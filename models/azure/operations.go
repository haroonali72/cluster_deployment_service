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
	"github.com/aws/aws-sdk-go/service/ec2"
	"strings"
	"antelope/models/networks"
	"io/ioutil"
	"time"
	"os/exec"
	"fmt"
)
type CreatedPool struct {
	Instances    []*ec2.Instance
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
	network , err := networks.GetNetworkStatus(cluster.EnvironmentId,"azure")

	if err != nil {
		beego.Error(err.Error())
		return nil ,err
	}

	var createdPools []CreatedPool

	for _, pool := range cluster.NodePools {

		beego.Info("AWSOperations: creating key")
		var createdPool CreatedPool
		logging.SendLog("Creating Key " + pool.KeyName,"info",cluster.EnvironmentId)

		keyResponse,err := cloud.GenerateKeyPair(pool.KeyName)
		if err != nil {
			beego.Error(err.Error())
			logging.SendLog("Error in key creation: " + pool.KeyName,"info",cluster.EnvironmentId)
			logging.SendLog(err.Error(),"info",cluster.EnvironmentId)
			return nil , err
		}
		beego.Info("AWSOperations creating nodes")

		result, err :=  cloud.CreateInstance(pool,network)
		if err != nil {
			logging.SendLog("Error in instances creation: " + err.Error(),"info",cluster.EnvironmentId)
			beego.Error(err.Error())
			return nil, err
		}

		if result != nil && result.Instances != nil && len(result.Instances) > 0 {
			for index, instance := range result.Instances {
				err := cloud.updateInstanceTags(instance.InstanceId, pool.Name+"_"+strconv.Itoa(index))
				if err != nil {
					logging.SendLog("Error in instances creation: " + err.Error(),"info",cluster.EnvironmentId)
					beego.Error(err.Error())
					return nil, err
				}
			}
		}

		var latest_instances []*ec2.Instance

		latest_instances ,err= cloud.GetInstances(result,cluster.EnvironmentId)
		if err != nil {
			return nil, err
		}

		createdPool.KeyName =pool.KeyName
		createdPool.Key = keyResponse.PrivateKey
		createdPool.Instances= latest_instances
		createdPool.PoolName=pool.Name
		createdPools = append(createdPools,createdPool)
	}

	return createdPools,nil
}
func (cloud *AZURE) CreateInstance (pool *NodePool,networkData networks.AzureNetwork  )(*ec2.Reservation, error){


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
		addressClient.CreateOrUpdate(cloud.context,groupName, IPname, pipParameters)


		/*
		making network interface
		 */

		nicParameters := network.Interface{
			Location: &cloud.Region,
			InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
				IPConfigurations: &[]network.InterfaceIPConfiguration{
					{
						Name: to.StringPtr(fmt.Sprintf("IPconfig-%s",  strings.ToLower(pool.Name)+"-"+string(index))),
						InterfaceIPConfigurationPropertiesFormat: &network.InterfaceIPConfigurationPropertiesFormat{
							PrivateIPAllocationMethod: network.Dynamic,
							Subnet: &network.Subnet{ID: to.StringPtr(subnetId)},
							PublicIPAddress:
						},
					},
				},
			},
		}
		if sgIds!= nil {
			nicParameters.InterfacePropertiesFormat.NetworkSecurityGroup = &network.SecurityGroup{
				ID: (sgIds[0]),
			}
		}
		interfacesClient := network.NewInterfacesClient(cloud.Subscription)
		interfacesClient.Authorizer = cloud.Authorizer

		nicName := fmt.Sprintf("NIC-%s", pool.Name +"-"+ string (index))

		interfacesClient.CreateOrUpdate(cloud.context, groupName, nicName, nicParameters)

		osDisk := &compute.OSDisk{
			CreateOption: compute.DiskCreateOptionTypesFromImage,
			Name : to.StringPtr(pool.Name+"-"+string(index)),
			ManagedDisk: &compute.ManagedDiskParameters{
				StorageAccountType:compute.StorageAccountTypesStandardSSDLRS,
			},
		}

		storageName := "ext" + pool.Name + strconv.Itoa(index)
		disk := compute.DataDisk{
			Lun:          to.Int32Ptr(int32(index)),
			Name:         to.StringPtr(storageName),
			CreateOption: compute.DiskCreateOptionTypesFromImage,
			DiskSizeGB:   to.Int32Ptr(int32(1023)),
		}

		var storage = []compute.DataDisk{}
		storage = append(storage, disk)

		vm = compute.VirtualMachine{
			Name:     to.StringPtr(pool.Name+"-"+string(index)),
			Location: to.StringPtr(cloud.Region),
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				HardwareProfile: &compute.HardwareProfile{
					VMSize: compute.VirtualMachineSizeTypes(pool.MachineType),
				},
				StorageProfile: &compute.StorageProfile{
					ImageReference: &compute.ImageReference{
						Offer: 		to.StringPtr(pool.Image.Offer),
						Sku:		to.StringPtr(pool.Image.Sku),
						Publisher:	to.StringPtr(pool.Image.Publisher),
						Version:	to.StringPtr(pool.Image.Version),
					},
					OsDisk:    osDisk,
					DataDisks: &storage,
				},
				OsProfile: &compute.OSProfile{
					ComputerName:to.StringPtr(pool.Name),
					AdminUsername:to.StringPtr(""),
					AdminPassword:to.StringPtr(""),
				},
				NetworkProfile: &compute.NetworkProfile{
					NetworkInterfaces: &[]compute.NetworkInterfaceReference{
						{
							NetworkInterfaceReferenceProperties: &compute.NetworkInterfaceReferenceProperties{
								Primary: to.BoolPtr(true),
							},
						},
					},
				},

			},

		}


	}
	if err != nil {
		beego.Warn(err.Error())
		return nil, err
	}
	return result, nil

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
