package vault

import (
	"antelope/models"
	"antelope/models/logging"
	"antelope/models/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"io/ioutil"
)

type Key struct {
	KeyInfo interface{} `json:"key_info"`
	KeyName string      `json:"key_name"`
	Cloud   string      `json:"cloud_type"`
}
type awsKey struct {
	KeyName     string         `json:"key_name" bson:"key_name"`
	KeyType     models.KeyType `json:"key_type" bson:"key_type"`
	KeyMaterial string         `json:"private_key" bson:"private_key"`
	Cloud       models.Cloud   `json:"cloud" bson:"cloud"`
}
type azureKey struct {
	CredentialType models.CredentialsType `json:"credential_type"  bson:"credential_type"`
	NewKey         models.KeyType         `json:"key_type"  bson:"key_type"`
	KeyName        string                 `json:"key_name" bson:"key_name"`
	AdminPassword  string                 `json:"admin_password" bson:"admin_password",omitempty"`
	PrivateKey     string                 `json:"private_key" bson:"private_key",omitempty"`
	PublicKey      string                 `json:"public_key" bson:"public_key",omitempty"`
	Cloud          models.Cloud           `json:"cloud" bson:"cloud"`
}

func GetSSHKey(cloudType string, keyName string) (interface{}, error) {

	req, err := utils.CreateGetRequest(getVaultHost() + "/template/sshKey/" + cloudType + "/" + keyName)
	if err != nil {
		beego.Error("%s", err)
		return awsKey{}, err
	}
	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		beego.Error("%s", err)
		return awsKey{}, err
	}
	defer response.Body.Close()

	var key awsKey
	beego.Info(response.StatusCode)
	beego.Info(response.Status)
	if response.StatusCode == 500 {
		return awsKey{}, errors.New("not found")
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		beego.Error("%s", err)
		return awsKey{}, err
	}

	err = json.Unmarshal(contents, &key)
	if err != nil {
		beego.Error("%s", err)
		return awsKey{}, err
	}
	return key, nil

}
func getVaultHost() string {
	return beego.AppConfig.String("vault_url")
}
func PostSSHKey(keyRaw interface{}) (int, error) {

	b, e := json.Marshal(keyRaw)
	if e != nil {
		beego.Error(e.Error())
		return 400, e
	}
	var key awsKey
	e = json.Unmarshal(b, &key)
	if e != nil {
		beego.Error(e.Error())
		return 400, e
	}
	key.Cloud = "aws"

	var keyObj Key
	keyObj.KeyInfo = key
	keyObj.Cloud = "aws"
	keyObj.KeyName = key.KeyName
	client := utils.InitReq()
	fmt.Print("+%v", key)
	fmt.Print("+%v", keyObj)
	request_data, err := logging.TransformData(keyObj)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	req, err := utils.CreatePostRequest(request_data, getVaultHost()+"/template/sshKey/")
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	response, err := client.SendRequest(req)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	beego.Error(response.StatusCode)
	if response.StatusCode == 500 {
		return 0, errors.New("error in saving key")
	}
	return response.StatusCode, err

}
func PostAzureSSHKey(keyRaw interface{}) (int, error) {
	b, e := json.Marshal(keyRaw)
	if e != nil {
		beego.Error(e.Error())
		return 400, e
	}
	var key azureKey
	e = json.Unmarshal(b, &key)
	if e != nil {
		beego.Error(e.Error())
		return 400, e
	}
	key.Cloud = "azure"

	var keyObj Key
	keyObj.KeyInfo = key
	keyObj.Cloud = "azure"
	keyObj.KeyName = key.KeyName

	client := utils.InitReq()

	request_data, err := logging.TransformData(keyObj)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	req, err := utils.CreatePostRequest(request_data, getVaultHost()+"/template/sshKey/")
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	response, err := client.SendRequest(req)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}
	if response.StatusCode == 500 {
		return 0, errors.New("error in saving key")
	}
	return response.StatusCode, err

}
func GetAzureSSHKey(cloudType string, keyName string) (interface{}, error) {

	req, err := utils.CreateGetRequest(getVaultHost() + "/template/sshKey/" + cloudType + "/" + keyName)
	if err != nil {
		beego.Error("%s", err)
		return azureKey{}, err
	}
	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		beego.Error("%s", err)
		return azureKey{}, err
	}
	defer response.Body.Close()

	var key azureKey
	beego.Info(response.StatusCode)
	beego.Info(response.Status)
	if response.StatusCode == 500 {
		return azureKey{}, errors.New("not found")
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		beego.Error("%s", err)
		return azureKey{}, err
	}

	err = json.Unmarshal(contents, &key)
	if err != nil {
		beego.Error("%s", err)
		return azureKey{}, err
	}
	return key, nil

}
func GetAllSSHKey(cloudType string) ([]string, error) {
	var keys []string
	req, err := utils.CreateGetRequest(getVaultHost() + "/template/sshKey/" + cloudType)
	if err != nil {
		beego.Error("%s", err)
		return keys, err
	}
	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		beego.Error("%s", err)
		return keys, err
	}
	defer response.Body.Close()

	beego.Info(response.StatusCode)
	beego.Info(response.Status)
	if response.StatusCode == 500 {
		return keys, errors.New("not found")
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		beego.Error("%s", err)
		return keys, err
	}

	err = json.Unmarshal(contents, &keys)
	if err != nil {
		beego.Error("%s", err)
		return keys, err
	}
	return keys, nil

}
