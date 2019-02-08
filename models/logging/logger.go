package logging

import (
	"antelope/models/utils"
	"encoding/json"
	"github.com/astaxie/beego"
)

type HeadersData struct {
	Key   string `bson:"key" json:"key"`
	Value string `bson:"value" json:"value"`
}

type Data struct {
	Message     string `json:"message" bson : "message"`
	ID          string `json:"id" bson : "id"`
	Environment string `json:"environment" bson : "environment"`
	Service     string `json:"service" bson : "service"`
	Level       string `json:"level" bson : "level"`
}

func SendLog(msg, message_type, env_id string) (int, error) {

	var data Data

	data.ID = env_id
	data.Service = "antelope"
	data.Environment = "environment"
	data.Level = message_type
	data.Message = msg

	logger := utils.InitReq()

	request_data, err := transformData(data)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	req, err := utils.CreatePostRequest(request_data, getLoggerHost())
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

func transformData(data interface{}) ([]byte, error) {

	request_data, err := json.Marshal(data)
	return request_data, err

}
func getLoggerHost() string {
	return beego.AppConfig.String("logger_url")
}
