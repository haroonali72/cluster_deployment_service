package gke

import (
	"antelope/models"
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
	"time"
)

type GKECluster struct {
	ID                             bson.ObjectId                   `json:"-" bson:"_id,omitempty"`
	ProjectId                      string                          `json:"project_id" bson:"project_id"`
	Cloud                          models.Cloud                    `json:"cloud" bson:"cloud"`
	CreationDate                   time.Time                       `json:"-" bson:"creation_date"`
	ModificationDate               time.Time                       `json:"-" bson:"modification_date"`
	CloudplexStatus                string                          `json:"status" bson:"status"`
	CompanyId                      string                          `json:"company_id" bson:"company_id"`
	IsExpert                       bool                            `json:"is_expert" bson:"is_expert"`
	IsAdvance                      bool                            `json:"is_advance" bson:"is_advance"`
	AddonsConfig                   *AddonsConfig                   `json:"addons_config,omitempty" bson:"addons_config,omitempty"`
	ClusterIpv4Cidr                string                          `json:"cluster_ipv4_cidr,omitempty" bson:"cluster_ipv4_cidr,omitempty"`
	Conditions                     []*StatusCondition              `json:"conditions,omitempty" bson:"conditions,omitempty"`
	CreateTime                     string                          `json:"create_time,omitempty" bson:"create_time,omitempty"`
	CurrentMasterVersion           string                          `json:"current_master_version,omitempty" bson:"current_master_version,omitempty"`
	CurrentNodeCount               int64                           `json:"current_node_count,omitempty" bson:"current_node_count,omitempty"`
	DefaultMaxPodsConstraint       *MaxPodsConstraint              `json:"default_max_pods_constraint,omitempty" bson:"default_max_pods_constraint,omitempty"`
	Description                    string                          `json:"description,omitempty" bson:"description,omitempty"`
	EnableKubernetesAlpha          bool                            `json:"enable_kubernetes_alpha,omitempty" bson:"enable_kubernetes_alpha,omitempty"`
	EnableTpu                      bool                            `json:"enable_tpu,omitempty" bson:"enable_tpu,omitempty"`
	Endpoint                       string                          `json:"endpoint,omitempty" bson:"endpoint,omitempty"`
	ExpireTime                     string                          `json:"expire_time,omitempty" bson:"expire_time,omitempty"`
	InitialClusterVersion          string                          `json:"initial_cluster_version,omitempty" bson:"initial_cluster_version,omitempty"`
	IpAllocationPolicy             *IPAllocationPolicy             `json:"ip_allocation_policy,omitempty" bson:"ip_allocation_policy,omitempty"`
	LabelFingerprint               string                          `json:"label_fingerprint,omitempty" bson:"label_fingerprint,omitempty"`
	LegacyAbac                     *LegacyAbac                     `json:"legacy_abac,omitempty" bson:"legacy_abac,omitempty"`
	Location                       string                          `json:"location,omitempty" bson:"location,omitempty"`
	Locations                      []string                        `json:"locations,omitempty" bson:"locations,omitempty"`
	LoggingService                 string                          `json:"logging_service,omitempty" bson:"logging_service,omitempty"`
	MaintenancePolicy              *MaintenancePolicy              `json:"maintenance_policy,omitempty" bson:"maintenance_policy,omitempty"`
	MasterAuth                     *MasterAuth                     `json:"master_auth,omitempty" bson:"master_auth,omitempty"`
	MasterAuthorizedNetworksConfig *MasterAuthorizedNetworksConfig `json:"master_authorized_networks_config,omitempty" bson:"master_authorized_networks_config,omitempty"`
	MonitoringService              string                          `json:"monitoring_service,omitempty" bson:"monitoring_service,omitempty"`
	Name                           string                          `json:"name,omitempty" bson:"name,omitempty"`
	Network                        string                          `json:"network,omitempty" bson:"network,omitempty"`
	NetworkConfig                  *NetworkConfig                  `json:"network_config,omitempty" bson:"network_config,omitempty"`
	NetworkPolicy                  *NetworkPolicy                  `json:"network_policy,omitempty" bson:"network_policy,omitempty"`
	NodeIpv4CidrSize               int64                           `json:"node_ipv4_cidr_size,omitempty" bson:"node_ipv4_cidr_size,omitempty"`
	NodePools                      []*NodePool                     `json:"node_pools,omitempty" bson:"node_pools,omitempty"`
	PrivateClusterConfig           *PrivateClusterConfig           `json:"private_cluster_config,omitempty" bson:"private_cluster_config,omitempty"`
	ResourceLabels                 map[string]string               `json:"resource_labels,omitempty" bson:"resource_labels,omitempty"`
	ResourceUsageExportConfig      *ResourceUsageExportConfig      `json:"resource_usage_export_config,omitempty" bson:"resource_usage_export_config,omitempty"`
	SelfLink                       string                          `json:"self_link,omitempty" bson:"self_link,omitempty"`
	ServicesIpv4Cidr               string                          `json:"services_ipv4_cidr,omitempty" bson:"services_ipv4_cidr,omitempty"`
	Status                         string                          `json:"cloud_status,omitempty" bson:"cloud_status,omitempty"`
	StatusMessage                  string                          `json:"status_message,omitempty" bson:"status_message,omitempty"`
	Subnetwork                     string                          `json:"subnetwork,omitempty" bson:"subnetwork,omitempty"`
	TpuIpv4CidrBlock               string                          `json:"tpu_ipv4_cidr_block,omitempty" bson:"tpu_ipv4_cidr_block,omitempty"`
	Zone                           string                          `json:"zone,omitempty" bson:"zone,omitempty"`
}

type AddonsConfig struct {
	HorizontalPodAutoscaling *HorizontalPodAutoscaling `json:"horizontal_pod_autoscaling,omitempty" bson:"horizontal_pod_autoscaling,omitempty"`
	HttpLoadBalancing        *HttpLoadBalancing        `json:"http_load_balancing,omitempty" bson:"http_load_balancing,omitempty"`
	KubernetesDashboard      *KubernetesDashboard      `json:"kubernetes_dashboard,omitempty" bson:"kubernetes_dashboard,omitempty"`
	NetworkPolicyConfig      *NetworkPolicyConfig      `json:"network_policy_config,omitempty" bson:"network_policy_config,omitempty"`
}

type HorizontalPodAutoscaling struct {
	Disabled bool `json:"disabled,omitempty" bson:"disabled,omitempty"`
}

type HttpLoadBalancing struct {
	Disabled bool `json:"disabled,omitempty" bson:"disabled,omitempty"`
}

type KubernetesDashboard struct {
	Disabled bool `json:"disabled,omitempty" bson:"disabled,omitempty"`
}

type NetworkPolicyConfig struct {
	Disabled bool `json:"disabled,omitempty" bson:"disabled,omitempty"`
}

type StatusCondition struct {
	Code    string `json:"code,omitempty" bson:"code,omitempty"`
	Message string `json:"message,omitempty" bson:"message,omitempty"`
}

type MaxPodsConstraint struct {
	MaxPodsPerNode int64 `json:"max_pods_per_node,omitempty" bson:"max_pods_per_node,omitempty"`
}

type IPAllocationPolicy struct {
	ClusterIpv4Cidr            string `json:"cluster_ipv4_cidr,omitempty" bson:"cluster_ipv4_cidr,omitempty"`
	ClusterIpv4CidrBlock       string `json:"cluster_ipv4_cidr_block,omitempty" bson:"cluster_ipv4_cidr_block,omitempty"`
	ClusterSecondaryRangeName  string `json:"cluster_secondary_range_name,omitempty" bson:"cluster_secondary_range_name,omitempty"`
	CreateSubnetwork           bool   `json:"create_subnetwork,omitempty" bson:"create_subnetwork,omitempty"`
	NodeIpv4Cidr               string `json:"node_ipv4_cidr,omitempty" bson:"node_ipv4_cidr,omitempty"`
	NodeIpv4CidrBlock          string `json:"node_ipv4_cidr_block,omitempty" bson:"node_ipv4_cidr_block,omitempty"`
	ServicesIpv4Cidr           string `json:"services_ipv4_cidr,omitempty" bson:"services_ipv4_cidr,omitempty"`
	ServicesIpv4CidrBlock      string `json:"services_ipv4_cidr_block,omitempty" bson:"services_ipv4_cidr_block,omitempty"`
	ServicesSecondaryRangeName string `json:"services_secondary_range_name,omitempty" bson:"services_secondary_range_name,omitempty"`
	SubnetworkName             string `json:"subnetwork_name,omitempty" bson:"subnetwork_name,omitempty"`
	TpuIpv4CidrBlock           string `json:"tpu_ipv4_cidr_block,omitempty" bson:"tpu_ipv4_cidr_block,omitempty"`
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
	Config            *NodeConfig          `json:"config,omitempty" bson:"config,omitempty"`
	InitialNodeCount  int64                `json:"initial_node_count,omitempty" bson:"initial_node_count,omitempty"`
	InstanceGroupUrls []string             `json:"instance_group_urls,omitempty" bson:"instance_group_urls,omitempty"`
	Management        *NodeManagement      `json:"management,omitempty" bson:"management,omitempty"`
	MaxPodsConstraint *MaxPodsConstraint   `json:"max_pods_constraint,omitempty" bson:"max_pods_constraint,omitempty"`
	Name              string               `json:"name,omitempty" bson:"name,omitempty"`
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
	DiskSizeGb     int64                `json:"disk_size_gb,omitempty" bson:"disk_size_gb,omitempty"`
	DiskType       string               `json:"disk_type,omitempty" bson:"disk_type,omitempty"`
	ImageType      string               `json:"image_type,omitempty" bson:"image_type,omitempty"`
	Labels         map[string]string    `json:"labels,omitempty" bson:"labels,omitempty"`
	LocalSsdCount  int64                `json:"local_ssd_count,omitempty" bson:"local_ssd_count,omitempty"`
	MachineType    string               `json:"machine_type,omitempty" bson:"machine_type,omitempty"`
	Metadata       map[string]string    `json:"metadata,omitempty" bson:"metadata,omitempty"`
	MinCpuPlatform string               `json:"min_cpu_platform,omitempty" bson:"min_cpu_platform,omitempty"`
	OauthScopes    []string             `json:"oauth_scopes,omitempty" bson:"oauth_scopes,omitempty"`
	Preemptible    bool                 `json:"preemptible,omitempty" bson:"preemptible,omitempty"`
	ServiceAccount string               `json:"service_account,omitempty" bson:"service_account,omitempty"`
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

func GetGKECluster( ctx utils.Context) (cluster GKECluster, err error) {
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
	err = c.Find(bson.M{"project_id":ctx.Data.ProjectId, "company_id": ctx.Data.Company}).One(&cluster)
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
	_, err := GetGKECluster( ctx)
	if err == nil {
		text := fmt.Sprintf("GKEAddClusterModel:  Add - Cluster for project '%s' already exists in the database.", cluster.ProjectId)
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
	oldCluster, err := GetGKECluster( ctx)
	if err != nil {
		text := "GKEUpdateClusterModel:  Update - Cluster '" + cluster.Name + "' does not exist in the database: " + err.Error()
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

func DeleteGKECluster( ctx utils.Context) error {
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
		PrintError(errr, cluster.Name,ctx)
		ctx.SendLogs(errr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.CustomCPError{StatusCode:"500",Description:errr.Error()}
	}

	gkeOps, err := GetGKE(credentials)
	if err.Description != "" {
		ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	err = gkeOps.init()
	if err.Description != "" {
		ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.CloudplexStatus = "Cluster creation failed"
		confError := UpdateGKECluster(cluster, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name,ctx)
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
		return err
	}

	_, _ = utils.SendLog(ctx.Data.Company, "Creating Cluster : "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.ProjectId)
	err = gkeOps.CreateCluster(cluster, token, ctx)
	if err.Description != "" {
		cluster.CloudplexStatus = "Cluster creation failed"
		confError := UpdateGKECluster(cluster, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, ctx)
			ctx.SendLogs("GKEDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		utils.SendLog(ctx.Data.Company, "Error in cluster creation : "+err.Description, models.LOGGING_LEVEL_ERROR, ctx.Data.ProjectId)
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
		return types.CustomCPError{}
	}
	confError = ApplyAgent(credentials, token, ctx, cluster.Name)
	if confError.Description != "" {
		cluster.CloudplexStatus = "Cluster creation failed"
		_ = UpdateGKECluster(cluster, ctx)
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
		return types.CustomCPError{}
	}
	cluster.CloudplexStatus = "Cluster Created"

	err1 := UpdateGKECluster(cluster, ctx)
	if err1 != nil {
		PrintError(err1, cluster.Name,ctx)
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
		return confError
	}

	_, _ = utils.SendLog(ctx.Data.Company, "Cluster created successfully "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.ProjectId)
	publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
	return types.CustomCPError{}
}

func FetchStatus(credentials gcp.GcpCredentials, token string, ctx utils.Context) (GKECluster, types.CustomCPError) {
	cluster, err := GetGKECluster( ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel:  Fetch -  Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, types.CustomCPError{Description:err.Error()}
	}

	gkeOps, err1 := GetGKE(credentials)
	if err1.Description != ""{
		ctx.SendLogs("GKEClusterModel:  Fetch -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err1
	}

	err1 = gkeOps.init()
	if err1.Description != "" {
		ctx.SendLogs("GKEClusterModel:  Fetch -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err1
	}

	latestCluster, err1 := gkeOps.fetchClusterStatus(cluster.Name, ctx)
	if err1.Description != "" {
		ctx.SendLogs("GKEClusterModel:  Fetch - Failed to get latest status "+err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err1
	}

	latestCluster.ProjectId =ctx.Data.ProjectId
	latestCluster.CompanyId = ctx.Data.Company
	latestCluster.CloudplexStatus = cluster.CloudplexStatus
	latestCluster.IsExpert = cluster.IsExpert
	latestCluster.IsAdvance = cluster.IsAdvance

	return latestCluster, types.CustomCPError{}
}

func TerminateCluster(credentials gcp.GcpCredentials, ctx utils.Context)types.CustomCPError {
	publisher := utils.Notifier{}
	pubErr := publisher.Init_notifier()
	if pubErr != nil {
		return types.CustomCPError{Description:pubErr.Error()}
	}

	cluster, err := GetGKECluster( ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Terminate ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.CustomCPError{Description:err.Error()}
	}

	if cluster.CloudplexStatus == "" || cluster.CloudplexStatus == "new" {
		text := "GKEClusterModel : Terminate - Cannot terminate a new cluster"
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
		return types.CustomCPError{Description:text}
	}

	gkeOps, err1 := GetGKE(credentials)
	if err1.Description != "" {
		ctx.SendLogs("GKEClusterModel : Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err1
	}

	cluster.CloudplexStatus = string(models.Terminating)

	err1 = gkeOps.init()
	if err1.Description != "" {
		ctx.SendLogs("GKEClusterModel : Terminate -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.CloudplexStatus = "Cluster Termination Failed"
		err = UpdateGKECluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("GKEClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			_, _ = utils.SendLog(ctx.Data.Company, "Error in cluster updation in mongo: "+cluster.Name, models.LOGGING_LEVEL_ERROR, cluster.ProjectId)
			_, _ = utils.SendLog(ctx.Data.Company, err.Error(), "error", ctx.Data.ProjectId)
			return types.CustomCPError{StatusCode:"500",Description:err.Error(),}
		}
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
		return err1
	}

	_, _ = utils.SendLog(ctx.Data.Company, "Terminating Cluster : "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.ProjectId)
	errr := gkeOps.deleteCluster(cluster, ctx)
	if errr.Description != "" {
		_, _ = utils.SendLog(ctx.Data.Company, "Cluster termination failed: "+cluster.Name , models.LOGGING_LEVEL_ERROR, ctx.Data.ProjectId)
		utils.SendLog(ctx.Data.Company, err.Error(), models.LOGGING_LEVEL_ERROR, ctx.Data.ProjectId)
		cluster.CloudplexStatus = "Cluster Termination Failed"
		err = UpdateGKECluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("GKEClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			_, _ = utils.SendLog(ctx.Data.Company, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			_, _ = utils.SendLog(ctx.Data.Company, err.Error(), "error", ctx.Data.ProjectId)
			publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
			return types.CustomCPError{StatusCode:"500",Description:err.Error(),}
		}
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
		return errr
	}

	cluster.CloudplexStatus = "Cluster Terminated"

	err = UpdateGKECluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Terminate "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		_, _ = utils.SendLog(ctx.Data.Company, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
		_, _ = utils.SendLog(ctx.Data.Company, err.Error(), "error", ctx.Data.ProjectId)
		publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
		return types.CustomCPError{StatusCode:"500",Description:err.Error(),}
	}
	_, _ = utils.SendLog(ctx.Data.Company, "Cluster terminated successfully "+cluster.Name, models.LOGGING_LEVEL_INFO, ctx.Data.ProjectId)
	publisher.Notify(ctx.Data.ProjectId, "Status Available", ctx)
	return types.CustomCPError{}
}

func GetServerConfig(credentials gcp.GcpCredentials, ctx utils.Context) (*gke.ServerConfig, types.CustomCPError) {
	gkeOps, err := GetGKE(credentials)
	if err.Description != "" {
		ctx.SendLogs("GKEClusterModel : GetServerConfig - "+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	err = gkeOps.init()
	if err.Description != "" {
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
		return types.CustomCPError{StatusCode:"500",Description:err.Error(),}
	}
	filePath := "/tmp/" + ctx.Data.Company + "/" + ctx.Data.ProjectId + "/"
	cmd := "mkdir -p " + filePath + " && echo '" + data2 + "'>" + filePath + "agent.yaml && echo '" + credentials.RawData + "'>" + filePath + "gcp-auth.json"
	output, err := models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.CustomCPError{StatusCode:"500",Description:err.Error(),}
	}

	if credentials.Zone != "" {
		cmd = "sudo docker run --rm --name " + ctx.Data.Company+ ctx.Data.ProjectId + " -e gcpProject=" + credentials.AccountData.ProjectId + " -e cluster=" + clusterName + " -e zone=" + credentials.Region + "-" + credentials.Zone + " -e serviceAccount=" + filePath + "gcp-auth.json" + " -e yamlFile=" + filePath + "agent.yaml -v " + filePath + ":" + filePath + " " + models.GKEAuthContainerName
	} else {
		cmd = "sudo docker run --rm --name " + ctx.Data.Company + ctx.Data.ProjectId + " -e gcpProject=" + credentials.AccountData.ProjectId + " -e cluster=" + clusterName + " -e region=" + credentials.Region + " -e serviceAccount=" + filePath + "gcp-auth.json" + " -e yamlFile=" + filePath + "agent.yaml -v " + filePath + ":" + filePath + " " + models.GKEAuthContainerName
	}

	output, err = models.RemoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.CustomCPError{StatusCode:"500",Description:err.Error(),}
	}
	return types.CustomCPError{}
}
