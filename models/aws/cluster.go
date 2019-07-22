package aws

import (
	"antelope/models"
	"antelope/models/db"
	"antelope/models/logging"
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
	NodePools        []*NodePool   `json:"node_pools" bson:"node_pools"`
	NetworkName      string        `json:"network_name" bson:"network_name" valid:"required"`
}

type NodePool struct {
	ID                 bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name               string        `json:"name" bson:"name" valid:"required"`
	NodeCount          int64         `json:"node_count" bson:"node_count" valid:"required"`
	MachineType        string        `json:"machine_type" bson:"machine_type" valid:"required"`
	Ami                Ami           `json:"ami" bson:"ami"`
	PoolSubnet         string        `json:"subnet_id" bson:"subnet_id" valid:"required"`
	PoolSecurityGroups []*string     `json:"security_group_id" bson:"security_group_id" valid:"required"`
	Nodes              []*Node       `json:"nodes" bson:"nodes"`
	KeyInfo            Key           `json:"key_info" bson:"key_info"`
	PoolRole           string        `json:"pool_role" bson:"pool_role" valid:"required"`
	EnableScaling      bool          `json:"enable_scaling" bson:"enable_scaling"`
	Scaling            AutoScaling   `json:"auto_scaling" bson:"auto_scaling"`
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
	KeyType     models.KeyType `json:"key_type" bson:"key_type" valid:"required in(new|cp|aws|user)"`
	KeyMaterial string         `json:"private_key" bson:"private_key"`
	Cloud       models.Cloud   `json:"cloud" bson:"cloud"`
}
type Ami struct {
	ID       bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name     string        `json:"name" bson:"name" valid:"required"`
	AmiId    string        `json:"ami_id" bson:"ami_id" valid:"required"`
	Username string        `json:"username" bson:"username" valid:"required"`

	RootVolume     Volume `json:"root_volume" bson:"root_volume" valid:"required"`
	IsExternal     bool   `json:"is_external" bson:"is_external"`
	ExternalVolume Volume `json:"external_volume" bson:"external_volume"`
}
type Volume struct {
	VolumeType          string `json:"volume_type" bson:"volume_type"`
	VolumeSize          int64  `json:"volume_size" bson:"volume_size"`
	DeleteOnTermination bool   `json:"delete_on_termination" bson:"delete_on_termination"`
	Iops                int64  `json:"iops" bson:"iops"`
}

func checkClusterSize(cluster Cluster_Def, ctx logging.Context) error {
	for _, pools := range cluster.NodePools {
		if pools.NodeCount > 3 {
			return errors.New("Nodepool can't have more than 3 nodes")
		}
	}
	return nil
}
func GetProfile(profileId string, region string, ctx logging.Context) (vault.AwsProfile, error) {
	data, err := vault.GetCredentialProfile("aws", profileId, ctx)
	awsProfile := vault.AwsProfile{}
	err = json.Unmarshal(data, &awsProfile)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return vault.AwsProfile{}, err
	}
	awsProfile.Profile.Region = region
	return awsProfile, nil

}
func GetRegion(projectId string, ctx logging.Context) (string, error) {
	url := beego.AppConfig.String("raccon_host") + "/" + projectId

	data, err := utils.GetAPIStatus(url, ctx)
	region := ""
	err = json.Unmarshal(data.([]byte), &region)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return region, err
	}
	return region, nil

}
func CreateCluster(cluster Cluster_Def, ctx logging.Context) error {
	_, err := GetCluster(cluster.ProjectId, ctx)
	if err == nil { //cluster found
		ctx.SendSDLog("Cluster model: Create - Cluster  already exists in the database: "+cluster.Name, "error")
		return errors.New("Cluster model: Create - Cluster  already exists in the database: " + cluster.Name)
	}
	err = checkClusterSize(cluster, ctx)
	if err != nil { //cluster found
		ctx.SendSDLog(err.Error(), "error")
		return err
	}
	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoAwsClusterCollection, cluster)
	if err != nil {
		ctx.SendSDLog("Cluster model: Create - Got error inserting cluster to the database: "+err.Error(), "error")
		return err
	}

	return nil
}

func GetCluster(projectId string, ctx logging.Context) (cluster Cluster_Def, err error) {

	session, err1 := db.GetMongoSession()
	if err1 != nil {
		ctx.SendSDLog("Cluster model: Get - Got error while connecting to the database: "+err.Error(), "error")
		return Cluster_Def{}, err1
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAwsClusterCollection)
	err = c.Find(bson.M{"project_id": projectId}).One(&cluster)
	if err != nil {
		ctx.SendSDLog("Cluster model: Get - Got error while connecting to the database: "+err.Error(), "error")
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
	c := session.DB(mc.MongoDb).C(mc.MongoAwsClusterCollection)
	err = c.Find(bson.M{}).All(&clusters)
	if err != nil {
		ctx.SendSDLog("Cluster model: GetAll - Got error while connecting to the database: "+err1.Error(), "error")
		return nil, err
	}
	return clusters, nil
}

func UpdateCluster(cluster Cluster_Def, update bool, ctx logging.Context) error {
	oldCluster, err := GetCluster(cluster.ProjectId, ctx)
	if err != nil {
		ctx.SendSDLog("Cluster model: Update - Cluster   does not exist in the database: "+cluster.Name+err.Error(), "error")
		return err
	}
	if oldCluster.Status == "Cluster Created" && update {
		ctx.SendSDLog("Cluster is in runnning state ", "error")
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
		ctx.SendSDLog("Cluster model: Update - Got error deleting cluster: "+err.Error(), "error")
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
	c := session.DB(mc.MongoDb).C(mc.MongoAwsClusterCollection)
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
func DeployCluster(cluster Cluster_Def, credentials vault.AwsCredentials, ctx logging.Context) (confError error) {

	aws := AWS{
		AccessKey: credentials.AccessKey,
		SecretKey: credentials.SecretKey,
		Region:    credentials.Region,
	}
	confError = aws.init()
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
	createdPools, confError := aws.createCluster(cluster, ctx)

	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx)

		confError = aws.CleanUp(cluster, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, ctx)
		}

		cluster.Status = "Cluster Creation Failed"
		confError = UpdateCluster(cluster, false, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, ctx)
		}
		publisher.Notify(cluster.ProjectId, "Status Available")
		return confError
	}

	cluster = updateNodePool(createdPools, cluster, ctx)

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
func FetchStatus(credentials vault.AwsProfile, projectId string, ctx logging.Context) (Cluster_Def, error) {

	cluster, err := GetCluster(projectId, ctx)
	if err != nil {
		ctx.SendSDLog("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), "error")
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
		ctx.SendSDLog(err.Error(), "error")
		return Cluster_Def{}, err
	}

	c, e := aws.fetchStatus(cluster, ctx)
	if e != nil {
		ctx.SendSDLog("Cluster model: Status - Failed to get lastest status "+e.Error(), "error")
		return Cluster_Def{}, e
	}
	/*	err = UpdateCluster(c)
		if err != nil {
			beego.Error("Cluster model: Deploy - Got error while connecting to the database: ", err.Error())
			return Cluster_Def{}, err
		}*/
	return c, nil
}
func TerminateCluster(cluster Cluster_Def, profile vault.AwsProfile, ctx logging.Context) error {

	aws := AWS{
		AccessKey: profile.Profile.AccessKey,
		SecretKey: profile.Profile.SecretKey,
		Region:    profile.Profile.Region,
	}
	err := aws.init()
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		beego.Error(err.Error())
		return err
	}

	publisher := utils.Notifier{}
	pub_err := publisher.Init_notifier()
	if pub_err != nil {
		ctx.SendSDLog(pub_err.Error(), "error")
		return pub_err
	}

	if cluster.Status != "Cluster Created" && cluster.Status == "Cluster Termination Failed" {
		ctx.SendSDLog("Cluster model: Cluster is not in created state ", "error")
		publisher.Notify(cluster.ProjectId, "Status Available")
		return err
	}

	logging.SendLog("Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	err = aws.terminateCluster(cluster, ctx)

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

	for _, pools := range cluster.NodePools {
		var nodes []*Node
		pools.Nodes = nodes
	}
	cluster.Status = "Cluster Terminated"
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
func updateNodePool(createdPools []CreatedPool, cluster Cluster_Def, ctx logging.Context) Cluster_Def {
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
		ctx.SendSDLog("Cluster model: updated nodes in pools", "info")
		cluster.NodePools[index].Nodes = updatedNodes
	}
	cluster.Status = "Cluster Created"
	return cluster
}
func GetAllSSHKeyPair(ctx logging.Context) (keys []string, err error) {

	keys, err = vault.GetAllSSHKey("aws", ctx)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
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
func GetAWSAmi(credentials vault.AwsProfile, amiId string, ctx logging.Context) ([]*ec2.BlockDeviceMapping, error) {

	aws := AWS{
		AccessKey: credentials.Profile.AccessKey,
		SecretKey: credentials.Profile.SecretKey,
		Region:    credentials.Profile.Region,
	}
	err := aws.init()
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return nil, err
	}

	amis, e := aws.describeAmi(&amiId, ctx)
	if e != nil {
		ctx.SendSDLog("Cluster model: Status - Failed to get ami details "+e.Error(), "error")
		return nil, e
	}
	return amis, nil
}
func EnableScaling(credentials string, cluster Cluster_Def, ctx logging.Context) error {

	splits := strings.Split(credentials, ":")
	aws := AWS{
		AccessKey: splits[0],
		SecretKey: splits[1],
		Region:    splits[2],
	}
	err := aws.init()
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return err
	}

	e := aws.enableScaling(cluster, ctx)
	if e != nil {
		ctx.SendSDLog("Cluster model: Status - Failed to enable  scaling"+e.Error(), "error")
		return e
	}
	return nil
}
