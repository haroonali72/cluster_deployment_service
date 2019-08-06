package azure

import (
	"antelope/models/aws"
	"antelope/models/azure"
	"antelope/models/utils"
	"encoding/json"
	"github.com/asaskevich/govalidator"
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

	ctx := new(utils.Context)
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
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /all [get]
func (c *AzureClusterController) GetAll() {
	beego.Info("AzureClusterController: GetAll clusters.")

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "")

	clusters, err := azure.GetAllCluster(*ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
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
// @Success 400 {"msg": "error message"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router / [post]
func (c *AzureClusterController) Post() {
	var cluster azure.Cluster_Def
	json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, cluster.ProjectId)

	ctx.SendSDLog("AzureClusterController: Post new cluster with name: "+cluster.Name, "error ")

	cluster.CreationDate = time.Now()
	beego.Info(cluster.ResourceGroup)
	err := azure.GetNetwork(cluster.ProjectId, *ctx, cluster.ResourceGroup)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	res, err := govalidator.ValidateStruct(cluster)
	if !res || err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	err = azure.CreateCluster(cluster, *ctx)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "cluster against same project id already exists"}
			c.ServeJSON()
			return
		}
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
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
// @Failure 500 {"error": "internal server error <error msg>"}
// @router / [put]
func (c *AzureClusterController) Patch() {

	var cluster azure.Cluster_Def
	json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	ctx := new(utils.Context)
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
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
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
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /:projectId [delete]
func (c *AzureClusterController) Delete() {
	id := c.GetString(":projectId")

	ctx := new(utils.Context)
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
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"msg": "cluster deleted successfully"}
	c.ServeJSON()
}

// @Title Start
// @Description starts a  cluster
// @Param	projectId	path	string	true	"Id of the project"
// @Param	X-Profile-Id	header	string	false	""
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 400 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /start/:projectId [post]
func (c *AzureClusterController) StartCluster() {

	projectId := c.GetString(":projectId")

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId)

	ctx.SendSDLog("AzureClusterController: POST cluster with project id: "+projectId, "error")

	beego.Info("AzureClusterController: StartCluster.")
	profileId := c.Ctx.Input.Header("X-Profile-Id")
	region, err := azure.GetRegion(projectId, *ctx)

	azureProfile, err := azure.GetProfile(profileId, region, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
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

	cluster, err = azure.GetCluster(projectId, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
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

	go azure.DeployCluster(cluster, azureProfile, *ctx)

	c.Data["json"] = map[string]string{"msg": "cluster creation in progress"}
	c.ServeJSON()
}

// @Title Status
// @Description returns status of nodes
// @Param	projectId	path	string	true	"Id of the project"
// @Param	X-Profile-Id	header	string	false	""
// @Success 200 {object} azure.Cluster_Def
// @Failure 404 {"error": "project id is empty"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /status/:projectId/ [get]
func (c *AzureClusterController) GetStatus() {

	projectId := c.GetString(":projectId")

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	region, err := azure.GetRegion(projectId, *ctx)

	azureProfile, err := azure.GetProfile(profileId, region, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
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

	cluster, err := azure.FetchStatus(azureProfile, projectId, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = cluster
	c.ServeJSON()
}

// @Title Terminate
// @Description terminates a  cluster
// @Param	projectId	path	string	true	"Id of the project"
// @Param	X-Profile-Id	header	string	false	""
// @Success 200 {"msg": "cluster terminated successfully"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /terminate/:projectId/ [post]
func (c *AzureClusterController) TerminateCluster() {

	projectId := c.GetString(":projectId")

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId)

	ctx.SendSDLog("AzureClusterController: TerminateCluster.", "info")

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	region, err := azure.GetRegion(projectId, *ctx)

	azureProfile, err := azure.GetProfile(profileId, region, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
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

	cluster, err = azure.GetCluster(projectId, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendSDLog("AzureClusterController: Terminating Cluster. "+cluster.Name, "info")

	go azure.TerminateCluster(cluster, azureProfile, *ctx)

	c.Data["json"] = map[string]string{"msg": "cluster termination is in progress"}
	c.ServeJSON()
}

// @Title SSHKeyPair
// @Description returns ssh key pairs
// @Success 200 {object} []string
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /sshkeys [get]
func (c *AzureClusterController) GetSSHKeys() {

	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "")
	ctx.SendSDLog("AWSNetworkController: FetchExistingSSHKeys.", "info")

	keys, err := azure.GetAllSSHKeyPair(*ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = keys
	c.ServeJSON()
}