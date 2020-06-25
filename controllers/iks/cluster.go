package iks

import (
	"antelope/models"
	"antelope/models/iks"
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

// Operations about ikscluster [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/cluster/{cloud}/]
type IKSClusterController struct {
	beego.Controller
}

// @Title Get Instance List
// @Description Get all available instance list
// @Param	X-Profile-Id header	X-Profile-Id	string	true "Vault credentials profile id"
// @Param	X-Auth-Token header	string	true "Token"
// @Param	zone	path	string	true	"Region of the cloud"
// @Success 200 {object} iks.AllInstancesResponse
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime error"}
// @Failure 512 {object} types.CustomCPError
// @router /getallmachines/:zone/ [get]
func (c *IKSClusterController) GetAllMachineTypes() {
	ctx := new(utils.Context)
	ctx.SendLogs("IKSClusterController: Get All Machines.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	region := c.GetString(":zone")
	if region == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "region is empty"}
		c.ServeJSON()
		return
	}

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this project id"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	statusCode, ibmProfile, err := iks.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("IKSClusterController: Getting All Machines. "+"", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	machineTypes, cpErr := iks.GetAllMachines(ibmProfile, *ctx)
	if cpErr != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(int(models.CloudStatusCode))
		c.Data["json"] = cpErr
		c.ServeJSON()
		return
	}

	ctx.SendLogs("IKSClusterController: All machines fetched", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs("IKSClusterController: All machines fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = machineTypes
	c.ServeJSON()
}

// @Title Get Regions
// @Description fetch regions of IKS
// @Param	X-Auth-Token	header	string	true	"Token"
// @Success 200 {object} []iks.Regions
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /getallregions/ [get]
func (c *IKSClusterController) FetchRegions() {
	ctx := new(utils.Context)
	ctx.SendLogs("IKSClusterController: Fetch regions", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	ctx.SendLogs("IKSClusterController: Getting regions", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	regions, err := iks.GetRegions(*ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("IKSClusterController: Regions fetched", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs(" Iks Regions fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = regions
	c.ServeJSON()
}

// @Title Get Kubernetes Versions
// @Description fetch version of kubernetes cluster
// @Param region path string true "region of the cloud"
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Param   X-Auth-Token	header string true "Token"
// @Success 200 {object} []iks.Versions
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Unauthorized"}
// @Failure 512 {object} types.CustomCPError
// @router /getallkubeversions/:region [get]
func (c *IKSClusterController) FetchKubeVersions() {
	ctx := new(utils.Context)
	ctx.SendLogs("IKSClusterController: Get kubernetes versions", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	region := c.GetString(":region")
	if region == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "region is empty"}
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

	statusCode, ibmProfile, err := iks.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("IKSClusterController: Getting All kubernetes versions "+"", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	versions, cpErr := iks.GetAllVersions(ibmProfile, *ctx)
	if cpErr != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(int(models.CloudStatusCode))
		c.Data["json"] = cpErr
		c.ServeJSON()
		return
	}

	ctx.SendLogs("IKSClusterController: Kubernetes versions fetched", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs("IKSClusterController: Kubernetes versions fetched", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = versions.Kubernetes
	c.ServeJSON()
}

// @Title Get Zone
// @Description get zone against region
// @Param	region	path	string	true	"Region of the cloud"
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 200 {object} []iks.Zone
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /getzones/:region/ [get]
func (c *IKSClusterController) FetchZones() {
	ctx := new(utils.Context)
	ctx.SendLogs("IKSClusterController:Fetch Zones ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
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

	statusCode, userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	ctx.SendLogs("IKSClusterController: Fetching zones ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	regions, err := iks.GetZones(region, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("IKSClusterController: Zones fetched ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs(" IKS zones fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = regions
	c.ServeJSON()
}

// @Title Get
// @Description Get cluster against the projectId
// @Param	projectId	path	string	true	"Id of the project"
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 200 {object} iks.Cluster_Def
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /:projectId/ [get]
func (c *IKSClusterController) Get() {

	ctx := new(utils.Context)
	ctx.SendLogs("IKSClusterController: Get Cluster", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
		c.ServeJSON()
		return
	}

	projectId := c.GetString(":projectId")
	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbac_athentication.Authenticate(models.IKS, "cluster", projectId, "View", token, *ctx)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this project id"}
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

	//====================================================================================//

	ctx.SendLogs("IKSClusterController: Getting cluster of project "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err := iks.GetCluster(projectId, userInfo.CompanyId, *ctx)
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

	ctx.SendLogs("IKSClusterController: Cluster"+cluster.Name+" of project "+projectId+" fetched", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs(" Iks cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Get All
// @Description get all the clusters
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 200 {object} []iks.Cluster
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Erorr"}
// @router /all [get]
func (c *IKSClusterController) GetAll() {

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
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, err, data := rbac_athentication.GetAllAuthenticate("cluster", userInfo.CompanyId, token, models.IKS, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	//====================================================================================//

	ctx.SendLogs("IKSClusterController: Getting all clusters.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.Data.Company = userInfo.CompanyId
	clusters, err := iks.GetAllCluster(*ctx, data)
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

	ctx.SendLogs("IKSClusterController: All Iks clusters fetched.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs(" All Iks clusters fetched ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	c.Data["json"] = clusters
	c.ServeJSON()
}

// @Title Create
// @Description Add a new cluster
// @Param	body	body 	iks.Cluster_Def		true	"Body for cluster content"
// @Param	X-Auth-Token header	string	true "Token"
// @Success 201 {"msg": "Cluster added successfully"}
// @Success 400 {"msg": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Cluster against this project already exists"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [post]
func (c *IKSClusterController) Post() {
	ctx := new(utils.Context)
	var cluster iks.Cluster_Def

	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
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
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
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
	err = iks.GetNetwork(token, cluster.ProjectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	//custom data validation
	err = iks.ValidateIKSData(cluster, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, cluster.ProjectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.IKS, "cluster", cluster.ProjectId, "Create", token, *ctx)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this project id"}
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

	ctx.SendLogs("IKSClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster.CompanyId = userInfo.CompanyId

	ctx.SendLogs("IKSClusterController: Add new cluster "+cluster.Name+" in project "+cluster.ProjectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err = iks.CreateCluster(cluster, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "cluster against this project id  already exists"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("IKSClusterController: New cluster "+cluster.Name+" in project "+cluster.ProjectId+" added", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs(" Iks cluster "+cluster.Name+" in project "+cluster.ProjectId+" added ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = map[string]string{"msg": "cluster added successfully"}
	c.ServeJSON()
}

// @Title Update
// @Description Update an existing cluster
// @Param	X-Auth-Token	header	string	true "token"
// @Param	body	body 	iks.Cluster_Def	true	"Body for cluster content"
// @Success 200 {"msg": "Cluster updated successfully"}
// @Failure 409 {"error": "Cluster is in Creating/Terminating/TerminationFailed state"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router / [put]
func (c *IKSClusterController) Patch() {
	var cluster iks.Cluster_Def
	ctx := new(utils.Context)

	token := c.Ctx.Input.Header("X-Auth-Token")
	if token == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Auth-Token is empty"}
		c.ServeJSON()
		return
	}

	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "error while unmarshalling " + err.Error()}
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

	if cluster.Status != models.New && cluster.Status != models.ClusterCreationFailed && cluster.Status != models.ClusterTerminated {
		ctx.SendLogs("IKSClusterController : Cluster is in "+string(cluster.Status)+" state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Can't update.Cluster is in " + string(cluster.Status) + " state"}
		c.ServeJSON()
		return
	}
	if cluster.Status == (models.Deploying) {
		ctx.SendLogs("IKSClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Terminating) {
		ctx.SendLogs("IKSClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.ClusterTerminationFailed) {
		ctx.SendLogs("IKSClusterController: Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": " Cluster creation is in termination failed state"}
		c.ServeJSON()
		return
	}
	if cluster.Status == (models.ClusterCreated) {
		c.Data["json"] = map[string]string{"msg": "Cluster updated successfully"}
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

	//custom data validation
	err = iks.ValidateIKSData(cluster, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, cluster.ProjectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.IKS, "cluster", cluster.ProjectId, "Update", token, *ctx)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this project id"}
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

	ctx.SendLogs("IKSClusterController:Update cluster "+cluster.Name+" of project"+cluster.ProjectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster.CompanyId = userInfo.CompanyId
	err = iks.UpdateCluster(cluster, true, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "not found") {
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

	ctx.SendLogs("IKSClusterController: Cluster "+cluster.Name+" of project"+cluster.ProjectId+" updated", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs(" Iks cluster "+cluster.Name+" of project  "+cluster.ProjectId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = map[string]string{"msg": "Cluster updated successfully"}
	c.ServeJSON()
}

// @Title Delete
// @Description delete a cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	projectId	path 	string	true	"Project id of the cluster"
// @Param	forceDelete path    boolean	true    "Forcefully delete cluster"
// @Success 204 {"msg": "Cluster deleted successfully"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in Creating/Created/Terminating/TerminationFailed state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /:projectId/:forceDelete [delete]
func (c *IKSClusterController) Delete() {
	ctx := new(utils.Context)
	ctx.SendLogs("IKSClusterController: Delete cluster", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, id, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	statusCode, allowed, err := rbac_athentication.Authenticate(models.IKS, "cluster", id, "Delete", token, *ctx)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this project id"}
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

	cluster, err := iks.GetCluster(id, userInfo.CompanyId, *ctx)
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

	if cluster.Status == models.ClusterCreated && !forceDelete {
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in created state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Deploying) && !forceDelete {
		ctx.SendLogs("cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Terminating) && !forceDelete {
		ctx.SendLogs("cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.ClusterTerminationFailed) && !forceDelete {
		ctx.SendLogs("DOKSClusterController: Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": " Cluster creation is in termination failed state"}
		c.ServeJSON()
		return
	}
	if cluster.Status == (models.ClusterTerminationFailed) {
		ctx.SendLogs("IKSClusterController: Cluster is in termination failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "cluster is in termination failed state"}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("IKSClusterController: Delete cluster with project id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	err = iks.DeleteCluster(id, userInfo.CompanyId, *ctx)
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

	ctx.SendLogs("IKSClusterController: Cluster of project "+id+" deleted", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs(" Iks cluster "+cluster.Name+" of project "+cluster.ProjectId+" deleted ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(204)
	c.Data["json"] = map[string]string{"msg": "cluster deleted successfully"}
	c.ServeJSON()
}

// @Title Start
// @Description Deploy a kubernetes cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	X-Profile-Id	header	string	true "Vault credentials profile id"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 201 {"msg": "Cluster created successfully"}
// @Success 202 {"msg": "Cluster creation initiated"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in Created/Creating/Terminating/TerminationFailed state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /start/:projectId [post]
func (c *IKSClusterController) StartCluster() {
	var cluster iks.Cluster_Def
	ctx := new(utils.Context)
	ctx.SendLogs("IKSClusterController:Start Cluster ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
		c.ServeJSON()
		return
	}

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
	/*	if err != nil {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
			return
		}
	*/
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbac_athentication.Authenticate(models.IKS, "cluster", projectId, "Start", token, *ctx)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this project id"}
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

	ctx.SendLogs("DONetworkController: StartCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	cluster, err = iks.GetCluster(projectId, userInfo.CompanyId, *ctx)
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

	region, err := iks.GetRegion(token, projectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	statusCode, ibmProfile, err := iks.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		utils.SendLog(userInfo.CompanyId, err.Error(), "error", cluster.ProjectId)
		utils.SendLog(userInfo.CompanyId, "Cluster creation failed: "+cluster.Name, "error", cluster.ProjectId)
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	/*	cluster.Status = (models.Deploying)
		err = iks.UpdateCluster(cluster, false, *ctx)
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

	ctx.SendLogs("IKSClusterController: Creating Cluster. "+cluster.Name+" of project"+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go iks.DeployCluster(cluster, ibmProfile.Profile, *ctx, userInfo.CompanyId, token)

	ctx.SendLogs("IKSClusterController: Cluster. "+cluster.Name+" of project "+projectId+" started", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs(" Iks cluster "+cluster.Name+" of project Id: "+cluster.ProjectId+" started ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(202)
	c.Data["json"] = map[string]string{"msg": "Cluster creation initiated"}
	c.ServeJSON()
}

// @Title Status
// @Description Get live status of the running cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile Id"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {object} iks.KubeClusterStatus1
// @Failure 401 {"error": "Unauthorized"}
// @Failure 404 {"error": "Not Found"}
// @Failure 409 {"error": "Cluster is in deploying/terminating state"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 512 {object} types.CustomCPError
// @router /status/:projectId/ [get]
func (c *IKSClusterController) GetStatus() {
	ctx := new(utils.Context)
	ctx.SendLogs("IKSClusterController: Fetch Status.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	if profileId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "X-Profile-Id is empty"}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbac_athentication.Authenticate(models.IKS, "cluster", projectId, "View", token, *ctx)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this project id"}
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

	ctx.SendLogs("IKSClusterController: Fetch Cluster Status of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	region, err := iks.GetRegion(token, projectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	statusCode, ibmProfile, err := iks.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("IKSClusterController: Fetching Status of project"+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, cpErr := iks.FetchStatus(ibmProfile, projectId, *ctx, userInfo.CompanyId, token)
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
	ctx.SendLogs("IKSClusterController: Status fetched of project"+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs("IKSClusterController: Status fetched of project"+projectId, models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Terminate
// @Description Terminate a running cluster
// @Param	X-Profile-Id header	X-Profile-Id	string	true "Vault credentials profile Id"
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 202 {"msg": "Cluster termination initiated"}
// @Success 204 {"msg": "Cluster terminated successfully"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 409 {"error": "Cluster is in New/Creating/Creation Failed /Terminated/Terminating state"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @router /terminate/:projectId/ [post]
func (c *IKSClusterController) TerminateCluster() {
	ctx := new(utils.Context)
	ctx.SendLogs("IKSClusterController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	projectId := c.GetString(":projectId")
	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
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

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	statusCode, allowed, err := rbac_athentication.Authenticate(models.IKS, "cluster", projectId, "Terminate", token, *ctx)
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this project id"}
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

	region, err := iks.GetRegion(token, projectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	statusCode, ibmProfile, err := iks.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	var cluster iks.Cluster_Def

	cluster, err = iks.GetCluster(projectId, userInfo.CompanyId, *ctx)
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
		ctx.SendLogs("IKSClusterController: Cluster is not in created state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is not in created state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Deploying) {
		ctx.SendLogs("IKSClusterController: Cluster is in creating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in creating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.Terminating) {
		ctx.SendLogs("IKSClusterController: Cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminating state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.ClusterTerminated) {
		ctx.SendLogs("IKSClusterController: Cluster is in terminated state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster is in terminated state"}
		c.ServeJSON()
		return
	} else if cluster.Status == (models.ClusterCreationFailed) {
		ctx.SendLogs("IKSClusterController: Cluster  creation is in failed state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(409)
		c.Data["json"] = map[string]string{"error": "Cluster  creation is in failed state"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("IKSClusterController: Terminating Cluster "+cluster.Name+" of project"+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go iks.TerminateCluster(cluster, ibmProfile, *ctx, userInfo.CompanyId, token)

	ctx.SendLogs("IKSClusterController: Cluster "+cluster.Name+" of project"+projectId+" terminated", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	/*	err = iks.UpdateCluster(cluster, false, *ctx)
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

	ctx.SendLogs("IKSClusterController: Cluster "+cluster.Name+" of project"+projectId+" terminated", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Ctx.Output.SetStatus(202)
	c.Data["json"] = map[string]string{"msg": "Cluster termination initiated"}
	c.ServeJSON()
}

// @Title Apply agent
// @Description Apply cloudplex agent file to eks cluster
// @Param	X-Auth-Token	header	string	true "Token"
// @Param	X-Profile-Id	header	string	true	"Vault credentials profile id"
// @Param	resourceGroup	header	string	true "Resource Group"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {"msg": "Agent Applied successfully"}
// @Failure 404 {"error": "Not Found"}
// @Failure 401 {"error": "Unauthorized"}
// @Failure 500 {"error": "Runtime Error"}
// @router /applyagent/:projectId [post]
func (c *IKSClusterController) ApplyAgent() {

	ctx := new(utils.Context)
	ctx.SendLogs("IKSClusterController: Apply agent ", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

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

	projectId := c.GetString(":projectId")
	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
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

	cluster, err := iks.GetCluster(projectId, userInfo.CompanyId, *ctx)
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

	if cluster.Status != "Cluster Created" {
		text := "IKSClusterController: Cannot apply agent until cluster is in created state. Cluster is in " + string(cluster.Status) + " state."
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": text}
		c.ServeJSON()
		return
	}

	resourceGroup := c.Ctx.Input.Header("resourceGroup")
	if resourceGroup == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "resourceGroup is empty"}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	statusCode, allowed, err := rbac_athentication.Authenticate(models.IKS, "cluster", projectId, "Start", token, utils.Context{})
	if err != nil {
		if statusCode == 404 && strings.Contains(strings.ToLower(err.Error()), "policy") {
			c.Ctx.Output.SetStatus(statusCode)
			c.Data["json"] = map[string]string{"error": "No policy exist against this project id"}
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

	region, err := iks.GetRegion(token, projectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	statusCode, ibmProfile, err := iks.GetProfile(profileId, region, token, *ctx)
	if err != nil {
		utils.SendLog(userInfo.CompanyId, err.Error(), "error", projectId)
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("IKSubernetesClusterController: Applying agent on cluster of project "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	go iks.ApplyAgent(ibmProfile.Profile, token, *ctx, cluster.Name, resourceGroup)

	ctx.SendLogs("IKSubernetesClusterController: Agent applied on cluster of project "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	ctx.SendLogs("IKSubernetesClusterController: Agent applied on cluster of project "+projectId, models.LOGGING_LEVEL_INFO, models.Audit_Trails)

	c.Data["json"] = map[string]string{"msg": "agent deployment is in progress"}
	c.ServeJSON()
}

// @Title Validate Profile
// @Description Validate IBM profile
// @Param	body	body 	vault.IBMCredentials		true	"Body for cluster content"
// @Param	X-Auth-Token	header	string	true "Token"
// @Success 200 {"msg": "Valid Profile"}
// @Failure 400 {"error": "Bad Request"}
// @Failure 404 {"error": "Not Found"}
// @Failure 500 {"error": "Runtime Error"}
// @Failure 512 {object} types.CustomCPError
// @router /validateProfile [post]
func (c *IKSClusterController) ValidateProfile() {
	ctx := new(utils.Context)
	var profile vault.IBMProfile

	profile.Profile = vault.IBMCredentials{}
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &profile.Profile)
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
		c.Ctx.Output.SetStatus(statusCode)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	ctx.SendLogs("IKSClusterController: Validating profile.", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	region := "us-east"
	profile.Profile.Region = region

	cpErr := iks.ValidateProfile(profile, *ctx)
	if cpErr != (types.CustomCPError{}) {
		c.Ctx.Output.SetStatus(cpErr.StatusCode)
		c.Data["json"] = cpErr
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"msg": "Valid Profile"}
	c.ServeJSON()
}
