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
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	"github.com/aws/aws-sdk-go/service/ec2"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

type Cluster_Def struct {
	ID               bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	ProjectId        string        `json:"project_id" bson:"project_id" valid:"required"`
	Kube_Credentials interface{}   `json:"kube_credentials" bson:"kube_credentials"`
	Name             string        `json:"name" bson:"name" valid:"required"`
	Status           string        `json:"status" bson:"status" valid:"in(New|new)"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud" valid:"in(AWS|aws)"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools        []*NodePool   `json:"node_pools" bson:"node_pools" valid:"required"`
	NetworkName      string        `json:"network_name" bson:"network_name" valid:"required"`
	CompanyId        string        `json:"company_id" bson:"company_id"`
	TokenName        string        `json:"token_name" bson:"token_name"`
}

type NodePool struct {
	ID                 bson.ObjectId    `json:"_id" bson:"_id,omitempty"`
	Name               string           `json:"name" bson:"name" valid:"required"`
	NodeCount          int64            `json:"node_count" bson:"node_count" valid:"required,matches(^[0-9]+$)"`
	MachineType        string           `json:"machine_type" bson:"machine_type" valid:"required"`
	Ami                Ami              `json:"ami" bson:"ami"`
	PoolSubnet         string           `json:"subnet_id" bson:"subnet_id" valid:"required"`
	PoolSecurityGroups []*string        `json:"security_group_id" bson:"security_group_id" valid:"required"`
	Nodes              []*Node          `json:"nodes" bson:"nodes"`
	KeyInfo            key_utils.AWSKey `json:"key_info" bson:"key_info"`
	PoolRole           models.PoolRole  `json:"pool_role" bson:"pool_role" valid:"required"`
	EnableScaling      bool             `json:"enable_scaling" bson:"enable_scaling"`
	Scaling            AutoScaling      `json:"auto_scaling" bson:"auto_scaling"`
	IsExternal         bool             `json:"is_external" bson:"is_external"`
	ExternalVolume     Volume           `json:"external_volume" bson:"external_volume"`
	EnablePublicIP     bool             `json:"enable_public_ip" bson:"enable_public_ip"`
}
type AutoScaling struct {
	MaxScalingGroupSize int64       `json:"max_scaling_group_size" bson:"max_scaling_group_size"`
	State               models.Type `json:"status" bson:"status"`
}
type Node struct {
	CloudId    string `json:"cloud_id" bson:"cloud_id",omitempty"`
	KeyName    string `json:"key_name" bson:"key_name",omitempty"`
	SSHKey     string `json:"ssh_key" bson:"ssh_key",omitempty"`
	NodeState  string `json:"node_state" bson:"node_state",omitempty"`
	Name       string `json:"name" bson:"name",omitempty"`
	PrivateIP  string `json:"private_ip" bson:"private_ip",omitempty"`
	PublicIP   string `json:"public_ip" bson:"public_ip",omitempty"`
	PublicDNS  string `json:"public_dns" bson:"public_dns",omitempty"`
	PrivateDNS string `json:"private_dns" bson:"private_dns",omitempty"`
	UserName   string `json:"user_name" bson:"user_name",omitempty"`
}

type Ami struct {
	ID         bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name       string        `json:"name" bson:"name"`
	AmiId      string        `json:"ami_id" bson:"ami_id" valid:"required"`
	Username   string        `json:"username" bson:"username" valid:"required"`
	RootVolume Volume        `json:"root_volume" bson:"root_volume" valid:"required"`
}
type Volume struct {
	VolumeType          string `json:"volume_type" bson:"volume_type"`
	VolumeSize          int64  `json:"volume_size" bson:"volume_size"`
	DeleteOnTermination bool   `json:"delete_on_termination" bson:"delete_on_termination"`
	Iops                int64  `json:"iops" bson:"iops"`
}
type Project struct {
	ProjectData Data `json:"data"`
}
type Data struct {
	Region string `json:"region"`
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

func checkClusterSize(cluster Cluster_Def, ctx utils.Context) error {
	for _, pools := range cluster.NodePools {
		if pools.NodeCount > 3 {
			return errors.New("Nodepool can't have more than 3 nodes")
		}
	}
	return nil
}
func GetProfile(profileId string, region string, token string, ctx utils.Context) (vault.AwsProfile, error) {
	data, err := vault.GetCredentialProfile("aws", profileId, token, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return vault.AwsProfile{}, err
	}
	awsProfile := vault.AwsProfile{}
	err = json.Unmarshal(data, &awsProfile)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return vault.AwsProfile{}, err
	}
	awsProfile.Profile.Region = region
	return awsProfile, nil

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
	/*
		err = checkClusterSize(cluster, ctx)
		if err != nil { //cluster size limit exceed
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	*/

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
func DeployCluster(cluster Cluster_Def, credentials vault.AwsCredentials, ctx utils.Context, companyId string, token string) (confError error) {
	publisher := utils.Notifier{}
	confError = publisher.Init_notifier()
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		return confError
	}
	aws := AWS{
		AccessKey: credentials.AccessKey,
		SecretKey: credentials.SecretKey,
		Region:    credentials.Region,
	}
	confError = aws.init()
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
	createdPools, confError := aws.createCluster(cluster, ctx, companyId, token)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		cluster.Status = "Cluster creation failed"
		confError = aws.CleanUp(cluster, ctx)
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
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return confError
	}
	utils.SendLog(companyId, "Cluster created successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)

	return nil
}
func FetchStatus(credentials vault.AwsProfile, projectId string, ctx utils.Context, companyId string, token string) (Cluster_Def, error) {

	cluster, err := GetCluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, err
	}
	//splits := strings.Split(credentials, ":")
	aws := AWS{
		AccessKey: credentials.Profile.AccessKey,
		SecretKey: credentials.Profile.SecretKey,
		Region:    credentials.Profile.Region,
	}
	err = aws.init()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, err
	}

	_, e := aws.fetchStatus(&cluster, ctx, companyId, token)
	if e != nil {

		ctx.SendLogs("Cluster model: Status - Failed to get lastest status "+e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, e
	}
	/*	err = UpdateCluster(c)
		if err != nil {
			beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
			return Cluster_Def{}, err
		}*/
	return cluster, nil
}
func TerminateCluster(cluster Cluster_Def, profile vault.AwsProfile, ctx utils.Context, companyId, token string) error {

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
		ctx.SendLogs("AwsClusterModel : "+text+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return errors.New(text)
	}

	aws := AWS{
		AccessKey: profile.Profile.AccessKey,
		SecretKey: profile.Profile.SecretKey,
		Region:    profile.Profile.Region,
	}

	cluster.Status = string(models.Terminating)
	utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	err = aws.init()
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

	flag := aws.terminateCluster(cluster, ctx, companyId)
	if flag {
		utils.SendLog(companyId, "Cluster termination failed: "+cluster.Name, "error", cluster.ProjectId)

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

	var flagcheck bool
	for {
		flagcheck = false
		_, err = aws.fetchStatus(&cluster, ctx, companyId, token)
		if err != nil {
			beego.Error(err)
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
	cluster.Status = "Cluster Created"
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
func GetAWSAmi(credentials vault.AwsProfile, amiId string, ctx utils.Context, token string) ([]*ec2.BlockDeviceMapping, error) {

	aws := AWS{
		AccessKey: credentials.Profile.AccessKey,
		SecretKey: credentials.Profile.SecretKey,
		Region:    credentials.Profile.Region,
	}
	err := aws.init()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	amis, e := aws.describeAmi(&amiId, ctx)
	if e != nil {
		ctx.SendLogs("Cluster model: Status - Failed to get ami details "+e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return nil, e
	}
	return amis, nil
}
func EnableScaling(credentials vault.AwsProfile, cluster Cluster_Def, ctx utils.Context, token string) error {

	aws := AWS{
		AccessKey: credentials.Profile.AccessKey,
		SecretKey: credentials.Profile.SecretKey,
		Region:    credentials.Profile.Region,
	}
	err := aws.init()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return err
	}

	e := aws.enableScaling(cluster, ctx, token)
	if e != nil {
		ctx.SendLogs("Cluster model: Status - Failed to enable  scaling"+e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return e
	}
	UpdateScalingStatus(&cluster)
	err = UpdateCluster(cluster, false, ctx)
	if e != nil {
		ctx.SendLogs("Cluster model: Status - Failed to enable  scaling"+e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}
func UpdateScalingStatus(cluster *Cluster_Def) {
	for _, pool := range cluster.NodePools {
		pool.Scaling.State = models.Created
	}
}
func CreateSSHkey(keyName string, credentials vault.AwsCredentials, token, teams, region string, ctx utils.Context) (keyMaterial string, err error) {

	keyMaterial, err = GenerateAWSKey(keyName, credentials, token, teams, region, ctx)
	if err != nil {
		return "", err
	}

	return keyMaterial, err
}

func DeleteSSHkey(keyName, token string, credentials vault.AwsCredentials, ctx utils.Context) error {

	err := DeleteAWSKey(keyName, token, credentials, ctx)
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

func GetRegions(ctx utils.Context) ([]models.Region, error) {

	regions, err := api_handler.GetAwsRegions()
	if err != nil {
		ctx.SendLogs("Cluster model: Status - Failed to get aws regions "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []models.Region{}, err
	}

	return regions, nil
}
func GetZones(credentials vault.AwsProfile, ctx utils.Context) ([]*string, error) {

	aws := AWS{
		AccessKey: credentials.Profile.AccessKey,
		SecretKey: credentials.Profile.SecretKey,
		Region:    credentials.Profile.Region,
	}
	err := aws.init()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	zones, e := aws.GetZones(ctx)
	if e != nil {
		ctx.SendLogs("Cluster model: Status - Failed to get aws regions "+e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return nil, e
	}
	//var zone []string
	/*for _,z := range zones{
		z := z[len(z)-1:]
		zone = append(zone,*z)
	}*/
	return zones, nil
}
func GetAllMachines() ([]string, error) {
	machines, err := api_handler.GetAwsMachines()
	if err != nil {
		return []string{}, nil
	}

	return machines, nil
}

func ValidateProfile(key, secret, region string, ctx utils.Context) error {

	aws := AWS{
		AccessKey: key,
		SecretKey: secret,
		Region:    region,
	}
	err := aws.init()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	err = aws.validateProfile(ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Status - Failed to get aws regions "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}
