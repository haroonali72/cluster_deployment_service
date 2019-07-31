package azure

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/key_utils"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-02-01/network"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2017-10-01/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/astaxie/beego"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type CreatedPool struct {
	Instances []*compute.VirtualMachine
	KeyName   string
	Key       string
	PoolName  string
}
type AZURE struct {
	Authorizer       *autorest.BearerAuthorizer
	AddressClient    network.PublicIPAddressesClient
	InterfacesClient network.InterfacesClient
	VMSSCLient       compute.VirtualMachineScaleSetsClient
	VMSSVMClient     compute.VirtualMachineScaleSetVMsClient
	VMClient         compute.VirtualMachinesClient
	DiskClient       compute.DisksClient
	AccountClient    storage.AccountsClient
	context          context.Context
	ID               string
	Key              string
	Tenant           string
	Subscription     string
	Region           string
	Resources        map[string]interface{}
}

func (cloud *AZURE) init() error {
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

	cloud.Resources = make(map[string]interface{})

	return nil
}
func getNetworkHost(cloudType string) string {
	host := beego.AppConfig.String("network_url")
	if strings.Contains(host, "{cloud_provider}") {
		host = strings.Replace(host, "{cloud_provider}", cloudType, -1)
	}
	return host

}
func (cloud *AZURE) createCluster(cluster Cluster_Def, ctx utils.Context) (Cluster_Def, error) {

	if cloud == nil {
		err := cloud.init()
		if err != nil {
			beego.Error(err.Error())
			return cluster, err
		}
	}

	var azureNetwork types.AzureNetwork
	url := getNetworkHost("azure") + "/" + cluster.ProjectId
	network, err := api_handler.GetAPIStatus(url, ctx)
	err = json.Unmarshal(network.([]byte), &azureNetwork)

	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return cluster, err
	}

	for i, pool := range cluster.NodePools {

		ctx.SendSDLog("AZUREOperations creating nodes", "info")

		result, private_key, err := cloud.CreateInstance(pool, azureNetwork, cluster.ResourceGroup, cluster.ProjectId, i, ctx)
		if err != nil {
			ctx.SendSDLog(err.Error(), "error")
			return cluster, err
		}
		if pool.EnableVolume {
			err = cloud.mountVolume(result, private_key, pool.KeyInfo.KeyName, cluster.ProjectId, pool.AdminUser, cluster.ResourceGroup, pool.Name, ctx)
			if err != nil {
				utils.SendLog("Error in volume mounting : "+err.Error(), "info", cluster.ProjectId)
				return cluster, err
			}
		}
		cluster.NodePools[i].Nodes = result
	}

	return cluster, nil
}
func (cloud *AZURE) CreateInstance(pool *NodePool, networkData types.AzureNetwork, resourceGroup string, projectId string, poolIndex int, ctx utils.Context) ([]*VM, string, error) {

	var cpVms []*VM

	subnetId := cloud.GetSubnets(pool, networkData)
	sgIds := cloud.GetSecurityGroups(pool, networkData)
	//subnetId := "/subscriptions/aa94b050-2c52-4b7b-9ce3-2ac18253e61e/resourceGroups/testsadaf/providers/Microsoft.Network/virtualNetworks/testsadaf-vnet/subnets/default"
	//var sgIds []*string
	//sid := "/subscriptions/aa94b050-2c52-4b7b-9ce3-2ac18253e61e/resourceGroups/testsadaf/providers/Microsoft.Network/networkSecurityGroups/fgfdnsg"
	//sgIds = append(sgIds, &sid)
	if pool.PoolRole == "master" {
		IPname := "pip-" + pool.Name
		utils.SendLog("Creating Public IP : "+projectId, "info", projectId)
		publicIPaddress, err := cloud.createPublicIp(pool, resourceGroup, IPname, ctx)
		if err != nil {
			return nil, "", err
		}
		utils.SendLog("Public IP created successfully : "+IPname, "info", projectId)
		cloud.Resources[projectId+IPname] = IPname
		/*
			making network interface
		*/
		nicName := "NIC-" + pool.Name
		utils.SendLog("Creating NIC : "+nicName, "info", projectId)
		nicParameters, err := cloud.createNIC(pool, resourceGroup, publicIPaddress, subnetId, sgIds, nicName, ctx)
		if err != nil {
			return nil, "", err
		}
		utils.SendLog("NIC created successfully : "+nicName, "info", projectId)
		cloud.Resources[projectId+nicName] = nicName

		utils.SendLog("Creating node  : "+pool.Name, "info", projectId)
		vm, private_key, _, err := cloud.createVM(pool, poolIndex, nicParameters, resourceGroup, ctx)
		if err != nil {
			return nil, "", err
		}
		utils.SendLog("Node created successfully : "+pool.Name, "info", projectId)
		cloud.Resources["Disk-"+pool.Name] = pool.Name
		cloud.Resources["NodeName-"+pool.Name] = pool.Name

		var vmObj VM
		vmObj.Name = vm.Name
		vmObj.CloudId = vm.ID
		vmObj.PrivateIP = (*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PrivateIPAddress
		vmObj.PublicIP = publicIPaddress.PublicIPAddressPropertiesFormat.IPAddress
		vmObj.NodeState = vm.VirtualMachineProperties.ProvisioningState
		vmObj.UserName = vm.VirtualMachineProperties.OsProfile.AdminUsername
		vmObj.PAssword = vm.VirtualMachineProperties.OsProfile.AdminPassword

		cpVms = append(cpVms, &vmObj)

		return cpVms, private_key, nil

	} else {
		vms, err, private_key := cloud.createVMSS(resourceGroup, projectId, pool, poolIndex, subnetId, sgIds, ctx)
		if err != nil {
			return nil, "", err
		}
		cloud.Resources["vmss-"+pool.Name] = pool.Name
		for _, vm := range vms.Values() {
			var vmObj VM
			vmObj.Name = vm.Name
			vmObj.CloudId = vm.ID
			nicId := ""
			for _, nic := range *vm.NetworkProfile.NetworkInterfaces {
				nicId = *nic.ID
				break
			}
			beego.Info(nicId)
			arr := strings.Split(nicId, "/")
			nicName := arr[12]
			beego.Info(nicName)
			beego.Info(arr[10])
			nicParameters, err := cloud.GetNIC(resourceGroup, projectId+strconv.Itoa(poolIndex), arr[10], nicName, ctx)
			if err != nil {
				return nil, "", err
			}
			vmObj.PrivateIP = (*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PrivateIPAddress
			pipId := (*(*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PublicIPAddress.ID)
			arr = strings.Split(pipId, "/")
			pipConf := arr[14]
			pipAddress := arr[16]
			beego.Info(pipId)
			beego.Info(arr[10])
			pip, err := cloud.GetPIP(resourceGroup, projectId+strconv.Itoa(poolIndex), arr[10], nicName, pipConf, pipAddress, ctx)
			if err != nil {
				return nil, "", err
			}
			vmObj.PublicIP = pip.IPAddress
			vmObj.NodeState = vm.ProvisioningState
			vmObj.UserName = vm.OsProfile.AdminUsername
			vmObj.PAssword = vm.OsProfile.AdminPassword

			cpVms = append(cpVms, &vmObj)

		}

		return cpVms, private_key, nil
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

func (cloud *AZURE) fetchStatus(cluster Cluster_Def, ctx utils.Context) (Cluster_Def, error) {
	if cloud.Authorizer == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendSDLog("Cluster model: Status - Failed to get lastest status "+err.Error(), "error")
			return Cluster_Def{}, err
		}
	}
	var cpVms []*VM
	for in, pool := range cluster.NodePools {
		var keyInfo utils.Key
		if pool.KeyInfo.CredentialType == models.SSHKey {
			k1, err := vault.GetAzureSSHKey("azure", pool.KeyInfo.KeyName, ctx)
			if err != nil {
				ctx.SendSDLog(err.Error(), "error")
				return Cluster_Def{}, err
			}
			keyInfo, err = key_utils.KeyConversion(k1, ctx)
			if err != nil {
				return Cluster_Def{}, err
			}
		}
		pool.KeyInfo = keyInfo
		if pool.PoolRole == "master" {

			beego.Info("getting instance")
			vm, err := cloud.GetInstance(pool.Name, cluster.ResourceGroup, ctx)
			if err != nil {
				ctx.SendSDLog(err.Error(), "error")
				return Cluster_Def{}, err
			}
			beego.Info("getting nic")
			nicName := "NIC-" + pool.Name
			nicParameters, err := cloud.GetVMNIC(cluster.ResourceGroup, nicName, ctx)
			if err != nil {
				ctx.SendSDLog(err.Error(), "error")
				return Cluster_Def{}, err
			}
			beego.Info("getting pip")
			IPname := "pip-" + pool.Name
			publicIPaddress, err := cloud.GetVMSSPIP(cluster.ResourceGroup, IPname, ctx)
			if err != nil {
				ctx.SendSDLog(err.Error(), "error")
				return Cluster_Def{}, err
			}

			var vmObj VM
			vmObj.Name = vm.Name
			vmObj.CloudId = vm.ID
			vmObj.PrivateIP = (*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PrivateIPAddress
			vmObj.PublicIP = publicIPaddress.PublicIPAddressPropertiesFormat.IPAddress
			vmObj.NodeState = vm.ProvisioningState
			vmObj.UserName = vm.OsProfile.AdminUsername
			vmObj.PAssword = vm.OsProfile.AdminPassword

			//cpVms = append(cpVms, &vmObj)
			beego.Info("updated node pool")
			cluster.NodePools[in].Nodes = ([]*VM{&vmObj})

		} else {
			vms, err := cloud.VMSSVMClient.List(cloud.context, cluster.ResourceGroup, pool.Name, "", "", "")
			if err != nil {
				ctx.SendSDLog(err.Error(), "error")
				return Cluster_Def{}, err
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
				nicParameters, err := cloud.GetNIC(cluster.ResourceGroup, cluster.ProjectId+strconv.Itoa(in), arr[10], nicName, ctx)
				if err != nil {
					return Cluster_Def{}, err
				}
				vmObj.PrivateIP = (*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PrivateIPAddress
				pipId := (*(*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PublicIPAddress.ID)
				arr = strings.Split(pipId, "/")
				pipConf := arr[14]
				pipAddress := arr[16]
				pip, err := cloud.GetPIP(cluster.ResourceGroup, cluster.ProjectId+strconv.Itoa(in), arr[10], nicName, pipConf, pipAddress, ctx)
				if err != nil {
					return Cluster_Def{}, err
				}
				vmObj.PublicIP = pip.IPAddress
				vmObj.NodeState = vm.ProvisioningState
				vmObj.UserName = vm.OsProfile.AdminUsername
				vmObj.PAssword = vm.OsProfile.AdminPassword

				cpVms = append(cpVms, &vmObj)

			}

			beego.Info("updated node pool")
			cluster.NodePools[in].Nodes = cpVms
		}
	}
	beego.Info("updated cluster")
	return cluster, nil
}

func (cloud *AZURE) GetInstance(name string, resourceGroup string, ctx utils.Context) (compute.VirtualMachine, error) {

	vm, err := cloud.VMClient.Get(cloud.context, resourceGroup, name, compute.InstanceView)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return compute.VirtualMachine{}, err
	}
	return vm, nil
}
func (cloud *AZURE) GetNIC(resourceGroup, vmss, vm, nicName string, ctx utils.Context) (network.Interface, error) {
	nicParameters, err := cloud.InterfacesClient.GetVirtualMachineScaleSetNetworkInterface(cloud.context, resourceGroup, vmss, vm, nicName, "")
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return network.Interface{}, err
	}
	return nicParameters, nil
}
func (cloud *AZURE) GetPIP(resourceGroup, vmss, vm, nic, ipConfig, ipAddress string, ctx utils.Context) (network.PublicIPAddress, error) {
	publicIPaddress, err := cloud.AddressClient.GetVirtualMachineScaleSetPublicIPAddress(cloud.context, resourceGroup, vmss, vm, nic, ipConfig, ipAddress, "")
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return network.PublicIPAddress{}, err
	}
	return publicIPaddress, nil
}
func (cloud *AZURE) GetVMNIC(resourceGroup, nicName string, ctx utils.Context) (network.Interface, error) {
	nicParameters, err := cloud.InterfacesClient.Get(cloud.context, resourceGroup, nicName, "")
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return network.Interface{}, err
	}
	return nicParameters, nil
}
func (cloud *AZURE) GetVMSSPIP(resourceGroup, IPname string, ctx utils.Context) (network.PublicIPAddress, error) {
	publicIPaddress, err := cloud.AddressClient.Get(cloud.context, resourceGroup, IPname, "")
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return network.PublicIPAddress{}, err
	}
	return publicIPaddress, nil
}
func (cloud *AZURE) terminateCluster(cluster Cluster_Def, ctx utils.Context) error {
	if cloud.Authorizer == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendSDLog(err.Error(), "error")
			return err
		}
	}
	utils.SendLog("Terminating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	for poolIndex, pool := range cluster.NodePools {

		utils.SendLog("Terminating node pool: "+pool.Name, "info", cluster.ProjectId)
		if pool.PoolRole == "master" {

			utils.SendLog("Terminating node pool: "+pool.Name, "info", cluster.ProjectId)
			err := cloud.TerminateMasterNode(*pool.Nodes[0].Name, cluster.ProjectId, cluster.ResourceGroup, ctx)
			if err != nil {
				return err
			}

			nicName := "NIC-" + pool.Name
			err = cloud.deleteNIC(nicName, cluster.ResourceGroup, cluster.ProjectId, ctx)
			if err != nil {
				return err
			}

			IPname := "pip-" + pool.Name
			err = cloud.deletePublicIp(IPname, cluster.ResourceGroup, cluster.ProjectId, ctx)
			if err != nil {
				return err
			}
			err = cloud.deleteStorageAccount(cluster.ResourceGroup, pool.Name, ctx)
			if err != nil {
				return err
			}
			err = cloud.deleteDisk(cluster.ResourceGroup, pool.Name, ctx)
			if err != nil {
				return err
			}

		} else {
			err := cloud.TerminatePool(pool.Name, cluster.ResourceGroup, cluster.ProjectId, ctx)
			if err != nil {
				return err
			}

			err = cloud.deleteStorageAccount(cluster.ResourceGroup, cluster.ProjectId+strconv.Itoa(poolIndex), ctx)
			if err != nil {
				return err
			}
		}
		utils.SendLog("Node Pool terminated successfully: "+pool.Name, "info", cluster.ProjectId)
	}
	return nil
}
func (cloud *AZURE) TerminatePool(name string, resourceGroup string, projectId string, ctx utils.Context) error {

	ctx.SendSDLog("AZUREOperations: terminating node pools", "info")

	future, err := cloud.VMSSCLient.Delete(cloud.context, resourceGroup, name)

	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return err
	} else {
		err = future.WaitForCompletion(cloud.context, cloud.VMSSCLient.Client)
		if err != nil {
			beego.Error("vm deletion failed")
			ctx.SendSDLog(err.Error(), "error")
			return err
		}
	}
	ctx.SendSDLog("Node pool terminated successfully: "+name, "info")

	return nil
}
func (cloud *AZURE) TerminateMasterNode(name, projectId, resourceGroup string, ctx utils.Context) error {

	beego.Info("AZUREOperations: terminating nodes")
	ctx.SendSDLog("Terminating node: "+name, "info")
	vmClient := compute.NewVirtualMachinesClient(cloud.Subscription)
	vmClient.Authorizer = cloud.Authorizer
	future, err := vmClient.Delete(cloud.context, resourceGroup, name)

	if err != nil {
		beego.Error(err)
		return err
	} else {
		err = future.WaitForCompletion(cloud.context, vmClient.Client)
		if err != nil {
			beego.Error("vm deletion failed")
			beego.Error(err)
			return err
		}
		beego.Info("Deleted Node" + name)
	}
	ctx.SendSDLog("Node terminated successfully: "+name, "info")
	return nil
}

func (cloud *AZURE) createPublicIp(pool *NodePool, resourceGroup string, IPname string, ctx utils.Context) (network.PublicIPAddress, error) {

	pipParameters := network.PublicIPAddress{
		Location: &cloud.Region,
		PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
			DNSSettings: &network.PublicIPAddressDNSSettings{
				DomainNameLabel: to.StringPtr(strings.ToLower(IPname)),
			},
		},
	}

	address, err := cloud.AddressClient.CreateOrUpdate(cloud.context, resourceGroup, IPname, pipParameters)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return network.PublicIPAddress{}, err
	} else {
		err = address.WaitForCompletionRef(cloud.context, cloud.AddressClient.Client)
		if err != nil {
			ctx.SendSDLog(err.Error(), "error")
			return network.PublicIPAddress{}, err
		}
	}

	ctx.SendSDLog("Get public IP address info...", "info")
	publicIPaddress, err := cloud.GetVMSSPIP(resourceGroup, IPname, ctx)
	return publicIPaddress, err
}

func (cloud *AZURE) deletePublicIp(IPname, resourceGroup string, projectId string, ctx utils.Context) error {
	utils.SendLog("Deleting Public IP: "+IPname, "info", projectId)
	address, err := cloud.AddressClient.Delete(cloud.context, resourceGroup, IPname)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return err
	} else {
		err = address.WaitForCompletionRef(cloud.context, cloud.AddressClient.Client)
		if err != nil {
			ctx.SendSDLog(err.Error(), "error")
			return err
		}
	}
	utils.SendLog("Public IP delete successfully: "+IPname, "info", projectId)
	return nil
}
func (cloud *AZURE) createNIC(pool *NodePool, resourceGroup string, publicIPaddress network.PublicIPAddress, subnetId string, sgIds []*string, nicName string, ctx utils.Context) (network.Interface, error) {

	nicParameters := network.Interface{
		Location: &cloud.Region,
		InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
			IPConfigurations: &[]network.InterfaceIPConfiguration{
				{
					Name: to.StringPtr(fmt.Sprintf("IPconfig-" + pool.Name)),
					InterfaceIPConfigurationPropertiesFormat: &network.InterfaceIPConfigurationPropertiesFormat{
						PrivateIPAllocationMethod: network.Dynamic,
						Subnet:          &network.Subnet{ID: to.StringPtr(subnetId)},
						PublicIPAddress: &publicIPaddress,
					},
				},
			},
		},
	}
	if sgIds != nil {
		nicParameters.InterfacePropertiesFormat.NetworkSecurityGroup = &network.SecurityGroup{
			ID: (sgIds[0]),
		}
	}

	future, err := cloud.InterfacesClient.CreateOrUpdate(cloud.context, resourceGroup, nicName, nicParameters)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return network.Interface{}, err
	} else {
		err := future.WaitForCompletion(cloud.context, cloud.InterfacesClient.Client)
		if err != nil {
			ctx.SendSDLog(err.Error(), "error")
			return network.Interface{}, err
		}
	}

	nicParameters, err = cloud.GetVMNIC(resourceGroup, nicName, ctx)
	return nicParameters, nil
}
func (cloud *AZURE) deleteNIC(nicName, resourceGroup string, proId string, ctx utils.Context) error {
	utils.SendLog("Deleting NIC: "+nicName, "info", proId)
	future, err := cloud.InterfacesClient.Delete(cloud.context, resourceGroup, nicName)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return err
	} else {
		err := future.WaitForCompletion(cloud.context, cloud.InterfacesClient.Client)
		if err != nil {
			ctx.SendSDLog(err.Error(), "error")
			return err
		}
	}
	utils.SendLog("NIC delete successfully: "+nicName, "info", proId)
	return nil
}

func (cloud *AZURE) createVM(pool *NodePool, index int, nicParameters network.Interface, resourceGroup string, ctx utils.Context) (compute.VirtualMachine, string, string, error) {
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
	storage = append(storage, disk)

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
		vm.StorageProfile.DataDisks = &storage
	}
	private := ""
	public := ""
	if pool.KeyInfo.CredentialType == models.SSHKey && pool.KeyInfo.NewKey == models.NEWKey {
		k, err := vault.GetAzureSSHKey("azure", pool.KeyInfo.KeyName, ctx)

		if err != nil && err.Error() != "not found" {
			ctx.SendSDLog("vm creation failed", "error")
			ctx.SendSDLog(err.Error(), "error")
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

			res, err := key_utils.GenerateKeyPair(pool.KeyInfo.KeyName, ctx)
			if err != nil {
				ctx.SendSDLog(err.Error(), "error")
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

			_, err = vault.PostAzureSSHKey(pool.KeyInfo, ctx)
			if err != nil {
				ctx.SendSDLog(err.Error(), "error")
				return compute.VirtualMachine{}, "", "", err
			}

			public = res.PublicKey
			private = res.PrivateKey
		}

	} else if pool.KeyInfo.CredentialType == models.SSHKey && pool.KeyInfo.NewKey == models.CPKey {

		k, err := vault.GetAzureSSHKey("azure", pool.KeyInfo.KeyName, ctx)
		if err != nil {
			ctx.SendSDLog(err.Error(), "error")
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
				ctx.SendSDLog(err.Error(), "error")
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
		ctx.SendSDLog(err.Error(), "error")
		return compute.VirtualMachine{}, "", "", err
	} else {
		err = vmFuture.WaitForCompletion(cloud.context, cloud.VMClient.Client)
		if err != nil {
			ctx.SendSDLog(err.Error(), "error")
			return compute.VirtualMachine{}, "", "", err
		}
	}
	beego.Info("Get VM  by name", pool.Name)
	vm, err = cloud.GetInstance(pool.Name, resourceGroup, ctx)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return compute.VirtualMachine{}, "", "", err
	}
	return vm, private, public, nil
}
func (cloud *AZURE) createStorageAccount(resouceGroup string, acccountName string, ctx utils.Context) error {
	accountParameters := storage.AccountCreateParameters{
		Sku: &storage.Sku{
			Name: storage.StandardLRS,
		},
		Location: &cloud.Region,
		AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{},
	}
	acccountName = strings.ToLower(acccountName)
	future, err := cloud.AccountClient.Create(context.Background(), resouceGroup, acccountName, accountParameters)
	if err != nil {
		beego.Error("Storage account creation failed")
		beego.Info(err)
		return err
	}
	err = future.WaitForCompletion(context.Background(), cloud.AccountClient.Client)
	if err != nil {

		beego.Error("Storage account creation failed")
		beego.Info(err)
		return err
	}
	/*account, err := cloud.AccountClient.GetProperties(cloud.context, resouceGroup, acccountName)
	if err != nil {
		beego.Error(err.Error())
		return "", err
	}
	beego.Info(*account.ID)*/
	return nil
}
func (cloud *AZURE) deleteDisk(resouceGroup string, diskName string, ctx utils.Context) error {

	_, err := cloud.AccountClient.Delete(context.Background(), resouceGroup, diskName)
	if err != nil {
		beego.Error("Disk deletion failed")
		ctx.SendSDLog(err.Error(), "error")
		return err
	}
	return nil
}
func (cloud *AZURE) deleteStorageAccount(resouceGroup string, acccountName string, ctx utils.Context) error {

	acccountName = strings.ToLower(acccountName)
	_, err := cloud.AccountClient.Delete(context.Background(), resouceGroup, acccountName)
	if err != nil {
		beego.Error("Storage account deletion failed")
		ctx.SendSDLog(err.Error(), "error")
		return err
	}
	return nil
}
func (cloud *AZURE) CleanUp(cluster Cluster_Def, ctx utils.Context) error {
	for _, pool := range cluster.NodePools {
		if pool.PoolRole == "master" {
			if cloud.Resources["NodeName-"+pool.Name] != nil {
				name := cloud.Resources["NodeName-"+pool.Name]
				nodeName := ""
				b, e := json.Marshal(name)
				if e != nil {
					beego.Info(e.Error())
					return e
				}
				e = json.Unmarshal(b, &nodeName)
				if e != nil {
					beego.Info(e.Error())
					return e
				}

				err := cloud.TerminateMasterNode(nodeName, cluster.ProjectId, cluster.ResourceGroup, ctx)
				if err != nil {
					beego.Info(e.Error())
					return err
				}
			}
			if cloud.Resources[cluster.ProjectId+pool.Name] != nil {
				name := cloud.Resources[cluster.ProjectId+pool.Name]
				nicName := ""
				b, e := json.Marshal(name)
				if e != nil {
					beego.Info(e.Error())
					return e
				}
				e = json.Unmarshal(b, &nicName)
				if e != nil {
					beego.Info(e.Error())
					return e
				}
				err := cloud.deleteNIC(nicName, cluster.ResourceGroup, cluster.ProjectId, ctx)
				if err != nil {
					beego.Info(e.Error())
					return err
				}
			}
			if cloud.Resources[cluster.ProjectId+pool.Name] != nil {
				name := cloud.Resources[cluster.ProjectId+pool.Name]
				IPname := ""
				b, e := json.Marshal(name)
				if e != nil {
					return e
				}
				e = json.Unmarshal(b, &IPname)
				if e != nil {
					return e
				}
				err := cloud.deletePublicIp(IPname, cluster.ResourceGroup, cluster.ProjectId, ctx)
				if err != nil {
					return err
				}
			}
			if cloud.Resources["SA-"+pool.Name] != nil {
				name := cloud.Resources["SA-"+pool.Name]
				SAname := ""
				b, e := json.Marshal(name)
				if e != nil {
					return e
				}
				e = json.Unmarshal(b, &SAname)
				if e != nil {
					return e
				}
				err := cloud.deleteStorageAccount(cluster.ResourceGroup, SAname, ctx)
				if err != nil {
					return err
				}
			}
			if cloud.Resources["Disk-"+pool.Name] != nil {
				name := cloud.Resources["Disk-"+pool.Name]
				diskName := ""
				b, e := json.Marshal(name)
				if e != nil {
					return e
				}
				e = json.Unmarshal(b, &diskName)
				if e != nil {
					return e
				}
				err := cloud.deleteDisk(cluster.ResourceGroup, diskName, ctx)
				if err != nil {
					return err
				}
			}
		} else {

			if cloud.Resources["vmss-"+pool.Name] != nil {
				name := cloud.Resources["vmss-"+pool.Name]
				vmssName := ""
				b, e := json.Marshal(name)
				if e != nil {
					return e
				}
				e = json.Unmarshal(b, &vmssName)
				if e != nil {
					return e
				}
				beego.Info(vmssName)
				err := cloud.TerminatePool(vmssName, cluster.ResourceGroup, cluster.ProjectId, ctx)
				if err != nil {
					return err
				}
			}

			if cloud.Resources["SA-"+pool.Name] != nil {
				name := cloud.Resources["SA-"+pool.Name]
				SAname := ""
				b, e := json.Marshal(name)
				if e != nil {
					return e
				}
				e = json.Unmarshal(b, &SAname)
				if e != nil {
					return e
				}
				err := cloud.deleteStorageAccount(cluster.ResourceGroup, SAname, ctx)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
func (cloud *AZURE) mountVolume(vms []*VM, privateKey string, KeyName string, projectId string, user string, resourceGroup string, poolName string, ctx utils.Context) error {

	for _, vm := range vms {
		err := fileWrite(privateKey, KeyName)
		if err != nil {
			return err
		}
		err = setPermission(KeyName)
		if err != nil {
			return err
		}

		if vm.PublicIP == nil {
			ctx.SendSDLog("waiting for public ip", "warning")
			time.Sleep(time.Second * 50)
			ctx.SendSDLog("waited for public ip", "warning")
			IPname := fmt.Sprintf("pip-%s", *vm.Name)
			beego.Info(IPname)
			publicIp, err := cloud.GetPIP(resourceGroup, poolName, *vm.Name, projectId+"Nic", projectId+"IpConfig", "pub", ctx)
			if err != nil {
				return err
			}
			vm.PublicIP = publicIp.IPAddress
		}

		start := time.Now()
		timeToWait := 60 //seconds
		retry := true
		var errCopy error

		for retry && int64(time.Since(start).Seconds()) < int64(timeToWait) {

			errCopy = copyFile(KeyName, user, *vm.PublicIP)
			if errCopy != nil && strings.Contains(errCopy.Error(), "exit status 1") {

				//ctx.SendSDLog(("time passed %6.2f sec\n"+ strconv.Itoa( int( time.Since(start).Seconds())))+"warning")
				ctx.SendSDLog("waiting 5 seconds before retry", "warning")
				time.Sleep(5 * time.Second)
			} else {
				retry = false
			}
		}
		if errCopy != nil {
			return errCopy
		}
		err = setScriptPermision(KeyName, user, *vm.PublicIP, ctx)
		if err != nil {
			return err
		}
		err = runScript(KeyName, user, *vm.PublicIP, ctx)
		if err != nil {
			return err
		}
		err = deleteScript(KeyName, user, *vm.PublicIP, ctx)
		if err != nil {
			return err
		}
		err = deleteFile(KeyName, ctx)
		if err != nil {
			return err
		}
	}
	return nil

}
func fileWrite(key string, keyName string) error {

	f, err := os.Create("../antelope/keys/" + keyName + ".pem")
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	defer f.Close()
	d2 := []byte(key)
	n2, err := f.Write(d2)
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	beego.Info("wrote %d bytes\n", n2)

	err = os.Chmod("../antelope/keys/"+keyName+".pem", 0777)
	if err != nil {
		beego.Error(err)
		return err
	}
	return nil
}
func setPermission(keyName string) error {
	keyPath := "../antelope/keys/" + keyName + ".pem"
	cmd1 := "chmod"
	beego.Info(keyPath)
	args := []string{"600", keyPath}
	cmd := exec.Command(cmd1, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	return nil
}
func copyFile(keyName string, userName string, instanceId string) error {

	keyPath := "../antelope/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId + ":/home/" + userName
	cmd1 := "scp"
	beego.Info(keyPath)
	beego.Info(ip)
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, "../antelope/scripts/azure-volume-mount.sh", ip}
	cmd := exec.Command(cmd1, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	return nil
}
func setScriptPermision(keyName string, userName string, instanceId string, ctx utils.Context) error {
	keyPath := "../antelope/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId
	cmd1 := "ssh"
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, ip, "chmod 700 /home/" + userName + "/azure-volume-mount.sh"}
	cmd := exec.Command(cmd1, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return nil
	}
	return nil
}
func runScript(keyName string, userName string, instanceId string, ctx utils.Context) error {
	keyPath := "../antelope/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId
	cmd1 := "ssh"
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, ip, "/home/" + userName + "/azure-volume-mount.sh"}
	cmd := exec.Command(cmd1, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return nil
	}
	return nil
}

func deleteScript(keyName string, userName string, instanceId string, ctx utils.Context) error {
	keyPath := "../antelope/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId
	cmd1 := "ssh"
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, ip, "rm", "/home/" + userName + "/azure-volume-mount.sh"}
	cmd := exec.Command(cmd1, args...)
	err := cmd.Run()
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return err
	}
	return nil
}

func deleteFile(keyName string, ctx utils.Context) error {
	keyPath := "../antelope/keys/" + keyName + ".pem"
	err := os.Remove(keyPath)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return err
	}
	return nil
}
func (cloud *AZURE) createVMSS(resourceGroup string, projectId string, pool *NodePool, poolIndex int, subnetId string, sgIds []*string, ctx utils.Context) (compute.VirtualMachineScaleSetVMListResultPage, error, string) {

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
		//Name:         to.StringPtr(projectId + "-" + strconv.Itoa(poolIndex)),
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

	storageName := "ext-" + pool.Name
	disk := compute.VirtualMachineScaleSetDataDisk{
		Lun:          to.Int32Ptr(int32(poolIndex)),
		Name:         to.StringPtr(storageName),
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
											PublicIPAddressConfiguration: &compute.VirtualMachineScaleSetPublicIPAddressConfiguration{
												Name: to.StringPtr("pip-" + pool.Name),
												VirtualMachineScaleSetPublicIPAddressConfigurationProperties: &compute.VirtualMachineScaleSetPublicIPAddressConfigurationProperties{
													DNSSettings: &compute.VirtualMachineScaleSetPublicIPAddressConfigurationDNSSettings{
														DomainNameLabel: to.StringPtr(pool.Name),
													},
												},
											},
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

	if pool.EnableVolume {
		params.VirtualMachineProfile.StorageProfile.DataDisks = &storage
	}

	private := ""
	// public := ""

	if pool.KeyInfo.CredentialType == models.SSHKey && pool.KeyInfo.NewKey == models.NEWKey {
		k, err := vault.GetAzureSSHKey("azure", pool.KeyInfo.KeyName, ctx)

		if err != nil && err.Error() != "not found" {
			ctx.SendSDLog(err.Error(), "error")
			return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
		} else if err == nil {

			existingKey, err := key_utils.KeyConversion(k, ctx)
			if err != nil {
				return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
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
		} else if err != nil && err.Error() == "not found" {

			res, err := key_utils.GenerateKeyPair(pool.KeyInfo.KeyName, ctx)
			if err != nil {
				ctx.SendSDLog(err.Error(), "error")
				return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
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
			params.VirtualMachineProfile.OsProfile.LinuxConfiguration = linux
			pool.KeyInfo.PublicKey = res.PublicKey
			pool.KeyInfo.PrivateKey = res.PrivateKey

			_, err = vault.PostAzureSSHKey(pool.KeyInfo, ctx)
			if err != nil {
				ctx.SendSDLog(err.Error(), "error")
				return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
			}

			//public = res.PublicKey
			private = res.PrivateKey
		}

	} else if pool.KeyInfo.CredentialType == models.SSHKey && pool.KeyInfo.NewKey == models.CPKey {

		k, err := vault.GetAzureSSHKey("azure", pool.KeyInfo.KeyName, ctx)
		if err != nil {
			ctx.SendSDLog(err.Error(), "error")
			return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
		}

		existingKey, err := key_utils.KeyConversion(k, ctx)
		if err != nil {
			return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
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
		params.VirtualMachineProfile.OsProfile.LinuxConfiguration = linux

		private = existingKey.PrivateKey
		//	public = existingKey.PublicKey
	} else {
		params.VirtualMachineProfile.OsProfile.AdminPassword = to.StringPtr(pool.KeyInfo.AdminPassword)
	}

	if pool.BootDiagnostics.Enable {

		if pool.BootDiagnostics.NewStroageAccount {
			sName := strings.Replace(pool.Name, "-", "", -1)
			sName = strings.ToLower(sName)
			storageId := "https://" + sName + ".blob.core.windows.net/"
			err := cloud.createStorageAccount(resourceGroup, sName, ctx)
			if err != nil {
				ctx.SendSDLog(err.Error(), "error")
				return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
			}
			params.VirtualMachineProfile.DiagnosticsProfile = &compute.DiagnosticsProfile{
				&compute.BootDiagnostics{
					Enabled: to.BoolPtr(true), StorageURI: &storageId,
				},
			}
			cloud.Resources["SA-"+pool.Name] = pool.Name
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
		ctx.SendSDLog(err.Error(), "error")
		return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
	} else {
		err = address.WaitForCompletionRef(cloud.context, cloud.AddressClient.Client)
		if err != nil {
			ctx.SendSDLog(err.Error(), "error")
			return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
		}
	}

	vms, err := cloud.VMSSVMClient.List(cloud.context, resourceGroup, pool.Name, "", "", "")
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
	}
	return vms, nil, private
}
