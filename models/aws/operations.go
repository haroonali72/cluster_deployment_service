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
type CreatedPool struct {
	Instances    []*ec2.Instance
	KeyName    	 string
	Key     	 string
	PoolName string
}
type AWS struct {
	Client    *ec2.EC2
	AccessKey string
	SecretKey string
	Region    string
}

func (cloud *AWS) createCluster(cluster Cluster_Def ) ([]CreatedPool , error){
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			beego.Error(err.Error())
			return nil ,err
		}
	}
	 var createdPools []CreatedPool

	for _, pool := range cluster.NodePools {
		beego.Info("AWSOperations: creating key")
		var createdPool CreatedPool
		keyMaterial,_,err  := cloud.KeyPairGenerator(pool.KeyName)
		if err != nil {
			beego.Warn(err.Error())
			return nil , err
		}
		beego.Info("AWSOperations creating nodes")
		input := &ec2.RunInstancesInput{
			ImageId:          aws.String(pool.Ami.AmiId),
			SubnetId:         aws.String(pool.SubnetId),
			SecurityGroupIds: pool.SecurityGroupId,
			MaxCount:         aws.Int64(pool.NodeCount),
			KeyName:          aws.String(pool.KeyName),
			MinCount: aws.Int64(1),
			InstanceType: aws.String(pool.MachineType),
		}

		result, err := cloud.Client.RunInstances(input)
		if err != nil {
			beego.Warn(err.Error())
			return nil, err
		}

		if result != nil && result.Instances != nil && len(result.Instances) > 0 {
			for index, instance := range result.Instances {
				cloud.updateInstanceTags(instance.InstanceId, pool.Name+"_"+strconv.Itoa(index))
			}
		}
		var latest_instances []*ec2.Instance
		if result != nil && result.Instances != nil && len(result.Instances) > 0 {
			for _, instance := range result.Instances {
				var ids []*string
				ids = append(ids,aws.String(*instance.InstanceId))
				instance_input := ec2.DescribeInstancesInput{InstanceIds: ids}
				new_instances , new_err := cloud.Client.DescribeInstances(&instance_input)
				if new_err != nil {
					beego.Warn(new_err.Error())
					return nil, new_err
				}
				latest_instances = append(latest_instances,new_instances.Reservations[0].Instances[0])
			}
		}
		createdPool.KeyName =pool.KeyName
		createdPool.Key = keyMaterial
		createdPool.Instances= latest_instances
		createdPool.PoolName=pool.Name
		createdPools = append(createdPools,createdPool)
	}

	return createdPools,nil
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
			beego.Error("Cluster model: Status - Failed to get lastest status ", err.Error())

			return Cluster_Def{},err
		}
	}
	for in, pool := range cluster.NodePools {
		for index, node :=range pool.Nodes {
			name := "instance-id"
			ids := []*string{&node.CloudId}
			request := &ec2.DescribeInstancesInput{Filters: []*ec2.Filter{&ec2.Filter{Name: &name, Values: ids}}}
			out, err := cloud.Client.DescribeInstances(request)
			if err != nil {
				beego.Error("Cluster model: Status - Failed to get lastest status ", err.Error())
				return Cluster_Def{}, err
			}
			pool.Nodes[index].NodeState=*out.Reservations[0].Instances[0].State.Name
			if out.Reservations[0].Instances[0].PublicIpAddress != nil {
				pool.Nodes[index].PublicIP = *out.Reservations[0].Instances[0].PublicIpAddress
			}
		}
		cluster.NodePools[in]=pool
	}

	return cluster,nil
}

func (cloud *AWS) getSSHKey ()( []*ec2.KeyPairInfo, error){
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return nil,err
		}
	}
 	input :=	&ec2.DescribeKeyPairsInput{}
	keys, err := cloud.Client.DescribeKeyPairs(input)
	if err != nil{
		return nil,err
	}
	return keys.KeyPairs, nil
}

func (cloud *AWS) KeyPairGenerator(keyName string) ( string ,string, error) {
	params := &ec2.CreateKeyPairInput{
		KeyName: aws.String(keyName),
		DryRun:  aws.Bool(false),
	}
	resp, err := cloud.Client.CreateKeyPair(params)
	if err != nil {
		return "","" ,err
	}

	return *resp.KeyMaterial, *resp.KeyFingerprint, nil
}