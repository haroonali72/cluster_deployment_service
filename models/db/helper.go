package db

import (
	"antelope/models"
	"antelope/models/types"
	"antelope/models/utils"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

func DeleteError(infraId, companyId string, ctx utils.Context) error {
	session, err := GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"AKSDeleteClusterModel:  Delete - Got error while connecting to the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	defer session.Close()
	mc := GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoClusterErrorCollection)
	err = c.Remove(bson.M{"infra_id": infraId, "company_id": companyId})
	if err != nil {
		ctx.SendLogs(
			"AKSDeleteClusterModel:  Delete - Got error while deleting from the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}

	return nil
}
func AddError(ctx utils.Context, errDef types.ClusterError) error {
	session, err := GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs(
			"AKSAddClusterModel:  Add - Got error while connecting to the database: "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}
	defer session.Close()

	mc := GetMongoConf()
	err = InsertInMongo(mc.MongoClusterErrorCollection, errDef)
	if err != nil {
		ctx.SendLogs(
			"AKSAddClusterModel:  Add - Got error while inserting cluster to the database:  "+err.Error(),
			models.LOGGING_LEVEL_ERROR,
			models.Backend_Logging,
		)
		return err
	}
	return nil
}
func CreateError(infraId, companyId string, cloud models.Cloud, ctx utils.Context, errDef types.CustomCPError) error {

	var customErr types.ClusterError
	customErr.InfraId = infraId
	customErr.Cloud = cloud
	customErr.CompanyId = companyId
	customErr.Err = errDef

	obj, err := GetError(infraId, companyId, cloud, ctx)
	if err != nil && strings.Contains(err.Error(), "not found") {
		err = AddError(ctx, customErr)
		if err != nil {
			ctx.SendLogs(
				"Update - Got error inserting in db "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
			return err
		}
	} else if err == nil && obj != (types.ClusterError{}) {
		err = DeleteError(infraId, companyId, ctx)
		if err != nil {
			ctx.SendLogs(
				"Update - Got error deleting from db "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
		}
		err = AddError(ctx, customErr)
		if err != nil {
			ctx.SendLogs(
				"Add - Got error while inserting cluster to the database:  "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
		}
	} else if err == nil && obj == (types.ClusterError{}) {
		err = AddError(ctx, customErr)
		if err != nil {
			ctx.SendLogs(
				"Add - Got error while inserting cluster to the database:  "+err.Error(),
				models.LOGGING_LEVEL_ERROR,
				models.Backend_Logging,
			)
			return err
		}
	}
	return nil
}
func GetError(infraId, companyId string, cloud models.Cloud, ctx utils.Context) (err types.ClusterError, err1 error) {

	session, err1 := GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Cluster model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.ClusterError{}, err1
	}

	defer session.Close()
	mc := GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoClusterErrorCollection)
	err1 = c.Find(bson.M{"infra_id": infraId, "company_id": companyId, "cloud": cloud}).One(&err)
	if err1 != nil {
		if err1 != nil && strings.Contains(strings.ToLower(err1.Error()), "not found") {
			return err, nil
		}
		ctx.SendLogs("Cluster model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return types.ClusterError{}, err1
	}
	return err, nil
}
