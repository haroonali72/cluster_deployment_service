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
			"AKS cluster validation for '"+aksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	request := cloud.generateClusterCreateRequest(aksCluster)
	cloud.ProjectId = aksCluster.ProjectId
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
		pool.Name = *aksAgentPool.Name
		pool.VnetSubnetID = *aksAgentPool.VnetSubnetID
		pool.Count = *aksAgentPool.Count
		pool.MaxPods = *aksAgentPool.MaxPods
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
		//Tags:       tags,
	}
}

func generateClusterNodePools(c AKSCluster) *[]containerservice.ManagedClusterAgentPoolProfile {
	var AKSNodePools []containerservice.ManagedClusterAgentPoolProfile
	if c.ClusterProperties.IsAdvanced {
		for _, nodepool := range c.ClusterProperties.AgentPoolProfiles {
			var AKSnodepool containerservice.ManagedClusterAgentPoolProfile
			AKSnodepool.Name = &nodepool.Name
			AKSnodepool.Count = &nodepool.Count
			AKSnodepool.OsType = nodepool.OsType
			AKSnodepool.VMSize = nodepool.VMSize
			AKSnodepool.OsDiskSizeGB = &nodepool.OsDiskSizeGB
			AKSnodepool.MaxPods = &nodepool.MaxPods

			nodelabels := make(map[string]*string)
			for key, value := range nodepool.NodeLabels {
				nodelabels[key] = &value
			}
			AKSnodepool.NodeLabels = nodelabels

			var nodeTaints []string
			for key, value := range nodepool.NodeTaints {
				nodeTaints = append(nodeTaints, key+"="+value)
			}
			AKSnodepool.NodeTaints = &nodeTaints

			if nodepool.EnableAutoScaling {
				AKSnodepool.EnableAutoScaling = &nodepool.EnableAutoScaling
				AKSnodepool.MinCount = &nodepool.MinCount
				AKSnodepool.MaxCount = &nodepool.MaxCount
				AKSnodepool.Type = "VirtualMachineScaleSets"
			}
			AKSNodePools = append(AKSNodePools, AKSnodepool)
		}
	} else {
		for _, nodepool := range c.ClusterProperties.AgentPoolProfiles {
			var AKSnodepool containerservice.ManagedClusterAgentPoolProfile
			AKSnodepool.Name = &nodepool.Name
			AKSnodepool.Count = &nodepool.Count
			AKSnodepool.OsType = "Linux"
			AKSnodepool.VMSize = nodepool.VMSize

			nodelabels := make(map[string]*string)
			nodelabels["AKS-Custer"] = to.StringPtr(c.ProjectId)
			AKSnodepool.NodeLabels = nodelabels

			AKSNodePools = append(AKSNodePools, AKSnodepool)
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
		//AKSapiServerAccessProfile.AuthorizedIPRanges =
	}

	return &AKSapiServerAccessProfile
}

func (cloud *AKS) generateServicePrincipal() *containerservice.ManagedClusterServicePrincipalProfile {
	var AKSservicePrincipal containerservice.ManagedClusterServicePrincipalProfile
	AKSservicePrincipal.ClientID = &cloud.ID
	AKSservicePrincipal.Secret = &cloud.Key
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
			ServicePrincipalProfile: cloud.generateServicePrincipal(),
			//APIServerAccessProfile:  cloud.generateApiServerAccessProfile(c),
			EnableRBAC: &c.ClusterProperties.EnableRBAC,
		},
	}
	return &request
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
		cluster.ClusterProperties.AgentPoolProfiles[index].Name = *agentPool.Name
		cluster.ClusterProperties.AgentPoolProfiles[index].OsDiskSizeGB = *agentPool.OsDiskSizeGB
		cluster.ClusterProperties.AgentPoolProfiles[index].VnetSubnetID = *agentPool.VnetSubnetID
		cluster.ClusterProperties.AgentPoolProfiles[index].VMSize = agentPool.VMSize
		cluster.ClusterProperties.AgentPoolProfiles[index].OsType = agentPool.OsType
		cluster.ClusterProperties.AgentPoolProfiles[index].Count = *agentPool.Count
	}
	cluster.ResourceID = *AKScluster.ID
	cluster.Type = *AKScluster.Type
	cluster.ClusterProperties.ProvisioningState = *AKScluster.ProvisioningState
	cluster.ClusterProperties.KubernetesVersion = *AKScluster.KubernetesVersion
	cluster.ClusterProperties.DNSPrefix = *AKScluster.DNSPrefix
	cluster.ClusterProperties.Fqdn = *AKScluster.Fqdn
	cluster.ClusterProperties.EnableRBAC = *AKScluster.EnableRBAC

	var networkProfile NetworkProfileType
	networkProfile.DNSServiceIP = *AKScluster.ManagedClusterProperties.NetworkProfile.DNSServiceIP
	networkProfile.PodCidr = *AKScluster.ManagedClusterProperties.NetworkProfile.PodCidr
	networkProfile.ServiceCidr = *AKScluster.ManagedClusterProperties.NetworkProfile.ServiceCidr
	networkProfile.DockerBridgeCidr = *AKScluster.ManagedClusterProperties.NetworkProfile.DockerBridgeCidr

	cluster.ClusterProperties.NetworkProfile = networkProfile
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
