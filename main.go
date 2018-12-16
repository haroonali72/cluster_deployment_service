package main

import (
	"github.com/astaxie/beego/plugins/cors"
	"os"
	"antelope/models/db"
	_ "antelope/routers"

	"github.com/astaxie/beego"
)

func SecretAuth(username, password string) bool {
	// TODO configure basic authentication properly
	return username == "username" && password == "password"
}

func main() {
	if !db.IsMongoAlive() {
		os.Exit(1)
	}

	beego.BConfig.WebConfig.DirectoryIndex = true
	beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"

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
