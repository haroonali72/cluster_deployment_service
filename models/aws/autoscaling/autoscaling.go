package autoscaling

import (
	"antelope/models/aws/IAMRoles"
	"errors"
	"github.com/astaxie/beego"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
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
func (cloud *AWSAutoScaler) ConfigLauncher(projectId string, nodeIp string, imageId string) (error, map[string]string) {

	m := make(map[string]string)
	config_input := autoscaling.CreateLaunchConfigurationInput{}
	beego.Info("getting project" + projectId)
	config_input.ImageId = &imageId
	config_input.InstanceId = &nodeIp
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

	_, err := roles.CreateRole(projectId)
	if err != nil {
		beego.Error(err.Error())
		return err, m
	}
	m[projectId+"_role"] = projectId
	_, err = roles.CreatePolicy(projectId, autoscale_policy)
	if err != nil {
		beego.Error(err.Error())
		return err, m
	}
	m[projectId+"_policy"] = projectId
	id, err := roles.CreateIAMProfile(projectId)
	if err != nil {
		beego.Error(err.Error())
		return err, m
	}
	m[projectId+"_iamProfile"] = projectId
	config_input.IamInstanceProfile = &id
	_, config_err := cloud.AutoScaling.CreateLaunchConfiguration(&config_input)

	if config_err != nil {
		beego.Error(config_err.Error())
		return config_err, m
	}
	m[projectId+"_launchConfig"] = projectId
	return nil, m
}
func (cloud *AWSAutoScaler) DeleteConfiguration(projectId string) error {
	config_input := autoscaling.DeleteLaunchConfigurationInput{
		LaunchConfigurationName: aws.String(projectId),
	}
	_, config_err := cloud.AutoScaling.DeleteLaunchConfiguration(&config_input)

	if config_err != nil {
		beego.Error(config_err.Error())
		return config_err
	}
	return nil
}
func (cloud *AWSAutoScaler) AutoScaler(projectId string, nodeIp string, imageId string, subnetId string, maxSize int64) (error, map[string]string) {

	err, m := cloud.ConfigLauncher(projectId, nodeIp, imageId)
	if err != nil {
		beego.Error(err.Error())
		return err, m
	}
	min := int64(0)

	config_input := autoscaling.CreateAutoScalingGroupInput{}

	config_input.AutoScalingGroupName = &projectId
	config_input.InstanceId = &nodeIp
	config_input.MinSize = &min
	config_input.MaxSize = &maxSize
	config_input.VPCZoneIdentifier = &subnetId
	config_input.LaunchConfigurationName = &projectId

	var tags []*autoscaling.Tag
	tag := autoscaling.Tag{
		Key:   aws.String("k8s.io/cluster-autoscaler/enabled"),
		Value: aws.String("true"),
	}
	tags = append(tags, &tag)
	tag = autoscaling.Tag{
		Key:   aws.String("k8s.io/cluster-autoscaler/" + projectId),
		Value: aws.String(projectId),
	}
	tags = append(tags, &tag)
	config_input.Tags = tags

	_, config_err := cloud.AutoScaling.CreateAutoScalingGroup(&config_input)

	if config_err != nil {
		beego.Error(config_err.Error())
		return config_err, m
	}
	m[projectId+"_autoScaler"] = projectId
	return nil, m

}
func (cloud *AWSAutoScaler) DeleteAutoScaler(projectId string) error {
	config_input := autoscaling.DeleteAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(projectId),
		ForceDelete:          aws.Bool(true),
	}
	_, config_err := cloud.AutoScaling.DeleteAutoScalingGroup(&config_input)

	if config_err != nil {
		beego.Error(config_err.Error())
		return config_err
	}
	return nil
}
