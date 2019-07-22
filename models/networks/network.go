package networks

import (
	"antelope/models/logging"
	"antelope/models/utils"
	"io/ioutil"
)

func GetAPIStatus(host string, ctx logging.Context) (interface{}, error) {

	client := utils.InitReq()

	req, err := utils.CreateGetRequest(host)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return nil, err
	}

	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return nil, err
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
