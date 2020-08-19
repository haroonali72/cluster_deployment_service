package do

import (
	"antelope/models"
	"antelope/models/aws"
	"antelope/models/do"
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

// Operations about DO cluster [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/cluster/{cloud}/]
type DOClusterController struct {
	beego.Controller
}

// @Title Get
// @Description get cluster
// @Param infraId path string true "Id of the infrastructure"
// @Param X-Auth-Token	header string true "token"
// @Success 200 {object} do.Cluster_Def
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /:infraId/ [get]
func (c *DOClusterController) Get() {

	infraId := c.GetString(":infraId")
	if infraId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "infrastructure id is empty"}
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
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbac_athentication.Authenticate(models.DO, "cluster", infraId, "View", token, *ctx)
	if err != nil {
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

	ctx.SendLogs("DOClusterController: Get cluster with infrastructure id: "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := do.GetCluster(infraId, userInfo.CompanyId, *ctx)

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
	ctx.SendLogs(" DO cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+" fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Get All
// @Description get all the clusters
// @Param X-Auth-Token header string token ""
// @Success 200 {object} []do.Cluster_Def
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /all [get]
func (c *DOClusterController) GetAll() {
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

	statusCode, err, data := rbac_athentication.GetAllAuthenticate("cluster", userInfo.CompanyId, token, models.DO, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	//====================================================================================//
	ctx.SendLogs("DOClusterController: GetAll clusters.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	clusters, err := do.GetAllCluster(*ctx, data)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" All DO clusters fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = clusters
	c.ServeJSON()
}

// @Title Create
// @Description create a new cluster
// @Param body body do.Cluster_Def true "body for cluster content"
// @Param X-Auth-Token header string token ""
// @Success 201 {"msg": "Cluster added successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not found"}
// @Failure 409 {"error": "Cluster against same infrastructure already exists"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [post]
func (c *DOClusterController) Post() {

	var cluster do.Cluster_Def
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		c.Ctx.ResponseWriter.WriteHeader(400)
		c.Data["json"] = map[string]string{"error": "error while unmarshalling " + err.Error()}
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
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, cluster.InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.DO, "cluster", cluster.InfraId, "Create", token, *ctx)
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

	ctx.SendLogs("DOClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	validate := validator.New()
	err = validate.Struct(cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	err = do.ValidateDOData(cluster, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
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
		}*/

	err = do.GetNetwork(token, cluster.InfraId, *ctx)
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
			c.Data["json"] = map[string]string{"error": "cluster against this infrastructure id  already exists"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" DO cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+" created ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster added successfully"}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing cluster
// @Param X-Auth-Token header string token ""
// @Param body body do.Cluster_Def true "body for cluster content"
// @Success 200 {"msg": "Cluster updated successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in Creating/Terminating/TerminationFailed state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [put]
func (c *DOClusterController) Patch() {
	var cluster do.Cluster_Def
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "error while unmarshalling " + err.Error()}
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
	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, cluster.InfraId, userInfo.CompanyId, userInfo.UserId)

	if cluster.Status == (models.Deploying) {
		ctx.SendLogs("DoClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Terminating) {
		ctx.SendLogs("DoClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.ClusterTerminationFailed) {
		ctx.SendLogs("DoClusterController: Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": " Cluster creation is in termination failed state"}
		c.ServeJSON()
		return
	}
	if cluster.Status == (models.ClusterCreated) {
		c.Data["json"] = map[string]string{"msg": "No changes are applicable"}
		c.ServeJSON()
	}

	validate := validator.New()
	err = validate.Struct(cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	err = do.ValidateDOData(cluster, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.DO, "cluster", cluster.InfraId, "Update", token, *ctx)
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
	/*	if cluster.Status != models.New && cluster.Status != models.ClusterCreationFailed && cluster.Status != models.ClusterTerminated {
		ctx.SendLogs("DOClusterController : Cluster is in "+string(cluster.Status)+" state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Can't Update.Cluster is in " + string(cluster.Status) + " state"}
		c.ServeJSON()
		return
	}*/
	ctx.SendLogs("DOClusterController: Patch cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	cluster.CompanyId = userInfo.CompanyId
	err = do.UpdateCluster(cluster, true, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "no cluster exists with this name"}
			c.ServeJSON()
			return
		}
		if strings.Contains(err.Error(), "cluster is in deploying state") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		}
		if strings.Contains(err.Error(), "cluster is in terminating state") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		} else if strings.Contains(err.Error(), "Cluster is in termination failed state") {
			c.Ctx.Output.SetStatus(int(models.StateConflict))
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		} else if strings.Contains(err.Error(), "No changes are applicable") {
			c.Data["json"] = map[string]string{"msg": string(models.SuccessfullyUpdated)}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" DO cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster updated successfully"}
	c.ServeJSON()
}

// @Title Delete
// @Description delete a cluster
// @Param X-Auth-Token header string token ""
// @Param infraId path string true "infrastructure id of the cluster"
// @Param forceDelete path boolean true ""
// @Success 204 {"msg": "Cluster deleted successfully"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in Creating/Created/Terminating/TerminationFailed state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /:infraId/:forceDelete [delete]
func (c *DOClusterController) Delete() {
	id := c.GetString(":infraId")
	if id == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "infrastructure id is empty"}
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
	statusCode, allowed, err := rbac_athentication.Authenticate(models.DO, "cluster", id, "Delete", token, *ctx)
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

	ctx.SendLogs("DOClusterController: Delete cluster with infrastructure id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := do.GetCluster(id, userInfo.CompanyId, *ctx)
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
		c.Data["json"] = map[string]string{"error": err.Error() + " + Cluster is in running state"}
		c.ServeJSON()
		return
	}

	if cluster.Status == (models.Deploying) && !forceDelete {
		ctx.SendLogs("cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "cluster is in deploying state"}
		c.ServeJSON()
		return
	}

	if cluster.Status == (models.Terminating) && !forceDelete {
		ctx.SendLogs("cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "cluster is in terminating state"}
		c.ServeJSON()
		return
	}
	if cluster.Status == (models.ClusterTerminationFailed) && !forceDelete {
		ctx.SendLogs("DOClusterController: Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "cluster is in termination failed state"}
		c.ServeJSON()
		return
	}
	err = do.DeleteCluster(id, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" DO cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+" deleted ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster deleted successfully"}
	c.ServeJSON()
}

// @Title Start
// @Description starts a  cluster
// @Param X-Auth-Token header string true "Token"
// @Param X-Profile-Id header string true "profileId"
// @Param infraId path string	true "Id of the infrastructure"
// @Success 202 {"msg": "Cluster creation initiated"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in Created/Creating/Terminating/TerminationFailed state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /start/:infraId [post]
func (c *DOClusterController) StartCluster() {

	infraId := c.GetString(":infraId")
	if infraId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "infrastructure id is empty"}
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
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.DO, "cluster", infraId, "Start", token, *ctx)
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

	ctx.SendLogs("DONetworkController: StartCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	var cluster do.Cluster_Def

	ctx.SendLogs("DOClusterController: Getting Cluster of infrastructure. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = do.GetCluster(infraId, userInfo.CompanyId, *ctx)

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

	if cluster.Status == models.ClusterCreated {
		ctx.SendLogs("DOClusterController : Cluster is already running", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is already in running state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Deploying) {
		ctx.SendLogs("DOClusterController: Cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Terminating) {
		ctx.SendLogs("DOClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.ClusterTerminationFailed) {
		ctx.SendLogs("DOClusterController: Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in termination failed state"}
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
	statusCode, doProfile, err := do.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		utils.SendLog(userInfo.CompanyId, err.Error(), "error", cluster.InfraId)
		utils.SendLog(userInfo.CompanyId, "Cluster creation failed: "+cluster.Name, "error", cluster.InfraId)
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	cluster.Status = (models.Deploying)
	err = do.UpdateCluster(cluster, false, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("DOClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go do.DeployCluster(cluster, doProfile.Profile, *ctx, userInfo.CompanyId, token)

	ctx.SendLogs(" DO cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+" deployed ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster creation in progress"}
	c.ServeJSON()
}

// @Title Status
// @Description returns status of nodes
// @Param X-Auth-Token header string true "Token"
// @Param X-Profile-Id header string true "profileId"
// @Param infraId path string	true "Id of the infrastructure"
// @Success 200 {object} do.Cluster_Def
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Cluster is in deploying/terminating state"}
// @Failure 500 {"error": "Runtime Error"}
// @router /status/:infraId/ [get]
func (c *DOClusterController) GetStatus() {

	infraId := c.GetString(":infraId")
	if infraId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "infrastructure id is empty"}
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
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.DO, "cluster", infraId, "View", token, *ctx)
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
	ctx.SendLogs("DOClusterController: FetchStatus.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("DOClusterController: Fetch Cluster Status of infrastructure. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	region, err := do.GetRegion(token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	statusCode, awsProfile, err := do.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	cluster, cpErr := do.FetchStatus(awsProfile, infraId, *ctx, userInfo.CompanyId, token)
	if cpErr != (types.CustomCPError{}) && strings.Contains(strings.ToLower(cpErr.Description), "state") || cpErr != (types.CustomCPError{}) && strings.Contains(strings.ToLower(cpErr.Description), "not deployed") {
		c.Ctx.Output.SetStatus(cpErr.StatusCode)
		c.Data["json"] = cpErr.Description
		c.ServeJSON()
		return
	}
	if cpErr != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(int(models.CloudStatusCode))
		c.Data["json"] = cpErr
		c.ServeJSON()
	}

	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Terminate
// @Description terminates a  cluster
// @Param X-Profile-Id header X-Profile-Id string true "profileId"
// @Param X-Auth-Token header string true "Token"
// @Param infraId path string true "Id of the infrastructure"
// @Success 202 {"msg": "cluster termination initiated"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in New/Creating/Creation Failed /Terminated/Terminating state"}/
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /terminate/:infraId/ [post]
func (c *DOClusterController) TerminateCluster() {

	infraId := c.GetString(":infraId")
	if infraId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "infrastructure id is empty"}
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
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.DO, "cluster", infraId, "Terminate", token, *ctx)
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
	ctx.SendLogs("DOClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
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

	statusCode, doProfile, err := do.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	var cluster do.Cluster_Def

	ctx.SendLogs("DOClusterController: Getting Cluster of infrastructure. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = do.GetCluster(infraId, userInfo.CompanyId, *ctx)
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
	if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		ctx.SendLogs("DOClusterController: Cluster is not in created state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is not in created state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Deploying) {
		ctx.SendLogs("DOClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Terminating) {
		ctx.SendLogs("DOClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.ClusterTerminated) {
		ctx.SendLogs("DOClusterController: Cluster is in terminated state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminated state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.ClusterCreationFailed) {
		ctx.SendLogs("DOClusterController: Cluster creation is in failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": " Cluster creation is in failed state"}
		c.ServeJSON()
		return
	}

	if cluster.Status == (models.Deploying) {
		ctx.SendLogs("DOClusterController: Cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is in deploying state"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("DOClusterController: Terminating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	cluster.Status = (models.Terminating)
	err = do.UpdateCluster(cluster, false, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	go do.TerminateCluster(cluster, doProfile, *ctx, userInfo.CompanyId, token)

	ctx.SendLogs(" DO cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+" terminated", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster termination is in progress"}
	c.ServeJSON()
}

// @Title SSHKeyPair
// @Description returns ssh key pairs
// @Param X-Auth-Token header string token ""
// @Success 200 {object} []string
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /sshkeys [get]
func (c *DOClusterController) GetSSHKeys() {

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

	//=============================================================================//

	ctx.SendLogs("DOClusterController: FetchExistingSSHKeys.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	keys, err := do.GetAllSSHKeyPair(*ctx, token)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = keys
	c.ServeJSON()
}

// @Title CreateSSHKey
// @Description Generates new SSH key
// @Param infraId path string true "Id of the infrastructure"
// @Param keyname path string true "SSHKey"
// @Param X-Profile-Id header string true "profileId"
// @Param X-Auth-Token header string true "Token"
// @Param teams header string true "teams"
// @Param region path string true "region"
// @Success 200 {object} key_utils.AZUREKey
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 512 {object} types.CustomCPError
// @router /sshkey/:infraId/:keyname/:region [post]
func (c *DOClusterController) PostSSHKey() {

	ctx := new(utils.Context)
	ctx.SendLogs("DOClusterController: CreateSSHKey.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	infraId := c.GetString(":infraId")
	if infraId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "infrastructure id is empty"}
		c.ServeJSON()
		return
	}

	keyName := c.GetString(":keyname")
	if keyName == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "keyname is empty"}
		c.ServeJSON()
		return
	}
	//==========================RBAC Authentication==============================//

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
		c.ServeJSON()
		return
	}

	teams := c.Ctx.Input.Header("teams")

	region := c.GetString(":region")
	if region == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "region id is empty"}
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

	ctx.SendLogs("DOClusterController: PostSSHKey.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	//==========================RBAC Authentication==============================//

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	statusCode, doProfile, err := do.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	//==========================RBAC Authentication==============================//

	keyMaterial, err_ := do.CreateSSHkey(keyName, doProfile.Profile, token, teams, region, *ctx)
	if err_ != (types.CustomCPError{}) {
		if err_.StatusCode != 500 {
			c.Ctx.Output.SetStatus(err_.StatusCode)
			c.Data["json"] = err_
			c.ServeJSON()
		}
		ctx.SendLogs("DOClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" DO cluster key "+keyName+" created in region "+region, models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = keyMaterial
	c.ServeJSON()
}

// @Title GetRegions
// @Description return regions and their supported machine sizes
// @Param X-Profile-Id header string true "X-Profile-Id"
// @Param X-Auth-Token header string true "Token"
// @Success 200 {object} []godo.Region
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime error"}
// @Failure 512 {object} types.CustomCPError
// @router /getallregions/ [get]
func (c *DOClusterController) GetRegions() {

	ctx := new(utils.Context)
	ctx.SendLogs("DOClusterController: GetallZones.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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
	ctx.SendLogs("DOClusterController: GetAllZones.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, doProfile, err := do.GetProfile(profileId, "", token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("DOClusterController: Get Zones. ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	zones, err_ := do.GetRegionsAndCores(doProfile.Profile, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(err_.StatusCode)
		c.Data["json"] = err_
		c.ServeJSON()
		return
	}
	c.Data["json"] = zones
	c.ServeJSON()
}

// @Title DeleteSSHKey
// @Description Delete SSH key
// @Param keyname path string true "keyname"
// @Param X-Profile-Id header string true "profileId"
// @Param X-Auth-Token header string true "Token"
// @Param region path string true "region"
// @Success 200 {"msg": "key deleted successfully"}
// @Failure 409 {"error": "key is in used by some infrastructures"}
// @Failure 404 {"error": "Not Found"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 500 {"error": "Runtime Error"}
// @router /sshkey/:keyname [delete]
func (c *DOClusterController) DeleteSSHKey() {

	ctx := new(utils.Context)
	ctx.SendLogs("DOClusterController: DeleteSSHKey.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token id is empty"}
		c.ServeJSON()
		return
	}

	region := c.Ctx.Input.Header("region")
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

	ctx.SendLogs("DOClusterController: DeleteSSHKey.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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
		c.Data["json"] = map[string]string{"error": "key is used in other infrastructures and can't be deleted"}
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

	statusCode, doProfile, err := do.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	cpErr := do.DeleteSSHkey(keyName, token, doProfile.Profile, *ctx)
	if cpErr != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(cpErr.StatusCode)
		c.Data["json"] = map[string]string{"error": cpErr.Description}
		c.ServeJSON()
		return
	}
	ctx.SendLogs(" DO cluster key "+keyName+"of region"+region+" deleted", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "key deleted successfully"}
	c.ServeJSON()
}

// @Title ValidateProfile
// @Description validate if profile is valid
// @Param X-Auth-Token header string true "token"
// @Param body body vault.DOCredentials true	"body for cluster content"
// @Success 200 {"msg": "Profile is valid"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Invalid Profile"}
// @Failure 404 {"error": "Not Found"}
// @router /validateProfile/ [post]
func (c *DOClusterController) ValidateProfile() {

	var credentials vault.DOCredentials
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &credentials)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "error while unmarshalling " + err.Error()}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)
	ctx.SendLogs("DOClusterController:Check if profile is valid.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("DOClusterController: GetAllZones.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs("DOClusterController: Get Zones. ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err_ := do.ValidateProfile(credentials.AccessKey, *ctx)
	if err_ != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Profile is not valid"}
		c.ServeJSON()
		return
	}
	c.Data["json"] = map[string]string{"msg": "profile is valid"}
	c.ServeJSON()
}
