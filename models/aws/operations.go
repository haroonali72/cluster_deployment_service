package aws

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/aws/IAMRoles"
	autoscaling2 "antelope/models/aws/autoscaling"
	"antelope/models/key_utils"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var docker_master_policy = []byte(`{
  "Version": "2012-10-17",
  "Statement": [
	{
            "Effect": "Allow",
            "Action": [
                "autoscaling:DescribeAutoScalingGroups",
                "autoscaling:DescribeAutoScalingInstances",
                "autoscaling:DescribeLaunchConfigurations",
                "autoscaling:SetDesiredCapacity",
				"autoscaling:DescribeTags",
                "autoscaling:TerminateInstanceInAutoScalingGroup"
            ],
            "Resource": "*"
	 },
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

type CreatedPool struct {
	Instances []*ec2.Instance
	PoolName  string
}

type AWS struct {
	Client    *ec2.EC2
	AccessKey string
	SecretKey string
	Region    string
	Resources map[string]interface{}

	Scaler autoscaling2.AWSAutoScaler
	Roles  IAMRoles.AWSIAMRoles
}

func (cloud *AWS) createCluster(cluster Cluster_Def, ctx utils.Context, companyId string, token string) ([]CreatedPool, error) {

	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			return nil, err
		}
	}
	var awsNetwork types.AWSNetwork
	url := getNetworkHost("aws", cluster.ProjectId)
	network, err := api_handler.GetAPIStatus(token, url, ctx)

	/*bytes, err := json.Marshal(network)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}*/

	err = json.Unmarshal(network.([]byte), &awsNetwork)

	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}
	var createdPools []CreatedPool

	for _, pool := range cluster.NodePools {
		var createdPool CreatedPool
		keyMaterial, err := cloud.getKey(*pool, cluster.ProjectId, ctx, companyId, token)
		if err != nil {
			return nil, err
		}
		beego.Info("AWSOperations creating nodes")

		result, err, subnetId := cloud.CreateInstance(pool, awsNetwork, ctx)
		if err != nil {
			utils.SendLog(companyId, "Error in instances creation: "+err.Error(), "info", cluster.ProjectId)
			return nil, err
		}

		if result != nil && result.Instances != nil && len(result.Instances) > 0 {
			for index, instance := range result.Instances {
				err := cloud.updateInstanceTags(instance.InstanceId, pool.Name+"-"+strconv.Itoa(index), cluster.ProjectId, ctx)
				if err != nil {
					utils.SendLog(companyId, "Error in instances creation: "+err.Error(), "info", cluster.ProjectId)
					return nil, err
				}
			}
			if pool.IsExternal {
				pool.KeyInfo.KeyMaterial = keyMaterial
				err = cloud.mountVolume(result.Instances, pool.Ami, pool.KeyInfo, cluster.ProjectId, ctx, companyId)
				if err != nil {
					utils.SendLog(companyId, "Error in volume mounting : "+err.Error(), "info", cluster.ProjectId)
					return nil, err
				}
			}
			if pool.EnableScaling {

				err, m := cloud.Scaler.AutoScaler(pool.Name, *result.Instances[0].InstanceId, pool.Ami.AmiId, subnetId, pool.Scaling.MaxScalingGroupSize, ctx, cluster.ProjectId)

				if m[pool.Name+"_scale_autoScaler"] != "" {
					cloud.Resources[pool.Name+"_scale_autoScaler"] = pool.Name + "-scale"
				}
				if m[pool.Name+"_scale_launchConfig"] != "" {
					cloud.Resources[pool.Name+"_scale_launchConfig"] = pool.Name + "-scale"
				}
				if m[pool.Name+"_scale_role"] != "" {
					cloud.Resources[pool.Name+"_scale_role"] = pool.Name + "-scale"
				}
				if m[pool.Name+"_scale_policy"] != "" {
					cloud.Resources[pool.Name+"_scale_policy"] = pool.Name + "-scale"
				}
				if m[pool.Name+"_iamProfile"] != "" {
					cloud.Resources[pool.Name+"_scale_iamProfile"] = pool.Name + "-scale"
				}
				if err != nil {
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
			latest_instances, err = cloud.GetInstances(ids, cluster.ProjectId, true, ctx, companyId)
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

func (cloud *AWS) updateInstanceTags(instance_id *string, nodepool_name string, projectId string, ctx utils.Context) error {
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
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	ctx.SendLogs(out.String(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	return nil
}

func (cloud *AWS) init() error {
	if cloud.Client != nil {
		return nil
	}
	beego.Info(cloud.AccessKey)
	beego.Info(cloud.SecretKey)
	beego.Info(cloud.Region)
	if cloud.AccessKey == "" || cloud.SecretKey == "" || cloud.Region == "" {
		text := "invalid cloud credentials"
		beego.Error(text)
		return errors.New(text)
	}

	region := cloud.Region
	creds := credentials.NewStaticCredentials(cloud.AccessKey, cloud.SecretKey, "")
	cloud.Client = ec2.New(session.New(&aws.Config{Region: &region, Credentials: creds}))
	cloud.Resources = make(map[string]interface{})

	scaler := autoscaling2.AWSAutoScaler{
		AccessKey: cloud.AccessKey,
		SecretKey: cloud.SecretKey,
		Region:    cloud.Region,
	}
	confError := scaler.Init()
	if confError != nil {
		return confError
	}

	cloud.Scaler = scaler

	roles := IAMRoles.AWSIAMRoles{
		AccessKey: cloud.AccessKey,
		SecretKey: cloud.SecretKey,
		Region:    cloud.Region,
	}
	confError = roles.Init()
	if confError != nil {
		return confError
	}

	cloud.Roles = roles
	return nil
}

func (cloud *AWS) fetchStatus(cluster Cluster_Def, ctx utils.Context, companyId string, token string) (Cluster_Def, error) {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs("Failed to get latest status"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return Cluster_Def{}, err
		}
	}
	for in, pool := range cluster.NodePools {
		/*
			fetching aws scaled node
		*/
		if pool.EnableScaling {
			beego.Info("getting scaler nodes")
			err, instances := cloud.Scaler.GetAutoScaler(cluster.ProjectId, pool.Name, ctx)
			if err != nil {
				return Cluster_Def{}, err
			}
			if instances != nil {
				for _, inst := range instances {
					cluster.NodePools[in].Nodes = append(cluster.NodePools[in].Nodes, &Node{CloudId: *inst.InstanceId})
					beego.Info(*inst.InstanceId)
				}
			}
		}
		for index, node := range cluster.NodePools[in].Nodes {

			var nodeId []*string
			beego.Info(node.CloudId)
			nodeId = append(nodeId, &node.CloudId)
			out, err := cloud.GetInstances(nodeId, cluster.ProjectId, false, ctx, companyId)
			if err != nil {
				return Cluster_Def{}, err
			}
			if out != nil {
				cluster.NodePools[in].Nodes[index].NodeState = *out[0].State.Name

				if out[0].PublicIpAddress != nil {
					cluster.NodePools[in].Nodes[index].PublicIP = *out[0].PublicIpAddress
				}
				if out[0].PrivateDnsName != nil {
					cluster.NodePools[in].Nodes[index].PrivateDNS = *out[0].PrivateDnsName
				}
				if out[0].PublicDnsName != nil {
					cluster.NodePools[in].Nodes[index].PublicDNS = *out[0].PublicDnsName
				}
				if out[0].PrivateIpAddress != nil {
					cluster.NodePools[in].Nodes[index].PrivateIP = *out[0].PrivateIpAddress
				}
				if pool.Ami.Username != "" {
					cluster.NodePools[in].Nodes[index].UserName = pool.Ami.Username
				}
				for _, tag := range out[0].Tags {
					if *tag.Key == "Name" {
						cluster.NodePools[in].Nodes[index].Name = *tag.Value
					}
				}
			}
		}

		keyInfo, err := vault.GetSSHKey(string(models.AWS), pool.KeyInfo.KeyName, token, ctx)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return Cluster_Def{}, err
		}
		k, err := key_utils.AWSKeyCoverstion(keyInfo, ctx)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return Cluster_Def{}, err
		}
		cluster.NodePools[in].KeyInfo = k
	}
	return cluster, nil
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
func (cloud *AWS) terminateCluster(cluster Cluster_Def, ctx utils.Context, companyId string) error {
	if cloud.Client == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	roles := IAMRoles.AWSIAMRoles{
		AccessKey: cloud.AccessKey,
		SecretKey: cloud.SecretKey,
		Region:    cloud.Region,
	}
	confError := roles.Init()
	if confError != nil {
		ctx.SendLogs(confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return confError
	}

	for _, pool := range cluster.NodePools {
		if pool.EnableScaling {
			err := cloud.Scaler.DeleteAutoScaler(pool.Name)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return err
			}
			err = cloud.Scaler.DeleteConfiguration(pool.Name)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return err
			}
		}

		err := cloud.TerminatePool(pool, cluster.ProjectId, ctx, companyId)
		if err != nil {
			return err
		}
		err = cloud.Roles.DeleteIAMRole(pool.Name, ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
func (cloud *AWS) CleanUp(cluster Cluster_Def, ctx utils.Context) error {

	for _, pool := range cluster.NodePools {

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
			err := cloud.Roles.DeleteIAMProfile(name, ctx)
			if err != nil {
				return err
			}
		}

		if cloud.Resources[pool.Name+"_role"] != nil {
			beego.Info(cloud.Resources[pool.Name+"_role"])
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
			beego.Info(name)
			err := cloud.Roles.DeleteRole(name, ctx)
			if err != nil {
				return err
			}
		}

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
			err := cloud.Roles.DeletePolicy(name, ctx)
			if err != nil {
				return err
			}
		}
		if cloud.Resources[pool.Name+"_scale_launchConfig"] != nil {

			err := cloud.Scaler.DeleteConfiguration(pool.Name)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return err
			}
		}
		if cloud.Resources[pool.Name+"_scale_autoScaler"] != nil {

			err := cloud.Scaler.DeleteAutoScaler(pool.Name)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return err
			}
		}
		if cloud.Resources[pool.Name+"_scale_iamProfile"] != nil {

			iamProfile := cloud.Resources[pool.Name+"_scale_iamProfile"]
			name := ""
			b, e := json.Marshal(iamProfile)
			if e != nil {
				return e
			}
			e = json.Unmarshal(b, &name)
			if e != nil {
				return e
			}
			err := cloud.Roles.DeleteIAMProfile(name, ctx)
			if err != nil {
				return err
			}
		}

		if cloud.Resources[pool.Name+"_scale_role"] != nil {
			role := cloud.Resources[pool.Name+"_scale_role"]
			name := ""
			b, e := json.Marshal(role)
			if e != nil {
				return e
			}
			e = json.Unmarshal(b, &name)
			if e != nil {
				return e
			}
			err := cloud.Roles.DeleteRole(name, ctx)
			if err != nil {
				return err
			}
		}

		if cloud.Resources[pool.Name+"_scale_policy"] != nil {
			policy := cloud.Resources[pool.Name+"_scale_policy"]
			name := ""
			b, e := json.Marshal(policy)
			if e != nil {
				return e
			}
			e = json.Unmarshal(b, &name)
			if e != nil {
				return e
			}
			err := cloud.Roles.DeletePolicy(name, ctx)
			if err != nil {
				return err
			}
		}

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
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return err
			}
		}
	}
	/*	if cloud.Resources[cluster.ProjectId+"_launchConfig"] != nil {
			id := cloud.Resources[cluster.ProjectId+"_launchConfig"]
			name := ""
			b, e := json.Marshal(id)
			if e != nil {
				return e
			}
			e = json.Unmarshal(b, &name)
			if e != nil {
				return e
			}
			err := cloud.Scaler.DeleteConfiguration(cluster.ProjectId)
			if err != nil {
				ctx.SendSDLog(err.Error(), "error")
				return err
			}
		}
		if cloud.Resources[cluster.ProjectId+"_autoScaler"] != nil {
			id := cloud.Resources[cluster.ProjectId+"_autoScaler"]
			name := ""
			b, e := json.Marshal(id)
			if e != nil {
				return e
			}
			e = json.Unmarshal(b, &name)
			if e != nil {
				return e
			}
			err := cloud.Scaler.DeleteAutoScaler(cluster.ProjectId)
			if err != nil {
				ctx.SendSDLog(err.Error(), "error")
				return err
			}
		}
		if cloud.Resources[cluster.ProjectId+"_role"] != nil {
			role := cloud.Resources[cluster.ProjectId+"_role"]
			name := ""
			b, e := json.Marshal(role)
			if e != nil {
				return e
			}
			e = json.Unmarshal(b, &name)
			if e != nil {
				return e
			}
			err := cloud.Roles.DeleteRole(name, ctx)
			if err != nil {
				return err
			}
		}
		if cloud.Resources[cluster.ProjectId+"_policy"] != nil {
			policy := cloud.Resources[cluster.ProjectId+"_policy"]
			name := ""
			b, e := json.Marshal(policy)
			if e != nil {
				return e
			}
			e = json.Unmarshal(b, &name)
			if e != nil {
				return e
			}
			err := cloud.Roles.DeletePolicy(name, ctx)
			if err != nil {
				return err
			}
		}

		if cloud.Resources[cluster.ProjectId+"_iamProfile"] != "" {
			policy := cloud.Resources[cluster.ProjectId+"_iamProfile"]
			name := ""
			b, e := json.Marshal(policy)
			if e != nil {
				return e
			}
			e = json.Unmarshal(b, &name)
			if e != nil {
				return e
			}
			err := cloud.Roles.DeleteIAMProfile(name, ctx)
			if err != nil {
				return err
			}
		}*/

	return nil
}
func (cloud *AWS) CreateInstance(pool *NodePool, network types.AWSNetwork, ctx utils.Context) (*ec2.Reservation, error, string) {

	subnetId := cloud.GetSubnets(pool, network)
	sgIds := cloud.GetSecurityGroups(pool, network)
	//subnetId := "subnet-0fca4a3b97b1a1d8e"
	//sid := "sg-060473b5f6720b8e3"
	//sgIds := []*string{&sid}
	_, err := cloud.Roles.CreateRole(pool.Name)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err, ""
	}
	cloud.Resources[pool.Name+"_role"] = pool.Name
	_, err = cloud.Roles.CreatePolicy(pool.Name, docker_master_policy, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err, ""
	}
	cloud.Resources[pool.Name+"_policy"] = pool.Name
	_, err = cloud.Roles.CreateIAMProfile(pool.Name, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err, ""
	}
	cloud.Resources[pool.Name+"_iamProfile"] = pool.Name
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
	ebs, err := cloud.describeAmi(&pool.Ami.AmiId, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err, ""
	}
	if ebs != nil && ebs[0].Ebs != nil && ebs[0].Ebs.VolumeSize != nil {
		ebs[0].Ebs.VolumeSize = &pool.Ami.RootVolume.VolumeSize
		ebs[0].Ebs.VolumeType = &pool.Ami.RootVolume.VolumeType
		if pool.Ami.RootVolume.VolumeType == "io1" {
			ebs[0].Ebs.Iops = &pool.Ami.RootVolume.Iops
		}
	}
	ctx.SendLogs("attaching external volume", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	if pool.IsExternal {
		var external_volume ec2.BlockDeviceMapping

		var external_ebs ec2.EbsBlockDevice

		external_ebs.VolumeType = &pool.ExternalVolume.VolumeType
		external_ebs.VolumeSize = &pool.ExternalVolume.VolumeSize
		if pool.ExternalVolume.VolumeType == "io1" {
			external_ebs.Iops = &pool.ExternalVolume.Iops
		}
		if pool.ExternalVolume.DeleteOnTermination {
			external_ebs.DeleteOnTermination = aws.Bool(true)
		} else {
			external_ebs.DeleteOnTermination = aws.Bool(false)
		}
		external_volume.Ebs = &external_ebs
		external_volume.DeviceName = aws.String("/dev/sdf")
		ebs = append(ebs, &external_volume)
	}
	input.BlockDeviceMappings = ebs

	ok := cloud.Roles.CheckInstanceProfile(pool.Name)
	if !ok {
		iamProfile := ec2.IamInstanceProfileSpecification{Name: aws.String(pool.Name)}
		input.IamInstanceProfile = &iamProfile
	} else {
		ctx.SendLogs("failed in attaching", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	}

	result, err := cloud.Client.RunInstances(input)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err, ""
	}
	if result != nil && result.Instances != nil && len(result.Instances) > 0 {

		var ids []*string
		for _, instance := range result.Instances {
			ids = append(ids, aws.String(*instance.InstanceId))
		}
		cloud.Resources[pool.Name+"_instances"] = ids
	}

	return result, nil, subnetId

}
func (cloud *AWS) GetSecurityGroups(pool *NodePool, network types.AWSNetwork) []*string {
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
func (cloud *AWS) GetSubnets(pool *NodePool, network types.AWSNetwork) string {
	for _, definition := range network.Definition {
		for _, subnet := range definition.Subnets {
			if subnet.Name == pool.PoolSubnet {
				return subnet.SubnetId
			}
		}
	}
	return ""
}

func (cloud *AWS) GetInstances(ids []*string, projectId string, creation bool, ctx utils.Context, companyId string) (latest_instances []*ec2.Instance, err error) {

	instance_input := ec2.DescribeInstancesInput{InstanceIds: ids}
	updated_instances, err := cloud.Client.DescribeInstances(&instance_input)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	if updated_instances == nil || updated_instances.Reservations == nil || updated_instances.Reservations[0].Instances == nil {

		return nil, errors.New("Nodes not found")
	}
	for _, instance := range updated_instances.Reservations[0].Instances {
		if creation {
			utils.SendLog(companyId, "Instance created successfully: "+*instance.InstanceId, "info", projectId)
		}
		latest_instances = append(latest_instances, instance)
	}
	return latest_instances, nil

}
func (cloud *AWS) GetInstancesByDNS(privateDns []*string, projectId string, ctx utils.Context) (latest_instances []*ec2.Instance, err error) {

	//dns :=[]*string {privateDns}
	filters := []*ec2.Filter{&ec2.Filter{Name: aws.String("private-dns-name"), Values: privateDns}}

	instance_input := ec2.DescribeInstancesInput{Filters: filters}
	updated_instances, err := cloud.Client.DescribeInstances(&instance_input)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	if updated_instances == nil || updated_instances.Reservations == nil || updated_instances.Reservations[0].Instances == nil {

		return nil, errors.New("Nodes not found")
	}
	for _, instance := range updated_instances.Reservations[0].Instances {
		latest_instances = append(latest_instances, instance)
	}
	return latest_instances, nil

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
func (cloud *AWS) TerminatePool(pool *NodePool, projectId string, ctx utils.Context, companyId string) error {

	beego.Info("AWSOperations terminating nodes")
	instance_ids := cloud.getIds(pool)

	err := cloud.TerminateIns(instance_ids)
	if err != nil {
		ctx.SendLogs("Cluster model: Status - Failed to terminate node pool "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	utils.SendLog(companyId, "Cluster pool terminated successfully: "+pool.Name, models.LOGGING_LEVEL_INFO, projectId)
	return nil
}

/*func (cloud *AWS) GetNetworkStatus(projectId string,ctx logging.Context) (Network, error) {

	url := getNetworkHost()

	url = strings.Replace(url, "{cloud_provider}", "aws", -1)

	client := utils.InitReq()

	url = url + "/" + projectId

	req, err := utils.CreateGetRequest(url)
	if err != nil {
		ctx.SendSDLog( err.Error(),"error")
		return Network{}, err
	}

	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog( err.Error(),"error")
		return Network{}, err
	}

	defer response.Body.Close()

	var network Network

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.SendSDLog( err.Error(),"error")
		return Network{}, err
	}

	err = json.Unmarshal(contents, &network)
	if err != nil {
		ctx.SendSDLog( err.Error(),"error")
		return Network{}, err
	}
	return network, nil

}*/
func getKubeEngineHost() string {
	return beego.AppConfig.String("network_url")
}
func getNetworkHost(cloudType, projectId string) string {

	host := beego.AppConfig.String("network_url") + models.WeaselGetEndpoint

	if strings.Contains(host, "{cloud}") {
		host = strings.Replace(host, "{cloud}", cloudType, -1)
	}

	if strings.Contains(host, "{projectId}") {
		host = strings.Replace(host, "{projectId}", projectId, -1)
	}

	return host
}
func (cloud *AWS) describeAmi(ami *string, ctx utils.Context) ([]*ec2.BlockDeviceMapping, error) {
	var amis []*string
	var ebsVolumes []*ec2.BlockDeviceMapping
	amis = append(amis, ami)
	amiInput := &ec2.DescribeImagesInput{ImageIds: amis}
	res, err := cloud.Client.DescribeImages(amiInput)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
func (cloud *AWS) checkInstanceState(id string, projectId string, ctx utils.Context, companyId string) (error, string) {
	ids := []*string{&id}
	latest_instances, err := cloud.GetInstances(ids, projectId, false, ctx, companyId)
	if err != nil {
		return err, ""
	} else {
		if latest_instances[0].PublicIpAddress != nil {

			return nil, *latest_instances[0].PublicIpAddress
		} else {

			return errors.New("public ip not assigned"), ""
		}
	}
}
func (cloud *AWS) getKey(pool NodePool, projectId string, ctx utils.Context, companyId string, token string) (keyMaterial string, err error) {

	//if pool.KeyInfo.KeyType == models.NEWKey {
	keyInfo, err := vault.GetSSHKey(string(models.AWS), pool.KeyInfo.KeyName, token, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(companyId, "Error in getting key: "+pool.KeyInfo.KeyName, models.LOGGING_LEVEL_INFO, projectId)
		utils.SendLog(companyId, err.Error(), models.LOGGING_LEVEL_INFO, projectId)
		return "", err
	}
	key, err := key_utils.AWSKeyCoverstion(keyInfo, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(companyId, "Error in getting key: "+pool.KeyInfo.KeyName, models.LOGGING_LEVEL_INFO, projectId)
		utils.SendLog(companyId, err.Error(), models.LOGGING_LEVEL_INFO, projectId)
		return "", err
	}
	keyMaterial = key.KeyMaterial

	//} else if pool.KeyInfo.KeyType == models.AWSKey { //not integrated
	//	_, err = vault.PostSSHKey(pool.KeyInfo, ctx, token)
	//	if err != nil {
	//		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	//		utils.SendLog(companyId, "Error in key insertion: "+pool.KeyInfo.KeyName, models.LOGGING_LEVEL_INFO, projectId)
	//		utils.SendLog(companyId, err.Error(), models.LOGGING_LEVEL_INFO, projectId)
	//		return "", err
	//	}
	//	keyMaterial = pool.KeyInfo.KeyMaterial
	//} else if pool.KeyInfo.KeyType == models.USERKey { ///not integrated
	//
	//	_, err = cloud.ImportSSHKeyPair(pool.KeyInfo.KeyName, pool.KeyInfo.KeyMaterial)
	//
	//	if err != nil {
	//		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	//		utils.SendLog(companyId, "Error in importing key: "+pool.KeyInfo.KeyName, models.LOGGING_LEVEL_INFO, projectId)
	//		utils.SendLog(companyId, err.Error(), models.LOGGING_LEVEL_INFO, projectId)
	//		return "", err
	//	}
	//
	//	_, err = vault.PostSSHKey(pool.KeyInfo, pool.KeyInfo.KeyName,pool.KeyInfo.Cloud,ctx, token, "")
	//	if err != nil {
	//		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	//		utils.SendLog(companyId, "Error in key insertion: "+pool.KeyInfo.KeyName, models.LOGGING_LEVEL_INFO, projectId)
	//		utils.SendLog(companyId, err.Error(), models.LOGGING_LEVEL_INFO, projectId)
	//		return "", err
	//	}
	//	keyMaterial = pool.KeyInfo.KeyMaterial
	//}
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

func (cloud *AWS) mountVolume(ids []*ec2.Instance, ami Ami, key key_utils.AWSKey, projectId string, ctx utils.Context, companyId string) error {

	for _, id := range ids {
		err := fileWrite(key.KeyMaterial, key.KeyName)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
		err = setPermission(key.KeyName)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
		publicIp := ""
		if id.PublicIpAddress == nil {
			beego.Error("waiting for public ip")
			time.Sleep(time.Second * 50)
			beego.Error("waited for public ip")
			err, publicIp = cloud.checkInstanceState(*id.InstanceId, projectId, ctx, companyId)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return err
			}
		}

		start := time.Now()
		timeToWait := 60 //seconds
		retry := true
		var errCopy error

		for retry && int64(time.Since(start).Seconds()) < int64(timeToWait) {

			errCopy = copyFile(key.KeyName, ami.Username, publicIp)
			if errCopy != nil && strings.Contains(errCopy.Error(), "exit status 1") {

				beego.Info("time passed %6.2f sec\n", time.Since(start).Seconds())
				beego.Info("waiting 5 seconds before retry")
				time.Sleep(5 * time.Second)
			} else {
				retry = false
			}
		}
		if errCopy != nil {
			ctx.SendLogs(errCopy.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return errCopy
		}
		err = setScriptPermision(key.KeyName, ami.Username, publicIp)
		if err != nil {
			ctx.SendLogs(errCopy.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

			return err
		}
		err = runScript(key.KeyName, ami.Username, publicIp)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
		err = deleteScript(key.KeyName, ami.Username, publicIp)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
		err = deleteFile(key.KeyName)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}
	return nil

}
func (cloud *AWS) enableScaling(cluster Cluster_Def, ctx utils.Context, token string) error {

	for _, pool := range cluster.NodePools {
		if pool.EnableScaling {
			var awsNetwork types.AWSNetwork
			url := getNetworkHost("aws", cluster.ProjectId)
			network, err := api_handler.GetAPIStatus(token, url, ctx)

			/*bytes, err := json.Marshal(network)
			if err != nil {
				beego.Error(err.Error())
				return err
			}
			*/
			err = json.Unmarshal(network.([]byte), &awsNetwork)

			if err != nil {
				beego.Error(err.Error())
				return err
			}
			subnetId := cloud.GetSubnets(pool, awsNetwork)
			err, m := cloud.Scaler.AutoScaler(pool.Name, pool.Nodes[0].CloudId, pool.Ami.AmiId, subnetId, pool.Scaling.MaxScalingGroupSize, ctx, cluster.ProjectId)
			if err != nil {
				if m[cluster.ProjectId+"_scale_launchConfig"] != "" {

					err := cloud.Scaler.DeleteConfiguration(pool.Name)
					if err != nil {
						ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
						return err
					}
				}
				if m[cluster.ProjectId+"_scale_autoScaler"] != "" {

					err := cloud.Scaler.DeleteAutoScaler(pool.Name)
					if err != nil {
						ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
						return err
					}
				}
				if m[cluster.ProjectId+"_scale_role"] != "" {

					err := cloud.Roles.DeleteRole(pool.Name+"-scale", ctx)
					if err != nil {
						return err
					}
				}
				if m[cluster.ProjectId+"_scale_policy"] != "" {

					err := cloud.Roles.DeletePolicy(pool.Name+"-scale", ctx)
					if err != nil {
						return err
					}
				}

				if m[cluster.ProjectId+"_scale_iamProfile"] != "" {

					err := cloud.Roles.DeleteIAMProfile(pool.Name+"-scale", ctx)
					if err != nil {
						return err
					}
				}
				return err
			}
		}
	}
	return nil
}
func fileWrite(key string, keyName string) error {

	f, err := os.Create("/app/keys/" + keyName + ".pem")
	if err != nil {
		return err
	}
	defer f.Close()
	d2 := []byte(key)
	n2, err := f.Write(d2)
	if err != nil {
		return err
	}
	beego.Info("wrote %d bytes\n", n2)

	err = os.Chmod("/app/keys/"+keyName+".pem", 0777)
	if err != nil {
		return err
	}
	return nil
}
func setPermission(keyName string) error {
	keyPath := "/app/keys/" + keyName + ".pem"
	cmd1 := "chmod"
	beego.Info(keyPath)
	args := []string{"600", keyPath}
	cmd := exec.Command(cmd1, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
func copyFile(keyName string, userName string, instanceId string) error {

	keyPath := "/app/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId + ":/home/" + userName
	cmd1 := "scp"
	beego.Info(keyPath)
	beego.Info(ip)
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, "/app/scripts/mount.sh", ip}
	cmd := exec.Command(cmd1, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
func setScriptPermision(keyName string, userName string, instanceId string) error {
	keyPath := "/app/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId
	cmd1 := "ssh"
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, ip, "chmod 700 /home/" + userName + "/mount.sh"}
	cmd := exec.Command(cmd1, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return nil
	}
	return nil
}
func runScript(keyName string, userName string, instanceId string) error {
	keyPath := "/app/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId
	cmd1 := "ssh"
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, ip, "/home/" + userName + "/mount.sh"}
	cmd := exec.Command(cmd1, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return nil
	}
	return nil
}

func deleteScript(keyName string, userName string, instanceId string) error {
	keyPath := "/app/keys/" + keyName + ".pem"
	ip := userName + "@" + instanceId
	cmd1 := "ssh"
	args := []string{"-o", "StrictHostKeyChecking=no", "-i", keyPath, ip, "rm", "/home/" + userName + "/mount.sh"}
	cmd := exec.Command(cmd1, args...)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func deleteFile(keyName string) error {
	keyPath := "/app/keys/" + keyName + ".pem"
	err := os.Remove(keyPath)
	if err != nil {
		return err
	}
	return nil
}
func GenerateAWSKey(keyName string, credentials vault.AwsCredentials, token, teams string, ctx utils.Context) (string, error) {
	aws := AWS{
		AccessKey: credentials.AccessKey,
		SecretKey: credentials.SecretKey,
		Region:    credentials.Region,
	}
	confError := aws.init()
	if confError != nil {
		return "", confError
	}

	_, err := vault.GetSSHKey(string(models.AWS), keyName, token, ctx)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		beego.Error("Key Already Exist ")
		return "", err
	}

	keyMaterial, _, err := aws.KeyPairGenerator(keyName)

	if err != nil {
		return "", err
	}

	var keyInfo key_utils.AWSKey
	keyInfo.KeyName = keyName
	keyInfo.KeyMaterial = keyMaterial
	keyInfo.KeyType = models.NEWKey
	keyInfo.Cloud = models.AWS

	ctx.SendLogs("SSHKey Created. ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	_, err = vault.PostSSHKey(keyInfo, keyInfo.KeyName, keyInfo.Cloud, ctx, token, teams)
	if err != nil {
		beego.Error("vm creation failed with error: " + err.Error())
		return "", err
	}

	return keyMaterial, err
}
