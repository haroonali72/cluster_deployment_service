package queue

import (
	"antelope/models"
	"antelope/models/aws"
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
