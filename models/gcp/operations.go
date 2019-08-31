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

func getNetworkHost(cloudType, projectId string) string {
	host := beego.AppConfig.String("network_url") + models.WeaselGetEndpoint

	if strings.Contains(host, "{cloud}") {
		host = strings.Replace(host, "{cloud}", cloudType, -1)
	}
	if strings.Contains(host, "{projectId}") {
		host = strings.Replace(host, "{projectId}", projectId, -1)
	}
	return host
}

func (cloud *GCP) createCluster(cluster Cluster_Def, token string, ctx utils.Context) (Cluster_Def, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return cluster, err
		}
	}

	var gcpNetwork types.GCPNetwork
	url := getNetworkHost("gcp", cluster.ProjectId)

	network, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return cluster, err
	}

	err = json.Unmarshal(network.([]byte), &gcpNetwork)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return cluster, err
	}

	cluster.NetworkName = gcpNetwork.Name
	for _, pool := range cluster.NodePools {
		beego.Info("GCPOperations creating nodes")

		if pool.PoolRole == "master" {
			err = cloud.deployMaster(pool, gcpNetwork, token, ctx)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				beego.Error(err.Error())
				return cluster, err
			}
		} else {
			err = cloud.deployWorkers(pool, gcpNetwork, token, ctx)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				beego.Error(err.Error())
				return cluster, err
			}
		}
	}

	return cluster, nil
}

func (cloud *GCP) deployMaster(pool *NodePool, network types.GCPNetwork, token string, ctx utils.Context) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	privateKey, err := fetchOrGenerateKey(pool.KeyInfo.KeyName, token, "", ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	externalIp, err := cloud.reserveExternalIp(pool.Name, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	reqCtx := context.Background()
	result, err := cloud.Client.Instances.Insert(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, &instance).Context(reqCtx).Do()
	if err != nil && !strings.Contains(err.Error(), "alreadyExists") {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return err
	}

	err = cloud.waitForZonalCompletion(result, cloud.Region+"-"+cloud.Zone, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	newNode, err := cloud.fetchNodeInfo(instance.Name, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

func (cloud *GCP) deployWorkers(pool *NodePool, network types.GCPNetwork, token string, ctx utils.Context) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	privateKey, err := fetchOrGenerateKey(pool.KeyInfo.KeyName, token, "", ctx)
	if err != nil {
		return err
	}

	instanceTemplateUrl, err := cloud.createInstanceTemplate(pool, network, token, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	instanceGroup := compute.InstanceGroupManager{
		Name:             strings.ToLower(pool.Name),
		BaseInstanceName: strings.ToLower(pool.Name),
		TargetSize:       pool.NodeCount,
		InstanceTemplate: instanceTemplateUrl,
	}

	reqCtx := context.Background()
	result, err := cloud.Client.InstanceGroupManagers.Insert(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, &instanceGroup).Context(reqCtx).Do()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return err
	}

	err = cloud.waitForZonalCompletion(result, cloud.Region+"-"+cloud.Zone, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	createdNodes := &compute.InstanceGroupManagersListManagedInstancesResponse{}
	allNodesDeployed := false
	for !allNodesDeployed {
		time.Sleep(5 * time.Second)
		createdNodes, err = cloud.Client.InstanceGroupManagers.ListManagedInstances(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, instanceGroup.Name).Context(reqCtx).Do()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

		newNode, err := cloud.fetchNodeInfo(nodeName, ctx)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

func (cloud *GCP) createInstanceTemplate(pool *NodePool, network types.GCPNetwork, token string, ctx utils.Context) (string, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return "", err
		}
	}

	_, err := fetchOrGenerateKey(pool.KeyInfo.KeyName, token, "", ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	reqCtx := context.Background()
	result, err := cloud.Client.InstanceTemplates.Insert(cloud.ProjectId, &instanceTemplate).Context(reqCtx).Do()
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "already exists") {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return "", err
	}

	err = cloud.waitForGlobalCompletion(result, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}

	createdTemplate, err := cloud.Client.InstanceTemplates.Get(cloud.ProjectId, instanceTemplate.Name).Context(reqCtx).Do()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return "", err
	}

	return createdTemplate.SelfLink, nil
}

func (cloud *GCP) fetchNodeInfo(nodeName string, ctx utils.Context) (Node, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return Node{}, err
		}
	}

	reqCtx := context.Background()
	createdNode, err := cloud.Client.Instances.Get(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, nodeName).Context(reqCtx).Do()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

func (cloud *GCP) deleteCluster(cluster Cluster_Def, ctx utils.Context) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	for _, pool := range cluster.NodePools {
		err := cloud.deletePool(pool, ctx)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	return nil
}

func (cloud *GCP) deletePool(pool *NodePool, ctx utils.Context) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return err
		}
	}

	if pool.PoolRole == "master" {
		reqCtx := context.Background()
		result, err := cloud.Client.Instances.Delete(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, pool.Name).Context(reqCtx).Do()
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
			beego.Error(err.Error())
			return err
		}

		err = cloud.waitForZonalCompletion(result, cloud.Region+"-"+cloud.Zone, ctx)
		if err != nil {
			return err
		}

		err = cloud.releaseExternalIp(pool.Name, ctx)
		if err != nil {
			return err
		}
	} else {
		reqCtx := context.Background()
		instanceGroupManager, err := cloud.Client.InstanceGroupManagers.Get(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, pool.Name).Context(reqCtx).Do()
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
			beego.Error(err.Error())
			return err
		}

		result, err := cloud.Client.InstanceGroupManagers.Delete(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, pool.Name).Context(reqCtx).Do()
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
			beego.Error(err.Error())
			return err
		}

		err = cloud.waitForZonalCompletion(result, cloud.Region+"-"+cloud.Zone, ctx)
		if err != nil {
			return err
		}

		if instanceGroupManager != nil {
			splits := strings.Split(instanceGroupManager.InstanceTemplate, "/")
			instanceTemplateName := splits[len(splits)-1]
			result, err := cloud.Client.InstanceTemplates.Delete(cloud.ProjectId, instanceTemplateName).Context(reqCtx).Do()
			if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
				beego.Error(err.Error())
				return err
			}

			err = cloud.waitForGlobalCompletion(result, ctx)
			if err != nil {
				return err
			}
		}
	}

	pool.Nodes = []*Node{}
	return nil
}

func (cloud *GCP) fetchClusterStatus(cluster *Cluster_Def, ctx utils.Context) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	for _, pool := range cluster.NodePools {
		err := cloud.fetchPoolStatus(pool, ctx)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	return nil
}

func (cloud *GCP) fetchPoolStatus(pool *NodePool, ctx utils.Context) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	reqCtx := context.Background()
	if pool.PoolRole == "master" {
		newNode, err := cloud.fetchNodeInfo(pool.Name, ctx)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			beego.Error(err.Error())
			return err
		}

		newNode.Username = pool.KeyInfo.Username
		pool.Nodes = []*Node{&newNode}
	} else {
		managedGroup, err := cloud.Client.InstanceGroupManagers.Get(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, pool.Name).Context(reqCtx).Do()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			beego.Error(err.Error())
			return err
		}
		pool.PoolId = managedGroup.InstanceGroup
		createdNodes, err := cloud.Client.InstanceGroupManagers.ListManagedInstances(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, pool.Name).Context(reqCtx).Do()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			beego.Error(err.Error())
			return err
		}

		pool.Nodes = []*Node{}
		for _, node := range createdNodes.ManagedInstances {
			splits := strings.Split(node.Instance, "/")
			nodeName := splits[len(splits)-1]

			newNode, err := cloud.fetchNodeInfo(nodeName, ctx)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				beego.Error(err.Error())
				return err
			}

			newNode.Username = pool.KeyInfo.Username
			pool.Nodes = append(pool.Nodes, &newNode)
		}
	}

	return nil
}

func (cloud *GCP) reserveExternalIp(nodeName string, ctx utils.Context) (string, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return "", err
		}
	}

	address := compute.Address{Name: "ip-" + strings.ToLower(nodeName) + "z"}
	reqCtx := context.Background()
	result, err := cloud.Client.Addresses.Insert(cloud.ProjectId, cloud.Region, &address).Context(reqCtx).Do()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return "", err
	}

	err = cloud.waitForRegionalCompletion(result, cloud.Region, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return "", err
	}

	externalIp := ""
	for externalIp == "" {
		time.Sleep(1 * time.Second)
		result, err := cloud.Client.Addresses.Get(cloud.ProjectId, cloud.Region, address.Name).Context(reqCtx).Do()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			beego.Error(err.Error())
			return "", err
		}
		externalIp = result.Address
	}

	return externalIp, nil
}

func (cloud *GCP) releaseExternalIp(nodeName string, ctx utils.Context) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return err
		}
	}

	addressName := "ip-" + strings.ToLower(nodeName) + "z"
	reqCtx := context.Background()
	result, err := cloud.Client.Addresses.Delete(cloud.ProjectId, cloud.Region, addressName).Context(reqCtx).Do()
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	err = cloud.waitForRegionalCompletion(result, cloud.Region, ctx)
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}

func (cloud *GCP) listServiceAccounts(ctx utils.Context) ([]string, error) {
	if cloud.Iam == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return nil, err
		}
	}

	reqCtx := context.Background()
	accounts, err := cloud.Iam.Projects.ServiceAccounts.List("projects/" + cloud.ProjectId).Context(reqCtx).Do()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return nil, err
	}

	accountList := []string{}
	for _, account := range accounts.Accounts {
		accountList = append(accountList, account.Email)
	}

	return accountList, nil
}

func (cloud *GCP) waitForGlobalCompletion(op *compute.Operation, ctx utils.Context) error {
	if op == nil {
		return nil
	}

	reqCtx := context.Background()
	status := ""
	for status != "DONE" {
		time.Sleep(5 * time.Second)
		result, err := cloud.Client.GlobalOperations.Get(cloud.ProjectId, op.Name).Context(reqCtx).Do()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			beego.Error(err.Error())
		}
		if result != nil {
			status = result.Status
		}
	}

	return nil
}

func (cloud *GCP) waitForRegionalCompletion(op *compute.Operation, region string, ctx utils.Context) error {
	if op == nil {
		return nil
	}

	reqCtx := context.Background()
	status := ""
	for status != "DONE" {
		time.Sleep(5 * time.Second)
		result, err := cloud.Client.RegionOperations.Get(cloud.ProjectId, region, op.Name).Context(reqCtx).Do()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			beego.Error(err.Error())
		}
		if result != nil {
			status = result.Status
		}
	}

	return nil
}

func (cloud *GCP) waitForZonalCompletion(op *compute.Operation, zone string, ctx utils.Context) error {
	if op == nil {
		return nil
	}

	reqCtx := context.Background()
	status := ""
	for status != "DONE" {
		time.Sleep(5 * time.Second)
		result, err := cloud.Client.ZoneOperations.Get(cloud.ProjectId, zone, op.Name).Context(reqCtx).Do()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

func fetchOrGenerateKey(keyName, token, teams string, ctx utils.Context) (string, error) {
	var keyInfo utils.Key
	key, err := vault.GetAzureSSHKey(string(models.GCP), keyName, token, teams, ctx)

	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {

		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return "", err
	}
	existingKey, err := key_utils.KeyConversion(key, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("vm creation failed with error: " + err.Error())
		return "", err
	}

	username := "cloudplex"
	if existingKey.PublicKey != "" && existingKey.PrivateKey != "" {
		keyInfo.PrivateKey = existingKey.PrivateKey
		keyInfo.PublicKey = strings.TrimSuffix(existingKey.PublicKey, "\n")

		keySplits := strings.Split(keyInfo.PublicKey, " ")
		if len(keySplits) >= 3 && keySplits[2] != username {
			keyInfo.PublicKey = keySplits[0] + " " + keySplits[1] + " " + username
		}

		return keyInfo.PrivateKey, nil
	}

	res, err := key_utils.GenerateKeyPair(keyName, username, ctx)
	if err != nil {
		beego.Error("vm creation failed with error: " + err.Error())
		return "", err
	}

	keyInfo.KeyName = keyName
	keyInfo.Username = username
	keyInfo.Cloud = models.GCP
	keyInfo.PrivateKey = res.PrivateKey
	keyInfo.PublicKey = strings.TrimSuffix(res.PublicKey, "\n")
	beego.Info("Private Key in fetch ", keyInfo.PrivateKey)
	_, err = vault.PostGcpSSHKey(keyInfo, ctx, token)
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
