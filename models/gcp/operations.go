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
	"google.golang.org/api/option"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type GCP struct {
	Client      *compute.Service
	Credentials string
	ProjectId   string
	Region      string
	Zone        string
}

func getNetworkHost(cloudType string) string {
	host := beego.AppConfig.String("network_url")
	if strings.Contains(host, "{cloud_provider}") {
		host = strings.Replace(host, "{cloud_provider}", cloudType, -1)
	}
	return host
}

func (cloud *GCP) createCluster(cluster Cluster_Def) (Cluster_Def, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return cluster, err
		}
	}

	var gcpNetwork types.GCPNetwork
	url := getNetworkHost("gcp") + "/" + cluster.ProjectId

	network, err := api_handler.GetAPIStatus(url, utils.Context{})
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
			err = cloud.deployMaster(pool, gcpNetwork)
			if err != nil {
				beego.Error(err.Error())
				return cluster, err
			}
		} else {
			err = cloud.deployWorkers(pool, gcpNetwork)
			if err != nil {
				beego.Error(err.Error())
				return cluster, err
			}
		}
	}

	return cluster, nil
}

func (cloud *GCP) deployMaster(pool *NodePool, network types.GCPNetwork) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return err
		}
	}

	err := fetchOrGenerateKey(&pool.KeyInfo)
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

	if pool.EnableVolume {
		secondaryDisk := compute.AttachedDisk{
			AutoDelete: true,
			Boot:       false,
			InitializeParams: &compute.AttachedDiskInitializeParams{
				DiskSizeGb: pool.Volume.Size,
				DiskType:   "projects/" + pool.Image.Project + "/zones/" + cloud.Region + "-" + cloud.Zone + "/diskTypes/" + string(pool.Volume.DiskType),
			},
		}

		if !pool.Volume.IsBlank {
			secondaryDisk.InitializeParams.SourceImage = "projects/" + pool.Image.Project + "/global/images/family/" + pool.Image.Family
		}

		instance.Disks = append(instance.Disks, &secondaryDisk)
	}

	ctx := context.Background()
	result, err := cloud.Client.Instances.Insert(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, &instance).Context(ctx).Do()
	if err != nil {
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

	return nil
}

func (cloud *GCP) deployWorkers(pool *NodePool, network types.GCPNetwork) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return err
		}
	}

	instanceTemplateUrl, err := cloud.createInstanceTemplate(pool, network)
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
	}

	return nil
}

func (cloud *GCP) createInstanceTemplate(pool *NodePool, network types.GCPNetwork) (string, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return "", err
		}
	}

	err := fetchOrGenerateKey(&pool.KeyInfo)
	if err != nil {
		return "", err
	}

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

	if pool.EnableVolume {
		secondaryDisk := compute.AttachedDisk{
			AutoDelete: true,
			Boot:       false,
			InitializeParams: &compute.AttachedDiskInitializeParams{
				DiskSizeGb: pool.Volume.Size,
				DiskType:   string(pool.Volume.DiskType),
			},
		}

		if !pool.Volume.IsBlank {
			secondaryDisk.InitializeParams.SourceImage = "projects/" + pool.Image.Project + "/global/images/family/" + pool.Image.Family
		}

		instanceProperties.Disks = append(instanceProperties.Disks, &secondaryDisk)
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

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

func fetchOrGenerateKey(keyInfo *utils.Key) error {
	key, err := vault.GetAzureSSHKey(string(models.GCP), keyInfo.KeyName, utils.Context{})
	
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		beego.Error("vm creation failed with error: " + err.Error())
		return err
	}
	
	existingKey, err := key_utils.KeyConversion(key, utils.Context{})
	if err != nil {
		beego.Error("vm creation failed with error: " + err.Error())
		return err
	}
	
	username := "user@cloudplex.io"
	if keyInfo.Username != "" {
		username = keyInfo.Username
	}

	if existingKey.PublicKey != "" && existingKey.PrivateKey != "" {
		keyInfo.PrivateKey = existingKey.PrivateKey
		keyInfo.PublicKey = strings.TrimSuffix(existingKey.PublicKey, "\n")
		return nil
	}

	res, err := key_utils.GenerateKeyPair(keyInfo.KeyName, username, utils.Context{})
	if err != nil {
		beego.Error("vm creation failed with error: " + err.Error())
		return err
	}

	keyInfo.Username = username
	keyInfo.Cloud = models.GCP
	keyInfo.PrivateKey = res.PrivateKey
	keyInfo.PublicKey = strings.TrimSuffix(res.PublicKey, "\n")

	_, err = vault.PostGcpSSHKey(keyInfo, utils.Context{})
	if err != nil {
		beego.Error("vm creation failed with error: " + err.Error())
		return err
	}

	return nil
}
