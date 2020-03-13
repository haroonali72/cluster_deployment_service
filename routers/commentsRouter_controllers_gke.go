package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

    beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"],
        beego.ControllerComments{
            Method: "Patch",
            Router: `/`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"],
        beego.ControllerComments{
            Method: "Get",
            Router: `/:projectId/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:projectId/:forceDelete`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/all`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"],
        beego.ControllerComments{
            Method: "ApplyAgent",
            Router: `/applyagent/:projectId`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"],
        beego.ControllerComments{
            Method: "GetServerConfig",
            Router: `/config/:projectId`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"],
        beego.ControllerComments{
            Method: "StartCluster",
            Router: `/start/:projectId`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"],
        beego.ControllerComments{
            Method: "GetStatus",
            Router: `/status/:projectId/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/gke:GKEClusterController"],
        beego.ControllerComments{
            Method: "TerminateCluster",
            Router: `/terminate/:projectId/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/gke:GKETemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/gke:GKETemplateController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/gke:GKETemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/gke:GKETemplateController"],
        beego.ControllerComments{
            Method: "Patch",
            Router: `/`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/gke:GKETemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/gke:GKETemplateController"],
        beego.ControllerComments{
            Method: "Get",
            Router: `/:templateId`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/gke:GKETemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/gke:GKETemplateController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:templateId`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/gke:GKETemplateController"] = append(beego.GlobalControllerRouter["antelope/controllers/gke:GKETemplateController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/all`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
