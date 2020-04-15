package gke

import (
	"antelope/models"
	"antelope/models/gcp"
	"antelope/models/gke"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/astaxie/beego"
	"strings"
)

// Operations about GKE cluster [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/cluster/{cloud}/]
type GKEClusterController struct {
	beego.Controller
}

// @Title Get
// @Description get cluster version and image type details
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	token	header	string	true "token"
// @Param	region	path	string	true	"region"
// @Success 200 {object} gke.ServerConfig
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router /config/:region [get]
func (c *GKEClusterController) GetServerConfig() {
	ctx := new(utils.Context)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("GKEClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	zone := c.GetString(":region")
	if zone == "" {
		ctx.SendLogs("GKEClusterController: Region field is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "region is empty"}
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

	isValid, credentials := gcp.IsValidGcpCredentials(profileId, "", token, zone, *ctx)
	if !isValid {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	config, err := gke.GetServerConfig(credentials, *ctx)
	if err.Description != "" {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = err
		c.ServeJSON()
		return
	}

	c.Data["json"] = config
	c.ServeJSON()
}

// @Title Get
// @Description get cluster against the projectId
// @Param	projectId	path	string	true	"Id of the project"
// @Param	token	header	string	true "token"
// @Success 200 {object} gke.GKECluster
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router /:projectId/ [get]
func (c *GKEClusterController) Get() {
	ctx := new(utils.Context)

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("GKEClusterController: projectId is empty", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.Data.Company=userInfo.CompanyId
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId, ctx.Data.Company, userInfo.UserId)
	ctx.SendLogs("GKEClusterController: Get cluster with project id "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", projectId, "View", token, utils.Context{})
	if err != nil {
		c.Ctx.Output.SetStatus(500)
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

	ctx.SendLogs("GKEClusterController: Get cluster with project id: "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := gke.GetGKECluster( *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "no cluster exists for this name"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs(" GKE cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Get All
// @Description get all of the company clusters
// @Param	token	header	string	true "token"
// @Success 200 {object} []gke.GKECluster
// @Failure 400 {"error": "Bad Request"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router /all [get]
func (c *GKEClusterController) GetAll() {
	ctx := new(utils.Context)
	ctx.SendLogs("GKEClusterController: GetAll clusters.", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", ctx.Data.Company, userInfo.UserId)

	ctx.SendLogs("GKEClusterController: Getting all clusters ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err, data := rbacAuthentication.GetAllAuthenticate("cluster", ctx.Data.Company, token, models.GKE, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	clusters, err := gke.GetAllGKECluster(data, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("All GKE cluster fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = clusters
	c.ServeJSON()
}

// @Title Add
// @Description add a new cluster
// @Param	token	header	string	true "token"
// @Param	body	body 	gke.GKECluster		true	"body for cluster content"
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "cluster against same project id already exists"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router / [post]
func (c *GKEClusterController) Post() {

	var cluster gke.GKECluster

	ctx := new(utils.Context)

	err:= json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "error while unmarshalling " + err.Error()}
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

	_, err = govalidator.ValidateStruct(cluster)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, cluster.ProjectId, ctx.Data.Company, userInfo.UserId)
	ctx.SendLogs("GKEClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.Data.Company=userInfo.CompanyId

	allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", cluster.ProjectId, "Create", token, utils.Context{})
	if err != nil {
		c.Ctx.Output.SetStatus(500)
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

	ctx.SendLogs("DOKSClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs("GKEClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	beego.Info("GKEClusterController: JSON Payload: ", cluster)

	cluster.CompanyId = ctx.Data.Company

	err = gke.AddGKECluster(cluster, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "cluster against same project id already exists"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("GKE cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" created ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster added successfully"}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing cluster
// @Param	token	header	string	true "true"
// @Param	body	body 	gke.GKECluster	true	"body for cluster content"
// @Success 200 {"msg": "cluster updated successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 402 {"error": "Cluster is in running state"}
// @Failure 404 {"error": "Not found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router / [put]
func (c *GKEClusterController) Patch() {
	ctx := new(utils.Context)

	var cluster gke.GKECluster
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "error while unmarshalling " + err.Error()}
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

	userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, cluster.ProjectId, ctx.Data.Company, userInfo.UserId)
	ctx.SendLogs("GKEClusterController: update cluster cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.Data.Company=userInfo.CompanyId

	allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", cluster.ProjectId, "Update", token, utils.Context{})
	if err != nil {
		c.Ctx.Output.SetStatus(500)
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
	ctx.SendLogs("GKEClusterController: Patch cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	beego.Info("GKEClusterController: JSON Payload: ", cluster)

	err = gke.UpdateGKECluster(cluster, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "no cluster exists with this name"}
			c.ServeJSON()
			return
		}
		if strings.Contains(err.Error(), "Cluster is in running state") {
			c.Ctx.Output.SetStatus(402)
			c.Data["json"] = map[string]string{"error": "Cluster is in running state"}
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

	ctx.SendLogs("GKE cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster updated successfully"}
	c.ServeJSON()
}

// @Title Delete
// @Description delete a existing cluster
// @Param	projectId	path	string	true	"project id of the cluster"
// @Param	forceDelete path  boolean	true ""
// @Param	token	header	string	true "token"
// @Success 200 {"msg": "Cluster deleted successfully"}
// @Success 204 {"msg": "Cluster deleted successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 402 {"error": "Cluster is in running state"}
// @Failure 404 {"error": "Not found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router /:projectId/:forceDelete  [delete]
func (c *GKEClusterController) Delete() {
	ctx := new(utils.Context)

	id := c.GetString(":projectId")
	if id == "" {
		ctx.SendLogs("GKEClusterController: ProjectId field is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, id, ctx.Data.Company, userInfo.UserId)
	ctx.SendLogs("GKEClusterController: Delete cluster with id "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.Data.Company=userInfo.CompanyId

	allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", id, "Delete", token, utils.Context{})
	if err != nil {
		c.Ctx.Output.SetStatus(500)
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

	ctx.SendLogs("GKEClusterController: Delete cluster with project id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := gke.GetGKECluster( *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if strings.ToLower(cluster.CloudplexStatus) == string(models.ClusterCreated) && !forceDelete {
		ctx.SendLogs("GKEClusterController: Cluster is in running state ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(402)
		c.Data["json"] = map[string]string{"error": "Cluster is in running state"}
		c.ServeJSON()
		return
	}

	if cluster.CloudplexStatus == string(models.Deploying) && !forceDelete {
		ctx.SendLogs("GKEClusterController: Cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in deploying state"}
		c.ServeJSON()
		return
	}

	if cluster.CloudplexStatus == string(models.Terminating) && !forceDelete {
		ctx.SendLogs("GKEClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in terminating state"}
		c.ServeJSON()
		return
	}

	err = gke.DeleteGKECluster( *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("GKE cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" deleted ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster deleted successfully"}
	c.ServeJSON()
}

// @Title Start
// @Description starts a cluster
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	token	header	string	true "token"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 402 {"error": "Cluster is in running state"}
// @Failure 404 {"error": "Not found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router /start/:projectId [post]
func (c *GKEClusterController) StartCluster() {
	ctx := new(utils.Context)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("GKEClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("GKEClusterController: ProjectId field is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId, ctx.Data.Company, userInfo.UserId)
	ctx.Data.Company=userInfo.CompanyId
	allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", projectId, "Start", token, utils.Context{})
	if err != nil {
		c.Ctx.Output.SetStatus(500)
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
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	isValid, credentials := gcp.IsValidGcpCredentials(profileId, region, token, zone, *ctx)
	if !isValid {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("GKEClusterController: Getting Cluster of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := gke.GetGKECluster(*ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if cluster.CloudplexStatus == "Cluster Created" {
		ctx.SendLogs("GKEClusterController : Cluster is already running", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(402)
		c.Data["json"] = map[string]string{"error": "cluster is already in running state"}
		c.ServeJSON()
		return
	}

	if cluster.CloudplexStatus == string(models.Deploying) {
		ctx.SendLogs("GKEClusterController: Cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in deploying state"}
		c.ServeJSON()
		return
	}

	if cluster.CloudplexStatus == string(models.Terminating) {
		ctx.SendLogs("GKEClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in terminating state"}
		c.ServeJSON()
		return
	}

	cluster.CloudplexStatus = string(models.Deploying)
	err = gke.UpdateGKECluster(cluster, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("GKEClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go gke.DeployGKECluster(cluster, credentials, token, *ctx)

	ctx.SendLogs(" GKE cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" deployed ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster creation in progress"}
	c.ServeJSON()
}

// @Title Status
// @Description returns latest status of the running cluster
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	token	header	string	true "token"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {object} gke.GKECluster
// @Failure 206 {object} gke.GKECluster
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router /status/:projectId/ [get]
func (c *GKEClusterController) GetStatus() {
	ctx := new(utils.Context)


	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("GKEClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("GKEClusterController: ProjectId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId, ctx.Data.Company, userInfo.UserId)
	ctx.SendLogs("GKEClusterController: FetchStatus.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", projectId, "View", token, utils.Context{})
	if err != nil {
		c.Ctx.Output.SetStatus(500)
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
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	isValid, credentials := gcp.IsValidGcpCredentials(profileId, region, token, zone, *ctx)
	if !isValid {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("GKEClusterController: Fetch Cluster Status of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.Data.Company=userInfo.CompanyId
	cluster, err1 := gke.FetchStatus(credentials, token,  *ctx)
	if err1.Description != "" {
		c.Ctx.Output.SetStatus(206)
	}

	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Terminate
// @Description terminates a running cluster
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	projectId	path	string	true	"Id of the project"
// @Param	token	header	string	true "token"
// @Success 200 {"msg": "Cluster is terminating"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router /terminate/:projectId/ [post]
func (c *GKEClusterController) TerminateCluster() {

	ctx := new(utils.Context)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("GKEClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("GKEClusterController: ProjectId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.Data.Company=userInfo.CompanyId
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId, ctx.Data.Company, userInfo.UserId)
	ctx.SendLogs("GKEClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", projectId, "Terminate", token, utils.Context{})
	if err != nil {

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
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	isValid, credentials := gcp.IsValidGcpCredentials(profileId, region, token, zone, *ctx)
	if !isValid {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("GKEClusterController: Getting Cluster of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := gke.GetGKECluster( *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if cluster.CloudplexStatus == string(models.Deploying) {
		ctx.SendLogs("GKEClusterController: cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in deploying state"}
		c.ServeJSON()
		return
	}

	if cluster.CloudplexStatus == string(models.Terminating) {
		ctx.SendLogs("GKEClusterController: cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in terminating state"}
		c.ServeJSON()
		return
	}

	go gke.TerminateCluster(credentials,  *ctx)

	err = gke.UpdateGKECluster(cluster, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" GKE cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" terminated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster termination is in progress"}
	c.ServeJSON()
}

// @Title Start
// @Description Apply cloudplex Agent file to a gke cluster
// @Param	clusterName	header	string	clusterName "Name of the cluster"
// @Param	token	header	string	true "token"
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {"msg": "Agent Applied successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router /applyagent/:projectId [post]
func (c *GKEClusterController) ApplyAgent() {

	ctx := new(utils.Context)
	ctx.SendLogs("GKEClusterController: Apply agent.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("GKEClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("GKEClusterController: ProjectId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	clusterName := c.Ctx.Input.Header("clusterName")
	if clusterName == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "clusterName is empty"}
		c.ServeJSON()
		return
	}
	userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId, ctx.Data.Company, userInfo.UserId)
	ctx.SendLogs("GKEClusterController: Apply Agent.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", projectId, "Start", token, utils.Context{})
	if err != nil {
		c.Ctx.Output.SetStatus(500)
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
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	isValid, credentials := gcp.IsValidGcpCredentials(profileId, region, token, zone, *ctx)
	if !isValid {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("GKEClusterController: Applying agent on cluster . "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go gke.ApplyAgent(credentials, token, *ctx, clusterName)

	c.Data["json"] = map[string]string{"msg": "agent deployment in progress"}
	c.ServeJSON()
}
