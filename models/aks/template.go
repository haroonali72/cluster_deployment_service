package aks

import (
	"antelope/models"
	"antelope/models/db"
	"antelope/models/utils"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type AKSClusterTemplate struct {
	ID               bson.ObjectId `json:"-" bson:"_id,omitempty"`
	TemplateId       string        `json:"template_id" bson:"template_id"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	CloudplexStatus  string        `json:"status" bson:"status"`
	CompanyId        string        `json:"company_id" bson:"company_id"`
	Status           string        `json:"status" bson:"status"`
	ResourceGoup     string        `json:"resource_group" bson:"resource_group"`
	// ManagedClusterProperties - Properties of a managed cluster.
	ClusterProperties *ManagedClusterPropertiesTemplate `json:"properties,omitempty"`
	// ID - Resource Id
	ResourceID *string `json:"id,omitempty"`
	// Name - Resource name
	Name *string `json:"name,omitempty"`
	// Type - Resource type
	Type *string `json:"type,omitempty"`
	// Location - Resource location
	Location *string `json:"location,omitempty"`
	// Tags - Resource tags
	Tags map[string]*string `json:"tags"`
}

type ManagedClusterPropertiesTemplate struct {
	// ProvisioningState - The current deployment or provisioning state, which only appears in the response.
	ProvisioningState *string `json:"provisioningState,omitempty"`
	// KubernetesVersion - Version of Kubernetes specified when creating the managed cluster.
	KubernetesVersion *string `json:"kubernetesVersion,omitempty"`
	// AgentPoolProfiles - Properties of the agent pool. Currently only one agent pool can exist.
	AgentPoolProfiles []ManagedClusterAgentPoolProfileTemplate `json:"agentPoolProfiles,omitempty"`
}

// ManagedClusterAgentPoolProfile profile for the container service agent pool.
type ManagedClusterAgentPoolProfileTemplate struct {
	// Name - Unique name of the agent pool profile in the context of the subscription and resource group.
	Name *string `json:"name,omitempty"`
	// Count - Number of agents (VMs) to host docker containers. Allowed values must be in the range of 1 to 100 (inclusive). The default value is 1.
	Count *int32 `json:"count,omitempty"`
	// VMSize - Size of agent VMs. Possible values include: 'StandardA1', 'StandardA10', 'StandardA11', 'StandardA1V2', 'StandardA2', 'StandardA2V2', 'StandardA2mV2', 'StandardA3', 'StandardA4', 'StandardA4V2', 'StandardA4mV2', 'StandardA5', 'StandardA6', 'StandardA7', 'StandardA8', 'StandardA8V2', 'StandardA8mV2', 'StandardA9', 'StandardB2ms', 'StandardB2s', 'StandardB4ms', 'StandardB8ms', 'StandardD1', 'StandardD11', 'StandardD11V2', 'StandardD11V2Promo', 'StandardD12', 'StandardD12V2', 'StandardD12V2Promo', 'StandardD13', 'StandardD13V2', 'StandardD13V2Promo', 'StandardD14', 'StandardD14V2', 'StandardD14V2Promo', 'StandardD15V2', 'StandardD16V3', 'StandardD16sV3', 'StandardD1V2', 'StandardD2', 'StandardD2V2', 'StandardD2V2Promo', 'StandardD2V3', 'StandardD2sV3', 'StandardD3', 'StandardD32V3', 'StandardD32sV3', 'StandardD3V2', 'StandardD3V2Promo', 'StandardD4', 'StandardD4V2', 'StandardD4V2Promo', 'StandardD4V3', 'StandardD4sV3', 'StandardD5V2', 'StandardD5V2Promo', 'StandardD64V3', 'StandardD64sV3', 'StandardD8V3', 'StandardD8sV3', 'StandardDS1', 'StandardDS11', 'StandardDS11V2', 'StandardDS11V2Promo', 'StandardDS12', 'StandardDS12V2', 'StandardDS12V2Promo', 'StandardDS13', 'StandardDS132V2', 'StandardDS134V2', 'StandardDS13V2', 'StandardDS13V2Promo', 'StandardDS14', 'StandardDS144V2', 'StandardDS148V2', 'StandardDS14V2', 'StandardDS14V2Promo', 'StandardDS15V2', 'StandardDS1V2', 'StandardDS2', 'StandardDS2V2', 'StandardDS2V2Promo', 'StandardDS3', 'StandardDS3V2', 'StandardDS3V2Promo', 'StandardDS4', 'StandardDS4V2', 'StandardDS4V2Promo', 'StandardDS5V2', 'StandardDS5V2Promo', 'StandardE16V3', 'StandardE16sV3', 'StandardE2V3', 'StandardE2sV3', 'StandardE3216sV3', 'StandardE328sV3', 'StandardE32V3', 'StandardE32sV3', 'StandardE4V3', 'StandardE4sV3', 'StandardE6416sV3', 'StandardE6432sV3', 'StandardE64V3', 'StandardE64sV3', 'StandardE8V3', 'StandardE8sV3', 'StandardF1', 'StandardF16', 'StandardF16s', 'StandardF16sV2', 'StandardF1s', 'StandardF2', 'StandardF2s', 'StandardF2sV2', 'StandardF32sV2', 'StandardF4', 'StandardF4s', 'StandardF4sV2', 'StandardF64sV2', 'StandardF72sV2', 'StandardF8', 'StandardF8s', 'StandardF8sV2', 'StandardG1', 'StandardG2', 'StandardG3', 'StandardG4', 'StandardG5', 'StandardGS1', 'StandardGS2', 'StandardGS3', 'StandardGS4', 'StandardGS44', 'StandardGS48', 'StandardGS5', 'StandardGS516', 'StandardGS58', 'StandardH16', 'StandardH16m', 'StandardH16mr', 'StandardH16r', 'StandardH8', 'StandardH8m', 'StandardL16s', 'StandardL32s', 'StandardL4s', 'StandardL8s', 'StandardM12832ms', 'StandardM12864ms', 'StandardM128ms', 'StandardM128s', 'StandardM6416ms', 'StandardM6432ms', 'StandardM64ms', 'StandardM64s', 'StandardNC12', 'StandardNC12sV2', 'StandardNC12sV3', 'StandardNC24', 'StandardNC24r', 'StandardNC24rsV2', 'StandardNC24rsV3', 'StandardNC24sV2', 'StandardNC24sV3', 'StandardNC6', 'StandardNC6sV2', 'StandardNC6sV3', 'StandardND12s', 'StandardND24rs', 'StandardND24s', 'StandardND6s', 'StandardNV12', 'StandardNV24', 'StandardNV6'
	//VMSize VMSizeTypes `json:"vmSize,omitempty"`
	// OsDiskSizeGB - OS Disk Size in GB to be used to specify the disk size for every machine in this master/agent pool. If you specify 0, it will apply the default osDisk size according to the vmSize specified.
	OsDiskSizeGB *int32 `json:"osDiskSizeGB,omitempty"`
	// StorageProfile - Storage profile specifies what kind of storage used. Defaults to ManagedDisks. Possible values include: 'StorageAccount', 'ManagedDisks'
	//StorageProfile StorageProfileTypes `json:"storageProfile,omitempty"`
	// VnetSubnetID - VNet SubnetID specifies the vnet's subnet identifier.
	VnetSubnetID *string `json:"vnetSubnetID,omitempty"`
	// MaxPods - Maximum number of pods that can run on a node.
	MaxPods *int32 `json:"maxPods,omitempty"`
	// OsType - OsType to be used to specify os type. Choose from Linux and Windows. Default to Linux. Possible values include: 'Linux', 'Windows'
	//OsType OSType `json:"osType,omitempty"`
}

func GetAKSClusterTemplate(templateId string, companyId string, ctx utils.Context) (cluster AKSClusterTemplate, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs(
			"AKSGetClusterTemplateModel:  Get - Got error while connecting to the database: "+err1.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAKSTemplateCollection)
	err = c.Find(bson.M{"template_id": templateId, "company_id": companyId}).One(&cluster)
	if err != nil {
		ctx.SendLogs(
			"AKSGetClusterTemplateModel:  Get - Got error while fetching from database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err
	}

	return cluster, nil
}

func GetAllAKSClusterTemplate(ctx utils.Context) (templates []AKSClusterTemplate, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("AKSGetAllClusterTemplateModel : GetAll - Got error while connecting to the database:  "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAKSTemplateCollection)
	err = c.Find(bson.M{}).All(&templates)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}
	return templates, nil
}

func AddAKSClusterTemplate(cluster AKSClusterTemplate, ctx utils.Context) (string, error) {
	_, err := GetAKSClusterTemplate(cluster.TemplateId, cluster.CompanyId, ctx)
	if err == nil {
		text := fmt.Sprintf("AKSAddClusterTemplateModel:  Add - Cluster template for project '%s' already exists in the database.", cluster.TemplateId)
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", errors.New(text)
	}

	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"AKSAddClusterTemplateModel:  Add - Got error while connecting to the database: "+err.Error(),
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
		cluster.Cloud = models.AKS
	}

	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoAKSTemplateCollection, cluster)
	if err != nil {
		ctx.SendLogs(
			"AKSAddClusterTemplateModel:  Add - Got error while inserting cluster to the database:  "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return "", err
	}

	return cluster.TemplateId, nil
}

func UpdateAKSClusterTemplate(cluster AKSClusterTemplate, ctx utils.Context) error {
	oldCluster, err := GetAKSClusterTemplate(cluster.TemplateId, cluster.CompanyId, ctx)
	if err != nil {
		text := "AKSUpdateClusterTemplateModel:  Update - Cluster '" + *cluster.Name + "' does not exist in the database: " + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	err = DeleteAKSClusterTemplate(cluster.TemplateId, cluster.CompanyId, ctx)
	if err != nil {
		ctx.SendLogs(
			"AKSUpdateClusterTemplateModel:  Update - Got error deleting cluster template: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	cluster.CreationDate = oldCluster.CreationDate
	cluster.ModificationDate = time.Now()

	_, err = AddAKSClusterTemplate(cluster, ctx)
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

func DeleteAKSClusterTemplate(templateId, companyId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"AKSDeleteClusterTemplateModel:  Delete - Got error while connecting to the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAKSTemplateCollection)
	err = c.Remove(bson.M{"template_id": templateId, "company_id": companyId})
	if err != nil {
		ctx.SendLogs(
			"AKSDeleteClusterTemplateModel:  Delete - Got error while deleting from the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}
