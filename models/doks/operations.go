package doks

import (
	"antelope/models"
	"antelope/models/api_handler"
	"fmt"
	"strings"

	"antelope/models/types"
	"antelope/models/utils"
	"context"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
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
func getWoodpecker() string {
	return beego.AppConfig.String("woodpecker_url") + models.WoodpeckerEnpoint
}
func (cloud *DOKS) init(ctx utils.Context) error {
	if cloud.Client != nil {
		return nil
	}

	if cloud.AccessKey == "" {
		text := "invalid cloud credentials"
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(text)
		return errors.New(text)
	}

	tokenSource := &TokenSource{
		AccessToken: cloud.AccessKey,
	}

	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	cloud.Client = godo.NewClient(oauthClient)
	cloud.Resources = make(map[string][]string)
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

func (cloud *DOKS) createCluster(cluster KubernetesCluster, ctx utils.Context, companyId string, token string) (KubernetesCluster, error) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != nil {
			return cluster, err
		}
	}

	var doksNetwork types.DONetwork
	url := getNetworkHost("do", cluster.ProjectId)

	network, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil || network == nil {
		return cluster, errors.New("error in fetching network")
	}
	err = json.Unmarshal(network.([]byte), &doksNetwork)
	if err != nil {
		beego.Error(err.Error())
		return cluster, err
	}

	utils.SendLog(companyId, "Creating DOKS Cluster With ID : "+cluster.ProjectId, "info", cluster.ProjectId)

	input :=godo.KubernetesClusterCreateRequest{
		Name:              cluster.Name,
		RegionSlug:        cluster.Region,
		VersionSlug:       cluster.Version,
		Tags:              cluster.Tags,
		VPCUUID:           cluster.VPCUUID,
		//NodePools:         cluster.NodePools,
		//MaintenancePolicy: cluster.MaintenancePolicy,
		AutoUpgrade:       cluster.AutoUpgrade,
	}

	clus,resp,err :=cloud.Client.Kubernetes.Create(context.Background(),&input)
	if err == nil{
		utils.SendLog(companyId, "Error in cluster creation: "+err.Error(), "info", cluster.ProjectId)
		return cluster, err
	}
	cluster.ID=clus.ID
	fmt.Println(resp)

	//cloud.Resources["project"] = append(cloud.Resources["project"], cluster.DOProjectId)
	utils.SendLog(companyId, "DOKS cluster created Successfully : "+cluster.ProjectId, "info", cluster.ProjectId)

	utils.SendLog(companyId, "Creating Node Pools : "+cluster.Name, "info", cluster.ProjectId)
	for index, nodepool := range cluster.NodePools {
		nodepool, err := cloud.createNodePool(nodepool, ctx,cluster.ProjectId,companyId, cluster.ID, token )
		if err != nil {
			utils.SendLog(companyId, "Error in instances creation: "+err.Error(), "info", cluster.ProjectId)
			return cluster, err
		}
		cluster.NodePools[index].Nodes = nodepool.Nodes
	}
		utils.SendLog(companyId, "Node Pool Created Successfully : "+cluster.Name, "info", cluster.ProjectId)

	return cluster, nil
}
func (cloud *DOKS) createNodePool(nodepool *KubernetesNodePool, ctx utils.Context,projectId,companyId, clusterId, token string) (KubernetesNodePool, error) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != nil {
			return *nodepool, err
		}
	}
	var doksNetwork types.DONetwork
	url := getNetworkHost("do", projectId)
	network, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil || network == nil {
		return *nodepool, errors.New("error in fetching network")
	}
	err = json.Unmarshal(network.([]byte), &doksNetwork)
	if err != nil {
		beego.Error(err.Error())
		return *nodepool, err
	}

	utils.SendLog(companyId, "Creating DOKS nodepool With ID : "+projectId, "info", projectId)

	input:= godo.KubernetesNodePoolCreateRequest{
			Name:       nodepool.Name,
			Size:      nodepool.Size,
			Count:     nodepool.Count,
			Tags:      nodepool.Tags,
			Labels:    nodepool.Labels,
			AutoScale: nodepool.AutoScale,
			MinNodes:  nodepool.MinNodes,
			MaxNodes:  nodepool.MaxNodes,
	}

	_,resp,err :=cloud.Client.Kubernetes.CreateNodePool(context.Background(),clusterId,&input)
	if err == nil{
		utils.SendLog(companyId, "Error in cluster creation: "+err.Error(), "info", projectId)
		return *nodepool, err
	}

	fmt.Println(resp)

	return *nodepool, nil
}
func (cloud *DOKS) deleteCluster(cluster KubernetesCluster, ctx utils.Context,projectId,companyId string) (KubernetesNodePool, error) {
	return KubernetesNodePool{}, nil
}
func (cloud *DOKS) deleteNodepool(nodepool *KubernetesNodePool, ctx utils.Context,projectId,companyId, clusterId, token string) (KubernetesNodePool, error) {
	return KubernetesNodePool{}, nil
}
func (cloud *DOKS) deleteNode(nodepool *KubernetesNodePool, ctx utils.Context,projectId,companyId, clusterId, token string) (KubernetesNodePool, error) {
	return KubernetesNodePool{}, nil
}
func (cloud *DOKS) GetCluster(nodepool *KubernetesNodePool, ctx utils.Context,projectId,companyId, clusterId, token string) (KubernetesNodePool, error) {
	return KubernetesNodePool{}, nil
}
func (cloud *DOKS) GetNodePool(nodepool *KubernetesNodePool, ctx utils.Context,projectId,companyId, clusterId, token string) (KubernetesNodePool, error) {
	return KubernetesNodePool{}, nil
}
func (cloud *DOKS) GetKubeConfig(nodepool *KubernetesNodePool, ctx utils.Context,projectId,companyId, clusterId, token string) (KubernetesNodePool, error) {
	return KubernetesNodePool{}, nil
}
func (cloud *DOKS) ListCluster(nodepool *KubernetesNodePool, ctx utils.Context,projectId,companyId, clusterId, token string) (KubernetesNodePool, error) {
	return KubernetesNodePool{}, nil
}
func (cloud *DOKS) ListNodePool(nodepool *KubernetesNodePool, ctx utils.Context,projectId,companyId, clusterId, token string) (KubernetesNodePool, error) {
	return KubernetesNodePool{}, nil
}
func (cloud *DOKS) UpdateCluster(nodepool *KubernetesNodePool, ctx utils.Context,projectId,companyId, clusterId, token string) (KubernetesNodePool, error) {
	return KubernetesNodePool{}, nil
}
func (cloud *DOKS) UpdateNodePool(nodepool *KubernetesNodePool, ctx utils.Context,projectId,companyId, clusterId, token string) (KubernetesNodePool, error) {
	return KubernetesNodePool{}, nil
}
func (cloud *DOKS) UpgradeVersion(nodepool *KubernetesNodePool, ctx utils.Context,projectId,companyId, clusterId, token string) (KubernetesNodePool, error) {
	return KubernetesNodePool{}, nil
}
func (cloud *DOKS) getVersion(nodepool *KubernetesNodePool, ctx utils.Context,projectId,companyId, clusterId, token string) (KubernetesNodePool, error) {
	return KubernetesNodePool{}, nil
}
func (cloud *DOKS) fetchStatus(ctx utils.Context, clusterId,companyId,projectId string) (KubernetesCluster, error) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != nil {
			return KubernetesCluster{}, err
		}
	}
	clusterId ="fb29ad6f-085e-4328-b1cf-39f71d26de84"
	//list := godo.ListOptions{}
	status,_,err := cloud.Client.Kubernetes.Get(context.Background(),clusterId)
	// cloud.Client.Kubernetes.List(context.Background(),&list)

	if err != nil{
		utils.SendLog(companyId, "Error in cluster creation: "+err.Error(), "info", projectId)
		return KubernetesCluster{}, err
	}
	var cluster *KubernetesCluster
	println(status)

	return &status, nil
}

