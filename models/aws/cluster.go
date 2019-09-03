package aws

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
}

type NodePool struct {
	ID                 bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name               string        `json:"name" bson:"name" valid:"required"`
	NodeCount          int64         `json:"node_count" bson:"node_count" valid:"required,matches(^[0-9]+$)"`
	MachineType        string        `json:"machine_type" bson:"machine_type" valid:"required"`
	Ami                Ami           `json:"ami" bson:"ami"`
	PoolSubnet         string        `json:"subnet_id" bson:"subnet_id" valid:"required"`
	PoolSecurityGroups []*string     `json:"security_group_id" bson:"security_group_id" valid:"required"`
	Nodes              []*Node       `json:"nodes" bson:"nodes"`
	KeyInfo            Key           `json:"key_info" bson:"key_info"`
	PoolRole           string        `json:"pool_role" bson:"pool_role" valid:"required"`
	EnableScaling      bool          `json:"enable_scaling" bson:"enable_scaling"`
	Scaling            AutoScaling   `json:"auto_scaling" bson:"auto_scaling"`
	IsExternal         bool          `json:"is_external" bson:"is_external"`
	ExternalVolume     Volume        `json:"external_volume" bson:"external_volume"`
}
type AutoScaling struct {
	MaxScalingGroupSize int64 `json:"max_scaling_group_size" bson:"max_scaling_group_size"`
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
type Key struct {
	KeyName     string         `json:"key_name" bson:"key_name" valid:"required"`
	KeyType     models.KeyType `json:"key_type" bson:"key_type" valid:"required, in(new|cp|aws|user)"`
	KeyMaterial string         `json:"private_key" bson:"private_key"`
	Cloud       models.Cloud   `json:"cloud" bson:"cloud"`
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

func GetNetwork(token, projectId string, ctx utils.Context) error {

	url := getNetworkHost("aws", projectId)

	_, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}
func CreateCluster(cluster Cluster_Def, ctx utils.Context) error {
	_, err := GetCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err == nil { //cluster found
		ctx.SendLogs("Cluster model: Create - Cluster  already exists in the database: "+cluster.Name, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("Cluster model: Create - Cluster  already exists in the database: " + cluster.Name)
	}
	err = checkClusterSize(cluster, ctx)
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
	ctx.SendLogs(" AWS Cluster: "+cluster.Name+" of Project Id: "+cluster.ProjectId+" created ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	return nil
}

func GetCluster(projectId, companyId string, ctx utils.Context) (cluster Cluster_Def, err error) {

	session, err1 := db.GetMongoSession()
	if err1 != nil {
		ctx.SendLogs("Cluster model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
	ctx.SendLogs(" Get AWS Cluster "+cluster.Name+" of Project Id: "+cluster.ProjectId+"", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	return cluster, nil
}

func GetAllCluster(ctx utils.Context, input rbac_athentication.List) (clusters []Cluster_Def, err error) {
	beego.Info("mongo session")
	session, err1 := db.GetMongoSession()
	if err1 != nil {
		ctx.SendLogs("Cluster model: GetAll - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return nil, err1
	}
	defer session.Close()
	mc := db.GetMongoConf()
	beego.Info("cluster aws")
	c := session.DB(mc.MongoDb).C(mc.MongoAwsClusterCollection)
	err = c.Find(bson.M{}).All(&clusters)
	beego.Info("getting all clusters")
	if err != nil {
		ctx.SendLogs("Cluster model: GetAll - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	ctx.SendLogs(" Get all AWS Cluster ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	return clusters, nil
}

func UpdateCluster(cluster Cluster_Def, update bool, ctx utils.Context) error {
	oldCluster, err := GetCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Update - Cluster   does not exist in the database: "+cluster.Name+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	if oldCluster.Status == "Cluster Created" && update {
		ctx.SendLogs("Cluster is in runnning state ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("Cluster is in runnning state")
	}
	err = DeleteCluster(cluster.ProjectId, ctx)
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

	ctx.SendLogs(" AWS Cluster "+cluster.Name+" of Project Id: "+cluster.ProjectId+"updated in database ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	return nil
}

func DeleteCluster(projectId string, ctx utils.Context) error {
	session, err := db.GetMongoSession()
	if err != nil {

		ctx.SendLogs("Cluster model: Delete - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAwsClusterCollection)
	err = c.Remove(bson.M{"project_id": projectId})
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	ctx.SendLogs(" Aws Cluster of Project Id: "+projectId+"deleted from database ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
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

	aws := AWS{
		AccessKey: credentials.AccessKey,
		SecretKey: credentials.SecretKey,
		Region:    credentials.Region,
	}
	confError = aws.init()
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		return confError
	}

	publisher := utils.Notifier{}
	confError = publisher.Init_notifier()
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		return confError
	}

	utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	createdPools, confError := aws.createCluster(cluster, ctx, companyId, token)

	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)

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

	confError = UpdateCluster(cluster, false, ctx)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return confError
	}
	utils.SendLog(companyId, "Cluster created successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	ctx.SendLogs(" AWS Cluster "+cluster.Name+" of Project Id: "+cluster.ProjectId+"deployed ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
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

	c, e := aws.fetchStatus(cluster, ctx, companyId, token)
	if e != nil {

		ctx.SendLogs("Cluster model: Status - Failed to get lastest status "+e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, e
	}
	ctx.SendLogs(" AWS Cluster "+cluster.Name+" of Project Id: "+projectId+"fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	/*	err = UpdateCluster(c)
		if err != nil {
			beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
			return Cluster_Def{}, err
		}*/
	return c, nil
}
func TerminateCluster(cluster Cluster_Def, profile vault.AwsProfile, ctx utils.Context, companyId string) error {

	aws := AWS{
		AccessKey: profile.Profile.AccessKey,
		SecretKey: profile.Profile.SecretKey,
		Region:    profile.Profile.Region,
	}
	err := aws.init()
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return err
	}

	publisher := utils.Notifier{}
	pub_err := publisher.Init_notifier()
	if pub_err != nil {
		ctx.SendLogs(pub_err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return pub_err
	}

	if cluster.Status != "Cluster Created" && cluster.Status == "Cluster Termination Failed" {
		ctx.SendLogs("Cluster model: Cluster is not in created state ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}

	utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	err = aws.terminateCluster(cluster, ctx, companyId)

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

	for _, pools := range cluster.NodePools {
		var nodes []*Node
		pools.Nodes = nodes
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
	ctx.SendLogs("Cluster "+cluster.Name+" of Project Id: "+cluster.ProjectId+"terminated by ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
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
			ctx.SendLogs("Pool "+nodepool.Name+" created ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

		}

		ctx.SendLogs("Cluster model: updated nodes in pools", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		cluster.NodePools[index].Nodes = updatedNodes
	}
	cluster.Status = "Cluster Created"
	return cluster
}
func GetAllSSHKeyPair(ctx utils.Context, token string) (keys []string, err error) {

	keys, err = vault.GetAllSSHKey("aws", ctx, token)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
	err = c.Find(bson.M{"cloud": models.AWS, "key_name": keyname}).One(&keys)
	if err != nil {
		return keys, err
	}
	return keys, nil
}
func InsertSSHKeyPair(key Key) (err error) {
	key.Cloud = models.AWS
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

	ctx.SendLogs("Cluster: "+cluster.Name+" scaled", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	return nil
}

func GetSSHkey(keyName, userName, token, teams string, ctx utils.Context) (privateKey string, err error) {

	privateKey, err = key_utils.GenerateKey(models.AWS, keyName, userName, token, teams, ctx)

	if err != nil {

		return "", err
	}
	return privateKey, err
}
