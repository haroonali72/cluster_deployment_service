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
	)
	beego.AddNamespace(ns)
}
