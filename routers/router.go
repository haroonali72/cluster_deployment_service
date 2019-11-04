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
	"antelope/controllers/aws"
	"antelope/controllers/azure"
	"antelope/controllers/gcp"
	"antelope/controllers/register_customer_templates"
	"github.com/astaxie/beego"
)

func init() {
	ns := beego.NewNamespace("/antelope",
		beego.NSNamespace("/health",
			beego.NSInclude(
				&controllers.HealthController{},
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
		beego.NSNamespace("/customerTemplate",
			beego.NSInclude(
				&register_customer_templates.CustomerTemplateController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
