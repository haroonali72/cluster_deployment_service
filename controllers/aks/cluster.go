package aks

import (
	"antelope/models"
	"antelope/models/aks"
	"antelope/models/azure"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/types"
	"antelope/models/utils"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/go-playground/validator/v10"
	"strings"
)

// Operations about AKS cluster [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/cluster/{cloud}/]
type AKSClusterController struct {
	beego.Controller
}

// @Title Get
// @Description Get cluster against the infraId
// @Param	infraId	path	string	true	"Id of the Infrastructuret"
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 200 {object} aks.AKSCluster
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /:infraId/ [get]
func (c *AKSClusterController) Get() {
	ctx := new(utils.Context)

	infraId := c.GetString(":infraId")
	if infraId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "Infrastructuret id is empty"}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("AKSClusterController: Get cluster with Infrastructuret id "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.AKS, "cluster", infraId, "View", token, utils.Context{})
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructuret id"}
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

	ctx.SendLogs("AKSClusterController: Get cluster with Infrastructuret id: "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := aks.GetAKSCluster(infraId, userInfo.CompanyId, *ctx)
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

	ctx.SendLogs(" AKS cluster "+cluster.Name+" of Infrastructuret Id: "+cluster.InfraId+" fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Get All
// @Description get all the clusters
// @Param	X-Auth-Token	header	string	true "token"
// @Success 200 {object} []aks.AKSCluster
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /all [get]
func (c *AKSClusterController) GetAll() {

	ctx := new(utils.Context)
	ctx.SendLogs("AKSClusterController: GetAll clusters.", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

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

	ctx.SendLogs("AKSClusterController: Getting all clusters ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, err, data := rbacAuthentication.GetAllAuthenticate("cluster", userInfo.CompanyId, token, models.AKS, *ctx)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructuret id"}
			c.ServeJSON()
			return
		}
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.Data.Company = userInfo.CompanyId
	clusters, err := aks.GetAllAKSCluster(data, *ctx)
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

	ctx.SendLogs("All AKS cluster fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = clusters
	c.ServeJSON()
}

// @Title Create
// @Description add a new cluster
// @Param	body body aks.AKSCluster true "body for cluster content"
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 201 {"msg": "Cluster created successfully"}
// @Success 400 {"msg": "Runtime Error"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Cluster against this Infrastructuret already exists"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [post]
func (c *AKSClusterController) Post() {

	var cluster aks.AKSCluster

	ctx := new(utils.Context)

	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		ctx.SendLogs("AKSClusterController: "+err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
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
		ctx.SendLogs("AKSClusterController: "+err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, cluster.InfraId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("AKSClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.AKS, "cluster", cluster.InfraId, "Create", token, utils.Context{})
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructuret id"}
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

	ctx.SendLogs("AKSClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	beego.Info("AKSClusterController: JSON Payload: ", cluster)

	validate := validator.New()
	err = validate.Struct(cluster)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	err = aks.ValidateAKSData(&cluster, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController: "+err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	err = aks.GetNetwork(token, cluster.InfraId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	cluster.CompanyId = userInfo.CompanyId
	err = aks.AddAKSCluster(cluster, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "cluster against same Infrastructuret id already exists"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AKS cluster "+cluster.Name+" of Infrastructuret Id: "+cluster.InfraId+" created ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Ctx.Output.SetStatus(201)
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
// @Failure 409 {"error": "Cluster is in Creating/Terminating/TerminationFailed state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [put]
func (c *AKSClusterController) Patch() {
	ctx := new(utils.Context)

	var cluster aks.AKSCluster
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		ctx.SendLogs("AKSClusterController: "+err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
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
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	if cluster.Status == (models.Deploying) {
		ctx.SendLogs("AKSClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Terminating) {
		ctx.SendLogs("AKSClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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

	err = aks.ValidateAKSData(&cluster, *ctx)
	if err != nil {
		ctx.SendLogs("AKSClusterController: "+err.Error(), models.LOGGING_LEVEL_INFO, models.Backend_Logging)
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, cluster.InfraId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("AKSClusterController: update cluster cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.AKS, "cluster", cluster.InfraId, "Update", token, utils.Context{})
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructuret id"}
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
	ctx.SendLogs("AKSClusterController: Patch cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	beego.Info("AKSClusterController: JSON Payload: ", cluster)

	cluster.CompanyId = userInfo.CompanyId
	if cluster.Status == (models.ClusterCreated) || cluster.Status == (models.ClusterTerminationFailed) || cluster.Status == (models.ClusterUpdateFailed) {
		err := aks.UpdatePreviousAKSCluster(cluster, *ctx)
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

		ctx.SendLogs("AKS running cluster "+cluster.Name+" in Infrastructuret Id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

		c.Data["json"] = map[string]string{"msg": "Running cluster updated successfully"}
		c.ServeJSON()
		return
	}
	err = aks.UpdateAKSCluster(cluster, *ctx)
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

	ctx.SendLogs("AKS cluster "+cluster.Name+" of Infrastructuret Id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = map[string]string{"msg": "cluster updated successfully"}
	c.ServeJSON()
}

// @Title Delete
// @Description Delete a cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	infraId	path 	string	true	"Infrastructuret id of the cluster"
// @Param	forceDelete path    boolean	true    "Forcefully delete cluster"
// @Success 204 {"msg": "Cluster deleted successfully"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in Cluster Created/Creating/Terminating/Termination Failed state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /:infraId/:forceDelete [delete]
func (c *AKSClusterController) Delete() {
	ctx := new(utils.Context)

	id := c.GetString(":infraId")
	if id == "" {
		ctx.SendLogs("AKSClusterController: InfraId field is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "Infrastructuret id is empty"}
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
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructuret id"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, id, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("AKSClusterController: Delete cluster with id "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.AKS, "cluster", id, "Delete", token, utils.Context{})
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructuret id"}
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

	ctx.SendLogs("AKSClusterController: Delete cluster with Infrastructuret id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := aks.GetAKSCluster(id, userInfo.CompanyId, *ctx)
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
	err = aks.DeleteAKSCluster(id, userInfo.CompanyId, *ctx)
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

	ctx.SendLogs("AKS cluster "+cluster.Name+" of Infrastructuret Id: "+cluster.InfraId+" deleted ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Ctx.Output.SetStatus(204)
	c.Data["json"] = map[string]string{"msg": "cluster deleted successfully"}
	c.ServeJSON()
}

// @Title Start
// @Description Deploy a kubernetes cluster
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	infraId	path	string	true	"Id of the Infrastructuret"
// @Success 201 {"msg": "Cluster created successfully"}
// @Success 202 {"msg": "Cluster creation initiated"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in Created/Creating/Terminating/TerminationFailed state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /start/:infraId [post]
func (c *AKSClusterController) StartCluster() {

	ctx := new(utils.Context)
	ctx.SendLogs("AKSClusterController: StartCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("AKSClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	infraId := c.GetString(":infraId")
	if infraId == "" {
		ctx.SendLogs("AKSClusterController: InfraId field is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "Infrastructuret id is empty"}
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
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructuret id"}
			c.ServeJSON()
			return
		}
		ctx.SendLogs("AKSClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.AKS, "cluster", infraId, "Start", token, utils.Context{})
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructuret id"}
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

	region, err := azure.GetRegion(token, infraId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
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

	ctx.SendLogs("AKSClusterController: Getting Cluster of Infrastructuret. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := aks.GetAKSCluster(infraId, userInfo.CompanyId, *ctx)
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
		ctx.SendLogs("AKSClusterController : Cluster is already running", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is already in running state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Deploying {
		ctx.SendLogs("AKSClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Terminating {
		ctx.SendLogs("AKSClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.ClusterTerminationFailed {
		ctx.SendLogs("AKSClusterController: Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in termination failed state"}
		c.ServeJSON()
		return
	}

	//cluster.Status = models.Deploying
	//err = aks.UpdateAKSCluster(cluster, *ctx)
	//if err != nil {
	//	c.Ctx.Output.SetStatus(500)
	//	c.Data["json"] = map[string]string{"error": err.Error()}
	//	c.ServeJSON()
	//	return
	//}
	ctx.SendLogs("AKSClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go aks.DeployAKSCluster(cluster, azureProfile, userInfo.CompanyId, token, *ctx)

	ctx.SendLogs(" AKS cluster "+cluster.Name+" of Infrastructuret Id: "+cluster.InfraId+" deployed ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(202)
	c.Data["json"] = map[string]string{"msg": "Cluster creation initiated"}
	c.ServeJSON()
}

// @Title Status
// @Description Get live status of the running cluster
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	infraId	path	string	true	"Id of the Infrastructuret"
// @Success 200 {object} aks.AKSCluster
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Cluster is in deploying/terminating state"}
// @Failure 500 {"error": "Internal Server Error"}
// @Failure 512 {object} types.CustomCPError
// @router /status/:infraId/ [get]
func (c *AKSClusterController) GetStatus() {
	ctx := new(utils.Context)
	ctx.SendLogs("AKSClusterController: FetchStatus.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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
		c.Data["json"] = map[string]string{"error": "Infrastructuret id is empty"}
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
		ctx.SendLogs("AKSClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("AKSClusterController: FetchStatus.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.AKS, "cluster", infraId, "View", token, utils.Context{})
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructuret id"}
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
	region, err := azure.GetRegion(token, infraId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
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

	ctx.SendLogs("AKSClusterController: Fetch Cluster Status of Infrastructuret. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, cpErr := aks.FetchStatus(azureProfile.Profile, token, infraId, userInfo.CompanyId, *ctx)
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
// @Description Terminate a running cluster
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	infraId	path	string	true	"Id of the Infrastructuret"
// @Success 202 {"msg": "Cluster termination started successfully"}
// @Success 204 {"msg": "Cluster terminated successfully"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in New/Creating/Creation Failed /Terminated/Terminating state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /terminate/:infraId/ [post]
func (c *AKSClusterController) TerminateCluster() {
	ctx := new(utils.Context)
	ctx.SendLogs("AKSClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		ctx.SendLogs("AKSClusterController: ProfileId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

	infraId := c.GetString(":infraId")
	if infraId == "" {
		ctx.SendLogs("AKSClusterController: InfraId is empty ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "Infrastructuret id is empty"}
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
		ctx.SendLogs("AKSClusterController: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("AKSClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.AKS, "cluster", infraId, "Terminate", token, utils.Context{})
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructuret id"}
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

	region, err := azure.GetRegion(token, infraId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
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

	ctx.SendLogs("AKSClusterController: Getting Cluster of Infrastructuret. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := aks.GetAKSCluster(infraId, userInfo.CompanyId, *ctx)
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
		ctx.SendLogs("AKSClusterController : Cluster is not in created state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is not in created state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Deploying {
		ctx.SendLogs("AKSClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.Terminating {
		ctx.SendLogs("AKSClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.ClusterTerminated {
		ctx.SendLogs("AKSClusterController: Cluster is in terminated state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminated state"}
		c.ServeJSON()
		return
	} else if cluster.Status == models.ClusterCreationFailed {
		ctx.SendLogs("AKSClusterController: Cluster creation is in failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster creation is in failed statee"}
		c.ServeJSON()
		return
	}

	//cluster.Status = models.Terminating
	//err = aks.UpdateAKSCluster(cluster, *ctx)
	//if err != nil {
	//	c.Ctx.Output.SetStatus(500)
	//	c.Data["json"] = map[string]string{"error": err.Error()}
	//	c.ServeJSON()
	//	return
	//}

	go aks.TerminateCluster(azureProfile, infraId, userInfo.CompanyId, token, *ctx)

	ctx.SendLogs(" AKS cluster "+cluster.Name+" of Infrastructuret Id: "+cluster.InfraId+" terminated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(202)
	c.Data["json"] = map[string]string{"msg": "Cluster termination initiated"}
	c.ServeJSON()
}

// @Title GetAKSVmsTypes
// @Description get aks vm types
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	region path	string	true "Cloud region"
// @Success 200 {object} []aks.VMSizeTypes
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /getvms/:region [get]
func (c *AKSClusterController) GetAKSVms() {
	ctx := new(utils.Context)
	ctx.SendLogs("AKSClusterController: GetVms.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
		c.ServeJSON()
		return
	}

	region := c.GetString(":region")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "region is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructuret id"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("AKSClusterController: GetVms.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	aksVms, err := aks.GetVms(region, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = aksVms
	c.ServeJSON()
}

// @Title Kubeconfig
// @Description get cluter kubeconfig
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	infraId	path	string	true	"Id of the Infrastructuret"
// @Failure 404 {"error": "Not Found"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 500 {"error": "Internal Server Error"}
// @Failure 512 {object} types.CustomCPError
// @router /kubeconfig/:infraId [get]
func (c *AKSClusterController) GetKubeConfig() {

	ctx := new(utils.Context)
	ctx.SendLogs("AKSClusterController: GetKubeConfig.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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
		c.Data["json"] = map[string]string{"error": "Infrastructuret id is empty"}
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
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructuret id"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)

	region, err := azure.GetRegion(token, infraId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
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

	ctx.SendLogs("AKSClusterController: Getting Cluster of Infrastructuret. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := aks.GetAKSCluster(infraId, userInfo.CompanyId, *ctx)
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
	ctx.SendLogs("AKSClusterController: GetKubeConfig. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	kubeconfig, CpErr := aks.GetKubeCofing(azureProfile.Profile, cluster, *ctx)
	if CpErr != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(int(models.CloudStatusCode))
		c.Data["json"] = CpErr
		c.ServeJSON()
		return
	}

	c.Data["json"] = kubeconfig
	c.ServeJSON()
}

// @Title Get Kube Versions
// @Description fetch version of kubernetes cluster
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Param	region	path	string	true	"Cloud region"
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 200 {object} []string
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Internal Server Error"}
// @Failure 512 {object} types.CustomCPError
// @router /getallkubeversions/:region [get]
func (c *AKSClusterController) FetchKubeVersions() {

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

	statusCode, userInfo, err := rbacAuthentication.GetInfo(token)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructuret id"}
			c.ServeJSON()
			return
		}
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)
	ctx.SendLogs("AKSClusterController: GetAllKubernetesVersions.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	statusCode, azureProfile, err := azure.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	kubeVersions, CpErr := aks.GetKubeVersions(azureProfile, *ctx)
	if CpErr != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(int(models.CloudStatusCode))
		c.Data["json"] = CpErr
		c.ServeJSON()
		return
	}
	c.Data["json"] = kubeVersions
	c.ServeJSON()
}

// @Title Update
// @Description Update a running kubernetes cluster
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	infraId	path	string	true	"Id of the Infrastructuret"
// @Success 201 {"msg": "Running cluster updated successfully"}
// @Success 202 {"msg": "Running cluster updation initiated"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in New/Creating/Creation Failed/Terminating/Terminated state"}
// @Failure 404 {"error": "Not found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /update/:infraId [put]
func (c *AKSClusterController) PatchRunningCluster() {

	ctx := new(utils.Context)

	ctx.SendLogs("AKSClusterController: Update running cluster ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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
		c.Data["json"] = map[string]string{"error": "Infrastructuret id is empty"}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, infraId, userInfo.CompanyId, userInfo.UserId)

	ctx.SendLogs("AKSClusterController: Updating cluster of Infrastructuret. "+infraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.Data.Company = userInfo.CompanyId
	ctx.Data.InfraId = infraId

	statusCode, allowed, err := rbacAuthentication.Authenticate(models.AKS, "cluster", infraId, "Start", token, utils.Context{})
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this Infrastructuret id"}
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

	region, err := azure.GetRegion(token, infraId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
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

	cluster, err := aks.GetAKSCluster(infraId, userInfo.CompanyId, *ctx)
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

	go aks.PatchRunningAKSCluster(cluster, azureProfile, userInfo.CompanyId, token, *ctx)

	ctx.SendLogs("AKSClusterController: Running cluster "+cluster.Name+" of Infrastructuret Id: "+cluster.InfraId+"updated", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs(" AKS running cluster "+cluster.Name+" of Infrastructuret Id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(202)
	c.Data["json"] = map[string]string{"msg": "Running cluster update initiated"}
	c.ServeJSON()
}
