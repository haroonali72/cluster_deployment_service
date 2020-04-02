package eks

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/aws/IAMRoles"
	autoscaling2 "antelope/models/aws/autoscaling"
	"antelope/models/gcp"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"context"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	eks "google.golang.org/api/container/v1"
	"google.golang.org/api/option"
	"strings"
	"time"
)

const (
	statusRunning = "RUNNING"
)

type EKS struct {
	Client    *eks.Cluster
	AccessKey string
	SecretKey string
	Region    string
	Resources map[string]interface{}

	Scaler autoscaling2.AWSAutoScaler
	Roles  IAMRoles.AWSIAMRoles
}

func (cloud *EKS) ListClusters(ctx utils.Context) ([]EKSCluster, error) {
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
			"EKS list clusters for '"+cloud.ProjectId+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return nil, err
	}

	result := []EKSCluster{}
	for _, v := range list.Clusters {
		if v != nil {
			result = append(result, GenerateClusterFromResponse(*v))
		}
	}

	return result, nil
}

func (cloud *EKS) CreateCluster(eksCluster EKSCluster, token string, ctx utils.Context) error {
	err := Validate(eksCluster)
	if err != nil {
		ctx.SendLogs(
			"EKS cluster validation for '"+eksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	clusterRequest := GenerateClusterCreateRequest(cloud.ProjectId, cloud.Region, cloud.Zone, eksCluster)
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

	requestJson, _ := json.Marshal(clusterRequest)
	ctx.SendLogs(
		"EKS cluster creation request for '"+eksCluster.Name+"' submitted: "+string(requestJson),
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)

	if err != nil && !strings.Contains(err.Error(), "alreadyExists") {
		ctx.SendLogs(
			"EKS cluster creation request for '"+eksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	} else if err != nil && strings.Contains(err.Error(), "alreadyExists") {
		ctx.SendLogs(
			"EKS cluster '"+eksCluster.Name+"' already exists.",
			models.LOGGING_LEVEL_INFO,
			models.Backend_Logging,
		)
		return nil
	}

	return cloud.waitForCluster(eksCluster.Name, ctx)
}

func (cloud *EKS) UpdateMasterVersion(clusterName, newVersion string, ctx utils.Context) error {
	if newVersion == "" {
		return nil
	}

	_, err := cloud.Client.Projects.Zones.Clusters.Update(
		cloud.ProjectId,
		cloud.Region+"-"+cloud.Zone,
		clusterName,
		&eks.UpdateClusterRequest{
			Update: &eks.ClusterUpdate{
				DesiredMasterVersion: newVersion,
			},
		},
	).Context(context.Background()).Do()
	if err != nil {
		ctx.SendLogs(
			"EKS cluster update request for '"+clusterName+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return cloud.waitForCluster(clusterName, ctx)
}

func (cloud *EKS) UpdateNodeVersion(clusterName, nodeName, newVersion string, ctx utils.Context) error {
	if newVersion == "" {
		return nil
	}

	_, err := cloud.Client.Projects.Zones.Clusters.NodePools.Update(
		cloud.ProjectId,
		cloud.Region+"-"+cloud.Zone,
		clusterName,
		nodeName,
		&eks.UpdateNodePoolRequest{
			NodeVersion: newVersion,
		},
	).Context(context.Background()).Do()
	if err != nil {
		ctx.SendLogs(
			"EKS node update request for cluster '"+clusterName+"' and node '"+nodeName+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return cloud.waitForNodePool(clusterName, nodeName, ctx)
}

func (cloud *EKS) UpdateNodeCount(clusterName, nodeName string, newCount int64, ctx utils.Context) error {
	if newCount == 0 {
		return nil
	}

	_, err := cloud.Client.Projects.Zones.Clusters.NodePools.SetSize(
		cloud.ProjectId,
		cloud.Region+"-"+cloud.Zone,
		clusterName,
		nodeName,
		&eks.SetNodePoolSizeRequest{
			NodeCount: newCount,
		},
	).Context(context.Background()).Do()
	if err != nil {
		ctx.SendLogs(
			"EKS node update request for cluster '"+clusterName+"' and node '"+nodeName+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return cloud.waitForNodePool(clusterName, nodeName, ctx)
}

func (cloud *EKS) DeleteCluster(clusterName string, ctx utils.Context) error {
	_, err := cloud.Client.Projects.Zones.Clusters.Delete(
		cloud.ProjectId,
		cloud.Region+"-"+cloud.Zone,
		clusterName,
	).Context(context.Background()).Do()

	if err != nil && !strings.Contains(err.Error(), "notFound") {
		ctx.SendLogs(
			"EKS cluster deletion for '"+clusterName+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	} else if err != nil && strings.Contains(err.Error(), "notFound") {
		ctx.SendLogs(
			"EKS cluster '"+clusterName+"' was not found.",
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func (cloud *EKS) waitForCluster(clusterName string, ctx utils.Context) error {
	message := ""
	for {
		cluster, err := cloud.Client.Projects.Zones.Clusters.Get(
			cloud.ProjectId,
			cloud.Region+"-"+cloud.Zone,
			clusterName,
		).Context(context.Background()).Do()
		if err != nil {
			ctx.SendLogs(
				"EKS cluster creation/updation for '"+clusterName+"' failed: "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
		}
		if cluster.Status == statusRunning {
			ctx.SendLogs(
				"EKS cluster '"+clusterName+"' is running.",
				models.LOGGING_LEVEL_INFO,
				models.Backend_Logging,
			)
			return nil
		}
		if cluster.Status != message {
			ctx.SendLogs(
				"EKS cluster '"+clusterName+"' is creating/updating.",
				models.LOGGING_LEVEL_INFO,
				models.Backend_Logging,
			)
			message = cluster.Status
		}
		time.Sleep(time.Second * 5)
	}
}

func (cloud *EKS) waitForNodePool(clusterName, nodeName string, ctx utils.Context) error {
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
				"EKS node creation/updation for cluster '"+clusterName+"' and node '"+nodeName+"' failed: "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
		}
		if nodepool.Status == statusRunning {
			ctx.SendLogs(
				"EKS node '"+nodeName+"' for cluster '"+clusterName+"' is running.",
				models.LOGGING_LEVEL_INFO,
				models.Backend_Logging,
			)
			return nil
		}
		if nodepool.Status != message {
			ctx.SendLogs(
				"EKS node '"+nodeName+"' for cluster '"+clusterName+"' is creating/updating.",
				models.LOGGING_LEVEL_INFO,
				models.Backend_Logging,
			)
			message = nodepool.Status
		}
		time.Sleep(time.Second * 5)
	}
}

func (cloud *EKS) getEKSVersions(ctx utils.Context) (*eks.ServerConfig, error) {
	config, err := cloud.Client.Projects.Zones.GetServerconfig("*", cloud.Zone).
		Context(context.Background()).
		Do()

	if err != nil {
		ctx.SendLogs(
			"EKS server config for '"+cloud.ProjectId+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return nil, err
	}

	return config, nil
}

func (cloud *EKS) getGCPNetwork(token string, ctx utils.Context) (gcpNetwork types.GCPNetwork) {
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

func (cloud *EKS) init() error {
	if cloud.Client != nil {
		return nil
	}

	var err error
	ctx := context.Background()

	cloud.Client, err = eks.NewService(ctx, option.WithCredentialsJSON([]byte(cloud.Credentials)))
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}

func (cloud *EKS) fetchClusterStatus(cluster *EKSCluster, ctx utils.Context) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	latestCluster, err := cloud.Client.Projects.Zones.Clusters.Get(cloud.ProjectId, cloud.Region+"-"+cloud.Zone, cluster.Name).Do()
	if err != nil && !strings.Contains(err.Error(), "not exist") {
		ctx.SendLogs(
			"EKS get cluster for '"+cloud.ProjectId+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	if latestCluster != nil {
		cluster.NodePools = GenerateNodePoolFromResponse(latestCluster.NodePools)
	}

	return nil
}

func (cloud *EKS) deleteCluster(cluster EKSCluster, ctx utils.Context) error {
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
			"EKS delete cluster for '"+cloud.ProjectId+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func Validate(eksCluster EKSCluster) error {
	if eksCluster.ProjectId == "" {
		return errors.New("project id is required")
	} else if eksCluster.Name == "" {
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

func GetEKS(credentials vault.AwsCredentials) (EKS, error) {
	return EKS{
		Credentials: credentials.RawData,
		ProjectId:   credentials.AccountData.ProjectId,
		Region:      credentials.Region,
		Zone:        credentials.Zone,
	}, nil
}
