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
	"strconv"
	"strings"
	"time"
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

func (cloud *EKS) CreateCluster(eksCluster *EKSCluster, token string, ctx utils.Context) types.CustomCPError {
	if cloud.Svc == nil {
		cloud.init()
	}

	err := Validate(*eksCluster)
	if err != nil {
		ctx.SendLogs(
			"EKS cluster validation for '"+eksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		cpErr := ApiError(errors.New("ProjectId or Cluster Name is missing"), "Cluster Info is missing", 500)
		return cpErr
	}

	//fetch aws network
	subnets, sgs, err := cloud.getAWSNetwork(token, ctx)
	if err != nil {
		ctx.SendLogs(
			"EKS cluster creation request for '"+eksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, "unable to fetch network against this application.\n"+err.Error(), "error", eksCluster.ProjectId)
		cpErr := ApiError(err, "unable to fetch network against this application", 500)
		return cpErr
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
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			utils.SendLog(ctx.Data.Company, err.Error(), "error", eksCluster.ProjectId)
			cpErr := ApiError(err, "KMSKey Creation Failed", 512)
			return cpErr
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
		secret := "secrets"
		var resources []*string
		resources = append(resources, &secret)
		eksCluster.EncryptionConfig.Resources = resources
	}
	/**/

	//create cluster IAM role
	eksCluster.RoleArn, _, err = cloud.createClusterIAMRole(eksCluster.ProjectId)
	if err != nil {
		ctx.SendLogs(
			"EKS cluster creation request for '"+eksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, err.Error(), "error", eksCluster.ProjectId)
		cpErr := ApiError(err, "IAM Role Creation Failed", 512)
		return cpErr
	}
	ctx.SendLogs(
		"EKS cluster creation: Cluster IAM role created for '"+eksCluster.Name+"'",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	if eksCluster.ResourcesVpcConfig.EndpointPrivateAccess == nil {
		flag := false
		eksCluster.ResourcesVpcConfig.EndpointPrivateAccess = &flag
	}
	if eksCluster.ResourcesVpcConfig.EndpointPublicAccess == nil {
		flag := false
		eksCluster.ResourcesVpcConfig.EndpointPublicAccess = &flag
	}

	//generate cluster create request
	if eksCluster.ResourcesVpcConfig.EndpointPrivateAccess == nil {
		cidr := "0.0.0.0/0"
		var cidrs []*string
		cidrs = append(cidrs, &cidr)
		eksCluster.ResourcesVpcConfig.PublicAccessCidrs = cidrs
	}

	clusterRequest := GenerateClusterCreateRequest(*eksCluster)
	/**/

	//submit cluster creation request to AWS
	time.Sleep(time.Second * 120)
	beego.Info("waited for role activation")
	var result *eks.CreateClusterOutput
	for {
		result, err = cloud.Svc.CreateCluster(clusterRequest)
		if err != nil && strings.Contains(err.Error(), "AccessDeniedException: status code: 403") {
			time.Sleep(time.Second * 60)
			continue
		} else if err != nil && !strings.Contains(err.Error(), "exists") {
			ctx.SendLogs(
				"EKS cluster creation request for '"+eksCluster.Name+"' failed: "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			utils.SendLog(ctx.Data.Company, err.Error(), "error", eksCluster.ProjectId)
			cpErr := ApiError(err, "EKS Cluster Creation Failed", 512)
			return cpErr
		} else {
			break
		}
	}
	ctx.SendLogs(
		"EKS cluster creation request sent for '"+eksCluster.Name+"'",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	/*if result != nil && result.Cluster != nil {
		eksCluster.OutputArn = result.Cluster.Arn
	}*/
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
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, err.Error(), "error", eksCluster.ProjectId)
		cpErr := ApiError(err, "EKS Cluster Creation Failed", 512)
		return cpErr
	}
	ctx.SendLogs(
		"EKS cluster creation: Cluster '"+eksCluster.Name+"' created, adding node groups",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	/**/
	result_, err := cloud.Svc.DescribeCluster(&eks.DescribeClusterInput{Name: aws.String(eksCluster.Name)})
	if err != nil {
		ctx.SendLogs(
			"EKS cluster creation request for '"+eksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, err.Error(), "error", eksCluster.ProjectId)
		cpErr := ApiError(err, "EKS Cluster Creation Failed", 512)
		return cpErr
	}
	if result_ != nil && result_.Cluster != nil {
		eksCluster.OutputArn = result.Cluster.Arn
		beego.Info(eksCluster.OutputArn)
	}

	//add node groups
	for in, nodePool := range eksCluster.NodePools {
		if nodePool != nil {
			err := cloud.addNodePool(nodePool, eksCluster.Name, subnets, sgs, ctx)
			if err != (types.CustomCPError{}) {
				return err
			}
			eksCluster.NodePools[in].PoolStatus = true
		}
	}
	/**/

	return types.CustomCPError{}
}

func (cloud *EKS) addNodePool(nodePool *NodePool, clusterName string, subnets []*string, sgs []*string, ctx utils.Context) types.CustomCPError {
	if nodePool == nil {
		return types.CustomCPError{}
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
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			utils.SendLog(ctx.Data.Company, err.Error(), "error", ctx.Data.ProjectId)
			cpErr := ApiError(err, "EKS Cluster Creation Failed", 512)
			return cpErr
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
	var err_ error
	//create node group IAM role
	nodePool.NodeRole, nodePool.RoleName, err_ = cloud.createNodePoolIAMRole(nodePool.NodePoolName)
	if err_ != nil {
		ctx.SendLogs(
			"EKS cluster creation request for cluster '"+clusterName+"', node group '"+nodePool.NodePoolName+"'"+err_.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		ctx.SendLogs(err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, err_.Error(), "error", ctx.Data.ProjectId)
		cpErr := ApiError(err_, "EKS Cluster Creation Failed", 512)
		return cpErr
	}
	ctx.SendLogs(
		"EKS cluster creation: NodePool IAM role created for cluster '"+clusterName+"', node group '"+nodePool.NodePoolName+"'",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	/**/

	nodePool.Subnets = subnets

	tags := make(map[string]*string)
	tags["name"] = &nodePool.NodePoolName
	nodePool.Tags = tags
	//generate cluster create request
	beego.Info("printin==== eks desired size +" + strconv.Itoa(int(*nodePool.ScalingConfig.DesiredSize)))
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
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, err.Error(), "error", ctx.Data.ProjectId)
		cpErr := ApiError(err, "EKS Cluster Creation Failed", 512)
		return cpErr
	} else if err != nil && strings.Contains(err.Error(), "exists") {
		ctx.SendLogs(
			"EKS node group '"+nodePool.NodePoolName+"' for cluster '"+clusterName+"' already exists.",
			models.LOGGING_LEVEL_INFO,
			models.Backend_Logging,
		)
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, err.Error(), "error", ctx.Data.ProjectId)
		cpErr := ApiError(err, "EKS Cluster Creation Failed", 512)
		return cpErr
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
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, err.Error(), "error", ctx.Data.ProjectId)
		cpErr := ApiError(err, "EKS Cluster Creation Failed", 512)
		return cpErr
	}
	ctx.SendLogs(
		"EKS cluster creation: node group '"+nodePool.NodePoolName+"' for cluster '"+clusterName+"' created",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)
	/**/

	return types.CustomCPError{}
}

func (cloud *EKS) DeleteCluster(eksCluster *EKSCluster, ctx utils.Context) types.CustomCPError {
	if eksCluster == nil {
		return types.CustomCPError{}
	}

	if cloud.Svc == nil {
		cloud.init()
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
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			utils.SendLog(ctx.Data.Company, err.Error()+"\n Nodepool Deletion Failed - "+nodePool.NodePoolName, "error", eksCluster.ProjectId)
			cpErr := ApiError(err, "NodePool Deletion Failed", 512)
			return cpErr
		}
		//delete extra resources
		if nodePool.RoleName != nil {
			err = cloud.deleteIAMRoleFromInstanceProfile(*nodePool.RoleName)
			if err != nil {
				ctx.SendLogs(
					"EKS delete IAM role for cluster '"+eksCluster.Name+"', node group '"+nodePool.NodePoolName+"' failed: "+err.Error(),
					models.LOGGING_LEVEL_ERROR,
					models.Backend_Logging,
				)
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				utils.SendLog(ctx.Data.Company, err.Error()+"\n Nodepool Deletion Failed - "+nodePool.NodePoolName, "error", eksCluster.ProjectId)
				cpErr := ApiError(err, "NodePool Deletion Failed", 512)
				return cpErr
			}
			err = cloud.deleteIAMRole(*nodePool.RoleName)
			if err != nil {
				ctx.SendLogs(
					"EKS delete IAM role for cluster '"+eksCluster.Name+"', node group '"+nodePool.NodePoolName+"' failed: "+err.Error(),
					models.LOGGING_LEVEL_ERROR,
					models.Backend_Logging,
				)
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				utils.SendLog(ctx.Data.Company, err.Error()+"\n Nodepool Deletion Failed - "+nodePool.NodePoolName, "error", eksCluster.ProjectId)
				cpErr := ApiError(err, "NodePool Deletion Failed", 512)
				return cpErr
			}
		}
		if nodePool.RemoteAccess != nil && nodePool.RemoteAccess.EnableRemoteAccess {
			err = cloud.deleteSSHKey(nodePool.RemoteAccess.Ec2SshKey)
			if err != nil {
				ctx.SendLogs(
					"EKS delete SSH key for cluster '"+eksCluster.Name+"', node group '"+nodePool.NodePoolName+"' failed: "+err.Error(),
					models.LOGGING_LEVEL_ERROR,
					models.Backend_Logging,
				)
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				utils.SendLog(ctx.Data.Company, err.Error()+"\n Nodepool Deletion Failed - "+nodePool.NodePoolName, "error", eksCluster.ProjectId)
				cpErr := ApiError(err, "NodePool Deletion Failed", 512)
				return cpErr
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
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, err.Error()+"\n Cluster Deletion Failed - "+eksCluster.Name, "error", eksCluster.ProjectId)
		cpErr := ApiError(err, "Cluster Deletion Failed", 512)
		return cpErr
	}
	/**/
	err = cloud.deleteIAMRoleFromInstanceProfile("eks-cluster-" + eksCluster.ProjectId)
	if err != nil {
		ctx.SendLogs(
			"EKS delete IAM role for cluster '"+eksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, err.Error()+"\n Cluster Deletion Failed - "+eksCluster.Name, "error", eksCluster.ProjectId)
		cpErr := ApiError(err, "Cluster Deletion Failed", 512)
		return cpErr
	}

	//delete extra resources
	err = cloud.deleteClusterIAMRole("eks-cluster-" + eksCluster.ProjectId)
	if err != nil {
		ctx.SendLogs(
			"EKS delete IAM role for cluster '"+eksCluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, err.Error()+"\n Cluster Deletion Failed - "+eksCluster.Name, "error", eksCluster.ProjectId)
		cpErr := ApiError(err, "Cluster Deletion Failed", 512)
		return cpErr
	}
	err = cloud.Svc.WaitUntilClusterDeleted(&eks.DescribeClusterInput{Name: aws.String(eksCluster.Name)})
	if err != nil {
		ctx.SendLogs(
			"EKS cluster '"+eksCluster.Name+"'deletion failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, err.Error()+"\n Cluster Deletion Failed - "+eksCluster.Name, "error", eksCluster.ProjectId)
		cpErr := ApiError(err, "Cluster Deletion Failed", 512)
		return cpErr
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
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				utils.SendLog(ctx.Data.Company, err.Error()+"\n Cluster Deletion Failed - "+eksCluster.Name, "error", eksCluster.ProjectId)
				cpErr := ApiError(err, "Cluster Deletion Failed", 512)
				return cpErr
			}
		}
	}
	/**/

	eksCluster.RoleName = nil
	eksCluster.RoleArn = nil
	eksCluster.OutputArn = nil

	return types.CustomCPError{}
}
func (cloud *EKS) CleanUpCluster(eksCluster *EKSCluster, ctx utils.Context) types.CustomCPError {
	if eksCluster == nil {
		return types.CustomCPError{}
	}

	if cloud.Svc == nil {
		cloud.init()
	}

	//try deleting all node groups first
	for _, nodePool := range eksCluster.NodePools {
		if nodePool.OutputArn != nil && *nodePool.OutputArn != "" {
			err := cloud.deleteNodePool(eksCluster.Name, nodePool.NodePoolName)
			if err != nil {
				ctx.SendLogs(
					"EKS delete node group for cluster '"+eksCluster.Name+"', node group '"+nodePool.NodePoolName+"' failed: "+err.Error(),
					models.LOGGING_LEVEL_ERROR,
					models.Backend_Logging,
				)
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				utils.SendLog(ctx.Data.Company, err.Error()+"\n Nodepool Deletion Failed - "+nodePool.NodePoolName, "error", eksCluster.ProjectId)
				cpErr := ApiError(err, "NodePool Deletion Failed", 512)
				return cpErr
			}
		}
		//delete extra resources
		if nodePool.NodeRole != nil && *nodePool.NodeRole != "" {
			err := cloud.deleteIAMRole(*nodePool.RoleName)
			if err != nil {
				ctx.SendLogs(
					"EKS delete IAM role for cluster '"+eksCluster.Name+"', node group '"+nodePool.NodePoolName+"' failed: "+err.Error(),
					models.LOGGING_LEVEL_ERROR,
					models.Backend_Logging,
				)
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				utils.SendLog(ctx.Data.Company, err.Error()+"\n Nodepool Deletion Failed - "+nodePool.NodePoolName, "error", eksCluster.ProjectId)
				cpErr := ApiError(err, "NodePool Deletion Failed", 512)
				return cpErr
			}
		}
		if nodePool.RemoteAccess != nil && nodePool.RemoteAccess.EnableRemoteAccess && nodePool.RemoteAccess.Ec2SshKey != nil && *nodePool.RemoteAccess.Ec2SshKey != "" {
			err := cloud.deleteSSHKey(nodePool.RemoteAccess.Ec2SshKey)
			if err != nil {
				ctx.SendLogs(
					"EKS delete SSH key for cluster '"+eksCluster.Name+"', node group '"+nodePool.NodePoolName+"' failed: "+err.Error(),
					models.LOGGING_LEVEL_ERROR,
					models.Backend_Logging,
				)
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				utils.SendLog(ctx.Data.Company, err.Error()+"\n Nodepool Deletion Failed - "+nodePool.NodePoolName, "error", eksCluster.ProjectId)
				cpErr := ApiError(err, "NodePool Deletion Failed", 512)
				return cpErr
			}
			nodePool.RemoteAccess.Ec2SshKey = nil
		}
		nodePool.RoleName = nil
		nodePool.NodeRole = nil
		nodePool.OutputArn = nil

	}
	/**/

	//try deleting cluster
	if eksCluster.OutputArn != nil && *eksCluster.OutputArn != "" {
		_, err := cloud.Svc.DeleteCluster(&eks.DeleteClusterInput{Name: aws.String(eksCluster.Name)})
		if err != nil {
			ctx.SendLogs(
				"EKS delete cluster for '"+eksCluster.Name+"' failed: "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			utils.SendLog(ctx.Data.Company, err.Error()+"\n Cluster Deletion Failed - "+eksCluster.Name, "error", eksCluster.ProjectId)
			cpErr := ApiError(err, "Cluster Deletion Failed", 512)
			return cpErr
		}

		eksCluster.OutputArn = nil
	}
	/**/

	//delete extra resources
	if eksCluster.RoleArn != nil {

		err := cloud.deleteClusterIAMRole("eks-cluster-" + eksCluster.ProjectId)
		if err != nil {
			ctx.SendLogs(
				"EKS delete IAM role for cluster '"+eksCluster.Name+"' failed: "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			utils.SendLog(ctx.Data.Company, err.Error()+"\n Cluster Deletion Failed - "+eksCluster.Name, "error", eksCluster.ProjectId)
			cpErr := ApiError(err, "Cluster Deletion Failed", 512)
			return cpErr
		}

		eksCluster.RoleName = nil
		eksCluster.RoleArn = nil
	}
	if eksCluster.EncryptionConfig != nil && eksCluster.EncryptionConfig.EnableEncryption {
		if eksCluster.EncryptionConfig.Provider != nil {
			err := cloud.scheduleKMSKeyDeletion(eksCluster.EncryptionConfig.Provider.KeyId)
			if err != nil {
				ctx.SendLogs(
					"EKS scheduling KMS key deletion for cluster '"+eksCluster.Name+"' failed: "+err.Error(),
					models.LOGGING_LEVEL_ERROR,
					models.Backend_Logging,
				)
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				utils.SendLog(ctx.Data.Company, err.Error()+"\n Cluster Deletion Failed - "+eksCluster.Name, "error", eksCluster.ProjectId)
				cpErr := ApiError(err, "Cluster Deletion Failed", 512)
				return cpErr
			}
		}
	}
	/**/

	return types.CustomCPError{}
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
			//subnets = append(subnets, aws.String("subnet-52864a1b"))
			//subnets = append(subnets, aws.String("subnet-a204dac5"))

			for _, sg := range def.SecurityGroups {
				sgs = append(sgs, aws.String(sg.SecurityGroupId))
			}
			//sgs = append(sgs, aws.String("sg-ab31bccd"))
			break
		}
	}

	if len(subnets) < 2 {
		return nil, nil, errors.New("no vpc found with at least 2 subnets")
	}

	return subnets, sgs, nil
}
func (cloud *EKS) createClusterIAMRole(projectId string) (*string, *string, error) {
	roleName := "eks-cluster-" + projectId

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
	managedPolicies := []string{
		"arn:aws:iam::aws:policy/AutoScalingFullAccess",
		"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
	}

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
func (cloud *EKS) createNodePoolIAMRole(nodePoolName string) (*string, *string, error) {
	roleName := "eks-worker-" + nodePoolName
	managedPolicies := []string{
		"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
		"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
		"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
		"arn:aws:iam::aws:policy/AutoScalingFullAccess",
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
func (cloud *EKS) dettachIAMPolicy(roleName, managedPolicy string) error {
	_, err := cloud.IAM.DetachRolePolicy(&iam.DetachRolePolicyInput{
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
	if err != nil {
		return err
	}
	err = cloud.Svc.WaitUntilNodegroupDeleted(&eks.DescribeNodegroupInput{ClusterName: aws.String(clusterName), NodegroupName: aws.String(nodePoolName)})
	if err != nil {
		return err
	}
	return nil
}
func (cloud *EKS) deleteIAMRoleFromInstanceProfile(roleName string) error {
	output, err := cloud.IAM.ListInstanceProfilesForRole(&iam.ListInstanceProfilesForRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return err
	}

	if output.InstanceProfiles == nil || output.InstanceProfiles[0].InstanceProfileName == nil {
		return nil
	}

	for _, names := range output.InstanceProfiles {
		if names.InstanceProfileName != nil {
			beego.Info("profile name := " + *names.InstanceProfileName)
		}
		for _, roles := range names.Roles {
			if roles.RoleName != nil {
				beego.Info("role name := " + *roles.RoleName)
			}
		}
	}

	_, err = cloud.IAM.RemoveRoleFromInstanceProfile(&iam.RemoveRoleFromInstanceProfileInput{
		RoleName:            aws.String(roleName),
		InstanceProfileName: output.InstanceProfiles[0].InstanceProfileName,
	})
	if err != nil {
		return err
	}

	return nil
}
func (cloud *EKS) deleteIAMRole(roleName string) error {
	managedPolicies := []string{
		"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
		"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
		"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
		"arn:aws:iam::aws:policy/AutoScalingFullAccess",
	}

	for _, managedPolicy := range managedPolicies {
		err := cloud.dettachIAMPolicy(roleName, managedPolicy)
		if err != nil {
			return err
		}
	}
	_, err := cloud.IAM.DeleteRole(&iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return err
	}

	return nil
}
func (cloud *EKS) deleteClusterIAMRole(roleName string) error {

	managedPolicies := []string{
		"arn:aws:iam::aws:policy/AutoScalingFullAccess",
		"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
	}

	for _, managedPolicy := range managedPolicies {
		err := cloud.dettachIAMPolicy(roleName, managedPolicy)
		if err != nil {
			return err
		}
	}
	_, err := cloud.IAM.DeleteRole(&iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return err
	}

	return nil
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
func (cloud *EKS) fetchStatus(cluster *EKSCluster, ctx utils.Context, companyId string) (EKSClusterStatus, types.CustomCPError) {

	var response EKSClusterStatus
	clusterInput := eks.DescribeClusterInput{Name: aws.String(cluster.Name)}
	clusterOutput, err := cloud.Svc.DescribeCluster(&clusterInput)
	if err != nil {

		ctx.SendLogs(
			"EKS cluster state request for '"+cluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, "unable to fetch cluster"+err.Error(), "error", cluster.ProjectId)
		cpErr := ApiError(err, "unable to fetch cluster status", 512)

		return EKSClusterStatus{}, cpErr
	}
	response.Name = clusterOutput.Cluster.Name
	response.Status = clusterOutput.Cluster.Status
	response.ClusterEndpoint = clusterOutput.Cluster.Endpoint
	response.KubeVersion = clusterOutput.Cluster.Version
	response.ClusterArn = clusterOutput.Cluster.Arn

	for _, pool := range cluster.NodePools {
		if pool.PoolStatus {

			//getting nodes
			nodes, cpErr := cloud.getNodes(pool.NodePoolName, ctx)
			if cpErr != (types.CustomCPError{}) {
				return EKSClusterStatus{}, cpErr
			}

			//getting pool details
			var poolResponse EKSPoolStatus
			poolInput := eks.DescribeNodegroupInput{ClusterName: aws.String(cluster.Name),
				NodegroupName: aws.String(pool.NodePoolName)}
			poolOutput, err := cloud.Svc.DescribeNodegroup(&poolInput)
			if err != nil {

				ctx.SendLogs(
					"EKS cluster state request for '"+cluster.Name+"' failed: "+err.Error(),
					models.LOGGING_LEVEL_ERROR,
					models.Backend_Logging,
				)
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				utils.SendLog(ctx.Data.Company, "unable to fetch cluster"+err.Error(), "error", cluster.ProjectId)
				cpErr := ApiError(err, "unable to fetch cluster status", 512)

				return EKSClusterStatus{}, cpErr
			}
			poolResponse.NodePoolArn = poolOutput.Nodegroup.NodegroupArn
			poolResponse.Name = poolOutput.Nodegroup.NodegroupName
			poolResponse.Status = poolOutput.Nodegroup.Status
			poolResponse.AMI = poolOutput.Nodegroup.AmiType

			var scaling AutoScaling

			scaling.DesiredSize = poolOutput.Nodegroup.ScalingConfig.DesiredSize
			scaling.MinCount = poolOutput.Nodegroup.ScalingConfig.MinSize
			scaling.MaxCount = poolOutput.Nodegroup.ScalingConfig.MaxSize
			scaling.AutoScale = true

			poolResponse.Scaling = scaling
			poolResponse.Name = poolOutput.Nodegroup.InstanceTypes[0]

			poolResponse.Nodes = nodes
			response.NodePools = append(response.NodePools, poolResponse)
		} else {
			var poolResponse EKSPoolStatus
			poolResponse.Name = &pool.NodePoolName
			poolResponse.Status = aws.String("new")
			poolResponse.AMI = pool.AmiType
			poolResponse.MachineType = pool.InstanceType

			var scaling AutoScaling

			scaling.DesiredSize = pool.ScalingConfig.DesiredSize
			scaling.MinCount = pool.ScalingConfig.MinSize
			scaling.MaxCount = pool.ScalingConfig.MaxSize
			scaling.AutoScale = pool.ScalingConfig.IsEnabled
			poolResponse.Scaling = scaling

			response.NodePools = append(response.NodePools, poolResponse)
		}

	}
	return response, types.CustomCPError{}

}
func (cloud *EKS) getEKSCluster(ctx utils.Context) ([]*string, types.CustomCPError) {
	clusterInput := eks.ListClustersInput{}
	clusterOutput, err := cloud.Svc.ListClusters(&clusterInput)
	if err != nil {
		ctx.SendLogs("Failed to get instances list "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Failed to get running eks instances", 512)

		return nil, cpErr
	}
	return clusterOutput.Clusters, types.CustomCPError{}
}
func (cloud *EKS) getNodes(poolName string, ctx utils.Context) ([]EKSNodesStatus, types.CustomCPError) {
	var nodes []EKSNodesStatus

	var values []*string
	values = append(values, &poolName)
	var tags []*ec2.Filter
	tag := ec2.Filter{Name: aws.String("tag:eks:nodegroup-name"), Values: values}
	tags = append(tags, &tag)

	instance_input := ec2.DescribeInstancesInput{Filters: tags}
	updated_instances, err := cloud.EC2.DescribeInstances(&instance_input)

	if err != nil {
		ctx.SendLogs(
			"EKS cluster state request for failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "unable to fetch cluster status", 512)
		return []EKSNodesStatus{}, cpErr
	}
	if updated_instances == nil || updated_instances.Reservations == nil || updated_instances.Reservations[0].Instances == nil {

		return nil, ApiError(errors.New("Error in fetching instance"), "Nodes not found", 512)
	}
	for _, instance := range updated_instances.Reservations {
		var node EKSNodesStatus

		node.Name = instance.Instances[0].InstanceId
		node.ID = instance.Instances[0].InstanceId
		node.State = instance.Instances[0].State.Name
		node.PrivateIP = instance.Instances[0].PrivateIpAddress
		node.PublicIP = instance.Instances[0].PublicIpAddress
		nodes = append(nodes, node)
	}
	return nodes, types.CustomCPError{}

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
func GetEKS(projectId string, credentials vault.AwsCredentials) EKS {
	return EKS{
		AccessKey: credentials.AccessKey,
		SecretKey: credentials.SecretKey,
		Region:    credentials.Region,
		ProjectId: projectId,
	}
}
func (cloud *EKS) UpdateLogging(name string, logging Logging, ctx utils.Context) types.CustomCPError {
	clusterRequest := GenerateClusterUpdateLoggingRequest(name, logging)
	_, err := cloud.Svc.UpdateClusterConfig(clusterRequest)
	if err != nil {
		ctx.SendLogs(
			"EKS running cluster logging update request of "+name+" failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return types.CustomCPError{
			StatusCode:  512,
			Error:       "Error in running cluster logging update",
			Description: err.Error(),
		}
	}

	oldCluster, err := GetPreviousEKSCluster(ctx)
	if err != nil {
		ctx.SendLogs(
			"EKS running cluster logging update request of "+name+" failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return types.CustomCPError{
			StatusCode:  512,
			Error:       "Error in running cluster logging update",
			Description: err.Error(),
		}
	}

	oldCluster.Logging = logging

	err = AddPreviousEKSCluster(oldCluster, ctx, true)
	if err != nil {
		ctx.SendLogs(
			"EKS running cluster logging update request of "+name+" failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return types.CustomCPError{
			StatusCode:  512,
			Error:       "Error in running cluster logging update",
			Description: err.Error(),
		}
	}
	return types.CustomCPError{}
}
func (cloud *EKS) UpdateNetworking(name string, network VpcConfigRequest, ctx utils.Context) types.CustomCPError {
	clusterRequest := GenerateClusterUpdateNetworkRequest(name, network)
	_, err := cloud.Svc.UpdateClusterConfig(clusterRequest)
	if err != nil {
		ctx.SendLogs(
			"EKS running cluster network update request of "+name+" failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return types.CustomCPError{
			StatusCode:  512,
			Error:       "Error in running cluster network update",
			Description: err.Error(),
		}
	}

	oldCluster, err := GetPreviousEKSCluster(ctx)
	if err != nil {
		ctx.SendLogs(
			"EKS running cluster network update request of "+name+" failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return types.CustomCPError{
			StatusCode:  512,
			Error:       "Error in running cluster logging update",
			Description: err.Error(),
		}
	}

	oldCluster.ResourcesVpcConfig = network

	err = AddPreviousEKSCluster(oldCluster, ctx, true)
	if err != nil {
		ctx.SendLogs(
			"EKS running cluster network update request of "+name+" failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return types.CustomCPError{
			StatusCode:  512,
			Error:       "Error in running cluster network update",
			Description: err.Error(),
		}
	}

	return types.CustomCPError{}
}
func (cloud *EKS) UpdateNodeConfig(clusterName, poolName string, scalingConfig NodePoolScalingConfig, ctx utils.Context) types.CustomCPError {
	clusterRequest := GeneratNodeConfigUpdateRequest(clusterName, poolName, scalingConfig)
	_, err := cloud.Svc.UpdateNodegroupConfig(clusterRequest)
	if err != nil {
		ctx.SendLogs(
			"EKS running cluster nodepool config update request of "+clusterName+" failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return types.CustomCPError{
			StatusCode:  512,
			Error:       "Error in running cluster nodepool config update",
			Description: err.Error(),
		}
	}
	oldCluster, err := GetPreviousEKSCluster(ctx)
	if err != nil {
		ctx.SendLogs(
			"EKS running cluster nodepool config update request of "+clusterName+" failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return types.CustomCPError{
			StatusCode:  512,
			Error:       "Error in running cluster logging update",
			Description: err.Error(),
		}
	}
	for ind, pools := range oldCluster.NodePools {
		if pools.NodePoolName == poolName {
			oldCluster.NodePools[ind].ScalingConfig = &scalingConfig
		}
	}

	err = AddPreviousEKSCluster(oldCluster, ctx, true)
	if err != nil {
		ctx.SendLogs(
			"EKS running cluster nodepool config update request of "+clusterName+" failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return types.CustomCPError{
			StatusCode:  512,
			Error:       "Error in running cluster network update",
			Description: err.Error(),
		}
	}
	return types.CustomCPError{}
}
func (cloud *EKS) UpdateClusterVersion(clusterName, version string, ctx utils.Context) types.CustomCPError {
	clusterRequest := GenerateUpdateClusterVersionRequest(clusterName, version)
	_, err := cloud.Svc.UpdateClusterVersion(clusterRequest)
	if err != nil {
		ctx.SendLogs(
			"EKS running cluster version update request of "+clusterName+" failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return types.CustomCPError{
			StatusCode:  512,
			Error:       "Error in running cluster version update",
			Description: err.Error(),
		}
	}
	oldCluster, err := GetPreviousEKSCluster(ctx)
	if err != nil {
		ctx.SendLogs(
			"EKS running cluster version update request of "+clusterName+" failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return types.CustomCPError{
			StatusCode:  512,
			Error:       "Error in running cluster logging update",
			Description: err.Error(),
		}
	}

	oldCluster.Version = &version

	err = AddPreviousEKSCluster(oldCluster, ctx, true)
	if err != nil {
		ctx.SendLogs(
			"EKS running cluster version update request of "+clusterName+" failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return types.CustomCPError{
			StatusCode:  512,
			Error:       "Error in running cluster network update",
			Description: err.Error(),
		}
	}
	return types.CustomCPError{}
}
func (cloud *EKS) GetClusterStatus(name string, ctx utils.Context) (EKSClusterStatus, types.CustomCPError) {
	var response EKSClusterStatus
	clusterInput := eks.DescribeClusterInput{Name: aws.String(name)}
	clusterOutput, err := cloud.Svc.DescribeCluster(&clusterInput)
	if err != nil {

		ctx.SendLogs(
			"EKS cluster state request for '"+name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, "unable to fetch cluster"+err.Error(), "error", ctx.Data.ProjectId)
		cpErr := ApiError(err, "unable to fetch cluster status", 512)

		return EKSClusterStatus{}, cpErr
	}
	response.Name = clusterOutput.Cluster.Name
	response.Status = clusterOutput.Cluster.Status
	response.ClusterEndpoint = clusterOutput.Cluster.Endpoint
	response.KubeVersion = clusterOutput.Cluster.Version
	response.ClusterArn = clusterOutput.Cluster.Arn
	return response, types.CustomCPError{}
}
