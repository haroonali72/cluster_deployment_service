package autoscaling

import (
	"antelope/models/aws/IAMRoles"
	"antelope/models/utils"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"time"
)

type AWSAutoScaler struct {
	AutoScaling *autoscaling.AutoScaling
	AccessKey   string
	SecretKey   string
	Region      string
}

var autoscale_policy = []byte(`{
   "Version": "2012-10-17",
   "Statement": [
       {
           "Effect": "Allow",
           "Action": [
               "autoscaling:DescribeAutoScalingGroups",
               "autoscaling:DescribeAutoScalingInstances",
               "autoscaling:DescribeLaunchConfigurations",
               "autoscaling:SetDesiredCapacity",
				"autoscaling:DescribeTags",
               "autoscaling:TerminateInstanceInAutoScalingGroup"
           ],
           "Resource": "*"
       }
   ]
}`)

func (cloud *AWSAutoScaler) Init() error {
	if cloud.AutoScaling != nil {
		return nil
	}

	if cloud.AccessKey == "" || cloud.SecretKey == "" || cloud.Region == "" {
		text := "invalid cloud credentials"
		beego.Error(text)
		return errors.New(text)
	}

	region := cloud.Region
	creds := credentials.NewStaticCredentials(cloud.AccessKey, cloud.SecretKey, "")
	cloud.AutoScaling = autoscaling.New(session.New(&aws.Config{Region: &region, Credentials: creds}))

	return nil
}
func (cloud *AWSAutoScaler) ConfigLauncher(projectId string, nodeId string, imageId string, ctx utils.Context) (error, map[string]string) {
	fmt.Println(nodeId)
	m := make(map[string]string)

	config_input := autoscaling.CreateLaunchConfigurationInput{}

	config_input.ImageId = &imageId
	config_input.InstanceId = &nodeId
	config_input.LaunchConfigurationName = &projectId

	roles := IAMRoles.AWSIAMRoles{
		AccessKey: cloud.AccessKey,
		SecretKey: cloud.SecretKey,
		Region:    cloud.Region,
	}
	confError := roles.Init()
	if confError != nil {
		return confError, m
	}
	/*
		_, err := roles.CreateRole(projectId+"-scale")
		if err != nil {
			ctx.SendSDLog(err.Error(), "error")
			return err, m
		}
		m[projectId+"_scale_role"] = projectId+"-scale"
		_, err = roles.CreatePolicy(projectId+"-scale", autoscale_policy, ctx)
		if err != nil {
			ctx.SendSDLog(err.Error(), "error")
			return err, m
		}
		m[projectId+"_scale_policy"] = projectId+"-scale"*/
	/*id, err := roles.CreateIAMProfile(projectId+"-scale", ctx)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return err, m
	}
	m[projectId+"_scale_iamProfile"] = projectId+"-scale"
	ok := roles.CheckInstanceProfile(projectId+"-scale")
	if !ok {
		//config_input.IamInstanceProfile = &id
	} else {
		ctx.SendSDLog("failed in attaching", "info")
	}*/
	_, config_err := cloud.AutoScaling.CreateLaunchConfiguration(&config_input)

	if config_err != nil {
		ctx.SendSDLog(config_err.Error(), "error")
		return config_err, m
	}
	m[projectId+"_scale_launchConfig"] = projectId
	return nil, m
}
func (cloud *AWSAutoScaler) DeleteConfiguration(projectId string) error {
	config_input := autoscaling.DeleteLaunchConfigurationInput{
		LaunchConfigurationName: aws.String(projectId),
	}
	_, config_err := cloud.AutoScaling.DeleteLaunchConfiguration(&config_input)

	if config_err != nil {
		return config_err
	}
	return nil
}
func (cloud *AWSAutoScaler) AutoScaler(name string, nodeIp string, imageId string, subnetId string, maxSize int64, ctx utils.Context, projectId string) (error, map[string]string) {
	beego.Info("before sleep")
	time.Sleep(time.Second * 180)
	beego.Info("after sleep")
	err, m := cloud.ConfigLauncher(name, nodeIp, imageId, ctx)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return err, m
	}
	min := int64(0)

	config_input := autoscaling.CreateAutoScalingGroupInput{}

	config_input.AutoScalingGroupName = &name
	//	config_input.InstanceId = &nodeIp
	config_input.MinSize = &min
	config_input.MaxSize = &maxSize
	config_input.VPCZoneIdentifier = &subnetId
	config_input.LaunchConfigurationName = &name

	var tags []*autoscaling.Tag
	//tag := autoscaling.Tag{
	//	Key:   aws.String("k8s.io/cluster-autoscaler/enabled"),
	//	Value: aws.String("true"),
	//}
	//tags = append(tags, &tag)
	tag_ := autoscaling.Tag{Key: aws.String("KubernetesCluster"), Value: aws.String(projectId)}
	tags = append(tags, &tag_)
	tag := autoscaling.Tag{
		Key:   aws.String("Name"),
		Value: aws.String(name),
	}
	tags = append(tags, &tag)
	config_input.Tags = tags

	_, config_err := cloud.AutoScaling.CreateAutoScalingGroup(&config_input)

	if config_err != nil {
		ctx.SendSDLog(config_err.Error(), "error")
		return config_err, m
	}
	m[name+"_scale_autoScaler"] = name
	return nil, m

}
func (cloud *AWSAutoScaler) DeleteAutoScaler(projectId string) error {
	config_input := autoscaling.DeleteAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(projectId),
		ForceDelete:          aws.Bool(true),
	}
	_, config_err := cloud.AutoScaling.DeleteAutoScalingGroup(&config_input)

	if config_err != nil {
		return config_err
	}
	return nil
}
func (cloud *AWSAutoScaler) GetAutoScaler(projectId string, name string, ctx utils.Context) (error, []*autoscaling.Instance) {
	str := []*string{&name}

	config_input := autoscaling.DescribeAutoScalingGroupsInput{}

	config_input.AutoScalingGroupNames = str

	out, config_err := cloud.AutoScaling.DescribeAutoScalingGroups(&config_input)

	if config_err != nil {
		ctx.SendSDLog(config_err.Error(), "error")
		return config_err, nil
	}
	if out != nil && out.AutoScalingGroups != nil && out.AutoScalingGroups[0].Instances != nil {
		return nil, out.AutoScalingGroups[0].Instances
	} else {
		return nil, nil
	}

}
