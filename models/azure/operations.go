package azure

import (
	"antelope/models"
	"antelope/models/logging"
	"antelope/models/networks"
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
	"io/ioutil"
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

	cloud.Resources = make(map[string]interface{})

	return nil
}
func getNetworkHost() string {
	return beego.AppConfig.String("network_url")

}
func (cloud *AZURE) createCluster(cluster Cluster_Def) (Cluster_Def, error) {

	if cloud == nil {
		err := cloud.init()
		if err != nil {
			beego.Error(err.Error())
			return cluster, err
		}
	}

	var azureNetwork networks.AzureNetwork
	network, err := networks.GetAPIStatus(getNetworkHost(), cluster.ProjectId, "azure")
	bytes, err := json.Marshal(network)
	if err != nil {
		beego.Error(err.Error())
		return cluster, err
	}

	err = json.Unmarshal(bytes, &azureNetwork)

	if err != nil {
		beego.Error(err.Error())
		return cluster, err
	}

	for i, pool := range cluster.NodePools {

		beego.Info("AZUREOperations creating nodes")

		result, private_key, err := cloud.CreateInstance(pool, azureNetwork, cluster.ResourceGroup, cluster.ProjectId, i)
		if err != nil {
			beego.Error(err.Error())
			return cluster, err
		}
		beego.Info("private key")
		beego.Info(private_key)
		err = cloud.mountVolume(result, private_key, pool.KeyInfo.KeyName, cluster.ProjectId, pool.AdminUser, cluster.ResourceGroup, i)
		if err != nil {
			logging.SendLog("Error in volume mounting : "+err.Error(), "info", cluster.ProjectId)
			return cluster, err
		}
		cluster.NodePools[i].Nodes = result
	}

	return cluster, nil
}
func (cloud *AZURE) CreateInstance(pool *NodePool, networkData networks.AzureNetwork, resourceGroup string, projectId string, poolIndex int) ([]*VM, string, error) {

	var cpVms []*VM

	subnetId := cloud.GetSubnets(pool, networkData)
	sgIds := cloud.GetSecurityGroups(pool, networkData)

	private_key := ""

	vms, err, private_key := cloud.createVMSS(resourceGroup, projectId, pool, poolIndex, subnetId, sgIds)
	if err != nil {
		return nil, "", err
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
		nicParameters, err := cloud.GetNIC(resourceGroup, projectId+"-"+strconv.Itoa(poolIndex), *vm.Name, nicName)
		if err != nil {
			return nil, "", err
		}
		vmObj.PrivateIP = (*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PrivateIPAddress
		pipId := (*(*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PublicIPAddress.ID)
		arr = strings.Split(pipId, "/")
		pipConf := arr[14]
		pipAddress := arr[16]
		pip, err := cloud.GetPIP(resourceGroup, projectId+"-"+strconv.Itoa(poolIndex), *vm.Name, nicName, pipConf, pipAddress)
		if err != nil {
			return nil, "", err
		}
		vmObj.PublicIP = pip.IPAddress
		vmObj.NodeState = vm.ProvisioningState
		vmObj.UserName = vm.OsProfile.AdminUsername
		vmObj.PAssword = vm.OsProfile.AdminPassword

		cpVms = append(cpVms, &vmObj)

	}
	cloud.Resources["vmss-"+projectId+"-"+strconv.Itoa(poolIndex)] = projectId + "-" + strconv.Itoa(poolIndex)
	return cpVms, private_key, err

}
func (cloud *AZURE) GetSecurityGroups(pool *NodePool, network networks.AzureNetwork) []*string {
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
func (cloud *AZURE) GetSubnets(pool *NodePool, network networks.AzureNetwork) string {
	for _, definition := range network.Definition {
		for _, subnet := range definition.Subnets {
			if subnet.Name == pool.PoolSubnet {
				return subnet.SubnetId
			}
		}
	}
	return ""
}
func (cloud *AZURE) GenerateKeyPair(keyName string) (KeyPairResponse, error) {

	res := KeyPairResponse{}

	t := time.Now().Local()
	tstamp := t.Format("20060102150405")
	keyName = keyName + "_" + tstamp

	cmd := "ssh-keygen"
	args := []string{"-t", "rsa", "-b", "4096", "-C", "azure@example.com", "-f", keyName}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		beego.Error(err)
		return KeyPairResponse{}, err
	}
	beego.Info("Successfully generated sshkeys")

	arr, err1 := ioutil.ReadFile(keyName)
	str := string(arr)
	if err1 != nil {
		beego.Error(err1)
		return KeyPairResponse{}, err1
	}

	res.PrivateKey = str
	res.Key_name = keyName

	arr, err1 = ioutil.ReadFile(keyName + ".pub")
	str = string(arr)
	if err1 != nil {
		beego.Error(err1)
		return KeyPairResponse{}, err1
	}
	res.PublicKey = str
	return res, nil
}

type KeyPairResponse struct {
	Key_name   string `json:"key_name"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

func keyCoverstion(keyInfo interface{}) (Key, error) {
	b, e := json.Marshal(keyInfo)
	var k Key
	if e != nil {
		beego.Error(e)
		return Key{}, e
	}
	e = json.Unmarshal(b, &k)
	if e != nil {
		beego.Error(e)
		return Key{}, e
	}
	return k, nil
}
func (cloud *AZURE) fetchStatus(cluster Cluster_Def) (Cluster_Def, error) {
	if cloud.Authorizer == nil {
		err := cloud.init()
		if err != nil {
			beego.Error("Cluster model: Status - Failed to get lastest status ", err.Error())

			return Cluster_Def{}, err
		}
	}
	var cpVms []*VM
	for in, pool := range cluster.NodePools {
		k1, err := vault.GetAzureSSHKey("azure", pool.KeyInfo.KeyName)
		if err != nil {
			beego.Error(err)
			return Cluster_Def{}, err
		}
		keyInfo, err := keyCoverstion(k1)
		if err != nil {
			return Cluster_Def{}, err
		}

		pool.KeyInfo = keyInfo
		vms, err := cloud.VMSSVMClient.List(cloud.context, cluster.ResourceGroup, cluster.ProjectId+"-"+strconv.Itoa(in), "", "", "")
		if err != nil {
			beego.Error(err)
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
			nicParameters, err := cloud.GetNIC(cluster.ResourceGroup, cluster.ProjectId+"-"+strconv.Itoa(in), *vm.Name, nicName)
			if err != nil {
				return Cluster_Def{}, err
			}
			vmObj.PrivateIP = (*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PrivateIPAddress
			pipId := (*(*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PublicIPAddress.ID)
			arr = strings.Split(pipId, "/")
			pipConf := arr[14]
			pipAddress := arr[16]
			pip, err := cloud.GetPIP(cluster.ResourceGroup, cluster.ProjectId+"-"+strconv.Itoa(in), *vm.Name, nicName, pipConf, pipAddress)
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
	beego.Info("updated cluster")
	return cluster, nil
}

/*func (cloud *AZURE) GetInstance(name string, resourceGroup string) (compute.VirtualMachine, error) {

	vm, err := cloud.VMClient.Get(cloud.context, resourceGroup, name, compute.InstanceView)
	if err != nil {
		beego.Error(err)
		return compute.VirtualMachine{}, err
	}
	return vm, nil
}*/
func (cloud *AZURE) GetNIC(resourceGroup, vmss, vm, nicName string) (network.Interface, error) {
	nicParameters, err := cloud.InterfacesClient.GetVirtualMachineScaleSetNetworkInterface(cloud.context, resourceGroup, vmss, vm, nicName, "")
	if err != nil {
		beego.Info(err.Error())
		return network.Interface{}, err
	}
	return nicParameters, nil
}
func (cloud *AZURE) GetPIP(resourceGroup, vmss, vm, nic, ipConfig, ipAddress string) (network.PublicIPAddress, error) {
	publicIPaddress, err := cloud.AddressClient.GetVirtualMachineScaleSetPublicIPAddress(cloud.context, resourceGroup, vmss, vm, nic, ipConfig, ipAddress, "")
	if err != nil {
		beego.Error(err.Error())
		return network.PublicIPAddress{}, err
	}
	return publicIPaddress, nil
}
func (cloud *AZURE) terminateCluster(cluster Cluster_Def) error {
	if cloud.Authorizer == nil {
		err := cloud.init()
		if err != nil {
			beego.Error(err.Error())
			return err
		}
	}
	logging.SendLog("Terminating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	for poolIndex, pool := range cluster.NodePools {

		logging.SendLog("Terminating node pool: "+pool.Name, "info", cluster.ProjectId)
		err := cloud.TerminatePool(cluster.ProjectId, cluster.ResourceGroup, strconv.Itoa(poolIndex))
		if err != nil {
			return err
		}

		err = cloud.deleteStorageAccount(cluster.ResourceGroup, cluster.ProjectId+strconv.Itoa(poolIndex))
		if err != nil {
			return err
		}

		logging.SendLog("Node Pool terminated successfully: "+pool.Name, "info", cluster.ProjectId)
	}
	return nil
}
func (cloud *AZURE) TerminatePool(projectId string, resourceGroup string, poolIndex string) error {

	beego.Info("AZUREOperations: terminating node pools")

	future, err := cloud.VMSSCLient.Delete(cloud.context, resourceGroup, projectId+"-"+poolIndex)

	if err != nil {
		beego.Error(err)
		return err
	} else {
		err = future.WaitForCompletion(cloud.context, cloud.VMSSCLient.Client)
		if err != nil {
			beego.Error("vm deletion failed")
			beego.Error(err)
			return err
		}
	}
	logging.SendLog("Node pool terminated successfully: "+projectId+"-"+poolIndex, "info", projectId)

	return nil
}

/*func (cloud *AZURE) createPublicIp(pool *NodePool, resourceGroup string, IPname string, index int) (network.PublicIPAddress, error) {

	pipParameters := network.PublicIPAddress{
		Location: &cloud.Region,
		PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
			DNSSettings: &network.PublicIPAddressDNSSettings{
				DomainNameLabel: to.StringPtr(fmt.Sprintf("%s", strings.ToLower(pool.Name)+"-"+strconv.Itoa(index))),
			},
		},
	}

	address, err := cloud.AddressClient.CreateOrUpdate(cloud.context, resourceGroup, IPname, pipParameters)
	if err != nil {
		beego.Error(err)
		return network.PublicIPAddress{}, err
	} else {
		err = address.WaitForCompletionRef(cloud.context, cloud.AddressClient.Client)
		if err != nil {
			beego.Error(err)
			return network.PublicIPAddress{}, err
		}
	}

	beego.Info("Get public IP address info...")
	publicIPaddress, err := cloud.GetPIP(resourceGroup, IPname)
	return publicIPaddress, err
}

func (cloud *AZURE) deletePublicIp(IPname, resourceGroup string, projectId string) error {
	logging.SendLog("Deleting Public IP: "+IPname, "info", projectId)
	address, err := cloud.AddressClient.Delete(cloud.context, resourceGroup, IPname)
	if err != nil {
		beego.Error(err)
		return err
	} else {
		err = address.WaitForCompletionRef(cloud.context, cloud.AddressClient.Client)
		if err != nil {
			beego.Error(err)
			return err
		}
	}
	logging.SendLog("Public IP delete successfully: "+IPname, "info", projectId)
	return nil
}
func (cloud *AZURE) createNIC(pool *NodePool, index int, resourceGroup string, publicIPaddress network.PublicIPAddress, subnetId string, sgIds []*string) (network.Interface, error) {

	nicName := fmt.Sprintf("NIC-%s", pool.Name+"-"+strconv.Itoa(index))

	nicParameters := network.Interface{
		Location: &cloud.Region,
		InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
			IPConfigurations: &[]network.InterfaceIPConfiguration{
				{
					Name: to.StringPtr(fmt.Sprintf("IPconfig-%s", strings.ToLower(pool.Name)+"-"+strconv.Itoa(index))),
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
		beego.Info(err.Error())
		return network.Interface{}, err
	} else {
		err := future.WaitForCompletion(cloud.context, cloud.InterfacesClient.Client)
		if err != nil {
			beego.Info(err.Error())
			return network.Interface{}, err
		}
	}

	nicParameters, err = cloud.GetNIC(resourceGroup, nicName)
	return nicParameters, nil
}
func (cloud *AZURE) deleteNIC(nicName, resourceGroup string, proId string) error {
	logging.SendLog("Deleting NIC: "+nicName, "info", proId)
	future, err := cloud.InterfacesClient.Delete(cloud.context, resourceGroup, nicName)
	if err != nil {
		beego.Info(err.Error())
		return err
	} else {
		err := future.WaitForCompletion(cloud.context, cloud.InterfacesClient.Client)
		if err != nil {
			beego.Info(err.Error())
			return err
		}
	}
	logging.SendLog("NIC delete successfully: "+nicName, "info", proId)
	return nil
}
*/
/*func (cloud *AZURE) createVM(pool *NodePool, index int, nicParameters network.Interface, resourceGroup string) (compute.VirtualMachine, string, string, error) {
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
		Name:         to.StringPtr(pool.Name + "-" + strconv.Itoa(index)),
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

	storageName := "ext-" + pool.Name + "-" + strconv.Itoa(index)
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
		Name:     to.StringPtr(pool.Name + "-" + strconv.Itoa(index)),
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
	if pool.Volume.EnableVolume {
		vm.StorageProfile.DataDisks = &storage
	}
	private := ""
	public := ""
	if pool.KeyInfo.CredentialType == models.SSHKey && pool.KeyInfo.NewKey == models.NEWKey {
		k, err := vault.GetAzureSSHKey("azure", pool.KeyInfo.KeyName)

		if err != nil && err.Error() != "not found" {
			beego.Error("vm creation failed")
			beego.Error(err)
			return compute.VirtualMachine{}, "", "", err
		} else if err == nil {

			existingKey, err := keyCoverstion(k)
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

			res, err := cloud.GenerateKeyPair(pool.KeyInfo.KeyName)
			if err != nil {
				beego.Info(err.Error())
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

			_, err = vault.PostAzureSSHKey(pool.KeyInfo)
			if err != nil {
				beego.Error("vm creation failed")
				beego.Error(err)
				return compute.VirtualMachine{}, "", "", err
			}

			public = res.PublicKey
			private = res.PrivateKey
		}

	} else if pool.KeyInfo.CredentialType == models.SSHKey && pool.KeyInfo.NewKey == models.CPKey {

		k, err := vault.GetAzureSSHKey("azure", pool.KeyInfo.KeyName)
		if err != nil {
			beego.Error("vm creation failed")
			beego.Error(err)
			return compute.VirtualMachine{}, "", "", err
		}

		existingKey, err := keyCoverstion(k)
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

			storageId := "https://" + pool.Name + strconv.Itoa(index) + ".blob.core.windows.net/"
			err := cloud.createStorageAccount(resourceGroup, pool.Name+strconv.Itoa(index))
			if err != nil {
				beego.Error("vm creation failed")
				beego.Error(err)
				return compute.VirtualMachine{}, "", "", err
			}
			vm.VirtualMachineProperties.DiagnosticsProfile = &compute.DiagnosticsProfile{
				&compute.BootDiagnostics{
					Enabled: to.BoolPtr(true), StorageURI: &storageId,
				},
			}
			cloud.Resources[pool.Name+"-SA"+strconv.Itoa(index)] = pool.Name + strconv.Itoa(index)
		} else {

			storageId := "https://" + pool.BootDiagnostics.StorageAccountId + ".blob.core.windows.net/"
			vm.VirtualMachineProperties.DiagnosticsProfile = &compute.DiagnosticsProfile{
				&compute.BootDiagnostics{
					Enabled: to.BoolPtr(true), StorageURI: &storageId,
				},
			}
		}
	}
	vmFuture, err := cloud.VMClient.CreateOrUpdate(cloud.context, resourceGroup, pool.Name+"-"+strconv.Itoa(index), vm)
	if err != nil {
		beego.Error(err)
		return compute.VirtualMachine{}, "", "", err
	} else {
		err = vmFuture.WaitForCompletion(cloud.context, cloud.VMClient.Client)
		if err != nil {
			beego.Error("vm creation failed")
			beego.Error(err)
			return compute.VirtualMachine{}, "", "", err
		}
	}
	beego.Info("Get VM  by name", pool.Name+"-"+strconv.Itoa(index))
	vm, err = cloud.GetInstance(pool.Name+"-"+strconv.Itoa(index), resourceGroup)
	if err != nil {
		beego.Error(err)
		return compute.VirtualMachine{}, "", "", err
	}
	return vm, private, public, nil
}*/
func (cloud *AZURE) createStorageAccount(resouceGroup string, acccountName string) error {
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
func (cloud *AZURE) deleteStorageAccount(resouceGroup string, acccountName string) error {

	acccountName = strings.ToLower(acccountName)
	_, err := cloud.AccountClient.Delete(context.Background(), resouceGroup, acccountName)
	if err != nil {
		beego.Error("Storage account deletion failed")
		beego.Info(err)
		return err
	}
	return nil
}
func (cloud *AZURE) CleanUp(cluster Cluster_Def) error {
	for i, _ := range cluster.NodePools {
		if cloud.Resources["vmss-"+cluster.ProjectId+"-"+strconv.Itoa(i)] != nil {
			name := cloud.Resources["vmss-"+cluster.ProjectId+"-"+strconv.Itoa(i)]
			vmssName := ""
			b, e := json.Marshal(name)
			if e != nil {
				return e
			}
			e = json.Unmarshal(b, &vmssName)
			if e != nil {
				return e
			}

			err := cloud.TerminatePool(vmssName, cluster.ProjectId, cluster.ResourceGroup)
			if err != nil {
				return err
			}
		}

		if cloud.Resources[cluster.ProjectId+"-SA"+strconv.Itoa(i)] != nil {
			name := cloud.Resources[cluster.ProjectId+"-SA"+strconv.Itoa(i)]
			SAname := ""
			b, e := json.Marshal(name)
			if e != nil {
				return e
			}
			e = json.Unmarshal(b, &SAname)
			if e != nil {
				return e
			}
			err := cloud.deleteStorageAccount(cluster.ResourceGroup, SAname)
			if err != nil {
				return err
			}
		}
		i = i + 1

	}

	return nil
}
func (cloud *AZURE) mountVolume(vms []*VM, privateKey string, KeyName string, projectId string, user string, resourceGroup string, poolIndex int) error {

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
			beego.Info("waiting for public ip")
			time.Sleep(time.Second * 50)
			beego.Info("waited for public ip")
			IPname := fmt.Sprintf("pip-%s", *vm.Name)
			beego.Info(IPname)
			publicIp, err := cloud.GetPIP(resourceGroup, projectId+"-"+strconv.Itoa(poolIndex), *vm.Name, projectId+"Nic", projectId+"IpConfig", "pub")
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

				beego.Info("time passed %6.2f sec\n", time.Since(start).Seconds())
				beego.Info("waiting 5 seconds before retry")
				time.Sleep(5 * time.Second)
			} else {
				retry = false
			}
		}
		if errCopy != nil {
			return errCopy
		}
		err = setScriptPermision(KeyName, user, *vm.PublicIP)
		if err != nil {
			return err
		}
		err = runScript(KeyName, user, *vm.PublicIP)
		if err != nil {
			return err
		}
		err = deleteScript(KeyName, user, *vm.PublicIP)
		if err != nil {
			return err
		}
		err = deleteFile(KeyName)
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
func setScriptPermision(keyName string, userName string, instanceId string) error {
	keyPath := "../antelope/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId
	cmd1 := "ssh"
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, ip, "chmod 700 /home/" + userName + "/azure-volume-mount.sh"}
	cmd := exec.Command(cmd1, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		beego.Warn(err.Error())
		return nil
	}
	return nil
}
func runScript(keyName string, userName string, instanceId string) error {
	keyPath := "../antelope/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId
	cmd1 := "ssh"
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, ip, "/home/" + userName + "/azure-volume-mount.sh"}
	cmd := exec.Command(cmd1, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		beego.Warn(err.Error())
		return nil
	}
	return nil
}

func deleteScript(keyName string, userName string, instanceId string) error {
	keyPath := "../antelope/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId
	cmd1 := "ssh"
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, ip, "rm", "/home/" + userName + "/azure-volume-mount.sh"}
	cmd := exec.Command(cmd1, args...)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func deleteFile(keyName string) error {
	keyPath := "../antelope/keys/" + keyName + ".pem"
	err := os.Remove(keyPath)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}
func (cloud *AZURE) createVMSS(resourceGroup string, projectId string, pool *NodePool, poolIndex int, subnetId string, sgIds []*string) (compute.VirtualMachineScaleSetVMListResultPage, error, string) {

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
		Name:         to.StringPtr(projectId + "-" + strconv.Itoa(poolIndex)),
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

	storageName := "ext-" + projectId + "-" + strconv.Itoa(poolIndex)
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
		Name:     to.StringPtr(projectId + "-" + strconv.Itoa(poolIndex)),
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
							Name: to.StringPtr("nic-" + projectId + "-" + strconv.Itoa(poolIndex)),
							VirtualMachineScaleSetNetworkConfigurationProperties: &compute.VirtualMachineScaleSetNetworkConfigurationProperties{
								Primary: to.BoolPtr(true),
								IPConfigurations: &[]compute.VirtualMachineScaleSetIPConfiguration{
									{
										Name: to.StringPtr(pool.Name),
										VirtualMachineScaleSetIPConfigurationProperties: &compute.VirtualMachineScaleSetIPConfigurationProperties{
											Subnet: &compute.APIEntityReference{ID: to.StringPtr(subnetId)},
											PublicIPAddressConfiguration: &compute.VirtualMachineScaleSetPublicIPAddressConfiguration{
												Name: to.StringPtr("pip-" + projectId + "-" + strconv.Itoa(poolIndex)),
												VirtualMachineScaleSetPublicIPAddressConfigurationProperties: &compute.VirtualMachineScaleSetPublicIPAddressConfigurationProperties{
													DNSSettings: &compute.VirtualMachineScaleSetPublicIPAddressConfigurationDNSSettings{
														DomainNameLabel: to.StringPtr(fmt.Sprintf("%s", strings.ToLower(projectId)+"-"+strconv.Itoa(poolIndex))),
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

	if pool.Volume.EnableVolume {
		params.VirtualMachineProfile.StorageProfile.DataDisks = &storage
	}

	private := ""
	// public := ""

	if pool.KeyInfo.CredentialType == models.SSHKey && pool.KeyInfo.NewKey == models.NEWKey {
		k, err := vault.GetAzureSSHKey("azure", pool.KeyInfo.KeyName)

		if err != nil && err.Error() != "not found" {
			beego.Error("vm creation failed")
			beego.Error(err)
			return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
		} else if err == nil {

			existingKey, err := keyCoverstion(k)
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

			res, err := cloud.GenerateKeyPair(pool.KeyInfo.KeyName)
			if err != nil {
				beego.Info(err.Error())
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

			_, err = vault.PostAzureSSHKey(pool.KeyInfo)
			if err != nil {
				beego.Error("vm creation failed")
				beego.Error(err)
				return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
			}

			//public = res.PublicKey
			private = res.PrivateKey
		}

	} else if pool.KeyInfo.CredentialType == models.SSHKey && pool.KeyInfo.NewKey == models.CPKey {

		k, err := vault.GetAzureSSHKey("azure", pool.KeyInfo.KeyName)
		if err != nil {
			beego.Error("vm creation failed")
			beego.Error(err)
			return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
		}

		existingKey, err := keyCoverstion(k)
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

			storageId := "https://" + projectId + strconv.Itoa(poolIndex) + ".blob.core.windows.net/"
			err := cloud.createStorageAccount(resourceGroup, projectId+strconv.Itoa(poolIndex))
			if err != nil {
				beego.Error("vm creation failed")
				beego.Error(err)
				return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
			}
			params.VirtualMachineProfile.DiagnosticsProfile = &compute.DiagnosticsProfile{
				&compute.BootDiagnostics{
					Enabled: to.BoolPtr(true), StorageURI: &storageId,
				},
			}
			cloud.Resources[projectId+"-SA"+strconv.Itoa(poolIndex)] = projectId + strconv.Itoa(poolIndex)
		} else {

			storageId := "https://" + pool.BootDiagnostics.StorageAccountId + ".blob.core.windows.net/"
			params.VirtualMachineProfile.DiagnosticsProfile = &compute.DiagnosticsProfile{
				&compute.BootDiagnostics{
					Enabled: to.BoolPtr(true), StorageURI: &storageId,
				},
			}
		}
	}
	address, err := cloud.VMSSCLient.CreateOrUpdate(cloud.context, resourceGroup, projectId+"-"+strconv.Itoa(poolIndex), params)
	if err != nil {
		beego.Error(err)
		return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
	} else {
		err = address.WaitForCompletionRef(cloud.context, cloud.AddressClient.Client)
		if err != nil {
			beego.Error(err)
			return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
		}
	}

	vms, err := cloud.VMSSVMClient.List(cloud.context, resourceGroup, projectId+"-"+strconv.Itoa(poolIndex), "", "", "")
	if err != nil {
		beego.Error(err)
		return compute.VirtualMachineScaleSetVMListResultPage{}, err, ""
	}
	return vms, nil, private
}
