package customer_template

import (
	rbac "antelope/models/rbac_authentication"
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
// @Param token header string token ""
// @Success 200 {"msg": "template created successfully"}
// @Failure 404 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /register [post]
func (c *CustomerTempalteController) Post() {

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	statusCode,userInfo, err := rbac.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	awstemplates, azuretemplates, gcptemplates, dotemplates, ibmtemplates, err := register_customer_template.GetCustomerTemplate(*ctx)
	if err != nil {

		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	err = register_customer_template.RegisterCustomerTemplate(awstemplates, azuretemplates, gcptemplates, dotemplates, ibmtemplates, userInfo.CompanyId, *ctx)
	if err != nil {

		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	//==========================RBAC Policy Creation==============================//

	beego.Info("creating policy " + token + " user id " + userInfo.UserId + " company id " + userInfo.CompanyId)

	err = register_customer_template.CreatePolicy(awstemplates, azuretemplates, gcptemplates, dotemplates, ibmtemplates, token, *ctx)
	if err != nil {
		//beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = map[string]string{"msg": "templates registered successfully "}
	c.ServeJSON()
}
