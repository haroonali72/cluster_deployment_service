package gcp

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/key_utils"
	"antelope/models/types"
	userData2 "antelope/models/userData"
	"antelope/models/utils"
	"antelope/models/vault"
	"context"
	"encoding/json"
	"errors"
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

func (cloud *GCP) createCluster(cluster Cluster_Def, token string, ctx utils.Context) (Cluster_Def, types.CustomCPError) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return cluster, err
		}
	}

	var gcpNetwork types.GCPNetwork
	url := getNetworkHost("gcp", cluster.ProjectId)

	network, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, ApiErrors(err, "Error in cluster creation")
	}

	err = json.Unmarshal(network.([]byte), &gcpNetwork)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, ApiErrors(err, "Error in cluster creation")
	}

	cluster.NetworkName = gcpNetwork.Name
	for _, pool := range cluster.NodePools {
		beego.Info("GCPOperations creating nodes")

		if pool.PoolRole == "master" {
			err := cloud.deployMaster(cluster.ProjectId, pool, gcpNetwork, token, ctx)
			if err != (types.CustomCPError{}) && strings.Contains(err.Description, "was not found, notFound") {
				ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return cluster, types.CustomCPError{
					StatusCode:  502,
					Error:       "Instance creation failed due to limit reached out",
					Description: err.Description,
				}
			} else if err != (types.CustomCPError{}) {
				ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return cluster, err
			}
		} else {
			err := cloud.deployWorkers(cluster.ProjectId, pool, gcpNetwork, token, ctx)
			if err != (types.CustomCPError{}) && strings.Contains(err.Description, "was not found, notFound") {
				ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return cluster, types.CustomCPError{
					StatusCode:  502,
					Error:       "Instance creation failed due to limit reached out",
					Description: err.Description,
				}
			} else if err != (types.CustomCPError{}) {
				ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return cluster, err
			}
		}
	}

	return cluster, types.CustomCPError{}
}

func (cloud *GCP) deployMaster(projectId string, pool *NodePool, network types.GCPNetwork, token string, ctx utils.Context) types.CustomCPError {
	if cloud.Client == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}
	bytes, err := vault.GetSSHKey(string(models.GCP), pool.KeyInfo.KeyName, token, ctx, "")
	if err != nil {
		ctx.SendLogs("vm creation failed with error: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiErrors(err, "Error in deploying master")
	}
	fetchedKey, err := key_utils.AzureKeyConversion(bytes, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiErrors(err, "Error in deploying master")
	}
	pool.KeyInfo.PrivateKey = fetchedKey.PrivateKey
	pool.KeyInfo.PublicKey = fetchedKey.PublicKey

	pool.KeyInfo.PrivateKey = fetchedKey.PrivateKey
	pool.KeyInfo.PublicKey = fetchedKey.PublicKey
	zone := getSubnetZone(pool.PoolSubnet, network.Definition[0].Subnets)
	instance := compute.Instance{
		Name:        strings.ToLower(pool.Name),
		MachineType: "zones/" + zone + "/machineTypes/" + pool.MachineType,
		Tags: &compute.Tags{
			Items: append(pool.Tags, projectId),
		},
		Labels: map[string]string{
			"network": network.Definition[0].Vpc.Name,
		},
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				Subnetwork: getSubnet(pool.PoolSubnet, network.Definition[0].Subnets),
			},
		},
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				InitializeParams: &compute.AttachedDiskInitializeParams{
					SourceImage: "projects/" + pool.Image.Project + "/global/images/family/" + pool.Image.Family,
					DiskSizeGb:  pool.RootVolume.Size,
					DiskType:    "projects/" + pool.Image.Project + "/zones/" + zone + "/diskTypes/" + string(pool.RootVolume.DiskType),
				},
			},
		},
	}

	externalIp := ""
	var err1 types.CustomCPError
	if pool.EnablePublicIP {
		externalIp, err1 = cloud.reserveExternalIp(pool.Name, ctx)
		if err1 != (types.CustomCPError{}) {
			ctx.SendLogs(err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			beego.Warn("cannot reserve any external ip for: " + pool.Name)
			beego.Warn("creating instance '" + pool.Name + "' without external ip")
		}

		instance.NetworkInterfaces[0].AccessConfigs = []*compute.AccessConfig{
			{
				NatIP: externalIp,
			},
		}
	}
	items := []*compute.MetadataItems{
		{
			Key:   "block-project-ssh-keys",
			Value: to.StringPtr("true"),
		},
		{
			Key:   "ssh-keys",
			Value: to.StringPtr(pool.KeyInfo.Username + ":" + pool.KeyInfo.PublicKey),
		},
	}
	var fileName []string
	if pool.EnableVolume {
		fileName = append(fileName, "gcp-volume-mount.sh")
	}
	userData, err := userData2.GetUserData(token, getWoodpecker()+"/"+projectId, fileName, pool.PoolRole, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiErrors(err, "Error in deploying agent")
	}
	if userData != "no user data found" {
		items = append(items, &compute.MetadataItems{
			Key:   "user-data",
			Value: to.StringPtr(userData),
		})
	}

	instance.Metadata = &compute.Metadata{Items: items}
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
				DiskType:   "projects/" + pool.Image.Project + "/zones/" + zone + "/diskTypes/" + string(pool.Volume.DiskType),
			},
		}

		instance.Disks = append(instance.Disks, &secondaryDisk)
	}

	reqCtx := context.Background()
	result, err := cloud.Client.Instances.Insert(cloud.ProjectId, zone, &instance).Context(reqCtx).Do()
	if err != nil && !strings.Contains(err.Error(), "alreadyExists") {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiErrors(err, "Error in deploying instance")
	}

	err1 = cloud.waitForZonalCompletion(result, zone, ctx)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err1
	}

	newNode, err1 := cloud.fetchNodeInfo(instance.Name, zone, ctx)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs(err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err1
	}

	pool.Nodes = []*Node{&newNode}

	/*if pool.EnableVolume {
		err = mountVolume(pool.KeyInfo.PrivateKey, pool.KeyInfo.KeyName, pool.KeyInfo.Username, newNode.PublicIp)
		if err != nil {
			beego.Error(err.Error())
			return err
		}
	}*/

	return types.CustomCPError{}
}
func getWoodpecker() string {
	return beego.AppConfig.String("woodpecker_url") + models.WoodpeckerEnpoint
}
func (cloud *GCP) deployWorkers(projectId string, pool *NodePool, network types.GCPNetwork, token string, ctx utils.Context) types.CustomCPError {
	if cloud.Client == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	bytes, err := vault.GetSSHKey(string(models.GCP), pool.KeyInfo.KeyName, token, ctx, "")
	if err != nil {
		ctx.SendLogs("vm creation failed with error: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiErrors(err, "Error in deploying worker")
	}
	fetchedKey, err := key_utils.AzureKeyConversion(bytes, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiErrors(err, "Error in deploying worker")
	}
	pool.KeyInfo.PrivateKey = fetchedKey.PrivateKey
	pool.KeyInfo.PublicKey = fetchedKey.PublicKey

	instanceTemplateUrl, err1 := cloud.createInstanceTemplate(projectId, pool, network, token, ctx)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs(err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err1
	}

	instanceGroup := compute.InstanceGroupManager{
		Name:             strings.ToLower(pool.Name),
		BaseInstanceName: strings.ToLower(pool.Name),
		TargetSize:       pool.NodeCount,
		InstanceTemplate: instanceTemplateUrl,
	}
	zone := getSubnetZone(pool.PoolSubnet, network.Definition[0].Subnets)
	reqCtx := context.Background()
	result, err := cloud.Client.InstanceGroupManagers.Insert(cloud.ProjectId, zone, &instanceGroup).Context(reqCtx).Do()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiErrors(err, "Error in deploying worker")
	}

	err1 = cloud.waitForZonalCompletion(result, zone, ctx)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs(err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err1
	}

	createdNodes := &compute.InstanceGroupManagersListManagedInstancesResponse{}
	allNodesDeployed := false
	for !allNodesDeployed {
		time.Sleep(5 * time.Second)
		createdNodes, err = cloud.Client.InstanceGroupManagers.ListManagedInstances(cloud.ProjectId, zone, instanceGroup.Name).Context(reqCtx).Do()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return ApiErrors(err, "Error in deploying cluster")
		}
		for _, nodes := range createdNodes.ManagedInstances {
			if nodes.LastAttempt.Errors != nil && nodes.LastAttempt.Errors.Errors[0].Message != "QUOTA_EXCEEDED" {
				ctx.SendLogs(nodes.LastAttempt.Errors.Errors[0].Message, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return ApiErrors(errors.New(nodes.LastAttempt.Errors.Errors[0].Message), "Error in deploying cluster.Quota limit reached")
			} else if nodes.LastAttempt.Errors != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return ApiErrors(err, "Error in deploying cluster")
			}
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

		newNode, err := cloud.fetchNodeInfo(nodeName, zone, ctx)
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}

		pool.Nodes = append(pool.Nodes, &newNode)

		/*	if pool.EnableVolume {
			err = mountVolume(pool.KeyInfo.PrivateKey, pool.KeyInfo.KeyName, pool.KeyInfo.Username, newNode.PublicIp)
			if err != nil {
				beego.Error(err.Error())
				return err
			}
		}*/
	}

	return types.CustomCPError{}
}

func (cloud *GCP) createInstanceTemplate(projectId string, pool *NodePool, network types.GCPNetwork, token string, ctx utils.Context) (string, types.CustomCPError) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return "", err
		}
	}
	_, err := vault.GetSSHKey(string(models.GCP), pool.KeyInfo.KeyName, token, ctx, "")
	if err != nil {
		ctx.SendLogs("Key Not Found:  "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", ApiErrors(err, "Error in instance creation")
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	instanceProperties := compute.InstanceProperties{
		MachineType: pool.MachineType,
		Tags: &compute.Tags{
			Items: append(pool.Tags, projectId),
		},
		Labels: map[string]string{
			"network": network.Definition[0].Vpc.Name,
		},
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				Subnetwork: getSubnet(pool.PoolSubnet, network.Definition[0].Subnets),
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
	}
	items := []*compute.MetadataItems{
		{
			Key:   "block-project-ssh-keys",
			Value: to.StringPtr("true"),
		},
		{
			Key:   "ssh-keys",
			Value: to.StringPtr(pool.KeyInfo.Username + ":" + pool.KeyInfo.PublicKey),
		},
	}
	var fileName []string
	if pool.EnableVolume {
		fileName = append(fileName, "gcp-volume-mount.sh")
	}
	userData, err := userData2.GetUserData(token, getWoodpecker()+"/"+projectId, fileName, pool.PoolRole, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", ApiErrors(err, "Error in agent deployment")
	}
	if userData != "no user data found" {
		items = append(items, &compute.MetadataItems{
			Key:   "user-data",
			Value: to.StringPtr(userData),
		})
	}
	instanceProperties.Metadata = &compute.Metadata{Items: items}
	if pool.EnablePublicIP {
		instanceProperties.NetworkInterfaces[0].AccessConfigs = []*compute.AccessConfig{
			{
				Name: "External NAT",
				Type: "ONE_TO_ONE_NAT",
			},
		}
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
		return "", ApiErrors(err, "Error in instance creation")
	}

	err1 := cloud.waitForGlobalCompletion(result, ctx)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs(err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err1
	}

	createdTemplate, err := cloud.Client.InstanceTemplates.Get(cloud.ProjectId, instanceTemplate.Name).Context(reqCtx).Do()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", ApiErrors(err, "Error in instance creation")
	}

	return createdTemplate.SelfLink, types.CustomCPError{}
}

func (cloud *GCP) fetchNodeInfo(nodeName, zone string, ctx utils.Context) (Node, types.CustomCPError) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return Node{}, err
		}
	}

	reqCtx := context.Background()
	createdNode, err := cloud.Client.Instances.Get(cloud.ProjectId, zone, nodeName).Context(reqCtx).Do()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Node{}, ApiErrors(err, "Erro in fetching node info")
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

	return newNode, types.CustomCPError{}
}

func (cloud *GCP) deleteCluster(cluster Cluster_Def, token string, ctx utils.Context) types.CustomCPError {
	if cloud.Client == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	var gcpNetwork types.GCPNetwork
	url := getNetworkHost("gcp", cluster.ProjectId)

	network, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiErrors(err, "Error in cluster creation")
	}

	err = json.Unmarshal(network.([]byte), &gcpNetwork)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiErrors(err, "Error in cluster creation")
	}

	for _, pool := range cluster.NodePools {
		availabilityZone := getSubnetZone(pool.PoolSubnet, gcpNetwork.Definition[0].Subnets)

		err := cloud.deletePool(pool, availabilityZone, ctx)
		if err {
			return ApiErrors(errors.New("Error occured during termination"), "Error occured during termination")
		}
	}

	return types.CustomCPError{}
}

func (cloud *GCP) deletePool(pool *NodePool, zone string, ctx utils.Context) bool {
	error_occured := false
	if cloud.Client == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
			return true
		}
	}

	if pool.PoolRole == "master" {
		reqCtx := context.Background()
		result, err := cloud.Client.Instances.Delete(cloud.ProjectId, zone, pool.Name).Context(reqCtx).Do()
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "not found") {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				error_occured = true
			} else {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
			}
		} else {
			err1 := cloud.waitForZonalCompletion(result, zone, ctx)
			if err1 != (types.CustomCPError{}) {
				ctx.SendLogs(err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				error_occured = true
			}
		}

		err1 := cloud.releaseExternalIp(pool.Name, ctx)
		if err1 != (types.CustomCPError{}) {
			if !strings.Contains(strings.ToLower(err1.Description), "not found") {
				ctx.SendLogs(err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				error_occured = true
			} else {
				ctx.SendLogs(err1.Description, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
			}
		}
	} else {
		reqCtx := context.Background()

		result, err := cloud.Client.InstanceGroupManagers.Delete(cloud.ProjectId, zone, pool.Name).Context(reqCtx).Do()
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "not found") {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				error_occured = true
			} else {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
			}
		} else {
			err1 := cloud.waitForZonalCompletion(result, zone, ctx)
			if err1 != (types.CustomCPError{}) {
				ctx.SendLogs(err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				error_occured = true
			}
		}

		instanceGroupManager, err := cloud.Client.InstanceGroupManagers.Get(cloud.ProjectId, zone, pool.Name).Context(reqCtx).Do()

		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "not found") {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				error_occured = true
			} else {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
			}
		} else {
			if instanceGroupManager != nil {
				splits := strings.Split(instanceGroupManager.InstanceTemplate, "/")
				instanceTemplateName := splits[len(splits)-1]
				result, err := cloud.Client.InstanceTemplates.Delete(cloud.ProjectId, instanceTemplateName).Context(reqCtx).Do()
				if err != nil {
					if !strings.Contains(strings.ToLower(err.Error()), "not found") {
						ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
						error_occured = true
					} else {
						ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
					}
				} else {
					err1 := cloud.waitForZonalCompletion(result, zone, ctx)
					if err1 != (types.CustomCPError{}) {
						ctx.SendLogs(err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
						error_occured = true
					}
				}
			}
		}
	}

	pool.Nodes = []*Node{}
	return error_occured
}

func (cloud *GCP) fetchClusterStatus(cluster *Cluster_Def, token string, ctx utils.Context) types.CustomCPError {
	if cloud.Client == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	var gcpNetwork types.GCPNetwork
	url := getNetworkHost("gcp", cluster.ProjectId)

	network, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiErrors(err, "Error in cluster creation")
	}

	err = json.Unmarshal(network.([]byte), &gcpNetwork)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiErrors(err, "Error in cluster creation")
	}

	for _, pool := range cluster.NodePools {
		availabilityZone := getSubnetZone(pool.PoolSubnet, gcpNetwork.Definition[0].Subnets)
		//zone:=strings.Split(availabilityZone , "-")
		err := cloud.fetchPoolStatus(pool, availabilityZone, ctx)
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	return types.CustomCPError{}
}

func (cloud *GCP) fetchPoolStatus(pool *NodePool, zone string, ctx utils.Context) types.CustomCPError {
	if cloud.Client == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	reqCtx := context.Background()
	if pool.PoolRole == "master" {
		newNode, err := cloud.fetchNodeInfo(pool.Name, zone, ctx)
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}

		newNode.Username = pool.KeyInfo.Username
		pool.Nodes = []*Node{&newNode}
	} else {
		managedGroup, err := cloud.Client.InstanceGroupManagers.Get(cloud.ProjectId, zone, pool.Name).Context(reqCtx).Do()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return ApiErrors(err, "Erro in fechting pool status")
		}
		pool.PoolId = managedGroup.InstanceGroup
		createdNodes, err := cloud.Client.InstanceGroupManagers.ListManagedInstances(cloud.ProjectId, zone, pool.Name).Context(reqCtx).Do()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return ApiErrors(err, "Error in fetching nodepool status")
		}

		pool.Nodes = []*Node{}
		for _, node := range createdNodes.ManagedInstances {
			splits := strings.Split(node.Instance, "/")
			nodeName := splits[len(splits)-1]

			newNode, err := cloud.fetchNodeInfo(nodeName, zone, ctx)
			if err != (types.CustomCPError{}) {
				ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return err
			}

			newNode.Username = pool.KeyInfo.Username
			pool.Nodes = append(pool.Nodes, &newNode)
		}
	}

	return types.CustomCPError{}
}

func (cloud *GCP) reserveExternalIp(nodeName string, ctx utils.Context) (string, types.CustomCPError) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return "", err
		}
	}

	address := compute.Address{Name: "ip-" + strings.ToLower(nodeName) + "z"}
	reqCtx := context.Background()
	result, err := cloud.Client.Addresses.Insert(cloud.ProjectId, cloud.Region, &address).Context(reqCtx).Do()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", ApiErrors(err, "Error in deploying master")
	}

	err1 := cloud.waitForRegionalCompletion(result, cloud.Region, ctx)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs(err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err1
	}

	externalIp := ""
	for externalIp == "" {
		time.Sleep(1 * time.Second)
		result, err := cloud.Client.Addresses.Get(cloud.ProjectId, cloud.Region, address.Name).Context(reqCtx).Do()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return "", ApiErrors(err, "Error in deploying master")
		}
		externalIp = result.Address
	}

	return externalIp, types.CustomCPError{}
}

func (cloud *GCP) releaseExternalIp(nodeName string, ctx utils.Context) types.CustomCPError {
	if cloud.Client == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			return err
		}
	}

	addressName := "ip-" + strings.ToLower(nodeName) + "z"
	reqCtx := context.Background()
	result, err := cloud.Client.Addresses.Delete(cloud.ProjectId, cloud.Region, addressName).Context(reqCtx).Do()
	if err != nil {
		beego.Error(err.Error())
		return ApiErrors(err, "Error in pool deletion")
	}

	err1 := cloud.waitForRegionalCompletion(result, cloud.Region, ctx)
	if err1 != (types.CustomCPError{}) {
		beego.Error(err1.Description)
		return err1
	}

	return types.CustomCPError{}
}

func (cloud *GCP) listServiceAccounts(ctx utils.Context) ([]string, types.CustomCPError) {
	if cloud.Iam == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return nil, err
		}
	}

	reqCtx := context.Background()
	accounts, err := cloud.Iam.Projects.ServiceAccounts.List("projects/" + cloud.ProjectId).Context(reqCtx).Do()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, ApiErrors(err, "")
	}

	accountList := []string{}
	for _, account := range accounts.Accounts {
		accountList = append(accountList, account.Email)
	}

	return accountList, types.CustomCPError{}
}

func (cloud *GCP) waitForGlobalCompletion(op *compute.Operation, ctx utils.Context) types.CustomCPError {
	if op == nil {
		return types.CustomCPError{}
	}

	reqCtx := context.Background()
	status := ""
	for status != "DONE" {
		time.Sleep(5 * time.Second)
		result, err := cloud.Client.GlobalOperations.Get(cloud.ProjectId, op.Name).Context(reqCtx).Do()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		if result != nil {
			status = result.Status
		}
	}

	return types.CustomCPError{}
}

func (cloud *GCP) waitForRegionalCompletion(op *compute.Operation, region string, ctx utils.Context) types.CustomCPError {
	if op == nil {
		return types.CustomCPError{}
	}

	reqCtx := context.Background()
	status := ""
	for status != "DONE" {
		time.Sleep(5 * time.Second)
		result, err := cloud.Client.RegionOperations.Get(cloud.ProjectId, region, op.Name).Context(reqCtx).Do()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		if result != nil {
			status = result.Status
		}
	}

	return types.CustomCPError{}
}

func (cloud *GCP) waitForZonalCompletion(op *compute.Operation, zone string, ctx utils.Context) types.CustomCPError {
	if op == nil {
		return types.CustomCPError{}
	}

	reqCtx := context.Background()
	status := ""
	for status != "DONE" {
		time.Sleep(5 * time.Second)
		result, err := cloud.Client.ZoneOperations.Get(cloud.ProjectId, zone, op.Name).Context(reqCtx).Do()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		if result != nil {
			status = result.Status
		}
	}

	return types.CustomCPError{}
}

func (cloud *GCP) init() types.CustomCPError {
	if cloud.Client != nil {
		return types.CustomCPError{}
	}

	var err error
	ctx := context.Background()

	cloud.Client, err = compute.NewService(ctx, option.WithCredentialsJSON([]byte(cloud.Credentials)))
	if err != nil {
		beego.Error(err.Error())
		return ApiErrors(err, "Error in initializing APIs")
	}

	cloud.Iam, err = iam.NewService(ctx, option.WithCredentialsJSON([]byte(cloud.Credentials)))
	if err != nil {
		beego.Error(err.Error())
		return ApiErrors(err, "Error in initializing APIs")
	}

	return types.CustomCPError{}
}

func GetGCP(credentials GcpCredentials) (GCP, types.CustomCPError) {
	return GCP{
		Credentials: credentials.RawData,
		ProjectId:   credentials.AccountData.ProjectId,
		Region:      credentials.Region,
		Zone:        credentials.Zone,
	}, types.CustomCPError{}
}

func getSubnet(subnetName string, subnets []*types.Subnet) string {
	for _, subnet := range subnets {
		if subnet.Name == subnetName {
			return subnet.Link
		}
	}
	return ""
}

func getSubnetZone(subnetName string, subnets []*types.Subnet) string {
	for _, subnet := range subnets {
		if subnet.Name == subnetName {
			return subnet.Zone
		}
	}
	return ""
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

func (cloud *GCP) GetAllMachines(ctx utils.Context) (*compute.MachineTypeList, types.CustomCPError) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return &compute.MachineTypeList{}, err
		}
	}

	reqCtx := context.Background()

	machines, err := cloud.Client.MachineTypes.List(cloud.ProjectId, cloud.Region+"-"+cloud.Zone).Context(reqCtx).Do()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return &compute.MachineTypeList{}, ApiErrors(err, "Error in fetching machines")
	}

	return machines, types.CustomCPError{}
}

func (cloud *GCP) GetRegions(ctx utils.Context) (*compute.RegionList, types.CustomCPError) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return &compute.RegionList{}, err
		}
	}
	reqCtx := context.Background()
	regionInfo, err := cloud.Client.Regions.List(cloud.ProjectId).Context(reqCtx).Do()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return &compute.RegionList{}, ApiErrors(err, "Error in fetching ")
	}

	return regionInfo, types.CustomCPError{}
}
func (cloud *GCP) GetZones(ctx utils.Context) (*compute.Region, types.CustomCPError) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return &compute.Region{}, err
		}
	}

	reqCtx := context.Background()

	regionInfo, err := cloud.Client.Regions.Get(cloud.ProjectId, cloud.Region).Context(reqCtx).Do()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return &compute.Region{}, ApiErrors(err, "Error in fetching zones")
	}

	return regionInfo, types.CustomCPError{}
}
