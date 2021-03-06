package gke

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/db"
	"antelope/models/gcp"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/woodpecker"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/r3labs/diff"
	gke "google.golang.org/api/container/v1"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
	"time"
)

type GKECluster struct {
	ID                             bson.ObjectId                   `json:"-" bson:"_id,omitempty"`
	InfraId                        string                          `json:"infra_id" bson:"infra_id" validate:"required" description:"ID of infrastructure [required]"`
	Cloud                          models.Cloud                    `json:"cloud" bson:"cloud"`
	CreationDate                   time.Time                       `json:"-" bson:"creation_date"`
	ModificationDate               time.Time                       `json:"-" bson:"modification_date"`
	CloudplexStatus                models.Type                     `json:"status" bson:"status" validate:"eq=new|eq=New|eq=NEW|eq=Cluster Creation Failed|eq=Cluster Terminated|eq=Cluster Created|eq=Cluster Update Failed" description:"Status of cluster [optional]"`
	CompanyId                      string                          `json:"company_id" bson:"company_id" description:"ID of company [optional]"`
	IsExpert                       bool                            `json:"is_expert" bson:"is_expert"`
	IsAdvance                      bool                            `json:"is_advance" bson:"is_advance"`
	AddonsConfig                   *AddonsConfig                   `json:"addons_config,omitempty" bson:"addons_config,omitempty"`
	ClusterIpv4Cidr                string                          `json:"cluster_ipv4_cidr,omitempty" bson:"cluster_ipv4_cidr,omitempty" description:"The IP address range of the container pods in this cluster [optional]"`
	Conditions                     []*StatusCondition              `json:"conditions,omitempty" bson:"conditions,omitempty"`
	CreateTime                     string                          `json:"create_time,omitempty" bson:"create_time,omitempty" description:"The time the cluster was created [readonly]"`
	CurrentMasterVersion           string                          `json:"current_master_version,omitempty" bson:"current_master_version,omitempty" description:"The current software version of master endpoint [readonly]"`
	CurrentNodeCount               int64                           `json:"current_node_count,omitempty" bson:"current_node_count,omitempty" description:"The number of nodes currently in the cluster [readonly]"`
	DefaultMaxPodsConstraint       *MaxPodsConstraint              `json:"default_max_pods_constrainty" bson:"default_max_pods_constraint" `
	Description                    string                          `json:"description,omitempty" bson:"description,omitempty" description:"An optional description of this cluster [optional]"`
	EnableKubernetesAlpha          bool                            `json:"enable_kubernetes_alpha,omitempty" bson:"enable_kubernetes_alpha,omitempty" description:"Alpha enabled clusters are automatically deleted thirty days after [optional]"`
	EnableTpu                      bool                            `json:"enable_tpu,omitempty" bson:"enable_tpu,omitempty" description:"Enable the ability to use Cloud TPUs in this cluster [optional]"`
	Endpoint                       string                          `json:"endpoint,omitempty" bson:"endpoint,omitempty" description:"IP address of this cluster's master endpoint [readonly]"`
	ExpireTime                     string                          `json:"expire_time,omitempty" bson:"expire_time,omitempty" description:"Time the cluster will be automatically deleted [readonly]"`
	InitialClusterVersion          string                          `json:"initial_cluster_version,omitempty" bson:"initial_cluster_version,omitempty" description:"Initial kubernetes version for this cluster [optional]"`
	IpAllocationPolicy             *IPAllocationPolicy             `json:"ip_allocation_policy" bson:"ip_allocation_policy"`
	LabelFingerprint               string                          `json:"label_fingerprint,omitempty" bson:"label_fingerprint,omitempty" description:"The fingerprint of the set of labels for this cluster [optional]"`
	LegacyAbac                     *LegacyAbac                     `json:"legacy_abac,omitempty" bson:"legacy_abac,omitempty"`
	Location                       string                          `json:"location" bson:"location"  description:"The name of GCP zone or region in which cluster resides [required]"`
	Locations                      []string                        `json:"locations,omitempty" bson:"locations,omitempty" description:"The name of GCP zones in which cluster nodes located [optional]"`
	LoggingService                 string                          `json:"logging_service,omitempty" bson:"logging_service,omitempty" description:"The logging service the cluster should use to write logs [optional]"`
	MaintenancePolicy              *MaintenancePolicy              `json:"maintenance_policy,omitempty" bson:"maintenance_policy,omitempty"`
	MasterAuth                     *MasterAuth                     `json:"master_auth,omitempty" bson:"master_auth,omitempty"`
	MasterAuthorizedNetworksConfig *MasterAuthorizedNetworksConfig `json:"master_authorized_networks_config,omitempty" bson:"master_authorized_networks_config,omitempty"`
	MonitoringService              string                          `json:"monitoring_service,omitempty" bson:"monitoring_service,omitempty" description:"The monitoring service the cluster should use to write metrics [optional]"`
	Name                           string                          `json:"name" bson:"name" validate:"required" description:"The name of this cluster [required]"`
	Network                        string                          `json:"network" bson:"network" description:"The name of GCP network to which the cluster connected [required]"`
	NetworkConfig                  *NetworkConfig                  `json:"network_config,omitempty" bson:"network_config,omitempty"`
	NetworkPolicy                  *NetworkPolicy                  `json:"network_policy,omitempty" bson:"network_policy,omitempty"`
	NodeIpv4CidrSize               int64                           `json:"node_ipv4_cidr_size,omitempty" bson:"node_ipv4_cidr_size,omitempty" description:"The size of the address space on each node [readonly]"`
	NodePools                      []*NodePool                     `json:"node_pools" bson:"node_pools" validate:"required,dive"`
	PrivateClusterConfig           *PrivateClusterConfig           `json:"private_cluster_config,omitempty" bson:"private_cluster_config,omitempty"`
	ResourceLabels                 map[string]string               `json:"resource_labels,omitempty" bson:"resource_labels,omitempty"`
	ResourceUsageExportConfig      *ResourceUsageExportConfig      `json:"resource_usage_export_config,omitempty" bson:"resource_usage_export_config,omitempty"`
	SelfLink                       string                          `json:"self_link,omitempty" bson:"self_link,omitempty" description:"Server-defined URL for the resource [readonly]"`
	ServicesIpv4Cidr               string                          `json:"services_ipv4_cidr,omitempty" bson:"services_ipv4_cidr,omitempty" description:"The IP address range of the kubernetes services in the cluster [readonly]"`
	Status                         models.Type                     `json:"cloud_status,omitempty" bson:"cloud_status,omitempty" description:"The current status of this cluster [readonly]"`
	StatusMessage                  string                          `json:"status_message,omitempty" bson:"status_message,omitempty" description:"Additional information about the current status [readonly]"`
	Subnetwork                     string                          `json:"subnetwork,omitempty" bson:"subnetwork,omitempty" description:"The name of the GCP subnetwork cluster is connected to [required]"`
	TpuIpv4CidrBlock               string                          `json:"tpu_ipv4_cidr_block,omitempty" bson:"tpu_ipv4_cidr_block,omitempty" description:"The IP address range of the Cloud TPUs in the cluster [readonly]"`
	Zone                           string                          `json:"zone" bson:"zone" validate:"required" description:"Name of GCP zone in which cluster resides [required]"`
}

type AddonsConfig struct {
	HorizontalPodAutoscaling *HorizontalPodAutoscaling `json:"horizontal_pod_autoscaling,omitempty" bson:"horizontal_pod_autoscaling,omitempty"`
	HttpLoadBalancing        *HttpLoadBalancing        `json:"http_load_balancing,omitempty" bson:"http_load_balancing,omitempty"`
	KubernetesDashboard      *KubernetesDashboard      `json:"kubernetes_dashboard,omitempty" bson:"kubernetes_dashboard,omitempty"`
	NetworkPolicyConfig      *NetworkPolicyConfig      `json:"network_policy_config,omitempty" bson:"network_policy_config,omitempty"`
}

type HorizontalPodAutoscaling struct {
	Disabled bool `json:"disabled,omitempty" bson:"disabled,omitempty" description:"Whether the Horizontal Pod Autoscaling feature is enabled in cluster [optional]"`
}

type HttpLoadBalancing struct {
	Disabled bool `json:"disabled,omitempty" bson:"disabled,omitempty" description:"Whether the HTTP Load Balancing controller is enabled in cluster [optional]"`
}

type KubernetesDashboard struct {
	Disabled bool `json:"disabled,omitempty" bson:"disabled,omitempty" description:"Whether the Kubernetes Dashboard is enabled for this cluster [optional]"`
}

type NetworkPolicyConfig struct {
	Disabled bool `json:"disabled,omitempty" bson:"disabled,omitempty" description:"Whether NetworkPolicy is enabled for this cluster [optional]"`
}

type StatusCondition struct {
	Code    string `json:"code,omitempty" bson:"code,omitempty"`
	Message string `json:"message,omitempty" bson:"message,omitempty"`
}

type MaxPodsConstraint struct {
	MaxPodsPerNode int64 `json:"max_pods_per_node" bson:"max_pods_per_node"  description:"Constraint enforced on the max num of pods per node [required]"`
}

type IPAllocationPolicy struct {
	ClusterIpv4Cidr            string `json:"cluster_ipv4_cidr" bson:"cluster_ipv4_cidr" `
	ClusterIpv4CidrBlock       string `json:"cluster_ipv4_cidr_block" bson:"cluster_ipv4_cidr_block" `
	ClusterSecondaryRangeName  string `json:"cluster_secondary_range_name,omitempty" bson:"cluster_secondary_range_name,omitempty"`
	CreateSubnetwork           bool   `json:"create_subnetwork,omitempty" bson:"create_subnetwork,omitempty"`
	NodeIpv4Cidr               string `json:"node_ipv4_cidr" bson:"node_ipv4_cidr" `
	NodeIpv4CidrBlock          string `json:"node_ipv4_cidr_block" bson:"node_ipv4_cidr_block" `
	ServicesIpv4Cidr           string `json:"services_ipv4_cidr" bson:"services_ipv4_cidr" `
	ServicesIpv4CidrBlock      string `json:"services_ipv4_cidr_block" bson:"services_ipv4_cidr_block" `
	ServicesSecondaryRangeName string `json:"services_secondary_range_name,omitempty" bson:"services_secondary_range_name,omitempty"`
	SubnetworkName             string `json:"subnetwork_name,omitempty" bson:"subnetwork_name,omitempty"`
	TpuIpv4CidrBlock           string `json:"tpu_ipv4_cidr_block" bson:"tpu_ipv4_cidr_block" `
	UseIpAliases               bool   `json:"use_ip_aliases,omitempty" bson:"use_ip_aliases,omitempty"`
}

type LegacyAbac struct {
	Enabled bool `json:"enabled,omitempty" bson:"enabled,omitempty"`
}

type MaintenancePolicy struct {
	Window *MaintenanceWindow `json:"window,omitempty" bson:"window,omitempty"`
}

type MaintenanceWindow struct {
	DailyMaintenanceWindow *DailyMaintenanceWindow `json:"daily_maintenance_window,omitempty" bson:"daily_maintenance_window,omitempty"`
}

type DailyMaintenanceWindow struct {
	Duration  string `json:"duration,omitempty" bson:"duration,omitempty"`
	StartTime string `json:"start_time,omitempty" bson:"start_time,omitempty"`
}

type MasterAuth struct {
	ClientCertificate       string                   `json:"client_certificate,omitempty" bson:"client_certificate,omitempty"`
	ClientCertificateConfig *ClientCertificateConfig `json:"client_certificate_config,omitempty" bson:"client_certificate_config,omitempty"`
	ClientKey               string                   `json:"client_key,omitempty" bson:"client_key,omitempty"`
	ClusterCaCertificate    string                   `json:"cluster_ca_certificate,omitempty" bson:"cluster_ca_certificate,omitempty"`
	Password                string                   `json:"password,omitempty" bson:"password,omitempty"`
	Username                string                   `json:"username,omitempty" bson:"username,omitempty"`
}

type ClientCertificateConfig struct {
	IssueClientCertificate bool `json:"issue_client_certificate,omitempty" bson:"issue_client_certificate,omitempty"`
}

type MasterAuthorizedNetworksConfig struct {
	CidrBlocks []*CidrBlock `json:"cidr_blocks,omitempty" bson:"cidr_blocks,omitempty"`
	Enabled    bool         `json:"enabled,omitempty" bson:"enabled,omitempty"`
}

type CidrBlock struct {
	CidrBlock   string `json:"cidr_block,omitempty" bson:"cidr_block,omitempty"`
	DisplayName string `json:"display_name,omitempty" bson:"display_name,omitempty"`
}

type NetworkConfig struct {
	Network    string `json:"network,omitempty" bson:"network,omitempty"`
	Subnetwork string `json:"subnetwork,omitempty" bson:"subnetwork,omitempty"`
}

type NetworkPolicy struct {
	Enabled  bool   `json:"enabled,omitempty" bson:"enabled,omitempty"`
	Provider string `json:"provider,omitempty" bson:"provider,omitempty"`
}

type PrivateClusterConfig struct {
	EnablePrivateEndpoint bool   `json:"enable_private_endpoint,omitempty" bson:"enable_private_endpoint,omitempty"`
	EnablePrivateNodes    bool   `json:"enable_private_nodes,omitempty" bson:"enable_private_nodes,omitempty"`
	MasterIpv4CidrBlock   string `json:"master_ipv4_cidr_block,omitempty" bson:"master_ipv4_cidr_block,omitempty"`
	PrivateEndpoint       string `json:"private_endpoint,omitempty" bson:"private_endpoint,omitempty"`
	PublicEndpoint        string `json:"public_endpoint,omitempty" bson:"public_endpoint,omitempty"`
}

type ResourceUsageExportConfig struct {
	BigqueryDestination         *BigQueryDestination       `json:"bigquery_destination,omitempty" bson:"bigquery_destination,omitempty"`
	ConsumptionMeteringConfig   *ConsumptionMeteringConfig `json:"consumption_metering_config,omitempty" bson:"consumption_metering_config,omitempty"`
	EnableNetworkEgressMetering bool                       `json:"enable_network_egress_metering,omitempty" bson:"enable_network_egress_metering,omitempty"`
}

type BigQueryDestination struct {
	DatasetId string `json:"dataset_id,omitempty" bson:"dataset_id,omitempty"`
}

type ConsumptionMeteringConfig struct {
	Enabled bool `json:"enabled,omitempty" bson:"enabled,omitempty"`
}

type NodePool struct {
	Autoscaling       *NodePoolAutoscaling `json:"autoscaling,omitempty" bson:"autoscaling,omitempty"`
	Conditions        []*StatusCondition   `json:"conditions,omitempty" bson:"conditions,omitempty"`
	Config            *NodeConfig          `json:"config" bson:"config" validate:"required,dive"`
	InitialNodeCount  int64                `json:"initial_node_count" bson:"initial_node_count" validate:"required,gte=1"`
	InstanceGroupUrls []string             `json:"instance_group_urls,omitempty" bson:"instance_group_urls,omitempty"`
	Management        *NodeManagement      `json:"management,omitempty" bson:"management,omitempty"`
	MaxPodsConstraint *MaxPodsConstraint   `json:"max_pods_constraint,omitempty" bson:"max_pods_constraint,omitempty"`
	Name              string               `json:"name" bson:"name" validate:"required"`
	PodIpv4CidrSize   int64                `json:"pod_ipv4_cidr_size,omitempty" bson:"pod_ipv4_cidr_size,omitempty"`
	SelfLink          string               `json:"self_link,omitempty" bson:"self_link,omitempty"`
	Status            string               `json:"status,omitempty" bson:"status,omitempty"`
	StatusMessage     string               `json:"status_message,omitempty" bson:"status_message,omitempty"`
	Version           string               `json:"version,omitempty" bson:"version,omitempty"`
	PoolStatus        bool                 `json:"pool_status,omitempty" bson:"pool_status,omitempty"`
}

type NodePoolAutoscaling struct {
	Enabled      bool  `json:"enabled,omitempty" bson:"enabled"`
	MaxNodeCount int64 `json:"max_node_count,omitempty" bson:"max_node_count,omitempty"`
	MinNodeCount int64 `json:"min_node_count,omitempty" bson:"min_node_count,omitempty"`
}

type NodeConfig struct {
	Accelerators   []*AcceleratorConfig `json:"accelerators,omitempty" bson:"accelerators,omitempty"`
	DiskSizeGb     int64                `json:"disk_size_gb" bson:"disk_size_gb" validate:"required,gte=30"`
	DiskType       string               `json:"disk_type" bson:"disk_type" validate:"required"`
	ImageType      string               `json:"image_type" bson:"image_type" validate:"required"`
	Labels         map[string]string    `json:"labels,omitempty" bson:"labels,omitempty"`
	LocalSsdCount  int64                `json:"local_ssd_count,omitempty" bson:"local_ssd_count,omitempty"`
	MachineType    string               `json:"machine_type" bson:"machine_type" validate:"required"`
	Metadata       map[string]string    `json:"metadata,omitempty" bson:"metadata,omitempty"`
	MinCpuPlatform string               `json:"min_cpu_platform,omitempty" bson:"min_cpu_platform,omitempty"`
	OauthScopes    []string             `json:"oauth_scopes,omitempty" bson:"oauth_scopes,omitempty"`
	Preemptible    bool                 `json:"preemptible,omitempty" bson:"preemptible,omitempty"`
	ServiceAccount string               `json:"service_account" bson:"service_account" validate:"required"`
	Tags           []string             `json:"tags,omitempty" bson:"tags,omitempty"`
	Taints         []*NodeTaint         `json:"taints,omitempty" bson:"taints,omitempty"`
}

type AcceleratorConfig struct {
	AcceleratorCount int64  `json:"accelerator_count,omitempty" bson:"accelerator_count,omitempty"`
	AcceleratorType  string `json:"accelerator_type,omitempty" bson:"accelerator_type,omitempty"`
}

type NodeTaint struct {
	Effect string `json:"effect,omitempty" bson:"effect,omitempty"`
	Key    string `json:"key,omitempty" bson:"key,omitempty"`
	Value  string `json:"value,omitempty" bson:"value,omitempty"`
}

type NodeManagement struct {
	AutoRepair     bool                `json:"auto_repair,omitempty" bson:"auto_repair,omitempty"`
	AutoUpgrade    bool                `json:"auto_upgrade,omitempty" bson:"auto_upgrade,omitempty"`
	UpgradeOptions *AutoUpgradeOptions `json:"upgrade_options,omitempty" bson:"upgrade_options,omitempty"`
}

type AutoUpgradeOptions struct {
	AutoUpgradeStartTime string `json:"auto_upgrade_start_time,omitempty" bson:"auto_upgrade_start_time,omitempty"`
	Description          string `json:"description,omitempty" bson:"description,omitempty"`
}

type Cluster struct {
	Name    string      `json:"name,omitempty" bson:"name,omitempty" description:"Cluster name"`
	InfraId string      `json:"infra_id" bson:"infra_id"  description:"ID of infrastructure"`
	Status  models.Type `json:"status,omitempty" bson:"status,omitempty" description:"Status of cluster"`
}

type KubeClusterStatus struct {
	Id                string                 `json:"id,omitempty"`
	Name              string                 `json:"name,omitempty"`
	Region            string                 `json:"region,omitempty"`
	Status            models.Type            `json:"status,omitempty"`
	State             string                 `json:"state,omitempty"`
	KubernetesVersion string                 `json:"kubernetes_version,omitempty"`
	Network           string                 `json:"network,omitempty"`
	PoolCount         int64                  `json:"nodepool_count,omitempty"`
	ClusterIP         string                 `json:"cluster_ip,omitempty"`
	WorkerPools       []KubeWorkerPoolStatus `json:"node_pools,omitempty"`
}

type KubeWorkerPoolStatus struct {
	Id          string            `json:"id,omitempty"`
	Name        string            `json:"name,omitempty"`
	Link        string            `json:"-"`
	NodeCount   int64             `json:"node_count,omitempty"`
	MachineType string            `json:"machine_type,omitempty"`
	Autoscaling AutoScaling       `json:"auto_scaling,omitempty"`
	Subnet      string            `json:"subnet_id,omitempty"`
	Nodes       []KubeNodesStatus `json:"nodes"`
}
type AutoScaling struct {
	AutoScale bool  `json:"auto_scale,omitempty"`
	MinCount  int64 `json:"min_scaling_group_size,omitempty"`
	MaxCount  int64 `json:"max_scaling_group_size,omitempty"`
}
type KubeNodesStatus struct {
	Id        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	State     string `json:"state,omitempty"`
	PrivateIp string `json:"private_ip,omitempty"`
	PublicIp  string `json:"public_ip,omitempty"`
}

func GetNetwork(token, infraId string, ctx utils.Context) error {

	url := getNetworkHost("gcp", infraId)

	_, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}

func GetGKECluster(ctx utils.Context) (cluster GKECluster, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs(
			"GKEGetClusterModel:  Get - Got error while connecting to the database: "+err1.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGKEClusterCollection)
	err = c.Find(bson.M{"infra_id": ctx.Data.InfraId, "company_id": ctx.Data.Company}).One(&cluster)
	if err != nil {
		ctx.SendLogs(
			"GKEGetClusterModel:  Get - Got error while fetching from database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err
	}

	return cluster, nil
}
func GetAllGKECluster(data rbacAuthentication.List, ctx utils.Context) (gkeClusters []Cluster, err error) {
	var clusters []GKECluster
	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs(
			"GKEGetAllClusterModel:  GetAll - Got error while connecting to the database: "+err1.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return gkeClusters, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGKEClusterCollection)
	err = c.Find(bson.M{"infra_id": bson.M{"$in": copyData}, "company_id": ctx.Data.Company}).All(&clusters)
	if err != nil {
		ctx.SendLogs(
			"GKEGetAllClusterModel:  GetAll - Got error while fetching from database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return gkeClusters, err
	}

	for _, cluster := range clusters {
		temp := Cluster{Name: cluster.Name, InfraId: cluster.InfraId, Status: cluster.CloudplexStatus}
		gkeClusters = append(gkeClusters, temp)
	}
	return gkeClusters, nil
}
func AddGKECluster(cluster GKECluster, ctx utils.Context) error {
	_, err := GetGKECluster(ctx)
	if err == nil {
		text := fmt.Sprintf("GKEAddClusterModel:  Add - Cluster for infrastructure '%s' already exists in the database."+err.Error(), cluster.InfraId)
		ctx.SendLogs(text+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"GKEAddClusterModel:  Add - Got error while connecting to the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}
	defer session.Close()

	if cluster.CreationDate.IsZero() {
		cluster.CreationDate = time.Now()
		cluster.ModificationDate = time.Now()
		if cluster.CloudplexStatus == "" {
			cluster.CloudplexStatus = "new"
		}
		cluster.Cloud = models.GKE
		cluster.CompanyId = ctx.Data.Company
	}

	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoGKEClusterCollection, cluster)
	if err != nil {
		ctx.SendLogs(
			"GKEAddClusterModel:  Add - Got error while inserting cluster to the database:  "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}
func UpdateGKECluster(cluster GKECluster, ctx utils.Context) error {
	oldCluster, err := GetGKECluster(ctx)
	if err != nil {
		text := "GKEUpdateClusterModel:  Update - Cluster '" + cluster.Name + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	err = DeleteGKECluster(ctx)
	if err != nil {
		ctx.SendLogs(
			"GKEUpdateClusterModel:  Update - Got error deleting old cluster ",
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	cluster.CreationDate = oldCluster.CreationDate
	cluster.ModificationDate = time.Now()
	cluster.CompanyId = oldCluster.CompanyId

	err = AddGKECluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs(
			"GKEUpdateClusterModel:  Update - Got error creating new cluster ",
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}
func DeleteGKECluster(ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"GKEDeleteClusterModel:  Delete - Got error while connecting to the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGKEClusterCollection)
	err = c.Remove(bson.M{"infra_id": ctx.Data.InfraId, "company_id": ctx.Data.Company})
	if err != nil {
		ctx.SendLogs(
			"GKEDeleteClusterModel:  Delete - Got error while deleting from the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func AddPreviousGKECluster(cluster GKECluster, ctx utils.Context, patch bool) error {
	var oldCluster GKECluster
	_, err := GetPreviousGKECluster(ctx)
	if err == nil {
		err := DeletePreviousGKECluster(ctx)
		if err != nil {
			ctx.SendLogs(
				"GKEAddClusterModel:  Add previous cluster - "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
		}
	}

	if patch == false {
		oldCluster, err = GetGKECluster(ctx)
		if err != nil {
			ctx.SendLogs(
				"GKEAddClusterModel:  Add previous cluster - "+err.Error(),
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
			"GKEAddClusterModel:  Add previous cluster - "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	defer session.Close()

	if cluster.CreationDate.IsZero() {
		cluster.CreationDate = time.Now()
		cluster.ModificationDate = time.Now()
		cluster.Cloud = models.GKE
		cluster.CompanyId = ctx.Data.Company
	}

	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoGKEPreviousClusterCollection, oldCluster)
	if err != nil {
		ctx.SendLogs(
			"GKEAddClusterModel:  Add previous cluster -  "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}
func GetPreviousGKECluster(ctx utils.Context) (cluster GKECluster, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs(
			"GKEGetClusterModel:  Get previous cluster - Got error while connecting to the database: "+err1.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGKEPreviousClusterCollection)
	err = c.Find(bson.M{"infra_id": ctx.Data.InfraId, "company_id": ctx.Data.Company}).One(&cluster)
	if err != nil {
		ctx.SendLogs(
			"GKEGetClusterModel:  Get previous cluster- Got error while fetching from database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err
	}

	return cluster, nil
}
func UpdatePreviousGKECluster(cluster GKECluster, ctx utils.Context) error {

	err := AddPreviousGKECluster(cluster, ctx, false)
	if err != nil {
		text := "GKEClusterModel:  Update  previous cluster -'" + cluster.Name + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	err = UpdateGKECluster(cluster, ctx)
	if err != nil {
		text := "GKEClusterModel:  Update previous cluster - '" + cluster.Name + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		err = DeletePreviousGKECluster(ctx)
		if err != nil {
			text := "GKEDeleteClusterModel:  Delete  previous cluster - '" + cluster.Name + err.Error()
			ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return errors.New(text)
		}
		return err
	}

	return nil
}
func DeletePreviousGKECluster(ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"GKEDeleteClusterModel:  Delete  previous cluster - "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGKEPreviousClusterCollection)
	err = c.Remove(bson.M{"infra_id": ctx.Data.InfraId, "company_id": ctx.Data.Company})
	if err != nil {
		ctx.SendLogs(
			"GKEDeleteClusterModel:  Delete  previous cluster - "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func DeployGKECluster(cluster GKECluster, credentials gcp.GcpCredentials, token string, ctx utils.Context) (confError types.CustomCPError) {

	publisher := utils.Notifier{}
	errr := publisher.Init_notifier()

	if errr != nil {
		PrintError(errr, cluster.Name, ctx)
		ctx.SendLogs(errr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := types.CustomCPError{StatusCode: 500, Error: "Error in deploying GKE Cluster", Description: errr.Error()}
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.GKE, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return cpErr
	}

	gkeOps, err := GetGKE(credentials)
	if err != (types.CustomCPError{}) {
		ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.GKE, ctx, err)
		if err_ != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return err
	}

	err = gkeOps.init()
	if err != (types.CustomCPError{}) {
		ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.CloudplexStatus = models.ClusterCreationFailed
		confError := UpdateGKECluster(cluster, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, ctx)
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.GKE, ctx, err)
		if err_ != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error + "\n" + err.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return err
	}

	_, _ = utils.SendLog(ctx.Data.Company, "Creating Cluster : "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
	cluster.CloudplexStatus = (models.Deploying)
	err_ := UpdateGKECluster(cluster, ctx)
	if err_ != nil {
		utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.InfraId)
		cpErr := types.CustomCPError{Description: confError.Error, Error: "Error occurred while updating cluster status in database", StatusCode: 500}
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.GKE, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err_.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return cpErr
	}

	err = gkeOps.CreateCluster(cluster, token, ctx)
	if err != (types.CustomCPError{}) {

		cluster.CloudplexStatus = models.ClusterCreationFailed

		utils.SendLog(ctx.Data.Company, "Cluster creation failed : "+cluster.Name, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
		utils.SendLog(ctx.Data.Company, err.Description, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)

		confError := UpdateGKECluster(cluster, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, ctx)
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.GKE, ctx, err)
		if err_ != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error + "\n" + err.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return err
	}

	pubSub := publisher.Subscribe(ctx.Data.InfraId, ctx)

	confError = ApplyAgent(credentials, token, ctx, cluster.Name)
	if confError != (types.CustomCPError{}) {
		cluster.CloudplexStatus = models.ClusterCreationFailed
		PrintError(errors.New(confError.Error), cluster.Name, ctx)
		PrintError(errors.New("Cleaning up resources"), cluster.Name, ctx)
		_ = TerminateCluster(credentials, token, ctx)

		_ = UpdateGKECluster(cluster, ctx)
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.GKE, ctx, confError)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: confError.Error + "\n" + confError.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		return confError
	}

	cluster.CloudplexStatus = models.ClusterCreated

	err1 := UpdateGKECluster(cluster, ctx)
	if err1 != nil {
		PrintError(err1, cluster.Name, ctx)
		cpErr := types.CustomCPError{StatusCode: 500, Error: "Error in applying agent", Description: err1.Error()}

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err1.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.GKE, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return cpErr
	}

	_, _ = utils.SendLog(ctx.Data.Company, "Cluster created successfully "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
	notify := publisher.RecieveNotification(ctx.Data.InfraId, ctx, pubSub)
	if notify {
		ctx.SendLogs("GKEClusterModel:  Notification recieved from agent", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Cluster created successfully",
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Create,
		}, ctx)
	} else {
		ctx.SendLogs("GKEClusterModel:  Notification not recieved from agent", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		cluster.Status = models.ClusterCreationFailed
		utils.SendLog(ctx.Data.Company, "Notification not recieved from agent", models.LOGGING_LEVEL_INFO, cluster.InfraId)
		confError_ := UpdateGKECluster(cluster, ctx)
		if confError_ != nil {
			ctx.SendLogs("GKEDeployClusterModel:"+confError_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		}
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.GKE, ctx, types.CustomCPError{Description: confError_.Error(), Error: confError_.Error(), StatusCode: 512})
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Agent  - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

func PatchRunningGKECluster(cluster GKECluster, credentials gcp.GcpCredentials, token string, ctx utils.Context) (confError types.CustomCPError) {
	/*
		publisher := utils.Notifier{}

		errr := publisher.Init_notifier()
		if errr != nil {
			PrintError(errr, cluster.Name, ctx)
			ctx.SendLogs(errr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			cpErr := types.CustomCPError{StatusCode: int(models.CloudStatusCode), Error: "Error in deploying GKE Cluster", Description: errr.Error()}
			err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.GKE, ctx, cpErr)
			if err != nil {
				ctx.SendLogs("GKEUpdateRunningClusterModel:  Update - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			return cpErr
		}*/

	gkeOps, err := GetGKE(credentials)
	if err != (types.CustomCPError{}) {
		ctx.SendLogs("GKEUpdateRunningClusterModel: Update - "+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.GKE, ctx, err)
		if err_ != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return err
	}

	err = gkeOps.init()
	if err != (types.CustomCPError{}) {
		ctx.SendLogs("GKEUpdateRunningClusterModel:  Update - "+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.CloudplexStatus = models.ClusterCreationFailed
		confError := UpdateGKECluster(cluster, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, ctx)
			ctx.SendLogs("GKEUpdateRunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.GKE, ctx, err)
		if err_ != nil {
			ctx.SendLogs("GKEUpdateRunningClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	difCluster, err1 := CompareClusters(ctx)
	if err1 != nil {
		if strings.Contains(err1.Error(), "Nothing to update") {
			cluster.CloudplexStatus = models.ClusterCreated
			confError_ := UpdateGKECluster(cluster, ctx)
			if confError_ != nil {
				ctx.SendLogs("GKERunningClusterModel:"+confError_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			utils.Publisher(utils.ResponseSchema{
				Status:  true,
				Message: "Nothing to update",
				InfraId: cluster.InfraId,
				Token:   token,
				Action:  models.Update,
			}, ctx)
			return types.CustomCPError{}
		}
	}

	_, _ = utils.SendLog(ctx.Data.Company, "Updating running cluster : "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)

	previousCluster, err1 := GetPreviousGKECluster(ctx)
	if err1 != nil {
		return types.CustomCPError{Error: "Error in updating running cluster", StatusCode: 512, Description: err1.Error()}
	}
	previousPoolCount := len(previousCluster.NodePools)

	/*	if previousPoolCount < newPoolCount {
			var pools []*NodePool
			for i := previousPoolCount; i < newPoolCount; i++ {
				pools = append(pools, cluster.NodePools[i])
			}

			err := AddNodepool(cluster, ctx, gkeOps, pools, previousPoolCount)
			if err != (types.CustomCPError{}) {
				return err
			}
		} else if previousPoolCount > newPoolCount {

			previousCluster, err := GetPreviousGKECluster(ctx)
			if err != nil {
				return types.CustomCPError{Error: "Error in updating running cluster", StatusCode: 512, Description: err.Error()}
			}

			for _, oldpool := range previousCluster.NodePools {
				delete := true
				for _, pool := range cluster.NodePools {

					if pool.Name == oldpool.Name {
						delete = false
					}
				}
				if delete == true {
					DeleteNodepool(cluster, ctx, gkeOps, oldpool.Name)
				}
			}
		}
	*/

	var addpools []*NodePool
	var addedIndex []int
	addincluster := false
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
		err = AddNodepool(cluster, ctx, gkeOps, addpools,token)
		if err != (types.CustomCPError{}) {
			return err
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
			DeleteNodepool(cluster, ctx, gkeOps, prePool.Name,token)
		}

	}

	for _, dif := range difCluster {

		if len(dif.Path) > 2 {
			poolIndex, _ := strconv.Atoi(dif.Path[1])
			if poolIndex > (previousPoolCount - 1) {
				continue
			}
			for _, index := range addedIndex {
				if index == poolIndex {
					continue
				}
			}
		}
		if dif.Type == "update" {
			if dif.Path[0] == "MasterAuthorizedNetworksConfig" {
				err := UpdateMasterAuthorizedNetworksConfig(cluster, ctx, gkeOps,token )
				if err != (types.CustomCPError{}) {
					return err
				}
			} else if dif.Path[0] == "NetworkPolicy" {
				err := UpdateNetworkPolicy(cluster, ctx, gkeOps,token)
				if err != (types.CustomCPError{}) {
					return err
				}
			} else if dif.Path[0] == "AddonsConfig" {
				err := UpdateHttpLoadBalancing(cluster, ctx, gkeOps,token)
				if err != (types.CustomCPError{}) {
					return err
				}
			} else if dif.Path[0] == "InitialClusterVersion" {
				err := UpdateVersion(cluster, ctx, gkeOps,token)
				if err != (types.CustomCPError{}) {
					return err
				}
			} else if dif.Path[0] == "LoggingService" || dif.Path[0] == "MonitoringService" {
				err := UpdateMonitoringAndLoggingService(cluster, ctx, gkeOps,token)
				if err != (types.CustomCPError{}) {
					return err
				}
			} else if dif.Path[0] == "LegacyAbac" {
				err := UpdateLegacyAbac(cluster, ctx, gkeOps,token)
				if err != (types.CustomCPError{}) {
					return err
				}
			} else if dif.Path[0] == "MaintenancePolicy" {
				err := UpdateMaintenancePolicy(cluster, ctx, gkeOps,token)
				if err != (types.CustomCPError{}) {
					return err
				}
			} else if dif.Path[0] == "ResourceUsageExportConfig" {
				err := UpdateResourceUsageExportConfig(cluster, ctx, gkeOps,token)
				if err != (types.CustomCPError{}) {
					return err
				}
			} else if len(dif.Path) >= 4 && dif.Path[0] == "NodePools" && dif.Path[2] == "Config" && dif.Path[3] == "ImageType" {
				poolIndex, _ := strconv.Atoi(dif.Path[1])
				err := UpdateNodeImage(cluster, ctx, gkeOps, cluster.NodePools[poolIndex], poolIndex,token)
				if err != (types.CustomCPError{}) {
					return err
				}
			} else if len(dif.Path) >= 3 && dif.Path[0] == "NodePools" && dif.Path[2] == "InitialNodeCount" {
				poolIndex, _ := strconv.Atoi(dif.Path[1])
				err := UpdateNodeCount(cluster, ctx, gkeOps, cluster.NodePools[poolIndex], poolIndex,token)
				if err != (types.CustomCPError{}) {
					return err
				}
			} else if len(dif.Path) >= 3 && dif.Path[0] == "NodePools" && dif.Path[2] == "Management" {
				poolIndex, _ := strconv.Atoi(dif.Path[1])
				err := UpdateNodePoolManagement(cluster, ctx, gkeOps, cluster.NodePools[poolIndex], poolIndex,token)
				if err != (types.CustomCPError{}) {
					return err
				}
			} else if len(dif.Path) > 3 && dif.Path[0] == "NodePools" && dif.Path[2] == "Autoscaling" {
				poolIndex, _ := strconv.Atoi(dif.Path[1])
				err := UpdateNodePoolScaling(cluster, ctx, gkeOps, cluster.NodePools[poolIndex], poolIndex,token)
				if err != (types.CustomCPError{}) {
					return err
				}
			}
		}
	}

	DeletePreviousGKECluster(ctx)

	latestCluster, err2 := gkeOps.fetchClusterStatus(cluster.Name, ctx)
	if err2 != (types.CustomCPError{}) {
		return err
	}

	for strings.ToLower(string(latestCluster.Status)) != strings.ToLower("running") {
		time.Sleep(time.Second * 60)
	}

	cluster.CloudplexStatus = models.ClusterCreated
	confError_ := UpdateGKECluster(cluster, ctx)
	if confError_ != nil {
		ctx.SendLogs("GKERunningClusterModel:"+confError_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

	}

	_, _ = utils.SendLog(ctx.Data.Company, "Running Cluster updated successfully "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
	utils.Publisher(utils.ResponseSchema{
		Status:  true,
		Message: "cluster updated successfully",
		InfraId: cluster.InfraId,
		Token:   token,
		Action:  models.Update,
	}, ctx)

	return types.CustomCPError{}

}

func FetchStatus(credentials gcp.GcpCredentials, token string, ctx utils.Context) (*KubeClusterStatus, types.CustomCPError) {
	cluster, err := GetGKECluster(ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel:  Fetch -  Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return &KubeClusterStatus{}, types.CustomCPError{StatusCode: 500, Error: "Error in fetching status", Description: err.Error()}
	}
	if string(cluster.CloudplexStatus) == strings.ToLower(string(models.New)) {
		cpErr := types.CustomCPError{Error: "Unable to fetch status - Cluster is not deployed yet", Description: "Unable to fetch state - Cluster is not deployed yet", StatusCode: 409}
		return &KubeClusterStatus{}, cpErr
	}
	if cluster.CloudplexStatus == models.Deploying || cluster.CloudplexStatus == models.Terminating || cluster.CloudplexStatus == models.ClusterTerminated {
		cpErr := types.CustomCPError{Error: "Cluster is in " +
			string(cluster.CloudplexStatus) + " state", Description: "Cluster is in " +
			string(cluster.CloudplexStatus) + " state", StatusCode: 409}
		return &KubeClusterStatus{}, cpErr
	}
	if cluster.CloudplexStatus != models.ClusterCreated {
		customErr, err := db.GetError(cluster.InfraId, ctx.Data.Company, models.GKE, ctx)
		if err != nil {
			return &KubeClusterStatus{}, types.CustomCPError{Error: "Error occurred while getting cluster status in database",
				Description: "Error occurred while getting cluster status in database",
				StatusCode:  500}
		}
		if customErr.Err != (types.CustomCPError{}) {
			return &KubeClusterStatus{}, customErr.Err
		}
	}
	gkeOps, err1 := GetGKE(credentials)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GKEClusterModel:  Fetch -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return &KubeClusterStatus{}, err1
	}

	err1 = gkeOps.init()
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GKEClusterModel:  Fetch -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return &KubeClusterStatus{}, err1
	}

	latestCluster, err1 := gkeOps.fetchClusterStatus(cluster.Name, ctx)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GKEClusterModel:  Fetch - Failed to get latest status "+err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return &KubeClusterStatus{}, err1
	}
	customStatus := fillStatusInfo(latestCluster)
	err1 = gkeOps.fetchNodePool(latestCluster, &customStatus, ctx)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GKEClusterModel:  Fetch - Failed to get latest status "+err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return &KubeClusterStatus{}, err1
	}
	customStatus.Status = cluster.Status
	latestCluster.InfraId = ctx.Data.InfraId
	latestCluster.CompanyId = ctx.Data.Company
	latestCluster.CloudplexStatus = cluster.CloudplexStatus
	latestCluster.IsExpert = cluster.IsExpert
	latestCluster.IsAdvance = cluster.IsAdvance

	return &customStatus, types.CustomCPError{}
}

func TerminateCluster(credentials gcp.GcpCredentials, token string, ctx utils.Context) types.CustomCPError {
	/*publisher := utils.Notifier{}
	pubErr := publisher.Init_notifier()
	if pubErr != nil {
		err_ := types.CustomCPError{StatusCode: 500, Error: "Error in terminating cluster", Description: pubErr.Error()}
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.GKE, ctx, err_)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return err_
	}*/

	cluster, err := GetGKECluster(ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Terminate ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		err_ := types.CustomCPError{StatusCode: 500, Error: "Error in terminating cluster", Description: err.Error()}
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.GKE, ctx, err_)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return err_
	}

	if cluster.CloudplexStatus == "" || cluster.CloudplexStatus == "new" {
		text := "GKEClusterModel : Terminate - Cannot terminate a new cluster"
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		err_ := types.CustomCPError{StatusCode: 500, Error: "Error in terminating error", Description: text}
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.GKE, ctx, err_)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: text,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Terminate,
		}, ctx)
		return err_
	}

	gkeOps, err1 := GetGKE(credentials)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GKEClusterModel : Terminate - "+err1.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.GKE, ctx, err1)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return err1
	}

	err1 = gkeOps.init()
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GKEClusterModel : Terminate -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.CloudplexStatus = models.ClusterTerminationFailed
		err = UpdateGKECluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("GKEClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			_, _ = utils.SendLog(ctx.Data.Company, "Error in cluster updation in mongo: "+cluster.Name, models.LOGGING_LEVEL_ERROR, cluster.InfraId)
			_, _ = utils.SendLog(ctx.Data.Company, err.Error(), "error", ctx.Data.InfraId)
			err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.GKE, ctx, err1)
			if err != nil {
				ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			return err1
		}
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.GKE, ctx, err1)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err1.Error + "\n" + err1.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Terminate,
		}, ctx)
		return err1
	}

	_, err_ := CompareClusters(ctx)
	if err_ != nil  &&  !(strings.Contains(err_.Error(),"Nothing to update")){
		oldCluster,err_ := GetPreviousGKECluster(ctx)
		if err_ != nil {

			utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.InfraId)
			cpErr := types.CustomCPError{Description: err_.Error(), Error: "Error occurred while updating cluster status in database", StatusCode: 500}
			err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.GKE, ctx, cpErr)
			if err != nil {
				ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

		err_ = UpdateGKECluster(oldCluster, ctx)
		if err_ != nil {

			utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.InfraId)
			cpErr := types.CustomCPError{Description: err_.Error(), Error: "Error occurred while updating cluster status in database", StatusCode: 500}
			err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.GKE, ctx, cpErr)
			if err != nil {
				ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
	}

	_, _ = utils.SendLog(ctx.Data.Company, "Terminating Cluster : "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
	cluster.CloudplexStatus = models.Terminating
	err_ = UpdateGKECluster(cluster, ctx)
	if err_ != nil {

		utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.InfraId)
		cpErr := types.CustomCPError{Description: err_.Error(), Error: "Error occurred while updating cluster status in database", StatusCode: 500}
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.GKE, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
	errr := gkeOps.deleteCluster(cluster, ctx)
	if errr != (types.CustomCPError{}) {
		_, _ = utils.SendLog(ctx.Data.Company, "Cluster termination failed: "+cluster.Name, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
		utils.SendLog(ctx.Data.Company, errr.Description, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
		cluster.CloudplexStatus = models.ClusterTerminationFailed
		err = UpdateGKECluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("GKEClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			_, _ = utils.SendLog(ctx.Data.Company, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.InfraId)
			_, _ = utils.SendLog(ctx.Data.Company, err.Error(), "error", ctx.Data.InfraId)
			err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.GKE, ctx, errr)
			if err != nil {
				ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			utils.Publisher(utils.ResponseSchema{
				Status:  false,
				Message: errr.Error + "\n" + errr.Description,
				InfraId: cluster.InfraId,
				Token:   token,
				Action:  models.Terminate,
			}, ctx)
			return errr
		}
		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.GKE, ctx, errr)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: errr.Error + "\n" + errr.Description,
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Terminate,
		}, ctx)
		return errr
	}

	cluster.CloudplexStatus = models.ClusterTerminated

	err = UpdateGKECluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Terminate "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		_, _ = utils.SendLog(ctx.Data.Company, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.InfraId)
		_, _ = utils.SendLog(ctx.Data.Company, err.Error(), "error", ctx.Data.InfraId)
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: cluster.InfraId,
			Token:   token,
			Action:  models.Terminate,
		}, ctx)
		err_ := types.CustomCPError{StatusCode: 500, Error: "Error in terminating Cluster", Description: err.Error()}

		err := db.CreateError(ctx.Data.InfraId, ctx.Data.Company, models.GKE, ctx, err_)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return err_
	}
	_, _ = utils.SendLog(ctx.Data.Company, "Cluster terminated successfully "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.InfraId)
	utils.Publisher(utils.ResponseSchema{
		Status:  true,
		Message: "Cluster terminated successfully",
		InfraId: cluster.InfraId,
		Token:   token,
		Action:  models.Terminate,
	}, ctx)
	return types.CustomCPError{}
}

func GetServerConfig(credentials gcp.GcpCredentials, ctx utils.Context) (*gke.ServerConfig, types.CustomCPError) {
	gkeOps, err := GetGKE(credentials)
	if err != (types.CustomCPError{}) {
		ctx.SendLogs("GKEClusterModel : GetServerConfig - "+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	err = gkeOps.init()
	if err != (types.CustomCPError{}) {
		ctx.SendLogs("GKEClusterModel : GetServerConfig -"+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	return gkeOps.getGKEVersions(ctx)
}

func PrintError(confError error, name string, ctx utils.Context) {
	if confError != nil {
		_, _ = utils.SendLog(ctx.Data.Company, "Cluster creation failed : "+name, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
		_, _ = utils.SendLog(ctx.Data.Company, confError.Error(), models.LOGGING_LEVEL_ERROR, ctx.Data.Company)
	}
}

func ApplyAgent(credentials gcp.GcpCredentials, token string, ctx utils.Context, clusterName string) (confError types.CustomCPError) {

	data2, err := woodpecker.GetCertificate(ctx.Data.InfraId, token, ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Apply Agent -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.CustomCPError{StatusCode: 500, Error: "Error in applying agent", Description: err.Error()}
	}
	filePath := "/tmp/" + ctx.Data.Company + "/" + ctx.Data.InfraId + "/"
	cmd := "mkdir -p " + filePath + " && echo '" + data2 + "'>" + filePath + "agent.yaml && echo '" + credentials.RawData + "'>" + filePath + "gcp-auth.json"
	output, err := models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.CustomCPError{StatusCode: 500, Error: "Error in applying agent", Description: err.Error()}
	}

	if credentials.Zone != "" {
		cmd = "sudo docker run --rm --name " + ctx.Data.Company + ctx.Data.InfraId + " -e gcpProject=" + credentials.AccountData.InfraId + " -e cluster=" + clusterName + " -e zone=" + credentials.Region + "-" + credentials.Zone + " -e serviceAccount=" + filePath + "gcp-auth.json" + " -e yamlFile=" + filePath + "agent.yaml -v " + filePath + ":" + filePath + " " + models.GKEAuthContainerName
	} else {
		cmd = "sudo docker run --rm --name " + ctx.Data.Company + ctx.Data.InfraId + " -e gcpProject=" + credentials.AccountData.InfraId + " -e cluster=" + clusterName + " -e region=" + credentials.Region + " -e serviceAccount=" + filePath + "gcp-auth.json" + " -e yamlFile=" + filePath + "agent.yaml -v " + filePath + ":" + filePath + " " + models.GKEAuthContainerName
	}

	output, err = models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.CustomCPError{StatusCode: 500, Error: "Error in applying agent", Description: err.Error()}
	}
	return types.CustomCPError{}
}

func Validate(gkeCluster GKECluster) error {
	if gkeCluster.InfraId == "" {
		return errors.New("infrastructure id is empty")
	} else if gkeCluster.Name == "" {
		return errors.New("cluster name is empty")
	} else if len(gkeCluster.NodePools) > 0 {
		for _, nodepool := range gkeCluster.NodePools {
			if nodepool.Config != nil {
				isDiskExist, err := validateGKEDiskType(nodepool.Config.DiskType)
				if err != nil && !isDiskExist {
					text := "availabe disks are " + err.Error()
					return errors.New(text)
				}

				isImageExist, err := validateGKEImageType(nodepool.Config.ImageType)
				if err != nil && !isImageExist {
					text := "availabe images are " + err.Error()
					return errors.New(text)
				}

			}
		}
	}
	return nil
}

func validateGKEDiskType(diskType string) (bool, error) {
	diskList := []string{"pd-ssd", "pd-standard"}

	for _, v := range diskList {
		if v == diskType {
			return true, nil
		}
	}

	var errData string
	for _, v := range diskList {
		errData += v + ", "
	}

	return false, errors.New(errData)
}

func validateGKEImageType(imageType string) (bool, error) {
	imageList := []string{"COS_CONTAINERD", "COS", "UBUNTU"}

	for _, v := range imageList {
		if v == imageType {
			return true, nil
		}
	}

	var errData string
	for _, v := range imageList {
		errData += v + ", "
	}

	return false, errors.New(errData)
}

func fillStatusInfo(cluster GKECluster) (status KubeClusterStatus) {
	status.Id = cluster.Name
	status.Name = cluster.Name
	status.Region = cluster.Location
	status.Status = cluster.CloudplexStatus
	status.Network = cluster.Network
	status.PoolCount = cluster.CurrentNodeCount
	status.KubernetesVersion = cluster.CurrentMasterVersion
	status.ClusterIP = cluster.ClusterIpv4Cidr
	status.PoolCount = cluster.CurrentNodeCount
	status.ClusterIP = cluster.ClusterIpv4Cidr
	status.State = cluster.NodePools[0].Status
	for _, pool := range cluster.NodePools {
		workerpool := KubeWorkerPoolStatus{}
		workerpool.Name = pool.Name
		workerpool.Id = pool.Name
		workerpool.NodeCount = pool.InitialNodeCount
		workerpool.MachineType = pool.Config.MachineType
		workerpool.Link = pool.InstanceGroupUrls[0]
		if pool.Autoscaling != nil && pool.Autoscaling.Enabled == true {
			workerpool.Autoscaling.AutoScale = pool.Autoscaling.Enabled
			workerpool.Autoscaling.MinCount = pool.Autoscaling.MinNodeCount
			workerpool.Autoscaling.MaxCount = pool.Autoscaling.MaxNodeCount

			workerpool.Subnet = cluster.Subnetwork
		} else {
			workerpool.Autoscaling.AutoScale = false
		}
		status.WorkerPools = append(status.WorkerPools, workerpool)

	}

	return status
}

func UpdateVersion(cluster GKECluster, ctx utils.Context, gkeOps GKE,token string) types.CustomCPError {

	err := gkeOps.UpdateMasterVersion(cluster, ctx)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}
	for _, node := range cluster.NodePools {
		err := gkeOps.UpdateNodePoolVersion(cluster.Name, node.Name, cluster.InitialClusterVersion, ctx)
		if err != (types.CustomCPError{}) {
			updationFailedError(cluster, ctx, err,token)
			return err
		}
	}

	oldCluster, err1 := GetPreviousGKECluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{
			StatusCode:  int(models.CloudStatusCode),
			Error:       "Error in updating running cluster version",
			Description: err1.Error(),
		},token)
	}

	if cluster.InitialClusterVersion != "" {
		oldCluster.InitialClusterVersion = cluster.InitialClusterVersion
	}

	err1 = AddPreviousGKECluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in updating running cluster version", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}
	return types.CustomCPError{}
}

func CompareClusters(ctx utils.Context) (diff.Changelog, error) {
	cluster, err := GetGKECluster(ctx)
	if err != nil {
		return diff.Changelog{}, err
	}

	oldCluster, err := GetPreviousGKECluster(ctx)
	if err != nil && strings.Contains(err.Error(), "not found") {
		return diff.Changelog{}, errors.New("Nothing to update")
	}

	previousPoolCount := len(oldCluster.NodePools)
	newPoolCount := len(cluster.NodePools)

	difCluster, err := diff.Diff(oldCluster, cluster)
	if len(difCluster) < 2 && previousPoolCount == newPoolCount {
		return diff.Changelog{}, errors.New("Nothing to update")
	} else if err != nil {
		return diff.Changelog{}, errors.New("Error in comparing differences:" + err.Error())
	}
	return difCluster, nil
}

func UpdateMasterAuthorizedNetworksConfig(cluster GKECluster, ctx utils.Context, gkeOps GKE,token string) types.CustomCPError {
	err := gkeOps.UpdateMasterAuthorizedNetworksConfig(cluster, ctx)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}

	oldCluster, err1 := GetPreviousGKECluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, err,token)
	}

	if cluster.MasterAuthorizedNetworksConfig != nil {
		oldCluster.MasterAuthorizedNetworksConfig = cluster.MasterAuthorizedNetworksConfig
	} else {
		oldCluster.MasterAuthorizedNetworksConfig = nil
	}

	err1 = AddPreviousGKECluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in updating running cluster master authorized networks config", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}
	return types.CustomCPError{}
}

func UpdateMonitoringAndLoggingService(cluster GKECluster, ctx utils.Context, gkeOps GKE,token string) types.CustomCPError {

	err := gkeOps.UpdateMonitoringAndLoggingService(cluster, ctx)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}

	oldCluster, err1 := GetPreviousGKECluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{
			StatusCode:  int(models.CloudStatusCode),
			Error:       "Error in running cluster monitoring and logging update",
			Description: err1.Error(),
		},token)
	}

	if cluster.MonitoringService != "" {
		oldCluster.MonitoringService = cluster.MonitoringService
		oldCluster.LoggingService = cluster.LoggingService
	} else {
		oldCluster.MonitoringService = ""
		oldCluster.LoggingService = ""
	}

	err1 = AddPreviousGKECluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in running cluster monitoring and logging update", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}
	return types.CustomCPError{}
}

func UpdateLegacyAbac(cluster GKECluster, ctx utils.Context, gkeOps GKE,token string) types.CustomCPError {

	err := gkeOps.UpdateLegacyAbac(cluster, ctx)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}

	oldCluster, err1 := GetPreviousGKECluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{
			StatusCode:  int(models.CloudStatusCode),
			Error:       "Error in running cluster legacyAbac update",
			Description: err1.Error(),
		},token)
	}

	if cluster.LegacyAbac != nil {
		oldCluster.LegacyAbac = cluster.LegacyAbac
	} else {
		oldCluster.LegacyAbac = nil
	}

	err1 = AddPreviousGKECluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in running cluster legacyAbac update", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}

	return types.CustomCPError{}
}

func UpdateMaintenancePolicy(cluster GKECluster, ctx utils.Context, gkeOps GKE,token string) types.CustomCPError {

	err := gkeOps.UpdateMaintenancePolicy(cluster, ctx)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}

	oldCluster, err1 := GetPreviousGKECluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{
			StatusCode:  int(models.CloudStatusCode),
			Error:       "Error in running cluster maintenance policy update",
			Description: err1.Error(),
		},token)
	}

	if cluster.MaintenancePolicy != nil {
		oldCluster.MaintenancePolicy = cluster.MaintenancePolicy
	} else {
		oldCluster.MaintenancePolicy = nil
	}

	err1 = AddPreviousGKECluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in running cluster maintenance policy update", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}
	return types.CustomCPError{}
}

func UpdateResourceUsageExportConfig(cluster GKECluster, ctx utils.Context, gkeOps GKE,token string) types.CustomCPError {
	if cluster.ResourceUsageExportConfig == nil {
		cluster.ResourceUsageExportConfig.BigqueryDestination = nil
		cluster.ResourceUsageExportConfig.ConsumptionMeteringConfig = &ConsumptionMeteringConfig{}
		cluster.ResourceUsageExportConfig.ConsumptionMeteringConfig.Enabled = false
		cluster.ResourceUsageExportConfig.EnableNetworkEgressMetering = false
	}
	err := gkeOps.UpdateResourceUsageExportConfig(cluster, ctx)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}

	oldCluster, err1 := GetPreviousGKECluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{
			StatusCode:  int(models.CloudStatusCode),
			Error:       "Error in running cluster resource usage export config update",
			Description: err1.Error(),
		},token)
	}

	if cluster.MaintenancePolicy != nil {
		oldCluster.MaintenancePolicy = cluster.MaintenancePolicy
	} else {
		oldCluster.MaintenancePolicy = nil
	}

	err1 = AddPreviousGKECluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in running cluster resource usage export config update", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}
	return types.CustomCPError{}
}

func UpdateNetworkPolicy(cluster GKECluster, ctx utils.Context, gkeOps GKE,token string) types.CustomCPError {

	err := gkeOps.UpdateNetworkPolicy(cluster, ctx)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}

	oldCluster, err1 := GetPreviousGKECluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{
			StatusCode:  int(models.CloudStatusCode),
			Error:       "Error in running cluster network policy update",
			Description: err1.Error(),
		},token)
	}

	if cluster.NetworkPolicy != nil {
		oldCluster.NetworkPolicy = cluster.NetworkPolicy
	} else {
		oldCluster.NetworkPolicy = nil
	}

	err1 = AddPreviousGKECluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in running cluster network policy update", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}
	return types.CustomCPError{}
}

func UpdateHttpLoadBalancing(cluster GKECluster, ctx utils.Context, gkeOps GKE,token string ) types.CustomCPError {

	err := gkeOps.UpdateHttpLoadBalancing(cluster, ctx)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}

	oldCluster, err1 := GetPreviousGKECluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{
			StatusCode:  int(models.CloudStatusCode),
			Error:       "Error in running cluster http load balancing update",
			Description: err1.Error(),
		},token)
	}

	if cluster.AddonsConfig != nil {
		oldCluster.AddonsConfig = cluster.AddonsConfig
	} else {
		oldCluster.AddonsConfig = nil
	}

	err1 = AddPreviousGKECluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in running cluster http load balancing update", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}
	return types.CustomCPError{}
}

func UpdateNodeCount(cluster GKECluster, ctx utils.Context, gkeOps GKE, pool *NodePool, poolIndex int,token string) types.CustomCPError {

	err := gkeOps.UpdateNodePoolCount(cluster.Name, *pool, ctx)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}

	oldCluster, err1 := GetPreviousGKECluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{
			StatusCode:  int(models.CloudStatusCode),
			Error:       "Error in running cluster nodepool node count update",
			Description: err1.Error(),
		},token)
	}

	if cluster.NodePools[poolIndex].InitialNodeCount != 0 {
		oldCluster.NodePools[poolIndex].InitialNodeCount = cluster.NodePools[poolIndex].InitialNodeCount
	}

	err1 = AddPreviousGKECluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in running cluster nodepool node count update", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}
	return types.CustomCPError{}
}

func UpdateNodeImage(cluster GKECluster, ctx utils.Context, gkeOps GKE, pool *NodePool, poolIndex int,token string) types.CustomCPError {

	err := gkeOps.UpdateNodePoolImageType(cluster.Name, *pool, ctx)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}

	oldCluster, err1 := GetPreviousGKECluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{
			StatusCode:  int(models.CloudStatusCode),
			Error:       "Error in running cluster nodepool node image update",
			Description: err1.Error(),
		},token)
	}

	if cluster.NodePools[poolIndex].Config.ImageType != "" {
		oldCluster.NodePools[poolIndex].Config.ImageType = cluster.NodePools[poolIndex].Config.ImageType
	}

	err1 = AddPreviousGKECluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in running cluster nodepool node image update", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}

	return types.CustomCPError{}
}

func UpdateNodePoolManagement(cluster GKECluster, ctx utils.Context, gkeOps GKE, pool *NodePool, poolIndex int,token string) types.CustomCPError {

	err := gkeOps.UpdateNodepoolManagement(cluster.Name, *pool, ctx)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}

	oldCluster, err1 := GetPreviousGKECluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{
			StatusCode:  int(models.CloudStatusCode),
			Error:       "Error in running cluster nodepool management update",
			Description: err1.Error(),
		},token)
	}

	if oldCluster.NodePools[poolIndex].Management != nil {
		oldCluster.NodePools[poolIndex].Management = cluster.NodePools[poolIndex].Management
	} else {
		oldCluster.NodePools[poolIndex].Management = nil
	}

	err1 = AddPreviousGKECluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in running cluster nodepool management update", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}
	return types.CustomCPError{}
}

func UpdateNodePoolScaling(cluster GKECluster, ctx utils.Context, gkeOps GKE, pool *NodePool, poolIndex int,token string) types.CustomCPError {

	err := gkeOps.UpdateNodepoolScaling(cluster.Name, *pool, ctx)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}

	oldCluster, err1 := GetPreviousGKECluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{
			StatusCode:  int(models.CloudStatusCode),
			Error:       "Error in running cluster nodepool scaling update",
			Description: err1.Error(),
		},token)
	}

	if cluster.NodePools[poolIndex].Config.ImageType != "" {
		oldCluster.NodePools[poolIndex].Config.ImageType = cluster.NodePools[poolIndex].Config.ImageType
	}

	err1 = AddPreviousGKECluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in running cluster nodepool scaling update", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}

	return types.CustomCPError{}
}

func AddNodepool(cluster GKECluster, ctx utils.Context, gkeOps GKE, pools []*NodePool,token string) types.CustomCPError {

	err := gkeOps.AddNodePool(cluster.Name, pools, ctx)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}

	oldCluster, err1 := GetPreviousGKECluster(ctx)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{
			StatusCode:  int(models.CloudStatusCode),
			Error:       "Error in adding nodepool in running cluster",
			Description: err1.Error(),
		},token)
	}

	for _, pool := range pools {
		pool.PoolStatus = true
		oldCluster.NodePools = append(oldCluster.NodePools, pool)
	}

	err1 = AddPreviousGKECluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in adding nodepool in running cluster", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}
	return types.CustomCPError{}
}

func DeleteNodepool(cluster GKECluster, ctx utils.Context, gkeOps GKE, poolName ,token string) types.CustomCPError {

	err := gkeOps.DeleteNodePool(cluster.Name, poolName, ctx)
	if err != (types.CustomCPError{}) {
		updationFailedError(cluster, ctx, err,token)
		return err
	}

	oldCluster, err1 := GetPreviousGKECluster(ctx)
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
	err1 = AddPreviousGKECluster(oldCluster, ctx, true)
	if err1 != nil {
		return updationFailedError(cluster, ctx, types.CustomCPError{Error: "Error in deleting nodepool in running cluster", Description: err1.Error(), StatusCode: int(models.CloudStatusCode)},token)
	}
	return types.CustomCPError{}
}

func updationFailedError(cluster GKECluster, ctx utils.Context, err types.CustomCPError,token string) types.CustomCPError {
	publisher := utils.Notifier{}

	errr := publisher.Init_notifier()
	if errr != nil {
		PrintError(errr, cluster.Name, ctx)
		ctx.SendLogs(errr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := types.CustomCPError{StatusCode: 500, Error: "Error in deploying GKE Cluster", Description: errr.Error()}
		err := db.CreateError(cluster.InfraId, ctx.Data.Company, models.GKE, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("GKERunningClusterModel: Update - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return cpErr
	}

	cluster.CloudplexStatus = models.ClusterUpdateFailed
	confError := UpdateGKECluster(cluster, ctx)
	if confError != nil {
		PrintError(confError, cluster.Name, ctx)
		ctx.SendLogs("GKERunningClusterModel:  Update - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	}

	utils.SendLog(ctx.Data.Company, "Error in running cluster update : "+err.Description, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)

	err_ := db.CreateError(cluster.InfraId, ctx.Data.Company, models.GKE, ctx, err)
	if err_ != nil {
		ctx.SendLogs("GKERunningClusterModel:  Update - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	}

	utils.SendLog(ctx.Data.Company, "Deployed cluster update failed : "+cluster.Name, models.LOGGING_LEVEL_ERROR, ctx.Data.InfraId)
	utils.SendLog(ctx.Data.Company, err.Description, models.LOGGING_LEVEL_ERROR, ctx.Data.Company)

	utils.Publisher(utils.ResponseSchema{
		Status:  false,
		Message: "Cluster update failed",
		InfraId: cluster.InfraId,
		Token:   token,
		Action:  models.Update,
	}, ctx)

	return err
}
