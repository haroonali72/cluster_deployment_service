package gcp

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

type Template struct {
	ID               bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	TemplateId       string        `json:"template_id" bson:"template_id"`
	Name             string        `json:"name" bson:"name"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools        []*NodePoolT  `json:"node_pools" bson:"node_pools"`
	NetworkName      string        `json:"network_name" bson:"network_name"`
	VPCName          string        `json:"vpc_name" bson:"vpc_name"`
	CompanyId        string        `json:"company_id" bson:"company_id"`
	IsCloudplex      bool          `json:"is_cloudplex" bson:"is_cloudplex"`
}

type NodePoolT struct {
	ID             bson.ObjectId `json:"-" bson:"_id,omitempty"`
	PoolId         string        `json:"pool_id" bson:"pool_id"`
	NodeCount      int64         `json:"node_count" bson:"node_count"`
	MachineType    string        `json:"machine_type" bson:"machine_type"`
	Image          Image         `json:"image" bson:"image"`
	Volume         Volume        `json:"volume" bson:"volume"`
	RootVolume     Volume        `json:"root_volume" bson:"root_volume"`
	EnableVolume   bool          `json:"is_external" bson:"is_external"`
	PoolSubnet     string        `json:"subnet_id" bson:"subnet_id"`
	PoolRole       string        `json:"pool_role" bson:"pool_role"`
	Nodes          []*Node       `json:"nodes" bson:"nodes"`
	EnableScaling  bool          `json:"enable_scaling" bson:"enable_scaling"`
	Scaling        AutoScaling   `json:"auto_scaling" bson:"auto_scaling"`
	EnablePublicIP bool          `json:"enable_public_ip" bson:"enable_public_ip"`
}
type TemplateMetadata struct {
	TemplateId  string `json:"template_id" bson:"template_id"`
	IsCloudplex bool   `json:"is_cloudplex" bson:"is_cloudplex"`
	PoolCount   int64  `json:"pool_count" bson:"pool_count"`
}

func CheckRole(roles types.UserRole) bool {
	for _, role := range roles.Roles {
		if role.Name == models.SuperUser || role.Name == models.Admin {
			return true
		}
	}
	return false
}

func GetCustomerTemplate(templateId string, ctx utils.Context) (template Template, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Template model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Template{}, err1
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoGcpCustomerTemplateCollection)
	err = c.Find(bson.M{"template_id": templateId}).One(&template)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Template{}, err
	}

	return template, nil
}
func CreateCustomerTemplate(template Template, ctx utils.Context) (error, string) {

	_, err := GetCustomerTemplate(template.TemplateId, ctx)
	if err == nil { //template found
		text := fmt.Sprintf("Template model: Create - Template '%s' already exists in the database: ", template.Name)
		beego.Error(text)
		return errors.New(text), ""
	}

	template.CreationDate = time.Now()

	s := db.GetMongoConf()
	err = db.InsertInMongo(s.MongoGcpCustomerTemplateCollection, template)
	if err != nil {
		ctx.SendLogs("Template model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err, ""
	}

	return nil, template.TemplateId
}
func UpdateCustomerTemplate(template Template, ctx utils.Context) error {

	oldTemplate, err := GetCustomerTemplate(template.TemplateId, ctx)
	if err != nil {
		text := fmt.Sprintf("Template model: UpdateCustomerTemplate '%s' does not exist in the database: ", template.TemplateId)
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	err = DeleteCustomerTemplate(template.TemplateId, ctx)
	if err != nil {
		ctx.SendLogs("Template model: UpdateCustomerTemplate - Got error deleting template: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	template.CreationDate = oldTemplate.CreationDate
	template.ModificationDate = time.Now()

	err, _ = CreateCustomerTemplate(template, ctx)
	if err != nil {
		ctx.SendLogs("Template model: Update - Got error creating template: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func DeleteCustomerTemplate(templateId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs("Template model: DeleteCustomerTemplate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoGcpCustomerTemplateCollection)
	err = c.Remove(bson.M{"template_id": templateId})
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func GetAllCustomerTemplates(ctx utils.Context) (templates []Template, err error) {

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		beego.Error("Template model: GetAll - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoGcpCustomerTemplateCollection)
	err = c.Find(bson.M{}).All(&templates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	return templates, nil
}

func CreateTemplate(template Template, ctx utils.Context) (error, string) {

	_, err := GetTemplate(template.TemplateId, template.CompanyId, ctx)
	if err == nil { //template found
		text := fmt.Sprintf("Template model: Create - Template '%s' already exists in the database: ", template.Name)
		ctx.SendLogs("gcpTemplateModel :"+text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text), ""
	}

	template.CreationDate = time.Now()
	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoGcpTemplateCollection, template)
	if err != nil {
		beego.Error("Template model: Create - Got error inserting template to the database: ", err)
		return err, ""
	}
	return nil, template.TemplateId
}
func GetTemplate(templateId, companyId string, ctx utils.Context) (template Template, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("GcpTemplateModel :"+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Template model: Get - Got error while connecting to the database: ", err1)
		return Template{}, err1
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGcpTemplateCollection)
	err = c.Find(bson.M{"template_id": templateId, "company_id": companyId}).One(&template)
	if err != nil {
		ctx.SendLogs("GcpTemplateModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Template{}, err
	}

	return template, nil
}
func GetTemplates(ctx utils.Context, data rbac_athentication.List, companyId string) (templates []Template, err error) {
	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		beego.Error("Template model: GetAll - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoGcpTemplateCollection)
	err = c.Find(bson.M{"template_id": bson.M{"$in": copyData}, "company_id": companyId}).All(&templates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return nil, err
	}
	return templates, nil
}
func GetAllTemplate(ctx utils.Context) (templates []Template, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("GcpTemplateModel : GetAll - Got error while connecting to the database:  "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGcpTemplateCollection)
	err = c.Find(bson.M{}).All(&templates)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}
	return templates, nil
}
func UpdateTemplate(template Template, ctx utils.Context) error {
	oldTemplate, err := GetTemplate(template.TemplateId, template.CompanyId, ctx)
	if err != nil {
		text := fmt.Sprintf("Template model: Update - Template '%s' does not exist in the database: ", template.Name)
		ctx.SendLogs("GcpTemplateModel "+text+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	err = DeleteTemplate(template.TemplateId, template.CompanyId, ctx)
	if err != nil {
		ctx.SendLogs("GcpTemplateModel : Update - Got error deleting template: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	template.CreationDate = oldTemplate.CreationDate
	template.ModificationDate = time.Now()

	err, _ = CreateTemplate(template, ctx)
	if err != nil {
		ctx.SendLogs("GcpTemplateModel :Update - Got error creating template:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func DeleteTemplate(templateId, companyId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs("GcpTemplateModel : Delete - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGcpTemplateCollection)
	err = c.Remove(bson.M{"template_id": templateId, "company_id": companyId})
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	return nil
}

func GetTemplatesMetadata(ctx utils.Context, data rbac_athentication.List, companyId string) (metadatat []TemplateMetadata, err error) {

	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		beego.Error("Template model: Get meta data - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()

	var templates []Template

	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoGcpTemplateCollection)
	err = c.Find(bson.M{"template_id": bson.M{"$in": copyData}, "company_id": companyId}).All(&templates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return nil, err
	}

	templatemetadata := make([]TemplateMetadata, len(templates))

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

func GetCustomerTemplatesMetadata(ctx utils.Context, data rbac_athentication.List, companyId string) (metadatat []TemplateMetadata, err error) {

	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		beego.Error("Template model: Get meta data - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()

	var customerTemplates []Template

	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoGcpCustomerTemplateCollection)
	err = c.Find(bson.M{}).All(&customerTemplates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	templatemetadata := make([]TemplateMetadata, len(customerTemplates))

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
