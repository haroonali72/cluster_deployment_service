package aws

import (
	"errors"
	"github.com/astaxie/beego"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"strconv"
)

type AWS struct {
	Client    *ec2.EC2
	AccessKey string
	SecretKey string
	Region    string
}

func (cloud *AWS) createCluster(cluster Cluster_Def ) []*ec2.Instance {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return nil
		}
	}
	for _, pool := range cluster.Clusters[0].NodePools {
		input := &ec2.RunInstancesInput{
			ImageId:          aws.String(pool.Ami.AmiId),
			SubnetId:         aws.String(pool.SubnetId),
			SecurityGroupIds: pool.SecurityGroupId,
			MaxCount:         aws.Int64(pool.NodeCount),
			KeyName:          aws.String(pool.KeyName),
		}

		result, err := cloud.Client.RunInstances(input)
		if err != nil {
			beego.Warn(err.Error())
			return nil
		}

		if result != nil && result.Instances != nil && len(result.Instances) > 0 {
			for index, instance := range result.Instances {
				cloud.updateInstanceTags(instance.InstanceId, pool.Name+"_"+strconv.Itoa(index))
			}
		}
		return result.Instances
	}

	return nil
}

func (cloud *AWS) updateInstanceTags(instance_id * string ,nodepool_name string){
	var resource []*string
	resource = append(resource, instance_id)

	var tags []*ec2.Tag
	tag := ec2.Tag{Key: aws.String("Name"), Value: aws.String(nodepool_name)}
	tags = append(tags, &tag)

	input := ec2.CreateTagsInput{Resources: resource,
		Tags: tags,
	}
	out, err := cloud.Client.CreateTags(&input)
	if err != nil {
		beego.Warn(err.Error())
	}

	beego.Warn(out.String())
}

func (cloud *AWS) init() error {
	if cloud.Client != nil {
		return nil
	}

	if cloud.AccessKey == "" || cloud.SecretKey == "" || cloud.Region == "" {
		text := "invalid cloud credentials"
		beego.Error(text)
		return errors.New(text)
	}

	region := cloud.Region
	creds := credentials.NewStaticCredentials(cloud.AccessKey, cloud.SecretKey, "")

	cloud.Client = ec2.New(session.New(&aws.Config{Region: &region, Credentials: creds}))

	return nil
}

func (cloud *AWS) fetchStatus(cluster Cluster_Def ) (Cluster_Def, error){
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return Cluster_Def{},err
		}
	}
	for in, pool := range cluster.Clusters[0].NodePools {
		for index, node :=range pool.Nodes {
			name := "instance-id"
			ids := []*string{&node.CloudId}
			request := &ec2.DescribeInstancesInput{Filters: []*ec2.Filter{&ec2.Filter{Name: &name, Values: ids}}}
			out, err := cloud.Client.DescribeInstances(request)
			if err != nil {
				return Cluster_Def{}, err
			}
			pool.Nodes[index].NodeState=*out.Reservations[0].Instances[0].State.Name
		}
		cluster.Clusters[0].NodePools[in]=pool
	}

	return cluster,nil
}
