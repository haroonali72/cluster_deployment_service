package register_customer_template

import (
	"antelope/models"
	"antelope/models/aws"
	"antelope/models/azure"
	"antelope/models/db"
	"antelope/models/do"
	"antelope/models/gcp"
	"antelope/models/ibm"
	rbac "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"errors"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"strconv"
	"time"
)

func RegisterCustomerTemplate(awsTemplates []aws.Template, azureTemplates []azure.Template, gcpTemplates []gcp.Template, doTemplates []do.Template, ibmTemplates []ibm.Template, companyId string, ctx utils.Context) error {

	for index, template := range awsTemplates {
		awsTemplates[index].CompanyId = companyId
		awsTemplates[index].CreationDate = time.Now()
		if template.TemplateId == "" {
			i := rand.Int()
			awsTemplates[index].TemplateId = template.Name + strconv.Itoa(i)
		}
		awsTemplates[index].ID = bson.NewObjectId()
		awsTemplates[index].IsCloudplex = true
	}

	for index, template := range azureTemplates {
		azureTemplates[index].CompanyId = companyId
		azureTemplates[index].CreationDate = time.Now()
		if template.TemplateId == "" {
			i := rand.Int()
			azureTemplates[index].TemplateId = template.Name + strconv.Itoa(i)
		}
		azureTemplates[index].ID = bson.NewObjectId()
		azureTemplates[index].IsCloudplex = true
	}

	for index, template := range gcpTemplates {
		gcpTemplates[index].CompanyId = companyId
		gcpTemplates[index].CreationDate = time.Now()
		if template.TemplateId == "" {
			i := rand.Int()
			gcpTemplates[index].TemplateId = template.Name + strconv.Itoa(i)
		}
		gcpTemplates[index].ID = bson.NewObjectId()
		gcpTemplates[index].IsCloudplex = true
	}

	for index, template := range doTemplates {
		doTemplates[index].CompanyId = companyId
		doTemplates[index].CreationDate = time.Now()
		if template.TemplateId == "" {
			i := rand.Int()
			doTemplates[index].TemplateId = template.Name + strconv.Itoa(i)
		}
		doTemplates[index].ID = bson.NewObjectId()
		doTemplates[index].IsCloudplex = true
	}

	for index, template := range ibmTemplates {
		ibmTemplates[index].CompanyId = companyId
		ibmTemplates[index].CreationDate = time.Now()
		if template.TemplateId == "" {
			i := rand.Int()
			doTemplates[index].TemplateId = template.Name + strconv.Itoa(i)
		}
		ibmTemplates[index].ID = bson.NewObjectId()
		ibmTemplates[index].IsCloudplex = true
	}

	s := db.GetMongoConf()

	var awsInterface []interface{}
	for _, template := range awsTemplates {
		awsInterface = append(awsInterface, template)
	}

	err := db.InsertManyInMongo(s.MongoAwsTemplateCollection, awsInterface)
	if err != nil {
		ctx.SendLogs("Template model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	var azureInterface []interface{}
	for _, template := range azureTemplates {
		azureInterface = append(azureInterface, template)
	}

	err = db.InsertManyInMongo(s.MongoAzureTemplateCollection, azureInterface)
	if err != nil {
		ctx.SendLogs("Template model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	var gcpInterface []interface{}
	for _, template := range gcpTemplates {
		gcpInterface = append(gcpInterface, template)
	}

	err = db.InsertManyInMongo(s.MongoGcpTemplateCollection, gcpInterface)
	if err != nil {
		ctx.SendLogs("Template model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	var doInterface []interface{}
	for _, template := range doTemplates {
		doInterface = append(doInterface, template)
	}

	err = db.InsertManyInMongo(s.MongoDOTemplateCollection, doInterface)
	if err != nil {
		ctx.SendLogs("Template model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	var ibmInterface []interface{}
	for _, template := range ibmTemplates {
		ibmInterface = append(ibmInterface, template)
	}

	err = db.InsertManyInMongo(s.MongoIBMTemplateCollection, ibmInterface)
	if err != nil {
		ctx.SendLogs("Template model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func GetCustomerTemplate(ctx utils.Context) ([]aws.Template, []azure.Template, []gcp.Template, []do.Template, []ibm.Template, error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Template model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, nil, nil, nil, nil, err1
	}

	defer session.Close()
	s := db.GetMongoConf()

	var awsTemplates []aws.Template
	c := session.DB(s.MongoDb).C(s.MongoAwsCustomerTemplateCollection)
	err := c.Find(bson.M{}).All(&awsTemplates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, nil, nil, nil, nil, err1
	}

	var azureTemplates []azure.Template
	c = session.DB(s.MongoDb).C(s.MongoAzureCustomerTemplateCollection)
	err = c.Find(bson.M{}).All(&azureTemplates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, nil, nil, nil, nil, err1
	}

	var gcpTemplates []gcp.Template
	c = session.DB(s.MongoDb).C(s.MongoGcpCustomerTemplateCollection)
	err = c.Find(bson.M{}).All(&gcpTemplates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, nil, nil, nil, nil, err1
	}

	var doTemplates []do.Template
	c = session.DB(s.MongoDb).C(s.MongoDOCustomerTemplateCollection)
	err = c.Find(bson.M{}).All(&doTemplates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, nil, nil, nil, nil, err1
	}

	var ibmTemplates []ibm.Template
	c = session.DB(s.MongoDb).C(s.MongoIBMCustomerTemplateCollection)
	err = c.Find(bson.M{}).All(&ibmTemplates)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, nil, nil, nil, nil, err1
	}

	return awsTemplates, azureTemplates, gcpTemplates, doTemplates, ibmTemplates, nil
}
func CreatePolicy(awsTemplates []aws.Template, azureTemplates []azure.Template, gcpTemplates []gcp.Template, doTemplates []do.Template, ibmTemplates []ibm.Template, token string, ctx utils.Context) error {

	for _, template := range awsTemplates {
		statusCode, err := rbac.CreatePolicy(template.TemplateId, token, ctx.Data.UserId, ctx.Data.Company, models.POST, nil, models.AWS, ctx)
		if err != nil || statusCode != 200 {
			return errors.New("error occured in creation policy")
		}
	}

	for _, template := range azureTemplates {
		statusCode, err := rbac.CreatePolicy(template.TemplateId, token, ctx.Data.UserId, ctx.Data.Company, models.POST, nil, models.Azure, ctx)
		if err != nil || statusCode != 200 {
			return errors.New("error occured in creation policy")
		}
	}

	for _, template := range gcpTemplates {
		statusCode, err := rbac.CreatePolicy(template.TemplateId, token, ctx.Data.UserId, ctx.Data.Company, models.POST, nil, models.GCP, ctx)
		if err != nil || statusCode != 200 {
			return errors.New("error occured in creation policy")
		}
	}

	for _, template := range doTemplates {
		statusCode, err := rbac.CreatePolicy(template.TemplateId, token, ctx.Data.UserId, ctx.Data.Company, models.POST, nil, models.DO, ctx)
		if err != nil || statusCode != 200 {
			return errors.New("error occured in creation policy")
		}
	}

	for _, template := range ibmTemplates {
		statusCode, err := rbac.CreatePolicy(template.TemplateId, token, ctx.Data.UserId, ctx.Data.Company, models.POST, nil, models.IBM, ctx)
		if err != nil || statusCode != 200 {
			return errors.New("error occured in creation policy")
		}
	}
	return nil
}
