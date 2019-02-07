package utils

import (
	"github.com/astaxie/beego"
	"github.com/urfave/cli"
	"log"
	"os"
)

var (
	mongo                           = ""
	mongo_auth                      = "" // false
	mongo_db                        = ""
	mongo_user                      = ""
	mongo_pass                      = ""
	mongo_aws_template_collection   = ""
	mongo_aws_cluster_collection    = ""
	mongo_azure_template_collection = ""
	mongo_azure_cluster_collection  = ""
	redis_url                       = ""
	logger_url                      = ""
	network_url                     = ""
)

func InitFlags() error {
	//	os.Setenv()

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
	}
	app.Action = func(c *cli.Context) error {
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Println(err)
		return err
	}

	beego.AppConfig.Set("mongo_host", mongo)
	beego.AppConfig.Set("mongo_user", mongo_user)
	beego.AppConfig.Set("mongo_pass", mongo_pass)
	beego.AppConfig.Set("mongo_auth", mongo_auth)
	beego.AppConfig.Set("mongo_db", mongo_db)
	beego.AppConfig.Set("mongo_aws_template_collection", mongo_aws_cluster_collection)
	beego.AppConfig.Set("mongo_aws_cluster_collection", mongo_aws_template_collection)
	beego.AppConfig.Set("mongo_azure_template_collection", mongo_azure_template_collection)
	beego.AppConfig.Set("mongo_azure_cluster_collection", mongo_azure_cluster_collection)
	beego.AppConfig.Set("redis_url", redis_url)
	beego.AppConfig.Set("logger_url", logger_url)
	beego.AppConfig.Set("network_url", network_url)
	return nil
}
