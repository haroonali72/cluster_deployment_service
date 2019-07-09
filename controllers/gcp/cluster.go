package gcp

import (
	"antelope/models/gcp"
	"antelope/models/utils"
	"encoding/json"
	"github.com/astaxie/beego"
	"strings"
	"time"
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

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	cluster, err := gcp.GetCluster(projectId)
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
// @Success 200 {object} []gcp.Cluster_Def
// @Failure 500 {"error": "internal server error"}
// @router /all [get]
func (c *GcpClusterController) GetAll() {
	beego.Info("GcpClusterController: GetAll clusters.")

	clusters, err := gcp.GetAllCluster()
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
// @Param	body	body 	gcp.Cluster_Def		true	"body for cluster content"
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 409 {"error": "cluster against same project id already exists"}
// @Failure 500 {"error": "internal server error"}
// @router / [post]
func (c *GcpClusterController) Post() {
	var cluster gcp.Cluster_Def
	json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	beego.Info("GcpClusterController: Post new cluster with name: ", cluster.Name)
	beego.Info("GcpClusterController: JSON Payload: ", cluster)

	cluster.CreationDate = time.Now()

	err := gcp.CreateCluster(cluster)
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

	err := gcp.UpdateCluster(cluster, true)
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
func (c *GcpClusterController) Delete() {
	id := c.GetString(":projectId")

	beego.Info("GcpClusterController: Delete cluster with project id: ", id)

	if id == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "name is empty"}
		c.ServeJSON()
		return
	}
	cluster, err := gcp.GetCluster(id)
	if err == nil && cluster.Status == "Cluster Created" {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + "Cluster is in running state"}
		c.ServeJSON()
		return
	}
	err = gcp.DeleteCluster(id)
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
// @Param	Authorization	header	string	false	"base64 encoded service_account_json"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 400 {"error": "exception_message"}
// @Failure 401 {"error": "Authorization format should be 'base64 encoded service_account_json'"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 500 {"error": "internal server error"}
// @router /start/:projectId [post]
func (c *GcpClusterController) StartCluster() {
	beego.Info("GcpClusterController: StartCluster.")
	credentials := c.Ctx.Input.Header("Authorization")

	valid, _ := utils.IsValdidGcpCredentials(credentials)
	if !valid {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Authorization format should be 'base64 encoded service_account_json'"}
		c.ServeJSON()
		return
	}

	var cluster gcp.Cluster_Def

	projectId := c.GetString(":projectId")

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	beego.Info("GcpClusterController: Getting Cluster of project. ", projectId)

	cluster, err := gcp.GetCluster(projectId)

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
	beego.Info("GcpClusterController: Creating Cluster. ", cluster.Name)

	go gcp.DeployCluster(cluster, credentials)

	c.Data["json"] = map[string]string{"msg": "cluster creation in progress"}
	c.ServeJSON()
}

// @Title Status
// @Description returns status of nodes
// @Param	Authorization	header	string	false	"base64 encoded service_account_json"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {object} gcp.Cluster_Def
// @Failure 401 {"error": "Authorization format should be 'base64 encoded service_account_json'"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 500 {"error": "internal server error"}
// @router /status/:projectId/ [get]
func (c *GcpClusterController) GetStatus() {
	beego.Info("GcpClusterController: FetchStatus.")
	credentials := c.Ctx.Input.Header("Authorization")

	valid, _ := utils.IsValdidGcpCredentials(credentials)
	if !valid {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Authorization format should be 'base64 encoded service_account_json'"}
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

	beego.Info("GcpClusterController: Fetch Cluster Status of project. ", projectId)

	cluster, err := gcp.FetchStatus(credentials, projectId)

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
// @Param	Authorization	header	string	false	"base64 encoded service_account_json"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {"msg": "cluster terminated successfully"}
// @Failure 401 {"error": "Authorization format should be 'base64 encoded service_account_json'"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 500 {"error": "internal server error"}
// @router /terminate/:projectId/ [post]
func (c *GcpClusterController) TerminateCluster() {
	beego.Info("GcpClusterController: TerminateCluster.")
	credentials := c.Ctx.Input.Header("Authorization")

	valid, _ := utils.IsValdidGcpCredentials(credentials)
	if !valid {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Authorization format should be 'base64 encoded service_account_json'"}
		c.ServeJSON()
		return
	}

	var cluster gcp.Cluster_Def

	projectId := c.GetString(":projectId")

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	beego.Info("GcpClusterController: Getting Cluster of project. ", projectId)

	cluster, err := gcp.GetCluster(projectId)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}
	beego.Info("GcpClusterController: Terminating Cluster. ", cluster.Name)

	go gcp.TerminateCluster(cluster, credentials)

	c.Data["json"] = map[string]string{"msg": "cluster termination is in progress"}
	c.ServeJSON()
}
