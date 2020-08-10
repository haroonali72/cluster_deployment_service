package azure

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/key_utils"
	"antelope/models/types"
	userData2 "antelope/models/userData"
	"antelope/models/utils"
	"antelope/models/vault"
	"context"
	"crypto/rand"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	c "github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-02-01/network"
	"github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2018-01-01-preview/authorization"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2016-06-01/subscriptions"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2017-10-01/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/astaxie/beego"
	"os"
	"os/exec"
	"strings"
	"time"
)

type CreatedPool struct {
	Instances []*compute.VirtualMachine
	KeyName   string
	Key       string
	PoolName  string
}

type azureVM struct {
	ID       *string            `json:"id,omitempty"`
	Name     *string            `json:"name,omitempty"`
	Type     *string            `json:"type,omitempty"`
	Location *string            `json:"location,omitempty"`
	Tags     map[string]*string `json:"tags"`
}

type AZURE struct {
	Authorizer       *autorest.BearerAuthorizer
	AddressClient    network.PublicIPAddressesClient
	InterfacesClient network.InterfacesClient
	Location         subscriptions.Client
	VMSSCLient       compute.VirtualMachineScaleSetsClient
	VMSSVMClient     compute.VirtualMachineScaleSetVMsClient
	VMClient         compute.VirtualMachinesClient
	DiskClient       compute.DisksClient
	AccountClient    storage.AccountsClient
	context          context.Context

	ID             string
	Key            string
	Tenant         string
	Subscription   string
	Region         string
	Resources      map[string]interface{}
	RoleAssignment authorization.RoleAssignmentsClient
	RoleDefinition authorization.RoleDefinitionsClient
}

func (cloud *AZURE) init() types.CustomCPError {
	if cloud.Authorizer != nil {
		return types.CustomCPError{}
	}

	if cloud.ID == "" || cloud.Key == "" || cloud.Tenant == "" || cloud.Subscription == "" || cloud.Region == "" {
		text := "invalid cloud credentials"
		beego.Error(text)
		return ApiError(errors.New(text), "Error in initialising cloud", int(models.CloudStatusCode))
	}

	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, cloud.Tenant)
	if err != nil {
		panic(err)
	}

	spt, err := adal.NewServicePrincipalToken(*oauthConfig, cloud.ID, cloud.Key, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return ApiError(err, "Error in initilising cloud", int(models.CloudStatusCode))
	}

	cloud.context = context.Background()
	cloud.Authorizer = autorest.NewBearerAuthorizer(spt)

	cloud.AddressClient = network.NewPublicIPAddressesClient(cloud.Subscription)
	cloud.AddressClient.Authorizer = cloud.Authorizer

	cloud.InterfacesClient = network.NewInterfacesClient(cloud.Subscription)
	cloud.InterfacesClient.Authorizer = cloud.Authorizer

	cloud.AccountClient = storage.NewAccountsClient(cloud.Subscription)
	cloud.AccountClient.Authorizer = cloud.Authorizer

	cloud.VMClient = compute.NewVirtualMachinesClient(cloud.Subscription)
	cloud.VMClient.Authorizer = cloud.Authorizer

	cloud.VMSSVMClient = compute.NewVirtualMachineScaleSetVMsClient(cloud.Subscription)
	cloud.VMSSVMClient.Authorizer = cloud.Authorizer

	cloud.VMSSCLient = compute.NewVirtualMachineScaleSetsClient(cloud.Subscription)
	cloud.VMSSCLient.Authorizer = cloud.Authorizer

	cloud.DiskClient = compute.NewDisksClient(cloud.Subscription)
	cloud.DiskClient.Authorizer = cloud.Authorizer

	cloud.RoleAssignment = authorization.NewRoleAssignmentsClient(cloud.Subscription)
	cloud.RoleAssignment.Authorizer = cloud.Authorizer

	cloud.RoleDefinition = authorization.NewRoleDefinitionsClient(cloud.Subscription)
	cloud.RoleDefinition.Authorizer = cloud.Authorizer

	cloud.Resources = make(map[string]interface{})
	cloud.Location = subscriptions.NewClient()
	cloud.Location.Authorizer = cloud.Authorizer
	return types.CustomCPError{}
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
func (cloud *AZURE) createCluster(cluster Cluster_Def, ctx utils.Context, companyId string, token string) (Cluster_Def, types.CustomCPError) {

	if cloud == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			return cluster, err
		}
	}

	var azureNetwork types.AzureNetwork
	url := getNetworkHost("azure", cluster.ProjectId)
	network, err1 := api_handler.GetAPIStatus(token, url, ctx)
	if err1 != nil {
		beego.Error(err1.Error())
		return cluster, ApiError(err1, "Error in cluster creation", int(models.CloudStatusCode))
	}

	err1 = json.Unmarshal(network.([]byte), &azureNetwork)
	if err1 != nil {
		ctx.SendLogs(err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, ApiError(err1, "Error in cluster creation", int(models.CloudStatusCode))
	}

	for i, pool := range cluster.NodePools {
		ctx.SendLogs("AZUREOperations creating nodes", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		result, private_key, err := cloud.CreateInstance(pool, azureNetwork, cluster.ResourceGroup, cluster.ProjectId, i, ctx, companyId, token)
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return cluster, err
		}
		beego.Info(private_key)
		//if pool.EnableVolume {
		//	err = cloud.mountVolume(result, private_key, pool.KeyInfo.KeyName, cluster.ProjectId, pool.AdminUser, cluster.ResourceGroup, pool.Name, ctx, string(pool.PoolRole), false)
		//	if err != (types.CustomCPError{}) {
		//		utils.SendLog(companyId, "Error in volume mounting : ", "info", cluster.ProjectId)
		//		return cluster, err
		//	}
		//}
		//if pool.PoolRole == "master" {
		//	err = cloud.mountVolume(result, private_key, pool.KeyInfo.KeyName, cluster.ProjectId, pool.AdminUser, cluster.ResourceGroup, pool.Name, ctx, string(pool.PoolRole), true)
		//	if err != (types.CustomCPError{}) {
		//		utils.SendLog(companyId, "Error in volume mounting on master pool: ", "info", cluster.ProjectId)
		//		return cluster, err
		//	}
		//}
		cluster.NodePools[i].Nodes = result
	}

	return cluster, types.CustomCPError{}
}

func (cloud *AZURE) AddRoles(ctx utils.Context, companyId string, resourceGroup string, projectId string, vmId *string, vmPrincipalId *string) types.CustomCPError {
	RolesID := []string{models.VM_CONTRIBUTOR_GUID, models.NETWORK_CONTRIBUTOR_GUID, models.STORAGE_CONTRIBUTOR_GUID, models.AVERE_CONTRIBUTER_GUID}
	BasePath := "/subscriptions/" + cloud.Subscription + "/providers/Microsoft.Authorization/roleDefinitions/"
	scope := "/subscriptions/" + cloud.Subscription + "/resourceGroups/" + resourceGroup
	RoleAssignmentParam := authorization.RoleAssignmentCreateParameters{}
	RoleAssignmentParam.RoleAssignmentProperties = &authorization.RoleAssignmentProperties{
		PrincipalID: vmPrincipalId,
	}
	utils.SendLog(companyId, "Attaching access roles to "+*vmId, "info", projectId)
	for _, id := range RolesID {
		RoleAssignmentParam.RoleAssignmentProperties.RoleDefinitionID = to.StringPtr(BasePath + id)
		bytes := make([]byte, 16)
		_, err := rand.Read(bytes)
		if err != nil {
			utils.SendLog(companyId, "Error creating UUID for role: "+err.Error(), "error", projectId)
			return ApiError(err, "Error in cluster creation", int(models.CloudStatusCode))
		}
		uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
			bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:])

		result, err := cloud.RoleAssignment.Create(context.Background(), scope, uuid, RoleAssignmentParam)
		if err != nil && !strings.Contains(err.Error(), "RoleAssignmentExists") {
			utils.SendLog(companyId, err.Error(), "error", projectId)
			return ApiError(err, "Error in cluster creation", int(models.CloudStatusCode))
		} else {
			x, _ := json.Marshal(result)
			utils.SendLog(companyId, "Role: "+string(x), "info", projectId)
		}
	}
	return types.CustomCPError{}
}
func (cloud *AZURE) CreateInstance(pool *NodePool, networkData types.AzureNetwork, resourceGroup string, projectId string, poolIndex int, ctx utils.Context, companyId string, token string) ([]*VM, string, types.CustomCPError) {

	var cpVms []*VM
	subnetId := cloud.GetSubnets(pool, networkData)
	sgIds := cloud.GetSecurityGroups(pool, networkData)
	zones := cloud.GetZones(pool,networkData)
	vpcName := networkData.Definition[0].Vnet.Name

	if pool.PoolRole == "master" {
		var publicIPaddress network.PublicIPAddress
		var err types.CustomCPError
		if pool.EnablePublicIP {
			IPname := "pip-" + pool.Name
			utils.SendLog(companyId, "Creating Public IP : "+projectId, "info", projectId)
			publicIPaddress, err = cloud.createPublicIp(pool, resourceGroup, IPname, ctx,zones[0])
			if err != (types.CustomCPError{}) {
				return nil, "", err
			}
			utils.SendLog(companyId, "Public IP created successfully : "+IPname, "info", projectId)
			cloud.Resources["Pip-"+projectId] = IPname
		}
		/*
			making network interface
		*/
		nicName := "NIC-" + pool.Name
		utils.SendLog(companyId, "Creating NIC : "+nicName, "info", projectId)
		nicParameters, err1 := cloud.createNIC(pool, resourceGroup, publicIPaddress, subnetId, sgIds, nicName, ctx)
		if err1 != (types.CustomCPError{}) {
			return nil, "", err1
		}
		utils.SendLog(companyId, "NIC created successfully : "+nicName, "info", projectId)
		cloud.Resources["Nic-"+projectId] = nicName

		utils.SendLog(companyId, "Creating node  : "+pool.Name, "info", projectId)
		vm, private_key, _, err1 := cloud.createVM(pool, poolIndex, nicParameters, resourceGroup, ctx, token, projectId, vpcName,zones)
		if err1 != (types.CustomCPError{}) {
			return nil, "", err1
		}
		utils.SendLog(companyId, "Node created successfully : "+pool.Name, "info", projectId)
		cloud.Resources["Disk-"+projectId] = pool.Name
		cloud.Resources["NodeName-"+projectId] = pool.Name

		var vmObj VM
		vmObj.Name = vm.Name
		vmObj.CloudId = vm.ID
		//vmObj.PrivateIP = (*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PrivateIPAddress
		//vmObj.PublicIP = publicIPaddress.PublicIPAddressPropertiesFormat.IPAddress
		//vmObj.NodeState = vm.VirtualMachineProperties.ProvisioningState
		vmObj.UserName = vm.VirtualMachineProperties.OsProfile.AdminUsername
		vmObj.PAssword = vm.VirtualMachineProperties.OsProfile.AdminPassword
		vmObj.ComputerName = vm.OsProfile.ComputerName
		vmObj.IdentityPrincipalId = vm.Identity.PrincipalID
		cpVms = append(cpVms, &vmObj)
		err1 = cloud.AddRoles(ctx, companyId, resourceGroup, projectId, vm.Name, vm.Identity.PrincipalID)
		if err1 != (types.CustomCPError{}) {
			return nil, "", err1
		}
		return cpVms, private_key, types.CustomCPError{}

	} else {
		vms, err, private_key := cloud.createVMSS(resourceGroup, projectId, pool, poolIndex, subnetId, sgIds, ctx, token, vpcName,zones)
		if err != (types.CustomCPError{}) {
			return nil, "", err
		}
		cloud.Resources["Vmss-"+pool.Name] = pool.Name
		for _, vm := range vms.Values() {
			var vmObj VM
			vmObj.Name = vm.Name
			vmObj.CloudId = vm.ID
			/*nicId := ""
			for _, nic := range *vm.NetworkProfile.NetworkInterfaces {
				nicId = *nic.ID
				break
			}
			arr := strings.Split(nicId, "/")
			nicName := arr[12]

			nicParameters, err := cloud.GetNIC(resourceGroup, pool.Name, arr[10], nicName, ctx)
			if err != nil {
				return nil, "", err
			}*/
			//vmObj.PrivateIP = (*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PrivateIPAddress
			/*pipId := *(*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PublicIPAddress.ID
			arr = strings.Split(pipId, "/")
			pipConf := arr[14]
			pipAddress := arr[16]
			_, err = cloud.GetPIP(resourceGroup, pool.Name, arr[10], nicName, pipConf, pipAddress, ctx)
			if err != nil {
				return nil, "", err
			}*/
			//vmObj.PublicIP = pip.IPAddress
			//vmObj.NodeState = vm.ProvisioningState
			vmObj.UserName = vm.OsProfile.AdminUsername
			vmObj.PAssword = vm.OsProfile.AdminPassword
			vmObj.ComputerName = vm.OsProfile.ComputerName
			cpVms = append(cpVms, &vmObj)

		}
		vmScaleSet, err1 := cloud.VMSSCLient.Get(context.Background(), resourceGroup, pool.Name)
		if err1 != nil {
			return nil, "", ApiError(err1, "Error in cluster creation", int(models.CloudStatusCode))
		}

		err = cloud.AddRoles(ctx, companyId, resourceGroup, projectId, vmScaleSet.Name, vmScaleSet.Identity.PrincipalID)
		if err != (types.CustomCPError{}) {
			return nil, "", err
		}
		return cpVms, private_key, types.CustomCPError{}
	}

}
func (cloud *AZURE) GetSecurityGroups(pool *NodePool, network types.AzureNetwork) []*string {
	var sgId []*string
	for _, definition := range network.Definition {
		for _, sg := range definition.SecurityGroups {
			for _, sgName := range pool.PoolSecurityGroups {
				if *sgName == sg.Name {
					sgId = append(sgId, &sg.SecurityGroupId)
				}
			}
		}
	}
	return sgId
}
func (cloud *AZURE) GetSubnets(pool *NodePool, network types.AzureNetwork) string {
	for _, definition := range network.Definition {
		for _, subnet := range definition.Subnets {
			if subnet.Name == pool.PoolSubnet {
				return subnet.SubnetId
			}
		}
	}
	return ""
}
func (cloud *AZURE) GetZones(pool *NodePool, network types.AzureNetwork) []string {
	for _, definition := range network.Definition {
		for _, subnet := range definition.Subnets {
			if subnet.Name == pool.PoolSubnet {
				return subnet.Zone
			}
		}
	}
	return []string{}
}
func (cloud *AZURE) fetchStatus(cluster *Cluster_Def, token string, ctx utils.Context) (*Cluster_Def, types.CustomCPError) {
	if cloud.Authorizer == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs("Cluster model: Status - Failed to get lastest status "+err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return &Cluster_Def{}, err
		}
	}
	for in, pool := range cluster.NodePools {
		var keyInfo key_utils.AZUREKey

		if pool.KeyInfo.CredentialType == models.SSHKey {
			bytes, err := vault.GetSSHKey(string(models.Azure), pool.KeyInfo.KeyName, token, ctx, "")
			if err != nil {
				ctx.SendLogs("vm creation failed with error: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return &Cluster_Def{}, ApiError(err, "Error in fetching status", int(models.CloudStatusCode))
			}
			keyInfo, err = key_utils.AzureKeyConversion(bytes, ctx)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return &Cluster_Def{}, ApiError(err, "Error in fetching status", int(models.CloudStatusCode))
			}

		}
		pool.KeyInfo = keyInfo
		if pool.PoolRole == "master" {
			var vmObj VM
			beego.Info("getting instance")
			vm, err := cloud.GetInstance(pool.Name, cluster.ResourceGroup, ctx)
			if err != (types.CustomCPError{}) {
				ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return &Cluster_Def{}, err
			}
			beego.Info("getting nic")
			nicName := "NIC-" + pool.Name
			nicParameters, err := cloud.GetVMNIC(cluster.ResourceGroup, nicName, ctx)
			if err != (types.CustomCPError{}) {
				ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return &Cluster_Def{}, err
			}
			beego.Info("getting pip")
			IPname := "pip-" + pool.Name
			if pool.EnablePublicIP {
				publicIPaddress, err := cloud.GetVMPIP(cluster.ResourceGroup, IPname, ctx)
				if err != (types.CustomCPError{}) {
					ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
					return &Cluster_Def{}, err
				}
				vmObj.PublicIP = publicIPaddress.PublicIPAddressPropertiesFormat.IPAddress
			}

			vmObj.Name = vm.Name
			vmObj.CloudId = vm.ID
			vmObj.PrivateIP = (*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PrivateIPAddress

			vmObj.NodeState = vm.ProvisioningState
			vmObj.UserName = vm.OsProfile.AdminUsername
			vmObj.PAssword = vm.OsProfile.AdminPassword
			vmObj.ComputerName = vm.OsProfile.ComputerName
			//cpVms = append(cpVms, &vmObj)
			beego.Info("updated node pool")
			cluster.NodePools[in].Nodes = []*VM{&vmObj}

		} else {
			var cpVms []*VM
			vms, err := cloud.VMSSVMClient.List(cloud.context, cluster.ResourceGroup, pool.Name, "", "", "")
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return &Cluster_Def{}, ApiError(err, "Error in fetching status", int(models.CloudStatusCode))
			}
			for _, vm := range vms.Values() {
				var vmObj VM
				vmObj.Name = vm.Name
				vmObj.CloudId = vm.ID
				nicId := ""
				for _, nic := range *vm.NetworkProfile.NetworkInterfaces {
					nicId = *nic.ID
					break
				}
				arr := strings.Split(nicId, "/")
				nicName := arr[12]
				nicParameters, err := cloud.GetNIC(cluster.ResourceGroup, pool.Name, arr[10], nicName, ctx)
				if err != (types.CustomCPError{}) {
					return &Cluster_Def{}, err
				}
				vmObj.PrivateIP = (*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PrivateIPAddress
				if pool.EnablePublicIP {
					pipId := *(*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PublicIPAddress.ID
					arr = strings.Split(pipId, "/")
					pipConf := arr[14]
					pipAddress := arr[16]
					pip, err := cloud.GetPIP(cluster.ResourceGroup, pool.Name, arr[10], nicName, pipConf, pipAddress, ctx)
					if err != (types.CustomCPError{}) {
						return &Cluster_Def{}, err
					}
					vmObj.PublicIP = pip.IPAddress
				}
				vmObj.NodeState = vm.ProvisioningState
				vmObj.UserName = vm.OsProfile.AdminUsername
				vmObj.PAssword = vm.OsProfile.AdminPassword
				vmObj.ComputerName = vm.OsProfile.ComputerName
				cpVms = append(cpVms, &vmObj)

			}

			beego.Info("updated node pool")
			cluster.NodePools[in].Nodes = cpVms
		}
	}
	beego.Info("updated cluster")
	return cluster, types.CustomCPError{}
}

func (cloud *AZURE) GetInstance(name string, resourceGroup string, ctx utils.Context) (compute.VirtualMachine, types.CustomCPError) {

	vm, err := cloud.VMClient.Get(cloud.context, resourceGroup, name, compute.InstanceView)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return compute.VirtualMachine{}, ApiError(err, "Error in fetching instance", int(models.CloudStatusCode))
	}
	return vm, types.CustomCPError{}
}

func (cloud *AZURE) GetNIC(resourceGroup, vmss, vm, nicName string, ctx utils.Context) (network.Interface, types.CustomCPError) {

	nicParameters, err := cloud.InterfacesClient.GetVirtualMachineScaleSetNetworkInterface(cloud.context, resourceGroup, vmss, vm, nicName, "")
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return network.Interface{}, ApiError(err, "Error in fetching status", int(models.CloudStatusCode))
	}
	return nicParameters, types.CustomCPError{}
}
func (cloud *AZURE) GetPIP(resourceGroup, vmss, vm, nic, ipConfig, ipAddress string, ctx utils.Context) (network.PublicIPAddress, types.CustomCPError) {

	publicIPaddress, err := cloud.AddressClient.GetVirtualMachineScaleSetPublicIPAddress(cloud.context, resourceGroup, vmss, vm, nic, ipConfig, ipAddress, "")
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return network.PublicIPAddress{}, ApiError(err, "Errorin VMSS creation", int(models.CloudStatusCode))
	}
	return publicIPaddress, types.CustomCPError{}
}
func (cloud *AZURE) GetVMNIC(resourceGroup, nicName string, ctx utils.Context) (network.Interface, types.CustomCPError) {
	nicParameters, err := cloud.InterfacesClient.Get(cloud.context, resourceGroup, nicName, "")
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return network.Interface{}, ApiError(err, "Error in cluster creation", int(models.CloudStatusCode))
	}
	return nicParameters, types.CustomCPError{}
}
func (cloud *AZURE) GetVMPIP(resourceGroup, IPname string, ctx utils.Context) (network.PublicIPAddress, types.CustomCPError) {
	publicIPaddress, err := cloud.AddressClient.Get(cloud.context, resourceGroup, IPname, "")
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return network.PublicIPAddress{}, ApiError(err, "Error in fetching VM public IP ", int(models.CloudStatusCode))
	}
	return publicIPaddress, types.CustomCPError{}
}
func (cloud *AZURE) terminateCluster(cluster Cluster_Def, ctx utils.Context, companyId string) types.CustomCPError {

	terminate := true

	if cloud.Authorizer == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	utils.SendLog(companyId, "Terminating Cluster : "+cluster.Name, models.LOGGING_LEVEL_INFO, cluster.ProjectId)

	for _, pool := range cluster.NodePools {

		utils.SendLog(companyId, "Terminating node pool: "+pool.Name, models.LOGGING_LEVEL_INFO, cluster.ProjectId)
		if pool.PoolRole == "master" {

			utils.SendLog(companyId, "Terminating node pool: "+pool.Name, models.LOGGING_LEVEL_INFO, cluster.ProjectId)
			if pool != nil && pool.Nodes != nil && pool.Nodes[0].Name != nil {
				err := cloud.TerminateMasterNode(*pool.Nodes[0].Name, cluster.ProjectId, cluster.ResourceGroup, ctx, companyId)
				if err != (types.CustomCPError{}) {
					terminate = false
					break
				}
			} else {
				break
			}
			nicName := "NIC-" + pool.Name

			err := cloud.deleteNIC(nicName, cluster.ResourceGroup, cluster.ProjectId, ctx, companyId)
			if err != (types.CustomCPError{}) {
				terminate = false
			}

			IPname := "pip-" + pool.Name
			err = cloud.deletePublicIp(IPname, cluster.ResourceGroup, cluster.ProjectId, ctx, companyId)
			if err != (types.CustomCPError{}) {
				terminate = false
			}

			sName := strings.Replace(pool.Name, "-", "", -1)
			sName = strings.ToLower(sName)
			cloud.Resources["SA-"+pool.Name] = sName
			err = cloud.deleteStorageAccount(cluster.ResourceGroup, cloud.Resources["SA-"+pool.Name].(string), ctx)
			if err != (types.CustomCPError{}) {
				terminate = false
			}

			beego.Info("terminating master pool disk: " + pool.Name)

			err = cloud.deleteDisk(cluster.ResourceGroup, pool.Name, ctx)
			if err != (types.CustomCPError{}) {
				terminate = false
			}

			if pool.EnableVolume {
				err = cloud.deleteDisk(cluster.ResourceGroup, "ext-"+pool.Name, ctx)
				if err != (types.CustomCPError{}) {
					terminate = false
				}
			}
			//deleting master volume
			err = cloud.deleteDisk(cluster.ResourceGroup, "ext-master-"+pool.Name, ctx)
			if err != (types.CustomCPError{}) {
				terminate = false
			}

		} else {
			err := cloud.TerminatePool(pool.Name, cluster.ResourceGroup, cluster.ProjectId, ctx)
			if err != (types.CustomCPError{}) {
				terminate = false
				break
			}

			sName := strings.Replace(pool.Name, "-", "", -1)
			sName = strings.ToLower(sName)
			cloud.Resources["SA-"+pool.Name] = sName
			err = cloud.deleteStorageAccount(cluster.ResourceGroup, cloud.Resources["SA-"+pool.Name].(string), ctx)
			if err != (types.CustomCPError{}) {
				terminate = false
			}

		}
		utils.SendLog(companyId, "Node Pool terminated successfully: "+pool.Name, models.LOGGING_LEVEL_INFO, cluster.ProjectId)
	}
	if terminate == false {
		return ApiError(errors.New("Termination failed"), "Error in termination", int(models.CloudStatusCode))
	}

	return types.CustomCPError{}
}

func (cloud *AZURE) TerminatePool(name string, resourceGroup string, projectId string, ctx utils.Context) types.CustomCPError {

	ctx.SendLogs("AZUREOperations: terminating node pools", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	future, err := cloud.VMSSCLient.Delete(cloud.context, resourceGroup, name)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiError(err, "Error in pool termination", int(models.CloudStatusCode))
	} else if err != nil && strings.Contains(err.Error(), "not found") {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		return types.CustomCPError{}
	} else {
		err = future.WaitForCompletionRef(cloud.context, cloud.VMSSCLient.Client)
		if err != nil {
			ctx.SendLogs("vm deletion failed"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return ApiError(err, "Error in pool termination", int(models.CloudStatusCode))
		}
	}

	ctx.SendLogs("Node pool terminated successfully: "+name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	return types.CustomCPError{}
}
func (cloud *AZURE) TerminateMasterNode(name, projectId, resourceGroup string, ctx utils.Context, companyId string) types.CustomCPError {

	beego.Info("AZUREOperations: terminating nodes")

	ctx.SendLogs("Terminating node: "+name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	utils.SendLog(companyId, "Terminating node: "+name, "info", projectId)

	vmClient := compute.NewVirtualMachinesClient(cloud.Subscription)
	vmClient.Authorizer = cloud.Authorizer
	future, err := vmClient.Delete(cloud.context, resourceGroup, name)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return ApiError(err, "Error in master node termination", int(models.CloudStatusCode))
	} else if err != nil && strings.Contains(err.Error(), "not found") {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		utils.SendLog(companyId, err.Error(), "error", projectId)
		return types.CustomCPError{}
	} else {
		err = future.WaitForCompletionRef(cloud.context, vmClient.Client)
		if err != nil {
			utils.SendLog(companyId, err.Error(), "error", projectId)
			return ApiError(err, "Error in master node termination", int(models.CloudStatusCode))
		}
		ctx.SendLogs("Deleted Node"+name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	}
	ctx.SendLogs("Node terminated successfully: "+name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	return types.CustomCPError{}
}

func (cloud *AZURE) createPublicIp(pool *NodePool, resourceGroup string, IPname string, ctx utils.Context,zone string) (network.PublicIPAddress, types.CustomCPError) {

	pipParameters := network.PublicIPAddress{
		Location: &cloud.Region,
		Sku:&network.PublicIPAddressSku{Name:"standard"},
		Zones : &[]string{zone},
		PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: network.IPAllocationMethod("Static"),
			DNSSettings: &network.PublicIPAddressDNSSettings{
				DomainNameLabel: to.StringPtr(strings.ToLower(IPname)),
			},
		},
	}

	address, err := cloud.AddressClient.CreateOrUpdate(cloud.context, resourceGroup, IPname, pipParameters)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return network.PublicIPAddress{}, ApiError(err, "Error in cluster creation", int(models.CloudStatusCode))
	} else {
		err = address.WaitForCompletionRef(cloud.context, cloud.AddressClient.Client)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return network.PublicIPAddress{}, ApiError(err, "Error in cluster creation", int(models.CloudStatusCode))
		}
	}
	ctx.SendLogs("Get public IP address info...", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	publicIPaddress, err1 := cloud.GetVMPIP(resourceGroup, IPname, ctx)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs(err1.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return network.PublicIPAddress{}, err1
	}
	return publicIPaddress, types.CustomCPError{}
}

func (cloud *AZURE) deletePublicIp(IPname, resourceGroup string, projectId string, ctx utils.Context, companyId string) types.CustomCPError {

	utils.SendLog(companyId, "Deleting Public IP: "+IPname, "info", projectId)

	address, err := cloud.AddressClient.Delete(cloud.context, resourceGroup, IPname)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiError(err, "Error in deleting public IP", int(models.CloudStatusCode))
	} else if err != nil && strings.Contains(err.Error(), "not found") {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		return types.CustomCPError{}
	} else {
		err = address.WaitForCompletionRef(cloud.context, cloud.AddressClient.Client)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return ApiError(err, "Error in deleting public IP", int(models.CloudStatusCode))
		}
	}

	utils.SendLog(companyId, "Public IP deleted successfully: "+IPname, models.LOGGING_LEVEL_INFO, projectId)
	return types.CustomCPError{}
}
func (cloud *AZURE) createNIC(pool *NodePool, resourceGroup string, publicIPaddress network.PublicIPAddress, subnetId string, sgIds []*string, nicName string, ctx utils.Context) (network.Interface, types.CustomCPError) {

	nicParameters := network.Interface{
		Location: &cloud.Region,
		InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
			IPConfigurations: &[]network.InterfaceIPConfiguration{
				{
					Name: to.StringPtr(fmt.Sprintf("IPconfig-" + pool.Name)),
					InterfaceIPConfigurationPropertiesFormat: &network.InterfaceIPConfigurationPropertiesFormat{
						PrivateIPAllocationMethod: network.Dynamic,
						Subnet:                    &network.Subnet{ID: to.StringPtr(subnetId)},
					},
				},
			},
		},
	}
	if pool.EnablePublicIP {
		(*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].InterfaceIPConfigurationPropertiesFormat.PublicIPAddress = &publicIPaddress
	}
	if sgIds != nil {
		nicParameters.InterfacePropertiesFormat.NetworkSecurityGroup = &network.SecurityGroup{
			ID: sgIds[0],
		}
	}
	future, err := cloud.InterfacesClient.CreateOrUpdate(cloud.context, resourceGroup, nicName, nicParameters)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return network.Interface{}, ApiError(err, "Error in cluster creation", int(models.CloudStatusCode))
	} else {
		err := future.WaitForCompletionRef(cloud.context, cloud.InterfacesClient.Client)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return network.Interface{}, ApiError(err, "Error in cluster creation", int(models.CloudStatusCode))
		}
	}

	nicParameters, err1 := cloud.GetVMNIC(resourceGroup, nicName, ctx)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs(err1.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return network.Interface{}, err1
	}
	return nicParameters, types.CustomCPError{}
}
func (cloud *AZURE) deleteNIC(nicName, resourceGroup string, proId string, ctx utils.Context, companyId string) types.CustomCPError {

	utils.SendLog(companyId, "Deleting NIC: "+nicName, "info", proId)

	future, err := cloud.InterfacesClient.Delete(cloud.context, resourceGroup, nicName)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiError(err, "Error in deleting NIC", int(models.CloudStatusCode))
	} else if err != nil && strings.Contains(err.Error(), "not found") {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		return types.CustomCPError{}
	} else {

		err := future.WaitForCompletionRef(cloud.context, cloud.InterfacesClient.Client)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return ApiError(err, "Error in deleting NIC", int(models.CloudStatusCode))
		}
	}
	utils.SendLog(companyId, "NIC deleted successfully: "+nicName, models.LOGGING_LEVEL_INFO, proId)
	return types.CustomCPError{}
}

/*
func (cloud *AZURE) createVM(pool *NodePool, index int, nicParameters network.Interface, resourceGroup string, ctx utils.Context, token string) (compute.VirtualMachine, string, string, error) {
	var satype compute.StorageAccountTypes
	if pool.OsDisk == models.StandardSSD {
		satype = compute.StorageAccountTypesStandardSSDLRS
	} else if pool.OsDisk == models.PremiumSSD {
		satype = compute.StorageAccountTypesPremiumLRS
	} else if pool.OsDisk == models.StandardHDD {
		satype = compute.StorageAccountTypesStandardLRS

	}
	osDisk := &compute.OSDisk{
		CreateOption: compute.DiskCreateOptionTypesFromImage,
		Name:         to.StringPtr(pool.Name),
		ManagedDisk: &compute.ManagedDiskParameters{
			StorageAccountType: satype,
		},
	}
	if pool.Volume.DataDisk == models.StandardSSD {
		satype = compute.StorageAccountTypesStandardSSDLRS
	} else if pool.Volume.DataDisk == models.PremiumSSD {
		satype = compute.StorageAccountTypesPremiumLRS
	} else if pool.Volume.DataDisk == models.StandardHDD {
		satype = compute.StorageAccountTypesStandardLRS

	}

	storageName := "ext-" + pool.Name
	disk := compute.DataDisk{
		Lun:          to.Int32Ptr(int32(index)),
		Name:         to.StringPtr(storageName),
		CreateOption: compute.DiskCreateOptionTypesEmpty,
		DiskSizeGB:   to.Int32Ptr(pool.Volume.Size),
		ManagedDisk: &compute.ManagedDiskParameters{
			StorageAccountType: satype,
		},
	}

	var storage = []compute.DataDisk{}

	staticVolume := compute.DataDisk{
		Lun:          to.Int32Ptr(int32(index)),
		Name:         to.StringPtr("ext-master-" + pool.Name),
		CreateOption: compute.DiskCreateOptionTypesEmpty,
		DiskSizeGB:   to.Int32Ptr(5),
		ManagedDisk: &compute.ManagedDiskParameters{
			StorageAccountType: satype,
		},
	}
	cloud.Resources["ext-master-"+pool.Name] = "ext-master-" + pool.Name
	storage = append(storage, staticVolume)

	vm := compute.VirtualMachine{
		Name:     to.StringPtr(pool.Name),
		Location: to.StringPtr(cloud.Region),
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			HardwareProfile: &compute.HardwareProfile{
				VMSize: compute.VirtualMachineSizeTypes(pool.MachineType),
			},
			StorageProfile: &compute.StorageProfile{
				ImageReference: &compute.ImageReference{
					Offer:     to.StringPtr(pool.Image.Offer),
					Sku:       to.StringPtr(pool.Image.Sku),
					Publisher: to.StringPtr(pool.Image.Publisher),
					Version:   to.StringPtr(pool.Image.Version),
				},
				OsDisk: osDisk,
				//DataDisks: &storage,
			},
			OsProfile: &compute.OSProfile{
				ComputerName:  to.StringPtr(pool.Name),
				AdminUsername: to.StringPtr(pool.AdminUser),
			},
			NetworkProfile: &compute.NetworkProfile{

				NetworkInterfaces: &[]compute.NetworkInterfaceReference{
					{
						ID: &(*nicParameters.ID),
						NetworkInterfaceReferenceProperties: &compute.NetworkInterfaceReferenceProperties{
							Primary: to.BoolPtr(true),
						},
					},
				},
			},
		},
	}
	if pool.EnableVolume {
		storage = append(storage, disk)
		cloud.Resources["ext-"+pool.Name] = "ext-" + pool.Name
	}
	vm.StorageProfile.DataDisks = &storage
	privateKey := ""
	publicKey := ""
	if pool.KeyInfo.CredentialType == models.SSHKey  {
		_, err := vault.GetAzureSSHKey("azure", pool.KeyInfo.KeyName, token, ctx)

		if err != nil && err.Error() != "not found" {
			ctx.SendLogs("vm creation failed", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return compute.VirtualMachine{}, "", "", err

		} else if err == nil {
			key,err := key_utils.FetchAzureKey("keyName", "userName", token, ctx )
			if err != nil {
				return compute.VirtualMachine{}, "", "", err
			}
			if key.PublicKey != "" && key.PrivateKey != "" {
				keyy := []compute.SSHPublicKey{{
					Path:    to.StringPtr("/home/" + pool.AdminUser + "/.ssh/authorized_keys"),
					KeyData: to.StringPtr(key.PublicKey),
				},
				}
				linux := &compute.LinuxConfiguration{
					SSH: &compute.SSHConfiguration{
						PublicKeys: &keyy,
					},
				}
				vm.OsProfile.LinuxConfiguration = linux
				privateKey = key.PrivateKey
				publicKey = key.PublicKey
			}
		} else if err != nil && err.Error() == "not found" {
			privateKey,publicKey,err =key_utils.GenerateAzureKey("keyName", "userName", token, "teams", ctx )
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return compute.VirtualMachine{}, "", "", err
			}
			key := []compute.SSHPublicKey{{
				Path:    to.StringPtr("/home/" + pool.AdminUser + "/.ssh/authorized_keys"),
				KeyData: to.StringPtr(publicKey),
			},
			}
			linux := &compute.LinuxConfiguration{
				SSH: &compute.SSHConfiguration{
					PublicKeys: &key,
				},
			}
			vm.OsProfile.LinuxConfiguration = linux
			pool.KeyInfo.PublicKey = publicKey
			pool.KeyInfo.PrivateKey = privateKey


		}

	}

	if pool.BootDiagnostics.Enable {

		if pool.BootDiagnostics.NewStroageAccount {
			sName := strings.Replace(pool.Name, "-", "", -1)
			sName = strings.ToLower(sName)
			storageId := "https://" + sName + ".blob.core.windows.net/"
			err := cloud.createStorageAccount(resourceGroup, sName, ctx)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return compute.VirtualMachine{}, "", "", err
			}
			vm.VirtualMachineProperties.DiagnosticsProfile = &compute.DiagnosticsProfile{
				&compute.BootDiagnostics{
					Enabled: to.BoolPtr(true), StorageURI: &storageId,
				},
			}
			cloud.Resources["SA-"+pool.Name] = pool.Name
		} else {

			storageId := "https://" + pool.BootDiagnostics.StorageAccountId + ".blob.core.windows.net/"
			vm.VirtualMachineProperties.DiagnosticsProfile = &compute.DiagnosticsProfile{
				&compute.BootDiagnostics{
					Enabled: to.BoolPtr(true), StorageURI: &storageId,
				},
			}
		}
	}
	vmFuture, err := cloud.VMClient.CreateOrUpdate(cloud.context, resourceGroup, pool.Name, vm)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return compute.VirtualMachine{}, "", "", err
	} else {
		err = vmFuture.WaitForCompletion(cloud.context, cloud.VMClient.Client)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return compute.VirtualMachine{}, "", "", err
		}
	}
	beego.Info("Get VM  by name", pool.Name)
	vm, err = cloud.GetInstance(pool.Name, resourceGroup, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return compute.VirtualMachine{}, "", "", err
	}
	return vm, privateKey, publicKey, nil
}
*/
func getWoodpecker() string {
	return beego.AppConfig.String("woodpecker_url") + models.WoodpeckerEnpoint
}
func (cloud *AZURE) createVM(pool *NodePool, index int, nicParameters network.Interface, resourceGroup string, ctx utils.Context, token, projectId, vpcName string,zones []string) (compute.VirtualMachine, string, string, types.CustomCPError) {
	var zone []string
	if zones != nil{
		zone = []string{zones[0]}
	}else {
		zone = zones
	}
	var satype compute.StorageAccountTypes
	if pool.OsDisk == models.StandardSSD {
		satype = compute.StorageAccountTypesStandardSSDLRS
	} else if pool.OsDisk == models.PremiumSSD {
		satype = compute.StorageAccountTypesPremiumLRS
	} else if pool.OsDisk == models.StandardHDD {
		satype = compute.StorageAccountTypesStandardLRS

	}
	osDisk := &compute.OSDisk{
		CreateOption: compute.DiskCreateOptionTypesFromImage,
		Name:         to.StringPtr(pool.Name),
		ManagedDisk: &compute.ManagedDiskParameters{
			StorageAccountType: satype,
		},
	}
	if pool.Volume.DataDisk == models.StandardSSD {
		satype = compute.StorageAccountTypesStandardSSDLRS
	} else if pool.Volume.DataDisk == models.PremiumSSD {
		satype = compute.StorageAccountTypesPremiumLRS
	} else if pool.Volume.DataDisk == models.StandardHDD {
		satype = compute.StorageAccountTypesStandardLRS

	}

	storageName := "ext-" + pool.Name
	disk := compute.DataDisk{
		Lun:          to.Int32Ptr(int32(60)),
		Name:         to.StringPtr(storageName),
		CreateOption: compute.DiskCreateOptionTypesEmpty,
		DiskSizeGB:   to.Int32Ptr(pool.Volume.Size),
		ManagedDisk: &compute.ManagedDiskParameters{
			StorageAccountType: satype,
		},
	}

	var storage = []compute.DataDisk{}

	staticVolume := compute.DataDisk{
		Lun:          to.Int32Ptr(int32(index)),
		Name:         to.StringPtr("ext-master-" + pool.Name),
		CreateOption: compute.DiskCreateOptionTypesEmpty,
		DiskSizeGB:   to.Int32Ptr(5),
		ManagedDisk: &compute.ManagedDiskParameters{
			StorageAccountType: satype,
		},
	}
	cloud.Resources["ext-master-"+pool.Name] = "ext-master-" + pool.Name
	storage = append(storage, staticVolume)
	password := "Cloudplex1"
	vm := compute.VirtualMachine{
		Name:     to.StringPtr(pool.Name),
		Location: to.StringPtr(cloud.Region),
		Identity: &compute.VirtualMachineIdentity{
			Type: compute.ResourceIdentityTypeSystemAssigned,
		},
		Tags: map[string]*string{
			"network": to.StringPtr(vpcName),
		},
		Zones: &zone,
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			HardwareProfile: &compute.HardwareProfile{
				VMSize: compute.VirtualMachineSizeTypes(pool.MachineType),
			},
			StorageProfile: &compute.StorageProfile{
				ImageReference: &compute.ImageReference{
					Offer:     to.StringPtr(pool.Image.Offer),
					Sku:       to.StringPtr(pool.Image.Sku),
					Publisher: to.StringPtr(pool.Image.Publisher),
					Version:   to.StringPtr(pool.Image.Version),
				},
				OsDisk: osDisk,
				//DataDisks: &storage,
			},

			OsProfile: &compute.OSProfile{
				ComputerName:  to.StringPtr(pool.Name),
				AdminUsername: to.StringPtr(pool.AdminUser),
				AdminPassword: to.StringPtr(password),
			},
			NetworkProfile: &compute.NetworkProfile{

				NetworkInterfaces: &[]compute.NetworkInterfaceReference{
					{
						ID: &(*nicParameters.ID),
						NetworkInterfaceReferenceProperties: &compute.NetworkInterfaceReferenceProperties{
							Primary: to.BoolPtr(true),
						},
					},
				},
			},
		},
	}
	var fileName []string
	fileName = append(fileName, "static_volume.sh")
	if pool.EnableVolume {
		storage = append(storage, disk)
		cloud.Resources["ext-"+pool.Name] = "ext-" + pool.Name
		fileName = append(fileName, "azure-volume-mount.sh")
	}
	userData, err := userData2.GetUserData(token, getWoodpecker()+"/"+projectId, fileName, pool.PoolRole, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return compute.VirtualMachine{}, "", "", ApiError(err, "Error in creating VM", int(models.CloudStatusCode))
	}
	if userData != "no user data found" {
		encodedData := b64.StdEncoding.EncodeToString([]byte(userData))
		vm.OsProfile.CustomData = to.StringPtr(encodedData)
	}

	vm.StorageProfile.DataDisks = &storage

	private := ""
	public := ""
	if pool.KeyInfo.CredentialType == models.SSHKey {

		bytes, err := vault.GetSSHKey(string(models.Azure), pool.KeyInfo.KeyName, token, ctx, "")
		if err != nil {
			ctx.SendLogs("vm creation failed with error: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return compute.VirtualMachine{}, "", "", ApiError(err, "Error in VM creation", int(models.CloudStatusCode))
		}
		existingKey, err := key_utils.AzureKeyConversion(bytes, ctx)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return compute.VirtualMachine{}, "", "", ApiError(err, "Error in VM creation", int(models.CloudStatusCode))
		}

		if existingKey.PublicKey != "" && existingKey.PrivateKey != "" {
			key := []compute.SSHPublicKey{{
				Path:    to.StringPtr("/home/" + pool.AdminUser + "/.ssh/authorized_keys"),
				KeyData: to.StringPtr(existingKey.PublicKey),
			},
			}
			linux := &compute.LinuxConfiguration{
				SSH: &compute.SSHConfiguration{
					PublicKeys: &key,
				},
			}
			vm.OsProfile.LinuxConfiguration = linux
			private = existingKey.PrivateKey
			public = existingKey.PublicKey
		}
	} else if pool.KeyInfo.CredentialType == models.Password {
		vm.OsProfile.AdminPassword = to.StringPtr(pool.KeyInfo.AdminPassword)
	}

	if pool.BootDiagnostics.Enable {

		if pool.BootDiagnostics.NewStroageAccount {
			sName := strings.Replace(pool.Name, "-", "", -1)
			sName = strings.ToLower(sName)
			storageId := "https://" + sName + ".blob.core.windows.net/"
			err := cloud.createStorageAccount(resourceGroup, sName, ctx)
			if err != (types.CustomCPError{}) {
				ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return compute.VirtualMachine{}, "", "", err
			}
			vm.VirtualMachineProperties.DiagnosticsProfile = &compute.DiagnosticsProfile{
				&compute.BootDiagnostics{
					Enabled: to.BoolPtr(true), StorageURI: &storageId,
				},
			}
			cloud.Resources["SA-"+pool.Name] = sName
		} else {

			storageId := "https://" + pool.BootDiagnostics.StorageAccountId + ".blob.core.windows.net/"
			vm.VirtualMachineProperties.DiagnosticsProfile = &compute.DiagnosticsProfile{
				&compute.BootDiagnostics{
					Enabled: to.BoolPtr(true), StorageURI: &storageId,
				},
			}
		}
	}
	vmFuture, err := cloud.VMClient.CreateOrUpdate(cloud.context, resourceGroup, pool.Name, vm)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return compute.VirtualMachine{}, "", "", ApiError(err, "Error in VM creation", int(models.CloudStatusCode))
	} else {
		err = vmFuture.WaitForCompletionRef(cloud.context, cloud.VMClient.Client)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return compute.VirtualMachine{}, "", "", ApiError(err, "Error in VM creation", int(models.CloudStatusCode))
		}
	}
	beego.Info("Get VM  by name", pool.Name)
	vm, err1 := cloud.GetInstance(pool.Name, resourceGroup, ctx)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs(err1.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return compute.VirtualMachine{}, "", "", err1
	}
	return vm, private, public, types.CustomCPError{}
}

/*
func (cloud *AZURE) createVM(pool *NodePool, index int, nicParameters network.Interface, resourceGroup string, ctx utils.Context, token string) (compute.VirtualMachine, string, string, error) {
	var satype compute.StorageAccountTypes
	if pool.OsDisk == models.StandardSSD {
		satype = compute.StorageAccountTypesStandardSSDLRS
	} else if pool.OsDisk == models.PremiumSSD {
		satype = compute.StorageAccountTypesPremiumLRS
	} else if pool.OsDisk == models.StandardHDD {
		satype = compute.StorageAccountTypesStandardLRS

	}
	osDisk := &compute.OSDisk{
		CreateOption: compute.DiskCreateOptionTypesFromImage,
		Name:         to.StringPtr(pool.Name),
		ManagedDisk: &compute.ManagedDiskParameters{
			StorageAccountType: satype,
		},
	}
	if pool.Volume.DataDisk == models.StandardSSD {
		satype = compute.StorageAccountTypesStandardSSDLRS
	} else if pool.Volume.DataDisk == models.PremiumSSD {
		satype = compute.StorageAccountTypesPremiumLRS
	} else if pool.Volume.DataDisk == models.StandardHDD {
		satype = compute.StorageAccountTypesStandardLRS

	}

	storageName := "ext-" + pool.Name
	disk := compute.DataDisk{
		Lun:          to.Int32Ptr(int32(index)),
		Name:         to.StringPtr(storageName),
		CreateOption: compute.DiskCreateOptionTypesEmpty,
		DiskSizeGB:   to.Int32Ptr(pool.Volume.Size),
		ManagedDisk: &compute.ManagedDiskParameters{
			StorageAccountType: satype,
		},
	}

	var storage = []compute.DataDisk{}

	staticVolume := compute.DataDisk{
		Lun:          to.Int32Ptr(int32(index)),
		Name:         to.StringPtr("ext-master-" + pool.Name),
		CreateOption: compute.DiskCreateOptionTypesEmpty,
		DiskSizeGB:   to.Int32Ptr(5),
		ManagedDisk: &compute.ManagedDiskParameters{
			StorageAccountType: satype,
		},
	}
	cloud.Resources["ext-master-"+pool.Name] = "ext-master-" + pool.Name
	storage = append(storage, staticVolume)

	vm := compute.VirtualMachine{
		Name:     to.StringPtr(pool.Name),
		Location: to.StringPtr(cloud.Region),
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			HardwareProfile: &compute.HardwareProfile{
				VMSize: compute.VirtualMachineSizeTypes(pool.MachineType),
			},
			StorageProfile: &compute.StorageProfile{
				ImageReference: &compute.ImageReference{
					Offer:     to.StringPtr(pool.Image.Offer),
					Sku:       to.StringPtr(pool.Image.Sku),
					Publisher: to.StringPtr(pool.Image.Publisher),
					Version:   to.StringPtr(pool.Image.Version),
				},
				OsDisk: osDisk,
				//DataDisks: &storage,
			},
			OsProfile: &compute.OSProfile{
				ComputerName:  to.StringPtr(pool.Name),
				AdminUsername: to.StringPtr(pool.AdminUser),
			},
			NetworkProfile: &compute.NetworkProfile{

				NetworkInterfaces: &[]compute.NetworkInterfaceReference{
					{
						ID: &(*nicParameters.ID),
						NetworkInterfaceReferenceProperties: &compute.NetworkInterfaceReferenceProperties{
							Primary: to.BoolPtr(true),
						},
					},
				},
			},
		},
	}
	if pool.EnableVolume {
		storage = append(storage, disk)
		cloud.Resources["ext-"+pool.Name] = "ext-" + pool.Name
	}
	vm.StorageProfile.DataDisks = &storage
	private := ""
	public := ""
	if pool.KeyInfo.CredentialType == models.SSHKey && pool.KeyInfo.NewKey == models.NEWKey {
		k, err := vault.GetAzureSSHKey("azure", pool.KeyInfo.KeyName, token, ctx)

		if err != nil && err.Error() != "not found" {
			ctx.SendLogs("vm creation failed", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return compute.VirtualMachine{}, "", "", err
		} else if err == nil {

			existingKey, err := key_utils.KeyConversion(k, ctx)
			if err != nil {
				return compute.VirtualMachine{}, "", "", err
			}
			if existingKey.PublicKey != "" && existingKey.PrivateKey != "" {
				key := []compute.SSHPublicKey{{
					Path:    to.StringPtr("/home/" + pool.AdminUser + "/.ssh/authorized_keys"),
					KeyData: to.StringPtr(existingKey.PublicKey),
				},
				}
				linux := &compute.LinuxConfiguration{
					SSH: &compute.SSHConfiguration{
						PublicKeys: &key,
					},
				}
				vm.OsProfile.LinuxConfiguration = linux
				private = existingKey.PrivateKey
				public = existingKey.PublicKey
			}
		} else if err != nil && err.Error() == "not found" {

			res, err := key_utils.GenerateKeyPair(pool.KeyInfo.KeyName, "azure@example.com", ctx)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return compute.VirtualMachine{}, "", "", err
			}
			key := []compute.SSHPublicKey{{
				Path:    to.StringPtr("/home/" + pool.AdminUser + "/.ssh/authorized_keys"),
				KeyData: to.StringPtr(res.PublicKey),
			},
			}
			linux := &compute.LinuxConfiguration{
				SSH: &compute.SSHConfiguration{
					PublicKeys: &key,
				},
			}
			vm.OsProfile.LinuxConfiguration = linux
			pool.KeyInfo.PublicKey = res.PublicKey
			pool.KeyInfo.PrivateKey = res.PrivateKey

			_, err = vault.PostAzureSSHKey(pool.KeyInfo, ctx, token)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return compute.VirtualMachine{}, "", "", err
			}

			public = res.PublicKey
			private = res.PrivateKey
		}

	} else if pool.KeyInfo.CredentialType == models.SSHKey && pool.KeyInfo.NewKey == models.CPKey {

		k, err := vault.GetAzureSSHKey("azure", pool.KeyInfo.KeyName, token, ctx)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return compute.VirtualMachine{}, "", "", err
		}

		existingKey, err := key_utils.KeyConversion(k, ctx)
		if err != nil {
			return compute.VirtualMachine{}, "", "", err
		}

		key := []compute.SSHPublicKey{{
			Path:    to.StringPtr("/home/" + pool.AdminUser + "/.ssh/authorized_keys"),
			KeyData: to.StringPtr(existingKey.PublicKey),
		}}

		linux := &compute.LinuxConfiguration{
			SSH: &compute.SSHConfiguration{

				PublicKeys: &key,
			},
		}
		vm.OsProfile.LinuxConfiguration = linux

		private = existingKey.PrivateKey
		public = existingKey.PublicKey
	} else {
		vm.OsProfile.AdminPassword = to.StringPtr(pool.KeyInfo.AdminPassword)
	}

	if pool.BootDiagnostics.Enable {

		if pool.BootDiagnostics.NewStroageAccount {
			sName := strings.Replace(pool.Name, "-", "", -1)
			sName = strings.ToLower(sName)
			storageId := "https://" + sName + ".blob.core.windows.net/"
			err := cloud.createStorageAccount(resourceGroup, sName, ctx)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return compute.VirtualMachine{}, "", "", err
			}
			vm.VirtualMachineProperties.DiagnosticsProfile = &compute.DiagnosticsProfile{
				&compute.BootDiagnostics{
					Enabled: to.BoolPtr(true), StorageURI: &storageId,
				},
			}
			cloud.Resources["SA-"+pool.Name] = pool.Name
		} else {

			storageId := "https://" + pool.BootDiagnostics.StorageAccountId + ".blob.core.windows.net/"
			vm.VirtualMachineProperties.DiagnosticsProfile = &compute.DiagnosticsProfile{
				&compute.BootDiagnostics{
					Enabled: to.BoolPtr(true), StorageURI: &storageId,
				},
			}
		}
	}
	vmFuture, err := cloud.VMClient.CreateOrUpdate(cloud.context, resourceGroup, pool.Name, vm)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return compute.VirtualMachine{}, "", "", err
	} else {
		err = vmFuture.WaitForCompletion(cloud.context, cloud.VMClient.Client)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return compute.VirtualMachine{}, "", "", err
		}
	}
	beego.Info("Get VM  by name", pool.Name)
	vm, err = cloud.GetInstance(pool.Name, resourceGroup, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return compute.VirtualMachine{}, "", "", err
	}
	return vm, private, public, nil
}
*/
func (cloud *AZURE) createStorageAccount(resouceGroup string, acccountName string, ctx utils.Context) types.CustomCPError {
	accountParameters := storage.AccountCreateParameters{
		Sku: &storage.Sku{
			Name: storage.StandardLRS,
		},
		Location:                          &cloud.Region,
		AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{},
	}
	acccountName = strings.ToLower(acccountName)
	future, err := cloud.AccountClient.Create(context.Background(), resouceGroup, acccountName, accountParameters)
	if err != nil {
		beego.Error("Storage account creation failed")
		beego.Info(err)
		return ApiError(err, "Error in storage account creation", int(models.CloudStatusCode))
	}
	err = future.WaitForCompletionRef(context.Background(), cloud.AccountClient.Client)
	if err != nil {

		beego.Error("Storage account creation failed")
		beego.Info(err)
		return ApiError(err, "Error in storage account creation", int(models.CloudStatusCode))
	}
	/*account, err := cloud.AccountClient.GetProperties(cloud.context, resouceGroup, acccountName)
	if err != nil {
		beego.Error(err.Error())
		return "", err
	}
	beego.Info(*account.ID)*/
	return types.CustomCPError{}
}
func (cloud *AZURE) deleteDisk(resouceGroup string, diskName string, ctx utils.Context) types.CustomCPError {

	_, err := cloud.DiskClient.Delete(context.Background(), resouceGroup, diskName)

	if err != nil && !strings.Contains(err.Error(), "not found") {
		ctx.SendLogs("Disk deletion failed"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiError(err, "Error in deleting disk", int(models.CloudStatusCode))
	} else if err != nil && strings.Contains(err.Error(), "not found") {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	}
	return types.CustomCPError{}
}
func (cloud *AZURE) deleteStorageAccount(resouceGroup string, acccountName string, ctx utils.Context) types.CustomCPError {

	acccountName = strings.ToLower(acccountName)
	_, err := cloud.AccountClient.Delete(context.Background(), resouceGroup, acccountName)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		ctx.SendLogs("Storage account deletion failed"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiError(err, "Error in deleting storage account", int(models.CloudStatusCode))
	} else if err != nil && strings.Contains(err.Error(), "not found") {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	}
	return types.CustomCPError{}
}
func (cloud *AZURE) CleanUp(cluster Cluster_Def, ctx utils.Context, companyId string) types.CustomCPError {
	for _, pool := range cluster.NodePools {
		if pool.PoolRole == "master" {
			if cloud.Resources["NodeName-"+cluster.ProjectId] != nil {
				name := cloud.Resources["NodeName-"+cluster.ProjectId]
				nodeName := ""
				b, e := json.Marshal(name)
				if e != nil {
					beego.Info(e.Error())
					return ApiError(e, "Error in cluster cleanup", int(models.CloudStatusCode))
				}
				e = json.Unmarshal(b, &nodeName)
				if e != nil {
					beego.Info(e.Error())
					return ApiError(e, "Error in cluster cleanup", int(models.CloudStatusCode))
				}

				err := cloud.TerminateMasterNode(nodeName, cluster.ProjectId, cluster.ResourceGroup, ctx, companyId)
				if err != (types.CustomCPError{}) {
					beego.Info(e.Error())
					return err
				}
			}
			if cloud.Resources["Nic-"+cluster.ProjectId] != nil {
				name := cloud.Resources["Nic-"+cluster.ProjectId]
				nicName := ""
				b, e := json.Marshal(name)
				if e != nil {
					beego.Info(e.Error())
					return ApiError(e, "Error in cluster cleanup", int(models.CloudStatusCode))
				}
				e = json.Unmarshal(b, &nicName)
				if e != nil {
					beego.Info(e.Error())
					return ApiError(e, "Error in cluster cleanup", int(models.CloudStatusCode))
				}
				err := cloud.deleteNIC(nicName, cluster.ResourceGroup, cluster.ProjectId, ctx, companyId)
				if err != (types.CustomCPError{}) {
					beego.Info(err.Error)
					return err
				}
			}
			if cloud.Resources["Pip-"+cluster.ProjectId] != nil {
				name := cloud.Resources["Pip-"+cluster.ProjectId]
				IPname := ""
				b, e := json.Marshal(name)
				if e != nil {
					return ApiError(e, "Error in cluster cleanup", int(models.CloudStatusCode))
				}
				e = json.Unmarshal(b, &IPname)
				if e != nil {
					return ApiError(e, "Error in cluster cleanup", int(models.CloudStatusCode))
				}
				err := cloud.deletePublicIp(IPname, cluster.ResourceGroup, cluster.ProjectId, ctx, companyId)
				if err != (types.CustomCPError{}) {
					return err
				}
			}
			if cloud.Resources["SA-"+pool.Name] != nil {
				name := cloud.Resources["SA-"+pool.Name]
				SAname := ""
				b, e := json.Marshal(name)
				if e != nil {
					return ApiError(e, "Error in cluster cleanup", int(models.CloudStatusCode))
				}
				e = json.Unmarshal(b, &SAname)
				if e != nil {
					return ApiError(e, "Error in cluster cleanup", int(models.CloudStatusCode))
				}
				err := cloud.deleteStorageAccount(cluster.ResourceGroup, SAname, ctx)
				if err != (types.CustomCPError{}) {
					return err
				}
			}
			if cloud.Resources["Disk-"+cluster.ProjectId] != nil {
				name := cloud.Resources["Disk-"+cluster.ProjectId]
				diskName := ""
				b, e := json.Marshal(name)
				if e != nil {
					return ApiError(e, "Error in cluster cleanup", int(models.CloudStatusCode))
				}
				e = json.Unmarshal(b, &diskName)
				if e != nil {
					return ApiError(e, "Error in cluster cleanup", int(models.CloudStatusCode))
				}
				err := cloud.deleteDisk(cluster.ResourceGroup, diskName, ctx)
				if err != (types.CustomCPError{}) {
					return err
				}
			}
			if cloud.Resources["ext-"+pool.Name] != nil {
				err := cloud.deleteDisk(cluster.ResourceGroup, "ext-"+pool.Name, ctx)
				if err != (types.CustomCPError{}) {
					return err
				}
			}
			if cloud.Resources["ext-master-"+pool.Name] != nil {
				err := cloud.deleteDisk(cluster.ResourceGroup, "ext-master-"+pool.Name, ctx)
				if err != (types.CustomCPError{}) {
					return err
				}
			}
		} else {

			if cloud.Resources["Vmss-"+pool.Name] != nil {
				name := cloud.Resources["Vmss-"+pool.Name]
				vmssName := ""
				b, e := json.Marshal(name)
				if e != nil {
					return ApiError(e, "Error in cluster cleanup", int(models.CloudStatusCode))
				}
				e = json.Unmarshal(b, &vmssName)
				if e != nil {
					return ApiError(e, "Error in cluster cleanup", int(models.CloudStatusCode))
				}
				beego.Info(vmssName)
				err := cloud.TerminatePool(vmssName, cluster.ResourceGroup, cluster.ProjectId, ctx)
				if err != (types.CustomCPError{}) {
					return err
				}
			}

			if cloud.Resources["SA-"+pool.Name] != nil {
				name := cloud.Resources["SA-"+pool.Name]
				SAname := ""
				b, e := json.Marshal(name)
				if e != nil {
					return ApiError(e, "Error in cluster cleanup", int(models.CloudStatusCode))
				}
				e = json.Unmarshal(b, &SAname)
				if e != nil {
					return ApiError(e, "Error in cluster cleanup", int(models.CloudStatusCode))
				}
				err := cloud.deleteStorageAccount(cluster.ResourceGroup, SAname, ctx)
				if err != (types.CustomCPError{}) {
					return err
				}
			}
		}
	}

	return types.CustomCPError{}
}
func (cloud *AZURE) mountVolume(vms []*VM, privateKey string, KeyName string, projectId string, user string, resourceGroup string, poolName string, ctx utils.Context, poleRole string, masterVolume bool) types.CustomCPError {

	for _, vm := range vms {
		err := fileWrite(privateKey, KeyName)
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
		err = setPermission(KeyName)
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
		start := time.Now()
		timeToWait := 90 //seconds
		retry := true
		var errPublicIP types.CustomCPError

		if vm.PublicIP == nil {
			for retry && int64(time.Since(start).Seconds()) < int64(timeToWait) {

				if poleRole == "master" {
					IPname := "pip-" + poolName
					publicIp, errPublicIP := cloud.GetVMPIP(resourceGroup, IPname, ctx)
					if errPublicIP != (types.CustomCPError{}) {
						ctx.SendLogs(errPublicIP.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
						return errPublicIP
					} else if publicIp.IPAddress == nil {
						ctx.SendLogs("waiting 15 seconds before retry", models.LOGGING_LEVEL_WARNING, models.Backend_Logging)
						time.Sleep(15 * time.Second)

					} else {
						vm.PublicIP = publicIp.IPAddress
						retry = false

					}
				} else {
					IPname := "pip-" + poolName
					vmname := string(*vm.Name)
					vMname := vmname[len(vmname)-1:]
					nic := "nic-" + poolName
					ipConfig := poolName
					publicIp, errPublicIP := cloud.GetPIP(resourceGroup, poolName, vMname, nic, ipConfig, IPname, ctx)
					if errPublicIP != (types.CustomCPError{}) {
						ctx.SendLogs(errPublicIP.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
						return errPublicIP
					} else if publicIp.IPAddress == nil {
						ctx.SendLogs("waiting 15 seconds before retry", models.LOGGING_LEVEL_WARNING, models.Backend_Logging)
						time.Sleep(15 * time.Second)

					} else {
						vm.PublicIP = publicIp.IPAddress
						retry = false

					}
				}
			}
		}
		if errPublicIP != (types.CustomCPError{}) {
			return errPublicIP
		}
		if vm.PublicIP == nil {
			str := "Public IP is not available. Cannot mount volume"
			ctx.SendLogs(str, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return ApiError(errors.New(str), "Error in mounting volume", int(models.CloudStatusCode))

		}
		start = time.Now()
		timeToWait = 90 //seconds
		retry = true
		fileName := ""
		var errCopy types.CustomCPError
		if masterVolume {
			fileName = "static_volume.sh"
		} else {
			fileName = "azure-volume-mount.sh"
		}

		for retry && int64(time.Since(start).Seconds()) < int64(timeToWait) {

			errCopy = copyFile(KeyName, user, *vm.PublicIP, fileName)
			if errCopy != (types.CustomCPError{}) && strings.Contains(errCopy.Error, "exit status 1") {
				ctx.SendLogs(errCopy.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				ctx.SendLogs("waiting 5 seconds before retry", models.LOGGING_LEVEL_WARNING, models.Backend_Logging)
				time.Sleep(5 * time.Second)
			} else {
				retry = false
			}
		}
		if errCopy != (types.CustomCPError{}) {
			return errCopy
		}
		err = setScriptPermision(KeyName, user, *vm.PublicIP, fileName, ctx)
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
		err = runScript(KeyName, user, *vm.PublicIP, fileName, ctx)
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
		err = deleteScript(KeyName, user, *vm.PublicIP, fileName, ctx)
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
		err = deleteFile(KeyName, ctx)
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}
	return types.CustomCPError{}

}
func fileWrite(key string, keyName string) types.CustomCPError {

	f, err := os.Create("/app/keys/" + keyName + ".pem")
	if err != nil {
		beego.Error(err.Error())
		return ApiError(err, "Error in mounting volume", int(models.CloudStatusCode))
	}
	defer f.Close()
	d2 := []byte(key)
	n2, err := f.Write(d2)
	if err != nil {
		beego.Error(err.Error())
		return ApiError(err, "Error in mounting volume", int(models.CloudStatusCode))
	}
	beego.Info("wrote %d bytes\n", n2)

	err = os.Chmod("/app/keys/"+keyName+".pem", 0777)
	if err != nil {
		beego.Error(err)
		return ApiError(err, "Error in mounting volume", int(models.CloudStatusCode))
	}
	return types.CustomCPError{}
}
func setPermission(keyName string) types.CustomCPError {
	keyPath := "/app/keys/" + keyName + ".pem"
	cmd1 := "chmod"
	beego.Info(keyPath)
	args := []string{"600", keyPath}
	cmd := exec.Command(cmd1, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		beego.Error(err.Error())
		return ApiError(err, "Error in mounting volume", int(models.CloudStatusCode))
	}
	return types.CustomCPError{}
}

func copyFile(keyName string, userName string, instanceId string, file string) types.CustomCPError {

	keyPath := "/app/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId + ":/home/" + userName
	cmd1 := "scp"
	beego.Info(keyPath)
	beego.Info(ip)
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, "/app/scripts/" + file, ip}
	cmd := exec.Command(cmd1, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		beego.Error(err.Error())
		return ApiError(err, "Error in mounting volume", int(models.CloudStatusCode))
	}
	return types.CustomCPError{}
}

func setScriptPermision(keyName string, userName string, instanceId, fileName string, ctx utils.Context) types.CustomCPError {
	keyPath := "/app/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId
	cmd1 := "ssh"
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, ip, "chmod 700 /home/" + userName + "/" + fileName}
	cmd := exec.Command(cmd1, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiError(err, "Error in cluster creation", int(models.CloudStatusCode))
	}
	return types.CustomCPError{}
}

func runScript(keyName string, userName string, instanceId string, fileName string, ctx utils.Context) types.CustomCPError {
	keyPath := "/app/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId
	cmd1 := "ssh"
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, ip, "/home/" + userName + "/" + fileName}
	cmd := exec.Command(cmd1, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiError(err, "Error in mounting volume", int(models.CloudStatusCode))
	}
	return types.CustomCPError{}
}

func deleteScript(keyName string, userName string, instanceId string, fileName string, ctx utils.Context) types.CustomCPError {
	keyPath := "/app/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId
	cmd1 := "ssh"
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, ip, "rm", "/home/" + userName + "/" + fileName}
	cmd := exec.Command(cmd1, args...)
	err := cmd.Run()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiError(err, "Error in mounting volume", int(models.CloudStatusCode))
	}
	return types.CustomCPError{}
}

func deleteFile(keyName string, ctx utils.Context) types.CustomCPError {
	keyPath := "/app/keys/" + keyName + ".pem"
	err := os.Remove(keyPath)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiError(err, "Error in mounting volume", int(models.CloudStatusCode))
	}
	return types.CustomCPError{}
}
func (cloud *AZURE) createVMSS(resourceGroup string, projectId string, pool *NodePool, poolIndex int, subnetId string, sgIds []*string, ctx utils.Context, token, vpcName string,zones []string) (compute.VirtualMachineScaleSetVMListResultPage, types.CustomCPError, string) {

	var satype compute.StorageAccountTypes
	if pool.OsDisk == models.StandardSSD {
		satype = compute.StorageAccountTypesStandardSSDLRS
	} else if pool.OsDisk == models.PremiumSSD {
		satype = compute.StorageAccountTypesPremiumLRS
	} else if pool.OsDisk == models.StandardHDD {
		satype = compute.StorageAccountTypesStandardLRS

	}
	osDisk := &compute.VirtualMachineScaleSetOSDisk{
		CreateOption: compute.DiskCreateOptionTypesFromImage,
		//Name:         to.StringPtr(pool.Name),
		ManagedDisk: &compute.VirtualMachineScaleSetManagedDiskParameters{
			StorageAccountType: satype,
		},
	}
	if pool.Volume.DataDisk == models.StandardSSD {
		satype = compute.StorageAccountTypesStandardSSDLRS
	} else if pool.Volume.DataDisk == models.PremiumSSD {
		satype = compute.StorageAccountTypesPremiumLRS
	} else if pool.Volume.DataDisk == models.StandardHDD {
		satype = compute.StorageAccountTypesStandardLRS

	}

	//storageName := "ext-" + pool.Name
	disk := compute.VirtualMachineScaleSetDataDisk{
		Lun: to.Int32Ptr(int32(poolIndex)),
		//Name:         to.StringPtr(storageName),
		CreateOption: compute.DiskCreateOptionTypesEmpty,
		DiskSizeGB:   to.Int32Ptr(pool.Volume.Size),
		ManagedDisk: &compute.VirtualMachineScaleSetManagedDiskParameters{
			StorageAccountType: satype,
		},
	}
	storage := []compute.VirtualMachineScaleSetDataDisk{disk}
	params := compute.VirtualMachineScaleSet{
		Name:     to.StringPtr(pool.Name),
		Location: to.StringPtr(cloud.Region),
		Zones: to.StringSlicePtr(zones),
		Identity: &compute.VirtualMachineScaleSetIdentity{
			Type: compute.ResourceIdentityTypeSystemAssigned,
		},
		Tags: map[string]*string{
			"network": to.StringPtr(vpcName),
		},
		Sku: &compute.Sku{
			Capacity: to.Int64Ptr(pool.NodeCount),
			Name:     to.StringPtr(pool.MachineType),
		},
		VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
			VirtualMachineProfile: &compute.VirtualMachineScaleSetVMProfile{

				StorageProfile: &compute.VirtualMachineScaleSetStorageProfile{
					ImageReference: &compute.ImageReference{
						Offer:     to.StringPtr(pool.Image.Offer),
						Sku:       to.StringPtr(pool.Image.Sku),
						Publisher: to.StringPtr(pool.Image.Publisher),
						Version:   to.StringPtr(pool.Image.Version),
					},
					OsDisk: osDisk,
				},
				OsProfile: &compute.VirtualMachineScaleSetOSProfile{
					ComputerNamePrefix: to.StringPtr(pool.Name),
					AdminUsername:      to.StringPtr(pool.AdminUser),
				},
				NetworkProfile: &compute.VirtualMachineScaleSetNetworkProfile{

					NetworkInterfaceConfigurations: &[]compute.VirtualMachineScaleSetNetworkConfiguration{
						{
							Name: to.StringPtr("nic-" + pool.Name),
							VirtualMachineScaleSetNetworkConfigurationProperties: &compute.VirtualMachineScaleSetNetworkConfigurationProperties{
								Primary: to.BoolPtr(true),
								IPConfigurations: &[]compute.VirtualMachineScaleSetIPConfiguration{
									{
										Name: to.StringPtr(pool.Name),
										VirtualMachineScaleSetIPConfigurationProperties: &compute.VirtualMachineScaleSetIPConfigurationProperties{
											Subnet: &compute.APIEntityReference{ID: to.StringPtr(subnetId)},
										},
									},
								},
								NetworkSecurityGroup: &compute.SubResource{
									ID: to.StringPtr(*sgIds[0]),
								},
							},
						},
					},
				},
			},
		},
	}
	if pool.EnablePublicIP {
		p := (*params.VirtualMachineScaleSetProperties.VirtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations)[0].VirtualMachineScaleSetNetworkConfigurationProperties.IPConfigurations
		(*p)[0].VirtualMachineScaleSetIPConfigurationProperties.PublicIPAddressConfiguration = &compute.VirtualMachineScaleSetPublicIPAddressConfiguration{
			Name: to.StringPtr("pip-" + pool.Name),
			VirtualMachineScaleSetPublicIPAddressConfigurationProperties: &compute.VirtualMachineScaleSetPublicIPAddressConfigurationProperties{
				DNSSettings: &compute.VirtualMachineScaleSetPublicIPAddressConfigurationDNSSettings{
					DomainNameLabel: to.StringPtr(strings.ToLower(pool.Name)),
				},
			},
		}
	}
	var fileName []string
	if pool.EnableVolume {
		params.VirtualMachineProfile.StorageProfile.DataDisks = &storage
		fileName = append(fileName, "azure-volume-slave-mount.sh")
	}

	userData, err := userData2.GetUserData(token, getWoodpecker()+"/"+projectId, fileName, pool.PoolRole, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return compute.VirtualMachineScaleSetVMListResultPage{}, ApiError(err, "Error in VMSS creation", int(models.CloudStatusCode)), ""
	}
	if userData != "no user data found" {
		encodedData := b64.StdEncoding.EncodeToString([]byte(userData))
		params.VirtualMachineScaleSetProperties.VirtualMachineProfile.OsProfile.CustomData = to.StringPtr(encodedData)
	}

	private := ""
	// public := ""

	if pool.KeyInfo.CredentialType == models.SSHKey {

		bytes, err := vault.GetSSHKey(string(models.Azure), pool.KeyInfo.KeyName, token, ctx, "")
		if err != nil {
			ctx.SendLogs("vm creation failed with error: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return compute.VirtualMachineScaleSetVMListResultPage{}, ApiError(err, "Error in VMSS creation", int(models.CloudStatusCode)), ""
		}
		existingKey, err := key_utils.AzureKeyConversion(bytes, ctx)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return compute.VirtualMachineScaleSetVMListResultPage{}, ApiError(err, "Error in VMSS creation", int(models.CloudStatusCode)), ""
		}

		if existingKey.PublicKey != "" && existingKey.PrivateKey != "" {
			key := []compute.SSHPublicKey{{
				Path:    to.StringPtr("/home/" + pool.AdminUser + "/.ssh/authorized_keys"),
				KeyData: to.StringPtr(existingKey.PublicKey),
			},
			}
			linux := &compute.LinuxConfiguration{
				SSH: &compute.SSHConfiguration{
					PublicKeys: &key,
				},
			}
			params.VirtualMachineProfile.OsProfile.LinuxConfiguration = linux
			private = existingKey.PrivateKey
			//public = existingKey.PublicKey
		}
	} else if pool.KeyInfo.CredentialType == models.Password {
		params.VirtualMachineProfile.OsProfile.AdminPassword = to.StringPtr(pool.KeyInfo.AdminPassword)
	}

	if pool.BootDiagnostics.Enable {

		if pool.BootDiagnostics.NewStroageAccount {
			sName := strings.Replace(pool.Name, "-", "", -1)
			sName = strings.ToLower(sName)
			storageId := "https://" + sName + ".blob.core.windows.net/"
			err := cloud.createStorageAccount(resourceGroup, sName, ctx)
			if err != (types.CustomCPError{}) {
				ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
			}
			params.VirtualMachineProfile.DiagnosticsProfile = &compute.DiagnosticsProfile{
				&compute.BootDiagnostics{
					Enabled: to.BoolPtr(true), StorageURI: &storageId,
				},
			}
			cloud.Resources["SA-"+pool.Name] = sName
		} else {

			storageId := "https://" + pool.BootDiagnostics.StorageAccountId + ".blob.core.windows.net/"
			params.VirtualMachineProfile.DiagnosticsProfile = &compute.DiagnosticsProfile{
				&compute.BootDiagnostics{
					Enabled: to.BoolPtr(true), StorageURI: &storageId,
				},
			}
		}
	}
	params.UpgradePolicy = &compute.UpgradePolicy{
		Mode: compute.Manual,
	}
	address, err := cloud.VMSSCLient.CreateOrUpdate(cloud.context, resourceGroup, pool.Name, params)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return compute.VirtualMachineScaleSetVMListResultPage{}, ApiError(err, "Error in VMSS creation", int(models.CloudStatusCode)), ""
	} else {
		err = address.WaitForCompletionRef(cloud.context, cloud.AddressClient.Client)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return compute.VirtualMachineScaleSetVMListResultPage{}, ApiError(err, "Error in VMSS creation", int(models.CloudStatusCode)), ""
		}
	}
	vms, err := cloud.VMSSVMClient.List(cloud.context, resourceGroup, pool.Name, "", "", "")
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return compute.VirtualMachineScaleSetVMListResultPage{}, ApiError(err, "Error in VMSS creation", int(models.CloudStatusCode)), ""
	}

	return vms, types.CustomCPError{}, private
}

func (cloud *AZURE) getAllInstances() ([]azureVM, types.CustomCPError) {
	if cloud == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			return []azureVM{}, err
		}
	}

	var instanceList []azureVM
	VmResult, err1 := cloud.VMClient.ListAll(context.Background())
	if err1 != nil {
		return []azureVM{}, ApiError(err1, "Error in fetching all instances", int(models.CloudStatusCode))
	}
	for _, instance := range VmResult.Values() {
		bytes, err1 := json.Marshal(instance)
		if err1 != nil {
			beego.Error(err1.Error())
			return []azureVM{}, ApiError(err1, "Error in fetching all instances", int(models.CloudStatusCode))
		}

		var vm azureVM
		err1 = json.Unmarshal(bytes, &vm)
		if err1 != nil {
			beego.Error(err1.Error())
			return []azureVM{}, ApiError(err1, "Error in fetching all instances", int(models.CloudStatusCode))
		}
		instanceList = append(instanceList, vm)
	}
	VMSSResult, err1 := cloud.VMSSCLient.ListAll(context.Background())
	if err1 != nil {
		beego.Error(err1.Error())
		return []azureVM{}, ApiError(err1, "Error in fetching all instances", int(models.CloudStatusCode))
	}

	for _, instance := range VMSSResult.Values() {
		bytes, err1 := json.Marshal(instance)
		if err1 != nil {
			beego.Error(err1.Error())
			return []azureVM{}, ApiError(err1, "Error in fetching all instances", int(models.CloudStatusCode))
		}

		var vm azureVM
		err1 = json.Unmarshal(bytes, &vm)
		if err1 != nil {
			beego.Error(err1.Error())
			return []azureVM{}, ApiError(err1, "Error in fetching all instances", int(models.CloudStatusCode))
		}
		instanceList = append(instanceList, vm)
	}
	return instanceList, types.CustomCPError{}
}

func (cloud *AZURE) getRegions(ctx utils.Context) (region []models.Region, err types.CustomCPError) {
	var reg models.Region
	if cloud == nil {
		err := cloud.init()
		if err != (types.CustomCPError{}) {
			return []models.Region{}, err
		}
	}

	LocResult, err2 := cloud.Location.ListLocations(cloud.context, cloud.Subscription)
	if err2 != nil {
		beego.Error(err2.Error())
		return []models.Region{}, ApiError(err2, "Error in listing locations", int(models.CloudStatusCode))
	}
	for _, loc := range *LocResult.Value {
		reg.Name = *loc.DisplayName
		reg.Location = *loc.Name
		region = append(region, reg)
	}
	return region, types.CustomCPError{}
}

func getAllVMSizes() ([]string, types.CustomCPError) {

	VmResult := c.PossibleVirtualMachineSizeTypesValues()
	if VmResult == nil {
		return []string{}, ApiError(errors.New("VM Machine Type Not Fetched"), "Error in fetching VM Machine", int(models.CloudStatusCode))
	}
	var machine []string
	for _, vm := range VmResult {
		fmt.Println(string(vm))
		machine = append(machine, string(vm))
	}
	return machine, types.CustomCPError{}
}
