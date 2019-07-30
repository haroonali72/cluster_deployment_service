package azure

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/db"
	"antelope/models/logging"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type SSHKeyPair struct {
	Name        string `json:"name" bson:"name",omitempty"`
	FingerPrint string `json:"fingerprint" bson:"fingerprint"`
}
type Cluster_Def struct {
	ID               bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	ProjectId        string        `json:"project_id" bson:"project_id" valid:"required matches:[A-Za-z0-9]`
	Name             string        `json:"name" bson:"name" valid:"required"`
	Status           string        `json:"status" bson:"status" valid:"in(NEW|new|New)"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud" valid:"in(AZURE|azure)"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools        []*NodePool   `json:"node_pools" bson:"node_pools" valid:"required"`
	NetworkName      string        `json:"network_name" bson:"network_name" valid:"required"`
	ResourceGroup    string        `json:"resource_group" bson:"resource_group" valid:"required"`
}

type NodePool struct {
	ID                 bson.ObjectId      `json:"_id" bson:"_id,omitempty"`
	Name               string             `json:"name" bson:"name" valid:"required"`
	NodeCount          int64              `json:"node_count" bson:"node_count" valid:"required matches:[0-9]"`
	MachineType        string             `json:"machine_type" bson:"machine_type" valid:"required"`
	Image              ImageReference     `json:"image" bson:"image" valid:"required"`
	Volume             Volume             `json:"volume" bson:"volume"`
	PoolSubnet         string             `json:"subnet_id" bson:"subnet_id" valid:"required"`
	PoolSecurityGroups []*string          `json:"security_group_id" bson:"security_group_id" valid:"required"`
	Nodes              []*VM              `json:"nodes" bson:"nodes"`
	PoolRole           string             `json:"pool_role" bson:"pool_role"`
	AdminUser          string             `json:"user_name" bson:"user_name,omitempty"`
	KeyInfo            utils.Key          `json:"key_info" bson:"key_info"`
	BootDiagnostics    DiagnosticsProfile `json:"boot_diagnostics" bson:"boot_diagnostics"`
	OsDisk             models.OsDiskType  `json:"os_disk_type" bson:"os_disk_type" valid:"required in(standard hdd|standard ssd|premium ssd)"`
	EnableScaling      bool               `json:"enable_scaling" bson:"enable_scaling"`
	Scaling            AutoScaling        `json:"auto_scaling" bson:"auto_scaling"`
}
type AutoScaling struct {
	MaxScalingGroupSize int64 `json:"max_scaling_group_size" bson:"max_scaling_group_size"`
}
type Key struct {
	CredentialType models.CredentialsType `json:"credential_type"  bson:"credential_type" valid:"required, in(password|key)"`
	NewKey         models.KeyType         `json:"key_type"  bson:"key_type" valid:"required in(new|cp|azure|user")"`
	KeyName        string                 `json:"key_name" bson:"key_name" valid:"required"`
	AdminPassword  string                 `json:"admin_password" bson:"admin_password",omitempty"`
	PrivateKey     string                 `json:"private_key" bson:"private_key",omitempty"`
	PublicKey      string                 `json:"public_key" bson:"public_key",omitempty"`
	Cloud          models.Cloud           `json:"cloud" bson:"cloud"`
}
type Volume struct {
	DataDisk     models.OsDiskType `json:"disk_type" bson:"disk_type"`
	Size         int32             `json:"disk_size" bson:"disk_size"`
	EnableVolume bool              `json:"enable_volume" bson:"enable_volume"`
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
	Enable            bool   `json:"enable" bson:"enable"`
	NewStroageAccount bool   `json:"new_storage_account" bson:"new_storage_account"`
	StorageAccountId  string `json:"storage_account_id" bson:"storage_account_id"`
}

type ImageReference struct {
	ID        bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Publisher string        `json:"publisher" bson:"publisher,omitempty" valid:"required"`
	Offer     string        `json:"offer" bson:"offer,omitempty" valid:"required"`
	Sku       string        `json:"sku" bson:"sku,omitempty" valid:"required"`
	Version   string        `json:"version" bson:"version,omitempty" valid:"required"`
	ImageId   string        `json:"image_id" bson:"image_id,omitempty"`
}
type Project struct {
	ProjectData Data `json:"data"`
}
type Data struct {
	Region string `json:"region"`
}

func GetRegion(projectId string, ctx logging.Context) (string, error) {
	url := beego.AppConfig.String("raccoon_url") + "/" + projectId

	data, err := api_handler.GetAPIStatus(url, ctx)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return "", err
	}
	var region Project
	err = json.Unmarshal(data.([]byte), &region)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return region.ProjectData.Region, err
	}
	return region.ProjectData.Region, nil

}
func GetNetwork(projectId string, ctx logging.Context, resourceGroup string) error {

	url := getNetworkHost("azure") + "/" + projectId

	data, err := api_handler.GetAPIStatus(url, ctx)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return err
	}

	var network types.AzureNetwork
	err = json.Unmarshal(data.([]byte), &network)
	beego.Info(string(data.([]byte)))
	beego.Info(network)
	beego.Info(network.Definition[0].ResourceGroup + " " + resourceGroup)
	if network.Definition[0].ResourceGroup != resourceGroup {
		ctx.SendSDLog("Resource group is incorrect", "error")
		return errors.New("Resource Group is in correct")
	}

	return nil
}
func GetProfile(profileId string, region string, ctx logging.Context) (vault.AzureProfile, error) {
	data, err := vault.GetCredentialProfile("azure", profileId, ctx)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return vault.AzureProfile{}, err
	}
	azureProfile := vault.AzureProfile{}
	err = json.Unmarshal(data, &azureProfile)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return vault.AzureProfile{}, err
	}
	azureProfile.Profile.Location = region
	return azureProfile, nil

}
func checkClusterSize(cluster Cluster_Def) error {
	for _, pools := range cluster.NodePools {
		if pools.NodeCount > 3 {
			return errors.New("Nodepool can't have more than 3 nodes")
		}
	}
	return nil
}
func CreateCluster(cluster Cluster_Def, ctx logging.Context) error {

	_, err := GetCluster(cluster.ProjectId, ctx)
	if err == nil { //cluster found
		text := fmt.Sprintf("Cluster model: Create - Cluster for project'%s' already exists in the database: ", cluster.Name)
		ctx.SendSDLog(text+err.Error(), "error")
		return errors.New(text)
	}
	session, err := db.GetMongoSession()
	if err != nil {
		ctx.SendSDLog("Cluster model: Delete - Got error while connecting to the database: "+err.Error(), "error")
		return err
	}
	defer session.Close()

	err = checkClusterSize(cluster)
	if err != nil { //cluster found
		ctx.SendSDLog(err.Error(), "error")
		return err
	}
	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoAzureClusterCollection, cluster)
	if err != nil {
		ctx.SendSDLog("Cluster model: Create - Got error inserting cluster to the database: "+err.Error(), "error")
		return err
	}

	return nil
}

func GetCluster(projectId string, ctx logging.Context) (cluster Cluster_Def, err error) {

	session, err1 := db.GetMongoSession()
	if err1 != nil {
		ctx.SendSDLog("Cluster model: Get - Got error while connecting to the database: "+err1.Error(), "error")
		return Cluster_Def{}, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAzureClusterCollection)
	err = c.Find(bson.M{"project_id": projectId}).One(&cluster)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return Cluster_Def{}, err
	}

	return cluster, nil
}

func GetAllCluster(ctx logging.Context) (clusters []Cluster_Def, err error) {
	session, err1 := db.GetMongoSession()
	if err1 != nil {
		ctx.SendSDLog("Cluster model: GetAll - Got error while connecting to the database: "+err1.Error(), "error")
		return nil, err1
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAzureClusterCollection)
	err = c.Find(bson.M{}).All(&clusters)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return nil, err
	}

	return clusters, nil
}

func UpdateCluster(cluster Cluster_Def, update bool, ctx logging.Context) error {
	oldCluster, err := GetCluster(cluster.ProjectId, ctx)
	if err != nil {
		text := fmt.Sprintf("Cluster model: Update - Cluster '%s' does not exist in the database: ", cluster.Name)
		ctx.SendSDLog(text, "error")
		return errors.New(text)
	}
	if oldCluster.Status == "Cluster Created" && update {
		ctx.SendSDLog("Cluster is in runnning state", "error")
		return errors.New("Cluster is in runnning state")
	}
	err = DeleteCluster(cluster.ProjectId, ctx)
	if err != nil {
		ctx.SendSDLog("Cluster model: Update - Got error deleting cluster: "+err.Error(), "error")
		return err
	}

	cluster.CreationDate = oldCluster.CreationDate
	cluster.ModificationDate = time.Now()

	err = CreateCluster(cluster, ctx)
	if err != nil {
		ctx.SendSDLog("Cluster model: Update - Got error creating cluster: "+err.Error(), "error")
		return err
	}

	return nil
}

func DeleteCluster(projectId string, ctx logging.Context) error {
	session, err := db.GetMongoSession()
	if err != nil {
		ctx.SendSDLog("Cluster model: Delete - Got error while connecting to the database: "+err.Error(), "error")
		return err
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAzureClusterCollection)
	err = c.Remove(bson.M{"project_id": projectId})
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return err
	}

	return nil
}
func PrintError(confError error, name, projectId string, ctx logging.Context) {
	if confError != nil {
		ctx.SendSDLog(confError.Error(), "error")
		logging.SendLog("Cluster creation failed : "+name, "error", projectId)
		logging.SendLog(confError.Error(), "error", projectId)
	}
}
func DeployCluster(cluster Cluster_Def, credentials vault.AzureProfile, ctx logging.Context) (confError error) {

	azure := AZURE{
		ID:           credentials.Profile.ClientId,
		Key:          credentials.Profile.ClientSecret,
		Tenant:       credentials.Profile.TenantId,
		Subscription: credentials.Profile.SubscriptionId,
		Region:       credentials.Profile.Location,
	}
	confError = azure.init()
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx)
		return confError
	}

	publisher := utils.Notifier{}
	confError = publisher.Init_notifier()
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx)
		return confError
	}

	logging.SendLog("Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	cluster, confError = azure.createCluster(cluster, ctx)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx)
		beego.Info("going to cleanup")
		confError = azure.CleanUp(cluster, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, ctx)
		}

		cluster.Status = "Cluster creation failed"
		confError = UpdateCluster(cluster, false, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, ctx)
		}
		publisher.Notify(cluster.ProjectId, "Status Available")
		return nil

	}
	cluster.Status = "Cluster Created"

	confError = UpdateCluster(cluster, false, ctx)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx)
		publisher.Notify(cluster.ProjectId, "Status Available")
		return confError
	}
	logging.SendLog("Cluster created successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available")

	return nil
}
func FetchStatus(credentials vault.AzureProfile, projectId string, ctx logging.Context) (Cluster_Def, error) {

	cluster, err := GetCluster(projectId, ctx)
	if err != nil {
		ctx.SendSDLog("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), "error")
		return Cluster_Def{}, err
	}

	azure := AZURE{
		ID:           credentials.Profile.ClientId,
		Key:          credentials.Profile.ClientSecret,
		Tenant:       credentials.Profile.TenantId,
		Subscription: credentials.Profile.SubscriptionId,
		Region:       credentials.Profile.Location,
	}
	err = azure.init()
	if err != nil {
		return Cluster_Def{}, err
	}

	c, e := azure.fetchStatus(cluster, ctx)
	if e != nil {
		ctx.SendSDLog("Cluster model: Status - Failed to get lastest status "+e.Error(), "error")
		return Cluster_Def{}, e
	}
	/*err = UpdateCluster(c)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
		return Cluster_Def{}, err
	}*/
	return c, nil
}
func TerminateCluster(cluster Cluster_Def, credentials vault.AzureProfile, ctx logging.Context) error {

	publisher := utils.Notifier{}
	pub_err := publisher.Init_notifier()
	if pub_err != nil {
		ctx.SendSDLog(pub_err.Error(), "error")
		return pub_err
	}

	cluster, err := GetCluster(cluster.ProjectId, ctx)
	if err != nil {
		ctx.SendSDLog("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), "error")
		return err
	}
	if cluster.Status != "Cluster Created" {
		ctx.SendSDLog("Cluster model: Cluster is not in created state ", "error")
		publisher.Notify(cluster.ProjectId, "Status Available")
		return err
	}

	azure := AZURE{
		ID:           credentials.Profile.ClientId,
		Key:          credentials.Profile.ClientSecret,
		Tenant:       credentials.Profile.TenantId,
		Subscription: credentials.Profile.SubscriptionId,
		Region:       credentials.Profile.Location,
	}
	err = azure.init()
	if err != nil {
		return err
	}

	err = azure.terminateCluster(cluster, ctx)

	if err != nil {

		ctx.SendSDLog(err.Error(), "error")

		logging.SendLog("Cluster termination failed: "+cluster.Name, "error", cluster.ProjectId)
		logging.SendLog(err.Error(), "error", cluster.ProjectId)

		cluster.Status = "Cluster Termination Failed"
		err = UpdateCluster(cluster, false, ctx)
		if err != nil {
			ctx.SendSDLog("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), "error")
			logging.SendLog("Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			logging.SendLog(err.Error(), "error", cluster.ProjectId)
			publisher.Notify(cluster.ProjectId, "Status Available")
			return err
		}
		publisher.Notify(cluster.ProjectId, "Status Available")
		return nil
	}

	cluster.Status = "Cluster Terminated"

	for _, pools := range cluster.NodePools {
		var nodes []*VM
		pools.Nodes = nodes
	}
	err = UpdateCluster(cluster, false, ctx)
	if err != nil {
		ctx.SendSDLog("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), "error")
		logging.SendLog("Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
		logging.SendLog(err.Error(), "error", cluster.ProjectId)
		publisher.Notify(cluster.ProjectId, "Status Available")
		return err
	}
	logging.SendLog("Cluster terminated successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available")

	return nil
}
func InsertSSHKeyPair(key utils.Key) (err error) {
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
func GetAllSSHKeyPair(ctx logging.Context) (keys []string, err error) {

	keys, err = vault.GetAllSSHKey("azure", ctx)
	if err != nil {
		beego.Error(err.Error())
		return keys, err
	}
	return keys, nil
}
func GetSSHKeyPair(keyname string) (keys *utils.Key, err error) {

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
