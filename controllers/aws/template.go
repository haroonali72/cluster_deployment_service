package aws

import (
	"antelope/models"
	"antelope/models/aws"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"encoding/json"
	"github.com/astaxie/beego"
	"strings"
)

// Operations about AWS template [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/template/{cloud}/]
type AWSTemplateController struct {
	beego.Controller
}

// @Title Get
// @Description get template
// @Param	templateId	path	string	true	"Template Id of the template"
// @Param	token	header	string	token ""
// @Success 200 {object} aws.Template
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /:templateId/ [get]
func (c *AWSTemplateController) Get() {

	templateId := c.GetString(":templateId")
	if templateId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "templateId is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("token")
	if templateId == "" {
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
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, templateId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	allowed, err := rbac_athentication.Authenticate(models.AWS, "clusterTemplate", templateId, "View", token, *ctx)
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

	//=============================================================================//
	ctx.SendLogs("AWSTemplateController: Get template  id : "+templateId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	template, err := aws.GetTemplate(templateId, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "no template exists for this id"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = template
	c.ServeJSON()
}

// @Title Get All
// @Description get all the templates
// @Param	token	header	string	token ""
// @Success 200 {object} []aws.Template
// @Failure 400 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /all [get]
func (c *AWSTemplateController) GetAll() {

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

	//==========================RBAC Authentication==============================//

	err, data := rbac_athentication.GetAllAuthenticate("clusterTemplate", userInfo.CompanyId, token, models.AWS, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	//=============================================================================//
	ctx.SendLogs("AWSTemplateController: GetAll template.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	templates, err := aws.GetTemplates(*ctx, data)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = templates
	c.ServeJSON()
}

// @Title Create
// @Description create a new template
// @Param	token	header	string	token ""
// @Param	teams	header	string	teams ""
// @Param	body	body	aws.Template	true	"body for template content"
// @Success 200 {"msg": "template created successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 409 {"error": "template with same name already exists"}
// @Failure 500 {"error": "error msg"}
// @router / [post]
func (c *AWSTemplateController) Post() {
	var template aws.Template
	json.Unmarshal(c.Ctx.Input.RequestBody, &template)

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

	//==========================RBAC Authentication==============================//

	allowed, err := rbac_athentication.Evaluate("Create", token, *ctx)
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

	ctx.SendLogs("AWSTemplateController: Post new template with name: "+template.Name, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

	template.CompanyId = userInfo.CompanyId

	err, id := aws.CreateTemplate(template, *ctx)
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
	//==========================RBAC Policy Creation==============================//

	team := c.Ctx.Input.Header("teams")


	var teams []string
	if team != "" {
		teams = strings.Split(team, ";")
	}

	statusCode, err := rbac_athentication.CreatePolicy(id, token, userInfo.UserId, userInfo.CompanyId, models.POST, teams, models.AWS, *ctx)
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

	c.Data["json"] = map[string]string{"msg": "template generated successfully with id " + id}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing template
// @Param	token	header	string	token ""
// @Param	teams	header	string	token ""
// @Param	body	body	aws.Template	true	"body for template content"
// @Success 200 {"msg": "template updated successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "no template exists with this name"}
// @Failure 500 {"error": "error msg"}
// @router / [put]
func (c *AWSTemplateController) Patch() {
	var template aws.Template
	json.Unmarshal(c.Ctx.Input.RequestBody, &template)

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
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, template.TemplateId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	allowed, err := rbac_athentication.Authenticate(models.AWS, "clusterTemplate", template.TemplateId, "Update", token, *ctx)
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

	//=============================================================================//
	ctx.SendLogs("AWSTemplateController: Patch template with template id : "+template.TemplateId, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

	err = aws.UpdateTemplate(template, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "no template exists with this project id"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error":  err.Error()}
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

	c.Data["json"] = map[string]string{"msg": "template updated successfully"}
	c.ServeJSON()
}

// @Title Delete
// @Description delete a templates
// @Param	token	header	string	token ""
// @Param	templateId	path	string	true	"template id of the template"
// @Success 200 {"msg": "template deleted successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 500 {"error": "error msg"}
// @router /:templateId [delete]
func (c *AWSTemplateController) Delete() {

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

	//==========================RBAC Authentication==============================//
	allowed, err := rbac_athentication.Authenticate(models.AWS, "clusterTemplate", templateId, "Delete", token, *ctx)
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

	//=============================================================================//
	beego.Info("AWSTemplateController: Delete template with template Id ", templateId)

	err = aws.DeleteTemplate(templateId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	//==========================RBAC Authentication==============================//

	status_code, err := rbac_athentication.DeletePolicy(models.AWS, templateId, token, utils.Context{})
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

	//==================================================================================
	c.Data["json"] = map[string]string{"msg": "template deleted successfully"}
	c.ServeJSON()
}
