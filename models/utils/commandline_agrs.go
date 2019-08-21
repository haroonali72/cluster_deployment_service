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
	network_url                     = ""
	raccoon_url                     = ""
	rbac_url                        = ""
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
			Usage:       "mongo user name",
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
			Name:        "raccoon_url",
			Usage:       "raccoon_url",
			Destination: &raccoon_url,
			EnvVar:      "racoon_url",
		},
		cli.StringFlag{
			Name:        "rbac_url",
			Usage:       "rbac_url",
			Destination: &rbac_url,
			EnvVar:      "rbac_url",
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

	mongo = mongo + ":32180"
	redis_url = redis_url + ":6379"
	network_url = "http://" + network_url + ":9080/weasel/network/{cloud_provider}"
	vault_url = "http://" + vault_url + ":8092/robin/api/v1"
	raccoon_url := "http://" + raccoon_url + "/raccoon/projects/"

	beego.AppConfig.Set("mongo_host", mongo)
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
	beego.AppConfig.Set("redis_url", redis_url)
	beego.AppConfig.Set("logger_url", logger_url)
	beego.AppConfig.Set("network_url", network_url)
	beego.AppConfig.Set("vault_url", vault_url)
	beego.AppConfig.Set("raccoon_url", raccoon_url)
	beego.AppConfig.Set("rbac_url", rbac_url)
	return nil
}
