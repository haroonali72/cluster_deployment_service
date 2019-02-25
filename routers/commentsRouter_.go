package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
		beego.ControllerComments{
			Method:           "Patch",
			Router:           `/`,
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           `/:environmentId`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
		beego.ControllerComments{
			Method:           "Get",
			Router:           `/:environmentId/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           `/all`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
		beego.ControllerComments{
			Method:           "GetSSHKeyPairs",
			Router:           `/sshkeys`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
		beego.ControllerComments{
			Method:           "StartCluster",
			Router:           `/start/:environmentId`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
		beego.ControllerComments{
			Method:           "GetStatus",
			Router:           `/status/:environmentId/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
		beego.ControllerComments{
			Method:           "TerminateCluster",
			Router:           `/terminate/:environmentId/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"],
		beego.ControllerComments{
			Method:           "Patch",
			Router:           `/`,
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"],
		beego.ControllerComments{
			Method:           "Get",
			Router:           `/:name`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           `/:name`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           `/all`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
