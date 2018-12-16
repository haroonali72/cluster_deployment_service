package aws

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"strings"
	"antelope/models/aws"
)

// Operations about AWS template [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/template/{cloud}/]
type AWSTemplateController struct {
	beego.Controller
}

// @Title Get
// @Description get template
// @Param	name	path	string	true	"Name of the template"
// @Success 200 {object} aws.Template
// @Failure 404 {"error": exception_message}
// @Failure 500 {"error": "internal server error"}
// @router /:name [get]
func (c *AWSTemplateController) Get() {
	name := c.GetString(":name")

	beego.Info("AWSTemplateController: Get template with name: ", name)

	if name == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "name is empty"}
		c.ServeJSON()
		return
	}

	template, err := aws.GetTemplate(name)
	if err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "no template exists for this name"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = template
	c.ServeJSON()
}

// @Title Get All
// @Description get all the templates
// @Success 200 {object} []aws.Template
// @Failure 500 {"error": "internal server error"}
// @router /all [get]
func (c *AWSTemplateController) GetAll() {
	beego.Info("AWSTemplateController: GetAll template.")

	templates, err := aws.GetAllTemplate()
	if err != nil {
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
// @Param	body	body	aws.Template	true	"body for template content"
// @Success 201 {"msg": "template created successfully"}
// @Failure 409 {"error": "template with same name already exists"}
// @Failure 500 {"error": "internal server error"}
// @router / [post]
func (c *AWSTemplateController) Post() {
	var template aws.Template
	json.Unmarshal(c.Ctx.Input.RequestBody, &template)

	beego.Info("AWSTemplateController: Post new template with name: ", template.Name)
	beego.Info("AWSTemplateController: JSON Payload: ", template)

	err := aws.CreateTemplate(template)
	if err != nil {
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

	c.Data["json"] = map[string]string{"msg": "template added successfully"}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing template
// @Param	body	body	aws.Template	true	"body for template content"
// @Success 200 {"msg": "template updated successfully"}
// @Failure 404 {"error": "no template exists with this name"}
// @Failure 500 {"error": "internal server error"}
// @router / [put]
func (c *AWSTemplateController) Patch() {
	var template aws.Template
	json.Unmarshal(c.Ctx.Input.RequestBody, &template)

	beego.Info("AWSTemplateController: Patch template with name: ", template.Name)
	beego.Info("AWSTemplateController: JSON Payload: ", template)

	err := aws.UpdateTemplate(template)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "no template exists with this name"}
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
// @router /:name [delete]
func (c *AWSTemplateController) Delete() {
	name := c.GetString(":name")

	beego.Info("AWSTemplateController: Delete template with name: ", name)

	if name == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "name is empty"}
		c.ServeJSON()
		return
	}

	err := aws.DeleteTemplate(name)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"msg": "template deleted successfully"}
	c.ServeJSON()
}
