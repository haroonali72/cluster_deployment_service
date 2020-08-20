package utils

import (
	"antelope/models"
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
	data.Type = "infrastructure"
	data.Level = message_type
	data.Message = msg
	data.CompanyId = companyId

	logger := InitReq() //returns httpclient

	request_data, err := TransformData(data) //transforms data to json
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	req, err := CreatePostRequest(request_data, getLoggerHost()) // req is generated
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}
	m := make(map[string]string)
	m["Content-Type"] = "application/json"
	SetHeaders(req, m)
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
	return beego.AppConfig.String("logger_url") + models.LoggingEndpoint + models.FrontEndLoggingURI
}
