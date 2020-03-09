package gke

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/gcp"
	"antelope/models/types"
	"antelope/models/utils"
	"context"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
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
	Credentials string
	ProjectId   string
	Region      string
	Zone        string
}

func (cloud *GKE) ListClusters(ctx utils.Context) ([]GKECluster, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return nil, err
		}
	}

	list, err := cloud.Client.Projects.Zones.Clusters.List(cloud.ProjectId, cloud.Zone).Do()
	if err != nil {
		ctx.SendLogs(
			"GKE list clusters for '"+cloud.ProjectId+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return nil, err
	}

	result := []GKECluster{}
	for _, v := range list.Clusters {
		if v != nil {
			result = append(result, cloud.generateClusterFromResponse(*v))
		}
	}

	return result, nil
}

func (cloud *GKE) CreateCluster(gkeCluster GKECluster, token string, ctx utils.Context) error {
	err := validate(gkeCluster)
	if err != nil {
		ctx.SendLogs(
			"GKE cluster validation for '"+gkeCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	clusterRequest := cloud.generateClusterCreateRequest(gkeCluster)
	networkInformation := cloud.getGCPNetwork(token, ctx)

	// overriding network configurations with network from current project
	if len(networkInformation.Definition) > 0 {
		clusterRequest.Cluster.Network = networkInformation.Definition[0].Vpc.Name
		if len(networkInformation.Definition[0].Subnets) > 0 {
			clusterRequest.Cluster.Subnetwork = networkInformation.Definition[0].Subnets[0].Name
		}
	}

	_, err = cloud.Client.Projects.Zones.Clusters.Create(
		cloud.ProjectId,
		cloud.Zone,
		clusterRequest,
	).Context(context.Background()).Do()

	requestJson, _ := json.Marshal(clusterRequest)
	ctx.SendLogs(
		"GKE cluster creation request for '"+gkeCluster.Name+"' submitted: "+string(requestJson),
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)

	if err != nil && !strings.Contains(err.Error(), "alreadyExists") {
		ctx.SendLogs(
			"GKE cluster creation request for '"+gkeCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	} else if err != nil && strings.Contains(err.Error(), "alreadyExists") {
		ctx.SendLogs(
			"GKE cluster '"+gkeCluster.Name+"' already exists.",
			models.LOGGING_LEVEL_INFO,
			models.Backend_Logging,
		)
		return nil
	}

	return cloud.waitForCluster(gkeCluster.Name, ctx)
}

func (cloud *GKE) UpdateMasterVersion(clusterName, newVersion string, ctx utils.Context) error {
	if newVersion == "" {
		return nil
	}

	_, err := cloud.Client.Projects.Zones.Clusters.Update(
		cloud.ProjectId,
		cloud.Zone,
		clusterName,
		&gke.UpdateClusterRequest{
			Update: &gke.ClusterUpdate{
				DesiredMasterVersion: newVersion,
			},
		},
	).Context(context.Background()).Do()
	if err != nil {
		ctx.SendLogs(
			"GKE cluster update request for '"+clusterName+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return cloud.waitForCluster(clusterName, ctx)
}

func (cloud *GKE) UpdateNodeVersion(clusterName, nodeName, newVersion string, ctx utils.Context) error {
	if newVersion == "" {
		return nil
	}

	_, err := cloud.Client.Projects.Zones.Clusters.NodePools.Update(
		cloud.ProjectId,
		cloud.Zone,
		clusterName,
		nodeName,
		&gke.UpdateNodePoolRequest{
			NodeVersion: newVersion,
		},
	).Context(context.Background()).Do()
	if err != nil {
		ctx.SendLogs(
			"GKE node update request for cluster '"+clusterName+"' and node '"+nodeName+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return cloud.waitForNodePool(clusterName, nodeName, ctx)
}

func (cloud *GKE) UpdateNodeCount(clusterName, nodeName string, newCount int64, ctx utils.Context) error {
	if newCount == 0 {
		return nil
	}

	_, err := cloud.Client.Projects.Zones.Clusters.NodePools.SetSize(
		cloud.ProjectId,
		cloud.Zone,
		clusterName,
		nodeName,
		&gke.SetNodePoolSizeRequest{
			NodeCount: newCount,
		},
	).Context(context.Background()).Do()
	if err != nil {
		ctx.SendLogs(
			"GKE node update request for cluster '"+clusterName+"' and node '"+nodeName+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return cloud.waitForNodePool(clusterName, nodeName, ctx)
}

func (cloud *GKE) DeleteCluster(clusterName string, ctx utils.Context) error {
	_, err := cloud.Client.Projects.Zones.Clusters.Delete(
		cloud.ProjectId,
		cloud.Zone,
		clusterName,
	).Context(context.Background()).Do()

	if err != nil && !strings.Contains(err.Error(), "notFound") {
		ctx.SendLogs(
			"GKE cluster deletion for '"+clusterName+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	} else if err != nil && strings.Contains(err.Error(), "notFound") {
		ctx.SendLogs(
			"GKE cluster '"+clusterName+"' was not found.",
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func (cloud *GKE) waitForCluster(clusterName string, ctx utils.Context) error {
	message := ""
	for {
		cluster, err := cloud.Client.Projects.Zones.Clusters.Get(
			cloud.ProjectId,
			cloud.Zone,
			clusterName,
		).Context(context.TODO()).Do()
		if err != nil {
			ctx.SendLogs(
				"GKE cluster creation/updation for '"+clusterName+"' failed: "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
		}
		if cluster.Status == statusRunning {
			ctx.SendLogs(
				"GKE cluster '"+clusterName+"' is running.",
				models.LOGGING_LEVEL_INFO,
				models.Backend_Logging,
			)
			return nil
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

func (cloud *GKE) waitForNodePool(clusterName, nodeName string, ctx utils.Context) error {
	message := ""
	for {
		nodepool, err := cloud.Client.Projects.Zones.Clusters.NodePools.Get(
			cloud.ProjectId,
			cloud.Zone,
			clusterName,
			nodeName,
		).Context(context.TODO()).Do()
		if err != nil {
			ctx.SendLogs(
				"GKE node creation/updation for cluster '"+clusterName+"' and node '"+nodeName+"' failed: "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
		}
		if nodepool.Status == statusRunning {
			ctx.SendLogs(
				"GKE node '"+nodeName+"' for cluster '"+clusterName+"' is running.",
				models.LOGGING_LEVEL_INFO,
				models.Backend_Logging,
			)
			return nil
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

func (cloud *GKE) getGCPNetwork(token string, ctx utils.Context) (gcpNetwork types.GCPNetwork) {
	url := getNetworkHost(string(models.GCP), cloud.ProjectId)

	network, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return gcpNetwork
	}

	err = json.Unmarshal(network.([]byte), &gcpNetwork)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return gcpNetwork
	}

	return gcpNetwork
}

func (cloud *GKE) generateClusterFromResponse(v gke.Cluster) GKECluster {
	return GKECluster{
		ProjectId:                      cloud.ProjectId,
		Cloud:                          models.GKE,
		AddonsConfig:                   v.AddonsConfig,
		ClusterIpv4Cidr:                v.ClusterIpv4Cidr,
		Conditions:                     v.Conditions,
		CreateTime:                     v.CreateTime,
		CurrentMasterVersion:           v.CurrentMasterVersion,
		DefaultMaxPodsConstraint:       v.DefaultMaxPodsConstraint,
		Description:                    v.Description,
		EnableKubernetesAlpha:          v.EnableKubernetesAlpha,
		EnableTpu:                      v.EnableTpu,
		Endpoint:                       v.Endpoint,
		ExpireTime:                     v.ExpireTime,
		InitialClusterVersion:          v.InitialClusterVersion,
		IpAllocationPolicy:             v.IpAllocationPolicy,
		LabelFingerprint:               v.LabelFingerprint,
		LegacyAbac:                     v.LegacyAbac,
		Location:                       v.Location,
		Locations:                      v.Locations,
		LoggingService:                 v.LoggingService,
		MaintenancePolicy:              v.MaintenancePolicy,
		MasterAuth:                     v.MasterAuth,
		MasterAuthorizedNetworksConfig: v.MasterAuthorizedNetworksConfig,
		MonitoringService:              v.MonitoringService,
		Name:                           v.Name,
		Network:                        v.Network,
		NetworkConfig:                  v.NetworkConfig,
		NetworkPolicy:                  v.NetworkPolicy,
		NodeIpv4CidrSize:               v.NodeIpv4CidrSize,
		NodePools:                      v.NodePools,
		PrivateClusterConfig:           v.PrivateClusterConfig,
		ResourceLabels:                 v.ResourceLabels,
		ResourceUsageExportConfig:      v.ResourceUsageExportConfig,
		SelfLink:                       v.SelfLink,
		ServicesIpv4Cidr:               v.ServicesIpv4Cidr,
		Status:                         v.Status,
		StatusMessage:                  v.StatusMessage,
		Subnetwork:                     v.Subnetwork,
		TpuIpv4CidrBlock:               v.TpuIpv4CidrBlock,
		Zone:                           v.Zone,
	}
}

func (cloud *GKE) generateClusterCreateRequest(c GKECluster) *gke.CreateClusterRequest {
	request := gke.CreateClusterRequest{
		Cluster: &gke.Cluster{
			AddonsConfig:                   c.AddonsConfig,
			ClusterIpv4Cidr:                c.ClusterIpv4Cidr,
			DefaultMaxPodsConstraint:       c.DefaultMaxPodsConstraint,
			Description:                    c.Description,
			EnableKubernetesAlpha:          c.EnableKubernetesAlpha,
			EnableTpu:                      c.EnableTpu,
			InitialClusterVersion:          c.InitialClusterVersion,
			IpAllocationPolicy:             c.IpAllocationPolicy,
			LabelFingerprint:               c.LabelFingerprint,
			LegacyAbac:                     c.LegacyAbac,
			Locations:                      c.Locations,
			LoggingService:                 c.LoggingService,
			MaintenancePolicy:              c.MaintenancePolicy,
			MonitoringService:              c.MonitoringService,
			MasterAuthorizedNetworksConfig: c.MasterAuthorizedNetworksConfig,
			MasterAuth:                     c.MasterAuth,
			Name:                           c.Name,
			Network:                        c.Network,
			NetworkConfig:                  c.NetworkConfig,
			NetworkPolicy:                  c.NetworkPolicy,
			NodePools:                      c.NodePools,
			PrivateClusterConfig:           c.PrivateClusterConfig,
			ResourceLabels:                 c.ResourceLabels,
			ResourceUsageExportConfig:      c.ResourceUsageExportConfig,
			Subnetwork:                     c.Subnetwork,
		},
		ProjectId: cloud.ProjectId,
		Zone:      cloud.Zone,
	}
	return &request
}

func (cloud *GKE) init() error {
	if cloud.Client != nil {
		return nil
	}

	var err error
	ctx := context.Background()

	cloud.Client, err = gke.NewService(ctx, option.WithCredentialsJSON([]byte(cloud.Credentials)))
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}

func (cloud *GKE) fetchClusterStatus(cluster *GKECluster, ctx utils.Context) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	latestCluster, err := cloud.Client.Projects.Zones.Clusters.Get(cloud.ProjectId, cloud.Zone, cluster.Name).Do()
	if err != nil && !strings.Contains(err.Error(), "not exist") {
		ctx.SendLogs(
			"GKE get cluster for '"+cloud.ProjectId+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	if latestCluster != nil {
		cluster.NodePools = latestCluster.NodePools
	}

	return nil
}

func (cloud *GKE) deleteCluster(cluster GKECluster, ctx utils.Context) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	_, err := cloud.Client.Projects.Zones.Clusters.Delete(cloud.ProjectId, cloud.Zone, cluster.Name).Do()
	if err != nil {
		ctx.SendLogs(
			"GKE delete cluster for '"+cloud.ProjectId+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func validate(gkeCluster GKECluster) error {
	if gkeCluster.ProjectId == "" {
		return errors.New("project id is required")
	} else if gkeCluster.Zone == "" {
		return errors.New("zone is required")
	} else if gkeCluster.Name == "" {
		return errors.New("cluster name is required")
	}
	return nil
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

func GetGKE(credentials gcp.GcpCredentials) (GKE, error) {
	return GKE{
		Credentials: credentials.RawData,
		ProjectId:   credentials.AccountData.ProjectId,
		Region:      credentials.Region,
		Zone:        credentials.Zone,
	}, nil
}
