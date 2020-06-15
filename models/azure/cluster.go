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
	ID               bson.ObjectId `json:"-" bson:"_id,omitempty"`
	ProjectId        string        `json:"project_id" bson:"project_id" valid:"required" description:"Id of project [required]"`
	Name             string        `json:"name" bson:"name" valid:"required" description:"Unique name of the cluster [required]"`
	Status           models.Type   `json:"status" bson:"status" validate:"eq=new|eq=New|eq=NEW|eq=Cluster Creation Failed|eq=Cluster Terminated|eq=Cluster Created" description:"Status of the cluster [optional]"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud" validate:"eq=AZURE|eq=azure|eq=Azure" description:"Name of the cloud [optional]"`
	CreationDate     time.Time     `json:"creation_date" bson:"creation_date"`
	ModificationDate time.Time     `json:"modification_date" bson:"modification_date"`
	NodePools        []*NodePool   `json:"node_pools" bson:"node_pools" validate:"required,dive" description:"Nodepools of the cluster.Atleast 1 nodepool [required]"`
	NetworkName      string        `json:"network_name" bson:"network_name" valid:"required" description:"Network name to deploy the cluster [required]"`
	ResourceGroup    string        `json:"resource_group" bson:"resource_group" valid:"required" description:"Resource group to deploy the cluster [required]"`
	CompanyId        string        `json:"company_id" bson:"company_id"  description:"Id of the company [optional]"`
	TokenName        string        `json:"token_name" bson:"token_name"`
}

type NodePool struct {
	ID                 bson.ObjectId      `json:"-" bson:"_id,omitempty"`
	Name               string             `json:"name" bson:"name" valid:"required" description:"Unique name of the nodepool.[required]"`
	NodeCount          int64              `json:"node_count" bson:"node_count" valid:"required,matches(^[1-9]+$)" description:"Count of the nodepool. Atleast 1 [required]"`
	MachineType        string             `json:"machine_type" bson:"machine_type" valid:"required" description:"Machine type of the nodepool.[required]"`
	Image              ImageReference     `json:"image" bson:"image" valid:"required,dive" description:"VM image of the nodepool.[required]"`
	Volume             Volume             `json:"volume" bson:"volume" description:"Volume to attach with the nodepool.[required]"`
	EnableVolume       bool               `json:"is_external" bson:"is_external"  valid:"required" description:"Enable if volume is external [required]"`
	PoolSubnet         string             `json:"subnet_id" bson:"subnet_id" description:"Subnet to deploy the nodepool of the cluster.[required]"`
	PoolSecurityGroups []*string          `json:"security_group_id" bson:"security_group_id" description:"Security group to attach with the nodepool.[required]"`
	Nodes              []*VM              `json:"nodes" bson:"nodes" valid:"required" description:"Nodes in the nodepool.Atleast 1 [required]"`
	PoolRole           models.PoolRole    `json:"pool_role" bson:"pool_role" valid:"required,eq=master|eq=slave" description:"Role of the nodepool.Valid values are 'master' and 'slave'[required]"`
	AdminUser          string             `json:"user_name" bson:"user_name"  valid:"required" description:"User of the nodepool.[optional]"`
	KeyInfo            key_utils.AZUREKey `json:"key_info" bson:"key_info" valid:"required" description:"SSH key details.[requiredl]"`
	BootDiagnostics    DiagnosticsProfile `json:"boot_diagnostics" bson:"boot_diagnostics" description:"Storage account details.[optional]"`
	OsDisk             models.OsDiskType  `json:"os_disk_type" bson:"os_disk_type" valid:"required,eq=standard hdd|standard ssd|premium ssd" description:"Type of the OS disk.[requiredl]`
	EnableScaling      bool               `json:"enable_scaling" bson:"enable_scaling"  valid:"required" description:"For enabling scaling [required]"`
	Scaling            AutoScaling        `json:"auto_scaling" bson:"auto_scaling"  valid:"required"  description:"Details of auto scaling [required]"`
	EnablePublicIP     bool               `json:"enable_public_ip" bson:"enable_public_ip"  valid:"required" description:"Enable to assign public Ip to the nodepool [required]"`
}
type AutoScaling struct {
	MaxScalingGroupSize int64       `json:"max_scaling_group_size" bson:"max_scaling_group_size" valid:"required" description:"Max count for scaling [required]"`
	State               models.Type `json:"status" bson:"status" description:"Status of scaling [required]"`
}
type Key struct {
	CredentialType models.CredentialsType `json:"credential_type"  bson:"credential_type" valid:"required,eq=password|eq=key" description:"Credentials type to connect to the VM.Valid values are 'key' and 'password' [required]"`
	NewKey         models.KeyType         `json:"key_type"  bson:"key_type" valid:"required in(new|cp|azure|user")" description:"Type of key to use.Valid values are 'new','cp','azure' and 'user' [required]"`
	KeyName        string                 `json:"key_name" bson:"key_name" valid:"required" description:"Unique name of the key [required]"`
	AdminPassword  string                 `json:"admin_password" bson:"admin_password",omitempty" description:"Password to log in [required]"`
	PrivateKey     string                 `json:"private_key" bson:"private_key",omitempty" description:"Private SSH key [required]"`
	PublicKey      string                 `json:"public_key" bson:"public_key",omitempty" description:"Public SSH key [required]"`
	Cloud          models.Cloud           `json:"cloud" bson:"cloud" validate:"eq=AZURE|eq=azure|eq=Azure" description:"Name of the cloud [optional]"`
}
type Volume struct {
	DataDisk models.OsDiskType `json:"disk_type" bson:"disk_type" valid:"required,eq=standard hdd|standard ssd|premium ssd" description:"Type of the OS disk.[required]"`
	Size     int32             `json:"disk_size" bson:"disk_size" valid:"required" description:"Size of the disk.[required]"`
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
	Enable            bool   `json:"enable" bson:"enable" valid:"required" description:"To enable diagnostics profile [required]`
	NewStroageAccount bool   `json:"new_storage_account" bson:"new_storage_account" valid:"required" description:"Enable to make new storage account[required]`
	StorageAccountId  string `json:"storage_account_id" bson:"storage_account_id" description:"Id of the storage account [optional]`
}

type ImageReference struct {
	ID        bson.ObjectId `json:"-" bson:"_id,omitempty"`
	Publisher string        `json:"publisher" bson:"publisher" valid:"required" description:"Publisher of the VM image [required]"`
	Offer     string        `json:"offer" bson:"offer,omitempty" valid:"required" description:"Offer of the VM image [required]"`
	Sku       string        `json:"sku" bson:"sku,omitempty" valid:"required" description:"Sku of the Vm image [required]"`
	Version   string        `json:"version" bson:"version,omitempty" valid:"required" description:"Version of the Vm image [required]"`
	ImageId   string        `json:"image_id" bson:"image_id,omitempty"`
}
type Project struct {
	ProjectData Data `json:"data"`
}

type Data struct {
	Region string `json:"region"`
}

type AzureCluster struct {
	Name      string      `json:"name,omitempty" bson:"name,omitempty" v description:"Cluster name"`
	ProjectId string      `json:"project_id" bson:"project_id"  description:"ID of project"`
	Status    models.Type `json:"status,omitempty" bson:"status,omitempty" " description:"Status of cluster"`
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
		ctx.SendLogs("Error in fetching region"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}
	var region Project
	err = json.Unmarshal(data.([]byte), &region.ProjectData)
	if err != nil {
		ctx.SendLogs("Error in fetching region"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
			return types.AzureNetwork{}, errors.New("Resource Group is incorrect")
		}
	} else {
		return types.AzureNetwork{}, errors.New("Network not found")
	}
	return network, nil
}
func GetProfile(profileId string, region string, token string, ctx utils.Context) (int, vault.AzureProfile, error) {
	statusCode, data, err := vault.GetCredentialProfile("azure", profileId, token, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return statusCode, vault.AzureProfile{}, err
	}
	azureProfile := vault.AzureProfile{}
	err = json.Unmarshal(data, &azureProfile)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 500, vault.AzureProfile{}, err
	}
	azureProfile.Profile.Location = region
	return 0, azureProfile, nil

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
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

func GetAllCluster(ctx utils.Context, list rbac_athentication.List) (azurecluster []AzureCluster, err error) {
	var copyData []string
	var clusters []Cluster_Def
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
	err = c.Find(bson.M{"project_id": bson.M{"$in": copyData}, "company_id": ctx.Data.Company}).All(&clusters)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	for _, cluster := range clusters {
		temp := AzureCluster{Name: cluster.Name, ProjectId: cluster.ProjectId, Status: cluster.Status}
		azurecluster = append(azurecluster, temp)
	}
	return azurecluster, nil
}

func UpdateCluster(cluster Cluster_Def, update bool, ctx utils.Context) error {
	oldCluster, err := GetCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err != nil {
		text := fmt.Sprintf("Cluster model: Update - Cluster '%s' does not exist in the database: ", cluster.Name)
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	if oldCluster.Status == models.Deploying && update {
		ctx.SendLogs("Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("Cluster is in creating state")
	} else if oldCluster.Status == models.Terminating && update {
		ctx.SendLogs("Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("Cluster is in terminating state")
	} else if oldCluster.Status == models.ClusterTerminationFailed && update {
		ctx.SendLogs("Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("Cluster is in termination failed state")
	} else if oldCluster.Status == models.ClusterCreated && update {
		if !checkScalingChanges(&oldCluster, &cluster) {
			ctx.SendLogs("Cluster is in created state ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return errors.New("Cluster is in created state")
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
	cluster.CompanyId = oldCluster.CompanyId

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
func DeployCluster(cluster Cluster_Def, credentials vault.AzureProfile, ctx utils.Context, companyId string, token string) types.CustomCPError {

	publisher := utils.Notifier{}
	confError := publisher.Init_notifier()
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		customError := ApiError(confError, "Error in cluster creation", int(models.CloudStatusCode))
		err := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.Azure, ctx, customError)
		if err != nil {
			ctx.SendLogs("AzureClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return customError
	}

	azure := AZURE{
		ID:           credentials.Profile.ClientId,
		Key:          credentials.Profile.ClientSecret,
		Tenant:       credentials.Profile.TenantId,
		Subscription: credentials.Profile.SubscriptionId,
		Region:       credentials.Profile.Location,
	}
	err := azure.init()
	if err != (types.CustomCPError{}) {
		PrintError(errors.New(err.Error), cluster.Name, cluster.ProjectId, ctx, companyId)
		PrintError(errors.New(err.Description), cluster.Name, cluster.ProjectId, ctx, companyId)
		cluster.Status = models.ClusterCreationFailed
		confError = UpdateCluster(cluster, false, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		}

		err1 := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.Azure, ctx, err)
		if err1 != nil {
			ctx.SendLogs("AzureClusterModel:  Deploy - "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}


	utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)

	pubSub:= publisher.Subscribe(ctx.Data.ProjectId ,ctx)

	cluster, err = azure.createCluster(cluster, ctx, companyId, token)
	if err != (types.CustomCPError{}) {
		PrintError(errors.New(err.Error), cluster.Name, cluster.ProjectId, ctx, companyId)
		PrintError(errors.New(err.Description), cluster.Name, cluster.ProjectId, ctx, companyId)
		cluster.Status = models.ClusterCreationFailed
		beego.Info("going to cleanup")
		err = azure.CleanUp(cluster, ctx, companyId)
		if err != (types.CustomCPError{}) {
			PrintError(errors.New(err.Error), cluster.Name, cluster.ProjectId, ctx, companyId)
		}

		cluster.Status = models.ClusterCreationFailed
		confError = UpdateCluster(cluster, false, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		}
		err1 := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.Azure, ctx, err)
		if err1 != nil {
			ctx.SendLogs("AzureClusterModel:  Deploy - "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}

	cluster.Status = models.ClusterCreated

	confError = UpdateCluster(cluster, false, ctx)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		customError := ApiError(confError, "Error in cluster creation", int(models.CloudStatusCode))
		err1 := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.Azure, ctx, customError)
		if err1 != nil {
			ctx.SendLogs("AzureClusterModel:  Deploy - "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return customError
	}

	utils.SendLog(companyId, "Cluster created successfully "+cluster.Name, "info", cluster.ProjectId)

	notify:= publisher.RecieveNotification(ctx.Data.ProjectId,ctx,pubSub)
	if notify{
		ctx.SendLogs("AzureClusterModel:  Notification recieved from agent", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
	}else{
		ctx.SendLogs("AzureClusterModel:  Notification not recieved from agent", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	}

	return types.CustomCPError{}
}
func FetchStatus(credentials vault.AzureProfile, token, projectId string, companyId string, ctx utils.Context) (Cluster_Def, types.CustomCPError) {

	cluster, err := GetCluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, ApiError(err, "Error in fetching status.", int(models.CloudStatusCode))
	}

	azure := AZURE{
		ID:           credentials.Profile.ClientId,
		Key:          credentials.Profile.ClientSecret,
		Tenant:       credentials.Profile.TenantId,
		Subscription: credentials.Profile.SubscriptionId,
		Region:       credentials.Profile.Location,
	}
	err1 := azure.init()
	if err1 != (types.CustomCPError{}) {
		return Cluster_Def{}, err1
	}

	_, e := azure.fetchStatus(&cluster, token, ctx)
	if e != (types.CustomCPError{}) {
		ctx.SendLogs("Cluster model: Status - Failed to get lastest status "+e.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, e
	}
	/*err = UpdateCluster(c)
	if err != nil {
		beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
		return Cluster_Def{}, err
	}*/
	return cluster, types.CustomCPError{}
}
func TerminateCluster(cluster Cluster_Def, credentials vault.AzureProfile, ctx utils.Context, companyId string) types.CustomCPError {

	publisher := utils.Notifier{}
	pub_err := publisher.Init_notifier()
	if pub_err != nil {
		ctx.SendLogs(pub_err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		customError := ApiError(pub_err, "Error in cluster termination", int(models.CloudStatusCode))
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.Azure, ctx, customError)
		if err != nil {
			ctx.SendLogs("AzureClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return customError
	}

	cluster, err := GetCluster(cluster.ProjectId, companyId, ctx)
	if err != nil {
		customError := ApiError(pub_err, "Error in cluster termination", int(models.CloudStatusCode))
		err1 := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.Azure, ctx, customError)
		if err1 != nil {
			ctx.SendLogs("AzureClusterModel:  Terminate Cluster - "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return customError
	}

	if cluster.Status == "" || cluster.Status == "new" {
		text := "Cannot terminate a new cluster"
		ctx.SendLogs("AzureClusterModel : "+text+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return ApiError(errors.New("Error in cluster termination"), text, int(models.CloudStatusCode))
	}

	azure := AZURE{
		ID:           credentials.Profile.ClientId,
		Key:          credentials.Profile.ClientSecret,
		Tenant:       credentials.Profile.TenantId,
		Subscription: credentials.Profile.SubscriptionId,
		Region:       credentials.Profile.Location,
	}

	cluster.Status = models.Terminating
	utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	err1 := azure.init()
	if err1 != (types.CustomCPError{}) {
		utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
		cluster.Status = models.ClusterTerminationFailed
		err = UpdateCluster(cluster, false, ctx)
		if err != nil {
			ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
			return err1
		}
		err2 := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.Azure, ctx, err1)
		if err2 != nil {
			ctx.SendLogs("AzureClusterModel:  Terminate Cluster - "+err2.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err1
	}

	err1 = azure.terminateCluster(cluster, ctx, companyId)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs(err1.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(companyId, "Cluster termination failed: "+cluster.Name, "error", cluster.ProjectId)
		utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)

		cluster.Status = models.ClusterTerminationFailed
		err = UpdateCluster(cluster, false, ctx)
		if err != nil {
			ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
			publisher.Notify(cluster.ProjectId, "Status Available", ctx)
			return ApiError(err, "Error in cluster termination", int(models.CloudStatusCode))
		}
		err2 := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.Azure, ctx, err1)
		if err2 != nil {
			ctx.SendLogs("AzureClusterModel:  Terminate Cluster - "+err2.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)

		return types.CustomCPError{}
	}

	cluster.Status = models.ClusterTerminated

	for _, pools := range cluster.NodePools {
		var nodes []*VM
		pools.Nodes = nodes
	}
	err = UpdateCluster(cluster, false, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
		utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
		customError := ApiError(err, "Error in cluster deletion", int(models.CloudStatusCode))
		err1 := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.Azure, ctx, customError)
		if err1 != nil {
			ctx.SendLogs("AzureClusterModel:  Terminate Cluster - "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return customError
	}
	utils.SendLog(companyId, "Cluster terminated successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)

	return types.CustomCPError{}
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

func GetInstances(credentials vault.AzureProfile, ctx utils.Context) ([]azureVM, types.CustomCPError) {

	azure := AZURE{
		ID:           credentials.Profile.ClientId,
		Key:          credentials.Profile.ClientSecret,
		Tenant:       credentials.Profile.TenantId,
		Subscription: credentials.Profile.SubscriptionId,
		Region:       credentials.Profile.Location,
	}
	err := azure.init()
	if err != (types.CustomCPError{}) {
		return []azureVM{}, err
	}

	instances, err := azure.getAllInstances()
	if err != (types.CustomCPError{}) {
		beego.Error(err.Error)
		return []azureVM{}, err
	}
	return instances, types.CustomCPError{}
}
func GetRegions(credentials vault.AzureProfile, ctx utils.Context) ([]models.Region, types.CustomCPError) {

	azure := AZURE{
		ID:           credentials.Profile.ClientId,
		Key:          credentials.Profile.ClientSecret,
		Tenant:       credentials.Profile.TenantId,
		Subscription: credentials.Profile.SubscriptionId,
		Region:       credentials.Profile.Location,
	}
	err := azure.init()
	if err != (types.CustomCPError{}) {
		return []models.Region{}, err
	}

	regions, err := azure.getRegions(ctx)
	if err != (types.CustomCPError{}) {
		beego.Error(err.Error)
		return []models.Region{}, err
	}
	return regions, types.CustomCPError{}
}
func GetAllMachines() ([]string, types.CustomCPError) {

	regions, err := getAllVMSizes()
	if err != (types.CustomCPError{}) {
		beego.Error(err.Error)
		return []string{}, err
	}
	return regions, types.CustomCPError{}
}

func ValidateProfile(clientId, clientSecret, subscriptionId, tenantId, region string, ctx utils.Context) types.CustomCPError {

	azure := AZURE{
		ID:           clientId,
		Key:          clientSecret,
		Tenant:       tenantId,
		Subscription: subscriptionId,
		Region:       region,
	}
	err := azure.init()
	if err != (types.CustomCPError{}) {
		return err
	}

	_, err = azure.getRegions(ctx)
	if err != (types.CustomCPError{}) {
		beego.Error("Profile is not valid")
		return ApiError(errors.New("Profile is not valid"), "Profile is not valid", int(models.CloudStatusCode))
	}
	return types.CustomCPError{}
}

func ApplyAgent(credentials vault.AzureProfile, token string, ctx utils.Context, clusterName, resourceGroup string) (confError error) {
	companyId := ctx.Data.Company
	projetcID := ctx.Data.ProjectId
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
