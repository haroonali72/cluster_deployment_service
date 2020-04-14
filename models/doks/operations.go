package doks

import (
	"antelope/models"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"context"
	"github.com/astaxie/beego"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
	"log"
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
		return types.CustomCPError{StatusCode:"500",Description:text,}
	}

	tokenSource := &TokenSource{
		AccessToken: cloud.AccessKey,
	}

	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	cloud.Client = godo.NewClient(oauthClient)
	cloud.Resources = make(map[string][]string)
	return types.CustomCPError{}
}

func (cloud *DOKS) createCluster(cluster KubernetesCluster, ctx utils.Context, token string, credentials vault.DOCredentials) (KubernetesCluster, types.CustomCPError) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err.Description != "" {
			return cluster, err
		}
	}
	ctx.SendLogs(
		"DOKS cluster creation of "+cluster.Name+"' submitted ",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
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
		NodePools:   nodepool,
		//MaintenancePolicy: cluster.MaintenancePolicy,
		AutoUpgrade: cluster.AutoUpgrade,
	}

	clus, _, err := cloud.Client.Kubernetes.Create(context.Background(), &input)
	if err != nil {
		utils.SendLog(ctx.Data.Company, "Error in cluster creation : "+err.Error(), models.LOGGING_LEVEL_ERROR, cluster.ProjectId)
		ctx.SendLogs("DOKS cluster creation of '"+cluster.Name+"' failed: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging, )
		return cluster, ApiError(err, credentials, ctx)
	}

	cluster.ID = clus.ID

	time.Sleep(2 * 30 * time.Second)

	status, _, err := cloud.Client.Kubernetes.Get(context.Background(), clus.ID)

	for status.Status.State != "running" {
		time.Sleep(30 * time.Second)
		status, _, err = cloud.Client.Kubernetes.Get(context.Background(), clus.ID)
	}

	time.Sleep(15 * time.Second)

	return cluster, types.CustomCPError{}
}

func (cloud *DOKS) deleteCluster(cluster KubernetesCluster, ctx utils.Context) types.CustomCPError {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err.Description != ""{
			return  err
		}
	}

	_, err := cloud.Client.Kubernetes.Delete(context.Background(), cluster.ID)
	if err != nil {
		utils.SendLog(ctx.Data.Company, "Error in cluster termination: "+err.Error(), models.LOGGING_LEVEL_INFO, ctx.Data.ProjectId)
		return ApiError(err,vault.DOCredentials{},ctx)
	}

	return types.CustomCPError{}
}

func (cloud *DOKS) GetKubeConfig(ctx utils.Context, cluster KubernetesCluster) (KubernetesConfig, types.CustomCPError) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err.Description != "" {
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
		return KubernetesConfig{}, ApiError(err,vault.DOCredentials{},ctx)
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

func (cloud *DOKS) UpdateCluster(nodepool *KubernetesNodePool, ctx utils.Context, clusterId, token string) (KubernetesNodePool,types.CustomCPError) {
	return KubernetesNodePool{}, types.CustomCPError{}
}

func (cloud *DOKS) UpdateNodePool(nodepool *KubernetesNodePool, ctx utils.Context,  clusterId, token string) (KubernetesNodePool, types.CustomCPError) {
	return KubernetesNodePool{},types.CustomCPError{}
}

func (cloud *DOKS) UpgradeVersion(nodepool *KubernetesNodePool, ctx utils.Context, clusterId, token string) (KubernetesNodePool,types.CustomCPError) {
	return KubernetesNodePool{}, types.CustomCPError{}
}

func (cloud *DOKS) fetchStatus(ctx utils.Context, clusterId string) (*godo.KubernetesCluster,  types.CustomCPError) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err.Description !="" {
			return &godo.KubernetesCluster{},err
		}
	}

	status, _, err := cloud.Client.Kubernetes.Get(context.Background(), clusterId)
	if err != nil {
		ctx.SendLogs("DOKS get cluster status  failed: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging, )
		return &godo.KubernetesCluster{}, ApiError(err,vault.DOCredentials{},ctx)
	}

	return status, types.CustomCPError{}
}

func (cloud *DOKS) GetServerConfig(ctx utils.Context) (*godo.KubernetesOptions,types.CustomCPError) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err.Description != "" {
			return &godo.KubernetesOptions{}, err
		}
	}

	options, _, err := cloud.Client.Kubernetes.GetOptions(context.Background())
	if err != nil {
		return &godo.KubernetesOptions{}, ApiError(err,vault.DOCredentials{},ctx)
	}

	return options,types.CustomCPError{}
}
