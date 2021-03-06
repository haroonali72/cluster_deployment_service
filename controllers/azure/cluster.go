package azure

import (
	"antelope/models"
	"antelope/models/azure"
	"antelope/models/cores"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/astaxie/beego"
	"github.com/go-playground/validator/v10"
	"strings"
	"time"
)

// Operations about azure cluster [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/cluster/{cloud}/]
type AzureClusterController struct {
	beego.Controller
}

// @Title Get
// @Description Get cluster against the infraId
// @Param	infraId	path	string	true	"Id of the Infrastructure"
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 200 {object} azure.Cluster_Def
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /:infraId/ [get]
func (c *AzureClusterController) Get() {
	ctx := new(utils.Context)
	ctx.SendLogs("AzureClusterController: Get cluster ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.Token) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	infraId := c.GetString(":infraId")
	if infraId == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.InfraId) + string(models.IsEmpty)}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbac_athentication.Authenticate(models.Azure, "cluster", infraId, "View", token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(int(models.Unauthorized))
		c.Data["json"] = map[string]string{"error": string(models.Notauthorized)}
		c.ServeJSON()
		return
	}

	//==================================================================================//

	ctx.SendLogs("AzureClusterController: Getting cluster with infrastructure Id "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := azure.GetCluster(infraId, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs(" Azure cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+" fetched ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs(" Azure cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+" fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Get All
// @Description get all the clusters
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 200 {object} []azure.AzureCluster
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /all [get]
func (c *AzureClusterController) GetAll() {
	ctx := new(utils.Context)
	ctx.SendLogs("AzureClusterController: Get All Clusters.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.Token) + string(models.IsEmpty)}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, err, data := rbac_athentication.GetAllAuthenticate("cluster", userInfo.CompanyId, token, models.Azure, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	//=============================================================================//

	clusters, err := azure.GetAllCluster(*ctx, data)
	if err != nil {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs(" All Azure clusters fetched", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs(" All Azure clusters fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = clusters
	c.ServeJSON()
}

// @Title Create
// @Description create a new cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	body	body 	azure.Cluster_Def		true	"Body for cluster content"
// @Success 201 {"msg": "Cluster created successfully"}
// @Success 400 {"msg": "Bad Request"}
// @Success 401 {"msg": "Unauthorized"}
// @Success 404 {"msg": "Not Found"}
// @Failure 409 {"error": "Cluster against same infrastructure Id already exists"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [post]
func (c *AzureClusterController) Post() {
	var cluster azure.Cluster_Def
	ctx := new(utils.Context)
	ctx.SendLogs("AzureClusterController: Add Cluster ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.Token) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		ctx.SendLogs("Error in unmarshal "+err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		c.Ctx.Output.SetStatus(int(models.BadRequest))
		c.Data["json"] = map[string]string{"error": "Internal Server Error"}
		c.ServeJSON()
		return
	}

	_, err = govalidator.ValidateStruct(cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(int(models.BadRequest))
		c.Data["json"] = map[string]string{"error": err.Error()}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, cluster.InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbac_athentication.Authenticate(models.Azure, string(models.Cluster), cluster.InfraId, "Create", token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(int(models.Unauthorized))
		c.Data["json"] = map[string]string{"error": string(models.Notauthorized)}
		c.ServeJSON()
		return
	}

	//=============================================================================//

	ctx.SendLogs("AzureClusterController: Adding new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster.CreationDate = time.Now()

	network, err := azure.GetNetwork(cluster.InfraId, *ctx, cluster.ResourceGroup, token)
	if err != nil {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	for _, node := range cluster.NodePools {
		node.EnablePublicIP = !network.IsPrivate

	}

	cluster.CompanyId = userInfo.CompanyId

	err = azure.CreateCluster(cluster, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(int(models.Conflict))
			c.Data["json"] = map[string]string{"error": string(models.AlreadyExist)}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs(" Azure cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+" created ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs(" Azure cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+" created ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = map[string]string{"msg": string(models.SuccessfullyAdded)}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	body	body 	azure.Cluster_Def	true	"Body for cluster content"
// @Success 200 {"msg": "Cluster updated successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Cluster is in Creating/Terminating/TerminationFailed state"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [put]
func (c *AzureClusterController) Patch() {
	ctx := new(utils.Context)
	var cluster azure.Cluster_Def
	ctx.SendLogs("AzureClusterController: Update cluster ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.Token) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(int(models.BadRequest))
		c.Data["json"] = map[string]string{"error": err.Error()}
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

	savedCluster, err := azure.GetCluster(cluster.InfraId, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if savedCluster.Status == (models.Deploying) {
		ctx.SendLogs("AzureClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if savedCluster.Status == (models.Terminating) {
		ctx.SendLogs("AzureClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if savedCluster.Status == (models.ClusterTerminationFailed) {
		ctx.SendLogs("AzureClusterController: Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, cluster.InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbac_athentication.Authenticate(models.Azure, string(models.Cluster), cluster.InfraId, "Update", token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(int(models.Unauthorized))
		c.Data["json"] = map[string]string{"error": string(models.Notauthorized)}
		c.ServeJSON()
		return
	}

	//=============================================================================//

	network, err := azure.GetNetwork(cluster.InfraId, *ctx, cluster.ResourceGroup, token)
	if err != nil {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	for _, node := range cluster.NodePools {
		node.EnablePublicIP = !network.IsPrivate
	}

	ctx.SendLogs("AzureClusterController: Update cluster "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster.CompanyId = userInfo.CompanyId

	err = azure.UpdateCluster(cluster, true, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			c.Ctx.Output.SetStatus(int(models.NotFound))
			c.Data["json"] = map[string]string{"error": "No cluster exists with this name"}
			c.ServeJSON()
			return
		} else if strings.Contains(err.Error(), "No changes are applicable") {
			c.Data["json"] = map[string]string{"msg": string(models.SuccessfullyUpdated)}
			c.ServeJSON()
			return
		} else if strings.Contains(err.Error(), "Cluster is in creating state") {
			c.Ctx.Output.SetStatus(int(models.StateConflict))
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		} else if strings.Contains(err.Error(), "Cluster is in terminating state") {
			c.Ctx.Output.SetStatus(int(models.StateConflict))
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		} else if strings.Contains(err.Error(), "Cluster is in termination failed state") {
			c.Ctx.Output.SetStatus(int(models.StateConflict))
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		}

		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs(" Azure cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs(" Azure cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = map[string]string{"msg": string(models.SuccessfullyUpdated)}
	c.ServeJSON()
}

// @Title Delete
// @Description delete a cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	infraId	path	string	true	"infrastructure Id of the cluster"
// @Param	forceDelete path    boolean	true     ""
// @Success 204 {"msg": "Cluster deleted successfully"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Cluster is in Created/Creating/Terminating/TerminationFailed state"}
// @Failure 500 {"error": "Runtime Error"}
// @router /:infraId/:forceDelete  [delete]
func (c *AzureClusterController) Delete() {
	ctx := new(utils.Context)
	ctx.SendLogs("AzureClusterController: Delete cluster", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

	id := c.GetString(":infraId")
	if id == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.InfraId) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.Token) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	forceDelete, err := c.GetBool(":forceDelete")
	if err != nil {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": err.Error()}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, id, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbac_athentication.Authenticate(models.Azure, string(models.Cluster), id, "Delete", token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(int(models.Unauthorized))
		c.Data["json"] = map[string]string{"error": string(models.Notauthorized)}
		c.ServeJSON()
		return
	}
	//==============================================================================//

	cluster, err := azure.GetCluster(id, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if cluster.Status == models.ClusterCreated && !forceDelete {
		ctx.SendLogs("AzureClusterController: Cluster is in created state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(int(models.StateConflict))
		c.Data["json"] = map[string]string{"error": "Cluster is in created state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Deploying && !forceDelete {
		ctx.SendLogs("AzureClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(int(models.StateConflict))
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Terminating && !forceDelete {
		ctx.SendLogs("AzureClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(int(models.StateConflict))
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.ClusterTerminationFailed && !forceDelete {
		ctx.SendLogs("AzureClusterController: Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(int(models.StateConflict))
		c.Data["json"] = map[string]string{"error": "Cluster is in termination failed state"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AzureClusterController: Delete cluster of Infrastructure "+id, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

	err = azure.DeleteCluster(id, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs(" Azure cluster "+cluster.Name+" of Infrastructure  "+cluster.InfraId+" deleted ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs(" Azure cluster "+cluster.Name+" of Infrastructure  "+cluster.InfraId+" deleted ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(204)
	c.Data["json"] = map[string]string{"msg": string(models.SuccessfullyDeleted)}
	c.ServeJSON()
}

// @Title Create
// @Description Deploy a  cluster
// @Param	infraId	path	string	true	"Id of the Infrastructure"
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Success 201 {"msg": "Cluster created successfully"}
// @Success 202 {"msg": "Cluster creation initiated"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in Created/Creating/Terminating/TerminationFailed state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 512 {object} types.CustomCPError
// @router /start/:infraId [post]
func (c *AzureClusterController) StartCluster() {
	ctx := new(utils.Context)
	ctx.SendLogs("AzureClusterController: Create cluster ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.ProfileId) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.Token) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	infraId := c.GetString(":infraId")
	if infraId == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.InfraId) + string(models.IsEmpty)}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbac_athentication.Authenticate(models.Azure, string(models.Cluster), infraId, "Start", token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(int(models.Unauthorized))
		c.Data["json"] = map[string]string{"error": string(models.Notauthorized)}
		c.ServeJSON()
		return
	}

	//===========================================================================//

	region, err := azure.GetRegion(token, infraId, *ctx)
	if region == "" {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": "region is empty"}
		c.ServeJSON()
		return
	}

	var cluster azure.Cluster_Def

	ctx.SendLogs("AzureClusterController: Getting Cluster of Infrastructure. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = azure.GetCluster(infraId, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if cluster.Status == models.ClusterCreated {
		c.Ctx.Output.SetStatus(int(models.StateConflict))
		c.Data["json"] = map[string]string{"error": "Cluster is already in created state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Deploying {
		ctx.SendLogs("Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(int(models.StateConflict))
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Terminating {
		ctx.SendLogs("Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(int(models.StateConflict))
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.ClusterTerminationFailed {
		ctx.SendLogs("Cluster termination is in failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(int(models.StateConflict))
		c.Data["json"] = map[string]string{"error": "Cluster termination is in failed state"}
		c.ServeJSON()
		return
	}

	statusCode, azureProfile, err := azure.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		utils.SendLog(userInfo.CompanyId, err.Error(), "error", infraId)
		utils.SendLog(userInfo.CompanyId, "Cluster creation failed: "+cluster.Name, "error", cluster.InfraId)
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	cluster.Status = models.Deploying
	err = azure.UpdateCluster(cluster, false, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AzureClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go azure.DeployCluster(cluster, azureProfile, *ctx, userInfo.CompanyId, token)

	ctx.SendLogs(" Azure cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+" created ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs(" Azure cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+" created ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(202)
	c.Data["json"] = map[string]string{"msg": string(models.CreationInitialised)}
	c.ServeJSON()
}

// @Title Status
// @Description Get live status of the cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	infraId	path	string	true	"Id of the Infrastructure"
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Success 200 {object} azure.Cluster_Def
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Cluster is in creating/terminating state"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 512 {object} types.CustomCPError
// @router /status/:infraId/ [get]
func (c *AzureClusterController) GetStatus() {
	ctx := new(utils.Context)
	ctx.SendLogs("AzureClusterController: Fetch Status of cluster ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	infraId := c.GetString(":infraId")
	if infraId == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.InfraId) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.Token) + string(models.IsEmpty)}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.Azure, string(models.Cluster), infraId, "View", token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(int(models.Unauthorized))
		c.Data["json"] = map[string]string{"error": string(models.Notauthorized)}
		c.ServeJSON()
		return
	}

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.ProfileId) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}
	//===========================================================================//

	region, err := azure.GetRegion(token, infraId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	statusCode, azureProfile, err := azure.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AzureClusterController: Fetching cluster status of Infrastructure "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err1 := azure.FetchStatus(azureProfile, token, infraId, userInfo.CompanyId, *ctx)
	if err1 != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(err1.StatusCode)
		c.Data["json"] = err1
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AzureClusterController:Cluster status of Infrastructure "+infraId+" fetched", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs("AzureClusterController:Cluster status of Infrastructure "+infraId+" fetched", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Terminate
// @Description terminates a running cluster
// @Param	infraId	path	string	true	"Id of the Infrastructure"
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 202 {"msg": "Cluster termination initialized"}
// @Success 204 {"msg": "Cluster terminated successfully"}
// @Failure 404 {"error": "Not Found"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 512 {object} types.CustomCPError
// @router /terminate/:infraId/ [post]
func (c *AzureClusterController) TerminateCluster() {
	ctx := new(utils.Context)
	var cluster azure.Cluster_Def

	ctx.SendLogs("AzureClusterController: Terminate Cluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	infraId := c.GetString(":infraId")
	if infraId == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.InfraId) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.ProfileId) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.Token) + string(models.IsEmpty)}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbac_athentication.Authenticate(models.Azure, string(models.Cluster), infraId, "Terminate", token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(int(models.Unauthorized))
		c.Data["json"] = map[string]string{"error": string(models.Notauthorized)}
		c.ServeJSON()
		return
	}

	//============================================================================//

	ctx.SendLogs("AzureClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	region, err := azure.GetRegion(token, infraId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	statusCode, azureProfile, err := azure.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AzureClusterController: Getting Cluster of Infrastructure. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = azure.GetCluster(infraId, userInfo.CompanyId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		ctx.SendLogs("AZUREClusterController: Cluster is not in created state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(int(models.StateConflict))
		c.Data["json"] = map[string]string{"error": " Cluster is not in created state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Deploying {
		ctx.SendLogs("AzureClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(int(models.StateConflict))
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Terminating {
		ctx.SendLogs("AzureClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(int(models.StateConflict))
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.ClusterTerminated {
		ctx.SendLogs("AzureClusterController: Cluster is in terminated state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(int(models.StateConflict))
		c.Data["json"] = map[string]string{"error": "Cluster is in terminated state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.ClusterCreationFailed {
		ctx.SendLogs("AzureClusterController: Cluster creation is in failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(int(models.StateConflict))
		c.Data["json"] = map[string]string{"error": "Cluster creation is in failed state"}
		c.ServeJSON()
		return
	}
	cluster.Status = models.Terminating
	err = azure.UpdateCluster(cluster, false, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("AzureClusterController: Terminating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	go azure.TerminateCluster(cluster, azureProfile, *ctx, userInfo.CompanyId, token)

	ctx.SendLogs(" Azure cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+" terminated", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(202)
	c.Data["json"] = map[string]string{"msg": string(models.TerminationInitialised)}
	c.ServeJSON()
}

// @Title SSHKeyPair
// @Description returns ssh key pairs
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 200 {object} []string
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /sshkeys [get]
func (c *AzureClusterController) GetSSHKeys() {
	ctx := new(utils.Context)
	ctx.SendLogs("AZUREClusterController: Fetch Existing SSHKeys.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.Token) + string(models.IsEmpty)}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	//=============================================================================//
	ctx.SendLogs("AZUREClusterController: Fetching Existing SSHKeys", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	keys, err := azure.GetAllSSHKeyPair(*ctx, token)
	if err != nil {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("AZUREClusterController: Existing SSHKeys fetched", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	c.Data["json"] = keys
	c.ServeJSON()
}

// @Title CreateSSHKey
// @Description Generates new SSH key
// @Param	infraId	path	string	true	"Id of the Infrastructure"
// @Param	keyname	 	path	string	true	"SSHKey"
// @Param	X-Auth-Token		header	string	true 	"Token"
// @Param	teams		header	string	teams 	""
// @Success 200 		{object} key_utils.AZUREKey
// @Failure 404 		{"error": "Not Found"}
// @Failure 500 		{"error": "Runtime Error"}
// @router /sshkey/:keyname/:infraId [post]
func (c *AzureClusterController) PostSSHKey() {

	ctx := new(utils.Context)
	ctx.SendLogs("AzureClusterController: Create SSH Key ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	infraId := c.GetString(":infraId")
	if infraId == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.InfraId) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.Token) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	keyName := c.GetString(":keyname")
	if keyName == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.KeyName) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	teams := c.Ctx.Input.Header("teams")

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	//==========================RBAC Authentication==============================//
	//=============================================================================//

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)

	ctx.SendLogs("AZURENetworkController: Creating SSH Key ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	privateKey, err := azure.CreateSSHkey(keyName, token, teams, *ctx)
	if err != nil {
		ctx.SendLogs("AzureClusterController :"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs(" Azure cluster key "+keyName+" created ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs(" Azure cluster key "+keyName+" created ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = privateKey
	c.ServeJSON()
}

// @Title GetCores
// @Description Get Azure Machine instance cores
// @Success 200 			{object} models.Machine
// @Failure 500 			{"error": "error msg"}
// @router /machine/info [get]
func (c *AzureClusterController) GetCores() {

	var machine []models.Machine
	if err := json.Unmarshal(cores.AzureCores, &machine); err != nil {
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
// @Param	keyname	 	path	string	true	"Unique name of the key"
// @Param	X-Auth-Token		header	string	true 	"Token"
// @Success 204 		{"msg": Key deleted successfully}
// @Failure 401 		{"error": "Unauthorized"}
// @Failure 404 		{"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /sshkey/:keyname [delete]
func (c *AzureClusterController) DeleteSSHKey() {

	ctx := new(utils.Context)
	ctx.SendLogs("AzureClusterController: Delete SSH key ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.Token) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	keyName := c.GetString(":keyname")
	if keyName == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.KeyName) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	//==========================RBAC Authentication==============================//
	//==========================================================================//

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	ctx.SendLogs("AZURENetworkController: Deleting SSH Key "+keyName, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	alreadyUsed := azure.CheckKeyUsage(keyName, userInfo.CompanyId, *ctx)
	if alreadyUsed {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	err = azure.DeleteSSHkey(keyName, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AZURENetworkController:: Key "+keyName+" deleted ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs(" Azure cluster key "+keyName+" deleted ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Ctx.Output.SetStatus(204)
	c.Data["json"] = map[string]string{"msg": string(models.KeySuccessfullyDeleted)}
	c.ServeJSON()
}

// @Title Get Instances
// @Description Get All Instances
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	X-Profile-Id	header	string	true	"profile id"
// @Param	region	path	string	true	"Cloud region"
// @Success 200 []compute.VirtualMachines
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 512 {object} types.CustomCPError
// @router /getAllInstances [get]
func (c *AzureClusterController) GetInstances() {
	ctx := new(utils.Context)
	ctx.SendLogs("AZURENetworkController:Get all instances ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.ProfileId) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.Token) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	region := c.Ctx.Input.Header("region")
	if token == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.RegionV) + string(models.IsEmpty)}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	statusCode, azureProfile, err := azure.GetProfile(profileId, region, token, *ctx)
	beego.Info("AzureClusterController: Get All Instances..")

	statusCode, azureProfile, err = azure.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("AZURENetworkController:: Get all instances ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	instances, err1 := azure.GetInstances(azureProfile, *ctx)
	if err1 != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(err1.StatusCode)
		c.Data["json"] = err1
		c.ServeJSON()
		return
	}
	ctx.SendLogs("AZURENetworkController:All instance fetched ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	c.Data["json"] = instances
	c.ServeJSON()
}

// @Title Get Azure Regions
// @Description Get List of the Azure Regions
// @Param	X-Auth-Token	header	string	 true "Token"
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Success 200 []model.Region
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /getallregions [get]
func (c *AzureClusterController) GetRegions() {
	ctx := new(utils.Context)
	ctx.SendLogs("AZURENetworkController: Get all regions ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.Token) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}
	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if token == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.ProfileId) + string(models.IsEmpty)}
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
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	var regions []models.Region
	if err := json.Unmarshal(cores.AzureRegions, &regions); err != nil {
		beego.Error("Unmarshalling of regions failed ", err.Error())
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	statusCode, azureProfile, err := azure.GetProfile(profileId, "useast", token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AzureClusterController:: Getting all instances ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	reg, err1 := azure.GetRegions(azureProfile, *ctx)
	if err1 != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err1.Error}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("AzureClusterController:: All instances fetched ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs("AzureClusterController:: All instances fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = reg
	c.ServeJSON()
}

// @Title Get Availability Zone
// @Description return zones against a region
// @Param	region	 	path	string	true	"Azure region"
// @Success 200  []string
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 			{"error": "Bad Request"}
// @router /getzones/:region [get]
func (c *AzureClusterController) GetZones() {
	ctx := new(utils.Context)

	region := c.GetString(":region")
	if region == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "region is empty"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AWSClusterController: fetch availability zones.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	az, err1 := azure.GetZones(region)
	if err1 != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(err1.StatusCode)
		c.Data["json"] = err1
		c.ServeJSON()
		return
	}

	ctx.SendLogs("Zones fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = az
	c.ServeJSON()
}

// @Title Get VM Sizes
// @Description Get list of azure VM sizes
// @Success 200 []string
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /getallmachines [get]
func (c *AzureClusterController) GetAllMachines() {

	instances, err1 := azure.GetAllMachines()
	if err1 != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(int(models.InternalServerError))
		c.Data["json"] = map[string]string{"error": err1.Error}
		c.ServeJSON()
		return
	}
	c.Data["json"] = instances
	c.ServeJSON()
}

// @Title Validate Profile
// @Description Check if profile is valid
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	body	body 	vault.AzureCredentials		true	"Body for cluster content"
// @Success 200 {"msg": "Profile is valid"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 512 {object} types.CustomCPError
// @router /validateProfile/ [post]
func (c *AzureClusterController) ValidateProfile() {

	ctx := new(utils.Context)
	ctx.SendLogs("AzureClusterController:: Validate Profile ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	var credentials vault.AzureCredentials
	json.Unmarshal(c.Ctx.Input.RequestBody, &credentials)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.Token) + string(models.IsEmpty)}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	var regions []models.Region
	if err := json.Unmarshal(cores.AzureRegions, &regions); err != nil {
		beego.Error("Unmarshalling of machine instances failed ", err.Error())
		c.Ctx.Output.SetStatus(int(models.BadRequest))
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("Checking Profile Validity", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	for _, region := range regions {
		err1 := azure.ValidateProfile(credentials.ClientId, credentials.ClientSecret, credentials.SubscriptionId, credentials.TenantId, region.Location, *ctx)
		if err1 != (types.CustomCPError{}) {
			ctx.SendLogs("AzureClusterController: Profile not valid", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			c.Ctx.Output.SetStatus(err1.StatusCode)
			c.Data["json"] = err1
			c.ServeJSON()
			return
		}
		if err1 == (types.CustomCPError{}) {
			break
		}
	}
	ctx.SendLogs("Profile Validated", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs("Profile Validated", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = map[string]string{"msg": string(models.ValidProfile)}
	c.ServeJSON()
}

// @Title Start
// @Description Apply cloudplex Agent file to a aks cluster
// @Param	clusterName	header	string	true "Cluster Name"
// @Param	resourceGroup	header	string	true "ResourceGroup"
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 200 {"msg": "Agent Applied successfully"}
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Param	infraId	path	string	true	"Id of the Infrastructure"
// @Failure 404 {"error": "Not Found"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 500 {"error": "Runtime Error"}
// @router /applyagent/:infraId [post]
func (c *AzureClusterController) ApplyAgent() {

	ctx := new(utils.Context)
	ctx.SendLogs("AzureClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("AzureClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.ProfileId) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	infraId := c.GetString(":infraId")
	if infraId == "" {
		ctx.SendLogs("AzureClusterController: InfraId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.InfraId) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.Token) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	clusterName := c.Ctx.Input.Header("clusterName")
	if clusterName == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.ClusterName) + string(models.IsEmpty)}
		c.ServeJSON()
		return
	}

	resourceGroup := c.Ctx.Input.Header("resourceGroup")
	if resourceGroup == "" {
		c.Ctx.Output.SetStatus(int(models.ParamMissing))
		c.Data["json"] = map[string]string{"error": string(models.ResourceGroup) + string(models.IsEmpty)}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("AzureClusterController: Apply Agent.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, allowed, err := rbac_athentication.Authenticate(models.Azure, string(models.Cluster), infraId, "Start", token, utils.Context{})
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(int(models.Unauthorized))
		c.Data["json"] = map[string]string{"error": string(models.Notauthorized)}
		c.ServeJSON()
		return
	}

	statusCode, azureProfile, err := azure.GetProfile(profileId, "", token, *ctx)
	if err != nil {
		utils.SendLog(userInfo.CompanyId, err.Error(), "error", infraId)
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AKSClusterController: applying agent on cluster . "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go azure.ApplyAgent(azureProfile, token, *ctx, clusterName, resourceGroup)

	c.Data["json"] = map[string]string{"msg": "Agent deployment in progress"}
	c.ServeJSON()
}
