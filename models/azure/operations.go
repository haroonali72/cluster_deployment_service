package azure

import (
	"antelope/models"
	"antelope/models/logging"
	"antelope/models/networks"
	"context"
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
	"github.com/aws/aws-sdk-go/aws"
	"io/ioutil"
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
	VMClient         compute.VirtualMachinesClient
	DiskClient       compute.DisksClient
	AccountClient    storage.AccountsClient
	context          context.Context
	ID               string
	Key              string
	Tenant           string
	Subscription     string
	Region           string
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

	cloud.VMClient = compute.NewVirtualMachinesClient(cloud.Subscription)
	cloud.VMClient.Authorizer = cloud.Authorizer

	cloud.AccountClient = storage.NewAccountsClient(cloud.Subscription)
	cloud.AccountClient.Authorizer = cloud.Authorizer
	return nil
}

func (cloud *AZURE) createCluster(cluster Cluster_Def) (Cluster_Def, error) {

	if cloud == nil {
		err := cloud.init()
		if err != nil {
			beego.Error(err.Error())
			return cluster, err
		}
	}
	/*
		var azureNetwork networks.AzureNetwork
		network, err := networks.GetNetworkStatus(cluster.ProjectId, "azure")
		bytes, err := json.Marshal(network)
		if err != nil {
			beego.Error(err.Error())
			return nil, err
		}

		err = json.Unmarshal(bytes, &azureNetwork)*/

	/*if err != nil {
		beego.Error(err.Error())
		return nil, err
	}*/

	for i, pool := range cluster.NodePools {

		beego.Info("AZUREOperations creating nodes")

		result, err := cloud.CreateInstance(pool, networks.AzureNetwork{}, cluster.ResourceGroup)
		if err != nil {
			logging.SendLog("Error in instances creation: "+err.Error(), "info", cluster.ProjectId)
			return cluster, err
		}

		cluster.NodePools[i].Nodes = result
	}

	return cluster, nil
}
func (cloud *AZURE) CreateInstance(pool *NodePool, networkData networks.AzureNetwork, resourceGroup string) ([]*VM, error) {

	var vms []*VM

	//subnetId := cloud.GetSubnets(pool, networkData)
	//sgIds := cloud.GetSecurityGroups(pool, networkData)

	subnetId := "/subscriptions/aa94b050-2c52-4b7b-9ce3-2ac18253e61e/resourceGroups/azureCluster/providers/Microsoft.Network/virtualNetworks/vnet-cloudNative/subnets/subnet-cloudNative"
	var sgIds []*string
	sid := "/subscriptions/aa94b050-2c52-4b7b-9ce3-2ac18253e61e/resourceGroups/azureCluster/providers/Microsoft.Network/networkSecurityGroups/sg-cloudNative"
	sgIds = append(sgIds, &sid)

	i := 0

	for i < int(pool.NodeCount) {

		/*
			Making public ip
		*/
		IPname := fmt.Sprintf("pip-%s", pool.Name+"-"+strconv.Itoa(i))
		publicIPaddress, err := cloud.createPublicIp(pool, resourceGroup, IPname, i)
		if err != nil {
			return nil, err
		}
		/*
			making network interface
		*/
		nicParameters, err := cloud.createNIC(pool, i, resourceGroup, publicIPaddress, subnetId, sgIds)
		if err != nil {
			return nil, err
		}
		vm, err := cloud.createVM(pool, i, nicParameters, resourceGroup)
		if err != nil {
			return nil, err
		}

		var vmObj VM
		vmObj.Name = vm.Name
		vmObj.CloudId = vm.ID
		vmObj.PrivateIP = (*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PrivateIPAddress
		vmObj.PublicIP = publicIPaddress.PublicIPAddressPropertiesFormat.IPAddress
		vmObj.NodeState = vm.VirtualMachineProperties.ProvisioningState
		vmObj.UserName = vm.VirtualMachineProperties.OsProfile.AdminUsername
		vmObj.PAssword = vm.VirtualMachineProperties.OsProfile.AdminPassword

		vms = append(vms, &vmObj)
		i = i + 1
	}
	return vms, nil

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

func (cloud *AZURE) fetchStatus(cluster Cluster_Def) (Cluster_Def, error) {
	if cloud.Authorizer == nil {
		err := cloud.init()
		if err != nil {
			beego.Error("Cluster model: Status - Failed to get lastest status ", err.Error())

			return Cluster_Def{}, err
		}
	}
	for in, pool := range cluster.NodePools {
		index := 0
		for index < int(pool.NodeCount) {
			for nodeIndex, n := range pool.Nodes {
				vm, err := cloud.GetInstance(*n.Name, cluster.ResourceGroup)
				if err != nil {
					beego.Error(err)
					return Cluster_Def{}, err
				}
				nicName := fmt.Sprintf("NIC-%s", pool.Name+"-"+strconv.Itoa(index))
				nicParameters, err := cloud.GetNIC(cluster.ResourceGroup, nicName)
				if err != nil {
					beego.Error(err)
					return Cluster_Def{}, err
				}
				IPname := fmt.Sprintf("pip-%s", *n.Name)
				publicIPaddress, err := cloud.GetPIP(cluster.ResourceGroup, IPname)
				if err != nil {
					beego.Error(err)
					return Cluster_Def{}, err
				}
				pool.Nodes[nodeIndex].Name = vm.Name
				pool.Nodes[nodeIndex].CloudId = vm.ID
				pool.Nodes[nodeIndex].PrivateIP = (*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PrivateIPAddress
				pool.Nodes[nodeIndex].PublicIP = publicIPaddress.PublicIPAddressPropertiesFormat.IPAddress

				pool.Nodes[nodeIndex].NodeState = vm.VirtualMachineProperties.ProvisioningState
				pool.Nodes[nodeIndex].UserName = vm.VirtualMachineProperties.OsProfile.AdminUsername
				pool.Nodes[nodeIndex].PAssword = vm.VirtualMachineProperties.OsProfile.AdminPassword
			}
		}
		cluster.NodePools[in] = pool
	}
	return cluster, nil
}
func (cloud *AZURE) GetInstance(name string, resourceGroup string) (compute.VirtualMachine, error) {

	vm, err := cloud.VMClient.Get(cloud.context, resourceGroup, name, compute.InstanceView)
	if err != nil {
		beego.Error(err)
		return compute.VirtualMachine{}, err
	}
	return vm, nil
}
func (cloud *AZURE) GetNIC(resourceGroup, nicName string) (network.Interface, error) {
	nicParameters, err := cloud.InterfacesClient.Get(cloud.context, resourceGroup, nicName, "")
	if err != nil {
		beego.Info(err.Error())
		return network.Interface{}, err
	}
	return nicParameters, nil
}
func (cloud *AZURE) GetPIP(resourceGroup, IPname string) (network.PublicIPAddress, error) {
	publicIPaddress, err := cloud.AddressClient.Get(cloud.context, resourceGroup, IPname, "")
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

	for _, pool := range cluster.NodePools {
		for index, node := range pool.Nodes {

			err := cloud.TerminatePool(node, cluster.ProjectId, cluster.ResourceGroup)
			if err != nil {
				return err
			}

			nicName := fmt.Sprintf("NIC-%s", pool.Name+"-"+strconv.Itoa(index))
			err = cloud.deleteNIC(nicName, cluster.ResourceGroup)
			if err != nil {
				return err
			}

			IPname := fmt.Sprintf("pip-%s", pool.Name+"-"+strconv.Itoa(index))
			err = cloud.deletePublicIp(IPname, cluster.ResourceGroup)
			if err != nil {
				return err
			}

		}
		logging.SendLog("Node Pool terminated successfully: "+pool.Name, "info", cluster.ProjectId)
	}
	return nil
}
func (cloud *AZURE) TerminatePool(node *VM, envId string, resourceGroup string) error {

	beego.Info("AZUREOperations: terminating nodes")

	vmClient := compute.NewVirtualMachinesClient(cloud.Subscription)
	vmClient.Authorizer = cloud.Authorizer
	future, err := vmClient.Delete(cloud.context, resourceGroup, *node.Name)

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
		beego.Info("Deleted Node" + *node.Name)
	}
	logging.SendLog("Node terminated successfully: "+*node.Name, "info", envId)

	future_, err := cloud.DiskClient.Delete(cloud.context, resourceGroup, *node.Name)
	if err != nil {
		beego.Error(err)
		return err
	} else {
		err = future_.WaitForCompletion(cloud.context, vmClient.Client)
		if err != nil {
			beego.Error("vm deletion failed")
			beego.Error(err)
			return err
		}
		beego.Info("Deleted Disk" + *node.Name)
	}
	logging.SendLog("Disk deleted successfully: "+*node.Name, "info", envId)

	future_1, err := cloud.DiskClient.Delete(cloud.context, resourceGroup, "ext-"+*node.Name)
	if err != nil {
		beego.Error(err)
		return err
	} else {
		err = future_1.WaitForCompletion(cloud.context, vmClient.Client)
		if err != nil {
			beego.Error("vm deletion failed")
			beego.Error(err)
			return err
		}
		beego.Info("Deleted Disk" + *node.Name)
	}
	logging.SendLog("Disk deleted successfully: "+*node.Name, "info", envId)

	return nil
}
func (cloud *AZURE) createPublicIp(pool *NodePool, resourceGroup string, IPname string, index int) (network.PublicIPAddress, error) {

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
func (cloud *AZURE) deletePublicIp(IPname, resourceGroup string) error {

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
func (cloud *AZURE) deleteNIC(nicName, resourceGroup string) error {

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
	return nil
}

func (cloud *AZURE) createVM(pool *NodePool, index int, nicParameters network.Interface, resourceGroup string) (compute.VirtualMachine, error) {
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

	storageName := "ext-" + pool.Name + "-" + strconv.Itoa(index)
	disk := compute.DataDisk{
		Lun:          to.Int32Ptr(int32(index)),
		Name:         to.StringPtr(storageName),
		CreateOption: compute.DiskCreateOptionTypesEmpty,
		DiskSizeGB:   to.Int32Ptr(int32(1023)),
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
				OsDisk:    osDisk,
				DataDisks: &storage,
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

	if pool.KeyInfo.CredentialType == "SSH Key" && pool.KeyInfo.NewKey == models.NEWKey {
		res, err := cloud.GenerateKeyPair(pool.KeyInfo.KeyName)
		if err != nil {
			beego.Info(err.Error())
			return compute.VirtualMachine{}, err
		}
		key := []compute.SSHPublicKey{{

			KeyData: to.StringPtr(res.PublicKey),
		},
		}
		vm.OsProfile.LinuxConfiguration.SSH.PublicKeys = &key
		pool.KeyInfo.PublicKey = res.PublicKey
		pool.KeyInfo.PrivateKey = res.PrivateKey

		err = InsertSSHKeyPair(pool.KeyInfo)

		if err != nil {
			beego.Error("vm creation failed")
			beego.Error(err)
			return compute.VirtualMachine{}, err
		}

	} else if pool.KeyInfo.CredentialType == "SSH Key" && pool.KeyInfo.NewKey == models.CPKey {

		existingKey, err := GetSSHKeyPair(pool.KeyInfo.KeyName)
		if err != nil {
			beego.Error("vm creation failed")
			beego.Error(err)
			return compute.VirtualMachine{}, err
		}

		key := []compute.SSHPublicKey{{

			KeyData: to.StringPtr(existingKey.PublicKey),
		}}

		vm.OsProfile.LinuxConfiguration.SSH.PublicKeys = &key

		err = InsertSSHKeyPair(pool.KeyInfo)

		if err != nil {
			beego.Error("vm creation failed")
			beego.Error(err)
			return compute.VirtualMachine{}, err
		}
	} else {
		vm.OsProfile.AdminPassword = to.StringPtr(pool.KeyInfo.AdminPassword)
	}

	if pool.BootDiagnostics.Enable {

		if pool.BootDiagnostics.NewStroageAccount {

			storageId := "https://" + pool.Name + strconv.Itoa(index) + ".blob.core.windows.net/"
			cloud.createStorageAccount(resourceGroup, pool.Name+strconv.Itoa(index))
			vm.VirtualMachineProperties.DiagnosticsProfile = &compute.DiagnosticsProfile{
				&compute.BootDiagnostics{
					Enabled: aws.Bool(true), StorageURI: &storageId,
				},
			}
		} else {

			storageId := "https://" + pool.BootDiagnostics.StorageAccountId + ".blob.core.windows.net/"
			vm.VirtualMachineProperties.DiagnosticsProfile = &compute.DiagnosticsProfile{
				&compute.BootDiagnostics{
					Enabled: aws.Bool(true), StorageURI: &storageId,
				},
			}
		}
	}

	vmClient := compute.NewVirtualMachinesClient(cloud.Subscription)
	vmClient.Authorizer = cloud.Authorizer
	vmFuture, err := vmClient.CreateOrUpdate(cloud.context, resourceGroup, pool.Name+"-"+strconv.Itoa(index), vm)
	if err != nil {
		beego.Error(err)
		return compute.VirtualMachine{}, err
	} else {
		err = vmFuture.WaitForCompletion(cloud.context, vmClient.Client)
		if err != nil {
			beego.Error("vm creation failed")
			beego.Error(err)
			return compute.VirtualMachine{}, err
		}
	}
	beego.Info("Get VM  by name", pool.Name+"-"+strconv.Itoa(index))
	vm, err = cloud.GetInstance(pool.Name+"-"+strconv.Itoa(index), resourceGroup)
	if err != nil {
		beego.Error(err)
		return compute.VirtualMachine{}, err
	}
	return vm, nil
}
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
