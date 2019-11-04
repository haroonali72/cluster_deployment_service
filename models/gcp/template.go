package gcp

import (
	"antelope/models"
	"antelope/models/db"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"strconv"
	"time"
)

type Template struct {
	ID               bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	TemplateId       string        `json:"template_id" bson:"template_id"`
	Name             string        `json:"name" bson:"name"`
	Status           string        `json:"status" bson:"status"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools        []*NodePoolT  `json:"node_pools" bson:"node_pools"`
	NetworkName      string        `json:"network_name" bson:"network_name"`
	CompanyId        string        `json:"company_id" bson:"company_id"`
}

type NodePoolT struct {
	ID                  bson.ObjectId `json:"-" bson:"_id,omitempty"`
	Name                string        `json:"name" bson:"name"`
	PoolId              string        `json:"pool_id" bson:"pool_id"`
	NodeCount           int64         `json:"node_count" bson:"node_count"`
	MachineType         string        `json:"machine_type" bson:"machine_type"`
	Image               Image         `json:"image" bson:"image"`
	Volume              Volume        `json:"volume" bson:"volume"`
	RootVolume          Volume        `json:"root_volume" bson:"root_volume"`
	EnableVolume        bool          `json:"is_external" bson:"is_external"`
	PoolSubnet          string        `json:"subnet_id" bson:"subnet_id"`
	PoolRole            string        `json:"pool_role" bson:"pool_role"`
	ServiceAccountEmail string        `json:"service_account_email" bson:"service_account_email"`
	Nodes               []*Node       `json:"nodes" bson:"nodes"`
	EnableScaling       bool          `json:"enable_scaling" bson:"enable_scaling"`
	Scaling             AutoScaling   `json:"auto_scaling" bson:"auto_scaling"`
}

func CreateTemplate(template Template, ctx utils.Context) (error, string) {
	_, err := GetTemplate(template.TemplateId, template.CompanyId, ctx)
	if err == nil { //template found
		text := fmt.Sprintf("Template model: Create - Template '%s' already exists in the database: ", template.Name)
		ctx.SendLogs("gcpTemplateModel :"+text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(text)
		return errors.New(text), ""
	}

	if template.TemplateId == "" {
		i := rand.Int()
		template.TemplateId = template.Name + strconv.Itoa(i)
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
func RegisterCustomerTemplate(templates []Template, companyId string, ctx utils.Context) error {

	for _, template := range templates {
		template.CompanyId = companyId
		template.CreationDate = time.Now()
		if template.TemplateId == "" {
			i := rand.Int()
			template.TemplateId = template.Name + strconv.Itoa(i)
		}
	}
	var inter []interface{}
	for _, template := range templates {
		inter = append(inter, template)
	}
	s := db.GetMongoConf()
	err := db.InsertManyInMongo(s.MongoAwsTemplateCollection, inter)
	if err != nil {
		ctx.SendLogs("Template model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}
func GetCustomerTemplate(ctx utils.Context) (template []Template, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Template model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []Template{}, err1
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoGcpCustomerTemplateCollection)
	err = c.Find(bson.M{}).All(&template)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []Template{}, err
	}

	return template, nil
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
		beego.Error(err.Error())
		return Template{}, err
	}

	return template, nil
}
func GetTemplates(ctx utils.Context, data rbac_athentication.List) (templates []Template, err error) {
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
	err = c.Find(bson.M{"template_id": bson.M{"$in": copyData}}).All(&templates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return nil, err
	}

	return templates, nil
}
func GetAllTemplate(ctx utils.Context) (templates []Template, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("GcpTemplateModel : error connecting to database "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Template model: GetAll - Got error while connecting to the database: ", err1)
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
		beego.Error(text, err)
		return errors.New(text)
	}

	err = DeleteTemplate(template.TemplateId, ctx)
	if err != nil {
		ctx.SendLogs("GcpTemplateModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Template model: Update - Got error deleting template: ", err)
		return err
	}

	template.CreationDate = oldTemplate.CreationDate
	template.ModificationDate = time.Now()

	err, _ = CreateTemplate(template, ctx)
	if err != nil {
		ctx.SendLogs("GcpTemplateModel :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Template model: Update - Got error creating template: ", err)
		return err
	}

	return nil
}

func DeleteTemplate(templateId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs("GcpTemplateModel : erro with connecting database "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("Template model: Delete - Got error while connecting to the database: ", err)
		return err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGcpTemplateCollection)
	err = c.Remove(bson.M{"template_id": templateId})
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}
