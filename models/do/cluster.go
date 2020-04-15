package do

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/db"
	"antelope/models/key_utils"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"antelope/models/vault"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/digitalocean/godo"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

type Cluster_Def struct {
	ID               bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	ProjectId        string        `json:"project_id" bson:"project_id" valid:"required"`
	DOProjectId      string        `json:"do_project_id" bson:"do_project_id"`
	Kube_Credentials interface{}   `json:"kube_credentials" bson:"kube_credentials"`
	Name             string        `json:"name" bson:"name" valid:"required"`
	Status           string        `json:"status" bson:"status" valid:"in(New|new)"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud" valid:"in(DO|do)"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools        []*NodePool   `json:"node_pools" bson:"node_pools" valid:"required"`
	NetworkName      string        `json:"network_name" bson:"network_name" valid:"required"`
	CompanyId        string        `json:"company_id" bson:"company_id"`
	TokenName        string        `json:"token_name" bson:"token_name"`
}

type NodePool struct {
	ID                 bson.ObjectId      `json:"_id" bson:"_id,omitempty"`
	Name               string             `json:"name" bson:"name" valid:"required"`
	NodeCount          int64              `json:"node_count" bson:"node_count" valid:"required,matches(^[0-9]+$)"`
	MachineType        string             `json:"machine_type" bson:"machine_type" valid:"required"`
	Image              ImageReference     `json:"image" bson:"image"`
	PoolSecurityGroups []*string          `json:"security_group_id" bson:"security_group_id" valid:"required"`
	Nodes              []*Node            `json:"nodes" bson:"nodes"`
	KeyInfo            key_utils.AZUREKey `json:"key_info" bson:"key_info"`
	PoolRole           models.PoolRole    `json:"pool_role" bson:"pool_role" valid:"required"`
	IsExternal         bool               `json:"is_external" bson:"is_external"`
	ExternalVolume     Volume             `json:"external_volume" bson:"external_volume"`
	PrivateNetworking  bool               `json:"private_networking" bson:"private_networking"`
}

type Node struct {
	CloudId    int    `json:"cloud_id" bson:"cloud_id",omitempty"`
	NodeState  string `json:"node_state" bson:"node_state",omitempty"`
	Name       string `json:"name" bson:"name",omitempty"`
	PrivateIP  string `json:"private_ip" bson:"private_ip",omitempty"`
	PublicIP   string `json:"public_ip" bson:"public_ip",omitempty"`
	PublicDNS  string `json:"public_dns" bson:"public_dns",omitempty"`
	PrivateDNS string `json:"private_dns" bson:"private_dns",omitempty"`
	UserName   string `json:"user_name" bson:"user_name",omitempty"`
	VolumeId   string `json:"volume_id" bson:"volume_id"`
}

type ImageReference struct {
	ID      bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Slug    string        `json:"slug" bson:"slug,omitempty"`
	ImageId int           `json:"image_id" bson:"image_id,omitempty"`
}
type Volume struct {
	VolumeSize int64 `json:"volume_size" bson:"volume_size"`
}
type Project struct {
	ProjectData Data `json:"data"`
}
type Data struct {
	Region string `json:"region"`
}

func GetCluster(projectId, companyId string, ctx utils.Context) (cluster Cluster_Def, err error) {

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Cluster model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoDOClusterCollection)
	err = c.Find(bson.M{"project_id": projectId, "company_id": companyId}).One(&cluster)
	if err != nil {
		ctx.SendLogs("Cluster model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, err
	}
	return cluster, nil
}
func GetAllCluster(ctx utils.Context, input rbac_athentication.List) (clusters []Cluster_Def, err error) {

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Cluster model: GetAll - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return nil, err1
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoDOClusterCollection)
	err = c.Find(bson.M{}).All(&clusters)
	if err != nil {
		ctx.SendLogs("Cluster model: GetAll - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	return clusters, nil
}
func GetNetwork(token, projectId string, ctx utils.Context) error {

	url := getNetworkHost("do", projectId)

	_, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
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

func CreateCluster(cluster Cluster_Def, ctx utils.Context) error {
	_, err := GetCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err == nil { //cluster found
		ctx.SendLogs("Cluster model: Create - Cluster  already exists in the database: "+cluster.Name, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("Cluster model: Create - Cluster  already exists in the database: " + cluster.Name)
	}
	err = checkMasterPools(cluster)
	if err != nil { //cluster found
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	//err = checkClusterSize(cluster, ctx)
	//if err != nil { //cluster size limit exceed
	//	ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	//	return err
	//}

	/*	if subscriptionID != "" {
		err = checkCoresLimit(cluster, subscriptionID, ctx)
		if err != nil { //core size limit exceed
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}*/
	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoDOClusterCollection, cluster)
	if err != nil {
		ctx.SendLogs("Cluster model: Create - Got error inserting cluster to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}
func UpdateCluster(cluster Cluster_Def, update bool, ctx utils.Context) error {
	oldCluster, err := GetCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Update - Cluster   does not exist in the database: "+cluster.Name+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
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
		//if !checkScalingChanges(&oldCluster, &cluster) {
		ctx.SendLogs("Cluster is in runnning state ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("Cluster is in runnning state")
		//} else {
		//	cluster = oldCluster
		//}
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
		ctx.SendLogs("Cluster model: Update - Got error deleting cluster: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
	c := session.DB(mc.MongoDb).C(mc.MongoDOClusterCollection)
	err = c.Remove(bson.M{"project_id": projectId, "company_id": companyId})
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}
func GetRegion(token string, ctx utils.Context) (string, error) {
	fmt.Println(ctx.Data.ProjectId)
	url := beego.AppConfig.String("raccoon_url") + models.ProjectGetEndpoint
	if strings.Contains(url, "{projectId}") {
		url = strings.Replace(url, "{projectId}",ctx.Data.ProjectId, -1)
	}
	data, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs("Fetch region:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
func GetProfile(profileId string, region string, token string, ctx utils.Context) (vault.DOProfile, error) {
	data, err := vault.GetCredentialProfile("do", profileId, token, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return vault.DOProfile{}, err
	}
	doProfile := vault.DOProfile{}
	err = json.Unmarshal(data, &doProfile)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return vault.DOProfile{}, err
	}
	doProfile.Profile.Region = region
	return doProfile, nil

}
func PrintError(confError error, name, projectId string, ctx utils.Context, companyId string) {
	if confError != nil {
		ctx.SendLogs(confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(companyId, "Cluster creation failed : "+name, "error", projectId)
		utils.SendLog(companyId, confError.Error(), "error", projectId)

	}
}
func DeployCluster(cluster Cluster_Def, credentials vault.DOCredentials, ctx utils.Context, companyId string, token string) (confError error) {
	publisher := utils.Notifier{}
	confError = publisher.Init_notifier()
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		return confError
	}
	do := DO{
		AccessKey: credentials.AccessKey,
		Region:    credentials.Region,
	}
	confError = do.init(ctx)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		cluster.Status = "Cluster Creation Failed"
		confError = UpdateCluster(cluster, false, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return confError
	}

	utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	cluster, confError = do.createCluster(cluster, ctx, companyId, token)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		confError = do.CleanUp(ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		}

		cluster.Status = "Cluster Creation Failed"
		confError = UpdateCluster(cluster, false, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return confError
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
func FetchStatus(credentials vault.DOProfile, projectId string, ctx utils.Context, companyId string, token string) (Cluster_Def, error) {

	cluster, err := GetCluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, err
	}
	//splits := strings.Split(credentials, ":")
	do := DO{
		AccessKey: credentials.Profile.AccessKey,
		Region:    credentials.Profile.Region,
	}
	err = do.init(ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, err
	}

	e := do.fetchStatus(&cluster, ctx, companyId, token)
	if e != nil {

		ctx.SendLogs("Cluster model: Status - Failed to get lastest status "+e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, e
	}
	return cluster, nil
}
func TerminateCluster(cluster Cluster_Def, profile vault.DOProfile, ctx utils.Context, companyId, token string) error {

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
		ctx.SendLogs("DOClusterModel : "+text+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return errors.New(text)
	}

	do := DO{
		AccessKey: profile.Profile.AccessKey,
		Region:    profile.Profile.Region,
	}

	cluster.Status = string(models.Terminating)
	utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	err = do.init(ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.Status = "Cluster Termination Failed"
		err = UpdateCluster(cluster, false, ctx)
		if err != nil {
			ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}

	err = do.terminateCluster(&cluster, ctx, companyId)
	if err != nil {
		utils.SendLog(companyId, "Cluster termination failed: "+err.Error()+cluster.Name, "error", cluster.ProjectId)

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

	//var flagcheck bool
	//for {
	//	flagcheck = false
	//	err = do.fetchStatus(&cluster, ctx, companyId, token)
	//	if err != nil {
	//		beego.Error(err)
	//	}
	//	for _, nodePools := range cluster.NodePools {
	//		for _, node := range nodePools.Nodes {
	//			if node.NodeState != "terminated" {
	//				flagcheck = true
	//				break
	//			}
	//		}
	//	}
	//	if !flagcheck {
	//		break
	//	}
	//	time.Sleep(time.Second * 5)
	//}

	for _, pools := range cluster.NodePools {
		var nodes []*Node
		pools.Nodes = nodes
		pools.KeyInfo.KeyType = models.CPKey
	}
	cluster.Status = "Cluster Terminated"
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
func GetAllSSHKeyPair(ctx utils.Context, token string) (keys interface{}, err error) {

	keys, err = vault.GetAllSSHKey("do", ctx, token, "")
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return keys, err
	}
	return keys, nil
}
func CreateSSHkey(keyName string, credentials vault.DOCredentials, token, teams, region string, ctx utils.Context) (keyMaterial string, err error) {

	keyInfo, err := key_utils.GenerateKey(models.DO, keyName, "do@example.com", token, teams, ctx)
	if err != nil {
		return "", err
	}
	do := DO{
		AccessKey: credentials.AccessKey,
		Region:    credentials.Region,
	}
	err = do.init(ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err

	}
	err, key := do.importKey(keyInfo.KeyName, keyInfo.PublicKey, ctx)
	if err != nil {
		return "", err

	}
	keyInfo.FingerPrint = key.Fingerprint
	keyInfo.ID = key.ID
	_, err = vault.PostSSHKey(keyInfo, keyInfo.KeyName, keyInfo.Cloud, ctx, token, teams, "")
	if err != nil {
		return "", err
	}

	return keyInfo.PrivateKey, err
}

func DeleteSSHkey(keyName, token string, credentials vault.DOCredentials, ctx utils.Context) error {

	bytes, err := vault.GetSSHKey(string(models.DO), keyName, token, ctx, "")

	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		return err
	}

	key, err := key_utils.AzureKeyConversion(bytes, ctx)
	if err != nil {
		return err
	}

	err = vault.DeleteSSHkey(string(models.DO), keyName, token, ctx, credentials.Region)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	do := DO{
		AccessKey: credentials.AccessKey,
		Region:    credentials.Region,
	}

	confError := do.init(ctx)
	if confError != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return confError
	}

	err = do.deleteKey(key.ID, ctx)
	if err != nil {
		return err
	}

	return nil
}
func GetRegionsAndCores(credentials vault.DOCredentials, ctx utils.Context) ([]godo.Region, error) {
	do := DO{
		AccessKey: credentials.AccessKey,
	}
	err := do.init(ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	regions, err := do.getCores(ctx)
	if err != nil {

		return nil, err
	}
	return regions, err
}

func ValidateProfile(key string, ctx utils.Context) error {
	do := DO{
		AccessKey: key,
	}
	err := do.init(ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	_, err = do.getCores(ctx)
	if err != nil {

		return err
	}
	return err
}
