package queue

import (
	"antelope/models"
	"antelope/models/aks"
	"antelope/models/aws"
	"antelope/models/azure"
	"antelope/models/do"
	"antelope/models/doks"
	"antelope/models/eks"
	"antelope/models/gcp"
	"antelope/models/gke"
	"antelope/models/iks"
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
			Status:  true,
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
			Action:  models.Terminate,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure id is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action:  models.Terminate,
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
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}
	if !allowed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "User is not allowed to perform this action",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	if cluster.Status == models.Deploying {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't Terminate. Cluster is in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.Status == models.Terminating {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.Status == models.ClusterTerminated {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "cluster is already in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return

	} else if cluster.Status == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Cluster state is invalid for termination",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Can't Terminate an undeployed  cluster",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action:  models.Terminate,
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
			Action:  models.Terminate,
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
			Action:  models.Terminate,
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
			Status:  true,
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
			Action:  models.Terminate,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure id is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action:  models.Terminate,
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
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}
	if !allowed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "User is not allowed to perform this action",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	if cluster.Status == models.Deploying {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't Terminate. Cluster is in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action: models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.Status == models.Terminating {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.Status == models.ClusterTerminated {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "cluster is already in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action: models.Terminate,
		}, utils.Context{})

		return

	} else if cluster.Status == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Cluster is not in created state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Can't Terminate an undeployed  cluster",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action:  models.Terminate,
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
			Action: models.Terminate,
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
			Action:  models.Terminate,
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
			Status:  true,
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
			Action:  models.Terminate,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure id is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action: models.Terminate,
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
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action: models.Terminate,
		}, utils.Context{})

		return
	}
	if !allowed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "User is not allowed to perform this action",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	if cluster.Status == models.Deploying {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't Terminate. Cluster is in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.Status == models.Terminating {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.Status == models.ClusterTerminated {
		utils.Publisher(utils.ResponseSchema{
			Status: true,
			Message: "cluster is already in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return

	} else if cluster.Status == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Cluster state is invalid for termination",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action: models.Terminate,
		}, utils.Context{})

		return
	} else if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Can't Terminate an undeployed  cluster",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	isValid, credentials := gcp.IsValidGcpCredentials(infraData.infrastructureData.ProfileId, region, task.Token, zone, *ctx)
	if !isValid {
		return
	}
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action:  models.Terminate,
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
			Status:  true,
			Message: "Cluster is in deployed state",
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
			Action:  models.Terminate,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure id is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action:  models.Terminate,
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
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}
	if !allowed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "User is not allowed to perform this action",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action: models.Terminate,
		}, utils.Context{})

		return
	}

	if cluster.Status == models.Deploying {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't Terminate. Cluster is in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.Status == models.Terminating {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.Status == models.ClusterTerminated {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "cluster is already in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action: models.Terminate,
		}, utils.Context{})

		return

	} else if cluster.Status == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Cluster state is invalid for termination",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Can't Terminate an undeployed  cluster",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action: models.Terminate,
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
			Action:  models.Terminate,
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
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	ctx.SendLogs("DOClusterController: Terminating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	go do.TerminateCluster(cluster, doProfile, *ctx, userInfo.CompanyId, task.Token)
}
func AKSClusterStartHelpler(task WorkSchema, infraData Infrastructure) {
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
	_, allowed, err := rbac_athentication.Authenticate(models.AKS, "cluster", task.InfraId, "Start", task.Token, *ctx)
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

	var cluster aks.AKSCluster

	ctx.SendLogs("AKSClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = aks.GetAKSCluster(task.InfraId, userInfo.CompanyId, *ctx)
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
			Status:  true,
			Message: "Cluster is in deployed state",
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
	err = aks.UpdateAKSCluster(cluster, *ctx)
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

	ctx.SendLogs("AKSClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	aks.DeployAKSCluster(cluster, azureProfile, userInfo.CompanyId, task.Token, *ctx)

}
func AKSClusterTerminateHelper(task WorkSchema, infraData Infrastructure) {

	if task.Token == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Token is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure id is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger("", "POST", "", task.InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	_, allowed, err := rbac_athentication.Authenticate(models.AKS, "cluster", task.InfraId, "Terminate", task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}
	if !allowed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "User is not allowed to perform this action",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action: models.Terminate,
		}, utils.Context{})

		return
	}

	var cluster aks.AKSCluster
	ctx.SendLogs("AKSClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = aks.GetAKSCluster(task.InfraId, userInfo.CompanyId, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action: models.Terminate,
		}, utils.Context{})

		return
	}

	if cluster.Status == models.Deploying {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't Terminate. Cluster is in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.Status == models.Terminating {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.Status == models.ClusterTerminated {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "cluster is already in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action: models.Terminate,
		}, utils.Context{})

		return

	} else if cluster.Status == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Cluster state is invalid for termination",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Can't Terminate an undeployed  cluster",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action: models.Terminate,
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
			Action:  models.Terminate,
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
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	cluster.Status = models.Terminating
	err = aks.UpdateAKSCluster(cluster, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	ctx.SendLogs("AKSClusterController: Terminating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	go aks.TerminateCluster(azureProfile, task.InfraId, userInfo.CompanyId, task.Token, *ctx)
}
func EKSClusterStartHelpler(task WorkSchema, infraData Infrastructure) {
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
	_, allowed, err := rbac_athentication.Authenticate(models.EKS, "cluster", task.InfraId, "Start", task.Token, *ctx)
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

	var cluster eks.EKSCluster

	ctx.SendLogs("EKSClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = eks.GetEKSCluster(task.InfraId, userInfo.CompanyId, *ctx)
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
			Status:  true,
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
	err = eks.UpdateEKSCluster(cluster, *ctx)
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

	ctx.SendLogs("EKSClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	eks.DeployEKSCluster(cluster, awsProfile, userInfo.CompanyId, task.Token, *ctx)

}
func EKSClusterTerminateHelper(task WorkSchema, infraData Infrastructure) {

	if task.Token == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Token is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure id is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger("", "POST", "", task.InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	_, allowed, err := rbac_athentication.Authenticate(models.EKS, "cluster", task.InfraId, "Terminate", task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}
	if !allowed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "User is not allowed to perform this action",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action: models.Terminate,
		}, utils.Context{})

		return
	}

	var cluster eks.EKSCluster
	ctx.SendLogs("EKSClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = eks.GetEKSCluster(task.InfraId, userInfo.CompanyId, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action: models.Terminate,
		}, utils.Context{})

		return
	}

	if cluster.Status == models.Deploying {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't Terminate. Cluster is in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.Status == models.Terminating {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.Status == models.ClusterTerminated {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "cluster is already in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return

	} else if cluster.Status == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Cluster state is invalid for termination",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status: true,
			Message: "Can't Terminate an undeployed  cluster",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action: models.Terminate,
		}, utils.Context{})

		return
	}
	_, azureProfile, err := aws.GetProfile(infraData.infrastructureData.ProfileId, region, task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	cluster.Status = models.Terminating
	err = eks.UpdateEKSCluster(cluster, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	ctx.SendLogs("EKSClusterController: Terminating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	go eks.TerminateCluster(cluster, azureProfile, task.InfraId, userInfo.CompanyId, task.Token, *ctx)
}
func DOKSClusterStartHelpler(task WorkSchema, infraData Infrastructure) {
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
	_, allowed, err := rbac_athentication.Authenticate(models.DOKS, "cluster", task.InfraId, "Start", task.Token, *ctx)
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

	var cluster doks.KubernetesCluster

	ctx.SendLogs("DOKSClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = doks.GetKubernetesCluster(*ctx)
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

	if cluster.CloudplexStatus == models.ClusterCreated {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Cluster is already in deployed state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return

	} else if cluster.CloudplexStatus == models.Deploying {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.CloudplexStatus == models.Terminating {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't deploy. Cluster is in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.CloudplexStatus == models.ClusterTerminationFailed {

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

	cluster.CloudplexStatus = models.Deploying
	err = doks.UpdateKubernetesCluster(cluster, *ctx)
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

	ctx.SendLogs("DOKSClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	doks.DeployKubernetesCluster(cluster, doProfile.Profile, task.Token, *ctx)

}
func DOKSClusterTerminateHelper(task WorkSchema, infraData Infrastructure) {

	if task.Token == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Token is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure id is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger("", "POST", "", task.InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	_, allowed, err := rbac_athentication.Authenticate(models.DOKS, "cluster", task.InfraId, "Terminate", task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}
	if !allowed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "User is not allowed to perform this action",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	var cluster doks.KubernetesCluster
	ctx.SendLogs("DOKSClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = doks.GetKubernetesCluster(*ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action: models.Terminate,
		}, utils.Context{})

		return
	}

	if cluster.CloudplexStatus == models.Deploying {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't Terminate. Cluster is in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action: models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.CloudplexStatus == models.Terminating {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.CloudplexStatus == models.ClusterTerminated {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "cluster is already in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return

	} else if cluster.CloudplexStatus == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Cluster state is invalid for termination",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if strings.ToLower(string(cluster.CloudplexStatus)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Can't Terminate an undeployed  cluster",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action:  models.Terminate,
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
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	cluster.CloudplexStatus = models.Terminating
	err = doks.UpdateKubernetesCluster(cluster, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	ctx.SendLogs("DOKSClusterController: Terminating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	go doks.TerminateCluster(doProfile.Profile, task.Token, *ctx)
}
func GKEClusterStartHelpler(task WorkSchema, infraData Infrastructure) {
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
	_, allowed, err := rbac_athentication.Authenticate(models.GKE, "cluster", task.InfraId, "Start", task.Token, *ctx)
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

	var cluster gke.GKECluster

	ctx.SendLogs("GKEClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = gke.GetGKECluster(*ctx)
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

	if cluster.CloudplexStatus == models.ClusterCreated {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Cluster is already in deployed state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return

	} else if cluster.CloudplexStatus == models.Deploying {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.CloudplexStatus == models.Terminating {

		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't deploy. Cluster is in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.CloudplexStatus == models.ClusterTerminationFailed {

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
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Invalid Credentials",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
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

	cluster.CloudplexStatus = models.Deploying
	err = gke.UpdateGKECluster(cluster, *ctx)
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

	ctx.SendLogs("GKEClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	gke.DeployGKECluster(cluster, credentials, task.Token, *ctx)

}
func GKEClusterTerminateHelper(task WorkSchema, infraData Infrastructure) {

	if task.Token == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Token is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action: models.Terminate,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure id is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger("", "POST", "", task.InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	_, allowed, err := rbac_athentication.Authenticate(models.GKE, "cluster", task.InfraId, "Terminate", task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}
	if !allowed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "User is not allowed to perform this action",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action: models.Terminate,
		}, utils.Context{})

		return
	}

	var cluster gke.GKECluster
	ctx.SendLogs("GKEClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = gke.GetGKECluster(*ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	if cluster.CloudplexStatus == models.Deploying {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't Terminate. Cluster is in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.CloudplexStatus == models.Terminating {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.CloudplexStatus == models.ClusterTerminated {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "cluster is already in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return

	} else if cluster.CloudplexStatus == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Cluster state is invalid for termination",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if strings.ToLower(string(cluster.CloudplexStatus)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Can't Terminate an undeployed  cluster",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action: models.Terminate,
		}, utils.Context{})

		return
	}
	isValid, credentials := gcp.IsValidGcpCredentials(infraData.infrastructureData.ProfileId, region, task.Token, zone, *ctx)
	if !isValid {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Invalid Credentials",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	cluster.CloudplexStatus = models.Terminating
	err = gke.UpdateGKECluster(cluster, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	ctx.SendLogs("GKEClusterController: Terminating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	go gke.TerminateCluster(credentials, task.Token, *ctx)
}
func IKSClusterStartHelpler(task WorkSchema, infraData Infrastructure) {
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
	_, allowed, err := rbac_athentication.Authenticate(models.IKS, "cluster", task.InfraId, "Start", task.Token, *ctx)
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

	var cluster iks.Cluster_Def

	ctx.SendLogs("IKSClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = iks.GetCluster(task.InfraId, userInfo.CompanyId, *ctx)
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
			Status:  true,
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

	region, err := iks.GetRegion(task.Token, task.InfraId, *ctx)
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
	_, ibmProfile, err := iks.GetProfile(infraData.infrastructureData.ProfileId, region, task.Token, *ctx)

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
	err = iks.UpdateCluster(cluster, false, *ctx)
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

	ctx.SendLogs("IKSClusterController: Creating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	iks.DeployCluster(cluster, ibmProfile.Profile, *ctx, userInfo.CompanyId, task.Token)

}
func IKSClusterTerminateHelper(task WorkSchema, infraData Infrastructure) {

	if task.Token == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Token is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action: models.Terminate,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure id is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
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
			Action: models.Terminate,
		}, utils.Context{})

		return
	}
	ctx := new(utils.Context)
	ctx.InitializeLogger("", "POST", "", task.InfraId, userInfo.CompanyId, userInfo.UserId)

	//==========================RBAC Authentication==============================//
	_, allowed, err := rbac_athentication.Authenticate(models.IKS, "cluster", task.InfraId, "Terminate", task.Token, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}
	if !allowed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "User is not allowed to perform this action",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	var cluster iks.Cluster_Def
	ctx.SendLogs("IKSClusterController: Getting Cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	cluster, err = iks.GetCluster(task.InfraId, userInfo.CompanyId, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	if cluster.Status == models.Deploying {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't Terminate. Cluster is in deploying state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.Status == models.Terminating {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Cluster is already in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if cluster.Status == models.ClusterTerminated {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "cluster is already in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return

	} else if cluster.Status == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Cluster is not in created state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	} else if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "Cluster is undeployed",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}
	region, err := iks.GetRegion(task.Token, task.InfraId, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})
		return
	}
	_, ibmProfile, err := iks.GetProfile(infraData.infrastructureData.ProfileId, region, task.Token, *ctx)

	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action: models.Terminate,
		}, utils.Context{})
		return
	}

	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	cluster.Status = models.Terminating
	err = iks.UpdateCluster(cluster, false, *ctx)
	if err != nil {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: err.Error(),
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})

		return
	}

	ctx.SendLogs("IKSClusterController: Terminating Cluster. "+cluster.Name, models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	go iks.TerminateCluster(cluster, ibmProfile, *ctx, userInfo.CompanyId, task.Token)
}

func OPClusterStartHelper(task WorkSchema, infraData Infrastructure) {

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

		utils.Publisher(utils.ResponseSchema{
			Status:  true,
			Message: "No changes made",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}

func OPClusterTerminateHelper(task WorkSchema, infraData Infrastructure) {

	if task.Token == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Token is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})
		return
	}
	if task.InfraId == "" {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Infrastructure id is missing",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Terminate,
		}, utils.Context{})
		return
	}
	utils.Publisher(utils.ResponseSchema{
		Status:  true,
		Message: "No changes made",
		InfraId: task.InfraId,
		Token:   task.Token,
		Action:  models.Terminate,
	}, utils.Context{})
	return
}

func UpdateEKSRunningCluster(task WorkSchema, infraData Infrastructure) {

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
	_, allowed, err := rbac_athentication.Authenticate(models.EKS, "cluster", task.InfraId, "Terminate", task.Token, *ctx)
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

	ctx.InitializeLogger("", "POST", "", task.InfraId, ctx.Data.Company, userInfo.UserId)

	ctx.SendLogs("EKSClusterController: Updating cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.Data.Company = userInfo.CompanyId
	ctx.Data.InfraId = task.InfraId

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

	_, eksProfile, err := aws.GetProfile(infraData.infrastructureData.ProfileId, region, task.Token, *ctx)
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
	cluster, err := eks.GetEKSCluster(task.InfraId, userInfo.CompanyId, *ctx)
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

	if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in new state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in failed state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == (models.Deploying) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in creating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == (models.Terminating) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == (models.ClusterTerminated) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}

	go eks.PatchRunningEKSCluster(cluster, eksProfile.Profile, task.Token, *ctx)

	ctx.SendLogs("EKSClusterController: Running cluster "+cluster.Name+" of infrastructure id: "+cluster.InfraId+"updated", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs(" EKS running cluster "+cluster.Name+" of infrastructure id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

}
func UpdateAKSRunningCluster(task WorkSchema, infraData Infrastructure) {

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
	_, allowed, err := rbac_athentication.Authenticate(models.AKS, "cluster", task.InfraId, "Terminate", task.Token, *ctx)
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

	ctx.InitializeLogger("", "POST", "", task.InfraId, ctx.Data.Company, userInfo.UserId)

	ctx.SendLogs("AKSClusterController: Updating cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.Data.Company = userInfo.CompanyId
	ctx.Data.InfraId = task.InfraId

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

	_, aksProfile, err := azure.GetProfile(infraData.infrastructureData.ProfileId, region, task.Token, *ctx)
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
	cluster, err := aks.GetAKSCluster(task.InfraId, userInfo.CompanyId, *ctx)
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

	if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in new state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in failed state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == (models.Deploying) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in creating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == (models.Terminating) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == (models.ClusterTerminated) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}

	go aks.PatchRunningAKSCluster(cluster, aksProfile, userInfo.CompanyId, task.Token, *ctx)

	ctx.SendLogs("AKSClusterController: Running cluster "+cluster.Name+" of infrastructure id: "+cluster.InfraId+"updated", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs("AKS running cluster "+cluster.Name+" of infrastructure id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

}
func UpdateDOKSRunningCluster(task WorkSchema, infraData Infrastructure) {

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
	_, allowed, err := rbac_athentication.Authenticate(models.DOKS, "cluster", task.InfraId, "Terminate", task.Token, *ctx)
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

	ctx.InitializeLogger("", "POST", "", task.InfraId, ctx.Data.Company, userInfo.UserId)

	ctx.SendLogs("DOKSClusterController: Updating cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.Data.Company = userInfo.CompanyId
	ctx.Data.InfraId = task.InfraId

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

	_, doksProfile, err := do.GetProfile(infraData.infrastructureData.ProfileId, region, task.Token, *ctx)
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
	cluster, err := doks.GetKubernetesCluster(*ctx)
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

	if strings.ToLower(string(cluster.CloudplexStatus)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in new state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.CloudplexStatus == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in failed state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.CloudplexStatus == (models.Deploying) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in creating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.CloudplexStatus == (models.Terminating) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.CloudplexStatus == (models.ClusterTerminated) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}

	go doks.PatchRunningDOKSCluster(cluster, doksProfile.Profile, task.Token, *ctx)

	ctx.SendLogs("DOKSClusterController: Running cluster "+cluster.Name+" of infrastructure id: "+cluster.InfraId+"updated", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs("DOKS running cluster "+cluster.Name+" of infrastructure id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

}
func UpdateIKSRunningCluster(task WorkSchema, infraData Infrastructure) {

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
	_, allowed, err := rbac_athentication.Authenticate(models.IKS, "cluster", task.InfraId, "Terminate", task.Token, *ctx)
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

	ctx.InitializeLogger("", "POST", "", task.InfraId, ctx.Data.Company, userInfo.UserId)

	ctx.SendLogs("IKSClusterController: Updating cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.Data.Company = userInfo.CompanyId
	ctx.Data.InfraId = task.InfraId

	region, err := iks.GetRegion(task.Token, task.InfraId, *ctx)
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

	_, iksProfile, err := iks.GetProfile(infraData.infrastructureData.ProfileId, region, task.Token, *ctx)
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
	cluster, err := iks.GetCluster(task.InfraId, userInfo.CompanyId, *ctx)
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

	if strings.ToLower(string(cluster.Status)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in new state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in failed state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == (models.Deploying) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in creating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == (models.Terminating) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.Status == (models.ClusterTerminated) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}

	go iks.PatchRunningIKSCluster(cluster, iksProfile.Profile, task.Token, *ctx)

	ctx.SendLogs("IKSClusterController: Running cluster "+cluster.Name+" of infrastructure id: "+cluster.InfraId+"updated", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs("IKS running cluster "+cluster.Name+" of infrastructure id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

}
func UpdateGKERunningCluster(task WorkSchema, infraData Infrastructure) {

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
	_, allowed, err := rbac_athentication.Authenticate(models.GKE, "cluster", task.InfraId, "Terminate", task.Token, *ctx)
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

	ctx.InitializeLogger("", "POST", "", task.InfraId, ctx.Data.Company, userInfo.UserId)

	ctx.SendLogs("GKEClusterController: Updating cluster of infrastructure. "+task.InfraId, models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.Data.Company = userInfo.CompanyId
	ctx.Data.InfraId = task.InfraId

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
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Invalid credentials",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})

		return
	}
	cluster, err := gke.GetGKECluster(*ctx)
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

	if strings.ToLower(string(cluster.CloudplexStatus)) == strings.ToLower(string(models.New)) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in new state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.CloudplexStatus == models.ClusterCreationFailed {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in failed state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.CloudplexStatus == (models.Deploying) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in creating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.CloudplexStatus == (models.Terminating) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in terminating state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	} else if cluster.CloudplexStatus == (models.ClusterTerminated) {
		utils.Publisher(utils.ResponseSchema{
			Status:  false,
			Message: "Can't redeploy.Cluster is in terminated state",
			InfraId: task.InfraId,
			Token:   task.Token,
			Action:  models.Create,
		}, utils.Context{})
		return
	}

	go gke.PatchRunningGKECluster(cluster, credentials, task.Token, *ctx)

	ctx.SendLogs("GKEClusterController: Running cluster "+cluster.Name+" of infrastructure id: "+cluster.InfraId+"updated", models.LOGGING_LEVEL_INFO, models.Backend_Logging)

	ctx.SendLogs("GKE running cluster "+cluster.Name+" of infrastructure id: "+cluster.InfraId+" updated ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)

}
