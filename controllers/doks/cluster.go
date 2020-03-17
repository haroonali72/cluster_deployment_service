package doks

import (
	"antelope/models"
	"antelope/models/do"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/astaxie/beego"
	"strings"
	"time"
)

type DOKSClusterController struct {
	beego.Controller
}

// @Title Create
// @Description create a new managed cluster
// @Param	body	body 	do.DOKSCluster	true	"body for cluster  content"
// @Param	token	header	string	token ""
// @Success 200 {"msg": "kubernetes cluster created successfully"}
// @Success 400 {"msg": "error msg"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "not found"}
// @Failure 409 {"error": "cluster against this project already exists"}
// @Failure 500 {"error": "error msg"}
// @router / [post]
func (c *DOKSClusterController) Post() {

	var cluster do.DOKSCluster
	json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	cluster.CreationDate = time.Now()

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
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, cluster.ProjectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	allowed, err := rbac_athentication.Authenticate(models.DO, "cluster", cluster.ProjectId, "Create", token, *ctx)
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

	//=============================================================================//

	ctx.SendLogs("DOClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	res, err := govalidator.ValidateStruct(cluster)
	if !res || err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	err = do.GetNetwork(token, cluster.ProjectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	cluster.CompanyId = userInfo.CompanyId
	err = do.CreateCluster(cluster, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "cluster against this project id  already exists"}
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
	ctx.SendLogs(" DO cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" created ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster added successfully"}
	c.ServeJSON()
}

// @Title Start
// @Description starts a  cluster
// @Param	token	header	string	token ""
// @Param	X-Profile-Id	header	string	profileId	""
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 401 {"error": "error msg"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 400 {"error": "error msg"}
// @Failure 500 {"error": "error msg"}
// @router /start/:projectId [post]
func (c *DOClusterController) StartCluster() {

	projectId := c.GetString(":projectId")
	if projectId == "" {
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
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	allowed, err := rbac_athentication.Authenticate(models.DO, "cluster", projectId, "Start", token, *ctx)
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

	//=============================================================================//

	ctx.SendLogs("DONetworkController: StartCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	var cluster do.Cluster_Def

	ctx.SendLogs("DOClusterController: Getting Cluster of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = do.GetCluster(projectId, userInfo.CompanyId, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if cluster.Status == "Cluster Created" {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is already in running state"}
		c.ServeJSON()
		return
	}

	if cluster.Status == string(models.Deploying) {
		ctx.SendLogs("cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in deploying state"}
		c.ServeJSON()
		return
	}

	if cluster.Status == string(models.Terminating) {
		ctx.SendLogs("cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in terminating state"}
		c.ServeJSON()
		return
	}

	region, err := do.GetRegion(token, projectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	doProfile, err := do.GetProfile(profileId, region, token, *ctx)

	if err != nil {
		utils.SendLog(userInfo.CompanyId, err.Error(), "error", cluster.ProjectId)
		utils.SendLog(userInfo.CompanyId, "Cluster creation failed: "+cluster.Name, "error", cluster.ProjectId)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	cluster.Status = string(models.Deploying)
	err = do.UpdateCluster(cluster, false, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("DOClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go do.DeployCluster(cluster, doProfile.Profile, *ctx, userInfo.CompanyId, token)

	ctx.SendLogs(" DO cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" deployed ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster creation in progress"}
	c.ServeJSON()
}
