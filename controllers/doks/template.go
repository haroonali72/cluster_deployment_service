package doks

import (
	"antelope/models"
	"antelope/models/doks"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"encoding/json"
	"github.com/astaxie/beego"
	"strings"
)

// Operations about DOKS template [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/template/{cloud}/]
type DOKSTemplateController struct {
	beego.Controller
}

// @Title Get
// @Description get kubernetes cluster template
// @Param	templateId	path	string	true	"Template Id of the template"
// @Param	token	header	string	true "token"
// @Success 200 {object} doks.KubernetesTemplate
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /:templateId/ [get]
func (c *DOKSTemplateController) Get() {

	templateId := c.GetString(":templateId")
	if templateId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "templateId is empty"}
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

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, templateId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, _, err = rbac_athentication.Authenticate(models.DOKS, "clusterTemplate", templateId, "View", token, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	/*if !allowed {
		c.Ctx.Output.SetStatus(403)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}
	*/
	//=============================================================================//
	ctx.SendLogs("DOKSTemplateController: Get template with id : "+templateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	template, err := doks.GetTemplate(templateId, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "no template exists for this id"}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("DOKS template of template id "+template.TemplateId+"fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = template
	c.ServeJSON()
}

// @Title Get All
// @Description get all kubernetes cluster templates
// @Param	token	header	string	true "token"
// @Success 200 {object} []doks.KubernetesTemplate
// @Failure 400 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /all [get]
func (c *DOKSTemplateController) GetAll() {

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, err, data := rbac_athentication.GetAllAuthenticate("clusterTemplate", userInfo.CompanyId, token, models.DOKS, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	//=============================================================================//
	ctx.SendLogs("DOKSTemplateController: GetAll template.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	templates, err := doks.GetTemplates(*ctx, data, userInfo.CompanyId)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("All DOKS Template fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = templates
	c.ServeJSON()
}

// @Title Create
// @Description create a new kubernetes cluster template
// @Param	token	header	string	true "token"
// @Param	teams	header	string	true "teams"
// @Param	body	body	doks.KubernetesTemplate	true	"body for template content"
// @Success 200 {"msg": "template created successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 409 {"error": "template with same name already exists"}
// @Failure 500 {"error": "error msg"}
// @router / [post]
func (c *DOKSTemplateController) Post() {
	var template doks.KubernetesTemplate
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &template)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": err.Error()}
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

	if template.TemplateId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "templateId is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	_, err = rbac_athentication.Evaluate("Create", token, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	/*if !allowed {
		c.Ctx.Output.SetStatus(403)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}
	*/
	ctx.SendLogs("DOKSTemplateController: Post new template with name: "+template.Name, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

	template.CompanyId = userInfo.CompanyId
	template.IsCloudplex = false

	err, id := doks.CreateTemplate(template, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "template with same id already exists"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	//==========================RBAC Policy Creation==============================//

	team := c.Ctx.Input.Header("teams")

	var teams []string
	if team != "" {
		teams = strings.Split(team, ";")
	}

	statusCode, err = rbac_athentication.CreatePolicy(id, token, userInfo.UserId, userInfo.CompanyId, models.POST, teams, models.DOKS, *ctx)
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
	ctx.SendLogs("DOKS template of template id "+template.TemplateId+" created", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "template generated successfully with id " + id}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing kubernetes template
// @Param	token	header	string	true "token"
// @Param	teams	header	string	true "token"
// @Param	body	body	doks.KubernetesTemplate	true	"body for template content"
// @Success 200 {"msg": "template updated successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "no template exists with this name"}
// @Failure 500 {"error": "error msg"}
// @router / [put]
func (c *DOKSTemplateController) Patch() {
	var template doks.KubernetesTemplate
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

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, template.TemplateId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, _, err = rbac_athentication.Authenticate(models.DOKS, "clusterTemplate", template.TemplateId, "Update", token, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	/*	if !allowed {
			c.Ctx.Output.SetStatus(403)
			c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
			c.ServeJSON()
			return
		}
	*/
	//=============================================================================//
	ctx.SendLogs("DOKSTemplateController: Patch template with template id : "+template.TemplateId, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

	err = doks.UpdateTemplate(template, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "no template exists with this project id"}
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

	statusCode, err = rbac_athentication.CreatePolicy(template.TemplateId, token, userInfo.UserId, userInfo.CompanyId, models.PUT, teams, models.DOKS, *ctx)
	if err != nil {
		beego.Error("error" + err.Error())
		c.Ctx.Output.SetStatus(400)
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
	ctx.SendLogs("DOKS template of template id "+template.TemplateId+" updated", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "template updated successfully"}
	c.ServeJSON()
}

// @Title Delete
// @Description delete kubernetes template
// @Param	token	header	string	true "token"
// @Param	templateId	path	string	true	"template id of the template"
// @Success 200 {"msg": "template deleted successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 500 {"error": "error msg"}
// @router /:templateId [delete]
func (c *DOKSTemplateController) Delete() {

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

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, templateId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, _, err = rbac_athentication.Authenticate(models.DOKS, "clusterTemplate", templateId, "Delete", token, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	/*	if !allowed {
			c.Ctx.Output.SetStatus(403)
			c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
			c.ServeJSON()
			return
		}
	*/
	//=============================================================================//
	ctx.SendLogs("DOKSTemplateController: Delete template with template Id "+templateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err = doks.DeleteTemplate(templateId, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	//==========================RBAC Authentication==============================//

	status_code, err := rbac_athentication.DeletePolicy(models.DOKS, templateId, token, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if status_code != 200 {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "RBAC Policy Deletion Failed"}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("DOKS template of template id "+templateId+" deleted", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	//==================================================================================
	c.Data["json"] = map[string]string{"msg": "template deleted successfully"}
	c.ServeJSON()
}

// @Title Create Customer Template
// @Description create a new customer template
// @Param	token	header	string	true "token"
// @Param	body	body	doks.KubernetesTemplate	true	"body for template content"
// @Success 200 {"msg": "customer template created successfully"}
// @Failure 400 {"error": "error message"}
// @Failure 401 {"error": "error message"}
// @Failure 404 {"error": "error message"}
// @Failure 409 {"error": "template with same name already exists"}
// @Failure 500 {"error": "error msg"}
// @router /customerTemplate [post]
func (c *DOKSTemplateController) PostCustomerTemplate() {

	var template doks.KubernetesTemplate
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
	//==============================RBAC Role Authentication====================================//

	roleInfo, err := rbac_athentication.GetRole(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !doks.CheckRole(roleInfo) {
		c.Ctx.Output.SetStatus(403)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//===========================================================================================//

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, "", "", "")

	ctx.SendLogs("DOKSTemplateController: Post new customer template with id: "+template.TemplateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err, id := doks.CreateCustomerTemplate(template, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "template with same id already exists"}
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
// @Success 200 {object} doks.KubernetesTemplate
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @router /customerTemplate/:templateId [get]
func (c *DOKSTemplateController) GetCustomerTemplate() {

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

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
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

	ctx.SendLogs("DOKSCustomerTemplateController: Get customer template  id : "+tempId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	template, err := doks.GetCustomerTemplate(tempId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "no customer template exists for this id"}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("DOKS customer template of template id "+template.TemplateId+" fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = template
	c.ServeJSON()
}

// @Title Update customer kubernetes template
// @Description update an existing kubernetes customer template
// @Param	token	header	string	true "token"
// @Param	body	body	doks.KubernetesTemplate	true	"body for template content"
// @Success 200 {"msg": "customer template updated successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "no template exists with this name"}
// @Failure 500 {"error": "error msg"}
// @router /customerTemplate [put]
func (c *DOKSTemplateController) PatchCustomerTemplate() {

	var template doks.KubernetesTemplate
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

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, template.TemplateId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Role Authentication=============================//

	roleInfo, err := rbac_athentication.GetRole(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if !doks.CheckRole(roleInfo) {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//=============================================================================//

	ctx.SendLogs("DOKSCustomerTemplateController: Patch template with template id : "+template.TemplateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err = doks.UpdateCustomerTemplate(template, *ctx)
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

	ctx.SendLogs("DOKS customer template of template id "+template.TemplateId+" updated", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
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
func (c *DOKSTemplateController) DeleteCustomerTemplate() {

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

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, templateId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Role Authentication=============================//

	_, err = rbac_athentication.GetRole(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	/*if !doks.CheckRole(roleInfo) {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}
	*/
	//=============================================================================//

	ctx.SendLogs("DOKSCustomerTemplateController: Delete customer template with template Id "+templateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err = doks.DeleteCustomerTemplate(templateId, *ctx)
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
// @Success 200 {object} []doks.KubernetesTemplate
// @Failure 400 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /allCustomerTemplates [get]
func (c *DOKSTemplateController) AllCustomerTemplates() {

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
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

	ctx.SendLogs("DOKSTemplateController: GetAllCustomerTemplate.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	templates, err := doks.GetAllCustomerTemplates(*ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("All DOKS Customer Template fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = templates
	c.ServeJSON()
}

// @Title   GetAllKubernetesTemplateInfo
// @Description get all kubernetes templates info
// @Param	token	header	string	true "token"
// @Success 200 {object} []doks.KubernetesTemplateMetadata
// @Failure 400 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /allTemplatesInfo [get]
func (c *DOKSTemplateController) GetAllTemplateInfo() {

	ctx := new(utils.Context)
	ctx.SendLogs("DOKSTemplateController:  Get Templates MetaData.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, err, data := rbac_athentication.GetAllAuthenticate("clusterTemplate", userInfo.CompanyId, token, models.DOKS, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	//==================================================================================
	templates, err := doks.GetTemplatesMetadata(*ctx, data, userInfo.CompanyId)
	if err != nil {
		ctx.SendLogs("DOKSTemplateController: Internal server error "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("DOKS templates meta data fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = templates
	c.ServeJSON()
}

// @Title   GetAllCustomerTemplateInfo
// @Description get all the customer templates info
// @Param	token	header	string	true "token"
// @Success 200 {object} []doks.KubernetesTemplateMetadata
// @Failure 400 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /allCustomerTemplatesInfo [get]
func (c *DOKSTemplateController) GetAllCustomerTemplateInfo() {

	ctx := new(utils.Context)
	ctx.SendLogs("DOKSTemplateController:  Get all customer Templates Info.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, err, data := rbac_athentication.GetAllAuthenticate("clusterTemplate", userInfo.CompanyId, token, models.GCP, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	//==================================================================================
	templates, err := doks.GetCustomerTemplatesMetadata(*ctx, data, userInfo.CompanyId)
	if err != nil {
		ctx.SendLogs("DOKSTemplateController: Internal server error "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("DOKS customer templates info fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = templates
	c.ServeJSON()
}
