package aws

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"strings"
	"antelope/models/aws"
)

// Operations about AWS cluster [BASE URL WILL BE CHANGED TO STANDARD URLs IN FUTURE e.g. /antelope/cluster/{cloud}/]
type AWSClusterController struct {
	beego.Controller
}

// @Title Get
// @Description get cluster
// @Param	name	path	string	true	"Name of the cluster"
// @Success 200 {object} aws.Cluster_Def
// @Failure 404 {"error": exception_message}
// @Failure 500 {"error": "internal server error"}
// @router /:name [get]
func (c *AWSClusterController) Get() {
	name := c.GetString(":name")

	beego.Info("AWSClusterController: Get cluster with name: ", name)

	if name == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "name is empty"}
		c.ServeJSON()
		return
	}

	cluster, err := aws.GetCluster(name)
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
// @Success 201 {"msg": "cluster created successfully"}
// @Failure 409 {"error": "cluster with same name already exists"}
// @Failure 500 {"error": "internal server error"}
// @router / [post]
func (c *AWSClusterController) Post() {
	var cluster aws.Cluster_Def
	json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	beego.Info("AWSClusterController: Post new cluster with name: ", cluster.Name)
	beego.Info("AWSClusterController: JSON Payload: ", cluster)

	err := aws.CreateCluster(cluster)
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
// @Param	body	body 	aws.Cluster_Def	true	"body for cluster content"
// @Success 200 {"msg": "cluster updated successfully"}
// @Failure 404 {"error": "no cluster exists with this name"}
// @Failure 500 {"error": "internal server error"}
// @router / [put]
func (c *AWSClusterController) Patch() {
	var cluster aws.Cluster_Def
	json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	beego.Info("AWSClusterController: Patch cluster with name: ", cluster.Name)
	beego.Info("AWSClusterController: JSON Payload: ", cluster)

	err := aws.UpdateCluster(cluster)
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
// @Param	name	path	string	true	"Name of the cluster"
// @Success 200 {"msg": "cluster deleted successfully"}
// @Failure 404 {"error": "name is empty"}
// @Failure 500 {"error": "internal server error"}
// @router /:name [delete]
func (c *AWSClusterController) Delete() {
	name := c.GetString(":name")

	beego.Info("AWSClusterController: Delete cluster with name: ", name)

	if name == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "name is empty"}
		c.ServeJSON()
		return
	}

	err := aws.DeleteCluster(name)
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
// @Param	name	path	string	true	"Name of the cluster"
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 404 {"error": "name is empty"}
// @Failure 401 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error"}
// @router /start/:name [post]
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

	name := c.GetString(":name")

	if name == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "name is empty"}
		c.ServeJSON()
		return
	}

	beego.Info("AWSClusterController: Getting Cluster. ", name)

	cluster , err :=aws.GetCluster(name)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}
	beego.Info("AWSClusterController: Creating Cluster. ", name)

	err = aws.DeployCluster(cluster,credentials)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"msg": "cluster created successfully"}
	c.ServeJSON()
}
// @Title Status
// @Description returns status of nodes
// @Param	Authorization	header	string	false	"{access_key}:{secret_key}:{region}"
// @Param	name	path	string	true	"Name of the cluster"
// @Success 200 {object} aws.Cluster_Def
// @Failure 404 {"error": "name is empty"}
// @Failure 401 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error"}
// @router /status/:name [get]
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

	name := c.GetString(":name")

	if name == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "name is empty"}
		c.ServeJSON()
		return
	}

	beego.Info("AWSClusterController: Fetch Cluster Status. ", name)

	cluster , err :=aws.FetchStatus(name,credentials)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = cluster
	c.ServeJSON()
}
// @Title SSHKeyPair
// @Description returns ssh key pairs
// @Param	Authorization	header	string	false	"{access_key}:{secret_key}:{region}"
// @Success 200 {object} aws.SSHKeyPair
// @Failure 401 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error"}
// @router /sshkeys [get]
func (c *AWSClusterController) GetSSHKeyPairs() {

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

	beego.Info("AWSClusterController: Get Keys ")

	keys , err :=aws.GetSSHKeyPair(credentials)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = keys
	c.ServeJSON()
}