package ibm

import (
	"antelope/models"
	"antelope/models/ibm"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"encoding/json"
	"github.com/astaxie/beego"
	"strings"
)

// Operations about Gcp template [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/template/{cloud}/]
type IKSTemplateController struct {
	beego.Controller
}

// @Title Get
// @Description get template
// @Param	token	header	string	token ""
// @Param	templateId	path	string	true	"Id of the template"
// @Success 200 {object} ibm.Template
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @router /:templateId [get]
func (c *IKSTemplateController) Get() {
	ctx := new(utils.Context)
	ctx.SendLogs("IKSTemplateController: Get template", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	id := c.GetString(":templateId")
	if id == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "template Id is empty"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("IKSTemplateController: Get template with id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, id, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("IKSTemplateController: Get template  id : "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	allowed, err := rbac_athentication.Authenticate(models.IBM, "clusterTemplate", id, "View", token, utils.Context{})
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

	//==================================================================================//

	template, err := ibm.GetTemplate(id, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("IKSTemplateController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "no template exists for this id"}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("Ibm template of template id "+template.TemplateId+" fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = template
	c.ServeJSON()
}

// @Title Get All
// @Description get all the templates
// @Param	token	header	string	token ""
// @Success 200 {object} []ibm.Template
// @Failure 400 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /all [get]
func (c *IKSTemplateController) GetAll() {

	ctx := new(utils.Context)
	ctx.SendLogs("IKSTemplateController: GetAll template.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("IKSTemplateController: GetAll template.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	err, data := rbac_athentication.GetAllAuthenticate("clusterTemplate", userInfo.CompanyId, token, models.IBM, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	//==================================================================================

	templates, err := ibm.GetTemplates(*ctx, data, userInfo.CompanyId)
	if err != nil {
		ctx.SendLogs("IKSTemplateController: Internal server error "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("Ibm templates fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = templates
	c.ServeJSON()
}

// @Title Create
// @Description create a new template
// @Param	body	body	ibm.Template	true	"body for template content"
// @Param	token	header	string	token ""
// @Param	teams	header	string	teams ""
// @Success 200 {"msg": "template created successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 409 {"error": "template with same name already exists"}
// @Failure 500 {"error": "error msg"}
// @router / [post]
func (c *IKSTemplateController) Post() {

	var template ibm.Template
	json.Unmarshal(c.Ctx.Input.RequestBody, &template)

	ctx := new(utils.Context)
	ctx.SendLogs("IKSTemplateController: Post new template with name: "+template.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, template.TemplateId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("IKSTemplateController: Posting  new template .", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//
	allowed, err := rbac_athentication.Evaluate("Create", token, utils.Context{})
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
	template.IsCloudplex = false

	err, id := ibm.CreateTemplate(template, *ctx)
	if err != nil {
		ctx.SendLogs("IKSTemplateController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

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

	//==========================RBAC Policy Creation==============================//

	team := c.Ctx.Input.Header("teams")

	var teams []string
	if team != "" {
		teams = strings.Split(team, ";")
	}

	statusCode, err := rbac_athentication.CreatePolicy(id, token, userInfo.UserId, userInfo.CompanyId, models.POST, teams, models.IBM, *ctx)
	if err != nil {
		//beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "Policy creation failed"}
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
	ctx.SendLogs("Ibm template of template id "+template.TemplateId+" created", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "template generated successfully with id " + id}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing template
// @Param	token	header	string	token ""
// @Param	teams	header	string	teams ""
// @Param	body	body	ibm.Template	true	"body for template content"
// @Success 200 {"msg": "template updated successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "no template exists with this name"}
// @Failure 500 {"error": "error msg"}
// @router / [put]
func (c *IKSTemplateController) Patch() {

	var template ibm.Template
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

	ctx := new(utils.Context)

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, template.TemplateId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("IKSTemplateController: Patch template with templateId "+template.TemplateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	allowed, err := rbac_athentication.Authenticate(models.IBM, "clusterTemplate", template.TemplateId, "Update", token, utils.Context{})
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

	//==================================================================================
	ctx.SendLogs("IKSTemplateController: Patch template with id: "+template.TemplateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	beego.Info("IKSTemplateController: JSON Payload: ", template)

	err = ibm.UpdateTemplate(template, *ctx)
	if err != nil {
		ctx.SendLogs("IKSTemplateController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	statusCode, err := rbac_athentication.CreatePolicy(template.TemplateId, token, userInfo.UserId, userInfo.CompanyId, models.PUT, teams, models.IBM, *ctx)
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
	ctx.SendLogs("Ibm template of template id "+template.TemplateId+" updated", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "template updated successfully"}
	c.ServeJSON()
}

// @Title Delete
// @Description delete a templates
// @Param	token	header	string	token ""
// @Param	templateId	path	string	true	"Name of the template"
// @Success 200 {"msg": "template deleted successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "name is empty"}
// @Failure 500 {"error": "error msg"}
// @router /:templateId [delete]
func (c *IKSTemplateController) Delete() {

	id := c.GetString(":templateId")
	if id == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "template id is empty"}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)
	ctx.SendLogs("IKSTemplateController: Delete template with id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, id, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	allowed, err := rbac_athentication.Authenticate(models.IBM, "clusterTemplate", id, "Delete", token, utils.Context{})
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

	//==================================================================================

	err = ibm.DeleteTemplate(id, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("IKSTemplateController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("IKSTemplateController: Deleting template with templateId "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	//==========================RBAC Authentication==============================//

	status_code, err := rbac_athentication.DeletePolicy(models.IBM, id, token, utils.Context{})
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
	ctx.SendLogs("Ibm template of template id "+id+" deleted", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	//==================================================================================
	c.Data["json"] = map[string]string{"msg": "template deleted successfully"}
	c.ServeJSON()
}

// @Title Create Customer Template
// @Description create a new customer template
// @Param	token	header	string	token ""
// @Param	body	body	ibm.Template	true	"body for template content"
// @Success 200 {"msg": "template created successfully"}
// @Failure 409 {"error": "template with same name already exists"}
// @Failure 500 {"error": "error msg"}
// @router /customerTemplate [post]
func (c *IKSTemplateController) PostCustomerTemplate() {

	var template ibm.Template
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

	roleInfo, err := rbac_athentication.GetRole(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !ibm.CheckRole(roleInfo) {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, "", "", "")

	ctx.SendLogs("IKSTemplateController: Post new customer template with name: "+template.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err, id := ibm.CreateCustomerTemplate(template, *ctx)
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
// @Param	token	header	string	token ""
// @Success 200 {object} ibm.Template
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @router /customerTemplate/:templateId [get]
func (c *IKSTemplateController) GetCustomerTemplate() {

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

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
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

	ctx.SendLogs("IbmCustomerTemplateController: Get customer template  id : "+tempId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	template, err := ibm.GetCustomerTemplate(tempId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "no customer template exists for this id"}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("Ibm customer template of template id "+template.TemplateId+" fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = template
	c.ServeJSON()
}

// @Title Update customer templates
// @Description update an existing customer template
// @Param	token	header	string	token ""
// @Param	teams	header	string	token ""
// @Param	body	body	ibm.Template	true	"body for template content"
// @Success 200 {"msg": "customer template updated successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "no template exists with this name"}
// @Failure 500 {"error": "error msg"}
// @router /customerTemplate [put]
func (c *IKSTemplateController) PatchCustomerTemplate() {

	var template ibm.Template
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

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, template.TemplateId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Role Authentication=============================//

	roleInfo, err := rbac_athentication.GetRole(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if !ibm.CheckRole(roleInfo) {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//=============================================================================//

	ctx.SendLogs("ibmCustomerTemplateController: Patch template with template id : "+template.TemplateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err = ibm.UpdateCustomerTemplate(template, *ctx)
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

	ctx.SendLogs("ibm customer template of template id "+template.TemplateId+" updated", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": " customer template updated successfully"}
	c.ServeJSON()
}

// @Title Delete customer template
// @Description delete a customer template
// @Param	token	header	string	token ""
// @Param	templateId	path	string	true	"template id of the template"
// @Success 200 {"msg": "customer template deleted successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 500 {"error": "error msg"}
// @router /customerTemplate/:templateId [delete]
func (c *IKSTemplateController) DeleteCustomerTemplate() {

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

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, templateId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Role Authentication=============================//

	roleInfo, err := rbac_athentication.GetRole(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if !ibm.CheckRole(roleInfo) {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//=============================================================================//
	ctx.SendLogs("IbmCustomerTemplateController: Delete customer template with template Id "+templateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err = ibm.DeleteCustomerTemplate(templateId, *ctx)
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
// @Param	token	header	string	token ""
// @Success 200 {object} []ibm.Template
// @Failure 400 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /allCustomerTemplates [get]
func (c *IKSTemplateController) AllCustomerTemplates() {

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
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

	ctx.SendLogs("IKSTemplateController: GetAllCustomerTemplate.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	templates, err := ibm.GetAllCustomerTemplates(*ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("All Gcp Customer Template fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = templates
	c.ServeJSON()
}

// @Title   GetAllTemplateInfo
// @Description get all the templates info
// @Param	token	header	string	token ""
// @Success 200 {object} []ibm.TemplateMetadata
// @Failure 400 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /allTemplatesInfo [get]
func (c *IKSTemplateController) GetAllTemplateInfo() {

	ctx := new(utils.Context)
	ctx.SendLogs("IKSTemplateController:  Get Templates MetaData.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	err, data := rbac_athentication.GetAllAuthenticate("clusterTemplate", userInfo.CompanyId, token, models.IBM, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	//==================================================================================
	templates, err := ibm.GetTemplatesMetadata(*ctx, data, userInfo.CompanyId)
	if err != nil {
		ctx.SendLogs("IKSTemplateController: Internal server error "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("Ibm templates meta data fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = templates
	c.ServeJSON()
}

// @Title   GetAllCustomerTemplateInfo
// @Description get all the customer templates info
// @Param	token	header	string	token ""
// @Success 200 {object} []ibm.TemplateMetadata
// @Failure 400 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /allCustomerTemplatesInfo [get]
func (c *IKSTemplateController) GetAllCustomerTemplateInfo() {

	ctx := new(utils.Context)
	ctx.SendLogs("IKSTemplateController:  Get all customer Templates Info.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	err, data := rbac_athentication.GetAllAuthenticate("clusterTemplate", userInfo.CompanyId, token, models.IBM, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	//==================================================================================
	templates, err := ibm.GetCustomerTemplatesMetadata(*ctx, data, userInfo.CompanyId)
	if err != nil {
		ctx.SendLogs("IKSTemplateController: Internal server error "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("Ibm customer templates info fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = templates
	c.ServeJSON()
}
