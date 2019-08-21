package vault

import (
	"antelope/models"
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

type AzureProfile struct {
	Profile AzureCredentials `json:"credentials"`
}
type AzureCredentials struct {
	ClientId       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	SubscriptionId string `json:"subscription_id"`
	TenantId       string `json:"tenant_id"`
	Location       string `json:"region"`
}
type AwsProfile struct {
	Profile AwsCredentials `json:"credentials"`
}
type AwsCredentials struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"access_secret"`
	Region    string `json:"region"`
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

func GetSSHKey(cloudType string, keyName string, ctx utils.Context, token string) (interface{}, error) {

	req, err := utils.CreateGetRequest(getVaultHost() + "/template/sshKey/" + cloudType + "/" + keyName)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return awsKey{}, err
	}
	req.Header.Set("token", token)
	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return awsKey{}, err
	}
	defer response.Body.Close()

	var key awsKey
	beego.Info(response.StatusCode)
	beego.Info(response.Status)
	if response.StatusCode == 500 || response.StatusCode == 404 {
		return awsKey{}, errors.New("not found")
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return awsKey{}, err
	}

	err = json.Unmarshal(contents, &key)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return awsKey{}, err
	}
	return key, nil

}
func getVaultHost() string {
	return beego.AppConfig.String("vault_url")
}
func PostSSHKey(keyRaw interface{}, ctx utils.Context, token string) (int, error) {

	b, e := json.Marshal(keyRaw)
	if e != nil {
		ctx.SendSDLog(e.Error(), "error")
		return 400, e
	}
	var key awsKey
	e = json.Unmarshal(b, &key)
	if e != nil {
		ctx.SendSDLog(e.Error(), "error")
		return 400, e
	}
	key.Cloud = "aws"

	var keyObj Key
	keyObj.KeyInfo = key
	keyObj.Cloud = "aws"
	keyObj.KeyName = key.KeyName
	client := utils.InitReq()
	request_data, err := utils.TransformData(keyObj)
	if err != nil {
		ctx.SendSDLog(e.Error(), "error")
		return 400, err
	}

	req, err := utils.CreatePostRequest(request_data, getVaultHost()+"/template/sshKey/")
	if err != nil {
		ctx.SendSDLog(e.Error(), "error")
		return 400, err
	}
	req.Header.Set("token", token)

	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(e.Error(), "error")
		return 400, err
	}

	beego.Error(response.StatusCode)
	if response.StatusCode == 500 {
		return 0, errors.New("error in saving key")
	}
	return response.StatusCode, err

}
func PostAzureSSHKey(keyRaw interface{}, ctx utils.Context, token string) (int, error) {
	b, e := json.Marshal(keyRaw)
	if e != nil {
		ctx.SendSDLog(e.Error(), "error")
		return 400, e
	}
	var key azureKey
	e = json.Unmarshal(b, &key)
	if e != nil {
		ctx.SendSDLog(e.Error(), "error")
		return 400, e
	}
	key.Cloud = "azure"

	var keyObj Key
	keyObj.KeyInfo = key
	keyObj.Cloud = "azure"
	keyObj.KeyName = key.KeyName

	client := utils.InitReq()

	request_data, err := utils.TransformData(keyObj)
	if err != nil {
		ctx.SendSDLog(e.Error(), "error")
		return 400, err
	}

	req, err := utils.CreatePostRequest(request_data, getVaultHost()+"/template/sshKey/")
	if err != nil {
		ctx.SendSDLog(e.Error(), "error")
		return 400, err
	}
	req.Header.Set("token", token)
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(e.Error(), "error")
		return 400, err
	}
	if response.StatusCode == 500 {
		return 0, errors.New("error in saving key")
	}
	return response.StatusCode, err

}
func PostGcpSSHKey(keyRaw interface{}, ctx utils.Context, token string) (int, error) {
	b, e := json.Marshal(keyRaw)
	if e != nil {
		ctx.SendSDLog(e.Error(), "error")
		return 400, e
	}
	var key azureKey
	e = json.Unmarshal(b, &key)
	if e != nil {
		ctx.SendSDLog(e.Error(), "error")
		return 400, e
	}
	key.Cloud = models.GCP

	var keyObj Key
	keyObj.KeyInfo = key
	keyObj.Cloud = string(models.GCP)
	keyObj.KeyName = key.KeyName

	client := utils.InitReq()

	request_data, err := utils.TransformData(keyObj)
	if err != nil {
		ctx.SendSDLog(e.Error(), "error")
		return 400, err
	}

	req, err := utils.CreatePostRequest(request_data, getVaultHost()+"/template/sshKey/")
	if err != nil {
		ctx.SendSDLog(e.Error(), "error")
		return 400, err
	}
	req.Header.Set("token", token)

	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(e.Error(), "error")
		return 400, err
	}
	if response.StatusCode == 500 {
		return 0, errors.New("error in saving key")
	}
	return response.StatusCode, err

}
func GetAzureSSHKey(cloudType string, keyName string, ctx utils.Context) (interface{}, error) {

	fmt.Print(getVaultHost() + "/template/sshKey/" + cloudType + "/" + keyName)
	req, err := utils.CreateGetRequest(getVaultHost() + "/template/sshKey/" + cloudType + "/" + keyName)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return azureKey{}, err
	}
	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
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
		ctx.SendSDLog(err.Error(), "error")
		return azureKey{}, err
	}

	err = json.Unmarshal(contents, &key)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return azureKey{}, err
	}
	return key, nil

}
func GetAllSSHKey(cloudType string, ctx utils.Context, token string) ([]string, error) {
	var keys []string
	req, err := utils.CreateGetRequest(getVaultHost() + "/template/sshKey/" + cloudType)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return keys, err
	}
	client := utils.InitReq()
	req.Header.Set("token", token)
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")

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
		ctx.SendSDLog(err.Error(), "error")

		return keys, err
	}

	err = json.Unmarshal(contents, &keys)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")

		return keys, err
	}
	return keys, nil

}
func GetCredentialProfile(cloudType string, profileId string, token string, ctx utils.Context) ([]byte, error) {

	req, err := utils.CreateGetRequest(getVaultHost() + "/template/" + cloudType + "/credentials/" + profileId)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return []byte{}, err
	}
	req.Header.Add("token", token)
	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return []byte{}, err
	}
	defer response.Body.Close()

	beego.Info(response.StatusCode)
	beego.Info(response.Status)
	if response.StatusCode == 500 || response.StatusCode == 404 {
		return []byte{}, errors.New("not found")
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return []byte{}, err
	}
	return contents, nil

}
