package gcp

import (
	"antelope/models"
	"antelope/models/cores"
	"antelope/models/gcp"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/types"
	"antelope/models/utils"
	"encoding/json"

	"github.com/astaxie/beego"
	"strings"
)

// Operations about Gcp cluster [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/cluster/{cloud}/]
type GcpClusterController struct {
	beego.Controller
}

// @Title Get
// @Description get cluster
// @Param	infraId	path	string	true	"Id of the Infrastructure"
// @Param	X-Auth-Token	header	string	true "token"
// @Success 200 {object} gcp.Cluster_Def
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /:infraId/ [get]
func (c *GcpClusterController) Get() {

	ctx := new(utils.Context)

	infraId := c.GetString(":infraId")
	if infraId == "" {
		ctx.SendLogs("GcpClusterController: infraId is empty", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: Get cluster with Infrastructure id "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbac_athentication.Authenticate(models.GCP, "cluster", infraId, "View", token, utils.Context{})
	if err != nil {
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

	ctx.SendLogs("GcpClusterController: Get cluster with Infrastructure id: "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := gcp.GetCluster(infraId, userInfo.CompanyId, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "No cluster exists for this name"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs(" GCP cluster "+cluster.Name+" of Infrastructure Id: "+cluster.InfraId+" fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Get All
// @Description get all the clusters
// @Param	X-Auth-Token	header	string	true "token"
// @Success 200 {object} []gcp.Cluster_Def
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /all [get]
func (c *GcpClusterController) GetAll() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: GetAll clusters.", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	ctx.SendLogs("GcpClusterController: Getting all clusters ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	statusCode, err, data := rbac_athentication.GetAllAuthenticate("cluster", userInfo.CompanyId, token, models.GCP, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	clusters, err := gcp.GetAllCluster(data, *ctx)
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

	ctx.SendLogs("All GCP cluster fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = clusters
	c.ServeJSON()
}

// @Title Create
// @Description create a new cluster
// @Param	X-Auth-Token	header	string	true "token"
// @Param	body	body 	gcp.Cluster_Def		true	"body for cluster content"
// @Success 201 {"msg": "Cluster created successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Conflict"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [post]
func (c *GcpClusterController) Post() {

	var cluster gcp.Cluster_Def

	ctx := new(utils.Context)

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

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, cluster.InfraId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbac_athentication.Authenticate(models.GCP, "cluster", cluster.InfraId, "Create", token, utils.Context{})
	if err != nil {
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

	ctx.SendLogs("GcpClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	beego.Info("GcpClusterController: JSON Payload: ", cluster)

	network, err := gcp.GetNetwork(token, cluster.InfraId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	for _, node := range cluster.NodePools {
		node.EnablePublicIP = !network.IsPrivate
	}
	cluster.CompanyId = userInfo.CompanyId
	err = gcp.ValidateData(cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "Invalid data: " + err.Error()}
		c.ServeJSON()
		return
	}
	err = gcp.CreateCluster(cluster, *ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "Cluster against same Infrastructure id already exists"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" GCP cluster "+cluster.Name+" of Infrastructure Id: "+cluster.InfraId+" created ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = map[string]string{"msg": "Cluster added successfully"}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing cluster
// @Param	X-Auth-Token	header	string	true "token"
// @Param	body	body 	gcp.Cluster_Def	true	"body for cluster content"
// @Success 200 {"msg": "Cluster updated successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Cluster is in Creating/Terminating/TerminationFailed state"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [put]
func (c *GcpClusterController) Patch() {

	ctx := new(utils.Context)
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

	var cluster gcp.Cluster_Def
	err = json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "Error while unmarshalling " + err.Error()}
		c.ServeJSON()
		return
	}

	savedCluster, err := gcp.GetCluster(cluster.InfraId, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if savedCluster.Status == (models.Deploying) {
		ctx.SendLogs("GcplusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if savedCluster.Status == (models.Terminating) {
		ctx.SendLogs("GcpClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if savedCluster.Status == (models.ClusterTerminationFailed) {
		ctx.SendLogs("GcpClusterController: Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": " Cluster creation is in termination failed state"}
		c.ServeJSON()
		return
	}
	if savedCluster.Status == (models.ClusterCreated) {
		c.Data["json"] = map[string]string{"msg": "No changes are applicable"}
		c.ServeJSON()
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, cluster.InfraId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: update cluster cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbac_athentication.Authenticate(models.GCP, "cluster", cluster.InfraId, "Update", token, utils.Context{})
	if err != nil {
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

	/*	if cluster.Status == (models.Deploying) {
			ctx.SendLogs("GCPClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
			c.ServeJSON()
			return
		} else if cluster.Status == (models.Terminating) {
			ctx.SendLogs("GCPClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
			c.ServeJSON()
			return
		}*/

	ctx.SendLogs("GcpClusterController: Patch cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	beego.Info("GcpClusterController: JSON Payload: ", cluster)
	network, err := gcp.GetNetwork(token, cluster.InfraId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	for _, node := range cluster.NodePools {
		node.EnablePublicIP = !network.IsPrivate
	}
	err = gcp.UpdateCluster(cluster, true, *ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "not found") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "No cluster exists with this name"}
			c.ServeJSON()
			return
		}
		if strings.Contains(err.Error(), "Cluster is in deploying state") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "Cluster is in deploying state"}
			c.ServeJSON()
			return
		}
		if strings.Contains(err.Error(), "No changes are applicable") {
			c.Data["json"] = map[string]string{"msg": string(models.SuccessfullyUpdated)}
			c.ServeJSON()
			return
		}
		if strings.Contains(err.Error(), "Cluster is in terminating state") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
			c.ServeJSON()
			return
		}
		if strings.Contains(err.Error(), "Cluster is in termination failed state") {
			c.Ctx.Output.SetStatus(int(models.StateConflict))
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" GCP cluster "+cluster.Name+" of Infrastructure Id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "Cluster updated successfully"}
	c.ServeJSON()
}

// @Title Delete
// @Description delete a cluster
// @Param	infraId	path	string	true	"Infrastructure id of the cluster"
// @Param	forceDelete path  boolean	true "deleting cluster forcefully"
// @Param	X-Auth-Token	header	string	true "token"
// @Success 204 {"msg": "Cluster deleted successfully"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /:infraId/:forceDelete  [delete]
func (c *GcpClusterController) Delete() {
	ctx := new(utils.Context)

	id := c.GetString(":infraId")
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
	ctx.SendLogs("GcpClusterController: Delete cluster with id "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.GCP, "cluster", id, "Delete", token, utils.Context{})
	if err != nil {
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

	ctx.SendLogs("GcpClusterController: Delete cluster with Infrastructure id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := gcp.GetCluster(id, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if cluster.Status == models.ClusterCreated && !forceDelete {
		ctx.SendLogs("GcpClusterController: Cluster is in running state ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": err.Error() + "Cluster is in created state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Deploying && !forceDelete {
		ctx.SendLogs("GcpClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Terminating && !forceDelete {
		ctx.SendLogs("GcpClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.ClusterTerminationFailed && !forceDelete {
		ctx.SendLogs("GcpClusterController: Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in termination failed state"}
		c.ServeJSON()
		return
	}

	err = gcp.DeleteCluster(id, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs(" GCP cluster "+cluster.Name+" of Infrastructure Id: "+cluster.InfraId+" deleted ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Ctx.Output.SetStatus(204)
	c.Data["json"] = map[string]string{"msg": "cluster deleted successfully"}
	c.ServeJSON()
}

// @Title Start
// @Description starts a  cluster
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	X-Auth-Token	header	string	true "token"
// @Param	infraId	path	string	true	"Id of the Infrastructure"
// @Success 201 {"msg": "Cluster created successfully"}
// @Success 202 {"msg": "Cluster creation started successfully"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Conflict"}
// @Failure 500 {"error": "Runtime Error"}
// @router /start/:infraId [post]
func (c *GcpClusterController) StartCluster() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: StartCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	infraId := c.GetString(":infraId")
	if infraId == "" {
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.GCP, "cluster", infraId, "Start", token, utils.Context{})
	if err != nil {
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

	region, zone, err := gcp.GetRegion(token, infraId, *ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	isValid, credentials := gcp.IsValidGcpCredentials(profileId, region, token, zone, *ctx)
	if !isValid {
		ctx.SendLogs("GcpClusterController : Unable to get profile", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	var cluster gcp.Cluster_Def

	ctx.SendLogs("GcpClusterController: Getting Cluster of Infrastructure. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = gcp.GetCluster(infraId, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if cluster.Status == models.ClusterCreated {
		ctx.SendLogs("GcpClusterController : Cluster is in created state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in created state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Deploying {
		ctx.SendLogs("GcpClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Terminating {
		ctx.SendLogs("GcpClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.ClusterTerminationFailed {
		ctx.SendLogs("GcpClusterController: Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in termination failed state"}
		c.ServeJSON()
		return
	}

	cluster.Status = models.Deploying
	err = gcp.UpdateCluster(cluster, false, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("GcpClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go gcp.DeployCluster(cluster, credentials, userInfo.CompanyId, token, *ctx)

	ctx.SendLogs(" GCP cluster "+cluster.Name+" of Infrastructure Id: "+cluster.InfraId+" deployed ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Ctx.Output.SetStatus(202)
	c.Data["json"] = map[string]string{"msg": "Cluster creation initiated"}
	c.ServeJSON()
}

// @Title Status
// @Description returns status of nodes
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	X-Auth-Token	header	string	true "token"
// @Param	infraId	path	string	true	"Id of the Infrastructure"
// @Success 200 {object} gcp.Cluster_Def
// @Failure 404 {"error": "Not Found"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 500 {"error": "Runtime Error"}
// @router /status/:infraId/ [get]
func (c *GcpClusterController) GetStatus() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: FetchStatus.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	infraId := c.GetString(":infraId")
	if infraId == "" {
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: FetchStatus.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.GCP, "cluster", infraId, "View", token, utils.Context{})
	if err != nil {
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
	region, zone, err := gcp.GetRegion(token, infraId, *ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	isValid, credentials := gcp.IsValidGcpCredentials(profileId, region, token, zone, *ctx)
	if !isValid {
		ctx.SendLogs("GcpClusterController : Gcp credentials not valid ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("GcpClusterController: Fetch Cluster Status of Infrastructure. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err1 := gcp.FetchStatus(credentials, token, infraId, userInfo.CompanyId, *ctx)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GcpClusterController :"+err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(err1.StatusCode)
		c.Data["json"] = err1
	}

	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Terminate
// @Description terminates a  cluster
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	infraId	path	string	true	"Id of the Infrastructure"
// @Param	X-Auth-Token	header	string	true "token"
// @Success 202 {"msg": "Cluster termination initiated"}
// @Success 204 {"msg": "Cluster terminated successfully"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Conflict"}
// @Failure 500 {"error": "Runtime Error"}
// @router /terminate/:infraId/ [post]
func (c *GcpClusterController) TerminateCluster() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	infraId := c.GetString(":infraId")
	if infraId == "" {
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbac_athentication.Authenticate(models.GCP, "cluster", infraId, "Terminate", token, utils.Context{})
	if err != nil {
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

	region, zone, err := gcp.GetRegion(token, infraId, *ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	isValid, credentials := gcp.IsValidGcpCredentials(profileId, region, token, zone, *ctx)
	if !isValid {
		ctx.SendLogs("GcpClusterController: athorization params missing or invalid", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	var cluster gcp.Cluster_Def

	ctx.SendLogs("GcpClusterController: Getting Cluster of Infrastructure. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	cluster.InfraId = infraId
	cluster, err = gcp.GetCluster(infraId, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if cluster.Status == models.Deploying {
		ctx.SendLogs("GcpClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Terminating {
		ctx.SendLogs("GcpClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.ClusterTerminated {
		ctx.SendLogs("GcpClusterController: Cluster is in terminated state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminated state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.ClusterCreationFailed {
		ctx.SendLogs("GcpClusterController: Cluster is in cluster creation failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in cluster creation failed state"}
		c.ServeJSON()
		return
	} else if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		ctx.SendLogs("GcpClusterController: Cluster is not in created status", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is not in created status"}
		c.ServeJSON()
		return
	}

	cluster.Status = models.Terminating
	err = gcp.UpdateCluster(cluster, false, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	go gcp.TerminateCluster(cluster, credentials, token, userInfo.CompanyId, *ctx)

	c.Ctx.Output.SetStatus(202)
	ctx.SendLogs(" GCP cluster "+cluster.Name+" of Infrastructure Id: "+cluster.InfraId+" terminated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "Cluster termination is in progress"}
	c.ServeJSON()
}

// @Title SSHKeyPair
// @Description returns ssh key pairs
// @Param	X-Auth-Token	header	string	true "token"
// @Success 200 {object} []string
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /sshkeys [get]
func (c *GcpClusterController) GetSSHKeys() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: FetchExistingSSHKeys.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: FetchExistingSSHKeys.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//
	keys, err := gcp.GetAllSSHKeyPair(token, *ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = keys
	c.ServeJSON()
}

// @Title ListServiceAccounts
// @Description returns list of service account emails
// @Param	X-Auth-Token	header	string	true "token"
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Success 200 {object} []string
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error":  "Runtime Error"}
// @router /serviceaccounts [get]
func (c *GcpClusterController) GetServiceAccounts() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: FetchExistingServiceAccounts.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: Getting service accounts ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	isValid, credentials := gcp.IsValidGcpCredentials(profileId, "", token, "", *ctx)
	if !isValid {
		ctx.SendLogs("GcpClusterController: authorization params missing or invalid ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	serviceAccounts, err1 := gcp.GetAllServiceAccounts(credentials, *ctx)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("GcpClusterController :"+err1.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(err1.StatusCode)
		c.Data["json"] = err
		c.ServeJSON()
		return
	}

	c.Data["json"] = serviceAccounts
	c.ServeJSON()
}

// @Title CreateSSHKey
// @Description Generates new SSH key
// @Param	infraId	path	string	true	"Id of the Infrastructure"
// @Param	keyname	 	path	string	true	"SSHKey"
// @Param	username	path	string	true	"UserName"
// @Param	X-Auth-Token		header	string	true "Token"
// @Param	teams		header	string	true "Teams"
// @Success 200 		{object} key_utils.AZUREKey
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 		{"error": "Unauthorized"}
// @Failure 404 		{"error": "Not Found"}
// @Failure 500 		{"error": "Runtime Error"}
// @router /sshkey/:keyname/:username/:infraId [post]
func (c *GcpClusterController) PostSSHKey() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: CreateSSHKey.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	infraId := c.GetString(":infraId")
	if infraId == "" {
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

	teams := c.Ctx.Input.Header("teams")

	//==========================RBAC Authentication==============================//

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: PostSSHKey.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	keyName := c.GetString(":keyname")
	if keyName == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "key name is empty"}
		c.ServeJSON()
		return
	}

	userName := c.GetString(":username")
	if userName == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "username is empty"}
		c.ServeJSON()
		return
	}

	privateKey, err := gcp.GetSSHkey(keyName, userName, token, teams, *ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" GCP cluster key "+keyName+" created ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = privateKey
	c.ServeJSON()
}

// @Title DeleteSSHKey
// @Description Delete SSH key
// @Param	keyname	 	path	string	true	"keyname"
// @Param	X-Auth-Token		header	string	true "token"
// @Success 204 		{"msg": "Key deleted successfully"}
// @Failure 401 		{"error": "Unauthorized"}
// @Failure 404 		{"error": "Not Found"}
// @Failure 409 		{"error": "Conflict"}
// @Failure 500 		{"error": "Runtime Error"}
// @router /sshkey/:keyname [delete]
func (c *GcpClusterController) DeleteSSHKey() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: CreateSSHKey.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
		c.ServeJSON()
		return
	}
	//==========================RBAC Authentication==============================//
	//===================================================================//

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GCPClusterController: DeleteSSHKey.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	keyName := c.GetString(":keyname")
	if keyName == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "key name is empty"}
		c.ServeJSON()
		return
	}
	alreadyUsed := gcp.CheckKeyUsage(keyName, userInfo.CompanyId, *ctx)
	if alreadyUsed {
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Key is used in other Infrastructures and can't be deleted"}
		c.ServeJSON()
		return
	}
	err = gcp.DeleteSSHkey(keyName, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" GCP cluster key "+keyName+" deleted ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Ctx.Output.SetStatus(204)
	c.Data["json"] = map[string]string{"msg": "Key deleted successfully"}
	c.ServeJSON()
}

// @Title GetAllMachines
// @Description return machines against a region and zone
// @Param	X-Profile-Id	header	string	true "profileId"
// @Param	X-Auth-Token	header	string	true "token"
// @Param	zone	path	string	true	"zone of GCP"
// @Success 200 []string
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /getallmachines/:zone [get]
func (c *GcpClusterController) GetAllMachines() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: GellAllMachines.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
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
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: GetAllMachines.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	zone := c.GetString(":zone")
	if zone == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "zone is empty"}
		c.ServeJSON()
		return
	}
	region := strings.Split(zone, "-")
	isValid, credentials := gcp.IsValidGcpCredentials(profileId, region[0]+"-"+region[1], token, region[2], *ctx)
	if !isValid {
		ctx.SendLogs("GcpClusterController : Gcp credentials not valid ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("GcpClusterController: Get All Machines. ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	machines, err1 := gcp.GetAllMachines(credentials, *ctx)
	if err1 != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(err1.StatusCode)
		c.Data["json"] = err1
		c.ServeJSON()
		return
	}

	c.Data["json"] = machines.MachineName
	c.ServeJSON()
}

// @Title Get All Regions
// @Description return all regions
// @Success 200 {object} []string
// @Failure 404 {"error": "Not Found"}
// @Failure 401 {"error": "Unauthorized"}
// @router /getallregions [get]
func (c *GcpClusterController) GetAllRegions() {
	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: GetAllRegions.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	regions, err := gcp.GetRegions()
	if err != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(err.StatusCode)
		c.Data["json"] = err
		c.ServeJSON()
		return
	}
	ctx.SendLogs("GcpClusterController: Region fetched ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	c.Data["json"] = regions
	c.ServeJSON()
}

// @Title Validate Profile
// @Description check if profile is valid
// @Param	X-Auth-Token	header	string	true "token"
// @Param	body	body 	gcp.GcpCredentials	true	"body for cluster content"
// @Success 200 {"msg": "Profile is valid"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /validateProfile [post]
func (c *GcpClusterController) ValidateProfile() {

	ctx := new(utils.Context)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
		c.ServeJSON()
		return
	}

	var profile gcp.GcpCredentials

	prof := c.Ctx.Input.RequestBody

	json.Unmarshal(c.Ctx.Input.RequestBody, &profile)

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	ctx.SendLogs("Check Profile Validity", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	var regions []models.Region
	if err := json.Unmarshal(cores.GCPRegions, &regions); err != nil {
		beego.Error("Unmarshalling of machine instances failed ", err.Error())
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	for _, region := range regions {
		err1 := gcp.ValidateProfile(prof, region.Location, "b", *ctx)
		if err1 != (types.CustomCPError{}) {
			ctx.SendLogs("GcpClusterController: Profile not valid", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			c.Ctx.Output.SetStatus(err1.StatusCode)
			c.Data["json"] = err
			c.ServeJSON()
			return
		}
		if err == nil {
			break
		}
	}

	ctx.SendLogs("Profile Validated", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "profile is valid"}
	c.ServeJSON()
}

// @Title GetZonesAgainstRegion
// @Description return zones against a region
// @Param	X-Profile-Id	header	string	true "X-Profile-Id"
// @Param	X-Auth-Token	header	string	true "token"
// @Param	region	path	string	true	"region of GCP"
// @Success 200 {object} []string
// @Failure 404 {"error": "Not Found"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 500 {"error": "Runtime Error"}
// @router /getzones/:region [get]
func (c *GcpClusterController) GetZones() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: GellZones.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: GetAllZones.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	region := c.GetString(":region")
	if region == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "region is empty"}
		c.ServeJSON()
		return
	}

	isValid, credentials := gcp.IsValidGcpCredentials(profileId, region, token, "", *ctx)
	if !isValid {
		ctx.SendLogs("GcpClusterController : Gcp credentials not valid ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "authorization params missing or invalid"}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("GcpClusterController: Get Zones. ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	zones, err1 := gcp.GetZones(credentials, *ctx)
	if err1 != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(err1.StatusCode)
		c.Data["json"] = err1
		c.ServeJSON()
		return
	}
	c.Data["json"] = zones
	c.ServeJSON()
}
