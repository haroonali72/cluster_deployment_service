package azure

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
	ResourceGroup    string        `json:"resource_group" bson:"resource_group"`
	CompanyId        string        `json:"company_id" bson:"company_id"`
	IsCloudplex      bool          `json:"is_cloudplex" bson:"is_cloudplex"`
}

type NodePoolT struct {
	ID                 bson.ObjectId      `json:"_id" bson:"_id,omitempty"`
	NodeCount          int64              `json:"node_count" bson:"node_count"`
	MachineType        string             `json:"machine_type" bson:"machine_type"`
	Image              ImageReference     `json:"image" bson:"image"`
	Volume             Volume             `json:"volume" bson:"volume"`
	EnableVolume       bool               `json:"is_external" bson:"is_external"`
	PoolSubnet         string             `json:"subnet_id" bson:"subnet_id"`
	PoolSecurityGroups []*string          `json:"security_group_id" bson:"security_group_id"`
	Nodes              []*VM              `json:"nodes" bson:"nodes"`
	PoolRole           string             `json:"pool_role" bson:"pool_role"`
	AdminUser          string             `json:"user_name" bson:"user_name",omitempty"`
	BootDiagnostics    DiagnosticsProfile `json:"boot_diagnostics" bson:"boot_diagnostics"`
	OsDisk             models.OsDiskType  `json:"os_disk_type" bson:"os_disk_type"`
	EnablePublicIP     bool               `json:"enable_public_ip" bson:"enable_public_ip"`
}

type TemplateMetadata struct {
	TemplateId  string `json:"template_id" bson:"template_id"`
	IsCloudplex bool   `json:"is_cloudplex" bson:"is_cloudplex"`
	PoolCount   int    `json:"pool_count" bson:"pool_count"`
}

func checkTemplateSize(cluster Template) error {
	for _, pools := range cluster.NodePools {
		if pools.NodeCount > 3 {
			return errors.New("Nodepool can't have more than 3 nodes")
		}
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
func GetCustomerTemplate(templateId string, ctx utils.Context) (template Template, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Template model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Template{}, err1
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoAzureCustomerTemplateCollection)
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
	err = db.InsertInMongo(s.MongoAzureCustomerTemplateCollection, template)
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
		ctx.SendLogs("Template model: DeleteCustomerTempalte - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoAzureCustomerTemplateCollection)
	err = c.Remove(bson.M{"template_id": templateId})
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func CreateTemplate(template Template, ctx utils.Context) error {

	_, err := GetTemplate(template.TemplateId, template.CompanyId, ctx)
	if err == nil { //template found
		text := fmt.Sprintf("Template model: Create - Template '%s' already exists in the database: ", template.Name)
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	template.CreationDate = time.Now()

	//err = checkTemplateSize(template)
	//if err != nil { //cluster found
	//
	//	ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	//	return err, ""
	//}
	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoAzureTemplateCollection, template)
	if err != nil {

		ctx.SendLogs("Template model: Create - Got error inserting template to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}

func GetTemplate(templateId, companyId string, ctx utils.Context) (template Template, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Template model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Template{}, err1
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAzureTemplateCollection)
	err = c.Find(bson.M{"template_id": templateId, "company_id": companyId}).One(&template)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
	c := session.DB(s.MongoDb).C(s.MongoAzureTemplateCollection)
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
		ctx.SendLogs("Template model: GetAll - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err1
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAzureTemplateCollection)
	err = c.Find(bson.M{}).All(&templates)
	if err != nil {
		ctx.SendLogs("Template model: GetAll - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	return templates, nil
}

func UpdateTemplate(template Template, ctx utils.Context) error {
	oldTemplate, err := GetTemplate(template.TemplateId, template.CompanyId, ctx)
	if err != nil {
		text := fmt.Sprintf("Template model: Update - Template '%s' does not exist in the database: ", template.Name)
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	err = DeleteTemplate(template.TemplateId, template.CompanyId, ctx)
	if err != nil {
		beego.Error("Template model: Update - Got error deleting template: ", err)
		return err
	}

	template.CreationDate = oldTemplate.CreationDate
	template.ModificationDate = time.Now()

	err = CreateTemplate(template, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}

func DeleteTemplate(templateId, companyId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAzureTemplateCollection)
	err = c.Remove(bson.M{"template_id": templateId, "company_id": companyId})
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
	c := session.DB(s.MongoDb).C(s.MongoAzureCustomerTemplateCollection)
	err = c.Find(bson.M{}).All(&templates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	return templates, nil
}

func GetAllTemplateMetadata(ctx utils.Context, data rbac_athentication.List, companyId string) []TemplateMetadata {
	var templateList []Template
	templates, err := GetTemplates(ctx, data, companyId)
	if err != nil {
		beego.Error(err.Error())
		return nil
	}

	for _, template := range templates {
		templateList = append(templateList, template)
	}
	templateMetadata := make([]TemplateMetadata, len(templateList))
	for i, template := range templateList {
		templateMetadata[i].TemplateId = template.TemplateId
		poolCount := 0

		if template.IsCloudplex {
			templateMetadata[i].IsCloudplex = true
		} else {
			templateMetadata[i].IsCloudplex = false
		}

		for i := 0; i < len(template.NodePools); i++ {
			poolCount++
		}
		templateMetadata[i].PoolCount = poolCount
	}

	return templateMetadata
}

func GetAllCustomerTemplateMetadata(ctx utils.Context) []TemplateMetadata {
	var templateList []Template
	templates, err := GetAllCustomerTemplates(ctx)
	if err != nil {
		beego.Error(err.Error())
		return nil
	}

	for _, template := range templates {
		templateList = append(templateList, template)
	}
	templateMetadata := make([]TemplateMetadata, len(templateList))
	for i, template := range templateList {
		templateMetadata[i].TemplateId = template.TemplateId
		poolCount := 0

		if template.IsCloudplex {
			templateMetadata[i].IsCloudplex = true
		} else {
			templateMetadata[i].IsCloudplex = false
		}

		for i := 0; i < len(template.NodePools); i++ {
			poolCount++
		}
		templateMetadata[i].PoolCount = poolCount
	}

	return templateMetadata
}

func GetAzureDefault(ctx utils.Context) (template Template, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Template model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Template model: Get - Got error while connecting to the database: ", err1)
		return Template{}, err1
	}
	defer session.Close()
	conf := db.GetMongoConf()

	c := session.DB(conf.MongoDb).C(conf.MongoDefaultTemplateCollection)
	err = c.Find(bson.M{"cloud": "azure"}).One(&template)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return Template{}, err
	}
	return template, nil
}
