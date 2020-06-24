package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["antelope/controllers/default:DefaultTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/default:DefaultTemplateController"],
		beego.ControllerComments{
			Method:           "Get",
			Router:           `/:cloudtype`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
