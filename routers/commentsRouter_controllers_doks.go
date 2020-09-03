package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "Patch",
			Router:           `/`,
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "Get",
			Router:           `/:infraId/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           `/:infraId/:forceDelete`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           `/all`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "ApplyAgent",
			Router:           `/applyagent/:infraId`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "GetServerConfig",
			Router:           `/config`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "GetKubeConfig",
			Router:           `/kubeconfig/:infraId`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "StartCluster",
			Router:           `/start/:infraId`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "GetStatus",
			Router:           `/status/:infraId/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "TerminateCluster",
			Router:           `/terminate/:infraId/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "PatchRunningCluster",
			Router:           `/update/:infraId`,
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"],
		beego.ControllerComments{
			Method:           "Patch",
			Router:           `/`,
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           `/:templateId`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"],
		beego.ControllerComments{
			Method:           "Get",
			Router:           `/:templateId/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           `/all`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"],
		beego.ControllerComments{
			Method:           "AllCustomerTemplates",
			Router:           `/allCustomerTemplates`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"],
		beego.ControllerComments{
			Method:           "GetAllCustomerTemplateInfo",
			Router:           `/allCustomerTemplatesInfo`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"],
		beego.ControllerComments{
			Method:           "GetAllTemplateInfo",
			Router:           `/allTemplatesInfo`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"],
		beego.ControllerComments{
			Method:           "PatchCustomerTemplate",
			Router:           `/customerTemplate`,
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"],
		beego.ControllerComments{
			Method:           "PostCustomerTemplate",
			Router:           `/customerTemplate`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"],
		beego.ControllerComments{
			Method:           "GetCustomerTemplate",
			Router:           `/customerTemplate/:templateId`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSTemplateController"],
		beego.ControllerComments{
			Method:           "DeleteCustomerTemplate",
			Router:           `/customerTemplate/:templateId`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
