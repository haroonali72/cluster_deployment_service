package azure

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/db"
	"antelope/models/key_utils"
	"antelope/models/rbac_authentication"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"antelope/models/woodpecker"
	"encoding/json"
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
	ProjectId        string        `json:"project_id" bson:"project_id" valid:"required"`
	Name             string        `json:"name" bson:"name" valid:"required"`
	Status           string        `json:"status" bson:"status" valid:"in(NEW|new|New)"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud" valid:"in(AZURE|azure)"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools        []*NodePool   `json:"node_pools" bson:"node_pools" valid:"required"`
	NetworkName      string        `json:"network_name" bson:"network_name" valid:"required"`
	ResourceGroup    string        `json:"resource_group" bson:"resource_group" valid:"required"`
	CompanyId        string        `json:"company_id" bson:"company_id"`
	TokenName        string        `json:"token_name" bson:"token_name"`
}

type NodePool struct {
	ID                 bson.ObjectId      `json:"_id" bson:"_id,omitempty"`
	Name               string             `json:"name" bson:"name" valid:"required"`
	NodeCount          int64              `json:"node_count" bson:"node_count" valid:"required,matches(^[0-9]+$)"`
	MachineType        string             `json:"machine_type" bson:"machine_type" valid:"required"`
	Image              ImageReference     `json:"image" bson:"image" valid:"required"`
	Volume             Volume             `json:"volume" bson:"volume"`
	EnableVolume       bool               `json:"is_external" bson:"is_external"`
	PoolSubnet         string             `json:"subnet_id" bson:"subnet_id" valid:"required"`
	PoolSecurityGroups []*string          `json:"security_group_id" bson:"security_group_id" valid:"required"`
	Nodes              []*VM              `json:"nodes" bson:"nodes"`
	PoolRole           models.PoolRole    `json:"pool_role" bson:"pool_role"`
	AdminUser          string             `json:"user_name" bson:"user_name,omitempty"`
	KeyInfo            key_utils.AZUREKey `json:"key_info" bson:"key_info"`
	BootDiagnostics    DiagnosticsProfile `json:"boot_diagnostics" bson:"boot_diagnostics"`
	OsDisk             models.OsDiskType  `json:"os_disk_type" bson:"os_disk_type" valid:"required, in(standard hdd|standard ssd|premium ssd)"`
	EnableScaling      bool               `json:"enable_scaling" bson:"enable_scaling"`
	Scaling            AutoScaling        `json:"auto_scaling" bson:"auto_scaling"`
	EnablePublicIP     bool               `json:"enable_public_ip" bson:"enable_public_ip"`
}
type AutoScaling struct {
	MaxScalingGroupSize int64       `json:"max_scaling_group_size" bson:"max_scaling_group_size"`
	State               models.Type `json:"status" bson:"status"`
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
	DataDisk models.OsDiskType `json:"disk_type" bson:"disk_type"`
	Size     int32             `json:"disk_size" bson:"disk_size"`
}
type VM struct {
	CloudId             *string `json:"cloud_id" bson:"cloud_id,omitempty"`
	NodeState           *string `json:"node_state" bson:"node_state,omitempty"`
	Name                *string `json:"name" bson:"name,omitempty"`
	PrivateIP           *string `json:"private_ip" bson:"private_ip,omitempty"`
	PublicIP            *string `json:"public_ip" bson:"public_ip,omitempty"`
	UserName            *string `json:"user_name" bson:"user_name,omitempty"`
	PAssword            *string `json:"password" bson:"password,omitempty"`
	ComputerName        *string `json:"computer_name" bson:"computer_name,omitempty"`
	IdentityPrincipalId *string `json:"identity_principal_id" bson:"identity_principal_id"`
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
func GetRegion(token, projectId string, ctx utils.Context) (string, error) {
	url := beego.AppConfig.String("raccoon_url") + models.ProjectGetEndpoint
	if strings.Contains(url, "{projectId}") {
		url = strings.Replace(url, "{projectId}", projectId, -1)
	}
	data, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}
	var region Project
	err = json.Unmarshal(data.([]byte), &region)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return region.ProjectData.Region, err
	}
	return region.ProjectData.Region, nil

}
func GetNetwork(projectId string, ctx utils.Context, resourceGroup string, token string) (types.AzureNetwork, error) {

	url := getNetworkHost("azure", projectId)

	data, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.AzureNetwork{}, err
	}

	var network types.AzureNetwork
	err = json.Unmarshal(data.([]byte), &network)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.AzureNetwork{}, err
	}

	if network.Definition != nil {
		if network.Definition[0].ResourceGroup != resourceGroup {
			ctx.SendLogs("Resource group is incorrect", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return types.AzureNetwork{},errors.New("Resource Group is in correct")
		}
	} else {
		return types.AzureNetwork{}, errors.New("Network not found")
	}
	return network, nil
}
func GetProfile(profileId string, region string, token string, ctx utils.Context) (vault.AzureProfile, error) {
	data, err := vault.GetCredentialProfile("azure", profileId, token, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return vault.AzureProfile{}, err
	}
	azureProfile := vault.AzureProfile{}
	err = json.Unmarshal(data, &azureProfile)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
func CreateCluster(cluster Cluster_Def, ctx utils.Context) error {

	_, err := GetCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err == nil { //cluster found
		text := fmt.Sprintf("Cluster model: Create - Cluster for project'%s' already exists in the database: ", cluster.Name)
		ctx.SendLogs(text+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}
	err = checkMasterPools(cluster)
	if err != nil { //cluster found
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Delete - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer session.Close()

	//err = checkClusterSize(cluster)
	//if err != nil { //cluster found
	//	ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	//	return err
	//}
	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoAzureClusterCollection, cluster)
	if err != nil {

		ctx.SendLogs("Cluster model: Create - Got error inserting cluster to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}

func GetCluster(projectId, companyId string, ctx utils.Context) (cluster Cluster_Def, err error) {

	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Create - Got error inserting cluster to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAzureClusterCollection)
	err = c.Find(bson.M{"project_id": projectId, "company_id": companyId}).One(&cluster)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, err
	}
	return cluster, nil
}

func GetAllCluster(ctx utils.Context, list rbac_athentication.List) (clusters []Cluster_Def, err error) {
	var copyData []string
	for _, d := range list.Data {
		copyData = append(copyData, d)
	}
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Cluster model: GetAll - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err1
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAzureClusterCollection)
	err = c.Find(bson.M{"project_id": bson.M{"$in": copyData}}).All(&clusters)
	if err != nil {

		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	return clusters, nil
}

func UpdateCluster(cluster Cluster_Def, update bool, ctx utils.Context) error {
	oldCluster, err := GetCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err != nil {
		text := fmt.Sprintf("Cluster model: Update - Cluster '%s' does not exist in the database: ", cluster.Name)
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
		ctx.SendLogs("Cluster model: Update - Got error deleting cluster: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	cluster.CreationDate = oldCluster.CreationDate
	cluster.ModificationDate = time.Now()

	err = CreateCluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Update - Got error creating cluster: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}

func DeleteCluster(projectId, companyId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Delete - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAzureClusterCollection)
	err = c.Remove(bson.M{"project_id": projectId, "company_id": companyId})
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func PrintError(confError error, name, projectId string, ctx utils.Context, companyId string) {
	if confError != nil {
		ctx.SendLogs(confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(companyId, "Cluster creation failed : "+name, "error", projectId)
		utils.SendLog(companyId, confError.Error(), "error", projectId)
	}
}
func DeployCluster(cluster Cluster_Def, credentials vault.AzureProfile, ctx utils.Context, companyId string, token string) (confError error) {

	publisher := utils.Notifier{}
	confError = publisher.Init_notifier()
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		return confError
	}

	azure := AZURE{
		ID:           credentials.Profile.ClientId,
		Key:          credentials.Profile.ClientSecret,
		Tenant:       credentials.Profile.TenantId,
		Subscription: credentials.Profile.SubscriptionId,
		Region:       credentials.Profile.Location,
	}
	confError = azure.init()
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)

		cluster.Status = "Cluster creation failed"
		confError = UpdateCluster(cluster, false, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return confError
	}

	utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	cluster, confError = azure.createCluster(cluster, ctx, companyId, token)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		cluster.Status = "Cluster creation failed"
		beego.Info("going to cleanup")
		confError = azure.CleanUp(cluster, ctx, companyId)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		}

		cluster.Status = "Cluster creation failed"
		confError = UpdateCluster(cluster, false, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return nil

	}
	cluster.Status = "Cluster Created"

	confError = UpdateCluster(cluster, false, ctx)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return confError
	}
	utils.SendLog(companyId, "Cluster created successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)

	return nil
}
func FetchStatus(credentials vault.AzureProfile, token, projectId string, companyId string, ctx utils.Context) (Cluster_Def, error) {

	cluster, err := GetCluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	_, e := azure.fetchStatus(&cluster, token, ctx)
	if e != nil {

		ctx.SendLogs("Cluster model: Status - Failed to get lastest status "+e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return cluster, e
	}
	/*err = UpdateCluster(c)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
		return Cluster_Def{}, err
	}*/
	return cluster, nil
}
func TerminateCluster(cluster Cluster_Def, credentials vault.AzureProfile, ctx utils.Context, companyId string) error {

	publisher := utils.Notifier{}
	pub_err := publisher.Init_notifier()
	if pub_err != nil {
		ctx.SendLogs(pub_err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return pub_err
	}

	cluster, err := GetCluster(cluster.ProjectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	if cluster.Status == "" || cluster.Status == "new" {
		text := "Cannot terminate a new cluster"
		ctx.SendLogs("AzureClusterModel : "+text+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return errors.New(text)
	}

	azure := AZURE{
		ID:           credentials.Profile.ClientId,
		Key:          credentials.Profile.ClientSecret,
		Tenant:       credentials.Profile.TenantId,
		Subscription: credentials.Profile.SubscriptionId,
		Region:       credentials.Profile.Location,
	}

	cluster.Status = string(models.Terminating)
	utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	err = azure.init()
	if err != nil {
		utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
		cluster.Status = "Cluster Termination Failed"
		err = UpdateCluster(cluster, false, ctx)
		if err != nil {
			ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)

			return err
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}

	err = azure.terminateCluster(cluster, ctx, companyId)

	if err != nil {

		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		utils.SendLog(companyId, "Cluster termination failed: "+cluster.Name, "error", cluster.ProjectId)
		utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)

		cluster.Status = "Cluster Termination Failed"
		err = UpdateCluster(cluster, false, ctx)
		if err != nil {
			ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

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
		var nodes []*VM
		pools.Nodes = nodes
	}
	err = UpdateCluster(cluster, false, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
		utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}
	utils.SendLog(companyId, "Cluster terminated successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)

	return nil
}
func InsertSSHKeyPair(key key_utils.AZUREKey) (err error) {
	key.Cloud = models.Azure
	ctx := new(utils.Context)
	session, err := db.GetMongoSession(*ctx)
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
func GetAllSSHKeyPair(ctx utils.Context, token string) (keys interface{}, err error) {

	keys, err = vault.GetAllSSHKey("azure", ctx, token, "")
	if err != nil {
		beego.Error(err.Error())
		return keys, err
	}
	return keys, nil
}
func GetSSHKeyPair(keyname string) (keys *key_utils.AZUREKey, err error) {

	ctx := new(utils.Context)
	session, err := db.GetMongoSession(*ctx)
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

func CreateSSHkey(keyName, token, teams string, ctx utils.Context) (privateKey string, err error) {

	keyInfo, err := key_utils.GenerateKey(models.Azure, keyName, "azure@example.com", token, teams, ctx)
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

	err := vault.DeleteSSHkey(string(models.Azure), keyName, token, ctx, "")
	if err != nil {
		return err
	}

	return err
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

func GetInstances(credentials vault.AzureProfile, ctx utils.Context) ([]azureVM, error) {

	azure := AZURE{
		ID:           credentials.Profile.ClientId,
		Key:          credentials.Profile.ClientSecret,
		Tenant:       credentials.Profile.TenantId,
		Subscription: credentials.Profile.SubscriptionId,
		Region:       credentials.Profile.Location,
	}
	err := azure.init()
	if err != nil {
		return []azureVM{}, err
	}

	instances, err := azure.getAllInstances()
	if err != nil {
		beego.Error(err.Error())
		return []azureVM{}, err
	}
	return instances, nil
}
func GetRegions(credentials vault.AzureProfile, ctx utils.Context) ([]models.Region, error) {

	azure := AZURE{
		ID:           credentials.Profile.ClientId,
		Key:          credentials.Profile.ClientSecret,
		Tenant:       credentials.Profile.TenantId,
		Subscription: credentials.Profile.SubscriptionId,
		Region:       credentials.Profile.Location,
	}
	err := azure.init()
	if err != nil {
		return []models.Region{}, err
	}

	regions, err := azure.getRegions(ctx)
	if err != nil {
		beego.Error(err.Error())
		return []models.Region{}, err
	}
	return regions, nil
}
func GetAllMachines() ([]string, error) {

	regions, err := getAllVMSizes()
	if err != nil {
		beego.Error(err.Error())
		return []string{}, err
	}
	return regions, nil
}
func ValidateProfile(clientId, clientSecret, subscriptionId, tenantId, region string, ctx utils.Context) error {

	azure := AZURE{
		ID:           clientId,
		Key:          clientSecret,
		Tenant:       tenantId,
		Subscription: subscriptionId,
		Region:       region,
	}
	err := azure.init()
	if err != nil {
		return err
	}

	_, err = azure.getRegions(ctx)
	if err != nil {
		beego.Error("Profile is not valid")
		return err
	}
	return nil
}
func ApplyAgent(credentials vault.AzureProfile, companyId, token string, ctx utils.Context, projetcID, clusterName, resourceGroup string) (confError error) {

	data2, err := woodpecker.GetCertificate(projetcID, token, ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterModel : Apply Agent -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	filePath := "/tmp/" + companyId + "/" + projetcID + "/"
	cmd := "mkdir -p " + filePath + " && echo '" + data2 + "'>" + filePath + "agent.yaml"
	output, err := models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("AKSClusterModel : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	cmd = "sudo docker run --rm --name " + companyId + projetcID + " -e resourceGroup=" + resourceGroup + " -e cluster=" + clusterName + " -e clientID=" + credentials.Profile.ClientId + " -e tenant=" + credentials.Profile.TenantId + " -e clientSecret=" + credentials.Profile.ClientSecret + " -e yamlFile=" + filePath + "agent.yaml -v " + filePath + ":" + filePath + " " + models.AKSAuthContainerName

	output, err = models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("AKSClusterModel : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
