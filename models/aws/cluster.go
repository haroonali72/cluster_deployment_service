package aws

import (
	"antelope/models"
	"antelope/models/db"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"time"
	"strings"
	"github.com/aws/aws-sdk-go/service/ec2"
)
type SSHKeyPair struct {
	Name              string `json:"name" bson:"name",omitempty"`
	FingerPrint    string        `json:"fingerprint" bson:"fingerprint"`
}
type Cluster_Def struct {
	ID               bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	EnvironmentId    string        `json:"environment_id" bson:"environment_id"`
	Name             string        `json:"name" bson:"name"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	Clusters         []*Cluster `json:"clusters" bson:"clusters"`
}

type Cluster struct {
	ID        bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name      string        `json:"name" bson:"name"`
	NodePools []*NodePool   `json:"node_pools" bson:"node_pools"`
}

type NodePool struct {
	ID              bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name            string        `json:"name" bson:"name"`
	NodeCount       int64         `json:"node_count" bson:"node_count"`
	MachineType     string        `json:"machine_type" bson:"machine_type"`
	Ami             Ami           `json:"ami" bson:"ami"`
	SubnetId        string		  `json:"subnet_id" bson:"subnet_id"`
	SecurityGroupId []*string     `json:"security_group_id" bson:"security_group_id"`
	Nodes 			[]*Node		  `json:"nodes" bson:"nodes"`
	KeyName 		string 		  `json:"key_name" bson:"key_name"`
	PoolRole string               `json:"pool_role" bson:"pool_role"`
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
	AmiId string        `json:"ami_id" bson:"ami_id"`
	Username string        `json:"username" bson:"username"`
}

func CreateCluster(cluster Cluster_Def) error {
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

func GetCluster(clusterName string) (cluster Cluster_Def, err error) {
	session, err1 := db.GetMongoSession()
	if err1 != nil {
		beego.Error("Cluster model: Get - Got error while connecting to the database: ", err1)
		return Cluster_Def{}, err1
	}
	defer session.Close()

	c := session.DB(db.MongoDb).C(db.MongoAwsClusterCollection)
	err = c.Find(bson.M{"name": clusterName}).One(&cluster)
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

	c := session.DB(db.MongoDb).C(db.MongoAwsClusterCollection)
	err = c.Find(bson.M{}).All(&clusters)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}

	return clusters, nil
}

func UpdateCluster(cluster Cluster_Def) error {
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
func DeployCluster(cluster Cluster_Def, credentials string) error {
	splits := strings.Split(credentials, ":")

	aws := AWS{
		AccessKey: splits[0],
		SecretKey: splits[1],
		Region:    splits[2],
	}
	err := aws.init()
	if err != nil {
		return err
	}

	instances := aws.createCluster(cluster)
	var updatedCluster Cluster_Def
	updatedCluster = cluster

	for index, nodepool := range updatedCluster.Clusters[0].NodePools {
		var updatedNodes []*Node
		for _, inst := range instances {

			var node Node
			node.Name=""
			node.KeyName = *inst.KeyName
			node.CloudId = *inst.InstanceId
			node.NodeState = *inst.State.Name
			node.PrivateIP = *inst.PrivateIpAddress
			node.PublicIP = *inst.PublicIpAddress
			node.UserName = nodepool.Ami.Username
			node.SSHKey = ""
			updatedNodes = append(updatedNodes, &node)
		}
		updatedCluster.Clusters[0].NodePools[index].Nodes = updatedNodes
	}

	err = UpdateCluster(updatedCluster)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err)
		return err
	}
	return nil
}
func FetchStatus(clusterName string, credentials string) (Cluster_Def , error){

	cluster, err := GetCluster(clusterName)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err)
		return Cluster_Def{},err
	}
	splits := strings.Split(credentials, ":")
	aws := AWS{
		AccessKey: splits[0],
		SecretKey: splits[1],
		Region:    splits[2],
	}
	err = aws.init()
	if err != nil {
		return Cluster_Def{},err
	}

	c , e := aws.fetchStatus(cluster)
	if e != nil {
		beego.Error("Cluster model: Status - Failed to get lastest status ", err)
		return Cluster_Def{}, e
	}
	err = UpdateCluster(c)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err)
		return Cluster_Def{}, err
	}
	return c, nil
}
func GetSSHKeyPair( credentials string) ([]*SSHKeyPair , error){


	splits := strings.Split(credentials, ":")
	aws := AWS{
		AccessKey: splits[0],
		SecretKey: splits[1],
		Region:    splits[2],
	}
	err := aws.init()
	if err != nil {
		return nil,err
	}

	keys , e := aws.getSSHKey()
	k:= fillKeyInfo(keys)
	if e != nil {
		beego.Error("Cluster model: Status - Failed to get ssh key pairs ", err)
		return nil, e
	}
	return  k,nil
}
func fillKeyInfo(keys_raw []*ec2.KeyPairInfo)  (keys []*SSHKeyPair) {
	for _, key := range keys_raw {

			keys = append(keys, &SSHKeyPair{
				FingerPrint: *key.KeyFingerprint,
				Name:            *key.KeyName,
			})

	}

	return keys
}