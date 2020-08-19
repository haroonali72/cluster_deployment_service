package vault

import (
	"antelope/models"
	"antelope/models/types"
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
	Region  string       `json:"region"`
}

type AzureProfile struct {
	Profile AzureCredentials `json:"credentials" validate:"required,dive" description:"AzureCredentials [required]"`
}

type AzureCredentials struct {
	ClientId       string `json:"client_id" validate:"required" description:"Client Id [required]"`
	ClientSecret   string `json:"client_secret" validate:"required" description:"Client secret key [required]"`
	SubscriptionId string `json:"subscription_id" validate:"required" description:"SubscriptionId of azure account [required]`
	TenantId       string `json:"tenant_id" validate:"required" description:"TenantId of azure account [required]`
	Location       string `json:"region" description:"Cloud location [optional]`
}

type AwsProfile struct {
	Profile AwsCredentials `json:"credentials"`
}
type AwsCredentials struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"access_secret"`
	Region    string `json:"region"`
}
type DOProfile struct {
	Profile DOCredentials `json:"credentials" validate:"required,dive" description:"DO Credentials [required]`
}
type DOCredentials struct {
	AccessKey string `json:"access_token" validate:"required" description:"Access key [required]"`
	Region    string `json:"region"  description:"Cloud Region [optional]"`
}
type IBMProfile struct {
	Profile IBMCredentials `json:"credentials" validate:"required,dive" description:"IBM Credentials [required]"`
}
type IBMCredentials struct {
	IAMKey string `json:"iam_key" validate:"required" description:"Cluster IAM key [required]"`
	Region string `json:"region"  description:"Cloud region [optional]"`
}

func getVaultHost() string {
	return beego.AppConfig.String("vault_url") + models.VaultEndpoint
}
func PostSSHKey(keyRaw interface{}, keyName string, cloudType models.Cloud, ctx utils.Context, token, teams, region string) (int, error) {
	var keyObj Key

	keyObj.KeyInfo = keyRaw
	keyObj.Cloud = cloudType
	keyObj.KeyName = keyName
	keyObj.Region = region

	client := utils.InitReq()

	request_data, err := utils.TransformData(keyObj)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 400, err
	}

	ctx.SendLogs(ctx.ReqRespData(types.ReqResPayload{
		Token:   token,
		Url:     getVaultHost() + models.VaultCreateKeyURI,
		ReqType: types.POST,
		ReqBody: string(request_data),
	}), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	req, err := utils.CreatePostRequest(request_data, getVaultHost()+models.VaultCreateKeyURI)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 500, err
	}
	m := make(map[string]string)

	m["Content-Type"] = "application/json"
	m["X-Auth-Token"] = token
	m["teams"] = teams
	utils.SetHeaders(req, m)
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 500, err
	}
	if response.StatusCode != 201 {
		return response.StatusCode, errors.New("Error in saving key")
	}
	return response.StatusCode, err

}
func GetSSHKey(cloudType, keyName, token string, ctx utils.Context, region string) ([]byte, error) {

	host := getVaultHost() + models.VaultGetKeyURI

	if strings.Contains(host, "{cloud}") {
		host = strings.Replace(host, "{cloud}", cloudType, -1)
	}

	if region != "" {
		if strings.Contains(host, "{region}") {
			host = strings.Replace(host, "{region}", region, -1)
		}
	}

	if strings.Contains(host, "{keyName}") {
		host = strings.Replace(host, "{keyName}", keyName, -1)
	}

	ctx.SendLogs(ctx.ReqRespData(types.ReqResPayload{
		Token:   token,
		Url:     host,
		ReqType: types.GET,
	}), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	req, err := utils.CreateGetRequest(host)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []byte{}, err
	}
	client := utils.InitReq()
	req.Header.Set("X-Auth-Token", token)
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
	} else if response.StatusCode == 404 {
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

	ctx.SendLogs(ctx.ReqRespData(types.ReqResPayload{
		Token:   token,
		Url:     host,
		ReqType: types.GET,
		Resp:    string(contents),
	}), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	return contents, nil

}
func GetAllSSHKey(cloudType string, ctx utils.Context, token, region string) (interface{}, error) {
	var keys interface{}
	host := getVaultHost() + models.VaultGetAllKeysURI

	if strings.Contains(host, "{cloud}") {
		host = strings.Replace(host, "{cloud}", cloudType, -1)
	}

	if region != "" {
		if strings.Contains(host, "{region}") {
			host = strings.Replace(host, "{region}", region, -1)
		}
	}
	req, err := utils.CreateGetRequest(host)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return keys, err
	}
	client := utils.InitReq()
	req.Header.Set("X-Auth-Token", token)
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
func GetCredentialProfile(cloudType string, profileId string, token string, ctx utils.Context) (int, []byte, error) {
	host := getVaultHost() + models.VaultGetProfileURI

	if strings.Contains(host, "{cloud}") {
		host = strings.Replace(host, "{cloud}", cloudType, -1)
	}

	if strings.Contains(host, "{profileId}") {
		host = strings.Replace(host, "{profileId}", profileId, -1)
	}

	ctx.SendLogs(ctx.ReqRespData(types.ReqResPayload{
		Token:   token,
		Url:     host,
		ReqType: types.GET,
	}), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	req, err := utils.CreateGetRequest(host)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 500, []byte{}, err
	}
	req.Header.Add("X-Auth-Token", token)
	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 500, []byte{}, err
	}
	defer response.Body.Close()

	beego.Info(response.StatusCode)
	beego.Info(response.Status)
	if response.StatusCode == 403 {
		return 401, []byte{}, errors.New("User is not authorized for credential profile - " + profileId)
	} else if response.StatusCode == 404 {
		return response.StatusCode, []byte{}, errors.New("profile not found")
	}

	if response.StatusCode != 200 {
		return response.StatusCode, []byte{}, errors.New("profile not found " + response.Status)
	}

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 500, []byte{}, err
	}
	ctx.SendLogs(ctx.ReqRespData(types.ReqResPayload{
		Token:   token,
		Url:     host,
		ReqType: types.GET,
		Resp:    string(contents),
	}), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	return 0, contents, nil

}

func DeleteSSHkey(cloudType, keyName, token string, ctx utils.Context, region string) error {
	host := getVaultHost() + models.VaultDeleteKeyURI
	if strings.Contains(host, "{cloudType}") {
		host = strings.Replace(host, "{cloudType}", cloudType, -1)
	}

	if region != "" {
		if strings.Contains(host, "{region}") {
			host = strings.Replace(host, "{region}", region, -1)
		}
	}

	if strings.Contains(host, "{name}") {
		host = strings.Replace(host, "{name}", keyName, -1)
	}

	ctx.SendLogs(ctx.ReqRespData(types.ReqResPayload{
		Token:   token,
		Url:     host,
		ReqType: types.DELETE,
	}), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	req, err := utils.CreateDeleteRequest(host)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	client := utils.InitReq()

	m := make(map[string]string)
	m["Content-Type"] = "application/json"
	m["X-Auth-Token"] = token
	utils.SetHeaders(req, m)

	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer response.Body.Close()

	beego.Info(response.StatusCode)
	beego.Info(response.Status)
	if response.StatusCode == 403 {
		return errors.New("User is not authorized to delete this key - " + keyName)
	} else if response.StatusCode == 404 {
		return errors.New("key not found")
	}
	if response.StatusCode != 200 {
		return errors.New("Status Code: " + strconv.Itoa(response.StatusCode))
	}

	return nil

}
