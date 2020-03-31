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
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-09-01/skus"
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-02-01/containerservice"
	"github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2018-01-01-preview/authorization"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2016-06-01/subscriptions"
	"github.com/Azure/go-autorest/autorest/to"
	"time"

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
	ResourceSkuClient skus.ResourceSkusClient
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

	cloud.ResourceSkuClient = skus.NewResourceSkusClient(cloud.Subscription)
	cloud.ResourceSkuClient.Authorizer = cloud.Authorizer

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
			"AKS cluster validation for '"+aksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	request := cloud.generateClusterCreateRequest(aksCluster)
	cloud.ProjectId = aksCluster.ProjectId
	//networkInformation := cloud.getAzureNetwork(token, ctx)

	//if len(networkInformation.Definition) > 0 {
	//	for _, AKSnodePool := range *request.ManagedClusterProperties.AgentPoolProfiles {
	//		for _, subnet := range networkInformation.Definition[0].Subnets {
	//			if subnet.Name == *AKSnodePool.VnetSubnetID {
	//				*AKSnodePool.VnetSubnetID = subnet.SubnetId
	//				break
	//			}
	//		}
	//	}
	//}
	cloud.Context = context.Background()
	future, err := cloud.MCClient.CreateOrUpdate(cloud.Context, aksCluster.ResourceGoup, *request.Name, *request)
	if err != nil {
		ctx.SendLogs(
			"AKS cluster creation for '"+aksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}
	err = future.WaitForCompletionRef(context.Background(), cloud.MCClient.Client)
	if err != nil {
		ctx.SendLogs(
			"AKS cluster creation for '"+aksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	AKSclusterResp, err := future.Result(cloud.MCClient)
	if err != nil {
		ctx.SendLogs(
			"AKS cluster creation for '"+aksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}
	if *AKSclusterResp.ProvisioningState != "Succeeded" {
		ctx.SendLogs(
			"AKS cluster creation for '"+aksCluster.Name+"' failed",
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return errors.New("AKS cluster provision state is not succeeded")
	}

	return nil
}

func (cloud *AKS) TerminateCluster(cluster AKSCluster, ctx utils.Context) error {
	if cloud == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	cloud.Context = context.Background()
	_, err := cloud.MCClient.Delete(cloud.Context, cluster.ResourceGoup, cluster.Name)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		ctx.SendLogs(
			"AKS cluster deletion for '"+cluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	for {
		akscluster, err := cloud.MCClient.Get(cloud.Context, cluster.ResourceGoup, cluster.Name)
		if err != nil {
			ctx.SendLogs(
				"AKS cluster deletion for '"+cluster.Name+"' failed: "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
		}

		if akscluster.ProvisioningState == to.StringPtr("Deleting") {
			ctx.SendLogs(
				"AKS cluster deletion for '"+cluster.Name+"' is in progress ",
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)

			time.Sleep(10 * time.Second)
		} else if akscluster.ProvisioningState == to.StringPtr("Deleted") {
			break
		} else {
			ctx.SendLogs(
				"AKS cluster deletion for '"+cluster.Name+"' failed: ",
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)

			return errors.New("AKS cluster deletion for '" + cluster.Name + "' failed: ")
		}
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
	results, err := cloud.MCClient.ListClusterUserCredentials(cloud.Context, cluster.ResourceGoup, cluster.Name)
	if err != nil {
		ctx.SendLogs(
			"AKS getting user credentials for '"+cluster.Name+"' failed: "+err.Error(),
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

	tags := make(map[string]string)
	for key, value := range v.Tags {
		tags[key] = *value
	}

	return AKSCluster{
		ProjectId: cloud.ProjectId,
		Cloud:     models.AKS,
		ClusterProperties: ManagedClusterProperties{
			ProvisioningState: *v.ProvisioningState,
			KubernetesVersion: *v.KubernetesVersion,
			AgentPoolProfiles: agentPoolArr,
			EnableRBAC:        *v.EnableRBAC,
		},
		ResourceID: *v.ID,
		Name:       *v.Name,
		Type:       *v.Type,
		Location:   *v.Location,
	}
}

func generateClusterNodePools(c AKSCluster) *[]containerservice.ManagedClusterAgentPoolProfile {
	AKSNodePools := make([]containerservice.ManagedClusterAgentPoolProfile, len(c.ClusterProperties.AgentPoolProfiles))
	if c.ClusterProperties.IsAdvanced {
		for i, nodepool := range c.ClusterProperties.AgentPoolProfiles {
			AKSNodePools[i].Name = nodepool.Name
			AKSNodePools[i].Count = nodepool.Count
			AKSNodePools[i].OsType = "Linux"
			AKSNodePools[i].VMSize = *nodepool.VMSize
			AKSNodePools[i].OsDiskSizeGB = nodepool.OsDiskSizeGB
			AKSNodePools[i].MaxPods = nodepool.MaxPods
			AKSNodePools[i].Type = "VirtualMachineScaleSets"

			nodelabels := make(map[string]*string)
			for key, value := range nodepool.NodeLabels {
				nodelabels[key] = value
			}
			AKSNodePools[i].NodeLabels = nodelabels

			var nodeTaints []string
			for key, value := range nodepool.NodeTaints {
				nodeTaints = append(nodeTaints, key+"="+*value)
			}
			AKSNodePools[i].NodeTaints = &nodeTaints

			if *nodepool.EnableAutoScaling {
				AKSNodePools[i].EnableAutoScaling = nodepool.EnableAutoScaling
				AKSNodePools[i].MinCount = nodepool.MinCount
				AKSNodePools[i].MaxCount = nodepool.MaxCount
			}

		}
	} else {
		for i, nodepool := range c.ClusterProperties.AgentPoolProfiles {
			AKSNodePools[i].Name = nodepool.Name
			AKSNodePools[i].Count = nodepool.Count
			AKSNodePools[i].OsType = "Linux"
			AKSNodePools[i].VMSize = *nodepool.VMSize
			AKSNodePools[i].Type = "VirtualMachineScaleSets"

			nodelabels := make(map[string]*string)
			nodelabels["AKS-Custer-Node-Pool"] = to.StringPtr(c.ProjectId)
			AKSNodePools[i].NodeLabels = nodelabels
		}
	}

	return &AKSNodePools
}

func generateApiServerAccessProfile(c AKSCluster) *containerservice.ManagedClusterAPIServerAccessProfile {
	var AKSapiServerAccessProfile containerservice.ManagedClusterAPIServerAccessProfile

	if c.ClusterProperties.IsAdvanced {
		if c.ClusterProperties.APIServerAccessProfile.EnablePrivateCluster {
			AKSapiServerAccessProfile.EnablePrivateCluster = to.BoolPtr(true)
		} else {
			AKSapiServerAccessProfile.EnablePrivateCluster = to.BoolPtr(false)
		}

		var authIpRanges []string
		for _, val := range c.ClusterProperties.APIServerAccessProfile.AuthorizedIPRanges {
			authIpRanges = append(authIpRanges, val)
		}

		AKSapiServerAccessProfile.AuthorizedIPRanges = &authIpRanges
	} else {
		AKSapiServerAccessProfile.EnablePrivateCluster = to.BoolPtr(false)
	}

	return &AKSapiServerAccessProfile
}

func (cloud *AKS) generateServicePrincipal(c AKSCluster) *containerservice.ManagedClusterServicePrincipalProfile {
	var AKSservicePrincipal containerservice.ManagedClusterServicePrincipalProfile
	if c.ClusterProperties.IsAdvanced && c.ClusterProperties.IsServicePrincipal {
		AKSservicePrincipal.ClientID = &c.ClusterProperties.ClientID
		AKSservicePrincipal.Secret = &c.ClusterProperties.Secret
	} else {
		AKSservicePrincipal.ClientID = &cloud.ID
		AKSservicePrincipal.Secret = &cloud.Key
	}
	return &AKSservicePrincipal
}

func (cloud *AKS) generateClusterCreateRequest(c AKSCluster) *containerservice.ManagedCluster {
	request := containerservice.ManagedCluster{
		Name:     &c.Name,
		Location: &c.Location,
		ManagedClusterProperties: &containerservice.ManagedClusterProperties{
			DNSPrefix:               generateDnsPrefix(c),
			KubernetesVersion:       generateKubernetesVersion(c),
			AgentPoolProfiles:       generateClusterNodePools(c),
			ServicePrincipalProfile: cloud.generateServicePrincipal(c),
			APIServerAccessProfile:  generateApiServerAccessProfile(c),
			EnableRBAC:              &c.ClusterProperties.EnableRBAC,
			AddonProfiles:           generateAddonProfiles(c),
			NetworkProfile:          generateNetworkProfile(c),
			//WindowsProfile:          generateWindowsProfile(),
		},
		Identity: generateClusterIdentity(),
		Tags:     generateClusterTags(c),
	}
	return &request
}

func generateWindowsProfile() *containerservice.ManagedClusterWindowsProfile {
	var AKSwindowProfile containerservice.ManagedClusterWindowsProfile
	AKSwindowProfile.AdminPassword = to.StringPtr("cloudplex")
	AKSwindowProfile.AdminUsername = to.StringPtr("cloudplex")
	return &AKSwindowProfile
}

func (cloud *AKS) GetKubernetesVersions(ctx utils.Context) (*containerservice.OrchestratorVersionProfileListResult, error) {
	if cloud == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return nil, err
		}
	}

	cloud.Context = context.Background()
	result, err := cloud.KubeVersionClient.ListOrchestrators(cloud.Context, cloud.Region, "Microsoft.ContainerService")
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	return &result, nil
}

func generateNetworkProfile(c AKSCluster) *containerservice.NetworkProfileType {

	if c.ClusterProperties.IsExpert {
		var AKSnetworkProfile containerservice.NetworkProfileType
		AKSnetworkProfile.PodCidr = &c.ClusterProperties.PodCidr
		AKSnetworkProfile.DNSServiceIP = &c.ClusterProperties.DNSServiceIP
		AKSnetworkProfile.ServiceCidr = &c.ClusterProperties.ServiceCidr
		AKSnetworkProfile.DockerBridgeCidr = &c.ClusterProperties.DockerBridgeCidr
		return &AKSnetworkProfile
	}
	return nil
}

func generateClusterTags(c AKSCluster) map[string]*string {
	AKSclusterTags := make(map[string]*string)
	if c.ClusterProperties.IsAdvanced {
		for key, value := range c.ClusterProperties.ClusterTags {
			AKSclusterTags[key] = &value
		}
	} else {
		AKSclusterTags["AKS-Cluster"] = &c.ProjectId
	}

	return AKSclusterTags
}

func generateClusterIdentity() *containerservice.ManagedClusterIdentity {
	var AKSclusterIdentity containerservice.ManagedClusterIdentity
	AKSclusterIdentity.Type = "SystemAssigned"
	return &AKSclusterIdentity
}

func generateAddonProfiles(c AKSCluster) map[string]*containerservice.ManagedClusterAddonProfile {
	AKSaddon := make(map[string]*containerservice.ManagedClusterAddonProfile)
	if c.ClusterProperties.IsAdvanced && c.ClusterProperties.IsHttpRouting {
		var httpAddOn containerservice.ManagedClusterAddonProfile
		httpAddOn.Enabled = to.BoolPtr(true)
		AKSaddon["httpApplicationRouting"] = &httpAddOn
	} else {
		var httpAddOn containerservice.ManagedClusterAddonProfile
		httpAddOn.Enabled = to.BoolPtr(false)
		AKSaddon["httpApplicationRouting"] = &httpAddOn
	}

	return AKSaddon
}

func generateKubernetesVersion(c AKSCluster) *string {
	if c.ClusterProperties.IsAdvanced {
		return to.StringPtr(c.ClusterProperties.KubernetesVersion)
	} else {
		return to.StringPtr("1.15.10")
	}
}

func generateDnsPrefix(c AKSCluster) *string {
	if c.ClusterProperties.IsAdvanced {
		return to.StringPtr(c.ClusterProperties.DNSPrefix + "-dns")
	} else {
		return to.StringPtr(c.Name + "-dns")
	}
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
	AKScluster, err := cloud.MCClient.Get(cloud.Context, cluster.ResourceGoup, cluster.Name)
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
		cluster.ClusterProperties.AgentPoolProfiles[index].VMSize = &agentPool.VMSize
		cluster.ClusterProperties.AgentPoolProfiles[index].OsType = &agentPool.OsType
		cluster.ClusterProperties.AgentPoolProfiles[index].Count = agentPool.Count
	}
	cluster.ResourceID = *AKScluster.ID
	cluster.Type = *AKScluster.Type
	cluster.ClusterProperties.ProvisioningState = *AKScluster.ProvisioningState
	cluster.ClusterProperties.KubernetesVersion = *AKScluster.KubernetesVersion
	cluster.ClusterProperties.DNSPrefix = *AKScluster.DNSPrefix
	cluster.ClusterProperties.Fqdn = *AKScluster.Fqdn
	cluster.ClusterProperties.EnableRBAC = *AKScluster.EnableRBAC

	cluster.ClusterProperties.DNSServiceIP = *AKScluster.ManagedClusterProperties.NetworkProfile.DNSServiceIP
	cluster.ClusterProperties.PodCidr = *AKScluster.ManagedClusterProperties.NetworkProfile.PodCidr
	cluster.ClusterProperties.ServiceCidr = *AKScluster.ManagedClusterProperties.NetworkProfile.ServiceCidr
	cluster.ClusterProperties.DockerBridgeCidr = *AKScluster.ManagedClusterProperties.NetworkProfile.DockerBridgeCidr

	return nil
}

func GetAKSSupportedVms(ctx utils.Context) []containerservice.VMSizeTypes {
	return containerservice.PossibleVMSizeTypesValues()
}

func validate(aksCluster AKSCluster) error {
	if aksCluster.ProjectId == "" {
		return errors.New("project id is required")
	} else if aksCluster.Name == "" {
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
