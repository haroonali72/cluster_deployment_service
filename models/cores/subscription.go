package cores

import (
	d_duck "antelope/d-duck"
	"antelope/models"
	"encoding/json"
	"github.com/astaxie/beego"
	"strings"
)

func GetCoresLimit(subscriptionId string) (int64, error) {
	s := strings.Split(beego.AppConfig.String("subscription_host"), ":")
	ip, port := string(s[0]), string(s[1])
	beego.Info("ip", ip)
	beego.Info("port", port)
	subscriptionClient := d_duck.Init{Client: d_duck.Client{
		Host: ip,
		Port: port,
	}}

	limits, err := subscriptionClient.GetLimitsWithSubscriptionId(subscriptionId)
	if err != nil {
		beego.Error(err.Error())
		return 0, err
	}

	coresLimit, err := json.MarshalIndent(limits, "", "  ")
	if err != nil {
		beego.Error(err.Error())
		return 0, err
	}

	var limit models.Limits
	if err := json.Unmarshal(coresLimit, &limit); err != nil {
		beego.Error("Unmarshalling of cores limits failed ", err.Error())
		return 0, err
	}

	return int64(limit.CoreCount), nil
}
