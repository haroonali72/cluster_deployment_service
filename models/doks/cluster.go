package doks

import (
	"antelope/models"
	"antelope/models/api_handler"
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
	"github.com/r3labs/diff"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
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
	InfraId          string                `json:"infra_id" bson:"infra_id" validate:"required" description:"ID of infrastructure [required]"`
	CompanyId        string                `json:"company_id" bson:"company_id" validate:"required" description:"ID of compnay [optional]"`
	Cloud            models.Cloud          `json:"cloud" bson:"cloud" validate:"eq=DOKS|eq=doks|eq=Doks"`
	CreationDate     time.Time             `json:"-" bson:"creation_date"`
	ModificationDate time.Time             `json:"-" bson:"modification_date"`
	CloudplexStatus  models.Type           `json:"status" bson:"status" validate:"eq=new|eq=New|eq=NEW|eq=Cluster Creation Failed|eq=Cluster Terminated|eq=Cluster Created|eq=Cluster Update Failed" description:"Status of cluster [required]"`
	Name             string                `json:"name,omitempty" bson:"name" validate:"required" description:"Cluster name [required]"`
	Region           string                `json:"region,omitempty" bson:"region"  description:"Location for cluster provisioning [required]"`
	KubeVersion      string                `json:"version,omitempty" bson:"version" validate:"required" description:"Kubernetes version to be provisioned [required]"`
	Tags             []string              `json:"tags,omitempty" bson:"tags"`
	NodePools        []*KubernetesNodePool `json:"node_pools,omitempty" bson:"node_pools" validate:"required,dive"`
	AutoUpgrade      bool                  `json:"auto_upgrade,omitempty" bson:"auto_upgrade" description:"Auto upgradation of cluster on new kubernetes version [optional]"`
	IsAdvance        bool                  `json:"is_advance" bson:"is_advance"`
	IsExpert         bool                  `json:"is_expert" bson:"is_expert"`
	NetworkName      string                `json:"network_name" bson:"network_name" valid:"required"`
	VPCUUID          string                `json:"vpc_id" bson:"vpc_uuid"`
}

type KubernetesNodePool struct {
	ID          string            `json:"id"  bson:"id"`
	Name        string            `json:"name,omitempty"  bson:"name" validate:"required" description:"Cluster pool name [required]"`
	MachineType string            `json:"machine_type,omitempty"  bson:"machine_type" validate:"required" description:"Machine type for pool [required]"` //machine size
	NodeCount   int               `json:"node_count,omitempty"  bson:"node_count" validate:"required,gte=1" description:"Pool node count [required]"`
	Tags        []string          `json:"tags,omitempty"  bson:"tags" description:"Node pool tags [optional]"`
	Labels      map[string]string `json:"labels,omitempty"  bson:"labels" description:"Node pool labels, it would be key value pair [optional]"`
	AutoScale   bool              `json:"auto_scale,omitempty"  bson:"auto_scale" description:"Autoscaling configuration, possible value 'true' or 'false' [required]"`
	MinNodes    int               `json:"min_nodes,omitempty"  bson:"min_nodes" description:"Min VM count ['required' if autoscaling is enabled]"`
	MaxNodes    int               `json:"max_nodes,omitempty"  bson:"max_nodes" description:"Max VM count, must be greater than min count ['required' if autoscaling is enabled]"`
	Nodes       []*KubernetesNode `json:"nodes,omitempty"  bson:"nodes"`
	PoolStatus  bool              `json:"pool_status" bson:"pool_status"`
}

type KubernetesNode struct {
	ID        string    `json:"-,omitempty" bson:"id"`
	Name      string    `json:"name,omitempty" bson:"name" description:"Name of the node [optional]"`
	DropletID string    `json:"-" bson:"droplet_id"`
	CreatedAt time.Time `json:"-" bson:"created_at"`
	UpdatedAt time.Time `json:"-" bson:"updated_at"`
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
type KubeClusterStatus struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Status            models.Type            `json:"status"`
	State             string                 `json:"state"`
	RegionSlug        string                 `json:"region"`
	KubernetesVersion string                 `json:"kubernetes_version"`
	ClusterIp         string                 `json:"cluster_ip"`
	NodePoolCount     int                    `json:"nodepool_count"`
	Endpoint          string                 `json:"endpoint"`
	WorkerPools       []KubeWorkerPoolStatus `json:"node_pools"`
}

type KubeWorkerPoolStatus struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Size        string      `json:"machine_type"`
	Nodes       []PoolNodes `json:"nodes"`
	Count       int         `json:"node_count"`
	AutoScaling AutoScaling `json:"auto_scaling,omitempty"`
}
type AutoScaling struct {
	AutoScale bool `json:"auto_scale,omitempty"`
	MinCount  int  `json:"min_scaling_group_size,omitempty"`
	MaxCount  int  `json:"max_scaling_group_size,omitempty"`
}

type PoolNodes struct {
	Name      string `json:"name,omitempty"`
	State     string `json:"state,omitempty"`
	DropletID string `json:"id,omitempty"`
	PublicIp  string `json:"public_ip,omitempty"`
	PrivateIp string `json:"private_ip,omitempty"`
}

func getNetworkHost(cloudType, infraId string) string {

	host := beego.AppConfig.String("network_url") + models.WeaselGetEndpoint

	if strings.Contains(host, "{cloud}") {
		host = strings.Replace(host, "{cloud}", cloudType, -1)
	}

	if strings.Contains(host, "{infraId}") {
		host = strings.Replace(host, "{infraId}", infraId, -1)
	}

	return host
}
func GetNetwork(token, infraId string, ctx utils.Context) error {

	url := getNetworkHost("do", infraId)

	_, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}

type DOKSCluster struct {
	Name    string      `json:"name,omitempty" bson:"name,omitempty" v description:"Cluster name"`
	InfraId string      `json:"infra_id" bson:"infra_id"  description:"ID of infrastructure"`
	Status  models.Type `json:"status,omitempty" bson:"status,omitempty" " description:"Status of cluster"`
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
	err = c.Find(bson.M{"infra_id": ctx.Data.InfraId, "company_id": ctx.Data.Company}).One(&cluster)
	if err != nil {
		ctx.SendLogs("DOKSGetClusterModel:  Get - Got error while fetching from database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err
	}

	return cluster, nil
}
func GetAllKubernetesCluster(data rbacAuthentication.List, ctx utils.Context) (dokscluster []DOKSCluster, err error) {
	var clusters []KubernetesCluster
	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("DOKSGetAllClusterModel:  GetAll - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []DOKSCluster{}, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoDOKSClusterCollection)
	err = c.Find(bson.M{"infra_id": bson.M{"$in": copyData}, "company_id": ctx.Data.Company}).All(&clusters)
	if err != nil {
		ctx.SendLogs("DOKSGetAllClusterModel:  GetAll - Got error while fetching from database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return dokscluster, err
	}
	for _, cluster := range clusters {
		temp := DOKSCluster{Name: cluster.Name, InfraId: cluster.InfraId, Status: cluster.CloudplexStatus}
		dokscluster = append(dokscluster, temp)
	}
	return dokscluster, nil
}
func AddKubernetesCluster(cluster KubernetesCluster, ctx utils.Context) error {
	_, err := GetKubernetesCluster(ctx)
	if err == nil {
		text := fmt.Sprintf("DOKSAddClusterModel:  Add - Cluster for infrastructure '%s' already exists in the database.", cluster.InfraId)
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
func DeleteKubernetesCluster(ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs("DOKSDeleteClusterModel:  Delete - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoDOKSClusterCollection)
	err = c.Remove(bson.M{"infra_id": ctx.Data.InfraId, "company_id": ctx.Data.Company})
	if err != nil {
		ctx.SendLogs("DOKSDeleteClusterModel:  Delete - Got error while deleting from the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
	cluster.CompanyId = oldCluster.CompanyId
	if oldCluster.ID != "" && oldCluster.CloudplexStatus != models.Terminating {
		cluster.ID = oldCluster.ID
	}
	for index, pool := range cluster.NodePools {
		if len(cluster.NodePools) >= len(oldCluster.NodePools) && index < len(oldCluster.NodePools) && oldCluster.CloudplexStatus != models.Terminating {
			if oldCluster.NodePools[index] != nil && oldCluster.NodePools[index].ID != "" {
				pool.ID = oldCluster.NodePools[index].ID
			}
		}
	}
	err = AddKubernetesCluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs("DOKSUpdateClusterModel:  Update - Got error creating new cluster "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}

func UpdatePreviousDOKSCluster(cluster KubernetesCluster, ctx utils.Context) error {

	err := AddPreviousDOKSCluster(cluster, ctx, false)
	if err != nil {
		text := "DOKSClusterModel:  Update  previous cluster -'" + cluster.Name + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	err = UpdateKubernetesCluster(cluster, ctx)
	if err != nil {
		text := "DOKSClusterModel:  Update previous cluster - '" + cluster.Name + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		err = DeletePreviousDOKSCluster(ctx)
		if err != nil {
			text := "DOKSDeleteClusterModel:  Delete  previous cluster - '" + cluster.Name + err.Error()
			ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return errors.New(text)
		}
		return err
	}

	return nil
}
func AddPreviousDOKSCluster(cluster KubernetesCluster, ctx utils.Context, patch bool) error {

	var oldCluster KubernetesCluster

	_, err := GetPreviousDOKSCluster(ctx)
	if err == nil {
		err := DeletePreviousDOKSCluster(ctx)
		if err != nil {
			ctx.SendLogs(
				"DOKSAddClusterModel:  Add previous cluster - "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
		}
	}

	if patch == false {
		oldCluster, err = GetKubernetesCluster(ctx)
		if err != nil {
			ctx.SendLogs(
				"DOKSAddClusterModel:  Add previous cluster - "+err.Error(),
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
			"DOKSAddClusterModel:  Add previous cluster - "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	defer session.Close()

	if cluster.CreationDate.IsZero() {
		cluster.CreationDate = time.Now()
		cluster.ModificationDate = time.Now()
		cluster.Cloud = models.DOKS
		cluster.CompanyId = ctx.Data.Company
	}

	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoDOKSPreviousClusterCollection, oldCluster)
	if err != nil {
		ctx.SendLogs(
			"DOKSAddClusterModel:  Add previous cluster -  "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}
func DeletePreviousDOKSCluster(ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"DOKSDeleteClusterModel:  Delete  previous cluster - "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoDOKSPreviousClusterCollection)
	err = c.Remove(bson.M{"infra_id": ctx.Data.InfraId, "company_id": ctx.Data.Company})
	if err != nil {
		ctx.SendLogs(
			"DOKSDeleteClusterModel:  Delete  previous cluster - "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}
func GetPreviousDOKSCluster(ctx utils.Context) (cluster KubernetesCluster, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs(
			"DOKSGetClusterModel:  Get previous cluster - Got error while connecting to the database: "+err1.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoDOKSPreviousClusterCollection)
	err = c.Find(bson.M{"infra_id": ctx.Data.InfraId, "company_id": ctx.Data.Company}).One(&cluster)
	if err != nil {
		ctx.SendLogs(
			"DOKSGetClusterModel:  Get previous cluster- Got error while fetching from database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err
	}

	return cluster, nil
}

func PrintError(ctx utils.Context, confError, name string) {
	if confError != "" {
		_, _ = utils.SendLog(ctx.Data.Company, "Cluster creation failed : "+name, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
		_, _ = utils.SendLog(ctx.Data.Company, confError, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
	}
}

func DeployKubernetesCluster(cluster KubernetesCluster, credentials vault.DOCredentials, token string, ctx utils.Context) (customError types.CustomCPError) {

	publisher := utils.Notifier{}
	confError := publisher.Init_notifier()
	if confError != nil {
		PrintError(ctx, confError.Error(), cluster.Name)
		customError.StatusCode = 500
		customError.Description = confError.Error()
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.DOKS, ctx, customError)
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
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.DOKS, ctx, customError)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return customError
	}

	err1 := doksOps.init(ctx)
	if err1 != (types.CustomCPError{}) {
		cluster.CloudplexStatus = models.ClusterCreationFailed
		confError = UpdateKubernetesCluster(cluster, ctx)
		if confError != nil {
			PrintError(ctx, confError.Error(), cluster.Name)
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.DOKS, ctx, err1)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err1.Error + "\n" + err1.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return err1
	}

	_, _ = utils.SendLog(ctx.Data.Company, "Creating Cluster : "+cluster.Name, models.LOGGING_LEVEL_INFO, cluster.InfraId)
	cluster.CloudplexStatus = (models.Deploying)
	err_ := UpdateKubernetesCluster(cluster, ctx)
	if err_ != nil {
		utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.InfraId)
		cpErr := types.CustomCPError{Description: err_.Error(), Error: "Error occurred while updating cluster status in database", StatusCode: 500}
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.DOKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err1.Error + "\n" + err1.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return cpErr
	}

	clus, errr := doksOps.createCluster(&cluster, ctx, token, credentials)
	if errr != (types.CustomCPError{}) {
		cluster.CloudplexStatus = models.ClusterCreationFailed

		confError = UpdateKubernetesCluster(cluster, ctx)
		if confError != nil {
			PrintError(ctx, confError.Error(), cluster.Name)
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.DOKS, ctx, errr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: errr.Error + "\n" + errr.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return errr
	}

	confError = UpdateKubernetesCluster(*clus, ctx)
	if confError != nil {
		PrintError(ctx, confError.Error(), cluster.Name)
		ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		CpErr := types.CustomCPError{StatusCode: 500, Error: "Error in updating cluster", Description: err.Error()}
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.DOKS, ctx, CpErr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: confError.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return CpErr
	}

	pubSub := publisher.Subscribe(ctx.Data.InfraId, ctx)
	confErr := ApplyAgent(credentials, token, ctx, cluster.Name)
	if confErr != (types.CustomCPError{}) {
		PrintError(ctx, confErr.Description, cluster.Name)
		utils.SendLog(ctx.Data.Company, "Cleaning up resources", "info", cluster.InfraId)
		cluster.CloudplexStatus = models.ClusterCreationFailed
		_ = TerminateCluster(credentials, token, ctx)
		confError = UpdateKubernetesCluster(cluster, ctx)
		if confError != nil {
			PrintError(ctx, confError.Error(), cluster.Name)
			ctx.SendLogs("DOKSDeployClusterModel:  Apply agent - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.DOKS, ctx, confErr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Apply agent  - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: confErr.Error + "\n" + confErr.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return confErr

	}

	cluster.CloudplexStatus = models.ClusterCreated
	confError = UpdateKubernetesCluster(cluster, ctx)
	if confError != nil {
		PrintError(ctx, confError.Error(), cluster.Name)
		ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		CpErr := types.CustomCPError{StatusCode: 500, Error: "Error in updating cluster", Description: err.Error()}
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.DOKS, ctx, CpErr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: CpErr.Error + "\n" + CpErr.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return CpErr
	}

	_, _ = utils.SendLog(ctx.Data.Company, "Cluster created successfully "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

	notify := publisher.RecieveNotification(ctx.Data.InfraId, ctx, pubSub)
	if notify {
		ctx.SendLogs("DOKSClusterModel:  Notification recieved from agent", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Cluster created successfully",
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
	} else {
		ctx.SendLogs("DOKSClusterModel:  Notification not recieved from agent", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		cluster.CloudplexStatus = models.ClusterCreationFailed
		PrintError(ctx, "Notification not recieved from agent", cluster.Name)
		confError = UpdateKubernetesCluster(cluster, ctx)
		if confError != nil {
			PrintError(ctx, confError.Error(), cluster.Name)
			ctx.SendLogs("DOKSDeployClusterModel:  Apply agent - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.DOKS, ctx, confErr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Apply agent  - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Notification not recieved from agent",
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)

	}

	return types.CustomCPError{}
}
func PatchRunningDOKSCluster(cluster KubernetesCluster, credentials vault.DOCredentials, token string, ctx utils.Context) (confError types.CustomCPError) {

	//publisher := utils.Notifier{}

	/*err := publisher.Init_notifier()
	if err != nil {
		PrintError(ctx, err.Error(), cluster.Name)
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := types.CustomCPError{StatusCode: int(models.CloudStatusCode), Error: "Error in updating DOKS cluster", Description: err.Error()}
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.DOKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("DOKSUpdateRunningClusterModel:  Update - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return cpErr
	}*/

	doksOps, err := GetDOKS(credentials)
	if err != nil {
		ctx.SendLogs("DOKSUpdateRunningClusterModel: Update running cluster : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		confError.StatusCode = int(models.CloudStatusCode)
		confError.Description = err.Error()
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.DOKS, ctx, confError)
		if err != nil {
			ctx.SendLogs("DOKSUpdateRunningClusterModel:  Update running cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return confError
	}

	err1 := doksOps.init(ctx)
	if err1 != (types.CustomCPError{}) {
		cluster.CloudplexStatus = models.ClusterCreationFailed
		confError := UpdateKubernetesCluster(cluster, ctx)
		if confError != nil {
			PrintError(ctx, confError.Error(), cluster.Name)
			ctx.SendLogs("DOKSUpdateRunningClusterModel:  Update running cluster - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.DOKS, ctx, err1)
		if err != nil {
			ctx.SendLogs("DOKSUpdateRunningClusterModel:  Update running cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err1.Error + "\n" + err1.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Update,
		}, ctx)
		return err1
	}

	difCluster, err2 := CompareClusters(ctx)
	if err2 != nil {
		if strings.Contains(err2.Error(), "Nothing to update") {
			cluster.CloudplexStatus = models.ClusterCreated
			confError_ := UpdateKubernetesCluster(cluster, ctx)
			if confError_ != nil {
				ctx.SendLogs("DOKSUpdateRunningClusterModel: "+confError_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			utils.Publisher(utils.ResponseSchema{
				Status:  true,
				Message: "Nothing to update",
				InfraId: cluster.InfraId,
				Token:   token,
				Action:  models.Update,
			}, ctx)
			return types.CustomCPError{}
		}
	}

	_, _ = utils.SendLog(ctx.Data.Company, "Updating running cluster : "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

	/*if previousPoolCount < newPoolCount {
		var pools []*KubernetesNodePool
		for i := previousPoolCount; i < newPoolCount; i++ {
			pools = append(pools, cluster.NodePools[i])
		}

		err := AddNodepool(cluster, ctx, doksOps, pools, credentials)
		if err != (types.CustomCPError{}) {
			return err
		}


	} else if previousPoolCount > newPoolCount {

		previousCluster, err := GetPreviousDOKSCluster(ctx)
		if err != nil {
			return types.CustomCPError{Error: "Error in updating running cluster", StatusCode: 512, Description: err.Error()}
		}

		for _, oldpool := range previousCluster.NodePools {
			delete := true
			for _, pool := range cluster.NodePools {

				if pool.Name == oldpool.Name {
					delete = false
				}
			}
			if delete == true {
				DeleteNodepool(cluster, ctx, doksOps, oldpool.ID,credentials)
				utils.SendLog(ctx.Data.Company, "Nodepool  "+ oldpool.Name +" deleted from "+ cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

			}
		}
	}
	*/
	previousCluster, err := GetPreviousDOKSCluster(ctx)
	if err != nil {
		return types.CustomCPError{Error: "Error in updating running cluster", StatusCode: 512, Description: err.Error()}
	}

	previousPoolCount := len(previousCluster.NodePools)

	var addpools []*KubernetesNodePool

	var addedIndex []int

	addincluster := false

	for index, pool := range cluster.NodePools {

		existInPrevious := false

		for _, prePool := range previousCluster.NodePools {
			if pool.Name == prePool.Name {
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

		err := AddNodepool(cluster, ctx, doksOps, addpools, credentials,token)
		if err != (types.CustomCPError{}) {
			return err
		}
	}

	for _, prePool := range previousCluster.NodePools {
		existInNew := false
		for _, pool := range cluster.NodePools {
			if pool.ID == prePool.ID {
				existInNew = true
			}
		}

		if existInNew == false {
			DeleteNodepool(cluster, ctx, doksOps, prePool.ID, credentials,token)
		}
	}

	done, done1, index := false, false, -1
	for _, dif := range difCluster {

		if len(dif.Path) > 2 {
			poolIndex, _ := strconv.Atoi(dif.Path[1])
			if poolIndex > (previousPoolCount - 1) {
				continue
			}

			if poolIndex > index {
				index = poolIndex
				done = false
			}

			for _, in := range addedIndex {
				if in == poolIndex {
					continue
				}
			}
		}

		if dif.Type == "update" || dif.Type == "create" {
			if dif.Path[0] == "AutoUpgrade" || dif.Path[0] == "Tags" {
				if !done1 {
					err := UpdateCluster(cluster, ctx, doksOps, credentials,token)
					if err != (types.CustomCPError{}) {
						return err
					}
					utils.SendLog(ctx.Data.Company, "Cluster Tags/AutoUpgrade updated ", models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
				}
				done1 = true
			} else if dif.Path[0] == "KubeVersion" {
				err := UpdateKubernetesVersion(cluster, ctx, doksOps, credentials,token)
				if err != (types.CustomCPError{}) {
					return err
				}
				utils.SendLog(ctx.Data.Company, "Cluster kubernetes version updated ", models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

			} else if dif.Path[0] == "NodePools" {
				if !done {
					poolIndex, _ := strconv.Atoi(dif.Path[1])
					if dif.Path[2] == "PoolStatus" || dif.Path[2] == string(models.ClusterUpdateFailed) || dif.Path[2] == "ID" {
						continue
					}
					err := UpdateNodePool(cluster, poolIndex, ctx, doksOps, credentials,token )
					if err != (types.CustomCPError{}) {
						return err
					}
					utils.SendLog(ctx.Data.Company, "Cluster nodepool updated ", models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

				}
				done = true
			}
		}
	}

	DeletePreviousDOKSCluster(ctx)

	/*latestCluster, err1 := doksOps.fetchStatus( ctx,cluster.ID)
	if err1 != (types.CustomCPError{}) {
		return err1
	}

	for strings.ToLower(string(latestCluster.State)) != strings.ToLower("running") {
		time.Sleep(time.Second * 60)
	}
	*/
	cluster.CloudplexStatus = models.ClusterCreated
	confError_ := UpdateKubernetesCluster(cluster, ctx)
	if confError_ != nil {
		ctx.SendLogs("DOKSRunningClusterModel:"+confError_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

	}

	_, _ = utils.SendLog(ctx.Data.Company, "Running Cluster updated successfully "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
	utils.Publisher(utils.ResponseSchema{
		Status:  true,
		Message: "CLuster updated successfully",
		InfraId: cluster.InfraId,
		Token:   token,
		Action:  models.Update,
	}, ctx)

	return types.CustomCPError{}

}

func FetchStatus(credentials vault.DOCredentials, ctx utils.Context) (KubeClusterStatus, types.CustomCPError) {

	cluster, err := GetKubernetesCluster(ctx)
	if err != nil {
		ctx.SendLogs("DOKSClusterModel:  Fetch -  Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubeClusterStatus{}, types.CustomCPError{StatusCode: 500, Error: "Got error while connecting to the database", Description: err.Error()}
	}
	if string(cluster.CloudplexStatus) == strings.ToLower(string(models.New)) {
		cpErr := types.CustomCPError{Error: "Unable to fetch status - Cluster is not deployed yet", Description: "Unable to fetch state - Cluster is not deployed yet", StatusCode: 409}
		return KubeClusterStatus{}, cpErr
	}
	if cluster.CloudplexStatus == models.Deploying || cluster.CloudplexStatus == models.Terminating || cluster.CloudplexStatus == models.ClusterTerminated {
		cpErr := types.CustomCPError{Error: "Cluster is in " +
			string(cluster.CloudplexStatus) + " state", Description: "Cluster is in " +
			string(cluster.CloudplexStatus) + " state", StatusCode: 409}
		return KubeClusterStatus{}, cpErr
	}

	if cluster.CloudplexStatus != models.ClusterCreated {
		customErr, err := db.GetError(cluster.InfraId, ctx.Data.Company, models.DOKS, ctx)
		if err != nil {
			return KubeClusterStatus{}, types.CustomCPError{Error: "Error occurred while getting cluster status from database",
				Description: "Error occurred while getting cluster status from database",
				StatusCode:  500}
		}
		if customErr.Err != (types.CustomCPError{}) {
			return KubeClusterStatus{}, customErr.Err
		}
	}
	doksOps, err := GetDOKS(credentials)
	if err != nil {
		ctx.SendLogs("DOKSClusterModel:  Fetch -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubeClusterStatus{}, types.CustomCPError{StatusCode: 500, Error: "Error in fetching cluster status ", Description: err.Error()}
	}

	err1 := doksOps.init(ctx)
	if err1 != (types.CustomCPError{}) {
		return KubeClusterStatus{}, err1
	}

	status, errr := doksOps.fetchStatus(ctx, cluster)
	if errr != (types.CustomCPError{}) {
		return KubeClusterStatus{}, errr
	}
	status.Status = cluster.CloudplexStatus

	return status, errr
}

func TerminateCluster(credentials vault.DOCredentials, token string, ctx utils.Context) (customError types.CustomCPError) {

	/*	publisher := utils.Notifier{}
		confError := publisher.Init_notifier()
		if confError != nil {
			ctx.SendLogs("DOKSClusterModel:  Terminate Cluster : "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			customError.StatusCode = 500
			customError.Description = confError.Error()
			err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.DOKS, ctx, customError)
			if err != nil {
				ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			return customError
		}*/

	doksOps, err := GetDOKS(credentials)
	if err != nil {
		ctx.SendLogs("DOKSClusterModel:  Terminate Cluster : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		customError.StatusCode = 500
		customError.Description = err.Error()
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.DOKS, ctx, customError)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return customError
	}
	cluster, err := GetKubernetesCluster(ctx)
	if err != nil {
		ctx.SendLogs("DOKSClusterModel : Terminate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		customError = types.CustomCPError{StatusCode: 500, Error: "Error in applying agent", Description: err.Error()}
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.DOKS, ctx, customError)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return customError
	}

	_, err2 := CompareClusters(ctx)
	if err2 != nil  &&  !(strings.Contains(err2.Error(),"Nothing to update")){
		oldCluster ,err := GetPreviousDOKSCluster(ctx)
		if err != nil {
			utils.SendLog(ctx.Data.Company, err.Error(), "error", cluster.InfraId)
			cpErr := types.CustomCPError{Description: err.Error(), Error: "Error occurred while updating cluster status in database", StatusCode: 500}
			err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.DOKS, ctx, cpErr)
			if err != nil {
				ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
		err_ := UpdateKubernetesCluster(oldCluster, ctx)
		if err_ != nil {
			utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.InfraId)
			cpErr := types.CustomCPError{Description: err_.Error(), Error: "Error occurred while updating cluster status in database", StatusCode: 500}
			err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.DOKS, ctx, cpErr)
			if err != nil {
				ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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


	if cluster.CloudplexStatus == "" || cluster.CloudplexStatus == "new" {
		text := "DOKSClusterModel : Terminate - Cannot terminate a new cluster"
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		customError = types.CustomCPError{StatusCode: 500, Error: "Error in applying agent", Description: err.Error()}
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.DOKS, ctx, customError)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: text,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Terminate,
		}, ctx)
		return types.CustomCPError{StatusCode: 500, Error: "Error in cluster termination", Description: text}
	}

	cluster.CloudplexStatus = (models.Terminating)
	_, _ = utils.SendLog(ctx.Data.Company, "Terminating cluster: "+cluster.Name, models.LOGGING_LEVEL_INFO, cluster.InfraId)

	err_ := UpdateKubernetesCluster(cluster, ctx)
	if err_ != nil {
		utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.InfraId)
		cpErr := types.CustomCPError{Description: err_.Error(), Error: "Error occurred while updating cluster status in database", StatusCode: 500}
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.DOKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
	errr := doksOps.init(ctx)
	if errr != (types.CustomCPError{}) {
		cluster.CloudplexStatus = models.ClusterTerminationFailed
		err = UpdateKubernetesCluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("DOKSClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			_, _ = utils.SendLog(ctx.Data.Company, "Error in cluster updation in mongo: "+cluster.Name, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
			_, _ = utils.SendLog(ctx.Data.Company, err.Error(), models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
			return errr
		}
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.DOKS, ctx, errr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: errr.Error,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Terminate,
		}, ctx)
		return errr
	}

	errr = doksOps.deleteCluster(cluster, ctx)
	if errr != (types.CustomCPError{}) {
		_, _ = utils.SendLog(ctx.Data.Company, "Cluster termination failed: "+cluster.Name, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
		cluster.CloudplexStatus = models.ClusterTerminationFailed
		err = UpdateKubernetesCluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("DOKSClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			_, _ = utils.SendLog(ctx.Data.Company, "Error in cluster updation in mongo: "+cluster.Name, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
			_, _ = utils.SendLog(ctx.Data.Company, err.Error(), models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
			err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.DOKS, ctx, errr)
			if err != nil {
				ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			utils.Publisher(utils.ResponseSchema{
				Status:  false,
				Message: errr.Error,
				InfraId: cluster.InfraId,
				Token:   token,
				Action:  models.Terminate,
			}, ctx)
			return errr
		}
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.DOKS, ctx, errr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: errr.Error,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Terminate,
		}, ctx)
		return types.CustomCPError{}
	}
	cluster.ID = ""
	for _, pool := range cluster.NodePools {
		pool.PoolStatus = false
		pool.ID = ""
	}
	cluster.CloudplexStatus = models.ClusterTerminated

	err = UpdateKubernetesCluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs("DOKSClusterModel : Terminate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		_, _ = utils.SendLog(ctx.Data.Company, "Error in cluster updation in mongo: "+cluster.Name, models.LOGGING_LEVEL_ERROR, cluster.InfraId)
		_, _ = utils.SendLog(ctx.Data.Company, err.Error(), models.LOGGING_LEVEL_ERROR, cluster.InfraId)
		cpErr := types.CustomCPError{StatusCode: 500, Error: "Error in applying agent", Description: err.Error()}
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.DOKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("DOKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	_, _ = utils.SendLog(ctx.Data.Company, "Cluster terminated successfully "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
	utils.Publisher(utils.ResponseSchema{
		Status:  true,
		Message: "Cluster terminated successfully",
		InfraId: cluster.InfraId,
		Token:   token,
		Action:  models.Terminate,
	}, ctx)
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
	if err1 != (types.CustomCPError{}) {
		return config, err1
	}

	config, errr := doksOps.GetKubeConfig(ctx, cluster)
	if errr != (types.CustomCPError{}) {
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
	infraID := ctx.Data.InfraId
	data2, err := woodpecker.GetCertificate(infraID, token, ctx)
	if err != nil {
		ctx.SendLogs("DOKubernetesClusterController : Apply Agent -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.CustomCPError{StatusCode: 500, Error: "Error in applying agent", Description: "Agent Deployment failed " + err.Error()}
	}

	filePath := "/tmp/" + companyId + "/" + infraID + "/"
	cmd := "mkdir -p " + filePath + " && echo '" + data2 + "'>" + filePath + "agent.yaml"

	//	key:= beego.AppConfig.String("jump_host_ssh_key")
	//	dat, err := ioutil.ReadFile(key)

	output, err := models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("DOKubernetesClusterController : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.CustomCPError{StatusCode: 500, Error: "Error in applying agent", Description: err.Error()}
	}

	cmd = "sudo docker run --rm --name " + companyId + infraID + " -e DIGITALOCEAN_ACCESS_TOKEN=" + credentials.AccessKey + " -e cluster=" + clusterName + " -e yamlFile=" + filePath + "agent.yaml -v " + filePath + ":" + filePath + " " + models.DOAuthContainerName

	output, err = models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("DOKubernetesClusterController : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.CustomCPError{StatusCode: 500, Error: "Error in applying agent", Description: "Agent Deployment failed " + err.Error()}
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
	if errr != (types.CustomCPError{}) {
		return &godo.KubernetesOptions{}, errr
	}

	options, confErrr := doksOps.GetServerConfig(ctx)
	if confErrr != (types.CustomCPError{}) {
		return &godo.KubernetesOptions{}, confErrr
	}

	return options, types.CustomCPError{}
}

func ValidateDOKSData(cluster KubernetesCluster, ctx utils.Context) error {
	if cluster.InfraId == "" {

		return errors.New("infrastructure Id is empty")

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

func CompareClusters(ctx utils.Context) (diff.Changelog, error) {
	cluster, err := GetKubernetesCluster(ctx)
	if err != nil {
		return diff.Changelog{}, err
	}

	oldCluster, err := GetPreviousDOKSCluster(ctx)
	if err != nil && strings.Contains(err.Error(), "not found") {
		return diff.Changelog{}, errors.New("Nothing to update")
	}

	previousPoolCount := len(oldCluster.NodePools)
	newPoolCount := len(cluster.NodePools)

	difCluster, err := diff.Diff(oldCluster, cluster)
	if len(difCluster) < 2 && previousPoolCount == newPoolCount {
		return diff.Changelog{}, errors.New("Nothing to update")
	} else if err != nil {
		return diff.Changelog{}, errors.New("Error in comparing differences:" + err.Error())
	}
	return difCluster, nil
}

func UpdateCluster(cluster KubernetesCluster, ctx utils.Context, doksOps DOKS, credentials vault.DOCredentials,token string) types.CustomCPError {
	err := doksOps.UpdateCluster(&cluster, ctx, credentials)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token )
		return err
	}

	oldCluster, err1 := GetPreviousDOKSCluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, err,token)
	}

	oldCluster.AutoUpgrade = cluster.AutoUpgrade

	err1 = AddPreviousDOKSCluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in updating running cluster autoupgrade", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}
	return types.CustomCPError{}
}
func UpdateKubernetesVersion(cluster KubernetesCluster, ctx utils.Context, doksOps DOKS, credentials vault.DOCredentials,token string) types.CustomCPError {

	err := doksOps.UpgradeKubernetesVersion(&cluster, ctx, credentials)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}

	oldCluster, err1 := GetPreviousDOKSCluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, err,token)
	}

	oldCluster.KubeVersion = cluster.KubeVersion

	err1 = AddPreviousDOKSCluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in updating running cluster kubernetes version", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}
	return types.CustomCPError{}
}
func UpdateNodePool(cluster KubernetesCluster, poolIndex int, ctx utils.Context, doksOps DOKS, credentials vault.DOCredentials,token string) types.CustomCPError {

	err := doksOps.UpdateNodePool(cluster.NodePools[poolIndex], ctx, cluster.ID, cluster.InfraId, credentials)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}

	oldCluster, err1 := GetPreviousDOKSCluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, err,token)
	}
	for _, pool := range oldCluster.NodePools {
		if pool.ID == cluster.NodePools[poolIndex].ID {
			pool = cluster.NodePools[poolIndex]
		}
	}

	err1 = AddPreviousDOKSCluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in updating nodepool " + cluster.NodePools[poolIndex].Name, Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}

	return types.CustomCPError{}
}
func AddNodepool(cluster KubernetesCluster, ctx utils.Context, doksOps DOKS, pools []*KubernetesNodePool, credentials vault.DOCredentials,token string) types.CustomCPError {

	for _, pool := range pools {

		oldCluster, err1 := GetPreviousDOKSCluster(ctx)
		if err1 != nil {
			return updationFailedError(cluster, ctx, types.CustomCPError{
				StatusCode:  int(models.CloudStatusCode),
				Error:       "Error in adding nodepool in running cluster",
				Description: err1.Error(),
			},token)
		}

		poolId, err := doksOps.addNodepool(*pool, ctx, cluster.ID, cluster.InfraId, credentials)
		if err != (types.CustomCPError{}) {
			updationFailedError(cluster, ctx, err,token)
			return err
		}
		pool.ID = poolId
		pool.PoolStatus = true
		oldCluster.NodePools = append(oldCluster.NodePools, pool)

		err1 = AddPreviousDOKSCluster(oldCluster, ctx, true)
		if err1 != nil {
			return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in adding nodepool in running cluster", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
		}
		err1 = AddKubernetesCluster(cluster, ctx)
		utils.SendLog(ctx.Data.Company, "Nodepool  "+pool.Name+" added in "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
	}

	return types.CustomCPError{}
}
func DeleteNodepool(cluster KubernetesCluster, ctx utils.Context, doksOps DOKS, poolId string, credentials vault.DOCredentials,token string) types.CustomCPError {

	err := doksOps.deleteNodepool(ctx, poolId, cluster.ID, cluster.InfraId, credentials)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}

	oldCluster, err1 := GetPreviousDOKSCluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{
			StatusCode:  int(models.CloudStatusCode),
			Error:       "Error in deleting nodepool in running cluster",
			Description: err1.Error(),
		},token)
	}

	for _, pool := range oldCluster.NodePools {
		if pool.ID == poolId {
			pool = nil
		}
	}
	err1 = AddPreviousDOKSCluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in deleting nodepool in running cluster", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}

	return types.CustomCPError{}
}

func updationFailedError(cluster KubernetesCluster, ctx utils.Context, err types.CustomCPError,token string) types.CustomCPError {
	publisher := utils.Notifier{}

	errr := publisher.Init_notifier()
	if errr != nil {
		PrintError(ctx, errr.Error(), cluster.Name)
		ctx.SendLogs(errr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := types.CustomCPError{StatusCode: 500, Error: "Error in deploying DOKS Cluster", Description: errr.Error()}

		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.DOKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("DOKSRunningClusterModel: Update - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		return cpErr
	}

	cluster.CloudplexStatus = models.ClusterUpdateFailed
	confError := UpdateKubernetesCluster(cluster, ctx)
	if confError != nil {
		PrintError(ctx, confError.Error(), cluster.Name)
		ctx.SendLogs("DOKSRunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	}

	utils.SendLog(ctx.Data.Company, "Error in running cluster update : "+err.Description, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)

	err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.DOKS, ctx, err)
	if err_ != nil {
		ctx.SendLogs("DOKSRunningClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
