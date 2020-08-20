package queue

import (
	"antelope/models"
	"antelope/models/aws"
	"antelope/models/azure"
	"antelope/models/do"
	"antelope/models/gcp"
	rbac_athentication "antelope/models/rbac_authentication"

	"antelope/models/utils"
	"strings"
)

func AWSClusterStartHelper(task WorkSchema, infraData Infrastructure) {

	if task.Token == "" {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Token is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	_, userInfo, err := rbac_athentication.GetInfo(task.Token)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger("", "POST", "", task.InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	_, allowed, err := rbac_athentication.Authenticate(models.AWS, "cluster", task.InfraId, "Start", task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	if !allowed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "User is not allowed to perform this action",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	var cluster aws.Cluster_Def

	ctx.SendLogs("AWSClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = aws.GetCluster(task.InfraId, userInfo.CompanyId, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	if cluster.Status == models.ClusterCreated {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in deployed state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return

	} else if cluster.Status == models.Deploying {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == models.Terminating {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't deploy. Cluster is in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == models.ClusterTerminationFailed {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "cluster is in invalid state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}

	region, err := aws.GetRegion(task.Token, task.InfraId, *ctx)
	if err != nil {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	_, awsProfile, err := aws.GetProfile(infraData.infrastructureData.ProfileId, region, task.Token, *ctx)
	if err != nil {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}

	cluster.Status = models.Deploying
	err = aws.UpdateCluster(cluster, false, *ctx)
	if err != nil {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}

	ctx.SendLogs("AWSClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	aws.DeployCluster(cluster, awsProfile.Profile, *ctx, userInfo.CompanyId, task.Token)

}
func AWSClusterTerminateHelper(task WorkSchema, infraData Infrastructure) {

	if task.Token == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Token is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure id is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	_, userInfo, err := rbac_athentication.GetInfo(task.Token)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger("", "POST", "", task.InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	_, allowed, err := rbac_athentication.Authenticate(models.AWS, "cluster", task.InfraId, "Terminate", task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	if !allowed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "User is not allowed to perform this action",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	var cluster aws.Cluster_Def

	ctx.SendLogs("AWSClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = aws.GetCluster(task.InfraId, userInfo.CompanyId, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	if cluster.Status == models.Deploying {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't Terminate. Cluster is in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	} else if cluster.Status == models.Terminating {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	} else if cluster.Status == models.ClusterTerminated {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "cluster is already in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return

	} else if cluster.Status == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster state is invalid for termination",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	} else if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't Terminate an undeployed  cluster",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	region, err := aws.GetRegion(task.Token, task.InfraId, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	_, awsProfile, err := aws.GetProfile(infraData.infrastructureData.ProfileId, region, task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	cluster.Status = models.Terminating
	err = aws.UpdateCluster(cluster, false, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	ctx.SendLogs("AWSClusterController: Terminating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	go aws.TerminateCluster(cluster, awsProfile, *ctx, userInfo.CompanyId, task.Token)

}
func AzureClusterStartHelper(task WorkSchema, infraData Infrastructure) {

	if task.Token == "" {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Token is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	_, userInfo, err := rbac_athentication.GetInfo(task.Token)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger("", "POST", "", task.InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	_, allowed, err := rbac_athentication.Authenticate(models.Azure, "cluster", task.InfraId, "Start", task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	if !allowed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "User is not allowed to perform this action",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	var cluster azure.Cluster_Def

	ctx.SendLogs("AZureClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = azure.GetCluster(task.InfraId, userInfo.CompanyId, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	if cluster.Status == models.ClusterCreated {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in deployed state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return

	} else if cluster.Status == models.Deploying {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == models.Terminating {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't deploy. Cluster is in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == models.ClusterTerminationFailed {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "cluster is in invalid state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}

	region, err := azure.GetRegion(task.Token, task.InfraId, *ctx)
	if err != nil {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	_, azureProfile, err := azure.GetProfile(infraData.infrastructureData.ProfileId, region, task.Token, *ctx)
	if err != nil {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}

	cluster.Status = models.Deploying
	err = azure.UpdateCluster(cluster, false, *ctx)
	if err != nil {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}

	ctx.SendLogs("AWSClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	azure.DeployCluster(cluster, azureProfile, *ctx, userInfo.CompanyId, task.Token)
}
func AzureClusterTerminateHelper(task WorkSchema, infraData Infrastructure) {

	if task.Token == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Token is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure id is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	_, userInfo, err := rbac_athentication.GetInfo(task.Token)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger("", "POST", "", task.InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	_, allowed, err := rbac_athentication.Authenticate(models.Azure, "cluster", task.InfraId, "Terminate", task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	if !allowed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "User is not allowed to perform this action",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	var cluster azure.Cluster_Def

	ctx.SendLogs("AZUREClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = azure.GetCluster(task.InfraId, userInfo.CompanyId, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	if cluster.Status == models.Deploying {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't Terminate. Cluster is in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	} else if cluster.Status == models.Terminating {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	} else if cluster.Status == models.ClusterTerminated {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "cluster is already in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return

	} else if cluster.Status == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster state is invalid for termination",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	} else if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't Terminate an undeployed  cluster",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	region, err := azure.GetRegion(task.Token, task.InfraId, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	_, azureProfile, err := azure.GetProfile(infraData.infrastructureData.ProfileId, region, task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	cluster.Status = models.Terminating
	err = azure.UpdateCluster(cluster, false, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	ctx.SendLogs("AWSClusterController: Terminating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	go azure.TerminateCluster(cluster, azureProfile, *ctx, userInfo.CompanyId, task.Token)
}
func GCPClusterStartHelper(task WorkSchema, infraData Infrastructure) {

	if task.Token == "" {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Token is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	_, userInfo, err := rbac_athentication.GetInfo(task.Token)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger("", "POST", "", task.InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	_, allowed, err := rbac_athentication.Authenticate(models.GCP, "cluster", task.InfraId, "Start", task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	if !allowed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "User is not allowed to perform this action",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	var cluster gcp.Cluster_Def

	ctx.SendLogs("GCPClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = gcp.GetCluster(task.InfraId, userInfo.CompanyId, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	if cluster.Status == models.ClusterCreated {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in deployed state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return

	} else if cluster.Status == models.Deploying {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == models.Terminating {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't deploy. Cluster is in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == models.ClusterTerminationFailed {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "cluster is in invalid state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}

	region, zone, err := gcp.GetRegion(task.Token, task.InfraId, *ctx)
	if err != nil {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	isValid, credentials := gcp.IsValidGcpCredentials(infraData.infrastructureData.ProfileId, region, task.Token, zone, *ctx)
	if !isValid {

		return
	}
	cluster.Status = models.Deploying
	err = gcp.UpdateCluster(cluster, false, *ctx)
	if err != nil {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}

	ctx.SendLogs("GCPClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	gcp.DeployCluster(cluster, credentials, userInfo.CompanyId, task.Token, *ctx)
}
func GCPClusterTerminateHelper(task WorkSchema, infraData Infrastructure) {

	if task.Token == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Token is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure id is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	_, userInfo, err := rbac_athentication.GetInfo(task.Token)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger("", "POST", "", task.InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	_, allowed, err := rbac_athentication.Authenticate(models.GCP, "cluster", task.InfraId, "Terminate", task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	if !allowed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "User is not allowed to perform this action",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	var cluster gcp.Cluster_Def

	ctx.SendLogs("GCPClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = gcp.GetCluster(task.InfraId, userInfo.CompanyId, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	if cluster.Status == models.Deploying {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't Terminate. Cluster is in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	} else if cluster.Status == models.Terminating {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	} else if cluster.Status == models.ClusterTerminated {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "cluster is already in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return

	} else if cluster.Status == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster state is invalid for termination",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	} else if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't Terminate an undeployed  cluster",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	region, zone, err := gcp.GetRegion(task.Token, task.InfraId, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	isValid, credentials := gcp.IsValidGcpCredentials(infraData.infrastructureData.ProfileId, region, task.Token, zone, *ctx)
	if !isValid {
		return
		return
	}
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	cluster.Status = models.Terminating
	err = gcp.UpdateCluster(cluster, false, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	ctx.SendLogs("GCPClusterController: Terminating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	go gcp.TerminateCluster(cluster, credentials, task.Token, userInfo.CompanyId, *ctx)
}
func DOClusterStartHelper(task WorkSchema, infraData Infrastructure) {

	if task.Token == "" {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Token is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	_, userInfo, err := rbac_athentication.GetInfo(task.Token)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger("", "POST", "", task.InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	_, allowed, err := rbac_athentication.Authenticate(models.DO, "cluster", task.InfraId, "Start", task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	if !allowed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "User is not allowed to perform this action",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	var cluster do.Cluster_Def

	ctx.SendLogs("DOClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = do.GetCluster(task.InfraId, userInfo.CompanyId, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	if cluster.Status == models.ClusterCreated {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in deployed state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return

	} else if cluster.Status == models.Deploying {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == models.Terminating {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't deploy. Cluster is in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == models.ClusterTerminationFailed {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "cluster is in invalid state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}

	region, err := do.GetRegion(task.Token, *ctx)
	if err != nil {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}

	_, doProfile, err := do.GetProfile(infraData.infrastructureData.ProfileId, region, task.Token, *ctx)
	if err != nil {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}

	cluster.Status = models.Deploying
	err = do.UpdateCluster(cluster, false, *ctx)
	if err != nil {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}

	ctx.SendLogs("DOClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	do.DeployCluster(cluster, doProfile.Profile, *ctx, userInfo.CompanyId, task.Token)
}
func DOClusterTerminateHelper(task WorkSchema, infraData Infrastructure) {

	if task.Token == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Token is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure id is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}
	_, userInfo, err := rbac_athentication.GetInfo(task.Token)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger("", "POST", "", task.InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	_, allowed, err := rbac_athentication.Authenticate(models.DO, "cluster", task.InfraId, "Terminate", task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	if !allowed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "User is not allowed to perform this action",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	var cluster do.Cluster_Def

	ctx.SendLogs("DOClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = do.GetCluster(task.InfraId, userInfo.CompanyId, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	if cluster.Status == models.Deploying {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't Terminate. Cluster is in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	} else if cluster.Status == models.Terminating {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	} else if cluster.Status == models.ClusterTerminated {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "cluster is already in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return

	} else if cluster.Status == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster state is invalid for termination",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	} else if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't Terminate an undeployed  cluster",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	region, err := do.GetRegion(task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	_, doProfile, err := do.GetProfile(infraData.infrastructureData.ProfileId, region, task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	cluster.Status = models.Terminating
	err = do.UpdateCluster(cluster, false, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

	ctx.SendLogs("DOClusterController: Terminating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	go do.TerminateCluster(cluster, doProfile, *ctx, userInfo.CompanyId, task.Token)
}
