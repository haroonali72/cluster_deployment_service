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
	os.Setenv("mongo_host", "35.246.150.221")
	os.Setenv("mongo_auth", "true")
	os.Setenv("mongo_db", "antelope")
	os.Setenv("mongo_user", "antelope")
	os.Setenv("mongo_pass", "antelope")
	os.Setenv("mongo_aws_template_collection", "aws_template")
	os.Setenv("mongo_aws_cluster_collection", "aws_cluster")
	os.Setenv("mongo_azure_template_collection", "azure_template")
	os.Setenv("mongo_azure_cluster_collection", "azure_cluster")
	os.Setenv("mongo_ssh_keys_collection", "ssh_key")
	os.Setenv("redis_url", "35.246.150.221")
	os.Setenv("logger_url", "https://dapis.cloudplex.cf/api/v1/logger")
	os.Setenv("network_url", "https://dapis.cloudplex.cf/weasel/network/{cloud_provider}")
	os.Setenv("vault_url", "https://dapis.cloudplex.cf/robin/api/v1")
}

/*func setAppConf(){
	iniconf, err := config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		beego.Error(err)
	}
	iniconf.Set("mongo_host","10.248.9.173")
	beego.Info(iniconf.String("appname"))
	beego.Info(iniconf.String("mongo_host"))
	beego.AppConfig.String("mongo_host")
	err = iniconf.SaveConfigFile("conf/app.conf")
	if err != nil {
		beego.Error(err)
	}
	beego.Info(iniconf.String("mongo_host"))
	beego.Info(iniconf.String("mongo_host"))
	beego.Info(beego.AppConfig.String("mongo_host"))
	beego.AppConfig.Set("mongo_host","10.248.9.173")
	beego.Info(beego.AppConfig.String("mongo_host"))
	/*beego.Info("going into sleep mode")
	time.Sleep(1*time.Minute)
    beego.Info("returing from set conf method")*/
//}
