package customer_template

import (
	"antelope/models/register_customer_template"
	"antelope/models/utils"
	"github.com/astaxie/beego"
)

// customer template endpoint
type CustomerTempalteController struct {
	beego.Controller
}

// @Title Post
// @Description register customer templates
// @Param	companyId	path	string	true	"Company Id"
// @Success 200 {"msg": "template created successfully"}
// @Failure 404 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /register/:companyId [post]
func (c *CustomerTempalteController) Post() {

	companyId := c.GetString(":companyId")
	if companyId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "company id is empty"}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, "", companyId, "")

	awstemplates, azuretemplates, gcptemplates, err := register_customer_template.GetCustomerTemplate(*ctx)
	if err != nil {

		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	err = register_customer_template.RegisterAWSCustomerTemplate(awstemplates, azuretemplates, gcptemplates, companyId, *ctx)
	if err != nil {

		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"msg": "templates registered successfully "}
	c.ServeJSON()
}
