package aws

import (
	"antelope/models"
	"antelope/models/db"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Cluster struct {
	ID               bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	EnvironmentId    string        `json:"environment_id" bson:"environment_id"`
	Name             string        `json:"name" bson:"name"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	Subclusters      []*Subcluster `json:"subclusters" bson:"subclusters"`
}

type Subcluster struct {
	ID        bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name      string        `json:"name" bson:"name"`
	NodePools []*NodePool   `json:"node_pools" bson:"node_pools"`
}

type NodePool struct {
	ID              bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name            string        `json:"name" bson:"name"`
	NodeCount       int32         `json:"node_count" bson:"node_count"`
	MachineType     string        `json:"machine_type" bson:"machine_type"`
	Ami             Ami           `json:"ami" bson:"ami"`
	SubnetId        bson.ObjectId `json:"subnet_id" bson:"subnet_id"`
	SecurityGroupId bson.ObjectId `json:"security_group_id" bson:"security_group_id"`
}

type Ami struct {
	ID       bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name     string        `json:"name" bson:"name"`
	Username string        `json:"username" bson:"username"`
}

func CreateCluster(cluster Cluster) error {
	_, err := GetCluster(cluster.Name)
	if err == nil { //cluster found
		text := fmt.Sprintf("Cluster model: Create - Cluster '%s' already exists in the database: ", cluster.Name)
		beego.Error(text, err)
		return errors.New(text)
	}

	cluster.CreationDate = time.Now()

	err = db.InsertInMongo(db.MongoAwsClusterCollection, cluster)
	if err != nil {
		beego.Error("Cluster model: Create - Got error inserting cluster to the database: ", err)
		return err
	}

	return nil
}

func GetCluster(clusterName string) (cluster Cluster, err error) {
	session, err1 := db.GetMongoSession()
	if err1 != nil {
		beego.Error("Cluster model: Get - Got error while connecting to the database: ", err1)
		return Cluster{}, err1
	}
	defer session.Close()

	c := session.DB(db.MongoDb).C(db.MongoAwsClusterCollection)
	err = c.Find(bson.M{"name": clusterName}).One(&cluster)
	if err != nil {
		beego.Error(err.Error())
		return Cluster{}, err
	}

	return cluster, nil
}

func GetAllCluster() (clusters []Cluster, err error) {
	session, err1 := db.GetMongoSession()
	if err1 != nil {
		beego.Error("Cluster model: GetAll - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()

	c := session.DB(db.MongoDb).C(db.MongoAwsClusterCollection)
	err = c.Find(bson.M{}).All(&clusters)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}

	return clusters, nil
}

func UpdateCluster(cluster Cluster) error {
	oldCluster, err := GetCluster(cluster.Name)
	if err != nil {
		text := fmt.Sprintf("Cluster model: Update - Cluster '%s' does not exist in the database: ", cluster.Name)
		beego.Error(text, err)
		return errors.New(text)
	}

	err = DeleteCluster(cluster.Name)
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

func DeleteCluster(clusterName string) error {
	session, err := db.GetMongoSession()
	if err != nil {
		beego.Error("Cluster model: Delete - Got error while connecting to the database: ", err)
		return err
	}
	defer session.Close()

	c := session.DB(db.MongoDb).C(db.MongoAwsClusterCollection)
	err = c.Remove(bson.M{"name": clusterName})
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}
