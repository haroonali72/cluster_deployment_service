package do

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/db"
	"antelope/models/key_utils"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/types"
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
	ID               bson.ObjectId `json:"-" bson:"_id,omitempty"`
	ProjectId        string        `json:"project_id" bson:"project_id" validate:"required" description:"ID of project [required]`
	DOProjectId      string        `json:"_" bson:"do_project_id"`
	Kube_Credentials interface{}   `json:"kube_credentials" bson:"kube_credentials"`
	Name             string        `json:"name" bson:"name" validate:"required" description:"Cluster Name [required]`
	Status           models.Type   `json:"status" bson:"status" validate:"eq=new|eq=New|eq=NEW|eq=Cluster Creation Failed|eq=Cluster Terminated|eq=Cluster Created" description:"Status of cluster  [required]"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud" validate:"eq=DO|eq=do|eq=Do)"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools        []*NodePool   `json:"node_pools" bson:"node_pools" validate:"required,dive"`
	NetworkName      string        `json:"network_name" bson:"network_name" validate:"required" description:"Network name of corresponding project [required]`
	CompanyId        string        `json:"_" bson:"company_id"`
	TokenName        string        `json:"token_name" bson:"token_name" description:"Rbac Token for Scaling Cluster [required]`
}

type NodePool struct {
	ID                 bson.ObjectId      `json:"_" bson:"_id,omitempty"`
	Name               string             `json:"name" bson:"name" validate:"required" description:"Name of pool [required]`
	NodeCount          int64              `json:"node_count"  bson:"node_count" validate:"required,gte=1" description:"Pool node count [required]"`
	MachineType        string             `json:"machine_type"  bson:"machine_type" validate:"required" description:"Machine type for pool [required]"` //machine size
	Image              ImageReference     `json:"image" bson:"image" description:"Image Information for cluster [required]"`
	PoolSecurityGroups []*string          `json:"security_group_id" bson:"security_group_id" validate:"required" description:"Security Group for cluster [required]"`
	Nodes              []*Node            `json:"nodes,omitempty" bson:"nodes"`
	KeyInfo            key_utils.AZUREKey `json:"key_info" bson:"key_info" description:"SSH Key information [required]"`
	PoolRole           models.PoolRole    `json:"pool_role" bson:"pool_role" validate:"required" description:"Pool role, possible values 'master' or 'slave'. [required]"`
	IsExternal         bool               `json:"is_external" bson:"is_external" description:"Enable Volume Option, possible values 'true' or 'false'  [optional]"`
	ExternalVolume     Volume             `json:"external_volume" bson:"external_volume" description:"Block Store Volume Information ['required' if external volume is enabled']"`
	PrivateNetworking  bool               `json:"private_networking" bson:"private_networking" description:"Option to enable private networking [optional]"`
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
	ID      bson.ObjectId `json:"_" bson:"_id,omitempty"`
	Slug    string        `json:"slug" bson:"slug" description:"Image Slug Information ['optional' if ImageId is provided']"`
	ImageId int           `json:"image_id" bson:"image_id" Image ID ['optional' if Slug is provided']`
}
type Volume struct {
	VolumeSize int64 `json:"volume_size" bson:"volume_size" description:"Block Store Volume Size ['required' if external volume is enabled']`
}
type Project struct {
	ProjectData Data `json:"data"`
}
type Data struct {
	Region string `json:"region"`
}

type Cluster struct {
	Name      string      `json:"name,omitempty" bson:"name,omitempty" v description:"Cluster name"`
	ProjectId string      `json:"project_id" bson:"project_id"  description:"ID of project"`
	Status    models.Type `json:"status,omitempty" bson:"status,omitempty" " description:"Status of cluster"`
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
func GetAllCluster(ctx utils.Context, input rbac_athentication.List) (doClusters []Cluster, err error) {
	var copyData []string
	var clusters []Cluster_Def
	for _, d := range input.Data {
		copyData = append(copyData, d)
	}
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Cluster model: GetAll - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return nil, err1
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoDOClusterCollection)
	err = c.Find(bson.M{"project_id": bson.M{"$in": copyData}, "company_id": ctx.Data.Company}).All(&clusters)
	if err != nil {
		ctx.SendLogs("Cluster model: GetAll - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	for _, cluster := range clusters {
		temp := Cluster{Name: cluster.Name, ProjectId: cluster.ProjectId, Status: cluster.Status}
		doClusters = append(doClusters, temp)
	}
	return doClusters, nil
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
	/*	err = checkMasterPools(cluster)
		if err != nil { //cluster found
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}*/

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

	if oldCluster.Status == (models.Deploying) && update {
		ctx.SendLogs("cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("cluster is in deploying state")
	}
	if oldCluster.Status == (models.Terminating) && update {
		ctx.SendLogs("cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("cluster is in terminating state")
	} else if oldCluster.Status == models.ClusterTerminationFailed && update {
		ctx.SendLogs("Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("Cluster is in termination failed state")
	} else if oldCluster.Status == models.ClusterCreated && update {
		//if !checkScalingChanges(&oldCluster, &cluster) {
		//ctx.SendLogs("Cluster is in runnning state ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		//return errors.New("Cluster is in runnning state")
		//} else {
		//		cluster = oldCluster
		//	}
		ctx.SendLogs("No changes are applicable", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("No changes are applicable")
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
		url = strings.Replace(url, "{projectId}", ctx.Data.ProjectId, -1)
	}
	data, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs("Error in fetching region: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}
	var region Project
	err = json.Unmarshal(data.([]byte), &region.ProjectData)
	if err != nil {
		ctx.SendLogs("Error in fetching region: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return region.ProjectData.Region, err
	}
	return region.ProjectData.Region, nil

}
func GetProfile(profileId string, region string, token string, ctx utils.Context) (int, vault.DOProfile, error) {
	statusCode, data, err := vault.GetCredentialProfile("do", profileId, token, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return statusCode, vault.DOProfile{}, err
	}
	doProfile := vault.DOProfile{}
	err = json.Unmarshal(data, &doProfile)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 500, vault.DOProfile{}, err
	}
	doProfile.Profile.Region = region
	return 0, doProfile, nil

}
func PrintError(confError error, name, projectId string, ctx utils.Context, companyId string) {
	if confError != nil {
		ctx.SendLogs(confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(companyId, "Cluster creation failed : "+name, "error", projectId)
		utils.SendLog(companyId, confError.Error(), "error", projectId)
	}
}
func DeployCluster(cluster Cluster_Def, credentials vault.DOCredentials, ctx utils.Context, companyId string, token string) types.CustomCPError {
	publisher := utils.Notifier{}
	publisher.Init_notifier()

	do := DO{
		AccessKey: credentials.AccessKey,
		Region:    credentials.Region,
	}
	confError := do.init(ctx)
	if confError != (types.CustomCPError{}) {
		cluster.Status = models.ClusterCreationFailed
		err := UpdateCluster(cluster, false, ctx)
		if err != nil {
			PrintError(err, cluster.Name, cluster.ProjectId, ctx, companyId)
		}
		err = db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DO, ctx, confError)
		if err != nil {
			ctx.SendLogs("DODeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return confError
	}

	utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)

	pubSub := publisher.Subscribe(ctx.Data.ProjectId, ctx)

	cluster, confError = do.createCluster(cluster, ctx, companyId, token)
	if confError != (types.CustomCPError{}) {
		PrintError(errors.New(confError.Description), cluster.Name, cluster.ProjectId, ctx, companyId)
		confError_ := do.CleanUp(ctx)
		if confError_ != (types.CustomCPError{}) {
			PrintError(errors.New(confError_.Description), cluster.Name, cluster.ProjectId, ctx, companyId)
		}

		cluster.Status = models.ClusterCreationFailed
		err := UpdateCluster(cluster, false, ctx)
		if err != nil {
			PrintError(err, cluster.Name, cluster.ProjectId, ctx, companyId)
		}
		err = db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DO, ctx, confError)
		if err != nil {
			ctx.SendLogs("DODeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return confError
	}

	cluster.Status = models.ClusterCreated
	err := UpdateCluster(cluster, false, ctx)
	if err != nil {
		confError = types.CustomCPError{StatusCode: 500, Error: "Error occured in updating cluster status in database", Description: "Error occured in updating cluster status in database"}
		PrintError(err, cluster.Name, cluster.ProjectId, ctx, companyId)
		err = db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DO, ctx, confError)
		if err != nil {
			ctx.SendLogs("DODeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return types.CustomCPError{StatusCode: 500, Description: err.Error(), Error: "Error occurred in updating cluster status in database"}

	}
	utils.SendLog(companyId, "Cluster created successfully "+cluster.Name, models.LOGGING_LEVEL_INFO, cluster.ProjectId)

	notify := publisher.RecieveNotification(ctx.Data.ProjectId, ctx, pubSub)
	if notify {
		ctx.SendLogs("DOClusterModel:  Notification received from agent", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
	} else {
		ctx.SendLogs("DOClusterModel:  Notification not received from agent", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	}

	return types.CustomCPError{}
}

func FetchStatus(credentials vault.DOProfile, projectId string, ctx utils.Context, companyId string, token string) (Cluster_Def, types.CustomCPError) {

	cluster, err := GetCluster(projectId, companyId, ctx)
	if err != nil {
		cpErr := types.CustomCPError{StatusCode: 500, Description: err.Error(), Error: "Error occurred in getting cluster"}
		if strings.Contains(err.Error(), "not found") {
			cpErr.StatusCode = 404
		}
		ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, cpErr
	}
	if string(cluster.Status) == strings.ToLower(string(models.New)) {
		cpErr := types.CustomCPError{Error: "Unable to fetch status - Cluster is not deployed yet", Description: "Unable to fetch state - Cluster is not deployed yet", StatusCode: 409}
		return Cluster_Def{}, cpErr
	}
	if cluster.Status == models.Deploying || cluster.Status == models.Terminating || cluster.Status == models.ClusterTerminated {
		cpErr := types.CustomCPError{Error: "Cluster is in " +
			string(cluster.Status) + " state", Description: "Cluster is in " +
			string(cluster.Status) + " state", StatusCode: 409}
		return Cluster_Def{}, cpErr
	}
	customErr, err := db.GetError(cluster.ProjectId, ctx.Data.Company, models.DOKS, ctx)
	if err != nil {
		return Cluster_Def{}, types.CustomCPError{Error: "Error occurred while getting cluster status from database",
			Description: "Error occurred while getting cluster status from database",
			StatusCode:  500}
	}
	if customErr.Err != (types.CustomCPError{}) {
		return Cluster_Def{}, customErr.Err
	}
	do := DO{
		AccessKey: credentials.Profile.AccessKey,
		Region:    credentials.Profile.Region,
	}
	err_ := do.init(ctx)
	if err_ != (types.CustomCPError{}) {
		return Cluster_Def{}, err_
	}

	e := do.fetchStatus(&cluster, ctx, companyId, token)
	if e != (types.CustomCPError{}) {

		ctx.SendLogs("Cluster model: Status - Failed to get lastest status "+e.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, e
	}
	return cluster, (types.CustomCPError{})
}
func TerminateCluster(cluster Cluster_Def, profile vault.DOProfile, ctx utils.Context, companyId, token string) types.CustomCPError {

	publisher := utils.Notifier{}

	publisher.Init_notifier()

	cluster, err := GetCluster(cluster.ProjectId, companyId, ctx)
	if err != nil {
		cpErr := types.CustomCPError{StatusCode: 500, Description: err.Error(), Error: "Error occurred in getting cluster"}

		if strings.Contains(err.Error(), "not found") {
			cpErr.StatusCode = 404
		}
		ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		err = db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DO, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("DODeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}

	if cluster.Status == "" || cluster.Status == "new" {
		text := "Cannot terminate a new cluster"
		ctx.SendLogs("DOClusterModel : "+text+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := types.CustomCPError{StatusCode: 409, Description: text, Error: text}
		err = db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DO, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("DODeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}

	do := DO{
		AccessKey: profile.Profile.AccessKey,
		Region:    profile.Profile.Region,
	}

	cluster.Status = (models.Terminating)
	utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	err_ := do.init(ctx)
	if err_ != (types.CustomCPError{}) {
		ctx.SendLogs(err_.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.Status = "Cluster Termination Failed"
		err = UpdateCluster(cluster, false, ctx)
		if err != nil {
			ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
		}
		err = db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DO, ctx, err_)
		if err != nil {
			ctx.SendLogs("DODeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err_
	}

	err_ = do.terminateCluster(&cluster, ctx, companyId)
	if err_ != (types.CustomCPError{}) {
		utils.SendLog(companyId, "Cluster termination failed: "+err_.Description+cluster.Name, "error", cluster.ProjectId)

		cluster.Status = "Cluster Termination Failed"
		err = UpdateCluster(cluster, false, ctx)
		if err != nil {
			ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)

		}
		err = db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DO, ctx, err_)
		if err != nil {
			ctx.SendLogs("DODeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err_
	}

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
		cpErr := types.CustomCPError{StatusCode: 500, Description: err.Error(), Error: "Error occurred in updating cluster status in database"}
		err = db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.DO, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("DODeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return cpErr
	}
	utils.SendLog(companyId, "Cluster terminated successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	return types.CustomCPError{}
}
func GetAllSSHKeyPair(ctx utils.Context, token string) (keys interface{}, err error) {

	keys, err = vault.GetAllSSHKey("do", ctx, token, "")
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return keys, err
	}
	return keys, nil
}
func CreateSSHkey(keyName string, credentials vault.DOCredentials, token, teams, region string, ctx utils.Context) (keyMaterial string, err types.CustomCPError) {

	keyInfo, err_ := key_utils.GenerateKey(models.DO, keyName, "do@example.com", token, teams, ctx)
	if err_ != nil {
		return "", types.CustomCPError{StatusCode: 500, Description: err_.Error(), Error: "Error occurred in key generation"}
	}
	do := DO{
		AccessKey: credentials.AccessKey,
		Region:    credentials.Region,
	}
	err = do.init(ctx)
	if err != (types.CustomCPError{}) {
		return "", err

	}
	err, key := do.importKey(keyInfo.KeyName, keyInfo.PublicKey, ctx)
	if err != (types.CustomCPError{}) {
		return "", err

	}
	keyInfo.FingerPrint = key.Fingerprint
	keyInfo.ID = key.ID
	_, err_ = vault.PostSSHKey(keyInfo, keyInfo.KeyName, keyInfo.Cloud, ctx, token, teams, "")
	if err_ != nil {
		return "", types.CustomCPError{StatusCode: 500, Description: err_.Error(), Error: "Error occurred in key generation"}
	}
	return keyInfo.PrivateKey, err
}

func DeleteSSHkey(keyName, token string, credentials vault.DOCredentials, ctx utils.Context) types.CustomCPError {

	bytes, err := vault.GetSSHKey(string(models.DO), keyName, token, ctx, "")

	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		return types.CustomCPError{StatusCode: 404, Description: err.Error(), Error: "key not found"}
	}

	key, err := key_utils.AzureKeyConversion(bytes, ctx)
	if err != nil {
		return types.CustomCPError{StatusCode: 404, Description: err.Error(), Error: "key not found"}
	}

	err = vault.DeleteSSHkey(string(models.DO), keyName, token, ctx, credentials.Region)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return types.CustomCPError{StatusCode: 404, Description: err.Error(), Error: "key not found"}
		} else if strings.Contains(err.Error(), "not authorized") {
			return types.CustomCPError{StatusCode: 401, Description: err.Error(), Error: err.Error()}
		} else {
			return types.CustomCPError{StatusCode: 500, Description: err.Error(), Error: "Error occured in deleting key"}
		}

	}
	do := DO{
		AccessKey: credentials.AccessKey,
		Region:    credentials.Region,
	}

	confError := do.init(ctx)
	if confError != (types.CustomCPError{}) {
		return confError
	}

	cpErr := do.deleteKey(key.ID, ctx)
	if cpErr != (types.CustomCPError{}) {
		return cpErr
	}

	return types.CustomCPError{}
}
func GetRegionsAndCores(credentials vault.DOCredentials, ctx utils.Context) ([]godo.Region, types.CustomCPError) {
	do := DO{
		AccessKey: credentials.AccessKey,
	}
	err := do.init(ctx)
	if err != (types.CustomCPError{}) {
		return nil, err
	}
	regions, err := do.getCores(ctx)
	if err != (types.CustomCPError{}) {

		return nil, err
	}
	return regions, err
}

func ValidateProfile(key string, ctx utils.Context) types.CustomCPError {
	do := DO{
		AccessKey: key,
	}
	err := do.init(ctx)
	if err != (types.CustomCPError{}) {
		return err
	}
	_, err = do.getCores(ctx)
	if err != (types.CustomCPError{}) {
		return err
	}
	return types.CustomCPError{}
}
func ValidateDOData(cluster Cluster_Def, ctx utils.Context) error {
	if cluster.ProjectId == "" {

		return errors.New("project Id is empty")

	} else if cluster.Name == "" {

		return errors.New("cluster name is empty")

	} else if cluster.NetworkName == "" {

		return errors.New("kubernetes version is empty")

	} else if len(cluster.NodePools) == 0 {

		return errors.New("node pool length must not be zero")

	} else {

		for _, nodepool := range cluster.NodePools {

			if nodepool.Name == "" {

				return errors.New("node pool name is empty")

			} else if nodepool.MachineType == "" {

				return errors.New("machine type is empty")

			} else if nodepool.NodeCount == 0 {

				return errors.New("node count must be greater than zero")

			} else if nodepool.PoolRole == "" {

				return errors.New("pool role is empty")

			} else if nodepool.KeyInfo.KeyName == "" {

				return errors.New("key name is empty")

			} else if len(nodepool.PoolSecurityGroups) == 0 {

				return errors.New("security group is empty")

			} else if nodepool.Image.Slug == "" && nodepool.Image.ImageId == 0 {

				if nodepool.Image.Slug == "" {

					return errors.New("image slug is empty")

				} else {

					return errors.New("image id is empty")

				}

			}

		}

	}

	return nil
}
