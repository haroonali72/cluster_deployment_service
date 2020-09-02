package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"],
		beego.ControllerComments{
			Method:           "Patch",
			Router:           `/`,
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"],
		beego.ControllerComments{
			Method:           "Get",
			Router:           `/:infraId/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           `/:infraId/:forceDelete`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           `/all`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"],
		beego.ControllerComments{
			Method:           "GetRegions",
			Router:           `/getallregions/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"],
		beego.ControllerComments{
			Method:           "PostSSHKey",
			Router:           `/sshkey/:infraId/:keyname/:region`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"],
		beego.ControllerComments{
			Method:           "DeleteSSHKey",
			Router:           `/sshkey/:keyname`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"],
		beego.ControllerComments{
			Method:           "GetSSHKeys",
			Router:           `/sshkeys`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"],
		beego.ControllerComments{
			Method:           "StartCluster",
			Router:           `/start/:infraId`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"],
		beego.ControllerComments{
			Method:           "GetStatus",
			Router:           `/status/:infraId/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"],
		beego.ControllerComments{
			Method:           "TerminateCluster",
			Router:           `/terminate/:infraId/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOClusterController"],
		beego.ControllerComments{
			Method:           "ValidateProfile",
			Router:           `/validateProfile/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           `/`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"],
		beego.ControllerComments{
			Method:           "Patch",
			Router:           `/`,
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           `/:templateId`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"],
		beego.ControllerComments{
			Method:           "Get",
			Router:           `/:templateId/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           `/all`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"],
		beego.ControllerComments{
			Method:           "AllCustomerTemplates",
			Router:           `/allCustomerTemplates`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"],
		beego.ControllerComments{
			Method:           "GetAllCustomerTemplateInfo",
			Router:           `/allCustomerTemplatesInfo`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"],
		beego.ControllerComments{
			Method:           "GetAllTemplateInfo",
			Router:           `/allTemplatesInfo`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"],
		beego.ControllerComments{
			Method:           "PatchCustomerTemplate",
			Router:           `/customerTemplate`,
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"],
		beego.ControllerComments{
			Method:           "PostCustomerTemplate",
			Router:           `/customerTemplate`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"],
		beego.ControllerComments{
			Method:           "GetCustomerTemplate",
			Router:           `/customerTemplate/:templateId`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/do:DOTemplateController"],
		beego.ControllerComments{
			Method:           "DeleteCustomerTemplate",
			Router:           `/customerTemplate/:templateId`,
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
