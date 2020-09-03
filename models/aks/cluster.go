package aks

import (
	"antelope/models"
	"antelope/models/api_handler"
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
	"github.com/r3labs/diff"
	"github.com/signalsciences/ipv4"
	"gopkg.in/mgo.v2/bson"
	"sort"
	"strconv"
	"strings"
	"time"
)

//swagger:model akscluster
type AKSCluster struct {
	ID                     bson.ObjectId                        `json:"-" bson:"_id,omitempty"`
	InfraId                string                               `json:"infra_id" bson:"infra_id" validate:"required" description:"ID of infrastructure [required]"`
	Cloud                  models.Cloud                         `json:"cloud" bson:"cloud"`
	CreationDate           time.Time                            `json:"-" bson:"creation_date"`
	ModificationDate       time.Time                            `json:"-" bson:"modification_date"`
	UpdateStatus           string                               `json:"-" bson:"update_status"`
	CompanyId              string                               `json:"company_id" bson:"company_id" description:"ID of compnay [optional]"`
	Status                 models.Type                          `json:"status,omitempty" bson:"status,omitempty" validate:"eq=new|eq=New|eq=NEW|eq=Cluster Creation Failed|eq=Cluster Terminated|eq=Cluster Created|eq=Cluster Update Failed" description:"Status of cluster [required]"`
	ProvisioningState      string                               `json:"-" bson:"provisioning_state,omitempty"`
	KubernetesVersion      string                               `json:"kubernetes_version" bson:"kubernetes_version" description:"Kubernetes version to be provisioned ['required' if advance settings enabled]"`
	DNSPrefix              string                               `json:"dns_prefix,omitempty" bson:"dns_prefix,omitempty" description:"Cluster DNS prefix ['required' if advance settings enabled]"`
	Fqdn                   string                               `json:"-" bson:"fqdn,omitempty"`
	AgentPoolProfiles      []ManagedClusterAgentPoolProfile     `json:"node_pools,omitempty" bson:"node_pools,omitempty" validate:"required,dive"`
	APIServerAccessProfile ManagedClusterAPIServerAccessProfile `json:"api_server_access_profile,omitempty" bson:"api_server_access_profile,omitempty"`
	EnableRBAC             bool                                 `json:"enable_rbac,omitempty" bson:"enable_rbac,omitempty" description:"Cluster RBAC configuration ['required' if advance settings enabled]"`
	IsHttpRouting          bool                                 `json:"enable_http_routing,omitempty" bson:"enable_http_routing,omitempty" description:"Cluster Http Routing configuration ['required' if advance settings enabled]"`
	IsServicePrincipal     bool                                 `json:"enable_service_principal,omitempty" bson:"enable_service_principal,omitempty" description:"Service principal configurations ['required' if advance settings enabled]"`
	ClientID               string                               `json:"client_id,omitempty" bson:"client_id,omitempty" description:"Client ID for service principal ['required' if service principal enabled]"`
	Secret                 string                               `json:"secret,omitempty" bson:"secret,omitempty" description:"Client secret for service principal ['required' if service principal enabled]"`
	ClusterTags            []Tag                                `json:"tags" bson:"tags" description:"Cluster tags [optional]"`
	IsAdvanced             bool                                 `json:"is_advance" bson:"is_advance" description:"Cluster advance level settings possible value 'true' or 'false'"`
	IsExpert               bool                                 `json:"is_expert" bson:"is_expert" description:"Cluster expert level settings possible value 'true' or 'false'"`
	PodCidr                string                               `json:"pod_cidr,omitempty" bson:"pod_cidr,omitempty" description:"Pod CIDR for cluster ['required' if expert settings enabled]"`
	ServiceCidr            string                               `json:"service_cidr,omitempty" bson:"service_cidr,omitempty" description:"Service CIDR for cluster ['required' if expert settings enabled]"`
	DNSServiceIP           string                               `json:"dns_service_ip,omitempty" bson:"dns_service_ip,omitempty" description:"DNS service IP for cluster ['required' if expert settings enabled]"`
	DockerBridgeCidr       string                               `json:"docker_bridge_cidr,omitempty" bson:"docker_bridge_cidr,omitempty" description:"Docker bridge CIDR for cluster ['required' if expert settings enabled]"`
	ResourceGoup           string                               `json:"resource_group" bson:"resource_group" validate:"required" description:"Resources would be created within resource_group [required]"`
	ResourceID             string                               `json:"-" bson:"cluster_id,omitempty"`
	Name                   string                               `json:"name,omitempty" bson:"name,omitempty" validate:"required" description:"Cluster name [required]"`
	Type                   string                               `json:"-" bson:"type,omitempty"`
	Location               string                               `json:"location,omitempty" bson:"location,omitempty" validate:"required" description:"Location for cluster provisioning [required]"`
}

type Tag struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

func GetNetwork(token, infraId string, ctx utils.Context) error {

	url := getNetworkHost("azure", infraId)

	_, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}

type KubeClusterStatus struct {
	Id                string                          `json:"id" bson:"id"  description:"Cluster id"`
	Name              string                          `json:"name" bson:"name"  description:"Cluster name"`
	Region            string                          `json:"region" bson:"region"  description:"Region for cluster provisioning"`
	Status            models.Type                     `json:"status" bson:"status"  description:"Status of cluster"`
	KubernetesVersion string                          `json:"kubernetes_version" bson:"kubernetes_version" description:"Kubernetes version to be provisioned"`
	ProvisioningState string                          `json:"state" bson:"state" description:"Kubernetes state"`
	NodePoolCount     int32                           `json:"nodepool_count" bson:"nodepool_count" description:"Node pool count"`
	ResourceGoup      string                          `json:"resource_group" bson:"resource_group"description:"Resources would be created within resource_group"`
	AgentPoolProfiles []ManagedClusterAgentPoolStatus `json:"node_pools" bson:"node_pools" `
}
type KubeWorkerPoolStatus struct {
	Name              *string `json:"name" bson:"name" validate:"required,alphanum" description:"Cluster pool name"`
	Count             *int32  `json:"node_count" bson:"node_count" description:"Pool node count"`
	VMSize            *string `json:"vm_size" bson:"vm_size"  description:"Machine type for pool"`
	OsDiskSizeGB      *int32  `json:"os_disk_size_gb" bson:"os_disk_size_gb" description:"Disk size for VMs"`
	Subnet            *string `json:"subnet" bson:"subnet" description:"ID of subnet in which pool will be created"`
	MaxPodsPerNode    *int32  `json:"max_pods_per_node" bson:"max_pods_per_node" description:"Max pods per node [required]"`
	MaxCount          *int32  `json:"max_count" bson:"max_count" description:"Max VM count, must be greater than min count"`
	MinCount          *int32  `json:"min_count" bson:"min_count" description:"Min VM count"`
	EnableAutoScaling *bool   `json:"auto_scaling" bson:"auto_scaling" description:"Autoscaling configuration"`
}
type AutoScaling struct {
	MaxCount          *int32 `json:"max_scaling_group_size" bson:"max_count" description:"Max VM count, must be greater than min count"`
	MinCount          *int32 `json:"min_scaling_group_size" bson:"min_count" description:"Min VM count"`
	EnableAutoScaling *bool  `json:"autoscale" bson:"auto_scaling" description:"Autoscaling configuration"`
}
type KubeNodesStatus struct {
	Id        *string `json:"id" bson:"id,omitempty"`
	NodeState *string `json:"state" bson:"state,omitempty"`
	Name      *string `json:"name" bson:"name,omitempty"`
	PrivateIP *string `json:"private_ip,omitempty" bson:"private_ip,omitempty"`
	PublicIP  *string `json:"public_ip,omitempty" bson:"public_ip,omitempty"`
}

type ManagedClusterAgentPoolStatus struct {
	Id           *string           `json:"id" bson:"id" description:"Cluster pool id"`
	Name         *string           `json:"name,omitempty" bson:"name,omitempty"  description:"Cluster pool name "`
	VnetSubnetID *string           `json:"subnet_id" bson:"subnet_id" description:"ID of subnet in which pool is created"`
	Count        *int64            `json:"node_count,omitempty" bson:"count,omitempty"  description:"Pool node count"`
	VMSize       *string           `json:"machine_type,omitempty" bson:"vm_size,omitempty" description:"Machine type for pool"`
	AutoScaling  AutoScaling       `json:"autoscaling" bson:"auto_scaling" description:"Autoscaling configuration"`
	KubeNodes    []KubeNodesStatus `json:"nodes" bson:"nodes" description:"Nodes "`
}

// ManagedClusterAPIServerAccessProfile access profile for managed cluster API server.
type ManagedClusterAPIServerAccessProfile struct {
	AuthorizedIPRanges   []string `json:"authorized_ip_ranges,omitempty" description:"Authorized IP ranges for accessing kube server [optional]"`
	EnablePrivateCluster bool     `json:"-" bson:"enable_private_cluster,omitempty"`
}

// ManagedClusterAgentPoolProfile profile for the container service agent pool.
type ManagedClusterAgentPoolProfile struct {
	Name              *string            `json:"name,omitempty" bson:"name,omitempty" validate:"required" description:"Cluster pool name [required]"`
	Count             *int32             `json:"count,omitempty" bson:"count,omitempty" validate:"required,gte=1" description:"Pool node count [required]"`
	VMSize            *string            `json:"vm_size,omitempty" bson:"vm_size,omitempty" validate:"required" description:"Machine type for pool [required]"`
	OsDiskSizeGB      *int32             `json:"os_disk_size_gb,omitempty" bson:"os_disk_size_gb,omitempty" description:"Disk size for VMs [required]"`
	VnetSubnetID      *string            `json:"subnet_id" bson:"subnet_id" description:"ID of subnet in which pool will be created [required]"`
	MaxPods           *int32             `json:"max_pods,omitempty" bson:"max_pods,omitempty" description:"Max pods per node [required]"`
	OsType            *aks.OSType        `json:"-" bson:"os_type,omitempty"`
	MaxCount          *int32             `json:"max_count,omitempty" bson:"max_count,omitempty" description:"Max VM count, must be greater than min count ['required' if autoscaling is enabled]"`
	MinCount          *int32             `json:"min_count,omitempty" bson:"min_count,omitempty" description:"Min VM count ['required' if autoscaling is enabled]"`
	EnableAutoScaling *bool              `json:"enable_auto_scaling,omitempty" bson:"enable_auto_scaling,omitempty" description:"Autoscaling configuration, possible value 'true' or 'false' [required]"`
	NodeLabels        []Tag              `json:"node_labels,omitempty" bson:"node_labels,omitempty"`
	NodeTaints        map[string]*string `json:"-" bson:"node_taints,omitempty"`
	EnablePublicIp    *bool              `json:"enable_public_ip" bson:"enable_public_ip"`
}

type AzureRegion struct {
	Region   string `json:"region"`
	Location string `json:"location"`
}

type Cluster struct {
	Name    string      `json:"name,omitempty" bson:"name,omitempty" description:"Cluster name"`
	InfraId string      `json:"infra_id" bson:"infra_id"  description:"ID of infrastructure"`
	Status  models.Type `json:"status,omitempty" bson:"status,omitempty" description:"Status of cluster"`
}

func AddPreviousAKSCluster(cluster AKSCluster, ctx utils.Context, patch bool) error {
	var oldCluster AKSCluster
	_, err := GetPreviousAKSCluster(cluster.InfraId, cluster.CompanyId, ctx)
	if err == nil {
		err := DeletePreviousAKSCluster(cluster.InfraId, cluster.CompanyId, ctx)
		if err != nil {
			ctx.SendLogs(
				"AKSAddClusterModel:  Add previous cluster - "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
		}
	}

	if patch == false {
		oldCluster, err = GetAKSCluster(cluster.InfraId, cluster.CompanyId, ctx)
		if err != nil {
			ctx.SendLogs(
				"AKSAddClusterModel:  Add previous cluster - "+err.Error(),
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
			"AKSAddClusterModel:  Add previous cluster - "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	defer session.Close()

	if cluster.CreationDate.IsZero() {
		cluster.CreationDate = time.Now()
		cluster.ModificationDate = time.Now()
		cluster.Cloud = models.AKS
	}

	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoAKSPreviousClusterCollection, oldCluster)
	if err != nil {
		ctx.SendLogs(
			"AKSAddClusterModel:  Add previous cluster -  "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func GetPreviousAKSCluster(infraId, companyId string, ctx utils.Context) (cluster AKSCluster, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs(
			"AKSGetClusterModel:  Get previous cluster - Got error while connecting to the database: "+err1.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAKSPreviousClusterCollection)
	err = c.Find(bson.M{"infra_id": infraId, "company_id": companyId}).One(&cluster)
	if err != nil {
		ctx.SendLogs(
			"AKSGetClusterModel:  Get previous cluster- Got error while fetching from database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err
	}

	return cluster, nil
}

func UpdatePreviousAKSCluster(cluster AKSCluster, ctx utils.Context) error {

	result, _ := GetAKSCluster(cluster.InfraId, cluster.CompanyId, ctx)
	if result.UpdateStatus != string(models.New) {
		err := AddPreviousAKSCluster(cluster, ctx, false)
		if err != nil {
			text := "AKSClusterModel:  Update  previous cluster -'" + cluster.Name + err.Error()
			ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return errors.New(text)
		}
	}

	cluster.UpdateStatus = string(models.New)
	err := UpdateAKSCluster(cluster, ctx)
	if err != nil {
		text := "AKSClusterModel:  Update previous cluster - '" + cluster.Name + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		err = DeletePreviousAKSCluster(cluster.InfraId, cluster.CompanyId, ctx)
		if err != nil {
			text := "AKSDeleteClusterModel:  Delete  previous cluster - '" + cluster.Name + err.Error()
			ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return errors.New(text)
		}
		return err
	}

	return nil
}

func DeletePreviousAKSCluster(infraId, companyId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"AKSDeleteClusterModel:  Delete  previous cluster - "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAKSPreviousClusterCollection)
	err = c.Remove(bson.M{"infra_id": infraId, "company_id": companyId})
	if err != nil {
		ctx.SendLogs(
			"AKSDeleteClusterModel:  Delete  previous cluster - "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func GetAKSCluster(infraId string, companyId string, ctx utils.Context) (cluster AKSCluster, err error) {
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
	err = c.Find(bson.M{"infra_id": infraId, "company_id": companyId}).One(&cluster)
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

func GetAllAKSCluster(data rbacAuthentication.List, ctx utils.Context) (aksClusters []Cluster, err error) {
	var clusters []AKSCluster
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
		return aksClusters, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAKSClusterCollection)
	err = c.Find(bson.M{"infra_id": bson.M{"$in": copyData}, "company_id": ctx.Data.Company}).All(&clusters)
	if err != nil {
		ctx.SendLogs(
			"AKSGetAllClusterModel:  GetAll - Got error while fetching from database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return aksClusters, err
	}
	for _, cluster := range clusters {
		temp := Cluster{Name: cluster.Name, InfraId: cluster.InfraId, Status: cluster.Status}
		aksClusters = append(aksClusters, temp)
	}

	return aksClusters, nil
}

func AddAKSCluster(cluster AKSCluster, ctx utils.Context) error {
	_, err := GetAKSCluster(cluster.InfraId, cluster.CompanyId, ctx)
	if err == nil {
		text := fmt.Sprintf("AKSAddClusterModel:  Add - Cluster for infrastructure '%s' already exists in the database.", cluster.InfraId)
		ctx.SendLogs(text+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text + err.Error())
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
	oldCluster, err := GetAKSCluster(cluster.InfraId, cluster.CompanyId, ctx)
	if err != nil {
		text := "AKSUpdateClusterModel:  Update - Cluster '" + cluster.Name + "' does not exist in the database: " + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	err = DeleteAKSCluster(cluster.InfraId, cluster.CompanyId, ctx)
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

func DeleteAKSCluster(infraId, companyId string, ctx utils.Context) error {
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
	err = c.Remove(bson.M{"infra_id": infraId, "company_id": companyId})
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

	/*publisher := utils.Notifier{}
	_ = publisher.Init_notifier()*/

	aksOps, _ := GetAKS(credentials.Profile)

	CpErr := aksOps.init()
	if CpErr != (types.CustomCPError{}) {
		ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+CpErr.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		cluster.Status = models.ClusterCreationFailed
		UpdationErr := UpdateAKSCluster(cluster, ctx)
		if UpdationErr != nil {
			_, _ = utils.SendLog(companyId, "Cluster creation failed : "+UpdationErr.Error(), "error", cluster.InfraId)

			ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+UpdationErr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		err := db.CreateError(cluster.InfraId, companyId, models.AKS, ctx, CpErr)
		if err != nil {

			ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: CpErr.Error + "\n" + CpErr.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)

		return CpErr
	}

	_, _ = utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.InfraId)
	cluster.Status = (models.Deploying)
	err_ := UpdateAKSCluster(cluster, ctx)
	if err_ != nil {
		utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.InfraId)

		CpErr = ApiError(err_, "Error occurred while updating cluster in database", 500)

		err := db.CreateError(cluster.InfraId, companyId, models.AKS, ctx, CpErr)
		if err != nil {
			ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: CpErr.Error + "\n" + CpErr.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return CpErr
	}
	err := aksOps.CreateCluster(cluster, token, ctx)
	if err != nil {
		cpErr := ApiError(err, "", 502)

		_, _ = utils.SendLog(companyId, "Cluster creation failed : "+cpErr.Error, "error", cluster.InfraId)
		_, _ = utils.SendLog(companyId, cpErr.Description, "error", cluster.InfraId)

		cluster.Status = models.ClusterCreationFailed
		UpdationErr := UpdateAKSCluster(cluster, ctx)
		if UpdationErr != nil {
			_, _ = utils.SendLog(companyId, "Cluster creation failed : "+UpdationErr.Error(), "error", cluster.InfraId)

			ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+UpdationErr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		err := db.CreateError(cluster.InfraId, companyId, models.AKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return cpErr
	}
	AgentErr := azure.ApplyAgent(credentials, token, ctx, cluster.Name, cluster.ResourceGoup)
	if AgentErr != nil {
		cpErr := ApiError(AgentErr, "agent deployment failed", 500)
		_, _ = utils.SendLog(companyId, "Cluster creation failed : "+cpErr.Error, "error", cluster.InfraId)
		_, _ = utils.SendLog(companyId, "Agent deployment failed : "+cpErr.Error+cpErr.Description, "error", cluster.InfraId)

		cluster.Status = models.ClusterCreationFailed
		utils.SendLog(companyId, "Cleaning up resources", "info", cluster.InfraId)
		_ = TerminateCluster(credentials, cluster.InfraId, companyId, token, ctx)
		UpdationErr := UpdateAKSCluster(cluster, ctx)
		if UpdationErr != nil {
			_, _ = utils.SendLog(companyId, "Cluster creation failed : "+UpdationErr.Error(), "error", cluster.InfraId)
			ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+UpdationErr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		err := db.CreateError(cluster.InfraId, companyId, models.AKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: AgentErr.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return cpErr
	}
	cluster.Status = models.ClusterCreated

	UpdationErr := UpdateAKSCluster(cluster, ctx)
	if UpdationErr != nil {
		CpErr = ApiError(err_, "Error occurred while updating cluster in database", 500)
		_, _ = utils.SendLog(companyId, "Cluster creation failed : "+UpdationErr.Error(), "error", cluster.InfraId)
		ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+UpdationErr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: UpdationErr.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		err := db.CreateError(cluster.InfraId, companyId, models.AKS, ctx, CpErr)
		if err != nil {
			ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return CpErr
	}
	utils.Publisher(utils.ResponseSchema{
		Status:  true,
		Message: "Cluster created successfully",
		InfraId: cluster.InfraId,
		Token:   token,
		Action:  models.Create,
	}, ctx)
	_, _ = utils.SendLog(companyId, "Cluster created successfully "+cluster.Name, "info", cluster.InfraId)

	return types.CustomCPError{}
}

func FetchStatus(credentials vault.AzureCredentials, token, infraId, companyId string, ctx utils.Context) (KubeClusterStatus, types.CustomCPError) {
	cluster, err := GetAKSCluster(infraId, companyId, ctx)
	if err != nil {
		return KubeClusterStatus{}, types.CustomCPError{Error: "Error occurred while getting cluster status in database",
			Description: err.Error(),
			StatusCode:  500}
	}
	if string(cluster.Status) == strings.ToLower(string(models.New)) {
		cpErr := types.CustomCPError{Error: "Unable to fetch status - Cluster is not deployed yet", Description: "Unable to fetch state - Cluster is not deployed yet", StatusCode: 409}
		return KubeClusterStatus{}, cpErr
	}
	if cluster.Status == models.Deploying || cluster.Status == models.Terminating || cluster.Status == models.ClusterTerminated {
		cpErr := types.CustomCPError{Error: "Cluster is in " +
			string(cluster.Status) + " state", Description: "Cluster is in " +
			string(cluster.Status) + " state", StatusCode: 409}
		return KubeClusterStatus{}, cpErr
	}
	if cluster.Status != models.ClusterCreated {
		customErr, err := db.GetError(cluster.InfraId, ctx.Data.Company, models.GKE, ctx)
		if err != nil {
			return KubeClusterStatus{}, types.CustomCPError{Error: "Error occurred while getting cluster status in database",
				Description: "Error occurred while getting cluster status in database",
				StatusCode:  500}
		}
		if customErr.Err != (types.CustomCPError{}) {
			return KubeClusterStatus{}, customErr.Err
		}
	}
	aksOps, _ := GetAKS(credentials)

	CpErr := aksOps.init()
	if CpErr != (types.CustomCPError{}) {
		ctx.SendLogs("AKSClusterModel:  Fetch -"+CpErr.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubeClusterStatus{}, CpErr
	}

	clusterstatus, CpErr := aksOps.fetchClusterStatus(credentials, &cluster, ctx)
	if CpErr != (types.CustomCPError{}) {
		return KubeClusterStatus{}, CpErr
	}

	return clusterstatus, types.CustomCPError{}
}

func TerminateCluster(credentials vault.AzureProfile, infraId, companyId, token string, ctx utils.Context) types.CustomCPError {
	publisher := utils.Notifier{}
	_ = publisher.Init_notifier()

	cluster, err := GetAKSCluster(infraId, companyId, ctx)
	if err != nil {
		cpErr := ApiError(err, "Error wile getting cluster from database", 500)
		err := db.CreateError(cluster.InfraId, companyId, models.AKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("AKSDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return cpErr
	}

	if cluster.Status == "" || cluster.Status == "new" {

		text := "AKSClusterModel : Terminate - Cannot terminate a new cluster"
		cpErr := ApiError(errors.New(text), text, 400)
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		err := db.CreateError(cluster.InfraId, companyId, models.AKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("AKSDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: text,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Terminate,
		}, ctx)
		return cpErr
	}

	aksOps, _ := GetAKS(credentials.Profile)
	_, _ = utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.InfraId)

	cluster.Status = (models.Terminating)
	err_ := UpdateAKSCluster(cluster, ctx)
	if err_ != nil {
		cpErr := ApiError(err_, "Error while updating cluster in database", 500)
		utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.InfraId)
		err := db.CreateError(cluster.InfraId, companyId, models.AKS, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("AKSDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err_.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Terminate,
		}, ctx)
		return cpErr
	}

	CpErr := aksOps.init()
	if CpErr != (types.CustomCPError{}) {
		ctx.SendLogs("AKSClusterModel : Terminate -"+CpErr.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		utils.SendLog(companyId, CpErr.Description, "error", cluster.InfraId)

		cluster.Status = models.ClusterTerminationFailed
		err = UpdateAKSCluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("AKSClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		err := db.CreateError(cluster.InfraId, companyId, models.AKS, ctx, CpErr)
		if err != nil {
			ctx.SendLogs("AKSDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: CpErr.Error + "\n" + CpErr.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Terminate,
		}, ctx)

		return CpErr
	}

	CpErr = aksOps.TerminateCluster(cluster, ctx)
	if CpErr != (types.CustomCPError{}) {

		_, _ = utils.SendLog(companyId, "Cluster termination failed: "+CpErr.Error, "error", cluster.InfraId)
		_, _ = utils.SendLog(companyId, CpErr.Description, "error", cluster.InfraId)

		cluster.Status = models.ClusterTerminationFailed
		err = UpdateAKSCluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("AKSClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		err := db.CreateError(cluster.InfraId, companyId, models.AKS, ctx, CpErr)
		if err != nil {
			ctx.SendLogs("AKSDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: CpErr.Error + "\n" + CpErr.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Terminate,
		}, ctx)
		return CpErr
	}

	cluster.Status = models.ClusterTerminated

	err = UpdateAKSCluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterModel : Terminate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		_, _ = utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.InfraId)
		_, _ = utils.SendLog(companyId, err.Error(), "error", cluster.InfraId)

		CpErr := ApiError(err, "Error while updating cluster in database", 500)

		err := db.CreateError(cluster.InfraId, companyId, models.AKS, ctx, CpErr)
		if err != nil {
			ctx.SendLogs("AKSDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: CpErr.Error + "\n" + CpErr.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Terminate,
		}, ctx)
		return ApiError(err, "Error while updating cluster in database", 500)
	}
	_, _ = utils.SendLog(companyId, "Cluster terminated successfully "+cluster.Name, "info", cluster.InfraId)
	utils.Publisher(utils.ResponseSchema{
		Status:  true,
		Message: "Cluster terminated successfully",
		InfraId: cluster.InfraId,
		Token:   token,
		Action:  models.Terminate,
	}, ctx)
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

func PrintError(confError error, name, infraId string, companyId string) {
	if confError != nil {
		beego.Error(confError.Error())
		_, _ = utils.SendLog(companyId, "Cluster creation failed : "+name, "error", infraId)
		_, _ = utils.SendLog(companyId, confError.Error(), "error", infraId)
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

	sort.Slice(versions, func(i, j int) bool {
		return versions[i] > versions[j]
	})
	return versions, types.CustomCPError{}

}

func ValidateAKSData(cluster *AKSCluster, ctx utils.Context) error {
	if !cluster.IsAdvanced {
		cluster.DNSPrefix = cluster.Name + "-dns"
	}
	if cluster.InfraId == "" {

		return errors.New("infrastructure ID is empty")

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

	if len(cluster.AgentPoolProfiles) == 0 {

		return errors.New("length of node pools must be greater than zero")

	} else if cluster.IsAdvanced {

		if cluster.KubernetesVersion == "" {

			return errors.New("kubernetes version is empty")

		} else if cluster.DNSPrefix == "" {

			return errors.New("DNS prefix is empty")

		} else if cluster.IsServicePrincipal {

			if cluster.ClientID == "" || cluster.Secret == "" {

				return errors.New("client id or secret is empty")

			}
		}

		for _, pool := range cluster.AgentPoolProfiles {

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

	if cluster.IsExpert {
		if cluster.PodCidr == "" {

			return errors.New("pod CIDR must not be empty")

		} else {

			isValidCidr := ipv4.IsIPv4(cluster.PodCidr)
			if !isValidCidr {
				return errors.New("pod CIDR is not valid")
			}

		}

		if cluster.DNSServiceIP == "" {

			return errors.New("DNS service IP must not be empty")

		} else {

			isValidIp := ipv4.IsIPv4(cluster.DNSServiceIP)
			if !isValidIp {
				return errors.New("DNS service IP is not valid")
			}

		}

		if cluster.DockerBridgeCidr == "" {

			return errors.New("Docker Bridge CIDR must not be empty")

		} else {

			isValidCidr := ipv4.IsIPv4(cluster.DockerBridgeCidr)
			if !isValidCidr {
				return errors.New("docker bridge CIDR is not valid")
			}

		}

		if cluster.ServiceCidr == "" {

			return errors.New("Service CIDR must not be empty")

		} else {

			isValidCidr := ipv4.IsIPv4(cluster.ServiceCidr)
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
		if v1.Location == region {
			return true, nil
		}
	}

	var errData string
	for _, v1 := range regionList {
		errData += v1.Location + ", "
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

func PatchRunningAKSCluster(cluster AKSCluster, credentials vault.AzureProfile, companyId, token string, ctx utils.Context) (confError types.CustomCPError) {

	/*publisher := utils.Notifier{}
	_ = publisher.Init_notifier()*/

	aksOps, _ := GetAKS(credentials.Profile)
	CpErr := aksOps.init()
	if CpErr != (types.CustomCPError{}) {
		ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+CpErr.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		cluster.Status = models.ClusterUpdateFailed
		UpdationErr := UpdateAKSCluster(cluster, ctx)
		if UpdationErr != nil {
			_, _ = utils.SendLog(companyId, "Cluster creation failed : "+UpdationErr.Error(), "error", cluster.InfraId)

			ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+UpdationErr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		err := db.CreateError(cluster.InfraId, companyId, models.AKS, ctx, CpErr)
		if err != nil {

			ctx.SendLogs("AKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: CpErr.Error + "\n" + CpErr.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Update,
		}, ctx)

		return CpErr
	}

	difCluster, previousPoolCount, newPoolCount, err1 := CompareClusters(cluster.InfraId, companyId, ctx)
	if err1 != nil {
		if strings.Contains(err1.Error(), "Nothing to update") {
			cluster.Status = models.ClusterCreated
			confError_ := UpdateAKSCluster(cluster, ctx)
			if confError_ != nil {
				ctx.SendLogs("AKSRunningClusterModel:"+confError_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
	}

	_, _ = utils.SendLog(ctx.Data.Company, "Updating running cluster : "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

	aksOps.infraId = cluster.InfraId
	isClusterUpdated := false
	if previousPoolCount < newPoolCount {
		isPoolUpdated := make(map[int]bool)
		for poolIndex, nodePool := range cluster.AgentPoolProfiles {
			isPoolUpdated[poolIndex] = false
			for _, diff := range difCluster {
				if diff.Path[0] == "ID" || diff.Path[0] == "ModificationDate" || diff.Path[0] == "Status" || diff.Path[0] == "IsAdvanced" || diff.Path[0] == "IsExpert" || diff.Path[0] == "ClusterUpdated" {
					continue
				}
				if diff.Type == "create" && diff.Path[0] == "AgentPoolProfiles" && diff.Path[2] == "Name" && *diff.To.(*string) == *nodePool.Name {
					err := aksOps.CreatOrUpdateAgentPool(ctx, token, cluster.ResourceGoup, cluster.Name, nodePool)
					if err != nil {
						ctx.SendLogs("AKSRunningClusterModel:  Update - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
						updationFailedError(cluster, ctx, types.CustomCPError{
							StatusCode:  int(models.CloudStatusCode),
							Error:       "AKS cluster updation failed",
							Description: err.Error(),
						})
						return
					}
				} else if (diff.Type == "update" || diff.Type == "delete" || diff.Type == "create") && diff.Path[0] == "AgentPoolProfiles" && diff.Path[1] == strconv.Itoa(poolIndex) && !isPoolUpdated[poolIndex] {
					err := aksOps.CreatOrUpdateAgentPool(ctx, token, cluster.ResourceGoup, cluster.Name, nodePool)
					if err != nil {
						ctx.SendLogs("AKSRunningClusterModel:  Update - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
						updationFailedError(cluster, ctx, types.CustomCPError{
							StatusCode:  int(models.CloudStatusCode),
							Error:       "AKS cluster updation failed",
							Description: err.Error(),
						})
						return
					}
					isPoolUpdated[poolIndex] = true
				}
			}
		}
	} else if newPoolCount < previousPoolCount {
		previousCluster, err := GetPreviousAKSCluster(cluster.InfraId, companyId, ctx)
		if err != nil {
			updationFailedError(cluster, ctx, types.CustomCPError{
				StatusCode:  int(models.CloudStatusCode),
				Error:       "AKS cluster updation failed",
				Description: err.Error(),
			})
			ctx.SendLogs("Error in updating running cluster: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return types.CustomCPError{Error: "Error in updating running cluster", StatusCode: 512, Description: err.Error()}
		}
		for _, nodePool := range previousCluster.AgentPoolProfiles {
			for _, diff := range difCluster {
				if diff.Path[0] == "ID" || diff.Path[0] == "ModificationDate" || diff.Path[0] == "Status" || diff.Path[0] == "IsAdvanced" || diff.Path[0] == "IsExpert" || diff.Path[0] == "ClusterUpdated" {
					continue
				}
				if diff.Type == "delete" && diff.Path[0] == "AgentPoolProfiles" && diff.Path[2] == "Name" && *diff.From.(*string) == *nodePool.Name {
					err := aksOps.DeleteAgentPool(ctx, cluster.ResourceGoup, cluster.Name, nodePool)
					if err != nil {
						ctx.SendLogs("AKSRunningClusterModel:  Update - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
						updationFailedError(cluster, ctx, types.CustomCPError{
							StatusCode:  int(models.CloudStatusCode),
							Error:       "AKS cluster updation failed",
							Description: err.Error(),
						})
						return
					}
				}
			}
		}

		//handling nodepool update in case if one nodepool deleted and other is updated
		isPoolUpdated := make(map[int]bool)
		for poolIndex, nodePool := range cluster.AgentPoolProfiles {
			isPoolUpdated[poolIndex] = false
			for _, diff := range difCluster {
				if diff.Path[0] == "ID" || diff.Path[0] == "ModificationDate" || diff.Path[0] == "Status" || diff.Path[0] == "IsAdvanced" || diff.Path[0] == "IsExpert" || diff.Path[0] == "ClusterUpdated" {
					continue
				}
				if (diff.Type == "update" || diff.Type == "delete" || diff.Type == "create") && diff.Path[0] == "AgentPoolProfiles" && diff.Path[1] == strconv.Itoa(poolIndex) && !isPoolUpdated[poolIndex] {
					err := aksOps.CreatOrUpdateAgentPool(ctx, token, cluster.ResourceGoup, cluster.Name, nodePool)
					if err != nil {
						ctx.SendLogs("AKSRunningClusterModel:  Update - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
						updationFailedError(cluster, ctx, types.CustomCPError{
							StatusCode:  int(models.CloudStatusCode),
							Error:       "AKS cluster updation failed",
							Description: err.Error(),
						})
						return
					}
					isPoolUpdated[poolIndex] = true
				}
			}
		}
	} else if newPoolCount == previousPoolCount {
		isPoolUpdated := make(map[int]bool)
		for poolIndex, nodePool := range cluster.AgentPoolProfiles {
			isPoolUpdated[poolIndex] = false
			for _, diff := range difCluster {
				if diff.Path[0] == "ID" || diff.Path[0] == "ModificationDate" || diff.Path[0] == "Status" || diff.Path[0] == "IsAdvanced" || diff.Path[0] == "IsExpert" || diff.Path[0] == "ClusterUpdated" {
					continue
				}

				//If user delete existing nodepool and create again with same nodepool name
				if diff.Type == "update" && diff.Path[0] == "AgentPoolProfiles" && diff.Path[1] == strconv.Itoa(poolIndex) && (diff.Path[2] == "VMSize" || diff.Path[2] == "OsDiskSizeGB" || diff.Path[2] == "MaxPods" || diff.Path[2] == "Name") && !isPoolUpdated[poolIndex] {
					poolName := nodePool.Name
					if diff.Path[2] == "Name" {
						if _, ok := diff.To.(string); ok {
							v := diff.From.(string)
							poolName = &v
						}
					}
					err := aksOps.DeleteAgentPool(ctx, cluster.ResourceGoup, cluster.Name, ManagedClusterAgentPoolProfile{Name: poolName})
					if err != nil {
						ctx.SendLogs("AKSRunningClusterModel:  Update - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
						updationFailedError(cluster, ctx, types.CustomCPError{
							StatusCode:  int(models.CloudStatusCode),
							Error:       "AKS cluster updation failed",
							Description: err.Error(),
						})
						return
					}

					err = aksOps.CreatOrUpdateAgentPool(ctx, token, cluster.ResourceGoup, cluster.Name, nodePool)
					if err != nil {
						ctx.SendLogs("AKSRunningClusterModel:  Update - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
						updationFailedError(cluster, ctx, types.CustomCPError{
							StatusCode:  int(models.CloudStatusCode),
							Error:       "AKS cluster updation failed",
							Description: err.Error(),
						})
						return
					}
					isPoolUpdated[poolIndex] = true
				}

				if (diff.Type == "update" || diff.Type == "create" || diff.Type == "delete") && diff.Path[0] == "AgentPoolProfiles" && diff.Path[1] == strconv.Itoa(poolIndex) && !isPoolUpdated[poolIndex] {
					err := aksOps.CreatOrUpdateAgentPool(ctx, token, cluster.ResourceGoup, cluster.Name, nodePool)
					if err != nil {
						ctx.SendLogs("AKSRunningClusterModel:  Update - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
						updationFailedError(cluster, ctx, types.CustomCPError{
							StatusCode:  int(models.CloudStatusCode),
							Error:       "AKS cluster updation failed",
							Description: err.Error(),
						})
						return
					}
					isPoolUpdated[poolIndex] = true
				}
			}
		}
	}

	isClusterUpdated = false
	for _, diff := range difCluster {
		if diff.Path[0] == "ID" || diff.Path[0] == "ModificationDate" || diff.Path[0] == "Status" || diff.Path[0] == "IsAdvanced" || diff.Path[0] == "IsExpert" || diff.Path[0] == "ClusterUpdated" {
			continue
		} else if (diff.Type == "update" || diff.Type == "create" || diff.Type == "delete") && diff.Path[0] != "AgentPoolProfiles" && !isClusterUpdated {
			err := aksOps.CreateCluster(cluster, token, ctx)
			if err != nil {
				ctx.SendLogs("AKSRunningClusterModel:  Update - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				updationFailedError(cluster, ctx, types.CustomCPError{
					StatusCode:  int(models.CloudStatusCode),
					Error:       "AKS cluster updation failed",
					Description: err.Error(),
				})
				return
			}
			isClusterUpdated = true
		}
	}

	//_ = DeletePreviousAKSCluster(cluster.InfraId, companyId, ctx)

	//latestCluster, err2 := aksOps.fetchClusterStatus(credentials.Profile, &cluster, ctx)
	//if err2 != (types.CustomCPError{}) {
	//	return
	//}
	//
	//for strings.ToLower(string(latestCluster.Status)) != strings.ToLower("running") {
	//	time.Sleep(time.Second * 60)
	//}

	cluster.Status = models.ClusterCreated
	cluster.UpdateStatus = string(models.ClusterUpdated)
	confError_ := UpdateAKSCluster(cluster, ctx)
	if confError_ != nil {
		ctx.SendLogs("AKSRunningClusterModel:"+confError_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

	}

	_, _ = utils.SendLog(ctx.Data.Company, "Running Cluster updated successfully "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
	utils.Publisher(utils.ResponseSchema{
		Status:  true,
		Message: "Cluster updated sucessfully",
		InfraId: cluster.InfraId,
		Token:   token,
		Action:  models.Update,
	}, ctx)

	return types.CustomCPError{}

}

func updationFailedError(cluster AKSCluster, ctx utils.Context, err types.CustomCPError) types.CustomCPError {
	publisher := utils.Notifier{}
	_ = publisher.Init_notifier()

	cluster.Status = models.ClusterUpdateFailed
	confError := UpdateAKSCluster(cluster, ctx)
	if confError != nil {
		ctx.SendLogs("AKSRunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	}

	utils.SendLog(ctx.Data.Company, "Error in running cluster update : "+err.Description, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)

	err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.AKS, ctx, err)
	if err_ != nil {
		ctx.SendLogs("AKSRunningClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	}

	utils.SendLog(ctx.Data.Company, "Deployed cluster update failed : "+cluster.Name, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
	utils.SendLog(ctx.Data.Company, err.Description, models.LOGGING_LEVEL_ERROR, ctx.Data.Company)

	publisher.Notify(ctx.Data.InfraId, "Redeploy Status Available", ctx)
	return err
}

func CompareClusters(infraId, companyId string, ctx utils.Context) (diff.Changelog, int, int, error) {
	cluster, err := GetAKSCluster(infraId, companyId, ctx)
	if err != nil {
		return diff.Changelog{}, 0, 0, err
	}

	oldCluster, err := GetPreviousAKSCluster(infraId, companyId, ctx)
	if err != nil && strings.Contains(err.Error(), "not found") {
		return diff.Changelog{}, 0, 0, errors.New("Nothing to update")
	}

	previousPoolCount := len(oldCluster.AgentPoolProfiles)
	newPoolCount := len(cluster.AgentPoolProfiles)

	difCluster, err := diff.Diff(oldCluster, cluster)
	if len(difCluster) < 2 && previousPoolCount == newPoolCount {
		return diff.Changelog{}, 0, 0, errors.New("Nothing to update")
	} else if err != nil {
		return diff.Changelog{}, 0, 0, errors.New("Error in comparing differences:" + err.Error())
	}
	return difCluster, previousPoolCount, newPoolCount, nil
}
