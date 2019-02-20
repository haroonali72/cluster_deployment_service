package db

import (
	"crypto/tls"
	"errors"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2"
	"net"
)

func GetMongoSession() (session *mgo.Session, err error) {
	conf := GetMongoConf()
	beego.Info("connecting to mongo host: " + conf.mongoHost)

	if !conf.mongoAuth {
		session, err = mgo.Dial(conf.mongoHost)
		return session, err
	}
	session, err = mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{conf.mongoHost},
		Username: conf.mongoUser,
		Password: conf.mongoPass,
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

	conf := GetMongoConf()
	_, err := GetMongoSession()
	if err != nil {
		beego.Error(err.Error())
		beego.Error("unable to establish connection to " + conf.mongoHost + " mongo db")
		return false
	}

	beego.Info("successfully connected to mongo host: " + conf.mongoHost + "")
	return true
}

func InsertInMongo(collection string, data interface{}) error {
	conf := GetMongoConf()
	session, err := GetMongoSession()
	if err != nil {
		errorText := "unable to establish connection to " + conf.mongoHost + " db"
		beego.Error(errorText)
		return errors.New(errorText)
	}
	defer session.Close()

	c := session.DB(conf.MongoDb).C(collection)
	err = c.Insert(data)
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}
func GetMongoConf() mongConf {

	var conf mongConf
	conf.mongoHost = beego.AppConfig.String("mongo_host")
	conf.mongoUser = beego.AppConfig.String("mongo_user")
	conf.mongoPass = beego.AppConfig.String("mongo_pass")
	conf.mongoAuth, _ = beego.AppConfig.Bool("mongo_auth")
	conf.MongoDb = beego.AppConfig.String("mongo_db")
	conf.MongoAwsTemplateCollection = beego.AppConfig.String("mongo_aws_template_collection")
	conf.MongoAwsClusterCollection = beego.AppConfig.String("mongo_aws_cluster_collection")
	conf.MongoAzureTemplateCollection = beego.AppConfig.String("mongo_azure_template_collection")
	conf.MongoAzureClusterCollection = beego.AppConfig.String("mongo_azure_cluster_collection")
	return conf
}

type mongConf struct {
	mongoHost                    string
	mongoUser                    string
	mongoPass                    string
	mongoAuth                    bool
	MongoDb                      string
	MongoAwsTemplateCollection   string
	MongoAwsClusterCollection    string
	MongoAzureTemplateCollection string
	MongoAzureClusterCollection  string
}
