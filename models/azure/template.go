package azure

import (
	"antelope/models"
	"antelope/models/db"
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
	ID                 bson.ObjectId      `json:"_id" bson:"_id,omitempty"`
	Name               string             `json:"name" bson:"name"`
	NodeCount          int64              `json:"node_count" bson:"node_count"`
	MachineType        string             `json:"machine_type" bson:"machine_type"`
	Image              ImageReference     `json:"image" bson:"image"`
	PoolSubnet         string             `json:"subnet_id" bson:"subnet_id"`
	PoolSecurityGroups []*string          `json:"security_group_id" bson:"security_group_id"`
	Nodes              []*VM              `json:"nodes" bson:"nodes"`
	PoolRole           string             `json:"pool_role" bson:"pool_role"`
	AdminUser          string             `json:"user_name" bson:"user_name",omitempty"`
	KeyInfo            utils.Key          `json:"key_info" bson:"key_info"`
	BootDiagnostics    DiagnosticsProfile `json:"boot_diagnostics" bson:"boot_diagnostics"`
	OsDisk             models.OsDiskType  `json:"os_disk_type" bson:"os_disk_type"`
}

func checkTemplateSize(cluster Template) error {
	for _, pools := range cluster.NodePools {
		if pools.NodeCount > 3 {
			return errors.New("Nodepool can't have more than 3 nodes")
		}
	}
	return nil
}
func CreateTemplate(template Template, ctx utils.Context) (error, string) {
	_, err := GetTemplate(template.TemplateId, ctx)
	if err == nil { //template found
		text := fmt.Sprintf("Template model: Create - Template '%s' already exists in the database: ", template.Name)
		ctx.SendSDLog(text, "error")
		return errors.New(text), ""
	}
	i := rand.Int()

	template.TemplateId = template.Name + strconv.Itoa(i)

	template.CreationDate = time.Now()

	err = checkTemplateSize(template)
	if err != nil { //cluster found
		ctx.SendSDLog(err.Error(), "error")
		return err, ""
	}
	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoAzureTemplateCollection, template)
	if err != nil {
		ctx.SendSDLog("Template model: Create - Got error inserting template to the database: "+err.Error(), "error")
		return err, ""
	}

	return nil, template.TemplateId
}

func GetTemplate(templateName string, ctx utils.Context) (template Template, err error) {
	session, err1 := db.GetMongoSession()
	if err1 != nil {
		ctx.SendSDLog("Template model: Get - Got error while connecting to the database: "+err1.Error(), "error")
		return Template{}, err1
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAzureTemplateCollection)
	err = c.Find(bson.M{"name": templateName}).One(&template)
	if err != nil {
		ctx.SendSDLog(err1.Error(), "error")
		return Template{}, err
	}
	return template, nil
}

func GetAllTemplate(ctx utils.Context) (templates []Template, err error) {
	session, err1 := db.GetMongoSession()
	if err1 != nil {
		ctx.SendSDLog("Template model: GetAll - Got error while connecting to the database: "+err1.Error(), "error")
		return nil, err1
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAzureTemplateCollection)
	err = c.Find(bson.M{}).All(&templates)
	if err != nil {
		ctx.SendSDLog("Template model: GetAll - Got error while connecting to the database: "+err.Error(), "error")
		return nil, err
	}

	return templates, nil
}

func UpdateTemplate(template Template, ctx utils.Context) error {
	oldTemplate, err := GetTemplate(template.TemplateId, ctx)
	if err != nil {
		text := fmt.Sprintf("Template model: Update - Template '%s' does not exist in the database: ", template.Name)
		ctx.SendSDLog(err.Error(), "error")
		return errors.New(text)
	}

	err = DeleteTemplate(template.TemplateId, ctx)
	if err != nil {
		beego.Error("Template model: Update - Got error deleting template: ", err)
		return err
	}

	template.CreationDate = oldTemplate.CreationDate
	template.ModificationDate = time.Now()

	err, _ = CreateTemplate(template, ctx)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return err
	}

	return nil
}

func DeleteTemplate(templateName string, ctx utils.Context) error {
	session, err := db.GetMongoSession()
	if err != nil {

		ctx.SendSDLog(err.Error(), "error")
		return err
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoAzureTemplateCollection)
	err = c.Remove(bson.M{"name": templateName})
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return err
	}

	return nil
}
