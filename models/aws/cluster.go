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
	"antelope/models/logging"
)
type SSHKeyPair struct {
	Name              string `json:"name" bson:"name",omitempty"`
	FingerPrint    	  string        `json:"fingerprint" bson:"fingerprint"`
}
type Cluster_Def struct {
	ID               bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	EnvironmentId    string        `json:"environment_id" bson:"environment_id"`
	Kube_Credentials interface{}	`json:"kube_credentials" bson:"kube_credentials"`
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
	PublicDNS	 string	`json:"public_dns" bson:"public_dns,omitempty"`
	PrivateDNS string	`json:"private_dns" bson:"private_dns,omitempty"`
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


	err = db.InsertInMongo(db.MongoAwsClusterCollection, cluster)
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

	c := session.DB(db.MongoDb).C(db.MongoAwsClusterCollection)
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

	c := session.DB(db.MongoDb).C(db.MongoAwsClusterCollection)
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

	c := session.DB(db.MongoDb).C(db.MongoAwsClusterCollection)
	err = c.Remove(bson.M{"environment_id": envId})
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


	publisher := Notifier{}
	pub_err := publisher.init_notifier()
	if pub_err != nil {
		beego.Error(pub_err.Error())
		return pub_err
	}

	logging.SendLog("Creating Cluster : " + cluster.Name,"info",cluster.EnvironmentId)
	createdPools , err:= aws.createCluster(cluster)

	if err != nil {
		beego.Error(err.Error())

		logging.SendLog("Cluster creation failed : " + cluster.Name,"error",cluster.EnvironmentId)
		logging.SendLog(err.Error(),"error",cluster.EnvironmentId)

		cluster.Status = "Cluster creation failed"

		err = UpdateCluster(cluster)
		if err != nil {
			beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
			logging.SendLog("Cluster updation failed in mongo: " + cluster.Name,"error",cluster.EnvironmentId)
			logging.SendLog(err.Error(),"error",cluster.EnvironmentId)
			publisher.notify(cluster.Name,"Status Available")
			return err
		}
		publisher.notify(cluster.Name,"Status Available")

	}

	for index, nodepool := range cluster.NodePools {

		var updatedNodes []*Node

		for _, createdPool := range createdPools {

			if createdPool.PoolName == nodepool.Name {

				for _, inst := range createdPool.Instances {

					var node Node
					beego.Info(*inst.Tags[0].Value , *inst.Tags[0].Value)
					if *inst.Tags[0].Key== "Name" {
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
	beego.Info(cluster.Status + cluster.EnvironmentId)
	err = UpdateCluster(cluster)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
		logging.SendLog("Cluster updation failed in mongo: " + cluster.Name,"error",cluster.EnvironmentId)
		logging.SendLog(err.Error(),"error",cluster.EnvironmentId)
		publisher.notify(cluster.Name,"Status Available")
		return err
	}
	logging.SendLog("Cluster created successfully " + cluster.Name,"info",cluster.EnvironmentId)
	publisher.notify(cluster.Name,"Status Available")

	return nil
}
func FetchStatus( credentials string,envId string) (Cluster_Def , error){

	cluster, err := GetCluster(envId)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
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


	publisher := Notifier{}
	pub_err := publisher.init_notifier()
	if pub_err != nil {
		beego.Error(pub_err.Error())
		return pub_err
	}

	if cluster.Status != "Cluster Created" {
		beego.Error("Cluster model: Cluster is not in created state ")
		publisher.notify(cluster.Name,"Status Available")
		return err
	}

	logging.SendLog("Terminating cluster: " + cluster.Name,"info",cluster.EnvironmentId)


	err = aws.terminateCluster(cluster)

	if err != nil {

		beego.Error(err.Error())

		logging.SendLog("Cluster termination failed: " + cluster.Name,"error",cluster.EnvironmentId)
		logging.SendLog(err.Error(),"error",cluster.EnvironmentId)

		cluster.Status = "Cluster termination failed"
		err = UpdateCluster(cluster)
		if err != nil {
			beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
			logging.SendLog("Error in cluster updation in mongo: " + cluster.Name,"error",cluster.EnvironmentId)
			logging.SendLog(err.Error(),"error",cluster.EnvironmentId)
			publisher.notify(cluster.Name,"Status Available")
			return err
		}
		publisher.notify(cluster.Name,"Status Available")

	}

	cluster.Status = "Cluster terminated"
	err = UpdateCluster(cluster)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
		logging.SendLog("Error in cluster updation in mongo: " + cluster.Name,"error",cluster.EnvironmentId)
		logging.SendLog(err.Error(),"error",cluster.EnvironmentId)
		publisher.notify(cluster.Name,"Status Available")
		return err
	}
	logging.SendLog("Cluster terminated successfully " + cluster.Name,"info",cluster.EnvironmentId)
	publisher.notify(cluster.Name,"Status Available")

	return nil
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
	if e != nil {
		beego.Error("Cluster model: Status - Failed to get ssh key pairs ", e.Error())
		return nil, e
	}
	k:= fillKeyInfo(keys)

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