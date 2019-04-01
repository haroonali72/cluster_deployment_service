package vault

import (
	"antelope/models/aws"
	"antelope/models/azure"
	"antelope/models/logging"
	"antelope/models/utils"
	"encoding/json"
	"github.com/astaxie/beego"
	"io/ioutil"
)

func GetSSHKey(cloudType string, keyName string) (aws.Key, error) {

	req, err := utils.CreateGetRequest(getVaultHost() + "/" + cloudType + "/" + keyName)
	if err != nil {
		beego.Error("%s", err)
		return aws.Key{}, err
	}
	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		beego.Error("%s", err)
		return aws.Key{}, err
	}
	defer response.Body.Close()

	var key aws.Key

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		beego.Error("%s", err)
		return aws.Key{}, err
	}

	err = json.Unmarshal(contents, &key)
	if err != nil {
		beego.Error("%s", err)
		return aws.Key{}, err
	}
	return key, nil

}
func getVaultHost() string {
	return beego.AppConfig.String("vault_url")
}
func PostSSHKey(key aws.Key) (int, error) {

	key.Cloud = "aws"
	client := utils.InitReq()

	request_data, err := logging.TransformData(key)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	req, err := utils.CreatePostRequest(request_data, getVaultHost())
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	response, err := client.SendRequest(req)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}
	return response.StatusCode, err

}
func PostAzureSSHKey(key azure.Key) (int, error) {

	key.Cloud = "aws"
	client := utils.InitReq()

	request_data, err := logging.TransformData(key)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	req, err := utils.CreatePostRequest(request_data, getVaultHost())
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	response, err := client.SendRequest(req)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}
	return response.StatusCode, err

}
func GetAzureSSHKey(cloudType string, keyName string) (azure.Key, error) {

	req, err := utils.CreateGetRequest(getVaultHost() + "/" + cloudType + "/" + keyName)
	if err != nil {
		beego.Error("%s", err)
		return azure.Key{}, err
	}
	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		beego.Error("%s", err)
		return azure.Key{}, err
	}
	defer response.Body.Close()

	var key azure.Key

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		beego.Error("%s", err)
		return azure.Key{}, err
	}

	err = json.Unmarshal(contents, &key)
	if err != nil {
		beego.Error("%s", err)
		return azure.Key{}, err
	}
	return key, nil

}
