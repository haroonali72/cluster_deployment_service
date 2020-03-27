package aks

import (
	"antelope/models"
	"antelope/models/aks"
	"antelope/models/azure"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/astaxie/beego"
	"strings"
)

// Operations about AKS cluster [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/cluster/{cloud}/]
type AKSClusterController struct {
	beego.Controller
}

// @Title Get
// @Description get cluster
// @Param	projectId	path	string	true	"Id of the project"
// @Param	token	header	string	token ""
// @Success 200 {object} aks.AKSCluster
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /:projectId/ [get]
func (c *AKSClusterController) Get() {
	ctx := new(utils.Context)

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("AKSClusterController: projectId is empty", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("AKSClusterController: Get cluster with project id "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	allowed, err := rbacAuthentication.Authenticate(models.AKS, "cluster", projectId, "View", token, utils.Context{})
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

	ctx.SendLogs("AKSClusterController: Get cluster with project id: "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := aks.GetAKSCluster(projectId, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("AKSGetClusterController: error getting gke cluster "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "no cluster exists for this name"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs(" AKS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Get All
// @Description get all the clusters
// @Param	token	header	string	token ""
// @Success 200 {object} []aks.AKSCluster
// @Failure 400 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /all [get]
func (c *AKSClusterController) GetAll() {
	ctx := new(utils.Context)
	ctx.SendLogs("AKSClusterController: GetAll clusters.", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

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
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	ctx.SendLogs("AKSClusterController: Getting all clusters ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err, data := rbacAuthentication.GetAllAuthenticate("cluster", userInfo.CompanyId, token, models.AKS, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	clusters, err := aks.GetAllAKSCluster(data, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("All AKS cluster fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = clusters
	c.ServeJSON()
}

// @Title Add
// @Description add a new cluster
// @Param	token	header	string	token ""
// @Param body body aks.AKSCluster true	"body for cluster content"
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 409 {"error": "cluster against same project id already exists"}
// @Failure 410 {"error": "Core limit exceeded"}
// @Failure 500 {"error": "error msg"}
// @router / [post]
func (c *AKSClusterController) Post() {

	var cluster aks.AKSCluster

	ctx := new(utils.Context)

	_ = json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	token := c.Ctx.Input.Header("token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	res, err := govalidator.ValidateStruct(cluster)
	if !res || err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, cluster.ProjectId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("AKSClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	allowed, err := rbacAuthentication.Authenticate(models.AKS, "cluster", cluster.ProjectId, "Create", token, utils.Context{})
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

	ctx.SendLogs("AKSClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	beego.Info("AKSClusterController: JSON Payload: ", cluster)

	cluster.CompanyId = userInfo.CompanyId

	err = aks.AddAKSCluster(cluster, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	ctx.SendLogs("AKS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" created ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster added successfully"}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing cluster
// @Param	token	header	string	token ""
// @Param	body	body 	aks.AKSCluster	true	"body for cluster content"
// @Success 200 {"msg": "cluster updated successfully"}
// @Failure 400 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 402 {"error": "error msg"}
// @Failure 404 {"error": "no cluster exists with this name"}
// @Failure 500 {"error": "error msg"}
// @router / [put]
func (c *AKSClusterController) Patch() {
	ctx := new(utils.Context)

	var cluster aks.AKSCluster
	_ = json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

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
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, cluster.ProjectId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("AKSClusterController: update cluster cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	allowed, err := rbacAuthentication.Authenticate(models.AKS, "cluster", cluster.ProjectId, "Update", token, utils.Context{})
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
	ctx.SendLogs("AKSClusterController: Patch cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	beego.Info("AKSClusterController: JSON Payload: ", cluster)

	err = aks.UpdateAKSCluster(cluster, *ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	ctx.SendLogs("AKS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
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
// @router /:projectId/:forceDelete [delete]
func (c *AKSClusterController) Delete() {
	ctx := new(utils.Context)

	id := c.GetString(":projectId")
	if id == "" {
		ctx.SendLogs("AKSClusterController: ProjectId field is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, id, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("AKSClusterController: Delete cluster with id "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	allowed, err := rbacAuthentication.Authenticate(models.AKS, "cluster", id, "Delete", token, utils.Context{})
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

	ctx.SendLogs("AKSClusterController: Delete cluster with project id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := aks.GetAKSCluster(id, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if strings.ToLower(cluster.Status) == string(models.ClusterCreated) && !forceDelete {
		ctx.SendLogs("AKSClusterController: Cluster is in running state ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "Cluster is in running state"}
		c.ServeJSON()
		return
	}

	if cluster.Status == string(models.Deploying) && !forceDelete {
		ctx.SendLogs("AKSClusterController: Cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in deploying state"}
		c.ServeJSON()
		return
	}

	if cluster.Status == string(models.Terminating) && !forceDelete {
		ctx.SendLogs("AKSClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in terminating state"}
		c.ServeJSON()
		return
	}

	err = aks.DeleteAKSCluster(id, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AKS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" deleted ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
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
func (c *AKSClusterController) StartCluster() {

	ctx := new(utils.Context)
	ctx.SendLogs("AKSClusterController: StartCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("AKSClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("AKSClusterController: ProjectId field is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	allowed, err := rbacAuthentication.Authenticate(models.AKS, "cluster", projectId, "Start", token, utils.Context{})
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

	region, err := azure.GetRegion(token, projectId, *ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	azureProfile, err := azure.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController : Unable to get profile", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AKSClusterController: Getting Cluster of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := aks.GetAKSCluster(projectId, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if cluster.Status == "Cluster Created" {
		ctx.SendLogs("AKSClusterController : Cluster is already running", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is already in running state"}
		c.ServeJSON()
		return
	}

	//if cluster.Status == string(models.Deploying) {
	//	ctx.SendLogs("AKSClusterController: Cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	//	c.Ctx.Output.SetStatus(400)
	//	c.Data["json"] = map[string]string{"error": "cluster is in deploying state"}
	//	c.ServeJSON()
	//	return
	//}
	//
	//if cluster.Status == string(models.Terminating) {
	//	ctx.SendLogs("AKSClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	//	c.Ctx.Output.SetStatus(400)
	//	c.Data["json"] = map[string]string{"error": "cluster is in terminating state"}
	//	c.ServeJSON()
	//	return
	//}
	//
	//cluster.Status = string(models.Deploying)
	//err = aks.UpdateAKSCluster(cluster, *ctx)
	//if err != nil {
	//	c.Ctx.Output.SetStatus(500)
	//	c.Data["json"] = map[string]string{"error": err.Error()}
	//	c.ServeJSON()
	//	return
	//}
	ctx.SendLogs("AKSClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go aks.DeployAKSCluster(cluster, azureProfile, userInfo.CompanyId, token, *ctx)

	ctx.SendLogs(" AKS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" deployed ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster creation in progress"}
	c.ServeJSON()
}

// @Title Status
// @Description returns status of nodes
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	token	header	string	token ""
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {object} aks.AKSCluster
// @Failure 206 {object} aks.AKSCluster
// @Failure 400 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @Failure 401 {"error": "authorization params missing or invalid"}
// @Failure 500 {"error": "error msg"}
// @router /status/:projectId/ [get]
func (c *AKSClusterController) GetStatus() {
	ctx := new(utils.Context)
	ctx.SendLogs("AKSClusterController: FetchStatus.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("AKSClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("AKSClusterController: ProjectId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("AKSClusterController: FetchStatus.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	allowed, err := rbacAuthentication.Authenticate(models.AKS, "cluster", projectId, "View", token, utils.Context{})
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
	region, err := azure.GetRegion(token, projectId, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	azureProfile, err := azure.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController : Unable to get profile", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AKSClusterController: Fetch Cluster Status of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := aks.FetchStatus(azureProfile.Profile, token, projectId, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
func (c *AKSClusterController) TerminateCluster() {
	ctx := new(utils.Context)
	ctx.SendLogs("AKSClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("AKSClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("AKSClusterController: ProjectId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("AKSClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	allowed, err := rbacAuthentication.Authenticate(models.AKS, "cluster", projectId, "Terminate", token, utils.Context{})
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

	region, err := azure.GetRegion(token, projectId, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	azureProfile, err := azure.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController : Unable to get profile", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AKSClusterController: Getting Cluster of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := aks.GetAKSCluster(projectId, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if cluster.Status == "Cluster Terminated" {
		ctx.SendLogs("AKSClusterController : Cluster is terminated", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is already in terminated state"}
		c.ServeJSON()
		return
	}

	//if cluster.Status == string(models.Deploying) {
	//	ctx.SendLogs("AKSClusterController: cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	//	c.Ctx.Output.SetStatus(400)
	//	c.Data["json"] = map[string]string{"error": "cluster is in deploying state"}
	//	c.ServeJSON()
	//	return
	//}
	//
	//if cluster.Status == string(models.Terminating) {
	//	ctx.SendLogs("AKSClusterController: cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	//	c.Ctx.Output.SetStatus(400)
	//	c.Data["json"] = map[string]string{"error": "cluster is in terminating state"}
	//	c.ServeJSON()
	//	return
	//}
	//
	//cluster.Status = string(models.Terminating)
	//err = aks.UpdateAKSCluster(cluster, *ctx)
	//if err != nil {
	//	c.Ctx.Output.SetStatus(500)
	//	c.Data["json"] = map[string]string{"error": err.Error()}
	//	c.ServeJSON()
	//	return
	//}

	go aks.TerminateCluster(azureProfile, projectId, userInfo.CompanyId, *ctx)

	ctx.SendLogs(" AKS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" terminated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster termination is in progress"}
	c.ServeJSON()
}

// @Title GetAKSVmsTypes
// @Description get aks vm types
// @Param	token	header	string	token ""
// @Success 200 {object} []aks.VMSizeTypes
// @Failure 401 {"error": "Authorization format should be 'base64 encoded service_account_json'"}
// @Failure 400 {"error": "error_msg"}
// @Failure 404 {"error": "error_msg"}
// @Failure 500 {"error": "error msg"}
// @router /getvms/ [get]
func (c *AKSClusterController) GetAKSVms() {
	ctx := new(utils.Context)
	ctx.SendLogs("AKSClusterController: GetVms.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("AKSClusterController: GetVms.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	aksVms := aks.GetAKSVms(*ctx)

	c.Data["json"] = aksVms
	c.ServeJSON()
}

// @Title Kubeconfig
// @Description get cluter kubeconfig
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	token	header	string	token ""
// @Param	projectId	path	string	true	"Id of the project"
// @Failure 400 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /kubeconfig/:projectId [get]
func (c *AKSClusterController) GetKubeConfig() {

	ctx := new(utils.Context)
	ctx.SendLogs("AKSClusterController: GetKubeConfig.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("AKSClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("AKSClusterController: ProjectId field is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	region, err := azure.GetRegion(token, projectId, *ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	azureProfile, err := azure.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController : Unable to get profile", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AKSClusterController: Getting Cluster of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := aks.GetAKSCluster(projectId, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("AKSClusterController: GetKubeConfig. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	kubeconfig, err := aks.GetKubeCofing(azureProfile.Profile, cluster, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = kubeconfig
	c.ServeJSON()
}

// @Title Get Kube Versions
// @Description fetch version of kubernetes cluster
// @Param	token	header	string	token ""
// @Success 200 {object} []string
// @Failure 400 {"error": "error msg"}
// @Failure 404 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /getallkubeversions/ [get]
func (c *AKSClusterController) FetchKubeVersions() {

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
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("AKSClusterController: GetAllKubernetesVersions.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	kubeVersions := aks.GetKubeVersions(*ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	c.Data["json"] = kubeVersions
	c.ServeJSON()
}