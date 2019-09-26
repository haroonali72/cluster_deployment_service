package cores

import (
	"antelope/models"
	d_duck "bitbucket.org/cloudplex-devs/d-duck"
	"encoding/json"
	"github.com/astaxie/beego"
	"strings"
)

func GetCoresLimit(subscriptionId string) (int64, error) {
	subscriptionId = "88903349-acdc-4fa4-88e0-0a4763197feb"
	beego.Info("PORT:", beego.AppConfig.String("subscription_host"))
	s := strings.Split(beego.AppConfig.String("subscription_host"), ":")
	ip, port := string(s[0]), string(s[1])
	beego.Info("subscriptionId:", subscriptionId)
	beego.Info("IP:", ip)
	beego.Info("port:", port)

	subscriptionClient := d_duck.Init{Client: d_duck.Client{
		Host: "122.129.74.5",
		Port: "8080",
	}}

	limits, err := subscriptionClient.GetLimitsWithSubscriptionId(subscriptionId)
	if err != nil {
		beego.Error("subscription host not connected" + err.Error())
		return 0, err
	}

	coresLimit, err := json.MarshalIndent(limits, "", "  ")
	if err != nil {
		beego.Error("marshalling of cores limits failed" + err.Error())
		return 0, err
	}

	var limit models.Limits
	if err := json.Unmarshal(coresLimit, &limit); err != nil {
		beego.Error("Unmarshalling of cores limits failed ", err.Error())
		return 0, err
	}

	return int64(limit.CoreCount), nil
}
