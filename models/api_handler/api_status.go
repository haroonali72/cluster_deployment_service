package api_handler

import (
	"antelope/models"
	"antelope/models/utils"
	"errors"
	"io/ioutil"
)

func GetAPIStatus(token, host string, ctx utils.Context) (interface{}, error) {

	client := utils.InitReq()

	req, err := utils.CreateGetRequest(host)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	req.Header.Add("token", token)
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	if response.StatusCode != 200 {
		ctx.SendLogs("data not fetched", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, errors.New("data not fetched")
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	return contents, nil

}
