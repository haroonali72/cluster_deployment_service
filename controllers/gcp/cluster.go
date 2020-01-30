package gcp

import (
	"antelope/models"
	"antelope/models/cores"
	"antelope/models/gcp"
	rbac_athentication "antelope/models/rbac_authentication"
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
// @Param	projectId	path	string	true	"Id of the project"
// @Param	token	header	string	token ""
// @Success 200 {object} gcp.Cluster_Def
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /:projectId/ [get]
func (c *GcpClusterController) Get() {

	ctx := new(utils.Context)

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("GcpClusterController: projectId is empty", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: Get cluster with project id "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	allowed, err := rbac_athentication.Authenticate(models.GCP, "cluster", projectId, "View", token, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
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

	ctx.SendLogs("GcpClusterController: Get cluster with project id: "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := gcp.GetCluster(projectId, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("GcpGetClusterController: error getting gcp cluster "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "no cluster exists for this name"}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" GCP cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Get All
// @Description get all the clusters
// @Param	token	header	string	token ""
// @Success 200 {object} []gcp.Cluster_Def
// @Failure 400 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /all [get]
func (c *GcpClusterController) GetAll() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: GetAll clusters.", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	ctx.SendLogs("GcpClusterController: Getting all clusters ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	err, data := rbac_athentication.GetAllAuthenticate("cluster", userInfo.CompanyId, token, models.GCP, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	clusters, err := gcp.GetAllCluster(data, *ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
// @Param	subscription_id	header	string	subscriptionId ""
// @Param	token	header	string	token ""
// @Param	body	body 	gcp.Cluster_Def		true	"body for cluster content"
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 409 {"error": "cluster against same project id already exists"}
// @Failure 410 {"error": "Core limit exceeded"}
// @Failure 500 {"error": "error msg"}
// @router / [post]
func (c *GcpClusterController) Post() {

	var cluster gcp.Cluster_Def

	ctx := new(utils.Context)

	json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	subscriptionId := c.Ctx.Input.Header("subscription_id")
	if subscriptionId == "" {
		ctx.SendLogs("GcpClusterController: subscriptionId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400) //no need
		c.Data["json"] = map[string]string{"error": "subscriptionId is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, cluster.ProjectId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	allowed, err := rbac_athentication.Authenticate(models.GCP, "cluster", cluster.ProjectId, "Create", token, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
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

	err = gcp.GetNetwork(token, cluster.ProjectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	cluster.CompanyId = userInfo.CompanyId

	err = gcp.CreateCluster(subscriptionId, cluster, *ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "cluster against same project id already exists"}
			c.ServeJSON()
			return
		} else if strings.Contains(err.Error(), "Exceeds the cores limit") {
			c.Ctx.Output.SetStatus(410)
			c.Data["json"] = map[string]string{"error": "core limit exceeded"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" GCP cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" created ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster added successfully"}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing cluster
// @Param	token	header	string	token ""
// @Param	subscription_id	header	string	subscriptionId ""
// @Param	body	body 	gcp.Cluster_Def	true	"body for cluster content"
// @Success 200 {"msg": "cluster updated successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 402 {"error": "error msg"}
// @Failure 404 {"error": "no cluster exists with this name"}
// @Failure 500 {"error": "error msg"}
// @router / [put]
func (c *GcpClusterController) Patch() {

	ctx := new(utils.Context)

	var cluster gcp.Cluster_Def
	json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	subscriptionId := c.Ctx.Input.Header("subscription_id")
	if subscriptionId == "" {
		ctx.SendLogs("GcpClusterController: subscriptionId field is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "subscriptionId is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, cluster.ProjectId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: update cluster cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	allowed, err := rbac_athentication.Authenticate(models.GCP, "cluster", cluster.ProjectId, "Update", token, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
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
	ctx.SendLogs("GcpClusterController: Patch cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	beego.Info("GcpClusterController: JSON Payload: ", cluster)

	err = gcp.UpdateCluster(subscriptionId, cluster, true, *ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		if strings.Contains(err.Error(), "does not exist") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "no cluster exists with this name"}
			c.ServeJSON()
			return
		}
		if strings.Contains(err.Error(), "Cluster is in runnning state") {
			c.Ctx.Output.SetStatus(402)
			c.Data["json"] = map[string]string{"error": "Cluster is in runnning state"}
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
	ctx.SendLogs(" GCP cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster updated successfully"}
	c.ServeJSON()
}

// @Title Delete
// @Description delete a cluster
// @Param	projectId	path	string	true	"project id of the cluster"
// @Param	forceDelete path  boolean	true ""
// @Param	token	header	string	token ""
// @Success 200 {"msg": "cluster deleted successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 500 {"error": "error msg"}
// @router /:projectId/:forceDelete  [delete]
func (c *GcpClusterController) Delete() {
	ctx := new(utils.Context)

	id := c.GetString(":projectId")
	if id == "" {
		ctx.SendLogs("GcpClusterController: ProjectId field is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
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
	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, id, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: Delete cluster with id "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//
	allowed, err := rbac_athentication.Authenticate(models.GCP, "cluster", id, "Delete", token, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
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

	ctx.SendLogs("GcpClusterController: Delete cluster with project id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := gcp.GetCluster(id, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if cluster.Status == "Cluster Created" && !forceDelete {
		ctx.SendLogs("GcpClusterController: Cluster is in running state ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error() + "Cluster is in running state"}
		c.ServeJSON()
		return
	}

	if cluster.Status == string(models.Deploying) && !forceDelete {
		ctx.SendLogs("GcpClusterController: Cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in deploying state"}
		c.ServeJSON()
		return
	}

	if cluster.Status == string(models.Terminating) && !forceDelete {
		ctx.SendLogs("GcpClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in terminating state"}
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
	ctx.SendLogs(" GCP cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" deleted ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster deleted successfully"}
	c.ServeJSON()
}

// @Title Start
// @Description starts a  cluster
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	token	header	string	token ""
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /start/:projectId [post]
func (c *GcpClusterController) StartCluster() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: StartCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("GcpClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("GcpClusterController: ProjectId field is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	allowed, err := rbac_athentication.Authenticate(models.GCP, "cluster", projectId, "Start", token, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
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

	region, zone, err := gcp.GetRegion(token, projectId, *ctx)
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

	ctx.SendLogs("GcpClusterController: Getting Cluster of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = gcp.GetCluster(projectId, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if cluster.Status == "Cluster Created" {
		ctx.SendLogs("GcpClusterController : Cluster is already running", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is already in running state"}
		c.ServeJSON()
		return
	}

	if cluster.Status == string(models.Deploying) {
		ctx.SendLogs("GcpClusterController: Cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in deploying state"}
		c.ServeJSON()
		return
	}

	if cluster.Status == string(models.Terminating) {
		ctx.SendLogs("GcpClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in terminating state"}
		c.ServeJSON()
		return
	}

	cluster.Status = string(models.Deploying)
	err = gcp.UpdateCluster("", cluster, false, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("GcpClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go gcp.DeployCluster(cluster, credentials, userInfo.CompanyId, token, *ctx)

	ctx.SendLogs(" GCP cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" deployed ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster creation in progress"}
	c.ServeJSON()
}

// @Title Status
// @Description returns status of nodes
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	token	header	string	token ""
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {object} gcp.Cluster_Def
// @Failure 206 {object} gcp.Cluster_Def
// @Failure 400 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @Failure 401 {"error": "authorization params missing or invalid"}
// @Failure 500 {"error": "error msg"}
// @router /status/:projectId/ [get]
func (c *GcpClusterController) GetStatus() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: FetchStatus.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("GcpClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("GcpClusterController: ProjectId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: FetchStatus.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//
	allowed, err := rbac_athentication.Authenticate(models.GCP, "cluster", projectId, "View", token, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
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
	region, zone, err := gcp.GetRegion(token, projectId, *ctx)
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

	ctx.SendLogs("GcpClusterController: Fetch Cluster Status of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := gcp.FetchStatus(credentials, token, projectId, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(206)
	}

	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Terminate
// @Description terminates a  cluster
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	projectId	path	string	true	"Id of the project"
// @Param	token	header	string	token ""
// @Success 200 {"msg": "cluster terminated successfully"}
// @Failure 401 {"error": "Authorization format should be 'base64 encoded service_account_json'"}
// @Failure 400 {"error": "error_msg"}
// @Failure 404 {"error": "error_msg"}
// @Failure 500 {"error": "error msg"}
// @router /terminate/:projectId/ [post]
func (c *GcpClusterController) TerminateCluster() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("GcpClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("GcpClusterController: ProjectId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	allowed, err := rbac_athentication.Authenticate(models.GCP, "cluster", projectId, "Terminate", token, utils.Context{})
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
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

	region, zone, err := gcp.GetRegion(token, projectId, *ctx)
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

	ctx.SendLogs("GcpClusterController: Getting Cluster of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = gcp.GetCluster(projectId, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if cluster.Status == string(models.Deploying) {
		ctx.SendLogs("GcpClusterController: cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in deploying state"}
		c.ServeJSON()
		return
	}

	if cluster.Status == string(models.Terminating) {
		ctx.SendLogs("GcpClusterController: cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in terminating state"}
		c.ServeJSON()
		return
	}

	go gcp.TerminateCluster(cluster, credentials, userInfo.CompanyId, *ctx)

	err = gcp.UpdateCluster("", cluster, false, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" GCP cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" terminated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster termination is in progress"}
	c.ServeJSON()
}

// @Title SSHKeyPair
// @Description returns ssh key pairs
// @Param	token	header	string	token ""
// @Success 200 {object} []string
// @Failure 400 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /sshkeys [get]
func (c *GcpClusterController) GetSSHKeys() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: FetchExistingSSHKeys.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
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
// @Param	token	header	string	token ""
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Success 200 {object} []string
// @Failure 400 {"error": "profile id is empty"}
// @Failure 401 {"error": "authorization params missing or invalid"}
// @Failure 500 {"error":  "error msg"}
// @router /serviceaccounts [get]
func (c *GcpClusterController) GetServiceAccounts() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: FetchExistingServiceAccounts.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: Getting service accounts ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("GcpClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
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

	serviceAccounts, err := gcp.GetAllServiceAccounts(credentials, *ctx)
	if err != nil {
		ctx.SendLogs("GcpClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = serviceAccounts
	c.ServeJSON()
}

// @Title CreateSSHKey
// @Description Generates new SSH key
// @Param	projectId	path	string	true	"Id of the project"
// @Param	keyname	 	path	string	true	"SSHKey"
// @Param	username	path	string	true	"UserName"
// @Param	token		header	string	token 	""
// @Param	teams		header	string	teams 	""
// @Success 200 		{object} key_utils.AZUREKey
// @Failure 400 		{"error": "error msg"}
// @Failure 404 		{"error": "error msg"}
// @Failure 500 		{"error": "error msg"}
// @router /sshkey/:keyname/:username/:projectId [post]
func (c *GcpClusterController) PostSSHKey() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: CreateSSHKey.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("GcpClusterController: ProjectId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	teams := c.Ctx.Input.Header("teams")

	//==========================RBAC Authentication==============================//

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)
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

// @Title GetCores
// @Description Get GCP Machine instance cores
// @Success 200 			{object} models.Machine
// @Failure 500 			{"error":  "error msg"}
// @router /machine/info [get]
func (c *GcpClusterController) GetCores() {

	var machine []models.GCPMachine
	if err := json.Unmarshal(cores.GCPCores, &machine); err != nil {
		beego.Error("Unmarshalling of machine instances failed ", err.Error())
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	c.Data["json"] = machine
	c.ServeJSON()
}

// @Title DeleteSSHKey
// @Description Delete SSH key
// @Param	keyname	 	path	string	true	""
// @Param	token		header	string	token 	""
// @Success 200 		{"msg": key deleted successfully}
// @Failure 400 		{"error": "error msg"}
// @Failure 404 		{"error": error msg}
// @router /sshkey/:keyname [delete]
func (c *GcpClusterController) DeleteSSHKey() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: CreateSSHKey.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}
	//==========================RBAC Authentication==============================//
	//===================================================================//

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(404)
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
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "key is used in other projects and can't be deleted"}
		c.ServeJSON()
		return
	}
	err = gcp.DeleteSSHkey(keyName, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" GCP cluster key "+keyName+" deleted ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "key deleted successfully"}
	c.ServeJSON()
}

// @Title GetAllMachines
// @Description return machines against a region and zone
// @Param	profileid	header	string	true	"profile of GCP"
// @Param	token	header	string	token  true""
// @Param	region	path	string	true	"region of GCP"
// @Param	zone	path	string	true	"zone of GCP"
// @Success 200 {object} gcp.Machines
// @Failure 400 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @Failure 401 {"error": "authorization params missing or invalid"}
// @router /getallmachines/:region/:zone [get]
func (c *GcpClusterController) GetAllMachines() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: GellAllMachines.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("profileid")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profileid is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("GcpClusterController: GetAllMachines.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	region := c.GetString(":region")
	if region == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "region is empty"}
		c.ServeJSON()
		return
	}

	zone := c.GetString(":zone")
	if zone == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "zone is empty"}
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

	ctx.SendLogs("GcpClusterController: Get All Machines. ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	machines, err := gcp.GetAllMachines(credentials, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = machines
	c.ServeJSON()
}

// @Title GetZonesAgainstRegion
// @Description return zones against a region
// @Param	profileid	header	string	true	"profile of GCP"
// @Param	token	header	string	token  true""
// @Param	region	path	string	true	"region of GCP"
// @Success 200 {object} []string
// @Failure 400 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @Failure 401 {"error": "authorization params missing or invalid"}
// @router /getzones/:region [get]
func (c *GcpClusterController) GetZones() {

	ctx := new(utils.Context)
	ctx.SendLogs("GcpClusterController: GellZones.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("profileid")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profileid is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
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

	zones, err := gcp.GetZones(credentials, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	c.Data["json"] = zones
	c.ServeJSON()
}
