package api_handler

import (
	"antelope/models/utils"
	"errors"
	"io/ioutil"
)

func GetAPIStatus(token, host string, ctx utils.Context) (interface{}, error) {

	client := utils.InitReq()

	req, err := utils.CreateGetRequest(host)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return nil, err
	}
	req.Header.Add("token", token)
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return nil, err
	}

	if response.StatusCode == 404 {
		ctx.SendSDLog("no entity exists for this project id", "error")
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
