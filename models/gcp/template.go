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
	ResourceGroup    string        `json:"resource_group" bson:"resource_group"`
}

type NodePoolT struct {
	ID          bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name        string        `json:"name" bson:"name"`
	NodeCount   int64         `json:"node_count" bson:"node_count"`
	MachineType string        `json:"machine_type" bson:"machine_type"`
	Image       Image         `json:"image" bson:"image"`
	PoolSubnet  string        `json:"subnet_id" bson:"subnet_id"`
	KeyInfo     utils.Key     `json:"key_info" bson:"key_info"`
}

func CreateTemplate(template Template, ctx utils.Context) (error, string) {
	_, err := GetTemplate(template.TemplateId, ctx)
	if err == nil { //template found
		text := fmt.Sprintf("Template model: Create - Template '%s' already exists in the database: ", template.Name)
		ctx.SendSDLog("gcpTemplateModel :"+text+err.Error(), "error")
		beego.Error(text, err)
		return errors.New(text), ""
	}
	i := rand.Int()

	template.TemplateId = template.Name + strconv.Itoa(i)

	template.CreationDate = time.Now()
	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoGcpTemplateCollection, template)
	if err != nil {
		beego.Error("Template model: Create - Got error inserting template to the database: ", err)
		return err, ""
	}

	return nil, template.TemplateId
}

func GetTemplate(templateName string, ctx utils.Context) (template Template, err error) {
	session, err1 := db.GetMongoSession()
	if err1 != nil {
		ctx.SendSDLog("GcpTemplateModel :"+err1.Error(), "error")
		beego.Error("Template model: Get - Got error while connecting to the database: ", err1)
		return Template{}, err1
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGcpTemplateCollection)
	err = c.Find(bson.M{"name": templateName}).One(&template)
	if err != nil {
		ctx.SendSDLog("GcpTemplateModel :"+err.Error(), "error")
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
	session, err1 := db.GetMongoSession()
	if err1 != nil {
		beego.Error("Template model: GetAll - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoAwsTemplateCollection)
	err = c.Find(bson.M{"template_id": bson.M{"$in": copyData}}).All(&templates)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return nil, err
	}

	return templates, nil
}
func GetAllTemplate(ctx utils.Context) (templates []Template, err error) {
	session, err1 := db.GetMongoSession()
	if err1 != nil {
		ctx.SendSDLog("GcpTemplateModel : error connecting to database "+err1.Error(), "error")
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
	oldTemplate, err := GetTemplate(template.TemplateId, ctx)
	if err != nil {
		text := fmt.Sprintf("Template model: Update - Template '%s' does not exist in the database: ", template.Name)
		ctx.SendSDLog("GcpTemplateModel "+text+err.Error(), "error")
		beego.Error(text, err)
		return errors.New(text)
	}

	err = DeleteTemplate(template.TemplateId, ctx)
	if err != nil {
		ctx.SendSDLog("GcpTemplateModel :"+err.Error(), "error")
		beego.Error("Template model: Update - Got error deleting template: ", err)
		return err
	}

	template.CreationDate = oldTemplate.CreationDate
	template.ModificationDate = time.Now()

	err, _ = CreateTemplate(template, ctx)
	if err != nil {
		ctx.SendSDLog("GcpTemplateModel :"+err.Error(), "error")
		beego.Error("Template model: Update - Got error creating template: ", err)
		return err
	}

	return nil
}

func DeleteTemplate(templateName string, ctx utils.Context) error {
	session, err := db.GetMongoSession()
	if err != nil {
		ctx.SendSDLog("GcpTemplateModel : erro with connecting database "+err.Error(), "error")
		beego.Error("Template model: Delete - Got error while connecting to the database: ", err)
		return err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoGcpTemplateCollection)
	err = c.Remove(bson.M{"name": templateName})
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}
