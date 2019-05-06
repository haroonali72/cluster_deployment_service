package aws

import (
	"antelope/models/aws"
	"encoding/json"
	"github.com/astaxie/beego"
	"strings"
	"time"
)

// Operations about AWS cluster [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/cluster/{cloud}/]
type AWSClusterController struct {
	beego.Controller
}

// @Title Get
// @Description get cluster
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {object} aws.Cluster_Def
// @Failure 404 {"error": exception_message}
// @Failure 500 {"error": "internal server error"}
// @router /:projectId/ [get]
func (c *AWSClusterController) Get() {
	projectId := c.GetString(":projectId")

	beego.Info("AWSClusterController: Get cluster with project id: ", projectId)

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	cluster, err := aws.GetCluster(projectId)

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
// @Success 200 {object} []aws.Cluster_Def
// @Failure 500 {"error": "internal server error"}
// @router /all [get]
func (c *AWSClusterController) GetAll() {
	beego.Info("AWSClusterController: GetAll clusters.")

	clusters, err := aws.GetAllCluster()
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
// @Param	body	body 	aws.Cluster_Def		true	"body for cluster content"
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 409 {"error": "cluster against this project already exists"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router / [post]
func (c *AWSClusterController) Post() {
	var cluster aws.Cluster_Def
	json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	beego.Info("AWSClusterController: Post new cluster with name: ", cluster.Name)
	beego.Info("AWSClusterController: JSON Payload: ", cluster)

	cluster.CreationDate = time.Now()

	err := aws.CreateCluster(cluster)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.Ctx.Output.SetStatus(409)
			c.Data["json"] = map[string]string{"error": "cluster against this project id  already exists"}
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
// @Param	body	body 	aws.Cluster_Def	true	"body for cluster content"
// @Success 200 {"msg": "cluster updated successfully"}
// @Failure 404 {"error": "no cluster exists with this name"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router / [put]
func (c *AWSClusterController) Patch() {
	var cluster aws.Cluster_Def
	json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	beego.Info("AWSClusterController: Patch cluster with name: ", cluster.Name)
	beego.Info("AWSClusterController: JSON Payload: ", cluster)

	err := aws.UpdateCluster(cluster, true)
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
func (c *AWSClusterController) Delete() {
	id := c.GetString(":projectId")

	beego.Info("AWSClusterController: Delete cluster with project id: ", id)

	if id == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	cluster, err := aws.GetCluster(id)
	if err == nil && cluster.Status == "Cluster Created" {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error ," + "Cluster is in running state"}
		c.ServeJSON()
		return
	}
	err = aws.DeleteCluster(id)
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
// @Param	Authorization	header	string	false	"{access_key}:{secret_key}:{region}"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 404 {"error": "name is empty"}
// @Failure 401 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /start/:projectId [post]
func (c *AWSClusterController) StartCluster() {

	beego.Info("AWSNetworkController: StartCluster.")
	credentials := c.Ctx.Input.Header("Authorization")

	if credentials == "" ||
		strings.Contains(credentials, " ") ||
		strings.Contains(strings.ToLower(credentials), "bearer") ||
		strings.Contains(strings.ToLower(credentials), "aws") ||
		len(strings.Split(credentials, ":")) != 3 {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Authorization format should be '{access_key}:{secret_key}:{region}'"}
		c.ServeJSON()
		return
	}

	var cluster aws.Cluster_Def

	projectId := c.GetString(":projectId")

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	beego.Info("AWSClusterController: Getting Cluster of project. ", projectId)

	cluster, err := aws.GetCluster(projectId)

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
	beego.Info("AWSClusterController: Creating Cluster. ", cluster.Name)

	go aws.DeployCluster(cluster, credentials)

	c.Data["json"] = map[string]string{"msg": "cluster creation in progress"}
	c.ServeJSON()
}

// @Title Status
// @Description returns status of nodes
// @Param	Authorization	header	string	false	"{access_key}:{secret_key}:{region}"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {object} aws.Cluster_Def
// @Failure 404 {"error": "project id is empty"}
// @Failure 401 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /status/:projectId/ [get]
func (c *AWSClusterController) GetStatus() {

	beego.Info("AWSNetworkController: FetchStatus.")
	credentials := c.Ctx.Input.Header("Authorization")

	if credentials == "" ||
		strings.Contains(credentials, " ") ||
		strings.Contains(strings.ToLower(credentials), "bearer") ||
		strings.Contains(strings.ToLower(credentials), "aws") ||
		len(strings.Split(credentials, ":")) != 3 {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Authorization format should be '{access_key}:{secret_key}:{region}'"}
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

	beego.Info("AWSClusterController: Fetch Cluster Status of project. ", projectId)

	cluster, err := aws.FetchStatus(credentials, projectId)

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
// @Param	Authorization	header	string	false	"{access_key}:{secret_key}:{region}"
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {"msg": "cluster terminated successfully"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 401 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /terminate/:projectId/ [post]
func (c *AWSClusterController) TerminateCluster() {

	beego.Info("AWSNetworkController: TerminateCluster.")
	credentials := c.Ctx.Input.Header("Authorization")

	if credentials == "" ||
		strings.Contains(credentials, " ") ||
		strings.Contains(strings.ToLower(credentials), "bearer") ||
		strings.Contains(strings.ToLower(credentials), "aws") ||
		len(strings.Split(credentials, ":")) != 3 {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Authorization format should be '{access_key}:{secret_key}:{region}'"}
		c.ServeJSON()
		return
	}

	var cluster aws.Cluster_Def

	projectId := c.GetString(":projectId")

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	beego.Info("AWSClusterController: Getting Cluster of project. ", projectId)

	cluster, err := aws.GetCluster(projectId)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}
	beego.Info("AWSClusterController: Terminating Cluster. ", cluster.Name)

	go aws.TerminateCluster(cluster, credentials)

	c.Data["json"] = map[string]string{"msg": "cluster termination is in progress"}
	c.ServeJSON()
}

// @Title SSHKeyPair
// @Description returns ssh key pairs
// @Success 200 {object} []string
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /sshkeys [get]
func (c *AWSClusterController) GetSSHKeys() {

	beego.Info("AWSNetworkController: FetchExistingSSHKeys.")

	keys, err := aws.GetAllSSHKeyPair()

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = keys
	c.ServeJSON()
}

// @Title AwsAmis
// @Description returns aws ami details
// @Param	Authorization	header	string	false	"{access_key}:{secret_key}:{region}"
// @Param	amiId	path	string	true	"Id of the ami"
// @Success 200 {object} []*ec2.BlockDeviceMapping
// @Failure 401 {"error": "exception_message"}
// @Failure 404 {"error": "ami id is empty"}
// @Failure 500 {"error": "internal server error"}
// @router /amis/:amiId [get]
func (c *AWSClusterController) GetAMI() {

	beego.Info("AWSNetworkController: FetchExistingVpcs.")
	credentials := c.Ctx.Input.Header("Authorization")

	if credentials == "" ||
		strings.Contains(credentials, " ") ||
		strings.Contains(strings.ToLower(credentials), "bearer") ||
		strings.Contains(strings.ToLower(credentials), "aws") ||
		len(strings.Split(credentials, ":")) != 3 {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "Authorization format should be '{access_key}:{secret_key}:{region}'"}
		c.ServeJSON()
		return
	}
	amiId := c.GetString(":amiId")

	if amiId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "ami id is empty"}
		c.ServeJSON()
		return
	}
	beego.Info("AWSClusterController: Get Ami from AWS")

	keys, err := aws.GetAWSAmi(credentials, amiId)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = keys
	c.ServeJSON()
}
