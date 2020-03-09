package gke

import (
	"antelope/models"
	"antelope/models/db"
	"antelope/models/utils"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	gke "google.golang.org/api/container/v1"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type GKEClusterTemplate struct {
	ID               bson.ObjectId `json:"-" bson:"_id,omitempty"`
	TemplateId       string        `json:"template_id" bson:"template_id"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	CloudplexStatus  string        `json:"status" bson:"status"`
	CompanyId        string        `json:"company_id" bson:"company_id"`

	// AddonsConfig: Configurations for the various addons available to run
	// in the cluster.
	AddonsConfig *gke.AddonsConfig `json:"addons_config,omitempty" bson:"addons_config,omitempty"`

	// ClusterIpv4Cidr: The IP address range of the container pods in this
	// cluster,
	// in
	// [CIDR](http://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing)
	//
	// notation (e.g. `10.96.0.0/14`). Leave blank to have
	// one automatically chosen or specify a `/14` block in `10.0.0.0/8`.
	ClusterIpv4Cidr string `json:"cluster_ipv4_cidr,omitempty" bson:"cluster_ipv4_cidr,omitempty"`

	// Conditions: Which conditions caused the current cluster state.
	Conditions []*gke.StatusCondition `json:"conditions,omitempty" bson:"conditions,omitempty"`

	// CreateTime: [Output only] The time the cluster was created,
	// in
	// [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) text format.
	CreateTime string `json:"create_time,omitempty" bson:"create_time,omitempty"`

	// CurrentMasterVersion: [Output only] The current software version of
	// the master endpoint.
	CurrentMasterVersion string `json:"current_master_version,omitempty" bson:"current_master_version,omitempty"`

	// DefaultMaxPodsConstraint: The default constraint on the maximum
	// number of pods that can be run
	// simultaneously on a node in the node pool of this cluster. Only
	// honored
	// if cluster created with IP Alias support.
	DefaultMaxPodsConstraint *gke.MaxPodsConstraint `json:"default_max_pods_constraint,omitempty" bson:"default_max_pods_constraint,omitempty"`

	// Description: An optional description of this cluster.
	Description string `json:"description,omitempty" bson:"description,omitempty"`

	// EnableKubernetesAlpha: Kubernetes alpha features are enabled on this
	// cluster. This includes alpha
	// API groups (e.g. v1alpha1) and features that may not be production
	// ready in
	// the kubernetes version of the master and nodes.
	// The cluster has no SLA for uptime and master/node upgrades are
	// disabled.
	// Alpha enabled clusters are automatically deleted thirty days
	// after
	// creation.
	EnableKubernetesAlpha bool `json:"enable_kubernetes_alpha,omitempty" bson:"enable_kubernetes_alpha,omitempty"`

	// EnableTpu: Enable the ability to use Cloud TPUs in this cluster.
	EnableTpu bool `json:"enable_tpu,omitempty" bson:"enable_tpu,omitempty"`

	// Endpoint: [Output only] The IP address of this cluster's master
	// endpoint.
	// The endpoint can be accessed from the internet
	// at
	// `https://username:password@endpoint/`.
	//
	// See the `masterAuth` property of this resource for username
	// and
	// password information.
	Endpoint string `json:"endpoint,omitempty" bson:"endpoint,omitempty"`

	// ExpireTime: [Output only] The time the cluster will be
	// automatically
	// deleted in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) text
	// format.
	ExpireTime string `json:"expire_time,omitempty" bson:"expire_time,omitempty"`

	// InitialClusterVersion: The initial Kubernetes version for this
	// cluster.  Valid versions are those
	// found in validMasterVersions returned by getServerConfig.  The
	// version can
	// be upgraded over time; such upgrades are reflected
	// in
	// currentMasterVersion and currentNodeVersion.
	//
	// Users may specify either explicit versions offered by
	// Kubernetes Engine or version aliases, which have the following
	// behavior:
	//
	// - "latest": picks the highest valid Kubernetes version
	// - "1.X": picks the highest valid patch+gke.N patch in the 1.X
	// version
	// - "1.X.Y": picks the highest valid gke.N patch in the 1.X.Y version
	// - "1.X.Y-gke.N": picks an explicit Kubernetes version
	// - "","-": picks the default Kubernetes version
	InitialClusterVersion string `json:"initial_cluster_version,omitempty" bson:"initial_cluster_version,omitempty"`

	// IpAllocationPolicy: Configuration for cluster IP allocation.
	IpAllocationPolicy *gke.IPAllocationPolicy `json:"ip_allocation_policy,omitempty" bson:"ip_allocation_policy,omitempty"`

	// LabelFingerprint: The fingerprint of the set of labels for this
	// cluster.
	LabelFingerprint string `json:"label_fingerprint,omitempty" bson:"label_fingerprint,omitempty"`

	// LegacyAbac: Configuration for the legacy ABAC authorization mode.
	LegacyAbac *gke.LegacyAbac `json:"legacy_abac,omitempty" bson:"legacy_abac,omitempty"`

	// Location: [Output only] The name of the Google Compute
	// Engine
	// [zone](/compute/docs/regions-zones/regions-zones#available)
	// or
	// [region](/compute/docs/regions-zones/regions-zones#available) in
	// which
	// the cluster resides.
	Location string `json:"location,omitempty" bson:"location,omitempty"`

	// Locations: The list of Google Compute
	// Engine
	// [zones](/compute/docs/zones#available) in which the cluster's
	// nodes
	// should be located.
	Locations []string `json:"locations,omitempty" bson:"locations,omitempty"`

	// LoggingService: The logging service the cluster should use to write
	// logs.
	// Currently available options:
	//
	// * "logging.googleapis.com/kubernetes" - the Google Cloud
	// Logging
	// service with Kubernetes-native resource model in Stackdriver
	// * `logging.googleapis.com` - the Google Cloud Logging service.
	// * `none` - no logs will be exported from the cluster.
	// * if left as an empty string,`logging.googleapis.com` will be used.
	LoggingService string `json:"logging_service,omitempty" bson:"logging_service,omitempty"`

	// MaintenancePolicy: Configure the maintenance policy for this cluster.
	MaintenancePolicy *gke.MaintenancePolicy `json:"maintenance_policy,omitempty" bson:"maintenance_policy,omitempty"`

	// MasterAuth: The authentication information for accessing the master
	// endpoint.
	// If unspecified, the defaults are used:
	// For clusters before v1.12, if master_auth is unspecified, `username`
	// will
	// be set to "admin", a random password will be generated, and a
	// client
	// certificate will be issued.
	MasterAuth *gke.MasterAuth `json:"master_auth,omitempty" bson:"master_auth,omitempty"`

	// MasterAuthorizedNetworksConfig: The configuration options for master
	// authorized networks feature.
	MasterAuthorizedNetworksConfig *gke.MasterAuthorizedNetworksConfig `json:"master_authorized_networks_config,omitempty" bson:"master_authorized_networks_config,omitempty"`

	// MonitoringService: The monitoring service the cluster should use to
	// write metrics.
	// Currently available options:
	//
	// * `monitoring.googleapis.com` - the Google Cloud Monitoring
	// service.
	// * `none` - no metrics will be exported from the cluster.
	// * if left as an empty string, `monitoring.googleapis.com` will be
	// used.
	MonitoringService string `json:"monitoring_service,omitempty" bson:"monitoring_service,omitempty"`

	// Name: The name of this cluster. The name must be unique within this
	// project
	// and zone, and can be up to 40 characters with the following
	// restrictions:
	//
	// * Lowercase letters, numbers, and hyphens only.
	// * Must start with a letter.
	// * Must end with a number or a letter.
	Name string `json:"name,omitempty" bson:"name,omitempty"`

	// Network: The name of the Google Compute
	// Engine
	// [network](/compute/docs/networks-and-firewalls#networks) to which
	// the
	// cluster is connected. If left unspecified, the `default` network
	// will be used.
	Network string `json:"network,omitempty" bson:"network,omitempty"`

	// NetworkConfig: Configuration for cluster networking.
	NetworkConfig *gke.NetworkConfig `json:"network_config,omitempty" bson:"network_config,omitempty"`

	// NetworkPolicy: Configuration options for the NetworkPolicy feature.
	NetworkPolicy *gke.NetworkPolicy `json:"network_policy,omitempty" bson:"network_policy,omitempty"`

	// NodeIpv4CidrSize: [Output only] The size of the address space on each
	// node for hosting
	// containers. This is provisioned from within the
	// `container_ipv4_cidr`
	// range. This field will only be set when cluster is in route-based
	// network
	// mode.
	NodeIpv4CidrSize int64 `json:"node_ipv4_cidr_size,omitempty" bson:"node_ipv4_cidr_size,omitempty"`

	// NodePools: The node pools associated with this cluster.
	// This field should not be set if "node_config" or "initial_node_count"
	// are
	// specified.
	NodePools []*gke.NodePool `json:"node_pools,omitempty" bson:"node_pools,omitempty"`

	// PrivateClusterConfig: Configuration for private cluster.
	PrivateClusterConfig *gke.PrivateClusterConfig `json:"private_cluster_config,omitempty" bson:"private_cluster_config,omitempty"`

	// ResourceLabels: The resource labels for the cluster to use to
	// annotate any related
	// Google Compute Engine resources.
	ResourceLabels map[string]string `json:"resource_labels,omitempty" bson:"resource_labels,omitempty"`

	// ResourceUsageExportConfig: Configuration for exporting resource
	// usages. Resource usage export is
	// disabled when this config is unspecified.
	ResourceUsageExportConfig *gke.ResourceUsageExportConfig `json:"resource_usage_export_config,omitempty" bson:"resource_usage_export_config,omitempty"`

	// SelfLink: [Output only] Server-defined URL for the resource.
	SelfLink string `json:"self_link,omitempty" bson:"self_link,omitempty"`

	// ServicesIpv4Cidr: [Output only] The IP address range of the
	// Kubernetes services in
	// this cluster,
	// in
	// [CIDR](http://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing)
	//
	// notation (e.g. `1.2.3.4/29`). Service addresses are
	// typically put in the last `/16` from the container CIDR.
	ServicesIpv4Cidr string `json:"services_ipv4_cidr,omitempty" bson:"services_ipv4_cidr,omitempty"`

	// Status: [Output only] The current status of this cluster.
	//
	// Possible values:
	//   "STATUS_UNSPECIFIED" - Not set.
	//   "PROVISIONING" - The PROVISIONING state indicates the cluster is
	// being created.
	//   "RUNNING" - The RUNNING state indicates the cluster has been
	// created and is fully
	// usable.
	//   "RECONCILING" - The RECONCILING state indicates that some work is
	// actively being done on
	// the cluster, such as upgrading the master or node software. Details
	// can
	// be found in the `statusMessage` field.
	//   "STOPPING" - The STOPPING state indicates the cluster is being
	// deleted.
	//   "ERROR" - The ERROR state indicates the cluster may be unusable.
	// Details
	// can be found in the `statusMessage` field.
	//   "DEGRADED" - The DEGRADED state indicates the cluster requires user
	// action to restore
	// full functionality. Details can be found in the `statusMessage`
	// field.
	Status string `json:"cloud_status,omitempty" bson:"cloud_status,omitempty"`

	// StatusMessage: [Output only] Additional information about the current
	// status of this
	// cluster, if available.
	StatusMessage string `json:"status_message,omitempty" bson:"status_message,omitempty"`

	// Subnetwork: The name of the Google Compute
	// Engine
	// [subnetwork](/compute/docs/subnetworks) to which the
	// cluster is connected.
	Subnetwork string `json:"subnetwork,omitempty" bson:"subnetwork,omitempty"`

	// TpuIpv4CidrBlock: [Output only] The IP address range of the Cloud
	// TPUs in this cluster,
	// in
	// [CIDR](http://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing)
	//
	// notation (e.g. `1.2.3.4/29`).
	TpuIpv4CidrBlock string `json:"tpu_ipv4_cidr_block,omitempty" bson:"tpu_ipv4_cidr_block,omitempty"`

	// Zone: [Output only] The name of the Google Compute
	// Engine
	// [zone](/compute/docs/zones#available) in which the
	// cluster
	// resides.
	// This field is deprecated, use location instead.
	Zone string `json:"zone,omitempty"`
}

func GetGKEClusterTemplate(templateId string, companyId string, ctx utils.Context) (cluster GKEClusterTemplate, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs(
			"GKEGetClusterTemplateModel:  Get - Got error while connecting to the database: "+err1.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGKETemplateCollection)
	err = c.Find(bson.M{"template_id": templateId, "company_id": companyId}).One(&cluster)
	if err != nil {
		ctx.SendLogs(
			"GKEGetClusterTemplateModel:  Get - Got error while fetching from database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err
	}

	return cluster, nil
}

func GetAllGKEClusterTemplate(ctx utils.Context) (templates []GKEClusterTemplate, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("GKEGetAllClusterTemplateModel : GetAll - Got error while connecting to the database:  "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGKETemplateCollection)
	err = c.Find(bson.M{}).All(&templates)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}
	return templates, nil
}

func AddGKEClusterTemplate(cluster GKEClusterTemplate, ctx utils.Context) (string, error) {
	_, err := GetGKEClusterTemplate(cluster.TemplateId, cluster.CompanyId, ctx)
	if err == nil {
		text := fmt.Sprintf("GKEAddClusterTemplateModel:  Add - Cluster template for project '%s' already exists in the database.", cluster.TemplateId)
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", errors.New(text)
	}

	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"GKEAddClusterTemplateModel:  Add - Got error while connecting to the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return "", err
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
	err = db.InsertInMongo(mc.MongoGKETemplateCollection, cluster)
	if err != nil {
		ctx.SendLogs(
			"GKEAddClusterTemplateModel:  Add - Got error while inserting cluster to the database:  "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return "", err
	}

	return cluster.TemplateId, nil
}

func UpdateGKEClusterTemplate(cluster GKEClusterTemplate, ctx utils.Context) error {
	oldCluster, err := GetGKEClusterTemplate(cluster.TemplateId, cluster.CompanyId, ctx)
	if err != nil {
		text := "GKEUpdateClusterTemplateModel:  Update - Cluster '" + cluster.Name + "' does not exist in the database: " + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	err = DeleteGKEClusterTemplate(cluster.TemplateId, cluster.CompanyId, ctx)
	if err != nil {
		ctx.SendLogs(
			"GKEUpdateClusterTemplateModel:  Update - Got error deleting cluster template: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	cluster.CreationDate = oldCluster.CreationDate
	cluster.ModificationDate = time.Now()

	_, err = AddGKEClusterTemplate(cluster, ctx)
	if err != nil {
		ctx.SendLogs(
			"GKEUpdateClusterTemplateModel:  Update - Got error creating cluster template: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func DeleteGKEClusterTemplate(templateId, companyId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"GKEDeleteClusterTemplateModel:  Delete - Got error while connecting to the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGKETemplateCollection)
	err = c.Remove(bson.M{"template_id": templateId, "company_id": companyId})
	if err != nil {
		ctx.SendLogs(
			"GKEDeleteClusterTemplateModel:  Delete - Got error while deleting from the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}