package db

import (
	"crypto/tls"
	"errors"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2"
	"net"
)

var (
	mongoHost                  = beego.AppConfig.String("mongo_host")
	mongoUser                  = beego.AppConfig.String("mongo_user")
	mongoPass                  = beego.AppConfig.String("mongo_pass")
	mongoAuth, _               = beego.AppConfig.Bool("mongo_auth")
	MongoDb                    = beego.AppConfig.String("mongo_db")
	MongoAwsTemplateCollection = beego.AppConfig.String("mongo_aws_template_collection")
	MongoAwsClusterCollection  = beego.AppConfig.String("mongo_aws_cluster_collection")
	MongoAzureTemplateCollection = beego.AppConfig.String("mongo_azure_template_collection")
	MongoAzureClusterCollection  = beego.AppConfig.String("mongo_azure_cluster_collection")
)

func GetMongoSession() (session *mgo.Session, err error) {
	beego.Info("connecting to mongo host: " + mongoHost + "")

	if !mongoAuth {
		session, err = mgo.Dial(mongoHost)
		return session, err
	}

	session, err = mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{mongoHost},
		Username: mongoUser,
		Password: mongoPass,
		DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
			conf := &tls.Config{
				InsecureSkipVerify: true,
			}
			return tls.Dial("tcp", addr.String(), conf)
		},
	})

	return session, err
}

func IsMongoAlive() bool {
	_, err := GetMongoSession()
	if err != nil {
		beego.Error("unable to establish connection to " + mongoHost + " mongo db")
		return false
	}

	beego.Info("successfully connected to mongo host: " + mongoHost + "")
	return true
}

func InsertInMongo(collection string, data interface{}) error {
	session, err := GetMongoSession()
	if err != nil {
		errorText := "unable to establish connection to " + mongoHost + " db"
		beego.Error(errorText)
		return errors.New(errorText)
	}
	defer session.Close()

	c := session.DB(MongoDb).C(collection)
	err = c.Insert(data)
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}
