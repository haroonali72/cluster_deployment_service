package db

import (
	"antelope/models"
	"antelope/models/utils"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"net"
	"strings"
)

func GetMongoSession(ctx utils.Context) (session *mgo.Session, err error) {
	conf := GetMongoConf()

	beego.Info("connecting to mongo host: " + conf.mongoHost)

	if !conf.mongoAuth {
		session, err = mgo.Dial(conf.mongoHost)
		beego.Info("Mongo host connected: " + conf.mongoHost)
		return session, err
	}

	tlsconfig := getTLSCertificate()
	if tlsconfig == nil {
		return
	}

	session, err = mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    strings.Split(conf.mongoHost, ","),
		Username: conf.mongoUser,
		Password: conf.mongoPass,
		DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
			conf := tlsconfig
			dial, err := tls.Dial("tcp", addr.String(), conf)
			if err != nil {
				ctx.SendLogs(" Db connection: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			}
			return dial, err
		},
	})

	return session, err
}

func IsMongoAlive() bool {

	conf := GetMongoConf()
	ctx := new(utils.Context)
	_, err := GetMongoSession(*ctx)
	if err != nil {
		beego.Error(err.Error())
		beego.Error("unable to establish connection to " + conf.mongoHost + " mongo db")
		return false
	}

	beego.Info("successfully connected to mongo host: " + conf.mongoHost + "")
	return true
}
func InsertManyInMongo(collection string, data []interface{}) error {
	conf := GetMongoConf()
	ctx := new(utils.Context)
	session, err := GetMongoSession(*ctx)
	if err != nil {
		errorText := "unable to establish connection to " + conf.mongoHost + " db"
		beego.Error(errorText)
		return errors.New(errorText)
	}
	defer session.Close()

	c := session.DB(conf.MongoDb).C(collection)
	err = c.Insert(data...)
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	return nil
}
func InsertInMongo(collection string, data interface{}) error {
	conf := GetMongoConf()
	ctx := new(utils.Context)
	session, err := GetMongoSession(*ctx)
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
	conf.MongoSshKeyCollection = beego.AppConfig.String("mongo_ssh_keys_collection")
	conf.MongoAwsTemplateCollection = beego.AppConfig.String("mongo_aws_template_collection")
	conf.MongoAwsClusterCollection = beego.AppConfig.String("mongo_aws_cluster_collection")
	conf.MongoAzureClusterCollection = beego.AppConfig.String("mongo_azure_cluster_collection")
	conf.MongoAzureTemplateCollection = beego.AppConfig.String("mongo_azure_template_collection")
	conf.MongoGcpClusterCollection = beego.AppConfig.String("mongo_gcp_cluster_collection")
	conf.MongoGcpTemplateCollection = beego.AppConfig.String("mongo_gcp_template_collection")
	conf.MongoGKEClusterCollection = beego.AppConfig.String("mongo_gke_cluster_collection")
	conf.MongoGKETemplateCollection = beego.AppConfig.String("mongo_gke_template_collection")
	conf.MongoAKSClusterCollection = beego.AppConfig.String("mongo_aks_cluster_collection")
	conf.MongoAKSTemplateCollection = beego.AppConfig.String("mongo_aks_template_collection")
	conf.MongoDOClusterCollection = beego.AppConfig.String("mongo_do_cluster_collection")
	conf.MongoDOKSClusterCollection = beego.AppConfig.String("mongo_doks_cluster_collection")
	conf.MongoDOTemplateCollection = beego.AppConfig.String("mongo_do_template_collection")
	conf.MongoIKSClusterCollection = beego.AppConfig.String("mongo_iks_cluster_collection")
	conf.MongoIKSTemplateCollection = beego.AppConfig.String("mongo_iks_template_collection")
	conf.MongoOPClusterCollection = beego.AppConfig.String("mongo_op_cluster_collection")
	conf.MongoOPTemplateCollection = beego.AppConfig.String("mongo_op_template_collection")
	conf.MongoAwsCustomerTemplateCollection = "mongo_aws_customer_template_collection"
	conf.MongoAzureCustomerTemplateCollection = "mongo_azure_customer_template_collection"
	conf.MongoGcpCustomerTemplateCollection = "mongo_gcp_customer_template_collection"
	conf.MongoGKECustomerTemplateCollection = "mongo_gke_customer_template_collection"
	conf.MongoAKSCustomerTemplateCollection = "mongo_aks_customer_template_collection"
	conf.MongoDOCustomerTemplateCollection = "mongo_do_customer_template_collection"
	conf.MongoIKSCustomerTemplateCollection = "mongo_iks_customer_template_collection"
	return conf

}

type mongConf struct {
	mongoHost                            string
	mongoUser                            string
	mongoPass                            string
	mongoAuth                            bool
	MongoDb                              string
	MongoAwsTemplateCollection           string
	MongoAwsCustomerTemplateCollection   string
	MongoAwsClusterCollection            string
	MongoAzureTemplateCollection         string
	MongoAzureCustomerTemplateCollection string
	MongoAzureClusterCollection          string
	MongoGcpTemplateCollection           string
	MongoGcpCustomerTemplateCollection   string
	MongoGcpClusterCollection            string
	MongoGKETemplateCollection           string
	MongoGKECustomerTemplateCollection   string
	MongoGKEClusterCollection            string
	MongoAKSTemplateCollection           string
	MongoAKSCustomerTemplateCollection   string
	MongoAKSClusterCollection            string
	MongoSshKeyCollection                string
	MongoDOClusterCollection             string
	MongoDOTemplateCollection            string
	MongoDOCustomerTemplateCollection    string
	MongoIKSClusterCollection            string
	MongoIKSTemplateCollection           string
	MongoIKSCustomerTemplateCollection   string
	MongoOPClusterCollection             string
	MongoOPTemplateCollection            string
	MongoDOKSClusterCollection           string
}
type tlsConfig struct {
	ClientCert string
	ClientPem  string
	CaCert     string
}

func getTLSCertificate() *tls.Config {
	var conf tlsConfig
	conf.CaCert = beego.AppConfig.String("ca_certificate")
	conf.ClientCert = beego.AppConfig.String("client_cert")
	conf.ClientPem = beego.AppConfig.String("client_pem")
	tlsConfig := &tls.Config{}
	if conf.CaCert != "" {
		rootCAs := x509.NewCertPool()
		rootCert, err := ioutil.ReadFile(conf.CaCert)
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
