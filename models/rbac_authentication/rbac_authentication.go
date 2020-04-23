package rbac_athentication

import (
	"antelope/models"
	"antelope/models/types"
	"antelope/models/utils"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/astaxie/beego"
)

type Input struct {
	ResourceId  string       `json:"resource_id"`
	ResouceType string       `json:"resource_type"`
	Teams       []string     `json:"teams"`
	CompanyId   string       `json:"companyId"`
	UserName    string       `json:"username"`
	CloudType   models.Cloud `json:"sub_type"`
}
type List struct {
	Data []string `json:"data"`
}

func getRbacHost() string {
	return beego.AppConfig.String("rbac_url")
}
func GetAllAuthenticate(resourceType, companyId string, token string, cloudType models.Cloud, ctx utils.Context) (int,error , List) {

	req, err := utils.CreateGetRequest(getRbacHost() + models.RbacEndpoint + models.RbacListURI)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 500,err, List{}
	}
	q := req.URL.Query()
	q.Add("companyId", companyId)
	q.Add("resource_type", resourceType)
	q.Add("sub_type", string(cloudType))

	req.Header.Set("token", token)
	req.URL.RawQuery = q.Encode()

	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 500,err, List{}
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return response.StatusCode, errors.New( response.Status ), List{}
	}

	var data List
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 500,err, List{}
	}
	err = json.Unmarshal(contents, &data)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 500,err, List{}
	}

	return 0,nil, data
}
func Authenticate(cloud interface{}, resourceType, resourceId string, action string, token string, ctx utils.Context) (bool, error) {
	subType := ""
	b, err := json.Marshal(cloud)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return false, err
	}
	err = json.Unmarshal(b, &subType)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return false, err
	}
	req, err := utils.CreateGetRequest(getRbacHost() + models.RbacEndpoint + models.RbacAccessURI)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return false, err
	}
	q := req.URL.Query()
	q.Add("resource_id", resourceId)
	q.Add("resource_type", resourceType)
	q.Add("action", action)
	q.Add("sub_type", subType)
	req.Header.Set("token", token)
	req.URL.RawQuery = q.Encode()

	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return false, err
	}
	defer response.Body.Close()

	if response.StatusCode == 200 {
		return true, nil
	}
	return false, nil
}

func Evaluate(action string, token string, ctx utils.Context) (bool, error) {

	req, err := utils.CreateGetRequest(getRbacHost() + models.RbacEndpoint + models.RbacEvaluateURI)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return false, err
	}
	q := req.URL.Query()
	q.Add("resource", "clusterTemplate")
	q.Add("action", action)
	req.Header.Set("token", token)
	req.URL.RawQuery = q.Encode()

	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return false, err
	}
	defer response.Body.Close()

	if response.StatusCode == 200 {
		return true, nil
	}
	return false, nil

}

func GetInfo(token string) (types.Response, error) {

	req, err := utils.CreateGetRequest(getRbacHost() + models.RbacEndpoint + models.RbacInfoURI)
	if err != nil {
		return types.Response{}, err
	}
	q := req.URL.Query()
	req.Header.Set("token", token)
	req.URL.RawQuery = q.Encode()

	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		return types.Response{}, err
	}
	defer response.Body.Close()
	beego.Info(response.StatusCode)
	if response.StatusCode != 200 {
		return types.Response{}, errors.New("RBAC: Unauthorized , " + strconv.Itoa(response.StatusCode))
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {

		return types.Response{}, err
	}
	var res types.Response
	err = json.Unmarshal(contents, &res)
	if err != nil {
		return types.Response{}, err
	}
	return res, nil
}
func GetRole(token string) (types.UserRole, error) {

	req, err := utils.CreateGetRequest(getRbacHost() + models.RbacEndpoint + models.RbacExtractURI)
	if err != nil {
		return types.UserRole{}, err
	}
	q := req.URL.Query()
	req.Header.Set("token", token)
	req.URL.RawQuery = q.Encode()

	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		return types.UserRole{}, err
	}
	defer response.Body.Close()
	beego.Info(response.StatusCode)
	if response.StatusCode != 200 {
		return types.UserRole{}, errors.New("RBAC: Unauthorized , " + strconv.Itoa(response.StatusCode))
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {

		return types.UserRole{}, err
	}
	var res types.UserRole
	err = json.Unmarshal(contents, &res)
	if err != nil {
		return types.UserRole{}, err
	}
	return res, nil
}
func CreatePolicy(resourceId, token, userName, companyId string, requestType models.RequestType, teams []string, cloudType models.Cloud, ctx utils.Context) (int, error) {

	var input Input
	input.UserName = userName
	input.CompanyId = companyId
	input.ResouceType = "clusterTemplate"
	input.ResourceId = resourceId
	input.Teams = teams
	input.CloudType = cloudType

	client := utils.InitReq()
	request_data, err := utils.TransformData(input)
	if err != nil {

		beego.Info(err.Error())
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 400, err
	}
	var req *http.Request
	if requestType == models.POST {

		req, err = utils.CreatePostRequest(request_data, getRbacHost()+"/security/api/rbac/policy")
	} else if requestType == models.PUT {

		req, err = utils.CreatePutRequest(request_data, getRbacHost()+"/security/api/rbac/policy")
	}
	m := make(map[string]string)

	m["Content-Type"] = "application/json"
	utils.SetHeaders(req, m)
	if err != nil {
		beego.Info(err.Error())
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 400, err
	}
	req.Header.Set("token", token)
	response, err := client.SendRequest(req)
	if err != nil {
		beego.Info(err.Error())
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 400, err
	}
	contents, err := ioutil.ReadAll(response.Body)
	beego.Info(string(contents))
	beego.Info(response.StatusCode)

	return response.StatusCode, err

}
func DeletePolicy(cloud models.Cloud, resourceId string, token string, ctx utils.Context) (int, error) {

	client := utils.InitReq()

	req, err := utils.CreateDeleteRequest(getRbacHost() + models.RbacEndpoint + models.RbacPolicyURI)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 400, err
	}
	q := req.URL.Query()
	q.Add("resource_id", resourceId)
	q.Add("resource_type", "clusterTemplate")
	q.Add("sub_type", string(cloud))

	m := make(map[string]string)
	m["Content-Type"] = "application/json"
	m["token"] = token
	utils.SetHeaders(req, m)

	req.Header.Set("token", token)
	req.URL.RawQuery = q.Encode()

	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return 400, err
	}
	return response.StatusCode, err

}
