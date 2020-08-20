package eks

import (
	"antelope/models"
	"antelope/models/aws"
	"antelope/models/eks"
	"antelope/models/iks"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/types"
	"antelope/models/utils"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/go-playground/validator/v10"
	"strings"
)

// Operations about EKS cluster [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/cluster/{cloud}/]
type EKSClusterController struct {
	beego.Controller
}

// @Title Get
// @Description Get cluster against the infraId
// @Param	infraId	path	string	true	"Id of the infrastructure"
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 200 {object} eks.EKSCluster
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /:infraId/ [get]
func (c *EKSClusterController) Get() {
	ctx := new(utils.Context)

	infraId := c.GetString(":infraId")
	if infraId == "" {
		ctx.SendLogs("EKSClusterController: infraId is empty", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("EKSClusterController: Get cluster with infrastructure id "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.EKS, "cluster", infraId, "View", token, utils.Context{})
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this infrastructure id"}
			c.ServeJSON()
			return
		}
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

	ctx.SendLogs("EKSClusterController: Get cluster with infrastructure id: "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := eks.GetEKSCluster(infraId, userInfo.CompanyId, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		}
		ctx.SendLogs("EKSGetClusterController: error getting eks cluster "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs(" EKS cluster "+cluster.Name+" of infrastructure id: "+cluster.InfraId+" fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Get All
// @Description get all the clusters
// @Param	X-Auth-Token	header	string	true "token"
// @Success 200 {object} []eks.EKSCluster
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /all [get]
func (c *EKSClusterController) GetAll() {
	ctx := new(utils.Context)
	ctx.SendLogs("EKSClusterController: GetAll clusters.", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	ctx.SendLogs("EKSClusterController: Getting all clusters ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, err, data := rbacAuthentication.GetAllAuthenticate("cluster", userInfo.CompanyId, token, models.EKS, *ctx)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this infrastructure id"}
			c.ServeJSON()
			return
		}
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	clusters, err := eks.GetAllEKSCluster(data, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		}
		ctx.SendLogs("EKSClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("All EKS cluster fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = clusters
	c.ServeJSON()
}

// @Title Create
// @Description add a new cluster
// @Param	body body eks.EKSCluster true "body for cluster content"
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 201 {"msg": "Cluster created successfully"}
// @Success 400 {"msg": "Runtime Error"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Cluster against this infrastructure already exists"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [post]
func (c *EKSClusterController) Post() {

	var cluster eks.EKSCluster

	ctx := new(utils.Context)

	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		ctx.SendLogs("EKSClusterController: "+err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
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

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, cluster.InfraId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("EKSClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.EKS, "cluster", cluster.InfraId, "Create", token, utils.Context{})
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this infrastructure id"}
			c.ServeJSON()
			return
		}
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

	ctx.SendLogs("EKSClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	beego.Info("EKSClusterController: JSON Payload: ", cluster)

	validate := validator.New()
	err = validate.Struct(cluster)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	err = eks.ValidateEKSData(cluster, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController: "+err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	cluster.CompanyId = userInfo.CompanyId
	err = eks.GetNetwork(token, cluster.InfraId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	err = eks.AddEKSCluster(cluster, *ctx)
	if err != nil {
		ctx.SendLogs("EKSClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "cluster against same infrastructure id already exists"}
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

	ctx.SendLogs("EKS cluster "+cluster.Name+" of infrastructure id: "+cluster.InfraId+" created ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster added successfully"}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	body	body 	aks.AKSCluster	true	"Body for cluster content"
// @Success 200 {"msg": "Cluster updated successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in Creating/Created/Terminating/TerminationFailed state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [put]
func (c *EKSClusterController) Patch() {
	ctx := new(utils.Context)

	var cluster eks.EKSCluster
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		ctx.SendLogs("EKSClusterController: "+err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
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
	if cluster.Status == (models.Deploying) {
		ctx.SendLogs("EKSClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Terminating) {
		ctx.SendLogs("EKSClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	}
	validate := validator.New()
	err = validate.Struct(cluster)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	err = eks.ValidateEKSData(cluster, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController: "+err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, cluster.InfraId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("EKSClusterController: update cluster cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.EKS, "cluster", cluster.InfraId, "Update", token, utils.Context{})
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this infrastructure id"}
			c.ServeJSON()
			return
		}
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
	ctx.SendLogs("EKSClusterController: Patch cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	beego.Info("EKSClusterController: JSON Payload: ", cluster)
	if cluster.Status == (models.ClusterCreated) || cluster.Status == (models.ClusterTerminationFailed) {
		err := eks.UpdatePreviousEKSCluster(cluster, *ctx)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.Ctx.Output.SetStatus(404)
				c.Data["json"] = map[string]string{"error": err.Error()}
				c.ServeJSON()
				return
			}
			if strings.Contains(err.Error(), "does not exist") {
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

		ctx.SendLogs("EKS running cluster "+cluster.Name+" in infrastructure id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

		c.Data["json"] = map[string]string{"msg": "Running cluster updated successfully"}
		c.ServeJSON()
	} else if cluster.Status == (models.ClusterUpdateFailed) {
		err := eks.UpdateEKSCluster(cluster, *ctx)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.Ctx.Output.SetStatus(404)
				c.Data["json"] = map[string]string{"error": err.Error()}
				c.ServeJSON()
				return
			}
			if strings.Contains(err.Error(), "does not exist") {
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

		ctx.SendLogs("EKS running cluster "+cluster.Name+" in infrastructure id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

		c.Data["json"] = map[string]string{"msg": "Running cluster updated successfully"}
		c.ServeJSON()
	}

	cluster.CompanyId = userInfo.CompanyId
	err = eks.UpdateEKSCluster(cluster, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "not found") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "no cluster exists with this name"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("EKS cluster "+cluster.Name+" of infrastructure id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster updated successfully"}
	c.ServeJSON()
}

// @Title Delete
// @Description Delete a cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	infraId	path 	string	true	"infrastructure id of the cluster"
// @Param	forceDelete path    boolean	true    "Forcefully delete cluster"
// @Success 204 {"msg": "Cluster deleted successfully"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in Cluster Created/Creating/Terminating/Termination Failed state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /:infraId/:forceDelete [delete]
func (c *EKSClusterController) Delete() {
	ctx := new(utils.Context)

	id := c.GetString(":infraId")
	if id == "" {
		ctx.SendLogs("AKSClusterController: InfraId field is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this infrastructure id"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, id, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("EKSClusterController: Delete cluster with id "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.EKS, "cluster", id, "Delete", token, utils.Context{})
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this infrastructure id"}
			c.ServeJSON()
			return
		}
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

	ctx.SendLogs("EKSClusterController: Delete cluster with infrastructure id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := eks.GetEKSCluster(id, userInfo.CompanyId, *ctx)
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
	if strings.ToLower(string(cluster.Status)) == string(models.ClusterCreated) && !forceDelete {
		ctx.SendLogs("AKSClusterController: Cluster is in running state ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in created state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Deploying) && !forceDelete {
		ctx.SendLogs("AKSClusterController: Cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Terminating) && !forceDelete {
		ctx.SendLogs("AKSClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.ClusterTerminationFailed) && !forceDelete {
		ctx.SendLogs("AKSClusterController: Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster creation is in termination failed state"}
		c.ServeJSON()
		return
	}

	err = eks.DeleteEKSCluster(id, userInfo.CompanyId, *ctx)
	if err != nil {
		ctx.SendLogs("EKSClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("EKS cluster "+cluster.Name+" of infrastructure id: "+cluster.InfraId+" deleted ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster deleted successfully"}
	c.ServeJSON()
}

// @Title Start
// @Description Deploy a kubernetes cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	X-Profile-Id	header	string	true "Vault credentials profile id"
// @Param	infraId	path	string	true	"Id of the infrastructure"
// @Success 201 {"msg": "Cluster created successfully"}
// @Success 202 {"msg": "Cluster creation initiated"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in Created/Creating/Terminating/TerminationFailed state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /start/:infraId [post]
func (c *EKSClusterController) StartCluster() {
	ctx := new(utils.Context)
	ctx.SendLogs("EKSClusterController: StartCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("EKSClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	infraId := c.GetString(":infraId")
	if infraId == "" {
		ctx.SendLogs("EKSClusterController: InfraId field is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.EKS, "cluster", infraId, "Start", token, utils.Context{})
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this infrastructure id"}
			c.ServeJSON()
			return
		}
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

	region, err := aws.GetRegion(token, infraId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	statusCode, awsProfile, err := aws.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		ctx.SendLogs("EKSClusterController : Unable to get profile", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("EKSClusterController: Getting Cluster of infrastructure. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := eks.GetEKSCluster(infraId, userInfo.CompanyId, *ctx)
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
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is already in created state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Deploying) {
		ctx.SendLogs("cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Terminating) {
		ctx.SendLogs("Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.ClusterTerminationFailed) {
		ctx.SendLogs("Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in termination failed state"}
		c.ServeJSON()
		return
	}
	/*	cluster.Status = string(models.Deploying)
		err = eks.UpdateEKSCluster(cluster, *ctx)
		if err != nil {
			c.Ctx.Output.SetStatus(500)
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		}*/
	ctx.SendLogs("EKSClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go eks.DeployEKSCluster(cluster, awsProfile, userInfo.CompanyId, token, *ctx)

	ctx.SendLogs(" EKS cluster "+cluster.Name+" of infrastructure id: "+cluster.InfraId+" deployed ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Ctx.Output.SetStatus(202)
	c.Data["json"] = map[string]string{"msg": "cluster creation in progress"}
	c.ServeJSON()
}

// @Title Terminate
// @Description Terminate a running cluster
// @Param	X-Profile-Id header	X-Profile-Id	string	true "Vault credentials profile Id"
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	infraId	path	string	true	"Id of the infrastructure"
// @Success 202 {"msg": "Cluster termination initiated"}
// @Success 204 {"msg": "Cluster terminated successfully"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in New/Creating/Creation Failed /Terminated/Terminating state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /terminate/:infraId/ [post]
func (c *EKSClusterController) TerminateCluster() {
	ctx := new(utils.Context)
	ctx.SendLogs("EKSClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("EKSClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	infraId := c.GetString(":infraId")
	if infraId == "" {
		ctx.SendLogs("EKSClusterController: InfraId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("EKSClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.EKS, "cluster", infraId, "Terminate", token, utils.Context{})
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this infrastructure id"}
			c.ServeJSON()
			return
		}
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

	region, err := aws.GetRegion(token, infraId, *ctx)
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

	ctx.SendLogs("EKSClusterController: Getting Cluster of infrastructure. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := eks.GetEKSCluster(infraId, userInfo.CompanyId, *ctx)
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
	if string(cluster.Status) == strings.ToLower(string(models.New)) {
		ctx.SendLogs("EKSClusterController: Cluster is not in created state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is not in created state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Deploying) {
		ctx.SendLogs("EKSClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Terminating) {
		ctx.SendLogs("EKSClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.ClusterTerminated) {
		ctx.SendLogs("EKSClusterController: Cluster is in terminated state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminated state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.ClusterCreationFailed) {
		ctx.SendLogs("EKSClusterController: Cluster  creation is in failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster  creation is in failed state"}
		c.ServeJSON()
		return
	}

	go eks.TerminateCluster(cluster, awsProfile, infraId, userInfo.CompanyId, *ctx)

	c.Ctx.Output.SetStatus(202)
	c.Data["json"] = map[string]string{"msg": "Cluster termination initiated"}
	c.ServeJSON()
}

// @Title Get
// @Description Get Kubernetes Cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 200 {object} []string
// @Failure 404 {"error": "Not Found"}
// @router /kube/versions/:infraId/ [get]
func (c *EKSClusterController) GetkubeVersions() {
	ctx := new(utils.Context)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	versions := eks.KubeVersions(*ctx)
	c.Data["json"] = versions
	c.ServeJSON()
}

// @Title Get
// @Description Get AMIs for EKS Cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 200 {object} eks.AMI
// @Failure 404 {"error": "Not Found"}
// @router /ami/:infraId/ [get]
func (c *EKSClusterController) GetAMI() {
	ctx := new(utils.Context)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	amis := eks.GetAMIS()
	c.Data["json"] = amis
	c.ServeJSON()
}

// @Title Get
// @Description Get Instances for EKS Cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 200 {object} map[string][]string
// @Failure 404 {"error": "Not Found"}
// @router /instances/:amiType/:infraId/ [get]
func (c *EKSClusterController) GetInstances() {
	ctx := new(utils.Context)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
		c.ServeJSON()
		return
	}
	amiType := c.GetString(":amiType")
	if amiType == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "amiType is empty"}
		c.ServeJSON()
		return
	}
	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	instances := eks.GetInstances(amiType, *ctx)
	c.Data["json"] = instances
	c.ServeJSON()
}

// @Title Get
// @Description Get list of running EKS Cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	X-Profile-Id	header	string	true "Token"
// @Success 200 {object} []*string
// @Failure 404 {"error": "Not Found"}
// @router /clusters/:region/:infraId/ [get]
func (c *EKSClusterController) GetClusters() {
	ctx := new(utils.Context)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
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
	region := c.GetString(":region")
	if region == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "region is empty"}
		c.ServeJSON()
		return
	}
	infraId := c.GetString(":infraId")
	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)
	statusCode, awsProfile, err := aws.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	instances, cpErr := eks.GetEKSClusters(infraId, awsProfile, *ctx)
	if cpErr != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(int(models.CloudStatusCode))
		c.Data["json"] = cpErr
		c.ServeJSON()
	}
	c.Data["json"] = instances
	c.ServeJSON()
}

// @Title Status
// @Description Get live status of the running cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile Id"
// @Param	infraId	path	string	true	"Id of the infrastructure"
// @Success 200 {object} eks.EKSClusterStatus
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Cluster is in deploying/terminating state"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 512 {object} types.CustomCPError
// @router /status/:infraId/ [get]
func (c *EKSClusterController) GetStatus() {
	ctx := new(utils.Context)
	ctx.SendLogs("EKClusterController: Fetch Status.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.EKS, "cluster", infraId, "View", token, *ctx)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this infrastructure id"}
			c.ServeJSON()
			return
		}
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

	ctx.SendLogs("EKSClusterController: Fetch Cluster Status of infrastructure. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	region, err := iks.GetRegion(token, infraId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	statusCode, eksProfile, err := aws.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("EKSClusterController: Fetching Status of infrastructure"+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, cpErr := eks.FetchStatus(eksProfile, infraId, *ctx, userInfo.CompanyId, token)
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
	ctx.SendLogs("EKSClusterController: Status fetched of infrastructure"+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs("EKSClusterController: Status fetched of infrastructure"+infraId, models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Start Agent
// @Description Apply cloudplex Agent file to eks cluster
// @Param	clusterName	header	string	true "clusterName"
// @Param	X-Auth-Token	header	string	true "token"
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	infraId	path	string	true	"Id of the infrastructure"
// @Success 200 {"msg": "Agent Applied successfully"}
// @Failure 404 {"error": "Not Found"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 500 {"error": "Runtime Error"}
// @router /applyagent/:infraId [post]
func (c *EKSClusterController) ApplyAgent() {
	ctx := new(utils.Context)
	ctx.SendLogs("EKSClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	infraId := c.GetString(":infraId")
	if infraId == "" {
		ctx.SendLogs("EKSClusterController: InfraId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	clusterName := c.Ctx.Input.Header("clusterName")
	if clusterName == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "clusterName is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("EKSClusterController: Apply Agent.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", infraId, "Start", token, utils.Context{})
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
	region, err := aws.GetRegion(token, infraId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	statusCode, awsProfile, err := aws.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		utils.SendLog(userInfo.CompanyId, err.Error(), "error", infraId)
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("EKSClusterController: applying agent on cluster . "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go eks.ApplyAgent(awsProfile, token, *ctx, clusterName)

	c.Data["json"] = map[string]string{"msg": "agent deployment in progress"}
	c.ServeJSON()
}

// @Title Update
// @Description Update a running kubernetes cluster
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	infraId	path	string	true	"Id of the infrastructure"
// @Success 201 {"msg": "Running cluster updated successfully"}
// @Success 202 {"msg": "Running cluster updation initiated"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in New/Creating/Creation Failed/Terminating/Terminated state"}
// @Failure 404 {"error": "Not found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /update/:infraId [put]
func (c *EKSClusterController) PatchRunningCluster() {

	ctx := new(utils.Context)

	ctx.SendLogs("EKSClusterController: Update running cluster ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, infraId, ctx.Data.Company, userInfo.UserId)

	ctx.SendLogs("EKSClusterController: Updating cluster of infrastructure. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.Data.Company = userInfo.CompanyId
	ctx.Data.InfraId = infraId

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.EKS, "cluster", infraId, "Start", token, utils.Context{})
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this infrastructure id"}
			c.ServeJSON()
			return
		}
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

	region, err := aws.GetRegion(token, infraId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	statusCode, eksProfile, err := aws.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	cluster, err := eks.GetEKSCluster(infraId, userInfo.CompanyId, *ctx)
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
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is new state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.ClusterCreationFailed {
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is cluster creation failed state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Deploying) {
		ctx.SendLogs("GKEClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Terminating) {
		ctx.SendLogs("GKEClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.ClusterTerminated) {
		ctx.SendLogs("GKEClusterController: Cluster is in terminated state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminated state"}
		c.ServeJSON()
		return
	}

	go eks.PatchRunningEKSCluster(cluster, eksProfile.Profile, token, *ctx)

	ctx.SendLogs("GKEClusterController: Running cluster "+cluster.Name+" of infrastructure id: "+cluster.InfraId+"updated", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs(" GKE running cluster "+cluster.Name+" of infrastructure id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(202)
	c.Data["json"] = map[string]string{"msg": "Running cluster update initiated"}
	c.ServeJSON()
}
