package doks

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"context"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/digitalocean/godo"
	"github.com/ghodss/yaml"
	"golang.org/x/oauth2"
	"log"
	"strconv"
	"strings"
	"time"
)

type DOKS struct {
	AccessKey string
	Region    string
	Client    *godo.Client
	Resources map[string][]string
}

type TokenSource struct {
	AccessToken string
}

type KubernetesClusterCredentials struct {
	Server                   string    `json:"server"`
	CertificateAuthorityData []byte    `json:"certificate_authority_data"`
	ClientCertificateData    []byte    `json:"client_certificate_data"`
	ClientKeyData            []byte    `json:"client_key_data"`
	Token                    string    `json:"token"`
	ExpiresAt                time.Time `json:"expires_at"`
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func (cloud *DOKS) init(ctx utils.Context) types.CustomCPError {
	if cloud.Client != nil {
		return types.CustomCPError{}
	}

	if cloud.AccessKey == "" {
		text := "Invalid cloud credentials"
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(text)
		return types.CustomCPError{StatusCode: 500, Error: "Error in cloud credentials", Description: text}
	}

	tokenSource := &TokenSource{
		AccessToken: cloud.AccessKey,
	}

	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	cloud.Client = godo.NewClient(oauthClient)
	cloud.Resources = make(map[string][]string)
	return types.CustomCPError{}
}

func (cloud *DOKS) createCluster(cluster *KubernetesCluster, ctx utils.Context, token string, credentials vault.DOCredentials) (*KubernetesCluster, types.CustomCPError) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != (types.CustomCPError{}) {
			return cluster, err
		}
	}

	var doNetwork types.DONetwork
	url := getNetworkHost("do", cluster.ProjectId)

	network, err := api_handler.GetAPIStatus(token, url, ctx)
	if err !=nil && strings.Contains(err.Error(),"Not Found"){
		ctx.SendLogs("No Network found for this network", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	}else if err != nil || network == nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Error while fetching network",  credentials,ctx)
		return cluster, cpErr
	}

	err = json.Unmarshal(network.([]byte), &doNetwork)
	if err != nil {
		cpErr := ApiError(err, "Error while fetching network",  credentials,ctx)
		return cluster, cpErr
	}


	ctx.SendLogs(
		"DOKS cluster creation of "+cluster.Name+"' submitted ",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)

	var nodepool []*godo.KubernetesNodePoolCreateRequest

	for _, node := range cluster.NodePools {

		pool := godo.KubernetesNodePoolCreateRequest{

			Name:      node.Name,
			Size:      node.MachineType,
			Count:     node.NodeCount,
			Tags:      node.Tags,
			Labels:    node.Labels,
			AutoScale: node.AutoScale,
			MinNodes:  node.MinNodes,
			MaxNodes:  node.MaxNodes,
		}
		nodepool = append(nodepool, &pool)
	}

	input := godo.KubernetesClusterCreateRequest{
		Name:        cluster.Name,
		RegionSlug:  cluster.Region,
		VersionSlug: cluster.KubeVersion,
		Tags:        cluster.Tags,
		NodePools:   nodepool,
		//MaintenancePolicy: cluster.MaintenancePolicy,
		AutoUpgrade: cluster.AutoUpgrade,
		VPCUUID: doNetwork.Definition[0].VPCs[0].VPCId,
	}


	clus, _, err := cloud.Client.Kubernetes.Create(context.Background(), &input)
	if err != nil {
		utils.SendLog(ctx.Data.Company, "Error in cluster creation : "+err.Error(), models.LOGGING_LEVEL_ERROR, cluster.ProjectId)
		ctx.SendLogs("DOKS cluster creation of '"+cluster.Name+"' failed: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, ApiError(err, "Error in Cluster Creation", credentials, ctx)
	}

	cluster.ID = clus.ID
	for i, pool := range cluster.NodePools{
		pool.ID= clus.NodePools[i].ID
		pool.PoolStatus =true
	}

	status, _, err := cloud.Client.Kubernetes.Get(context.Background(), clus.ID)

	for status.Status.State != "running" {
		time.Sleep(30 * time.Second)
		status, _, err = cloud.Client.Kubernetes.Get(context.Background(), clus.ID)
	}

	time.Sleep(15 * time.Second)

	return cluster, types.CustomCPError{}
}
func (cloud *DOKS) addNodepool(nodepool KubernetesNodePool, ctx utils.Context, clusterId,projectId string, credentials vault.DOCredentials) ( string, types.CustomCPError) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != (types.CustomCPError{}) {
			return  "",err
		}
	}

	input := godo.KubernetesNodePoolCreateRequest{

		Name:      nodepool.Name,
		Size:      nodepool.MachineType,
		Count:     nodepool.NodeCount,
		Tags:      nodepool.Tags,
		Labels:    nodepool.Labels,
		AutoScale: nodepool.AutoScale,
		MinNodes:  nodepool.MinNodes,
		MaxNodes:  nodepool.MaxNodes,
	}



	pool, _, err := cloud.Client.Kubernetes.CreateNodePool(context.Background(),clusterId, &input)
	if err != nil {
		utils.SendLog(ctx.Data.Company, "Error in cluster updating : "+err.Error(), models.LOGGING_LEVEL_ERROR, projectId)
		ctx.SendLogs("DOKS ndepool creation of '"+nodepool.Name+"' failed: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "",ApiError(err, "Error in updating cluster", credentials, ctx)
	}

	_, res, err := cloud.Client.Kubernetes.GetNodePool(context.Background(), clusterId,pool.ID)
	for res.StatusCode != 200 {
		time.Sleep(30 * time.Second)
		_,res, err = cloud.Client.Kubernetes.GetNodePool(context.Background(),clusterId,pool.ID)
	}


	time.Sleep(60 * time.Second)

	return pool.ID,types.CustomCPError{}

}
func (cloud *DOKS) deleteNodepool( ctx utils.Context,nodepoolId , clusterId,projectId string, credentials vault.DOCredentials) ( types.CustomCPError) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != (types.CustomCPError{}) {
			return  err
		}
	}

	_, err := cloud.Client.Kubernetes.DeleteNodePool(context.Background(),clusterId,nodepoolId)
	if err != nil {
		utils.SendLog(ctx.Data.Company, "Error in nodepool deletion : "+err.Error(), models.LOGGING_LEVEL_ERROR, projectId)
		ctx.SendLogs("DOKS nodepool deletion of '"+nodepoolId+"' failed: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiError(err, "Error in deleting nodepool during updating", credentials, ctx)
	}

	_, response, err := cloud.Client.Kubernetes.GetNodePool(context.Background(), clusterId,nodepoolId)

	for response.StatusCode != 200 {
		time.Sleep(30 * time.Second)
		_, response, err = cloud.Client.Kubernetes.GetNodePool(context.Background(), clusterId,nodepoolId)
	}

	time.Sleep(15 * time.Second)

	return types.CustomCPError{}
}

func (cloud *DOKS) deleteCluster(cluster KubernetesCluster, ctx utils.Context) types.CustomCPError {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != (types.CustomCPError{}) {
			return err
		}
	}

	_, err := cloud.Client.Kubernetes.Delete(context.Background(), cluster.ID)
	if err != nil {
		utils.SendLog(ctx.Data.Company, "Error in cluster termination: "+err.Error(), models.LOGGING_LEVEL_INFO, ctx.Data.ProjectId)
		return ApiError(err, "Error in terminating kubernetes cluster", vault.DOCredentials{}, ctx)
	}

	time.Sleep(20 * time.Second)
	
	return types.CustomCPError{}
}


func (cloud *DOKS) UpdateCluster(cluster *KubernetesCluster, ctx utils.Context, credentials vault.DOCredentials) ( types.CustomCPError) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != (types.CustomCPError{}) {
			return err
		}
	}

	input := godo.KubernetesClusterUpdateRequest{
		Name:              cluster.Name,
		AutoUpgrade:       &cluster.AutoUpgrade,
		Tags:              cluster.Tags,
	}

	clus, _, err := cloud.Client.Kubernetes.Update(context.Background(),cluster.ID, &input)
	if err != nil {
		utils.SendLog(ctx.Data.Company, "Error in cluster creation : "+err.Error(), models.LOGGING_LEVEL_ERROR, cluster.ProjectId)
		ctx.SendLogs("DOKS cluster creation of '"+cluster.Name+"' failed: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return  ApiError(err, "Error in updating cluster autoupgrade", credentials, ctx)
	}

	status, _, err := cloud.Client.Kubernetes.Get(context.Background(), clus.ID)
	for status.Status.State != "running" {
		time.Sleep(30 * time.Second)
		status, _, err = cloud.Client.Kubernetes.Get(context.Background(), clus.ID)
	}
	return  types.CustomCPError{}
}
func (cloud *DOKS) UpdateNodePool(nodepool *KubernetesNodePool, ctx utils.Context, clusterId,projectId string, credentials vault.DOCredentials) ( types.CustomCPError) {

	input := godo.KubernetesNodePoolUpdateRequest{
		Name:      nodepool.Name,
		Count:     &nodepool.NodeCount,
		AutoScale: &nodepool.AutoScale,
		MinNodes:  &nodepool.MinNodes,
		MaxNodes:  &nodepool.MaxNodes,
		Tags:		nodepool.Tags,
	}

	_, _, err := cloud.Client.Kubernetes.UpdateNodePool(context.Background(),clusterId,nodepool.ID,&input)
	if err != nil {
		utils.SendLog(ctx.Data.Company, "Error in updating nodepool : "+err.Error(), models.LOGGING_LEVEL_ERROR, projectId)
		ctx.SendLogs("DOKS nodepool updating of '"+nodepool.Name+"' failed: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return  ApiError(err, "Error in updating nodepool "+nodepool.ID, credentials, ctx)

	}

	_,res, err := cloud.Client.Kubernetes.GetNodePool(context.Background(), clusterId,nodepool.ID)
	for res.StatusCode != 200 {
		time.Sleep(30 * time.Second)
		_, res, err = cloud.Client.Kubernetes.GetNodePool(context.Background(), clusterId,nodepool.ID)
	}

	return  types.CustomCPError{}

}
func (cloud *DOKS) UpgradeKubernetesVersion( cluster *KubernetesCluster, ctx utils.Context, credentials vault.DOCredentials ) ( types.CustomCPError) {

	input := godo.KubernetesClusterUpgradeRequest{
		VersionSlug : cluster.KubeVersion,
	}

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != (types.CustomCPError{}) {
			return  err
		}
	}

	_, err := cloud.Client.Kubernetes.Upgrade(context.Background(),cluster.ID, &input)
	if err != nil {
		utils.SendLog(ctx.Data.Company, "Error in upgrading cluster version  : "+err.Error(), models.LOGGING_LEVEL_ERROR, cluster.ProjectId)
		ctx.SendLogs("DOKS cluster kubernetes version of '"+cluster.Name+"' failed: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return  ApiError(err, "Error in cluster version upgrade ", credentials, ctx)
	}

	status, _, err := cloud.Client.Kubernetes.Get(context.Background(), cluster.ID)
	for status.Status.State != "running" {
		time.Sleep(30 * time.Second)
		status, _, err = cloud.Client.Kubernetes.Get(context.Background(), cluster.ID)
	}

	return  types.CustomCPError{}

}

func (cloud *DOKS) fetchStatus(ctx utils.Context, cluster KubernetesCluster) (KubeClusterStatus, types.CustomCPError) {
	var response KubeClusterStatus
	done :=false
	count := 0
	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err.Description != "" {
			return KubeClusterStatus{}, err
		}
	}

	status, _, err := cloud.Client.Kubernetes.Get(context.Background(), cluster.ID)
	if err != nil {
		ctx.SendLogs("DOKS get cluster status  failed: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubeClusterStatus{}, ApiError(err, "Error in fetching kubernetes cluster status", vault.DOCredentials{}, ctx)
	}

	response.Name = status.Name
	response.ID = status.ID
	response.RegionSlug = status.RegionSlug
	response.KubernetesVersion = status.VersionSlug
	response.ClusterIp = status.IPv4
	response.Endpoint = status.Endpoint
	response.State = string(status.Status.State)
	for _,p := range cluster.NodePools {

		for _, pool := range status.NodePools {
			for _, worker := range response.WorkerPools {
				done=false
				if pool.ID == worker.ID {
					done=true
				}
			}
			if p.ID == pool.ID && !done{
				count++
				var workerPool KubeWorkerPoolStatus
				workerPool.Name = pool.Name
				workerPool.ID = pool.ID
				workerPool.Size = pool.Size
				workerPool.Count = pool.Count
				if pool.AutoScale == true {
					workerPool.AutoScaling.AutoScale = pool.AutoScale
					workerPool.AutoScaling.MinCount = pool.MinNodes
					workerPool.AutoScaling.MaxCount = pool.MaxNodes
				}

				for _, nodes := range pool.Nodes {

					var poolNodes PoolNodes
					poolNodes.Name = nodes.Name
					poolNodes.DropletID = nodes.DropletID
					dropletId, _ := strconv.ParseInt(nodes.DropletID, 10, 64)
					droplet, _, err := cloud.Client.Droplets.Get(context.Background(), int(dropletId))
					if err != nil {
						ctx.SendLogs("Error in getting droplet status "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
						cpErr := ApiError(err, "Error in getting droplet status", vault.DOCredentials{}, ctx)
						return KubeClusterStatus{}, cpErr
					}
					poolNodes.Name = nodes.Name
					poolNodes.PrivateIp, _ = droplet.PrivateIPv4()
					poolNodes.PublicIp, _ = droplet.PublicIPv4()
					poolNodes.State = nodes.Status.State

					workerPool.Nodes = append(workerPool.Nodes, poolNodes)
				}
				response.WorkerPools = append(response.WorkerPools, workerPool)
				done =false
			}
		}
	}
	response.NodePoolCount = count
	return response, types.CustomCPError{}
}

func (cloud *DOKS) GetServerConfig(ctx utils.Context) (*godo.KubernetesOptions, types.CustomCPError) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != (types.CustomCPError{}) {
			return &godo.KubernetesOptions{}, err
		}
	}

	options, _, err := cloud.Client.Kubernetes.GetOptions(context.Background())
	if err != nil {
		return &godo.KubernetesOptions{}, ApiError(err, "Error in fetching valid version and machine sizes", vault.DOCredentials{}, ctx)
	}

	return options, types.CustomCPError{}
}
func (cloud *DOKS) GetKubeConfig(ctx utils.Context, cluster KubernetesCluster) (KubernetesConfig, types.CustomCPError) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != (types.CustomCPError{}) {
			return KubernetesConfig{}, err
		}
	}

	config, _, err := cloud.Client.Kubernetes.GetKubeConfig(context.Background(), cluster.ID)
	if err != nil {
		ctx.SendLogs(
			"DOKS terminate cluster for "+cluster.ProjectId+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return KubernetesConfig{}, ApiError(err, "Error in getting kubernetes config file", vault.DOCredentials{}, ctx)
	}

	var con KubernetesClusterConfig
	con.KubeconfigYAML = config.KubeconfigYAML
	kubeFile := KubernetesConfig{}

	err = yaml.Unmarshal([]byte(config.KubeconfigYAML), &kubeFile)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return kubeFile, types.CustomCPError{}
}
