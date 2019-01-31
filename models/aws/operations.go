package aws

import (
	"errors"
	"github.com/astaxie/beego"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"strconv"
	"github.com/aws/aws-sdk-go/service/iam"
	"antelope/models/utils"
	"io/ioutil"
	"encoding/json"
	"gopkg.in/mgo.v2/bson"
	"antelope/models"
	"time"
	"antelope/models/logging"
	"strings"
)

var (
	networkHost    = beego.AppConfig.String("network_url")
)
var testInstanceMap = map[string]string{
	"us-east-2":      "ami-9686a4f3",
	"sa-east-1":      "ami-a3e39ecf",
	"eu-central-1":   "ami-5a922335",
	"us-west-1":      "ami-2d5c6d4d",
	"us-west-2":      "ami-ecc63a94",
	"ap-northeast-2": "ami-0f6fb461",
	"ca-central-1":   "ami-e59c2581",
	"eu-west-2":      "ami-e1f2e185",
	"ap-southeast-1": "ami-e6d3a585",
	"eu-west-1":      "ami-17d11e6e",
	"ap-southeast-2": "ami-391ff95b",
	"ap-northeast-1": "ami-8422ebe2",
	"us-east-1":      "ami-d651b8ac",
	"ap-south-1":     "ami-08a5e367"}

var docker_master_policy = []byte(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "kopsK8sEC2MasterPermsDescribeResources",
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstances",
        "ec2:DescribeRegions",
        "ec2:DescribeRouteTables",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeSubnets",
        "ec2:DescribeVolumes"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Sid": "kopsK8sEC2MasterPermsAllResources",
      "Effect": "Allow",
      "Action": [
        "ec2:CreateSecurityGroup",
        "ec2:CreateTags",
        "ec2:CreateVolume",
        "ec2:ModifyInstanceAttribute"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Sid": "kopsK8sEC2MasterPermsTaggedResources",
      "Effect": "Allow",
      "Action": [
        "ec2:AttachVolume",
        "ec2:AuthorizeSecurityGroupIngress",
        "ec2:CreateRoute",
        "ec2:DeleteRoute",
        "ec2:DeleteSecurityGroup",
        "ec2:DeleteVolume",
        "ec2:DetachVolume",
        "ec2:RevokeSecurityGroupIngress"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Sid": "kopsMasterCertIAMPerms",
      "Effect": "Allow",
      "Action": [
        "iam:ListServerCertificates",
        "iam:GetServerCertificate"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Sid": "kopsK8sS3GetListBucket",
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::kops-tests"
      ]
    },
    {
      "Sid": "kopsK8sELB",
      "Effect": "Allow",
      "Action": [
		"elasticloadbalancing:DescribeTags",
		"elasticloadbalancing:CreateLoadBalancerListeners",
		"elasticloadbalancing:ConfigureHealthCheck",
		"elasticloadbalancing:DeleteLoadBalancerListeners",
		"elasticloadbalancing:RegisterInstancesWithLoadBalancer",
		"elasticloadbalancing:DescribeLoadBalancers",
		"elasticloadbalancing:CreateLoadBalancer",
		"elasticloadbalancing:DeleteLoadBalancer",
		"elasticloadbalancing:ModifyLoadBalancerAttributes",
		"elasticloadbalancing:DescribeLoadBalancerAttributes"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}`)
type Network struct {
	EnvironmentId    string        `json:"environment_id" bson:"environment_id"`
	Name             string        `json:"name" bson:"name"`
	Type             models.Type   `json:"type" bson:"type"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"`
	NetworkStatus    string        `json:"status" bson:"status"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	Definition       []*Definition `json:"definition" bson:"definition"`
}

type Definition struct {
	ID             bson.ObjectId    `json:"_id" bson:"_id,omitempty"`
	Vpc            Vpc              `json:"vpc" bson:"vpc"`
	Subnets        []*Subnet        `json:"subnets" bson:"subnets"`
	SecurityGroups []*SecurityGroup `json:"security_groups" bson:"security_groups"`
}

type Vpc struct {
	ID    bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	VpcId string        `json:"vpc_id" bson:"vpc_id"`
	Name  string        `json:"name" bson:"name"`
	CIDR  string        `json:"cidr" bson:"cidr"`
}

type Subnet struct {
	ID       bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	SubnetId string        `json:"subnet_id" bson:"subnet_id"`
	Name     string        `json:"name" bson:"name"`
	CIDR  	 string        `json:"cidr" bson:"cidr"`
}

type SecurityGroup struct {
	ID              bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	SecurityGroupId string        `json:"security_group_id" bson:"security_group_id"`
	Name            string        `json:"name" bson:"name"`
	Description     string        `json:"description" bson:"description"`
}

type CreatedPool struct {
	Instances    []*ec2.Instance
	KeyName    	 string
	Key     	 string
	PoolName string
}

type AWS struct {
	Client    	*ec2.EC2
	IAMService	*iam.IAM
	AccessKey 	string
	SecretKey 	string
	Region   	string
}

func (cloud *AWS) createCluster(cluster Cluster_Def ) ([]CreatedPool , error){

	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			beego.Error(err.Error())
			return nil ,err
		}
	}
	network , err := cloud.GetNetworkStatus(cluster.EnvironmentId)

	if err != nil {
		beego.Error(err.Error())
		return nil ,err
	}

	 var createdPools []CreatedPool

	for _, pool := range cluster.NodePools {

		beego.Info("AWSOperations: creating key")
		var createdPool CreatedPool
		logging.SendLog("Creating Key " + pool.KeyName,"info",cluster.EnvironmentId)

		keyMaterial,_,err  := cloud.KeyPairGenerator(pool.KeyName)
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
				err := cloud.updateInstanceTags(instance.InstanceId, pool.Name+"-"+strconv.Itoa(index))
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
		createdPool.Key = keyMaterial
		createdPool.Instances= latest_instances
		createdPool.PoolName=pool.Name
		createdPools = append(createdPools,createdPool)
	}

	return createdPools,nil
}

func (cloud *AWS) updateInstanceTags(instance_id * string ,nodepool_name string)(error){
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
		beego.Error(err.Error())
		return err
	}

	beego.Info(out.String())
	return nil
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
	cloud.IAMService = iam.New(session.New(&aws.Config{Region: &region, Credentials: creds}))

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

			out, err := cloud.GetInstanceStatus(node)
			if err != nil {
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
func (cloud *AWS) terminateCluster(cluster Cluster_Def ) ( error){
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			beego.Error(err.Error())
			return err
		}
	}

	for _, pool := range cluster.NodePools {
		err := cloud.TerminatePool(pool, cluster.EnvironmentId)
		if err != nil {
			return  err
		}
	}
	return nil
}
func (cloud *AWS) CreateInstance (pool *NodePool, network Network )(*ec2.Reservation, error){


	subnetId := cloud.GetSubnets(pool,network)
    sgIds := cloud.GetSecurityGroups(pool,network)

	_, err := cloud.createIAMRole(pool.Name)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}

	input := &ec2.RunInstancesInput{
		ImageId:          aws.String(pool.Ami.AmiId),
		SubnetId:         aws.String(subnetId),
		SecurityGroupIds: sgIds,
		MaxCount:         aws.Int64(pool.NodeCount),
		KeyName:          aws.String(pool.KeyName),
		MinCount: aws.Int64(1),
		InstanceType: aws.String(pool.MachineType),
	}
	ok := cloud.checkInstanceProfile(pool.Name)
	if !ok {
		iamProfile := ec2.IamInstanceProfileSpecification{Name: aws.String(pool.Name)}
		input.IamInstanceProfile = &iamProfile
	} else {
		beego.Info("failed in attaching")
	}

	result, err := cloud.Client.RunInstances(input)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}
	return result, nil

}
func (cloud *AWS) GetSecurityGroups (pool *NodePool, network Network )([]*string) {
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
func (cloud *AWS) GetSubnets (pool *NodePool, network Network )(string) {
	for _, definition := range network.Definition{
		for _, subnet := range definition.Subnets {
			if subnet.Name ==  pool.PoolSubnet{
				return subnet.SubnetId
			}
		}
	}
	return ""
}

func (cloud *AWS) GetInstances (result *ec2.Reservation, envId string)(latest_instances []*ec2.Instance,err error){

	if result != nil && result.Instances != nil && len(result.Instances) > 0 {

		var ids []*string
		for _, instance := range result.Instances {
			ids = append(ids, aws.String(*instance.InstanceId))
		}

		instance_input := ec2.DescribeInstancesInput{InstanceIds: ids}
		updated_instances, err := cloud.Client.DescribeInstances(&instance_input)

		if err != nil {
			beego.Error(err.Error())
			return nil, err
		}

		for _, instance := range updated_instances.Reservations[0].Instances{
				logging.SendLog("Instance created successfully: " + *instance.InstanceId ,"info",envId)
				latest_instances = append(latest_instances,instance)
		}
		return latest_instances, nil
	}
	return nil, nil
}
func (cloud *AWS) GetInstanceStatus (node *Node)(output *ec2.DescribeInstancesOutput,err error){

	name := "instance-id"
	ids := []*string{&node.CloudId}

	request := &ec2.DescribeInstancesInput{Filters: []*ec2.Filter{&ec2.Filter{Name: &name, Values: ids}}}
	output, err = cloud.Client.DescribeInstances(request)

	if err != nil {
		beego.Error("Cluster model: Status - Failed to get lastest status ", err.Error())
		return nil, err
	}
	return output, nil
}
func (cloud *AWS) TerminatePool(pool *NodePool, envId string ) ( error) {

	beego.Info("AWSOperations terminating nodes")
	var instance_ids []*string

	for _, id := range pool.Nodes {
		instance_ids = append(instance_ids, &id.CloudId)
	}

	input := &ec2.TerminateInstancesInput{
		InstanceIds: instance_ids,
	}

	_, err := cloud.Client.TerminateInstances(input)
	if err != nil {

		beego.Error("Cluster model: Status - Failed to terminate node pool ", err.Error())
		return err
	}
	logging.SendLog("Cluster pool terminated successfully: " + pool.Name,"info",envId)
	return nil
}

func (cloud *AWS) GetNetworkStatus(envId string ) ( Network, error){

	client := utils.InitReq()

	req , err :=utils.CreateGetRequest(envId, networkHost)

	response, err := client.SendRequest(req)

	defer response.Body.Close()

	var network Network

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		beego.Error("%s", err)
		return Network{},err
	}

	err = json.Unmarshal(contents,&network)
	if err != nil {
		beego.Error("%s", err)
		return Network{}, err
	}
	return network, nil

}
func (cloud *AWS) createIAMRole(name string) (string, error) {


		roleName := name


		raw_policy := docker_master_policy
		raw_role := []byte(`{
				  "Version": "2012-10-17",
				  "Statement": [
				    {
				      "Effect": "Allow",
				      "Principal": { "Service": "ec2.amazonaws.com"},
				      "Action": "sts:AssumeRole"
				    }
				  ]
				}`)
		role := string(raw_role)
		policy := string(raw_policy)

		roleInput := iam.CreateRoleInput{AssumeRolePolicyDocument: &role, RoleName: &roleName}
		out, err := cloud.IAMService.CreateRole(&roleInput)
		if err != nil {
			beego.Error(err)
			return "", err
		}

		beego.Info(out.GoString())


		policy_out, err_1 := cloud.IAMService.CreatePolicy(&iam.CreatePolicyInput{
			PolicyDocument: aws.String(policy),
			PolicyName:     &roleName,
		})

		if err_1 != nil {
			beego.Error(err_1)
			return "",err_1
		}

		attach := iam.AttachRolePolicyInput{RoleName: &roleName, PolicyArn: policy_out.Policy.Arn}
		_, err_2 := cloud.IAMService.AttachRolePolicy(&attach)

		if err_2 != nil {
			beego.Error(err_2)
			return "", err_2
		}

		profileInput := iam.CreateInstanceProfileInput{InstanceProfileName: &roleName}
		outtt, err := cloud.IAMService.CreateInstanceProfile(&profileInput)
		if err != nil {
			beego.Error(err)
			return "", err
		}

		testProfile := iam.AddRoleToInstanceProfileInput{InstanceProfileName: &roleName, RoleName: &roleName}
		_, err = cloud.IAMService.AddRoleToInstanceProfile(&testProfile)
		if err != nil {
			beego.Error(err)
			return "",  err
		}

		return  *outtt.InstanceProfile.Arn, nil

}

func (cloud *AWS) checkInstanceProfile(iamProfileName string ) bool {

	iamProfile := ec2.IamInstanceProfileSpecification{Name: aws.String(iamProfileName)}

	start := time.Now()
	timeToWait := 60 //seconds
	retry := true

	region := cloud.Region
	ami := testInstanceMap[region]

	for retry && int64(time.Since(start).Seconds()) < int64(timeToWait) {

		//this dummy instance run , to check the success of RunInstance call
		//this is to ensure that iamProfile is properly propagated
		_, err :=cloud.Client.RunInstances(&ec2.RunInstancesInput{
			// An Amazon Linux AMI ID for t2.micro instances in the us-west-2 region
			ImageId:            aws.String(ami),
			InstanceType:       aws.String("t2.micro"),
			MinCount:           aws.Int64(1),
			MaxCount:           aws.Int64(1),
			DryRun:             aws.Bool(true),
			IamInstanceProfile: &iamProfile,
		})

		beego.Error(err)

		if err != nil && strings.Contains(err.Error(), "DryRunOperation: Request would have succeeded") {
			retry = false
		} else {
			beego.Info("time passed %6.2f sec\n", time.Since(start).Seconds())
			beego.Info("waiting 5 seconds before retry")
			time.Sleep(5 * time.Second)
		}

	}
	beego.Info("retry", retry)
	return retry
}