package azure

import (
	"antelope/models"
	"antelope/models/db"
	"antelope/models/logging"
	"antelope/models/utils"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

type SSHKeyPair struct {
	Name        string `json:"name" bson:"name",omitempty"`
	FingerPrint string `json:"fingerprint" bson:"fingerprint"`
}
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
	ResourceGroup    string        `json:"resource_group" bson:"resource_group"`
}

type NodePool struct {
	ID                 bson.ObjectId      `json:"_id" bson:"_id,omitempty"`
	Name               string             `json:"name" bson:"name"`
	NodeCount          int64              `json:"node_count" bson:"node_count"`
	MachineType        string             `json:"machine_type" bson:"machine_type"`
	Image              ImageReference     `json:"image" bson:"image"`
	PoolSubnet         string             `json:"subnet_id" bson:"subnet_id"`
	PoolSecurityGroups []*string          `json:"security_group_id" bson:"security_group_id"`
	Nodes              []*VM              `json:"nodes" bson:"nodes"`
	PoolRole           string             `json:"pool_role" bson:"pool_role"`
	AdminUser          string             `json:"user_name" bson:"user_name",omitempty"`
	KeyInfo            Key                `json:"key_info" bson:"key_info"`
	BootDiagnostics    DiagnosticsProfile `json:"boot_diagnostics" bson:"boot_diagnostics"`
	OsDisk             models.OsDiskType  `json:"os_disk_type" bson:"os_disk_type"`
}
type Key struct {
	CredentialType models.CredentialsType `json:"credential_type"  bson:"credential_type"`
	NewKey         models.KeyType         `json:"key_type"  bson:"key_type"`
	KeyName        string                 `json:"key_name" bson:"key_name"`
	AdminPassword  string                 `json:"admin_password" bson:"admin_password",omitempty"`
	PrivateKey     string                 `json:"private_key" bson:"private_key",omitempty"`
	PublicKey      string                 `json:"public_key" bson:"public_key",omitempty"`
	Cloud          models.Cloud           `json:"cloud" bson:"cloud"`
}
type VM struct {
	CloudId   *string `json:"cloud_id" bson:"cloud_id,omitempty"`
	NodeState *string `json:"node_state" bson:"node_state,omitempty"`
	Name      *string `json:"name" bson:"name,omitempty"`
	PrivateIP *string `json:"private_ip" bson:"private_ip,omitempty"`
	PublicIP  *string `json:"public_ip" bson:"public_ip,omitempty"`
	UserName  *string `json:"user_name" bson:"user_name,omitempty"`
	PAssword  *string `json:"password" bson:"password,omitempty"`
}
type DiagnosticsProfile struct {
	Enable            bool   `json:"enable" bson :"enable"`
	NewStroageAccount bool   `json:"new_storage_account" bson:"new_storage_account"`
	StorageAccountId  string `json:"storage_account_id" bson:"storage_account_id"`
}

type ImageReference struct {
	ID        bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Publisher string        `json:"publisher" bson:"publisher,omitempty"`
	Offer     string        `json:"offer" bson:"offer,omitempty"`
	Sku       string        `json:"sku" bson:"sku,omitempty"`
	Version   string        `json:"version" bson:"version,omitempty"`
	ImageId   string        `json:"image_id" bson:"image_id,omitempty"`
}

func CreateCluster(cluster Cluster_Def) error {

	fmt.Printf("%+v", cluster.NodePools[0].BootDiagnostics)
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
	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoAzureClusterCollection, cluster)
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
	c := session.DB(mc.MongoDb).C(mc.MongoAzureClusterCollection)
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
	c := session.DB(mc.MongoDb).C(mc.MongoAzureClusterCollection)
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
	c := session.DB(mc.MongoDb).C(mc.MongoAzureClusterCollection)
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

	splits := strings.Split(credentials, ":")

	azure := AZURE{
		ID:           splits[0],
		Key:          splits[1],
		Tenant:       splits[2],
		Subscription: splits[3],
		Region:       splits[4],
	}
	err := azure.init()
	if err != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId)
		return err
	}

	publisher := utils.Notifier{}
	pub_err := publisher.Init_notifier()
	if pub_err != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId)
		return pub_err
	}

	logging.SendLog("Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	cluster, err = azure.createCluster(cluster)
	if err != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId)

		confError = azure.CleanUp(cluster)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId)
		}

		cluster.Status = "Cluster creation failed"
		err = UpdateCluster(cluster)
		if err != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId)
		}
		publisher.Notify(cluster.Name, "Status Available")
		return nil

	}
	cluster.Status = "Cluster Created"

	err = UpdateCluster(cluster)
	if err != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId)
		publisher.Notify(cluster.Name, "Status Available")
		return err
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
	splits := strings.Split(credentials, ":")
	azure := AZURE{
		ID:           splits[0],
		Key:          splits[1],
		Tenant:       splits[2],
		Subscription: splits[3],
		Region:       splits[4],
	}
	err = azure.init()
	if err != nil {
		return Cluster_Def{}, err
	}

	c, e := azure.fetchStatus(cluster)
	if e != nil {
		beego.Error("Cluster model: Status - Failed to get lastest status ", e.Error())
		return Cluster_Def{}, e
	}
	/*err = UpdateCluster(c)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
		return Cluster_Def{}, err
	}*/
	return c, nil
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
	splits := strings.Split(credentials, ":")
	azure := AZURE{
		ID:           splits[0],
		Key:          splits[1],
		Tenant:       splits[2],
		Subscription: splits[3],
		Region:       splits[4],
	}
	err = azure.init()
	if err != nil {
		return err
	}

	err = azure.terminateCluster(cluster)

	if err != nil {

		beego.Error(err.Error())

		logging.SendLog("Cluster termination failed: "+cluster.Name, "error", cluster.ProjectId)
		logging.SendLog(err.Error(), "error", cluster.ProjectId)

		cluster.Status = "Cluster Termination Failed"
		err = UpdateCluster(cluster)
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
		var nodes []*VM
		pools.Nodes = nodes
	}
	err = UpdateCluster(cluster)
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
func InsertSSHKeyPair(key Key) (err error) {
	key.Cloud = models.Azure
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
func GetAllSSHKeyPair() (keys []*Key, err error) {

	session, err := db.GetMongoSession()
	if err != nil {
		beego.Error("Cluster model: Get - Got error while connecting to the database: ", err)
		return keys, err
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoSshKeyCollection)
	err = c.Find(bson.M{"cloud": models.Azure}).All(&keys)
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
	err = c.Find(bson.M{"cloud": models.Azure, "key_name": keyname}).One(&keys)
	if err != nil {
		return keys, err
	}
	return keys, nil
}
