package iks

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/db"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"antelope/models/vault"
	"antelope/models/woodpecker"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

type Cluster_Def struct {
	ID               bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	ClusterId        string        `json:"cluster_id" bson:"cluster_id,omitempty"`
	ProjectId        string        `json:"project_id" bson:"project_id" valid:"required"`
	Kube_Credentials interface{}   `json:"kube_credentials" bson:"kube_credentials"`
	Name             string        `json:"name" bson:"name" valid:"required"`
	Status           string        `json:"status" bson:"status" valid:"in(New|new)"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud" valid:"in(DO|do)"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools        []*NodePool   `json:"node_pools" bson:"node_pools" valid:"required"`
	NetworkName      string        `json:"network_name" bson:"network_name" valid:"required"`
	PublicEndpoint   bool          `json:"disable_public_service_endpoint" bson:"disable_public_service_endpoint"`
	KubeVersion      string        `json:"kube_version" bson:"kube_version"`
	CompanyId        string        `json:"company_id" bson:"company_id"`
	TokenName        string        `json:"token_name" bson:"token_name"`
	VPCId            string        `json:"vpc_id" bson:"vpc_id"`
	ResourceGroup    string        `json:"resource_group" bson:"resource_group"`
}
type NodePool struct {
	ID          bson.ObjectId   `json:"_id" bson:"_id,omitempty"`
	Name        string          `json:"name" bson:"name" valid:"required"`
	NodeCount   int             `json:"node_count" bson:"node_count" valid:"required,matches(^[0-9]+$)"`
	MachineType string          `json:"machine_type" bson:"machine_type" valid:"required"`
	PoolRole    models.PoolRole `json:"pool_role" bson:"pool_role" valid:"required"`
	SubnetID    string          `json:"subnet_id" bson:"subnet_id"`
}

type Project struct {
	ProjectData Data `json:"data"`
}
type Data struct {
	Region string `json:"region"`
}

type Regions struct {
	Name  string   `json:"Name"`
	Location string   `json:"Location"`
	Zones []string `json:"zones"`
}

func getNetworkHost(cloudType, projectId string) string {

	host := beego.AppConfig.String("network_url") + models.WeaselGetEndpoint

	if strings.Contains(host, "{cloud}") {
		host = strings.Replace(host, "{cloud}", cloudType, -1)
	}

	if strings.Contains(host, "{projectId}") {
		host = strings.Replace(host, "{projectId}", projectId, -1)
	}

	return host
}

func GetProfile(profileId string, region string, token string, ctx utils.Context) (vault.IBMProfile, error) {
	data, err := vault.GetCredentialProfile("ibm", profileId, token, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return vault.IBMProfile{}, err
	}
	ibmProfile := vault.IBMProfile{}
	err = json.Unmarshal(data, &ibmProfile)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return vault.IBMProfile{}, err
	}
	ibmProfile.Profile.Region = region
	return ibmProfile, nil
}

func GetCluster(projectId, companyId string, ctx utils.Context) (cluster Cluster_Def, err error) {

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Cluster model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoIKSClusterCollection)
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
	c := session.DB(mc.MongoDb).C(mc.MongoIKSClusterCollection)
	err = c.Find(bson.M{}).All(&clusters)
	if err != nil {
		ctx.SendLogs("Cluster model: GetAll - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	return clusters, nil
}
func GetNetwork(token, projectId string, ctx utils.Context) error {

	url := getNetworkHost("ibm", projectId)

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

	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoIKSClusterCollection, cluster)
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
	c := session.DB(mc.MongoDb).C(mc.MongoIKSClusterCollection)
	err = c.Remove(bson.M{"project_id": projectId, "company_id": companyId})
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
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
func PrintError(confError error, name, projectId string, ctx utils.Context, companyId string) {
	if confError != nil {
		ctx.SendLogs(confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(companyId, "Cluster creation failed : "+name, "error", projectId)
		utils.SendLog(companyId, confError.Error(), "error", projectId)

	}
}
func DeployCluster(cluster Cluster_Def, credentials vault.IBMCredentials, ctx utils.Context, companyId string, token string) (confError error) {
	publisher := utils.Notifier{}
	confError = publisher.Init_notifier()
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		return confError
	}
	iks, err := GetIBM(credentials)
	if err != nil {
		return err
	}
	confError = iks.init(credentials.Region, ctx)
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
	cluster, confError = iks.create(cluster, ctx, companyId, token)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, ctx, companyId)
		confError = iks.terminateCluster(&cluster, ctx)
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
func FetchStatus(credentials vault.IBMProfile, projectId string, ctx utils.Context, companyId string, token string) (KubeClusterStatus, error) {

	cluster, err := GetCluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Deploy - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubeClusterStatus{}, err
	}
	iks, err := GetIBM(credentials.Profile)
	if err != nil {
		return KubeClusterStatus{}, err
	}
	err = iks.init(credentials.Profile.Region, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubeClusterStatus{}, err
	}

	response, e := iks.fetchStatus(&cluster, ctx, companyId)
	if e != nil {

		ctx.SendLogs("Cluster model: Status - Failed to get lastest status "+e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubeClusterStatus{}, e
	}
	return response, nil
}
func TerminateCluster(cluster Cluster_Def, profile vault.IBMProfile, ctx utils.Context, companyId, token string) error {

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
		ctx.SendLogs("IBMClusterModel : "+text+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return errors.New(text)
	}

	iks, err := GetIBM(profile.Profile)
	if err != nil {
		return err
	}

	cluster.Status = string(models.Terminating)
	utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	err = iks.init(profile.Profile.Region, ctx)
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

	err = iks.terminateCluster(&cluster, ctx)
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
func GetAllMachines(profile vault.IBMProfile, ctx utils.Context) (AllInstancesResponse, error) {
	iks, err := GetIBM(profile.Profile)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return AllInstancesResponse{}, err
	}

	err = iks.init(profile.Profile.Region, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return AllInstancesResponse{}, err
	}

	machineTypes, err := iks.GetAllInstances(ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return AllInstancesResponse{}, err
	}

	return machineTypes, nil
}
func GetRegions(ctx utils.Context) ([]Regions, error) {
	regionsDetails := []byte(`[
    {
      "Name": "Dallas",
      "Location": "us-south",
      "zones": [
        "us-south-1",
        "us-south-2",
        "us-south-3"
      ]
    },
    {
      "Name": "Washington DC",
      "Location": "us-east",
      "zones": [
        "us-east-1",
        "us-east-2",
        "us-east-3"
      ]
    },
    {
      "Name": "Frankfurt",
      "Location": "eu-de",
      "zones": [
        "eu-de-1",
        "eu-de-2",
        "eu-de-3"
      ]
    },
    {
      "Name": "Tokyo",
      "Location": "jp-tok",
      "zones": [
        "jp-tok-1",
        "jp-tok-2",
        "jp-tok-3"
      ]
    },
    {
      "Name": "London",
      "Location": "eu-gb",
      "zones": [
        "eu-gb-1",
        "eu-gb-2",
        "eu-gb-3"
      ]
    },
    {
      "Name": "Sydney",
      "Location": "au-syd",
      "zones": [
        "au-syd-1",
        "au-syd-2",
        "au-syd-3"
      ]
    }
  ]`)
	var regions []Regions
	err := json.Unmarshal(regionsDetails, &regions)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []Regions{}, err
	}
	return regions, nil
}
func GetAllVersions(profile vault.IBMProfile, ctx utils.Context) (Versions, error) {
	iks, err := GetIBM(profile.Profile)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Versions{}, err
	}

	err = iks.init(profile.Profile.Region, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Versions{}, err
	}

	machineTypes, err := iks.GetAllVersions(ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Versions{}, err
	}

	return machineTypes, nil
}
func ApplyAgent(credentials vault.IBMProfile, token string, ctx utils.Context, clusterName, resourceGroup string) (confError error) {
	companyId := ctx.Data.Company
	projetcID := ctx.Data.ProjectId
	data2, err := woodpecker.GetCertificate(projetcID, token, ctx)
	if err != nil {
		ctx.SendLogs("IKSKubernetesClusterController. : Apply Agent -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	filePath := "/tmp/" + companyId + "/" + projetcID + "/"
	cmd := "mkdir -p " + filePath + " && echo '" + data2 + "'>" + filePath + "agent.yaml"
	output, err := models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("IKSKubernetesClusterController. : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	cmd = "sudo docker run --rm --name " + companyId + projetcID + " -e resourceGroup=" + resourceGroup + " -e apikey=" + credentials.Profile.IAMKey + " -e cluster=" + clusterName + " -e yamlFile=" + filePath + "agent.yaml -v " + filePath + ":" + filePath + " " + models.IBMKSAuthContainerName

	output, err = models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("IKSKubernetesClusterController. : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func GeZones(region string, ctx utils.Context) ([]string, error) {
	var zones []string
	if region == "us-south" {
		zones = append(zones, "us-south-1")
		zones = append(zones, "us-south-2")
		zones = append(zones, "us-south-3")
	} else if region == "us-east" {
		zones = append(zones, "us-east-1")
		zones = append(zones, "us-east-2")
		zones = append(zones, "us-east-3")
	} else if region == "eu-de" {
		zones = append(zones, "eu-de-1")
		zones = append(zones, "eu-de-2")
		zones = append(zones, "eu-de-3")
	} else if region == "jp-tok" {
		zones = append(zones, "jp-tok-1")
		zones = append(zones, "jp-tok-2")
		zones = append(zones, "jp-tok-3")
	} else if region == "eu-gb" {
		zones = append(zones, "eu-gb-1")
		zones = append(zones, "eu-gb-2")
		zones = append(zones, "eu-gb-3")
	} else if region == "au-syd" {
		zones = append(zones, "au-syd-1")
		zones = append(zones, "au-syd-2")
		zones = append(zones, "au-syd-3")
	}

	return zones, nil
}
