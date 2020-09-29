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
	"github.com/r3labs/diff"
	"gopkg.in/mgo.v2/bson"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Cluster_Def struct {
	ID               bson.ObjectId `json:"-" bson:"_id,omitempty"`
	ClusterId        string        `json:"cluster_id" bson:"cluster_id,omitempty"`
	InfraId          string        `json:"infra_id" bson:"infra_id" validate:"required" description:"ID of infrastructure [required]"`
	Kube_Credentials interface{}   `json:"-" bson:"kube_credentials"`
	Name             string        `json:"name" bson:"name" validate:"required" description:"Cluster name [required]"`
	Status           models.Type   `json:"status" bson:"status" validate:"eq=new|eq=New|eq=NEW|eq=Cluster Creation Failed|eq=Cluster Terminated|eq=Cluster Created" description:"Status of cluster [required]"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud" validate:"eq=IKS|eq=iks"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools        []*NodePool   `json:"node_pools" bson:"node_pools" validate:"required,dive"`
	NetworkName      string        `json:"network_name" bson:"network_name" validate:"required" description:"Network name in which cluster will be provisioned [required]"`
	PublicEndpoint   bool          `json:"disable_public_service_endpoint" bson:"disable_public_service_endpoint" description:"[optional]"`
	KubeVersion      string        `json:"kube_version" bson:"kube_version" description:"Kubernetes version to be provisioned [optional]"`
	CompanyId        string        `json:"company_id" bson:"company_id" description:"ID of compnay [optional]"`
	TokenName        string        `json:"-" bson:"token_name"`
	VPCId            string        `json:"vpc_id" bson:"vpc_id" validate:"required" description:"Virtual private cloud ID in which cluster will be provisioned [required]"`
	IsAdvance        bool          `json:"is_advance" bson:"is_advance"`
	ResourceGroup    string        `json:"resource_group" bson:"resource_group" validate:"required" description:"Resources would be created within resource_group [required]"`
}
type NodePool struct {
	ID               bson.ObjectId `json:"-" bson:"_id,omitempty"`
	Name             string        `json:"name" bson:"name" validate:"required" description:"Cluster pool name [required]"`
	NodeCount        int           `json:"node_count" bson:"node_count" validate:"required,gte=1" description:"Pool node count [required]"`
	MachineType      string        `json:"machine_type" bson:"machine_type" validate:"required" description:"Machine type for pool [required]"`
	SubnetID         string        `json:"subnet_id" bson:"subnet_id" validate:"required" description:"ID of subnet in which pool will be created [required]"`
	AvailabilityZone string        `json:"availability_zone" bson:"availability_zone" validate:"required"`
	Autoscaling      Autoscaling   `json:"auto_scaling,omitempty"  bson:"autoscaling,omitempty" description:"Autoscaling configuration [optional]"`
	PoolStatus       bool          `json:"pool_status,omitempty" bson:"pool_status,omitempty"`
	PoolId           string        `json:"pool_Id" bson:"pool_Id"  description:"Cluster pool id [optional]"`
}

type infrastructure struct {
	infrastructureData Data `json:"data"`
}
type Data struct {
	Region string `json:"region"`
}

type Regions struct {
	Name     string   `json:"Name"`
	Location string   `json:"Location"`
	Zones    []string `json:"Zones"`
}

type Cluster struct {
	Name    string      `json:"name,omitempty" bson:"name,omitempty" v description:"Cluster name"`
	InfraId string      `json:"infra_id" bson:"infra_id"  description:"ID of infrastructure"`
	Status  models.Type `json:"status,omitempty" bson:"status,omitempty" " description:"Status of cluster"`
}

func getNetworkHost(cloudType, infraId string) string {

	host := beego.AppConfig.String("network_url") + models.WeaselGetEndpoint

	if strings.Contains(host, "{cloud}") {
		host = strings.Replace(host, "{cloud}", cloudType, -1)
	}

	if strings.Contains(host, "{infraId}") {
		host = strings.Replace(host, "{infraId}", infraId, -1)
	}

	return host
}

func GetProfile(profileId string, region string, token string, ctx utils.Context) (int, vault.IBMProfile, error) {
	statusCode, data, err := vault.GetCredentialProfile("ibm", profileId, token, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return statusCode, vault.IBMProfile{}, err
	}
	ibmProfile := vault.IBMProfile{}
	err = json.Unmarshal(data, &ibmProfile)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 500, vault.IBMProfile{}, err
	}
	ibmProfile.Profile.Region = region
	return 0, ibmProfile, nil
}
func GetCluster(infraId, companyId string, ctx utils.Context) (cluster Cluster_Def, err error) {

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Cluster model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoIKSClusterCollection)
	err = c.Find(bson.M{"infra_id": infraId, "company_id": companyId}).One(&cluster)
	if err != nil {
		ctx.SendLogs("Cluster model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, err
	}
	return cluster, nil
}
func GetAllCluster(ctx utils.Context, input rbac_athentication.List) (iksClusters []Cluster, err error) {
	var clusters []Cluster_Def
	var copyData []string

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
	c := session.DB(mc.MongoDb).C(mc.MongoIKSClusterCollection)
	err = c.Find(bson.M{"infra_id": bson.M{"$in": copyData}, "company_id": ctx.Data.Company}).All(&clusters)
	if err != nil {
		ctx.SendLogs("Cluster model: GetAll - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	for _, cluster := range clusters {
		temp := Cluster{Name: cluster.Name, InfraId: cluster.InfraId, Status: cluster.Status}
		iksClusters = append(iksClusters, temp)
	}

	return iksClusters, nil
}
func GetNetwork(token, infraId string, ctx utils.Context) error {

	url := getNetworkHost("ibm", infraId)

	_, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}
func CreateCluster(cluster Cluster_Def, ctx utils.Context) error {
	_, err := GetCluster(cluster.InfraId, cluster.CompanyId, ctx)
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
	oldCluster, err := GetCluster(cluster.InfraId, cluster.CompanyId, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Update - Cluster   does not exist in the database: "+cluster.Name+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	err = DeleteCluster(cluster.InfraId, cluster.CompanyId, ctx)
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
func DeleteCluster(infraId, companyId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {

		ctx.SendLogs("Cluster model: Delete - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoIKSClusterCollection)
	err = c.Remove(bson.M{"infra_id": infraId, "company_id": companyId})
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}
func GetRegion(token, infraId string, ctx utils.Context) (string, error) {
	url := beego.AppConfig.String("raccoon_url") + models.InfraGetEndpoint
	if strings.Contains(url, "{infraId}") {
		url = strings.Replace(url, "{infraId}", infraId, -1)
	}
	data, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs("Error in fetching region"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}
	var region infrastructure
	err = json.Unmarshal(data.([]byte), &region.infrastructureData)
	if err != nil {
		ctx.SendLogs("Error in fetching region"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return region.infrastructureData.Region, err
	}
	return region.infrastructureData.Region, nil

}
func DeployCluster(cluster Cluster_Def, credentials vault.IBMCredentials, ctx utils.Context, companyId string, token string) types.CustomCPError {
	publisher := utils.Notifier{}
	publisher.Init_notifier()

	iks := GetIBM(credentials)

	cpError := iks.init(credentials.Region, ctx)
	if cpError != (types.CustomCPError{}) {

		utils.SendLog(companyId, cpError.Error, "error", cluster.InfraId)
		utils.SendLog(companyId, cpError.Description, "error", cluster.InfraId)
		utils.SendLog(companyId, "Cluster creation failed : "+cluster.Name, "error", cluster.InfraId)

		cluster.Status = models.ClusterCreationFailed
		confError := UpdateCluster(cluster, false, ctx)

		if confError != nil {
			utils.SendLog(companyId, confError.Error(), "error", cluster.InfraId)

		}
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpError)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: cpError.Error + "\n" + cpError.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return cpError
	}

	utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.InfraId)
	cluster.Status = (models.Deploying)
	confError := UpdateCluster(cluster, false, ctx)
	if confError != nil {

		utils.SendLog(companyId, confError.Error(), "error", cluster.InfraId)
		cpErr := ApiError(confError, "Error occurred while updating cluster status in database", 500)

		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: confError.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return cpErr
	}

	cluster, cpError = iks.create(cluster, ctx, companyId, token)

	if cpError != (types.CustomCPError{}) {

		if cluster.ClusterId != "" {

			iks.terminateCluster(&cluster, ctx)

		}
		cluster.Status = models.ClusterCreationFailed
		confError := UpdateCluster(cluster, false, ctx)
		if confError != nil {

			utils.SendLog(companyId, confError.Error(), "error", cluster.InfraId)
			cpErr := ApiError(confError, "Error occurred while updating cluster status in database", 500)
			err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpError)
			if err != nil {
				ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			utils.Publisher(utils.ResponseSchema{
				Status:  false,
				Message: cpError.Error + "\n" + cpError.Description,
				InfraId: cluster.InfraId,
				Token:   token,
				Action:  models.Create,
			}, ctx)
			return cpErr

		}
		utils.SendLog(companyId, "Cluster creation failed : "+cluster.Name, "error", cluster.InfraId)
		ctx.SendLogs("Cluster creation failed", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpError)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: cpError.Error + "\n" + cpError.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return cpError
	}

	pubSub := publisher.Subscribe(ctx.Data.InfraId, ctx)

	confError = ApplyAgent(credentials, token, ctx, cluster.Name, cluster.ResourceGroup)
	if confError != nil {
		utils.SendLog(companyId, confError.Error(), "error", cluster.InfraId)

		cluster.Status = models.ClusterCreationFailed
		profile := vault.IBMProfile{Profile: credentials}
		_ = TerminateCluster(cluster, profile, ctx, companyId, token)
		utils.SendLog(companyId, "Cleaning up resources", "info", cluster.InfraId)
		confError = UpdateCluster(cluster, false, ctx)
		if confError != nil {
			utils.SendLog(companyId, confError.Error(), "error", cluster.InfraId)
		}

		cpErr := ApiError(confError, "Error occurred while deploying agent", 500)
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: confError.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return cpErr
	}
	cluster.Status = models.ClusterCreated

	confError = UpdateCluster(cluster, false, ctx)

	if confError != nil {

		utils.SendLog(companyId, confError.Error(), "error", cluster.InfraId)
		cpErr := ApiError(confError, "Error occurred while updating cluster status in database", 500)
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpError)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: confError.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return cpErr
	}
	utils.SendLog(companyId, "Cluster Created Sccessfully "+cluster.Name, "info", cluster.InfraId)

	notify := publisher.RecieveNotification(ctx.Data.InfraId, ctx, pubSub)
	if notify {
		ctx.SendLogs("IKSClusterModel:  Notification recieved from agent", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Cluster created successfully",
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
	} else {
		ctx.SendLogs("IKSClusterModel:  Notification not recieved from agent", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		cluster.Status = models.ClusterCreationFailed
		utils.SendLog(ctx.Data.Company, "Notification not recieved from agent", models.LOGGING_LEVEL_INFO, cluster.InfraId)
		confError_ := UpdateCluster(cluster, false, ctx)
		if confError_ != nil {
			ctx.SendLogs("IKSDeployClusterModel:"+confError_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.GKE, ctx, types.CustomCPError{Description: confError_.Error(), Error: confError_.Error(), StatusCode: 512})
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Agent  - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Notification not recieved from agent",
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
	}

	return types.CustomCPError{}
}
func FetchStatus(credentials vault.IBMProfile, infraId string, ctx utils.Context, companyId string, token string) (KubeClusterStatus1, types.CustomCPError) {

	cluster, err := GetCluster(infraId, companyId, ctx)
	if err != nil {
		cpErr := ApiError(err, "Error occurred while getting cluster status in database", 500)
		return KubeClusterStatus1{}, cpErr
	}
	if string(cluster.Status) == strings.ToLower(string(models.New)) {
		cpErr := types.CustomCPError{Error: "Unable to fetch status - Cluster is not deployed yet", Description: "Unable to fetch state - Cluster is not deployed yet", StatusCode: 409}
		return KubeClusterStatus1{}, cpErr
	}

	if cluster.Status == models.Deploying || cluster.Status == models.Terminating || cluster.Status == models.ClusterTerminated {
		cpErr := ApiError(errors.New("Cluster is in "+
			string(cluster.Status)), "Cluster is in "+
			string(cluster.Status)+" state", 409)
		return KubeClusterStatus1{}, cpErr
	}
	if cluster.Status != models.ClusterCreated {
		customErr, err := db.GetError(infraId, companyId, models.IKS, ctx)
		if err != nil {
			cpErr := ApiError(err, "Error occurred while getting cluster status in database", 500)
			return KubeClusterStatus1{}, cpErr
		}
		if customErr.Err != (types.CustomCPError{}) {
			return KubeClusterStatus1{}, customErr.Err
		}
	}
	iks := GetIBM(credentials.Profile)

	cpErr := iks.init(credentials.Profile.Region, ctx)
	if cpErr != (types.CustomCPError{}) {
		return KubeClusterStatus1{}, cpErr
	}

	response, e := iks.fetchStatus(&cluster, ctx, companyId)

	if e != (types.CustomCPError{}) {

		ctx.SendLogs("Cluster model: Status - Failed to get lastest status "+e.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubeClusterStatus1{}, e
	}
	var response1 KubeClusterStatus1
	response1.ID = response.ID
	response1.Name = response.Name
	response1.Region = response.Region
	response1.ResourceGroup = response.ResourceGroupName
	response1.PoolCount = 0
	response1.KubernetesVersion = response.KubernetesVersion
	response1.State = response.State
	for _, pool := range response.WorkerPools {
		response1.PoolCount = response1.PoolCount + 1
		var pool1 KubeWorkerPoolStatus1
		pool1.Name = pool.Name
		pool1.ID = pool.ID
		pool1.Flavour = pool.Flavour
		for _, pool := range cluster.NodePools {
			if pool.Autoscaling.AutoScale != false || pool.Autoscaling != (Autoscaling{}) {
				pool1.Autoscaling.AutoScale = pool.Autoscaling.AutoScale
				pool1.Autoscaling.MaxNodes = pool.Autoscaling.MaxNodes
				pool1.Autoscaling.MinNodes = pool.Autoscaling.MinNodes
			}
		}

		pool1.Count = pool.Count
		pool1.SubnetId = pool.Nodes[0].NetworkInterfaces[0].SubnetId
		for _, node := range pool.Nodes {
			var node1 KubeWorkerNodesStatus1
			node1.State = node.Lifecycle.State
			node1.PoolId = node.PoolId
			node1.PrivateIp = node.NetworkInterfaces[0].IpAddress
			node1.PublicIp = node.NetworkInterfaces[0].IpAddress
			node1.Name = node.PoolId
			pool1.Nodes = append(pool1.Nodes, node1)
		}
		response1.WorkerPools = append(response1.WorkerPools, pool1)
	}
	response1.Status = cluster.Status
	return response1, types.CustomCPError{}
}
func TerminateCluster(cluster Cluster_Def, profile vault.IBMProfile, ctx utils.Context, companyId, token string) types.CustomCPError {

	publisher := utils.Notifier{}
	publisher.Init_notifier()

	iks := GetIBM(profile.Profile)

	_, _, _, err1 := CompareClusters(ctx)
	if err1 != nil  &&  !(strings.Contains(err1.Error(),"Nothing to update")){
		oldCluster, err_ :=  GetPreviousIKSCluster(ctx)
		if err_ != nil {
			utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.InfraId)
			cpErr := types.CustomCPError{Description: err_.Error(), Error: "Error occurred while updating cluster status in database", StatusCode: 500}
			err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpErr)
			if err != nil {
				ctx.SendLogs("IKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}

			utils.Publisher(utils.ResponseSchema{
				Status:  false,
				Message: "Cluster termination failed",
				InfraId: cluster.InfraId,
				Token:   token,
				Action:  models.Terminate,
			}, ctx)

			return cpErr

		}

		err_ = UpdateCluster(oldCluster, false, ctx)
		if err_ != nil {

			utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.InfraId)
			cpErr := types.CustomCPError{Description: err_.Error(), Error: "Error occurred while updating cluster status in database", StatusCode: 500}
			err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpErr)
			if err != nil {
				ctx.SendLogs("IKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}

			utils.Publisher(utils.ResponseSchema{
				Status:  false,
				Message: "Cluster termination failed",
				InfraId: cluster.InfraId,
				Token:   token,
				Action:  models.Terminate,
			}, ctx)

			return cpErr

		}
	}

	cluster.Status = (models.Terminating)
	utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.InfraId)

	err_ := UpdateCluster(cluster, false, ctx)
	if err_ != nil {

		utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.InfraId)
		cpErr := types.CustomCPError{Description: err_.Error(), Error: "Error occurred while updating cluster status in database", StatusCode: 500}
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Terminate Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster termination failed",
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Terminate,
		}, ctx)


		return cpErr
	}
	cpErr := iks.init(profile.Profile.Region, ctx)
	if cpErr != (types.CustomCPError{}) {

		utils.SendLog(companyId, cpErr.Error, "error", cluster.InfraId)
		utils.SendLog(companyId, cpErr.Description, "error", cluster.InfraId)

		cluster.Status = models.ClusterTerminationFailed
		err := UpdateCluster(cluster, false, ctx)
		if err != nil {
			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.InfraId)
			utils.SendLog(companyId, err.Error(), "error", cluster.InfraId)
		}
		err = db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Update Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster update failed",
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Update,
		}, ctx)

		return cpErr
	}

	cpErr = iks.terminateCluster(&cluster, ctx)
	if cpErr != (types.CustomCPError{}) {

		utils.SendLog(companyId, "Cluster termination failed: "+cpErr.Description+cluster.Name, "error", cluster.InfraId)

		cluster.Status = models.ClusterTerminationFailed
		err := UpdateCluster(cluster, false, ctx)
		if err != nil {
			utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.InfraId)
			utils.SendLog(companyId, err.Error(), "error", cluster.InfraId)

		}
		err = db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Update Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster update failed",
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Update,
		}, ctx)

		return cpErr
	}

	cluster.Status = models.ClusterTerminated
	err := UpdateCluster(cluster, false, ctx)
	if err != nil {
		utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.InfraId)
		utils.SendLog(companyId, err.Error(), "error", cluster.InfraId)

		cpErr := ApiError(err, "Error occurred while updating cluster status in database", 500)
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("IKSDeployClusterModel:  Deploy Cluster - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster update failed",
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Update,
		}, ctx)

		return cpErr

	}

	utils.SendLog(companyId, "Cluster terminated successfully "+cluster.Name, "info", cluster.InfraId)
	utils.Publisher(utils.ResponseSchema{
		Status:  true,
		Message: "Cluster terminated successfully",
		InfraId: cluster.InfraId,
		Token:   token,
		Action:  models.Terminate,
	}, ctx)


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
		ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Versions{}, err
	}

	versions, err := iks.GetAllVersions(ctx)
	if err != (types.CustomCPError{}) {
		ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		ctx.SendLogs(err.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Versions{}, err
	}

	var kubeVersions []Kubernetes
	for _, kube := range versions.Kubernetes {
		if kube.Major == 1 && kube.Minor == 15 && kube.Patch == 12 {
			continue
		} else {
			kubeVersions = append(kubeVersions, kube)
		}
	}

	sort.Slice(kubeVersions, func(i, j int) bool {

		return kubeVersions[i].Minor > kubeVersions[j].Minor

	})

	versions.Kubernetes = kubeVersions

	return versions, types.CustomCPError{}
}
func ApplyAgent(credentials vault.IBMCredentials, token string, ctx utils.Context, clusterName, resourceGroup string) (confError error) {
	companyId := ctx.Data.Company
	infraID := ctx.Data.InfraId
	data2, err := woodpecker.GetCertificate(infraID, token, ctx)
	if err != nil {
		ctx.SendLogs("IKSKubernetesClusterController. : Apply Agent -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	filePath := "/tmp/" + companyId + "/" + infraID + "/"
	cmd := "mkdir -p " + filePath + " && echo '" + data2 + "'>" + filePath + "agent.yaml"
	output, err := models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("IKSKubernetesClusterController. : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	cmd = "sudo docker run --rm --name " + companyId + infraID + " -e resourceGroup=" + resourceGroup + " -e apikey=" + credentials.IAMKey + " -e cluster=" + clusterName + " -e yamlFile=" + filePath + "agent.yaml -v " + filePath + ":" + filePath + " " + models.IBMKSAuthContainerName

	output, err = models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("IKSKubernetesClusterController. : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func GetZones(region string, ctx utils.Context) ([]string, error) {
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

	if cluster.InfraId == "" {

		return errors.New("infrastructure id is empty")

	} else if cluster.Name == "" {

		return errors.New("cluster name is empty")

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
func AddPreviousIKSCluster(cluster Cluster_Def, ctx utils.Context, patch bool) error {
	var oldCluster Cluster_Def
	_, err := GetPreviousIKSCluster(ctx)
	if err == nil {
		err := DeletePreviousIKSCluster(ctx)
		if err != nil {
			ctx.SendLogs(
				"IKSAddClusterModel:  Add previous cluster - "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
		}
	}

	if patch == false {
		oldCluster, err = GetCluster(ctx.Data.InfraId, ctx.Data.Company, ctx)
		if err != nil {
			ctx.SendLogs(
				"IKEAddClusterModel:  Add previous cluster - "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
		}
	} else {
		oldCluster = cluster
	}
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"IKEAddClusterModel:  Add previous cluster - "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	defer session.Close()

	if cluster.CreationDate.IsZero() {
		cluster.CreationDate = time.Now()
		cluster.ModificationDate = time.Now()
		cluster.Cloud = models.IKS
		cluster.CompanyId = ctx.Data.Company
	}

	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoIKSPreviousClusterCollection, oldCluster)
	if err != nil {
		ctx.SendLogs(
			"IKEAddClusterModel:  Add previous cluster -  "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}
func GetPreviousIKSCluster(ctx utils.Context) (cluster Cluster_Def, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs(
			"IKSGetClusterModel:  Get previous cluster - Got error while connecting to the database: "+err1.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoIKSPreviousClusterCollection)
	err = c.Find(bson.M{"infra_id": ctx.Data.InfraId, "company_id": ctx.Data.Company}).One(&cluster)
	if err != nil {
		ctx.SendLogs(
			"IKSGetClusterModel:  Get previous cluster- Got error while fetching from database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err
	}

	return cluster, nil
}
func UpdatePreviousIKSCluster(cluster Cluster_Def, ctx utils.Context) error {

	err := AddPreviousIKSCluster(cluster, ctx, false)
	if err != nil {
		text := "EKSClusterModel:  Update  previous cluster - " + cluster.Name + " " + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	err = UpdateCluster(cluster, false, ctx)
	if err != nil {
		text := "IKSClusterModel:  Update previous cluster - " + cluster.Name + " " + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		err = DeletePreviousIKSCluster(ctx)
		if err != nil {
			text := "IKSDeleteClusterModel:  Delete  previous cluster - " + cluster.Name + " " + err.Error()
			ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return errors.New(text)
		}
		return err
	}

	return nil
}
func DeletePreviousIKSCluster(ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"IKSDeleteClusterModel:  Delete  previous cluster - "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoIKSPreviousClusterCollection)
	err = c.Remove(bson.M{"infra_id": ctx.Data.InfraId, "company_id": ctx.Data.Company})
	if err != nil {
		ctx.SendLogs(
			"ISKDeleteClusterModel:  Delete  previous cluster - "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}
func PatchRunningIKSCluster(cluster Cluster_Def, credentials vault.IBMCredentials, token string, ctx utils.Context) (confError types.CustomCPError) {

	/*	publisher := utils.Notifier{}
		publisher.Init_notifier()*/

	iks := GetIBM(credentials)

	iks.init(credentials.Region, ctx)
	utils.SendLog(ctx.Data.Company, "Updating running cluster : "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

	difCluster, _, _, err1 := CompareClusters(ctx)
	if err1 != nil &&  !(strings.Contains(err1.Error(),"Nothing to update")){
		ctx.SendLogs("IKSUpdateRunningClusterModel:  Update - "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, err1.Error()+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

		if !strings.Contains(err1.Error(), "Nothing to update") {
			utils.SendLog(ctx.Data.Company, "Nothing to update"+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

			cluster.Status = models.ClusterCreated
			confError := UpdateCluster(cluster, false, ctx)
			if confError != nil {
				ctx.SendLogs("IKSpdateRunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			err := ApiError(err1, "Nothing to update", 512)
			err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.IKS, ctx, err)
			if err_ != nil {
				ctx.SendLogs("IKSUpdateRunningClusterModel:  Update - "+"Nothing to update", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			utils.Publisher(utils.ResponseSchema{
				Status:  true,
				Message: "Nothing to update",
				InfraId: cluster.InfraId,
				Token:   token,
				Action:  models.Update,
			}, ctx)
			return err
		}

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err1.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Update,
		}, ctx)
		return types.CustomCPError{}
	}

	/*if previousPoolCount < newPoolCount {

		var pools []*NodePool
		for i := previousPoolCount; i < newPoolCount; i++ {
			pools = append(pools, cluster.NodePools[i])
		}

		err := AddNodepool(&cluster, ctx, iks, pools, previousPoolCount, token)
		if err != (types.CustomCPError{}) {
			utils.SendLog(ctx.Data.Company, "Cluster updation failed"+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

			cluster.Status = models.ClusterUpdateFailed
			confError := UpdateCluster(cluster, false, ctx)
			if confError != nil {
				ctx.SendLogs("IKSUpdateRunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			//err := ApiError(err, "Error occured while apply cluster changes", 500)
			err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.IKS, ctx, err)
			if err_ != nil {
				ctx.SendLogs("IKSUpdateRunningClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}

			return err
		}

	} else if previousPoolCount > newPoolCount {

		previousCluster, err := GetPreviousIKSCluster(ctx)
		if err != nil {
			err_ := types.CustomCPError{Error: "Error in updating running cluster", StatusCode: 512, Description: err.Error()}
			return updationFailedError(cluster, ctx, err_)
		}
		for _, oldpool := range previousCluster.NodePools {
			delete := true
			for _, pool := range cluster.NodePools {
				if pool.Name == oldpool.Name {
					delete = false
					break
				}
			}
			if delete == true {
				err_ := DeleteNodepool(cluster, ctx, iks, oldpool.Name, oldpool.PoolId)
				if err_ != (types.CustomCPError{}) {
					return err_
				}
			}
		}
	}*/
	previousCluster, err := GetPreviousIKSCluster(ctx)
	if err != nil {
		err_ := types.CustomCPError{Error: "Error in updating running cluster", StatusCode: 512, Description: err.Error()}
		return updationFailedError(cluster, ctx, err_,token)
	}
	previousPoolCount := len(previousCluster.NodePools)

	addincluster := false
	var addpools []*NodePool
	var addedIndex []int
	for index, pool := range cluster.NodePools {
		existInPrevious := false
		for _, prePool := range previousCluster.NodePools {
			if pool.Name == prePool.Name {
				existInPrevious = true

			}
		}
		if existInPrevious == false {
			addpools = append(addpools, pool)
			addedIndex = append(addedIndex, index)
			addincluster = true
		}
	}
	if addincluster == true {
		err2 := AddNodepool(&cluster, ctx, iks, addpools, token)
		if err2 != (types.CustomCPError{}) {
			utils.SendLog(ctx.Data.Company, "Cluster updation failed"+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

			cluster.Status = models.ClusterUpdateFailed
			confError := UpdateCluster(cluster, false, ctx)
			if confError != nil {
				ctx.SendLogs("IKSpdateRunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			//err := ApiError(err, "Error occured while apply cluster changes", 500)
			err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.EKS, ctx, err2)
			if err_ != nil {
				ctx.SendLogs("IKSUpdateRunningClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			utils.Publisher(utils.ResponseSchema{
				Status:  false,
				Message: err2.Error + "\n" + err2.Description,
				InfraId: cluster.InfraId,
				Token:   token,
				Action:  models.Update,
			}, ctx)
			return err2
		}
	}
	for _, prePool := range previousCluster.NodePools {
		existInNew := false
		for _, pool := range cluster.NodePools {
			if pool.Name == prePool.Name {
				existInNew = true
			}
		}
		if existInNew == false {
			DeleteNodepool(cluster, ctx, iks, prePool.Name, prePool.PoolId,token)
		}

	}

	poolIndex_, currentpoolIndex_ := -1, -1
	for _, dif := range difCluster {
		if dif.Type != "update" || len(dif.Path) < 1 {
			continue
		}

		if len(dif.Path) > 2 {
			currentpoolIndex_, _ = strconv.Atoi(dif.Path[1])
			poolIndex, _ := strconv.Atoi(dif.Path[1])
			if poolIndex > (previousPoolCount - 1) {
				break
			}
			for _, index := range addedIndex {
				if index == poolIndex {
					continue
				}
			}
		}
		if dif.Path[0] == "KubeVersion" {
			utils.SendLog(ctx.Data.Company, "Changing kubernetes version of cluster "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

			err := iks.updateMasterVersion(cluster.ResourceGroup, cluster.ClusterId, cluster.KubeVersion, ctx)
			if err != (types.CustomCPError{}) {

				utils.SendLog(ctx.Data.Company, err.Description+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
				utils.SendLog(ctx.Data.Company, "Cluster updation failed"+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

				cluster.Status = models.ClusterUpdateFailed
				confError := UpdateCluster(cluster, false, ctx)
				if confError != nil {
					ctx.SendLogs("IKSpdateRunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				}
				//err := ApiError(err1, "Error occured while apply cluster changes", 500)
				err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.IKS, ctx, err)
				if err_ != nil {
					ctx.SendLogs("IKSUpdateRunningClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				}
				utils.Publisher(utils.ResponseSchema{
					Status:  false,
					Message: err.Error + "\n" + err.Description,
					InfraId: cluster.InfraId,
					Token:   token,
					Action:  models.Update,
				}, ctx)
				return err
			}
			utils.SendLog(ctx.Data.Company, "Kubernetes version updated of cluster "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

		} else if len(dif.Path) >= 3 && dif.Path[0] == "NodePools" && currentpoolIndex_ != poolIndex_ && dif.Path[2] == "NodeCount" {

			poolIndex, _ := strconv.Atoi(dif.Path[1])
			utils.SendLog(ctx.Data.Company, "Changing nodepool size of nodepool "+cluster.NodePools[poolIndex].Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

			err := iks.updatePoolSize(cluster.ResourceGroup, cluster.ClusterId, cluster.NodePools[poolIndex].PoolId, cluster.NodePools[poolIndex].NodeCount, ctx)
			if err != (types.CustomCPError{}) {

				utils.SendLog(ctx.Data.Company, err.Description+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
				utils.SendLog(ctx.Data.Company, "Cluster updation failed"+" "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

				cluster.Status = models.ClusterUpdateFailed
				confError := UpdateCluster(cluster, false, ctx)
				if confError != nil {
					ctx.SendLogs("IKSpdateRunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				}
				//err := ApiError(err, "Error occured while apply cluster changes", 500)
				err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.IKS, ctx, err)
				if err_ != nil {
					ctx.SendLogs("IKSUpdateRunningClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				}
				utils.Publisher(utils.ResponseSchema{
					Status:  false,
					Message: err.Error + "\n" + err.Description,
					InfraId: cluster.InfraId,
					Token:   token,
					Action:  models.Update,
				}, ctx)
				return err
			}
			utils.SendLog(ctx.Data.Company, "Nodepool size updated successfully", models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

			currentpoolIndex_ = poolIndex_
		}

	}

	utils.SendLog(ctx.Data.Company, "Running Cluster updated successfully "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

	err = DeletePreviousIKSCluster(ctx)
	if err != nil {
		beego.Info("***********")
		beego.Info(err.Error())
	}
	/*cluster, err = GetEKSCluster(ctx.Data.InfraId, ctx.Data.Company, ctx)
	if err != nil {
		beego.Info("***********")
		beego.Info(err.Error())
	}*/

	/*	latestCluster, err2 := eks.GetClusterStatus(cluster.Name, ctx)
		if err2 != (types.CustomCPError{}) {
			return err2
		}

		beego.Info("*******" + *latestCluster.Status)
		for strings.ToLower(string(*latestCluster.Status)) != strings.ToLower("running") {
			time.Sleep(time.Second * 60)
		}*/
	cluster.Status = models.ClusterCreated
	err_update := UpdateCluster(cluster, false, ctx)
	if err_update != nil {

		ctx.SendLogs("IKSpdateRunningClusterModel:  Update - "+err_update.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	}

	utils.Publisher(utils.ResponseSchema{
		Status:  true,
		Message: "Cluster updated successfully",
		InfraId: cluster.InfraId,
		Token:   token,
		Action:  models.Update,
	}, ctx)

	return types.CustomCPError{}

}
func CompareClusters(ctx utils.Context) (diff.Changelog, int, int, error) {
	cluster, err := GetCluster(ctx.Data.InfraId, ctx.Data.Company, ctx)
	if err != nil {

		return diff.Changelog{}, 0, 0, errors.New("error in getting eks cluster")
	}

	oldCluster, err := GetPreviousIKSCluster(ctx)
	if err != nil && strings.Contains(err.Error(), "not found") {
		return diff.Changelog{}, 0, 0, errors.New("Nothing to update")
	}

	previousPoolCount := len(oldCluster.NodePools)
	newPoolCount := len(cluster.NodePools)

	difCluster, err := diff.Diff(oldCluster, cluster)
	if len(difCluster) < 2 && previousPoolCount == newPoolCount {
		return diff.Changelog{}, 0, 0, errors.New("Nothing to update")
	} else if err != nil {
		return diff.Changelog{}, 0, 0, errors.New("Error in comparing differences:" + err.Error())
	}
	return difCluster, previousPoolCount, newPoolCount, nil
}
func updationFailedError(cluster Cluster_Def, ctx utils.Context, err types.CustomCPError,token string) types.CustomCPError {
	publisher := utils.Notifier{}

	errr := publisher.Init_notifier()
	if errr != nil {
		PrintError(errr, cluster.Name, ctx)
		ctx.SendLogs(errr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := types.CustomCPError{StatusCode: 500, Error: "Error in deploying EKS Cluster", Description: errr.Error()}
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.IKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("EKSRunningClusterModel: Update - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return cpErr
	}

	cluster.Status = models.ClusterUpdateFailed
	confError := UpdateCluster(cluster, false, ctx)
	if confError != nil {
		PrintError(confError, cluster.Name, ctx)
		ctx.SendLogs("IKSRunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	}

	utils.SendLog(ctx.Data.Company, "Error in running cluster update : "+err.Description, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)

	err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.IKS, ctx, err)
	if err_ != nil {
		ctx.SendLogs("IKSRunningClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	}

	utils.SendLog(ctx.Data.Company, "Deployed cluster update failed : "+cluster.Name, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
	utils.SendLog(ctx.Data.Company, err.Description, models.LOGGING_LEVEL_ERROR, ctx.Data.Company)

	utils.Publisher(utils.ResponseSchema{
		Status:  true,
		Message: "Cluster updated successfully",
		InfraId: cluster.InfraId,
		Token:   token,
		Action:  models.Terminate,
	}, ctx)


	return err
}
func PrintError(confError error, name string, ctx utils.Context) {
	if confError != nil {
		utils.SendLog(ctx.Data.Company, "Cluster creation failed : "+name, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
		utils.SendLog(ctx.Data.Company, confError.Error(), models.LOGGING_LEVEL_ERROR, ctx.Data.Company)
	}
}
func AddNodepool(cluster *Cluster_Def, ctx utils.Context, iksOps IBM, pools []*NodePool, token string) types.CustomCPError {
	/*/
	  Fetching network
	*/
	network, vpcId, err := iksOps.getNetwork(*cluster, token, ctx)
	if err != (types.CustomCPError{}) {
		return err
	}

	for in, pool := range pools {
		utils.SendLog(ctx.Data.Company, "Adding nodepool "+pool.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
		wid, err := iksOps.createWorkerPool(cluster.ResourceGroup, cluster.ClusterId, vpcId, pool, network, ctx)
		if err != (types.CustomCPError{}) {
			ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			utils.SendLog(ctx.Data.Company, err.Description, "error", cluster.InfraId)

			return err
		}
		utils.SendLog(ctx.Data.Company, pool.Name+" nodepool added successfully", models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
		pools[in].PoolId = wid
	}

	oldCluster, err1 := GetPreviousIKSCluster(ctx)
	if err1 != nil {
		ctx.SendLogs(err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, err1.Error(), "error", cluster.InfraId)

		return types.CustomCPError{
			StatusCode:  int(models.CloudStatusCode),
			Error:       "Error in adding nodepool in running cluster",
			Description: err1.Error(),
		}
	}

	oldCluster.NodePools = cluster.NodePools
	for in, mainPool := range cluster.NodePools {
		cluster.NodePools[in].PoolStatus = true
		for _, pool := range pools {
			if pool.Name == mainPool.Name {
				cluster.NodePools[in].PoolId = pool.PoolId
			}
		}
	}

	err1 = AddPreviousIKSCluster(oldCluster, ctx, true)
	if err1 != nil {
		ctx.SendLogs(err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(ctx.Data.Company, err1.Error(), "error", cluster.InfraId)

		return types.CustomCPError{Error: "Error in adding nodepool in running cluster", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)}
	}
	return types.CustomCPError{}
}

func DeleteNodepool(cluster Cluster_Def, ctx utils.Context, iksOps IBM, poolName, poolId ,token string) types.CustomCPError {
	utils.SendLog(ctx.Data.Company, "Deleting nodePool "+poolId, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

	err := iksOps.removeWorkerPool(cluster.ResourceGroup, cluster.ClusterId, poolId, ctx)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}
	utils.SendLog(ctx.Data.Company, " NodePool "+poolId+"deleted successfully", models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

	oldCluster, err1 := GetPreviousIKSCluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{
			StatusCode:  int(models.CloudStatusCode),
			Error:       "Error in deleting nodepool in running cluster",
			Description: err1.Error(),
		},token)
	}

	for _, pool := range oldCluster.NodePools {
		if pool.Name == poolName {
			pool = nil
		}
	}
	err1 = AddPreviousIKSCluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx,
			types.CustomCPError{Error: "Error in deleting nodepool in running cluster", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}
	return types.CustomCPError{}
}
