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
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

type EKSCluster struct {
	ID                 bson.ObjectId      `json:"-" bson:"_id,omitempty"`
	ProjectId          string             `json:"project_id" bson:"project_id" validate:"required" description:"ID of project [required]"`
	Cloud              models.Cloud       `json:"cloud" bson:"cloud" validate:"eq=EKS|eq=eks"`
	CreationDate       time.Time          `json:"-" bson:"creation_date"`
	ModificationDate   time.Time          `json:"-" bson:"modification_date"`
	NodePools          []*NodePool        `json:"node_pools" bson:"node_pools" validate:"required,dive"`
	IsAdvanced         bool               `json:"is_advance" bson:"is_advance" description:"Cluster advance level settings possible value 'true' or 'false'"`
	Status             models.Type        `json:"status" bson:"status" validate:"eq=new|eq=New|eq=NEW|eq=Cluster Creation Failed|eq=Cluster Terminated|eq=Cluster Created" description:"Status of cluster [required]"`
	CompanyId          string             `json:"company_id" bson:"company_id" description:"ID of compnay [optional]"`
	OutputArn          *string            `json:"-" bson:"output_arn,omitempty"`
	EncryptionConfig   *EncryptionConfig  `json:"encryption_config,omitempty" bson:"encryption_config,omitempty" description:"Encryption Configurations [optional]"`
	Logging            Logging            `json:"logging" bson:"logging" description:"Logging Configurations [optional]"`
	Name               string             `json:"name" bson:"name" validate:"required" description:"Cluster name [required]"`
	ResourcesVpcConfig VpcConfigRequest   `json:"resources_vpc_config" bson:"resources_vpc_config" description:"Access Level Details [optional]`
	RoleArn            *string            `json:"-" bson:"role_arn"`
	RoleName           *string            `json:"-" bson:"role_name"`
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
	OutputArn     *string                `json:"-" bson:"output_arn,omitempty"`
	AmiType       *string                `json:"ami_type,omitempty" bson:"ami_type" validate:"required" description:"AMI for nodepool [required]"`
	DiskSize      *int64                 `json:"disk_size,omitempty" bson:"disk_size,omitempty" description:"Size of disk for nodes in nodepool [optional]"`
	InstanceType  *string                `json:"instance_type,omitempty" bson:"instance_type" validate:"required" description:"Instance type for nodes [required]"`
	Labels        map[string]*string     `json:"-" bson:"labels,omitempty"`
	NodeRole      *string                `json:"-" bson:"node_role"`
	RoleName      *string                `json:"-" bson:"role_name"`
	NodePoolName  string                 `json:"node_pool_name" bson:"node_pool_name" validate:"required" description:"Node Pool Name [required]"`
	RemoteAccess  *RemoteAccessConfig    `json:"remote_access,omitempty" bson:"remote_access,omitempty" description:"Access Levels (private of public)[optional]"`
	ScalingConfig *NodePoolScalingConfig `json:"scaling_config,omitempty" bson:"scaling_config,omitempty" description:"Scaling Configurations for nodepool [optional]"`
	Subnets       []*string              `json:"-" bson:"subnets"`
	Tags          map[string]*string     `json:"-" bson:"tags"`
}

type RemoteAccessConfig struct {
	EnableRemoteAccess   bool      `json:"enable_remote_access" bson:"enable_remote_access" description:"Enable Remote Access [optional]"`
	Ec2SshKey            *string   `json:"-" bson:"ec2_ssh_key"`
	SourceSecurityGroups []*string `json:"-" bson:"source_security_groups"`
}

type NodePoolScalingConfig struct {
	DesiredSize *int64 `json:"desired_size" bson:"desired_size"`
	MaxSize     *int64 `json:"max_size" bson:"max_size"`
	MinSize     *int64 `json:"min_size" bson:"min_size"`
	IsEnabled   bool   `json:"is_enabled" bson:"is_enabled"`
}

type EKSClusterStatus struct {
	ClusterEndpoint *string         `json:"endpoint"`
	Name            *string         `json:"name"`
	Status          *string         `json:"status"`
	KubeVersion     *string         `json:"kubernetes_version"`
	ClusterArn      *string         `json:"cluster_arn"`
	NodePools       []EKSPoolStatus `json:"node_pools"`
}
type EKSPoolStatus struct {
	NodePoolArn *string          `json:"pool_arn"`
	Name        *string          `json:"name"`
	Status      *string          `json:"status"`
	AMI         *string          `json:"ami_type"`
	MachineType *string          `json:"machine_type"`
	DesiredSize *int64           `json:"desired_size"`
	MinSize     *int64           `json:"min_size"`
	MaxSize     *int64           `json:"max_size"`
	Nodes       []EKSNodesStatus `json:"nodes"`
}
type EKSNodesStatus struct {
	Name      *string `json:"name"`
	PublicIP  *string `json:"public_ip"`
	PrivateIP *string `json:"private_ip"`
	State     *string `json:"state"`
	ID        *string `json:"id"`
}

func KubeVersions(ctx utils.Context) []string {
	var kubeVersions []string
	kubeVersions = append(kubeVersions, "1.14")
	kubeVersions = append(kubeVersions, "1.15")
	kubeVersions = append(kubeVersions, "1.16")
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
func GetEKSCluster(projectId string, companyId string, ctx utils.Context) (cluster EKSCluster, err error) {
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
	err = c.Find(bson.M{"project_id": projectId, "company_id": companyId}).One(&cluster)
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
	err = c.Find(bson.M{"project_id": bson.M{"$in": copyData}}).All(&clusters)
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
func GetNetwork(token, projectId string, ctx utils.Context) error {

	url := getNetworkHost("aws", projectId)

	_, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}
func AddEKSCluster(cluster EKSCluster, ctx utils.Context) error {
	_, err := GetEKSCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err == nil {
		text := fmt.Sprintf("EKSAddClusterModel:  Add - Cluster for project '%s' already exists in the database.", cluster.ProjectId)
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
	oldCluster, err := GetEKSCluster(cluster.ProjectId, cluster.CompanyId, ctx)
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
	err = DeleteEKSCluster(cluster.ProjectId, cluster.CompanyId, ctx)
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

func DeleteEKSCluster(projectId, companyId string, ctx utils.Context) error {
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
	err = c.Remove(bson.M{"project_id": projectId, "company_id": companyId})
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

	eksOps := GetEKS(cluster.ProjectId, credentials.Profile)
	eksOps.init()

	utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)

	cluster.Status = (models.Deploying)
	confError := UpdateEKSCluster(cluster, ctx)
	if confError != nil {

		utils.SendLog(companyId, confError.Error(), "error", cluster.ProjectId)
		cpErr := ApiError(confError, "Error occurred while updating cluster status in database", 500)

		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}

	cpError := eksOps.CreateCluster(&cluster, token, ctx)

	if cpError != (types.CustomCPError{}) {
		utils.SendLog(ctx.Data.Company, "EKS CLuster Creation Failed", "error", cluster.ProjectId)

		//if cluster.OutputArn != nil {

		eksOps.CleanUpCluster(&cluster, ctx)

		//}
		cluster.Status = models.ClusterCreationFailed
		confError := UpdateEKSCluster(cluster, ctx)
		if confError != nil {

			utils.SendLog(companyId, confError.Error(), "error", cluster.ProjectId)
			cpErr := ApiError(confError, "Error occurred while updating cluster status in database", 500)
			err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.IKS, ctx, cpError)
			if err != nil {
				ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			publisher.Notify(cluster.ProjectId, "Status Available", ctx)
			return cpErr

		}
		utils.SendLog(companyId, "Cluster creation failed : "+cluster.Name, "error", cluster.ProjectId)
		ctx.SendLogs("Cluster creation failed", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.IKS, ctx, cpError)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpError
	}
	/**
	  TODO : Add Agent Deployment Process.Due on @Ahmad.
	*/
	pubSub := publisher.Subscribe(ctx.Data.ProjectId, ctx)
	confError = ApplyAgent(credentials, token, ctx, cluster.Name)
	if confError != nil {
		utils.SendLog(companyId, confError.Error(), "error", cluster.ProjectId)

		cluster.Status = models.ClusterCreationFailed
		profile := vault.AwsProfile{Profile: credentials.Profile}
		_ = TerminateCluster(cluster, profile, cluster.ProjectId, companyId, ctx)
		utils.SendLog(companyId, "Cleaning up resources", "info", cluster.ProjectId)
		confError_ := UpdateEKSCluster(cluster, ctx)
		if confError_ != nil {
			utils.SendLog(companyId, confError_.Error(), "error", cluster.ProjectId)
		}

		cpErr := ApiError(confError, "Error occurred while deploying agent", 500)
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("EKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}
	cluster.Status = models.ClusterCreated

	confError = UpdateEKSCluster(cluster, ctx)

	if confError != nil {

		utils.SendLog(companyId, confError.Error(), "error", cluster.ProjectId)
		cpErr := ApiError(confError, "Error occurred while updating cluster status in database", 500)
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.IKS, ctx, cpError)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}
	utils.SendLog(companyId, "Cluster Created Sccessfully "+cluster.Name, "info", cluster.ProjectId)
	notify := publisher.RecieveNotification(ctx.Data.ProjectId, ctx, pubSub)
	if notify {
		ctx.SendLogs("EKSClusterModel:  Notification recieved from agent", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
	} else {
		ctx.SendLogs("EKSClusterModel:  Notification not recieved from agent", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	}

	return types.CustomCPError{}
}

func TerminateCluster(cluster EKSCluster, credentials vault.AwsProfile, projectId, companyId string, ctx utils.Context) types.CustomCPError {
	publisher := utils.Notifier{}
	publisher.Init_notifier()

	eksOps := GetEKS(projectId, credentials.Profile)

	cluster.Status = (models.Terminating)
	utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	err_ := UpdateEKSCluster(cluster, ctx)
	if err_ != nil {

		utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.ProjectId)
		cpErr := types.CustomCPError{Description: err_.Error(), Error: "Error occurred while updating cluster status in database", StatusCode: 500}
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}
	eksOps.init()

	cpErr := eksOps.DeleteCluster(&cluster, ctx)
	if cpErr != (types.CustomCPError{}) {

		utils.SendLog(companyId, "Cluster termination failed: "+cpErr.Description+cluster.Name, "error", cluster.ProjectId)

		cluster.Status = models.ClusterTerminationFailed
		err := UpdateEKSCluster(cluster, ctx)
		if err != nil {
			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)

		}
		err = db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}

	cluster.Status = models.ClusterTerminated
	err := UpdateEKSCluster(cluster, ctx)
	if err != nil {
		utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
		utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)

		cpErr := ApiError(err, "Error occurred while updating cluster status in database", 500)
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr

	}
	utils.SendLog(companyId, "Cluster terminated successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	return types.CustomCPError{}
}

func ValidateEKSData(cluster EKSCluster, ctx utils.Context) error {
	if cluster.ProjectId == "" {

		return errors.New("project ID is empty")

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
func FetchStatus(credentials vault.AwsProfile, projectId string, ctx utils.Context, companyId string, token string) (EKSClusterStatus, types.CustomCPError) {

	cluster, err := GetEKSCluster(projectId, companyId, ctx)
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
		customErr, err := db.GetError(projectId, companyId, models.IKS, ctx)
		if err != nil {
			cpErr := ApiError(err, "Error occurred while getting cluster status in database", 500)
			return EKSClusterStatus{}, cpErr
		}
		if customErr.Err != (types.CustomCPError{}) {
			return EKSClusterStatus{}, customErr.Err
		}
	}
	eks := GetEKS(cluster.ProjectId, credentials.Profile)

	eks.init()

	response, e := eks.fetchStatus(&cluster, ctx, companyId)

	if e != (types.CustomCPError{}) {

		ctx.SendLogs("Cluster model: Status - Failed to get lastest status "+e.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return EKSClusterStatus{}, e
	}

	return response, types.CustomCPError{}
}
func ApplyAgent(credentials vault.AwsProfile, token string, ctx utils.Context, clusterName string) (confError error) {
	companyId := ctx.Data.Company
	projetcID := ctx.Data.ProjectId
	data2, err := woodpecker.GetCertificate(projetcID, token, ctx)
	if err != nil {
		ctx.SendLogs("EKSClusterModel : Apply Agent -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	filePath := "/tmp/" + companyId + "/" + projetcID + "/"
	cmd := "mkdir -p " + filePath + " && echo '" + data2 + "'>" + filePath + "agent.yaml"
	output, err := models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("EKSClusterModel : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	cmd = "sudo docker run --rm --name " + companyId + projetcID + " -e accessKey=" + credentials.Profile.AccessKey + " -e cluster=" + clusterName + " -e secretKey=" + credentials.Profile.SecretKey + " -e region=" + credentials.Profile.Region + " -e yamlFile=" + filePath + "agent.yaml -v " + filePath + ":" + filePath + " " + models.EKSAuthContainerName

	output, err = models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("AKSClusterModel : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func GetEKSClusters(projectId string, credentials vault.AwsProfile, ctx utils.Context) ([]*string, types.CustomCPError) {
	eksOps := GetEKS(projectId, credentials.Profile)
	eksOps.init()
	clusters, cpError := eksOps.getEKSCluster(ctx)
	if cpError != (types.CustomCPError{}) {
		return nil, cpError
	}
	return clusters, types.CustomCPError{}
}
