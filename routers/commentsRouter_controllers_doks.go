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
			Router:           `/:projectId/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           `/:projectId/:forceDelete`,
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
			Router:           `/applyagent/:projectId`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "GetServerConfig",
			Router:           `/config/:projectId`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "GetKubeConfig",
			Router:           `/kubeconfig/:projectId`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "StartCluster",
			Router:           `/start/:projectId`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "GetStatus",
			Router:           `/status/:projectId/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
		beego.ControllerComments{
			Method:           "TerminateCluster",
			Router:           `/terminate/:projectId/`,
			AllowHTTPMethods: []string{"post"},
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
