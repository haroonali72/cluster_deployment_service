package vault

import (
	"antelope/models"
	"antelope/models/logging"
	"antelope/models/utils"
	"encoding/json"
	"github.com/astaxie/beego"
	"io/ioutil"
)

type Key struct {
	keyInfo interface{} `json:"key_info"`
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

func GetSSHKey(cloudType string, keyName string) (awsKey, error) {

	req, err := utils.CreateGetRequest(getVaultHost() + "template/sshkey/" + cloudType + "/" + keyName)
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
	keyObj.keyInfo = key
	keyObj.Cloud = "aws"
	keyObj.KeyName = key.KeyName
	client := utils.InitReq()

	request_data, err := logging.TransformData(keyObj)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	req, err := utils.CreatePostRequest(request_data, getVaultHost()+"template/sshkey/")
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
	key.Cloud = "aws"
	key.Cloud = "azure"

	var keyObj Key
	keyObj.keyInfo = key
	keyObj.Cloud = "azure"
	keyObj.KeyName = key.KeyName

	client := utils.InitReq()

	request_data, err := logging.TransformData(keyObj)
	if err != nil {
		beego.Error("%s", err)
		return 400, err
	}

	req, err := utils.CreatePostRequest(request_data, getVaultHost()+"template/sshkey/")
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
func GetAzureSSHKey(cloudType string, keyName string) (azureKey, error) {

	req, err := utils.CreateGetRequest(getVaultHost() + "template/sshkey/" + cloudType + "/" + keyName)
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
