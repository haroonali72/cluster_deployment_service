package gke

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/gcp"
	"antelope/models/types"
	"antelope/models/utils"
	"context"
	"encoding/json"
	"github.com/astaxie/beego"
	"google.golang.org/api/compute/v1"
	gke "google.golang.org/api/container/v1"
	"google.golang.org/api/option"
	"strings"
	"time"
)

const (
	statusRunning = "RUNNING"
)

type GKE struct {
	Client      *gke.Service
	Compute     *compute.Service
	Credentials string
	ProjectId   string
	Region      string
	Zone        string
}

func (cloud *GKE) ListClusters(ctx utils.Context) ([]GKECluster, types.CustomCPError) {
	if cloud.Client == nil {
		err := cloud.init()
		if err.Description != "" {
			return nil, err
		}
	}

	list, err := cloud.Client.Projects.Zones.Clusters.List(cloud.ProjectId, cloud.Region+"-"+cloud.Zone).Do()
	if err != nil {
		ctx.SendLogs(
			"GKE list clusters for '"+cloud.ProjectId+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return nil, ApiErrors(err, "Error in listing running clusters")
	}

	result := []GKECluster{}
	for _, v := range list.Clusters {
		if v != nil {
			result = append(result, GenerateClusterFromResponse(*v))
		}
	}

	return result, types.CustomCPError{}
}

func (cloud *GKE) CreateCluster(gkeCluster GKECluster, token string, ctx utils.Context) types.CustomCPError {

	clusterRequest := GenerateClusterCreateRequest(cloud.ProjectId, cloud.Region, cloud.Zone, gkeCluster)
	networkInformation := cloud.getGCPNetwork(token, ctx)

	// overriding network configurations with network from current project
	if len(networkInformation.Definition) > 0 {
		clusterRequest.Cluster.Network = networkInformation.Definition[0].Vpc.Name
		if len(networkInformation.Definition[0].Subnets) > 0 {
			clusterRequest.Cluster.Subnetwork = networkInformation.Definition[0].Subnets[0].Name
		}
	}

	_, err := cloud.Client.Projects.Zones.Clusters.Create(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, clusterRequest).Context(context.Background()).Do()

	requestJson, _ := json.Marshal(clusterRequest)
	ctx.SendLogs(
		"GKE cluster creation of "+gkeCluster.Name+" submitted: "+string(requestJson),
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)

	if err != nil && !strings.Contains(err.Error(), "alreadyExists") {
		ctx.SendLogs(
			"GKE cluster creation of '"+gkeCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return types.CustomCPError{
			StatusCode:  500,
			Error:       "Error in cluster creation",
			Description: err.Error(),
		}
	}

	return cloud.waitForCluster(gkeCluster.Name, ctx)
}

func (cloud *GKE) UpdateMasterVersion(clusterName, newVersion string, ctx utils.Context) types.CustomCPError {
	if newVersion == "" {
		return types.CustomCPError{}
	}

	_, err := cloud.Client.Projects.Zones.Clusters.Update(
		cloud.ProjectId,
		cloud.Region+"-"+cloud.Zone,
		clusterName,
		&gke.UpdateClusterRequest{
			Update: &gke.ClusterUpdate{
				DesiredMasterVersion: newVersion,
			},
		},
	).Context(context.Background()).Do()
	if err != nil {
		ctx.SendLogs(
			"GKE running cluster version update request of '"+clusterName+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return types.CustomCPError{
			StatusCode:  512,
			Error:       "Error in running cluster version update",
			Description: err.Error(),
		}
	}

	return cloud.waitForCluster(clusterName, ctx)
}

func (cloud *GKE) UpdateResourceUsageExportConfig(clusterName string, ctx utils.Context,resourceConfig ResourceUsageExportConfig) types.CustomCPError {
	if clusterName == "" || resourceConfig == (ResourceUsageExportConfig{}){
		return types.CustomCPError{}
	}

	response :=  GenerateResourceUsageExportConfigFromRequest(&resourceConfig)

	_, err := cloud.Client.Projects.Zones.Clusters.Update(
		cloud.ProjectId,
		cloud.Region+"-"+cloud.Zone,
		clusterName,
		&gke.UpdateClusterRequest{
			Update: &gke.ClusterUpdate{
				DesiredResourceUsageExportConfig: &gke.ResourceUsageExportConfig{
					BigqueryDestination:         response.BigqueryDestination,
					ConsumptionMeteringConfig:  response.ConsumptionMeteringConfig,
					EnableNetworkEgressMetering: response.EnableNetworkEgressMetering,
				},
			},
		},
	).Context(context.Background()).Do()
	if err != nil {
		ctx.SendLogs(
			"GKE running cluster resource usage export config update request of "+clusterName+" failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return types.CustomCPError{
			StatusCode:  512,
			Error:       "Error in running cluster resource usage export config update",
			Description: err.Error(),
		}
	}

	return cloud.waitForCluster(clusterName, ctx)
}

func (cloud *GKE) UpdateNodePoolVersion(clusterName, nodeName, newVersion string, ctx utils.Context) types.CustomCPError {
	if newVersion == "" {
		return types.CustomCPError{}
	}

	_, err := cloud.Client.Projects.Zones.Clusters.NodePools.Update(
		cloud.ProjectId,
		cloud.Region+"-"+cloud.Zone,
		clusterName,
		nodeName,
		&gke.UpdateNodePoolRequest{
			NodeVersion: newVersion,
		},
	).Context(context.Background()).Do()
	if err != nil {
		ctx.SendLogs(
			"GKE running cluster version update request of  "+clusterName+" nodepool "+nodeName+" failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return ApiErrors(err, "Error in running cluster update request of nodepool version")
	}

	return cloud.waitForCluster(clusterName,  ctx)
}

func (cloud *GKE) UpdateNodePoolImageType(clusterName, nodeName, imageType string, ctx utils.Context) types.CustomCPError {
	if imageType == "" {
		return types.CustomCPError{}
	}

	_, err := cloud.Client.Projects.Zones.Clusters.NodePools.Update(
		cloud.ProjectId,
		cloud.Region+"-"+cloud.Zone,
		clusterName,
		nodeName,
		&gke.UpdateNodePoolRequest{
			ImageType: imageType,
		},
	).Context(context.Background()).Do()
	if err != nil {
		ctx.SendLogs(
			"GKE  running cluster image type update request of "+clusterName+" nodepool "+nodeName+" failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return ApiErrors(err, "Error in running cluster nodepole image type update")
	}

	return cloud.waitForCluster(clusterName,  ctx)
}

func (cloud *GKE) UpdateNodePoolCount(clusterName, nodeName string, newCount int64, ctx utils.Context) types.CustomCPError {
	if newCount == 0 {
		return types.CustomCPError{
			StatusCode:  500,
			Error:       "Error in updating node count",
			Description: "Node Count can't be zero.Select a numerical value for node count.",
		}
	}

	_, err := cloud.Client.Projects.Zones.Clusters.NodePools.SetSize(
		cloud.ProjectId,
		cloud.Region+"-"+cloud.Zone,
		clusterName,
		nodeName,
		&gke.SetNodePoolSizeRequest{
			NodeCount: newCount,
		},
	).Context(context.Background()).Do()
	if err != nil {
		ctx.SendLogs(
			"GKE running cluster nodepool count update request of "+clusterName+" nodepool "+nodeName+" failed "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return ApiErrors(err, "Error in updating running cluster nodepool count")
	}

	return cloud.waitForCluster(clusterName, ctx)
}

func (cloud *GKE) UpdateClusterAddson(cluster GKECluster, ctx utils.Context) types.CustomCPError {
//	if cluster.Name == "" || cluster == (GKECluster{}){
//		return types.CustomCPError{}
//	}

	response :=  GenerateClusterCreateRequest(cloud.ProjectId, cloud.Region, cloud.Zone,cluster)

	_, err := cloud.Client.Projects.Zones.Clusters.Update(
		cloud.ProjectId,
		cloud.Region+"-"+cloud.Zone,
		cluster.Name,
		&gke.UpdateClusterRequest{
			Update: &gke.ClusterUpdate{
				DesiredAddonsConfig: response.Cluster.AddonsConfig,
				DesiredMonitoringService: response.Cluster.MonitoringService,
				DesiredLoggingService: response.Cluster.LoggingService,
				DesiredVerticalPodAutoscaling: response.Cluster.VerticalPodAutoscaling,
				DesiredMasterAuthorizedNetworksConfig: response.Cluster.MasterAuthorizedNetworksConfig,
				DesiredLocations:response.Cluster.Zone,
				DesiredResourceUsageExportConfig: response.Cluster.ResourceUsageExportConfig,
			},
		},
	).Context(context.Background()).Do()
	if err != nil {
		ctx.SendLogs(
			"GKE running cluster resource usage export config update request of "+cluster.Name+" failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return types.CustomCPError{
			StatusCode:  512,
			Error:       "Error in running cluster resource usage export config update",
			Description: err.Error(),
		}
	}

	return cloud.waitForCluster(cluster.Name, ctx)
}

func (cloud *GKE) UpdateNodepoolManagement(clusterName string , nodepool NodePool, ctx utils.Context) types.CustomCPError{
	request := GenerateNodePoolManagementFromRequest(cloud.ProjectId,cloud.Region,cloud.Zone,nodepool)

	_,err := cloud.Client.Projects.Zones.Clusters.NodePools.SetManagement(cloud.ProjectId,cloud.Region + "-" +cloud.Zone,clusterName,nodepool.Name,request).Context(context.Background()).Do()
	if err != nil {
		ctx.SendLogs(
			"GKE running cluster nodepool count update request of "+clusterName+" nodepool "+nodepool.Name+" failed "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return ApiErrors(err, "Error in updating running cluster nodepool management")
	}

	return cloud.waitForCluster(clusterName, ctx)
}
func (cloud *GKE) AutoscaleNodepool(projectId string, zone string, clusterName string, nodePoolName string , scaling NodePoolAutoscaling,ctx utils.Context )types.CustomCPError{
	if scaling.MinNodeCount==0 || scaling.MaxNodeCount == 0 {
		return types.CustomCPError{
			StatusCode:512,
			Error : "Error in running cluster autoscaling update",
			Description :"Min/max node count cannot be zero.Select numerical value for node count.",
		}
	}

	_, err := cloud.Client.Projects.Zones.Clusters.NodePools.Autoscaling(
		cloud.ProjectId,
		cloud.Region+"-"+cloud.Zone,
		clusterName,
		nodePoolName,
		&gke.SetNodePoolAutoscalingRequest{
			Autoscaling:     &gke.NodePoolAutoscaling{
				Enabled:         scaling.Enabled,
				MaxNodeCount:    scaling.MaxNodeCount,
				MinNodeCount:    scaling.MinNodeCount,
			},
			Name:            clusterName,
		},
	).Context(context.Background()).Do()
	if err != nil {
		ctx.SendLogs(
			"GKE node update of cluster "+clusterName+" and node "+nodePoolName+" failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return ApiErrors(err, "Error in updating node version")
	}

	return cloud.waitForNodePool(clusterName, nodePoolName, ctx)
}

func (cloud *GKE) AddNodePool(clusterName string, nodepool []*NodePool,ctx utils.Context) types.CustomCPError{

	nodepoolRequest := GenerateNodepoolCreateRequest(cloud.ProjectId, cloud.Region, cloud.Zone,clusterName, nodepool)
	//nodepoolRequest := GenerateNodePoolFromRequest( nodepool)

/*	networkInformation := cloud.getGCPNetwork(token, ctx)
	// overriding network configurations with network from current project
	if len(networkInformation.Definition) > 0 {
		clusterRequest.Cluster.Network = networkInformation.Definition[0].Vpc.Name
		if len(networkInformation.Definition[0].Subnets) > 0 {
			clusterRequest.Cluster.Subnetwork = networkInformation.Definition[0].Subnets[0].Name
		}
	}
*/
	_, err := cloud.Client.Projects.Zones.Clusters.NodePools.Create(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, clusterName,nodepoolRequest).Context(context.Background()).Do()
	if err !=nil{}
	requestJson, _ := json.Marshal(nodepoolRequest)
	ctx.SendLogs(
		"GKE cluster creation of "+clusterName+" submitted: "+string(requestJson),
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)

	if err != nil && !strings.Contains(err.Error(), "Already exists") {
		ctx.SendLogs(
			"GKE nodepool creation of '"+nodepool[1].Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return types.CustomCPError{
			StatusCode:  512,
			Error:       "Error in adding nodepool in running cluster.",
			Description: err.Error(),
		}
	}

	return cloud.waitForCluster(clusterName, ctx)
}

func (cloud *GKE) DeleteCluster(clusterName string, ctx utils.Context) types.CustomCPError {
	_, err := cloud.Client.Projects.Zones.Clusters.Delete(
		cloud.ProjectId,
		cloud.Region+"-"+cloud.Zone,
		clusterName,
	).Context(context.Background()).Do()

	if err != nil && !strings.Contains(err.Error(), "notFound") {
		ctx.SendLogs(
			"GKE cluster deletion for '"+clusterName+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return ApiErrors(err, "Error in cluster deletion")
	} else if err != nil && strings.Contains(err.Error(), "notFound") {
		ctx.SendLogs(
			"GKE cluster '"+clusterName+"' was not found.",
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return ApiErrors(err, "Error in cluster deletion")
	}

	return types.CustomCPError{}
}

func (cloud *GKE) waitForCluster(clusterName string, ctx utils.Context) types.CustomCPError {
	message := ""
	for {
		cluster, err := cloud.Client.Projects.Zones.Clusters.Get(
			cloud.ProjectId,
			cloud.Region+"-"+cloud.Zone,
			clusterName,
		).Context(context.Background()).Do()
		if err != nil {
			ctx.SendLogs(
				"GKE cluster creation/updation for '"+clusterName+"' failed: "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return ApiErrors(err, "Error in GKE cluster creation/Updation")
		}
		if cluster.Status == statusRunning {
			ctx.SendLogs(
				"GKE cluster '"+clusterName+"' is running.",
				models.LOGGING_LEVEL_INFO,
				models.Backend_Logging,
			)
			return types.CustomCPError{}
		}
		if cluster.Status != message {
			ctx.SendLogs(
				"GKE cluster '"+clusterName+"' is creating/updating.",
				models.LOGGING_LEVEL_INFO,
				models.Backend_Logging,
			)
			message = cluster.Status
		}
		time.Sleep(time.Second * 5)
	}
}

func (cloud *GKE) waitForNodePool(clusterName, nodeName string, ctx utils.Context) types.CustomCPError {
	message := ""
	for {
		nodepool, err := cloud.Client.Projects.Zones.Clusters.NodePools.Get(
			cloud.ProjectId,
			cloud.Region+"-"+cloud.Zone,
			clusterName,
			nodeName,
		).Context(context.Background()).Do()
		if err != nil {
			ctx.SendLogs(
				"GKE node creation/updation for cluster '"+clusterName+"' and node '"+nodeName+"' failed: "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return ApiErrors(err, "Error in cluster"+clusterName+"nodepool creation/updation")
		}
		if nodepool.Status == statusRunning {
			ctx.SendLogs(
				"GKE node '"+nodeName+"' for cluster '"+clusterName+"' is running.",
				models.LOGGING_LEVEL_INFO,
				models.Backend_Logging,
			)
			return types.CustomCPError{}
		}
		if nodepool.Status != message {
			ctx.SendLogs(
				"GKE node '"+nodeName+"' for cluster '"+clusterName+"' is creating/updating.",
				models.LOGGING_LEVEL_INFO,
				models.Backend_Logging,
			)
			message = nodepool.Status
		}
		time.Sleep(time.Second * 5)
	}
}

func (cloud *GKE) getGKEVersions(ctx utils.Context) (*gke.ServerConfig, types.CustomCPError) {
	config, err := cloud.Client.Projects.Zones.GetServerconfig("*", cloud.Zone).
		Context(context.Background()).
		Do()

	if err != nil {
		ctx.SendLogs(
			"GKE fetch options for '"+cloud.ProjectId+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return nil, ApiErrors(err, "Error in fetching GKE versions")
	}

	return config, types.CustomCPError{}
}

func (cloud *GKE) getGCPNetwork(token string, ctx utils.Context) (gcpNetwork types.GCPNetwork) {
	url := getNetworkHost(string(models.GCP), cloud.ProjectId)

	network, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs("GKE get network:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return gcpNetwork
	}

	err = json.Unmarshal(network.([]byte), &gcpNetwork)
	if err != nil {
		ctx.SendLogs("GKE get network: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return gcpNetwork
	}

	return gcpNetwork
}

func (cloud *GKE) init() types.CustomCPError {
	if cloud.Client != nil {
		return types.CustomCPError{}
	}

	var err error
	ctx := context.Background()

	cloud.Client, err = gke.NewService(ctx, option.WithCredentialsJSON([]byte(cloud.Credentials)))
	if err != nil {
		return ApiErrors(err, "Error in initializing cloud credentials")
	}

	cloud.Compute, err = compute.NewService(ctx, option.WithCredentialsJSON([]byte(cloud.Credentials)))
	if err != nil {
		beego.Error(err.Error())
		return ApiErrors(err, "Error in initializing cloud credentials")
	}
	return types.CustomCPError{}
}

func (cloud *GKE) fetchClusterStatus(clusterName string, ctx utils.Context) (cluster GKECluster, err types.CustomCPError) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs("GKE get status for '"+cloud.ProjectId+" failed: "+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return cluster, err
		}
	}

	latestCluster, err1 := cloud.Client.Projects.Zones.Clusters.Get(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, clusterName).Do()
	if err1 != nil && !strings.Contains(err1.Error(), "not exist") {
		ctx.SendLogs(
			"GKE get status for '"+cloud.ProjectId+" failed: "+err1.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, ApiErrors(err1, "Cluster is not in running state")
	}
	if latestCluster == nil {
		ctx.SendLogs(
			err1.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, ApiErrors(err1, "Error in fetching cluster status")
	}

	return GenerateClusterFromResponse(*latestCluster), types.CustomCPError{}
}

func (cloud *GKE) fetchNodePool(cluster GKECluster, status *KubeClusterStatus, ctx utils.Context) types.CustomCPError {

	if cloud.Client == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs("GKE get status for '"+cloud.ProjectId+" failed: "+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	for _, pool := range cluster.NodePools {
		npool := pool.InstanceGroupUrls[0]
		arr := strings.Split(npool, "/")
		createdNodes, err := cloud.Compute.InstanceGroupManagers.ListManagedInstances(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, arr[10]).Context(context.Background()).Do()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return ApiErrors(err, "Error in fetching cluster status")
		}
		nodes := []KubeNodesStatus{}
		for _, node := range createdNodes.ManagedInstances {

			splits := strings.Split(node.Instance, "/")
			nodeName := splits[len(splits)-1]
			createdNode, err := cloud.Compute.Instances.Get(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, nodeName).Context(context.Background()).Do()
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return ApiErrors(err, "Error in fetching cluster status")
			}
			node := KubeNodesStatus{}
			node.Id = createdNode.Name
			node.Name = createdNode.Name
			node.State = createdNode.Status
			if len(createdNode.NetworkInterfaces) > 0 {
				node.PrivateIp = createdNode.NetworkInterfaces[0].NetworkIP
				if len(createdNode.NetworkInterfaces[0].AccessConfigs) > 0 {
					node.PublicIp = createdNode.NetworkInterfaces[0].AccessConfigs[0].NatIP
				}
			}
			nodes = append(nodes, node)
		}
		for i, statuspool := range status.WorkerPools {
			if statuspool.Link == pool.InstanceGroupUrls[0] {

				status.WorkerPools[i].Nodes = nodes
			}
		}
	}

	return types.CustomCPError{}
}

func (cloud *GKE) deleteCluster(cluster GKECluster, ctx utils.Context) types.CustomCPError {
	if cloud.Client == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs("GKE terminate cluster for "+cloud.ProjectId+"' failed: "+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	_, err := cloud.Client.Projects.Zones.Clusters.Delete(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, cluster.Name).Do()
	if err != nil {
		ctx.SendLogs(
			"GKE terminate cluster for "+cloud.ProjectId+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return ApiErrors(err, "Error in deleting cluster")
	}

	return types.CustomCPError{}
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

func GetGKE(credentials gcp.GcpCredentials) (GKE, types.CustomCPError) {
	return GKE{
		Credentials: credentials.RawData,
		ProjectId:   credentials.AccountData.ProjectId,
		Region:      credentials.Region,
		Zone:        credentials.Zone,
	}, types.CustomCPError{}
}
