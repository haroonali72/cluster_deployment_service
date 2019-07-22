package gcp

import (
	"antelope/models/logging"
	"antelope/models/networks"
	"antelope/models/types"
	"antelope/models/utils"
	"context"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"strings"
)

type GCP struct {
	Client      *compute.Service
	Credentials string
	ProjectId   string
	Region      string
}

func getNetworkHost(cloudType string) string {
	host := beego.AppConfig.String("network_url")
	if strings.Contains(host, "{cloud_provider}") {
		host = strings.Replace(host, "{cloud_provider}", cloudType, -1)
	}
	return host
}
func (cloud *GCP) createCluster(cluster Cluster_Def) (Cluster_Def, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return cluster, err
		}
	}
	var gcpNetwork types.GCPNetwork
	url := getNetworkHost("gcp") + "/" + cluster.ProjectId
	network, err := networks.GetAPIStatus(url, logging.Context{})
	if err != nil {
		beego.Error(err.Error())
		return cluster, err
	}
	bytes, err := json.Marshal(network)
	if err != nil {
		beego.Error(err.Error())
		return cluster, err
	}

	err = json.Unmarshal(bytes, &gcpNetwork)

	if err != nil {
		beego.Error(err.Error())
		return cluster, err
	}
	for _, pool := range cluster.NodePools {
		beego.Info("GCPOperations creating nodes")

		if pool.PoolRole == "master" {
			instance := compute.Instance{
				Name:        strings.ToLower(pool.Name),
				MachineType: pool.MachineType,
				NetworkInterfaces: []*compute.NetworkInterface{
					{
						Subnetwork: getSubnet(pool.PoolSubnet, gcpNetwork.Definition[0].Subnets),
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
			ctx := context.Background()
			_, err = cloud.Client.Instances.Insert(cloud.ProjectId, "a", &instance).Context(ctx).Do()
			if err != nil {
				beego.Error(err.Error())
				return cluster, err
			}
		} else {

			instanceTemplate, err := cloud.createInstanceTemplate(pool, gcpNetwork)
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
	}

	return cluster, nil
}

func (cloud *GCP) createInstanceTemplate(pool *NodePool, network types.GCPNetwork) (string, error) {
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
	if pool.PoolRole == "master" {
		ctx := context.Background()
		_, err := cloud.Client.Instances.Delete(cloud.ProjectId, "", pool.Name).Context(ctx).Do()
		if err != nil {
			beego.Error(err.Error())
			return err
		}
	} else {
		ctx := context.Background()
		_, err := cloud.Client.InstanceGroupManagers.Delete(cloud.ProjectId, "", pool.Name).Context(ctx).Do()
		if err != nil {
			beego.Error(err.Error())
			return err
		}
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
	if pool.PoolRole == "master" {
		result, err := cloud.Client.Instances.Get(cloud.ProjectId, "a", pool.Name).Context(ctx).Do()
		if err != nil {
			beego.Error(err.Error())
			return err
		}

		nodes := []*Node{}
		nodes = append(nodes, &Node{
			Url:    result.SelfLink,
			Status: result.Status,
		})

		pool.Nodes = nodes
	}
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

func getSubnet(subnetName string, subnets []*types.Subnet) string {
	for _, subnet := range subnets {
		if subnet.Name == subnetName {
			return subnet.SubnetId
		}
	}
	return ""
}
