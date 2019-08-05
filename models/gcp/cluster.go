package gcp

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/db"
	"antelope/models/utils"
	"antelope/models/vault"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Cluster_Def struct {
	ID               bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	ProjectId        string        `json:"project_id" bson:"project_id"`
	Name             string        `json:"name" bson:"name"`
	Status           string        `json:"status" bson:"status"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools        []*NodePool   `json:"node_pools" bson:"node_pools"`
	NetworkName      string        `json:"network_name" bson:"network_name"`
}

type NodePool struct {
	ID            bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name          string        `json:"name" bson:"name"`
	NodeCount     int64         `json:"node_count" bson:"node_count"`
	MachineType   string        `json:"machine_type" bson:"machine_type"`
	Image         Image         `json:"image" bson:"image"`
	Volume        Volume        `json:"volume" bson:"volume"`
	EnableVolume  bool          `json:"is_external" bson:"is_external"`
	PoolSubnet    string        `json:"subnet_id" bson:"subnet_id"`
	PoolRole      string        `json:"pool_role" bson:"pool_role"`
	Nodes         []*Node       `json:"nodes" bson:"nodes"`
	KeyInfo       utils.Key     `json:"key_info" bson:"key_info"`
	EnableScaling bool          `json:"enable_scaling" bson:"enable_scaling"`
	Scaling       AutoScaling   `json:"auto_scaling" bson:"auto_scaling"`
}
type AutoScaling struct {
	MaxScalingGroupSize int64 `json:"max_scaling_group_size" bson:"max_scaling_group_size"`
}
type Node struct {
	ID            bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Url           string        `json:"url" bson:"url"`
	Status        string        `json:"status" bson:"status"`
	CurrentAction string        `json:"current_action" bson:"current_action"`
}

type Image struct {
	ID      bson.ObjectId `json:"_id" bson:"_id,omitempty"`
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

func GetRegion(projectId string) (string, string, error) {
	url := beego.AppConfig.String("raccoon_url") + "/" + projectId

	data, err := api_handler.GetAPIStatus(url, utils.Context{})
	if err != nil {
		beego.Error(err.Error(), "error")
		return "", "", err
	}
	var project Project
	err = json.Unmarshal(data.([]byte), &project)
	if err != nil {
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

func IsValidGcpCredentials(profileId, region, zone string) (bool, GcpCredentials) {
	credentials := GcpResponse{}

	response, err := vault.GetCredentialProfile("gcp", profileId, utils.Context{})
	if err != nil {
		return false, GcpCredentials{}
	}

	err = json.Unmarshal(response, &credentials)
	if err != nil {
		beego.Error(err.Error())
		return false, GcpCredentials{}
	}

	jsonData, err := json.Marshal(credentials.Credentials.AccountData)
	if err != nil {
		beego.Error(err.Error())
		return false, GcpCredentials{}
	}

	credentials.Credentials.RawData = string(jsonData)
	credentials.Credentials.Region = region
	credentials.Credentials.Zone = zone
	_, err = govalidator.ValidateStruct(credentials.Credentials)
	if err != nil {
		beego.Error(err.Error())
		return false, GcpCredentials{}
	}

	return true, credentials.Credentials
}

func CreateCluster(cluster Cluster_Def) error {
	_, err := GetCluster(cluster.ProjectId)
	if err == nil { //cluster found
		text := fmt.Sprintf("Cluster model: Create - Cluster for project'%s' already exists in the database: ", cluster.Name)
		beego.Error(text, err)
		return errors.New(text)
	}
	session, err := db.GetMongoSession()
	if err != nil {
		beego.Error("Cluster model: Delete - Got error while connecting to the database: ", err)
		return err
	}
	defer session.Close()

	err = checkClusterSize(cluster)
	if err != nil { //cluster found
		beego.Error(err.Error())
		return err
	}
	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoGcpClusterCollection, cluster)
	if err != nil {
		beego.Error("Cluster model: Create - Got error inserting cluster to the database: ", err)
		return err
	}

	return nil
}

func GetCluster(projectId string) (cluster Cluster_Def, err error) {
	session, err1 := db.GetMongoSession()
	if err1 != nil {
		beego.Error("Cluster model: Get - Got error while connecting to the database: ", err1)
		return Cluster_Def{}, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGcpClusterCollection)
	err = c.Find(bson.M{"project_id": projectId}).One(&cluster)
	if err != nil {
		beego.Error(err.Error())
		return Cluster_Def{}, err
	}

	return cluster, nil
}

func GetAllCluster() (clusters []Cluster_Def, err error) {
	session, err1 := db.GetMongoSession()
	if err1 != nil {
		beego.Error("Cluster model: GetAll - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGcpClusterCollection)
	err = c.Find(bson.M{}).All(&clusters)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}

	return clusters, nil
}

func UpdateCluster(cluster Cluster_Def, update bool) error {
	oldCluster, err := GetCluster(cluster.ProjectId)
	if err != nil {
		text := fmt.Sprintf("Cluster model: Update - Cluster '%s' does not exist in the database: ", cluster.Name)
		beego.Error(text, err)
		return errors.New(text)
	}
	if oldCluster.Status == "Cluster Created" && update {
		beego.Error("Cluster is in runnning state")
		return errors.New("Cluster is in runnning state")
	}
	err = DeleteCluster(cluster.ProjectId)
	if err != nil {
		beego.Error("Cluster model: Update - Got error deleting cluster: ", err)
		return err
	}

	cluster.CreationDate = oldCluster.CreationDate
	cluster.ModificationDate = time.Now()

	err = CreateCluster(cluster)
	if err != nil {
		beego.Error("Cluster model: Update - Got error creating cluster: ", err)
		return err
	}

	return nil
}

func DeleteCluster(projectId string) error {
	session, err := db.GetMongoSession()
	if err != nil {
		beego.Error("Cluster model: Delete - Got error while connecting to the database: ", err)
		return err
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGcpClusterCollection)
	err = c.Remove(bson.M{"project_id": projectId})
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}

func PrintError(confError error, name, projectId string) {
	if confError != nil {
		beego.Error(confError.Error())
		utils.SendLog("Cluster creation failed : "+name, "error", projectId)
		utils.SendLog(confError.Error(), "error", projectId)

	}
}

func DeployCluster(cluster Cluster_Def, credentials GcpCredentials) (confError error) {
	gcp, err := GetGCP(credentials)
	if err != nil {
		return err
	}
	err = gcp.init()
	if err != nil {
		return err
	}

	publisher := utils.Notifier{}
	confError = publisher.Init_notifier()
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId)
		return confError
	}

	utils.SendLog("Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	cluster, confError = gcp.createCluster(cluster)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId)

		confError = gcp.deleteCluster(cluster)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId)
		}

		cluster.Status = "Cluster creation failed"
		confError = UpdateCluster(cluster, false)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", utils.Context{})
		return nil

	}
	cluster.Status = "Cluster Created"

	confError = UpdateCluster(cluster, false)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId)
		publisher.Notify(cluster.ProjectId, "Status Available", utils.Context{})
		return confError
	}
	utils.SendLog("Cluster created successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", utils.Context{})

	return nil
}

func FetchStatus(credentials GcpCredentials, projectId string) (Cluster_Def, error) {
	cluster, err := GetCluster(projectId)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
		return Cluster_Def{}, err
	}

	gcp, err := GetGCP(credentials)
	if err != nil {
		return Cluster_Def{}, err
	}
	err = gcp.init()
	if err != nil {
		return Cluster_Def{}, err
	}

	c, e := gcp.fetchClusterStatus(cluster)
	if e != nil {
		beego.Error("Cluster model: Status - Failed to get lastest status ", e.Error())
		return Cluster_Def{}, e
	}

	return c, nil
}

func GetAllSSHKeyPair() (keys []string, err error) {
	keys, err = vault.GetAllSSHKey(string(models.GCP), utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		return keys, err
	}
	return keys, nil
}

func TerminateCluster(cluster Cluster_Def, credentials GcpCredentials) error {
	publisher := utils.Notifier{}
	pub_err := publisher.Init_notifier()
	if pub_err != nil {
		beego.Error(pub_err.Error())
		return pub_err
	}

	cluster, err := GetCluster(cluster.ProjectId)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
		return err
	}
	if cluster.Status != "Cluster Created" {
		beego.Error("Cluster model: Cluster is not in created state ")
		publisher.Notify(cluster.ProjectId, "Status Available", utils.Context{})
		return err
	}

	gcp, err := GetGCP(credentials)
	if err != nil {
		return err
	}
	err = gcp.init()
	if err != nil {
		return err
	}

	err = gcp.deleteCluster(cluster)

	if err != nil {
		beego.Error(err.Error())

		utils.SendLog("Cluster termination failed: "+cluster.Name, "error", cluster.ProjectId)
		utils.SendLog(err.Error(), "error", cluster.ProjectId)

		cluster.Status = "Cluster Termination Failed"
		err = UpdateCluster(cluster, false)
		if err != nil {
			beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
			utils.SendLog("Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			utils.SendLog(err.Error(), "error", cluster.ProjectId)
			publisher.Notify(cluster.ProjectId, "Status Available", utils.Context{})
			return err
		}
		publisher.Notify(cluster.ProjectId, "Status Available", utils.Context{})
		return nil
	}

	cluster.Status = "Cluster Terminated"

	for _, pools := range cluster.NodePools {
		var nodes []*Node
		pools.Nodes = nodes
	}
	err = UpdateCluster(cluster, false)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
		utils.SendLog("Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
		utils.SendLog(err.Error(), "error", cluster.ProjectId)
		publisher.Notify(cluster.ProjectId, "Status Available", utils.Context{})
		return err
	}
	utils.SendLog("Cluster terminated successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", utils.Context{})

	return nil
}
