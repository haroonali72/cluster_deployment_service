//package main
//
//import (
//	"antelope/models/db"
//	"antelope/models/utils"
//	_ "antelope/routers"
//	"github.com/astaxie/beego"
//	"github.com/astaxie/beego/plugins/cors"
//	"os"
//)
//
//func SecretAuth(username, password string) bool {
//	// TODO configure basic authentication properly
//	return username == "username" && password == "password"
//}
//
//func main() {
//	setEnv()
//	utils.InitFlags()
//	if !db.IsMongoAlive() {
//		os.Exit(1)
//	}
//
//	beego.BConfig.WebConfig.DirectoryIndex = true
//	beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
//
//	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
//		AllowOrigins:     []string{"*"},
//		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE", "OPTIONS"},
//		AllowHeaders:     []string{"Origin", "Authorization", "Token", "Content-type", "Accept"},
//		ExposeHeaders:    []string{"Content-Length"},
//		AllowCredentials: true,
//	}))
//
//	// TODO enable basic authentication if required
//	//authPlugin := auth.NewBasicAuthenticator(SecretAuth, "Authorization Required")
//	//beego.InsertFilter("*", beego.BeforeRouter, authPlugin)
//
//	beego.Run()
//}
//func setEnv() {
//	os.Setenv("mongo_host", "10.248.9.173")
//	os.Setenv("mongo_auth", "true")
//	os.Setenv("mongo_db", "antelope")
//	os.Setenv("mongo_user", "antelope")
//	os.Setenv("mongo_pass", "deltapsi@#22237")
//	os.Setenv("mongo_aws_template_collection", "aws_template")
//	os.Setenv("mongo_aws_cluster_collection", "aws_cluster")
//	os.Setenv("mongo_azure_template_collection", "azure_template")
//	os.Setenv("mongo_azure_cluster_collection", "azure_cluster")
//	os.Setenv("redis_url", "10.248.9.173")
//	os.Setenv("logger_url", "10.248.9.173")
//	os.Setenv("network_url", "10.248.9.173")
//}
//
///*func setAppConf(){
//	iniconf, err := config.NewConfig("ini", "conf/app.conf")
//	if err != nil {
//		beego.Error(err)
//	}
//	iniconf.Set("mongo_host","10.248.9.173")
//	beego.Info(iniconf.String("appname"))
//	beego.Info(iniconf.String("mongo_host"))
//	beego.AppConfig.String("mongo_host")
//	err = iniconf.SaveConfigFile("conf/app.conf")
//	if err != nil {
//		beego.Error(err)
//	}
//	beego.Info(iniconf.String("mongo_host"))
//	beego.Info(iniconf.String("mongo_host"))
//	beego.Info(beego.AppConfig.String("mongo_host"))
//	beego.AppConfig.Set("mongo_host","10.248.9.173")
//	beego.Info(beego.AppConfig.String("mongo_host"))
//	/*beego.Info("going into sleep mode")
//	time.Sleep(1*time.Minute)
//    beego.Info("returing from set conf method")*/
////}
package main

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/go-redis/redis"
)

/*var (
	redisHost    = beego.AppConfig.String("redis_url")
)*/

type Notifier struct {
	Client    *redis.Client
	redisHost string
}
type Response struct {
	Status    string `json:"status"`
	ID        string `json:"_id"`
	Component string `json:"component"`
}

func (notifier *Notifier) Notify(channel, status string) {
	msg := Response{
		Status:    status,
		ID:        channel,
		Component: "Cluster",
	}
	cmd := notifier.Client.Publish(channel, msg)
	beego.Info(*cmd)
}

func (notifier *Notifier) Init_notifier() error {
	if notifier.Client != nil {
		return nil
	}
	notifier.redisHost = "10.248.9.173:6379"
	options := redis.Options{}
	options.Addr = notifier.redisHost
	notifier.Client = redis.NewClient(&options)

	return nil
}
func main() {
	var client *redis.Client
	opt := redis.Options{}
	opt.Addr = "10.248.9.173:6379"
	client = redis.NewClient(&opt)
	msg := Response{
		Status:    "yes",
		ID:        "savgdaf",
		Component: "Cluster",
	}
	b, _ := json.Marshal(msg)
	client.Publish("sadaf", string(b))
	//fmt.Print(cmd)
}
