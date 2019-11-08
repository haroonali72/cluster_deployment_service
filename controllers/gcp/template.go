package gcp

import (
	"antelope/models"
	"antelope/models/gcp"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"encoding/json"
	"github.com/astaxie/beego"
	"strings"
)

// Operations about Gcp template [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/template/{cloud}/]
type GcpTemplateController struct {
	beego.Controller
}

// @Title Get
// @Description get template
// @Param	token	header	string	token ""
// @Param	templateId	path	string	true	"Name of the template"
// @Success 200 {object} gcp.Template
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @router /:templateId [get]
func (c *GcpTemplateController) Get() {

	id := c.GetString(":templateId")
	if id == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "template Id is empty"}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)

	beego.Info("GcpTemplateController: Get template with id: ", id)

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
	ctx.SendLogs("GcpTemplateController: Get template  id : "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	allowed, err := rbac_athentication.Authenticate(models.GCP, "clusterTemplate", id, "View", token, utils.Context{})
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

	template, err := gcp.GetTemplate(id, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("GcpTemplateController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "no template exists for this id"}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("Azure template of template id "+template.TemplateId+" fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = template
	c.ServeJSON()
}

// @Title Get All
// @Description get all the templates
// @Param	token	header	string	token ""
// @Success 200 {object} []gcp.Template
// @Failure 400 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /all [get]
func (c *GcpTemplateController) GetAll() {
	beego.Info("GcpTemplateController: GetAll template.")

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpTemplateController: GetAll template.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	err, data := rbac_athentication.GetAllAuthenticate("clusterTemplate", userInfo.CompanyId, token, models.GCP, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	//==================================================================================

	templates, err := gcp.GetTemplates(*ctx, data,userInfo.CompanyId)
	if err != nil {
		ctx.SendLogs("GcpTemplateController: internal server error "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("Azure templates fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = templates
	c.ServeJSON()
}

// @Title Create
// @Description create a new template
// @Param	body	body	gcp.Template	true	"body for template content"
// @Param	token	header	string	token ""
// @Param	teams	header	string	teams ""
// @Success 200 {"msg": "template created successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 409 {"error": "template with same name already exists"}
// @Failure 500 {"error": "error msg"}
// @router / [post]
func (c *GcpTemplateController) Post() {

	var template gcp.Template
	json.Unmarshal(c.Ctx.Input.RequestBody, &template)

	beego.Info("GcpTemplateController: Post new template with name: ", template.Name)
	beego.Info("GcpTemplateController: JSON Payload: ", template)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, template.TemplateId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpTemplateController: Posting  new template .", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	userInfo.CompanyId = template.CompanyId

	err, id := gcp.CreateTemplate(template, *ctx)
	if err != nil {
		ctx.SendLogs("GcpTemplateController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

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

	statusCode, err := rbac_athentication.CreatePolicy(id, token, userInfo.UserId, userInfo.CompanyId, models.POST, teams, models.GCP, *ctx)
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
	ctx.SendLogs("Azure template of template id "+template.TemplateId+" created", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "template generated successfully with id " + id}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing template
// @Param	token	header	string	token ""
// @Param	teams	header	string	teams ""
// @Param	body	body	gcp.Template	true	"body for template content"
// @Success 200 {"msg": "template updated successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "no template exists with this name"}
// @Failure 500 {"error": "error msg"}
// @router / [put]
func (c *GcpTemplateController) Patch() {

	var template gcp.Template
	json.Unmarshal(c.Ctx.Input.RequestBody, &template)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
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
	ctx.SendLogs("GcpTemplateController: Patch template with templateId "+template.TemplateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	allowed, err := rbac_athentication.Authenticate(models.GCP, "clusterTemplate", template.TemplateId, "Update", token, utils.Context{})
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
	beego.Info("GcpTemplateController: Patch template with id: ", template.TemplateId)
	beego.Info("GcpTemplateController: JSON Payload: ", template)

	err = gcp.UpdateTemplate(template, *ctx)
	if err != nil {
		ctx.SendLogs("GcpTemplateController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	statusCode, err := rbac_athentication.CreatePolicy(template.TemplateId, token, userInfo.UserId, userInfo.CompanyId, models.PUT, teams, models.AWS, *ctx)
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
	ctx.SendLogs("Azure template of template id "+template.TemplateId+" updated", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
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
func (c *GcpTemplateController) Delete() {

	id := c.GetString(":templateId")
	if id == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "template id is empty"}
		c.ServeJSON()
		return
	}

	beego.Info("GcpTemplateController: Delete template with id: ", id)

	ctx := new(utils.Context)

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
	allowed, err := rbac_athentication.Authenticate(models.GCP, "clusterTemplate", id, "Delete", token, utils.Context{})
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

	err = gcp.DeleteTemplate(id,userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("GcpTemplateController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("GcpTemplateController: deleting template with templateId "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	//==========================RBAC Authentication==============================//

	status_code, err := rbac_athentication.DeletePolicy(models.GCP, id, token, utils.Context{})
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
	ctx.SendLogs("Azure template of template id "+id+" deleted", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	//==================================================================================
	c.Data["json"] = map[string]string{"msg": "template deleted successfully"}
	c.ServeJSON()
}

// @Title Create Customer Template
// @Description create a new customer template
// @Param	token	header	string	token ""
// @Param	body	body	gcp.Template	true	"body for template content"
// @Success 200 {"msg": "template created successfully"}
// @Failure 409 {"error": "template with same name already exists"}
// @Failure 500 {"error": "error msg"}
// @router /create/customerTemplate [post]
func (c *GcpTemplateController) PostCustomerTemplate() {

	var template gcp.Template
	json.Unmarshal(c.Ctx.Input.RequestBody, &template)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
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
	if !gcp.CheckRole(roleInfo) {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, "", "", "")

	ctx.SendLogs("AWSTemplateController: Post new customer template with name: "+template.Name, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

	err, id := gcp.CreateCustomerTemplate(template, *ctx)
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
