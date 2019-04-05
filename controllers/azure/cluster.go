package azure

import (
	"antelope/models/azure"
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

	beego.Info("AzureClusterController: Get cluster with project id: ", projectId)

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	cluster, err := azure.GetCluster(projectId)
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

	clusters, err := azure.GetAllCluster()
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
// @Success 201 {"msg": "cluster created successfully"}
// @Failure 409 {"error": "cluster with same name already exists"}
// @Failure 500 {"error": "internal server error"}
// @router / [post]
func (c *AzureClusterController) Post() {
	var cluster azure.Cluster_Def
	json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	beego.Info("AzureClusterController: Post new cluster with name: ", cluster.Name)
	beego.Info("AzureClusterController: JSON Payload: ", cluster)

	cluster.CreationDate = time.Now()

	err := azure.CreateCluster(cluster)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "cluster with same project id already exists"}
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

	beego.Info("AzureClusterController: Patch cluster with name: ", cluster.Name)
	beego.Info("AzureClusterController: JSON Payload: ", cluster)

	err := azure.UpdateCluster(cluster)
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

	beego.Info("AzureClusterController: Delete cluster with project id: ", id)

	if id == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "name is empty"}
		c.ServeJSON()
		return
	}

	err := azure.DeleteCluster(id)
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
// @Failure 404 {"error": "name is empty"}
// @Failure 401 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error"}
// @router /start/:projectId [post]
func (c *AzureClusterController) StartCluster() {

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

	projectId := c.GetString(":projectId")

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	beego.Info("AzureClusterController: Getting Cluster of project. ", projectId)

	cluster, err := azure.GetCluster(projectId)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}
	beego.Info("AzureClusterController: Creating Cluster. ", cluster.Name)

	go azure.DeployCluster(cluster, credentials)

	c.Data["json"] = map[string]string{"msg": "cluster creation in progress"}
	c.ServeJSON()
}

// @Title Status
// @Description returns status of nodes
// @Param	Authorization	header	string	false	"{id}:{key}:{tenant}:{subscription}:{region}"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {object} azure.Cluster_Def
// @Failure 404 {"error": "name is empty"}
// @Failure 401 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error"}
// @router /status/:projectId/ [get]
func (c *AzureClusterController) GetStatus() {

	beego.Info("AzureClusterController: FetchStatus.")
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
	projectId := c.GetString(":projectId")

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "name is empty"}
		c.ServeJSON()
		return
	}

	beego.Info("AzureClusterController: Fetch Cluster Status of project. ", projectId)

	cluster, err := azure.FetchStatus(credentials, projectId)

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
// @Failure 404 {"error": "name is empty"}
// @Failure 401 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error"}
// @router /terminate/:projectId/ [post]
func (c *AzureClusterController) TerminateCluster() {

	beego.Info("AzureClusterController: TerminateCluster.")
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

	projectId := c.GetString(":projectId")

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	beego.Info("AzureClusterController: Getting Cluster of project. ", projectId)

	cluster, err := azure.GetCluster(projectId)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}
	beego.Info("AzureClusterController: Terminating Cluster. ", cluster.Name)

	go azure.TerminateCluster(cluster, credentials)

	c.Data["json"] = map[string]string{"msg": "cluster termination is in progress"}
	c.ServeJSON()
}

// @Title SSHKeyPair
// @Description returns ssh key pairs
// @Success 200 {object} aws.Key
// @Failure 401 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error"}
// @router /sshkeys [get]
func (c *AzureClusterController) GetSSHKeys() {

	beego.Info("AWSNetworkController: FetchExistingSSHKeys.")

	keys, err := azure.GetAllSSHKeyPair()

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = keys
	c.ServeJSON()
}
