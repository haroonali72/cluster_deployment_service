package gcp

import (
	"antelope/models"
	"antelope/models/db"
	"antelope/models/logging"
	"antelope/models/utils"
	"antelope/models/vault"
	"errors"
	"fmt"
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
	ID          bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name        string        `json:"name" bson:"name"`
	NodeCount   int64         `json:"node_count" bson:"node_count"`
	MachineType string        `json:"machine_type" bson:"machine_type"`
	Image       Image         `json:"image" bson:"image"`
	Volume      Volume        `json:"volume" bson:"volume"`
	PoolSubnet  string        `json:"subnet_id" bson:"subnet_id"`
	PoolRole    string        `json:"pool_role" bson:"pool_role"`
	Nodes       []*Node       `json:"nodes" bson:"nodes"`
	KeyInfo     utils.Key     `json:"key_info" bson:"key_info"`
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
	DiskType     models.GCPDiskType `json:"disk_type" bson:"disk_type"`
	IsBlank      bool               `json:"is_blank" bson:"is_blank"`
	Size         int64              `json:"disk_size" bson:"disk_size"`
	EnableVolume bool               `json:"enable_volume" bson:"enable_volume"`
}

func checkClusterSize(cluster Cluster_Def) error {
	for _, pools := range cluster.NodePools {
		if pools.NodeCount > 3 {
			return errors.New("Nodepool can't have more than 3 nodes")
		}
	}
	return nil
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
		logging.SendLog("Cluster creation failed : "+name, "error", projectId)
		logging.SendLog(confError.Error(), "error", projectId)

	}
}

func DeployCluster(cluster Cluster_Def, credentials string) (confError error) {
	gcp, err := GetGCP(credentials, "")
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

	logging.SendLog("Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)
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
		publisher.Notify(cluster.Name, "Status Available")
		return nil

	}
	cluster.Status = "Cluster Created"

	confError = UpdateCluster(cluster, false)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId)
		publisher.Notify(cluster.Name, "Status Available")
		return confError
	}
	logging.SendLog("Cluster created successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.Name, "Status Available")

	return nil
}

func FetchStatus(credentials string, projectId string) (Cluster_Def, error) {
	cluster, err := GetCluster(projectId)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
		return Cluster_Def{}, err
	}

	gcp, err := GetGCP(credentials, "")
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
	keys, err = vault.GetAllSSHKey(string(models.GCP), logging.Context{})
	if err != nil {
		beego.Error(err.Error())
		return keys, err
	}
	return keys, nil
}

func TerminateCluster(cluster Cluster_Def, credentials string) error {
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
		publisher.Notify(cluster.Name, "Status Available")
		return err
	}

	gcp, err := GetGCP(credentials, "")
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

		logging.SendLog("Cluster termination failed: "+cluster.Name, "error", cluster.ProjectId)
		logging.SendLog(err.Error(), "error", cluster.ProjectId)

		cluster.Status = "Cluster Termination Failed"
		err = UpdateCluster(cluster, false)
		if err != nil {
			beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
			logging.SendLog("Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			logging.SendLog(err.Error(), "error", cluster.ProjectId)
			publisher.Notify(cluster.Name, "Status Available")
			return err
		}
		publisher.Notify(cluster.Name, "Status Available")
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
		logging.SendLog("Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
		logging.SendLog(err.Error(), "error", cluster.ProjectId)
		publisher.Notify(cluster.Name, "Status Available")
		return err
	}
	logging.SendLog("Cluster terminated successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.Name, "Status Available")

	return nil
}
