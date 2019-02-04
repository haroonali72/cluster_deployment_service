package azure

import (
	"github.com/astaxie/beego"
	"errors"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-02-01/network"
	"github.com/Azure/go-autorest/autorest/to"
	"context"
	"github.com/Azure/go-autorest/autorest"
	"antelope/models/logging"
	"strconv"
	"strings"
	"antelope/models/networks"
	"io/ioutil"
	"time"
	"os/exec"
	"fmt"
	"encoding/json"
)
type CreatedPool struct {
	Instances    []*compute.VirtualMachine
	KeyName    	 string
	Key     	 string
	PoolName 	 string
}
type AZURE struct {
	Authorizer		*autorest.BearerAuthorizer
	context         context.Context
	ID 				string
	Key 			string
	Tenant    		string
	Subscription    string
	Region 			string
}

func (cloud *AZURE) init() error {
	if cloud.Authorizer != nil {
		return nil
	}

	if cloud.ID == "" || cloud.Key == "" || cloud.Tenant == "" || cloud.Subscription== "" || cloud.Region == "" {
		text := "invalid cloud credentials"
		beego.Error(text)
		return errors.New(text)
	}

	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, cloud.Tenant)
	if err != nil {
		panic(err)
	}

	spt, err := adal.NewServicePrincipalToken(*oauthConfig, cloud.ID, cloud.Key, "resourceManagerEndpoint")
	if err != nil {
		cloud.Authorizer =&autorest.BearerAuthorizer{}
		return  err
	}
	cloud.context = context.Background()
	cloud.Authorizer = autorest.NewBearerAuthorizer(spt)

	return nil
}

func (cloud *AZURE) createCluster(cluster Cluster_Def ) ([]CreatedPool , error){

	if cloud == nil {
		err := cloud.init()
		if err != nil {
			beego.Error(err.Error())
			return nil ,err
		}
	}

	var azureNetwork networks.AzureNetwork
	network , err := networks.GetNetworkStatus(cluster.EnvironmentId,"azure")
	bytes, err := json.Marshal(network)
	if err != nil {
		beego.Error(err.Error())
		return nil ,err
	}

	err = json.Unmarshal(bytes,&azureNetwork )

	if err != nil {
		beego.Error(err.Error())
		return nil ,err
	}

	var createdPools []CreatedPool

	for _, pool := range cluster.NodePools {

		var createdPool CreatedPool

		beego.Info("AZUREOperations creating nodes")

		result, err :=  cloud.CreateInstance(pool,azureNetwork, cluster.ResourceGroup)
		if err != nil {
			logging.SendLog("Error in instances creation: " + err.Error(),"info",cluster.EnvironmentId)
			beego.Error(err.Error())
			return nil, err
		}

		createdPool.Instances= result
		createdPool.PoolName=pool.Name
		createdPools = append(createdPools,createdPool)
	}

	return createdPools,nil
}
func (cloud *AZURE) CreateInstance (pool *NodePool,networkData networks.AzureNetwork , resourceGroup string )([] *compute.VirtualMachine, error){

	var vms []*compute.VirtualMachine
	subnetId := cloud.GetSubnets(pool,networkData)
	sgIds := cloud.GetSecurityGroups(pool,networkData)

	for index, node := range pool.Nodes {

		/*
		Making public ip
		 */

		addressClient := network.NewPublicIPAddressesClient(cloud.Subscription)
		addressClient.Authorizer = cloud.Authorizer

		IPname := fmt.Sprintf("pip-%s", pool.Name+"-"+string(index))

		pipParameters := network.PublicIPAddress{
			Location: &cloud.Region,
			PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
				DNSSettings: &network.PublicIPAddressDNSSettings{
					DomainNameLabel: to.StringPtr(fmt.Sprintf("%s", strings.ToLower(pool.Name)+"-"+string(index))),
				},
			},
		}

		address, err := addressClient.CreateOrUpdate(cloud.context, resourceGroup, IPname, pipParameters)
		if err != nil {
			beego.Error(err)
			return nil, err
		} else {
			err = address.WaitForCompletion(cloud.context, addressClient.Client)
			if err != nil {
				beego.Error("Public Ip address creation failed")
				beego.Error(err)
				return nil, err
			}
		}

		beego.Info("Get public IP address info...")
		publicIPaddress, err := addressClient.Get(cloud.context, resourceGroup, IPname, "")
		if err != nil {
			beego.Error("Getting public ip failed")
			return nil, err
		}
		/*
		making network interface
		 */

		interfacesClient := network.NewInterfacesClient(cloud.Subscription)
		interfacesClient.Authorizer = cloud.Authorizer

		nicName := fmt.Sprintf("NIC-%s", pool.Name+"-"+string(index))

		nicParameters := network.Interface{
			Location: &cloud.Region,
			InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
				IPConfigurations: &[]network.InterfaceIPConfiguration{
					{
						Name: to.StringPtr(fmt.Sprintf("IPconfig-%s", strings.ToLower(pool.Name)+"-"+string(index))),
						InterfaceIPConfigurationPropertiesFormat: &network.InterfaceIPConfigurationPropertiesFormat{
							PrivateIPAllocationMethod: network.Dynamic,
							Subnet:                    &network.Subnet{ID: to.StringPtr(subnetId)},
							PublicIPAddress:           &publicIPaddress,
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

		future, err := interfacesClient.CreateOrUpdate(cloud.context, resourceGroup, nicName, nicParameters)
		if err != nil {
			return nil, err
		} else {
			err := future.WaitForCompletion(cloud.context, interfacesClient.Client)
			if err != nil {
				return nil, err
			}
		}

		nicParameters, err = interfacesClient.Get(context.Background(), resourceGroup, nicName, "")
		if err != nil {
			return nil, err
		}

		osDisk := &compute.OSDisk{
			CreateOption: compute.DiskCreateOptionTypesFromImage,
			Name:         to.StringPtr(pool.Name + "-" + string(index)),
			ManagedDisk: &compute.ManagedDiskParameters{
				StorageAccountType: compute.StorageAccountTypesStandardSSDLRS,
			},
		}

		storageName := "ext" + pool.Name + strconv.Itoa(index)
		disk := compute.DataDisk{
			Lun:          to.Int32Ptr(int32(index)),
			Name:         to.StringPtr(storageName),
			CreateOption: compute.DiskCreateOptionTypesFromImage,
			DiskSizeGB:   to.Int32Ptr(int32(1023)),
		}

		var storage= []compute.DataDisk{}
		storage = append(storage, disk)

		vm := compute.VirtualMachine{
			Name:     to.StringPtr(pool.Name + "-" + string(index)),
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
					AdminUsername: to.StringPtr(node.AdminUser),
					AdminPassword: to.StringPtr(node.AdminPassword),
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
		vmFuture, err := vmClient.CreateOrUpdate(cloud.context, resourceGroup, pool.Name+"-"+string(index), vm)
		if err != nil {
			beego.Error(err)
			return nil, err
		} else {
			err = vmFuture.WaitForCompletion(context.Background(), vmClient.Client)
			if err != nil {
				beego.Error("vm creation failed")
				beego.Error(err)
				return nil, err
			}
		}

		beego.Info("Get VM '%s' by name\n", pool.Name+"-"+string(index))
		vm, err = vmClient.Get(cloud.context, resourceGroup, pool.Name+"-"+string(index), compute.InstanceView)
		if err != nil {
			beego.Error(err)
			return nil, err
		}
		vms= append(vms , &vm)
	}
	return vms , nil

}
func (cloud *AZURE) GetSecurityGroups (pool *NodePool, network networks.AzureNetwork )([]*string) {
	var sgId []*string
	for _, definition := range network.Definition{
		for _, sg := range definition.SecurityGroups {
			for _, sgName := range  pool.PoolSecurityGroups{
				if *sgName ==  sg.Name{
					sgId = append(sgId, &sg.SecurityGroupId)
				}
			}
		}
	}
	return sgId
}
func (cloud *AZURE) GetSubnets (pool *NodePool, network networks.AzureNetwork )(string) {
	for _, definition := range network.Definition{
		for _, subnet := range definition.Subnets {
			if subnet.Name ==  pool.PoolSubnet{
				return subnet.SubnetId
			}
		}
	}
	return ""
}
func (cloud *AZURE) GenerateKeyPair(keyName string) (KeyPairResponse,error) {

	res := KeyPairResponse{}

	t := time.Now().Local()
	tstamp := t.Format("20060102150405")
	keyName =keyName + "_" + tstamp

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
	return res,nil
}

type KeyPairResponse struct {
	Key_name    string `json:"key_name"`
	PrivateKey  string `json:"privatekey"`
}
func (cloud *AZURE) fetchStatus(cluster Cluster_Def ) (Cluster_Def, error){
	if cloud.Authorizer == nil {
		err := cloud.init()
		if err != nil {
			beego.Error("Cluster model: Status - Failed to get lastest status ", err.Error())

			return Cluster_Def{},err
		}
	}
	for in, pool := range cluster.NodePools {

		for index, node :=range pool.Nodes {

			vm, err := cloud.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            GetInstance(node, cluster.ResourceGroup)
			if err != nil {
				return Cluster_Def{}, err
			}

			pool.Nodes[index].NodeState=*out.Reservations[0].Instances[0].State.Name

			if out.Reservations[0].Instances[0].PublicIpAddress != nil {

				pool.Nodes[index].PublicIP = *out.Reservations[0].Instances[0].PublicIpAddress
			}
			cluster.NodePools[in].Nodes[index].VMs=pool
		}

	}
	return cluster,nil
}
func (cloud *AZURE) GetInstance(node Node, resourceGroup string )(compute.VirtualMachine, error){

	beego.Info("Get VM '%s' by name\n", *node.VMs.Name)

	vmClient := compute.NewVirtualMachinesClient(cloud.Subscription)
	vmClient.Authorizer = cloud.Authorizer
	vm, err := vmClient.Get(cloud.context, resourceGroup, *node.VMs.Name, compute.InstanceView)
	if err != nil {
		beego.Error(err)
		return compute.VirtualMachine{}, err
	}
	return vm, nil
}