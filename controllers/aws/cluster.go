package aws

import (
	"antelope/models/aws"
	"antelope/models/logging"
	"encoding/json"
	"github.com/asaskevich/govalidator"
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

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId)
	ctx.SendSDLog("AWSClusterController: Get cluster with project id: "+projectId, "info")

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	cluster, err := aws.GetCluster(projectId, *ctx)

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
	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "")

	ctx.SendSDLog("AWSClusterController: GetAll clusters.", "info")

	clusters, err := aws.GetAllCluster(*ctx)
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

	cluster.CreationDate = time.Now()

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, cluster.ProjectId)

	ctx.SendSDLog("AWSClusterController: Post new cluster with name: "+cluster.Name, "info")

	res, err := govalidator.ValidateStruct(cluster)
	if !res || err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	err = aws.CreateCluster(cluster, *ctx)
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

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "PUT", c.Ctx.Request.RequestURI, cluster.ProjectId)

	ctx.SendSDLog("AWSClusterController: Patch cluster with name: "+cluster.Name, "info")

	err := aws.UpdateCluster(cluster, true, *ctx)
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

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "DELETE", c.Ctx.Request.RequestURI, id)

	ctx.SendSDLog("AWSClusterController: Delete cluster with project id: "+id, "info")

	if id == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	cluster, err := aws.GetCluster(id, *ctx)
	if err == nil && cluster.Status == "Cluster Created" {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error ," + "Cluster is in running state"}
		c.ServeJSON()
		return
	}
	err = aws.DeleteCluster(id, *ctx)
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

	projectId := c.GetString(":projectId")

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId)

	ctx.SendSDLog("AWSNetworkController: StartCluster.", "info")

	profileId := c.Ctx.Input.Header("X-Profile-Id")

	var cluster aws.Cluster_Def

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	ctx.SendSDLog("AWSClusterController: Getting Cluster of project. "+projectId, "info")
	cluster, err := aws.GetCluster(projectId, *ctx)

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
	region, err := aws.GetRegion(projectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}
	awsProfile, err := aws.GetProfile(profileId, region, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendSDLog("AWSClusterController: Creating Cluster. "+cluster.Name, "info")

	go aws.DeployCluster(cluster, awsProfile.Profile, *ctx)

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

	projectId := c.GetString(":projectId")
	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId)

	ctx.SendSDLog("AWSNetworkController: FetchStatus.", "info")

	profileId := c.Ctx.Input.Header("X-Profile-Id")

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	ctx.SendSDLog("AWSClusterController: Fetch Cluster Status of project. "+projectId, "info")
	region, err := aws.GetRegion(projectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}
	awsProfile, err := aws.GetProfile(profileId, region, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	cluster, err := aws.FetchStatus(awsProfile, projectId, *ctx)

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

	projectId := c.GetString(":projectId")

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, projectId)

	ctx.SendSDLog("AWSNetworkController: TerminateCluster.", "info")

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	region, err := aws.GetRegion(projectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	awsProfile, err := aws.GetProfile(profileId, region, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	var cluster aws.Cluster_Def

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	ctx.SendSDLog("AWSClusterController: Getting Cluster of project. "+projectId, "info")

	cluster, err = aws.GetCluster(projectId, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendSDLog("AWSClusterController: Terminating Cluster. "+cluster.Name, "info")

	go aws.TerminateCluster(cluster, awsProfile, *ctx)

	c.Data["json"] = map[string]string{"msg": "cluster termination is in progress"}
	c.ServeJSON()
}

// @Title SSHKeyPair
// @Description returns ssh key pairs
// @Success 200 {object} []string
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /sshkeys [get]
func (c *AWSClusterController) GetSSHKeys() {

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "")

	ctx.SendSDLog("AWSNetworkController: FetchExistingSSHKeys.", "info")

	keys, err := aws.GetAllSSHKeyPair(*ctx)

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

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "")

	ctx.SendSDLog("AWSClusterController: FetchAMIs.", "info")

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	region := c.Ctx.Input.Header("X-Region")

	awsProfile, err := aws.GetProfile(profileId, region, *ctx)

	amiId := c.GetString(":amiId")

	if amiId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "ami id is empty"}
		c.ServeJSON()
		return
	}
	ctx.SendSDLog("AWSClusterController: Get Ami from AWS", "info")

	keys, err := aws.GetAWSAmi(awsProfile, amiId, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = keys
	c.ServeJSON()
}

// @Title EnableScaling
// @Description enables autoscaling
// @Param	Authorization	header	string	false	"{access_key}:{secret_key}:{region}"
// @Param	body	body 	aws.AutoScaling	true	"body for cluster content"
// @Success 200 {object} aws.AutoScaling
// @Success 200 {"msg": "cluster autoscaled successfully"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /enablescaling/:projectId/ [post]
func (c *AWSClusterController) EnableAutoScaling() {

	ctx := new(logging.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "POST", c.Ctx.Request.RequestURI, "")
	ctx.SendSDLog("AWSClusterController: EnableScaling.", "info")

	credentials := c.Ctx.Input.Header("Authorization")

	/*var scaler aws.AutoScaling
	json.Unmarshal(c.Ctx.Input.RequestBody, &scaler)*/

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
	cluster, err := aws.GetCluster(projectId, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	err = aws.EnableScaling(credentials, cluster, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]string{"msg": "cluster autoscaled successfully"}
	c.ServeJSON()
}
