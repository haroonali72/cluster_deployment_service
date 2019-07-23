package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "Patch",
			Router:           `/`,
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           `/:projectId`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "Get",
			Router:           `/:projectId/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           `/all`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "StartCluster",
			Router:           `/start/:projectId`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "GetStatus",
			Router:           `/status/:projectId/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "TerminateCluster",
			Router:           `/terminate/:projectId/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"],
		beego.ControllerComments{
			Method:           "Patch",
			Router:           `/`,
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"],
		beego.ControllerComments{
			Method:           "Get",
			Router:           `/:templateId`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           `/:templateId`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           `/all`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
