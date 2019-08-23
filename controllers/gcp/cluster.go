package gcp

import (
	"antelope/models/gcp"
	"antelope/models/utils"
	"encoding/json"
	"github.com/astaxie/beego"
	"strings"
)

// Operations about Gcp cluster [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/cluster/{cloud}/]
type GcpClusterController struct {
	beego.Controller
}

// @Title Get
// @Description get cluster
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {object} gcp.Cluster_Def
// @Failure 404 {"error": exception_message}
// @Failure 500 {"error": "internal server error"}
// @router /:projectId/ [get]
func (c *GcpClusterController) Get() {
	projectId := c.GetString(":projectId")
	beego.Info("GcpClusterController: Get cluster with project id: ", projectId)

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId)
	ctx.SendSDLog("GcpClusterController: Get cluster with project id "+projectId, "info")

	if projectId == "" {
		ctx.SendSDLog("GcpClusterController: projectId is empty", "error")

		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	cluster, err := gcp.GetCluster(projectId, *ctx)
	if err != nil {
		ctx.SendSDLog("GcpGetClusterController: error getting gcp cluster "+err.Error(), "error")

		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "no cluster exists for this name"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Get All
// @Description get all the clusters
// @Success 200 {object} []gcp.Cluster_Def
// @Failure 500 {"error": "internal server error"}
// @router /all [get]
func (c *GcpClusterController) GetAll() {
	beego.Info("GcpClusterController: GetAll clusters.")

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "")
	ctx.SendSDLog("GcpClusterController: getting all clusters ", "info")

	clusters, err := gcp.GetAllCluster(*ctx)
	if err != nil {
		ctx.SendSDLog("GcpClusterController: "+err.Error(), "error")

		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = clusters
	c.ServeJSON()
}

// @Title Create
// @Description create a new cluster
// @Param	body	body 	gcp.Cluster_Def		true	"body for cluster content"
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 409 {"error": "cluster against same project id already exists"}
// @Failure 500 {"error": "internal server error"}
// @router / [post]
func (c *GcpClusterController) Post() {
	var cluster gcp.Cluster_Def
	json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, cluster.ProjectId)
	ctx.SendSDLog("GcpClusterController: Post new cluster with name: "+cluster.Name, "info")

	beego.Info("GcpClusterController: Post new cluster with name: ", cluster.Name)
	beego.Info("GcpClusterController: JSON Payload: ", cluster)

	err := gcp.CreateCluster(cluster, *ctx)
	if err != nil {
		ctx.SendSDLog("GcpClusterController: "+err.Error(), "error")
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "cluster against same project id already exists"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"msg": "cluster added successfully"}
	c.ServeJSON()
}

// @Title Update
// @Description update an existing cluster
// @Param	body	body 	gcp.Cluster_Def	true	"body for cluster content"
// @Success 200 {"msg": "cluster updated successfully"}
// @Failure 404 {"error": "no cluster exists with this name"}
// @Failure 500 {"error": "internal server error"}
// @router / [put]
func (c *GcpClusterController) Patch() {
	var cluster gcp.Cluster_Def
	json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	beego.Info("GcpClusterController: Patch cluster with name: ", cluster.Name)
	beego.Info("GcpClusterController: JSON Payload: ", cluster)

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, cluster.ProjectId)
	ctx.SendSDLog("GcpClusterController: update cluster cluster with name: "+cluster.Name, "info")

	err := gcp.UpdateCluster(cluster, true, *ctx)
	if err != nil {
		ctx.SendSDLog("GcpClusterController: "+err.Error(), "error")
		if strings.Contains(err.Error(), "does not exist") {
			c.Ctx.Output.SetStatus(404)
			c.Data["json"] = map[string]string{"error": "no cluster exists with this name"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"msg": "cluster updated successfully"}
	c.ServeJSON()
}

// @Title Delete
// @Description delete a cluster
// @Param	projectId	path	string	true	"project id of the cluster"
// @Success 200 {"msg": "cluster deleted successfully"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 500 {"error": "internal server error"}
// @router /:projectId [delete]
func (c *GcpClusterController) Delete() {
	id := c.GetString(":projectId")

	beego.Info("GcpClusterController: Delete cluster with project id: ", id)

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, id)
	ctx.SendSDLog("GcpClusterController: update cluster cluster with name: ", "info")

	if id == "" {
		ctx.SendSDLog("GcpClusterController: projectId field is empty ", "error")
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "name is empty"}
		c.ServeJSON()
		return
	}

	cluster, err := gcp.GetCluster(id, *ctx)
	if err == nil && cluster.Status == "Cluster Created" {
		ctx.SendSDLog("GcpClusterController: Cluster is in running state ", "error")
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + "Cluster is in running state"}
		c.ServeJSON()
		return
	}
	err = gcp.DeleteCluster(id, *ctx)
	if err != nil {
		ctx.SendSDLog("GcpClusterController: "+err.Error(), "error")
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"msg": "cluster deleted successfully"}
	c.ServeJSON()
}

// @Title Start
// @Description starts a  cluster
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 400 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error"}
// @router /start/:projectId [post]
func (c *GcpClusterController) StartCluster() {
	beego.Info("GcpClusterController: StartCluster.")

	profileId := c.Ctx.Input.Header("X-Profile-Id")

	projectId := c.GetString(":projectId")

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId)
	ctx.SendSDLog("GcpClusterController: Start Cluster ", "info")

	if profileId == "" {
		ctx.SendSDLog("GcpClusterController: ProfileId is empty ", "error")
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	if projectId == "" {
		ctx.SendSDLog("GcpClusterController: ProjectId is empty ", "error")
		c.Ctx.Output.SetStatus(400) //no need
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	region, zone, err := gcp.GetRegion(projectId, *ctx)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	isValid, credentials := gcp.IsValidGcpCredentials(profileId, region, zone, *ctx)
	if !isValid {
		ctx.SendSDLog("gcpClusterController : authorization params missing or invalid", "error")
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	var cluster gcp.Cluster_Def

	beego.Info("GcpClusterController: Getting Cluster of project. ", projectId)

	cluster, err = gcp.GetCluster(projectId, *ctx)

	if err != nil {
		ctx.SendSDLog("gcpClusterController :"+err.Error(), "error")
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	if cluster.Status == "Cluster Created" {
		ctx.SendSDLog("gcpClusterController : cluster is already running", "error")
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is already in running state"}
		c.ServeJSON()
		return
	}
	beego.Info("GcpClusterController: Creating Cluster. ", cluster.Name)

	go gcp.DeployCluster(cluster, credentials, *ctx)

	c.Data["json"] = map[string]string{"msg": "cluster creation in progress"}
	c.ServeJSON()
}

// @Title Status
// @Description returns status of nodes
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {object} gcp.Cluster_Def
// @Failure 206 {object} gcp.Cluster_Def
// @Failure 400 {"error": "exception_message"}
// @Failure 401 {"error": "authorization params missing or invalid"}
// @Failure 500 {"error": "internal server error"}
// @router /status/:projectId/ [get]
func (c *GcpClusterController) GetStatus() {
	beego.Info("GcpClusterController: FetchStatus.")

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	projectId := c.GetString(":projectId")

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId)

	ctx.SendSDLog("GcpNetworkController: FetchStatus.", "info")

	if profileId == "" {
		ctx.SendSDLog("GcpClusterController: ProfileId is empty ", "error")
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	if projectId == "" {
		ctx.SendSDLog("GcpClusterController: ProjectId is empty ", "error")
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	region, zone, err := gcp.GetRegion(projectId, *ctx)
	if err != nil {
		ctx.SendSDLog("GcpClusterController :"+err.Error(), "error")
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	isValid, credentials := gcp.IsValidGcpCredentials(profileId, region, zone, *ctx)
	if !isValid {
		ctx.SendSDLog("GcpClusterController : Gcp credentials not valid ", "error")
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	beego.Info("GcpClusterController: Fetch Cluster Status of project. ", projectId)

	cluster, err := gcp.FetchStatus(credentials, projectId, *ctx)

	if err != nil {
		ctx.SendSDLog("gcpClusterController :"+err.Error(), "error")
		c.Ctx.Output.SetStatus(206)
	}

	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Terminate
// @Description terminates a  cluster
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {"msg": "cluster terminated successfully"}
// @Failure 401 {"error": "Authorization format should be 'base64 encoded service_account_json'"}
// @Failure 400 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error"}
// @router /terminate/:projectId/ [post]
func (c *GcpClusterController) TerminateCluster() {
	beego.Info("GcpClusterController: TerminateCluster.")

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	projectId := c.GetString(":projectId")

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId)

	ctx.SendSDLog("GcpNetworkController: TerminateCluster.", "info")

	if profileId == "" {
		ctx.SendSDLog("GcpClusterController: ProfileId is empty ", "error")
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	if projectId == "" {
		ctx.SendSDLog("GcpClusterController: ProjectId is empty ", "error")
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	region, zone, err := gcp.GetRegion(projectId, *ctx)
	if err != nil {
		ctx.SendSDLog("GcpClusterController :"+err.Error(), "error")
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	isValid, credentials := gcp.IsValidGcpCredentials(profileId, region, zone, *ctx)
	if !isValid {
		ctx.SendSDLog("GcpClusterController: athorization params missing or invalid", "error")
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	var cluster gcp.Cluster_Def

	beego.Info("GcpClusterController: Getting Cluster of project. ", projectId)

	cluster, err = gcp.GetCluster(projectId, *ctx)

	if err != nil {
		ctx.SendSDLog("GcpClusterController :"+err.Error(), "error")
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}
	beego.Info("GcpClusterController: Terminating Cluster. ", cluster.Name)

	go gcp.TerminateCluster(cluster, credentials, *ctx)

	c.Data["json"] = map[string]string{"msg": "cluster termination is in progress"}
	c.ServeJSON()
}

// @Title SSHKeyPair
// @Description returns ssh key pairs
// @Success 200 {object} []string
// @Failure 500 {"error": "internal server error"}
// @router /sshkeys [get]
func (c *GcpClusterController) GetSSHKeys() {
	beego.Info("GcpClusterController: FetchExistingSSHKeys.")

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "")

	ctx.SendSDLog("AWSNetworkController: FetchExistingSSHKeys.", "info")

	keys, err := gcp.GetAllSSHKeyPair(*ctx)

	if err != nil {
		ctx.SendSDLog("GcpClusterController :"+err.Error(), "error")

		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = keys
	c.ServeJSON()
}

// @Title ListServiceAccounts
// @Description returns list of service account emails
// @Param	X-Profile-Id	header	string	true	"vault credentials profile id"
// @Success 200 {object} []string
// @Failure 400 {"error": "profile id is empty"}
// @Failure 401 {"error": "authorization params missing or invalid"}
// @Failure 500 {"error": "internal server error"}
// @router /serviceaccounts [get]
func (c *GcpClusterController) GetServiceAccounts() {
	beego.Info("GcpClusterController: FetchExistingServiceAccounts.")

	profileId := c.Ctx.Input.Header("X-Profile-Id")

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "")
	ctx.SendSDLog("AWSNetworkController: Getting service accounts ", "info")

	if profileId == "" {
		ctx.SendSDLog("GcpClusterController: ProfileId is empty ", "error")
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "profile id is empty"}
		c.ServeJSON()
		return
	}

	isValid, credentials := gcp.IsValidGcpCredentials(profileId, "", "", *ctx)
	if !isValid {
		ctx.SendSDLog("GcpClusterController: authorization params missing or invalid ", "error")
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "authorization params missing or invalid"}
		c.ServeJSON()
		return
	}

	serviceAccounts, err := gcp.GetAllServiceAccounts(credentials, *ctx)
	if err != nil {
		ctx.SendSDLog("gcpClusterController :"+err.Error(), "error")
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = serviceAccounts
	c.ServeJSON()
}
