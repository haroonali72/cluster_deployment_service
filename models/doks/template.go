package doks

import (
	"antelope/models"
	"antelope/models/db"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/types"
	"antelope/models/utils"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type KubernetesTemplate struct {
	ID               bson.ObjectId          `json:"-" bson:"_id,omitempty"`
	TemplateId       string                 `json:"template_id" bson:"template_id"`
	Name             string                 `json:"name" bson:"name"`
	Cloud            models.Cloud           `json:"cloud" bson:"cloud"`
	CreationDate     time.Time              `json:"-" bson:"creation_date"`
	ModificationDate time.Time              `json:"-" bson:"modification_date"`
	NodePools        []*KubernetesNodePoolT `json:"node_pools" bson:"node_pools"`
	//NetworkName      string        				`json:"network_name" bson:"network_name"`
	CompanyId   string `json:"company_id" bson:"company_id"`
	IsCloudplex bool   `json:"is_cloudplex" bson:"is_cloudplex"`
}

type KubernetesNodePoolT struct {
	ID        bson.ObjectId `json:"-" bson:"_id,omitempty"`
	NodeCount int32         `json:"node_count" bson:"node_count"`
	Size      string        `json:"size,omitempty"  bson:"size"` //machine size
	//	Count     int               `json:"count,omitempty"  bson:"count"`
	//	SecurityGroupId   []string       			`json:"security_group_id" bson:"security_group_id"`
	AutoScale bool              `json:"auto_scale,omitempty"  bson:"auto_scale"`
	MinNodes  int               `json:"min_nodes,omitempty"  bson:"min_nodes"`
	MaxNodes  int               `json:"max_nodes,omitempty"  bson:"max_nodes"`
	Nodes     []*KubernetesNode `json:"nodes,omitempty"  bson:"nodes"`
}

type KubernetesTemplateMetadata struct {
	TemplateId  string `json:"template_id" bson:"template_id"`
	IsCloudplex bool   `json:"is_cloudplex" bson:"is_cloudplex"`
	PoolCount   int64  `json:"pool_count" bson:"pool_count"`
}

func GetCustomerTemplate(templateId string, ctx utils.Context) (template KubernetesTemplate, err error) {

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Customer KubernetesTemplate model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubernetesTemplate{}, err1
	}

	defer session.Close()

	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoDOKSCustomerTemplateCollection)
	err = c.Find(bson.M{"template_id": templateId}).One(&template)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubernetesTemplate{}, err
	}

	return template, nil
}

func CreateCustomerTemplate(template KubernetesTemplate, ctx utils.Context) (error, string) {

	//if template.TemplateId == "" {
	//	i := rand.Int()
	//	template.TemplateId = template.Name + strconv.Itoa(i)
	//}

	_, err := GetCustomerTemplate(template.TemplateId, ctx)
	if err == nil { //template found
		text := fmt.Sprintf("KubernetesTemplate model: Create - KubernetesTemplate '%s' already exists in the database: ", template.Name)
		beego.Error(text)
		return errors.New(text), ""
	}

	template.CreationDate = time.Now()

	s := db.GetMongoConf()

	err = db.InsertInMongo(s.MongoDOKSCustomerTemplateCollection, template)
	if err != nil {
		ctx.SendLogs("KubernetesTemplate model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err, ""
	}

	return nil, template.TemplateId
}
func UpdateCustomerTemplate(template KubernetesTemplate, ctx utils.Context) error {

	oldTemplate, err := GetCustomerTemplate(template.TemplateId, ctx)
	if err != nil {
		text := fmt.Sprintf("Template model: UpdateCustomerTemplate '%s' does not exist in the database: ", template.TemplateId)
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	err = DeleteCustomerTemplate(template.TemplateId, ctx)
	if err != nil {
		ctx.SendLogs("KubernetesTemplate model: UpdateCustomerTemplate - Got error deleting KubernetesTemplate: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	template.CreationDate = oldTemplate.CreationDate
	template.ModificationDate = time.Now()

	err, _ = CreateCustomerTemplate(template, ctx)
	if err != nil {
		ctx.SendLogs("KubernetesTemplate model: Update - Got error creating template: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func DeleteCustomerTemplate(templateId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs("KubernetesTemplate model: DeleteCustomerTempalte - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoDOKSCustomerTemplateCollection)

	err = c.Remove(bson.M{"template_id": templateId})
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}

func CheckRole(roles types.UserRole) bool {
	for _, role := range roles.Roles {
		if role.Name == models.SuperUser || role.Name == models.Admin {
			return true
		}
	}
	return false
}
func CreateTemplate(template KubernetesTemplate, ctx utils.Context) (error, string) {

	_, err := GetTemplate(template.TemplateId, template.CompanyId, ctx)
	if err == nil { //template found
		text := fmt.Sprintf("KubernetesTemplate model: Create - KubernetesTemplate '%s' already exists in the database: ", template.Name)
		beego.Error(text)
		return errors.New(text), ""
	}

	template.CreationDate = time.Now()

	//err = checkTemplateSize(template, ctx)
	//if err != nil { //cluster found
	//	ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	//	return err, ""
	//}

	s := db.GetMongoConf()
	err = db.InsertInMongo(s.MongoDOKSTemplateCollection, template)
	if err != nil {
		ctx.SendLogs("KubernetesTemplate model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err, ""
	}
	return nil, template.TemplateId
}

func GetTemplate(templateId, companyId string, ctx utils.Context) (template KubernetesTemplate, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("KubernetesTemplate model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubernetesTemplate{}, err1
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoDOKSTemplateCollection)
	err = c.Find(bson.M{"template_id": templateId, "company_id": companyId}).One(&template)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubernetesTemplate{}, err
	}
	return template, nil
}
func GetTemplates(ctx utils.Context, data rbac_athentication.List, companyId string) (templates []KubernetesTemplate, err error) {
	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		beego.Error("KubernetesTemplate model: GetAll - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoDOKSTemplateCollection)
	err = c.Find(bson.M{"template_id": bson.M{"$in": copyData}, "company_id": companyId}).All(&templates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	return templates, nil
}
func GetAllTemplate(ctx utils.Context) (templates []KubernetesTemplate, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		beego.Error("KubernetesTemplate model: GetAll - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoDOKSTemplateCollection)
	err = c.Find(bson.M{}).All(&templates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	return templates, nil
}

func UpdateTemplate(template KubernetesTemplate, ctx utils.Context) error {
	oldTemplate, err := GetTemplate(template.TemplateId, template.CompanyId, ctx)
	if err != nil {
		text := fmt.Sprintf("KubernetesTemplate model: Update - KubernetesTemplate '%s' does not exist in the database: ", template.TemplateId)
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	err = DeleteTemplate(template.TemplateId, template.CompanyId, ctx)
	if err != nil {

		ctx.SendLogs("KubernetesTemplate model: Update - Got error deleting template: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return err
	}

	template.CreationDate = oldTemplate.CreationDate
	template.ModificationDate = time.Now()

	err, _ = CreateTemplate(template, ctx)
	if err != nil {
		ctx.SendLogs("KubernetesTemplate model: Update - Got error creating template: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}

func DeleteTemplate(templateId, companyId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs("KubernetesTemplate model: Delete - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoDOKSTemplateCollection)
	err = c.Remove(bson.M{"template_id": templateId, "company_id": companyId})
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}

func GetAllCustomerTemplates(ctx utils.Context) (templates []KubernetesTemplate, err error) {

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		beego.Error("KubernetesTemplate model: GetAll - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoDOKSCustomerTemplateCollection)
	err = c.Find(bson.M{}).All(&templates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	return templates, nil
}

func GetTemplatesMetadata(ctx utils.Context, data rbac_athentication.List, companyId string) (metadatat []KubernetesTemplateMetadata, err error) {

	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		beego.Error("KubernetesTemplate model: Get meta data - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()

	var templates []KubernetesTemplate

	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoDOKSTemplateCollection)
	err = c.Find(bson.M{"template_id": bson.M{"$in": copyData}, "company_id": companyId}).All(&templates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return nil, err
	}

	templatemetadata := make([]KubernetesTemplateMetadata, len(templates))

	for i, template := range templates {
		templatemetadata[i].TemplateId = templates[i].TemplateId

		if template.IsCloudplex {
			templatemetadata[i].IsCloudplex = true
		} else {
			templatemetadata[i].IsCloudplex = false
		}

		for range template.NodePools {

			templatemetadata[i].PoolCount++
		}
	}

	return templatemetadata, nil
}

func GetCustomerTemplatesMetadata(ctx utils.Context, data rbac_athentication.List, companyId string) (metadatat []KubernetesTemplateMetadata, err error) {

	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		beego.Error("KubernetesTemplate model: Get meta data - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()

	var customerTemplates []KubernetesTemplate

	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoDOKSCustomerTemplateCollection)
	err = c.Find(bson.M{}).All(&customerTemplates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	templatemetadata := make([]KubernetesTemplateMetadata, len(customerTemplates))

	for i, template := range customerTemplates {
		templatemetadata[i].TemplateId = customerTemplates[i].TemplateId
		if template.IsCloudplex {
			templatemetadata[i].IsCloudplex = true
		} else {
			templatemetadata[i].IsCloudplex = false
		}
		for range template.NodePools {

			templatemetadata[i].PoolCount++
		}
	}

	return templatemetadata, nil
}
