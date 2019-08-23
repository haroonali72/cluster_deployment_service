package gcp

import (
	"antelope/models/gcp"
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
// @Param	name	path	string	true	"Name of the template"
// @Success 200 {object} gcp.Template
// @Failure 404 {"error": exception_message}
// @Failure 500 {"error": "internal server error"}
// @router /:templateId [get]
func (c *GcpTemplateController) Get() {
	id := c.GetString(":templateId")
	beego.Info("GcpTemplateController: Get template with id: ", id)

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, id)
	ctx.SendSDLog("GcpTemplateController: Get template  id : "+id, "info")

	if id == "" {
		ctx.SendSDLog("GcpTemplateController: template id is empty", "error")
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "template id is empty"}
		c.ServeJSON()
		return
	}

	template, err := gcp.GetTemplate(id, *ctx)
	if err != nil {
		ctx.SendSDLog("GcpTemplateController :"+err.Error(), "error")
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
// @Success 200 {object} []gcp.Template
// @Failure 500 {"error": "internal server error"}
// @router /all [get]
func (c *GcpTemplateController) GetAll() {
	beego.Info("GcpTemplateController: GetAll template.")

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "")
	ctx.SendSDLog("GcpTemplateController: GetAll template.", "info")

	templates, err := gcp.GetAllTemplate(*ctx)
	if err != nil {
		ctx.SendSDLog("GcpTemplateController: internal server error "+err.Error(), "error")
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = templates
	c.ServeJSON()
}

// @Title Create
// @Description create a new template
// @Param	body	body	gcp.Template	true	"body for template content"
// @Success 200 {"msg": "template created successfully"}
// @Failure 409 {"error": "template with same name already exists"}
// @Failure 500 {"error": "internal server error"}
// @router / [post]
func (c *GcpTemplateController) Post() {
	var template gcp.Template
	json.Unmarshal(c.Ctx.Input.RequestBody, &template)

	beego.Info("GcpTemplateController: Post new template with name: ", template.Name)
	beego.Info("GcpTemplateController: JSON Payload: ", template)

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, "")
	ctx.SendSDLog("GcpTemplateController: Posting  new template .", "info")

	err, id := gcp.CreateTemplate(template, *ctx)
	if err != nil {
		ctx.SendSDLog("GcpTemplateController :"+err.Error(), "error")
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "template with same name already exists"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"msg": "template generated successfully with id " + id}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing template
// @Param	body	body	gcp.Template	true	"body for template content"
// @Success 200 {"msg": "template updated successfully"}
// @Failure 404 {"error": "no template exists with this name"}
// @Failure 500 {"error": "internal server error"}
// @router / [put]
func (c *GcpTemplateController) Patch() {
	var template gcp.Template
	json.Unmarshal(c.Ctx.Input.RequestBody, &template)

	beego.Info("GcpTemplateController: Patch template with id: ", template.TemplateId)
	beego.Info("GcpTemplateController: JSON Payload: ", template)

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, "")
	ctx.SendSDLog("GcpTemplateController: Patch template with templateId "+template.TemplateId, "info")

	err := gcp.UpdateTemplate(template, *ctx)
	if err != nil {
		ctx.SendSDLog("GcpTemplateController :"+err.Error(), "error")
		if strings.Contains(err.Error(), "does not exist") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "no template exists with this id"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"msg": "template updated successfully"}
	c.ServeJSON()
}

// @Title Delete
// @Description delete a templates
// @Param	name	path	string	true	"Name of the template"
// @Success 200 {"msg": "template deleted successfully"}
// @Failure 404 {"error": "name is empty"}
// @Failure 500 {"error": "internal server error"}
// @router /:templateId [delete]
func (c *GcpTemplateController) Delete() {
	id := c.GetString(":templateId")

	beego.Info("GcpTemplateController: Delete template with id: ", id)

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, "")
	ctx.SendSDLog("GcpTemplateController: deleting template with templateId "+id, "info")

	if id == "" {
		ctx.SendSDLog("GcpTemplateController: templateId is empty", "error")
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "name is empty"}
		c.ServeJSON()
		return
	}

	err := gcp.DeleteTemplate(id, *ctx)
	if err != nil {
		ctx.SendSDLog("GcpTemplateController :"+err.Error(), "error")
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"msg": "template deleted successfully"}
	c.ServeJSON()
}
