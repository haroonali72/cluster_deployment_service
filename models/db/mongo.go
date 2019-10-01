package db

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"net"
	"strings"
)

func GetMongoSession() (session *mgo.Session, err error) {
	conf := GetMongoConf()

	beego.Info("connecting to mongo host: " + conf.mongoHost)

	tlsconfig := getTLSCertificate()
	if tlsconfig == nil {
		return
	}

	if !conf.mongoAuth {
		session, err = mgo.Dial(conf.mongoHost)
		beego.Info("Mongo host connected" + conf.mongoHost)
		return session, err
	}
	session, err = mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    strings.Split(conf.mongoHost, ","),
		Username: conf.mongoUser,
		Password: conf.mongoPass,
		DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
			conf := tlsconfig
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
	var conft tlsConfig
	conf.mongoHost = beego.AppConfig.String("mongo_host")
	conf.mongoUser = beego.AppConfig.String("mongo_user")
	conf.mongoPass = beego.AppConfig.String("mongo_pass")
	conf.mongoAuth, _ = beego.AppConfig.Bool("mongo_auth")
	conf.MongoDb = beego.AppConfig.String("mongo_db")
	conf.MongoSshKeyCollection = beego.AppConfig.String("mongo_ssh_keys_collection")
	conf.MongoAwsTemplateCollection = beego.AppConfig.String("mongo_aws_template_collection")
	conf.MongoAwsClusterCollection = beego.AppConfig.String("mongo_aws_cluster_collection")
	conf.MongoAzureClusterCollection = beego.AppConfig.String("mongo_azure_cluster_collection")
	conf.MongoAzureTemplateCollection = beego.AppConfig.String("mongo_azure_template_collection")
	conf.MongoGcpClusterCollection = beego.AppConfig.String("mongo_gcp_cluster_collection")
	conf.MongoGcpTemplateCollection = beego.AppConfig.String("mongo_gcp_template_collection")
	conft.CaCert = beego.AppConfig.String("ca_certificate")
	conft.ClientCert = beego.AppConfig.String("client_cert")
	conft.ClientPem = beego.AppConfig.String("client_pem")
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
	MongoGcpTemplateCollection   string
	MongoGcpClusterCollection    string
	MongoSshKeyCollection        string
}
type tlsConfig struct {
	ClientCert string
	ClientPem  string
	CaCert     string
}

func getTLSCertificate() *tls.Config {
	var conf tlsConfig
	tlsConfig := &tls.Config{}
	if conf.CaCert != "" {
		rootCAs := x509.NewCertPool()
		rootCert, err := ioutil.ReadFile(conf.ClientCert)
		if err != nil {
			return nil
		}
		rootCAs.AppendCertsFromPEM(rootCert)
		if conf.ClientCert == "" || conf.ClientPem == "" {
			return nil
		}
		clientCrt, err := ioutil.ReadFile(conf.ClientCert)
		if err != nil {
			return nil
		}
		clientPem, err := ioutil.ReadFile(conf.ClientPem)
		if err != nil {
			return nil
		}
		clientCertificate, err := tls.X509KeyPair(clientCrt, clientPem)
		if err != nil {
			return nil
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, clientCertificate)
		tlsConfig.RootCAs = rootCAs
		tlsConfig.BuildNameToCertificate()
	} else {
		tlsConfig.InsecureSkipVerify = true
	}
	return tlsConfig
}
