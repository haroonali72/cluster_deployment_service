package aks

import (
	"antelope/models"
	"antelope/models/azure"
	"antelope/models/db"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"antelope/models/vault"
	"encoding/json"
	"errors"
	"fmt"
	aks "github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-02-01/containerservice"
	"github.com/astaxie/beego"
	"github.com/ghodss/yaml"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type AKSCluster struct {
	ID               bson.ObjectId `json:"-" bson:"_id,omitempty"`
	ProjectId        string        `json:"project_id" bson:"project_id"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	//CloudplexStatus  string        `json:"status" bson:"status"`
	CompanyId    string `json:"company_id" bson:"company_id"`
	Status       string `json:"status,omitempty" bson:"status,omitempty"`
	ResourceGoup string `json:"resource_group" bson:"resource_group" validate:"required"`
	// ManagedClusterProperties - Properties of a managed cluster.
	ClusterProperties *ManagedClusterProperties `json:"properties" bson:"properties" validate:"required"`
	// ID - Resource Id
	ResourceID *string `json:"cluster_id,omitempty" bson:"cluster_id,omitempty"`
	// Name - Resource name
	Name *string `json:"name,omitempty" bson:"name,omitempty"`
	// Type - Resource type
	Type *string `json:"type,omitempty" bson:"type,omitempty"`
	// Location - Resource location
	Location *string `json:"location,omitempty" bson:"location,omitempty"`
	// Tags - Resource tags
	Tags map[string]*string `json:"tags" bson:"tags"`
}

type ManagedClusterProperties struct {
	// ProvisioningState - The current deployment or provisioning state, which only appears in the response.
	ProvisioningState *string `json:"provisioning_state,omitempty" bson:"provisioning_state,omitempty"`
	// KubernetesVersion - Version of Kubernetes specified when creating the managed cluster.
	KubernetesVersion *string `json:"kubernetes_version,omitempty" bson:"kubernetes_version,omitempty"`
	// DNSPrefix - DNS prefix specified when creating the managed cluster.
	DNSPrefix *string `json:"dns_prefix,omitempty" bson:"dns_prefix,omitempty"`
	// Fqdn - READ-ONLY; FQDN for the master pool.
	Fqdn *string `json:"fqdn,omitempty" bson:"fqdn,omitempty"`
	// NetworkProfile - Profile of network configuration.
	NetworkProfile *NetworkProfileType `json:"network_profile,omitempty" bson:"network_profile,omitempty"`
	// AgentPoolProfiles - Properties of the agent pool. Currently only one agent pool can exist.
	AgentPoolProfiles []ManagedClusterAgentPoolProfile `json:"agent_pool,omitempty" bson:"agent_pool,omitempty"`
	// APIServerAccessProfile - Access profile for managed cluster API server.
	APIServerAccessProfile *ManagedClusterAPIServerAccessProfile `json:"api_server_access_profile,omitempty" bson:"api_server_access_profile,omitempty"`
}

// ManagedClusterAPIServerAccessProfile access profile for managed cluster API server.
type ManagedClusterAPIServerAccessProfile struct {
	// EnablePrivateCluster - Whether to create the cluster as a private cluster or not.
	EnablePrivateCluster *bool `json:"enable_private_cluster,omitempty" bson:"enable_private_cluster,omitempty"`
}

// NetworkProfileType profile of network configuration.
type NetworkProfileType struct {
	// PodCidr - A CIDR notation IP range from which to assign pod IPs when kubenet is used.
	PodCidr *string `json:"pod_cidr,omitempty" bson:"pod_cidr,omitempty"`
	// ServiceCidr - A CIDR notation IP range from which to assign service cluster IPs. It must not overlap with any Subnet IP ranges.
	ServiceCidr *string `json:"service_cidr,omitempty" bson:"service_cidr,omitempty"`
	// DNSServiceIP - An IP address assigned to the Kubernetes DNS service. It must be within the Kubernetes service address range specified in serviceCidr.
	DNSServiceIP *string `json:"dns_service_ip,omitempty" bson:"dns_service_ip,omitempty"`
	// DockerBridgeCidr - A CIDR notation IP range assigned to the Docker bridge network. It must not overlap with any Subnet IP ranges or the Kubernetes service address range.
	DockerBridgeCidr *string `json:"docker_bridge_cidr,omitempty" bson:"docker_bridge_cidr,omitempty"`
}

// ManagedClusterAgentPoolProfile profile for the container service agent pool.
type ManagedClusterAgentPoolProfile struct {
	// Name - Unique name of the agent pool profile in the context of the subscription and resource group.
	Name *string `json:"name,omitempty" bson:"name,omitempty" validate:"required"`
	// Count - Number of agents (VMs) to host docker containers. Allowed values must be in the range of 1 to 100 (inclusive). The default value is 1.
	Count *int32 `json:"count,omitempty" bson:"count,omitempty" validate:"required"`
	// VMSize - Size of agent VMs. Possible values include: 'StandardA1', 'StandardA10', 'StandardA11', 'StandardA1V2', 'StandardA2', 'StandardA2V2', 'StandardA2mV2', 'StandardA3', 'StandardA4', 'StandardA4V2', 'StandardA4mV2', 'StandardA5', 'StandardA6', 'StandardA7', 'StandardA8', 'StandardA8V2', 'StandardA8mV2', 'StandardA9', 'StandardB2ms', 'StandardB2s', 'StandardB4ms', 'StandardB8ms', 'StandardD1', 'StandardD11', 'StandardD11V2', 'StandardD11V2Promo', 'StandardD12', 'StandardD12V2', 'StandardD12V2Promo', 'StandardD13', 'StandardD13V2', 'StandardD13V2Promo', 'StandardD14', 'StandardD14V2', 'StandardD14V2Promo', 'StandardD15V2', 'StandardD16V3', 'StandardD16sV3', 'StandardD1V2', 'StandardD2', 'StandardD2V2', 'StandardD2V2Promo', 'StandardD2V3', 'StandardD2sV3', 'StandardD3', 'StandardD32V3', 'StandardD32sV3', 'StandardD3V2', 'StandardD3V2Promo', 'StandardD4', 'StandardD4V2', 'StandardD4V2Promo', 'StandardD4V3', 'StandardD4sV3', 'StandardD5V2', 'StandardD5V2Promo', 'StandardD64V3', 'StandardD64sV3', 'StandardD8V3', 'StandardD8sV3', 'StandardDS1', 'StandardDS11', 'StandardDS11V2', 'StandardDS11V2Promo', 'StandardDS12', 'StandardDS12V2', 'StandardDS12V2Promo', 'StandardDS13', 'StandardDS132V2', 'StandardDS134V2', 'StandardDS13V2', 'StandardDS13V2Promo', 'StandardDS14', 'StandardDS144V2', 'StandardDS148V2', 'StandardDS14V2', 'StandardDS14V2Promo', 'StandardDS15V2', 'StandardDS1V2', 'StandardDS2', 'StandardDS2V2', 'StandardDS2V2Promo', 'StandardDS3', 'StandardDS3V2', 'StandardDS3V2Promo', 'StandardDS4', 'StandardDS4V2', 'StandardDS4V2Promo', 'StandardDS5V2', 'StandardDS5V2Promo', 'StandardE16V3', 'StandardE16sV3', 'StandardE2V3', 'StandardE2sV3', 'StandardE3216sV3', 'StandardE328sV3', 'StandardE32V3', 'StandardE32sV3', 'StandardE4V3', 'StandardE4sV3', 'StandardE6416sV3', 'StandardE6432sV3', 'StandardE64V3', 'StandardE64sV3', 'StandardE8V3', 'StandardE8sV3', 'StandardF1', 'StandardF16', 'StandardF16s', 'StandardF16sV2', 'StandardF1s', 'StandardF2', 'StandardF2s', 'StandardF2sV2', 'StandardF32sV2', 'StandardF4', 'StandardF4s', 'StandardF4sV2', 'StandardF64sV2', 'StandardF72sV2', 'StandardF8', 'StandardF8s', 'StandardF8sV2', 'StandardG1', 'StandardG2', 'StandardG3', 'StandardG4', 'StandardG5', 'StandardGS1', 'StandardGS2', 'StandardGS3', 'StandardGS4', 'StandardGS44', 'StandardGS48', 'StandardGS5', 'StandardGS516', 'StandardGS58', 'StandardH16', 'StandardH16m', 'StandardH16mr', 'StandardH16r', 'StandardH8', 'StandardH8m', 'StandardL16s', 'StandardL32s', 'StandardL4s', 'StandardL8s', 'StandardM12832ms', 'StandardM12864ms', 'StandardM128ms', 'StandardM128s', 'StandardM6416ms', 'StandardM6432ms', 'StandardM64ms', 'StandardM64s', 'StandardNC12', 'StandardNC12sV2', 'StandardNC12sV3', 'StandardNC24', 'StandardNC24r', 'StandardNC24rsV2', 'StandardNC24rsV3', 'StandardNC24sV2', 'StandardNC24sV3', 'StandardNC6', 'StandardNC6sV2', 'StandardNC6sV3', 'StandardND12s', 'StandardND24rs', 'StandardND24s', 'StandardND6s', 'StandardNV12', 'StandardNV24', 'StandardNV6'
	VMSize aks.VMSizeTypes `json:"vm_size,omitempty" bson:"vm_size,omitempty" validate:"required"`
	// OsDiskSizeGB - OS Disk Size in GB to be used to specify the disk size for every machine in this master/agent pool. If you specify 0, it will apply the default osDisk size according to the vmSize specified.
	OsDiskSizeGB *int32 `json:"os_disk_size_gb,omitempty" bson:"os_disk_size_gb,omitempty"`
	// VnetSubnetID - VNet SubnetID specifies the vnet's subnet identifier.
	VnetSubnetID *string `json:"subnet_id" bson:"subnet_id"`
	// MaxPods - Maximum number of pods that can run on a node.
	MaxPods *int32 `json:"max_pods,omitempty" bson:"max_pods,omitempty"`
	// OsType - OsType to be used to specify os type. Choose from Linux and Windows. Default to Linux. Possible values include: 'Linux', 'Windows'
	OsType aks.OSType `json:"os_type,omitempty" bson:"os_type,omitempty"`
	// MaxCount - Maximum number of nodes for auto-scaling
	MaxCount *int32 `json:"max_count,omitempty" bson:"max_count,omitempty"`
	// MinCount - Minimum number of nodes for auto-scaling
	MinCount *int32 `json:"min_count,omitempty" bson:"min_count,omitempty"`
	// EnableAutoScaling - Whether to enable auto-scaler
	EnableAutoScaling *bool `json:"enable_auto_scaling,omitempty" bson:"enable_auto_scaling,omitempty"`
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
		cluster.Cloud = models.GKE
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
		text := "AKSUpdateClusterModel:  Update - Cluster '" + *cluster.Name + "' does not exist in the database: " + err.Error()
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

func DeployAKSCluster(
	cluster AKSCluster,
	credentials vault.AzureProfile,
	companyId string,
	token string,
	ctx utils.Context,
) (confError error) {

	publisher := utils.Notifier{}
	confError = publisher.Init_notifier()

	if confError != nil {
		PrintError(confError, *cluster.Name, cluster.ProjectId, companyId)
		ctx.SendLogs(confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return confError
	}

	aksOps, err := GetAKS(credentials.Profile)
	if err != nil {
		ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	err = aksOps.init()
	if err != nil {
		ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.Status = "Cluster creation failed"
		confError = UpdateAKSCluster(cluster, ctx)
		if confError != nil {
			PrintError(confError, *cluster.Name, cluster.ProjectId, companyId)
			ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}

	_, _ = utils.SendLog(companyId, "Creating Cluster : "+*cluster.Name, "info", cluster.ProjectId)
	confError = aksOps.CreateCluster(cluster, token, ctx)

	if confError != nil {
		ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		PrintError(confError, *cluster.Name, cluster.ProjectId, companyId)

		cluster.Status = "Cluster creation failed"
		confError = UpdateAKSCluster(cluster, ctx)
		if confError != nil {
			PrintError(confError, *cluster.Name, cluster.ProjectId, companyId)
			ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return nil
	}
	confError = azure.ApplyAgent(credentials, token, ctx, *cluster.Name, cluster.ResourceGoup)
	if confError != nil {
		ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		PrintError(confError, *cluster.Name, cluster.ProjectId, companyId)

		cluster.Status = "Cluster creation failed"
		confError = UpdateAKSCluster(cluster, ctx)
		if confError != nil {
			PrintError(confError, *cluster.Name, cluster.ProjectId, companyId)
			ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return nil
	}
	cluster.Status = "Cluster Created"

	confError = UpdateAKSCluster(cluster, ctx)
	if confError != nil {
		PrintError(confError, *cluster.Name, cluster.ProjectId, companyId)
		ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return confError
	}

	_, _ = utils.SendLog(companyId, "Cluster created successfully "+*cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	return nil
}

func FetchStatus(credentials vault.AzureCredentials, token, projectId, companyId string, ctx utils.Context) (AKSCluster, error) {
	cluster, err := GetAKSCluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterModel:  Fetch -  Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err
	}

	aksOps, err := GetAKS(credentials)
	if err != nil {
		ctx.SendLogs("AKSClusterModel:  Fetch -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err
	}

	err = aksOps.init()
	if err != nil {
		ctx.SendLogs("AKSClusterModel:  Fetch -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err
	}

	err = aksOps.fetchClusterStatus(&cluster, ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterModel:  Fetch - Failed to get latest status "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err
	}

	return cluster, nil
}

func TerminateCluster(credentials vault.AzureCredentials, projectId, companyId string, ctx utils.Context) error {
	publisher := utils.Notifier{}
	pubErr := publisher.Init_notifier()
	if pubErr != nil {
		ctx.SendLogs("AKSClusterModel:  Terminate -"+pubErr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return pubErr
	}

	cluster, err := GetAKSCluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterModel : Terminate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	if cluster.Status == "" || cluster.Status == "new" {
		text := "AKSClusterModel : Terminate - Cannot terminate a new cluster"
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return errors.New(text)
	}

	aksOps, err := GetAKS(credentials)
	if err != nil {
		ctx.SendLogs("AKSClusterModel : Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	cluster.Status = string(models.Terminating)
	_, _ = utils.SendLog(companyId, "Terminating cluster: "+*cluster.Name, "info", cluster.ProjectId)

	err = aksOps.init()
	if err != nil {
		ctx.SendLogs("AKSClusterModel : Terminate -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.Status = "Cluster Termination Failed"
		err = UpdateAKSCluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("AKSClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			_, _ = utils.SendLog(companyId, "Error in cluster updation in mongo: "+*cluster.Name, "error", cluster.ProjectId)
			_, _ = utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
			return err
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}

	err = aksOps.DeleteCluster(cluster, ctx)
	if err != nil {
		_, _ = utils.SendLog(companyId, "Cluster termination failed: "+*cluster.Name, "error", cluster.ProjectId)

		cluster.Status = "Cluster Termination Failed"
		err = UpdateAKSCluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("AKSClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			_, _ = utils.SendLog(companyId, "Error in cluster updation in mongo: "+*cluster.Name, "error", cluster.ProjectId)
			_, _ = utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
			publisher.Notify(cluster.ProjectId, "Status Available", ctx)
			return err
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return nil
	}

	cluster.Status = "Cluster Terminated"

	err = UpdateAKSCluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterModel : Terminate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		_, _ = utils.SendLog(companyId, "Error in cluster updation in mongo: "+*cluster.Name, "error", cluster.ProjectId)
		_, _ = utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}
	_, _ = utils.SendLog(companyId, "Cluster terminated successfully "+*cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	return nil
}

func GetKubeCofing(credentials vault.AzureCredentials, cluster AKSCluster, ctx utils.Context) (interface{}, error) {
	aksOps, err := GetAKS(credentials)
	if err != nil {
		ctx.SendLogs("AKSClusterModel : GetKubeConfig - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}

	err = aksOps.init()
	if err != nil {
		ctx.SendLogs("AKSClusterModel : GetKubeConfig -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}

	aksKubeConfig, err := aksOps.GetKubeConfig(ctx, cluster)
	if err != nil {
		ctx.SendLogs("AKSClusterModel : GetKubeConfig -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}

	var kubeconfigobj interface{}
	bytes, _ := yaml.YAMLToJSON(*aksKubeConfig.Value)
	_ = json.Unmarshal(bytes, &kubeconfigobj)
	return kubeconfigobj, nil

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
