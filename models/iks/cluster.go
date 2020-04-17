package iks

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/db"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/types"
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
	ProjectId        string        `json:"project_id" bson:"project_id" validate:"required"`
	Kube_Credentials interface{}   `json:"kube_credentials" bson:"kube_credentials"`
	Name             string        `json:"name" bson:"name" validate:"required"`
	Status           string        `json:"status" bson:"status" validate:"eq=New|eq=new"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud" validate:"eq=IKS|eq=iks"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools        []*NodePool   `json:"node_pools" bson:"node_pools" validate:"required,dive"`
	NetworkName      string        `json:"network_name" bson:"network_name" validate:"required"`
	PublicEndpoint   bool          `json:"disable_public_service_endpoint" bson:"disable_public_service_endpoint"`
	KubeVersion      string        `json:"kube_version" bson:"kube_version" validate:"required"`
	CompanyId        string        `json:"company_id" bson:"company_id"`
	TokenName        string        `json:"token_name" bson:"token_name"`
	VPCId            string        `json:"vpc_id" bson:"vpc_id" validate:"required"`
	IsAdvance        bool          `json:"is_advance" bson:"is_advance"`
	ResourceGroup    string        `json:"resource_group" bson:"resource_group"`
}
type NodePool struct {
	ID               bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name             string        `json:"name" bson:"name" valid:"required"`
	NodeCount        int           `json:"node_count" bson:"node_count" valid:"required,matches(^[0-9]+$)"`
	MachineType      string        `json:"machine_type" bson:"machine_type" valid:"required"`
	SubnetID         string        `json:"subnet_id" bson:"subnet_id"`
	AvailabilityZone string        `json:"availability_zone" bson:"availability_zone"`
}

type Project struct {
	ProjectData Data `json:"data"`
}
type Data struct {
	Region string `json:"region"`
}

type Regions struct {
	Name     string   `json:"Name"`
	Location string   `json:"Location"`
	Zones    []string `json:"Zones"`
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
func GetError(projectId, companyId string, ctx utils.Context) (err types.ClusterError, err1 error) {

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Cluster model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.ClusterError{}, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoClusterErrorCollection)
	err1 = c.Find(bson.M{"project_id": projectId, "company_id": companyId, "cloud": models.IKS}).One(&err)
	if err1 != nil {
		ctx.SendLogs("Cluster model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.ClusterError{}, err1
	}
	return err, nil
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
func CreateCluster(cluster Cluster_Def, ctx utils.Context) error {
	_, err := GetCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err == nil { //cluster found
		ctx.SendLogs("Cluster model: Create - Cluster  already exists in the database: "+cluster.Name, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("Cluster model: Create - Cluster  already exists in the database: " + cluster.Name)
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

	if oldCluster.Status == (models.Deploying) && update {
		ctx.SendLogs("cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("cluster is in deploying state")
	}
	if oldCluster.Status == (models.Terminating) && update {
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
func DeployCluster(cluster Cluster_Def, credentials vault.IBMCredentials, ctx utils.Context, companyId string, token string) types.CustomCPError {
	publisher := utils.Notifier{}
	publisher.Init_notifier()

	iks := GetIBM(credentials)

	cpError := iks.init(credentials.Region, ctx)
	if cpError != (types.CustomCPError{}) {

		utils.SendLog(companyId, cpError.Message, "error", cluster.ProjectId)
		utils.SendLog(companyId, cpError.Description, "error", cluster.ProjectId)
		utils.SendLog(companyId, "Cluster creation failed : "+cluster.Name, "error", cluster.ProjectId)

		cluster.Status = "Cluster Creation Failed"
		confError := UpdateCluster(cluster, false, ctx)

		if confError != nil {
			utils.SendLog(companyId, confError.Error(), "error", cluster.ProjectId)

		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpError
	}

	utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	cluster.Status = models.Deploying
	confError := UpdateCluster(cluster, false, ctx)
	if confError != nil {

		utils.SendLog(companyId, confError.Error(), "error", cluster.ProjectId)
		cpErr := ApiError(confError, "Error occurred while updating cluster status in database", 500)

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}

	cluster, cpError = iks.create(cluster, ctx, companyId, token)

	if cpError != (types.CustomCPError{}) {

		if cluster.ClusterId != "" {

			iks.terminateCluster(&cluster, ctx)

		}
		cluster.Status = "Cluster Creation Failed"
		confError := UpdateCluster(cluster, false, ctx)
		if confError != nil {

			utils.SendLog(companyId, confError.Error(), "error", cluster.ProjectId)
			cpErr := ApiError(confError, "Error occurred while updating cluster status in database", 500)
			publisher.Notify(cluster.ProjectId, "Status Available", ctx)
			return cpErr

		}
		utils.SendLog(companyId, "Cluster creation failed : "+cluster.Name, "error", cluster.ProjectId)
		ctx.SendLogs("Cluster creation failed", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpError
	}

	confError = ApplyAgent(credentials, token, ctx, cluster.Name, cluster.ResourceGroup)
	if confError != nil {

		utils.SendLog(companyId, confError.Error(), "error", cluster.ProjectId)

		cluster.Status = "Cluster Creation Failed"
		confError = UpdateCluster(cluster, false, ctx)
		if confError != nil {
			utils.SendLog(companyId, confError.Error(), "error", cluster.ProjectId)
		}

		cpErr := ApiError(confError, "Error occurred deploying agent", 500)

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}
	cluster.Status = "Cluster Created"

	confError = UpdateCluster(cluster, false, ctx)

	if confError != nil {

		utils.SendLog(companyId, confError.Error(), "error", cluster.ProjectId)
		cpErr := ApiError(confError, "Error occurred while updating cluster status in database", 500)

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}
	utils.SendLog(companyId, "Cluster Created Sccessfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)

	return types.CustomCPError{}
}
func FetchStatus(credentials vault.IBMProfile, projectId string, ctx utils.Context, companyId string, token string) ([]KubeWorkerPoolStatus, types.CustomCPError) {

	cluster, err := GetCluster(projectId, companyId, ctx)
	if err != nil {
		cpErr := ApiError(err, "Error occurred while getting cluster status in database", 500)
		return []KubeWorkerPoolStatus{}, cpErr
	}
	customErr, err := GetError(projectId, companyId, ctx)
	if err != nil {
		cpErr := ApiError(err, "Error occurred while getting cluster status in database", 500)
		return []KubeWorkerPoolStatus{}, cpErr
	}
	if customErr.Err != (types.CustomCPError{}) {
		return []KubeWorkerPoolStatus{}, customErr.Err
	}
	iks := GetIBM(credentials.Profile)

	cpErr := iks.init(credentials.Profile.Region, ctx)
	if cpErr != (types.CustomCPError{}) {
		return []KubeWorkerPoolStatus{}, cpErr
	}

	response, e := iks.fetchStatus(&cluster, ctx, companyId)
	if e != (types.CustomCPError{}) {

		ctx.SendLogs("Cluster model: Status - Failed to get lastest status "+e.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []KubeWorkerPoolStatus{}, e
	}
	return response, types.CustomCPError{}
}
func TerminateCluster(cluster Cluster_Def, profile vault.IBMProfile, ctx utils.Context, companyId, token string) types.CustomCPError {

	publisher := utils.Notifier{}
	publisher.Init_notifier()

	iks := GetIBM(profile.Profile)

	cluster.Status = (models.Terminating)
	utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	err_ := UpdateCluster(cluster, false, ctx)
	if err_ != nil {

		utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.ProjectId)
		cpErr := types.CustomCPError{Description: err_.Error(), Message: "Error occurred while updating cluster status in database", StatusCode: 500}

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}
	cpErr := iks.init(profile.Profile.Region, ctx)
	if cpErr != (types.CustomCPError{}) {

		utils.SendLog(companyId, cpErr.Message, "error", cluster.ProjectId)
		utils.SendLog(companyId, cpErr.Description, "error", cluster.ProjectId)

		cluster.Status = "Cluster Termination Failed"
		err := UpdateCluster(cluster, false, ctx)
		if err != nil {
			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}

	cpErr = iks.terminateCluster(&cluster, ctx)
	if cpErr != (types.CustomCPError{}) {

		utils.SendLog(companyId, "Cluster termination failed: "+cpErr.Description+cluster.Name, "error", cluster.ProjectId)

		cluster.Status = "Cluster Termination Failed"
		err := UpdateCluster(cluster, false, ctx)
		if err != nil {
			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)

		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}

	cluster.Status = "Cluster Terminated"
	err := UpdateCluster(cluster, false, ctx)
	if err != nil {
		utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
		utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)

	}
	utils.SendLog(companyId, "Cluster terminated successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	return types.CustomCPError{}
}
func GetAllMachines(profile vault.IBMProfile, ctx utils.Context) (AllInstancesResponse, types.CustomCPError) {
	iks := GetIBM(profile.Profile)

	err := iks.init(profile.Profile.Region, ctx)
	if err != (types.CustomCPError{}) {
		return AllInstancesResponse{}, err
	}

	machineTypes, err := iks.GetAllInstances(ctx)
	if err != (types.CustomCPError{}) {
		return AllInstancesResponse{}, err
	}

	return machineTypes, types.CustomCPError{}
}
func GetRegions(ctx utils.Context) ([]Regions, error) {
	regionsDetails := []byte(`[
    {
      "Name": "Dallas",
      "Location": "us-south",
      "Zones": [
        "us-south-1",
        "us-south-2",
        "us-south-3"
      ]
    },
    {
      "Name": "Washington DC",
      "Location": "us-east",
      "Zones": [
        "us-east-1",
        "us-east-2",
        "us-east-3"
      ]
    },
    {
      "Name": "Frankfurt",
      "Location": "eu-de",
      "Zones": [
        "eu-de-1",
        "eu-de-2",
        "eu-de-3"
      ]
    },
    {
      "Name": "Tokyo",
      "Location": "jp-tok",
      "Zones": [
        "jp-tok-1",
        "jp-tok-2",
        "jp-tok-3"
      ]
    },
    {
      "Name": "London",
      "Location": "eu-gb",
      "Zones": [
        "eu-gb-1",
        "eu-gb-2",
        "eu-gb-3"
      ]
    },
    {
      "Name": "Sydney",
      "Location": "au-syd",
      "Zones": [
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
func GetAllVersions(profile vault.IBMProfile, ctx utils.Context) (Versions, types.CustomCPError) {
	iks := GetIBM(profile.Profile)

	err := iks.init(profile.Profile.Region, ctx)
	if err != (types.CustomCPError{}) {
		ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		ctx.SendLogs(err.Message, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Versions{}, err
	}

	versions, err := iks.GetAllVersions(ctx)
	if err != (types.CustomCPError{}) {
		ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		ctx.SendLogs(err.Message, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Versions{}, err
	}

	return versions, types.CustomCPError{}
}
func ApplyAgent(credentials vault.IBMCredentials, token string, ctx utils.Context, clusterName, resourceGroup string) (confError error) {
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

	cmd = "sudo docker run --rm --name " + companyId + projetcID + " -e resourceGroup=" + resourceGroup + " -e apikey=" + credentials.IAMKey + " -e cluster=" + clusterName + " -e yamlFile=" + filePath + "agent.yaml -v " + filePath + ":" + filePath + " " + models.IBMKSAuthContainerName

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
func ValidateProfile(profile vault.IBMProfile, ctx utils.Context) types.CustomCPError {
	iks := GetIBM(profile.Profile)

	err := iks.init(profile.Profile.Region, ctx)
	if err != (types.CustomCPError{}) {
		return err
	}

	_, err = iks.GetAllVersions(ctx)
	if err != (types.CustomCPError{}) {
		return err
	}

	return types.CustomCPError{}
}

func ValidateIKSData(cluster Cluster_Def, ctx utils.Context) error {

	if cluster.ProjectId == "" {

		return errors.New("project id is empty")

	} else if cluster.Name == "" {

		return errors.New("cluster name is empty")

	} else if cluster.KubeVersion == "" {

		return errors.New("kubernetes version is empty")

	} else if cluster.NetworkName == "" {

		return errors.New("network name is empty")

	} else if cluster.VPCId == "" {

		return errors.New("VPC name is empty")

	} else if len(cluster.NodePools) == 0 {

		return errors.New("node pool length must be greater than zero")

	} else if len(cluster.NodePools) > 0 {

		for _, nodepool := range cluster.NodePools {

			if nodepool.Name == "" {

				return errors.New("node pool name is empty")

			} else if nodepool.NodeCount == 0 {

				return errors.New("machine count must be greater than zero")

			} else if nodepool.MachineType == "" {

				return errors.New("machine type is empty")

			} else if nodepool.SubnetID == "" {

				return errors.New("subnet Id is empty")

			} else if nodepool.AvailabilityZone == "" {

				return errors.New("availability zone is empty")

			}

		}

		isZoneExist, err := validateIKSZone(cluster.NodePools[0].AvailabilityZone, ctx)
		if err != nil && !isZoneExist {
			text := "availabe zones are " + err.Error()
			ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return errors.New(text)
		}
	}

	return nil
}

func validateIKSZone(zone string, ctx utils.Context) (bool, error) {

	regionList, err := GetRegions(ctx)
	if err != nil {
		return false, err
	}

	for _, v1 := range regionList {
		for _, v2 := range v1.Zones {
			if zone == v2 {
				return true, nil
			}
		}
	}

	var errData string
	for _, v1 := range regionList {
		for _, v2 := range v1.Zones {
			errData += v2 + ", "
		}
	}

	return false, errors.New(errData)
}
