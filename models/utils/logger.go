package utils

import (
	"encoding/json"
	"github.com/astaxie/beego"
)

type HeadersData struct {
	Key   string `bson:"key" json:"key"`
	Value string `bson:"value" json:"value"`
}

type Data struct {
	Message   string `json:"message" bson : "message"`
	ID        string `json:"id" bson : "id"`
	Type      string `json:"type" bson : "type"`
	Service   string `json:"service" bson : "service"`
	Level     string `json:"level" bson : "level"`
	CompanyId string `json:"company_id" bson : "company_id"`
}

func SendLog(companyId, msg, message_type, env_id string) (int, error) {

	var data Data

	data.ID = env_id
	data.Service = "antelope"
	data.Type = "project"
	data.Level = message_type
	data.Message = msg
	data.CompanyId = companyId

	logger := InitReq()

	request_data, err := TransformData(data)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	req, err := CreatePostRequest(request_data, getLoggerHost())
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	response, err := logger.SendRequest(req)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}
	return response.StatusCode, err

}

func TransformData(data interface{}) ([]byte, error) {

	request_data, err := json.Marshal(data)
	return request_data, err

}
func getLoggerHost() string {

	return "http://" + beego.AppConfig.String("logger_url") + "/elephant/api/v1/frontend/logging/"
	//return "https://dapis.cloudplex.cf/elephant/api/v1/frontend/logging/"
}
