package cores

import (
	d_duck "antelope/d-duck"
	"antelope/models"
	"encoding/json"
	"github.com/astaxie/beego"
)

func GetCoresLimit(subscriptionId string) (int64, error) {

	subscriptionClient := d_duck.Init{Client: d_duck.Client{
		Host: beego.AppConfig.String("Host"),
		Port: beego.AppConfig.String("Port"),
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
