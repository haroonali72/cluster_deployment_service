package aws

import (
	"antelope/models"
	"antelope/models/db"
	"antelope/models/logging"
	"antelope/models/utils"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/aws/aws-sdk-go/service/ec2"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

type Cluster_Def struct {
	ID               bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	ProjectId        string        `json:"project_id" bson:"project_id"`
	Kube_Credentials interface{}   `json:"kube_credentials" bson:"kube_credentials"`
	Name             string        `json:"name" bson:"name"`
	Status           string        `json:"status" bson:"status"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools        []*NodePool   `json:"node_pools" bson:"node_pools"`
	NetworkName      string        `json:"network_name" bson:"network_name"`
}

type NodePool struct {
	ID                 bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name               string        `json:"name" bson:"name"`
	NodeCount          int64         `json:"node_count" bson:"node_count"`
	MachineType        string        `json:"machine_type" bson:"machine_type"`
	Ami                Ami           `json:"ami" bson:"ami"`
	PoolSubnet         string        `json:"subnet_id" bson:"subnet_id"`
	PoolSecurityGroups []*string     `json:"security_group_id" bson:"security_group_id"`
	Nodes              []*Node       `json:"nodes" bson:"nodes"`
	KeyInfo            Key           `json:"key_info" bson:"key_info"`
	PoolRole           string        `json:"pool_role" bson:"pool_role"`
}
type Node struct {
	CloudId    string `json:"cloud_id" bson:"cloud_id,omitempty"`
	KeyName    string `json:"key_name" bson:"key_name,omitempty"`
	SSHKey     string `json:"ssh_key" bson:"ssh_key,omitempty"`
	NodeState  string `json:"node_state" bson:"node_state,omitempty"`
	Name       string `json:"name" bson:"name,omitempty"`
	PrivateIP  string `json:"private_ip" bson:"private_ip,omitempty"`
	PublicIP   string `json:"public_ip" bson:"public_ip,omitempty"`
	PublicDNS  string `json:"public_dns" bson:"public_dns,omitempty"`
	PrivateDNS string `json:"private_dns" bson:"private_dns,omitempty"`
	UserName   string `json:"user_name" bson:"user_name,omitempty"`
}
type Key struct {
	KeyName     string         `json:"key_name" bson:"key_name"`
	KeyType     models.KeyType `json:"key_type" bson:"key_type"`
	KeyMaterial string         `json:"key_material" bson:"key_materials"`
}
type Ami struct {
	ID       bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name     string        `json:"name" bson:"name"`
	AmiId    string        `json:"ami_id" bson:"ami_id"`
	Username string        `json:"username" bson:"username"`
}

func CreateCluster(cluster Cluster_Def) error {
	_, err := GetCluster(cluster.ProjectId)
	if err == nil { //cluster found
		text := fmt.Sprintf("Cluster model: Create - Cluster '%s' already exists in the database: ", cluster.Name)
		beego.Error(text, err)
		return errors.New(text)
	}
	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoAwsClusterCollection, cluster)
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
	c := session.DB(mc.MongoDb).C(mc.MongoAwsClusterCollection)
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
	c := session.DB(mc.MongoDb).C(mc.MongoAwsClusterCollection)
	err = c.Find(bson.M{}).All(&clusters)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}

	return clusters, nil
}

func UpdateCluster(cluster Cluster_Def) error {
	oldCluster, err := GetCluster(cluster.ProjectId)
	if err != nil {
		text := fmt.Sprintf("Cluster model: Update - Cluster '%s' does not exist in the database: ", cluster.Name)
		beego.Error(text, err)
		return errors.New(text)
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
	c := session.DB(mc.MongoDb).C(mc.MongoAwsClusterCollection)
	err = c.Remove(bson.M{"project_id": projectId})
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
		beego.Error(err.Error())
		return err
	}

	publisher := utils.Notifier{}
	pub_err := publisher.Init_notifier()
	if pub_err != nil {
		beego.Error(pub_err.Error())
		return pub_err
	}

	logging.SendLog("Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	createdPools, err := aws.createCluster(cluster)

	if err != nil {
		beego.Error(err.Error())

		logging.SendLog("Cluster creation failed : "+cluster.Name, "error", cluster.ProjectId)
		logging.SendLog(err.Error(), "error", cluster.ProjectId)

		cluster.Status = "Cluster creation failed"

		err = UpdateCluster(cluster)
		if err != nil {
			beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
			logging.SendLog("Cluster updation failed in mongo: "+cluster.Name, "error", cluster.ProjectId)
			logging.SendLog(err.Error(), "error", cluster.ProjectId)
			publisher.Notify(cluster.ProjectId, "Status Available")
			return err
		}
		publisher.Notify(cluster.ProjectId, "Status Available")
		return err
	}

	cluster = updateNodePool(createdPools, cluster)

	err = UpdateCluster(cluster)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
		logging.SendLog("Cluster updation failed in mongo: "+cluster.Name, "error", cluster.ProjectId)
		logging.SendLog(err.Error(), "error", cluster.ProjectId)
		publisher.Notify(cluster.ProjectId, "Status Available")
		return err
	}
	logging.SendLog("Cluster created successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available")

	return nil
}
func FetchStatus(credentials string, projectId string) (Cluster_Def, error) {

	cluster, err := GetCluster(projectId)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
		return Cluster_Def{}, err
	}
	splits := strings.Split(credentials, ":")
	aws := AWS{
		AccessKey: splits[0],
		SecretKey: splits[1],
		Region:    splits[2],
	}
	err = aws.init()
	if err != nil {
		return Cluster_Def{}, err
	}

	c, e := aws.fetchStatus(cluster)
	if e != nil {
		beego.Error("Cluster model: Status - Failed to get lastest status ", e.Error())
		return Cluster_Def{}, e
	}
	err = UpdateCluster(c)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", e.Error())
		return Cluster_Def{}, err
	}
	return c, nil
}
func TerminateCluster(cluster Cluster_Def, credentials string) error {

	splits := strings.Split(credentials, ":")

	aws := AWS{
		AccessKey: splits[0],
		SecretKey: splits[1],
		Region:    splits[2],
	}
	err := aws.init()
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	publisher := utils.Notifier{}
	pub_err := publisher.Init_notifier()
	if pub_err != nil {
		beego.Error(pub_err.Error())
		return pub_err
	}

	if cluster.Status != "Cluster Created" {
		beego.Error("Cluster model: Cluster is not in created state ")
		publisher.Notify(cluster.ProjectId, "Status Available")
		return err
	}

	logging.SendLog("Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	err = aws.terminateCluster(cluster)

	if err != nil {

		beego.Error(err.Error())

		logging.SendLog("Cluster termination failed: "+cluster.Name, "error", cluster.ProjectId)
		logging.SendLog(err.Error(), "error", cluster.ProjectId)

		cluster.Status = "Cluster termination failed"
		err = UpdateCluster(cluster)
		if err != nil {
			beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
			logging.SendLog("Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			logging.SendLog(err.Error(), "error", cluster.ProjectId)
			publisher.Notify(cluster.ProjectId, "Status Available")
			return err
		}
		publisher.Notify(cluster.ProjectId, "Status Available")

	}

	for _, pools := range cluster.NodePools {
		var nodes []*Node
		pools.Nodes = nodes
	}
	cluster.Status = "Cluster terminated"
	err = UpdateCluster(cluster)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
		logging.SendLog("Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
		logging.SendLog(err.Error(), "error", cluster.ProjectId)
		publisher.Notify(cluster.ProjectId, "Status Available")
		return err
	}
	logging.SendLog("Cluster terminated successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available")

	return nil
}
func updateNodePool(createdPools []CreatedPool, cluster Cluster_Def) Cluster_Def {
	for index, nodepool := range cluster.NodePools {

		var updatedNodes []*Node

		for _, createdPool := range createdPools {

			if createdPool.PoolName == nodepool.Name {

				for _, inst := range createdPool.Instances {

					var node Node
					beego.Info(*inst.Tags[0].Value, *inst.Tags[0].Value)
					if *inst.Tags[0].Key == "Name" {
						node.Name = *inst.Tags[0].Value
					}
					node.KeyName = *inst.KeyName
					node.CloudId = *inst.InstanceId
					node.NodeState = *inst.State.Name
					node.PrivateIP = *inst.PrivateIpAddress

					if inst.PublicIpAddress != nil {
						node.PublicIP = *inst.PublicIpAddress
					}
					node.UserName = nodepool.Ami.Username
					node.SSHKey = createdPool.Key
					updatedNodes = append(updatedNodes, &node)
					beego.Info("Cluster model: Instances added")
				}
			}
		}
		beego.Info("Cluster model: updated nodes in pools")
		cluster.NodePools[index].Nodes = updatedNodes
	}
	cluster.Status = "Cluster Created"
	return cluster
}
func GetAllSSHKeyPair() (keys []*Key, err error) {

	session, err := db.GetMongoSession()
	if err != nil {
		beego.Error("Cluster model: Get - Got error while connecting to the database: ", err)
		return keys, err
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoSshKeyCollection)
	err = c.Find(bson.M{"cloud_type": "aws"}).All(&keys)
	if err != nil {
		beego.Error(err.Error())
		return keys, err
	}
	return keys, nil
}
func GetSSHKeyPair(keyname string) (keys *Key, err error) {

	session, err := db.GetMongoSession()
	if err != nil {
		beego.Error("Cluster model: Get - Got error while connecting to the database: ", err)
		return keys, err
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoSshKeyCollection)
	err = c.Find(bson.M{"cloud_type": "aws", "key_name": keyname}).All(&keys)
	if err != nil {
		return keys, err
	}
	return keys, nil
}
func InsertSSHKeyPair(key Key) (err error) {

	session, err := db.GetMongoSession()
	if err != nil {
		beego.Error("Cluster model: Get - Got error while connecting to the database: ", err)
		return err
	}
	defer session.Close()
	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoSshKeyCollection, key)
	if err != nil {
		return err
	}
	return nil
}
func GetAwsSSHKeyPair(credentials string) ([]*ec2.KeyPairInfo, error) {

	splits := strings.Split(credentials, ":")
	aws := AWS{
		AccessKey: splits[0],
		SecretKey: splits[1],
		Region:    splits[2],
	}
	err := aws.init()
	if err != nil {
		return nil, err
	}

	keys, e := aws.getSSHKey()
	if e != nil {
		beego.Error("Cluster model: Status - Failed to get ssh key pairs ", e.Error())
		return nil, e
	}

	return keys, nil
}
