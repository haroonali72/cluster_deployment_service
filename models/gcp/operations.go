package gcp

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/key_utils"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"context"
	"encoding/json"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/astaxie/beego"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type GCP struct {
	Client      *compute.Service
	Iam         *iam.Service
	Credentials string
	ProjectId   string
	Region      string
	Zone        string
}

func getNetworkHost(cloudType string) string {
	host := "http://" + beego.AppConfig.String("network_url") + "/weasel/network/{cloud_provider}"
	if strings.Contains(host, "{cloud_provider}") {
		host = strings.Replace(host, "{cloud_provider}", cloudType, -1)
	}
	return host
}

func (cloud *GCP) createCluster(cluster Cluster_Def, token string) (Cluster_Def, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return cluster, err
		}
	}

	var gcpNetwork types.GCPNetwork
	url := getNetworkHost("gcp") + "/" + cluster.ProjectId

	network, err := api_handler.GetAPIStatus(token, url, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		return cluster, err
	}

	err = json.Unmarshal(network.([]byte), &gcpNetwork)
	if err != nil {
		beego.Error(err.Error())
		return cluster, err
	}

	cluster.NetworkName = gcpNetwork.Name
	for _, pool := range cluster.NodePools {
		beego.Info("GCPOperations creating nodes")

		if pool.PoolRole == "master" {
			err = cloud.deployMaster(pool, gcpNetwork, token)
			if err != nil {
				beego.Error(err.Error())
				return cluster, err
			}
		} else {
			err = cloud.deployWorkers(pool, gcpNetwork, token)
			if err != nil {
				beego.Error(err.Error())
				return cluster, err
			}
		}
	}

	return cluster, nil
}

func (cloud *GCP) deployMaster(pool *NodePool, network types.GCPNetwork, token string) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return err
		}
	}

	privateKey, err := fetchOrGenerateKey(&pool.KeyInfo, token)
	if err != nil {
		return err
	}

	externalIp, err := cloud.reserveExternalIp(pool.Name)
	if err != nil {
		beego.Warn("cannot reserve any external ip for: " + pool.Name)
		beego.Warn("creating instance '" + pool.Name + "' without external ip")
	}

	instance := compute.Instance{
		Name:        strings.ToLower(pool.Name),
		MachineType: "zones/" + cloud.Region + "-" + cloud.Zone + "/machineTypes/" + pool.MachineType,
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				Subnetwork: getSubnet(pool.PoolSubnet, network.Definition[0].Subnets),
				AccessConfigs: []*compute.AccessConfig{
					{
						NatIP: externalIp,
					},
				},
			},
		},
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				InitializeParams: &compute.AttachedDiskInitializeParams{
					SourceImage: "projects/" + pool.Image.Project + "/global/images/family/" + pool.Image.Family,
					DiskSizeGb:  pool.RootVolume.Size,
					DiskType:    "projects/" + pool.Image.Project + "/zones/" + cloud.Region + "-" + cloud.Zone + "/diskTypes/" + string(pool.RootVolume.DiskType),
				},
			},
		},
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{
				{
					Key:   "block-project-ssh-keys",
					Value: to.StringPtr("true"),
				},
				{
					Key:   "ssh-keys",
					Value: to.StringPtr(pool.KeyInfo.Username + ":" + pool.KeyInfo.PublicKey),
				},
			},
		},
	}

	if pool.ServiceAccountEmail != "" {
		instance.ServiceAccounts = []*compute.ServiceAccount{
			{
				Email:  pool.ServiceAccountEmail,
				Scopes: []string{"https://www.googleapis.com/auth/compute"},
			},
		}
	}

	if pool.EnableVolume {
		secondaryDisk := compute.AttachedDisk{
			AutoDelete: true,
			Boot:       false,
			InitializeParams: &compute.AttachedDiskInitializeParams{
				DiskSizeGb: pool.Volume.Size,
				DiskType:   "projects/" + pool.Image.Project + "/zones/" + cloud.Region + "-" + cloud.Zone + "/diskTypes/" + string(pool.Volume.DiskType),
			},
		}

		instance.Disks = append(instance.Disks, &secondaryDisk)
	}

	ctx := context.Background()
	result, err := cloud.Client.Instances.Insert(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, &instance).Context(ctx).Do()
	if err != nil && !strings.Contains(err.Error(), "alreadyExists") {
		beego.Error(err.Error())
		return err
	}

	err = cloud.waitForZonalCompletion(result, cloud.Region+"-"+cloud.Zone)
	if err != nil {
		return err
	}

	newNode, err := cloud.fetchNodeInfo(instance.Name)
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	pool.Nodes = []*Node{&newNode}

	if pool.EnableVolume {
		err = mountVolume(privateKey, pool.KeyInfo.KeyName, pool.KeyInfo.Username, newNode.PublicIp)
		if err != nil {
			beego.Error(err.Error())
			return err
		}
	}

	return nil
}

func (cloud *GCP) deployWorkers(pool *NodePool, network types.GCPNetwork, token string) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return err
		}
	}

	privateKey, err := fetchOrGenerateKey(&pool.KeyInfo, token)
	if err != nil {
		return err
	}

	instanceTemplateUrl, err := cloud.createInstanceTemplate(pool, network, token)
	if err != nil {
		return err
	}

	instanceGroup := compute.InstanceGroupManager{
		Name:             strings.ToLower(pool.Name),
		BaseInstanceName: strings.ToLower(pool.Name),
		TargetSize:       pool.NodeCount,
		InstanceTemplate: instanceTemplateUrl,
	}

	ctx := context.Background()
	result, err := cloud.Client.InstanceGroupManagers.Insert(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, &instanceGroup).Context(ctx).Do()
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	err = cloud.waitForZonalCompletion(result, cloud.Region+"-"+cloud.Zone)
	if err != nil {
		return err
	}

	createdNodes := &compute.InstanceGroupManagersListManagedInstancesResponse{}
	allNodesDeployed := false
	for !allNodesDeployed {
		time.Sleep(5 * time.Second)
		createdNodes, err = cloud.Client.InstanceGroupManagers.ListManagedInstances(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, instanceGroup.Name).Context(ctx).Do()
		if err != nil {
			beego.Error(err.Error())
			return err
		}

		allNodesDeployed = true
		for _, node := range createdNodes.ManagedInstances {
			if node.InstanceStatus == "" {
				allNodesDeployed = false
				break
			}
		}
	}

	pool.Nodes = []*Node{}
	for _, node := range createdNodes.ManagedInstances {
		splits := strings.Split(node.Instance, "/")
		nodeName := splits[len(splits)-1]

		newNode, err := cloud.fetchNodeInfo(nodeName)
		if err != nil {
			beego.Error(err.Error())
			return err
		}

		pool.Nodes = append(pool.Nodes, &newNode)

		if pool.EnableVolume {
			err = mountVolume(privateKey, pool.KeyInfo.KeyName, pool.KeyInfo.Username, newNode.PublicIp)
			if err != nil {
				beego.Error(err.Error())
				return err
			}
		}
	}

	return nil
}

func (cloud *GCP) createInstanceTemplate(pool *NodePool, network types.GCPNetwork, token string) (string, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return "", err
		}
	}

	_, err := fetchOrGenerateKey(&pool.KeyInfo, token)
	if err != nil {
		return "", err
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	instanceProperties := compute.InstanceProperties{
		MachineType: pool.MachineType,
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				Subnetwork: getSubnet(pool.PoolSubnet, network.Definition[0].Subnets),
				AccessConfigs: []*compute.AccessConfig{
					{
						Name: "External NAT",
						Type: "ONE_TO_ONE_NAT",
					},
				},
			},
		},
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				InitializeParams: &compute.AttachedDiskInitializeParams{
					SourceImage: "projects/" + pool.Image.Project + "/global/images/family/" + pool.Image.Family,
					DiskSizeGb:  pool.RootVolume.Size,
					DiskType:    string(pool.RootVolume.DiskType),
				},
			},
		},
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{
				{
					Key:   "block-project-ssh-keys",
					Value: to.StringPtr("true"),
				},
				{
					Key:   "ssh-keys",
					Value: to.StringPtr(pool.KeyInfo.Username + ":" + pool.KeyInfo.PublicKey),
				},
			},
		},
	}

	if pool.ServiceAccountEmail != "" {
		instanceProperties.ServiceAccounts = []*compute.ServiceAccount{
			{
				Email:  pool.ServiceAccountEmail,
				Scopes: []string{"https://www.googleapis.com/auth/compute"},
			},
		}
	}

	if pool.EnableVolume {
		secondaryDisk := compute.AttachedDisk{
			AutoDelete: true,
			Boot:       false,
			InitializeParams: &compute.AttachedDiskInitializeParams{
				DiskSizeGb: pool.Volume.Size,
				DiskType:   string(pool.Volume.DiskType),
			},
		}

		instanceProperties.Disks = append(instanceProperties.Disks, &secondaryDisk)
	}

	instanceTemplate := compute.InstanceTemplate{
		Name:       strings.ToLower(pool.Name) + "-template" + strconv.Itoa(r1.Intn(1000)) + strconv.Itoa(r1.Intn(1000)),
		Properties: &instanceProperties,
	}

	ctx := context.Background()
	result, err := cloud.Client.InstanceTemplates.Insert(cloud.ProjectId, &instanceTemplate).Context(ctx).Do()
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "already exists") {
		beego.Error(err.Error())
		return "", err
	}

	err = cloud.waitForGlobalCompletion(result)
	if err != nil {
		return "", err
	}

	createdTemplate, err := cloud.Client.InstanceTemplates.Get(cloud.ProjectId, instanceTemplate.Name).Context(ctx).Do()
	if err != nil {
		beego.Error(err.Error())
		return "", err
	}

	return createdTemplate.SelfLink, nil
}

func (cloud *GCP) fetchNodeInfo(nodeName string) (Node, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return Node{}, err
		}
	}

	ctx := context.Background()
	createdNode, err := cloud.Client.Instances.Get(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, nodeName).Context(ctx).Do()
	if err != nil {
		beego.Error(err.Error())
		return Node{}, err
	}

	newNode := Node{
		CloudId:   strconv.Itoa(int(createdNode.Id)),
		Name:      nodeName,
		Url:       createdNode.SelfLink,
		NodeState: createdNode.Status,
	}

	if len(createdNode.NetworkInterfaces) > 0 {
		newNode.PrivateIp = createdNode.NetworkInterfaces[0].NetworkIP
		if len(createdNode.NetworkInterfaces[0].AccessConfigs) > 0 {
			newNode.PublicIp = createdNode.NetworkInterfaces[0].AccessConfigs[0].NatIP
		}
	}

	return newNode, nil
}

func (cloud *GCP) deleteCluster(cluster Cluster_Def) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return err
		}
	}

	for _, pool := range cluster.NodePools {
		err := cloud.deletePool(pool)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cloud *GCP) deletePool(pool *NodePool) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return err
		}
	}

	if pool.PoolRole == "master" {
		ctx := context.Background()
		result, err := cloud.Client.Instances.Delete(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, pool.Name).Context(ctx).Do()
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
			beego.Error(err.Error())
			return err
		}

		err = cloud.waitForZonalCompletion(result, cloud.Region+"-"+cloud.Zone)
		if err != nil {
			return err
		}

		err = cloud.releaseExternalIp(pool.Name)
		if err != nil {
			return err
		}
	} else {
		ctx := context.Background()
		instanceGroupManager, err := cloud.Client.InstanceGroupManagers.Get(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, pool.Name).Context(ctx).Do()
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
			beego.Error(err.Error())
			return err
		}

		result, err := cloud.Client.InstanceGroupManagers.Delete(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, pool.Name).Context(ctx).Do()
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
			beego.Error(err.Error())
			return err
		}

		err = cloud.waitForZonalCompletion(result, cloud.Region+"-"+cloud.Zone)
		if err != nil {
			return err
		}

		if instanceGroupManager != nil {
			splits := strings.Split(instanceGroupManager.InstanceTemplate, "/")
			instanceTemplateName := splits[len(splits)-1]
			result, err := cloud.Client.InstanceTemplates.Delete(cloud.ProjectId, instanceTemplateName).Context(ctx).Do()
			if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
				beego.Error(err.Error())
				return err
			}

			err = cloud.waitForGlobalCompletion(result)
			if err != nil {
				return err
			}
		}
	}

	pool.Nodes = []*Node{}
	return nil
}

func (cloud *GCP) fetchClusterStatus(cluster *Cluster_Def) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return err
		}
	}

	for _, pool := range cluster.NodePools {
		err := cloud.fetchPoolStatus(pool)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cloud *GCP) fetchPoolStatus(pool *NodePool) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return err
		}
	}

	ctx := context.Background()
	if pool.PoolRole == "master" {
		newNode, err := cloud.fetchNodeInfo(pool.Name)
		if err != nil {
			beego.Error(err.Error())
			return err
		}

		newNode.Username = pool.KeyInfo.Username
		pool.Nodes = []*Node{&newNode}
	} else {
		createdNodes, err := cloud.Client.InstanceGroupManagers.ListManagedInstances(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, pool.Name).Context(ctx).Do()
		if err != nil {
			beego.Error(err.Error())
			return err
		}

		pool.Nodes = []*Node{}
		for _, node := range createdNodes.ManagedInstances {
			splits := strings.Split(node.Instance, "/")
			nodeName := splits[len(splits)-1]

			newNode, err := cloud.fetchNodeInfo(nodeName)
			if err != nil {
				beego.Error(err.Error())
				return err
			}

			newNode.Username = pool.KeyInfo.Username
			pool.Nodes = append(pool.Nodes, &newNode)
		}
	}

	return nil
}

func (cloud *GCP) reserveExternalIp(nodeName string) (string, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return "", err
		}
	}

	address := compute.Address{Name: "ip-" + strings.ToLower(nodeName) + "z"}
	ctx := context.Background()
	result, err := cloud.Client.Addresses.Insert(cloud.ProjectId, cloud.Region, &address).Context(ctx).Do()
	if err != nil {
		beego.Error(err.Error())
		return "", err
	}

	err = cloud.waitForRegionalCompletion(result, cloud.Region)
	if err != nil {
		beego.Error(err.Error())
		return "", err
	}

	externalIp := ""
	for externalIp == "" {
		time.Sleep(1 * time.Second)
		result, err := cloud.Client.Addresses.Get(cloud.ProjectId, cloud.Region, address.Name).Context(ctx).Do()
		if err != nil {
			beego.Error(err.Error())
			return "", err
		}
		externalIp = result.Address
	}

	return externalIp, nil
}

func (cloud *GCP) releaseExternalIp(nodeName string) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return err
		}
	}

	addressName := "ip-" + strings.ToLower(nodeName) + "z"
	ctx := context.Background()
	result, err := cloud.Client.Addresses.Delete(cloud.ProjectId, cloud.Region, addressName).Context(ctx).Do()
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	err = cloud.waitForRegionalCompletion(result, cloud.Region)
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}

func (cloud *GCP) listServiceAccounts() ([]string, error) {
	if cloud.Iam == nil {
		err := cloud.init()
		if err != nil {
			return nil, err
		}
	}

	ctx := context.Background()
	accounts, err := cloud.Iam.Projects.ServiceAccounts.List("projects/" + cloud.ProjectId).Context(ctx).Do()
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}

	accountList := []string{}
	for _, account := range accounts.Accounts {
		accountList = append(accountList, account.Email)
	}

	return accountList, nil
}

func (cloud *GCP) waitForGlobalCompletion(op *compute.Operation) error {
	if op == nil {
		return nil
	}

	ctx := context.Background()
	status := ""
	for status != "DONE" {
		time.Sleep(5 * time.Second)
		result, err := cloud.Client.GlobalOperations.Get(cloud.ProjectId, op.Name).Context(ctx).Do()
		if err != nil {
			beego.Error(err.Error())
		}
		if result != nil {
			status = result.Status
		}
	}

	return nil
}

func (cloud *GCP) waitForRegionalCompletion(op *compute.Operation, region string) error {
	if op == nil {
		return nil
	}

	ctx := context.Background()
	status := ""
	for status != "DONE" {
		time.Sleep(5 * time.Second)
		result, err := cloud.Client.RegionOperations.Get(cloud.ProjectId, region, op.Name).Context(ctx).Do()
		if err != nil {
			beego.Error(err.Error())
		}
		if result != nil {
			status = result.Status
		}
	}

	return nil
}

func (cloud *GCP) waitForZonalCompletion(op *compute.Operation, zone string) error {
	if op == nil {
		return nil
	}

	ctx := context.Background()
	status := ""
	for status != "DONE" {
		time.Sleep(5 * time.Second)
		result, err := cloud.Client.ZoneOperations.Get(cloud.ProjectId, zone, op.Name).Context(ctx).Do()
		if err != nil {
			beego.Error(err.Error())
		}
		if result != nil {
			status = result.Status
		}
	}

	return nil
}

func (cloud *GCP) init() error {
	if cloud.Client != nil {
		return nil
	}

	var err error
	ctx := context.Background()

	cloud.Client, err = compute.NewService(ctx, option.WithCredentialsJSON([]byte(cloud.Credentials)))
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	cloud.Iam, err = iam.NewService(ctx, option.WithCredentialsJSON([]byte(cloud.Credentials)))
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}

func GetGCP(credentials GcpCredentials) (GCP, error) {
	return GCP{
		Credentials: credentials.RawData,
		ProjectId:   credentials.AccountData.ProjectId,
		Region:      credentials.Region,
		Zone:        credentials.Zone,
	}, nil
}

func getSubnet(subnetName string, subnets []*types.Subnet) string {
	for _, subnet := range subnets {
		if subnet.Name == subnetName {
			return subnet.Link
		}
	}
	return ""
}

func fetchOrGenerateKey(keyInfo *utils.Key, token string) (string, error) {
	key, err := vault.GetAzureSSHKey(string(models.GCP), keyInfo.KeyName, token, utils.Context{})

	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		beego.Error("vm creation failed with error: " + err.Error())
		return "", err
	}

	existingKey, err := key_utils.KeyConversion(key, utils.Context{})
	if err != nil {
		beego.Error("vm creation failed with error: " + err.Error())
		return "", err
	}

	username := "cloudplex"
	if keyInfo.Username != "" {
		username = keyInfo.Username
	}

	if existingKey.PublicKey != "" && existingKey.PrivateKey != "" {
		keyInfo.PrivateKey = existingKey.PrivateKey
		keyInfo.PublicKey = strings.TrimSuffix(existingKey.PublicKey, "\n")

		keySplits := strings.Split(keyInfo.PublicKey, " ")
		if len(keySplits) >= 3 && keySplits[2] != username {
			keyInfo.PublicKey = keySplits[0] + " " + keySplits[1] + " " + username
		}

		return keyInfo.PrivateKey, nil
	}

	res, err := key_utils.GenerateKeyPair(keyInfo.KeyName, username, utils.Context{})
	if err != nil {
		beego.Error("vm creation failed with error: " + err.Error())
		return "", err
	}

	keyInfo.Username = username
	keyInfo.Cloud = models.GCP
	keyInfo.PrivateKey = res.PrivateKey
	keyInfo.PublicKey = strings.TrimSuffix(res.PublicKey, "\n")

	_, err = vault.PostGcpSSHKey(keyInfo, utils.Context{}, token)
	if err != nil {
		beego.Error("vm creation failed with error: " + err.Error())
		return "", err
	}

	return keyInfo.PrivateKey, nil
}

func mountVolume(privateKey, keyName, username, ipAddress string) error {
	t := time.Now().Local()
	tstamp := t.Format("20060102150405")
	sshKeyFileName := "/app/keys/" + keyName + "_" + tstamp + ".pem"
	connectionString := username + "@" + ipAddress

	err := writeFile(privateKey, sshKeyFileName)
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	err = setFilePermission(sshKeyFileName)
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	err = copyScriptFile(sshKeyFileName, connectionString+":/home/"+username)
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	errPer := setScriptPermission(sshKeyFileName, username, connectionString)
	if errPer != nil {
		beego.Error(errPer.Error())
	}

	errCmd := runScript(sshKeyFileName, username, connectionString)
	if errCmd != nil {
		beego.Error(errCmd.Error())
	}

	err = deleteScript(sshKeyFileName, username, connectionString)
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	err = deleteFile(sshKeyFileName)
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return errCmd
}

func copyScriptFile(sshKeyFileName, connectionString string) error {
	args := []string{
		"-o",
		"StrictHostKeyChecking=no",
		"-i",
		sshKeyFileName,
		"/app/scripts/gcp-volume-mount.sh",
		connectionString,
	}

	i := 0
	for i < 5 {
		cmd := exec.Command("scp", args...)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			beego.Warn("error while copying script, sleeping 20s before retrying", err.Error())
			time.Sleep(20 * time.Second)
		} else {
			break
		}
		i++
	}

	return nil
}

func setScriptPermission(sshKeyFileName, username, connectionString string) error {
	args := []string{
		"-o",
		"StrictHostKeyChecking=no",
		"-i",
		sshKeyFileName,
		connectionString,
		"chmod 700 /home/" + username + "/gcp-volume-mount.sh",
	}
	cmd := exec.Command("ssh", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		beego.Error(err.Error())
		return nil
	}

	return nil
}

func runScript(sshKeyFileName, username, connectionString string) error {
	args := []string{
		"-o",
		"StrictHostKeyChecking=no",
		"-i",
		sshKeyFileName,
		connectionString,
		"/home/" + username + "/gcp-volume-mount.sh",
	}
	cmd := exec.Command("ssh", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		beego.Error(err.Error())
		return nil
	}

	return nil
}

func deleteScript(sshKeyFileName, username, connectionString string) error {
	args := []string{
		"-o",
		"StrictHostKeyChecking=no",
		"-i",
		sshKeyFileName,
		connectionString,
		"rm",
		"/home/" + username + "/gcp-volume-mount.sh",
	}
	cmd := exec.Command("ssh", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	return nil
}

func writeFile(content string, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	defer file.Close()

	fileContent := []byte(content)
	_, err = file.Write(fileContent)
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}

func setFilePermission(fileName string) error {
	args := []string{"600", fileName}
	cmd := exec.Command("chmod", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	return nil
}

func deleteFile(fileName string) error {
	err := os.Remove(fileName)
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}
