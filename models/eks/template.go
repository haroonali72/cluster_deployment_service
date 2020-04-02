package eks

import (
	"antelope/models"
	"antelope/models/db"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type EKSClusterTemplate struct {
	ID               bson.ObjectId `json:"-" bson:"_id,omitempty"`
	TemplateId       string        `json:"template_id" bson:"template_id"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	Status           string        `json:"status" bson:"status"`
	CompanyId        string        `json:"company_id" bson:"company_id"`
	IsCloudplex      bool          `json:"is_cloudplex" bson:"is_cloudplex"`

	ClientRequestToken *string             `json:"client_request_token,omitempty" bson:"client_request_token,omitempty"`
	EncryptionConfig   []*EncryptionConfig `json:"encryption_config,omitempty" bson:"encryption_config,omitempty"`
	Logging            *Logging            `json:"logging,omitempty" bson:"logging,omitempty"`
	Name               string              `json:"name" bson:"name"`
	ResourcesVpcConfig VpcConfigRequest    `json:"resources_vpc_config" bson:"resources_vpc_config"`
	RoleArn            string              `json:"role_arn" bson:"role_arn"`
	Tags               map[string]*string  `json:"tags,omitempty" bson:"tags,omitempty"`
	Version            *string             `json:"version,omitempty" bson:"version,omitempty"`
	Nodegroups         []*Nodegroup        `json:"node_groups" bson:"node_groups"`
}

type TemplateMetadata struct {
	TemplateId  string `json:"template_id" bson:"template_id"`
	IsCloudplex bool   `json:"is_cloudplex" bson:"is_cloudplex"`
	PoolCount   int64  `json:"pool_count" bson:"pool_count"`
}

func GetEKSClusterTemplate(templateId string, companyId string, ctx utils.Context) (cluster EKSClusterTemplate, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs(
			"EKSGetClusterTemplateModel:  Get - Got error while connecting to the database: "+err1.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoEKSTemplateCollection)
	err = c.Find(bson.M{"template_id": templateId, "company_id": companyId}).One(&cluster)
	if err != nil {
		ctx.SendLogs(
			"EKSGetClusterTemplateModel:  Get - Got error while fetching from database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return cluster, err
	}

	return cluster, nil
}

func GetAllEKSClusterTemplate(ctx utils.Context) (templates []EKSClusterTemplate, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("EKSGetAllClusterTemplateModel : GetAll - Got error while connecting to the database:  "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoEKSTemplateCollection)
	err = c.Find(bson.M{}).All(&templates)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}
	return templates, nil
}

func AddEKSClusterTemplate(cluster EKSClusterTemplate, ctx utils.Context) (string, error) {
	_, err := GetEKSClusterTemplate(cluster.TemplateId, cluster.CompanyId, ctx)
	if err == nil {
		text := fmt.Sprintf("EKSAddClusterTemplateModel:  Add - Cluster template for project '%s' already exists in the database.", cluster.TemplateId)
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", errors.New(text)
	}

	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"EKSAddClusterTemplateModel:  Add - Got error while connecting to the database: "+err.Error(),
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
		cluster.Cloud = models.EKS
	}

	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoEKSTemplateCollection, cluster)
	if err != nil {
		ctx.SendLogs(
			"EKSAddClusterTemplateModel:  Add - Got error while inserting cluster to the database:  "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return "", err
	}

	return cluster.TemplateId, nil
}

func UpdateEKSClusterTemplate(cluster EKSClusterTemplate, ctx utils.Context) error {
	oldCluster, err := GetEKSClusterTemplate(cluster.TemplateId, cluster.CompanyId, ctx)
	if err != nil {
		text := "EKSUpdateClusterTemplateModel:  Update - Cluster '" + cluster.Name + "' does not exist in the database: " + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	err = DeleteEKSClusterTemplate(cluster.TemplateId, cluster.CompanyId, ctx)
	if err != nil {
		ctx.SendLogs(
			"EKSUpdateClusterTemplateModel:  Update - Got error deleting cluster template: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	cluster.CreationDate = oldCluster.CreationDate
	cluster.ModificationDate = time.Now()

	_, err = AddEKSClusterTemplate(cluster, ctx)
	if err != nil {
		ctx.SendLogs(
			"EKSUpdateClusterTemplateModel:  Update - Got error creating cluster template: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func DeleteEKSClusterTemplate(templateId, companyId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"EKSDeleteClusterTemplateModel:  Delete - Got error while connecting to the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoEKSTemplateCollection)
	err = c.Remove(bson.M{"template_id": templateId, "company_id": companyId})
	if err != nil {
		ctx.SendLogs(
			"EKSDeleteClusterTemplateModel:  Delete - Got error while deleting from the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}

func GetEKSCustomerTemplate(templateId string, ctx utils.Context) (template EKSClusterTemplate, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("EKSClusterTemplate model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return EKSClusterTemplate{}, err1
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoEKSClusterCollection)
	err = c.Find(bson.M{"template_id": templateId}).One(&template)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return EKSClusterTemplate{}, err
	}

	return template, nil
}

func CreateEKSCustomerTemplate(template EKSClusterTemplate, ctx utils.Context) (error, string) {
	_, err := GetEKSCustomerTemplate(template.TemplateId, ctx)
	if err == nil { //template found
		text := fmt.Sprintf("EKSClusterTemplate model: Create - Template '%s' already exists in the database: ", template.Name)
		beego.Error(text)
		return errors.New(text), ""
	}

	template.CreationDate = time.Now()

	s := db.GetMongoConf()
	err = db.InsertInMongo(s.MongoEKSClusterCollection, template)
	if err != nil {
		ctx.SendLogs("EKSClusterTemplate model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err, ""
	}

	return nil, template.TemplateId
}

func UpdateEKSCustomerTemplate(template EKSClusterTemplate, ctx utils.Context) error {
	oldTemplate, err := GetEKSCustomerTemplate(template.TemplateId, ctx)
	if err != nil {
		text := fmt.Sprintf("EKSClusterTemplate model: UpdateCustomerTemplate '%s' does not exist in the database: ", template.TemplateId)
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	err = DeleteEKSCustomerTemplate(template.TemplateId, ctx)
	if err != nil {
		ctx.SendLogs("EKSClusterTemplate model: UpdateCustomerTemplate - Got error deleting template: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	template.CreationDate = oldTemplate.CreationDate
	template.ModificationDate = time.Now()

	err, _ = CreateEKSCustomerTemplate(template, ctx)
	if err != nil {
		ctx.SendLogs("EKSClusterTemplate model: Update - Got error creating template: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}

func DeleteEKSCustomerTemplate(templateId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs("EKSClusterTemplate model: DeleteCustomerTemplate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoEKSCustomerTemplateCollection)
	err = c.Remove(bson.M{"template_id": templateId})
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}

func GetAllEKSCustomerTemplates(ctx utils.Context) (templates []EKSClusterTemplate, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		beego.Error("EKSClusterTemplate model: GetAll - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoEKSCustomerTemplateCollection)
	err = c.Find(bson.M{}).All(&templates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	return templates, nil
}

func GetEKSTemplatesMetadata(ctx utils.Context, data rbacAuthentication.List, companyId string) (metadata []TemplateMetadata, err error) {
	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		beego.Error("EKSClusterTemplate model: Get meta data - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()

	var templates []EKSClusterTemplate

	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoEKSTemplateCollection)
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

		for range template.Nodegroups {
			templateMetadata[i].PoolCount++
		}
	}

	return templateMetadata, nil
}

func GetEKSCustomerTemplatesMetadata(ctx utils.Context, data rbacAuthentication.List, companyId string) (metadata []TemplateMetadata, err error) {
	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		beego.Error("EKSClusterTemplate model: Get meta data - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()

	var customerTemplates []EKSClusterTemplate

	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoEKSCustomerTemplateCollection)
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

		for range template.Nodegroups {
			templateMetadata[i].PoolCount++
		}
	}

	return templateMetadata, nil
}
