package main

import (
	"antelope/models"
	"antelope/models/eks"
	_ "antelope/routers"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	eks_ "github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	"os"
	"strings"
)

func SecretAuth(username, password string) bool {
	// TODO configure basic authentication properly
	return username == "username" && password == "password"
}
func EKStest() {
	inputJson := `{
    "project_id" : "application-h7fmcs",
    "cloud" : "eks",
    "node_pools" : [ 
        {
            "ami_type" : "AL2_x86_64_GPU",
            "disk_size" : 32,
            "instance_type" : "g4dn.4xlarge",
            "node_role" : null,
            "role_name" : null,
            "node_pool_name" : "application-h7fmcs-1001",
            "scaling_config" : {
                "desired_size" : 2,
                "max_size" : 3,
                "min_size" : 2
            },
            "subnets" : [],
            "tags" : {}
        }, 
        {
            "ami_type" : "AL2_x86_64_GPU",
            "disk_size" : 32,
            "instance_type" : "g4dn.4xlarge",
            "node_role" : null,
            "role_name" : null,
            "node_pool_name" : "application-h7fmcs-1111",
            "remote_access" : {
                "enable_remote_access" : false,
                "ec2_ssh_key" : null,
                "source_security_groups" : []
            },
            "scaling_config" : {
                "desired_size" : 1,
                "max_size" : 3,
                "min_size" : 2
            },
            "subnets" : [],
            "tags" : {}
        }
    ],
    "status" : "Cluster Creation Failed",
    "company_id" : "5d945edc2dcc2f00089d8476",
    "encryption_config" : {
        "enable_encryption" : true,
        "provider" : {
            "key_arn" : "arn:aws:kms:us-west-1:193819466102:key/a66eeadf-981c-4c53-921a-2bcc2607d6bf",
            "key_id" : "a66eeadf-981c-4c53-921a-2bcc2607d6bf"
        },
        "resources" : []
    },
    "logging" : {
        "enable_api" : false,
        "enable_audit" : false,
        "enable_authenticator" : false,
        "enable_controller_manager" : false,
        "enable_scheduler" : false
    },
    "name" : "cluster-sadaf-1",
    "resources_vpc_config" : {
        "endpoint_private_access" : null,
        "endpoint_public_access" : null,
        "public_access_cidrs" : [],
        "security_group_ids" : [ 
            "sg-031949c8c0614d7f2"
        ],
        "subnet_ids" : [ 
            "subnet-04752ebc389496165", 
            "subnet-05341bcf9fb21b764"
        ]
    },
    "role_arn" : "arn:aws:iam::193819466102:role/eks-cluster-application-h7fmcs",
    "role_name" : "eks-cluster-application-h7fmcs",
    "version" : "1.16"
}`
	var cluster eks.EKSCluster
	err := json.Unmarshal([]byte(inputJson), &cluster)
	if err != nil {
		beego.Error(err.Error())
		return
	}

	/*	var subnets []*string
		subnet1 := "subnet-52864a1b"
		subnet2 := "subnet-59c4ee1f"
		subnets = append(subnets, &subnet1)
		subnets = append(subnets, &subnet2)
		var sgs []*string
		sg1 := "sg-ab31bccd"
		sgs = append(sgs, &sg1)
		cluster.ResourcesVpcConfig.SubnetIds = subnets
		cluster.ResourcesVpcConfig.SecurityGroupIds = sgs


		var eksObj EKS
		eksObj.init()
		if cluster.EncryptionConfig != nil && cluster.EncryptionConfig.EnableEncryption {
			keyArn, keyId, err := eksObj.createKMSKey(cluster.Name)
			if err != nil {
				beego.Error(err.Error())
			}
			cluster.EncryptionConfig.Provider = &eks.Provider{
				KeyArn: keyArn,
				KeyId:  keyId,
			}
			a1:="secrets"
			var a []*string
			a= append(a,&a1)
			cluster.EncryptionConfig.Resources = a
		}

		role := "arn:aws:iam::193819466102:role/cloudplex-eks"
		cluster.RoleArn = &role
		//role = "cloudplex-eks"
		//cluster.RoleName = &role
		if cluster.ResourcesVpcConfig.EndpointPrivateAccess == nil {
			cidr := "0.0.0.0/0"
			var cidrs []*string
			cidrs = append(cidrs, &cidr)
			cluster.ResourcesVpcConfig.PublicAccessCidrs = cidrs
		}
		clusterRequest := eks.GenerateClusterCreateRequest(cluster)

		var result *eks_.CreateClusterOutput
		for {
			result, err = eksObj.Svc.CreateCluster(clusterRequest)
			if err != nil && strings.Contains(err.Error(), "AccessDeniedException") {
				beego.Error(err.Error() + "jhgkjhg")
				time.Sleep(time.Second * 60)
				continue
			}  else  if err!=nil && !strings.Contains(err.Error(),"exists"){
				beego.Error(err.Error())
				return
			} else {
				break
			}
		}
		beego.Info("EKS cluster creation request sent for '" + cluster.Name + "'")
	/*	if result != nil && result.Cluster != nil {
			cluster.OutputArn = result.Cluster.Arn
			beego.Info(cluster.OutputArn)
		}
	*/
	var eksObj EKS
	eksObj.init()
	/*err = eksObj.Svc.WaitUntilClusterActive(&eks_.DescribeClusterInput{Name: aws.String(cluster.Name)})
	if err != nil {
		beego.Error(err.Error())
		return
	}
	beego.Info(
		"EKS cluster creation: Cluster '"+cluster.Name+"' created, adding node groups",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)*/

	result_, err := eksObj.Svc.DescribeCluster(&eks_.DescribeClusterInput{Name: aws.String(cluster.Name)})
	if err != nil {
		beego.Error(err.Error())
		return
	}
	if result_ != nil && result_.Cluster != nil {
		cluster.OutputArn = result_.Cluster.Arn
		beego.Info(*cluster.OutputArn)
	}
	var sgs []*string
	sg1 := "sg-ab31bccd"
	sgs = append(sgs, &sg1)
	for index, nodePool := range cluster.NodePools {
		if nodePool != nil {
			var subnets []*string
			subnet1 := "subnet-52864a1b"
			subnet2 := "subnet-59c4ee1f"
			subnets = append(subnets, &subnet1)
			subnets = append(subnets, &subnet2)
			nodePool.Subnets = subnets
			beego.Info("======pool name=====" + nodePool.NodePoolName)
			beego.Info(index)
			err := eksObj.addNodePool(nodePool, cluster.Name, sgs)
			if err != nil {
				return
			}
		}
	}
}
func (cloud *EKS) createSSHKey(clusterName, nodePoolName string) (*string, error) {
	keyName := "eks-" + clusterName + "-" + nodePoolName
	_, err := cloud.EC2.CreateKeyPair(&ec2.CreateKeyPairInput{KeyName: aws.String(keyName)})
	if err != nil {
		return nil, err
	}

	return aws.String(keyName), nil
}
func (cloud *EKS) createNodePoolIAMRole(nodePoolName string) (*string, *string, error) {
	roleName := "eks-worker-" + nodePoolName
	managedPolicies := []string{
		"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
		"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
		"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
	}
	trustedEntity := []byte(`{
	  "Version": "2012-10-17",
	  "Statement": [
		{
		  "Effect": "Allow",
		  "Principal": {
			"Service": "ec2.amazonaws.com"
		  },
		  "Action": "sts:AssumeRole"
		}
	  ]
	}`)

	roleArn, err := cloud.createIAMRole(roleName, string(trustedEntity))
	if err != nil {
		return nil, nil, err
	}

	for _, managedPolicy := range managedPolicies {
		err = cloud.attachIAMPolicy(roleName, managedPolicy)
		if err != nil {
			return nil, nil, err
		}
	}

	return roleArn, aws.String(roleName), nil
}
func (cloud *EKS) createIAMRole(roleName, trustedEntity string) (*string, error) {
	result, err := cloud.IAM.CreateRole(&iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(trustedEntity),
		RoleName:                 aws.String(roleName),
	})

	if err != nil {
		return nil, err
	} else if result != nil {
		return result.Role.Arn, nil
	} else {
		return nil, errors.New("IAM Role ARN not found")
	}
}

func (cloud *EKS) attachIAMPolicy(roleName, managedPolicy string) error {
	_, err := cloud.IAM.AttachRolePolicy(&iam.AttachRolePolicyInput{
		PolicyArn: aws.String(managedPolicy),
		RoleName:  aws.String(roleName),
	})

	return err
}
func (cloud *EKS) addNodePool(nodePool *eks.NodePool, clusterName string, sgs []*string) error {
	beego.Info(nodePool.NodePoolName)
	//create SSH key if remote access is enabled
	if nodePool.RemoteAccess != nil && nodePool.RemoteAccess.EnableRemoteAccess {
		keyName, err := cloud.createSSHKey(clusterName, nodePool.NodePoolName)
		if err != nil {
			beego.Error(err.Error())
			return err
		}
		nodePool.RemoteAccess.Ec2SshKey = keyName
		nodePool.RemoteAccess.SourceSecurityGroups = sgs
	}
	/**/
	var err_ error
	//create node group IAM role
	nodePool.NodeRole, nodePool.RoleName, err_ = cloud.createNodePoolIAMRole(nodePool.NodePoolName)
	if err_ != nil {
		beego.Error(err_.Error())
		return err_

	}
	beego.Info(
		"EKS cluster creation: NodePool IAM role created for cluster '"+clusterName+"', node group '"+nodePool.NodePoolName+"'",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	/**/

	//generate cluster create request
	nodePoolRequest := eks.GenerateNodePoolCreateRequest(*nodePool, clusterName)
	/**/

	//submit cluster creation request to AWS
	beego.Info(*nodePool.ScalingConfig.DesiredSize)
	result, err := cloud.Svc.CreateNodegroup(nodePoolRequest)
	if err != nil && !strings.Contains(err.Error(), "exists") {
		beego.Error(err.Error())
		return err
	} else if err != nil && strings.Contains(err.Error(), "exists") {
		beego.Error(err.Error())
		return err
	}
	beego.Info(
		"EKS cluster node group creation request sent for cluster '"+clusterName+"', node group '"+nodePool.NodePoolName+"'",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	if result != nil && result.Nodegroup != nil {
		nodePool.OutputArn = result.Nodegroup.NodegroupArn
	}
	/**/

	//wait for node group creation
	beego.Info(
		"EKS cluster creation: Waiting for node group '"+nodePool.NodePoolName+"' for cluster '"+clusterName+"' to become active",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	err = cloud.Svc.WaitUntilNodegroupActive(&eks_.DescribeNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String(nodePool.NodePoolName),
	})
	if err != nil {

		beego.Error(err.Error())
	}
	beego.Info(
		"EKS cluster creation: node group '"+nodePool.NodePoolName+"' for cluster '"+clusterName+"' created",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	/**/
	return nil
}

type EKS struct {
	Svc *eks_.EKS
	IAM *iam.IAM
	KMS *kms.KMS
	EC2 *ec2.EC2
}

func (cloud *EKS) init() {
	region := "ap-southeast-1"
	creds := credentials.NewStaticCredentials("AKIAS2IEKLF3H2QLHEEA", "7JaZfNdZiIDHOkxGRELUfCsCJk6QsOEJTbJWtq0o", "")
	cloud.Svc = eks_.New(session.New(&aws.Config{Region: &region, Credentials: creds}))
	cloud.IAM = iam.New(session.New(&aws.Config{Region: &region, Credentials: creds}))
	cloud.KMS = kms.New(session.New(&aws.Config{Region: &region, Credentials: creds}))
	cloud.EC2 = ec2.New(session.New(&aws.Config{Region: &region, Credentials: creds}))

}

func (cloud *EKS) createKMSKey(clusterName string) (*string, *string, error) {
	result, err := cloud.KMS.CreateKey(&kms.CreateKeyInput{})
	if err != nil {
		return nil, nil, err
	}

	if result != nil {
		_, _ = cloud.KMS.CreateAlias(&kms.CreateAliasInput{
			AliasName:   aws.String("alias/eks-" + clusterName),
			TargetKeyId: result.KeyMetadata.KeyId,
		})
		if result.KeyMetadata != nil {
			return result.KeyMetadata.Arn, result.KeyMetadata.KeyId, nil
		}
	}

	return nil, nil, errors.New("KMS key ARN not found")
}
func main() {
	EKStest()
	//setEnv()
	/*utils.InitFlags()
	if !db.IsMongoAlive() {
		os.Exit(1)
	}
	beego.BConfig.AppName = "antelope"
	beego.BConfig.CopyRequestBody = true
	beego.BConfig.WebConfig.EnableDocs = true
	beego.BConfig.WebConfig.AutoRender = true
	beego.BConfig.RunMode = "dev"
	beego.BConfig.WebConfig.DirectoryIndex = true
	beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	beego.BConfig.Listen.HTTPPort = 9081

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Token", "Content-type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	//for getting azure resource-sku-list
	go aks.RunCronJob()

	// TODO enable basic authentication if required
	//authPlugin := auth.NewBasicAuthenticator(SecretAuth, "Authorization Required")
	//beego.InsertFilter("*", beego.BeforeRouter, authPlugin)

	beego.Run()*/
}
func setEnv() {

	os.Setenv("kill_bill_user", "admin")
	os.Setenv("kill_bill_password", "password")
	os.Setenv("kill_bill_secret_key", "cloudplex")
	os.Setenv("kill_bill_api_key", "cloudplex")
	os.Setenv("ca_cert", "/home/zunaira/Downloads/mongoCA.crt")
	os.Setenv("client_cert", "/home/zunaira/Downloads/antelope.crt")
	os.Setenv("client_pem", "/home/zunaira/Downloads/antelope.pem")
	os.Setenv("subscription_host", "35.246.150.221:30906")
	os.Setenv("rbac_url", "http://localhost:7777")
	os.Setenv("mongo_host", "cloudplex-mongodb.cloudplex-system.svc.cluster.local:27017,mongodb-secondary-0.cloudplex-mongodb-headless:27017,mongodb-arbiter-0.cloudplex-mongodb-headless:27017")
	//os.Setenv("mongo_host", "localhost:27017")

	os.Setenv("mongo_auth", "true")
	os.Setenv("mongo_db", "antelope")
	os.Setenv("mongo_user", "antelope")
	os.Setenv("mongo_pass", "DbSn3hAzJU6pPVRcn61apb3KDEKmcSb7Bl..")
	os.Setenv("mongo_aws_template_collection", "aws_template")
	os.Setenv("mongo_op_cluster_collection", "op_cluster")
	os.Setenv("mongo_do_cluster_collection", "do_cluster")
	os.Setenv("mongo_aws_cluster_collection", "aws_cluster")
	os.Setenv("mongo_azure_template_collection", "azure_template")
	os.Setenv("mongo_cluster_error_collection", "errors_cluster")
	os.Setenv("mongo_azure_cluster_collection", "azure_cluster")
	os.Setenv("mongo_gcp_template_collection", "gcp_template")
	os.Setenv("mongo_gcp_cluster_collection", "gcp_cluster")
	os.Setenv("mongo_doks_cluster_collection", "doks_cluster")
	os.Setenv("mongo_doks_template_collection", "doks_template")
	os.Setenv("mongo_gke_template_collection", "gke_template")
	os.Setenv("mongo_gke_cluster_collection", "gke_cluster")
	os.Setenv("mongo_aks_template_collection", "aks_template")
	os.Setenv("mongo_aks_cluster_collection", "aks_cluster")
	os.Setenv("mongo_iks_template_collection", "iks_template")
	os.Setenv("mongo_iks_cluster_collection", "iks_cluster")
	os.Setenv("mongo_default_template_collection", "default_template")
	os.Setenv("mongo_ssh_keys_collection", "ssh_key")
	os.Setenv("redis_url", "localhost:6379")
	os.Setenv("logger_url", "https://dapis.cloudplex.io")
	os.Setenv("network_url", "https://dapis.cloudplex.io")
	os.Setenv("vault_url", "http://localhost:5000")
	os.Setenv("raccoon_url", "http://localhost:8092")
	os.Setenv("vault_url", "http://localhost:8092")
	os.Setenv("raccoon_url", "http://localhost:5000")
	os.Setenv("jump_host_ip", "52.220.196.92")
	os.Setenv("jump_host_ssh_key", "/home/zunaira/Downloads/ahmad.txt")
	os.Setenv("jump_host_ip", "52.220.196.92")
	os.Setenv("woodpecker_url", "http://localhost:3300")

}
