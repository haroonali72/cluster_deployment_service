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

	ctx.SendLogs("Connecting to mongo host: "+conf.mongoHost, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	if !conf.mongoAuth {
		session, err = mgo.Dial(conf.mongoHost)
		ctx.SendLogs("Mongo host connected: "+conf.mongoHost, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		return session, err
	}

	tlsconfig := getTLSCertificate(ctx)
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
		ctx.SendLogs("Unable to establish connection to "+conf.mongoHost+" mongo db", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return false
	}

	ctx.SendLogs("Successfully connected to mongo host: "+conf.mongoHost+"", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	return true
}

func InsertManyInMongo(collection string, data []interface{}) error {
	conf := GetMongoConf()
	ctx := new(utils.Context)
	session, err := GetMongoSession(*ctx)
	if err != nil {
		errorText := "Unable to establish connection to " + conf.mongoHost + " db"
		ctx.SendLogs(errorText, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(errorText)
	}
	defer session.Close()

	c := session.DB(conf.MongoDb).C(collection)
	err = c.Insert(data...)
	if err != nil {
		ctx.SendLogs("Unable to insert data in mongo db "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}
func InsertInMongo(collection string, data interface{}) error {
	conf := GetMongoConf()
	ctx := new(utils.Context)
	session, err := GetMongoSession(*ctx)
	if err != nil {
		errorText := "Unable to establish connection to " + conf.mongoHost + " db"
		ctx.SendLogs(errorText, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(errorText)
	}
	defer session.Close()

	c := session.DB(conf.MongoDb).C(collection)
	err = c.Insert(data)
	if err != nil {
		ctx.SendLogs("Unable to insert data in mongo db "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
	conf.MongoDOKSClusterCollection = beego.AppConfig.String("mongo_doks_cluster_collection")
	conf.MongoDOKSTemplateCollection = beego.AppConfig.String("mongo_doks_template_collection")
	conf.MongoAzureClusterCollection = beego.AppConfig.String("mongo_azure_cluster_collection")
	conf.MongoAzureTemplateCollection = beego.AppConfig.String("mongo_azure_template_collection")
	conf.MongoDefaultTemplateCollection = beego.AppConfig.String("mongo_default_template_collection")
	conf.MongoGcpClusterCollection = beego.AppConfig.String("mongo_gcp_cluster_collection")
	conf.MongoGcpTemplateCollection = beego.AppConfig.String("mongo_gcp_template_collection")
	conf.MongoGKEClusterCollection = beego.AppConfig.String("mongo_gke_cluster_collection")
	conf.MongoGKETemplateCollection = beego.AppConfig.String("mongo_gke_template_collection")
	conf.MongoEKSClusterCollection = beego.AppConfig.String("mongo_eks_cluster_collection")
	conf.MongoEKSTemplateCollection = beego.AppConfig.String("mongo_eks_template_collection")
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
	conf.MongoEKSCustomerTemplateCollection = "mongo_eks_customer_template_collection"
	conf.MongoAKSCustomerTemplateCollection = "mongo_aks_customer_template_collection"
	conf.MongoDOCustomerTemplateCollection = "mongo_do_customer_template_collection"
	conf.MongoIKSCustomerTemplateCollection = "mongo_iks_customer_template_collection"
	conf.MongoDOKSCustomerTemplateCollection = "mongo_doks_customer_template_collection"
	conf.MongoClusterErrorCollection = "mongo_cluster_error_collection"
	conf.MongoEKSPreviousClusterCollection = "eks_previous_cluster"
	conf.MongoGKEPreviousClusterCollection = "gke_previous_cluster"
	conf.MongoAKSPreviousClusterCollection = "aks_previous_cluster"
	conf.MongoIKSPreviousClusterCollection = "iks_previous_cluster"
	conf.MongoDOKSPreviousClusterCollection = "doks_previous_cluster"
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
	MongoGKEPreviousClusterCollection    string
	MongoDOKSPreviousClusterCollection   string
	MongoAKSPreviousClusterCollection    string
	MongoEKSTemplateCollection           string
	MongoEKSCustomerTemplateCollection   string
	MongoEKSClusterCollection            string
	MongoEKSPreviousClusterCollection    string
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
	MongoIKSPreviousClusterCollection    string
	MongoOPClusterCollection             string
	MongoOPTemplateCollection            string
	MongoDOKSClusterCollection           string
	MongoClusterErrorCollection          string
	MongoDOKSTemplateCollection          string
	MongoDOKSCustomerTemplateCollection  string
	MongoDefaultTemplateCollection       string
}
type tlsConfig struct {
	ClientCert string
	ClientPem  string
	CaCert     string
}

func getTLSCertificate(ctx utils.Context) *tls.Config {
	var conf tlsConfig
	conf.CaCert = beego.AppConfig.String("ca_certificate")
	conf.ClientCert = beego.AppConfig.String("client_cert")
	conf.ClientPem = beego.AppConfig.String("client_pem")
	tlsConfig := &tls.Config{}
	if conf.CaCert != "" {
		rootCAs := x509.NewCertPool()
		rootCert, err := ioutil.ReadFile(conf.CaCert)
		if err != nil {
			ctx.SendLogs("Error in getting root certificate"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return nil
		}
		rootCAs.AppendCertsFromPEM(rootCert)
		if conf.ClientCert == "" || conf.ClientPem == "" {
			ctx.SendLogs("Error in getting certificate"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return nil
		}
		clientCrt, err := ioutil.ReadFile(conf.ClientCert)
		if err != nil {
			ctx.SendLogs("Error in getting client certificate"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return nil
		}
		clientPem, err := ioutil.ReadFile(conf.ClientPem)
		if err != nil {
			ctx.SendLogs("Error in getting client certificate"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return nil
		}
		clientCertificate, err := tls.X509KeyPair(clientCrt, clientPem)
		if err != nil {
			ctx.SendLogs("Error in getting client certificate"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
