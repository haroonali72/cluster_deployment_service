package utils

import (
	"antelope/models/logging"
	"errors"
	"io/ioutil"
)

func GetAPIStatus(host string, ctx logging.Context) (interface{}, error) {

	client := InitReq()

	req, err := CreateGetRequest(host)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return nil, err
	}

	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return nil, err
	}

	if response.StatusCode == 404 {
		ctx.SendSDLog("no network exists for this project id", "error")
		return nil, errors.New("no network exists for this project id")
	}
	defer response.Body.Close()
	//	var network AzureNetwork
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return nil, err
	}
	return contents, nil

}
