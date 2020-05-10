package eks

import (
	"antelope/models"
	"antelope/models/db"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"antelope/models/vault"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

type EKSCluster struct {
	ID               bson.ObjectId `json:"-" bson:"_id,omitempty"`
	ProjectId        string        `json:"project_id" bson:"project_id"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	Status           string        `json:"status" bson:"status"`
	CompanyId        string        `json:"company_id" bson:"company_id"`

	OutputArn          *string            `json:"output_arn,omitempty" bson:"output_arn,omitempty"`
	EncryptionConfig   *EncryptionConfig  `json:"encryption_config,omitempty" bson:"encryption_config,omitempty"`
	Logging            Logging            `json:"logging" bson:"logging"`
	Name               string             `json:"name" bson:"name"`
	ResourcesVpcConfig VpcConfigRequest   `json:"resources_vpc_config" bson:"resources_vpc_config"`
	RoleArn            *string            `json:"-" bson:"role_arn"`
	RoleName           *string            `json:"-" bson:"role_name"`
	Tags               map[string]*string `json:"tags,omitempty" bson:"tags,omitempty"`
	Version            *string            `json:"version,omitempty" bson:"version,omitempty"`
	NodePools          []*NodePool        `json:"node_pools" bson:"node_pools"`
}

type EncryptionConfig struct {
	EnableEncryption bool      `json:"enable_encryption" bson:"enable_encryption"`
	Provider         *Provider `json:"-" bson:"provider"`
	Resources        []*string `json:"-" bson:"resources"`
}

type Provider struct {
	KeyArn *string `json:"-" bson:"key_arn"`
	KeyId  *string `json:"-" bson:"key_id"`
}

type Logging struct {
	EnableApi               bool `json:"enable_api" bson:"enable_api"`
	EnableAudit             bool `json:"enable_audit" bson:"enable_audit"`
	EnableAuthenticator     bool `json:"enable_authenticator" bson:"enable_authenticator"`
	EnableControllerManager bool `json:"enable_controller_manager" bson:"enable_controller_manager"`
	EnableScheduler         bool `json:"enable_scheduler" bson:"enable_scheduler"`
}

type LogSetup struct {
	Enabled *bool     `json:"enabled" bson:"enabled"`
	Types   []*string `json:"types" bson:"types"`
}

type VpcConfigRequest struct {
	EndpointPrivateAccess *bool     `json:"endpoint_private_access" bson:"endpoint_private_access"`
	EndpointPublicAccess  *bool     `json:"endpoint_public_access" bson:"endpoint_public_access"`
	PublicAccessCidrs     []*string `json:"public_access_cidrs" bson:"public_access_cidrs"`
	SecurityGroupIds      []*string `json:"-" bson:"security_group_ids"`
	SubnetIds             []*string `json:"-" bson:"subnet_ids"`
}

type NodePool struct {
	OutputArn     *string                `json:"output_arn,omitempty" bson:"output_arn,omitempty"`
	AmiType       *string                `json:"ami_type,omitempty" bson:"ami_type,omitempty"`
	DiskSize      *int64                 `json:"disk_size,omitempty" bson:"disk_size,omitempty"`
	InstanceType  *string                `json:"instance_type,omitempty" bson:"instance_type,omitempty"`
	Labels        map[string]*string     `json:"labels,omitempty" bson:"labels,omitempty"`
	NodeRole      *string                `json:"-" bson:"node_role"`
	RoleName      *string                `json:"-" bson:"role_name"`
	NodePoolName  string                 `json:"node_pool_name" bson:"node_pool_name"`
	RemoteAccess  *RemoteAccessConfig    `json:"remote_access,omitempty" bson:"remote_access,omitempty"`
	ScalingConfig *NodePoolScalingConfig `json:"scaling_config,omitempty" bson:"scaling_config,omitempty"`
	Subnets       []*string              `json:"-" bson:"subnets"`
	Tags          map[string]*string     `json:"tags" bson:"tags"`
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

	if oldCluster.Status == string(models.Deploying) {
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

func DeployEKSCluster(cluster EKSCluster, credentials vault.AwsProfile, companyId string, token string, ctx utils.Context) (confError error) {
	publisher := utils.Notifier{}
	confError = publisher.Init_notifier()

	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
		ctx.SendLogs(confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return confError
	}

	eksOps, err := GetEKS(cluster.ProjectId, credentials.Profile)
	if err != nil {
		ctx.SendLogs("EKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	err = eksOps.init()
	if err != nil {
		ctx.SendLogs("EKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.Status = "Cluster creation failed"
		confError = UpdateEKSCluster(cluster, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
			ctx.SendLogs("EKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}

	_, _ = utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	confError = eksOps.CreateCluster(&cluster, token, ctx)

	if confError != nil {
		ctx.SendLogs("EKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		PrintError(confError, cluster.Name, cluster.ProjectId, companyId)

		cluster.Status = "Cluster creation failed"
		confError = UpdateEKSCluster(cluster, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
			ctx.SendLogs("EKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return nil
	}
	cluster.Status = "Cluster Created"

	confError = UpdateEKSCluster(cluster, ctx)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
		ctx.SendLogs("EKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return confError
	}

	_, _ = utils.SendLog(companyId, "Cluster created successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	return nil
}

func TerminateCluster(credentials vault.AwsProfile, projectId, companyId string, ctx utils.Context) error {
	publisher := utils.Notifier{}
	pubErr := publisher.Init_notifier()
	if pubErr != nil {
		ctx.SendLogs("EKSClusterModel:  Terminate -"+pubErr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return pubErr
	}

	cluster, err := GetEKSCluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("EKSClusterModel : Terminate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	if cluster.Status == "" || cluster.Status == "new" {
		text := "EKSClusterModel : Terminate - Cannot terminate a new cluster"
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return errors.New(text)
	}

	eksOps, err := GetEKS(projectId, credentials.Profile)
	if err != nil {
		ctx.SendLogs("EKSClusterModel : Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	cluster.Status = string(models.Terminating)
	_, _ = utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	err = eksOps.init()
	if err != nil {
		ctx.SendLogs("EKSClusterModel : Terminate -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.Status = "Cluster Termination Failed"
		err = UpdateEKSCluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("EKSClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			_, _ = utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			_, _ = utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
			return err
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}

	err = eksOps.DeleteCluster(&cluster, ctx)
	if err != nil {
		_, _ = utils.SendLog(companyId, "Cluster termination failed: "+cluster.Name, "error", cluster.ProjectId)

		cluster.Status = "Cluster Termination Failed"
		err = UpdateEKSCluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("EKSClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			_, _ = utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			_, _ = utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
			publisher.Notify(cluster.ProjectId, "Status Available", ctx)
			return err
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return nil
	}

	cluster.Status = "Cluster Terminated"

	err = UpdateEKSCluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs("EKSClusterModel : Terminate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		_, _ = utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
		_, _ = utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}
	_, _ = utils.SendLog(companyId, "Cluster terminated successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	return nil
}

func PrintError(confError error, name, projectId string, companyId string) {
	if confError != nil {
		beego.Error(confError.Error())
		_, _ = utils.SendLog(companyId, "Cluster creation failed : "+name, "error", projectId)
		_, _ = utils.SendLog(companyId, confError.Error(), "error", projectId)
	}
}
