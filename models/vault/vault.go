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
	KeyInfo interface{}  `json:"key_info"`
	KeyName string       `json:"key_name"`
	Cloud   models.Cloud `json:"cloud_type"`
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

func getVaultHost() string {
	return beego.AppConfig.String("vault_url") + models.VaultEndpoint
}
func PostSSHKey(keyRaw interface{}, keyName string, cloudType models.Cloud, ctx utils.Context, token, teams string) (int, error) {
	var keyObj Key

	keyObj.KeyInfo = keyRaw
	keyObj.Cloud = cloudType
	keyObj.KeyName = keyName

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
func GetSSHKey(cloudType, keyName, token string, ctx utils.Context) ([]byte, error) {

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
		return []byte{}, err
	}
	client := utils.InitReq()
	req.Header.Set("token", token)
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []byte{}, err
	}
	defer response.Body.Close()

	beego.Info(response.StatusCode)
	beego.Info(response.Status)
	if response.StatusCode == 403 {
		return []byte{}, errors.New("User is not authorized to use this key - " + keyName)
	}else if response.StatusCode == 404 {
		return []byte{}, errors.New("key not found")
	}
	if response.StatusCode != 200 {
		return []byte{}, errors.New("Status Code: " + strconv.Itoa(response.StatusCode))
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []byte{}, err
	}
	return contents, nil

}
func GetAllSSHKey(cloudType string, ctx utils.Context, token string) (interface{}, error) {
	var keys interface{}
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
	if response.StatusCode == 403 {
		return []byte{}, errors.New("User is not authorized for credential profile - " + profileId)
	}else if response.StatusCode == 404 {
		return []byte{}, errors.New("profile not found")
	}

	if response.StatusCode != 200 {
		return []byte{}, errors.New("profile not found")
	}

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []byte{}, err
	}
	return contents, nil

}

func DeleteSSHkey(cloudType, keyName, token string, ctx utils.Context) error {
	host := getVaultHost() + models.VaultDeleteKeyURI
	if strings.Contains(host, "{cloudType}") {
		host = strings.Replace(host, "{cloudType}", cloudType, -1)
	}

	if strings.Contains(host, "{name}") {
		host = strings.Replace(host, "{name}", keyName, -1)
	}

	req, err := utils.CreateDeleteRequest(host)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	client := utils.InitReq()
	req.Header.Set("token", token)
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer response.Body.Close()

	beego.Info(response.StatusCode)
	beego.Info(response.Status)
	if response.StatusCode == 403 {
		return  errors.New("User is not authorized to delete this key - " + keyName)
	}else if response.StatusCode == 404{
		return errors.New("key not found")
	}
	if response.StatusCode != 200 {
		return errors.New("Status Code: " + strconv.Itoa(response.StatusCode))
	}

	return nil

}
