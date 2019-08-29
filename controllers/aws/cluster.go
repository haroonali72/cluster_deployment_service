package aws

import (
	"antelope/models"
	"antelope/models/aws"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
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
// @Param	token	header	string	token ""
// @Success 200 {object} aws.Cluster_Def
// @Failure 404 {"error": exception_message}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /:projectId/ [get]
func (c *AWSClusterController) Get() {
	projectId := c.GetString(":projectId")

	token := c.Ctx.Input.Header("token")

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)

	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	allowed, err := rbac_athentication.Authenticate("cluster", projectId, "View", token, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//====================================================================================//

	ctx.SendLogs("AWSClusterController: Get cluster with project id: "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Log)

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
// @Param	token	header	string	token ""
// @Success 200 {object} []aws.Cluster_Def
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /all [get]
func (c *AWSClusterController) GetAll() {
	token := c.Ctx.Input.Header("token")

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	err, data := rbac_athentication.GetAllAuthenticate("cluster", userInfo.CompanyId, token, models.AWS, *ctx)

	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}

	//====================================================================================//

	ctx.SendLogs("AWSClusterController: GetAll clusters.", models.LOGGING_LEVEL_INFO, models.Backend_Log)

	clusters, err := aws.GetAllCluster(*ctx, data)
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
// @Param	body	body 	aws.Cluster_Def		true	"body for cluster content"
// @Param	token	header	string	token ""
// @Success 200 {"msg": "cluster created successfully"}
// @Success 400 {"msg": "error msg"}
// @Failure 409 {"error": "cluster against this project already exists"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router / [post]
func (c *AWSClusterController) Post() {

	var cluster aws.Cluster_Def
	json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	cluster.CreationDate = time.Now()

	token := c.Ctx.Input.Header("token")

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, cluster.ProjectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	allowed, err := rbac_athentication.Authenticate("cluster", cluster.ProjectId, "Create", token, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//=============================================================================//

	ctx.SendLogs("AWSClusterController: Post new cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Log)

	res, err := govalidator.ValidateStruct(cluster)
	if !res || err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	err = aws.GetNetwork(token, cluster.ProjectId, *ctx)
	if err != nil {
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
// @Param	token	header	string	token ""
// @Param	body	body 	aws.Cluster_Def	true	"body for cluster content"
// @Success 200 {"msg": "cluster updated successfully"}
// @Failure 404 {"error": "no cluster exists with this name"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router / [put]
func (c *AWSClusterController) Patch() {
	var cluster aws.Cluster_Def
	json.Unmarshal(c.Ctx.Input.RequestBody, &cluster)

	token := c.Ctx.Input.Header("token")

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, cluster.ProjectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	allowed, err := rbac_athentication.Authenticate("cluster", cluster.ProjectId, "Update", token, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//=============================================================================//

	ctx.SendLogs("AWSClusterController: Patch cluster with name: "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Log)

	err = aws.UpdateCluster(cluster, true, *ctx)
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
// @Param	token	header	string	token ""
// @Param	projectId	path	string	true	"project id of the cluster"
// @Success 200 {"msg": "cluster deleted successfully"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /:projectId [delete]
func (c *AWSClusterController) Delete() {
	id := c.GetString(":projectId")

	token := c.Ctx.Input.Header("token")

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, id, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	allowed, err := rbac_athentication.Authenticate("cluster", id, "Delete", token, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//=============================================================================//

	ctx.SendLogs("AWSClusterController: Delete cluster with project id: "+id, models.LOGGING_LEVEL_INFO, models.Backend_Log)

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
// @Param	token	header	string	token ""
// @Param	X-Profile-Id	header	string	profileId	""
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {"msg": "cluster created successfully"}
// @Failure 404 {"error": "name is empty"}
// @Failure 40 {"error": "exception_message"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /start/:projectId [post]
func (c *AWSClusterController) StartCluster() {

	projectId := c.GetString(":projectId")

	token := c.Ctx.Input.Header("token")

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	allowed, err := rbac_athentication.Authenticate("cluster", projectId, "Start", token, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//=============================================================================//

	ctx.SendLogs("AWSNetworkController: StartCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Log)

	profileId := c.Ctx.Input.Header("X-Profile-Id")

	var cluster aws.Cluster_Def

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}

	ctx.SendLogs("AWSClusterController: Getting Cluster of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Log)

	cluster, err = aws.GetCluster(projectId, *ctx)

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
	region, err := aws.GetRegion(token, projectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}
	awsProfile, err := aws.GetProfile(profileId, region, token, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("AWSClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Log)

	go aws.DeployCluster(cluster, awsProfile.Profile, *ctx, userInfo.CompanyId, token)

	c.Data["json"] = map[string]string{"msg": "cluster creation in progress"}
	c.ServeJSON()
}

// @Title Status
// @Description returns status of nodes
// @Param	token	header	string	token ""
// @Param	X-Profile-Id	header	string	profileId	""
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {object} aws.Cluster_Def
// @Failure 404 {"error": "project id is empty"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /status/:projectId/ [get]
func (c *AWSClusterController) GetStatus() {

	projectId := c.GetString(":projectId")
	token := c.Ctx.Input.Header("token")

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	allowed, err := rbac_athentication.Authenticate("cluster", projectId, "View", token, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {

		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//=============================================================================//
	ctx.SendLogs("AWSNetworkController: FetchStatus.", models.LOGGING_LEVEL_INFO, models.Backend_Log)

	profileId := c.Ctx.Input.Header("X-Profile-Id")

	if projectId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "project id is empty"}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("AWSClusterController: Fetch Cluster Status of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Log)

	region, err := aws.GetRegion(token, projectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}
	beego.Info("********" + token + "********")
	awsProfile, err := aws.GetProfile(profileId, region, token, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	cluster, err := aws.FetchStatus(awsProfile, projectId, *ctx, userInfo.CompanyId, token)

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
// @Param	X-Profile-Id header	X-Profile-Id	string	profileId	""
// @Param	token	header	string	token ""
// @Param	projectId	path	string	true	"Id of the project"
// @Success 200 {"msg": "cluster terminated successfully"}
// @Failure 404 {"error": "project id is empty"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /terminate/:projectId/ [post]
func (c *AWSClusterController) TerminateCluster() {

	projectId := c.GetString(":projectId")

	token := c.Ctx.Input.Header("token")

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, projectId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	allowed, err := rbac_athentication.Authenticate("cluster", projectId, "Terminate", token, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//=============================================================================//
	ctx.SendLogs("AWSNetworkController: TerminateCluster.", models.LOGGING_LEVEL_INFO, models.Backend_Log)
	profileId := c.Ctx.Input.Header("X-Profile-Id")
	region, err := aws.GetRegion(token, projectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	awsProfile, err := aws.GetProfile(profileId, region, token, *ctx)

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
	ctx.SendLogs("AWSClusterController: Getting Cluster of project. "+projectId, models.LOGGING_LEVEL_INFO, models.Backend_Log)
	cluster, err = aws.GetCluster(projectId, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("AWSClusterController: Terminating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Log)

	go aws.TerminateCluster(cluster, awsProfile, *ctx, userInfo.CompanyId)

	c.Data["json"] = map[string]string{"msg": "cluster termination is in progress"}
	c.ServeJSON()
}

// @Title SSHKeyPair
// @Description returns ssh key pairs
// @Param	token	header	string	token ""
// @Success 200 {object} []string
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /sshkeys [get]
func (c *AWSClusterController) GetSSHKeys() {

	token := c.Ctx.Input.Header("token")

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//

	//=============================================================================//
	ctx.SendLogs("AWSNetworkController: FetchExistingSSHKeys.", models.LOGGING_LEVEL_INFO, models.Backend_Log)
	keys, err := aws.GetAllSSHKeyPair(*ctx, token)

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
// @Param	X-Profile-Id	header	string	profileId	""
// @Param	token	header	string	token ""
// @Param	X-Region	header	string	false	""
// @Param	amiId	path	string	true	"Id of the ami"
// @Success 200 {object} []*ec2.BlockDeviceMapping
// @Failure 404 {"error": "ami id is empty"}
// @Failure 500 {"error": "internal server error"}
// @router /amis/:amiId [get]
func (c *AWSClusterController) GetAMI() {

	token := c.Ctx.Input.Header("token")

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	//err, _ = rbac_athentication.GetAllAuthenticate("cluster",userInfo.CompanyId, token, *ctx)
	//if err != nil {
	//	beego.Error(err.Error())
	//	c.Ctx.Output.SetStatus(400)
	//	c.Data["json"] = map[string]string{"error": err.Error()}
	//	c.ServeJSON()
	//	return
	//}

	//=============================================================================//
	ctx.SendLogs("AWSClusterController: FetchAMIs.", models.LOGGING_LEVEL_INFO, models.Backend_Log)

	profileId := c.Ctx.Input.Header("X-Profile-Id")
	region := c.Ctx.Input.Header("X-Region")

	awsProfile, err := aws.GetProfile(profileId, region, token, *ctx)

	amiId := c.GetString(":amiId")

	if amiId == "" {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = map[string]string{"error": "ami id is empty"}
		c.ServeJSON()
		return
	}
	ctx.SendLogs("AWSClusterController: Get Ami from AWS", models.LOGGING_LEVEL_INFO, models.Backend_Log)

	keys, err := aws.GetAWSAmi(awsProfile, amiId, *ctx, token)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	c.Data["json"] = keys
	c.ServeJSON()
}

// @Title EnableScaling
// @Description enables autoscaling
// @Param	X-Profile-Id	header	string	profileId	""
// @Param	projectId	path	string	true	"Id of the project"
// @Param	token	header	string	token ""
// @Param	body	body 	aws.AutoScaling	true	"body for cluster content"
// @Success 200 {object} aws.AutoScaling
// @Success 200 {"msg": "cluster autoscaled successfully"}
// @Failure 500 {"error": "internal server error <error msg>"}
// @router /enablescaling/:projectId/ [post]
func (c *AWSClusterController) EnableAutoScaling() {

	projectId := c.GetString(":projectId")

	token := c.Ctx.Input.Header("token")

	userInfo, err := rbac_athentication.GetInfo(token)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger(c.Ctx.Request.Host, "GET", c.Ctx.Request.RequestURI, "", userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	allowed, err := rbac_athentication.Authenticate("cluster", projectId, "Start", token, *ctx)
	if err != nil {
		beego.Error(err.Error())
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = map[string]string{"error": err.Error()}
		c.ServeJSON()
		return
	}
	if !allowed {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = map[string]string{"error": "User is unauthorized to perform this action"}
		c.ServeJSON()
		return
	}

	//=============================================================================//
	profileId := c.Ctx.Input.Header("X-Profile-Id")
	region, err := aws.GetRegion(token, projectId, *ctx)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	awsProfile, err := aws.GetProfile(profileId, region, token, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}
	cluster, err := aws.GetCluster(projectId, *ctx)

	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}

	go aws.EnableScaling(awsProfile, cluster, *ctx, token)

	/*err = aws.EnableScaling(awsProfile, cluster, *ctx, token)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = map[string]string{"error": "internal server error " + err.Error()}
		c.ServeJSON()
		return
	}
	*/
	c.Data["json"] = map[string]string{"msg": "cluster autoscaled successfully"}
	c.ServeJSON()
}
