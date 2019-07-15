package utils

import (
	"github.com/astaxie/beego"
	"github.com/urfave/cli"
	"log"
	"os"
)

var (
	mongo                           = ""
	mongo_auth                      = "" // boolean
	mongo_db                        = ""
	mongo_user                      = ""
	mongo_pass                      = ""
	mongo_ssh_keys_collection       = ""
	mongo_aws_template_collection   = ""
	mongo_aws_cluster_collection    = ""
	mongo_azure_template_collection = ""
	mongo_azure_cluster_collection  = ""
	mongo_gcp_template_collection   = ""
	mongo_gcp_cluster_collection    = ""
	redis_url                       = ""
	logger_url                      = ""
	vault_url                       = ""
	kube_engine_url                 = ""
	network_url                     = ""
)

func InitFlags() error {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "mongo_host",
			Usage:       "mongo db host",
			Destination: &mongo,
			EnvVar:      "mongo_host",
		},
		cli.StringFlag{
			Name:        "mongo_auth",
			Usage:       "mongo auth",
			Destination: &mongo_auth,
			EnvVar:      "mongo_auth",
		},
		cli.StringFlag{
			Name:        "mongo_db",
			Usage:       "mongo db name",
			Destination: &mongo_db,
			EnvVar:      "mongo_db",
		},
		cli.StringFlag{
			Name:        "mongo_user",
			Usage:       "mongo user name ",
			Destination: &mongo_user,
			EnvVar:      "mongo_user",
		},
		cli.StringFlag{
			Name:        "mongo_pass",
			Usage:       "mongo user password ",
			Destination: &mongo_pass,
			EnvVar:      "mongo_pass",
		},
		cli.StringFlag{
			Name:        "mongo_ssh_keys_collection",
			Usage:       "ssh keys collection name ",
			Destination: &mongo_ssh_keys_collection,
			EnvVar:      "mongo_ssh_keys_collection",
		},
		cli.StringFlag{
			Name:        "mongo_aws_template_collection",
			Usage:       "aws template collection name ",
			Destination: &mongo_aws_template_collection,
			EnvVar:      "mongo_aws_template_collection",
		},
		cli.StringFlag{
			Name:        "mongo_aws_cluster_collection",
			Usage:       "aws cluster collection name ",
			Destination: &mongo_aws_cluster_collection,
			EnvVar:      "mongo_aws_cluster_collection",
		},
		cli.StringFlag{
			Name:        "mongo_azure_template_collection",
			Usage:       "azure template collection name ",
			Destination: &mongo_azure_template_collection,
			EnvVar:      "mongo_azure_template_collection",
		},
		cli.StringFlag{
			Name:        "mongo_azure_cluster_collection",
			Usage:       "azure cluster collection name ",
			Destination: &mongo_azure_cluster_collection,
			EnvVar:      "mongo_azure_cluster_collection",
		},
		cli.StringFlag{
			Name:        "mongo_gcp_template_collection",
			Usage:       "gcp template collection name ",
			Destination: &mongo_gcp_template_collection,
			EnvVar:      "mongo_gcp_template_collection",
		},
		cli.StringFlag{
			Name:        "mongo_gcp_cluster_collection",
			Usage:       "gcp cluster collection name ",
			Destination: &mongo_gcp_cluster_collection,
			EnvVar:      "mongo_gcp_cluster_collection",
		},
		cli.StringFlag{
			Name:        "redis_url",
			Usage:       "redis host",
			Destination: &redis_url,
			EnvVar:      "redis_url",
		},
		cli.StringFlag{
			Name:        "logger_url",
			Usage:       "logger host ",
			Destination: &logger_url,
			EnvVar:      "logger_url",
		},
		cli.StringFlag{
			Name:        "network_url",
			Usage:       "weasel host",
			Destination: &network_url,
			EnvVar:      "network_url",
		},
		cli.StringFlag{
			Name:        "vault_url",
			Usage:       "vault host",
			Destination: &vault_url,
			EnvVar:      "vault_url",
		},
		cli.StringFlag{
			Name:        "kube_engine_url",
			Usage:       "kube_engine_url",
			Destination: &kube_engine_url,
			EnvVar:      "kube_engine_url",
		},
	}
	app.Action = func(c *cli.Context) error {
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Println(err)
		return err
	}

	host := mongo + ":32180"
	redis := redis_url + ":31845"
	elephant := "http://" + logger_url + ":3500/api/v1/logger"
	weasel := "http://" + network_url + ":9080/weasel/network/{cloud_provider}"
	vault := "http://" + vault_url + ":8092/robin/api/v1"
	kube := "http://" + kube_engine_url + "3300:/kube/api/v1/nodes"

	beego.AppConfig.Set("mongo_host", host)
	beego.AppConfig.Set("mongo_user", mongo_user)
	beego.AppConfig.Set("mongo_pass", mongo_pass)
	beego.AppConfig.Set("mongo_auth", mongo_auth)
	beego.AppConfig.Set("mongo_db", mongo_db)
	beego.AppConfig.Set("mongo_ssh_keys_collection", mongo_ssh_keys_collection)
	beego.AppConfig.Set("mongo_aws_template_collection", mongo_aws_template_collection)
	beego.AppConfig.Set("mongo_aws_cluster_collection", mongo_aws_cluster_collection)
	beego.AppConfig.Set("mongo_azure_template_collection", mongo_azure_template_collection)
	beego.AppConfig.Set("mongo_azure_cluster_collection", mongo_azure_cluster_collection)
	beego.AppConfig.Set("mongo_gcp_cluster_collection", mongo_gcp_cluster_collection)
	beego.AppConfig.Set("mongo_gcp_template_collection", mongo_gcp_template_collection)
	beego.AppConfig.Set("redis_url", redis)
	beego.AppConfig.Set("logger_url", elephant)
	beego.AppConfig.Set("network_url", weasel)
	beego.AppConfig.Set("vault_url", vault)
	beego.AppConfig.Set("kube_engine_url", kube)
	return nil
}
