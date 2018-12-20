package aws

import (
	"antelope/models"
	"antelope/models/db"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Template struct {
	ID               bson.ObjectId  `json:"_id" bson:"_id,omitempty"`
	EnvironmentId    string         `json:"environment_id" bson:"environment_id"`
	Name             string         `json:"name" bson:"name"`
	Cloud            models.Cloud   `json:"cloud" bson:"cloud"`
	CreationDate     time.Time      `json:"-" bson:"creation_date"`
	ModificationDate time.Time      `json:"-" bson:"modification_date"`
	NodePools []*NodePoolT   `json:"node_pools" bson:"node_pools"`
}

/*type SubclusterT struct {
	ID        bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name      string        `json:"name" bson:"name"`
	NodePools []*NodePool   `json:"node_pools" bson:"node_pools"`
}
*/
type NodePoolT struct {
	ID              bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name            string        `json:"name" bson:"name"`
	NodeCount       int32         `json:"node_count" bson:"node_count"`
	MachineType     string        `json:"machine_type" bson:"machine_type"`
	Ami             Ami           `json:"ami" bson:"ami"`
	SubnetId        string `json:"subnet_id" bson:"subnet_id"`
	SecurityGroupId []string `json:"security_group_id" bson:"security_group_id"`
}

type AmiT struct {
	ID       bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	Name     string        `json:"name" bson:"name"`
	Username string        `json:"username" bson:"username"`
}

func CreateTemplate(template Template) error {
	_, err := GetTemplate(template.Name)
	if err == nil { //template found
		text := fmt.Sprintf("Template model: Create - Template '%s' already exists in the database: ", template.Name)
		beego.Error(text, err)
		return errors.New(text)
	}

	template.CreationDate = time.Now()

	err = db.InsertInMongo(db.MongoAwsTemplateCollection, template)
	if err != nil {
		beego.Error("Template model: Create - Got error inserting template to the database: ", err)
		return err
	}

	return nil
}

func GetTemplate(templateName string) (template Template, err error) {
	session, err1 := db.GetMongoSession()
	if err1 != nil {
		beego.Error("Template model: Get - Got error while connecting to the database: ", err1)
		return Template{}, err1
	}
	defer session.Close()

	c := session.DB(db.MongoDb).C(db.MongoAwsTemplateCollection)
	err = c.Find(bson.M{"name": templateName}).One(&template)
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

	c := session.DB(db.MongoDb).C(db.MongoAwsTemplateCollection)
	err = c.Find(bson.M{}).All(&templates)
	if err != nil {
		beego.Error(err.Error())
		return nil, err
	}

	return templates, nil
}

func UpdateTemplate(template Template) error {
	oldTemplate, err := GetTemplate(template.Name)
	if err != nil {
		text := fmt.Sprintf("Template model: Update - Template '%s' does not exist in the database: ", template.Name)
		beego.Error(text, err)
		return errors.New(text)
	}

	err = DeleteTemplate(template.Name)
	if err != nil {
		beego.Error("Template model: Update - Got error deleting template: ", err)
		return err
	}

	template.CreationDate = oldTemplate.CreationDate
	template.ModificationDate = time.Now()

	err = CreateTemplate(template)
	if err != nil {
		beego.Error("Template model: Update - Got error creating template: ", err)
		return err
	}

	return nil
}

func DeleteTemplate(templateName string) error {
	session, err := db.GetMongoSession()
	if err != nil {
		beego.Error("Template model: Delete - Got error while connecting to the database: ", err)
		return err
	}
	defer session.Close()

	c := session.DB(db.MongoDb).C(db.MongoAwsTemplateCollection)
	err = c.Remove(bson.M{"name": templateName})
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}
