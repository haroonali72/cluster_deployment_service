package doks

import (
	"antelope/models"
	"antelope/models/utils"
	"antelope/models/vault"
	"context"
	"errors"
	"github.com/astaxie/beego"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
	"log"
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
func getWoodpecker() string {
	return beego.AppConfig.String("woodpecker_url") + models.WoodpeckerEnpoint
}

func (cloud *DOKS) init(ctx utils.Context) error {
	if cloud.Client != nil {
		return nil
	}

	if cloud.AccessKey == "" {
		text := "Invalid cloud credentials"
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

func (cloud *DOKS) createCluster(cluster KubernetesCluster, ctx utils.Context, companyId , token string, credentials vault.DOCredentials) (KubernetesCluster, CustomError) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != nil {
			return cluster, ApiError(err,credentials,ctx,companyId)
		}
	}


	utils.SendLog(companyId, "Creating DOKS Cluster With ID : "+cluster.ProjectId, "info", cluster.ProjectId)

	/*	list := godo.ListOptions{}
		re,_,err :=cloud.Client.Kubernetes.List(context.Background(),&list)
		fmt.Println(re)
	*/
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
		NodePools: nodepool,
		//MaintenancePolicy: cluster.MaintenancePolicy,
		AutoUpgrade: cluster.AutoUpgrade,
	}

	clus, _, err := cloud.Client.Kubernetes.Create(context.Background(), &input)
	if err != nil {
		cluErr :=ApiError(err,credentials,ctx,companyId)
		utils.SendLog(companyId, "Error in cluster creation : "+err.Error(), models.LOGGING_LEVEL_ERROR, cluster.ProjectId)
		return cluster,cluErr
	}
	cluster.ID = clus.ID
	time.Sleep(2 *30 * time.Second)
	status, _, err := cloud.Client.Kubernetes.Get(context.Background(), clus.ID)
	for  status.Status.State != "running"{
		time.Sleep(30 * time.Second)
		status, _, err = cloud.Client.Kubernetes.Get(context.Background(), clus.ID)
	}

	time.Sleep(15 * time.Second)
	return cluster, CustomError{}
}
func (cloud *DOKS) deleteCluster(cluster KubernetesCluster, ctx utils.Context, projectId, companyId string) CustomError {
	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != nil {
			return CustomError{}
		}
	}

	utils.SendLog(companyId, "Deleting DOKS Cluster With ID : "+cluster.ProjectId, "info", cluster.ProjectId)

	/*		list := godo.ListOptions{}
			re,_,err :=cloud.Client.Kubernetes.List(context.Background(),&list)
			fmt.Println(re)
	*/
	_, err := cloud.Client.Kubernetes.Delete(context.Background(), cluster.ID)
	if err != nil {
		utils.SendLog(companyId, "Error in cluster creation: "+err.Error(), "info", cluster.ProjectId)
		return ApiError(err,vault.DOCredentials{},ctx,companyId)
	}

	utils.SendLog(companyId, "DOKS cluster deleted successfully : "+cluster.ProjectId, "info", cluster.ProjectId)
	return CustomError{}
}
func (cloud *DOKS) GetKubeConfig(ctx utils.Context, cluster KubernetesCluster) (KubernetesConfig, CustomError) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != nil {
			return KubernetesConfig{}, CustomError{}
		}
	}

	config, _, err := cloud.Client.Kubernetes.GetKubeConfig(context.Background(),cluster.ID)
	if err != nil {
		utils.SendLog(cluster.CompanyId, "Error in getting kubernetes config file: "+err.Error(), "error", cluster.ProjectId)
		return KubernetesConfig{}, ApiError(err,vault.DOCredentials{},ctx,"")
	}

	var con KubernetesClusterConfig
	con.KubeconfigYAML = config.KubeconfigYAML
	kubeFile := KubernetesConfig{}

	err = yaml.Unmarshal([]byte(config.KubeconfigYAML), &kubeFile)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	 utils.SendLog(cluster.CompanyId, "DOKS kubernetes config file fetched successfully : "+cluster.ProjectId, "info", cluster.ProjectId)

	return kubeFile, CustomError{}
}

func (cloud *DOKS) UpdateCluster(nodepool *KubernetesNodePool, ctx utils.Context, projectId, companyId, clusterId, token string) (KubernetesNodePool, error) {
	return KubernetesNodePool{}, nil
}
func (cloud *DOKS) UpdateNodePool(nodepool *KubernetesNodePool, ctx utils.Context, projectId, companyId, clusterId, token string) (KubernetesNodePool, error) {
	return KubernetesNodePool{}, nil
}
func (cloud *DOKS) UpgradeVersion(nodepool *KubernetesNodePool, ctx utils.Context, projectId, companyId, clusterId, token string) (KubernetesNodePool, error) {
	return KubernetesNodePool{}, nil
}
func (cloud *DOKS) fetchStatus(ctx utils.Context, clusterId, companyId, projectId string) (*godo.KubernetesCluster,  CustomError) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != nil {
			return &godo.KubernetesCluster{}, CustomError{}
		}
	}
	//clusterId ="b01f9429-459b-4fc6-9726-ba9c21e88272"
	status, _, err := cloud.Client.Kubernetes.Get(context.Background(), clusterId)
	if err != nil {
		utils.SendLog(companyId, "Error in cluster creation: "+err.Error(), "info", projectId)
		return &godo.KubernetesCluster{}, ApiError(err,vault.DOCredentials{},ctx,companyId)
	}
	return status, CustomError{}
}
func (cloud *DOKS) GetServerConfig(ctx utils.Context, companyId string) (*godo.KubernetesOptions, error) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != nil {
			return &godo.KubernetesOptions{}, err
		}
	}
	options, _, err := cloud.Client.Kubernetes.GetOptions(context.Background())
	if err != nil {
		utils.SendLog(companyId, "Error in gettin kubernetes config file: "+err.Error(), "error", "")
		return &godo.KubernetesOptions{}, err
	}

	utils.SendLog(companyId, "DOKS kubernetes config file fetched successfully : ", "info", "")

	return options, nil
}
