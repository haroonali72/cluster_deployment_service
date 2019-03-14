package aws

import (
	"antelope/models"
	"antelope/models/db"
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
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools        []*NodePoolT  `json:"node_pools" bson:"node_pools"`
	NetworkName      string        `json:"network_name" bson:"network_name"`
}

type NodePoolT struct {
	ID              bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name            string        `json:"machine_type" bson:"machine_type"`
	Ami             Ami           `json:"name" bson:"name"`
	NodeCount       int32         `json:"node_count" bson:"node_count"`
	MachineType     string        `json:"ami" bson:"ami"`
	SubnetId        string        `json:"subnet_id" bson:"subnet_id"`
	SecurityGroupId []string      `json:"security_group_id" bson:"security_group_id"`
	KeyInfo         Key           `json:"key_info" bson:"key_info"`
	PoolRole        string        `json:"pool_role" bson:"pool_role"`
}

func CreateTemplate(template Template) (error, string) {
	_, err := GetTemplate(template.TemplateId)
	if err == nil { //template found
		text := fmt.Sprintf("Template model: Create - Template '%s' already exists in the database: ", template.Name)
		beego.Error(text, err)
		return errors.New(text), ""
	}

	template.CreationDate = time.Now()
	i := rand.Int()

	beego.Info(i)
	beego.Info(strconv.Itoa(i))

	template.TemplateId = template.Name + strconv.Itoa(i)

	beego.Info(template.TemplateId)
	s := db.GetMongoConf()
	err = db.InsertInMongo(s.MongoAwsTemplateCollection, template)
	if err != nil {
		beego.Error("Template model: Create - Got error inserting template to the database: ", err)
		return err, ""
	}

	return nil, template.TemplateId
}

func GetTemplate(templateId string) (template Template, err error) {
	session, err1 := db.GetMongoSession()
	if err1 != nil {
		beego.Error("Template model: Get - Got error while connecting to the database: ", err1)
		return Template{}, err1
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoAwsTemplateCollection)
	err = c.Find(bson.M{"template_id": templateId}).One(&template)
	if err != nil {
		beego.Error(err.Error())
		return Template{}, err
	}

	return template, nil
}

func GetAllTemplate() (templates []Template, err error) {
	session, err1 := db.GetMongoSession()
	if err1 != nil {
		beego.Error("Template model: GetAll - Got error while connecting to the database: ", err1)
		return nil, err1
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoAwsTemplateCollection)
	err = c.Find(bson.M{}).All(&templates)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}

	return templates, nil
}

func UpdateTemplate(template Template) error {
	oldTemplate, err := GetTemplate(template.TemplateId)
	if err != nil {
		text := fmt.Sprintf("Template model: Update - Template '%s' does not exist in the database: ", template.TemplateId)
		beego.Error(text, err)
		return errors.New(text)
	}

	err = DeleteTemplate(template.TemplateId)
	if err != nil {
		beego.Error("Template model: Update - Got error deleting template: ", err)
		return err
	}

	template.CreationDate = oldTemplate.CreationDate
	template.ModificationDate = time.Now()

	err, _ = CreateTemplate(template)
	if err != nil {
		beego.Error("Template model: Update - Got error creating template: ", err)
		return err
	}

	return nil
}

func DeleteTemplate(templateId string) error {
	session, err := db.GetMongoSession()
	if err != nil {
		beego.Error("Template model: Delete - Got error while connecting to the database: ", err)
		return err
	}
	defer session.Close()
	s := db.GetMongoConf()
	c := session.DB(s.MongoDb).C(s.MongoAwsTemplateCollection)
	err = c.Remove(bson.M{"template_id": templateId})
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}
