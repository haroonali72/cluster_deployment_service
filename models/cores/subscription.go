package cores

import (
	"antelope/models"
	d_duck "bitbucket.org/cloudplex-devs/d-duck"
	"encoding/json"
	"github.com/astaxie/beego"
	"strings"
)

func GetCoresLimit(subscriptionId string) (int64, error) {

	s := strings.Split(beego.AppConfig.String("subscription_host"), ":")
	ip, port := string(s[0]), string(s[1])
	user := beego.AppConfig.String("kill_bill_user")
	password := beego.AppConfig.String("kill_bill_password")
	apiKey := beego.AppConfig.String("kill_bill_api_key")
	apiSecret := beego.AppConfig.String("kill_bill_secret_key")

	subscriptionClient := d_duck.Init{Client: d_duck.Client{
		Host:      ip,
		Port:      port,
		Username:  user,
		Password:  password,
		ApiKey:    apiKey,
		ApiSecret: apiSecret,
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
