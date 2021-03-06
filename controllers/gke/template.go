package gke

import (
	"antelope/models"
	"antelope/models/gcp"
	"antelope/models/gke"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"encoding/json"
	"github.com/astaxie/beego"
	"strings"
)

// Operations about GKE template [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/template/{cloud}/]
type GKETemplateController struct {
	beego.Controller
}

// @Title Get
// @Description get template
// @Param	token	header	string	true "token"
// @Param	templateId	path	string	true	"Id of the template"
// @Success 200 {object} gke.GKEClusterTemplate
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @router /:templateId [get]
func (c *GKETemplateController) Get() {
	ctx := new(utils.Context)
	ctx.SendLogs("GKETemplateController: Get template", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	id := c.GetString(":templateId")
	if id == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "template Id is empty"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("GKETemplateController: Get template with id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, id, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GKETemplateController: Get template  id : "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.GKE, "clusterTemplate", id, "View", token, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	template, err := gke.GetGKEClusterTemplate(id, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("GKETemplateController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "no template exists for this id"}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("GKE template of template id "+template.TemplateId+" fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = template
	c.ServeJSON()
}

// @Title Get All
// @Description get all the templates
// @Param	token	header	string	true "token"
// @Success 200 {object} []gke.GKEClusterTemplate
// @Failure 400 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /all [get]
func (c *GKETemplateController) GetAll() {

	ctx := new(utils.Context)
	ctx.SendLogs("GKETemplateController: GetAll template.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GKETemplateController: GetAll template.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, err, _ = rbacAuthentication.GetAllAuthenticate("clusterTemplate", userInfo.CompanyId, token, models.GKE, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	templates, err := gke.GetAllGKEClusterTemplate(*ctx)
	if err != nil {
		ctx.SendLogs("GKETemplateController: Internal server error "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("GKE templates fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = templates
	c.ServeJSON()
}

// @Title Create
// @Description create a new template
// @Param	body	body	gke.GKEClusterTemplate	true	"body for template content"
// @Param	token	header	string	true "token"
// @Param	teams	header	string	true "teams"
// @Success 200 {"msg": "template created successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 409 {"error": "template with same name already exists"}
// @Failure 500 {"error": "error msg"}
// @router / [post]
func (c *GKETemplateController) Post() {
	var template gke.GKEClusterTemplate
	_ = json.Unmarshal(c.Ctx.Input.RequestBody, &template)

	ctx := new(utils.Context)
	ctx.SendLogs("GKETemplateController: Post new template with name: "+template.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	if template.TemplateId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "templateId is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, template.TemplateId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GKETemplateController: Posting  new template .", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	allowed, err := rbacAuthentication.Evaluate("Create", token, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	template.CompanyId = userInfo.CompanyId

	id, err := gke.AddGKEClusterTemplate(template, *ctx)
	if err != nil {
		ctx.SendLogs("GKETemplateController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "template with same name already exists"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	team := c.Ctx.Input.Header("teams")

	var teams []string
	if team != "" {
		teams = strings.Split(team, ";")
	}

	statusCode, err = rbacAuthentication.CreatePolicy(id, token, userInfo.UserId, userInfo.CompanyId, models.POST, teams, models.GKE, *ctx)
	if err != nil {
		//beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if statusCode != 200 {
		//beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "Policy creation failed"}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("GKE template of template id "+template.TemplateId+" created", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "template generated successfully with id " + id}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing template
// @Param	token	header	string	true "token"
// @Param	teams	header	string	true "teams"
// @Param	body	body	gke.GKEClusterTemplate	true	"body for template content"
// @Success 200 {"msg": "template updated successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "no template exists with this name"}
// @Failure 500 {"error": "error msg"}
// @router / [put]
func (c *GKETemplateController) Patch() {
	var template gke.GKEClusterTemplate
	_ = json.Unmarshal(c.Ctx.Input.RequestBody, &template)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	if template.TemplateId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "templateId is empty"}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, template.TemplateId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GKETemplateController: Patch template with templateId "+template.TemplateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.GKE, "clusterTemplate", template.TemplateId, "Update", token, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//==================================================================================
	ctx.SendLogs("GKETemplateController: Patch template with id: "+template.TemplateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	beego.Info("GKETemplateController: JSON Payload: ", template)

	err = gke.UpdateGKEClusterTemplate(template, *ctx)
	if err != nil {
		ctx.SendLogs("GKETemplateController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		if strings.Contains(err.Error(), "does not exist") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "no template exists with this id"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	team := c.Ctx.Input.Header("teams")

	var teams []string
	if team != "" {
		teams = strings.Split(team, ";")
	}

	statusCode, err = rbacAuthentication.CreatePolicy(template.TemplateId, token, userInfo.UserId, userInfo.CompanyId, models.PUT, teams, models.GKE, *ctx)
	if err != nil {
		beego.Error("error" + err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": "Policy creation failed"}
		c.ServeJSON()
		return
	}
	if statusCode != 200 {
		beego.Error(statusCode)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "Policy creation failed!"}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("GKE template of template id "+template.TemplateId+" updated", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "template updated successfully"}
	c.ServeJSON()
}

// @Title Delete
// @Description delete a templates
// @Param	token	header	string	true "token"
// @Param	templateId	path	string	true	"Name of the template"
// @Success 200 {"msg": "template deleted successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "name is empty"}
// @Failure 500 {"error": "error msg"}
// @router /:templateId [delete]
func (c *GKETemplateController) Delete() {
	id := c.GetString(":templateId")
	if id == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "template id is empty"}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)
	ctx.SendLogs("GKETemplateController: Delete template with id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, id, userInfo.CompanyId, userInfo.UserId)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.GKE, "clusterTemplate", id, "Delete", token, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	err = gke.DeleteGKEClusterTemplate(id, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("GKETemplateController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("GKETemplateController: Deleting template with templateId "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, err = rbacAuthentication.DeletePolicy(models.GKE, id, token, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if statusCode != 200 {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "RBAC Policy Deletion Failed"}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("GKE template of template id "+id+" deleted", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = map[string]string{"msg": "template deleted successfully"}
	c.ServeJSON()
}

// @Title Create Customer Template
// @Description create a new customer template
// @Param	token	header	string	true "token"
// @Param	body	body	gke.GKEClusterTemplate	true	"body for template content"
// @Success 200 {"msg": "template created successfully"}
// @Failure 409 {"error": "template with same name already exists"}
// @Failure 500 {"error": "error msg"}
// @router /customerTemplate [post]
func (c *GKETemplateController) PostCustomerTemplate() {
	var template gke.GKEClusterTemplate
	json.Unmarshal(c.Ctx.Input.RequestBody, &template)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	if template.TemplateId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "templateId is empty"}
		c.ServeJSON()
		return
	}

	roleInfo, err := rbacAuthentication.GetRole(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !gcp.CheckRole(roleInfo) {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, "", "", "")

	ctx.SendLogs("GKETemplateController: Post new customer template with name: "+template.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err, id := gke.CreateGKECustomerTemplate(template, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "template with same name already exists"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	c.Data["json"] = map[string]string{"msg": "template generated successfully with id " + id}
	c.ServeJSON()
}

// @Title Get customer template
// @Description get customer template
// @Param	templateId	path	string	true	"Template Id of the template"
// @Param	token	header	string	true "token"
// @Success 200 {object} gke.GKEClusterTemplate
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @router /customerTemplate/:templateId [get]
func (c *GKETemplateController) GetCustomerTemplate() {
	tempId := c.GetString(":templateId")
	if tempId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "templateId is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "templateId is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, tempId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC User Authentication==============================//

	check := strings.Contains(userInfo.UserId, "cloudplex.io")

	if !check {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Unauthorized to access this template"}
		c.ServeJSON()
		return
	}

	//=============================================================================//

	ctx.SendLogs("GKECustomerTemplateController: Get customer template  id : "+tempId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	template, err := gke.GetGKECustomerTemplate(tempId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "no customer template exists for this id"}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("GKE customer template of template id "+template.TemplateId+" fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = template
	c.ServeJSON()
}

// @Title Update customer templates
// @Description update an existing customer template
// @Param	token	header	string	true "token"
// @Param	teams	header	string	true "token"
// @Param	body	body	gke.GKEClusterTemplate	true	"body for template content"
// @Success 200 {"msg": "customer template updated successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "no template exists with this name"}
// @Failure 500 {"error": "error msg"}
// @router /customerTemplate [put]
func (c *GKETemplateController) PatchCustomerTemplate() {
	var template gke.GKEClusterTemplate
	json.Unmarshal(c.Ctx.Input.RequestBody, &template)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	if template.TemplateId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "templateId is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, template.TemplateId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Role Authentication=============================//

	roleInfo, err := rbacAuthentication.GetRole(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if !gcp.CheckRole(roleInfo) {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//=============================================================================//

	ctx.SendLogs("GKECustomerTemplateController: Patch template with template id : "+template.TemplateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err = gke.UpdateGKECustomerTemplate(template, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "no customer template exists with this project id"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("GKE customer template of template id "+template.TemplateId+" updated", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": " customer template updated successfully"}
	c.ServeJSON()
}

// @Title Delete customer template
// @Description delete a customer template
// @Param	token	header	string	true "token"
// @Param	templateId	path	string	true	"template id of the template"
// @Success 200 {"msg": "customer template deleted successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 500 {"error": "error msg"}
// @router /customerTemplate/:templateId [delete]
func (c *GKETemplateController) DeleteCustomerTemplate() {
	templateId := c.GetString(":templateId")
	if templateId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "template id is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, templateId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Role Authentication=============================//

	roleInfo, err := rbacAuthentication.GetRole(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if !gcp.CheckRole(roleInfo) {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//=============================================================================//
	ctx.SendLogs("GKECustomerTemplateController: Delete customer template with template Id "+templateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err = gke.DeleteGKECustomerTemplate(templateId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"msg": "customer template deleted successfully"}
	c.ServeJSON()
}

// @Title Get All Customer Template
// @Description get all the customer templates
// @Param	token	header	string	true "token"
// @Success 200 {object} []gke.GKEClusterTemplate
// @Failure 400 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /allCustomerTemplates [get]
func (c *GKETemplateController) AllCustomerTemplates() {
	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC User Authentication==============================//

	check := strings.Contains(userInfo.UserId, "cloudplex.io")

	if !check {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "Unauthorized to access this templates"}
		c.ServeJSON()
		return
	}

	//=============================================================================//

	ctx.SendLogs("GKETemplateController: GetAllCustomerTemplate.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	templates, err := gke.GetAllGKECustomerTemplates(*ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("All GKE Customer Template fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = templates
	c.ServeJSON()
}

// @Title   GetAllTemplateInfo
// @Description get all the templates info
// @Param	token	header	string	true "token"
// @Success 200 {object} []gke.GKEClusterTemplateMetadata
// @Failure 400 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /allTemplatesInfo [get]
func (c *GKETemplateController) GetAllTemplateInfo() {
	ctx := new(utils.Context)
	ctx.SendLogs("GKETemplateController:  Get Templates MetaData.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, err, data := rbacAuthentication.GetAllAuthenticate("clusterTemplate", userInfo.CompanyId, token, models.GKE, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	//==================================================================================
	templates, err := gke.GetGKETemplatesMetadata(*ctx, data, userInfo.CompanyId)
	if err != nil {
		ctx.SendLogs("GKETemplateController: Internal server error "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("GKE templates meta data fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = templates
	c.ServeJSON()
}

// @Title   GetAllCustomerTemplateInfo
// @Description get all the customer templates info
// @Param	token	header	string	true "token"
// @Success 200 {object} []gke.GKEClusterTemplateMetadata
// @Failure 400 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /allCustomerTemplatesInfo [get]
func (c *GKETemplateController) GetAllCustomerTemplateInfo() {
	ctx := new(utils.Context)
	ctx.SendLogs("GKETemplateController:  Get all customer Templates Info.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, err, data := rbacAuthentication.GetAllAuthenticate("clusterTemplate", userInfo.CompanyId, token, models.GKE, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	//==================================================================================
	templates, err := gke.GetGKECustomerTemplatesMetadata(*ctx, data, userInfo.CompanyId)
	if err != nil {
		ctx.SendLogs("GKETemplateController: Internal server error "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("GKE customer templates info fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = templates
	c.ServeJSON()
}
