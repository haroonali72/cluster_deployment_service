package gke

import (
	"antelope/models"
	"antelope/models/gcp"
	"antelope/models/gke"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/types"
	"antelope/models/utils"
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/astaxie/beego"
	"github.com/go-playground/validator/v10"
	"strings"
)

// Operations about GKE cluster [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/cluster/{cloud}/]
type GKEClusterController struct {
	beego.Controller
}

// @Title Get Options
// @Description Get cluster versions and image sizes
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	region	path	string	true	"Region of the cloud"
// @Success 200 {object} gke.ServerConfig
// @Failure 404 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 500 {"error": "Runtime error"}
// @Failure 512 {object} types.CustomCPError
// @router /config/:region [get]
func (c *GKEClusterController) GetServerConfig() {
	ctx := new(utils.Context)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	zone := c.GetString(":region")
	if zone == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "region is empty"}
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

	isValid, credentials := gcp.IsValidGcpCredentials(profileId, "", token, zone, *ctx)
	if !isValid {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	config, err := gke.GetServerConfig(credentials, *ctx)
	if err != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(int(models.CloudStatusCode))
		c.Data["json"] = err
		c.ServeJSON()
		return
	}

	c.Data["json"] = config
	c.ServeJSON()
}

// @Title Get
// @Description Get cluster against the infraId
// @Param	infraId	path	string	true	"Id of the infrastructure"
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 200 {object} gke.GKECluster
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /:infraId/ [get]
func (c *GKEClusterController) Get() {
	ctx := new(utils.Context)

	ctx.SendLogs("GKEClusterController: Get cluster", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	ctx.Data.Company = userInfo.CompanyId
	ctx.Data.InfraId = infraId
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, infraId, ctx.Data.Company, userInfo.UserId)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", infraId, "View", token, utils.Context{})
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

	ctx.SendLogs("GKEClusterController: Getting cluster of infrastructure "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := gke.GetGKECluster(*ctx)
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

	ctx.SendLogs("GKEClusterController: Cluster of infrastructure "+infraId+" fetched", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs(" GKE cluster "+cluster.Name+" of infrastructure "+cluster.InfraId+" fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Get All
// @Description get all the clusters
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 200 {object} []gke.Cluster
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /all [get]
func (c *GKEClusterController) GetAll() {

	ctx := new(utils.Context)

	ctx.SendLogs("GKEClusterController: Get all clusters", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", ctx.Data.Company, userInfo.UserId)

	statusCode, err, data := rbacAuthentication.GetAllAuthenticate("cluster", ctx.Data.Company, token, models.GKE, *ctx)
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

	ctx.SendLogs("GKEClusterController: Getting all clusters ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.Data.Company = userInfo.CompanyId
	clusters, err := gke.GetAllGKECluster(data, *ctx)
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

	ctx.SendLogs("GKEClusterController: All cluster fetched ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs("All GKE cluster fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = clusters
	c.ServeJSON()
}

// @Title Create
// @Description add a new cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	body	body 	gke.GKECluster		true	"Body for cluster content"
// @Success 201 {"msg": "Cluster created successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Cluster against same infrastructure already exists"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [post]
func (c *GKEClusterController) Post() {

	var cluster gke.GKECluster
	ctx := new(utils.Context)

	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("GKEClusterController: Add cluster", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
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

	err = gke.Validate(cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "error while unmarshalling " + err.Error()}
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

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, cluster.InfraId, ctx.Data.Company, userInfo.UserId)

	ctx.Data.Company = userInfo.CompanyId

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", cluster.InfraId, "Create", token, utils.Context{})
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

	beego.Info("GKEClusterController: JSON Payload: ", cluster)

	cluster.CompanyId = ctx.Data.Company
	err = gke.GetNetwork(token, cluster.InfraId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("GKEClusterController: Adding new cluster with name: "+cluster.Name+" in infrastructure "+cluster.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err = gke.AddGKECluster(cluster, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "Cluster against same infrastructure id already exists"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("GKEClusterController: New cluster with name: "+cluster.Name+" added in infrastructure "+cluster.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs("GKE cluster "+cluster.Name+" add in infrastructure "+cluster.InfraId, models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = map[string]string{"msg": "Cluster added successfully"}
	c.ServeJSON()
}

// @Title Update
// @Description Update an existing cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	body	body 	gke.GKECluster	true	"Body for cluster content"
// @Success 200 {"msg": "Cluster updated successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in Creating/Terminating/Termination Failed state"}
// @Failure 404 {"error": "Not found"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [put]
func (c *GKEClusterController) Patch() {
	ctx := new(utils.Context)
	ctx.SendLogs("GKEClusterController: Update cluster", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	var cluster gke.GKECluster
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

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if cluster.CloudplexStatus == (models.Deploying) {
		ctx.SendLogs("GKEClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.CloudplexStatus == (models.Terminating) {
		ctx.SendLogs("GKEClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
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

	err = gke.Validate(cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "error while unmarshalling " + err.Error()}
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
	ctx.Data.Company = userInfo.CompanyId
	ctx.Data.InfraId = cluster.InfraId

	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, cluster.InfraId, ctx.Data.Company, userInfo.UserId)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", cluster.InfraId, "Update", token, utils.Context{})
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

	ctx.SendLogs("GKEClusterController: Updating cluster "+cluster.Name+" of infrastructure id "+cluster.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	if cluster.CloudplexStatus == (models.ClusterCreated) || cluster.CloudplexStatus == (models.ClusterTerminationFailed) {
		err := gke.UpdatePreviousGKECluster(cluster, *ctx)
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

		ctx.SendLogs("GKE running cluster "+cluster.Name+" in infrastructure Id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

		c.Data["json"] = map[string]string{"msg": "Running cluster updated successfully"}
		c.ServeJSON()
	} else if cluster.CloudplexStatus == (models.ClusterUpdateFailed) {
		err := gke.UpdateGKECluster(cluster, *ctx)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
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

		ctx.SendLogs("GKEClusterController: Cluster "+cluster.Name+" of the infrastructure "+cluster.InfraId+" updated", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		ctx.SendLogs("GKE cluster "+cluster.Name+" of the infrastructure "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

		c.Data["json"] = map[string]string{"msg": "Cluster updated successfully"}
		c.ServeJSON()
	}

	err = gke.UpdateGKECluster(cluster, *ctx)
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

	ctx.SendLogs("GKEClusterController: Cluster "+cluster.Name+" updated of infrastructure id "+cluster.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs("GKE cluster "+cluster.Name+" in infrastructure Id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = map[string]string{"msg": "cluster updated successfully"}
	c.ServeJSON()
}

// @Title Delete
// @Description Delete a cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	infraId	path	string	true	"infrastructure id of the cluster"
// @Param	forceDelete path  boolean	true "Forcefully delete cluster"
// @Success 204 {"msg": "Cluster deleted successfully"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in deploying/running/terminating state"}
// @Failure 404 {"error": "Not found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 512 {object} types.CustomCPError
// @router /:infraId/:forceDelete  [delete]
func (c *GKEClusterController) Delete() {
	ctx := new(utils.Context)

	ctx.SendLogs("GKEClusterController: Delete cluster ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, id, ctx.Data.Company, userInfo.UserId)

	ctx.Data.Company = userInfo.CompanyId
	ctx.Data.InfraId = id
	statusCode, allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", id, "Delete", token, utils.Context{})
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

	cluster, err := gke.GetGKECluster(*ctx)
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

	if strings.ToLower(string(cluster.CloudplexStatus)) == string(string(models.ClusterCreated)) && !forceDelete {
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in running state"}
		c.ServeJSON()
		return
	} else if cluster.CloudplexStatus == models.Deploying && !forceDelete {
		ctx.SendLogs("GKEClusterController: Cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "cluster is in deploying state"}
		c.ServeJSON()
		return
	} else if cluster.CloudplexStatus == (models.Terminating) && !forceDelete {
		ctx.SendLogs("GKEClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.CloudplexStatus == (models.ClusterTerminationFailed) && !forceDelete {
		ctx.SendLogs("GKEClusterController: Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": " Cluster creation is in termination failed state"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("GKEClusterController: Deleting cluster"+cluster.Name+"of infrastructure "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	err = gke.DeleteGKECluster(*ctx)
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

	ctx.SendLogs("GKEClusterController: Cluster "+cluster.Name+" of infrastructure "+id+" deleted", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs("GKE cluster "+cluster.Name+" of infrastructure Id: "+id+" deleted ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(204)
	c.Data["json"] = map[string]string{"msg": "cluster deleted successfully"}
	c.ServeJSON()
}

// @Title Start
// @Description Deploy a kubernetes cluster
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	infraId	path	string	true	"Id of the infrastructure"
// @Success 201 {"msg": "Cluster created initiated"}
// @Success 202 {"msg": "Cluster creation started successfully"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in Created/Creating/Terminating/TerminationFailed state"}
// @Failure 404 {"error": "Not found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /start/:infraId [post]
func (c *GKEClusterController) StartCluster() {
	ctx := new(utils.Context)
	ctx.SendLogs("GKEClusterController: Start cluster ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
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

	ctx.SendLogs("GKEClusterController: Strat cluster of infrastructure. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.Data.Company = userInfo.CompanyId
	ctx.Data.InfraId = infraId
	statusCode, allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", infraId, "Start", token, utils.Context{})
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
	ctx.Data.InfraId = infraId
	region, zone, err := gcp.GetRegion(token, infraId, *ctx)
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

	cluster, err := gke.GetGKECluster(*ctx)
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

	if cluster.CloudplexStatus == models.ClusterCreated {
		ctx.SendLogs("GKEClusterController : Cluster is already running", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is already in running state"}
		c.ServeJSON()
		return
	} else if cluster.CloudplexStatus == (models.Deploying) {
		ctx.SendLogs("GKEClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.CloudplexStatus == (models.Terminating) {
		ctx.SendLogs("GKEClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.CloudplexStatus == (models.ClusterTerminationFailed) {
		ctx.SendLogs("GKEClusterController: Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in termination failed state"}
		c.ServeJSON()
		return
	}

	/*cluster.CloudplexStatus = (models.Deploying)
	err = gke.UpdateGKECluster(cluster, *ctx)
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
	}*/

	go gke.DeployGKECluster(cluster, credentials, token, *ctx)

	ctx.SendLogs("GKEClusterController: Cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+"started", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs(" GKE cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+" deployed ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(202)
	c.Data["json"] = map[string]string{"msg": "Cluster creation initiated"}
	c.ServeJSON()
}

// @Title Status
// @Description Get live status of the running cluster
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	infraId	path	string	true	"Id of the infrastructure"
// @Success 200 {object} gke.KubeClusterStatus
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not found"}
// @Failure 409 {"error": "Cluster is in deploying/terminating state"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 512 {object} types.CustomCPError
// @router /status/:infraId/ [get]
func (c *GKEClusterController) GetStatus() {
	ctx := new(utils.Context)
	ctx.SendLogs("GKEClusterController: FetchStatus.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, infraId, ctx.Data.Company, userInfo.UserId)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", infraId, "View", token, utils.Context{})
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
	ctx.Data.InfraId = infraId
	region, zone, err := gcp.GetRegion(token, infraId, *ctx)
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

	ctx.Data.Company = userInfo.CompanyId

	ctx.SendLogs("GKEClusterController: Fetching status of cluster of the infrastructure  "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	cluster, cpErr := gke.FetchStatus(credentials, token, *ctx)
	if cpErr != (types.CustomCPError{}) && strings.Contains(strings.ToLower(cpErr.Description), "state") || cpErr != (types.CustomCPError{}) && strings.Contains(strings.ToLower(cpErr.Description), "not deployed") {
		c.Ctx.Output.SetStatus(cpErr.StatusCode)
		c.Data["json"] = cpErr
		c.ServeJSON()
		return
	} else if cpErr != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(int(models.CloudStatusCode))
		c.Data["json"] = cpErr
		c.ServeJSON()
	}
	ctx.SendLogs("GKEClusterController: Status Fetched of "+cluster.Name+" of the infrastructure "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Terminate
// @Description Terminate a running cluster
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	infraId	path	string	true	"Id of the infrastructure"
// @Success 202 {"msg": "Cluster termination initiated"}
// @Success 204 {"msg": "Cluster terminated successfully"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in New/Creating/Creation Failed /Terminated/Terminating state"}
// @Failure 404 {"error": "Not found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /terminate/:infraId/ [post]
func (c *GKEClusterController) TerminateCluster() {

	ctx := new(utils.Context)
	ctx.SendLogs("GKEClusterController: Terminate Cluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
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

	ctx.Data.Company = userInfo.CompanyId
	ctx.Data.InfraId = infraId
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, infraId, ctx.Data.Company, userInfo.UserId)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", infraId, "Terminate", token, utils.Context{})
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
	ctx.Data.InfraId = infraId
	region, zone, err := gcp.GetRegion(token, infraId, *ctx)
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

	cluster, err := gke.GetGKECluster(*ctx)
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
	if strings.ToLower(string(cluster.CloudplexStatus)) == strings.ToLower(string(models.New)) {
		ctx.SendLogs("GKEClusterController: Cluster is not in created state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is not in created state"}
		c.ServeJSON()
		return
	} else if cluster.CloudplexStatus == (models.Deploying) {
		ctx.SendLogs("GKEClusterController: cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.CloudplexStatus == (models.Terminating) {
		ctx.SendLogs("GKEClusterController: cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.CloudplexStatus == (models.ClusterTerminated) {
		ctx.SendLogs("GKEClusterController: Cluster is in terminated state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminated state"}
		c.ServeJSON()
		return
	} else if cluster.CloudplexStatus == (models.ClusterCreationFailed) {
		ctx.SendLogs("GKEClusterController: Cluster creation is in failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster creation is in failed state"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("GKEClusterController: Terminating cluster"+cluster.Name+" of infrastructure "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go gke.TerminateCluster(credentials, *ctx)

	/*err = gke.UpdateGKECluster(cluster, *ctx)
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
	}*/

	ctx.SendLogs("GKEClusterController: Cluster "+cluster.Name+" of infrastructure "+infraId+" terminated", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs(" GKE cluster "+cluster.Name+" of infrastructure "+cluster.InfraId+" terminated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(202)
	c.Data["json"] = map[string]string{"msg": "cluster termination initiated"}
	c.ServeJSON()
}

// @Title Start
// @Description Apply cloudplex Agent file to a gke cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	infraId	path	string	true	"Id of the infrastructure"
// @Success 200 {"msg": "Agent Applied successfully"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 512 {object} types.CustomCPError
// @router /applyagent/:infraId [post]
func (c *GKEClusterController) ApplyAgent() {

	ctx := new(utils.Context)
	ctx.SendLogs("GKEClusterController: Apply agent.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", infraId, "Start", token, utils.Context{})
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

	region, zone, err := gcp.GetRegion(token, infraId, *ctx)
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

	cluster, err := gke.GetGKECluster(*ctx)
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

	if cluster.CloudplexStatus != "Cluster Created" {
		text := "GKEClusterController: Cannot apply agent until cluster is in created state. Cluster is in " + string(cluster.CloudplexStatus) + " state."
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": text}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("GKEClusterController: Applying agent on cluster of the infrastructure "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go gke.ApplyAgent(credentials, token, *ctx, cluster.Name)

	ctx.SendLogs("GKEClusterController: Agent Applied on cluster of the infrastructure "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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
func (c *GKEClusterController) PatchRunningCluster() {

	ctx := new(utils.Context)

	ctx.SendLogs("GKEClusterController: Update running cluster ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	ctx.SendLogs("GKEClusterController: Updating cluster of infrastructure. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.Data.Company = userInfo.CompanyId
	ctx.Data.InfraId = infraId

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.GKE, "cluster", infraId, "Start", token, utils.Context{})
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

	region, zone, err := gcp.GetRegion(token, infraId, *ctx)
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

	cluster, err := gke.GetGKECluster(*ctx)
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

	if strings.ToLower(string(cluster.CloudplexStatus)) == strings.ToLower(string(models.New)) {
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is new state"}
		c.ServeJSON()
		return
	} else if cluster.CloudplexStatus == models.ClusterCreationFailed {
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is cluster creation failed state"}
		c.ServeJSON()
		return
	} else if cluster.CloudplexStatus == (models.Deploying) {
		ctx.SendLogs("GKEClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.CloudplexStatus == (models.Terminating) {
		ctx.SendLogs("GKEClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.CloudplexStatus == (models.ClusterTerminated) {
		ctx.SendLogs("GKEClusterController: Cluster is in terminated state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminated state"}
		c.ServeJSON()
		return
	}

	go gke.PatchRunningGKECluster(cluster, credentials, token, *ctx)

	ctx.SendLogs("GKEClusterController: Running cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+"updated", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs(" GKE running cluster "+cluster.Name+" of infrastructure Id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(202)
	c.Data["json"] = map[string]string{"msg": "Running cluster update initiated"}
	c.ServeJSON()
}
