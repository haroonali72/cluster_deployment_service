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
	gke "google.golang.org/api/container/v1"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

type GKECluster struct {
	ID                             bson.ObjectId                   `json:"-" bson:"_id,omitempty"`
	ProjectId                      string                          `json:"project_id" bson:"project_id" validate:"required" description:"ID of project [required]"`
	Cloud                          models.Cloud                    `json:"cloud" bson:"cloud" validate:"eq=gcp|eq=GCP"`
	CreationDate                   time.Time                       `json:"-" bson:"creation_date"`
	ModificationDate               time.Time                       `json:"-" bson:"modification_date"`
	CloudplexStatus                models.Type                     `json:"status" bson:"status" validate:"eq=new" description:"Status of cluster [required]"`
	CompanyId                      string                          `json:"company_id" bson:"company_id" description:"ID of compnay [optional]"`
	IsExpert                       bool                            `json:"is_expert" bson:"is_expert"`
	IsAdvance                      bool                            `json:"is_advance" bson:"is_advance"`
	AddonsConfig                   *AddonsConfig                   `json:"addons_config,omitempty" bson:"addons_config,omitempty"`
	ClusterIpv4Cidr                string                          `json:"cluster_ipv4_cidr,omitempty" bson:"cluster_ipv4_cidr,omitempty" description:"The IP address range of the container pods in this cluster [optional]"`
	Conditions                     []*StatusCondition              `json:"conditions,omitempty" bson:"conditions,omitempty"`
	CreateTime                     string                          `json:"create_time,omitempty" bson:"create_time,omitempty" description:"The time the cluster was created [readonly]"`
	CurrentMasterVersion           string                          `json:"current_master_version,omitempty" bson:"current_master_version,omitempty" description:"The current software version of master endpoint [readonly]"`
	CurrentNodeCount               int64                           `json:"current_node_count,omitempty" bson:"current_node_count,omitempty" description:"The number of nodes currently in the cluster [readonly]"`
	DefaultMaxPodsConstraint       *MaxPodsConstraint              `json:"default_max_pods_constraint,omitempty" bson:"default_max_pods_constraint,omitempty" validate:"dive"`
	Description                    string                          `json:"description,omitempty" bson:"description,omitempty" description:"An optional description of this cluster [optional]"`
	EnableKubernetesAlpha          bool                            `json:"enable_kubernetes_alpha,omitempty" bson:"enable_kubernetes_alpha,omitempty" description:"Alpha enabled clusters are automatically deleted thirty days after [optional]"`
	EnableTpu                      bool                            `json:"enable_tpu,omitempty" bson:"enable_tpu,omitempty" description:"Enable the ability to use Cloud TPUs in this cluster [optional]"`
	Endpoint                       string                          `json:"endpoint,omitempty" bson:"endpoint,omitempty" description:"IP address of this cluster's master endpoint [readonly]"`
	ExpireTime                     string                          `json:"expire_time,omitempty" bson:"expire_time,omitempty" description:"Time the cluster will be automatically deleted [readonly]"`
	InitialClusterVersion          string                          `json:"initial_cluster_version,omitempty" bson:"initial_cluster_version,omitempty" description:"Initial kubernetes version for this cluster [optional]"`
	IpAllocationPolicy             *IPAllocationPolicy             `json:"ip_allocation_policy,omitempty" bson:"ip_allocation_policy,omitempty"`
	LabelFingerprint               string                          `json:"label_fingerprint,omitempty" bson:"label_fingerprint,omitempty" description:"The fingerprint of the set of labels for this cluster [optional]"`
	LegacyAbac                     *LegacyAbac                     `json:"legacy_abac,omitempty" bson:"legacy_abac,omitempty"`
	Location                       string                          `json:"location,omitempty" bson:"location,omitempty" description:"The name of GCP zone or region in which cluster resides [required]"`
	Locations                      []string                        `json:"locations,omitempty" bson:"locations,omitempty" description:"The name of GCP zones in which cluster nodes located [optional]"`
	LoggingService                 string                          `json:"logging_service,omitempty" bson:"logging_service,omitempty" description:"The logging service the cluster should use to write logs [optional]"`
	MaintenancePolicy              *MaintenancePolicy              `json:"maintenance_policy,omitempty" bson:"maintenance_policy,omitempty"`
	MasterAuth                     *MasterAuth                     `json:"master_auth,omitempty" bson:"master_auth,omitempty"`
	MasterAuthorizedNetworksConfig *MasterAuthorizedNetworksConfig `json:"master_authorized_networks_config,omitempty" bson:"master_authorized_networks_config,omitempty"`
	MonitoringService              string                          `json:"monitoring_service,omitempty" bson:"monitoring_service,omitempty" description:"The monitoring service the cluster should use to write metrics [optional]"`
	Name                           string                          `json:"name,omitempty" bson:"name,omitempty" validate:"required" description:"The name of this cluster [required]"`
	Network                        string                          `json:"network,omitempty" bson:"network,omitempty" description:"The name of GCP network to which the cluster connected [required]"`
	NetworkConfig                  *NetworkConfig                  `json:"network_config,omitempty" bson:"network_config,omitempty"`
	NetworkPolicy                  *NetworkPolicy                  `json:"network_policy,omitempty" bson:"network_policy,omitempty"`
	NodeIpv4CidrSize               int64                           `json:"node_ipv4_cidr_size,omitempty" bson:"node_ipv4_cidr_size,omitempty" description:"The size of the address space on each node [readonly]"`
	NodePools                      []*NodePool                     `json:"node_pools,omitempty" bson:"node_pools,omitempty" validate:"required,dive"`
	PrivateClusterConfig           *PrivateClusterConfig           `json:"private_cluster_config,omitempty" bson:"private_cluster_config,omitempty"`
	ResourceLabels                 map[string]string               `json:"resource_labels,omitempty" bson:"resource_labels,omitempty"`
	ResourceUsageExportConfig      *ResourceUsageExportConfig      `json:"resource_usage_export_config,omitempty" bson:"resource_usage_export_config,omitempty"`
	SelfLink                       string                          `json:"self_link,omitempty" bson:"self_link,omitempty" description:"Server-defined URL for the resource [readonly]"`
	ServicesIpv4Cidr               string                          `json:"services_ipv4_cidr,omitempty" bson:"services_ipv4_cidr,omitempty" description:"The IP address range of the kubernetes services in the cluster [readonly]"`
	Status                         string                          `json:"cloud_status,omitempty" bson:"cloud_status,omitempty" description:"The current status of this cluster [readonly]"`
	StatusMessage                  string                          `json:"status_message,omitempty" bson:"status_message,omitempty" description:"Additional information about the current status [readonly]"`
	Subnetwork                     string                          `json:"subnetwork,omitempty" bson:"subnetwork,omitempty" description:"The name of the GCP subnetwork cluster is connected to [required]"`
	TpuIpv4CidrBlock               string                          `json:"tpu_ipv4_cidr_block,omitempty" bson:"tpu_ipv4_cidr_block,omitempty" description:"The IP address range of the Cloud TPUs in the cluster [readonly]"`
	Zone                           string                          `json:"zone,omitempty" bson:"zone,omitempty" validate:"required" description:"Name of GCP zone in which cluster resides [required]"`
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
	MaxPodsPerNode int64 `json:"max_pods_per_node,omitempty" bson:"max_pods_per_node,omitempty" validate:"required" description:"Constraint enforced on the max num of pods per node [required]"`
}

type IPAllocationPolicy struct {
	ClusterIpv4Cidr            string `json:"cluster_ipv4_cidr,omitempty" bson:"cluster_ipv4_cidr,omitempty" validate:"cidrv4"`
	ClusterIpv4CidrBlock       string `json:"cluster_ipv4_cidr_block,omitempty" bson:"cluster_ipv4_cidr_block,omitempty" validate:"cidrv4"`
	ClusterSecondaryRangeName  string `json:"cluster_secondary_range_name,omitempty" bson:"cluster_secondary_range_name,omitempty"`
	CreateSubnetwork           bool   `json:"create_subnetwork,omitempty" bson:"create_subnetwork,omitempty"`
	NodeIpv4Cidr               string `json:"node_ipv4_cidr,omitempty" bson:"node_ipv4_cidr,omitempty" validate:"cidrv4"`
	NodeIpv4CidrBlock          string `json:"node_ipv4_cidr_block,omitempty" bson:"node_ipv4_cidr_block,omitempty" validate:"cidrv4"`
	ServicesIpv4Cidr           string `json:"services_ipv4_cidr,omitempty" bson:"services_ipv4_cidr,omitempty" validate:"cidrv4"`
	ServicesIpv4CidrBlock      string `json:"services_ipv4_cidr_block,omitempty" bson:"services_ipv4_cidr_block,omitempty" validate:"cidrv4"`
	ServicesSecondaryRangeName string `json:"services_secondary_range_name,omitempty" bson:"services_secondary_range_name,omitempty"`
	SubnetworkName             string `json:"subnetwork_name,omitempty" bson:"subnetwork_name,omitempty"`
	TpuIpv4CidrBlock           string `json:"tpu_ipv4_cidr_block,omitempty" bson:"tpu_ipv4_cidr_block,omitempty" validate:"cidrv4"`
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
	Config            *NodeConfig          `json:"config,omitempty" bson:"config,omitempty" validate:"required,dive"`
	InitialNodeCount  int64                `json:"initial_node_count,omitempty" bson:"initial_node_count,omitempty" validate:"required,gte=1"`
	InstanceGroupUrls []string             `json:"instance_group_urls,omitempty" bson:"instance_group_urls,omitempty"`
	Management        *NodeManagement      `json:"management,omitempty" bson:"management,omitempty"`
	MaxPodsConstraint *MaxPodsConstraint   `json:"max_pods_constraint,omitempty" bson:"max_pods_constraint,omitempty" validate:"required,dive"`
	Name              string               `json:"name,omitempty" bson:"name,omitempty" validate:"required"`
	PodIpv4CidrSize   int64                `json:"pod_ipv4_cidr_size,omitempty" bson:"pod_ipv4_cidr_size,omitempty"`
	SelfLink          string               `json:"self_link,omitempty" bson:"self_link,omitempty"`
	Status            string               `json:"status,omitempty" bson:"status,omitempty"`
	StatusMessage     string               `json:"status_message,omitempty" bson:"status_message,omitempty"`
	Version           string               `json:"version,omitempty" bson:"version,omitempty"`
}

type NodePoolAutoscaling struct {
	Enabled      bool  `json:"enabled,omitempty" bson:"enabled,omitempty"`
	MaxNodeCount int64 `json:"max_node_count,omitempty" bson:"max_node_count,omitempty"`
	MinNodeCount int64 `json:"min_node_count,omitempty" bson:"min_node_count,omitempty"`
}

type NodeConfig struct {
	Accelerators   []*AcceleratorConfig `json:"accelerators,omitempty" bson:"accelerators,omitempty"`
	DiskSizeGb     int64                `json:"disk_size_gb,omitempty" bson:"disk_size_gb,omitempty" validate:"required,gte=30"`
	DiskType       string               `json:"disk_type,omitempty" bson:"disk_type,omitempty" validate:"required"`
	ImageType      string               `json:"image_type,omitempty" bson:"image_type,omitempty" validate:"required"`
	Labels         map[string]string    `json:"labels,omitempty" bson:"labels,omitempty"`
	LocalSsdCount  int64                `json:"local_ssd_count,omitempty" bson:"local_ssd_count,omitempty"`
	MachineType    string               `json:"machine_type,omitempty" bson:"machine_type,omitempty" validate:"required"`
	Metadata       map[string]string    `json:"metadata,omitempty" bson:"metadata,omitempty"`
	MinCpuPlatform string               `json:"min_cpu_platform,omitempty" bson:"min_cpu_platform,omitempty"`
	OauthScopes    []string             `json:"oauth_scopes,omitempty" bson:"oauth_scopes,omitempty"`
	Preemptible    bool                 `json:"preemptible,omitempty" bson:"preemptible,omitempty"`
	ServiceAccount string               `json:"service_account,omitempty" bson:"service_account,omitempty" validate:"required"`
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

func GetNetwork(token, projectId string, ctx utils.Context) error {

	url := getNetworkHost("gke", projectId)

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
	err = c.Find(bson.M{"project_id": ctx.Data.ProjectId, "company_id": ctx.Data.Company}).One(&cluster)
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

func GetAllGKECluster(data rbacAuthentication.List, ctx utils.Context) (clusters []GKECluster, err error) {
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
		return clusters, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGKEClusterCollection)
	err = c.Find(bson.M{"project_id": bson.M{"$in": copyData}}).All(&clusters)
	if err != nil {
		ctx.SendLogs(
			"GKEGetAllClusterModel:  GetAll - Got error while fetching from database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return clusters, err
	}

	return clusters, nil
}

func AddGKECluster(cluster GKECluster, ctx utils.Context) error {
	_, err := GetGKECluster(ctx)
	if err == nil {
		text := fmt.Sprintf("GKEAddClusterModel:  Add - Cluster for project '%s' already exists in the database."+err.Error(), cluster.ProjectId)
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
	err = c.Remove(bson.M{"project_id": ctx.Data.ProjectId, "company_id": ctx.Data.Company})
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

func DeployGKECluster(cluster GKECluster, credentials gcp.GcpCredentials, token string, ctx utils.Context) (confError types.CustomCPError) {

	publisher := utils.Notifier{}
	errr := publisher.Init_notifier()

	if errr != nil {
		PrintError(errr, cluster.Name, ctx)
		ctx.SendLogs(errr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := types.CustomCPError{StatusCode: 500, Error: "Error in deploying GKE Cluster", Description: errr.Error()}
		err := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.GKE, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return cpErr
	}

	gkeOps, err := GetGKE(credentials)
	if err != (types.CustomCPError{}) {
		ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.GKE, ctx, err)
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
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.GKE, ctx, err)
		if err_ != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
		return err
	}

	_, _ = utils.SendLog(ctx.Data.Company, "Creating Cluster : "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.ProjectId)
	cluster.CloudplexStatus = (models.Deploying)
	err_ := UpdateGKECluster(cluster, ctx)
	if err_ != nil {

		utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.ProjectId)
		cpErr := types.CustomCPError{Description: confError.Error, Error: "Error occurred while updating cluster status in database", StatusCode: 500}
		err := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.GKE, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}

	err = gkeOps.CreateCluster(cluster, token, ctx)
	if err != (types.CustomCPError{}) {
		cluster.CloudplexStatus = models.ClusterCreationFailed
		confError := UpdateGKECluster(cluster, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, ctx)
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.SendLog(ctx.Data.Company, "Error in cluster creation : "+err.Description, models.LOGGING_LEVEL_ERROR, ctx.Data.ProjectId)
		err_ := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.GKE, ctx, err)
		if err_ != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
		return err
	}
	confError = ApplyAgent(credentials, token, ctx, cluster.Name)
	if confError != (types.CustomCPError{}) {
		cluster.CloudplexStatus = models.AgentDeploymentFailed
		_ = TerminateCluster(credentials, ctx)
		_ = UpdateGKECluster(cluster, ctx)
		err := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.GKE, ctx, confError)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
		return confError
	}

	cluster.CloudplexStatus = models.ClusterCreated

	err1 := UpdateGKECluster(cluster, ctx)
	if err1 != nil {
		PrintError(err1, cluster.Name, ctx)
		cpErr := types.CustomCPError{StatusCode: 500, Error: "Error in applying agent", Description: err1.Error()}

		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
		err := db.CreateError(cluster.ProjectId, ctx.Data.Company, models.GKE, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return cpErr
	}

	_, _ = utils.SendLog(ctx.Data.Company, "Cluster created successfully "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.ProjectId)
	publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
	return types.CustomCPError{}
}

func FetchStatus(credentials gcp.GcpCredentials, token string, ctx utils.Context) (GKECluster, types.CustomCPError) {
	cluster, err := GetGKECluster(ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel:  Fetch -  Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, types.CustomCPError{StatusCode: 500, Error: "Error in fetching status", Description: err.Error()}
	}
	if string(cluster.CloudplexStatus) == strings.ToLower(string(models.New)) {
		cpErr := types.CustomCPError{Error: "Unable to fetch status - Cluster is not deployed yet", Description: "Unable to fetch state - Cluster is not deployed yet", StatusCode: 409}
		return GKECluster{}, cpErr
	}
	if cluster.CloudplexStatus == models.Deploying || cluster.CloudplexStatus == models.Terminating || cluster.CloudplexStatus == models.ClusterTerminated {
		cpErr := types.CustomCPError{Error: "Cluster is in " +
			string(cluster.CloudplexStatus) + " state", Description: "Cluster is in " +
			string(cluster.CloudplexStatus) + " state", StatusCode: 409}
		return GKECluster{}, cpErr
	}

	customErr, err := db.GetError(cluster.ProjectId, ctx.Data.Company, models.GKE, ctx)
	if err != nil {
		return GKECluster{}, types.CustomCPError{Error: "Error occurred while getting cluster status in database",
			Description: "Error occurred while getting cluster status in database",
			StatusCode:  500}
	}
	if customErr.Err != (types.CustomCPError{}) {
		return GKECluster{}, customErr.Err
	}
	gkeOps, err1 := GetGKE(credentials)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GKEClusterModel:  Fetch -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err1
	}

	err1 = gkeOps.init()
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GKEClusterModel:  Fetch -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err1
	}

	latestCluster, err1 := gkeOps.fetchClusterStatus(cluster.Name, ctx)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GKEClusterModel:  Fetch - Failed to get latest status "+err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err1
	}

	latestCluster.ProjectId = ctx.Data.ProjectId
	latestCluster.CompanyId = ctx.Data.Company
	latestCluster.CloudplexStatus = cluster.CloudplexStatus
	latestCluster.IsExpert = cluster.IsExpert
	latestCluster.IsAdvance = cluster.IsAdvance

	return latestCluster, types.CustomCPError{}
}

func TerminateCluster(credentials gcp.GcpCredentials, ctx utils.Context) types.CustomCPError {
	publisher := utils.Notifier{}
	pubErr := publisher.Init_notifier()
	if pubErr != nil {
		err_ := types.CustomCPError{StatusCode: 500, Error: "Error in terminating cluster", Description: pubErr.Error()}
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.GKE, ctx, err_)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return err_
	}

	cluster, err := GetGKECluster(ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Terminate ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		err_ := types.CustomCPError{StatusCode: 500, Error: "Error in terminating cluster", Description: err.Error()}
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.GKE, ctx, err_)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return err_
	}

	if cluster.CloudplexStatus == "" || cluster.CloudplexStatus == "new" {
		text := "GKEClusterModel : Terminate - Cannot terminate a new cluster"
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		err_ := types.CustomCPError{StatusCode: 500, Error: "Error in terminating error", Description: text}
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.GKE, ctx, err_)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
		return err_
	}

	gkeOps, err1 := GetGKE(credentials)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GKEClusterModel : Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.GKE, ctx, err1)
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
			_, _ = utils.SendLog(ctx.Data.Company, "Error in cluster updation in mongo: "+cluster.Name, models.LOGGING_LEVEL_ERROR, cluster.ProjectId)
			_, _ = utils.SendLog(ctx.Data.Company, err.Error(), "error", ctx.Data.ProjectId)
			err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.GKE, ctx, err1)
			if err != nil {
				ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			return err1
		}
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.GKE, ctx, err1)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
		return err1
	}

	_, _ = utils.SendLog(ctx.Data.Company, "Terminating Cluster : "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.ProjectId)
	cluster.CloudplexStatus = models.Terminating
	err_ := UpdateGKECluster(cluster, ctx)
	if err_ != nil {

		utils.SendLog(ctx.Data.Company, err_.Error(), "error", cluster.ProjectId)
		cpErr := types.CustomCPError{Description: err_.Error(), Error: "Error occurred while updating cluster status in database", StatusCode: 500}
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.GKE, ctx, cpErr)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return cpErr
	}
	errr := gkeOps.deleteCluster(cluster, ctx)
	if errr != (types.CustomCPError{}) {
		_, _ = utils.SendLog(ctx.Data.Company, "Cluster termination failed: "+cluster.Name, models.LOGGING_LEVEL_ERROR, ctx.Data.ProjectId)
		utils.SendLog(ctx.Data.Company, err.Error(), models.LOGGING_LEVEL_ERROR, ctx.Data.ProjectId)
		cluster.CloudplexStatus = models.ClusterTerminationFailed
		err = UpdateGKECluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("GKEClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			_, _ = utils.SendLog(ctx.Data.Company, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			_, _ = utils.SendLog(ctx.Data.Company, err.Error(), "error", ctx.Data.ProjectId)
			err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.GKE, ctx, errr)
			if err != nil {
				ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
			return errr
		}
		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.GKE, ctx, errr)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
		return errr
	}

	cluster.CloudplexStatus = models.ClusterTerminated

	err = UpdateGKECluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Terminate "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		_, _ = utils.SendLog(ctx.Data.Company, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
		_, _ = utils.SendLog(ctx.Data.Company, err.Error(), "error", ctx.Data.ProjectId)
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
		err_ := types.CustomCPError{StatusCode: 500, Error: "Error in terminating Cluster", Description: err.Error()}

		err := db.CreateError(ctx.Data.ProjectId, ctx.Data.Company, models.GKE, ctx, err_)
		if err != nil {
			ctx.SendLogs("GKEDeployClusterModel:  Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		return err_
	}
	_, _ = utils.SendLog(ctx.Data.Company, "Cluster terminated successfully "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.ProjectId)
	publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
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
		_, _ = utils.SendLog(ctx.Data.Company, "Cluster creation failed : "+name, models.LOGGING_LEVEL_ERROR, ctx.Data.ProjectId)
		_, _ = utils.SendLog(ctx.Data.Company, confError.Error(), models.LOGGING_LEVEL_ERROR, ctx.Data.Company)
	}
}

func ApplyAgent(credentials gcp.GcpCredentials, token string, ctx utils.Context, clusterName string) (confError types.CustomCPError) {

	data2, err := woodpecker.GetCertificate(ctx.Data.ProjectId, token, ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Apply Agent -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.CustomCPError{StatusCode: 500, Error: "Error in applying agent", Description: err.Error()}
	}
	filePath := "/tmp/" + ctx.Data.Company + "/" + ctx.Data.ProjectId + "/"
	cmd := "mkdir -p " + filePath + " && echo '" + data2 + "'>" + filePath + "agent.yaml && echo '" + credentials.RawData + "'>" + filePath + "gcp-auth.json"
	output, err := models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.CustomCPError{StatusCode: 500, Error: "Error in applying agent", Description: err.Error()}
	}

	if credentials.Zone != "" {
		cmd = "sudo docker run --rm --name " + ctx.Data.Company + ctx.Data.ProjectId + " -e gcpProject=" + credentials.AccountData.ProjectId + " -e cluster=" + clusterName + " -e zone=" + credentials.Region + "-" + credentials.Zone + " -e serviceAccount=" + filePath + "gcp-auth.json" + " -e yamlFile=" + filePath + "agent.yaml -v " + filePath + ":" + filePath + " " + models.GKEAuthContainerName
	} else {
		cmd = "sudo docker run --rm --name " + ctx.Data.Company + ctx.Data.ProjectId + " -e gcpProject=" + credentials.AccountData.ProjectId + " -e cluster=" + clusterName + " -e region=" + credentials.Region + " -e serviceAccount=" + filePath + "gcp-auth.json" + " -e yamlFile=" + filePath + "agent.yaml -v " + filePath + ":" + filePath + " " + models.GKEAuthContainerName
	}

	output, err = models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.CustomCPError{StatusCode: 500, Error: "Error in applying agent", Description: err.Error()}
	}
	return types.CustomCPError{}
}

func Validate(gkeCluster GKECluster) error {
	if gkeCluster.ProjectId == "" {
		return errors.New("project id is empty")
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
