package eks

import (
	"antelope/models"
	"antelope/models/api_handler"
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
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	"strings"
)

type EKS struct {
	Svc       *eks.EKS
	IAM       *iam.IAM
	KMS       *kms.KMS
	EC2       *ec2.EC2
	AccessKey string
	SecretKey string
	Region    string
	Resources map[string]interface{}
	ProjectId string
}

func (cloud *EKS) CreateCluster(eksCluster *EKSCluster, token string, ctx utils.Context) error {
	if eksCluster == nil {
		return nil
	}
	if cloud.Svc == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	err := Validate(*eksCluster)
	if err != nil {
		ctx.SendLogs(
			"EKS cluster validation for '"+eksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	//fetch aws network
	subnets, sgs, err := cloud.getAWSNetwork(token, ctx)
	if err != nil {
		ctx.SendLogs(
			"EKS cluster creation request for '"+eksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}
	ctx.SendLogs(
		"EKS cluster creation: Subnets chosen for '"+eksCluster.Name+"'",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	eksCluster.ResourcesVpcConfig.SubnetIds = subnets
	eksCluster.ResourcesVpcConfig.SecurityGroupIds = sgs
	/**/

	//create KMS key if encryption is enabled
	if eksCluster.EncryptionConfig != nil && eksCluster.EncryptionConfig.EnableEncryption {
		keyArn, keyId, err := cloud.createKMSKey(eksCluster.Name)
		if err != nil {
			ctx.SendLogs(
				"EKS cluster creation request for '"+eksCluster.Name+"' failed: "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
		}
		ctx.SendLogs(
			"EKS cluster creation: KMS Key created for '"+eksCluster.Name+"'",
			models.LOGGING_LEVEL_INFO,
			models.Backend_Logging,
		)
		eksCluster.EncryptionConfig.Provider = &Provider{
			KeyArn: keyArn,
			KeyId:  keyId,
		}
	}
	/**/

	//create cluster IAM role
	eksCluster.RoleArn, eksCluster.RoleName, err = cloud.createClusterIAMRole(eksCluster.Name)
	if err != nil {
		ctx.SendLogs(
			"EKS cluster creation request for '"+eksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}
	ctx.SendLogs(
		"EKS cluster creation: Cluster IAM role created for '"+eksCluster.Name+"'",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	/**/

	//generate cluster create request
	clusterRequest := GenerateClusterCreateRequest(*eksCluster)
	/**/

	//submit cluster creation request to AWS
	result, err := cloud.Svc.CreateCluster(clusterRequest)
	if err != nil && !strings.Contains(err.Error(), "exists") {
		ctx.SendLogs(
			"EKS cluster creation request for '"+eksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	} else if err != nil && strings.Contains(err.Error(), "exists") {
		ctx.SendLogs(
			"EKS cluster '"+eksCluster.Name+"' already exists.",
			models.LOGGING_LEVEL_INFO,
			models.Backend_Logging,
		)
		return nil
	}
	ctx.SendLogs(
		"EKS cluster creation request sent for '"+eksCluster.Name+"'",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	if result != nil && result.Cluster != nil {
		eksCluster.OutputArn = result.Cluster.Arn
	}
	/**/

	//wait for cluster creation
	ctx.SendLogs(
		"EKS cluster creation: Waiting for cluster '"+eksCluster.Name+"' to become active",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	err = cloud.Svc.WaitUntilClusterActive(&eks.DescribeClusterInput{Name: aws.String(eksCluster.Name)})
	if err != nil {
		ctx.SendLogs(
			"EKS cluster creation request for '"+eksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}
	ctx.SendLogs(
		"EKS cluster creation: Cluster '"+eksCluster.Name+"' created, adding node groups",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	/**/

	//add node groups
	for _, nodePool := range eksCluster.NodePools {
		if nodePool != nil {
			err = cloud.addNodePool(nodePool, eksCluster.Name, sgs, ctx)
			if err != nil {
				return err
			}
		}
	}
	/**/

	return nil
}

func (cloud *EKS) addNodePool(nodePool *NodePool, clusterName string, sgs []*string, ctx utils.Context) (err error) {
	if nodePool == nil {
		return err
	}

	//create SSH key if remote access is enabled
	if nodePool.RemoteAccess != nil && nodePool.RemoteAccess.EnableRemoteAccess {
		keyName, err := cloud.createSSHKey(clusterName, nodePool.NodePoolName)
		if err != nil {
			ctx.SendLogs(
				"EKS cluster creation request for cluster '"+clusterName+"', node group '"+nodePool.NodePoolName+"'"+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
		}
		ctx.SendLogs(
			"EKS cluster creation: SSH Key created for cluster '"+clusterName+"', node group '"+nodePool.NodePoolName+"'",
			models.LOGGING_LEVEL_INFO,
			models.Backend_Logging,
		)
		nodePool.RemoteAccess.Ec2SshKey = keyName
		nodePool.RemoteAccess.SourceSecurityGroups = sgs
	}
	/**/

	//create node group IAM role
	nodePool.NodeRole, nodePool.RoleName, err = cloud.createNodePoolIAMRole(nodePool.NodePoolName)
	if err != nil {
		ctx.SendLogs(
			"EKS cluster creation request for cluster '"+clusterName+"', node group '"+nodePool.NodePoolName+"'"+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}
	ctx.SendLogs(
		"EKS cluster creation: NodePool IAM role created for cluster '"+clusterName+"', node group '"+nodePool.NodePoolName+"'",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	/**/

	//generate cluster create request
	nodePoolRequest := GenerateNodePoolCreateRequest(*nodePool, clusterName)
	/**/

	//submit cluster creation request to AWS
	result, err := cloud.Svc.CreateNodegroup(nodePoolRequest)
	if err != nil && !strings.Contains(err.Error(), "exists") {
		ctx.SendLogs(
			"EKS cluster creation request for cluster '"+clusterName+"', node group '"+nodePool.NodePoolName+"'"+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	} else if err != nil && strings.Contains(err.Error(), "exists") {
		ctx.SendLogs(
			"EKS node group '"+nodePool.NodePoolName+"' for cluster '"+clusterName+"' already exists.",
			models.LOGGING_LEVEL_INFO,
			models.Backend_Logging,
		)
		return err
	}
	ctx.SendLogs(
		"EKS cluster node group creation request sent for cluster '"+clusterName+"', node group '"+nodePool.NodePoolName+"'",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	if result != nil && result.Nodegroup != nil {
		nodePool.OutputArn = result.Nodegroup.NodegroupArn
	}
	/**/

	//wait for node group creation
	ctx.SendLogs(
		"EKS cluster creation: Waiting for node group '"+nodePool.NodePoolName+"' for cluster '"+clusterName+"' to become active",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	err = cloud.Svc.WaitUntilNodegroupActive(&eks.DescribeNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String(nodePool.NodePoolName),
	})
	if err != nil {
		ctx.SendLogs(
			"EKS cluster creation request for cluster '"+clusterName+"', node group '"+nodePool.NodePoolName+"'"+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}
	ctx.SendLogs(
		"EKS cluster creation: node group '"+nodePool.NodePoolName+"' for cluster '"+clusterName+"' created",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	/**/

	return err
}

func (cloud *EKS) DeleteCluster(eksCluster *EKSCluster, ctx utils.Context) error {
	if eksCluster == nil {
		return nil
	}

	if cloud.Svc == nil {
		err := cloud.init()
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	//try deleting all node groups first
	for _, nodePool := range eksCluster.NodePools {
		err := cloud.deleteNodePool(eksCluster.Name, nodePool.NodePoolName)
		if err != nil {
			ctx.SendLogs(
				"EKS delete node group for cluster '"+eksCluster.Name+"', node group '"+nodePool.NodePoolName+"' failed: "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
		}
		//delete extra resources
		err = cloud.deleteIAMRole(nodePool.RoleName)
		if err != nil {
			ctx.SendLogs(
				"EKS delete IAM role for cluster '"+eksCluster.Name+"', node group '"+nodePool.NodePoolName+"' failed: "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
		}
		if nodePool.RemoteAccess != nil && nodePool.RemoteAccess.EnableRemoteAccess {
			err = cloud.deleteSSHKey(nodePool.RemoteAccess.Ec2SshKey)
			if err != nil {
				ctx.SendLogs(
					"EKS delete SSH key for cluster '"+eksCluster.Name+"', node group '"+nodePool.NodePoolName+"' failed: "+err.Error(),
					models.LOGGING_LEVEL_ERROR,
					models.Backend_Logging,
				)
			}
			nodePool.RemoteAccess.Ec2SshKey = nil
		}
		nodePool.RoleName = nil
		nodePool.NodeRole = nil
		nodePool.OutputArn = nil
	}
	/**/

	//try deleting cluster
	_, err := cloud.Svc.DeleteCluster(&eks.DeleteClusterInput{Name: aws.String(eksCluster.Name)})
	if err != nil {
		ctx.SendLogs(
			"EKS delete cluster for '"+eksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
	}
	/**/

	//delete extra resources
	err = cloud.deleteIAMRole(eksCluster.RoleName)
	if err != nil {
		ctx.SendLogs(
			"EKS delete IAM role for cluster '"+eksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
	}
	if eksCluster.EncryptionConfig != nil && eksCluster.EncryptionConfig.EnableEncryption {
		if eksCluster.EncryptionConfig.Provider != nil {
			err = cloud.scheduleKMSKeyDeletion(eksCluster.EncryptionConfig.Provider.KeyId)
			if err != nil {
				ctx.SendLogs(
					"EKS scheduling KMS key deletion for cluster '"+eksCluster.Name+"' failed: "+err.Error(),
					models.LOGGING_LEVEL_ERROR,
					models.Backend_Logging,
				)
			}
		}
	}
	/**/

	eksCluster.RoleName = nil
	eksCluster.RoleArn = nil
	eksCluster.OutputArn = nil

	return nil
}

func (cloud *EKS) getAWSNetwork(token string, ctx utils.Context) ([]*string, []*string, error) {
	url := getNetworkHost(string(models.AWS), cloud.ProjectId)
	awsNetwork := types.AWSNetwork{}

	network, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		return nil, nil, err
	}

	err = json.Unmarshal(network.([]byte), &awsNetwork)
	if err != nil {
		return nil, nil, err
	}

	subnets := []*string{}
	sgs := []*string{}
	for _, def := range awsNetwork.Definition {
		if def != nil && len(def.Subnets) > 1 {
			for _, subnet := range def.Subnets {
				subnets = append(subnets, aws.String(subnet.SubnetId))
			}
			for _, sg := range def.SecurityGroups {
				sgs = append(sgs, aws.String(sg.SecurityGroupId))
			}
			break
		}
	}

	if len(subnets) < 2 {
		return nil, nil, errors.New("no vpc found with at least 2 subnets")
	}

	return subnets, sgs, nil
}

func (cloud *EKS) createClusterIAMRole(clusterName string) (*string, *string, error) {
	roleName := "eks-cluster-" + clusterName
	managedPolicy := "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
	trustedEntity := []byte(`{
	  "Version": "2012-10-17",
	  "Statement": [
		{
		  "Effect": "Allow",
		  "Principal": {
			"Service": "eks.amazonaws.com"
		  },
		  "Action": "sts:AssumeRole"
		}
	  ]
	}`)

	roleArn, err := cloud.createIAMRole(roleName, string(trustedEntity))
	if err != nil {
		return nil, nil, err
	}

	return roleArn, aws.String(roleName), cloud.attachIAMPolicy(roleName, managedPolicy)
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

func (cloud *EKS) createSSHKey(clusterName, nodePoolName string) (*string, error) {
	keyName := "eks-" + clusterName + "-" + nodePoolName
	_, err := cloud.EC2.CreateKeyPair(&ec2.CreateKeyPairInput{KeyName: aws.String(keyName)})
	if err != nil {
		return nil, err
	}

	return aws.String(keyName), nil
}

func (cloud *EKS) deleteNodePool(clusterName, nodePoolName string) error {
	_, err := cloud.Svc.DeleteNodegroup(&eks.DeleteNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String(nodePoolName),
	})

	return err
}

func (cloud *EKS) deleteIAMRole(roleName *string) error {
	_, err := cloud.IAM.DeleteRole(&iam.DeleteRoleInput{
		RoleName: roleName,
	})

	return err
}

func (cloud *EKS) scheduleKMSKeyDeletion(keyId *string) error {
	_, err := cloud.KMS.ScheduleKeyDeletion(&kms.ScheduleKeyDeletionInput{
		KeyId:               keyId,
		PendingWindowInDays: aws.Int64(7),
	})

	return err
}

func (cloud *EKS) deleteSSHKey(keyName *string) error {
	_, err := cloud.EC2.DeleteKeyPair(&ec2.DeleteKeyPairInput{
		KeyName: keyName,
	})

	return err
}

func (cloud *EKS) init() error {
	if cloud.Svc != nil {
		return nil
	}

	region := cloud.Region
	creds := credentials.NewStaticCredentials(cloud.AccessKey, cloud.SecretKey, "")
	cloud.Svc = eks.New(session.New(&aws.Config{Region: &region, Credentials: creds}))
	cloud.IAM = iam.New(session.New(&aws.Config{Region: &region, Credentials: creds}))
	cloud.KMS = kms.New(session.New(&aws.Config{Region: &region, Credentials: creds}))
	cloud.EC2 = ec2.New(session.New(&aws.Config{Region: &region, Credentials: creds}))

	return nil
}

func Validate(eksCluster EKSCluster) error {
	if eksCluster.ProjectId == "" {
		return errors.New("project id is required")
	} else if eksCluster.Name == "" {
		return errors.New("cluster name is required")
	}
	return nil
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

func GetEKS(projectId string, credentials vault.AwsCredentials) (EKS, error) {
	return EKS{
		AccessKey: credentials.AccessKey,
		SecretKey: credentials.SecretKey,
		Region:    credentials.Region,
		ProjectId: projectId,
	}, nil
}
