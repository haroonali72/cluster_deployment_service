package rbac_athentication

import (
	"antelope/models/utils"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	"io/ioutil"
)

type Response struct {
	Msg string `json:"msg"`
}
type Input struct {
	ResourceId  string `json:"resource_id"`
	ResouceType string `json:resource_type"`
	Teams       string `json:"teams"`
	CompanyId   string `json:"companyId"`
	UserName    string `json:"username"`
}

func getRbacHost() string {
	return beego.AppConfig.String("vault_url")
}
func Authenticate(resourceId string, action string, token string, ctx utils.Context) (bool, error) {

	req, err := utils.CreateGetRequest(getRbacHost() + "/security/api/rbac/allowed/")
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return false, err
	}
	q := req.URL.Query()
	q.Add("resource_id", resourceId)
	q.Add("resource_type", "Cluster")
	q.Add("action", action)
	q.Add("token", token)
	req.URL.RawQuery = q.Encode()

	client := utils.InitReq()
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return false, err
	}
	defer response.Body.Close()

	if response.StatusCode == 500 || response.StatusCode == 404 {
		return false, errors.New("not found")
	}
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return false, err
	}
	var output Response
	err = json.Unmarshal(contents, &output)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return false, err
	}
	if output.Msg == "user authenticated" {
		return true, nil
	}
	return false, nil
}

func CreatePolicy(resourceId, team, userName, companyId string, ctx utils.Context) (int, error) {

	var input Input
	input.UserName = userName
	input.CompanyId = companyId
	input.ResouceType = "Cluster"
	input.ResourceId = resourceId
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
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return 400, err
	}
	return response.StatusCode, err

}
func DeletePolicy(resourceId string, ctx utils.Context) (int, error) {

	var input Input
	input.ResouceType = "Cluster"
	input.ResourceId = resourceId
	client := utils.InitReq()
	request_data, err := utils.TransformData(input)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return 400, err
	}
	req, err := utils.CreateDeleteRequest(request_data, getRbacHost()+"/security/api/rbac/policy")
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return 400, err
	}
	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return 400, err
	}
	return response.StatusCode, err

}
