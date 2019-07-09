package gcp

import (
	"antelope/models/utils"
	"context"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"io/ioutil"
	"strings"
)

type GCP struct {
	Client      *compute.Service
	Credentials string
	ProjectId   string
	Region      string
}

type Network struct {
	Definition []*Definition `json:"definition" bson:"definition"`
}

type Definition struct {
	Vpc            Vpc              `json:"vpc" bson:"vpc"`
	Subnets        []*Subnet        `json:"subnets" bson:"subnets"`
	SecurityGroups []*SecurityGroup `json:"security_groups" bson:"security_groups"`
}

type Vpc struct {
	VpcId string `json:"vpc_id" bson:"vpc_id"`
	Name  string `json:"name" bson:"name"`
}

type Subnet struct {
	SubnetId string `json:"subnet_id" bson:"subnet_id"`
	Name     string `json:"name" bson:"name"`
}

type SecurityGroup struct {
	SecurityGroupId string `json:"security_group_id" bson:"security_group_id"`
	Name            string `json:"name" bson:"name"`
}

func (cloud *GCP) createCluster(cluster Cluster_Def) (Cluster_Def, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return cluster, err
		}
	}

	network, err := cloud.getNetworkStatus(cluster.ProjectId, "gcp")
	if err != nil {
		beego.Error(err.Error())
		return cluster, err
	}

	for _, pool := range cluster.NodePools {
		beego.Info("GCPOperations creating nodes")

		instanceTemplate, err := cloud.createInstanceTemplate(pool, network)
		if err != nil {
			return cluster, err
		}

		instanceGroup := compute.InstanceGroupManager{
			Name:             strings.ToLower(pool.Name),
			BaseInstanceName: strings.ToLower(pool.Name),
			TargetSize:       pool.NodeCount,
			InstanceTemplate: instanceTemplate,
		}

		ctx := context.Background()
		_, err = cloud.Client.InstanceGroupManagers.Insert(cloud.ProjectId, "a", &instanceGroup).Context(ctx).Do()
		if err != nil {
			beego.Error(err.Error())
			return cluster, err
		}
	}

	return cluster, nil
}

func (cloud *GCP) createInstanceTemplate(pool *NodePool, network Network) (string, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return "", err
		}
	}

	instanceProperties := compute.InstanceProperties{
		MachineType: pool.MachineType,
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				Subnetwork: getSubnet(pool.PoolSubnet, network.Definition[0].Subnets),
			},
		},
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				InitializeParams: &compute.AttachedDiskInitializeParams{
					SourceImage: "projects/" + pool.Image.Project + "/global/images/family/" + pool.Image.Family,
				},
			},
		},
	}

	instanceTemplate := compute.InstanceTemplate{
		Name:       strings.ToLower(pool.Name) + "-template",
		Properties: &instanceProperties,
	}

	ctx := context.Background()
	result, err := cloud.Client.InstanceTemplates.Insert(cloud.ProjectId, &instanceTemplate).Context(ctx).Do()
	if err != nil {
		beego.Error(err.Error())
		return "", err
	}

	return result.Name, nil
}

func (cloud *GCP) deleteCluster(cluster Cluster_Def) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return err
		}
	}

	for _, pool := range cluster.NodePools {
		cloud.deletePool(pool)
	}

	return nil
}

func (cloud *GCP) deletePool(pool *NodePool) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return err
		}
	}

	ctx := context.Background()
	_, err := cloud.Client.InstanceGroupManagers.Delete(cloud.ProjectId, "", pool.Name).Context(ctx).Do()
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}

func (cloud *GCP) fetchClusterStatus(cluster Cluster_Def) (Cluster_Def, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return cluster, err
		}
	}

	for _, pool := range cluster.NodePools {
		cloud.fetchPoolStatus(pool)
	}

	return cluster, nil
}

func (cloud *GCP) fetchPoolStatus(pool *NodePool) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return err
		}
	}

	ctx := context.Background()
	result, err := cloud.Client.InstanceGroupManagers.ListManagedInstances(cloud.ProjectId, "a", pool.Name).Context(ctx).Do()
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	nodes := []*Node{}
	for _, instance := range result.ManagedInstances {
		nodes = append(nodes, &Node{
			Url:           instance.Instance,
			Status:        instance.InstanceStatus,
			CurrentAction: instance.CurrentAction,
		})
	}
	pool.Nodes = nodes

	return nil
}

func (cloud *GCP) getNetworkStatus(envId string, cloudType string) (Network, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return Network{}, err
		}
	}

	networkUrl := strings.Replace(beego.AppConfig.String("network_url"), "{cloud_provider}", cloudType, -1)
	client := utils.InitReq()

	url := networkUrl + "/" + envId
	req, err := utils.CreateGetRequest(url)
	if err != nil {
		beego.Error("%s", err)
		return Network{}, err
	}

	response, err := client.SendRequest(req)
	if err != nil {
		beego.Error("%s", err)
		return Network{}, err
	}
	defer response.Body.Close()

	var gcpNetwork Network
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		beego.Error("%s", err)
		return Network{}, err
	}

	err = json.Unmarshal(contents, &gcpNetwork)
	if err != nil {
		beego.Error("%s", err)
		return Network{}, err
	}

	return gcpNetwork, nil
}

func (cloud *GCP) init() error {
	if cloud.Client != nil {
		return nil
	}

	var err error
	ctx := context.Background()

	cloud.Client, err = compute.NewService(ctx, option.WithCredentialsJSON([]byte(cloud.Credentials)))
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}

func GetGCP(credentials, region string) (GCP, error) {
	isValid, creds := utils.IsValdidGcpCredentials(credentials)
	if !isValid {
		text := "invalid cloud credentials"
		beego.Error(text)
		return GCP{}, errors.New(text)
	}

	return GCP{
		Credentials: creds.Raw,
		ProjectId:   creds.ProjectId,
		Region:      region,
	}, nil
}

func getSubnet(subnetName string, subnets []*Subnet) string {
	for _, subnet := range subnets {
		if subnet.Name == subnetName {
			return subnet.SubnetId
		}
	}
	return ""
}
