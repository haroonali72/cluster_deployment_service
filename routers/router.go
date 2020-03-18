// @APIVersion 1.0.0
// @Title antelope API
// @Description cloudPlex CNAP node pool solution
// @Contact info@cloudplex.io
// @TermsOfServiceUrl https://cloudplex.io/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"antelope/controllers"
	"antelope/controllers/aks"
	"antelope/controllers/aws"
	"antelope/controllers/azure"
	"antelope/controllers/customer_template"
	"antelope/controllers/do"
	"antelope/controllers/gcp"
	"antelope/controllers/gke"
	"antelope/controllers/op"
	"github.com/astaxie/beego"
)

func init() {
	ns := beego.NewNamespace("/antelope",
		beego.NSNamespace("/health",
			beego.NSInclude(
				&controllers.HealthController{},
			),
		),
		beego.NSNamespace("/customerTemplate",
			beego.NSInclude(
				&customer_template.CustomerTempalteController{},
			),
		),
		beego.NSNamespace("/template/aws",
			beego.NSInclude(
				&aws.AWSTemplateController{},
			),
		),
		beego.NSNamespace("/cluster/aws",
			beego.NSInclude(
				&aws.AWSClusterController{},
			),
		),
		beego.NSNamespace("/cluster/azure",
			beego.NSInclude(
				&azure.AzureClusterController{},
			),
		),
		beego.NSNamespace("/template/azure",
			beego.NSInclude(
				&azure.AzureTemplateController{},
			),
		),
		beego.NSNamespace("/cluster/gcp",
			beego.NSInclude(
				&gcp.GcpClusterController{},
			),
		),
		beego.NSNamespace("/template/gcp",
			beego.NSInclude(
				&gcp.GcpTemplateController{},
			),
		),
		beego.NSNamespace("/cluster/gke",
			beego.NSInclude(
				&gke.GKEClusterController{},
			),
		),
		beego.NSNamespace("/template/gke",
			beego.NSInclude(
				&gke.GKETemplateController{},
			),
		),
		beego.NSNamespace("/template/do",
			beego.NSInclude(
				&do.DOTemplateController{},
			),
		),
		beego.NSNamespace("/cluster/do",
			beego.NSInclude(
				&do.DOClusterController{},
			),
		),
		beego.NSNamespace("/template/op",
			beego.NSInclude(
				&op.OPTemplateController{},
			),
		),
		beego.NSNamespace("/cluster/op",
			beego.NSInclude(
				&op.OPClusterController{},
			),
		),
		beego.NSNamespace("/template/aks",
			beego.NSInclude(
				&aks.AKSTemplateController{},
			),
		),
		beego.NSNamespace("/cluster/aks",
			beego.NSInclude(
				&aks.AKSClusterController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
