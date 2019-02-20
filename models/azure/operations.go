package azure

import (
	"antelope/models/logging"
	"antelope/models/networks"
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-02-01/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/astaxie/beego"
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
	return nil
}

func (cloud *AZURE) createCluster(cluster Cluster_Def) (Cluster_Def, error) {

	if cloud == nil {
		err := cloud.init()
		if err != nil {
			beego.Error(err.Error())
			return Cluster_Def{}, err
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
			return Cluster_Def{}, err
		}
		for j, vm := range result {
			cluster.NodePools[i].Nodes[j].VMs = vm
		}

	}

	return cluster, nil
}
func (cloud *AZURE) CreateInstance(pool *NodePool, networkData networks.AzureNetwork, resourceGroup string) ([]*compute.VirtualMachine, error) {

	var vms []*compute.VirtualMachine

	//subnetId := cloud.GetSubnets(pool, networkData)
	//sgIds := cloud.GetSecurityGroups(pool, networkData)

	subnetId := "/subscriptions/aa94b050-2c52-4b7b-9ce3-2ac18253e61e/resourceGroups/sadaf-test/providers/Microsoft.Network/virtualNetworks/sadaf-test-vnet/subnets/default"

	var sgIds []*string
	sid := "/subscriptions/aa94b050-2c52-4b7b-9ce3-2ac18253e61e/resourceGroups/sadaf-test/providers/Microsoft.Network/networkSecurityGroups/sadaf123-nsg"
	sgIds = append(sgIds, &sid)

	for index, node := range pool.Nodes {

		/*
			Making public ip
		*/
		IPname := fmt.Sprintf("pip-%s", pool.Name+"-"+strconv.Itoa(index))
		publicIPaddress, err := cloud.createPublicIp(pool, resourceGroup, IPname, index)
		if err != nil {
			return nil, err
		}
		/*
			making network interface
		*/
		nicParameters, err := cloud.createNIC(pool, index, resourceGroup, publicIPaddress, subnetId, sgIds)
		if err != nil {
			return nil, err
		}
		vm, err := cloud.createVM(pool, index, nicParameters, node, resourceGroup)
		if err != nil {
			return nil, err
		}
		vms = append(vms, &vm)
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
	return res, nil
}

type KeyPairResponse struct {
	Key_name   string `json:"key_name"`
	PrivateKey string `json:"privatekey"`
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

		for index, node := range pool.Nodes {

			beego.Info("Get VM  by name", *node.VMs.Name)

			vm, err := cloud.GetInstance(*node.VMs.Name, cluster.ResourceGroup)
			if err != nil {
				beego.Error(err)
				return Cluster_Def{}, err
			}
			cluster.NodePools[in].Nodes[index].VMs = &vm
		}

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

func (cloud *AZURE) terminateCluster(cluster Cluster_Def) error {
	if cloud.Authorizer == nil {
		err := cloud.init()
		if err != nil {
			beego.Error(err.Error())
			return err
		}
	}

	for _, pool := range cluster.NodePools {
		for _, node := range pool.Nodes {
			err := cloud.TerminatePool(node, cluster.ProjectId, cluster.ResourceGroup)
			if err != nil {
				return err
			}
		}
		logging.SendLog("Node Pool terminated successfully: "+pool.Name, "info", cluster.ProjectId)
	}
	return nil
}
func (cloud *AZURE) TerminatePool(node *Node, envId string, resourceGroup string) error {

	beego.Info("AZUREOperations: terminating nodes")

	vmClient := compute.NewVirtualMachinesClient(cloud.Subscription)
	vmClient.Authorizer = cloud.Authorizer
	future, err := vmClient.Delete(cloud.context, resourceGroup, *node.VMs.Name)

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
		beego.Info("Deleted Node" + *node.VMs.Name)
	}
	logging.SendLog("Node terminated successfully: "+*node.VMs.Name, "info", envId)
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
	publicIPaddress, err := cloud.AddressClient.Get(cloud.context, resourceGroup, IPname, "")
	if err != nil {
		beego.Info(err.Error())
		return network.PublicIPAddress{}, err
	}
	return publicIPaddress, nil
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

	nicParameters, err = cloud.InterfacesClient.Get(cloud.context, resourceGroup, nicName, "")
	if err != nil {
		beego.Info(err.Error())
		return network.Interface{}, err
	}
	return nicParameters, nil
}
func (cloud *AZURE) createVM(pool *NodePool, index int, nicParameters network.Interface, node *Node, resourceGroup string) (compute.VirtualMachine, error) {
	osDisk := &compute.OSDisk{
		CreateOption: compute.DiskCreateOptionTypesFromImage,
		Name:         to.StringPtr(pool.Name + "-" + strconv.Itoa(index)),
		ManagedDisk: &compute.ManagedDiskParameters{
			StorageAccountType: compute.StorageAccountTypesStandardSSDLRS,
		},
	}

	storageName := "ext-" + pool.Name + strconv.Itoa(index)
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
				AdminPassword: to.StringPtr(pool.AdminPassword),
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
	beego.Info("Get VM  by name", pool.Name+"-"+string(index))
	vm, err = cloud.GetInstance(pool.Name+"-"+string(index), resourceGroup)
	if err != nil {
		beego.Error(err)
		return compute.VirtualMachine{}, err
	}
	return vm, nil
}
