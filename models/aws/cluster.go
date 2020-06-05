package aws

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
	"github.com/astaxie/beego"
	"github.com/aws/aws-sdk-go/service/ec2"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

type Cluster_Def struct {
	ID               bson.ObjectId `json:"-" bson:"_id,omitempty"`
	ProjectId        string        `json:"project_id" bson:"project_id" valid:"required" description:"Id of the project [required]"`
	Kube_Credentials interface{}   `json:"kube_credentials" bson:"kube_credentials"`
	Name             string        `json:"name" bson:"name" valid:"required" description:"Name of the project [required]"`
	Status           models.Type   `json:"status" bson:"status" validate:"required,eq=new|eq=New|eq=NEW|eq=Cluster Creation Failed|eq=Cluster Terminated|eq=Cluster Created"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud" valid:"eq=AWS|aws|Aws"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools        []*NodePool   `json:"node_pools" bson:"node_pools" valid:"required,dive" description:"Nodepools info [required]`
	NetworkName      string        `json:"network_name" bson:"network_name" valid:"required" description:"Network to create cluster [required]"`
	CompanyId        string        `json:"company_id" bson:"company_id" description:"Company Id of the project [optional]"`
	TokenName        string        `json:"token_name" bson:"token_name" description:"Token name [optional]"`
}

type NodePool struct {
	ID                 bson.ObjectId    `json:"-" bson:"_id,omitempty"`
	Name               string           `json:"name" bson:"name" valid:"required" description:"Name of the nodepool [required]"`
	NodeCount          int64            `json:"node_count" bson:"node_count" valid:"required,matches(^[0-9]+$)" description:"Number of nodes in nodepool [required]"`
	MachineType        string           `json:"machine_type" bson:"machine_type" valid:"required" description:"Machine type of nodes in nodepool [required]"`
	Ami                Ami              `json:"ami" bson:"ami" valid:"required" description:"Ami to create nodes od nodepool [required]"`
	PoolSubnet         string           `json:"subnet_id" bson:"subnet_id" valid:"required" description:"Subnet to create nodepool [required]"`
	PoolSecurityGroups []*string        `json:"security_group_id" bson:"security_group_id" valid:"gt=0,required,dive" description:"Security groups attached with the nodepool [required]"`
	Nodes              []*Node          `json:"nodes" bson:"nodes" valid:"required,dive" description:"Nodes in the nodepool [required]"`
	KeyInfo            key_utils.AWSKey `json:"key_info" bson:"key_info" valid:"required,dive" description:"Information of SSH key [required]"`
	PoolRole           models.PoolRole  `json:"pool_role" bson:"pool_role" valid:"required,dive" description:"Role of pool.Valid values are master and slave [required]"`
	EnableScaling      bool             `json:"enable_scaling" bson:"enable_scaling" description:"To enable scalimng of nodepool [optional]"`
	Scaling            AutoScaling      `json:"auto_scaling" bson:"auto_scaling" description:"Autoscaling details [optional]"`
	IsExternal         bool             `json:"is_external" bson:"is_external"`
	ExternalVolume     Volume           `json:"external_volume" bson:"external_volume" description:"External volume details [optional]"`
	EnablePublicIP     bool             `json:"enable_public_ip" bson:"enable_public_ip" description:"Assign public ip to nodepool's node [optional]"`
}
type AutoScaling struct {
	MaxScalingGroupSize int64       `json:"max_scaling_group_size" bson:"max_scaling_group_size" valid:"required" description:"Max count of node for scaling [required]"`
	State               models.Type `json:"status" bson:"status" description:"Status of autoscaling [optional]"`
}
type Node struct {
	CloudId    string `json:"cloud_id" bson:"cloud_id",omitempty"`
	KeyName    string `json:"key_name" bson:"key_name",omitempty" description:"Name of the key to be used with node [required]"`
	SSHKey     string `json:"ssh_key" bson:"ssh_key",omitempty" description:"Key to be used with node [optional]"`
	NodeState  string `json:"node_state" bson:"node_state",omitempty" description:"Current state of the node [optional]"`
	Name       string `json:"name" bson:"name",omitempty" valid:"required" description:"Name of the node [optional]"`
	PrivateIP  string `json:"private_ip" bson:"private_ip",omitempty" description:"PrivateIp of the node [optional]"`
	PublicIP   string `json:"public_ip" bson:"public_ip",omitempty" description:"PrivateIp of the node [optional]"`
	PublicDNS  string `json:"public_dns" bson:"public_dns",omitempty" description:"PublicDNS of the node [optional]"`
	PrivateDNS string `json:"private_dns" bson:"private_dns",omitempty" description:"PrivateDNS of the node [optional]"`
	UserName   string `json:"user_name" bson:"user_name",omitempty" description:"User name to use with node [optional]"`
}

type Ami struct {
	ID         bson.ObjectId `json:"-" bson:"_id,omitempty"`
	Name       string        `json:"name" bson:"name" description:"Name of the AMI [required]"`
	AmiId      string        `json:"ami_id" bson:"ami_id" valid:"required" description:"Ami id of the instance [required]"`
	Username   string        `json:"username" bson:"username" valid:"required" description:"Username of the instance [required]"`
	RootVolume Volume        `json:"root_volume" bson:"root_volume" valid:"required,dive" description:"Instance root volume details [required]"`
}
type Volume struct {
	VolumeType          string `json:"volume_type" bson:"volume_type" valid:"required" description:"Type of the volume.Valid values are General Purpose SSD,IOPS SSD,Magnetic volumes[required]"`
	VolumeSize          int64  `json:"volume_size" bson:"volume_size" valid:"required" description:"Size of the volume [required]"`
	DeleteOnTermination bool   `json:"delete_on_termination" bson:"delete_on_termination" description:"Select if volume should terminate on deletion [optional]"`
	Iops                int64  `json:"iops" bson:"iops" valid:"required" description:"IOPS of volume [required]"`
}
type Project struct {
	ProjectData Data `json:"data" description:"Project data of the cluster [optional]"`
}
type Data struct {
	Region string `json:"region" description:"Region of the cluster [optional]"`
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

func GetProfile(profileId string, region string, token string, ctx utils.Context) (int, vault.AwsProfile, error) {
	statusCode, data, err := vault.GetCredentialProfile("aws", profileId, token, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return statusCode, vault.AwsProfile{}, err
	}
	awsProfile := vault.AwsProfile{}
	err = json.Unmarshal(data, &awsProfile)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 500, vault.AwsProfile{}, err
	}
	awsProfile.Profile.Region = region
	return 0, awsProfile, nil

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

func GetNetwork(token, projectId string, ctx utils.Context) (types.AWSNetwork, error) {

	url := getNetworkHost("aws", projectId)

	data, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.AWSNetwork{}, err
	}
	var net types.AWSNetwork
	err = json.Unmarshal(data.([]byte), &net)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.AWSNetwork{}, err
	}

	return net, nil
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
	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoAwsClusterCollection, cluster)
	if err != nil {
		ctx.SendLogs("Cluster model: Create - Got error inserting cluster to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}

func GetCluster(projectId, companyId string, ctx utils.Context) (cluster Cluster_Def, err error) {

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Cluster model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAwsClusterCollection)
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
	c := session.DB(mc.MongoDb).C(mc.MongoAwsClusterCollection)
	err = c.Find(bson.M{}).All(&clusters)
	if err != nil {
		ctx.SendLogs("Cluster model: GetAll - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	return clusters, nil
}

func UpdateCluster(cluster Cluster_Def, update bool, ctx utils.Context) error {
	oldCluster, err := GetCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Update - Cluster   does not exist in the database: "+cluster.Name+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	/*if oldCluster.Status == string(models.Deploying) && update {
		ctx.SendLogs("cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("cluster is in deploying state")
	}*/
	if oldCluster.Status == models.ClusterTerminationFailed && update {
		ctx.SendLogs("Cluster creation is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("Cluster creation is in termination failed state")
	}

	if oldCluster.Status == models.ClusterCreated && update {
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
	c := session.DB(mc.MongoDb).C(mc.MongoAwsClusterCollection)
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
func DeployCluster(cluster Cluster_Def, credentials vault.AwsCredentials, ctx utils.Context, companyId string, token string) (err types.CustomCPError) {
	publisher := utils.Notifier{}
	confError := publisher.Init_notifier()
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		confErr:= ApiError(confError,"Error in deploying cluster")
		err := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.AWS, ctx,confErr)
		if err != nil {
			ctx.SendLogs("AWSClusterModel:  Deploy :Error in saving error "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return confErr
	}
	aws := AWS{
		AccessKey: credentials.AccessKey,
		SecretKey: credentials.SecretKey,
		Region:    credentials.Region,
	}
	err = aws.init()
	if err !=(types.CustomCPError{}) {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		cluster.Status = models.ClusterCreationFailed
		confError = UpdateCluster(cluster, false, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		}
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.AWS, ctx, err)
		if err_ != nil {
			ctx.SendLogs("AWSClusterModel:  Deploy - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}

	utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	createdPools, err := aws.createCluster(cluster, ctx, companyId, token)
	if err != (types.CustomCPError{}) {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		cluster.Status = models.ClusterCreationFailed
		err1:=ApiError(confError,"Error in cluster creation")
		err := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.AWS, ctx,err1)
		if err != nil {
			ctx.SendLogs("AWSClusterModel:  Deploy :Error in saving error "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		confErr := aws.CleanUp(cluster, ctx)
		if confErr != (types.CustomCPError{}) {
			PrintError(errors.New(confErr.Error), cluster.Name, cluster.ProjectId, ctx, companyId)
		}

		cluster.Status = models.ClusterCreationFailed
		confError = UpdateCluster(cluster, false, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		}

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err1
	}

	cluster = updateNodePool(createdPools, cluster, ctx)

	for _, pool := range cluster.NodePools {
		for _, node := range pool.Nodes {
			node.NodeState = ""
			node.PublicIP = ""
			node.PrivateIP = ""
		}
	}

	UpdateScalingStatus(&cluster)
	confError = UpdateCluster(cluster, false, ctx)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		confErr:=ApiError(confError,"Error in cluster creation")
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.AWS, ctx,confErr)
		if err_ != nil {
			ctx.SendLogs("AWSClusterModel:  Deploy - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)

		return confErr
	}

	utils.SendLog(companyId, "Cluster created successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)

	return types.CustomCPError{}
}
func FetchStatus(credentials vault.AwsProfile, projectId string, ctx utils.Context, companyId string, token string) (Cluster_Def, types.CustomCPError) {

	cluster, err := GetCluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, ApiError(err,"Error in fetching cluster")
	}
	//splits := strings.Split(credentials, ":")
	aws := AWS{
		AccessKey: credentials.Profile.AccessKey,
		SecretKey: credentials.Profile.SecretKey,
		Region:    credentials.Profile.Region,
	}
	err1 := aws.init()
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, ApiError(err,"Error in fetching cluster")
	}

	_, e := aws.fetchStatus(&cluster, ctx, companyId, token)
	if e != (types.CustomCPError{}) {
		ctx.SendLogs("Cluster model: Status - Failed to get lastest status "+e.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, e
	}
	/*	err = UpdateCluster(c)
		if err != nil {
			beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
			return Cluster_Def{}, err
		}*/
	return cluster, types.CustomCPError{}
}
func TerminateCluster(cluster Cluster_Def, profile vault.AwsProfile, ctx utils.Context, companyId, token string) types.CustomCPError {

	publisher := utils.Notifier{}

	pub_err := publisher.Init_notifier()
	if pub_err != nil {
		ctx.SendLogs(pub_err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		confErr:=ApiError(pub_err,"Error in terminating cluster")
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.AWS, ctx,confErr)
		if err_ != nil {
			ctx.SendLogs("AWSClusterModel:  terminate - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return confErr
	}

	cluster, err := GetCluster(cluster.ProjectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		confErr:=ApiError(err,"Error in terminating cluster")
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.AWS, ctx,confErr)
		if err_ != nil {
			ctx.SendLogs("AWSClusterModel:  terminate - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return  confErr
	}

	if cluster.Status == "" || cluster.Status == models.New {
		text := "Cannot terminate a new cluster"
		ctx.SendLogs("AwsClusterModel : "+text+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return  ApiError(errors.New(text),"Error in terminating cluster")
	}

	aws := AWS{
		AccessKey: profile.Profile.AccessKey,
		SecretKey: profile.Profile.SecretKey,
		Region:    profile.Profile.Region,
	}

	utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	err1 := aws.init()
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs(err1.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.Status =models.ClusterTerminationFailed
		err = UpdateCluster(cluster, false, ctx)
		if err != nil {
			ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
		}
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.AWS, ctx,err1)
		if err_ != nil {
			ctx.SendLogs("AWSClusterModel:  terminate - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err1
	}

	flag := aws.terminateCluster(cluster, ctx, companyId)
	if flag {
		utils.SendLog(companyId, "Cluster termination failed: "+cluster.Name, "error", cluster.ProjectId)

		cluster.Status = models.ClusterTerminationFailed
		err = UpdateCluster(cluster, false, ctx)
		if err != nil {
			ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
			publisher.Notify(cluster.ProjectId, "Status Available", ctx)
			return ApiError(err,"Error in terminating cluster")
		}
		confErr:=ApiError(err,"Error in terminating cluster")
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.AWS, ctx,confErr)
		if err_ != nil {
			ctx.SendLogs("AWSClusterModel:  terminate - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return types.CustomCPError{}
	}

	var flagcheck bool
	for {
		flagcheck = false
		_, err1 = aws.fetchStatus(&cluster, ctx, companyId, token)
		if err1 != (types.CustomCPError{}) {
			beego.Error(err1)
		}
		for _, nodePools := range cluster.NodePools {
			for _, node := range nodePools.Nodes {
				if node.NodeState != "terminated" {
					flagcheck = true
					break
				}
			}
		}
		if !flagcheck {
			break
		}
		time.Sleep(time.Second * 5)
	}

	for _, pools := range cluster.NodePools {
		var nodes []*Node
		pools.Nodes = nodes
		pools.KeyInfo.KeyType = models.CPKey
	}
	cluster.Status = models.ClusterTerminated
	err = UpdateCluster(cluster, false, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
		utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
		confErr:=ApiError(err,"Error in terminating cluster")
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.AWS, ctx,confErr)
		if err_ != nil {
			ctx.SendLogs("AWSClusterModel:  terminate - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return confErr
	}
	utils.SendLog(companyId, "Cluster terminated successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	return types.CustomCPError{}
}
func updateNodePool(createdPools []CreatedPool, cluster Cluster_Def, ctx utils.Context) Cluster_Def {
	for index, nodepool := range cluster.NodePools {

		var updatedNodes []*Node

		for _, createdPool := range createdPools {

			if createdPool.PoolName == nodepool.Name {

				for _, inst := range createdPool.Instances {

					var node Node
					beego.Info(*inst.Tags[0].Key, *inst.Tags[0].Value)
					for _, tag := range inst.Tags {
						beego.Info(*tag.Key, *tag.Value)
						if *tag.Key == "Name" {
							node.Name = *tag.Value
						}
					}
					node.CloudId = *inst.InstanceId
					node.NodeState = *inst.State.Name
					node.PrivateIP = *inst.PrivateIpAddress
					if inst.PublicIpAddress != nil {
						node.PublicIP = *inst.PublicIpAddress
					}
					node.UserName = nodepool.Ami.Username

					updatedNodes = append(updatedNodes, &node)
					beego.Info("Cluster model: Instances added")
				}
			}
		}

		ctx.SendLogs("Cluster model: updated nodes in pools", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		cluster.NodePools[index].Nodes = updatedNodes
	}
	cluster.Status = models.ClusterCreated
	return cluster
}
func GetAllSSHKeyPair(ctx utils.Context, token, region string) (keys interface{}, err error) {

	keys, err = vault.GetAllSSHKey("aws", ctx, token, region)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return keys, err
	}
	return keys, nil
}
func GetSSHKeyPair(keyname string) (keys *key_utils.AWSKey, err error) {
	ctx := new(utils.Context)
	session, err := db.GetMongoSession(*ctx)
	if err != nil {
		beego.Error("Cluster model: Get - Got error while connecting to the database: ", err)
		return keys, err
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoSshKeyCollection)
	err = c.Find(bson.M{"cloud": models.AWS, "key_name": keyname}).One(&keys)
	if err != nil {
		return keys, err
	}
	return keys, nil
}
func InsertSSHKeyPair(key key_utils.AWSKey) (err error) {
	key.Cloud = models.AWS
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
func GetAwsSSHKeyPair(credentials string) ([]*ec2.KeyPairInfo, types.CustomCPError) {

	splits := strings.Split(credentials, ":")
	aws := AWS{
		AccessKey: splits[0],
		SecretKey: splits[1],
		Region:    splits[2],
	}
	err := aws.init()
	if err != (types.CustomCPError{}) {
		return nil, err
	}

	keys, e := aws.getSSHKey()
	if e != (types.CustomCPError{}) {
		beego.Error("Cluster model: Status - Failed to get ssh key pairs ", e.Error)
		return nil, e
	}

	return keys, types.CustomCPError{}
}
func GetAWSAmi(credentials vault.AwsProfile, amiId string, ctx utils.Context, token string) ([]*ec2.BlockDeviceMapping, types.CustomCPError) {

	aws := AWS{
		AccessKey: credentials.Profile.AccessKey,
		SecretKey: credentials.Profile.SecretKey,
		Region:    credentials.Profile.Region,
	}
	err := aws.init()
	if err != (types.CustomCPError{}){
		ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	amis, e := aws.describeAmi(&amiId, ctx)
	if e != (types.CustomCPError{}) {
		ctx.SendLogs("Cluster model: Status - Failed to get ami details "+e.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return nil, e
	}
	return amis, types.CustomCPError{}
}

func EnableScaling(credentials vault.AwsProfile, cluster Cluster_Def, ctx utils.Context, token string) types.CustomCPError {

	aws := AWS{
		AccessKey: credentials.Profile.AccessKey,
		SecretKey: credentials.Profile.SecretKey,
		Region:    credentials.Profile.Region,
	}
	err := aws.init()
	if err != (types.CustomCPError{}) {
		ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return err
	}

	e := aws.enableScaling(cluster, ctx, token)
	if e != (types.CustomCPError{}) {
		ctx.SendLogs("Cluster model: Status - Failed to enable  scaling"+e.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return e
	}
	UpdateScalingStatus(&cluster)
	err1 := UpdateCluster(cluster, false, ctx)
	if err1 != nil {
		ctx.SendLogs("Cluster model: Status - Failed to enable  scaling"+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return ApiError(err1,"")
	}

	return (types.CustomCPError{})
}

func UpdateScalingStatus(cluster *Cluster_Def) {
	for _, pool := range cluster.NodePools {
		pool.Scaling.State = models.Created
	}
}
func CreateSSHkey(keyName string, credentials vault.AwsCredentials, token, teams, region string, ctx utils.Context) (keyMaterial string, err types.CustomCPError) {

	keyMaterial, err = GenerateAWSKey(keyName, credentials, token, teams, region, ctx)
	if err != (types.CustomCPError{}) {
		return "", err
	}

	return keyMaterial, err
}

func DeleteSSHkey(keyName, token string, credentials vault.AwsCredentials, ctx utils.Context) types.CustomCPError {

	err := DeleteAWSKey(keyName, token, credentials, ctx)
	if err != (types.CustomCPError{}) {
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

func GetRegions(ctx utils.Context) ([]models.Region, types.CustomCPError) {

	regions, err := api_handler.GetAwsRegions()
	if err != nil {
		ctx.SendLogs("Cluster model: Status - Failed to get aws regions "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []models.Region{}, ApiError(err,"Error in fetching region")
	}

	return regions, types.CustomCPError{}
}

func GetZones(credentials vault.AwsProfile, ctx utils.Context) ([]*string, types.CustomCPError) {

	aws := AWS{
		AccessKey: credentials.Profile.AccessKey,
		SecretKey: credentials.Profile.SecretKey,
		Region:    credentials.Profile.Region,
	}
	err := aws.init()
	if err != (types.CustomCPError{}) {
		ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	zones, e := aws.GetZones(ctx)
	if e != (types.CustomCPError{}) {
		ctx.SendLogs("Cluster model: Status - Failed to get aws regions "+e.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return nil, e
	}
	//var zone []string
	/*for _,z := range zones{
		z := z[len(z)-1:]
		zone = append(zone,*z)
	}*/
	return zones,types.CustomCPError{}
}

func GetAllMachines() ([]string, error) {
	machines, err := api_handler.GetAwsMachines()
	if err != nil {
		return []string{}, nil
	}

	return machines, nil
}

func ValidateProfile(key, secret, region string, ctx utils.Context) types.CustomCPError {

	aws := AWS{
		AccessKey: key,
		SecretKey: secret,
		Region:    region,
	}

	err := aws.init()
	if err != (types.CustomCPError{}) {
		ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	err = aws.validateProfile(ctx)
	if err != (types.CustomCPError{}) {
		ctx.SendLogs("Cluster model: Status - Failed to get aws regions "+err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return types.CustomCPError{}
}

func ApplyAgent(credentials vault.AwsProfile, token string, ctx utils.Context, clusterName string) (confError error) {
	companyId := ctx.Data.Company
	projetcID := ctx.Data.ProjectId
	data2, err := woodpecker.GetCertificate(projetcID, token, ctx)
	if err != nil {
		ctx.SendLogs("EKSClusterModel : Apply Agent -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	filePath := "/tmp/" + companyId + "/" + projetcID + "/"
	cmd := "mkdir -p " + filePath + " && echo '" + data2 + "'>" + filePath + "agent.yaml"
	output, err := models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("EKSClusterModel : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	cmd = "sudo docker run --rm --name " + companyId + projetcID + " -e accessKey=" + credentials.Profile.AccessKey + " -e cluster=" + clusterName + " -e secretKey=" + credentials.Profile.SecretKey + " -e region=" + credentials.Profile.Region + " -e yamlFile=" + filePath + "agent.yaml -v " + filePath + ":" + filePath + " " + models.EKSAuthContainerName

	output, err = models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("AKSClusterModel : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
