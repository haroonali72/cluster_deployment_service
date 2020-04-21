package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

    beego.GlobalControllerRouter["antelope/controllers/op:OPClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/op:OPClusterController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/op:OPClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/op:OPClusterController"],
        beego.ControllerComments{
            Method: "Patch",
            Router: `/`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/op:OPClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/op:OPClusterController"],
        beego.ControllerComments{
            Method: "Get",
            Router: `/:projectId/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/op:OPClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/op:OPClusterController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:projectId/:forceDelete`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/op:OPClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/op:OPClusterController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/all`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/op:OPClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/op:OPClusterController"],
        beego.ControllerComments{
            Method: "Validate",
            Router: `/checkCluster/:projectId/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/op:OPTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/op:OPTemplateController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/op:OPTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/op:OPTemplateController"],
        beego.ControllerComments{
            Method: "Patch",
            Router: `/`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/op:OPTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/op:OPTemplateController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:templateId`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/op:OPTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/op:OPTemplateController"],
        beego.ControllerComments{
            Method: "Get",
            Router: `/:templateId/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/op:OPTemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/op:OPTemplateController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/all`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
