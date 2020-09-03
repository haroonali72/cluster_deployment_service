package op

import (
	"antelope/models"
	"antelope/models/op"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/go-playground/validator/v10"
	"strings"
	"time"
)

// Operations about OP cluster [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/cluster/{cloud}/]
type OPClusterController struct {
	beego.Controller
}

// @Title Get
// @Description Get cluster against the InfraId
// @Param	InfraId	path	string	true	"Id of the Infrastructure"
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 200 {object} op.Cluster_Def
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /:InfraId/ [get]
func (c *OPClusterController) Get() {
	ctx := new(utils.Context)
	ctx.SendLogs("OPClusterController: Get cluster", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	InfraId := c.GetString(":InfraId")
	if InfraId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "Infrastructure id is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbac_athentication.Authenticate(models.OP, "cluster", InfraId, "View", token, *ctx)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructure id"}
			c.ServeJSON()
			return
		}
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": "Bad Request: " + err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//====================================================================================//

	ctx.SendLogs("OPClusterController: Get cluster with Infrastructure id: "+InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := op.GetCluster(InfraId, userInfo.CompanyId, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("OPClusterController: Cluster of Infrastructure "+InfraId+" fetched", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs(" OP cluster "+cluster.Name+" of Infrastructure Id: "+cluster.InfraId+" fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Create
// @Description add a new cluster
// @Param	body	body 	op.Cluster_Def		true	"Body for cluster content"
// @Param	teams	header	string	false "teams"
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 201 {"msg": "Cluster created successfully"}
// @Success 400 {"msg": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Cluster against this Infrastructure already exists"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [post]
func (c *OPClusterController) Post() {

	var cluster op.Cluster_Def
	ctx := new(utils.Context)
	ctx.SendLogs("OPClusterController: Add cluster", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "Error while unmarshalling " + err.Error()}
		c.ServeJSON()
		return
	}
	cluster.CreationDate = time.Now()

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
		c.ServeJSON()
		return
	}

	validate := validator.New()
	err = validate.Struct(cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	teams := c.Ctx.Input.Header("teams")

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, cluster.InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.OP, "cluster", cluster.InfraId, "Create", token, *ctx)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructure id"}
			c.ServeJSON()
			return
		}
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//=============================================================================//

	ctx.SendLogs("OPClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster.CompanyId = userInfo.CompanyId

	ctx.SendLogs("OPClusterController: Adding new cluster "+cluster.Name+" in Infrastructure "+cluster.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err = op.CreateCluster(cluster, *ctx, token, teams)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "Cluster against this Infrastructure id  already exists"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("OPClusterController: New cluster "+cluster.Name+" added in Infrastructure "+cluster.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs(" OP cluster "+cluster.Name+" added in Infrastructure "+cluster.InfraId, models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = map[string]string{"msg": "Cluster added successfully"}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing cluster
// @Param	X-Auth-Token	header	string	true "token"
// @Param	teams	header	string	false 'teams"
// @Param	body	body 	op.Cluster_Def	true	"body for cluster content"
// @Success 200 {"msg": "Cluster updated successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Cluster is in Creating/Created/Terminating/TerminationFailed state"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [put]
func (c *OPClusterController) Patch() {
	var cluster op.Cluster_Def
	ctx := new(utils.Context)
	ctx.SendLogs("OPClusterController: Update Cluster", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "Error while unmarshalling " + err.Error()}
		c.ServeJSON()
		return
	}
	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, cluster.InfraId, userInfo.CompanyId, userInfo.UserId)

	savedCluster, err := op.GetCluster(cluster.InfraId, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if savedCluster.Status == (models.Deploying) {
		ctx.SendLogs("OPClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if savedCluster.Status == (models.Terminating) {
		ctx.SendLogs("OPClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if savedCluster.Status == (models.ClusterCreated) {
		ctx.SendLogs("OPClusterController: Cluster is in created state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": " Cluster is already in a created state"}
		c.ServeJSON()
		return
	} else if savedCluster.Status == (models.ClusterTerminationFailed) {
		ctx.SendLogs("OPClusterController: Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": " Cluster creation is in termination failed state"}
		c.ServeJSON()
		return
	}

	validate := validator.New()
	err = validate.Struct(cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	teams := c.Ctx.Input.Header("teams")

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.OP, "cluster", cluster.InfraId, "Update", token, *ctx)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructure id"}
			c.ServeJSON()
			return
		}
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//=============================================================================//

	ctx.SendLogs("OPClusterController:Update cluster "+cluster.Name+" of Infrastructure"+cluster.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs("OPClusterController: Patch cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	if strings.ToLower(string(cluster.Status)) != strings.ToLower(string(models.New)) && cluster.Status != models.ClusterCreationFailed && cluster.Status != models.ClusterTerminated {
		ctx.SendLogs("OPClusterController : Cluster is in "+string(cluster.Status)+" state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Can't Update.Cluster is in " + string(cluster.Status) + " state"}
		c.ServeJSON()
		return
	}
	err = op.UpdateCluster(cluster, true, *ctx, teams, token)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "no cluster exists with this name"}
			c.ServeJSON()
			return
		}
		if strings.Contains(err.Error(), "Cluster is in created state") {
			c.Ctx.Output.SetStatus(400)
			c.Data["json"] = map[string]string{"error": "Cluster is in created state"}
			c.ServeJSON()
			return
		}
		if strings.Contains(err.Error(), "cluster is in deploying state") {
			c.Ctx.Output.SetStatus(400)
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		}
		if strings.Contains(err.Error(), "cluster is in terminating state") {
			c.Ctx.Output.SetStatus(400)
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" OP cluster "+cluster.Name+" of Infrastructure Id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "Cluster updated successfully"}
	c.ServeJSON()
}

// @Title Get All
// @Description get all the company's clusters
// @Param	X-Auth-Token	header	string	true "token"
// @Success 200 {object} []op.Cluster_Def
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /all [get]
func (c *OPClusterController) GetAll() {
	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, err, data := rbac_athentication.GetAllAuthenticate("cluster", userInfo.CompanyId, token, models.OP, *ctx)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructure id"}
			c.ServeJSON()
			return
		}
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	//====================================================================================//
	ctx.SendLogs("OPClusterController: GetAll clusters.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.Data.Company = userInfo.CompanyId
	clusters, err := op.GetAllCluster(*ctx, data)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" All OP clusters fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = clusters
	c.ServeJSON()
}

// @Title Delete
// @Description delete a cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	InfraId	path 	string	true	"Infrastructure id of the cluster"
// @Param	forceDelete path    boolean	true    "Forcefully delete cluster"
// @Success 204 {"msg": "Cluster deleted successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in Creating/Created/Terminating/TerminationFailed state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /:InfraId/:forceDelete [delete]
func (c *OPClusterController) Delete() {
	ctx := new(utils.Context)

	id := c.GetString(":InfraId")
	if id == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "Infrastructure id is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
		c.ServeJSON()
		return
	}
	forceDelete, err := c.GetBool(":forceDelete")
	if err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, id, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.OP, "cluster", id, "Delete", token, *ctx)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructure id"}
			c.ServeJSON()
			return
		}
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//=============================================================================//

	ctx.SendLogs("OPClusterController: Delete cluster with Infrastructure id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := op.GetCluster(id, userInfo.CompanyId, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if cluster.Status == "Cluster Created" && !forceDelete {
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in created state"}
		c.ServeJSON()
		return
	}

	if cluster.Status == (models.Deploying) && !forceDelete {
		ctx.SendLogs("Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	}

	if cluster.Status == (models.Terminating) && !forceDelete {
		ctx.SendLogs("cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	}
	if cluster.Status == (models.ClusterTerminationFailed) {
		ctx.SendLogs("OPClusterController: Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "cluster is in termination failed state"}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("OPClusterController: Deleting cluster of Infrastructure "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err = op.DeleteCluster(id, userInfo.CompanyId, *ctx, token)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" OP cluster "+cluster.Name+" of Infrastructure id: "+cluster.InfraId+" deleted ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster deleted successfully"}
	c.ServeJSON()
}

// @Title ValidateCluster
// @Description validates a cluster
// @Param	token	header	string	true "Token"
// @Param	InfraId	path	string	true	"Id of the Infrastructure"
// @Success 200 {"msg": "Cluster is created successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /checkCluster/:InfraId/ [get]
func (c *OPClusterController) Validate() {
	id := c.GetString(":InfraId")
	if id == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "Infrastructure id is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
		c.ServeJSON()
		return
	}

	stausCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(stausCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "Get", c.Ctx.Request.RequestURI, id, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.OP, "cluster", id, "View", token, *ctx)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructure id"}
			c.ServeJSON()
			return
		}
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//=============================================================================//

	ctx.SendLogs("OPClusterController: Check cluster with Infrastructure id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	if id == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "Infrastructure id is empty"}
		c.ServeJSON()
		return
	}

	err = op.CheckCluster(id, userInfo.CompanyId, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("OPClusterController: Cluster of Infrastructure"+id+" deleted", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs(" OP cluster of Infrastructure"+id+" deleted ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = map[string]string{"msg": "Cluster deleted successfully"}
	c.ServeJSON()
}
