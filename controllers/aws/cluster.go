package aws

import (
	"antelope/models"
	"antelope/models/aws"
	"antelope/models/cores"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/go-playground/validator/v10"
	"strings"
	"time"
)

// Operations about AWS cluster [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/cluster/{cloud}/]
type AWSClusterController struct {
	beego.Controller
}

// @Title Get
// @Description get cluster
// @Param	projectId	path	string	true	"Id of the project"
// @Param	X-Auth-Token	header	string	true "token"
// @Success 200 {object} aws.Cluster_Def
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /:projectId/ [get]
func (c *AWSClusterController) Get() {

	projectId := c.GetString(":projectId")
	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
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

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbac_athentication.Authenticate(models.AWS, "cluster", projectId, "View", token, *ctx)
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

	//====================================================================================//

	ctx.SendLogs("AWSClusterController: Get cluster with project id: "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	cluster, err := aws.GetCluster(projectId, userInfo.CompanyId, *ctx)
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
	ctx.SendLogs(" AWS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Get All
// @Description get all the clusters
// @Param	X-Auth-Token	header	string	true "token"
// @Success 200 {object} []aws.AwsCluster
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /all [get]
func (c *AWSClusterController) GetAll() {
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

	statusCode, err, data := rbac_athentication.GetAllAuthenticate("cluster", userInfo.CompanyId, token, models.AWS, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	//====================================================================================//
	ctx.SendLogs("AWSClusterController: GetAll clusters.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	clusters, err := aws.GetAllCluster(*ctx, data)
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
	ctx.SendLogs(" All AWS clusters fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = clusters
	c.ServeJSON()
}

// @Title Add
// @Description add a new cluster
// @Param	body	body 	aws.Cluster_Def		true	"body for cluster content"
// @Param	X-Auth-Token	header	string	true "token"
// @Success 201 {"msg": "cluster created successfully"}
// @Success 400 {"msg": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Conflict"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [post]
func (c *AWSClusterController) Post() {

	var cluster aws.Cluster_Def
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "Error while unmarshalling " + err.Error()}
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

	cluster.CreationDate = time.Now()

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
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, cluster.ProjectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.AWS, "cluster", cluster.ProjectId, "Create", token, *ctx)
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

	//=============================================================================//

	ctx.SendLogs("AWSClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	network, err := aws.GetNetwork(token, cluster.ProjectId, *ctx)
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
	err = aws.CreateCluster(cluster, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "Cluster against this project id  already exists"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" AWS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" created ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = map[string]string{"msg": "Cluster added successfully"}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing cluster
// @Param	X-Auth-Token	header	string	true "token"
// @Param	body	body 	aws.Cluster_Def	true	"body for cluster content"
// @Success 200 {"msg": "Cluster updated successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Conflict"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [put]
func (c *AWSClusterController) Patch() {
	var cluster aws.Cluster_Def

	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "Error while unmarshalling " + err.Error()}
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
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, cluster.ProjectId, userInfo.CompanyId, userInfo.UserId)

	if cluster.Status == (models.Deploying) {
		ctx.SendLogs("DOKSClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Terminating) {
		ctx.SendLogs("DOKSClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	}
	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.AWS, "cluster", cluster.ProjectId, "Update", token, *ctx)
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

	//=============================================================================//

	ctx.SendLogs("AWSClusterController: Patch cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	network, err := aws.GetNetwork(token, cluster.ProjectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	for _, node := range cluster.NodePools {
		node.EnablePublicIP = !network.IsPrivate
	}
	err = aws.UpdateCluster(cluster, true, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "not found") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "No cluster exists with this name"}
			c.ServeJSON()
			return
		}
		if strings.Contains(err.Error(), "Cluster is in created state") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "Cluster is in created state"}
			c.ServeJSON()
			return
		}
		if strings.Contains(err.Error(), "Cluster creation is in termination failed state") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "Cluster creation is in termination failed state"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" AWS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster updated successfully"}
	c.ServeJSON()
}

// @Title Delete
// @Description delete a cluster
// @Param	X-Auth-Token	header	string	true "token"
// @Param	projectId	path 	string	true	"project id of the cluster"
// @Param	forceDelete path    boolean	true    "deleting cluster forcefully"
// @Success 204 {"msg": "Cluster deleted successfully"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /delete/:projectId/:forceDelete [delete]
func (c *AWSClusterController) Delete() {
	id := c.GetString(":projectId")
	if id == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
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
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, id, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.AWS, "cluster", id, "Delete", token, *ctx)
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

	//=============================================================================//

	ctx.SendLogs("AWSClusterController: Delete cluster with project id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	if id == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	cluster, err := aws.GetCluster(id, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if cluster.Status == models.ClusterCreated && !forceDelete {
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": err.Error() + " + Cluster is in creaated state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Deploying && !forceDelete {
		ctx.SendLogs("cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in deploying state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Terminating && !forceDelete {
		ctx.SendLogs("cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.ClusterTerminationFailed && !forceDelete {
		ctx.SendLogs(" Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": " Cluster is in termination failed state"}
		c.ServeJSON()
		return
	}

	err = aws.DeleteCluster(id, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs(" AWS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" deleted ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Ctx.Output.SetStatus(204)
	c.Data["json"] = map[string]string{"msg": "Cluster deleted successfully"}
	c.ServeJSON()
}

// @Title Start
// @Description starts a  cluster
// @Param	X-Auth-Token	header	string	true "token"
// @Param	X-Profile-Id	header	string	true "profileId"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 201 {"msg": "Cluster created successfully"}
// @Success 202 {"msg": "Cluster creation started successfully"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Conflict"}
// @Failure 500 {"error": "Runtime Error"}
// @router /start/:projectId [post]
func (c *AWSClusterController) StartCluster() {

	projectId := c.GetString(":projectId")
	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
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
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.AWS, "cluster", projectId, "Start", token, *ctx)
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

	//=============================================================================//

	ctx.SendLogs("AWSNetworkController: StartCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	var cluster aws.Cluster_Def

	ctx.SendLogs("AWSClusterController: Getting Cluster of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = aws.GetCluster(projectId, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if cluster.Status == models.ClusterCreated {
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in created state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Deploying {
		ctx.SendLogs("cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Terminating {
		ctx.SendLogs("Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.ClusterTerminationFailed {
		ctx.SendLogs(": Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": ": Cluster is in termination failed state"}
		c.ServeJSON()
		return
	}

	region, err := aws.GetRegion(token, projectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	statusCode, awsProfile, err := aws.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		utils.SendLog(userInfo.CompanyId, err.Error(), "error", cluster.ProjectId)
		utils.SendLog(userInfo.CompanyId, "Cluster creation failed: "+cluster.Name, "error", cluster.ProjectId)
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	cluster.Status = models.Deploying
	err = aws.UpdateCluster(cluster, false, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AWSClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go aws.DeployCluster(cluster, awsProfile.Profile, *ctx, userInfo.CompanyId, token)

	ctx.SendLogs(" AWS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" deployed ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Ctx.Output.SetStatus(202)
	c.Data["json"] = map[string]string{"msg": "Cluster creation initiated"}
	c.ServeJSON()
}

// @Title Status
// @Description returns status of nodes
// @Param	X-Auth-Token	header	string	true "token"
// @Param	X-Profile-Id	header	string	true "profileId"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {object} aws.Cluster_Def
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /status/:projectId/ [get]
func (c *AWSClusterController) GetStatus() {

	projectId := c.GetString(":projectId")
	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
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

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.AWS, "cluster", projectId, "View", token, *ctx)
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

	//=============================================================================//
	ctx.SendLogs("AWSClusterController: FetchStatus.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AWSClusterController: Fetch Cluster Status of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	region, err := aws.GetRegion(token, projectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	statusCode, awsProfile, err := aws.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	cluster, cpErr := aws.FetchStatus(awsProfile, projectId, *ctx, userInfo.CompanyId, token)
	if cpErr != (types.CustomCPError{}) && strings.Contains(strings.ToLower(cpErr.Description), "state") || cpErr != (types.CustomCPError{}) && strings.Contains(strings.ToLower(cpErr.Description), "not deployed") {
		c.Ctx.Output.SetStatus(cpErr.StatusCode)
		c.Data["json"] = cpErr.Description
		c.ServeJSON()
		return
	} else if cpErr != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(int(models.CloudStatusCode))
		c.Data["json"] = cpErr
		c.ServeJSON()
	}

	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Terminate
// @Description terminates a  cluster
// @Param	X-Profile-Id header	X-Profile-Id	string	true "profileId"
// @Param	X-Auth-Token	header	string	true "token"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 202 {"msg": "Cluster termination initiated"}
// @Success 204 {"msg": "Cluster terminated successfully"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /terminate/:projectId/ [post]
func (c *AWSClusterController) TerminateCluster() {

	projectId := c.GetString(":projectId")
	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
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

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.AWS, "cluster", projectId, "Terminate", token, *ctx)
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

	//=============================================================================//
	ctx.SendLogs("AWSClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	region, err := aws.GetRegion(token, projectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	statusCode, awsProfile, err := aws.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	var cluster aws.Cluster_Def

	ctx.SendLogs("AWSClusterController: Getting Cluster of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = aws.GetCluster(projectId, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if cluster.Status == models.Deploying {
		ctx.SendLogs("AWSClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Terminating {
		ctx.SendLogs("AWSClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.ClusterTerminated {
		ctx.SendLogs("AWSClusterController: Cluster is in terminated state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "cluster is in terminated state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.ClusterCreationFailed {
		ctx.SendLogs("AWSClusterController: Cluster is in cluster creation failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "cluster is in cluster creation failed state"}
		c.ServeJSON()
		return
	} else if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		ctx.SendLogs("AWSClusterController: Cluster is not in created status", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is not in created status"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AWSClusterController: Terminating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	cluster.Status = models.Terminating
	err = aws.UpdateCluster(cluster, false, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	go aws.TerminateCluster(cluster, awsProfile, *ctx, userInfo.CompanyId, token)

	ctx.SendLogs(" AWS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" terminated", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Ctx.Output.SetStatus(202)
	c.Data["json"] = map[string]string{"msg": "cluster termination initiated"}
	c.ServeJSON()
}

// @Title SSHKeyPair
// @Description returns ssh key pairs
// @Param	X-Auth-Token	header	string	true "token"
// @Param	region  path	string	true	"region"
// @Success 200 {object} []string
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /sshkeys/:region [get]
func (c *AWSClusterController) GetSSHKeys() {

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

	region := c.GetString(":region")
	if region == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "region is empty"}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	//=============================================================================//

	ctx.SendLogs("AWSClusterController: FetchExistingSSHKeys.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	keys, err := aws.GetAllSSHKeyPair(*ctx, token, region)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = keys
	c.ServeJSON()
}

// @Title AwsAmis
// @Description returns aws ami details
// @Param	X-Profile-Id	header	string	true "profileId"
// @Param	X-Auth-Token	header	string	true "token"
// @Param	region	path	string	true	"cloud region"
// @Param	amiId	path	string	true	"Id of the ami"
// @Success 200 {object} []*ec2.BlockDeviceMapping
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /amis/:amiId/:region [get]
func (c *AWSClusterController) GetAMI() {

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
	//err, _ = rbac_athentication.GetAllAuthenticate("cluster",userInfo.CompanyId, token, *ctx)
	//if err != nil {
	//	beego.Error(err.Error())
	//	c.Ctx.Output.SetStatus(400)
	//	c.Data["json"] = map[string]string{"error": err.Error()}
	//	c.ServeJSON()
	//	return
	//}

	//=============================================================================//
	ctx.SendLogs("AWSClusterController: FetchAMIs.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	region := c.GetString(":region")
	if region == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "region is empty"}
		c.ServeJSON()
		return
	}

	statusCode, awsProfile, err := aws.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		ctx.SendLogs("AWSClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	amiId := c.GetString(":amiId")
	if amiId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "ami id is empty"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AWSClusterController: Get Ami from AWS", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	keys, err1 := aws.GetAWSAmi(awsProfile, amiId, *ctx, token)
	if err1 != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = err1
		c.ServeJSON()
		return
	}

	c.Data["json"] = keys
	c.ServeJSON()
}

// @Title EnableScaling
// @Description enables autoscaling
// @Param	X-Profile-Id	header	string	true "profileId"
// @Param	projectId	path	string	true	"Id of the project"
// @Param	X-Auth-Token	header	string	true "token"
// @Success 200 {object} aws.AutoScaling
// @Success 200 {"msg": "cluster autoscaled successfully"}
// @Failure 401 {"error": "Unathorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /enablescaling/:projectId/ [post]
func (c *AWSClusterController) EnableAutoScaling() {

	projectId := c.GetString(":projectId")
	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
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

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.AWS, "cluster", projectId, "Start", token, *ctx)
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

	//=============================================================================//

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	region, err := aws.GetRegion(token, projectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	statusCode, awsProfile, err := aws.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	cluster, err := aws.GetCluster(projectId, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	go aws.EnableScaling(awsProfile, cluster, *ctx, token)

	/*err = aws.EnableScaling(awsProfile, cluster, *ctx, token)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	*/

	c.Data["json"] = map[string]string{"msg": "cluster autoscaled successfully"}
	c.ServeJSON()
}

// @Title CreateSSHKey
// @Description Generates new SSH key
// @Param	projectId		path	string	true		"Id of the project"
// @Param	keyname	 		path	string	true		"SSHKey"
// @Param	X-Profile-Id	header	string	true "profileId"
// @Param	X-Auth-Token			header	string	true "token"
// @Param	teams			header	string	true "teams"
// @Param	region		path	string	true	"cloud region"
// @Success 201 			{object} key_utils.AWSKey
// @Failure 401 			{"error": "Unauthorized"}
// @Failure 404 			{"error": "Not Found"}
// @Failure 409 			{"error": "Conflict"}
// @Failure 500 			{"error": "Runtime Error"}
// @router /sshkey/:projectId/:keyname/:region [post]
func (c *AWSClusterController) PostSSHKey() {

	ctx := new(utils.Context)
	ctx.SendLogs("AWSClusterController: CreateSSHKey.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	projectId := c.GetString(":projectId")
	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	//==========================RBAC Authentication==============================//

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token id is empty"}
		c.ServeJSON()
		return
	}

	teams := c.Ctx.Input.Header("teams")

	region := c.GetString(":region")
	if region == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "region is empty"}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	ctx.SendLogs("AWSClusterController: PostSSHKey.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	keyName := c.GetString(":keyname")
	if keyName == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "keyname is empty"}
		c.ServeJSON()
		return
	}

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	statusCode, awsProfile, err := aws.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	//==========================RBAC Authentication==============================//

	keyMaterial, err1 := aws.CreateSSHkey(keyName, awsProfile.Profile, token, teams, region, *ctx)
	if err1 != (types.CustomCPError{}) {
		ctx.SendLogs("AWSClusterController :"+err1.Error, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(err1.StatusCode)
		c.Data["json"] = err1
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(201)
	ctx.SendLogs(" AWS cluster key "+keyName+" created in region "+region, models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = keyMaterial
	c.ServeJSON()
}

// @Title DeleteSSHKey
// @Description Delete SSH key
// @Param	keyname	 		path	string	true		"keyname"
// @Param	X-Profile-Id	header	string	true "profileId"
// @Param	X-Auth-Token			header	string	true "token"
// @Param	region		path	string	true	"cloud region"
// @Success 204 {"msg": "SSH key deleted successfully"}
// @Failure 401 			{"error": "Unauthorized"}
// @Failure 404 			{"error": "Not Found"}
// @Failure 409 {"error": "Conflict"}
// @router /sshkey/:keyname/:region [delete]
func (c *AWSClusterController) DeleteSSHKey() {

	ctx := new(utils.Context)
	ctx.SendLogs("AWSClusterController: DeleteSSHKey.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token id is empty"}
		c.ServeJSON()
		return
	}

	region := c.GetString(":region")
	if region == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "region is empty"}
		c.ServeJSON()
		return
	}
	//==========================RBAC Authentication==============================//

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	ctx.SendLogs("AWSClusterController: DeleteSSHKey.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	keyName := c.GetString(":keyname")
	if keyName == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "keyname is empty"}
		c.ServeJSON()
		return
	}
	alreadyUsed := aws.CheckKeyUsage(keyName, userInfo.CompanyId, *ctx)
	if alreadyUsed {
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Key is used in other projects and can't be deleted"}
		c.ServeJSON()
		return
	}

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	statusCode, awsProfile, err := aws.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	err1 := aws.DeleteSSHkey(keyName, token, awsProfile.Profile, *ctx)
	if err1 != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(err1.StatusCode)
		c.Data["json"] = err1
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" AWS cluster key "+keyName+"of region"+region+" deleted", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Ctx.Output.SetStatus(204)
	c.Data["json"] = map[string]string{"msg": "Key deleted successfully"}
	c.ServeJSON()
}

// @Title GetRegions
// @Description Get AWS Regions
// @Success 200  []models.Region
// @Failure 400 {"error": "Bad Request"}
// @Failure 500  {"error": "Runtime Error"}
// @router /getallregions [get]
func (c *AWSClusterController) GetAllRegions() {

	ctx := new(utils.Context)
	ctx.SendLogs("AWSClusterController: Fetch regions.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	regions, err1 := aws.GetRegions(*ctx)
	if err1 != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(err1.StatusCode)
		c.Data["json"] = err1
		c.ServeJSON()
		return
	}

	ctx.SendLogs("Region fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = regions
	c.ServeJSON()
}

// @Title Get Availability Zone
// @Description return zones against a region
// @Param	X-Auth-Token	header	string	true "token"
// @Param	region	path	string	true	"region of AWS"
// @Param	X-Profile-Id	header	string	true "profileId"
// @Success 200 			[]*string
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 			{"error": "Bad Request"}
// @router /getzones/:region [get]
func (c *AWSClusterController) GetZones() {
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	ctx.SendLogs("AWSClusterController: fetch availability zones.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	region := c.GetString(":region")
	if region == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "region is empty"}
		c.ServeJSON()
		return
	}

	statusCode, awsProfile, err := aws.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		ctx.SendLogs("AWSClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	az, err1 := aws.GetZones(awsProfile, *ctx)
	if err1 != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(err1.StatusCode)
		c.Data["json"] = err1
		c.ServeJSON()
		return
	}
	ctx.SendLogs("Region fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = az
	c.ServeJSON()
}

// @Title Get Machine Types
// @Description Get AWS  Machine Types
// @Success 200 []string
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 			{"error": "Runtime Error"}
// @router /getallmachines [get]
func (c *AWSClusterController) GetAllMachines() {
	ctx := new(utils.Context)

	ctx.SendLogs("AWSClusterController: Fetch machine Types.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	machine, err := aws.GetAllMachines()
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("Machine types fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = machine
	c.ServeJSON()
}

// @Title Validate Profile
// @Description check if profile is valid
// @Param	X-Auth-Token	header	string	true "token"
// @Param	body	body 	vault.AwsCredentials		true	"body for cluster content"
// @Success 200 {"msg": "Cluster created successfully"}
// @Failure 401 {"error": "Profile is invalid"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /validateProfile/ [post]
func (c *AWSClusterController) ValidateProfile() {

	ctx := new(utils.Context)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
		c.ServeJSON()
		return
	}

	var credentials vault.AwsCredentials
	json.Unmarshal(c.Ctx.Input.RequestBody, &credentials)

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
	if err := json.Unmarshal(cores.AWSRegions, &regions); err != nil {
		beego.Error("Unmarshalling of machine instances failed ", err.Error())
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	for _, region := range regions {
		err := aws.ValidateProfile(credentials.AccessKey, credentials.SecretKey, region.Location, *ctx)
		if err != (types.CustomCPError{}) {
			ctx.SendLogs("AWSClusterController: Profile not valid", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			c.Ctx.Output.SetStatus(err.StatusCode)
			c.Data["json"] = err
			c.ServeJSON()
			return
		}
		if err == (types.CustomCPError{}) {
			break
		}
	}

	ctx.SendLogs("Profile Validated", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = map[string]string{"msg": "profile is valid"}
	c.ServeJSON()
}

// @Title Start Agent
// @Description Apply cloudplex Agent file to eks cluster
// @Param	clusterName	header	string	true "clusterName"
// @Param	X-Auth-Token	header	string	true "token"
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {"msg": "Agent Applied successfully"}
// @Failure 404 {"error": "Not Found"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 500 {"error": "Runtime Error"}
// @router /applyagent/:projectId [post]
func (c *AWSClusterController) ApplyAgent() {
	ctx := new(utils.Context)
	ctx.SendLogs("EKSClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//profileId := c.Ctx.Input.Header("X-Profile-Id")
	//if profileId == "" {
	//	c.Ctx.Output.SetStatus(404)
	//	c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
	//	c.ServeJSON()
	//	return
	//}
	//
	//projectId := c.GetString(":projectId")
	//if projectId == "" {
	//	ctx.SendLogs("EKSClusterController: ProjectId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	//	c.Ctx.Output.SetStatus(404)
	//	c.Data["json"] = map[string]string{"error": "project id is empty"}
	//	c.ServeJSON()
	//	return
	//}
	//
	//token := c.Ctx.Input.Header("X-Auth-Token")
	//if token == "" {
	//	c.Ctx.Output.SetStatus(404)
	//	c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
	//	c.ServeJSON()
	//	return
	//}
	//
	//clusterName := c.Ctx.Input.Header("clusterName")
	//if clusterName == "" {
	//	c.Ctx.Output.SetStatus(404)
	//	c.Data["json"] = map[string]string{"error": "clusterName is empty"}
	//	c.ServeJSON()
	//	return
	//}
	//
	//statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	//if err != nil {
	//	beego.Error(err.Error())
	//	c.Ctx.Output.SetStatus(statusCode)
	//	c.Data["json"] = map[string]string{"error": err.Error()}
	//	c.ServeJSON()
	//	return
	//}
	//
	//ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)
	//ctx.SendLogs("EKSClusterController: Apply Agent.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	//
	//statusCode, allowed, err := rbac_athentication.Authenticate(models.GKE, "cluster", projectId, "Start", token, utils.Context{})
	//if err != nil {
	//	beego.Error(err.Error())
	//	c.Ctx.Output.SetStatus(statusCode)
	//	c.Data["json"] = map[string]string{"error": err.Error()}
	//	c.ServeJSON()
	//	return
	//}
	//if !allowed {
	//	c.Ctx.Output.SetStatus(401)
	//	c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
	//	c.ServeJSON()
	//	return
	//}
	//region, err := aws.GetRegion(token, projectId, *ctx)
	//if err != nil {
	//	c.Ctx.Output.SetStatus(500)
	//	c.Data["json"] = map[string]string{"error": err.Error()}
	//	c.ServeJSON()
	//	return
	//}
	//statusCode, awsProfile, err := aws.GetProfile(profileId, region, token, *ctx)
	//if err != nil {
	//	utils.SendLog(userInfo.CompanyId, err.Error(), "error", projectId)
	//	c.Ctx.Output.SetStatus(statusCode)
	//	c.Data["json"] = map[string]string{"error": err.Error()}
	//	c.ServeJSON()
	//	return
	//}
	//ctx.SendLogs("EKSClusterController: applying agent on cluster . "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	//
	////go aws.ApplyAgent(awsProfile, token, *ctx, clusterName)

	c.Data["json"] = map[string]string{"msg": "agent deployment in progress"}
	c.ServeJSON()
}
