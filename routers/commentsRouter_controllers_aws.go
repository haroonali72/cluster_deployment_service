package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "Patch",
            Router: `/`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "Get",
            Router: `/:projectId/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:projectId/:forceDelete`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/all`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "GetAMI",
            Router: `/amis/:amiId/:region`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "ApplyAgent",
            Router: `/applyagent/:projectId`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "EnableAutoScaling",
            Router: `/enablescaling/:projectId/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "GetAllMachines",
            Router: `/getallmachines`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "GetAllRegions",
            Router: `/getallregions/:cloud`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "GetZones",
            Router: `/getzones/:region`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "DeleteSSHKey",
            Router: `/sshkey/:keyname/:region`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "PostSSHKey",
            Router: `/sshkey/:projectId/:keyname/:region`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "GetSSHKeys",
            Router: `/sshkeys/:region`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "StartCluster",
            Router: `/start/:projectId`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "GetStatus",
            Router: `/status/:projectId/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "TerminateCluster",
            Router: `/terminate/:projectId/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSClusterController"],
        beego.ControllerComments{
            Method: "ValidateProfile",
            Router: `/validateProfile/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"],
        beego.ControllerComments{
            Method: "Patch",
            Router: `/`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:templateId`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"],
        beego.ControllerComments{
            Method: "Get",
            Router: `/:templateId/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/all`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"],
        beego.ControllerComments{
            Method: "AllCustomerTemplates",
            Router: `/allCustomerTemplates`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"],
        beego.ControllerComments{
            Method: "GetAllCustomerTemplateInfo",
            Router: `/allCustomerTemplatesInfo`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"],
        beego.ControllerComments{
            Method: "GetAllTemplateInfo",
            Router: `/allTemplatesInfo`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"],
        beego.ControllerComments{
            Method: "PatchCustomerTemplate",
            Router: `/customerTemplate`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"],
        beego.ControllerComments{
            Method: "PostCustomerTemplate",
            Router: `/customerTemplate`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"],
        beego.ControllerComments{
            Method: "GetCustomerTemplate",
            Router: `/customerTemplate/:templateId`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/aws:AWSTemplateController"],
        beego.ControllerComments{
            Method: "DeleteCustomerTemplate",
            Router: `/customerTemplate/:templateId`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
