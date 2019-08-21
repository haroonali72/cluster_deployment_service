package rbac_athentication

import (
	"antelope/models/types"
	"antelope/models/utils"
	"encoding/json"
	"io/ioutil"

	"github.com/astaxie/beego"
)

type Input struct {
	ResourceId  string   `json:"resource_id"`
	ResouceType string   `json:resource_type"`
	Teams       []string `json:"teams"`
	CompanyId   string   `json:"companyId"`
	UserName    string   `json:"username"`
}

func getRbacHost() string {
	return beego.AppConfig.String("rbac_url")
}
func GetAllAuthenticate(companyId string, token string, ctx utils.Context) (bool, error) {

	req, err := utils.CreateGetRequest(getRbacHost() + "/security/api/rbac/list")
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return false, err
	}
	q := req.URL.Query()
	q.Add("companyId", companyId)
	q.Add("resource_type", "cluster")

	req.Header.Set("token", token)
	req.URL.RawQuery = q.Encode()

	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return false, err
	}
	defer response.Body.Close()
	beego.Info(response.StatusCode)
	if response.StatusCode == 200 {
		return true, nil
	}
	return false, nil
}
func Authenticate(resourceId string, action string, token string, ctx utils.Context) (bool, error) {

	req, err := utils.CreateGetRequest(getRbacHost() + "/security/api/rbac/allowed/")
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return false, err
	}
	q := req.URL.Query()
	q.Add("resource_id", resourceId)
	q.Add("resource_type", "cluster")
	q.Add("action", action)

	req.Header.Set("token", token)
	req.URL.RawQuery = q.Encode()

	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return false, err
	}
	defer response.Body.Close()

	if response.StatusCode == 200 {
		return true, nil
	}
	return false, nil
}

func Evaluate(action string, token string, ctx utils.Context) (bool, error) {

	req, err := utils.CreateGetRequest(getRbacHost() + "/security/api/rbac/evaluate/")
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return false, err
	}
	q := req.URL.Query()
	q.Add("resource", "cluster")
	q.Add("action", action)
	req.Header.Set("token", token)
	req.URL.RawQuery = q.Encode()

	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return false, err
	}
	defer response.Body.Close()

	if response.StatusCode == 200 {
		return true, nil
	}
	return false, nil
}

func GetInfo(token string) (types.Response, error) {

	req, err := utils.CreateGetRequest(getRbacHost() + "/security/api/rbac/token/info")
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

	if response.StatusCode == 200 {
		return types.Response{}, nil
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

func CreatePolicy(resourceId, token, userName, companyId string, teams []string, ctx utils.Context) (int, error) {

	var input Input
	input.UserName = userName
	input.CompanyId = companyId
	input.ResouceType = "cluster"
	input.ResourceId = resourceId
	input.Teams = teams
	client := utils.InitReq()
	request_data, err := utils.TransformData(input)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return 400, err
	}
	req, err := utils.CreatePostRequest(request_data, getRbacHost()+"/security/api/rbac/policy")
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return 400, err
	}
	req.Header.Set("token", token)
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return 400, err
	}
	return response.StatusCode, err

}

func DeletePolicy(resourceId string, token string, ctx utils.Context) (int, error) {

	client := utils.InitReq()

	req, err := utils.CreateDeleteRequest(getRbacHost() + "/security/api/rbac/policy")
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return 400, err
	}
	q := req.URL.Query()
	q.Add("resource_id", resourceId)
	q.Add("resouce_type", "cluster")
	req.Header.Set("token", token)
	req.URL.RawQuery = q.Encode()

	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return 400, err
	}
	return response.StatusCode, err

}
