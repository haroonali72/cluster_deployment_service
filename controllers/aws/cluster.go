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
// @Success 200 {object} aws.Cluster
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
// @Success 200 {object} []aws.Cluster
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
// @Param	body	body	aws.Cluster	true	"body for cluster content"
// @Success 201 {"msg": "cluster created successfully"}
// @Failure 409 {"error": "cluster with same name already exists"}
// @Failure 500 {"error": "internal server error"}
// @router / [post]
func (c *AWSClusterController) Post() {
	var cluster aws.Cluster
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
// @Param	body	body	aws.Cluster	true	"body for cluster content"
// @Success 200 {"msg": "cluster updated successfully"}
// @Failure 404 {"error": "no cluster exists with this name"}
// @Failure 500 {"error": "internal server error"}
// @router / [put]
func (c *AWSClusterController) Patch() {
	var cluster aws.Cluster
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
// @Title Get AlL Machine Types
// @Description get all the clusters
// @Success 200 {object} []aws.Cluster
// @Failure 500 {"error": "internal server error"}
// @router /all [get]
func (c *AWSClusterController) GetAllMachineTypes() {
	beego.Info("AWSClusterController: GetAll machine types.")

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