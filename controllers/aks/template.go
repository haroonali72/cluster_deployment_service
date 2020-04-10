package aks

//
//import (
//	"antelope/models"
//	"antelope/models/aks"
//	"antelope/models/gcp"
//	rbacAuthentication "antelope/models/rbac_authentication"
//	"antelope/models/utils"
//	"encoding/json"
//	"github.com/astaxie/beego"
//	"strings"
//)
//
//// Operations about AKS template [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/template/{cloud}/]
//type AKSTemplateController struct {
//	beego.Controller
//}
//
//// @Title Get
//// @Description get template
//// @Param	token	header	string	token ""
//// @Param	templateId	path	string	true	"Id of the template"
//// @Success 200 {object} aks.AKSClusterTemplate
//// @Failure 401 {"error": "error msg"}
//// @Failure 404 {"error": "error msg"}
//// @router /:templateId [get]
//func (c *AKSTemplateController) Get() {
//	ctx := new(utils.Context)
//	ctx.SendLogs("AKSTemplateController: Get template", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//
//	id := c.GetString(":templateId")
//	if id == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "template Id is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	ctx.SendLogs("AKSTemplateController: Get template with id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//
//	token := c.Ctx.Input.Header("token")
//	if token == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "token is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	userInfo, err := rbacAuthentication.GetInfo(token)
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, id, userInfo.CompanyId, userInfo.UserId)
//	ctx.SendLogs("AKSTemplateController: Get template  id : "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//
//	allowed, err := rbacAuthentication.Authenticate(models.AKS, "clusterTemplate", id, "View", token, utils.Context{})
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//	if !allowed {
//		c.Ctx.Output.SetStatus(401)
//		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
//		c.ServeJSON()
//		return
//	}
//
//	template, err := aks.GetAKSClusterTemplate(id, userInfo.CompanyId, *ctx)
//	if err != nil {
//		ctx.SendLogs("AKSTemplateController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "no template exists for this id"}
//		c.ServeJSON()
//		return
//	}
//	ctx.SendLogs("AKS template of template id "+template.TemplateId+" fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
//	c.Data["json"] = template
//	c.ServeJSON()
//}
//
//// @Title Get All
//// @Description get all the templates
//// @Param	token	header	string	token ""
//// @Success 200 {object} []aks.AKSClusterTemplate
//// @Failure 400 {"error": "error msg"}
//// @Failure 500 {"error": "error msg"}
//// @router /all [get]
//func (c *AKSTemplateController) GetAll() {
//
//	ctx := new(utils.Context)
//	ctx.SendLogs("AKSTemplateController: GetAll template.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//
//	token := c.Ctx.Input.Header("token")
//	if token == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "token is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	userInfo, err := rbacAuthentication.GetInfo(token)
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
//	ctx.SendLogs("AKSTemplateController: GetAll template.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//
//	err, _ = rbacAuthentication.GetAllAuthenticate("clusterTemplate", userInfo.CompanyId, token, models.AKS, utils.Context{})
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	templates, err := aks.GetAllAKSClusterTemplate(*ctx)
//	if err != nil {
//		ctx.SendLogs("AKSTemplateController: Internal server error "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
//		c.Ctx.Output.SetStatus(500)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//	ctx.SendLogs("AKS templates fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
//	c.Data["json"] = templates
//	c.ServeJSON()
//}
//
//// @Title Create
//// @Description create a new template
//// @Param	body	body	aks.AKSClusterTemplate	true	"body for template content"
//// @Param	token	header	string	token ""
//// @Param	teams	header	string	teams ""
//// @Success 200 {"msg": "template created successfully"}
//// @Failure 400 {"error": "error msg"}
//// @Failure 401 {"error": "error msg"}
//// @Failure 409 {"error": "template with same name already exists"}
//// @Failure 500 {"error": "error msg"}
//// @router / [post]
//func (c *AKSTemplateController) Post() {
//	var template aks.AKSClusterTemplate
//	_ = json.Unmarshal(c.Ctx.Input.RequestBody, &template)
//
//	ctx := new(utils.Context)
//	ctx.SendLogs("AKSTemplateController: Post new template with name: "+*template.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//
//	token := c.Ctx.Input.Header("token")
//	if token == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "token is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	if template.TemplateId == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "templateId is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	userInfo, err := rbacAuthentication.GetInfo(token)
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, template.TemplateId, userInfo.CompanyId, userInfo.UserId)
//	ctx.SendLogs("AKSTemplateController: Posting  new template .", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//
//	allowed, err := rbacAuthentication.Evaluate("Create", token, utils.Context{})
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//	if !allowed {
//		c.Ctx.Output.SetStatus(401)
//		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
//		c.ServeJSON()
//		return
//	}
//
//	template.CompanyId = userInfo.CompanyId
//
//	id, err := aks.AddAKSClusterTemplate(template, *ctx)
//	if err != nil {
//		ctx.SendLogs("AKSTemplateController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
//
//		if strings.Contains(err.Error(), "already exists") {
//			c.Ctx.Output.SetStatus(409)
//			c.Data["json"] = map[string]string{"error": "template with same name already exists"}
//			c.ServeJSON()
//			return
//		}
//		c.Ctx.Output.SetStatus(500)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	team := c.Ctx.Input.Header("teams")
//
//	var teams []string
//	if team != "" {
//		teams = strings.Split(team, ";")
//	}
//
//	statusCode, err := rbacAuthentication.CreatePolicy(id, token, userInfo.UserId, userInfo.CompanyId, models.POST, teams, models.AKS, *ctx)
//	if err != nil {
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": "Policy creation failed"}
//		c.ServeJSON()
//		return
//	}
//	if statusCode != 200 {
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": "Policy creation failed"}
//		c.ServeJSON()
//		return
//	}
//	ctx.SendLogs("AKS template of template id "+template.TemplateId+" created", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
//	c.Data["json"] = map[string]string{"msg": "template generated successfully with id " + id}
//	c.ServeJSON()
//}
//
//// @Title Update
//// @Description update an existing template
//// @Param	token	header	string	token ""
//// @Param	teams	header	string	teams ""
//// @Param	body	body	aks.AKSClusterTemplate	true	"body for template content"
//// @Success 200 {"msg": "template updated successfully"}
//// @Failure 400 {"error": "error msg"}
//// @Failure 401 {"error": "error msg"}
//// @Failure 404 {"error": "no template exists with this name"}
//// @Failure 500 {"error": "error msg"}
//// @router / [put]
//func (c *AKSTemplateController) Patch() {
//	var template aks.AKSClusterTemplate
//	_ = json.Unmarshal(c.Ctx.Input.RequestBody, &template)
//
//	token := c.Ctx.Input.Header("token")
//	if token == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "token is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	if template.TemplateId == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "templateId is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	ctx := new(utils.Context)
//
//	userInfo, err := rbacAuthentication.GetInfo(token)
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, template.TemplateId, userInfo.CompanyId, userInfo.UserId)
//	ctx.SendLogs("AKSTemplateController: Patch template with templateId "+template.TemplateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//
//	allowed, err := rbacAuthentication.Authenticate(models.AKS, "clusterTemplate", template.TemplateId, "Update", token, utils.Context{})
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//	if !allowed {
//		c.Ctx.Output.SetStatus(401)
//		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
//		c.ServeJSON()
//		return
//	}
//
//	//==================================================================================
//	ctx.SendLogs("AKSTemplateController: Patch template with id: "+template.TemplateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//	beego.Info("AKSTemplateController: JSON Payload: ", template)
//
//	err = aks.UpdateAKSClusterTemplate(template, *ctx)
//	if err != nil {
//		ctx.SendLogs("AKSTemplateController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
//		if strings.Contains(err.Error(), "does not exist") {
//			c.Ctx.Output.SetStatus(404)
//			c.Data["json"] = map[string]string{"error": "no template exists with this id"}
//			c.ServeJSON()
//			return
//		}
//		c.Ctx.Output.SetStatus(500)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	team := c.Ctx.Input.Header("teams")
//
//	var teams []string
//	if team != "" {
//		teams = strings.Split(team, ";")
//	}
//
//	statusCode, err := rbacAuthentication.CreatePolicy(template.TemplateId, token, userInfo.UserId, userInfo.CompanyId, models.PUT, teams, models.AKS, *ctx)
//	if err != nil {
//		beego.Error("error" + err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": "Policy creation failed"}
//		c.ServeJSON()
//		return
//	}
//	if statusCode != 200 {
//		beego.Error(statusCode)
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": "Policy creation failed!"}
//		c.ServeJSON()
//		return
//	}
//	ctx.SendLogs("AKS template of template id "+template.TemplateId+" updated", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
//	c.Data["json"] = map[string]string{"msg": "template updated successfully"}
//	c.ServeJSON()
//}
//
//// @Title Delete
//// @Description delete a templates
//// @Param	token	header	string	token ""
//// @Param	templateId	path	string	true	"Name of the template"
//// @Success 200 {"msg": "template deleted successfully"}
//// @Failure 400 {"error": "error msg"}
//// @Failure 401 {"error": "error msg"}
//// @Failure 404 {"error": "name is empty"}
//// @Failure 500 {"error": "error msg"}
//// @router /:templateId [delete]
//func (c *AKSTemplateController) Delete() {
//	id := c.GetString(":templateId")
//	if id == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "template id is empty"}
//		c.ServeJSON()
//		return
//	}
//	ctx := new(utils.Context)
//	ctx.SendLogs("AKSTemplateController: Delete template with id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//
//	token := c.Ctx.Input.Header("token")
//	if token == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "token is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	userInfo, err := rbacAuthentication.GetInfo(token)
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, id, userInfo.CompanyId, userInfo.UserId)
//
//	allowed, err := rbacAuthentication.Authenticate(models.AKS, "clusterTemplate", id, "Delete", token, utils.Context{})
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//	if !allowed {
//		c.Ctx.Output.SetStatus(401)
//		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
//		c.ServeJSON()
//		return
//	}
//
//	err = aks.DeleteAKSClusterTemplate(id, userInfo.CompanyId, *ctx)
//	if err != nil {
//		ctx.SendLogs("AKSTemplateController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
//		c.Ctx.Output.SetStatus(500)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//	ctx.SendLogs("AKSTemplateController: Deleting template with templateId "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//
//	statusCode, err := rbacAuthentication.DeletePolicy(models.GKE, id, token, utils.Context{})
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//	if statusCode != 200 {
//		c.Ctx.Output.SetStatus(401)
//		c.Data["json"] = map[string]string{"error": "RBAC Policy Deletion Failed"}
//		c.ServeJSON()
//		return
//	}
//	ctx.SendLogs("GKE template of template id "+id+" deleted", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
//
//	c.Data["json"] = map[string]string{"msg": "template deleted successfully"}
//	c.ServeJSON()
//}
//
//// @Title Create Customer Template
//// @Description create a new customer template
//// @Param	token	header	string	token ""
//// @Param	body	body	aks.AKSClusterTemplate	true	"body for template content"
//// @Success 200 {"msg": "template created successfully"}
//// @Failure 409 {"error": "template with same name already exists"}
//// @Failure 500 {"error": "error msg"}
//// @router /customerTemplate [post]
//func (c *AKSTemplateController) PostCustomerTemplate() {
//	var template aks.AKSClusterTemplate
//	json.Unmarshal(c.Ctx.Input.RequestBody, &template)
//
//	token := c.Ctx.Input.Header("token")
//	if token == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "token is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	if template.TemplateId == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "templateId is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	roleInfo, err := rbacAuthentication.GetRole(token)
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//	if !gcp.CheckRole(roleInfo) {
//		c.Ctx.Output.SetStatus(401)
//		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
//		c.ServeJSON()
//		return
//	}
//	ctx := new(utils.Context)
//	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, "", "", "")
//
//	ctx.SendLogs("AKSTemplateController: Post new customer template with name: "+template.TemplateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//
//	err, id := aks.CreateAKSCustomerTemplate(template, *ctx)
//	if err != nil {
//		if strings.Contains(err.Error(), "already exists") {
//			c.Ctx.Output.SetStatus(409)
//			c.Data["json"] = map[string]string{"error": "template with same name already exists"}
//			c.ServeJSON()
//			return
//		}
//		c.Ctx.Output.SetStatus(500)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//	c.Data["json"] = map[string]string{"msg": "template generated successfully with id " + id}
//	c.ServeJSON()
//}
//
//// @Title Get customer template
//// @Description get customer template
//// @Param	templateId	path	string	true	"Template Id of the template"
//// @Param	token	header	string	token ""
//// @Success 200 {object} aks.AKSClusterTemplate
//// @Failure 400 {"error": "error msg"}
//// @Failure 401 {"error": "error msg"}
//// @Failure 404 {"error": "error msg"}
//// @router /customerTemplate/:templateId [get]
//func (c *AKSTemplateController) GetCustomerTemplate() {
//	tempId := c.GetString(":templateId")
//	if tempId == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "templateId is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	token := c.Ctx.Input.Header("token")
//	if token == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "templateId is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	userInfo, err := rbacAuthentication.GetInfo(token)
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	ctx := new(utils.Context)
//	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, tempId, userInfo.CompanyId, userInfo.UserId)
//
//	//==========================RBAC User Authentication==============================//
//
//	check := strings.Contains(userInfo.UserId, "cloudplex.io")
//
//	if !check {
//		c.Ctx.Output.SetStatus(401)
//		c.Data["json"] = map[string]string{"error": "Unauthorized to access this template"}
//		c.ServeJSON()
//		return
//	}
//
//	//=============================================================================//
//
//	ctx.SendLogs("AKSCustomerTemplateController: Get customer template  id : "+tempId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//
//	template, err := aks.GetAKSCustomerTemplate(tempId, *ctx)
//	if err != nil {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "no customer template exists for this id"}
//		c.ServeJSON()
//		return
//	}
//	ctx.SendLogs("AKS customer template of template id "+template.TemplateId+" fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
//	c.Data["json"] = template
//	c.ServeJSON()
//}
//
//// @Title Update customer templates
//// @Description update an existing customer template
//// @Param	token	header	string	token ""
//// @Param	teams	header	string	token ""
//// @Param	body	body	aks.AKSClusterTemplate	true	"body for template content"
//// @Success 200 {"msg": "customer template updated successfully"}
//// @Failure 400 {"error": "error msg"}
//// @Failure 401 {"error": "error msg"}
//// @Failure 404 {"error": "no template exists with this name"}
//// @Failure 500 {"error": "error msg"}
//// @router /customerTemplate [put]
//func (c *AKSTemplateController) PatchCustomerTemplate() {
//	var template aks.AKSClusterTemplate
//	json.Unmarshal(c.Ctx.Input.RequestBody, &template)
//
//	token := c.Ctx.Input.Header("token")
//	if token == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "token is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	if template.TemplateId == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "templateId is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	userInfo, err := rbacAuthentication.GetInfo(token)
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	ctx := new(utils.Context)
//	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, template.TemplateId, userInfo.CompanyId, userInfo.UserId)
//
//	//==========================RBAC Role Authentication=============================//
//
//	roleInfo, err := rbacAuthentication.GetRole(token)
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	if !gcp.CheckRole(roleInfo) {
//		c.Ctx.Output.SetStatus(401)
//		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
//		c.ServeJSON()
//		return
//	}
//
//	//=============================================================================//
//
//	ctx.SendLogs("AKSCustomerTemplateController: Patch template with template id : "+template.TemplateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//
//	err = aks.UpdateAKSCustomerTemplate(template, *ctx)
//	if err != nil {
//		if strings.Contains(err.Error(), "does not exist") {
//			c.Ctx.Output.SetStatus(404)
//			c.Data["json"] = map[string]string{"error": "no customer template exists with this project id"}
//			c.ServeJSON()
//			return
//		}
//		c.Ctx.Output.SetStatus(500)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	ctx.SendLogs("AKS customer template of template id "+template.TemplateId+" updated", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
//	c.Data["json"] = map[string]string{"msg": " customer template updated successfully"}
//	c.ServeJSON()
//}
//
//// @Title Delete customer template
//// @Description delete a customer template
//// @Param	token	header	string	token ""
//// @Param	templateId	path	string	true	"template id of the template"
//// @Success 200 {"msg": "customer template deleted successfully"}
//// @Failure 400 {"error": "error msg"}
//// @Failure 401 {"error": "error msg"}
//// @Failure 404 {"error": "project id is empty"}
//// @Failure 500 {"error": "error msg"}
//// @router /customerTemplate/:templateId [delete]
//func (c *AKSTemplateController) DeleteCustomerTemplate() {
//	templateId := c.GetString(":templateId")
//	if templateId == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "template id is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	token := c.Ctx.Input.Header("token")
//	if token == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "token is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	userInfo, err := rbacAuthentication.GetInfo(token)
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	ctx := new(utils.Context)
//	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, templateId, userInfo.CompanyId, userInfo.UserId)
//
//	//==========================RBAC Role Authentication=============================//
//
//	roleInfo, err := rbacAuthentication.GetRole(token)
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	if !gcp.CheckRole(roleInfo) {
//		c.Ctx.Output.SetStatus(401)
//		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
//		c.ServeJSON()
//		return
//	}
//
//	//=============================================================================//
//	ctx.SendLogs("AKSCustomerTemplateController: Delete customer template with template Id "+templateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//
//	err = aks.DeleteAKSCustomerTemplate(templateId, *ctx)
//	if err != nil {
//		c.Ctx.Output.SetStatus(500)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	c.Data["json"] = map[string]string{"msg": "customer template deleted successfully"}
//	c.ServeJSON()
//}
//
//// @Title Get All Customer Template
//// @Description get all the customer templates
//// @Param	token	header	string	token ""
//// @Success 200 {object} []aks.AKSClusterTemplate
//// @Failure 400 {"error": "error msg"}
//// @Failure 404 {"error": "error msg"}
//// @Failure 500 {"error": "error msg"}
//// @router /allCustomerTemplates [get]
//func (c *AKSTemplateController) AllCustomerTemplates() {
//	token := c.Ctx.Input.Header("token")
//	if token == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "token is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	userInfo, err := rbacAuthentication.GetInfo(token)
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	ctx := new(utils.Context)
//	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
//
//	//==========================RBAC User Authentication==============================//
//
//	check := strings.Contains(userInfo.UserId, "cloudplex.io")
//
//	if !check {
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": "Unauthorized to access this templates"}
//		c.ServeJSON()
//		return
//	}
//
//	//=============================================================================//
//
//	ctx.SendLogs("AKSTemplateController: GetAllCustomerTemplate.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//	templates, err := aks.GetAllAKSCustomerTemplates(*ctx)
//	if err != nil {
//		c.Ctx.Output.SetStatus(500)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//	ctx.SendLogs("All AKS Customer Template fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
//	c.Data["json"] = templates
//	c.ServeJSON()
//}
//
//// @Title   GetAllTemplateInfo
//// @Description get all the templates info
//// @Param	token	header	string	token ""
//// @Success 200 {object} []aks.AKSClusterTemplateMetadata
//// @Failure 400 {"error": "error msg"}
//// @Failure 500 {"error": "error msg"}
//// @router /allTemplatesInfo [get]
//func (c *AKSTemplateController) GetAllTemplateInfo() {
//	ctx := new(utils.Context)
//	ctx.SendLogs("AKSTemplateController:  Get Templates MetaData.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//
//	token := c.Ctx.Input.Header("token")
//	if token == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "token is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	userInfo, err := rbacAuthentication.GetInfo(token)
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
//
//	//==========================RBAC Authentication==============================//
//
//	err, data := rbacAuthentication.GetAllAuthenticate("clusterTemplate", userInfo.CompanyId, token, models.AKS, utils.Context{})
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	//==================================================================================
//	templates, err := aks.GetAKSTemplatesMetadata(*ctx, data, userInfo.CompanyId)
//	if err != nil {
//		ctx.SendLogs("AKSTemplateController: Internal server error "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
//		c.Ctx.Output.SetStatus(500)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//	ctx.SendLogs("AKS templates meta data fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
//	c.Data["json"] = templates
//	c.ServeJSON()
//}
//
//// @Title   GetAllCustomerTemplateInfo
//// @Description get all the customer templates info
//// @Param	token	header	string	token ""
//// @Success 200 {object} []aks.AKSClusterTemplateMetadata
//// @Failure 400 {"error": "error msg"}
//// @Failure 500 {"error": "error msg"}
//// @router /allCustomerTemplatesInfo [get]
//func (c *AKSTemplateController) GetAllCustomerTemplateInfo() {
//	ctx := new(utils.Context)
//	ctx.SendLogs("AKSTemplateController:  Get all customer Templates Info.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
//
//	token := c.Ctx.Input.Header("token")
//	if token == "" {
//		c.Ctx.Output.SetStatus(404)
//		c.Data["json"] = map[string]string{"error": "token is empty"}
//		c.ServeJSON()
//		return
//	}
//
//	userInfo, err := rbacAuthentication.GetInfo(token)
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
//
//	//==========================RBAC Authentication==============================//
//
//	err, data := rbacAuthentication.GetAllAuthenticate("clusterTemplate", userInfo.CompanyId, token, models.AKS, utils.Context{})
//	if err != nil {
//		beego.Error(err.Error())
//		c.Ctx.Output.SetStatus(400)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//
//	//==================================================================================
//	templates, err := aks.GetAKSCustomerTemplatesMetadata(*ctx, data, userInfo.CompanyId)
//	if err != nil {
//		ctx.SendLogs("AKSTemplateController: Internal server error "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
//		c.Ctx.Output.SetStatus(500)
//		c.Data["json"] = map[string]string{"error": err.Error()}
//		c.ServeJSON()
//		return
//	}
//	ctx.SendLogs("AKS customer templates info fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
//	c.Data["json"] = templates
//	c.ServeJSON()
//}
