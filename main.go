package main

import (
	"antelope/models/db"
	"antelope/models/utils"
	_ "antelope/routers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
	"os"
)

func SecretAuth(username, password string) bool {
	// TODO configure basic authentication properly
	return username == "username" && password == "password"
}

func main() {
	setEnv()
	utils.InitFlags()
	if !db.IsMongoAlive() {
		os.Exit(1)
	}
	beego.BConfig.AppName = "antelope"
	beego.BConfig.CopyRequestBody = true
	beego.BConfig.WebConfig.EnableDocs = true
	beego.BConfig.WebConfig.AutoRender = true
	beego.BConfig.RunMode = "dev"
	beego.BConfig.WebConfig.DirectoryIndex = true
	beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	beego.BConfig.Listen.HTTPPort = 9081

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Token", "Content-type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// TODO enable basic authentication if required
	//authPlugin := auth.NewBasicAuthenticator(SecretAuth, "Authorization Required")
	//beego.InsertFilter("*", beego.BeforeRouter, authPlugin)

	beego.Run()
}

func setEnv() {

	os.Setenv("kill_bill_user", "admin")
	os.Setenv("kill_bill_password", "password")
	os.Setenv("kill_bill_secret_key", "cloudplex")
	os.Setenv("kill_bill_api_key", "cloudplex")
	os.Setenv("ca_cert","/home/zunaira/Downloads/mongoCA.crt")
	os.Setenv("client_cert", "/home/zunaira/Downloads/antelope.crt")
	os.Setenv("client_pem", "/home/zunaira/Downloads/antelope.pem")
	os.Setenv("subscription_host", "35.246.150.221:30906")
	os.Setenv("rbac_url", "https://dapis.cloudplex.cf")
	os.Setenv("mongo_host", "cloudplex-mongodb.cloudplex-system.svc.cluster.local:27017,mongodb-secondary-0.cloudplex-mongodb-headless:27017,mongodb-arbiter-0.cloudplex-mongodb-headless:27017")
	//os.Setenv("mongo_host", "localhost:27017")

	os.Setenv("mongo_auth", "true")
	os.Setenv("mongo_db", "antelope")
	os.Setenv("mongo_user", "antelope")
	os.Setenv("mongo_pass", "DbSn3hAzJU6pPVRcn61apb3KDEKmcSb7Bl..")
	os.Setenv("mongo_aws_template_collection", "aws_template")
	os.Setenv("mongo_aws_cluster_collection", "aws_cluster")
	os.Setenv("mongo_azure_template_collection", "azure_template")
	os.Setenv("mongo_azure_cluster_collection", "azure_cluster")
	os.Setenv("mongo_gcp_template_collection", "gcp_template")
	os.Setenv("mongo_gcp_cluster_collection", "gcp_cluster")
	os.Setenv("mongo_gke_template_collection", "gke_template")
	os.Setenv("mongo_gke_cluster_collection", "gke_cluster")
	os.Setenv("mongo_ssh_keys_collection", "ssh_key")
	os.Setenv("redis_url", "35.242.236.151:31845")
	os.Setenv("logger_url", "https://dapis.cloudplex.cf")
	os.Setenv("network_url", "https://dapis.cloudplex.cf")
	os.Setenv("vault_url", "https://dapis.cloudplex.cf")
	os.Setenv("raccoon_url", "https://dapis.cloudplex.cf")
	os.Setenv("jump_host_ip", "52.220.196.92")
	os.Setenv("jump_host_ssh_key", "/home/mahmad/ahmadnewkey.pem")
	os.Setenv("woodpecker_url", "http://woodpecker:3300")
}
