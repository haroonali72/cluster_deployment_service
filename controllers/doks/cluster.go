package doks

import (
	"antelope/models"
	"antelope/models/do"
	"antelope/models/doks"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"encoding/json"
	"github.com/astaxie/beego"
	"strings"
)

type DOKSClusterController struct {
	beego.Controller
}

// @Title Get
// @Description get kubernetes version,machine types and regions
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	token	header	string	true "token"
// @Success 200 {object} doks.ServerConfig
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router /config [get]
func (c *DOKSClusterController) GetServerConfig() {

	ctx := new(utils.Context)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("DOKSClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("token")
	if token == "" {
		ctx.SendLogs("DOKSClusterController: Token is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	/*region, err := do.GetRegion(token, projectId, *ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	*/
	region := "nyc1"
	doProfile, err := do.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	config, err1 := doks.GetServerConfig(doProfile.Profile, *ctx)
	if err1.Description != "" {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err1.Description}
		c.ServeJSON()
		return
	}

	c.Data["json"] = config
	c.ServeJSON()
}

// @Title Get
// @Description get valid kubernetes cluster version and machine sizes
// @Param	token	header	string	true "token"
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {object} doks.KubernetesConfig
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router /kubeconfig/:projectId [get]
func (c *DOKSClusterController) GetKubeConfig() {

	ctx := new(utils.Context)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("DOKSClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("DOKSClusterController: ProjectId field is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("token")
	if token == "" {
		ctx.SendLogs("DOKSClusterController: Token is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	region, err := do.GetRegion(token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	doProfile, err := do.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	cluster, err := doks.GetKubernetesCluster(*ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	config, err1 := doks.GetKubeConfig(doProfile.Profile, *ctx, cluster)
	if err1.Description != "" {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err1.Description}
		c.ServeJSON()
		return
	}

	c.Data["json"] = config
	c.ServeJSON()
}

// @Title Get
// @Description  get cluster against the projectId
// @Param	projectId	path	string	true	"Id of the project"
// @Param	token	header	string	true "token"
// @Success 200 {object} doks.KubernetesCluster
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router /:projectId/ [get]
func (c *DOKSClusterController) Get() {
	ctx := new(utils.Context)

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("DOKSClusterController: projectId is empty", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("token")
	if token == "" {
		ctx.SendLogs("DOKSClusterController: token is empty", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "token is empty"}
		c.ServeJSON()
		return
	}

	userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		ctx.SendLogs("DOKSClusterController: RBAC:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.Data.Company=userInfo.CompanyId
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId, ctx.Data.Company, userInfo.UserId)
	ctx.SendLogs("DOKSClusterController: Get cluster with project id "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	allowed, err := rbacAuthentication.Authenticate(models.DOKS, "cluster", projectId, "View", token, utils.Context{})
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

	ctx.SendLogs("DOKSClusterController: Get cluster with project id: "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := doks.GetKubernetesCluster(*ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" DOKS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Get All
// @Description get all the clusters
// @Param	token	header	string	true "token"
// @Success 200 {object} []doks.KubernetesCluster
// @Failure 400 {"error": "Bad Request"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router /all [get]
func (c *DOKSClusterController) GetAll() {

	ctx := new(utils.Context)
	ctx.SendLogs("DOKSClusterController: GetAll clusters.", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

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

	ctx.SendLogs("DOKSClusterController: Getting all clusters ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	_, data := rbacAuthentication.GetAllAuthenticate("cluster",ctx.Data.Company, token, models.DOKS, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	clusters, err := doks.GetAllKubernetesCluster(data, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("All DOKS cluster fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = clusters
	c.ServeJSON()
}

// @Title Add
// @Description add a new cluster
// @Param	token	header	string	true "token"
// @Param	body	body 	doks.KubernetesCluster		true	"body for cluster content"
// @Success 200 {"msg": "Cluster added successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not found"}
// @Failure 409 {"error": "Cluster against same project id already exists"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router / [post]
func (c *DOKSClusterController) Post() {

	var cluster doks.KubernetesCluster

	ctx := new(utils.Context)

	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
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

/*	_, err = govalidator.ValidateStruct(cluster)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
*/

	userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, cluster.ProjectId, ctx.Data.Company, userInfo.UserId)

	allowed, err := rbacAuthentication.Authenticate(models.DOKS, "cluster", cluster.ProjectId, "Create", token, utils.Context{})
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

	ctx.SendLogs("DOKSClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	beego.Info("DOKSClusterController: JSON Payload: ", cluster)

	cluster.CompanyId = userInfo.CompanyId

	err = doks.AddKubernetesCluster(cluster, *ctx)
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

	ctx.SendLogs("DOKS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" created ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster added successfully"}
	c.ServeJSON()

}

// @Title Update
// @Description update an existing kubernetes cluster
// @Param	token	header	string	true "token"
// @Param	body	body 	doks.KubernetesCluster	true	"body for cluster content"
// @Success 200 {"msg": "Cluster updated successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 402 {"error": "Cluster is in running state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router / [put]
func (c *DOKSClusterController) Patch() {
	ctx := new(utils.Context)

	var cluster doks.KubernetesCluster
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
	ctx.SendLogs("DOKSClusterController: update cluster cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	_, err = rbacAuthentication.Authenticate(models.DOKS, "cluster", cluster.ProjectId, "Update", token, utils.Context{})
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
/*	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}
*/
	ctx.SendLogs("DOKSClusterController: Patch cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	beego.Info("DOKSClusterController: JSON Payload: ", cluster)
	ctx.Data.Company = userInfo.CompanyId
	err = doks.UpdateKubernetesCluster(cluster, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": err.Error()}
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

	ctx.SendLogs("DOKS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster updated successfully"}
	c.ServeJSON()
}

// @Title Delete
// @Description delete a kubernetes cluster
// @Param	projectId	path	string	true	"project id of the cluster"
// @Param	forceDelete path  boolean	true ""
// @Param	token	header	string	true "token"
// @Success 200 {"msg": "cluster deleted successfully"}
// @Success 204 {"msg": "cluster deleted successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router /:projectId/:forceDelete [delete]
func (c *DOKSClusterController) Delete() {
	ctx := new(utils.Context)

	id := c.GetString(":projectId")
	if id == "" {
		ctx.SendLogs("DOKSClusterController: ProjectId field is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
	ctx.SendLogs("DOKSClusterController: Delete cluster with id "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	allowed, err := rbacAuthentication.Authenticate(models.DOKS, "cluster", id, "Delete", token, utils.Context{})
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

	ctx.SendLogs("DOKSClusterController: Delete cluster with project id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := doks.GetKubernetesCluster( *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if strings.ToLower(cluster.CloudplexStatus) == string(models.ClusterCreated) && !forceDelete {
		ctx.SendLogs("DOKSClusterController: Cluster is in running state ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(402)
		c.Data["json"] = map[string]string{"error": "Cluster is in running state"}
		c.ServeJSON()
		return
	}

	if cluster.CloudplexStatus == string(models.Deploying) && !forceDelete {
		ctx.SendLogs("DOKSClusterController: Cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in deploying state"}
		c.ServeJSON()
		return
	}

	if cluster.CloudplexStatus == string(models.Terminating) && !forceDelete {
		ctx.SendLogs("DOKSClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in terminating state"}
		c.ServeJSON()
		return
	}

	err = doks.DeleteKubernetesCluster(*ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("DOKS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" deleted ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster deleted successfully"}
	c.ServeJSON()
}

// @Title Start
// @Description starts a kubernetes cluster
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	token	header	string	true "token"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {"msg": "Cluster created successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 402 {"error": "Cluster is in running state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router /start/:projectId [post]
func (c *DOKSClusterController) StartCluster() {

	ctx := new(utils.Context)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("DOKSClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("DOKSClusterController: ProjectId field is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
	_, err = rbacAuthentication.Authenticate(models.DOKS, "cluster", projectId, "Start", token, utils.Context{})
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
/*	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}
*/
	region:="nyc1"
/*	region, err := do.GetRegion(token,  *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
*/

	ctx.SendLogs("DOKSClusterController: Getting Cluster of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := doks.GetKubernetesCluster(*ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	doProfile, err := do.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if cluster.CloudplexStatus == "Cluster Created" {
		ctx.SendLogs("DOKSClusterController : Cluster is already running", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(402)
		c.Data["json"] = map[string]string{"error": "cluster is already in running state"}
		c.ServeJSON()
		return
	}

	if cluster.CloudplexStatus == string(models.Deploying) {
		ctx.SendLogs("DOKSClusterController: Cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in deploying state"}
		c.ServeJSON()
		return
	}

	if cluster.CloudplexStatus == string(models.Terminating) {
		ctx.SendLogs("DOKSClusterContro<ller: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in terminating state"}
		c.ServeJSON()
		return
	}

	cluster.CloudplexStatus = string(models.Deploying)
	err = doks.UpdateKubernetesCluster(cluster, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("DOKSClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go doks.DeployKubernetesCluster(cluster, doProfile.Profile, token, *ctx)

	ctx.SendLogs(" DOKS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" deployed ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster creation in progress"}
	c.ServeJSON()
}

// @Title Status
// @Description returns latest status of the running cluster
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	token	header	string	true "token"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {object} doks.DOKSCluster
// @Failure 206 {object} doks.DOKSCluster
// @Failure 400 {"error": "Bad Request"}
// @Failure 404 {"error": "Not Found"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router /status/:projectId/ [get]
func (c *DOKSClusterController) GetStatus() {

	ctx := new(utils.Context)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("DOKSClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("DOKSClusterController: ProjectId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId,ctx.Data.Company, userInfo.UserId)
	ctx.SendLogs("DOKSClusterController: FetchStatus.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	_, err = rbacAuthentication.Authenticate(models.DOKS, "cluster", projectId, "View", token, utils.Context{})
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
/*	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}
*/
region:="nyc1"
/*	region, err := do.GetRegion(token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
*/
	doProfile, err := do.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("DOKSClusterController: Fetch Cluster Status of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.Data.Company=userInfo.CompanyId
	cluster, errr := doks.FetchStatus(doProfile.Profile, *ctx)
	if errr.Description != "" {
		c.Ctx.Output.SetStatus(206)
	}

	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Terminate
// @Description terminates a  cluster
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	projectId	path	string	true	"Id of the project"
// @Param	token	header	string	true "token"
// @Success 200 {"msg": "cluster terminated successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 502 {"error": "Cloud API Error"}
// @router /terminate/:projectId/ [post]
func (c *DOKSClusterController) TerminateCluster() {
	ctx := new(utils.Context)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("DOKSClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("DOKSClusterController: ProjectId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
	ctx.SendLogs("DOKSClusterController: TerminateKubernetesCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.Data.Company=userInfo.CompanyId
	_, err = rbacAuthentication.Authenticate(models.DOKS, "cluster", projectId, "Terminate", token, utils.Context{})
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
/*	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}
*/
region:="nyc1"
/*
	region, err := do.GetRegion(token, *ctx)
	if err != nil {
		ctx.SendLogs("DOKSClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
*/
	ctx.SendLogs("DOKSClusterController: Getting Cluster of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.Data.Company=userInfo.CompanyId
	cluster, err := doks.GetKubernetesCluster(*ctx)
	if err != nil {
		ctx.SendLogs("DOKSClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if cluster.CloudplexStatus == string(models.Deploying) {
		ctx.SendLogs("DOKSClusterController: cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in deploying state"}
		c.ServeJSON()
		return
	}

	if cluster.CloudplexStatus == string(models.Terminating) {
		ctx.SendLogs("DOKSClusterController: cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in terminating state"}
		c.ServeJSON()
		return
	}

	doProfile, err := do.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	go doks.TerminateCluster(doProfile.Profile, *ctx)

	err = doks.UpdateKubernetesCluster(cluster, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" DOKS cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" terminated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster termination is in progress"}
	c.ServeJSON()
}

// @Title Start
// @Description Apply cloudplex Agent file to doks cluster
// @Param	clusterName	header	string	clusterName ""
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
func (c *DOKSClusterController) ApplyAgent() {

	ctx := new(utils.Context)
	ctx.SendLogs("DOKSClusterController: Apply Agent", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("DOKSClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	projectId := c.GetString(":projectId")
	if projectId == "" {
		ctx.SendLogs("DOKSClusterController: ProjectId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId, ctx.Data.Company, userInfo.UserId)
	ctx.SendLogs("DOKSClusterController: Apply Agent.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	allowed, err := rbacAuthentication.Authenticate(models.DOKS, "cluster", projectId, "Start", token, utils.Context{})
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

	region, err := do.GetRegion(token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	doProfile, err := do.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("DOKSClusterController: applying agent on cluster . "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go doks.ApplyAgent(doProfile.Profile, token, *ctx, clusterName)

	c.Data["json"] = map[string]string{"msg": "agent deployment in progress"}
	c.ServeJSON()
}
