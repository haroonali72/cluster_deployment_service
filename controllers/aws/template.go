package aws

import (
	"antelope/models/aws"
	"antelope/models/logging"
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
// @Success 200 {object} aws.Template
// @Failure 404 {"error": exception_message}
// @Failure 500 {"error": "internal server error "}
// @router /:templateId/ [get]
func (c *AWSTemplateController) Get() {
	templateId := c.GetString(":templateId")
	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, templateId)

	ctx.SendSDLog("AWSTemplateController: Get template  id : "+templateId, "info")

	if templateId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "template id is empty"}
		c.ServeJSON()
		return
	}

	template, err := aws.GetTemplate(templateId, *ctx)
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
// @Success 200 {object} []aws.Template
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /all [get]
func (c *AWSTemplateController) GetAll() {
	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "")

	ctx.SendSDLog("AWSTemplateController: GetAll template.", "info")

	templates, err := aws.GetAllTemplate(*ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = templates
	c.ServeJSON()
}

// @Title Create
// @Description create a new template
// @Param	body	body	aws.Template	true	"body for template content"
// @Success 200 {"msg": "template created successfully"}
// @Failure 409 {"error": "template with same name already exists"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router / [post]
func (c *AWSTemplateController) Post() {
	var template aws.Template
	json.Unmarshal(c.Ctx.Input.RequestBody, &template)
	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, "")
	ctx.SendSDLog("AWSTemplateController: Post new template with name: "+template.Name, "error")

	err, id := aws.CreateTemplate(template, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "template with same name already exists"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"msg": "template generated successfully with id " + id}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing template
// @Param	body	body	aws.Template	true	"body for template content"
// @Success 200 {"msg": "template updated successfully"}
// @Failure 404 {"error": "no template exists with this name"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router / [put]
func (c *AWSTemplateController) Patch() {
	var template aws.Template
	json.Unmarshal(c.Ctx.Input.RequestBody, &template)
	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, "")
	ctx.SendSDLog("AWSTemplateController: Patch template with template id : "+template.TemplateId, "error")

	err := aws.UpdateTemplate(template, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "no template exists with this project id"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"msg": "template updated successfully"}
	c.ServeJSON()
}

// @Title Delete
// @Description delete a templates
// @Param	templateId	path	string	true	"template id of the template"
// @Success 200 {"msg": "template deleted successfully"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /:templateId [delete]
func (c *AWSTemplateController) Delete() {
	templateId := c.GetString(":templateId")
	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, "")

	beego.Info("AWSTemplateController: Delete template with template Id ", templateId)

	if templateId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "template id is empty"}
		c.ServeJSON()
		return
	}

	err := aws.DeleteTemplate(templateId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"msg": "template deleted successfully"}
	c.ServeJSON()
}
