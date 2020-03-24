package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

    beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
        beego.ControllerComments{
            Method: "Patch",
            Router: `/`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
        beego.ControllerComments{
            Method: "Get",
            Router: `/:projectId/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:projectId/:forceDelete`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: `/all`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
        beego.ControllerComments{
            Method: "ApplyAgent",
            Router: `/applyagent/:projectId`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
        beego.ControllerComments{
            Method: "GetServerConfig",
            Router: `/config/:zone`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
        beego.ControllerComments{
            Method: "GetKubeConfig",
            Router: `/kubeconfig/{projectId}`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
        beego.ControllerComments{
            Method: "StartCluster",
            Router: `/start/:projectId`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
        beego.ControllerComments{
            Method: "GetStatus",
            Router: `/status/:projectId/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"] = append(beego.GlobalControllerRouter["antelope/controllers/doks:DOKSClusterController"],
        beego.ControllerComments{
            Method: "TerminateCluster",
            Router: `/terminate/:projectId/`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
