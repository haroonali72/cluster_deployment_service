package azure

import (
	"github.com/astaxie/beego"
	"antelope/models/azure"
	"strings"
	"encoding/json"
	"time"
)

// Operations about azure cluster [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/cluster/{cloud}/]
type AzureClusterController struct {
	beego.Controller
}

// @Title Get
// @Description get cluster
// @Param	environmentId	path	string	true	"Id of the environment"
// @Success 200 {object} azure.Cluster_Def
// @Failure 404 {"error": exception_message}
// @Failure 500 {"error": "internal server error"}
// @router /:environmentId/ [get]
func (c *AzureClusterController) Get() {
	envId := c.GetString(":environmentId")


	beego.Info("AzureClusterController: Get cluster with environment id: ", envId)

	if envId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "environment id is empty"}
		c.ServeJSON()
		return
	}

	cluster, err := azure.GetCluster(envId)
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
			c.Data["json"] = map[string]string{"error": "cluster with same name already exists"}
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
// @Param	environmentId	path	string	true	"Environment id of the cluster"
// @Success 200 {"msg": "cluster deleted successfully"}
// @Failure 404 {"error": "environment id is empty"}
// @Failure 500 {"error": "internal server error"}
// @router /:environmentId [delete]
func (c *AzureClusterController) Delete() {
	id := c.GetString(":environmentId")

	beego.Info("AzureClusterController: Delete cluster with environment id: ", id)

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
// @Param	Authorization	header	string	false	"{access_key}:{secret_key}:{region}"
// @Param	environmentId	path	string	true	"Id of the environment"
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 404 {"error": "name is empty"}
// @Failure 401 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error"}
// @router /start/:environmentId [post]
func (c *AzureClusterController) StartCluster() {

	beego.Info("AzureClusterController: StartCluster.")
	credentials := c.Ctx.Input.Header("Authorization")

	if credentials == "" ||
		strings.Contains(credentials, " ") ||
		strings.Contains(strings.ToLower(credentials), "bearer") ||
		strings.Contains(strings.ToLower(credentials), "azure") ||
		len(strings.Split(credentials, ":")) != 3 {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Authorization format should be '{access_key}:{secret_key}:{region}'"}
		c.ServeJSON()
		return
	}


	var cluster azure.Cluster_Def

	envId := c.GetString(":environmentId")


	if envId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "environment id is empty"}
		c.ServeJSON()
		return
	}

	beego.Info("AzureClusterController: Getting Cluster of environment. ", envId)

	cluster , err :=azure.GetCluster(envId)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}
	beego.Info("AzureClusterController: Creating Cluster. ", cluster.Name)

	go azure.DeployCluster(cluster,credentials)

	c.Data["json"] = map[string]string{"msg": "cluster creation in progress"}
	c.ServeJSON()
}

// @Title Status
// @Description returns status of nodes
// @Param	Authorization	header	string	false	"{access_key}:{secret_key}:{region}"
// @Param	environmentId	path	string	true	"Id of the environment"
// @Success 200 {object} azure.Cluster_Def
// @Failure 404 {"error": "name is empty"}
// @Failure 401 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error"}
// @router /status/:environmentId/ [get]
func (c *AzureClusterController) GetStatus() {

	beego.Info("AzureClusterController: FetchStatus.")
	credentials := c.Ctx.Input.Header("Authorization")

	if credentials == "" ||
		strings.Contains(credentials, " ") ||
		strings.Contains(strings.ToLower(credentials), "bearer") ||
		strings.Contains(strings.ToLower(credentials), "azure") ||
		len(strings.Split(credentials, ":")) != 3 {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Authorization format should be '{access_key}:{secret_key}:{region}'"}
		c.ServeJSON()
		return
	}
	envId := c.GetString(":environmentId")

	if envId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "name is empty"}
		c.ServeJSON()
		return
	}

	beego.Info("AzureClusterController: Fetch Cluster Status of environment. ", envId)

	cluster , err :=azure.FetchStatus(credentials,envId)

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
// @Param	Authorization	header	string	false	"{access_key}:{secret_key}:{region}"
// @Param	environmentId	path	string	true	"Id of the environment"
// @Success 200 {"msg": "cluster terminated successfully"}
// @Failure 404 {"error": "name is empty"}
// @Failure 401 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error"}
// @router /terminate/:environmentId/ [post]
func (c *AzureClusterController) TerminateCluster() {

	beego.Info("AzureClusterController: TerminateCluster.")
	credentials := c.Ctx.Input.Header("Authorization")

	if credentials == "" ||
		strings.Contains(credentials, " ") ||
		strings.Contains(strings.ToLower(credentials), "bearer") ||
		strings.Contains(strings.ToLower(credentials), "azure") ||
		len(strings.Split(credentials, ":")) != 3 {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Authorization format should be '{access_key}:{secret_key}:{region}'"}
		c.ServeJSON()
		return
	}


	var cluster azure.Cluster_Def

	envId := c.GetString(":environmentId")

	if envId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "environment id is empty"}
		c.ServeJSON()
		return
	}

	beego.Info("AzureClusterController: Getting Cluster of environment. ", envId)

	cluster , err :=azure.GetCluster(envId)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}
	beego.Info("AzureClusterController: Terminating Cluster. ", cluster.Name)

	go azure.TerminateCluster(cluster,credentials)

	c.Data["json"] = map[string]string{"msg": "cluster termination is in progress"}
	c.ServeJSON()
}