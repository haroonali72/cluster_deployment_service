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
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-09-01/skus"
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-02-01/containerservice"
	"github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2018-01-01-preview/authorization"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2016-06-01/subscriptions"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/astaxie/beego"
	"io/ioutil"
	"os"
	"strings"
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

type SkuResp struct {
	Locations    []string
	Name         string
	ResourceType string
}

func (cloud *AKS) init() types.CustomCPError {
	if cloud.Authorizer != nil {
		return types.CustomCPError{}
	}

	if cloud.ID == "" || cloud.Key == "" || cloud.Tenant == "" || cloud.Subscription == "" || cloud.Region == "" {
		text := "invalid cloud credentials"
		beego.Error(text)
		return ApiError(errors.New(text), text, 401)
	}

	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, cloud.Tenant)
	if err != nil {
		panic(err)
	}

	spt, err := adal.NewServicePrincipalToken(*oauthConfig, cloud.ID, cloud.Key, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return ApiError(err, "", 502)
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
	return types.CustomCPError{}
}

func (cloud *AKS) ListClustersByResourceGroup(ctx utils.Context, resourceGroupName string) ([]AKSCluster, types.CustomCPError) {
	if cloud == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
		return nil, ApiError(err, "Error while getting cluster list", 502)
	}

	result := []AKSCluster{}
	for pages.NotDone() {
		for _, v := range pages.Values() {
			result = append(result, cloud.generateClusterFromResponse(v))
		}
		_ = pages.Next()
	}

	return result, types.CustomCPError{}
}

func (cloud *AKS) ListClusters(ctx utils.Context) ([]AKSCluster, types.CustomCPError) {
	if cloud == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
		return nil, ApiError(err, "Error while getting cluster list", 502)
	}

	result := []AKSCluster{}
	for pages.NotDone() {
		for _, v := range pages.Values() {
			result = append(result, cloud.generateClusterFromResponse(v))
		}
		_ = pages.Next()
	}

	return result, types.CustomCPError{}
}

func (cloud *AKS) GetCluster(ctx utils.Context, resourceGroupName, clusterName string) (*AKSCluster, types.CustomCPError) {
	if cloud == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
		return nil, ApiError(err, "Error while getting cluster", 502)
	}

	aksCluster := cloud.generateClusterFromResponse(result)
	return &aksCluster, types.CustomCPError{}
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

	//Network will be added in every case BASIC, ADVANCE, EXPERT
	networkInformation := cloud.getAzureNetwork(token, ctx)
	if len(networkInformation.Definition) > 0 {
		for _, AKSnodePool := range *request.ManagedClusterProperties.AgentPoolProfiles {
			for _, subnet := range networkInformation.Definition[0].Subnets {
				if subnet.Name == *AKSnodePool.VnetSubnetID {
					*AKSnodePool.VnetSubnetID = subnet.SubnetId
					break
				}
			}
		}
	}
	//Network will be added in every case BASIC, ADVANCE, EXPERT

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

func (cloud *AKS) TerminateCluster(cluster AKSCluster, ctx utils.Context) types.CustomCPError {
	if cloud == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	cloud.Context = context.Background()
	future, err := cloud.MCClient.Delete(cloud.Context, cluster.ResourceGoup, cluster.Name)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		ctx.SendLogs(
			"AKS cluster deletion for '"+cluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return ApiError(err, "", 502)
	}

	//for {
	//	time.Sleep(10 * time.Second)
	//	akscluster, err := cloud.MCClient.Get(cloud.Context, cluster.ResourceGoup, cluster.Name)
	//	if err != nil {
	//		break
	//	}
	//
	//	if akscluster.ProvisioningState == to.StringPtr("Deleting") {
	//		ctx.SendLogs(
	//			"AKS cluster deletion for '"+cluster.Name+"' is in progress ",
	//			models.LOGGING_LEVEL_ERROR,
	//			models.Backend_Logging,
	//		)
	//	} else if akscluster.ProvisioningState == to.StringPtr("Deleted") {
	//		break
	//	} else {
	//		ctx.SendLogs(
	//			"AKS cluster deletion for '"+cluster.Name+"' failed: ",
	//			models.LOGGING_LEVEL_ERROR,
	//			models.Backend_Logging,
	//		)
	//
	//		return errors.New("AKS cluster deletion for '" + cluster.Name + "' failed: ")
	//	}
	//}
	err = future.WaitForCompletionRef(cloud.Context, cloud.MCClient.Client)
	if err != nil {
		ctx.SendLogs(
			"AKS cluster deletion for '"+cluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return ApiError(err, "", 502)
	}

	return types.CustomCPError{}
}

func (cloud *AKS) GetKubeConfig(ctx utils.Context, cluster AKSCluster) (*containerservice.CredentialResult, types.CustomCPError) {
	if cloud == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
		return nil, ApiError(err, "Error while getting kube config", 502)
	}

	for _, kubeconfig := range *results.Kubeconfigs {
		return &kubeconfig, types.CustomCPError{}
	}

	return nil, types.CustomCPError{}
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
		ProjectId:         cloud.ProjectId,
		Cloud:             models.AKS,
		ProvisioningState: *v.ProvisioningState,
		KubernetesVersion: *v.KubernetesVersion,
		AgentPoolProfiles: agentPoolArr,
		EnableRBAC:        *v.EnableRBAC,
		ResourceID:        *v.ID,
		Name:              *v.Name,
		Type:              *v.Type,
		Location:          *v.Location,
	}
}

func generateClusterNodePools(c AKSCluster) *[]containerservice.ManagedClusterAgentPoolProfile {
	AKSNodePools := make([]containerservice.ManagedClusterAgentPoolProfile, len(c.AgentPoolProfiles))
	if c.IsAdvanced {
		for i, nodepool := range c.AgentPoolProfiles {
			AKSNodePools[i].Name = nodepool.Name
			AKSNodePools[i].Count = nodepool.Count
			AKSNodePools[i].OsType = "Linux"
			AKSNodePools[i].VMSize = containerservice.VMSizeTypes(*nodepool.VMSize)
			AKSNodePools[i].OsDiskSizeGB = nodepool.OsDiskSizeGB
			AKSNodePools[i].MaxPods = nodepool.MaxPods
			AKSNodePools[i].VnetSubnetID = nodepool.VnetSubnetID
			AKSNodePools[i].Type = "VirtualMachineScaleSets"

			nodelabels := make(map[string]*string)
			for _, label := range nodepool.NodeLabels {
				nodelabels[label.Key] = &label.Value
			}
			AKSNodePools[i].NodeLabels = nodelabels

			var nodeTaints []string
			for key, value := range nodepool.NodeTaints {
				nodeTaints = append(nodeTaints, key+"="+*value)
			}
			AKSNodePools[i].NodeTaints = &nodeTaints

			if nodepool.EnableAutoScaling != nil && *nodepool.EnableAutoScaling {
				AKSNodePools[i].EnableAutoScaling = nodepool.EnableAutoScaling
				AKSNodePools[i].MinCount = nodepool.MinCount
				AKSNodePools[i].MaxCount = nodepool.MaxCount
			}

		}
	} else {
		for i, nodepool := range c.AgentPoolProfiles {
			if nodepool.Name == nil {
				AKSNodePools[i].Name = to.StringPtr("pool0")
			} else {
				AKSNodePools[i].Name = nodepool.Name
			}
			AKSNodePools[i].Count = nodepool.Count
			AKSNodePools[i].OsType = "Linux"
			AKSNodePools[i].VMSize = containerservice.VMSizeTypes(*nodepool.VMSize)
			AKSNodePools[i].VnetSubnetID = nodepool.VnetSubnetID
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

	if c.IsAdvanced {
		if c.APIServerAccessProfile.EnablePrivateCluster {
			AKSapiServerAccessProfile.EnablePrivateCluster = to.BoolPtr(true)
		} else {
			AKSapiServerAccessProfile.EnablePrivateCluster = to.BoolPtr(false)
		}

		var authIpRanges []string
		for _, val := range c.APIServerAccessProfile.AuthorizedIPRanges {
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
	if c.IsAdvanced && c.IsServicePrincipal {
		AKSservicePrincipal.ClientID = &c.ClientID
		AKSservicePrincipal.Secret = &c.Secret
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
			EnableRBAC:              &c.EnableRBAC,
			AddonProfiles:           generateAddonProfiles(c),
			NetworkProfile:          generateNetworkProfile(c),

			//WindowsProfile:          generateWindowsProfile(),
		},
		//Identity: generateClusterIdentity(),
		Tags: generateClusterTags(c),
	}
	return &request
}

func generateWindowsProfile() *containerservice.ManagedClusterWindowsProfile {
	var AKSwindowProfile containerservice.ManagedClusterWindowsProfile
	AKSwindowProfile.AdminPassword = to.StringPtr("cloudplex")
	AKSwindowProfile.AdminUsername = to.StringPtr("cloudplex")
	return &AKSwindowProfile
}

func (cloud *AKS) GetKubernetesVersions(ctx utils.Context) (*containerservice.OrchestratorVersionProfileListResult, types.CustomCPError) {
	if cloud == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return nil, err
		}
	}

	cloud.Context = context.Background()
	result, err := cloud.KubeVersionClient.ListOrchestrators(cloud.Context, cloud.Region, "Microsoft.ContainerService")
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, ApiError(err, "Error while getting kubernetes version", 502)
	}

	return &result, types.CustomCPError{}
}

func generateNetworkProfile(c AKSCluster) *containerservice.NetworkProfileType {

	var AKSnetworkProfile containerservice.NetworkProfileType
	if c.IsExpert {
		AKSnetworkProfile.PodCidr = &c.PodCidr
		AKSnetworkProfile.DNSServiceIP = &c.DNSServiceIP
		AKSnetworkProfile.ServiceCidr = &c.ServiceCidr
		AKSnetworkProfile.DockerBridgeCidr = &c.DockerBridgeCidr
	}

	if c.IsAdvanced && len(c.APIServerAccessProfile.AuthorizedIPRanges) > 0 { //standard load balancer is used if AuthorizedIpRanges defined
		AKSnetworkProfile.LoadBalancerSku = "standard"
	}
	return &AKSnetworkProfile

}

func generateClusterTags(c AKSCluster) map[string]*string {
	AKSclusterTags := make(map[string]*string)
	if c.IsAdvanced {
		for _, tag := range c.ClusterTags {
			AKSclusterTags[tag.Key] = &tag.Value
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
	if c.IsAdvanced && c.IsHttpRouting {
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
	if c.IsAdvanced {
		return to.StringPtr(c.KubernetesVersion)
	} else {
		return to.StringPtr("1.15.10")
	}
}

func generateDnsPrefix(c AKSCluster) *string {
	if c.IsAdvanced {
		return to.StringPtr(c.DNSPrefix + "-dns")
	} else {
		return to.StringPtr(c.Name + "-dns")
	}
}

func (cloud *AKS) fetchClusterStatus(cluster1 *AKSCluster, ctx utils.Context) (cluster KubeClusterStatus,error types.CustomCPError) {
	if cloud == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return KubeClusterStatus{},err
		}
	}

	cloud.Context = context.Background()
	AKScluster, err := cloud.MCClient.Get(cloud.Context, cluster1.ResourceGoup, cluster1.Name)
	if err != nil {
		ctx.SendLogs(
			"AKS get cluster within resource group '"+cluster.ResourceGoup+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return KubeClusterStatus{},ApiError(err, "Error in fetching statud", 512)
	}

	for _, agentPool := range *AKScluster.AgentPoolProfiles {
		agentPoolProfiles := ManagedClusterAgentPoolStatus{}
		agentPoolProfiles.Name = agentPool.Name
		vm := string(agentPool.VMSize)
		agentPoolProfiles.VMSize = &vm
		agentPoolProfiles.Count = agentPool.Count
		if *agentPool.EnableAutoScaling{
			agentPoolProfiles.EnableAutoScaling =agentPool.EnableAutoScaling
			agentPoolProfiles.MaxCount=agentPool.MaxCount
			agentPoolProfiles.MinCount=agentPool.MinCount
		}else {
			scaling:=false
			agentPoolProfiles.EnableAutoScaling=&scaling
	}
		cluster.AgentPoolProfiles = append(cluster.AgentPoolProfiles,agentPoolProfiles)
	}
	cluster.Name=*AKScluster.Name
	cluster.Status=cluster1.Status

	cluster.Region=*AKScluster.Location
	cluster.ResourceGoup=*AKScluster.NodeResourceGroup

	cluster.NodePoolCount=*AKScluster.MaxAgentPools
	cluster.ProvisioningState = *AKScluster.ProvisioningState
	cluster.KubernetesVersion = *AKScluster.KubernetesVersion


	return cluster,types.CustomCPError{}
}

func GetAKSSupportedVms(ctx utils.Context) []containerservice.VMSizeTypes {
	return containerservice.PossibleVMSizeTypesValues()
}

func GetVmSkus(ctx utils.Context) ([]SkuResp, error) {
	bytes, err := ioutil.ReadFile("/app/files/azure-list-skus.txt")
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []SkuResp{}, err
	}

	var skus []SkuResp
	err = json.Unmarshal(bytes, &skus)
	if err != nil {
		return []SkuResp{}, err
	}

	return skus, nil
}

func (cloud *AKS) WriteAzureSkus() {
	if cloud == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			fmt.Println(err.Description)
			return
		}
	}

	cloud.Context = context.Background()
	pages, err := cloud.ResourceSkuClient.List(cloud.Context)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var ResourceList []skus.ResourceSku
	for pages.NotDone() {
		for _, value := range pages.Values() {
			ResourceList = append(ResourceList, value)
		}
		_ = pages.Next()
	}

	err = os.Remove("/app/files/azure-list-skus.txt")
	if err != nil {
		fmt.Println(err.Error())
	}
	bytes, _ := json.Marshal(ResourceList)
	f, err := os.Create("/app/files/azure-list-skus.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer f.Close()

	_, err = f.Write(bytes)
	if err != nil {
		fmt.Println(err)
		return
	}
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
