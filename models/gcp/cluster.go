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
	VPCName          string        `json:"vpc_name" bson:"vpc_name"`
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
	PoolRole            models.PoolRole    `json:"pool_role" bson:"pool_role"`
	ServiceAccountEmail string             `json:"service_account_email" bson:"service_account_email"`
	Nodes               []*Node            `json:"nodes" bson:"nodes"`
	KeyInfo             key_utils.AZUREKey `json:"key_info" bson:"key_info"`
	EnableScaling       bool               `json:"enable_scaling" bson:"enable_scaling"`
	EnablePublicIP      bool               `json:"enable_public_ip" bson:"enable_public_ip"`
	Scaling             AutoScaling        `json:"auto_scaling" bson:"auto_scaling"`
	Tags                []string           `json:"tags" bson:"tags"`
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
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", "", err
	}
	var project Project
	err = json.Unmarshal(data.([]byte), &project)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	if oldCluster.Status == string(models.Deploying) && update {
		ctx.SendLogs("cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("cluster is in deploying state")
	}
	if oldCluster.Status == string(models.Terminating) && update {
		ctx.SendLogs("cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("cluster is in terminating state")
	}

	err = checkMasterPools(cluster)
	if err != nil { //cluster found
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
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

func DeployCluster(cluster Cluster_Def, credentials GcpCredentials, companyId string, token string, ctx utils.Context) (confError error) {
	publisher := utils.Notifier{}
	confError = publisher.Init_notifier()
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
		ctx.SendLogs(confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		//PrintError(confError, cluster.Name, cluster.ProjectId)
		return confError
	}
	gcp, err := GetGCP(credentials)
	if err != nil {
		ctx.SendLogs("gcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	err = gcp.init()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.Status = "Cluster creation failed"
		confError = UpdateCluster(cluster, false, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
			ctx.SendLogs("gcpClusterModel :"+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}

	utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	cluster, confError = gcp.createCluster(cluster, token, ctx)

	if confError != nil {
		ctx.SendLogs("gcpClusterModel :"+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		PrintError(confError, cluster.Name, cluster.ProjectId, companyId)

		cluster.Status = "Cluster creation failed"
		confError = UpdateCluster(cluster, false, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
			ctx.SendLogs("gcpClusterModel :"+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		}

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return nil

	}

	for _, pool := range cluster.NodePools {
		for _, node := range pool.Nodes {
			node.NodeState = ""
			node.PublicIp = ""
			node.PrivateIp = ""
		}
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
	return nil
}

func FetchStatus(credentials GcpCredentials, token, projectId, companyId string, ctx utils.Context) (Cluster_Def, error) {
	cluster, err := GetCluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterModel: Deploy - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
		ctx.SendLogs("GcpClusterModel : Status - Failed to get latest status "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err
	}

	for _, pool := range cluster.NodePools {
		var keyInfo key_utils.AZUREKey
		bytes, err := vault.GetSSHKey(string(models.GCP), pool.KeyInfo.KeyName, token, ctx, "")
		if err != nil {
			ctx.SendLogs("vm fetched failed with error: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return Cluster_Def{}, err
		}
		keyInfo, err = key_utils.AzureKeyConversion(bytes, ctx)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return Cluster_Def{}, err

		}
		pool.KeyInfo = keyInfo
	}
	return cluster, nil
}

func GetAllSSHKeyPair(token string, ctx utils.Context) (keys interface{}, err error) {
	keys, err = vault.GetAllSSHKey(string(models.GCP), ctx, token, "")
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
		ctx.SendLogs("GcpClusterModel ServiceAccounts - Failed to list service accounts "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	return serviceAccounts, err
}

func TerminateCluster(cluster Cluster_Def, credentials GcpCredentials, companyId string, ctx utils.Context) error {

	publisher := utils.Notifier{}
	pub_err := publisher.Init_notifier()
	if pub_err != nil {
		ctx.SendLogs("GcpClusterModel :"+pub_err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return pub_err
	}

	cluster, err := GetCluster(cluster.ProjectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterModel : Terminate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	if cluster.Status == "" || cluster.Status == "new" {
		text := "GcpClusterModel :Cannot terminate a new cluster"
		ctx.SendLogs(text+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return errors.New(text)
	}

	gcp, err := GetGCP(credentials)
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	cluster.Status = string(models.Terminating)
	utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	err = gcp.init()
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.Status = "Cluster Termination Failed"
		err = UpdateCluster(cluster, false, ctx)
		if err != nil {
			ctx.SendLogs("GcpClusterModel Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)

			return err
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}

	err = gcp.deleteCluster(cluster, ctx)
	if err != nil {

		utils.SendLog(companyId, "Cluster termination failed: "+cluster.Name, "error", cluster.ProjectId)

		cluster.Status = "Cluster Termination Failed"
		err = UpdateCluster(cluster, false, ctx)
		if err != nil {
			ctx.SendLogs("GcpClusterModel :Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
		ctx.SendLogs("GcpClusterModel :Terminate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
		utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}
	utils.SendLog(companyId, "Cluster terminated successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	return nil
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

	return err
}

func GetAllMachines(credentials GcpCredentials, ctx utils.Context) (Machines, error) {
	gcp, err := GetGCP(credentials)
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Machines{}, err
	}
	err = gcp.init()
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Machines{}, err
	}

	machines, err := gcp.GetAllMachines(ctx)
	if err != nil {
		return Machines{}, err
	}

	var mach Machines
	for _, machine := range machines.Items {
		mach.MachineName = append(mach.MachineName, machine.Name)
	}

	return mach, nil
}
func GetRegions() ([]models.Region, error) {


	regionInfo, err :=api_handler.GetGcpRegion()
	if err != nil {
		return []models.Region{}, err
	}

	return regionInfo, nil
}
func GetZones(credentials GcpCredentials, ctx utils.Context) ([]string, error) {
	gcp, err := GetGCP(credentials)
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []string{}, err
	}
	err = gcp.init()
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []string{}, err
	}

	regionInfo, err := gcp.GetZones(ctx)
	if err != nil {
		return []string{}, err
	}
	var zones []string
	for _, zone := range regionInfo.Zones {
		zone := zone[len(zone)-1:]
		zones = append(zones, zone)
	}

	return zones, nil
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
func ValidateProfile(profile []byte,region,zone string, ctx utils.Context)  error {
	credentials := GcpResponse{}

	err := json.Unmarshal(profile, &credentials.Credentials.AccountData)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	jsonData, err := json.Marshal(credentials.Credentials.AccountData)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	credentials.Credentials.RawData = string(jsonData)
	credentials.Credentials.Region = region
	credentials.Credentials.Zone = zone
	_, err = govalidator.ValidateStruct(credentials.Credentials)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	cre := GcpCredentials{
	AccountData: credentials.Credentials.AccountData,
	RawData:     string(jsonData),
	Region:      region,
	Zone:        zone,
	}
	gcp, err := GetGCP(cre)
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return  err
	}
	err = gcp.init()
	if err != nil {
		ctx.SendLogs("GcpClusterModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return  err
	}

	_, err = gcp.GetAllMachines(ctx)
	if err != nil {
		return  err
	}


	return  nil
}