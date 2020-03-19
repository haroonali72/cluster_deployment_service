package aks

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"context"
	"encoding/json"
	"errors"
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-02-01/containerservice"
	"github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2018-01-01-preview/authorization"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2016-06-01/subscriptions"
	"github.com/Azure/go-autorest/autorest/to"

	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/astaxie/beego"
)

type AKS struct {
	Authorizer        *autorest.BearerAuthorizer
	Location          subscriptions.Client
	MCClient          containerservice.ManagedClustersClient
	KubeVersionClient containerservice.ContainerServicesClient
	Context           context.Context
	ProjectId         string
	ID                string
	Key               string
	Tenant            string
	Subscription      string
	Region            string
	Resources         map[string]interface{}
	RoleAssignment    authorization.RoleAssignmentsClient
	RoleDefinition    authorization.RoleDefinitionsClient
}

func (cloud *AKS) init() error {
	if cloud.Authorizer != nil {
		return nil
	}

	if cloud.ID == "" || cloud.Key == "" || cloud.Tenant == "" || cloud.Subscription == "" || cloud.Region == "" {
		text := "invalid cloud credentials"
		beego.Error(text)
		return errors.New(text)
	}

	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, cloud.Tenant)
	if err != nil {
		panic(err)
	}

	spt, err := adal.NewServicePrincipalToken(*oauthConfig, cloud.ID, cloud.Key, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return err
	}
	cloud.Context = context.Background()
	cloud.Authorizer = autorest.NewBearerAuthorizer(spt)

	cloud.MCClient = containerservice.NewManagedClustersClient(cloud.Subscription)
	cloud.MCClient.Authorizer = cloud.Authorizer

	cloud.KubeVersionClient = containerservice.NewContainerServicesClient(cloud.Subscription)
	cloud.KubeVersionClient.Authorizer = cloud.Authorizer

	cloud.RoleAssignment = authorization.NewRoleAssignmentsClient(cloud.Subscription)
	cloud.RoleAssignment.Authorizer = cloud.Authorizer

	cloud.RoleDefinition = authorization.NewRoleDefinitionsClient(cloud.Subscription)
	cloud.RoleDefinition.Authorizer = cloud.Authorizer

	cloud.Resources = make(map[string]interface{})
	cloud.Location = subscriptions.NewClient()
	cloud.Location.Authorizer = cloud.Authorizer
	return nil
}

func (cloud *AKS) ListClustersByResourceGroup(ctx utils.Context, resourceGroupName string) ([]AKSCluster, error) {
	if cloud == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return nil, err
		}
	}

	cloud.Context = context.Background()
	pages, err := cloud.MCClient.ListByResourceGroup(cloud.Context, resourceGroupName)
	if err != nil {
		ctx.SendLogs(
			"AKS list clusters within resource group '"+resourceGroupName+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return nil, err
	}

	result := []AKSCluster{}
	for pages.NotDone() {
		for _, v := range pages.Values() {
			result = append(result, cloud.generateClusterFromResponse(v))
		}
		_ = pages.Next()
	}

	return result, nil
}

func (cloud *AKS) ListClusters(ctx utils.Context) ([]AKSCluster, error) {
	if cloud == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return nil, err
		}
	}

	cloud.Context = context.Background()
	pages, err := cloud.MCClient.List(cloud.Context)
	if err != nil {
		ctx.SendLogs(
			"AKS list clusters within specified subscription failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return nil, err
	}

	result := []AKSCluster{}
	for pages.NotDone() {
		for _, v := range pages.Values() {
			result = append(result, cloud.generateClusterFromResponse(v))
		}
		_ = pages.Next()
	}

	return result, nil
}

func (cloud *AKS) GetCluster(ctx utils.Context, resourceGroupName, clusterName string) (*AKSCluster, error) {
	if cloud == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return nil, err
		}
	}

	cloud.Context = context.Background()
	result, err := cloud.MCClient.Get(cloud.Context, resourceGroupName, clusterName)
	if err != nil {
		ctx.SendLogs(
			"AKS get cluster within resource group '"+resourceGroupName+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return nil, err
	}

	aksCluster := cloud.generateClusterFromResponse(result)
	return &aksCluster, nil
}

func (cloud *AKS) CreateCluster(aksCluster AKSCluster, token string, ctx utils.Context) error {
	err := validate(aksCluster)
	if err != nil {
		ctx.SendLogs(
			"AKS cluster validation for '"+*aksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	request := cloud.generateClusterCreateRequest(aksCluster)
	cloud.Context = context.Background()
	future, err := cloud.MCClient.CreateOrUpdate(cloud.Context, aksCluster.ResourceGoup, *request.Name, *request)
	if err != nil {
		ctx.SendLogs(
			"AKS cluster creation for '"+*aksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}
	err = future.WaitForCompletionRef(context.Background(), cloud.MCClient.Client)
	if err != nil {
		ctx.SendLogs(
			"AKS cluster creation for '"+*aksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	AKSclusterResp, err := future.Result(cloud.MCClient)
	if err != nil {
		ctx.SendLogs(
			"AKS cluster creation for '"+*aksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}
	if *AKSclusterResp.ProvisioningState != "Succeeded" {
		ctx.SendLogs(
			"AKS cluster creation for '"+*aksCluster.Name+"' failed",
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return errors.New("AKS cluster provision state is not succeeded")
	}

	return nil
}

//func (cloud *GKE) UpdateMasterVersion(clusterName, newVersion string, ctx utils.Context) error {
//	if newVersion == "" {
//		return nil
//	}
//
//	_, err := cloud.Client.Projects.Zones.Clusters.Update(
//		cloud.ProjectId,
//		cloud.Zone,
//		clusterName,
//		&gke.UpdateClusterRequest{
//			Update: &gke.ClusterUpdate{
//				DesiredMasterVersion: newVersion,
//			},
//		},
//	).Context(context.Background()).Do()
//	if err != nil {
//		ctx.SendLogs(
//			"GKE cluster update request for '"+clusterName+"' failed: "+err.Error(),
//			models.LOGGING_LEVEL_ERROR,
//			models.Backend_Logging,
//		)
//		return err
//	}
//
//	return cloud.waitForCluster(clusterName, ctx)
//}
//
//func (cloud *GKE) UpdateNodeVersion(clusterName, nodeName, newVersion string, ctx utils.Context) error {
//	if newVersion == "" {
//		return nil
//	}
//
//	_, err := cloud.Client.Projects.Zones.Clusters.NodePools.Update(
//		cloud.ProjectId,
//		cloud.Zone,
//		clusterName,
//		nodeName,
//		&gke.UpdateNodePoolRequest{
//			NodeVersion: newVersion,
//		},
//	).Context(context.Background()).Do()
//	if err != nil {
//		ctx.SendLogs(
//			"GKE node update request for cluster '"+clusterName+"' and node '"+nodeName+"' failed: "+err.Error(),
//			models.LOGGING_LEVEL_ERROR,
//			models.Backend_Logging,
//		)
//		return err
//	}
//
//	return cloud.waitForNodePool(clusterName, nodeName, ctx)
//}
//
//func (cloud *GKE) UpdateNodeCount(clusterName, nodeName string, newCount int64, ctx utils.Context) error {
//	if newCount == 0 {
//		return nil
//	}
//
//	_, err := cloud.Client.Projects.Zones.Clusters.NodePools.SetSize(
//		cloud.ProjectId,
//		cloud.Zone,
//		clusterName,
//		nodeName,
//		&gke.SetNodePoolSizeRequest{
//			NodeCount: newCount,
//		},
//	).Context(context.Background()).Do()
//	if err != nil {
//		ctx.SendLogs(
//			"GKE node update request for cluster '"+clusterName+"' and node '"+nodeName+"' failed: "+err.Error(),
//			models.LOGGING_LEVEL_ERROR,
//			models.Backend_Logging,
//		)
//		return err
//	}
//
//	return cloud.waitForNodePool(clusterName, nodeName, ctx)
//}
//
func (cloud *AKS) DeleteCluster(cluster AKSCluster, ctx utils.Context) error {
	if cloud == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	cloud.Context = context.Background()
	_, err := cloud.MCClient.Delete(cloud.Context, cluster.ResourceGoup, *cluster.Name)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		ctx.SendLogs(
			"AKS cluster deletion for '"+*cluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}
	//err = future.WaitForCompletionRef(cloud.Context, cloud.MCClient.Client)
	//if err != nil {
	//	ctx.SendLogs(
	//		"AKS cluster deletion for '"+*cluster.Name+"' failed: "+err.Error(),
	//		models.LOGGING_LEVEL_ERROR,
	//		models.Backend_Logging,
	//	)
	//	return err
	//}

	return nil
}

func (cloud *AKS) GetKubeConfig(ctx utils.Context, cluster AKSCluster) (*containerservice.CredentialResult, error) {
	if cloud == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return nil, err
		}
	}

	cloud.Context = context.Background()
	results, err := cloud.MCClient.ListClusterUserCredentials(cloud.Context, cluster.ResourceGoup, *cluster.Name)
	if err != nil {
		ctx.SendLogs(
			"AKS getting user credentials for '"+*cluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return nil, err
	}

	for _, kubeconfig := range *results.Kubeconfigs {
		return &kubeconfig, nil
	}

	return nil, nil
}

//
//func (cloud *GKE) waitForCluster(clusterName string, ctx utils.Context) error {
//	message := ""
//	for {
//		cluster, err := cloud.Client.Projects.Zones.Clusters.Get(
//			cloud.ProjectId,
//			cloud.Zone,
//			clusterName,
//		).Context(context.Background()).Do()
//		if err != nil {
//			ctx.SendLogs(
//				"GKE cluster creation/updation for '"+clusterName+"' failed: "+err.Error(),
//				models.LOGGING_LEVEL_ERROR,
//				models.Backend_Logging,
//			)
//			return err
//		}
//		if cluster.Status == statusRunning {
//			ctx.SendLogs(
//				"GKE cluster '"+clusterName+"' is running.",
//				models.LOGGING_LEVEL_INFO,
//				models.Backend_Logging,
//			)
//			return nil
//		}
//		if cluster.Status != message {
//			ctx.SendLogs(
//				"GKE cluster '"+clusterName+"' is creating/updating.",
//				models.LOGGING_LEVEL_INFO,
//				models.Backend_Logging,
//			)
//			message = cluster.Status
//		}
//		time.Sleep(time.Second * 5)
//	}
//}
//
//func (cloud *GKE) waitForNodePool(clusterName, nodeName string, ctx utils.Context) error {
//	message := ""
//	for {
//		nodepool, err := cloud.Client.Projects.Zones.Clusters.NodePools.Get(
//			cloud.ProjectId,
//			cloud.Zone,
//			clusterName,
//			nodeName,
//		).Context(context.Background()).Do()
//		if err != nil {
//			ctx.SendLogs(
//				"GKE node creation/updation for cluster '"+clusterName+"' and node '"+nodeName+"' failed: "+err.Error(),
//				models.LOGGING_LEVEL_ERROR,
//				models.Backend_Logging,
//			)
//			return err
//		}
//		if nodepool.Status == statusRunning {
//			ctx.SendLogs(
//				"GKE node '"+nodeName+"' for cluster '"+clusterName+"' is running.",
//				models.LOGGING_LEVEL_INFO,
//				models.Backend_Logging,
//			)
//			return nil
//		}
//		if nodepool.Status != message {
//			ctx.SendLogs(
//				"GKE node '"+nodeName+"' for cluster '"+clusterName+"' is creating/updating.",
//				models.LOGGING_LEVEL_INFO,
//				models.Backend_Logging,
//			)
//			message = nodepool.Status
//		}
//		time.Sleep(time.Second * 5)
//	}
//}
//
//func (cloud *GKE) getGKEVersions(ctx utils.Context) (*gke.ServerConfig, error) {
//	config, err := cloud.Client.Projects.Zones.GetServerconfig("*", cloud.Zone).
//		Context(context.Background()).
//		Do()
//
//	if err != nil {
//		ctx.SendLogs(
//			"GKE server config for '"+cloud.ProjectId+"' failed: "+err.Error(),
//			models.LOGGING_LEVEL_ERROR,
//			models.Backend_Logging,
//		)
//		return nil, err
//	}
//
//	return config, nil
//}

func (cloud *AKS) getAzureNetwork(token string, ctx utils.Context) (azureNetwork types.AzureNetwork) {
	url := getNetworkHost(string(models.Azure), cloud.ProjectId)

	network, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return azureNetwork
	}

	err = json.Unmarshal(network.([]byte), &azureNetwork)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return azureNetwork
	}

	return azureNetwork
}

func (cloud *AKS) generateClusterFromResponse(v containerservice.ManagedCluster) AKSCluster {
	var agentPoolArr []ManagedClusterAgentPoolProfile
	for _, aksAgentPool := range *v.AgentPoolProfiles {
		var pool ManagedClusterAgentPoolProfile
		pool.Name = aksAgentPool.Name
		pool.VnetSubnetID = aksAgentPool.VnetSubnetID
		pool.Count = aksAgentPool.Count
		pool.MaxPods = aksAgentPool.MaxPods
		agentPoolArr = append(agentPoolArr, pool)
	}
	return AKSCluster{
		ProjectId: cloud.ProjectId,
		Cloud:     models.AKS,
		ClusterProperties: &ManagedClusterProperties{
			ProvisioningState: v.ProvisioningState,
			KubernetesVersion: v.KubernetesVersion,
			AgentPoolProfiles: agentPoolArr,
		},
		ResourceID: v.ID,
		Name:       v.Name,
		Type:       v.Type,
		Location:   v.Location,
		Tags:       v.Tags,
	}
}

func (cloud *AKS) generateClusterCreateRequest(c AKSCluster) *containerservice.ManagedCluster {
	var AKSNodePools []containerservice.ManagedClusterAgentPoolProfile
	var AKSapiServerAccessProfile containerservice.ManagedClusterAPIServerAccessProfile
	if c.ClusterProperties != nil {
		for _, nodepool := range c.ClusterProperties.AgentPoolProfiles {
			var AKSnodepool containerservice.ManagedClusterAgentPoolProfile
			AKSnodepool.Name = nodepool.Name
			AKSnodepool.Count = nodepool.Count
			AKSnodepool.VnetSubnetID = nodepool.VnetSubnetID
			AKSnodepool.VMSize = nodepool.VMSize
			if nodepool.EnableAutoScaling != nil && *nodepool.EnableAutoScaling {
				AKSnodepool.EnableAutoScaling = nodepool.EnableAutoScaling
				AKSnodepool.MinCount = nodepool.MinCount
				AKSnodepool.MaxCount = nodepool.MaxCount
			}
			AKSNodePools = append(AKSNodePools, AKSnodepool)
		}

		if c.ClusterProperties.APIServerAccessProfile.EnablePrivateCluster != nil && *c.ClusterProperties.APIServerAccessProfile.EnablePrivateCluster {
			AKSapiServerAccessProfile.EnablePrivateCluster = to.BoolPtr(true)
		} else {
			AKSapiServerAccessProfile.EnablePrivateCluster = to.BoolPtr(true)
		}
	}

	var AKSservicePrincipal containerservice.ManagedClusterServicePrincipalProfile
	AKSservicePrincipal.ClientID = &cloud.ID
	AKSservicePrincipal.Secret = &cloud.Key
	request := containerservice.ManagedCluster{
		Name:     c.Name,
		Location: c.Location,
		ManagedClusterProperties: &containerservice.ManagedClusterProperties{
			DNSPrefix:               to.StringPtr(*c.Name + "-dns"),
			AgentPoolProfiles:       &AKSNodePools,
			ServicePrincipalProfile: &AKSservicePrincipal,
			APIServerAccessProfile:  &AKSapiServerAccessProfile,
		},
	}
	return &request
}

func (cloud *AKS) fetchClusterStatus(cluster *AKSCluster, ctx utils.Context) error {
	if cloud == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	cloud.Context = context.Background()
	AKScluster, err := cloud.MCClient.Get(cloud.Context, cluster.ResourceGoup, *cluster.Name)
	if err != nil {
		ctx.SendLogs(
			"AKS get cluster within resource group '"+cluster.ResourceGoup+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	for index, agentPool := range *AKScluster.AgentPoolProfiles {
		cluster.ClusterProperties.AgentPoolProfiles[index].Name = agentPool.Name
		cluster.ClusterProperties.AgentPoolProfiles[index].OsDiskSizeGB = agentPool.OsDiskSizeGB
		cluster.ClusterProperties.AgentPoolProfiles[index].VnetSubnetID = agentPool.VnetSubnetID
		cluster.ClusterProperties.AgentPoolProfiles[index].VMSize = agentPool.VMSize
		cluster.ClusterProperties.AgentPoolProfiles[index].OsType = agentPool.OsType
		cluster.ClusterProperties.AgentPoolProfiles[index].Count = agentPool.Count
	}
	cluster.ResourceID = AKScluster.ID
	cluster.Type = AKScluster.Type
	cluster.ClusterProperties.ProvisioningState = AKScluster.ProvisioningState
	cluster.ClusterProperties.KubernetesVersion = AKScluster.KubernetesVersion
	cluster.ClusterProperties.DNSPrefix = AKScluster.DNSPrefix
	cluster.ClusterProperties.Fqdn = AKScluster.Fqdn

	var networkProfile NetworkProfileType
	networkProfile.DNSServiceIP = AKScluster.ManagedClusterProperties.NetworkProfile.DNSServiceIP
	networkProfile.PodCidr = AKScluster.ManagedClusterProperties.NetworkProfile.PodCidr
	networkProfile.ServiceCidr = AKScluster.ManagedClusterProperties.NetworkProfile.ServiceCidr
	networkProfile.DockerBridgeCidr = AKScluster.ManagedClusterProperties.NetworkProfile.DockerBridgeCidr

	cluster.ClusterProperties.NetworkProfile = &networkProfile
	return nil
}

//
//func (cloud *GKE) deleteCluster(cluster GKECluster, ctx utils.Context) error {
//	if cloud.Client == nil {
//		err := cloud.init()
//		if err != nil {
//			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
//			return err
//		}
//	}
//
//	_, err := cloud.Client.Projects.Zones.Clusters.Delete(cloud.ProjectId, cloud.Zone, cluster.Name).Do()
//	if err != nil {
//		ctx.SendLogs(
//			"GKE delete cluster for '"+cloud.ProjectId+"' failed: "+err.Error(),
//			models.LOGGING_LEVEL_ERROR,
//			models.Backend_Logging,
//		)
//		return err
//	}
//
//	return nil
//}

func GetAKSSupportedVms(ctx utils.Context) []containerservice.VMSizeTypes {
	return containerservice.PossibleVMSizeTypesValues()
}

func validate(aksCluster AKSCluster) error {
	if aksCluster.ProjectId == "" {
		return errors.New("project id is required")
	} else if *aksCluster.Name == "" {
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

func GetAKS(credentials vault.AzureCredentials) (AKS, error) {
	return AKS{
		ID:           credentials.ClientId,
		Tenant:       credentials.TenantId,
		Key:          credentials.ClientSecret,
		Subscription: credentials.SubscriptionId,
		Region:       credentials.Location,
	}, nil
}
