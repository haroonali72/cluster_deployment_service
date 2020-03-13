package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"],
		beego.ControllerComments{
			Method:           "Patch",
			Router:           `/`,
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"],
		beego.ControllerComments{
			Method:           "Get",
			Router:           `/:projectId/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           `/:projectId/:forceDelete`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"],
		beego.ControllerComments{
			Method:           "GetAllMachines",
			Router:           `/getallmachines`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"],
		beego.ControllerComments{
			Method:           "GetRegions",
			Router:           `/getallregions`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           `/all`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"],
		beego.ControllerComments{
			Method:           "GetInstances",
			Router:           `/getAllInstances`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"],
		beego.ControllerComments{
			Method:           "GetCores",
			Router:           `/machine/info`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"],
		beego.ControllerComments{
			Method:           "DeleteSSHKey",
			Router:           `/sshkey/:keyname`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"],
		beego.ControllerComments{
			Method:           "PostSSHKey",
			Router:           `/sshkey/:keyname/:projectId`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"],
		beego.ControllerComments{
			Method:           "GetSSHKeys",
			Router:           `/sshkeys`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"],
		beego.ControllerComments{
			Method:           "StartCluster",
			Router:           `/start/:projectId`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"],
		beego.ControllerComments{
			Method:           "GetStatus",
			Router:           `/status/:projectId/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureClusterController"],
		beego.ControllerComments{
			Method:           "TerminateCluster",
			Router:           `/terminate/:projectId/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"],
		beego.ControllerComments{
			Method:           "Patch",
			Router:           `/`,
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"],
		beego.ControllerComments{
			Method:           "Get",
			Router:           `/:templateId`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           `/:templateId`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           `/all`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"],
		beego.ControllerComments{
			Method:           "AllCustomerTemplates",
			Router:           `/allCustomerTemplates`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"],
		beego.ControllerComments{
			Method:           "GetAllCustomerTemplateInfo",
			Router:           `/allCustomerTemplatesInfo`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"],
		beego.ControllerComments{
			Method:           "GetAllTemplateInfo",
			Router:           `/allTemplatesInfo`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"],
		beego.ControllerComments{
			Method:           "PostCustomerTemplate",
			Router:           `/customerTemplate`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"],
		beego.ControllerComments{
			Method:           "PatchCustomerTemplate",
			Router:           `/customerTemplate`,
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"],
		beego.ControllerComments{
			Method:           "GetCustomerTemplate",
			Router:           `/customerTemplate/:templateId`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/azure:AzureTemplateController"],
		beego.ControllerComments{
			Method:           "DeleteCustomerTemplate",
			Router:           `/customerTemplate/:templateId`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
