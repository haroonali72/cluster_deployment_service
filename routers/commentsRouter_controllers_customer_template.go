package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["antelope/controllers/customer_template:CustomerTempalteController"] = append(beego.GlobalControllerRouter["antelope/controllers/customer_template:CustomerTempalteController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/register/:companyId`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/customer_template:CustomerTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/customer_template:CustomerTemplateController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/:companyId`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
