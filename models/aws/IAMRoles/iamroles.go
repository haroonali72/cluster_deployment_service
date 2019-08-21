package IAMRoles

import (
	"antelope/constants"
	"antelope/models/utils"
	"errors"
	"github.com/astaxie/beego"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"strings"
	"time"
)

type AWSIAMRoles struct {
	IAMService *iam.IAM
	STS        *sts.STS
	Client     *ec2.EC2
	AccessKey  string
	SecretKey  string
	Region     string
}

func (cloud *AWSIAMRoles) Init() error {
	if cloud.IAMService != nil {
		return nil
	}

	if cloud.AccessKey == "" || cloud.SecretKey == "" || cloud.Region == "" {
		text := "invalid cloud credentials"
		beego.Error(text)
		return errors.New(text)
	}

	region := cloud.Region
	creds := credentials.NewStaticCredentials(cloud.AccessKey, cloud.SecretKey, "")
	cloud.Client = ec2.New(session.New(&aws.Config{Region: &region, Credentials: creds}))
	cloud.IAMService = iam.New(session.New(&aws.Config{Region: &region, Credentials: creds}))
	cloud.STS = sts.New(session.New(&aws.Config{Region: &region, Credentials: creds}))
	return nil
}

/*func (cloud *AWSIAMRoles) CreateIAMRole(name string, policyDef []byte) (string, error) {

	roleName := name

	raw_policy := policyDef
	raw_role := []byte(`{
				  "Version": "2012-10-17",
				  "Statement": [
				    {
				      "Effect": "Allow",
				      "Principal": { "Service": "ec2.amazonaws.com"},
				      "Action": "sts:AssumeRole"
				    }
				  ]
				}`)
	role := string(raw_role)
	policy := string(raw_policy)

	roleInput := iam.CreateRoleInput{AssumeRolePolicyDocument: &role, RoleName: &roleName}
	out, err := cloud.IAMService.CreateRole(&roleInput)
	if err != nil {
		beego.Error(err)
		return "", err
	}
	beego.Info(out.GoString())

	policy_out, err_1 := cloud.IAMService.CreatePolicy(&iam.CreatePolicyInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     &roleName,
	})

	if err_1 != nil {
		beego.Error(err_1)
		return "", err_1
	}

	attach := iam.AttachRolePolicyInput{RoleName: &roleName, PolicyArn: policy_out.Policy.Arn}
	_, err_2 := cloud.IAMService.AttachRolePolicy(&attach)

	if err_2 != nil {
		beego.Error(err_2)
		return "", err_2
	}

	profileInput := iam.CreateInstanceProfileInput{InstanceProfileName: &roleName}
	outtt, err := cloud.IAMService.CreateInstanceProfile(&profileInput)
	if err != nil {
		beego.Error(err)
		return "", err
	}
	testProfile := iam.AddRoleToInstanceProfileInput{InstanceProfileName: &roleName, RoleName: &roleName}
	_, err = cloud.IAMService.AddRoleToInstanceProfile(&testProfile)
	if err != nil {
		beego.Error(err)
		return "", err
	}

	return *outtt.InstanceProfile.Arn, nil

}*/
func (cloud *AWSIAMRoles) CreateRole(name string) (string, error) {

	roleName := name

	raw_role := []byte(`{
				  "Version": "2012-10-17",
				  "Statement": [
				    {
				      "Effect": "Allow",
				      "Principal": { "Service": "ec2.amazonaws.com"},
				      "Action": "sts:AssumeRole"
				    }
				  ]
				}`)
	role := string(raw_role)
	roleInput := iam.CreateRoleInput{AssumeRolePolicyDocument: &role, RoleName: &roleName}
	_, err := cloud.IAMService.CreateRole(&roleInput)
	if err != nil {
		return "", err
	}
	return roleName, nil

}
func (cloud *AWSIAMRoles) CreatePolicy(name string, policyDef []byte, ctx utils.Context) (string, error) {

	roleName := name

	raw_policy := policyDef

	policy := string(raw_policy)

	policy_out, err_1 := cloud.IAMService.CreatePolicy(&iam.CreatePolicyInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     &roleName,
	})

	if err_1 != nil {
		logType := []string{"backend-logging"}
		ctx.SendLogs(err_1.Error(), constants.LOGGING_LEVEL_ERROR, logType)
		return "", err_1
	}
	attach := iam.AttachRolePolicyInput{RoleName: &roleName, PolicyArn: policy_out.Policy.Arn}
	_, err_2 := cloud.IAMService.AttachRolePolicy(&attach)

	if err_2 != nil {
		logType := []string{"backend-logging"}
		ctx.SendLogs(err_2.Error(), constants.LOGGING_LEVEL_ERROR, logType)

		return "", err_2
	}

	return roleName, nil

}
func (cloud *AWSIAMRoles) CreateIAMProfile(name string, ctx utils.Context) (string, error) {

	roleName := name

	profileInput := iam.CreateInstanceProfileInput{InstanceProfileName: &roleName}
	outtt, err := cloud.IAMService.CreateInstanceProfile(&profileInput)
	if err != nil {
		logType := []string{"backend-logging"}
		ctx.SendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR, logType)
		return "", err
	}
	testProfile := iam.AddRoleToInstanceProfileInput{InstanceProfileName: &roleName, RoleName: &roleName}
	_, err = cloud.IAMService.AddRoleToInstanceProfile(&testProfile)
	if err != nil {
		logType := []string{"backend-logging"}
		ctx.SendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR, logType)
		return "", err
	}

	return *outtt.InstanceProfile.Arn, nil

}
func (cloud *AWSIAMRoles) CheckInstanceProfile(iamProfileName string) bool {

	iamProfile := ec2.IamInstanceProfileSpecification{Name: aws.String(iamProfileName)}

	start := time.Now()
	timeToWait := 60 //seconds
	retry := true

	region := cloud.Region
	ami := testInstanceMap[region]

	for retry && int64(time.Since(start).Seconds()) < int64(timeToWait) {

		//this dummy instance run , to check the success of RunInstance call
		//this is to ensure that iamProfile is properly propagated
		_, err := cloud.Client.RunInstances(&ec2.RunInstancesInput{
			// An Amazon Linux AMI ID for t2.micro instances in the us-west-2 region
			ImageId:            aws.String(ami),
			InstanceType:       aws.String("t2.micro"),
			MinCount:           aws.Int64(1),
			MaxCount:           aws.Int64(1),
			DryRun:             aws.Bool(true),
			IamInstanceProfile: &iamProfile,
		})

		beego.Error(err)

		if err != nil && strings.Contains(err.Error(), "DryRunOperation: Request would have succeeded") {
			retry = false
		} else {
			beego.Info("time passed %6.2f sec\n", time.Since(start).Seconds())
			beego.Info("waiting 5 seconds before retry")
			time.Sleep(5 * time.Second)
		}

	}
	beego.Info("retry", retry)
	return retry
}

var testInstanceMap = map[string]string{
	"us-east-2":      "ami-9686a4f3",
	"sa-east-1":      "ami-a3e39ecf",
	"eu-central-1":   "ami-5a922335",
	"us-west-1":      "ami-2d5c6d4d",
	"us-west-2":      "ami-ecc63a94",
	"ap-northeast-2": "ami-0f6fb461",
	"ca-central-1":   "ami-e59c2581",
	"eu-west-2":      "ami-e1f2e185",
	"ap-southeast-1": "ami-e6d3a585",
	"eu-west-1":      "ami-17d11e6e",
	"ap-southeast-2": "ami-391ff95b",
	"ap-northeast-1": "ami-8422ebe2",
	"us-east-1":      "ami-d651b8ac",
	"ap-south-1":     "ami-08a5e367",
}

func (cloud *AWSIAMRoles) DeletePolicy(policyName string, ctx utils.Context) error {
	err, policyArn := cloud.GetPolicyARN(policyName)
	if err != nil {
		beego.Error(err.Error())
		return err
	}
	policy_input := iam.DeletePolicyInput{PolicyArn: &policyArn}
	policy_out, err_1 := cloud.IAMService.DeletePolicy(&policy_input)

	if err_1 != nil {
		logType := []string{"backend-logging"}
		ctx.SendLogs(err_1.Error(), constants.LOGGING_LEVEL_ERROR, logType)
		return err_1
	}

	beego.Info(policy_out.GoString())
	return nil
}
func (cloud *AWSIAMRoles) GetPolicyARN(policyName string) (error, string) {
	id, err := cloud.getAccountId()
	if err != nil {
		beego.Error(err.Error())
		return err, ""
	}
	policyArn := "arn:aws:iam::" + id + ":policy/" + policyName
	return nil, policyArn
}
func (cloud *AWSIAMRoles) DeleteRole(roleName string, ctx utils.Context) error {
	err, policyArn := cloud.GetPolicyARN(roleName)
	if err != nil {
		logType := []string{"backend-logging"}
		ctx.SendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR, logType)

		return err
	}
	policy := iam.DetachRolePolicyInput{RoleName: &roleName, PolicyArn: &policyArn}
	out, err := cloud.IAMService.DetachRolePolicy(&policy)
	if err != nil {
		logType := []string{"backend-logging"}
		ctx.SendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR, logType)
		return err
	}

	beego.Info(out.GoString())

	roleInput := iam.DeleteRoleInput{RoleName: &roleName}
	out_, err := cloud.IAMService.DeleteRole(&roleInput)
	if err != nil {
		logType := []string{"backend-logging"}
		ctx.SendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR, logType)
		return err
	}

	beego.Info(out_.GoString())
	return nil
}
func (cloud *AWSIAMRoles) DeleteIAMProfile(roleName string, ctx utils.Context) error {
	profile := iam.RemoveRoleFromInstanceProfileInput{InstanceProfileName: &roleName, RoleName: &roleName}
	outtt, err := cloud.IAMService.RemoveRoleFromInstanceProfile(&profile)
	if err != nil {
		logType := []string{"backend-logging"}
		ctx.SendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR, logType)
		return err
	}
	beego.Info(outtt.GoString())

	profileInput := iam.DeleteInstanceProfileInput{InstanceProfileName: &roleName}
	outt, err := cloud.IAMService.DeleteInstanceProfile(&profileInput)
	if err != nil {
		logType := []string{"backend-logging"}
		ctx.SendLogs(err.Error(), constants.LOGGING_LEVEL_ERROR, logType)
		return err
	}
	beego.Info(outt.GoString())
	return nil
}
func (cloud *AWSIAMRoles) DeleteIAMRole(name string, ctx utils.Context) error {

	roleName := name
	err := cloud.DeleteIAMProfile(roleName, ctx)
	if err != nil {
		return err
	}
	err = cloud.DeleteRole(roleName, ctx)
	if err != nil {
		return err
	}
	err = cloud.DeletePolicy(roleName, ctx)
	if err != nil {
		return err
	}
	return nil

}
func (cloud *AWSIAMRoles) getAccountId() (string, error) {
	input := sts.GetCallerIdentityInput{}
	resp, err := cloud.STS.GetCallerIdentity(&input)
	if err != nil {
		beego.Error(err.Error())
		return "", err
	}
	return *resp.Account, nil

}
