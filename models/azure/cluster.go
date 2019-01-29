package azure

import (
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"antelope/models"
	"time"
	"fmt"
	"errors"
	"antelope/models/db"
)
type SSHKeyPair struct {
	Name              string `json:"name" bson:"name",omitempty"`
	FingerPrint    	  string        `json:"fingerprint" bson:"fingerprint"`
}
type Cluster_Def struct {
	ID               bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	EnvironmentId    string        `json:"environment_id" bson:"environment_id"`
	Name             string        `json:"name" bson:"name"`
	Status           string        `json:"status" bson:"status"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools		 []*NodePool   `json:"node_pools" bson:"node_pools"`
	NetworkName      string		   `json:"network_name" bson:"network_name"`
}

type NodePool struct {
	ID              	bson.ObjectId 		`json:"_id" bson:"_id,omitempty"`
	Name           		string       		`json:"name" bson:"name"`
	NodeCount       	int64        	 	`json:"node_count" bson:"node_count"`
	MachineType     	string        		`json:"machine_type" bson:"machine_type"`
	Ami             	Ami           		`json:"ami" bson:"ami"`
	PoolSubnet          string		  		`json:"subnet_id" bson:"subnet_id"`
	PoolSecurityGroups 	[]*string           `json:"security_group_id" bson:"security_group_id"`
	Nodes 				[]*Node		  		`json:"nodes" bson:"nodes"`
	KeyName 			string 		  		`json:"key_name" bson:"key_name"`
	PoolRole 			string              `json:"pool_role" bson:"pool_role"`
}
type Node struct {
	CloudId 	 string `json:"cloud_id" bson:"cloud_id,omitempty"`
	KeyName		 string	`json:"key_name" bson:"key_name,omitempty"`
	SSHKey 		 string	`json:"ssh_key" bson:"ssh_key,omitempty"`
	NodeState	 string	`json:"node_state" bson:"node_state,omitempty"`
	Name 		 string	`json:"name" bson:"name,omitempty"`
	PrivateIP	 string	`json:"private_ip" bson:"private_ip,omitempty"`
	PublicIP 	 string	`json:"public_ip" bson:"public_ip,omitempty"`
	UserName	 string `json:"user_name" bson:"user_name,omitempty"`
}

type Ami struct {
	ID       bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name     string        `json:"name" bson:"name"`
	AmiId 	 string        `json:"ami_id" bson:"ami_id"`
	Username string        `json:"username" bson:"username"`
}


func CreateCluster(cluster Cluster_Def) error {
	_, err := GetCluster(cluster.EnvironmentId)
	if err == nil { //cluster found
		text := fmt.Sprintf("Cluster model: Create - Cluster '%s' already exists in the database: ", cluster.Name)
		beego.Error(text, err)
		return errors.New(text)
	}


	err = db.InsertInMongo(db.MongoAzureClusterCollection, cluster)
	if err != nil {
		beego.Error("Cluster model: Create - Got error inserting cluster to the database: ", err)
		return err
	}

	return nil
}

func GetCluster(envId string ) (cluster Cluster_Def, err error) {

	session, err1 := db.GetMongoSession()
	if err1 != nil {
		beego.Error("Cluster model: Get - Got error while connecting to the database: ", err1)
		return Cluster_Def{}, err1
	}
	defer session.Close()

	c := session.DB(db.MongoDb).C(db.MongoAzureClusterCollection)
	err = c.Find(bson.M{ "environment_id":envId}).One(&cluster)
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

	c := session.DB(db.MongoDb).C(db.MongoAzureClusterCollection)
	err = c.Find(bson.M{}).All(&clusters)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}

	return clusters, nil
}

func UpdateCluster(cluster Cluster_Def) error {
	oldCluster, err := GetCluster(cluster.EnvironmentId)
	if err != nil {
		text := fmt.Sprintf("Cluster model: Update - Cluster '%s' does not exist in the database: ", cluster.Name)
		beego.Error(text, err)
		return errors.New(text)
	}

	err = DeleteCluster(cluster.EnvironmentId)
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

func DeleteCluster(envId string) error {
	session, err := db.GetMongoSession()
	if err != nil {
		beego.Error("Cluster model: Delete - Got error while connecting to the database: ", err)
		return err
	}
	defer session.Close()

	c := session.DB(db.MongoDb).C(db.MongoAzureClusterCollection)
	err = c.Remove(bson.M{"environment_id": envId})
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}
func DeployCluster(cluster Cluster_Def, credentials string) error {


	return nil
}
func FetchStatus( credentials string,envId string) (Cluster_Def , error){


	return Cluster_Def{}, nil
}
func TerminateCluster(cluster Cluster_Def, credentials string) error {


	return nil
}
