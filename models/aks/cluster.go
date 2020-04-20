package aks

import (
	"antelope/models"
	"antelope/models/azure"
	"antelope/models/cores"
	"antelope/models/db"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"encoding/json"
	"errors"
	"fmt"
	aks "github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-02-01/containerservice"
	"github.com/astaxie/beego"
	"github.com/ghodss/yaml"
	"github.com/jasonlvhit/gocron"
	"github.com/signalsciences/ipv4"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type AKSCluster struct {
	ID                bson.ObjectId            `json:"-" bson:"_id,omitempty"`
	ProjectId         string                   `json:"project_id" bson:"project_id"`
	Cloud             models.Cloud             `json:"cloud" bson:"cloud"`
	CreationDate      time.Time                `json:"-" bson:"creation_date"`
	ModificationDate  time.Time                `json:"-" bson:"modification_date"`
	CompanyId         string                   `json:"company_id" bson:"company_id"`
	Status            string                   `json:"status,omitempty" bson:"status,omitempty" validate:"eq=New|eq=new|eq=NEW"`
	ResourceGoup      string                   `json:"resource_group" bson:"resource_group" validate:"required"`
	ClusterProperties ManagedClusterProperties `json:"property" bson:"property" validate:"required,dive"`
	ResourceID        string                   `json:"cluster_id,omitempty" bson:"cluster_id,omitempty"`
	Name              string                   `json:"name,omitempty" bson:"name,omitempty" validate:"required"`
	Type              string                   `json:"type,omitempty" bson:"type,omitempty"`
	Location          string                   `json:"location,omitempty" bson:"location,omitempty" validate:"required"`
}

type ManagedClusterProperties struct {
	ProvisioningState      string                               `json:"provisioning_state,omitempty" bson:"provisioning_state,omitempty"`
	KubernetesVersion      string                               `json:"kubernetes_version,omitempty" bson:"kubernetes_version,omitempty" validate:"required"`
	DNSPrefix              string                               `json:"dns_prefix,omitempty" bson:"dns_prefix,omitempty" validate:"required"`
	Fqdn                   string                               `json:"fqdn,omitempty" bson:"fqdn,omitempty"`
	AgentPoolProfiles      []ManagedClusterAgentPoolProfile     `json:"agent_pool,omitempty" bson:"agent_pool,omitempty" validate:"required,dive"`
	APIServerAccessProfile ManagedClusterAPIServerAccessProfile `json:"api_server_access_profile,omitempty" bson:"api_server_access_profile,omitempty"`
	EnableRBAC             bool                                 `json:"enable_rbac,omitempty" bson:"enable_rbac,omitempty"`
	IsHttpRouting          bool                                 `json:"enable_http_routing,omitempty" bson:"enable_http_routing,omitempty"`
	IsServicePrincipal     bool                                 `json:"enable_service_principal,omitempty" bson:"enable_service_principal,omitempty"`
	ClientID               string                               `json:"client_id,omitempty" bson:"client_id,omitempty"`
	Secret                 string                               `json:"secret,omitempty" bson:"secret,omitempty"`
	ClusterTags            []Tag                                `json:"tags" bson:"tags"`
	IsAdvanced             bool                                 `json:"is_advance" bson:"is_advance"`
	IsExpert               bool                                 `json:"is_expert" bson:"is_expert"`
	PodCidr                string                               `json:"pod_cidr,omitempty" bson:"pod_cidr,omitempty" validate:"cidrv4"`
	ServiceCidr            string                               `json:"service_cidr,omitempty" bson:"service_cidr,omitempty" validate:"cidrv4"`
	DNSServiceIP           string                               `json:"dns_service_ip,omitempty" bson:"dns_service_ip,omitempty" validate:"ipv4"`
	DockerBridgeCidr       string                               `json:"docker_bridge_cidr,omitempty" bson:"docker_bridge_cidr,omitempty" validate:"cidrv4"`
}

type Tag struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

// ManagedClusterAPIServerAccessProfile access profile for managed cluster API server.
type ManagedClusterAPIServerAccessProfile struct {
	AuthorizedIPRanges   []string `json:"authorized_ip_ranges,omitempty"`
	EnablePrivateCluster bool     `json:"enable_private_cluster,omitempty" bson:"enable_private_cluster,omitempty"`
}

// ManagedClusterAgentPoolProfile profile for the container service agent pool.
type ManagedClusterAgentPoolProfile struct {
	Name              *string            `json:"name,omitempty" bson:"name,omitempty" validate:"required"`
	Count             *int32             `json:"count,omitempty" bson:"count,omitempty" validate:"required,gte=1"`
	VMSize            *aks.VMSizeTypes   `json:"vm_size,omitempty" bson:"vm_size,omitempty" validate:"required"`
	OsDiskSizeGB      *int32             `json:"os_disk_size_gb,omitempty" bson:"os_disk_size_gb,omitempty"`
	VnetSubnetID      *string            `json:"subnet_id" bson:"subnet_id"`
	MaxPods           *int32             `json:"max_pods,omitempty" bson:"max_pods,omitempty" validate:"required"`
	OsType            *aks.OSType        `json:"os_type,omitempty" bson:"os_type,omitempty"`
	MaxCount          *int32             `json:"max_count,omitempty" bson:"max_count,omitempty"`
	MinCount          *int32             `json:"min_count,omitempty" bson:"min_count,omitempty"`
	EnableAutoScaling *bool              `json:"enable_auto_scaling,omitempty" bson:"enable_auto_scaling,omitempty"`
	NodeLabels        []Tag              `json:"node_labels,omitempty" bson:"node_labels,omitempty"`
	NodeTaints        map[string]*string `json:"node_taints,omitempty" bson:"node_taints,omitempty"`
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

type AzureRegion struct {
	region   string
	location string
}

func GetAKSCluster(projectId string, companyId string, ctx utils.Context) (cluster AKSCluster, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs(
			"AKSGetClusterModel:  Get - Got error while connecting to the database: "+err1.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAKSClusterCollection)
	err = c.Find(bson.M{"project_id": projectId, "company_id": companyId}).One(&cluster)
	if err != nil {
		ctx.SendLogs(
			"AKSGetClusterModel:  Get - Got error while fetching from database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err
	}

	return cluster, nil
}

func GetAllAKSCluster(data rbacAuthentication.List, ctx utils.Context) (clusters []AKSCluster, err error) {
	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs(
			"AKSGetAllClusterModel:  GetAll - Got error while connecting to the database: "+err1.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return clusters, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAKSClusterCollection)
	err = c.Find(bson.M{"project_id": bson.M{"$in": copyData}}).All(&clusters)
	if err != nil {
		ctx.SendLogs(
			"AKSGetAllClusterModel:  GetAll - Got error while fetching from database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return clusters, err
	}

	return clusters, nil
}

func AddAKSCluster(cluster AKSCluster, ctx utils.Context) error {
	_, err := GetAKSCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err == nil {
		text := fmt.Sprintf("AKSAddClusterModel:  Add - Cluster for project '%s' already exists in the database.", cluster.ProjectId)
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"AKSAddClusterModel:  Add - Got error while connecting to the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}
	defer session.Close()

	if cluster.CreationDate.IsZero() {
		cluster.CreationDate = time.Now()
		cluster.ModificationDate = time.Now()
		if cluster.Status == "" {
			cluster.Status = "new"
		}
		cluster.Cloud = models.AKS
	}

	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoAKSClusterCollection, cluster)
	if err != nil {
		ctx.SendLogs(
			"AKSAddClusterModel:  Add - Got error while inserting cluster to the database:  "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func UpdateAKSCluster(cluster AKSCluster, ctx utils.Context) error {
	oldCluster, err := GetAKSCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err != nil {
		text := "AKSUpdateClusterModel:  Update - Cluster '" + cluster.Name + "' does not exist in the database: " + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	//if oldCluster.Status == string(models.Deploying) {
	//	ctx.SendLogs(
	//		"AKSUpdateClusterModel:  Update - Cluster is in deploying state.",
	//		models.LOGGING_LEVEL_ERROR,
	//		models.Backend_Logging,
	//	)
	//	return errors.New("cluster is in deploying state")
	//}
	//if oldCluster.Status == string(models.Terminating) {
	//	ctx.SendLogs(
	//		"AKSUpdateClusterModel:  Update - Cluster is in terminating state.",
	//		models.LOGGING_LEVEL_ERROR,
	//		models.Backend_Logging,
	//	)
	//	return errors.New("cluster is in terminating state")
	//}
	//if strings.ToLower(oldCluster.Status) == strings.ToLower(string(models.ClusterCreated)) {
	//	ctx.SendLogs(
	//		"AKSUpdateClusterModel:  Update - Cluster is in running state.",
	//		models.LOGGING_LEVEL_ERROR,
	//		models.Backend_Logging,
	//	)
	//	return errors.New("cluster is in running state")
	//}

	err = DeleteAKSCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err != nil {
		ctx.SendLogs(
			"AKSUpdateClusterModel:  Update - Got error deleting cluster "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	cluster.CreationDate = oldCluster.CreationDate
	cluster.ModificationDate = time.Now()

	err = AddAKSCluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs(
			"AKSUpdateClusterModel:  Update - Got error creating cluster "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func DeleteAKSCluster(projectId, companyId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"AKSDeleteClusterModel:  Delete - Got error while connecting to the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAKSClusterCollection)
	err = c.Remove(bson.M{"project_id": projectId, "company_id": companyId})
	if err != nil {
		ctx.SendLogs(
			"AKSDeleteClusterModel:  Delete - Got error while deleting from the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func DeployAKSCluster(cluster AKSCluster, credentials vault.AzureProfile, companyId string, token string, ctx utils.Context) (confError types.CustomCPError) {

	publisher := utils.Notifier{}
	_ = publisher.Init_notifier()

	aksOps, _ := GetAKS(credentials.Profile)

	CpErr := aksOps.init()
	if CpErr != (types.CustomCPError{}) {
		ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+CpErr.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		cluster.Status = "Cluster creation failed"
		UpdationErr := UpdateAKSCluster(cluster, ctx)
		if UpdationErr != nil {
			_, _ = utils.SendLog(companyId, "Cluster creation failed : "+UpdationErr.Error(), "error", cluster.ProjectId)

			ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+UpdationErr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return CpErr
	}

	_, _ = utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	cluster.Status = string(models.Deploying)
	err_ := UpdateAKSCluster(cluster, ctx)
	if err_ != nil {

		utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.ProjectId)

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return ApiError(err_, "Error occurred while updating cluster in database", 500)
	}
	err := aksOps.CreateCluster(cluster, token, ctx)

	if err != nil {
		cpErr := ApiError(err, "", 502)

		_, _ = utils.SendLog(companyId, "Cluster creation failed : "+cpErr.Message, "error", cluster.ProjectId)
		_, _ = utils.SendLog(companyId, cpErr.Description, "error", cluster.ProjectId)

		cluster.Status = "Cluster creation failed"
		UpdationErr := UpdateAKSCluster(cluster, ctx)
		if UpdationErr != nil {
			_, _ = utils.SendLog(companyId, "Cluster creation failed : "+UpdationErr.Error(), "error", cluster.ProjectId)

			ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+UpdationErr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}
	AgentErr := azure.ApplyAgent(credentials, token, ctx, cluster.Name, cluster.ResourceGoup)
	if AgentErr != nil {
		cpErr := ApiError(AgentErr, "agent deployment failed", 500)

		_, _ = utils.SendLog(companyId, "Cluster creation failed : "+cpErr.Message, "error", cluster.ProjectId)
		_, _ = utils.SendLog(companyId, cpErr.Description, "error", cluster.ProjectId)

		cluster.Status = "Cluster creation failed"
		UpdationErr := UpdateAKSCluster(cluster, ctx)
		if UpdationErr != nil {
			_, _ = utils.SendLog(companyId, "Cluster creation failed : "+UpdationErr.Error(), "error", cluster.ProjectId)

			ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+UpdationErr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}
	cluster.Status = "Cluster Created"

	UpdationErr := UpdateAKSCluster(cluster, ctx)
	if UpdationErr != nil {

		_, _ = utils.SendLog(companyId, "Cluster creation failed : "+UpdationErr.Error(), "error", cluster.ProjectId)
		ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+UpdationErr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return ApiError(err_, "Error occurred while updating cluster in database", 500)
	}

	_, _ = utils.SendLog(companyId, "Cluster created successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	return types.CustomCPError{}
}

func FetchStatus(credentials vault.AzureCredentials, token, projectId, companyId string, ctx utils.Context) (AKSCluster, types.CustomCPError) {
	cluster, err := GetAKSCluster(projectId, companyId, ctx)
	if err != nil {
		return cluster, types.CustomCPError{Message: err.Error(),
			Description: "Error occurred while getting cluster status in database",
			StatusCode:  500}
	}
	cpErr, err := GetError(cluster.ProjectId, ctx.Data.Company, ctx)
	if err != nil {
		return AKSCluster{}, types.CustomCPError{Message: "Error occurred while getting cluster status in database",
			Description: "Error occurred while getting cluster status in database",
			StatusCode:  500}
	}
	if cpErr.Err != (types.CustomCPError{}) {
		return AKSCluster{}, cpErr.Err
	}
	aksOps, _ := GetAKS(credentials)

	CpErr := aksOps.init()
	if CpErr != (types.CustomCPError{}) {
		ctx.SendLogs("AKSClusterModel:  Fetch -"+CpErr.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return AKSCluster{}, CpErr
	}

	CpErr = aksOps.fetchClusterStatus(&cluster, ctx)
	if CpErr != (types.CustomCPError{}) {
		return cluster, CpErr
	}

	return cluster, types.CustomCPError{}
}

func TerminateCluster(credentials vault.AzureProfile, projectId, companyId string, ctx utils.Context) types.CustomCPError {
	publisher := utils.Notifier{}
	_ = publisher.Init_notifier()

	cluster, err := GetAKSCluster(projectId, companyId, ctx)
	if err != nil {
		return ApiError(err, "Error wile getting cluster from database", 500)
	}

	if cluster.Status == "" || cluster.Status == "new" {
		text := "AKSClusterModel : Terminate - Cannot terminate a new cluster"
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return ApiError(errors.New(text), text, 400)
	}

	aksOps, _ := GetAKS(credentials.Profile)
	_, _ = utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	cluster.Status = string(models.Terminating)
	err_ := UpdateAKSCluster(cluster, ctx)
	if err_ != nil {
		utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.ProjectId)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return ApiError(err_, "Error while updating cluster in database", 500)
	}

	CpErr := aksOps.init()
	if CpErr != (types.CustomCPError{}) {
		ctx.SendLogs("AKSClusterModel : Terminate -"+CpErr.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		utils.SendLog(companyId, CpErr.Description, "error", cluster.ProjectId)

		cluster.Status = "Cluster Termination Failed"
		err = UpdateAKSCluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("AKSClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return CpErr
	}

	CpErr = aksOps.TerminateCluster(cluster, ctx)
	if CpErr != (types.CustomCPError{}) {

		_, _ = utils.SendLog(companyId, "Cluster termination failed: "+CpErr.Message, "error", cluster.ProjectId)
		_, _ = utils.SendLog(companyId, CpErr.Description, "error", cluster.ProjectId)

		cluster.Status = "Cluster Termination Failed"
		err = UpdateAKSCluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("AKSClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return CpErr
	}

	cluster.Status = "Cluster Terminated"

	err = UpdateAKSCluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterModel : Terminate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		_, _ = utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
		_, _ = utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return ApiError(err, "Error while updating cluster in database", 500)
	}
	_, _ = utils.SendLog(companyId, "Cluster terminated successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	return types.CustomCPError{}
}

func GetKubeCofing(credentials vault.AzureCredentials, cluster AKSCluster, ctx utils.Context) (interface{}, types.CustomCPError) {
	aksOps, _ := GetAKS(credentials)

	CpErr := aksOps.init()
	if CpErr != (types.CustomCPError{}) {
		ctx.SendLogs("AKSClusterModel : GetKubeConfig -"+CpErr.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", CpErr
	}

	aksKubeConfig, err := aksOps.GetKubeConfig(ctx, cluster)
	if err != (types.CustomCPError{}) {
		ctx.SendLogs("AKSClusterModel : GetKubeConfig -"+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}

	var kubeconfigobj interface{}
	bytes, _ := yaml.YAMLToJSON(*aksKubeConfig.Value)
	_ = json.Unmarshal(bytes, &kubeconfigobj)
	return kubeconfigobj, types.CustomCPError{}

}

func PrintError(confError error, name, projectId string, companyId string) {
	if confError != nil {
		beego.Error(confError.Error())
		_, _ = utils.SendLog(companyId, "Cluster creation failed : "+name, "error", projectId)
		_, _ = utils.SendLog(companyId, confError.Error(), "error", projectId)
	}
}

func GetAKSVms(ctx utils.Context) []string {
	aksvms := GetAKSSupportedVms(ctx)
	var vms []string
	for _, v := range aksvms {
		if v == "Standard_A0" || v == "Standard_A1" || v == "Standard_A1_v2" || v == "Standard_B1s" || v == "Standard_B1ms" || v == "Standard_F1" || v == "Standard_F1s" {
			continue
		} else {
			vms = append(vms, string(v))
		}
	}
	return vms
}

func GetVms(region string, ctx utils.Context) ([]string, error) {
	skusList, err := GetVmSkus(ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []string{}, err
	}

	var vms []string
	for _, v := range skusList {
		for _, location := range v.Locations {
			if location != region {
				continue
			} else {
				if v.ResourceType == "virtualMachines" {
					if v.Name == "Standard_A0" || v.Name == "Standard_A1" || v.Name == "Standard_A1_v2" || v.Name == "Standard_B1s" || v.Name == "Standard_B1ms" || v.Name == "Standard_F1" || v.Name == "Standard_F1s" {
						continue
					} else {
						vms = append(vms, v.Name)
						break
					}
				}
			}
		}
	}

	return vms, nil
}

func GetKubeVersions(credentials vault.AzureProfile, ctx utils.Context) ([]string, types.CustomCPError) {
	aksOps, _ := GetAKS(credentials.Profile)

	CpErr := aksOps.init()
	if CpErr != (types.CustomCPError{}) {
		ctx.SendLogs(CpErr.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []string{}, CpErr
	}

	result, err := aksOps.GetKubernetesVersions(ctx)
	if err != (types.CustomCPError{}) {
		ctx.SendLogs(err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []string{}, err
	}

	var versions []string
	for _, versionProfile := range *result.Orchestrators {
		if *versionProfile.OrchestratorVersion == "1.6.9" {
			continue
		}
		if *versionProfile.OrchestratorType == "Kubernetes" {
			versions = append(versions, *versionProfile.OrchestratorVersion)
		}
	}

	return versions, types.CustomCPError{}

}

func ValidateAKSData(cluster AKSCluster, ctx utils.Context) error {
	if cluster.ProjectId == "" {

		return errors.New("project ID is empty")

	} else if cluster.ResourceGoup == "" {

		return errors.New("Resource group name must is empty")

	} else if cluster.Location == "" {

		return errors.New("location is empty")

	} else {

		isRegionExist, err := validateAKSRegion(cluster.Location)
		if err != nil && !isRegionExist {
			text := "availabe locations are " + err.Error()
			ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return errors.New(text)
		}

	}

	if len(cluster.ClusterProperties.AgentPoolProfiles) == 0 {

		return errors.New("length of node pools must be greater than zero")

	} else if cluster.ClusterProperties.IsAdvanced {

		if cluster.ClusterProperties.KubernetesVersion == "" {

			return errors.New("kubernetes version is empty")

		} else if cluster.ClusterProperties.DNSPrefix == "" {

			return errors.New("DNS prefix is empty")

		} else if cluster.ClusterProperties.IsServicePrincipal {

			if cluster.ClusterProperties.ClientID == "" || cluster.ClusterProperties.Secret == "" {

				return errors.New("client id or secret is empty")

			}
		}

		for _, pool := range cluster.ClusterProperties.AgentPoolProfiles {

			if pool.Name != nil && *pool.Name == "" {

				return errors.New("Node Pool name is empty")

			} else if pool.VMSize != nil && *pool.VMSize == "" {

				return errors.New("machine type with pool " + *pool.Name + " is empty")

			} else if pool.Count != nil && *pool.Count == 0 {

				return errors.New("node count value is zero within pool " + *pool.Name)

			} else if pool.OsDiskSizeGB != nil && (*pool.OsDiskSizeGB == 0 || *pool.OsDiskSizeGB < 40 || *pool.OsDiskSizeGB > 2048) {

				return errors.New("Disk size must be greater than 40 and less than 2048 within pool " + *pool.Name)

			} else if pool.MaxPods != nil && (*pool.MaxPods == 0 || *pool.MaxPods < 40) {

				return errors.New("max pods must be greater than or equal to 40 within pool " + *pool.Name)

			} else if pool.EnableAutoScaling != nil && *pool.EnableAutoScaling {

				if *pool.MinCount > *pool.MaxCount {
					return errors.New("min count should be less than or equal to max count within pool " + *pool.Name)
				}

			}

		}
	}

	if cluster.ClusterProperties.IsExpert {
		if cluster.ClusterProperties.PodCidr == "" {

			return errors.New("pod CIDR must not be empty")

		} else {

			isValidCidr := ipv4.IsIPv4(cluster.ClusterProperties.PodCidr)
			if !isValidCidr {
				return errors.New("pod CIDR is not valid")
			}

		}

		if cluster.ClusterProperties.DNSServiceIP == "" {

			return errors.New("DNS service IP must not be empty")

		} else {

			isValidIp := ipv4.IsIPv4(cluster.ClusterProperties.DNSServiceIP)
			if !isValidIp {
				return errors.New("DNS service IP is not valid")
			}

		}

		if cluster.ClusterProperties.DockerBridgeCidr == "" {

			return errors.New("Docker Bridge CIDR must not be empty")

		} else {

			isValidCidr := ipv4.IsIPv4(cluster.ClusterProperties.DockerBridgeCidr)
			if !isValidCidr {
				return errors.New("docker bridge CIDR is not valid")
			}

		}

		if cluster.ClusterProperties.ServiceCidr == "" {

			return errors.New("Service CIDR must not be empty")

		} else {

			isValidCidr := ipv4.IsIPv4(cluster.ClusterProperties.ServiceCidr)
			if !isValidCidr {
				return errors.New("service CIDR is not valid")
			}

		}
	}

	return nil
}

func validateAKSRegion(region string) (bool, error) {

	bytes := cores.AzureRegions

	var regionList []AzureRegion

	err := json.Unmarshal(bytes, &regionList)
	if err != nil {
		return false, err
	}

	for _, v1 := range regionList {
		if v1.location == region {
			return true, nil
		}
	}

	var errData string
	for _, v1 := range regionList {
		errData += v1.location + ", "
	}

	return false, errors.New(errData)
}

func RunCronJob() {
	gocron.Every(1).Monday().Do(Task)
	gocron.Start()
}

func Task() {
	fmt.Println("running job")
	credentials := vault.AzureCredentials{
		ClientId:       "87b20591-1867-49ac-add7-ada0f22a70e4",
		ClientSecret:   "IV54Er?tiv8H3CSYwjZzPaAMl*UoFl?=",
		SubscriptionId: "aa94b050-2c52-4b7b-9ce3-2ac18253e61e",
		TenantId:       "959c117c-1656-470a-8403-947584c67e55",
		Location:       "new",
	}

	aksOps, err := GetAKS(credentials)
	if err != nil {
		fmt.Println(err)
		return
	}

	CpErr := aksOps.init()
	if CpErr != (types.CustomCPError{}) {
		fmt.Println(CpErr)
		return
	}

	aksOps.WriteAzureSkus()
}
