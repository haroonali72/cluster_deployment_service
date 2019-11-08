package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["antelope/controllers/customer_template:CustomerTempalteController"] = append(beego.GlobalControllerRouter["antelope/controllers/customer_template:CustomerTempalteController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/register/:token`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
