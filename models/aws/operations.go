package aws

import (
	"antelope/models"
	"antelope/models/logging"
	"antelope/models/utils"
	"antelope/models/vault"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
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
      	"Sid": "VisualEditor0",
		"Effect": "Allow",
         "Action": [
                "ec2:AttachVolume",
                "elasticloadbalancing:ModifyListener",
                "ec2:AuthorizeSecurityGroupIngress",
                "ec2:DescribeInstances",
                "iam:ListServerCertificates",
                "elasticloadbalancing:ConfigureHealthCheck",
                "elasticloadbalancing:RegisterTargets",
                "ec2:DescribeRegions",
                "elasticloadbalancing:ModifyTargetGroups",
                "elasticloadbalancing:DeleteLoadBalancer",
                "elasticloadbalancing:DescribeLoadBalancers",
                "ec2:DeleteVolume",
                "elasticloadbalancing:RemoveTags",
                "elasticloadbalancing:CreateListener",
                "elasticloadbalancing:DescribeListeners",
                "elasticloadbalancing:DeleteTargetGroups",
                "ec2:CreateRoute",
                "ec2:CreateSecurityGroup",
                "ec2:DescribeVolumes",
                "ec2:ModifyInstanceAttribute",
                "elasticloadbalancing:RegisterInstancesWithLoadBalancer",
                "elasticloadbalancing:DeleteListeners",
                "ec2:DescribeRouteTables",
                "ec2:DetachVolume",
                "iam:GetServerCertificate",
                "elasticloadbalancing:CreateLoadBalancer",
                "elasticloadbalancing:DescribeTags",
                "ec2:CreateTags",
                "elasticloadbalancing:CreateTargetGroup",
                "ec2:DeleteRoute",
                "elasticloadbalancing:DeregisterTargets",
                "elasticloadbalancing:*",
                "elasticloadbalancing:DeleteTargetGroup",
                "elasticloadbalancing:CreateLoadBalancerListeners",
                "ec2:DescribeSecurityGroups",
                "ec2:CreateVolume",
                "elasticloadbalancing:DescribeLoadBalancerAttributes",
                "ec2:RevokeSecurityGroupIngress",
                "elasticloadbalancing:AddTags",
                "ec2:DeleteSecurityGroup",
                "elasticloadbalancing:DescribeTargetGroups",
                "elasticloadbalancing:DeleteLoadBalancerListeners",
                "ec2:*",
                "ec2:DescribeSubnets",
                "elasticloadbalancing:ModifyLoadBalancerAttributes",
                "elasticloadbalancing:ModifyTargetGroup",
                "elasticloadbalancing:DeleteListener",
                "ecr:*"
            ],
         "Resource": ["*" ]
        },
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
	ProjectId        string        `json:"Project_id" bson:"Project_id"`
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
	CIDR     string        `json:"cidr" bson:"cidr"`
}

type SecurityGroup struct {
	ID              bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	SecurityGroupId string        `json:"security_group_id" bson:"security_group_id"`
	Name            string        `json:"name" bson:"name"`
	Description     string        `json:"description" bson:"description"`
}

type CreatedPool struct {
	Instances []*ec2.Instance
	PoolName  string
}

type AWS struct {
	Client     *ec2.EC2
	IAMService *iam.IAM
	STS        *sts.STS
	AccessKey  string
	SecretKey  string
	Region     string
	Resources  map[string]interface{}
}

func (cloud *AWS) createCluster(cluster Cluster_Def) ([]CreatedPool, error) {

	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return nil, err
		}
	}
	network, err := cloud.GetNetworkStatus(cluster.ProjectId)

	if err != nil {
		return nil, err
	}

	var createdPools []CreatedPool

	for _, pool := range cluster.NodePools {
		var createdPool CreatedPool
		keyMaterial, err := cloud.getKey(*pool, cluster.ProjectId)
		if err != nil {
			return nil, err
		}
		beego.Info("AWSOperations creating nodes")

		result, err := cloud.CreateInstance(pool, network)
		if err != nil {
			logging.SendLog("Error in instances creation: "+err.Error(), "info", cluster.ProjectId)
			return nil, err
		}

		if result != nil && result.Instances != nil && len(result.Instances) > 0 {
			for index, instance := range result.Instances {
				err := cloud.updateInstanceTags(instance.InstanceId, pool.Name+"-"+strconv.Itoa(index), cluster.ProjectId)
				if err != nil {
					logging.SendLog("Error in instances creation: "+err.Error(), "info", cluster.ProjectId)
					return nil, err
				}
			}
			if pool.Ami.IsExternal {
				pool.KeyInfo.KeyMaterial = keyMaterial
				err = cloud.mountVolume(result.Instances, pool.Ami, pool.KeyInfo, cluster.ProjectId)
				if err != nil {
					logging.SendLog("Error in instances creation: "+err.Error(), "info", cluster.ProjectId)
					return nil, err
				}
			}
		}

		var latest_instances []*ec2.Instance

		if result != nil && result.Instances != nil && len(result.Instances) > 0 {

			var ids []*string
			for _, instance := range result.Instances {
				ids = append(ids, aws.String(*instance.InstanceId))
			}
			latest_instances, err = cloud.GetInstances(ids, cluster.ProjectId, true)
			if err != nil {
				return nil, err
			}

		}

		createdPool.Instances = latest_instances
		createdPool.PoolName = pool.Name
		createdPools = append(createdPools, createdPool)
	}

	return createdPools, nil
}

func (cloud *AWS) updateInstanceTags(instance_id *string, nodepool_name string, projectId string) error {
	var resource []*string
	resource = append(resource, instance_id)

	var tags []*ec2.Tag
	tag := ec2.Tag{Key: aws.String("Name"), Value: aws.String(nodepool_name)}
	tag_ := ec2.Tag{Key: aws.String("KubernetesCluster"), Value: aws.String(projectId)}
	tags = append(tags, &tag)
	tags = append(tags, &tag_)

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
	cloud.STS = sts.New(session.New(&aws.Config{Region: &region, Credentials: creds}))
	cloud.Resources = make(map[string]interface{})
	return nil
}

func (cloud *AWS) fetchStatus(cluster Cluster_Def) (Cluster_Def, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			beego.Error("Cluster model: Status - Failed to get lastest status ", err.Error())

			return Cluster_Def{}, err
		}
	}
	for in, pool := range cluster.NodePools {

		for index, node := range pool.Nodes {
			keyInfo, err := vault.GetSSHKey("aws", pool.KeyInfo.KeyName)
			if err != nil {
				return Cluster_Def{}, err
			}
			var nodeId []*string
			nodeId = append(nodeId, &node.CloudId)
			out, err := cloud.GetInstances(nodeId, cluster.ProjectId, false)
			if err != nil {
				return Cluster_Def{}, err
			}
			if out != nil {
				pool.Nodes[index].NodeState = *out[0].State.Name

				if out[0].PublicIpAddress != nil {
					pool.Nodes[index].PublicIP = *out[0].PublicIpAddress
				}
				if out[0].PrivateDnsName != nil {
					pool.Nodes[index].PrivateDNS = *out[0].PrivateDnsName
				}
				if out[0].PublicDnsName != nil {
					pool.Nodes[index].PublicDNS = *out[0].PublicDnsName
				}
				if out[0].PrivateIpAddress != nil {
					pool.Nodes[index].PrivateIP = *out[0].PrivateIpAddress
				}

			}
			k, err := keyCoverstion(keyInfo)
			if err != nil {
				return Cluster_Def{}, err
			}
			pool.KeyInfo = k

		}
		cluster.NodePools[in] = pool
	}
	return cluster, nil
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
func (cloud *AWS) getSSHKey() ([]*ec2.KeyPairInfo, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return nil, err
		}
	}
	input := &ec2.DescribeKeyPairsInput{}
	keys, err := cloud.Client.DescribeKeyPairs(input)
	if err != nil {
		return nil, err
	}
	return keys.KeyPairs, nil
}

func (cloud *AWS) KeyPairGenerator(keyName string) (string, string, error) {
	params := &ec2.CreateKeyPairInput{
		KeyName: aws.String(keyName),
		DryRun:  aws.Bool(false),
	}
	resp, err := cloud.Client.CreateKeyPair(params)
	if err != nil {
		return "", "", err
	}

	return *resp.KeyMaterial, *resp.KeyFingerprint, nil
}
func (cloud *AWS) terminateCluster(cluster Cluster_Def) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			beego.Error(err.Error())
			return err
		}
	}

	for _, pool := range cluster.NodePools {
		err := cloud.TerminatePool(pool, cluster.ProjectId)
		if err != nil {
			return err
		}
		err = cloud.deleteIAMRole(pool.Name)
		if err != nil {
			return err
		}
	}
	return nil
}
func (cloud *AWS) CleanUp(cluster Cluster_Def) error {
	beego.Info("in clean up method")
	for _, pool := range cluster.NodePools {
		beego.Info("terminating pool" + pool.Name)
		beego.Info(cloud.Resources[pool.Name+"_iamProfile"])
		if cloud.Resources[pool.Name+"_iamProfile"] != nil {
			iamProfile := cloud.Resources[pool.Name+"_iamProfile"]
			name := ""
			b, e := json.Marshal(iamProfile)
			if e != nil {
				return e
			}
			e = json.Unmarshal(b, &name)
			if e != nil {
				return e
			}
			err := cloud.deleteIAMProfile(name)
			if err != nil {
				return err
			}
		}
		beego.Info(cloud.Resources[pool.Name+"_role"])
		if cloud.Resources[pool.Name+"_role"] != nil {
			role := cloud.Resources[pool.Name+"_role"]
			name := ""
			b, e := json.Marshal(role)
			if e != nil {
				return e
			}
			e = json.Unmarshal(b, &name)
			if e != nil {
				return e
			}
			err := cloud.deleteRole(name)
			if err != nil {
				return err
			}
		}
		beego.Info(cloud.Resources[pool.Name+"_policy"])
		if cloud.Resources[pool.Name+"_policy"] != nil {
			policy := cloud.Resources[pool.Name+"_policy"]
			name := ""
			b, e := json.Marshal(policy)
			if e != nil {
				return e
			}
			e = json.Unmarshal(b, &name)
			if e != nil {
				return e
			}
			err := cloud.deletePolicy(name)
			if err != nil {
				return err
			}
		}
		beego.Info(cloud.Resources[pool.Name+"_instances"])
		if cloud.Resources[pool.Name+"_instances"] != nil {
			value := cloud.Resources[pool.Name+"_instances"]
			var ids []*string
			b, e := json.Marshal(value)
			if e != nil {
				return e
			}
			e = json.Unmarshal(b, &ids)
			if e != nil {
				return e
			}
			err := cloud.TerminateIns(ids)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
func (cloud *AWS) CreateInstance(pool *NodePool, network Network) (*ec2.Reservation, error) {

	subnetId := cloud.GetSubnets(pool, network)
	sgIds := cloud.GetSecurityGroups(pool, network)

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
		KeyName:          aws.String(pool.KeyInfo.KeyName),
		MinCount:         aws.Int64(1),
		InstanceType:     aws.String(pool.MachineType),
	}
	/*
		setting 50 gb volume - temp work
	*/
	beego.Info("updating root volume ")
	ebs, err := cloud.describeAmi(&pool.Ami.AmiId)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}
	if ebs != nil && ebs[0].Ebs != nil && ebs[0].Ebs.VolumeSize != nil {
		ebs[0].Ebs.VolumeSize = &pool.Ami.RootVolume.VolumeSize
		ebs[0].Ebs.VolumeType = &pool.Ami.RootVolume.VolumeType
		if pool.Ami.RootVolume.VolumeType == "io1" {
			ebs[0].Ebs.Iops = &pool.Ami.RootVolume.Iops
		}
	}

	beego.Info("attaching external volume")
	if pool.Ami.IsExternal {
		var external_volume ec2.BlockDeviceMapping

		var external_ebs ec2.EbsBlockDevice

		external_ebs.VolumeType = &pool.Ami.ExternalVolume.VolumeType
		external_ebs.VolumeSize = &pool.Ami.ExternalVolume.VolumeSize
		if pool.Ami.ExternalVolume.VolumeType == "io1" {
			external_ebs.Iops = &pool.Ami.ExternalVolume.Iops
		}
		external_volume.Ebs = &external_ebs
		external_volume.DeviceName = aws.String("/dev/sdf")
		ebs = append(ebs, &external_volume)
	}
	input.BlockDeviceMappings = ebs

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
	if result != nil && result.Instances != nil && len(result.Instances) > 0 {

		var ids []*string
		for _, instance := range result.Instances {
			ids = append(ids, aws.String(*instance.InstanceId))
		}
		cloud.Resources[pool.Name+"_instances"] = ids
	}

	return result, nil

}
func (cloud *AWS) GetSecurityGroups(pool *NodePool, network Network) []*string {
	var sgId []*string
	for _, definition := range network.Definition {
		for _, sg := range definition.SecurityGroups {
			for _, sgName := range pool.PoolSecurityGroups {
				if sgName != nil {
					if *sgName == sg.Name {
						sgId = append(sgId, &sg.SecurityGroupId)
					}
				}
			}
		}
	}
	return sgId
}
func (cloud *AWS) GetSubnets(pool *NodePool, network Network) string {
	for _, definition := range network.Definition {
		for _, subnet := range definition.Subnets {
			if subnet.Name == pool.PoolSubnet {
				return subnet.SubnetId
			}
		}
	}
	return ""
}

func (cloud *AWS) GetInstances(ids []*string, projectId string, creation bool) (latest_instances []*ec2.Instance, err error) {

	instance_input := ec2.DescribeInstancesInput{InstanceIds: ids}
	updated_instances, err := cloud.Client.DescribeInstances(&instance_input)

	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}
	if updated_instances == nil || updated_instances.Reservations == nil || updated_instances.Reservations[0].Instances == nil {

		return nil, errors.New("Nodes not found")
	}
	for _, instance := range updated_instances.Reservations[0].Instances {
		if creation {
			logging.SendLog("Instance created successfully: "+*instance.InstanceId, "info", projectId)
		}
		latest_instances = append(latest_instances, instance)
	}
	return latest_instances, nil

	return nil, nil
}
func (cloud *AWS) getIds(pool *NodePool) []*string {
	var instance_ids []*string

	for _, id := range pool.Nodes {
		instance_ids = append(instance_ids, &id.CloudId)
	}
	return instance_ids
}

func (cloud *AWS) TerminateIns(instance_ids []*string) error {
	input := &ec2.TerminateInstancesInput{
		InstanceIds: instance_ids,
	}

	_, err := cloud.Client.TerminateInstances(input)

	return err
}
func (cloud *AWS) TerminatePool(pool *NodePool, projectId string) error {

	beego.Info("AWSOperations terminating nodes")
	instance_ids := cloud.getIds(pool)

	err := cloud.TerminateIns(instance_ids)
	if err != nil {

		beego.Error("Cluster model: Status - Failed to terminate node pool ", err.Error())
		return err
	}
	logging.SendLog("Cluster pool terminated successfully: "+pool.Name, "info", projectId)
	return nil
}
func (cloud *AWS) deleteIAMRole(name string) error {

	roleName := name
	err := cloud.deleteIAMProfile(roleName)
	if err != nil {
		return err
	}
	err = cloud.deleteRole(roleName)
	if err != nil {
		return err
	}
	err = cloud.deletePolicy(roleName)
	if err != nil {
		return err
	}
	return nil

}
func (cloud *AWS) deletePolicy(policyName string) error {
	err, policyArn := cloud.getPolicyARN(policyName)
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	policy_input := iam.DeletePolicyInput{PolicyArn: &policyArn}
	policy_out, err_1 := cloud.IAMService.DeletePolicy(&policy_input)

	if err_1 != nil {
		beego.Error(err_1.Error())
		return err_1
	}

	beego.Info(policy_out.GoString())
	return nil
}
func (cloud *AWS) getPolicyARN(policyName string) (error, string) {
	id, err := cloud.getAccountId()
	if err != nil {
		beego.Error(err.Error())
		return err, ""
	}
	policyArn := "arn:aws:iam::" + id + ":policy/" + policyName
	return nil, policyArn
}
func (cloud *AWS) deleteRole(roleName string) error {
	err, policyArn := cloud.getPolicyARN(roleName)
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	policy := iam.DetachRolePolicyInput{RoleName: &roleName, PolicyArn: &policyArn}
	out, err := cloud.IAMService.DetachRolePolicy(&policy)
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	beego.Info(out.GoString())

	roleInput := iam.DeleteRoleInput{RoleName: &roleName}
	out_, err := cloud.IAMService.DeleteRole(&roleInput)
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	beego.Info(out_.GoString())
	return nil
}
func (cloud *AWS) deleteIAMProfile(roleName string) error {
	profile := iam.RemoveRoleFromInstanceProfileInput{InstanceProfileName: &roleName, RoleName: &roleName}
	outtt, err := cloud.IAMService.RemoveRoleFromInstanceProfile(&profile)
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	beego.Info(outtt.GoString())

	profileInput := iam.DeleteInstanceProfileInput{InstanceProfileName: &roleName}
	outt, err := cloud.IAMService.DeleteInstanceProfile(&profileInput)
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	beego.Info(outt.GoString())
	return nil
}
func (cloud *AWS) GetNetworkStatus(projectId string) (Network, error) {

	url := getNetworkHost()

	url = strings.Replace(url, "{cloud_provider}", "aws", -1)

	client := utils.InitReq()

	url = url + "/" + projectId

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

	var network Network

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		beego.Error("%s", err)
		return Network{}, err
	}

	err = json.Unmarshal(contents, &network)
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
	cloud.Resources[name+"_role"] = roleName
	beego.Info(out.GoString())

	policy_out, err_1 := cloud.IAMService.CreatePolicy(&iam.CreatePolicyInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     &roleName,
	})

	if err_1 != nil {
		beego.Error(err_1)
		return "", err_1
	}
	cloud.Resources[name+"_policy"] = roleName
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
	cloud.Resources[name+"_iamProfile"] = roleName
	testProfile := iam.AddRoleToInstanceProfileInput{InstanceProfileName: &roleName, RoleName: &roleName}
	_, err = cloud.IAMService.AddRoleToInstanceProfile(&testProfile)
	if err != nil {
		beego.Error(err)
		return "", err
	}

	return *outtt.InstanceProfile.Arn, nil

}

func (cloud *AWS) checkInstanceProfile(iamProfileName string) bool {

	iamProfile := ec2.IamInstanceProfileSpecification{Name: aws.String(iamProfileName)}

	start := time.Now()
	timeToWait := 60 //seconds
	retry := true

	region := cloud.Region
	ami := testInstanceMap[region]

	for retry && int64(time.Since(start).Seconds()) < int64(timeToWait) {

		//this dummy instance run , to check the success of RunInstance call
		//this is to ensure that iamProfile is properly propagated
		_, err := cloud.Client.RunInstances(&ec2.RunInstancesInput{
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
func getNetworkHost() string {
	return beego.AppConfig.String("network_url")

}
func (cloud *AWS) describeAmi(ami *string) ([]*ec2.BlockDeviceMapping, error) {
	var amis []*string
	var ebsVolumes []*ec2.BlockDeviceMapping
	amis = append(amis, ami)
	amiInput := &ec2.DescribeImagesInput{ImageIds: amis}
	res, err := cloud.Client.DescribeImages(amiInput)
	if err != nil {
		beego.Error(err)
		return ebsVolumes, err
	}

	if len(res.Images) <= 0 {
		return ebsVolumes, errors.New("AMI not available in selected region or AMI not shared with the user")
	}
	for _, ebs := range res.Images[0].BlockDeviceMappings {
		if ebs.VirtualName == nil {
			beego.Info(*ebs.DeviceName)
			ebsVolumes = append(ebsVolumes, ebs)
		}
	}
	beego.Info(res.GoString())
	return ebsVolumes, nil
}

/*func (cloud *AWS) createVolume(ids []*ec2.Instance, volume Volume, projectId string) error {

	for _, id := range ids {
		/*start := time.Now()
		timeToWait := 60 //seconds
		retry := true

		for retry && int64(time.Since(start).Seconds()) < int64(timeToWait) {

			err, state := cloud.checkInstanceState(*id.InstanceId, projectId)

		}
		input := ec2.CreateVolumeInput{
			VolumeType: aws.String(volume.VolumeType),
			Size:       aws.Int64(volume.VolumeSize),
		AvailabilityZone:aws.String(cloud.Region+"a")}


		if input.VolumeType ==aws.String("io1"){
			input.Iops=aws.Int64(volume.Iops)
		}
		out, err := cloud.Client.CreateVolume(&input)
		if err != nil {
			beego.Error(err.Error())
			return err
		}

		attach := ec2.AttachVolumeInput{VolumeId: out.VolumeId, InstanceId: id.InstanceId, Device: aws.String("/dev/sdf")}

		out_, err_ := cloud.Client.AttachVolume(&attach)

		if err_ != nil {
			beego.Error(err_.Error())
			return err_
		}
		beego.Info(out_.State)
	}
	return nil

}*/
func (cloud *AWS) checkInstanceState(id string, projectId string) (error, string) {
	ids := []*string{&id}
	latest_instances, err := cloud.GetInstances(ids, projectId, false)
	if err != nil {
		return err, ""
	} else {
		return nil, *latest_instances[0].PublicIpAddress
	}

}
func (cloud *AWS) getKey(pool NodePool, projectId string) (keyMaterial string, err error) {

	if pool.KeyInfo.KeyType == models.NEWKey {

		keyInfo, err := vault.GetSSHKey("aws", pool.KeyInfo.KeyName)

		if err != nil && err.Error() != "not found" {

			beego.Error(err.Error())
			logging.SendLog("Error in getting key: "+pool.KeyInfo.KeyName, "info", projectId)
			logging.SendLog(err.Error(), "info", projectId)
			return "", err

		} else if err == nil {
			key, err := keyCoverstion(keyInfo)
			if err != nil {
				return "", err
			}
			pool.KeyInfo = key
			if key.KeyMaterial != "" && key.KeyMaterial != " " {
				keyMaterial = key.KeyMaterial
			}
		} else if err != nil && err.Error() == "not found" {
			beego.Info("AWSOperations: creating key")
			logging.SendLog("Creating Key "+pool.KeyInfo.KeyName, "info", projectId)

			keyMaterial, _, err = cloud.KeyPairGenerator(pool.KeyInfo.KeyName)

			if err != nil {
				beego.Error(err.Error())
				logging.SendLog("Error in key creation: "+pool.KeyInfo.KeyName, "info", projectId)
				logging.SendLog(err.Error(), "info", projectId)
				return "", err
			}
			pool.KeyInfo.KeyMaterial = keyMaterial
			_, err = vault.PostSSHKey(pool.KeyInfo)

			if err != nil {
				beego.Error(err.Error())
				logging.SendLog("Error in key insertion: "+pool.KeyInfo.KeyName, "info", projectId)
				logging.SendLog(err.Error(), "info", projectId)
				return "", err
			}
		}
	} else if pool.KeyInfo.KeyType == models.CPKey {

		k, err := vault.GetSSHKey("aws", pool.KeyInfo.KeyName)

		if err != nil {
			beego.Error(err.Error())
			logging.SendLog("Error in getting key: "+pool.KeyInfo.KeyName, "info", projectId)
			logging.SendLog(err.Error(), "info", projectId)
			return "", err
		}
		key, err := keyCoverstion(k)
		if err != nil {
			return "", err
		}
		keyMaterial = key.KeyMaterial

	} else if pool.KeyInfo.KeyType == models.AWSKey { //not integrated

		_, err = vault.PostSSHKey(pool.KeyInfo)

		if err != nil {
			beego.Error(err.Error())
			logging.SendLog("Error in key insertion: "+pool.KeyInfo.KeyName, "info", projectId)
			logging.SendLog(err.Error(), "info", projectId)
			return "", err
		}
		keyMaterial = pool.KeyInfo.KeyMaterial

	} else if pool.KeyInfo.KeyType == models.USERKey { ///not integrated

		_, err = cloud.ImportSSHKeyPair(pool.KeyInfo.KeyName, pool.KeyInfo.KeyMaterial)

		if err != nil {
			beego.Error(err.Error())
			logging.SendLog("Error in importing key: "+pool.KeyInfo.KeyName, "info", projectId)
			logging.SendLog(err.Error(), "info", projectId)
			return "", err
		}

		_, err = vault.PostSSHKey(pool.KeyInfo)

		if err != nil {
			beego.Error(err.Error())
			logging.SendLog("Error in key insertion: "+pool.KeyInfo.KeyName, "info", projectId)
			logging.SendLog(err.Error(), "info", projectId)
			return "", err
		}
		keyMaterial = pool.KeyInfo.KeyMaterial
	}
	return keyMaterial, nil
}
func (cloud *AWS) ImportSSHKeyPair(key_name string, publicKey string) (string, error) {

	input := &ec2.ImportKeyPairInput{
		KeyName:           aws.String(key_name),
		PublicKeyMaterial: []byte(publicKey),
	}
	resp, err := cloud.Client.ImportKeyPair(input)
	if err != nil {
		beego.Error(err.Error())
		return "", err
	}
	beego.Info("key name", *resp.KeyName, "key fingerprint", *resp.KeyFingerprint)

	return *resp.KeyName, err
}
func (cloud *AWS) getAccountId() (string, error) {
	input := sts.GetCallerIdentityInput{}
	resp, err := cloud.STS.GetCallerIdentity(&input)
	if err != nil {
		beego.Error(err.Error())
		return "", err
	}
	return *resp.Account, nil

}

func (cloud *AWS) mountVolume(ids []*ec2.Instance, ami Ami, key Key, projectId string) error {

	for _, id := range ids {

		err := fileWrite(key.KeyMaterial, key.KeyName)
		if err != nil {
			return err
		}
		err = setPermission(key.KeyName)
		if err != nil {
			return err
		}
		publicIp := ""
		if id.PublicIpAddress == nil {
			beego.Error("waiting for public ip")
			time.Sleep(time.Second * 40)
			beego.Error("waited for public ip")
			err, publicIp = cloud.checkInstanceState(*id.InstanceId, projectId)
			if err != nil {
				return err
			}
		}
		err = copyFile(key.KeyName, ami.Username, publicIp)
		if err != nil {
			return err
		}
		err = runScript(key.KeyName, ami.Username, publicIp)
		if err != nil {
			return err
		}
		err = deleteScript(key.KeyName, ami.Username, publicIp)
		if err != nil {
			return err
		}
		err = deleteFile(key.KeyName)
		if err != nil {
			return err
		}
	}
	return nil

}

func fileWrite(key string, keyName string) error {

	f, err := os.Create("../antelope/keys/" + keyName + ".pem")
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	defer f.Close()
	d2 := []byte(key)
	n2, err := f.Write(d2)
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	beego.Info("wrote %d bytes\n", n2)

	err = os.Chmod("../antelope/keys/"+keyName+".pem", 0777)
	if err != nil {
		beego.Error(err)
		return err
	}
	return nil
}
func setPermission(keyName string) error {
	keyPath := "../antelope/keys/" + keyName + ".pem"
	cmd1 := "chmod"
	beego.Info(keyPath)
	args := []string{"600", keyPath}
	cmd := exec.Command(cmd1, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	return nil
}
func copyFile(keyName string, userName string, instanceId string) error {

	keyPath := "../antelope/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId + ":/home/" + userName
	cmd1 := "scp"
	beego.Info(keyPath)
	beego.Info(ip)
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, "../antelope/scripts/mount.sh", ip}
	cmd := exec.Command(cmd1, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	return nil
}
func runScript(keyName string, userName string, instanceId string) error {
	keyPath := "../antelope/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId
	cmd1 := "ssh"
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, ip, "/home/" + userName + "/mount.sh"}
	cmd := exec.Command(cmd1, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		beego.Warn(err.Error())
		return nil
	}
	return nil
}

func deleteScript(keyName string, userName string, instanceId string) error {
	keyPath := "../antelope/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId
	cmd1 := "ssh"
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, ip, "rm", "/home/" + userName + "/mount.sh"}
	cmd := exec.Command(cmd1, args...)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func deleteFile(keyName string) error {
	keyPath := "../antelope/keys/" + keyName + ".pem"
	err := os.Remove(keyPath)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}
