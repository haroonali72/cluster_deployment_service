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
			Method:           "Get",
			Router:           `/:infraId/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           `/:infraId/:forceDelete`,
			AllowHTTPMethods: []string{"delete"},
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
			Method:           "GetAllMachines",
			Router:           `/getallmachines/:zone`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "GetAllRegions",
			Router:           `/getallregions`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "GetZones",
			Router:           `/getzones/:region`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "GetServiceAccounts",
			Router:           `/serviceaccounts`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "DeleteSSHKey",
			Router:           `/sshkey/:keyname`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "PostSSHKey",
			Router:           `/sshkey/:keyname/:username/:infraId`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "GetSSHKeys",
			Router:           `/sshkeys`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "StartCluster",
			Router:           `/start/:infraId`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "GetStatus",
			Router:           `/status/:infraId/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "TerminateCluster",
			Router:           `/terminate/:infraId/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpClusterController"],
		beego.ControllerComments{
			Method:           "ValidateProfile",
			Router:           `/validateProfile`,
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

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"],
		beego.ControllerComments{
			Method:           "AllCustomerTemplates",
			Router:           `/allCustomerTemplates`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"],
		beego.ControllerComments{
			Method:           "GetAllCustomerTemplateInfo",
			Router:           `/allCustomerTemplatesInfo`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"],
		beego.ControllerComments{
			Method:           "GetAllTemplateInfo",
			Router:           `/allTemplatesInfo`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"],
		beego.ControllerComments{
			Method:           "PatchCustomerTemplate",
			Router:           `/customerTemplate`,
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"],
		beego.ControllerComments{
			Method:           "PostCustomerTemplate",
			Router:           `/customerTemplate`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"],
		beego.ControllerComments{
			Method:           "GetCustomerTemplate",
			Router:           `/customerTemplate/:templateId`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/gcp:GcpTemplateController"],
		beego.ControllerComments{
			Method:           "DeleteCustomerTemplate",
			Router:           `/customerTemplate/:templateId`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
