package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["antelope/controllers/register_customer_templates:CustomerTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/register_customer_templates:CustomerTemplateController"],
		beego.ControllerComments{
			Method:           "RegisterCustomerTemplate",
			Router:           `/register/customerTemplates/:companyId`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
