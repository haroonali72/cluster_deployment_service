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

	list, err := cloud.Client.Projects.Zones.Clusters.List(cloud.ProjectId, cloud.Region+"-"+cloud.Zone).Do()
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
			result = append(result, GenerateClusterFromResponse(*v))
		}
	}

	return result, nil
}

func (cloud *GKE) CreateCluster(gkeCluster GKECluster, token string, ctx utils.Context) error {
	err := Validate(gkeCluster)
	if err != nil {
		ctx.SendLogs(
			"GKE cluster validation for '"+gkeCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	clusterRequest := GenerateClusterCreateRequest(cloud.ProjectId, cloud.Region, cloud.Zone, gkeCluster)
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
		cloud.Region+"-"+cloud.Zone,
		clusterRequest,
	).Context(context.Background()).Do()
	gcp.ApiErrors(err)
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
		cloud.Region+"-"+cloud.Zone,
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
		cloud.Region+"-"+cloud.Zone,
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
		cloud.Region+"-"+cloud.Zone,
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
			cloud.Region+"-"+cloud.Zone,
			clusterName,
		).Context(context.Background()).Do()
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

func (cloud *GKE) getGKEVersions(ctx utils.Context) (*gke.ServerConfig, error) {
	config, err := cloud.Client.Projects.Zones.GetServerconfig("*", cloud.Zone).
		Context(context.Background()).
		Do()

	if err != nil {
		ctx.SendLogs(
			"GKE server config for '"+cloud.ProjectId+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return nil, err
	}

	return config, nil
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

func (cloud *GKE) fetchClusterStatus(clusterName string, ctx utils.Context) (cluster GKECluster, err error) {
	if cloud.Client == nil {
		err = cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return cluster, err
		}
	}

	latestCluster, err := cloud.Client.Projects.Zones.Clusters.Get(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, clusterName).Do()
	if err != nil && !strings.Contains(err.Error(), "not exist") {
		ctx.SendLogs(
			"GKE get cluster for '"+cloud.ProjectId+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err
	}

	if latestCluster == nil {
		return cluster, err
	}

	return GenerateClusterFromResponse(*latestCluster), err
}

func (cloud *GKE) deleteCluster(cluster GKECluster, ctx utils.Context) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	_, err := cloud.Client.Projects.Zones.Clusters.Delete(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, cluster.Name).Do()
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

func Validate(gkeCluster GKECluster) error {
	if gkeCluster.ProjectId == "" {
		return errors.New("project id is required")
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
