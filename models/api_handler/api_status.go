package api_handler

import (
	"antelope/constants"
	"antelope/models/utils"
	"errors"
	"io/ioutil"
)

func GetAPIStatus(host string, ctx utils.Context) (interface{}, error) {

	client := utils.InitReq()

	req, err := utils.CreateGetRequest(host)
	if err != nil {
		logType := []string{"backend-logging"}
		ctx.SendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR, logType)
		return nil, err
	}

	response, err := client.SendRequest(req)
	if err != nil {
		logType := []string{"backend-logging"}
		ctx.SendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR, logType)
		return nil, err
	}

	if response.StatusCode == 404 {
		logType := []string{"backend-logging"}
		ctx.SendLogs("no network exists for this project id", constants.LOGGING_LEVEL_ERROR, logType)
		return nil, errors.New("no network exists for this project id")
	}
	defer response.Body.Close()
	//	var network AzureNetwork
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logType := []string{"backend-logging"}
		ctx.SendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR, logType)
		return nil, err
	}
	return contents, nil

}
