package vault

import (
	"antelope/models"
	"antelope/models/utils"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	"io/ioutil"
	"strconv"
	"strings"
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

	host := getVaultHost() + models.VaultGetKeyURI

	if strings.Contains(host, "{cloud}") {
		host = strings.Replace(host, "{cloud}", cloudType, -1)
	}

	if strings.Contains(host, "{keyName}") {
		host = strings.Replace(host, "{keyName}", keyName, -1)
	}
	req, err := utils.CreateGetRequest(host)
	if err != nil {

		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return awsKey{}, err
	}
	req.Header.Set("token", token)

	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {

		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return awsKey{}, err
	}
	defer response.Body.Close()

	var key awsKey
	//	beego.Info(response.StatusCode)
	beego.Info(response.Status)
	//if response.StatusCode == 500 || response.StatusCode == 404 {
	//	return awsKey{}, errors.New("not found")
	//}
	if response.StatusCode == 404 || response.StatusCode == 403 {

		return awsKey{}, errors.New("not found")
	}
	if response.StatusCode != 200 {
		return awsKey{}, errors.New("Error : " + response.Status)
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {

		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return awsKey{}, err
	}

	err = json.Unmarshal(contents, &key)
	if err != nil {

		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return awsKey{}, err
	}
	return key, nil

}
func getVaultHost() string {
	return beego.AppConfig.String("vault_url") + models.VaultEndpoint
}
func PostSSHKey(keyRaw interface{}, ctx utils.Context, token string) (int, error) {

	b, e := json.Marshal(keyRaw)
	if e != nil {

		ctx.SendLogs(e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 400, e
	}
	var key awsKey
	e = json.Unmarshal(b, &key)
	if e != nil {

		ctx.SendLogs(e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
		ctx.SendLogs(e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 400, err
	}

	req, err := utils.CreatePostRequest(request_data, getVaultHost()+models.VaultCreateKeyURI)
	if err != nil {
		ctx.SendLogs(e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 400, err
	}
	req.Header.Set("token", token)

	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendLogs(e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 400, err
	}

	beego.Error(response.StatusCode)
	if response.StatusCode == 500 {
		return 0, errors.New("error in saving key")
	}
	return response.StatusCode, err

}

func PostAzureSSHKey(cloud models.Cloud, keyRaw interface{}, ctx utils.Context, token, teams string) (int, error) {
	b, e := json.Marshal(keyRaw)
	if e != nil {
		ctx.SendLogs(e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 400, e
	}
	var key azureKey
	e = json.Unmarshal(b, &key)
	if e != nil {
		ctx.SendLogs(e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 400, e
	}
	key.Cloud = cloud

	var keyObj Key
	keyObj.KeyInfo = key
	keyObj.Cloud = string(cloud)
	keyObj.KeyName = key.KeyName

	client := utils.InitReq()

	request_data, err := utils.TransformData(keyObj)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 400, err
	}

	req, err := utils.CreatePostRequest(request_data, getVaultHost()+models.VaultCreateKeyURI)
	if err != nil {

		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 400, err
	}
	req.Header.Set("token", token)
	req.Header.Set("teams", teams)
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 400, err
	}
	if response.StatusCode == 500 {
		return 0, errors.New("error in saving key")
	}
	return response.StatusCode, err

}
func GetAzureSSHKey(cloudType, keyName, token string, ctx utils.Context) (interface{}, error) {

	host := getVaultHost() + models.VaultGetKeyURI

	if strings.Contains(host, "{cloud}") {
		host = strings.Replace(host, "{cloud}", cloudType, -1)
	}

	if strings.Contains(host, "{keyName}") {
		host = strings.Replace(host, "{keyName}", keyName, -1)
	}
	req, err := utils.CreateGetRequest(host)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return azureKey{}, err
	}
	client := utils.InitReq()
	req.Header.Set("token", token)
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return azureKey{}, err
	}
	defer response.Body.Close()

	var key azureKey
	beego.Info(response.StatusCode)
	beego.Info(response.Status)
	if response.StatusCode == 403 || response.StatusCode == 404 {
		return azureKey{}, errors.New("not found")
	}
	if response.StatusCode != 200 {
		return azureKey{}, errors.New("Status Code: " + strconv.Itoa(response.StatusCode))
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return azureKey{}, err
	}
	err = json.Unmarshal(contents, &key)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return azureKey{}, err
	}
	return key, nil

}
func GetAllSSHKey(cloudType string, ctx utils.Context, token string) ([]string, error) {
	var keys []string
	host := getVaultHost() + models.VaultGetAllKeysURI

	if strings.Contains(host, "{cloud}") {
		host = strings.Replace(host, "{cloud}", cloudType, -1)
	}
	req, err := utils.CreateGetRequest(host)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return keys, err
	}
	client := utils.InitReq()
	req.Header.Set("token", token)
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return keys, err
	}
	defer response.Body.Close()

	beego.Info(response.StatusCode)
	beego.Info(response.Status)
	if response.StatusCode == 500 {
		return keys, errors.New("not found")
	}
	if response.StatusCode != 200 {
		return keys, errors.New("Status Code : " + strconv.Itoa(response.StatusCode))
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return keys, err
	}

	err = json.Unmarshal(contents, &keys)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return keys, err
	}
	return keys, nil

}
func GetCredentialProfile(cloudType string, profileId string, token string, ctx utils.Context) ([]byte, error) {
	host := getVaultHost() + models.VaultGetProfileURI

	if strings.Contains(host, "{cloud}") {
		host = strings.Replace(host, "{cloud}", cloudType, -1)
	}

	if strings.Contains(host, "{profileId}") {
		host = strings.Replace(host, "{profileId}", profileId, -1)
	}
	req, err := utils.CreateGetRequest(host)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []byte{}, err
	}
	req.Header.Add("token", token)
	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []byte{}, err
	}
	defer response.Body.Close()

	beego.Info(response.StatusCode)
	beego.Info(response.Status)
	if response.StatusCode != 200 {
		return []byte{}, errors.New("not found")
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []byte{}, err
	}
	return contents, nil

}
