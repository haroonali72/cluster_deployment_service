package doks

import (
	"antelope/models"
	"antelope/models/cores"
	"antelope/models/db"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"antelope/models/woodpecker"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/digitalocean/godo"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type KubernetesClusterConfig struct {
	KubeconfigYAML []byte
}
type KubernetesConfig struct {
	ApiVersion     string     `yaml:"apiVersion"  json:"apiVersion"`
	Clusters       []Clusters `yaml:"clusters" json:"clusters"`
	Contexts       []Contexts `yaml:"contexts" json:"contexts"`
	CurrentContext string     `yaml:"current-context" json:"current-context"`
	Kind           string     `yaml:"kind" json:"kind"`
	Preferences    Preference `yaml:"preferences" json:"preferences"`
	Users          []Users    `yaml:"users" json:"users"`
}

type Clusters struct {
	Cluster Cluster `yaml:"cluster" json:"cluster"`
	Name    string  `yaml:"name" json:"name"`
}

type Cluster struct {
	Certificate string `yaml:"certificate-authority-data" json:"certificate-authority-data"`
	Server      string `yaml:"server" json:"server"`
}

type Contexts struct {
	Context Context `yaml:"context" json:"context"`
	Name    string  `yaml:"name" json:"name"`
}

type Context struct {
	Cluster string `yaml:"cluster" json:"cluster"`
	User    string `yaml:"user" json:"user"`
}

type Preference struct{}

type Users struct {
	Name string `yaml:"name" json:"name"`
	User User   `yaml:"user" json:"user"`
}

type User struct {
	Token string `yaml:"token" json:"token"`
}

type KubernetesCluster struct {
	ID               string                `json:"id" bson:"id"`
	ProjectId        string                `json:"project_id" bson:"project_id" validate:"required"`
	CompanyId        string                `json:"company_id" bson:"company_id" validate:"required"`
	Cloud            models.Cloud          `json:"cloud" bson:"cloud" validate:"eq=DOKS|eq=doks|eq=Doks"`
	CreationDate     time.Time             `json:"-" bson:"creation_date"`
	ModificationDate time.Time             `json:"-" bson:"modification_date"`
	CloudplexStatus  string                `json:"status" bson:"status"`
	Name             string                `json:"name,omitempty" bson:"name" validate:"required"`
	Region           string                `json:"region,omitempty" bson:"region" validate:"required"`
	KubeVersion      string                `json:"version,omitempty" bson:"version" validate:"required"`
	Tags             []string              `json:"tags,omitempty" bson:"tags"`
	NodePools        []*KubernetesNodePool `json:"node_pools,omitempty" bson:"node_pools" validate:"required,dive"`
	AutoUpgrade      bool                  `json:"auto_upgrade,omitempty" bson:"auto_upgrade"`
	IsAdvance        bool                  `json:"is_advance" bson:"is_advance"`
	IsExpert         bool                  `json:"is_expert" bson:"is_expert"`
	//NetworkName           string       `json:"network_name" bson:"network_name" valid:"required"`
	//ClusterSubnet 		string   	 `json:"cluster_subnet,omitempty" bson:"cluster_subnet"`
	//ServiceSubnet 		string   	 `json:"service_subnet,omitempty" bson:"service_subnet"`
	//IPv4          		string   	 `json:"ipv4,omitempty" bson:"ivp4"`
	//Endpoint      		string   	 `json:"endpoint,omitempty" bson:"endpoint"`
	//VPCUUID   			string       `json:"vpc_uuid" bson:"vpc_uuid"`
	//MaintenancePolicy     *KubernetesMaintenancePolicy 		`json:"maintenance_policy,omitempty" bson:"maintenance_policy"`
	//Status      *KubernetesClusterStatus `json:"kube_status,omitempty" bson:"kube_status"`
}

type KubernetesNodePool struct {
	ID          string            `json:"id,omitempty"  bson:"id"`
	Name        string            `json:"name,omitempty"  bson:"name" validate:"required"`
	MachineType string            `json:"machine_type,omitempty"  bson:"machine_type" validate:"required"` //machine size
	NodeCount   int               `json:"node_count,omitempty"  bson:"node_count" validate:"required,gte=1"`
	Tags        []string          `json:"tags,omitempty"  bson:"tags"`
	Labels      map[string]string `json:"labels,omitempty"  bson:"labels"`
	AutoScale   bool              `json:"auto_scale,omitempty"  bson:"auto_scale"`
	MinNodes    int               `json:"min_nodes,omitempty"  bson:"min_nodes"`
	MaxNodes    int               `json:"max_nodes,omitempty"  bson:"max_nodes"`
	Nodes       []*KubernetesNode `json:"nodes,omitempty"  bson:"nodes"`
}

type KubernetesNode struct {
	ID        string    `json:"id,omitempty" bson:"id"`
	Name      string    `json:"name,omitempty" bson:"name"`
	DropletID string    `json:"droplet_id,omitempty" bson:"droplet_id"`
	CreatedAt time.Time `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty" bson:"updated_at"`
	//	Status    *KubernetesNodeStatus `json:"status,omitempty" bson:"status"`
}

type KubernetesMaintenancePolicy struct {
	StartTime string `json:"start_time" bson:"start_time"`
	Duration  string `json:"duration" bson:"duration"`
	Day       string `json:"day" bson:"day"`
}

type KubernetesClusterStatus struct {
	State   string `json:"state,omitempty" bson:"state"`
	Message string `json:"message,omitempty" bson:"message"`
}

type KubernetesNodeStatus struct {
	State   string `json:"state,omitempty" bson:"state"`
	Message string `json:"message,omitempty" bson:"message"`
}

type KubernetesOptions struct {
	Versions []*KubernetesVersion  `json:"versions,omitempty"`
	Regions  []*KubernetesRegion   `json:"regions,omitempty"`
	Sizes    []*KubernetesNodeSize `json:"sizes,omitempty"`
}

type KubernetesVersion struct {
	Slug              string `json:"slug,omitempty"`
	KubernetesVersion string `json:"kubernetes_version,omitempty"`
}

type KubernetesRegion struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type KubernetesNodeSize struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func GetKubernetesCluster(ctx utils.Context) (cluster KubernetesCluster, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("DOKSGetClusterModel:  Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err1
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoDOKSClusterCollection)
	err = c.Find(bson.M{"project_id": ctx.Data.ProjectId, "company_id": ctx.Data.Company}).One(&cluster)
	if err != nil {
		ctx.SendLogs("DOKSGetClusterModel:  Get - Got error while fetching from database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err
	}

	return cluster, nil
}

func GetAllKubernetesCluster(data rbacAuthentication.List, ctx utils.Context) (clusters []KubernetesCluster, err error) {
	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("DOKSGetAllClusterModel:  GetAll - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return clusters, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoDOKSClusterCollection)
	err = c.Find(bson.M{"project_id": bson.M{"$in": copyData}}).All(&clusters)
	if err != nil {
		ctx.SendLogs("DOKSGetAllClusterModel:  GetAll - Got error while fetching from database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return clusters, err
	}

	return clusters, nil
}

func AddKubernetesCluster(cluster KubernetesCluster, ctx utils.Context) error {
	_, err := GetKubernetesCluster(ctx)
	if err == nil {
		text := fmt.Sprintf("DOKSAddClusterModel:  Add - Cluster for project '%s' already exists in the database.", cluster.ProjectId)
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs("DOKSAddClusterModel:  Add - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer session.Close()

	if cluster.CreationDate.IsZero() {
		cluster.CreationDate = time.Now()
		cluster.ModificationDate = time.Now()
		if cluster.CloudplexStatus == "" {
			cluster.CloudplexStatus = "new"
		}
		cluster.Cloud = models.DOKS
	}

	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoDOKSClusterCollection, cluster)
	if err != nil {
		ctx.SendLogs("DOKSAddClusterModel:  Add - Got error while inserting cluster to the database:  "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}

func UpdateKubernetesCluster(cluster KubernetesCluster, ctx utils.Context) error {
	oldCluster, err := GetKubernetesCluster(ctx)
	if err != nil {
		text := "DOKSUpdateClusterModel:  Update - Cluster '" + cluster.Name + "' does not exist in the database: " + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	err = DeleteKubernetesCluster(ctx)
	if err != nil {
		ctx.SendLogs("DOKSUpdateClusterModel:  Update - Got error deleting old cluster "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	cluster.CreationDate = oldCluster.CreationDate
	cluster.ModificationDate = time.Now()

	err = AddKubernetesCluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs("DOKSUpdateClusterModel:  Update - Got error creating new cluster "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}

func DeleteKubernetesCluster(ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs("DOKSDeleteClusterModel:  Delete - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoDOKSClusterCollection)
	err = c.Remove(bson.M{"project_id": ctx.Data.ProjectId, "company_id": ctx.Data.Company})
	if err != nil {
		ctx.SendLogs("DOKSDeleteClusterModel:  Delete - Got error while deleting from the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}

func PrintError(ctx utils.Context, confError, name string) {
	if confError != "" {
		_, _ = utils.SendLog(ctx.Data.Company, "Cluster creation failed : "+name, models.LOGGING_LEVEL_ERROR, ctx.Data.ProjectId)
		_, _ = utils.SendLog(ctx.Data.Company, confError, models.LOGGING_LEVEL_ERROR, ctx.Data.ProjectId)
	}
}

func DeployKubernetesCluster(cluster KubernetesCluster, credentials vault.DOCredentials, token string, ctx utils.Context) (customError types.CustomCPError) {

	publisher := utils.Notifier{}
	confError := publisher.Init_notifier()
	if confError != nil {
		PrintError(ctx, confError.Error(), cluster.Name)
		customError.StatusCode = 500
		customError.Description = confError.Error()
		err := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.DOKS, ctx, customError)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return customError
	}

	doksOps, err := GetDOKS(credentials)
	if err != nil {
		ctx.SendLogs("DOKSDeployClusterModel:  Deploy : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		customError.StatusCode = 500
		customError.Description = err.Error()
		err := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.DOKS, ctx, customError)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return customError
	}

	err1 := doksOps.init(ctx)
	if err1.Description != "" {
		cluster.CloudplexStatus = "Cluster creation failed"
		confError = UpdateKubernetesCluster(cluster, ctx)
		if confError != nil {
			PrintError(ctx, confError.Error(), cluster.Name)
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		err := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.DOKS, ctx, err1)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err1
	}

	_, _ = utils.SendLog(ctx.Data.Company, "Creating Cluster : "+cluster.Name, models.LOGGING_LEVEL_INFO, cluster.ProjectId)
	cluster.CloudplexStatus = string(models.Deploying)
	err_ := UpdateKubernetesCluster(cluster, ctx)
	if err_ != nil {

		utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.ProjectId)
		cpErr := types.CustomCPError{Description: err_.Error(), Message: "Error occurred while updating cluster status in database", StatusCode: 500}
		err := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.DOKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}
	cluster, errr := doksOps.createCluster(cluster, ctx, token, credentials)
	if errr.Description != "" {
		cluster.CloudplexStatus = "Cluster creation failed"
		confError = UpdateKubernetesCluster(cluster, ctx)
		if confError != nil {
			PrintError(ctx, confError.Error(), cluster.Name)
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		err := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.DOKS, ctx, errr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return errr
	}

	confErr := ApplyAgent(credentials, token, ctx, cluster.Name)
	if confErr.Description != "" {
		PrintError(ctx, confErr.Description, cluster.Name)
		cluster.CloudplexStatus = "Cluster creation failed"
		confError = UpdateKubernetesCluster(cluster, ctx)
		if confError != nil {
			PrintError(ctx, confError.Error(), cluster.Name)
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		err := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.DOKS, ctx, confErr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return confErr

	}

	cluster.CloudplexStatus = "Cluster Created"

	confError = UpdateKubernetesCluster(cluster, ctx)
	if confError != nil {
		PrintError(ctx, confError.Error(), cluster.Name)
		ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		CpErr := types.CustomCPError{StatusCode: 500,Message:"Error in applying agent", Description: err.Error()}
		err := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.DOKS, ctx, CpErr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return CpErr
	}

	_, _ = utils.SendLog(ctx.Data.Company, "Cluster created successfully "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.ProjectId)

	publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
	return types.CustomCPError{}
}

func FetchStatus(credentials vault.DOCredentials, ctx utils.Context) (*godo.KubernetesCluster, types.CustomCPError) {

	cluster, err := GetKubernetesCluster(ctx)
	if err != nil {
		ctx.SendLogs("DOKSClusterModel:  Fetch -  Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return &godo.KubernetesCluster{}, types.CustomCPError{StatusCode: 500, Message:"Error in applying agent",Description: err.Error()}
	}
	customErr, err := db.GetError(cluster.ProjectId, ctx.Data.Company, models.DOKS, ctx)
	if err != nil {
		return &godo.KubernetesCluster{}, types.CustomCPError{Message: "Error occurred while getting cluster status in database",
			Description: "Error occurred while getting cluster status in database",
			StatusCode:  500}
	}
	if customErr.Err != (types.CustomCPError{}) {
		return &godo.KubernetesCluster{}, customErr.Err
	}
	doksOps, err := GetDOKS(credentials)
	if err != nil {
		ctx.SendLogs("DOKSClusterModel:  Fetch -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return &godo.KubernetesCluster{}, types.CustomCPError{StatusCode: 500,Message:"Error in applying agent", Description: err.Error()}
	}

	err1 := doksOps.init(ctx)
	if err1.Description != "" {
		return &godo.KubernetesCluster{}, err1
	}

	status, errr := doksOps.fetchStatus(ctx, cluster.ID)
	if errr.Description != "" {
		return &godo.KubernetesCluster{}, errr
	}

	return status, errr
}

func TerminateCluster(credentials vault.DOCredentials, ctx utils.Context) (customError types.CustomCPError) {

	publisher := utils.Notifier{}
	confError := publisher.Init_notifier()
	if confError != nil {
		ctx.SendLogs("DOKSClusterModel:  Terminate Cluster : "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		customError.StatusCode = 500
		customError.Description = confError.Error()
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DOKS, ctx, customError)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return customError
	}

	doksOps, err := GetDOKS(credentials)
	if err != nil {
		ctx.SendLogs("DOKSClusterModel:  Terminate Cluster : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		customError.StatusCode = 500
		customError.Description = err.Error()
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DOKS, ctx, customError)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return customError
	}

	cluster, err := GetKubernetesCluster(ctx)
	if err != nil {
		ctx.SendLogs("DOKSClusterModel : Terminate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		customError = types.CustomCPError{StatusCode: 500,Message:"Error in applying agent", Description: err.Error()}
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DOKS, ctx, customError)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return customError
	}

	if cluster.CloudplexStatus == "" || cluster.CloudplexStatus == "new" {
		text := "DOKSClusterModel : Terminate - Cannot terminate a new cluster"
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		customError = types.CustomCPError{StatusCode: 500,Message:"Error in applying agent", Description: err.Error()}
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DOKS, ctx, customError)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return types.CustomCPError{StatusCode: 500, Description: text}
	}

	cluster.CloudplexStatus = string(models.Terminating)
	_, _ = utils.SendLog(ctx.Data.Company, "Terminating cluster: "+cluster.Name, models.LOGGING_LEVEL_INFO, cluster.ProjectId)

	err_ := UpdateKubernetesCluster(cluster, ctx)
	if err_ != nil {

		utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.ProjectId)
		cpErr := types.CustomCPError{Description: err_.Error(), Message: "Error occurred while updating cluster status in database", StatusCode: 500}
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DOKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}
	errr := doksOps.init(ctx)
	if errr.Description != "" {
		cluster.CloudplexStatus = "Cluster Termination Failed"
		err = UpdateKubernetesCluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("DOKSClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			_, _ = utils.SendLog(ctx.Data.Company, "Error in cluster updation in mongo: "+cluster.Name, models.LOGGING_LEVEL_ERROR, ctx.Data.ProjectId)
			_, _ = utils.SendLog(ctx.Data.Company, err.Error(), models.LOGGING_LEVEL_ERROR, ctx.Data.ProjectId)
			return errr
		}
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DOKS, ctx, errr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return errr
	}

	errr = doksOps.deleteCluster(cluster, ctx)
	if errr.Description != "" {
		_, _ = utils.SendLog(ctx.Data.Company, "Cluster termination failed: "+cluster.Name, models.LOGGING_LEVEL_ERROR, ctx.Data.ProjectId)
		cluster.CloudplexStatus = "Cluster Termination Failed"
		err = UpdateKubernetesCluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("DOKSClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			_, _ = utils.SendLog(ctx.Data.Company, "Error in cluster updation in mongo: "+cluster.Name, models.LOGGING_LEVEL_ERROR, ctx.Data.ProjectId)
			_, _ = utils.SendLog(ctx.Data.Company, err.Error(), models.LOGGING_LEVEL_ERROR, ctx.Data.ProjectId)
			err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DOKS, ctx, errr)
			if err != nil {
				ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			publisher.Notify(cluster.ProjectId, "Status Available", ctx)
			return errr
		}
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DOKS, ctx, errr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return types.CustomCPError{}
	}
	cluster.ID = ""
	cluster.CloudplexStatus = "Cluster Terminated"

	err = UpdateKubernetesCluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs("DOKSClusterModel : Terminate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		_, _ = utils.SendLog(ctx.Data.Company, "Error in cluster updation in mongo: "+cluster.Name, models.LOGGING_LEVEL_ERROR, cluster.ProjectId)
		_, _ = utils.SendLog(ctx.Data.Company, err.Error(), models.LOGGING_LEVEL_ERROR, cluster.ProjectId)
		cpErr := types.CustomCPError{StatusCode: 500,Message:"Error in applying agent", Description: err.Error()}
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DOKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}

	_, _ = utils.SendLog(ctx.Data.Company, "Cluster terminated successfully "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.ProjectId)
	publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
	return types.CustomCPError{}
}

func GetKubeConfig(credentials vault.DOCredentials, ctx utils.Context, cluster KubernetesCluster) (config KubernetesConfig, customError types.CustomCPError) {
	publisher := utils.Notifier{}
	confError := publisher.Init_notifier()
	if confError != nil {
		ctx.SendLogs("DOKSClusterModel:  Get kube config file : "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		customError.StatusCode = 500
		customError.Description = confError.Error()
		return config, customError
	}

	doksOps, err := GetDOKS(credentials)
	if err != nil {
		ctx.SendLogs("DOKSClusterModel:  Get kube config file : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		customError.StatusCode = 500
		customError.Description = err.Error()
		return config, customError
	}

	err1 := doksOps.init(ctx)
	if err1.Description != "" {
		return config, err1
	}

	config, errr := doksOps.GetKubeConfig(ctx, cluster)
	if errr.Description != "" {
		return config, errr
	}
	return config, customError
}

func GetDOKS(credentials vault.DOCredentials) (DOKS, error) {
	return DOKS{
		AccessKey: credentials.AccessKey,
		Region:    credentials.Region,
	}, nil
}

func ApplyAgent(credentials vault.DOCredentials, token string, ctx utils.Context, clusterName string) (confError types.CustomCPError) {

	companyId := ctx.Data.Company
	projetcID := ctx.Data.ProjectId
	data2, err := woodpecker.GetCertificate(projetcID, token, ctx)
	if err != nil {
		ctx.SendLogs("DOKubernetesClusterController : Apply Agent -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.CustomCPError{StatusCode: 500,Message:"Error in applying agent", Description: "Agent Deployment failed " + err.Error()}
	}

	filePath := "/tmp/" + companyId + "/" + projetcID + "/"
	cmd := "mkdir -p " + filePath + " && echo '" + data2 + "'>" + filePath + "agent.yaml"
	output, err := models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("DOKubernetesClusterController : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.CustomCPError{StatusCode: 500,Message:"Error in applying agent", Description: err.Error()}
	}

	cmd = "sudo docker run --rm --name " + companyId + projetcID + " -e DIGITALOCEAN_ACCESS_TOKEN=" + credentials.AccessKey + " -e cluster=" + clusterName + " -e yamlFile=" + filePath + "agent.yaml -v " + filePath + ":" + filePath + " " + models.DOAuthContainerName

	output, err = models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("DOKubernetesClusterController : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.CustomCPError{StatusCode: 500,Message:"Error in applying agent", Description: "Agent Deployment failed " + err.Error()}
	}
	return types.CustomCPError{}
}

func GetServerConfig(credentials vault.DOCredentials, ctx utils.Context) (options *godo.KubernetesOptions, customError types.CustomCPError) {

	publisher := utils.Notifier{}

	confError := publisher.Init_notifier()
	if confError != nil {
		ctx.SendLogs("DOKSClusterModel:  Get Options : "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		customError.StatusCode = 500
		customError.Description = confError.Error()
		return options, customError
	}

	doksOps, err := GetDOKS(credentials)
	if err != nil {
		ctx.SendLogs("DOKSClusterModel:  Get Options : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		customError.StatusCode = 500
		customError.Description = err.Error()
		return options, customError
	}

	errr := doksOps.init(ctx)
	if errr.Description != "" {
		return &godo.KubernetesOptions{}, errr
	}

	options, confErrr := doksOps.GetServerConfig(ctx)
	if confErrr.Description != "" {
		return &godo.KubernetesOptions{}, confErrr
	}

	return options, types.CustomCPError{}
}

func ValidateDOKSData(cluster KubernetesCluster, ctx utils.Context) error {
	if cluster.ProjectId == "" {

		return errors.New("project Id is empty")

	} else if cluster.Name == "" {

		return errors.New("cluster name is empty")

	} else if cluster.KubeVersion == "" {

		return errors.New("kubernetes version is empty")

	} else if len(cluster.NodePools) == 0 {

		return errors.New("node pool length must not be zero")

	} else {

		for _, nodepool := range cluster.NodePools {

			if nodepool.Name == "" {

				return errors.New("node pool name is empty")

			} else if nodepool.MachineType == "" {

				return errors.New("machine type is empty")

			} else if nodepool.NodeCount == 0 {

				return errors.New("node count must be greater than zero")

			} else if nodepool.AutoScale {

				if nodepool.MinNodes < 1 {

					return errors.New("min node count must be greater than zero")

				} else if nodepool.MaxNodes < 1 {

					return errors.New("max node count must be greater than zero")

				} else if nodepool.MaxNodes <= nodepool.MinNodes {

					return errors.New("max node count must be greater than min node count")

				} else if nodepool.MaxNodes > 25 {

					return errors.New("max node count msut be less than or equal to 25")

				}

			}

		}

	}

	if cluster.Region == "" {

		return errors.New("region is empty")

	} else {

		isRegionExist, err := validateDOKSRegion(cluster.Region)
		if err != nil && !isRegionExist {
			text := "availabe regions are " + err.Error()
			ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return errors.New(text)
		}

	}

	return nil
}

func validateDOKSRegion(region string) (bool, error) {

	bytes := cores.DORegions

	var regionList []KubernetesRegion

	err := json.Unmarshal(bytes, &regionList)
	if err != nil {
		return false, err
	}

	for _, v1 := range regionList {
		if v1.Slug == region {
			return true, nil
		}
	}

	var errData string
	for _, v1 := range regionList {
		errData += v1.Slug + ", "
	}

	return false, errors.New(errData)
}
