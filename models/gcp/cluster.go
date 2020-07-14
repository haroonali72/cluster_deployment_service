package gcp

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/db"
	"antelope/models/key_utils"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/astaxie/beego"

	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

var machines = []byte(`[
	"c2-standard-16",
	"c2-standard-30",
	"c2-standard-4",
	"c2-standard-60",
	"c2-standard-8",
	"e2-highcpu-16",
	"e2-highcpu-2",
	"e2-highcpu-4",
	"e2-highcpu-8",
	"e2-highmem-16",
	"e2-highmem-2",
	"e2-highmem-4",
	"e2-highmem-8",
	"e2-medium",
	"e2-micro",
	"e2-small",
	"e2-standard-16",
	"e2-standard-2",
	"e2-standard-4",
	"e2-standard-8",
	"f1-micro",
	"g1-small",
	"m1-megamem-96",
	"m1-ultramem-160",
	"m1-ultramem-40",
	"m1-ultramem-80",
	"n1-highcpu-16",
	"n1-highcpu-2",
	"n1-highcpu-32",
	"n1-highcpu-4",
	"n1-highcpu-64",
	"n1-highcpu-8",
	"n1-highcpu-96",
	"n1-highmem-16",
	"n1-highmem-2",
	"n1-highmem-32",
	"n1-highmem-4",
	"n1-highmem-64",
	"n1-highmem-8",
	"n1-highmem-96",
	"n1-megamem-96",
	"n1-standard-1",
	"n1-standard-16",
	"n1-standard-2",
	"n1-standard-32",
	"n1-standard-4",
	"n1-standard-64",
	"n1-standard-8",
	"n1-standard-96",
	"n1-ultramem-160",
	"n1-ultramem-40",
	"n1-ultramem-80",
	"n2-highcpu-16",
	"n2-highcpu-2",
	"n2-highcpu-32",
	"n2-highcpu-4",
	"n2-highcpu-48",
	"n2-highcpu-64",
	"n2-highcpu-8",
	"n2-highcpu-80",
	"n2-highmem-16",
	"n2-highmem-2",
	"n2-highmem-32",
	"n2-highmem-4",
	"n2-highmem-48",
	"n2-highmem-64",
	"n2-highmem-8",
	"n2-highmem-80",
	"n2-standard-16",
	"n2-standard-2",
	"n2-standard-32",
	"n2-standard-4",
	"n2-standard-48",
	"n2-standard-64",
	"n2-standard-8",
	"n2-standard-80"
]`)

type Cluster_Def struct {
	ID               bson.ObjectId `json:"-" bson:"_id,omitempty"`
	ProjectId        string        `json:"project_id" bson:"project_id" validate:"required" description:"Project ID of the cluster [required]"`
	Name             string        `json:"name" bson:"name" validate:"required" description:"Name of the cluster [required]"`
	Status           models.Type   `json:"status" bson:"status" validate:"eq=new|eq=New|eq=NEW|eq=Cluster Creation Failed|eq=Cluster Terminated" description:"Status of the project [required]"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"  validate:"eq=gcp|eq=Gcp|eq=GCP" description:"Cloud of the cluster.Valid value is gcp|GCP|Gcp [readonly]"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools        []*NodePool   `json:"node_pools" bson:"node_pools" validate:"required,dive" description:"Details of the nodepool [required]"`
	NetworkName      string        `json:"network_name" bson:"network_name" description:"Natwork of the cluster [required]"`
	VPCName          string        `json:"vpc_name" bson:"vpc_name" description:"VPC of the cluster [required]"`
	CompanyId        string        `json:"company_id" bson:"company_id" description:"CompanyId of the cluster [optional]"`
	TokenName        string        `json:"token_name" bson:"token_name"`
}

type NodePool struct {
	ID                  bson.ObjectId      `json:"-" bson:"_id,omitempty"`
	Name                string             `json:"name" bson:"name" validate:"required" description:"Name of the nodepool [required]"`
	PoolId              string             `json:"pool_id" bson:"pool_id" description:"Id of the nodepool [optional]"`
	NodeCount           int64              `json:"node_count" bson:"node_count" validate:"required" description:"Node count of pool [required]"`
	MachineType         string             `json:"machine_type" bson:"machine_type" validate:"required" description:"Machine type of the nodepool [required]"`
	Image               Image              `json:"image" bson:"image" validate:"required,dive" description:"Image of the nodepool[required]"`
	Volume              Volume             `json:"volume" bson:"volume" description:"Volume of the nodepool [optional]"`
	RootVolume          Volume             `json:"root_volume" bson:"root_volume" validate:"required,dive" description:"Root volume of the nodepool [required]"`
	EnableVolume        bool               `json:"is_external" bson:"is_external"`
	PoolSubnet          string             `json:"subnet_id" bson:"subnet_id"  description:"Subnet of the nodepool [required]"`
	PoolRole            models.PoolRole    `json:"pool_role" bson:"pool_role" validate:"required,dive" description:"Role of the nodepool.Valid values are master and slave. [required]"`
	ServiceAccountEmail string             `json:"service_account_email" bson:"service_account_email" validate:"required" description:"Service account of the nodepool [required]"`
	Nodes               []*Node            `json:"nodes" bson:"nodes" validate:"required,dive" description:"Details of the node [required]"`
	KeyInfo             key_utils.AZUREKey `json:"key_info" bson:"key_info" validate:"required,dive" description:"Details of the key [required]"`
	EnableScaling       bool               `json:"enable_scaling" bson:"enable_scaling" validate:"required" description:"To enable scaling [required]"`
	EnablePublicIP      bool               `json:"enable_public_ip" bson:"enable_public_ip" validate:"required" description:"To enable public ip of the instance [required]"`
	Scaling             AutoScaling        `json:"auto_scaling" bson:"auto_scaling"  description:"Details of the scaling [optional]"`
	Tags                []string           `json:"tags" bson:"tags"`
}

type AutoScaling struct {
	MaxScalingGroupSize int64       `json:"max_scaling_group_size" bson:"max_scaling_group_size" validate:"required" description:"Max count for the scaling [required]"`
	State               models.Type `json:"status" bson:"status" description:"State of the scaling [readonly]"`
}

type Node struct {
	ID        bson.ObjectId `json:"-" bson:"_id,omitempty"`
	CloudId   string        `json:"cloud_id" bson:"cloud_id,omitempty" description:"Cloud id of the node [readonly]"`
	Url       string        `json:"url" bson:"url,omitempty" description:"URL of the node [readonly]"`
	NodeState string        `json:"node_state" bson:"node_state,omitempty" description:"State of the node [readonly]"`
	Name      string        `json:"name" bson:"name,omitempty"  description:"Name of the node [readonly]"`
	PrivateIp string        `json:"private_ip" bson:"private_ip" description:"Private IP of the node [readonly]"`
	PublicIp  string        `json:"public_ip" bson:"public_ip" description:"Public IP of the node [readonly]"`
	Username  string        `json:"user_name" bson:"user_name" description:"Username of the node [optional]"`
}

type Image struct {
	ID      bson.ObjectId `json:"-" bson:"_id,omitempty" description:"Id of the image [readonly]"`
	Project string        `json:"project" bson:"project" validate:"required" description:"Project of the image [required]"`
	Family  string        `json:"family" bson:"family" validate:"required" description:"family of the image [required]"`
}

type Volume struct {
	DiskType models.GCPDiskType `json:"disk_type" bson:"disk_type" validate:"required" description:"Type of the disk [required]"`
	IsBlank  bool               `json:"is_blank" bson:"is_blank"`
	Size     int64              `json:"disk_size" bson:"disk_size" validate:"required" description:"Size of the disk [required]"`
}

type GcpResponse struct {
	Credentials GcpCredentials `json:"credentials" description:"Gcp credentials [readonly]"`
}
type GcpCredentials struct {
	AccountData AccountData `json:"account_data" description:"Account details [readonly]"`
	RawData     string      `json:"raw_account_data" valid:"required" description:"Account details [readonly]"`
	Region      string      `json:"region" description:"Region of the cloud [readonly]"`
	Zone        string      `json:"zone" description:"Zone of the cloud [readonly]"`
}

type AccountData struct {
	Type          string `json:"type" valid:"required" description:"Type of the account[readonly]"`
	ProjectId     string `json:"project_id" valid:"required" description:"Project Id of the account [readonly]"`
	PrivateKeyId  string `json:"private_key_id" valid:"required" description:"Private key Id of the account [readonly]"`
	PrivateKey    string `json:"private_key" valid:"required" description:"Private key of the account [readonly]"`
	ClientEmail   string `json:"client_email" valid:"required" description:"Client email of the account [readonly]"`
	ClientId      string `json:"client_id" valid:"required" description:"Client Id of the account [readonly]"`
	AuthUri       string `json:"auth_uri" valid:"required" description:"Auth Uri of the account [readonly]"`
	TokenUri      string `json:"token_uri" valid:"required" description:"Token Uri of the account [readonly]"`
	AuthProvider  string `json:"auth_provider_x509_cert_url" valid:"required" description:"Auth Provider of the account [readonly]"`
	ClientCertUrl string `json:"client_x509_cert_url" valid:"required" description:"Client Cert Url of the account [readonly]"`
}
type Project struct {
	ProjectData Data `json:"data"`
}
type Data struct {
	Region string `json:"region"`
	Zone   string `json:"zone"`
}
type NetworkType struct {
	IsPrivate bool `json:"is_private" bson:"is_private"`
}
type Machines struct {
	MachineName []string `json:"machine_name" bson:"machine_name"`
}

func checkMasterPools(cluster Cluster_Def) error {
	noOfMasters := 0
	for _, pools := range cluster.NodePools {
		if pools.PoolRole == models.Master {
			noOfMasters += 1
			if noOfMasters == 2 {
				return errors.New("Cluster can't have more than 1 master")
			}
		}
	}
	return nil
}
func checkScalingChanges(existingCluster, updatedCluster *Cluster_Def) bool {
	update := false
	for index, node_pool := range existingCluster.NodePools {
		if (!node_pool.EnableScaling && node_pool.EnableScaling != updatedCluster.NodePools[index].EnableScaling) || (node_pool.EnableScaling && node_pool.Scaling.MaxScalingGroupSize != updatedCluster.NodePools[index].Scaling.MaxScalingGroupSize) {
			update = true
			existingCluster.NodePools[index].EnableScaling = updatedCluster.NodePools[index].EnableScaling
			existingCluster.NodePools[index].Scaling.MaxScalingGroupSize = updatedCluster.NodePools[index].Scaling.MaxScalingGroupSize
			existingCluster.NodePools[index].Scaling.State = updatedCluster.NodePools[index].Scaling.State
		}
	}
	if update {
		existingCluster.TokenName = updatedCluster.TokenName
	}
	return update
}
func GetNetwork(token, projectId string, ctx utils.Context) (types.GCPNetwork, error) {

	url := getNetworkHost("gcp", projectId)

	data, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.GCPNetwork{}, err
	}
	var net types.GCPNetwork
	err = json.Unmarshal(data.([]byte), &net)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.GCPNetwork{}, err
	}
	return net, nil
}
func GetRegion(token, projectId string, ctx utils.Context) (string, string, error) {
	url := beego.AppConfig.String("raccoon_url") + models.ProjectGetEndpoint
	if strings.Contains(url, "{projectId}") {
		url = strings.Replace(url, "{projectId}", projectId, -1)
	}
	data, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs("Error in fetching region"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", "", err
	}
	var project Project
	err = json.Unmarshal(data.([]byte), &project.ProjectData)
	if err != nil {
		ctx.SendLogs("Error in fetching region"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", "", err
	}
	return project.ProjectData.Region, project.ProjectData.Zone, nil

}
func checkClusterSize(cluster Cluster_Def) error {
	for _, pools := range cluster.NodePools {
		if pools.NodeCount > 3 {
			return errors.New("Nodepool can't have more than 3 nodes")
		}
	}
	return nil
}

func IsValidGcpCredentials(profileId, region, token, zone string, ctx utils.Context) (bool, GcpCredentials) {
	credentials := GcpResponse{}

	_, response, err := vault.GetCredentialProfile("gcp", profileId, token, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return false, GcpCredentials{}
	}

	err = json.Unmarshal(response, &credentials)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return false, GcpCredentials{}
	}
	jsonData, err := json.Marshal(credentials.Credentials.AccountData)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return false, GcpCredentials{}
	}
	credentials.Credentials.RawData = string(jsonData)
	credentials.Credentials.Region = region
	credentials.Credentials.Zone = zone
	_, err = govalidator.ValidateStruct(credentials.Credentials)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return false, GcpCredentials{}
	}

	return true, credentials.Credentials
}

func CreateCluster(cluster Cluster_Def, ctx utils.Context) error {
	_, err := GetCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err == nil {
		text := fmt.Sprintf("Cluster model: Create - Cluster for project'%s' already exists in the database: ", cluster.Name)
		ctx.SendLogs("GcpClusterModel: "+text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	/*if subscriptionId != "" {
		err = checkCoresLimit(cluster, subscriptionId, ctx)
		if err != nil { //core size limit exceed
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}

	}
	*/
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterModel: Delete - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer session.Close()

	//err = checkClusterSize(cluster)
	//if err != nil { //cluster found
	//	ctx.SendLogs("GcpClusterModel: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	//	beego.Error(err.Error())
	//	return err
	//}

	if cluster.CreationDate.IsZero() {
		cluster.CreationDate = time.Now()
		cluster.ModificationDate = time.Now()
		if cluster.Status == "" {
			cluster.Status = "new"
		}
		cluster.Cloud = models.GCP
	}

	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoGcpClusterCollection, cluster)
	if err != nil {
		ctx.SendLogs("GcpClusterModel:  Create - Got error inserting cluster to the database:  "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}

func GetCluster(projectId string, companyId string, ctx utils.Context) (cluster Cluster_Def, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("GcpGetClusterModel:  Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGcpClusterCollection)
	err = c.Find(bson.M{"project_id": projectId, "company_id": companyId}).One(&cluster)
	if err != nil {
		beego.Error(err.Error())
		return Cluster_Def{}, err
	}
	return cluster, nil
}

func GetAllCluster(data rbac_athentication.List, ctx utils.Context) (clusters []Cluster_Def, err error) {
	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("GcpClusterModel:GetAll - Got error while connecting to the database:"+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err1
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGcpClusterCollection)
	err = c.Find(bson.M{"project_id": bson.M{"$in": copyData}}).All(&clusters)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}

	return clusters, nil
}

func UpdateCluster(cluster Cluster_Def, update bool, ctx utils.Context) error {
	oldCluster, err := GetCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err != nil {
		text := "Cluster model: Update - Cluster '%s' does not exist in the database: " + cluster.Name + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	if oldCluster.Status == models.Deploying && update {
		ctx.SendLogs("Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("Cluster is in creating state")
	}
	if oldCluster.Status == models.Terminating && update {
		ctx.SendLogs("Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("Cluster is in terminating state")
	}
	if oldCluster.Status == models.ClusterTerminationFailed && update {
		ctx.SendLogs("Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("Cluster is in termination failed state")
	}
	if oldCluster.Status == models.ClusterCreated && update {
		if !checkScalingChanges(&oldCluster, &cluster) {
			ctx.SendLogs("No changes are applicable", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return errors.New("No changes are applicable")
		} else {
			cluster = oldCluster
		}
	}

	err = DeleteCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterModel: Update - Got error deleting cluster "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	cluster.CreationDate = oldCluster.CreationDate
	cluster.ModificationDate = time.Now()

	err = CreateCluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterModel: Update - Got error creating cluster "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}

func DeleteCluster(projectId, companyId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterModel: Delete - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGcpClusterCollection)
	err = c.Remove(bson.M{"project_id": projectId, "company_id": companyId})
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	return nil
}

func PrintError(confError error, name, projectId string, companyId string) {
	if confError != nil {
		beego.Error(confError.Error())
		utils.SendLog(companyId, "Cluster creation failed : "+name, "error", projectId)
		utils.SendLog(companyId, confError.Error(), "error", projectId)
	}
}

func DeployCluster(cluster Cluster_Def, credentials GcpCredentials, companyId string, token string, ctx utils.Context) (confErr types.CustomCPError) {
	publisher := utils.Notifier{}
	confError := publisher.Init_notifier()
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
		ctx.SendLogs(confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.GCP, ctx, ApiErrors(confError, "Error in deploying cluster"))
		if err_ != nil {
			ctx.SendLogs("GCPDeployClusterModel:  Deploy - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return ApiErrors(confError, "Error in deploying cluster")
	}
	gcp, err := GetGCP(credentials)
	if err != (types.CustomCPError{}) {
		ctx.SendLogs("gcpClusterModel :"+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	err = gcp.init()
	if err != (types.CustomCPError{}) {
		ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.Status = "Cluster creation failed"
		confError = UpdateCluster(cluster, false, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
			ctx.SendLogs("gcpClusterModel :"+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.GCP, ctx, err)
		if err_ != nil {
			ctx.SendLogs("GCPDeployClusterModel:  Deploy - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}

	utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)

	pubSub := publisher.Subscribe(ctx.Data.ProjectId, ctx)

	cluster, confErr = gcp.createCluster(cluster, token, ctx)

	if confErr != (types.CustomCPError{}) {
		ctx.SendLogs("gcpClusterModel :"+confErr.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		PrintError(confError, cluster.Name, cluster.ProjectId, companyId)

		cluster.Status = models.ClusterCreationFailed
		confError = UpdateCluster(cluster, false, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
			ctx.SendLogs("gcpClusterModel :"+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		}
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.GCP, ctx, confErr)
		if err_ != nil {
			ctx.SendLogs("GCPDeployClusterModel:  Deploy - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return types.CustomCPError{}

	}

	for _, pool := range cluster.NodePools {
		for _, node := range pool.Nodes {
			node.NodeState = ""
			node.PublicIp = ""
			node.PrivateIp = ""
		}
	}
	cluster.Status = models.ClusterCreated

	confError = UpdateCluster(cluster, false, ctx)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
		ctx.SendLogs("gcpClusterModel :"+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.GCP, ctx, ApiErrors(confError, "Error in deploying cluster"))
		if err_ != nil {
			ctx.SendLogs("GCPDeployClusterModel:  Deploy - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return ApiErrors(confError, "Error in deploying cluster")
	}

	utils.SendLog(companyId, "Cluster created successfully "+cluster.Name, "info", cluster.ProjectId)

	notify := publisher.RecieveNotification(ctx.Data.ProjectId, ctx, pubSub)
	if notify {
		ctx.SendLogs("GCPClusterModel:  Notification recieved from agent", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	} else {
		ctx.SendLogs("GCPClusterModel:  Notification not recieved from agent", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		cluster.Status = models.ClusterCreationFailed
		PrintError(errors.New("Notification not recieved from the agent"), cluster.Name, cluster.ProjectId,  companyId)
		err := UpdateCluster(cluster, false, ctx)
		if err != nil {
			confErr := types.CustomCPError{StatusCode: 500, Error: "Error occured in updating cluster status in database", Description: "Error occured in updating cluster status in database"}
			PrintError(err, cluster.Name, cluster.ProjectId,  companyId)
			err = db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DO, ctx, confErr)
			if err != nil {
				ctx.SendLogs("GcpDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			publisher.Notify(cluster.ProjectId, "Status Available", ctx)
			return types.CustomCPError{StatusCode: 500, Description: err.Error(), Error: "Error occurred in updating cluster status in database"}
		}
	}

	return types.CustomCPError{}
}

func FetchStatus(credentials GcpCredentials, token, projectId, companyId string, ctx utils.Context) (Cluster_Def, types.CustomCPError) {

	cluster, err := GetCluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterModel: Deploy - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, ApiErrors(err, "Error in fetching status")
	}

	gcp, err1 := GetGCP(credentials)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GcpClusterModel :"+err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err1
	}
	err1 = gcp.init()
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GcpClusterModel :"+err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err1
	}


	err1 = gcp.fetchClusterStatus(&cluster,token, ctx)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GcpClusterModel : Status - Failed to get latest status "+err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err1
	}

	for _, pool := range cluster.NodePools {
		var keyInfo key_utils.AZUREKey
		bytes, err := vault.GetSSHKey(string(models.GCP), pool.KeyInfo.KeyName, token, ctx, "")
		if err != nil {
			ctx.SendLogs("vm fetched failed with error: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return Cluster_Def{}, ApiErrors(err, "Error in fetching status")
		}
		keyInfo, err = key_utils.AzureKeyConversion(bytes, ctx)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return Cluster_Def{}, ApiErrors(err, "Error in fetching status")

		}
		pool.KeyInfo = keyInfo
	}
	return cluster, types.CustomCPError{}
}

func GetAllSSHKeyPair(token string, ctx utils.Context) (keys interface{}, err error) {
	keys, err = vault.GetAllSSHKey(string(models.GCP), ctx, token, "")
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return keys, err
	}
	return keys, nil
}

func GetAllServiceAccounts(credentials GcpCredentials, ctx utils.Context) (serviceAccounts []string, err types.CustomCPError) {
	gcp, err := GetGCP(credentials)
	if err != (types.CustomCPError{}) {
		ctx.SendLogs("GcpClusterModel :"+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	err = gcp.init()
	if err != (types.CustomCPError{}) {
		ctx.SendLogs("GcpClusterModel :"+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	serviceAccounts, err = gcp.listServiceAccounts(ctx)
	if err != (types.CustomCPError{}) {
		ctx.SendLogs("GcpClusterModel ServiceAccounts - Failed to list service accounts "+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	return serviceAccounts, err
}

func TerminateCluster(cluster Cluster_Def, credentials GcpCredentials, companyId string, ctx utils.Context) types.CustomCPError {

	publisher := utils.Notifier{}
	pub_err := publisher.Init_notifier()
	if pub_err != nil {
		ctx.SendLogs("GcpClusterModel :"+pub_err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.GCP, ctx, ApiErrors(pub_err, "Error in terminating cluster"))
		if err_ != nil {
			ctx.SendLogs("GCPDeployClusterModel:  Deploy - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return ApiErrors(pub_err, "Error in initializing notifier")
	}

	cluster, err := GetCluster(cluster.ProjectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterModel : Terminate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.GCP, ctx, ApiErrors(err, "Error in terminating cluster"))
		if err_ != nil {
			ctx.SendLogs("GCPDeployClusterModel:  Deploy - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return ApiErrors(err, "Error in fetching cluster")
	}

	if cluster.Status == "" || strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		text := "GcpClusterModel :Cannot terminate a new cluster"
		ctx.SendLogs(text+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return ApiErrors(errors.New(text), text)
	}

	gcp, err1 := GetGCP(credentials)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GcpClusterModel :"+err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.GCP, ctx, err1)
		if err_ != nil {
			ctx.SendLogs("GCPDeployClusterModel:  Deploy - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return err1
	}

	cluster.Status = models.Terminating
	utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	err1 = gcp.init()
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GcpClusterModel :"+err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.Status = models.ClusterTerminationFailed
		err = UpdateCluster(cluster, false, ctx)
		if err != nil {
			ctx.SendLogs("GcpClusterModel Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)

			return ApiErrors(err, "Error in cluster termination")
		}
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.GCP, ctx, err1)
		if err_ != nil {
			ctx.SendLogs("GCPDeployClusterModel:  Deploy - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return ApiErrors(err, "Error in cluster termination")
	}

	err1 = gcp.deleteCluster(cluster, ctx)
	if err1 != (types.CustomCPError{}) {
		utils.SendLog(companyId, "Cluster termination failed: "+cluster.Name, "error", cluster.ProjectId)

		cluster.Status = models.ClusterTerminationFailed
		err = UpdateCluster(cluster, false, ctx)
		if err != nil {
			ctx.SendLogs("GcpClusterModel :Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
			publisher.Notify(cluster.ProjectId, "Status Available", ctx)
			return ApiErrors(err, "Error in cluster termination")
		}
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.GCP, ctx, err1)
		if err_ != nil {
			ctx.SendLogs("GCPDeployClusterModel:  Deploy - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return types.CustomCPError{}
	}

	cluster.Status = models.ClusterTerminated

	for _, pools := range cluster.NodePools {
		var nodes []*Node
		pools.Nodes = nodes
	}
	err = UpdateCluster(cluster, false, ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterModel :Terminate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
		utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.GCP, ctx, ApiErrors(err, "Error in terminating cluster"))
		if err_ != nil {
			ctx.SendLogs("GCPDeployClusterModel:  Deploy - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return ApiErrors(err, "Error in cluster termination")
	}
	utils.SendLog(companyId, "Cluster terminated successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	return types.CustomCPError{}
}

func GetSSHkey(keyName, userName, token, teams string, ctx utils.Context) (privateKey string, err error) {

	keyInfo, err := key_utils.GenerateKey(models.GCP, keyName, userName, token, teams, ctx)
	if err != nil {
		return "", err
	}
	_, err = vault.PostSSHKey(keyInfo, keyInfo.KeyName, keyInfo.Cloud, ctx, token, teams, "")
	if err != nil {
		beego.Error(err.Error())
		return "", err
	}

	return keyInfo.PrivateKey, err
}

func DeleteSSHkey(keyName, token string, ctx utils.Context) error {

	err := vault.DeleteSSHkey(string(models.GCP), keyName, token, ctx, "")
	if err != nil {
		return err
	}

	return nil
}

func GetAllMachines(credentials GcpCredentials, ctx utils.Context) (Machines, types.CustomCPError) {
	gcp, err := GetGCP(credentials)
	if err != (types.CustomCPError{}) {
		ctx.SendLogs("GcpClusterModel :"+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Machines{}, err
	}
	err = gcp.init()
	if err != (types.CustomCPError{}) {
		ctx.SendLogs("GcpClusterModel :"+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Machines{}, err
	}

	machines, err := gcp.GetAllMachines(ctx)
	if err != (types.CustomCPError{}) {
		return Machines{}, err
	}

	var mach Machines
	for _, machine := range machines.Items {
		mach.MachineName = append(mach.MachineName, machine.Name)
	}

	return mach, types.CustomCPError{}
}

func GetRegions() ([]models.Region, types.CustomCPError) {

	regionInfo, err := api_handler.GetGcpRegion()
	if err != nil {
		return []models.Region{}, ApiErrors(err, "Error in fetching regions")
	}

	return regionInfo, types.CustomCPError{}
}

func GetZones(credentials GcpCredentials, ctx utils.Context) ([]string, types.CustomCPError) {
	gcp, err := GetGCP(credentials)
	if err != (types.CustomCPError{}) {
		ctx.SendLogs("GcpClusterModel :"+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []string{}, err
	}
	err = gcp.init()
	if err != (types.CustomCPError{}) {
		ctx.SendLogs("GcpClusterModel :"+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []string{}, err
	}

	regionInfo, err := gcp.GetZones(ctx)
	if err != (types.CustomCPError{}) {
		return []string{}, err
	}
	var zones []string
	for _, zone := range regionInfo.Zones {
		zone := zone[len(zone)-1:]
		zones = append(zones, zone)
	}

	return zones, types.CustomCPError{}
}

func getCompanyAllCluster(companyId string, ctx utils.Context) (clusters []Cluster_Def, err error) {

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Cluster model: GetAllCompany - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err1
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAwsClusterCollection)
	err = c.Find(bson.M{"company_id": companyId}).All(&clusters)
	if err != nil {
		return nil, err
	}
	return clusters, nil
}

func CheckKeyUsage(keyName, companyId string, ctx utils.Context) bool {
	clusters, err := getCompanyAllCluster(companyId, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: GetAllCompany - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return true
	}
	for _, cluster := range clusters {
		for _, pool := range cluster.NodePools {
			if keyName == pool.KeyInfo.KeyName {
				ctx.SendLogs("Key is used in other projects ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return true
			}
		}
	}
	return false
}
func ValidateProfile(profile []byte, region, zone string, ctx utils.Context) types.CustomCPError {
	credentials := GcpResponse{}

	err := json.Unmarshal(profile, &credentials.Credentials.AccountData)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiErrors(err, "Error in profile validation")
	}
	jsonData, err := json.Marshal(credentials.Credentials.AccountData)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiErrors(err, "Error in profile validation")
	}
	credentials.Credentials.RawData = string(jsonData)
	credentials.Credentials.Region = region
	credentials.Credentials.Zone = zone
	_, err = govalidator.ValidateStruct(credentials.Credentials)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiErrors(err, "Error in profile validation")
	}

	cre := GcpCredentials{
		AccountData: credentials.Credentials.AccountData,
		RawData:     string(jsonData),
		Region:      region,
		Zone:        zone,
	}
	gcp, err1 := GetGCP(cre)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GcpClusterModel :"+err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err1
	}
	err1 = gcp.init()
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GcpClusterModel :"+err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err1
	}

	_, err1 = gcp.GetAllMachines(ctx)
	if err1 != (types.CustomCPError{}) {
		return err1
	}

	return types.CustomCPError{}
}
func ValidateData(cluster Cluster_Def) error {
	var machineList []string
	err := json.Unmarshal(machines, &machineList)
	if err != nil {
		return err
	}
	for _, nodepool := range cluster.NodePools {
		for _, mach := range machineList {
			if nodepool.MachineType == mach {
				return nil
			}
		}
	}
	return errors.New("Invalid machine types.Valid machines are:" + strings.Join(machineList, ","))

}
