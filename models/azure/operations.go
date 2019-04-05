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

	cloud.VMClient = compute.NewVirtualMachinesClient(cloud.Subscription)
	cloud.VMClient.Authorizer = cloud.Authorizer

	cloud.AccountClient = storage.NewAccountsClient(cloud.Subscription)
	cloud.AccountClient.Authorizer = cloud.Authorizer

	cloud.DiskClient = compute.NewDisksClient(cloud.Subscription)
	cloud.DiskClient.Authorizer = cloud.Authorizer
	cloud.Resources = make(map[string]interface{})

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

	var azureNetwork networks.AzureNetwork
	network, err := networks.GetNetworkStatus(cluster.ProjectId, "azure")
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

		result, _, _, err := cloud.CreateInstance(pool, azureNetwork, cluster.ResourceGroup, cluster.ProjectId)
		if err != nil {
			beego.Error(err.Error())
			return cluster, err
		}

		cluster.NodePools[i].Nodes = result
		//	cluster.NodePools[i].KeyInfo.PublicKey = public
		//	cluster.NodePools[i].KeyInfo.PrivateKey = private
	}

	return cluster, nil
}
func (cloud *AZURE) CreateInstance(pool *NodePool, networkData networks.AzureNetwork, resourceGroup string, projectId string) ([]*VM, string, string, error) {

	var vms []*VM

	subnetId := cloud.GetSubnets(pool, networkData)
	sgIds := cloud.GetSecurityGroups(pool, networkData)

	/*subnetId := "/subscriptions/aa94b050-2c52-4b7b-9ce3-2ac18253e61e/resourceGroups/azureCluster/providers/Microsoft.Network/virtualNetworks/vnet-cloudNative/subnets/subnet-cloudNative"
	var sgIds []*string
	sid := "/subscriptions/aa94b050-2c52-4b7b-9ce3-2ac18253e61e/resourceGroups/azureCluster/providers/Microsoft.Network/networkSecurityGroups/sg-cloudNative"
	sgIds = append(sgIds, &sid)*/

	i := 0
	private := ""
	public := ""
	for i < int(pool.NodeCount) {

		/*
			Making public ip
		*/

		IPname := fmt.Sprintf("pip-%s", pool.Name+"-"+strconv.Itoa(i))
		logging.SendLog("Creating Public IP : "+IPname, "info", projectId)
		publicIPaddress, err := cloud.createPublicIp(pool, resourceGroup, IPname, i)
		if err != nil {
			return nil, "", "", err
		}
		logging.SendLog("Public IP created successfully : "+IPname, "info", projectId)
		cloud.Resources["IPName-"+strconv.Itoa(i)] = IPname
		/*
			making network interface
		*/
		nicName := fmt.Sprintf("NIC-%s", pool.Name+"-"+strconv.Itoa(i))
		logging.SendLog("Creating NIC : "+nicName, "info", projectId)
		nicParameters, err := cloud.createNIC(pool, i, resourceGroup, publicIPaddress, subnetId, sgIds)
		if err != nil {
			return nil, "", "", err
		}
		logging.SendLog("NIC created successfully : "+nicName, "info", projectId)
		cloud.Resources["NicName-"+strconv.Itoa(i)] = nicName

		name := pool.Name + "-" + strconv.Itoa(i)

		logging.SendLog("Creating node  : "+name, "info", projectId)
		vm, private, public, err := cloud.createVM(pool, i, nicParameters, resourceGroup)
		if err != nil {
			return nil, "", "", err
		}
		logging.SendLog("Node created successfully : "+name, "info", projectId)
		cloud.Resources["NodeName-"+strconv.Itoa(i)] = name

		var vmObj VM
		vmObj.Name = vm.Name
		vmObj.CloudId = vm.ID
		vmObj.PrivateIP = (*nicParameters.InterfacePropertiesFormat.IPConfigurations)[0].PrivateIPAddress
		vmObj.PublicIP = publicIPaddress.PublicIPAddressPropertiesFormat.IPAddress
		vmObj.NodeState = vm.VirtualMachineProperties.ProvisioningState
		vmObj.UserName = vm.VirtualMachineProperties.OsProfile.AdminUsername
		vmObj.PAssword = vm.VirtualMachineProperties.OsProfile.AdminPassword

		vms = append(vms, &vmObj)
		beego.Info(private)
		beego.Info(public)
		i = i + 1
	}
	return vms, private, public, nil

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
		for nodeIndex, n := range pool.Nodes {

			beego.Info("getting instance")
			vm, err := cloud.GetInstance(*n.Name, cluster.ResourceGroup)
			if err != nil {
				beego.Error(err)
				return Cluster_Def{}, err
			}
			beego.Info("getting nic")
			nicName := fmt.Sprintf("NIC-%s", pool.Name+"-"+strconv.Itoa(nodeIndex))
			nicParameters, err := cloud.GetNIC(cluster.ResourceGroup, nicName)
			if err != nil {
				beego.Error(err)
				return Cluster_Def{}, err
			}
			beego.Info("getting pip")
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

			beego.Info("updated node")
		}

		pool.KeyInfo = keyInfo

		beego.Info("updated node pool")
		cluster.NodePools[in] = pool
	}
	beego.Info("updated cluster")
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
	logging.SendLog("Terminating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	for _, pool := range cluster.NodePools {
		for index, node := range pool.Nodes {
			logging.SendLog("Terminating node pool: "+pool.Name, "info", cluster.ProjectId)
			err := cloud.TerminatePool(*node.Name, cluster.ProjectId, cluster.ResourceGroup)
			if err != nil {
				return err
			}

			nicName := fmt.Sprintf("NIC-%s", pool.Name+"-"+strconv.Itoa(index))
			err = cloud.deleteNIC(nicName, cluster.ResourceGroup, cluster.ProjectId)
			if err != nil {
				return err
			}

			IPname := fmt.Sprintf("pip-%s", pool.Name+"-"+strconv.Itoa(index))
			err = cloud.deletePublicIp(IPname, cluster.ResourceGroup, cluster.ProjectId)
			if err != nil {
				return err
			}

		}
		logging.SendLog("Node Pool terminated successfully: "+pool.Name, "info", cluster.ProjectId)
	}
	return nil
}
func (cloud *AZURE) TerminatePool(name string, projectId string, resourceGroup string) error {

	beego.Info("AZUREOperations: terminating nodes")
	logging.SendLog("Terminating node: "+name, "info", projectId)
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
	logging.SendLog("Node terminated successfully: "+name, "info", projectId)

	future_, err := cloud.DiskClient.Delete(cloud.context, resourceGroup, name)
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
		beego.Info("Deleted Disk" + name)
	}
	logging.SendLog("Disk deleted successfully: "+name, "info", projectId)

	future_1, err := cloud.DiskClient.Delete(cloud.context, resourceGroup, "ext-"+name)
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
		beego.Info("Deleted Disk: " + name)
	}
	logging.SendLog("Disk deleted successfully: "+"ext-"+name, "info", projectId)

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

func (cloud *AZURE) createVM(pool *NodePool, index int, nicParameters network.Interface, resourceGroup string) (compute.VirtualMachine, string, string, error) {
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
	private := ""
	public := ""
	if pool.KeyInfo.CredentialType == models.SSHKey && pool.KeyInfo.NewKey == models.NEWKey {
		k, err := vault.GetAzureSSHKey("azure", pool.KeyInfo.KeyName)

		if err != nil && err.Error() != "not found" {
			beego.Error("vm creation failed")
			beego.Error(err)
			return compute.VirtualMachine{}, "", "", err
		} else {

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
			} else {

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
func (cloud *AZURE) CleanUp(cluster Cluster_Def) error {
	for _, pool := range cluster.NodePools {
		i := 0
		for i <= int(pool.NodeCount) {
			if cloud.Resources["NodeName-"+strconv.Itoa(i)] != "" {
				name := cloud.Resources["NodeName-"+strconv.Itoa(i)]
				nodeName := ""
				b, e := json.Marshal(name)
				if e != nil {
					return e
				}
				e = json.Unmarshal(b, &nodeName)
				if e != nil {
					return e
				}

				err := cloud.TerminatePool(nodeName, cluster.ProjectId, cluster.ResourceGroup)
				if err != nil {
					return err
				}
			}
			if cloud.Resources["NicName-"+strconv.Itoa(i)] != "" {
				name := cloud.Resources["NicName-"+strconv.Itoa(i)]
				nicName := ""
				b, e := json.Marshal(name)
				if e != nil {
					return e
				}
				e = json.Unmarshal(b, &nicName)
				if e != nil {
					return e
				}
				err := cloud.deleteNIC(nicName, cluster.ResourceGroup, cluster.ProjectId)
				if err != nil {
					return err
				}
			}
			if cloud.Resources["IPName-"+strconv.Itoa(i)] != "" {
				name := cloud.Resources["IPName-"+strconv.Itoa(i)]
				IPname := ""
				b, e := json.Marshal(name)
				if e != nil {
					return e
				}
				e = json.Unmarshal(b, &IPname)
				if e != nil {
					return e
				}
				err := cloud.deletePublicIp(IPname, cluster.ResourceGroup, cluster.ProjectId)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
