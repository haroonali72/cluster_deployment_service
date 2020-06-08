package eks

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/db"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type EKSCluster struct {
	ID                 bson.ObjectId     `json:"-" bson:"_id,omitempty"`
	ProjectId          string            `json:"project_id" bson:"project_id" validate:"required" description:"ID of project [required]"`
	Cloud              models.Cloud      `json:"cloud" bson:"cloud" validate:"eq=EKS|eq=eks"`
	CreationDate       time.Time         `json:"-" bson:"creation_date"`
	ModificationDate   time.Time         `json:"-" bson:"modification_date"`
	NodePools          []*NodePool       `json:"node_pools" bson:"node_pools" validate:"required,dive"`
	Status             models.Type       `json:"status" bson:"status" validate:"eq=new|eq=New|eq=NEW|eq=Cluster Creation Failed|eq=Cluster Terminated|eq=Cluster Created" description:"Status of cluster [required]"`
	CompanyId          string            `json:"company_id" bson:"company_id" description:"ID of compnay [optional]"`
	OutputArn          *string           `json:"-" bson:"output_arn,omitempty"`
	EncryptionConfig   *EncryptionConfig `json:"encryption_config,omitempty" bson:"encryption_config,omitempty" description:"Encryption Configurations [optional]"`
	Logging            Logging           `json:"logging" bson:"logging" description:"Logging Configurations [optional]"`
	Name               string            `json:"name" bson:"name" validate:"required" description:"Cluster name [required]"`
	ResourcesVpcConfig VpcConfigRequest  `json:"resources_vpc_config" bson:"resources_vpc_config" description:"Access Level Details [optional]`
	RoleArn            *string           `json:"-" bson:"role_arn"`
	RoleName           *string           `json:"-" bson:"role_name"`
	//Tags               map[string]*string `json:"tags,omitempty" bson:"tags,omitempty"`
	Version *string `json:"version,omitempty" bson:"version,omitempty" description:"Kubernetes Version [required]`
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
	OutputArn    *string `json:"-" bson:"output_arn,omitempty"`
	AmiType      *string `json:"ami_type,omitempty" bson:"ami_type,omitempty"`
	DiskSize     *int64  `json:"disk_size,omitempty" bson:"disk_size,omitempty"`
	InstanceType *string `json:"instance_type,omitempty" bson:"instance_type,omitempty"`
	//Labels        map[string]*string     `json:"labels,omitempty" bson:"labels,omitempty"`
	NodeRole      *string                `json:"-" bson:"node_role"`
	RoleName      *string                `json:"-" bson:"role_name"`
	NodePoolName  string                 `json:"node_pool_name" bson:"node_pool_name" description:"Node Pool Name [required]`
	RemoteAccess  *RemoteAccessConfig    `json:"remote_access,omitempty" bson:"remote_access,omitempty"`
	ScalingConfig *NodePoolScalingConfig `json:"scaling_config,omitempty" bson:"scaling_config,omitempty"`
	Subnets       []*string              `json:"-" bson:"subnets"`
	//Tags          map[string]*string     `json:"tags" bson:"tags"`
}

type RemoteAccessConfig struct {
	EnableRemoteAccess   bool      `json:"enable_remote_access" bson:"enable_remote_access"`
	Ec2SshKey            *string   `json:"-" bson:"ec2_ssh_key"`
	SourceSecurityGroups []*string `json:"-" bson:"source_security_groups"`
}

type NodePoolScalingConfig struct {
	DesiredSize *int64 `json:"desired_size" bson:"desired_size"`
	MaxSize     *int64 `json:"max_size" bson:"max_size"`
	MinSize     *int64 `json:"min_size" bson:"min_size"`
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

		if cluster.OutputArn != nil {

			eksOps.DeleteCluster(&cluster, ctx)

		}
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

	confError = ApplyAgent(credentials, token, ctx, cluster.Name, cluster.ResourceGroup)
	if confError != nil {
		utils.SendLog(companyId, confError.Error(), "error", cluster.ProjectId)

		cluster.Status = models.ClusterCreationFailed
		profile := vault.AwsProfile{Profile: credentials.Profile}
		_ = TerminateCluster(cluster, profile, cluster.ProjectId, companyId, ctx)
		utils.SendLog(companyId, "Cleaning up resources", "info", cluster.ProjectId)
		confError = UpdateEKSCluster(cluster, ctx)
		if confError != nil {
			utils.SendLog(companyId, confError.Error(), "error", cluster.ProjectId)
		}

		cpErr := ApiError(confError, "Error occurred while deploying agent", 500)
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)

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

func PrintError(confError error, name, projectId string, companyId string) {
	if confError != nil {
		beego.Error(confError.Error())
		_, _ = utils.SendLog(companyId, "Cluster creation failed : "+name, "error", projectId)
		_, _ = utils.SendLog(companyId, confError.Error(), "error", projectId)
	}
}
func ValidateEKSData(cluster EKSCluster, ctx utils.Context) error {
	if cluster.ProjectId == "" {

		return errors.New("project ID is empty")

	} else if cluster.Version == nil {

		return errors.New("kubernetes version is empty")

	}
	for _, pool := range cluster.NodePools {

		if pool.Name != nil && *pool.Name == "" {

			return errors.New("Node Pool name is empty")

		} else if pool.VMSize != nil && *pool.VMSize == "" {

			return errors.New("machine type with pool " + *pool.Name + " is empty")

		} else if pool.Count != nil && *pool.Count == 0 {

			return errors.New("node count value is zero within pool " + *pool.Name)

		} else if pool.OsDiskSizeGB != nil && (*pool.OsDiskSizeGB == 0 || *pool.OsDiskSizeGB < 40 || *pool.OsDiskSizeGB > 2048) {

			return errors.New("Disk size must be greater than 40 and less than 2048 within pool " + *pool.Name)

		} else if pool.MaxPods != nil && (*pool.MaxPods == 0 || *pool.MaxPods < 40) {

			return errors.New("max pods must be greater than or equal to 40 within pool " + *pool.Name)

		} else if pool.EnableAutoScaling != nil && *pool.EnableAutoScaling {

			if *pool.MinCount > *pool.MaxCount {
				return errors.New("min count should be less than or equal to max count within pool " + *pool.Name)
			}

		}

	}

	return nil
}
