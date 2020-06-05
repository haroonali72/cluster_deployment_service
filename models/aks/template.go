package aks

import (
	"antelope/models"
	"antelope/models/db"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"errors"
	"fmt"
	aks "github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-02-01/containerservice"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type AKSClusterTemplate struct {
	ID                bson.ObjectId             `json:"-" bson:"_id,omitempty"`
	TemplateId        string                    `json:"Template_id" bson:"Template_id"`
	Cloud             models.Cloud              `json:"cloud" bson:"cloud"`
	CreationDate      time.Time                 `json:"-" bson:"creation_date"`
	ModificationDate  time.Time                 `json:"-" bson:"modification_date"`
	CompanyId         string                    `json:"company_id" bson:"company_id"`
	Status            string                    `json:"status,omitempty" bson:"status,omitempty"`
	ResourceGoup      string                    `json:"resource_group" bson:"resource_group" validate:"required"`
	ClusterProperties ManagedClusterPropertiesT `json:"properties" bson:"properties" validate:"required"`
	ResourceID        *string                   `json:"cluster_id,omitempty" bson:"cluster_id,omitempty"`
	Name              *string                   `json:"name,omitempty" bson:"name,omitempty"`
	Type              *string                   `json:"type,omitempty" bson:"type,omitempty"`
	Location          *string                   `json:"location,omitempty" bson:"location,omitempty"`
	Tags              map[string]*string        `json:"tags" bson:"tags"`
	IsCloudplex       bool                      `json:"is_cloudplex" bson:"is_cloudplex"`
}

type ManagedClusterPropertiesT struct {
	AgentPoolProfiles      []ManagedClusterAgentPoolProfileT     `json:"agent_pool,omitempty" bson:"agent_pool,omitempty"`
	APIServerAccessProfile ManagedClusterAPIServerAccessProfileT `json:"api_server_access_profile,omitempty" bson:"api_server_access_profile,omitempty"`
	EnableRBAC             bool                                  `json:"enable_rbac,omitempty" bson:"enable_rbac,omitempty"`
}

// ManagedClusterAPIServerAccessProfile access profile for managed cluster API server.
type ManagedClusterAPIServerAccessProfileT struct {
	AuthorizedIPRanges   []string `json:"authorized_ip_ranges,omitempty"`
	EnablePrivateCluster bool     `json:"enable_private_cluster,omitempty" bson:"enable_private_cluster,omitempty"`
}

// ManagedClusterAgentPoolProfile profile for the container service agent pool.
type ManagedClusterAgentPoolProfileT struct {
	Name              string          `json:"name,omitempty" bson:"name,omitempty" validate:"required"`
	Count             int32           `json:"count,omitempty" bson:"count,omitempty" validate:"required"`
	VMSize            aks.VMSizeTypes `json:"vm_size,omitempty" bson:"vm_size,omitempty" validate:"required"`
	OsDiskSizeGB      int32           `json:"os_disk_size_gb,omitempty" bson:"os_disk_size_gb,omitempty"`
	VnetSubnetID      string          `json:"subnet_id" bson:"subnet_id"`
	MaxPods           int32           `json:"max_pods,omitempty" bson:"max_pods,omitempty"`
	OsType            aks.OSType      `json:"os_type,omitempty" bson:"os_type,omitempty"`
	MaxCount          int32           `json:"max_count,omitempty" bson:"max_count,omitempty"`
	MinCount          int32           `json:"min_count,omitempty" bson:"min_count,omitempty"`
	EnableAutoScaling bool            `json:"enable_auto_scaling,omitempty" bson:"enable_auto_scaling,omitempty"`
}

type TemplateMetadata struct {
	TemplateId  string `json:"template_id" bson:"template_id"`
	IsCloudplex bool   `json:"is_cloudplex" bson:"is_cloudplex"`
	PoolCount   int64  `json:"pool_count" bson:"pool_count"`
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

func GetAKSCustomerTemplate(templateId string, ctx utils.Context) (template AKSClusterTemplate, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("AKSClusterTemplate model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return AKSClusterTemplate{}, err1
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoAKSClusterCollection)
	err = c.Find(bson.M{"template_id": templateId}).One(&template)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return AKSClusterTemplate{}, err
	}

	return template, nil
}

func CreateAKSCustomerTemplate(template AKSClusterTemplate, ctx utils.Context) (error, string) {
	_, err := GetAKSCustomerTemplate(template.TemplateId, ctx)
	if err == nil { //template found
		text := fmt.Sprintf("AKSClusterTemplate model: Create - Template '%s' already exists in the database: ", template.Name)
		beego.Error(text)
		return errors.New(text), ""
	}

	template.CreationDate = time.Now()

	s := db.GetMongoConf()
	err = db.InsertInMongo(s.MongoAKSClusterCollection, template)
	if err != nil {
		ctx.SendLogs("AKSClusterTemplate model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err, ""
	}

	return nil, template.TemplateId
}

func UpdateAKSCustomerTemplate(template AKSClusterTemplate, ctx utils.Context) error {
	oldTemplate, err := GetAKSCustomerTemplate(template.TemplateId, ctx)
	if err != nil {
		text := fmt.Sprintf("AKSClusterTemplate model: UpdateCustomerTemplate '%s' does not exist in the database: ", template.TemplateId)
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	err = DeleteAKSCustomerTemplate(template.TemplateId, ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterTemplate model: UpdateCustomerTemplate - Got error deleting template: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	template.CreationDate = oldTemplate.CreationDate
	template.ModificationDate = time.Now()

	err, _ = CreateAKSCustomerTemplate(template, ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterTemplate model: Update - Got error creating template: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}

func DeleteAKSCustomerTemplate(templateId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterTemplate model: DeleteCustomerTemplate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoAKSCustomerTemplateCollection)
	err = c.Remove(bson.M{"template_id": templateId})
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}

func GetAllAKSCustomerTemplates(ctx utils.Context) (templates []AKSClusterTemplate, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		beego.Error("AKSClusterTemplate model: GetAll - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoAKSCustomerTemplateCollection)
	err = c.Find(bson.M{}).All(&templates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	return templates, nil
}

func GetAKSTemplatesMetadata(ctx utils.Context, data rbacAuthentication.List, companyId string) (metadata []TemplateMetadata, err error) {
	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		beego.Error("AKSClusterTemplate model: Get meta data - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()

	var templates []AKSClusterTemplate

	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoAKSTemplateCollection)
	err = c.Find(bson.M{"template_id": bson.M{"$in": copyData}, "company_id": companyId}).All(&templates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return nil, err
	}

	templateMetadata := make([]TemplateMetadata, len(templates))

	for i, template := range templates {
		templateMetadata[i].TemplateId = templates[i].TemplateId

		if template.IsCloudplex {
			templateMetadata[i].IsCloudplex = true
		} else {
			templateMetadata[i].IsCloudplex = false
		}

		for range template.ClusterProperties.AgentPoolProfiles {
			templateMetadata[i].PoolCount++
		}
	}

	return templateMetadata, nil
}

func GetAKSCustomerTemplatesMetadata(ctx utils.Context, data rbacAuthentication.List, companyId string) (metadata []TemplateMetadata, err error) {
	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		beego.Error("AKSClusterTemplate model: Get meta data - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()

	var customerTemplates []AKSClusterTemplate

	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoAKSCustomerTemplateCollection)
	err = c.Find(bson.M{}).All(&customerTemplates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	templateMetadata := make([]TemplateMetadata, len(customerTemplates))

	for i, template := range customerTemplates {
		templateMetadata[i].TemplateId = customerTemplates[i].TemplateId

		if template.IsCloudplex {
			templateMetadata[i].IsCloudplex = true
		} else {
			templateMetadata[i].IsCloudplex = false
		}

		for range template.ClusterProperties.AgentPoolProfiles {
			templateMetadata[i].PoolCount++
		}
	}

	return templateMetadata, nil
}

func GetAKSDefault(ctx utils.Context) (template AKSClusterTemplate, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Template model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Template model: Get - Got error while connecting to the database: ", err1)
		return AKSClusterTemplate{}, err1
	}
	defer session.Close()
	conf := db.GetMongoConf()

	c := session.DB(conf.MongoDb).C(conf.MongoDefaultTemplateCollection)
	err = c.Find(bson.M{"cloud": "aks"}).One(&template)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return AKSClusterTemplate{}, err
	}
	return template, nil
}
