package gcp

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/db"
	"antelope/models/key_utils"
	rbac_athentication "antelope/models/rbac_authentication"
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

type Cluster_Def struct {
	ID               bson.ObjectId `json:"-" bson:"_id,omitempty"`
	ProjectId        string        `json:"project_id" bson:"project_id"`
	Name             string        `json:"name" bson:"name"`
	Status           string        `json:"status" bson:"status"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools        []*NodePool   `json:"node_pools" bson:"node_pools"`
	NetworkName      string        `json:"network_name" bson:"network_name"`
	CompanyId        string        `json:"company_id" bson:"company_id"`
	TokenName        string        `json:"token_name" bson:"token_name"`
}

type NodePool struct {
	ID                  bson.ObjectId      `json:"-" bson:"_id,omitempty"`
	Name                string             `json:"name" bson:"name"`
	PoolId              string             `json:"pool_id" bson:"pool_id"`
	NodeCount           int64              `json:"node_count" bson:"node_count"`
	MachineType         string             `json:"machine_type" bson:"machine_type"`
	Image               Image              `json:"image" bson:"image"`
	Volume              Volume             `json:"volume" bson:"volume"`
	RootVolume          Volume             `json:"root_volume" bson:"root_volume"`
	EnableVolume        bool               `json:"is_external" bson:"is_external"`
	PoolSubnet          string             `json:"subnet_id" bson:"subnet_id"`
	PoolRole            string             `json:"pool_role" bson:"pool_role"`
	ServiceAccountEmail string             `json:"service_account_email" bson:"service_account_email"`
	Nodes               []*Node            `json:"nodes" bson:"nodes"`
	KeyInfo             key_utils.AZUREKey `json:"key_info" bson:"key_info"`
	EnableScaling       bool               `json:"enable_scaling" bson:"enable_scaling"`
	Scaling             AutoScaling        `json:"auto_scaling" bson:"auto_scaling"`
}

type AutoScaling struct {
	MaxScalingGroupSize int64       `json:"max_scaling_group_size" bson:"max_scaling_group_size"`
	State               models.Type `json:"status" bson:"status"`
}

type Node struct {
	ID        bson.ObjectId `json:"-" bson:"_id,omitempty"`
	CloudId   string        `json:"cloud_id" bson:"cloud_id,omitempty"`
	Url       string        `json:"url" bson:"url,omitempty"`
	NodeState string        `json:"node_state" bson:"node_state,omitempty"`
	Name      string        `json:"name" bson:"name,omitempty"`
	PrivateIp string        `json:"private_ip" bson:"private_ip"`
	PublicIp  string        `json:"public_ip" bson:"public_ip"`
	Username  string        `json:"user_name" bson:"user_name"`
}

type Image struct {
	ID      bson.ObjectId `json:"-" bson:"_id,omitempty"`
	Project string        `json:"project" bson:"project"`
	Family  string        `json:"family" bson:"family"`
}

type Volume struct {
	DiskType models.GCPDiskType `json:"disk_type" bson:"disk_type"`
	IsBlank  bool               `json:"is_blank" bson:"is_blank"`
	Size     int64              `json:"disk_size" bson:"disk_size"`
}
type GcpResponse struct {
	Credentials GcpCredentials `json:"credentials"`
}
type GcpCredentials struct {
	AccountData AccountData `json:"account_data"`
	RawData     string      `json:"raw_account_data" valid:"required"`
	Region      string      `json:"region"`
	Zone        string      `json:"zone"`
}

type AccountData struct {
	Type          string `json:"type" valid:"required"`
	ProjectId     string `json:"project_id" valid:"required"`
	PrivateKeyId  string `json:"private_key_id" valid:"required"`
	PrivateKey    string `json:"private_key" valid:"required"`
	ClientEmail   string `json:"client_email" valid:"required"`
	ClientId      string `json:"client_id" valid:"required"`
	AuthUri       string `json:"auth_uri" valid:"required"`
	TokenUri      string `json:"token_uri" valid:"required"`
	AuthProvider  string `json:"auth_provider_x509_cert_url" valid:"required"`
	ClientCertUrl string `json:"client_x509_cert_url" valid:"required"`
}
type Project struct {
	ProjectData Data `json:"data"`
}
type Data struct {
	Region string `json:"region"`
	Zone   string `json:"zone"`
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
	return update
}
func GetNetwork(token, projectId string, ctx utils.Context) error {

	url := getNetworkHost("gcp", projectId)

	_, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}
func GetRegion(token, projectId string, ctx utils.Context) (string, string, error) {
	url := beego.AppConfig.String("raccoon_url") + models.ProjectGetEndpoint
	if strings.Contains(url, "{projectId}") {
		url = strings.Replace(url, "{projectId}", projectId, -1)
	}
	data, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error(), "error")
		return "", "", err
	}
	var project Project
	err = json.Unmarshal(data.([]byte), &project)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error(), "error")
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

	response, err := vault.GetCredentialProfile("gcp", profileId, token, ctx)
	if err != nil {
		beego.Error(err.Error())
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return false, GcpCredentials{}
	}

	err = json.Unmarshal(response, &credentials)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return false, GcpCredentials{}
	}
	jsonData, err := json.Marshal(credentials.Credentials.AccountData)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return false, GcpCredentials{}
	}
	credentials.Credentials.RawData = string(jsonData)
	credentials.Credentials.Region = region
	credentials.Credentials.Zone = zone
	_, err = govalidator.ValidateStruct(credentials.Credentials)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return false, GcpCredentials{}
	}

	return true, credentials.Credentials
}

func CreateCluster(cluster Cluster_Def, ctx utils.Context) error {
	_, err := GetCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err == nil {
		text := fmt.Sprintf("Cluster model: Create - Cluster for project'%s' already exists in the database: ", cluster.Name)
		ctx.SendLogs("GcpClusterModel: "+text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(text, err)
		return errors.New(text)
	}

	session, err := db.GetMongoSession()
	if err != nil {
		ctx.SendLogs("GcpClusterModel: error while connecting to database "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		beego.Error("Cluster model: Delete - Got error while connecting to the database: ", err)
		return err
	}
	defer session.Close()

	err = checkClusterSize(cluster)
	if err != nil { //cluster found
		ctx.SendLogs("GcpClusterModel: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return err
	}

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
		ctx.SendLogs("GcpClusterModel: error inserting cluster to database "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Cluster model: Create - Got error inserting cluster to the database: ", err)
		return err
	}

	ctx.SendLogs(" GCP Cluster: "+cluster.Name+" of Project Id: "+cluster.ProjectId+" created ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	return nil
}

func GetCluster(projectId string, companyId string, ctx utils.Context) (cluster Cluster_Def, err error) {
	session, err1 := db.GetMongoSession()
	if err1 != nil {
		ctx.SendLogs("GcpGetClusterModel: error while connecting to database "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Cluster model: Get - Got error while connecting to the database: ", err1)
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
	ctx.SendLogs(" Get gcp Cluster "+cluster.Name+" of Project Id: "+cluster.ProjectId+"", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	return cluster, nil
}

func GetAllCluster(data rbac_athentication.List, ctx utils.Context) (clusters []Cluster_Def, err error) {
	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}
	session, err1 := db.GetMongoSession()
	if err1 != nil {
		ctx.SendLogs("GcpClusterModel: error while connecting to database "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Cluster model: GetAll - Got error while connecting to the database: ", err1)
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

	ctx.SendLogs(" Get all GCP Cluster ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	return clusters, nil
}

func UpdateCluster(cluster Cluster_Def, update bool, ctx utils.Context) error {
	oldCluster, err := GetCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err != nil {
		text := fmt.Sprintf("Cluster model: Update - Cluster '%s' does not exist in the database: ", cluster.Name)
		ctx.SendLogs("GcpClusterModel: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(text, err)
		return errors.New(text)
	}

	if oldCluster.Status == "Cluster Created" && update {
		if !checkScalingChanges(&oldCluster, &cluster) {
			ctx.SendLogs("Cluster is in runnning state ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return errors.New("Cluster is in runnning state")
		} else {
			cluster = oldCluster
		}
	}

	err = DeleteCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterModel: Update - Got error deleting cluster "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Cluster model: Update - Got error deleting cluster: ", err)
		return err
	}

	cluster.CreationDate = oldCluster.CreationDate
	cluster.ModificationDate = time.Now()

	err = CreateCluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterModel: Update - Got error creating cluster "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Cluster model: Update - Got error creating cluster: ", err)
		return err
	}
	ctx.SendLogs(" GCP Cluster "+cluster.Name+" of Project Id: "+cluster.ProjectId+"updated in database ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	return nil
}

func DeleteCluster(projectId, companyId string, ctx utils.Context) error {
	session, err := db.GetMongoSession()
	if err != nil {
		ctx.SendLogs("GcpClusterModel: error while connecting to database "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Cluster model: Delete - Got error while connecting to the database: ", err)
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
	ctx.SendLogs(" GCP Cluster of Project Id: "+projectId+"deleted from database ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	return nil
}

func PrintError(confError error, name, projectId string, companyId string) {
	if confError != nil {
		beego.Error(confError.Error())
		utils.SendLog(companyId, "Cluster creation failed : "+name, "error", projectId)
		utils.SendLog(companyId, confError.Error(), "error", projectId)

	}
}

func DeployCluster(cluster Cluster_Def, credentials GcpCredentials, companyId string, token string, ctx utils.Context) (confError error) {
	gcp, err := GetGCP(credentials)
	if err != nil {
		ctx.SendLogs("gcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	err = gcp.init()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	publisher := utils.Notifier{}
	confError = publisher.Init_notifier()
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
		ctx.SendLogs(confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		//PrintError(confError, cluster.Name, cluster.ProjectId)
		return confError
	}

	utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	cluster, confError = gcp.createCluster(cluster, token, ctx)

	if confError != nil {
		ctx.SendLogs("gcpClusterModel :"+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		PrintError(confError, cluster.Name, cluster.ProjectId, companyId)

		confError = gcp.deleteCluster(cluster, ctx)
		if confError != nil {
			ctx.SendLogs("gcpClusterModel :"+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
		}

		cluster.Status = "Cluster creation failed"
		confError = UpdateCluster(cluster, false, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
			ctx.SendLogs("gcpClusterModel :"+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return nil

	}
	cluster.Status = "Cluster Created"

	confError = UpdateCluster(cluster, false, ctx)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
		ctx.SendLogs("gcpClusterModel :"+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return confError
	}

	utils.SendLog(companyId, "Cluster created successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	ctx.SendLogs("GCP Cluster "+cluster.Name+" of Project Id: "+cluster.ProjectId+"deployed to GCP ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	return nil
}

func FetchStatus(credentials GcpCredentials, token, projectId, companyId string, ctx utils.Context) (Cluster_Def, error) {
	cluster, err := GetCluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
		return cluster, err
	}

	gcp, err := GetGCP(credentials)
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err
	}
	err = gcp.init()
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err
	}

	err = gcp.fetchClusterStatus(&cluster, ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Cluster model: Status - Failed to get latest status ", err.Error())
		return cluster, err
	}

	for _, pool := range cluster.NodePools {
		var keyInfo key_utils.AZUREKey
		bytes, err := vault.GetSSHKey(string(models.GCP), pool.KeyInfo.KeyName, token, ctx)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			beego.Error("vm fetched failed with error: " + err.Error())
			return Cluster_Def{}, err
		}
		keyInfo, err = key_utils.AzureKeyConversion(bytes, ctx)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return Cluster_Def{}, err

		}
		pool.KeyInfo = keyInfo
	}

	ctx.SendLogs(" GCP Cluster "+cluster.Name+" of Project Id: "+projectId+"fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	return cluster, nil
}

func GetAllSSHKeyPair(token string, ctx utils.Context) (keys interface{}, err error) {
	keys, err = vault.GetAllSSHKey(string(models.GCP), ctx, token)
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return keys, err
	}
	return keys, nil
}

func GetAllServiceAccounts(credentials GcpCredentials, ctx utils.Context) (serviceAccounts []string, err error) {
	gcp, err := GetGCP(credentials)
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	err = gcp.init()
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	serviceAccounts, err = gcp.listServiceAccounts(ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Cluster model: ServiceAccounts - Failed to list service accounts ", err.Error())
		return nil, err
	}

	return serviceAccounts, err
}

func TerminateCluster(cluster Cluster_Def, credentials GcpCredentials, companyId string, ctx utils.Context) error {
	publisher := utils.Notifier{}
	pub_err := publisher.Init_notifier()
	if pub_err != nil {
		ctx.SendLogs("GcpClusterModel :"+pub_err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(pub_err.Error())
		return pub_err
	}

	cluster, err := GetCluster(cluster.ProjectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Cluster model: Terminate - Got error while connecting to the database: ", err.Error())
		return err
	}
	if cluster.Status == "" || cluster.Status == "new" {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Cluster model: Cannot terminate a new cluster")
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}

	gcp, err := GetGCP(credentials)
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	err = gcp.init()
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	err = gcp.deleteCluster(cluster, ctx)

	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		beego.Error(err.Error())

		utils.SendLog(companyId, "Cluster termination failed: "+cluster.Name, "error", cluster.ProjectId)
		utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)

		cluster.Status = "Cluster Termination Failed"
		err = UpdateCluster(cluster, false, ctx)
		if err != nil {
			ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

			beego.Error("Cluster model: Terminate - Got error while connecting to the database: ", err.Error())
			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
			publisher.Notify(cluster.ProjectId, "Status Available", ctx)
			return err
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return nil
	}

	cluster.Status = "Cluster Terminated"

	for _, pools := range cluster.NodePools {
		var nodes []*Node
		pools.Nodes = nodes
	}
	err = UpdateCluster(cluster, false, ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		beego.Error("Cluster model: Terminate - Got error while connecting to the database: ", err.Error())
		utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
		utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}
	utils.SendLog(companyId, "Cluster terminated successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	ctx.SendLogs(" GCP Cluster "+cluster.Name+" of Project Id: "+cluster.ProjectId+"terminated by ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	return nil
}

func GetSSHkey(keyName, userName, token, teams string, ctx utils.Context) (privateKey string, err error) {

	privateKey, err = key_utils.GenerateKey(models.GCP, keyName, userName, token, teams, ctx)
	if err != nil {

		return "", err
	}
	return privateKey, err
}
