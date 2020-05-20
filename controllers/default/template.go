package _default

import (
	"antelope/models"
	"antelope/models/aks"
	"antelope/models/aws"
	"antelope/models/azure"
	"antelope/models/do"
	"antelope/models/doks"
	"antelope/models/gcp"
	"antelope/models/gke"
	"antelope/models/iks"
	"antelope/models/op"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"github.com/astaxie/beego"
	"strings"
)

type DefaultTemplateController struct {
	beego.Controller
}

// @Title Get
// @Description get template
// @Param	X-Auth-Token	header	string	true "token"
// @Param	cloudtype	path	string	true	"type of cloud provider"
// @Success 200 {object} aws.Template
// @Failure 400 {"error": "cloud type must not be empty"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "No template exists for this cloud type"}
// @router /:cloudtype [get]
func (c *DefaultTemplateController) Get() {

	cloudtype := c.GetString(":cloudtype")
	if cloudtype == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "Cloudtype is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "Token is empty"}
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
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI,"" , userInfo.UserId, userInfo.CompanyId)

	//====================================================================================//
	ctx.SendLogs("DefaultTemplateController: Get template for: "+cloudtype, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	if strings.ToLower(cloudtype) == string(models.AWS) {
		template, err := aws.GetAWSDefault(*ctx)
		if err != nil {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "No default template exists for this cloud"}
			c.ServeJSON()
			return
		}
		c.Data["json"] = template
		c.ServeJSON()
	}else if strings.ToLower(cloudtype) == string(models.Azure) {
		template, err := azure.GetAzureDefault(*ctx)
		if err != nil {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "No default template exists for this cloud"}
			c.ServeJSON()
			return
		}
		c.Data["json"] = template
		c.ServeJSON()
	} else if strings.ToLower(cloudtype) == string(models.DO) {
		template, err := do.GetDoDefault(*ctx)
		if err != nil {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "No default template exists for this cloud"}
			c.ServeJSON()
			return
		}
		c.Data["json"] = template
		c.ServeJSON()
	}else if strings.ToLower(cloudtype) == string(models.GCP) {
		template, err := gcp.GetGcpDefault(*ctx)
		if err != nil {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "No default template exists for this cloud"}
			c.ServeJSON()
			return
		}
		c.Data["json"] = template
		c.ServeJSON()
	}else if strings.ToLower(cloudtype) == string(models.OP) {
		template, err := op.GetOPDefault(*ctx)
		if err != nil {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "No default template exists for this cloud"}
			c.ServeJSON()
			return
		}
		c.Data["json"] = template
		c.ServeJSON()
	}else if strings.ToLower(cloudtype) == string(models.DOKS) {
		template, err := doks.GetDOKSDefault(*ctx)
		if err != nil {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "No default template exists for this cloud"}
			c.ServeJSON()
			return
		}
		c.Data["json"] = template
		c.ServeJSON()
	}else if strings.ToLower(cloudtype) == string(models.IKS) {
		template, err := iks.GetIKSDefault(*ctx)
		if err != nil {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "No default template exists for this cloud"}
			c.ServeJSON()
			return
		}
		c.Data["json"] = template
		c.ServeJSON()
	}else if strings.ToLower(cloudtype) == string(models.AKS){
		template, err := aks.GetAKSDefault(*ctx)
		if err != nil {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "No default template exists for this cloud"}
			c.ServeJSON()
			return
		}
		c.Data["json"] = template
		c.ServeJSON()
	}else if strings.ToLower(cloudtype) == string(models.GKE) {
		template, err := gke.GetGKEDefault(*ctx)
		if err != nil {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "No default template exists for this cloud"}
			c.ServeJSON()
			return
		}
		c.Data["json"] = template
		c.ServeJSON()
	}
}
