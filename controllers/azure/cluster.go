package azure

import (
	"antelope/models/aws"
	"antelope/models/azure"
	"antelope/models/logging"
	"encoding/json"
	"github.com/astaxie/beego"
	"strings"
	"time"
)

// Operations about azure cluster [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/cluster/{cloud}/]
type AzureClusterController struct {
	beego.Controller
}

// @Title Get
// @Description get cluster
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {object} azure.Cluster_Def
// @Failure 404 {"error": exception_message}
// @Failure 500 {"error": "internal server error"}
// @router /:projectId/ [get]
func (c *AzureClusterController) Get() {
	projectId := c.GetString(":projectId")

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId)

	ctx.SendSDLog("AWSClusterController: Get cluster with project id "+projectId, "info")

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	cluster, err := azure.GetCluster(projectId, *ctx)
	if err != nil {
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
// @Success 200 {object} []azure.Cluster_Def
// @Failure 500 {"error": "internal server error"}
// @router /all [get]
func (c *AzureClusterController) GetAll() {
	beego.Info("AzureClusterController: GetAll clusters.")

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "")

	clusters, err := azure.GetAllCluster(*ctx)
	if err != nil {
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
// @Param	body	body 	azure.Cluster_Def		true	"body for cluster content"
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 409 {"error": "cluster against same project id already exists"}
// @Failure 500 {"error": "internal server error"}
// @router / [post]
func (c *AzureClusterController) Post() {
	var cluster azure.Cluster_Def
	json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, cluster.ProjectId)

	ctx.SendSDLog("AzureClusterController: Post new cluster with name: "+cluster.Name, "error ")

	cluster.CreationDate = time.Now()

	err := azure.CreateCluster(cluster, *ctx)
	if err != nil {
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
// @Param	body	body 	azure.Cluster_Def	true	"body for cluster content"
// @Success 200 {"msg": "cluster updated successfully"}
// @Failure 404 {"error": "no cluster exists with this name"}
// @Failure 500 {"error": "internal server error"}
// @router / [put]
func (c *AzureClusterController) Patch() {

	var cluster azure.Cluster_Def
	json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, cluster.ProjectId)

	ctx.SendSDLog("AzureClusterController: Patch cluster with name: "+cluster.Name, "info")

	err := azure.UpdateCluster(cluster, true, *ctx)
	if err != nil {
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
func (c *AzureClusterController) Delete() {
	id := c.GetString(":projectId")

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, id)

	ctx.SendSDLog("AzureClusterController: Delete cluster with project id: "+id, "error")

	if id == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "name is empty"}
		c.ServeJSON()
		return
	}
	cluster, err := aws.GetCluster(id, *ctx)
	if err == nil && cluster.Status == "Cluster Created" {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + "Cluster is in running state"}
		c.ServeJSON()
		return
	}
	err = azure.DeleteCluster(id, *ctx)
	if err != nil {
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
// @Param	Authorization	header	string	false	"{id}:{key}:{tenant}:{subscription}:{region}"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 400 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error"}
// @router /start/:projectId [post]
func (c *AzureClusterController) StartCluster() {

	projectId := c.GetString(":projectId")

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId)

	ctx.SendSDLog("AzureClusterController: POST cluster with project id: "+projectId, "error")

	beego.Info("AzureClusterController: StartCluster.")
	credentials := c.Ctx.Input.Header("Authorization")

	if credentials == "" ||
		strings.Contains(credentials, " ") ||
		strings.Contains(strings.ToLower(credentials), "bearer") ||
		strings.Contains(strings.ToLower(credentials), "azure") ||
		len(strings.Split(credentials, ":")) != 5 {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Authorization format should be '{id}:{key}:{tenant}:{subscription}:{region}'"}
		c.ServeJSON()
		return
	}

	var cluster azure.Cluster_Def

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	ctx.SendSDLog("AzureClusterController: Getting Cluster of project. "+projectId, "info")

	cluster, err := azure.GetCluster(projectId, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	if cluster.Status == "Cluster Created" {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": "cluster is already in running state"}
		c.ServeJSON()
		return
	}
	ctx.SendSDLog("AzureClusterController: Creating Cluster. "+cluster.Name, "info")

	go azure.DeployCluster(cluster, credentials, *ctx)

	c.Data["json"] = map[string]string{"msg": "cluster creation in progress"}
	c.ServeJSON()
}

// @Title Status
// @Description returns status of nodes
// @Param	Authorization	header	string	false	"{id}:{key}:{tenant}:{subscription}:{region}"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {object} azure.Cluster_Def
// @Failure 404 {"error": "project id is empty"}
// @Failure 401 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error"}
// @router /status/:projectId/ [get]
func (c *AzureClusterController) GetStatus() {

	projectId := c.GetString(":projectId")

	credentials := c.Ctx.Input.Header("Authorization")

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId)

	ctx.SendSDLog("AzureClusterController: FetchStatus.", "info")
	if credentials == "" ||
		strings.Contains(credentials, " ") ||
		strings.Contains(strings.ToLower(credentials), "bearer") ||
		strings.Contains(strings.ToLower(credentials), "azure") ||
		len(strings.Split(credentials, ":")) != 5 {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Authorization format should be '{id}:{key}:{tenant}:{subscription}:{region}'"}
		c.ServeJSON()
		return
	}

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	ctx.SendSDLog("AzureClusterController: Fetch Cluster Status of project. "+projectId, "info")

	cluster, err := azure.FetchStatus(credentials, projectId, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Terminate
// @Description terminates a  cluster
// @Param	Authorization	header	string	false	"{id}:{key}:{tenant}:{subscription}:{region}"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {"msg": "cluster terminated successfully"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 401 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error"}
// @router /terminate/:projectId/ [post]
func (c *AzureClusterController) TerminateCluster() {

	projectId := c.GetString(":projectId")

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId)

	ctx.SendSDLog("AzureClusterController: TerminateCluster.", "info")

	credentials := c.Ctx.Input.Header("Authorization")

	if credentials == "" ||
		strings.Contains(credentials, " ") ||
		strings.Contains(strings.ToLower(credentials), "bearer") ||
		strings.Contains(strings.ToLower(credentials), "azure") ||
		len(strings.Split(credentials, ":")) != 5 {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Authorization format should be '{id}:{key}:{tenant}:{subscription}:{region}'"}
		c.ServeJSON()
		return
	}

	var cluster azure.Cluster_Def

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	ctx.SendSDLog("AzureClusterController: Getting Cluster of project. "+projectId, "info")

	cluster, err := azure.GetCluster(projectId, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}
	ctx.SendSDLog("AzureClusterController: Terminating Cluster. "+cluster.Name, "info")

	go azure.TerminateCluster(cluster, credentials, *ctx)

	c.Data["json"] = map[string]string{"msg": "cluster termination is in progress"}
	c.ServeJSON()
}

// @Title SSHKeyPair
// @Description returns ssh key pairs
// @Success 200 {object} []string
// @Failure 500 {"error": "internal server error"}
// @router /sshkeys [get]
func (c *AzureClusterController) GetSSHKeys() {

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "")
	ctx.SendSDLog("AWSNetworkController: FetchExistingSSHKeys.", "info")

	keys, err := azure.GetAllSSHKeyPair(*ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = keys
	c.ServeJSON()
}
