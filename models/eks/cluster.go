package eks

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/db"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"antelope/models/woodpecker"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/r3labs/diff"
	"gopkg.in/mgo.v2/bson"
	"strconv"

	"strings"
	"time"
)

type EKSCluster struct {
	ID                 bson.ObjectId      `json:"-" bson:"_id,omitempty"`
	InfraId            string             `json:"infra_id" bson:"infra_id" validate:"required" description:"ID of infrastructure [required]"`
	Cloud              models.Cloud       `json:"cloud" bson:"cloud" validate:"eq=EKS|eq=eks"`
	CreationDate       time.Time          `json:"-" bson:"creation_date"`
	ModificationDate   time.Time          `json:"-" bson:"modification_date"`
	NodePools          []*NodePool        `json:"node_pools" bson:"node_pools" validate:"required,dive"`
	IsAdvanced         bool               `json:"is_advance" bson:"is_advance" description:"Cluster advance level settings possible value 'true' or 'false'"`
	Status             models.Type        `json:"status" bson:"status" validate:"eq=new|eq=New|eq=NEW|eq=Cluster Update Failed|eq=Cluster Creation Failed|eq=Cluster Terminated|eq=Cluster Created" description:"Status of cluster [required]"`
	CompanyId          string             `json:"company_id" bson:"company_id" description:"ID of compnay [optional]"`
	OutputArn          *string            `json:"output_arn" bson:"output_arn,omitempty"`
	EncryptionConfig   *EncryptionConfig  `json:"encryption_config,omitempty" bson:"encryption_config,omitempty" description:"Encryption Configurations [optional]"`
	Logging            Logging            `json:"logging" bson:"logging" description:"Logging Configurations [optional]"`
	Name               string             `json:"name" bson:"name" validate:"required" description:"Cluster name [required]"`
	ResourcesVpcConfig VpcConfigRequest   `json:"resources_vpc_config" bson:"resources_vpc_config" description:"Access Level Details [optional]`
	RoleArn            *string            `json:"role_arn" bson:"role_arn"`
	RoleName           *string            `json:"role_name" bson:"role_name"`
	Tags               map[string]*string `json:"-" bson:"tags,omitempty"`
	Version            *string            `json:"version,omitempty" bson:"version,omitempty" description:"Kubernetes Version [required]`
}

type EncryptionConfig struct {
	EnableEncryption bool      `json:"enable_encryption" bson:"enable_encryption" description:"Option to enable or disable encryption [optional]`
	Provider         *Provider `json:"-" bson:"provider"`
	Resources        []*string `json:"-" bson:"resources"`
}

type Provider struct {
	KeyArn *string `json:"-" bson:"key_arn"`
	KeyId  *string `json:"-" bson:"key_id"`
}

type Logging struct {
	EnableApi               bool `json:"enable_api" bson:"enable_api" description:"Enable/Disable api logging [optional]"`
	EnableAudit             bool `json:"enable_audit" bson:"enable_audit" description:"Enable/Disable api audit logging [optional]"`
	EnableAuthenticator     bool `json:"enable_authenticator" bson:"enable_authenticator" description:"Enable/Disable authenticator logging  [optional]"`
	EnableControllerManager bool `json:"enable_controller_manager" bson:"enable_controller_manager" description:"Enable/Disable controller logging [optional]"`
	EnableScheduler         bool `json:"enable_scheduler" bson:"enable_scheduler" description:"Enable/Disable scheduler logging [optional]"`
}

type LogSetup struct {
	Enabled *bool     `json:"enabled" bson:"enabled"`
	Types   []*string `json:"types" bson:"types"`
}

type VpcConfigRequest struct {
	EndpointPrivateAccess *bool     `json:"endpoint_private_access" bson:"endpoint_private_access" description:"Enable/Disable private access endpoint [optional] `
	EndpointPublicAccess  *bool     `json:"endpoint_public_access" bson:"endpoint_public_access" description:"Enable/Disable public access endpoint [optional]`
	PublicAccessCidrs     []*string `json:"public_access_cidrs" bson:"public_access_cidrs" description:"Cidrs for public access [required, if public access is enabled]`
	SecurityGroupIds      []*string `json:"-" bson:"security_group_ids"`
	SubnetIds             []*string `json:"-" bson:"subnet_ids"`
}

type NodePool struct {
	OutputArn     *string                `json:"output_arn" bson:"output_arn,omitempty"`
	AmiType       *string                `json:"ami_type,omitempty" bson:"ami_type" validate:"required" description:"AMI for nodepool [required]"`
	DiskSize      *int64                 `json:"disk_size,omitempty" bson:"disk_size,omitempty" description:"Size of disk for nodes in nodepool [optional]"`
	InstanceType  *string                `json:"instance_type,omitempty" bson:"instance_type" validate:"required" description:"Instance type for nodes [required]"`
	Labels        map[string]*string     `json:"-" bson:"labels,omitempty"`
	NodeRole      *string                `json:"node_role" bson:"node_role"`
	RoleName      *string                `json:"role_name" bson:"role_name"`
	NodePoolName  string                 `json:"node_pool_name" bson:"node_pool_name" validate:"required" description:"Node Pool Name [required]"`
	RemoteAccess  *RemoteAccessConfig    `json:"remote_access,omitempty" bson:"remote_access,omitempty" description:"Access Levels (private of public)[optional]"`
	ScalingConfig *NodePoolScalingConfig `json:"auto_scaling,omitempty" bson:"scaling_config,omitempty" description:"Scaling Configurations for nodepool [optional]"`
	Subnets       []*string              `json:"subnets" bson:"subnets"`
	Tags          map[string]*string     `json:"-" bson:"tags"`
	PoolStatus    bool                   `json:"pool_status,omitempty" bson:"pool_status,omitempty"`
}

type RemoteAccessConfig struct {
	EnableRemoteAccess   bool      `json:"enable_remote_access" bson:"enable_remote_access" description:"Enable Remote Access [optional]"`
	Ec2SshKey            *string   `json:"ec2_ssh_key" bson:"ec2_ssh_key"`
	SourceSecurityGroups []*string `json:"-" bson:"source_security_groups"`
}

type NodePoolScalingConfig struct {
	DesiredSize *int64 `json:"desired_size" bson:"desired_size"`
	MaxSize     *int64 `json:"max_scaling_group_size" bson:"max_scaling_group_size"`
	MinSize     *int64 `json:"min_scaling_group_size" bson:"min_scaling_group_size"`
	IsEnabled   bool   `json:"autoscale" bson:"autoscale"`
}

type EKSClusterStatus struct {
	InfraId         string          `json:"infra_id"`
	ClusterEndpoint *string         `json:"endpoint"`
	Name            *string         `json:"name"`
	Status          *string         `json:"status"`
	KubeVersion     *string         `json:"kubernetes_version"`
	ClusterArn      *string         `json:"cluster_arn"`
	NodePools       []EKSPoolStatus `json:"node_pools"`
}
type EKSPoolStatus struct {
	NodePoolArn *string `json:"pool_arn"`
	Name        *string `json:"name"`
	Status      *string `json:"status"`
	AMI         *string `json:"ami_type"`
	MachineType *string `json:"machine_type"`

	Scaling AutoScaling `json:"auto_scaling"`
	//MaxSize *int64           `json:"max_size"`
	Nodes []EKSNodesStatus `json:"nodes"`
}
type EKSNodesStatus struct {
	Name      *string `json:"name"`
	PublicIP  *string `json:"public_ip"`
	PrivateIP *string `json:"private_ip"`
	State     *string `json:"state"`
	ID        *string `json:"id"`
}
type AutoScaling struct {
	AutoScale   bool   `json:"auto_scale,omitempty"`
	MinCount    *int64 `json:"min_scaling_group_size,omitempty"`
	MaxCount    *int64 `json:"max_scaling_group_size,omitempty"`
	DesiredSize *int64 `json:"desired_size"`
}

func KubeVersions(ctx utils.Context) []string {
	var kubeVersions []string
	kubeVersions = append(kubeVersions, "1.16")
	kubeVersions = append(kubeVersions, "1.15")
	kubeVersions = append(kubeVersions, "1.14")
	return kubeVersions
}

type AMI struct {
	Key   string `json:"name"`
	Value string `json:"value"`
}

func GetAMIS() []AMI {
	var amis []AMI

	var ami AMI
	ami.Key = "Amazon Linux 2"
	ami.Value = "AL2_x86_64"

	var ami2 AMI
	ami2.Key = "Amazon Linux 2 GPU Enabled"
	ami2.Value = "AL2_x86_64_GPU"

	amis = append(amis, ami)
	amis = append(amis, ami2)
	return amis
}
func GetInstances(amiType string, ctx utils.Context) []string {
	var list []string

	if amiType == "AL2_x86_64" {

		list = append(list, "t3.micro")
		list = append(list, "t3.small")
		list = append(list, "t3.medium")
		list = append(list, "t3.large")
		list = append(list, "t3.xlarge")
		list = append(list, "t3.2xlarge")
		list = append(list, "t3a.micro")
		list = append(list, "t3a.small")
		list = append(list, "t3a.medium")
		list = append(list, "t3a.large")
		list = append(list, "t3a.xlarge")
		list = append(list, "t3a.2xlarge")
		list = append(list, "m5.large")
		list = append(list, "m5.xlarge")
		list = append(list, "m5.2xlarge")
		list = append(list, "m5.4xlarge")
		list = append(list, "m5.8xlarge")
		list = append(list, "m5.12xlarge")
		list = append(list, "m5a.large")
		list = append(list, "m5a.xlarge")
		list = append(list, "m5a.2xlarge")
		list = append(list, "m5a.4xlarge")
		list = append(list, "c5.large")
		list = append(list, "c5.xlarge")
		list = append(list, "c5.2xlarge")
		list = append(list, "c5.4xlarge")
		list = append(list, "c5.9xlarge")
		list = append(list, "r5.large")
		list = append(list, "r5.xlarge")
		list = append(list, "r5.2xlarge")
		list = append(list, "r5.4xlarge")
		list = append(list, "r5a.large")
		list = append(list, "r5a.xlarge")
		list = append(list, "r5a.2xlarge")
		list = append(list, "r5a.4xlarge")
	} else {
		list = append(list, "g4dn.xlarge")
		list = append(list, "g4dn.2xlarge")
		list = append(list, "g4dn.4xlarge")
		list = append(list, "g4dn.8xlarge")
		list = append(list, "g4dn.12xlarge")
		list = append(list, "p2.xlarge")
		list = append(list, "p2.8xlarge")
		list = append(list, "p2.16xlarge")
		list = append(list, "p3.2xlarge")
		list = append(list, "p3.8xlarge")
		list = append(list, "p3.16xlarge")
		list = append(list, "p3dn.24xlarge")
	}
	return list

}
func GetEKSCluster(InfraId string, companyId string, ctx utils.Context) (cluster EKSCluster, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs(
			"EKSGetClusterModel:  Get - Got error while connecting to the database: "+err1.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoEKSClusterCollection)
	err = c.Find(bson.M{"infra_id": InfraId, "company_id": companyId}).One(&cluster)
	if err != nil {
		ctx.SendLogs(
			"EKSGetClusterModel:  Get - Got error while fetching from database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err
	}

	return cluster, nil
}
func GetAllEKSCluster(data rbacAuthentication.List, ctx utils.Context) (clusters []EKSCluster, err error) {
	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs(
			"EKSGetAllClusterModel:  GetAll - Got error while connecting to the database: "+err1.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return clusters, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoEKSClusterCollection)
	err = c.Find(bson.M{"infra_id": bson.M{"$in": copyData}}).All(&clusters)
	if err != nil {
		ctx.SendLogs(
			"EKSGetAllClusterModel:  GetAll - Got error while fetching from database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return clusters, err
	}

	return clusters, nil
}
func GetNetwork(token, InfraId string, ctx utils.Context) error {

	url := getNetworkHost("aws", InfraId)

	_, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}
func AddEKSCluster(cluster EKSCluster, ctx utils.Context) error {
	_, err := GetEKSCluster(cluster.InfraId, cluster.CompanyId, ctx)
	if err == nil {
		text := fmt.Sprintf("EKSAddClusterModel:  Add - Cluster for infrastructure '%s' already exists in the database.", cluster.InfraId)
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"EKSAddClusterModel:  Add - Got error while connecting to the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}
	defer session.Close()

	if cluster.CreationDate.IsZero() {
		cluster.CreationDate = time.Now()
		cluster.ModificationDate = time.Now()
		if cluster.Status == "" {
			cluster.Status = "new"
		}
		cluster.Cloud = models.EKS
	}

	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoEKSClusterCollection, cluster)
	if err != nil {
		ctx.SendLogs(
			"EKSAddClusterModel:  Add - Got error while inserting cluster to the database:  "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func UpdateEKSCluster(cluster EKSCluster, ctx utils.Context) error {
	oldCluster, err := GetEKSCluster(cluster.InfraId, cluster.CompanyId, ctx)
	if err != nil {
		text := "EKSUpdateClusterModel:  Update - Cluster '" + cluster.Name + "' does not exist in the database: " + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	/*if oldCluster.Status == string(models.Deploying) {
		ctx.SendLogs(
			"EKSUpdateClusterModel:  Update - Cluster is in deploying state.",
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return errors.New("cluster is in deploying state")
	}
	if oldCluster.Status == string(models.Terminating) {
		ctx.SendLogs(
			"EKSUpdateClusterModel:  Update - Cluster is in terminating state.",
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return errors.New("cluster is in terminating state")
	}
	if strings.ToLower(oldCluster.Status) == strings.ToLower(string(models.ClusterCreated)) {
		ctx.SendLogs(
			"EKSUpdateClusterModel:  Update - Cluster is in running state.",
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return errors.New("cluster is in running state")
	}
	*/
	err = DeleteEKSCluster(cluster.InfraId, cluster.CompanyId, ctx)
	if err != nil {
		ctx.SendLogs(
			"EKSUpdateClusterModel:  Update - Got error deleting cluster "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	cluster.CreationDate = oldCluster.CreationDate
	cluster.ModificationDate = time.Now()

	err = AddEKSCluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs(
			"EKSUpdateClusterModel:  Update - Got error creating cluster "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func DeleteEKSCluster(InfraId, companyId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"EKSDeleteClusterModel:  Delete - Got error while connecting to the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoEKSClusterCollection)
	err = c.Remove(bson.M{"infra_id": InfraId, "company_id": companyId})
	if err != nil {
		ctx.SendLogs(
			"EKSDeleteClusterModel:  Delete - Got error while deleting from the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func DeployEKSCluster(cluster EKSCluster, credentials vault.AwsProfile, companyId string, token string, ctx utils.Context) types.CustomCPError {
	publisher := utils.Notifier{}
	publisher.Init_notifier()

	eksOps := GetEKS(cluster.InfraId, credentials.Profile)
	eksOps.init()

	utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.InfraId)

	cluster.Status = (models.Deploying)
	confError := UpdateEKSCluster(cluster, ctx)
	if confError != nil {

		utils.SendLog(companyId, confError.Error(), "error", cluster.InfraId)
		cpErr := ApiError(confError, "Error occurred while updating cluster status in database", 500)

		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: confError.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return cpErr
	}

	cpError := eksOps.CreateCluster(&cluster, token, ctx)

	if cpError != (types.CustomCPError{}) {
		utils.SendLog(ctx.Data.Company, "EKS CLuster Creation Failed", "error", cluster.InfraId)

		//if cluster.OutputArn != nil {

		eksOps.CleanUpCluster(&cluster, ctx)

		//}
		cluster.Status = models.ClusterCreationFailed
		confError := UpdateEKSCluster(cluster, ctx)
		if confError != nil {

			utils.SendLog(companyId, confError.Error(), "error", cluster.InfraId)
			cpErr := ApiError(confError, "Error occurred while updating cluster status in database", 500)
			err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpError)
			if err != nil {
				ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			utils.Publisher(utils.ResponseSchema{
				Status:  false,
				Message: confError.Error(),
				InfraId: cluster.InfraId,
				Token:   token,
				Action:  models.Create,
			}, ctx)
			return cpErr

		}
		utils.SendLog(companyId, "Cluster creation failed : "+cluster.Name, "error", cluster.InfraId)
		ctx.SendLogs("Cluster creation failed", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpError)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: cpError.Error + "\n" + cpError.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return cpError
	}
	/**
	  TODO : Add Agent Deployment Process.Due on @Ahmad.
	*/
	pubSub := publisher.Subscribe(ctx.Data.InfraId, ctx)
	confError = ApplyAgent(credentials, token, ctx, cluster.Name)
	if confError != nil {
		utils.SendLog(companyId, confError.Error(), "error", cluster.InfraId)

		cluster.Status = models.ClusterCreationFailed
		profile := vault.AwsProfile{Profile: credentials.Profile}
		_ = TerminateCluster(cluster, profile, cluster.InfraId, companyId, token, ctx)
		utils.SendLog(companyId, "Cleaning up resources", "info", cluster.InfraId)
		confError_ := UpdateEKSCluster(cluster, ctx)
		if confError_ != nil {
			utils.SendLog(companyId, confError_.Error(), "error", cluster.InfraId)
		}

		cpErr := ApiError(confError, "Error occurred while deploying agent", 500)
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("EKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: confError.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return cpErr
	}
	cluster.Status = models.ClusterCreated

	confError = UpdateEKSCluster(cluster, ctx)

	if confError != nil {

		utils.SendLog(companyId, confError.Error(), "error", cluster.InfraId)
		cpErr := ApiError(confError, "Error occurred while updating cluster status in database", 500)
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpError)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: confError.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return cpErr
	}
	utils.SendLog(companyId, "Cluster Created Sccessfully "+cluster.Name, "info", cluster.InfraId)
	notify := publisher.RecieveNotification(ctx.Data.InfraId, ctx, pubSub)
	if notify {
		ctx.SendLogs("EKSClusterModel:  Notification recieved from agent", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Cluster Created Sccessfully",
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
	} else {
		ctx.SendLogs("EKSClusterModel:  Notification not recieved from agent", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		cluster.Status = models.ClusterCreationFailed
		utils.SendLog(companyId, confError.Error(), "Notification not recieved from agent", cluster.InfraId)
		confError_ := UpdateEKSCluster(cluster, ctx)
		if confError_ != nil {
			ctx.SendLogs("EKSDeployClusterModel:"+confError_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		}
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.EKS, ctx, types.CustomCPError{Description: confError_.Error(), Error: confError_.Error(), StatusCode: 512})
		if err != nil {
			ctx.SendLogs("EKSDeployClusterModel:  Agent  - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Notification not recieved from agent",
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
	}

	return types.CustomCPError{}
}

func TerminateCluster(cluster EKSCluster, credentials vault.AwsProfile, InfraId, companyId, token string, ctx utils.Context) types.CustomCPError {
	/*	publisher := utils.Notifier{}
		publisher.Init_notifier()*/

	eksOps := GetEKS(InfraId, credentials.Profile)

	_, _, _, err1 := CompareClusters(ctx)
	if err1 != nil &&  !(strings.Contains(err1.Error(),"Nothing to update")){
		oldCluster,err := GetPreviousEKSCluster(ctx)
		if err != nil {
			utils.SendLog(ctx.Data.Company, err.Error(), "error", cluster.InfraId)
			cpErr := types.CustomCPError{Description: err.Error(), Error: "Error occurred while updating cluster status in database", StatusCode: 500}
			err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.EKS, ctx, cpErr)
			if err != nil {
				ctx.SendLogs("EKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			utils.Publisher(utils.ResponseSchema{
				Status:  false,
				Message: err.Error(),
				InfraId: cluster.InfraId,
				Token:   token,
				Action:  models.Terminate,
			}, ctx)
			return cpErr
		}
		err_ := UpdateEKSCluster(oldCluster, ctx)
		if err_ != nil {
			utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.InfraId)
			cpErr := types.CustomCPError{Description: err_.Error(), Error: "Error occurred while updating cluster status in database", StatusCode: 500}
			err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpErr)
			if err != nil {
				ctx.SendLogs("EKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			utils.Publisher(utils.ResponseSchema{
				Status:  false,
				Message: err_.Error(),
				InfraId: cluster.InfraId,
				Token:   token,
				Action:  models.Terminate,
			}, ctx)
			return cpErr
		}

	}

	cluster.Status = (models.Terminating)
	utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.InfraId)

	err_ := UpdateEKSCluster(cluster, ctx)
	if err_ != nil {

		utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.InfraId)
		cpErr := types.CustomCPError{Description: err_.Error(), Error: "Error occurred while updating cluster status in database", StatusCode: 500}
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err_.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Terminate,
		}, ctx)
		return cpErr
	}

	eksOps.init()

	cpErr := eksOps.DeleteCluster(&cluster, ctx)
	if cpErr != (types.CustomCPError{}) {

		utils.SendLog(companyId, "Cluster termination failed: "+cpErr.Description+cluster.Name, "error", cluster.InfraId)

		cluster.Status = models.ClusterTerminationFailed
		err := UpdateEKSCluster(cluster, ctx)
		if err != nil {
			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.InfraId)
			utils.SendLog(companyId, err.Error(), "error", cluster.InfraId)

		}
		err = db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: cpErr.Error + "\n" + cpErr.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Terminate,
		}, ctx)
		return cpErr
	}

	cluster.Status = models.ClusterTerminated
	err := UpdateEKSCluster(cluster, ctx)
	if err != nil {
		utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.InfraId)
		utils.SendLog(companyId, err.Error(), "error", cluster.InfraId)

		cpErr := ApiError(err, "Error occurred while updating cluster status in database", 500)
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Terminate,
		}, ctx)
		return cpErr

	}
	utils.SendLog(companyId, "Cluster terminated successfully "+cluster.Name, "info", cluster.InfraId)
	utils.Publisher(utils.ResponseSchema{
		Status:  true,
		Message: "Cluster terminated successfully",
		InfraId: cluster.InfraId,
		Token:   token,
		Action:  models.Terminate,
	}, ctx)
	return types.CustomCPError{}
}

func ValidateEKSData(cluster EKSCluster, ctx utils.Context) error {
	if cluster.InfraId == "" {

		return errors.New("infrastructure ID is empty")

	} else if cluster.Version == nil {

		return errors.New("kubernetes version is empty")

	}
	if cluster.ResourcesVpcConfig.EndpointPublicAccess == nil && cluster.ResourcesVpcConfig.EndpointPrivateAccess == nil {
		return errors.New("both private and public access cannot be false")
	}
	if cluster.ResourcesVpcConfig.EndpointPublicAccess != nil && *cluster.ResourcesVpcConfig.EndpointPublicAccess == false && cluster.ResourcesVpcConfig.EndpointPrivateAccess == nil && *cluster.ResourcesVpcConfig.EndpointPrivateAccess == false {
		return errors.New("both private and public access cannot be false")
	}
	for _, pool := range cluster.NodePools {

		if pool.NodePoolName == "" {

			return errors.New("Node Pool name is empty")

		} else if pool.AmiType != nil && *pool.AmiType == "" {

			return errors.New("Ami Type is empty")

		} else if (pool.AmiType != nil) && (*pool.AmiType != "AL2_x86_64" && *pool.AmiType != "AL2_x86_64_GPU") {

			return errors.New("Ami Type is incorrect")

		} else if pool.InstanceType != nil && *pool.InstanceType == "" {

			return errors.New("Ami Type is empty")

		}
	}

	return nil
}
func FetchStatus(credentials vault.AwsProfile, InfraId string, ctx utils.Context, companyId string, token string) (EKSClusterStatus, types.CustomCPError) {

	cluster, err := GetEKSCluster(InfraId, companyId, ctx)
	if err != nil {
		cpErr := ApiError(err, "Error occurred while getting cluster status in database", 500)
		return EKSClusterStatus{}, cpErr
	}
	if string(cluster.Status) == strings.ToLower(string(models.New)) {
		cpErr := types.CustomCPError{Error: "Unable to fetch status - Cluster is not deployed yet", Description: "Unable to fetch state - Cluster is not deployed yet", StatusCode: 409}
		return EKSClusterStatus{}, cpErr
	}

	if cluster.Status == models.Deploying || cluster.Status == models.Terminating || cluster.Status == models.ClusterTerminated {
		cpErr := ApiError(errors.New("Cluster is in "+
			string(cluster.Status)), "Cluster is in "+
			string(cluster.Status)+" state", 409)
		return EKSClusterStatus{}, cpErr
	}
	if cluster.Status != models.ClusterCreated {
		customErr, err := db.GetError(InfraId, companyId, models.IKS, ctx)
		if err != nil {
			cpErr := ApiError(err, "Error occurred while getting cluster status in database", 500)
			return EKSClusterStatus{}, cpErr
		}
		if customErr.Err != (types.CustomCPError{}) {
			return EKSClusterStatus{}, customErr.Err
		}
	}
	eks := GetEKS(cluster.InfraId, credentials.Profile)

	eks.init()

	response, e := eks.fetchStatus(&cluster, ctx, companyId)

	if e != (types.CustomCPError{}) {

		ctx.SendLogs("Cluster model: Status - Failed to get lastest status "+e.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return EKSClusterStatus{}, e
	}
	response.InfraId = InfraId

	return response, types.CustomCPError{}
}
func ApplyAgent(credentials vault.AwsProfile, token string, ctx utils.Context, clusterName string) (confError error) {
	companyId := ctx.Data.Company
	infraID := ctx.Data.InfraId
	data2, err := woodpecker.GetCertificate(infraID, token, ctx)
	if err != nil {
		ctx.SendLogs("EKSClusterModel : Apply Agent -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	filePath := "/tmp/" + companyId + "/" + infraID + "/"
	cmd := "mkdir -p " + filePath + " && echo '" + data2 + "'>" + filePath + "agent.yaml"
	output, err := models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("EKSClusterModel : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	cmd = "sudo docker run --rm --name " + companyId + infraID + " -e accessKey=" + credentials.Profile.AccessKey + " -e cluster=" + clusterName + " -e secretKey=" + credentials.Profile.SecretKey + " -e region=" + credentials.Profile.Region + " -e yamlFile=" + filePath + "agent.yaml -v " + filePath + ":" + filePath + " " + models.EKSAuthContainerName

	output, err = models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("AKSClusterModel : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func GetEKSClusters(InfraId string, credentials vault.AwsProfile, ctx utils.Context) ([]*string, types.CustomCPError) {
	eksOps := GetEKS(InfraId, credentials.Profile)
	eksOps.init()
	clusters, cpError := eksOps.getEKSCluster(ctx)
	if cpError != (types.CustomCPError{}) {
		return nil, cpError
	}
	return clusters, types.CustomCPError{}
}
func PatchRunningEKSCluster(cluster EKSCluster, credentials vault.AwsCredentials, token string, ctx utils.Context) (confError types.CustomCPError) {

	/*publisher := utils.Notifier{}
	publisher.Init_notifier()*/

	eks := GetEKS(cluster.InfraId, credentials)

	eks.init()
	utils.SendLog(ctx.Data.Company, "Updating running cluster : "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

	difCluster, _, _, err1 := CompareClusters(ctx)
	if err1 != nil {
		ctx.SendLogs("EKSUpdateRunningClusterModel:  Update - "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, err1.Error()+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

		if !strings.Contains(err1.Error(), "Nothing to update") {
			utils.SendLog(ctx.Data.Company, "Cluster updation failed"+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

			cluster.Status = models.ClusterUpdateFailed
			confError := UpdateEKSCluster(cluster, ctx)
			if confError != nil {
				ctx.SendLogs("EKSpdateRunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			err := ApiError(err1, "Error occured while apply cluster changes", 500)
			err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.EKS, ctx, err)
			if err_ != nil {
				ctx.SendLogs("GKEUpdateRunningClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			utils.Publisher(utils.ResponseSchema{
				Status:  true,
				Message: "Nothing to update",
				InfraId: cluster.InfraId,
				Token:   token,
				Action:  models.Update,
			}, ctx)
			return err
		}

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err1.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Update,
		}, ctx)
		return types.CustomCPError{}
	}

	/*	if previousPoolCount < newPoolCount {

			var pools []*NodePool
			for i := previousPoolCount; i < newPoolCount; i++ {
				pools = append(pools, cluster.NodePools[i])
			}

			err := AddNodepool(&cluster, ctx, eks, pools, previousPoolCount, token)
			if err != (types.CustomCPError{}) {
				utils.SendLog(ctx.Data.Company, "Cluster updation failed"+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

				cluster.Status = models.ClusterUpdateFailed
				confError := UpdateEKSCluster(cluster, ctx)
				if confError != nil {
					ctx.SendLogs("EKSpdateRunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				}
				//err := ApiError(err, "Error occured while apply cluster changes", 500)
				err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.EKS, ctx, err)
				if err_ != nil {
					ctx.SendLogs("EKSUpdateRunningClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				}
				publisher.Notify(ctx.Data.InfraId, "Redeploy Status Available", ctx)
				return err
			}

		} else if previousPoolCount > newPoolCount {

			previousCluster, err := GetPreviousEKSCluster(ctx)
			if err != nil {
				err_ := types.CustomCPError{Error: "Error in updating running cluster", StatusCode: 512, Description: err.Error()}
				return updationFailedError(cluster, ctx, err_)
			}
			for _, oldpool := range previousCluster.NodePools {
				delete := true
				for _, pool := range cluster.NodePools {
					if pool.NodePoolName == oldpool.NodePoolName {
						delete = false
						break
					}
				}
				if delete == true {
					err_ := DeleteNodepool(cluster, ctx, eks, oldpool.NodePoolName)
					if err_ != (types.CustomCPError{}) {
						return err_
					}
				}
			}
		}
	*/

	previousCluster, err := GetPreviousEKSCluster(ctx)
	if err != nil {
		err_ := types.CustomCPError{Error: "Error in updating running cluster", StatusCode: 512, Description: err.Error()}
		return updationFailedError(cluster, ctx, err_,token)
	}
	previousPoolCount := len(previousCluster.NodePools)

	addincluster := false
	var addpools []*NodePool
	var addedIndex []int
	for index, pool := range cluster.NodePools {
		existInPrevious := false
		for _, prePool := range previousCluster.NodePools {
			if pool.NodePoolName == prePool.NodePoolName {
				existInPrevious = true

			}
		}
		if existInPrevious == false {
			addpools = append(addpools, pool)
			addedIndex = append(addedIndex, index)
			addincluster = true
		}
	}
	if addincluster == true {
		err2 := AddNodepool(&cluster, ctx, eks, addpools, token)
		if err2 != (types.CustomCPError{}) {
			utils.SendLog(ctx.Data.Company, "Cluster updation failed"+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

			cluster.Status = models.ClusterUpdateFailed
			confError := UpdateEKSCluster(cluster, ctx)
			if confError != nil {
				ctx.SendLogs("EKSpdateRunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			//err := ApiError(err, "Error occured while apply cluster changes", 500)
			err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.EKS, ctx, err2)
			if err_ != nil {
				ctx.SendLogs("EKSUpdateRunningClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			utils.Publisher(utils.ResponseSchema{
				Status:  false,
				Message: err2.Error + "\n" + err2.Description,
				InfraId: cluster.InfraId,
				Token:   token,
				Action:  models.Update,
			}, ctx)
			return err2
		}
	}
	for _, prePool := range previousCluster.NodePools {
		existInNew := false
		for _, pool := range cluster.NodePools {
			if pool.NodePoolName == prePool.NodePoolName {
				existInNew = true
			}
		}
		if existInNew == false {
			DeleteNodepool(cluster, ctx, eks, prePool.NodePoolName,token)
		}

	}

	loggingChanges, scalingChange := false, false

	poolIndex_ := -1
	for _, dif := range difCluster {
		if dif.Type != "update" || len(dif.Path) < 1 {
			continue
		}
		currentpoolIndex_:= 0
		if len(dif.Path) > 2 {
			currentpoolIndex_, _ = strconv.Atoi(dif.Path[1])
			poolIndex, _ := strconv.Atoi(dif.Path[1])
			if poolIndex > (previousPoolCount - 1) {
				break
			}
			for _, index := range addedIndex {
				if index == poolIndex {
					continue
				}
			}
		}

		if dif.Path[0] == "Logging" && !loggingChanges {
			time.Sleep(time.Second * 120)
			utils.SendLog(ctx.Data.Company, "Applying logging changes on  cluster "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
			err := eks.UpdateLogging(cluster.Name, cluster.Logging, ctx)

			loggingChanges = true
			if err != (types.CustomCPError{}) {

				utils.SendLog(ctx.Data.Company, err.Description+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
				utils.SendLog(ctx.Data.Company, "Cluster updation failed"+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

				cluster.Status = models.ClusterUpdateFailed
				confError := UpdateEKSCluster(cluster, ctx)
				if confError != nil {
					ctx.SendLogs("EKSpdateRunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				}
				//err := ApiError(err, "Error occured while apply cluster changes", 500)
				err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.EKS, ctx, err)
				if err_ != nil {
					ctx.SendLogs("EKSUpdateRunningClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				}
				utils.Publisher(utils.ResponseSchema{
					Status:  false,
					Message: err.Error + "\n" + err.Description,
					InfraId: cluster.InfraId,
					Token:   token,
					Action:  models.Update,
				}, ctx)
				return err
			}
			utils.SendLog(ctx.Data.Company, "Logging changes applied successfully on cluster "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

		} else if dif.Path[0] == "ResourcesVpcConfig" {
			time.Sleep(time.Second * 120)
			utils.SendLog(ctx.Data.Company, "Applying network changes on cluster "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

			err := eks.UpdateNetworking(cluster.Name, cluster.ResourcesVpcConfig, ctx)
			if err != (types.CustomCPError{}) {

				utils.SendLog(ctx.Data.Company, err.Description+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
				utils.SendLog(ctx.Data.Company, "Cluster updation failed"+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

				cluster.Status = models.ClusterUpdateFailed
				confError := UpdateEKSCluster(cluster, ctx)
				if confError != nil {
					ctx.SendLogs("EKSpdateRunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				}
				//	err := ApiError(err1, "Error occured while apply cluster changes", 500)
				err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.EKS, ctx, err)
				if err_ != nil {
					ctx.SendLogs("EKSUpdateRunningClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				}
				utils.Publisher(utils.ResponseSchema{
					Status:  false,
					Message: err.Error + "\n" + err.Description,
					InfraId: cluster.InfraId,
					Token:   token,
					Action:  models.Update,
				}, ctx)
				return err
			}
			utils.SendLog(ctx.Data.Company, "Network changes applied successfully on cluster "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

		} else if dif.Path[0] == "Version" {
			time.Sleep(time.Second * 120)
			utils.SendLog(ctx.Data.Company, "Changing kubernetes version of cluster "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

			err := eks.UpdateClusterVersion(cluster.Name, *cluster.Version, ctx)
			if err != (types.CustomCPError{}) {

				utils.SendLog(ctx.Data.Company, err.Description+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
				utils.SendLog(ctx.Data.Company, "Cluster updation failed"+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

				cluster.Status = models.ClusterUpdateFailed
				confError := UpdateEKSCluster(cluster, ctx)
				if confError != nil {
					ctx.SendLogs("EKSpdateRunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				}
				//err := ApiError(err1, "Error occured while apply cluster changes", 500)
				err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.EKS, ctx, err)
				if err_ != nil {
					ctx.SendLogs("EKSUpdateRunningClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				}
				utils.Publisher(utils.ResponseSchema{
					Status:  false,
					Message: err.Error + "\n" + err.Description,
					InfraId: cluster.InfraId,
					Token:   token,
					Action:  models.Update,
				}, ctx)
				return err
			}
			utils.SendLog(ctx.Data.Company, "Kubernetes version updated of cluster "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

		} else if len(dif.Path) >= 3 && dif.Path[0] == "NodePools" && currentpoolIndex_ != poolIndex_ && dif.Path[2] == "ScalingConfig" && dif.Path[3] != "IsEnabled" && !scalingChange {
			time.Sleep(time.Second * 120)
			poolIndex, _ := strconv.Atoi(dif.Path[1])
			utils.SendLog(ctx.Data.Company, "Changing scaling config of nodepool "+cluster.NodePools[poolIndex].NodePoolName, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

			err := eks.UpdateNodeConfig(cluster.Name, cluster.NodePools[poolIndex].NodePoolName, *cluster.NodePools[poolIndex].ScalingConfig, ctx)
			if err != (types.CustomCPError{}) {

				utils.SendLog(ctx.Data.Company, err.Description+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
				utils.SendLog(ctx.Data.Company, "Cluster updation failed"+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

				cluster.Status = models.ClusterUpdateFailed
				confError := UpdateEKSCluster(cluster, ctx)
				if confError != nil {
					ctx.SendLogs("EKSpdateRunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				}
				//err := ApiError(err, "Error occured while apply cluster changes", 500)
				err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.EKS, ctx, err)
				if err_ != nil {
					ctx.SendLogs("EKSUpdateRunningClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				}
				utils.Publisher(utils.ResponseSchema{
					Status:  false,
					Message: err.Error + "\n" + err.Description,
					InfraId: cluster.InfraId,
					Token:   token,
					Action:  models.Update,
				}, ctx)
				return err
			}
			utils.SendLog(ctx.Data.Company, "Scaling config updated successfully", models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
			scalingChange = true
			currentpoolIndex_ = poolIndex_
		}

	}

	utils.SendLog(ctx.Data.Company, "Running Cluster updated successfully "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

	err = DeletePreviousEKSCluster(ctx)
	if err != nil {
		beego.Info("***********")
		beego.Info(err.Error())
	}
	/*cluster, err = GetEKSCluster(ctx.Data.InfraId, ctx.Data.Company, ctx)
	if err != nil {
		beego.Info("***********")
		beego.Info(err.Error())
	}*/

	/*	latestCluster, err2 := eks.GetClusterStatus(cluster.Name, ctx)
		if err2 != (types.CustomCPError{}) {
			return err2
		}

		beego.Info("*******" + *latestCluster.Status)
		for strings.ToLower(string(*latestCluster.Status)) != strings.ToLower("running") {
			time.Sleep(time.Second * 60)
		}*/
	cluster.Status = models.ClusterCreated
	err_update := UpdateEKSCluster(cluster, ctx)
	if err_update != nil {

		ctx.SendLogs("EKSpdateRunningClusterModel:  Update - "+err_update.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	}

	utils.Publisher(utils.ResponseSchema{
		Status:  true,
		Message: "Cluster updated successfully",
		InfraId: cluster.InfraId,
		Token:   token,
		Action:  models.Update,
	}, ctx)

	return types.CustomCPError{}

}

func AddPreviousEKSCluster(cluster EKSCluster, ctx utils.Context, patch bool) error {
	var oldCluster EKSCluster
	_, err := GetPreviousEKSCluster(ctx)
	if err == nil {
		err := DeletePreviousEKSCluster(ctx)
		if err != nil {
			ctx.SendLogs(
				"GKEAddClusterModel:  Add previous cluster - "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
		}
	}

	if patch == false {
		oldCluster, err = GetEKSCluster(ctx.Data.InfraId, ctx.Data.Company, ctx)
		if err != nil {
			ctx.SendLogs(
				"GKEAddClusterModel:  Add previous cluster - "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
		}
	} else {
		oldCluster = cluster
	}
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"GKEAddClusterModel:  Add previous cluster - "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	defer session.Close()

	if cluster.CreationDate.IsZero() {
		cluster.CreationDate = time.Now()
		cluster.ModificationDate = time.Now()
		cluster.Cloud = models.EKS
		cluster.CompanyId = ctx.Data.Company
	}

	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoEKSPreviousClusterCollection, oldCluster)
	if err != nil {
		ctx.SendLogs(
			"GKEAddClusterModel:  Add previous cluster -  "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func GetPreviousEKSCluster(ctx utils.Context) (cluster EKSCluster, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs(
			"GKEGetClusterModel:  Get previous cluster - Got error while connecting to the database: "+err1.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoEKSPreviousClusterCollection)
	err = c.Find(bson.M{"infra_id": ctx.Data.InfraId, "company_id": ctx.Data.Company}).One(&cluster)
	if err != nil {
		ctx.SendLogs(
			"GKEGetClusterModel:  Get previous cluster- Got error while fetching from database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err
	}

	return cluster, nil
}

func UpdatePreviousEKSCluster(cluster EKSCluster, ctx utils.Context) error {

	err := AddPreviousEKSCluster(cluster, ctx, false)
	if err != nil {
		text := "EKSClusterModel:  Update  previous cluster - " + cluster.Name + " " + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	err = UpdateEKSCluster(cluster, ctx)
	if err != nil {
		text := "EKSClusterModel:  Update previous cluster - " + cluster.Name + " " + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		err = DeletePreviousEKSCluster(ctx)
		if err != nil {
			text := "EKSDeleteClusterModel:  Delete  previous cluster - " + cluster.Name + " " + err.Error()
			ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return errors.New(text)
		}
		return err
	}

	return nil
}

func DeletePreviousEKSCluster(ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"GKEDeleteClusterModel:  Delete  previous cluster - "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoEKSPreviousClusterCollection)
	err = c.Remove(bson.M{"infra_id": ctx.Data.InfraId, "company_id": ctx.Data.Company})
	if err != nil {
		ctx.SendLogs(
			"GKEDeleteClusterModel:  Delete  previous cluster - "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}
func CompareClusters(ctx utils.Context) (diff.Changelog, int, int, error) {
	cluster, err := GetEKSCluster(ctx.Data.InfraId, ctx.Data.Company, ctx)
	if err != nil {

		return diff.Changelog{}, 0, 0, errors.New("error in getting eks cluster")
	}

	oldCluster, err := GetPreviousEKSCluster(ctx)
	if err != nil && strings.Contains(err.Error(), "not found") {
		return diff.Changelog{}, 0, 0, errors.New("Nothing to update")
	}

	previousPoolCount := len(oldCluster.NodePools)
	newPoolCount := len(cluster.NodePools)

	difCluster, err := diff.Diff(oldCluster, cluster)
	if len(difCluster) < 2 && previousPoolCount == newPoolCount {
		return diff.Changelog{}, 0, 0, errors.New("Nothing to update")
	} else if err != nil {
		return diff.Changelog{}, 0, 0, errors.New("Error in comparing differences:" + err.Error())
	}
	return difCluster, previousPoolCount, newPoolCount, nil
}
func AddNodepool(cluster *EKSCluster, ctx utils.Context, eksOps EKS, pools []*NodePool, token string) types.CustomCPError {
	/*/
	  Fetching network
	*/
	subnets, sgs, err := eksOps.getAWSNetwork(token, ctx)
	if err != nil {
		ctx.SendLogs(
			"EKS cluster creation request for '"+cluster.Name+"' failed: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, "unable to fetch network against this application.\n"+err.Error(), "error", cluster.InfraId)
		cpErr := ApiError(err, "unable to fetch network against this application", 500)
		return cpErr
	}
	ctx.SendLogs(
		"EKS cluster creation: Subnets chosen for '"+cluster.Name+"'",
		models.LOGGING_LEVEL_INFO,
		models.Backend_Logging,
	)

	for _, pool := range pools {
		utils.SendLog(ctx.Data.Company, "Adding nodepool "+pool.NodePoolName, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

		err := eksOps.addNodePool(pool, cluster.Name, subnets, sgs, ctx)
		if err != (types.CustomCPError{}) {
			if pool.NodeRole != nil && *pool.NodeRole != "" {
				err := eksOps.deleteIAMRole(*pool.RoleName)
				if err != nil {
					ctx.SendLogs(
						err.Error(),
						models.LOGGING_LEVEL_ERROR,
						models.Backend_Logging,
					)
				}
			}
			return err
		}
		pool.PoolStatus = true
		utils.SendLog(ctx.Data.Company, pool.NodePoolName+" nodepool added successfully", models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

	}

	oldCluster, err1 := GetPreviousEKSCluster(ctx)
	if err1 != nil {
		ctx.SendLogs(err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, err1.Error(), "error", cluster.InfraId)

		return types.CustomCPError{
			StatusCode:  int(models.CloudStatusCode),
			Error:       "Error in adding nodepool in running cluster",
			Description: err1.Error(),
		}
	}

	oldCluster.NodePools = cluster.NodePools
	for in, mainPool := range pools {
		mainPool.PoolStatus = true
		cluster.NodePools[in].PoolStatus = true
		oldCluster.NodePools = append(oldCluster.NodePools, mainPool)
		for _, pool := range pools {
			if pool.NodePoolName == mainPool.NodePoolName {
				cluster.NodePools[in].RoleName = pool.RoleName
				cluster.NodePools[in].NodeRole = pool.NodeRole
				if pool.RemoteAccess != nil {
					var remoteAccess RemoteAccessConfig
					remoteAccess.SourceSecurityGroups = pool.RemoteAccess.SourceSecurityGroups
					remoteAccess.Ec2SshKey = pool.RemoteAccess.Ec2SshKey
					remoteAccess.EnableRemoteAccess = pool.RemoteAccess.EnableRemoteAccess
					cluster.NodePools[in].RemoteAccess = &remoteAccess
				}
			}
		}
	}

	err1 = AddPreviousEKSCluster(oldCluster, ctx, true)
	if err1 != nil {
		ctx.SendLogs(err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, err1.Error(), "error", cluster.InfraId)

		return types.CustomCPError{Error: "Error in adding nodepool in running cluster", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)}
	}
	return types.CustomCPError{}
}
func PrintError(confError error, name string, ctx utils.Context) {
	if confError != nil {
		utils.SendLog(ctx.Data.Company, "Cluster creation failed : "+name, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
		utils.SendLog(ctx.Data.Company, confError.Error(), models.LOGGING_LEVEL_ERROR, ctx.Data.Company)
	}
}

func DeleteNodepool(cluster EKSCluster, ctx utils.Context, eksOps EKS, poolName string,token string) types.CustomCPError {
	utils.SendLog(ctx.Data.Company, "Deleting nodePool "+poolName, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

	err := eksOps.deleteNodePool(cluster.Name, poolName)
	if err != nil {
		err_ := types.CustomCPError{Error: "Error in deleting nodepool in running cluster", Description: err.Error(), StatusCode: int(models.CloudStatusCode)}

		updationFailedError(cluster, ctx, err_,token)
		return err_
	}
	utils.SendLog(ctx.Data.Company, " NodePool "+poolName+"deleted successfully", models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

	oldCluster, err1 := GetPreviousEKSCluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{
			StatusCode:  int(models.CloudStatusCode),
			Error:       "Error in deleting nodepool in running cluster",
			Description: err1.Error(),
		},token)
	}

	for _, pool := range oldCluster.NodePools {
		if pool.NodePoolName == poolName {
			if pool.NodeRole != nil && *pool.NodeRole != "" {
				err = eksOps.deleteIAMRoleFromInstanceProfile(*pool.RoleName)
				if err != nil {
					ctx.SendLogs(
						"EKS delete IAM role for cluster'"+cluster.Name+"', node group '"+pool.NodePoolName+"' failed: "+err.Error(),
						models.LOGGING_LEVEL_ERROR,
						models.Backend_Logging,
					)
					ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
					utils.SendLog(ctx.Data.Company, err.Error()+"\n Nodepool Deletion Failed - "+pool.NodePoolName, "error", cluster.InfraId)
					cpErr := ApiError(err, "NodePool Deletion Failed", 512)
					return cpErr
				}
				err := eksOps.deleteIAMRole(*pool.RoleName)
				if err != nil {
					return updationFailedError(cluster, ctx, types.CustomCPError{
						StatusCode:  int(models.CloudStatusCode),
						Error:       "Error in deleting nodepool in running cluster",
						Description: err.Error(),
					},token)
				}
			}
			if pool.RemoteAccess != nil && pool.RemoteAccess.EnableRemoteAccess && pool.RemoteAccess.Ec2SshKey != nil && *pool.RemoteAccess.Ec2SshKey != "" {
				err = eksOps.deleteSSHKey(pool.RemoteAccess.Ec2SshKey)
				if err != nil {
					ctx.SendLogs(
						"EKS delete SSH key for cluster '"+cluster.Name+"', node group '"+pool.NodePoolName+"' failed: "+err.Error(),
						models.LOGGING_LEVEL_ERROR,
						models.Backend_Logging,
					)
					ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
					utils.SendLog(ctx.Data.Company, err.Error()+"\n Nodepool Deletion Failed - "+pool.NodePoolName, "error", cluster.InfraId)
					cpErr := ApiError(err, "NodePool Deletion Failed", 512)
					return cpErr
				}
				pool.RemoteAccess.Ec2SshKey = nil
			}
			pool = nil
		}
	}
	err1 = AddPreviousEKSCluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx,
			types.CustomCPError{Error: "Error in deleting nodepool in running cluster", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}
	return types.CustomCPError{}
}

func updationFailedError(cluster EKSCluster, ctx utils.Context, err types.CustomCPError,token string) types.CustomCPError {
	publisher := utils.Notifier{}

	errr := publisher.Init_notifier()
	if errr != nil {
		PrintError(errr, cluster.Name, ctx)
		ctx.SendLogs(errr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := types.CustomCPError{StatusCode: 500, Error: "Error in deploying EKS Cluster", Description: errr.Error()}
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.EKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("EKSRunningClusterModel: Update - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return cpErr
	}

	cluster.Status = models.ClusterUpdateFailed
	confError := UpdateEKSCluster(cluster, ctx)
	if confError != nil {
		PrintError(confError, cluster.Name, ctx)
		ctx.SendLogs("EKSRunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	}

	utils.SendLog(ctx.Data.Company, "Error in running cluster update : "+err.Description, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
	utils.SendLog(ctx.Data.Company, err.Error, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
	err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.EKS, ctx, err)
	if err_ != nil {
		ctx.SendLogs("EKSRunningClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	}

	utils.SendLog(ctx.Data.Company, "Deployed cluster update failed : "+cluster.Name, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
	utils.SendLog(ctx.Data.Company, err.Description, models.LOGGING_LEVEL_ERROR, ctx.Data.Company)

	utils.Publisher(utils.ResponseSchema{
		Status:  false,
		Message: "Cluster update failed",
		InfraId: cluster.InfraId,
		Token:   token,
		Action:  models.Update,
	}, ctx)


	return err
}
