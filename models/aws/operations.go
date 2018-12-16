package aws
//
//import (
//	"errors"
//	"github.com/astaxie/beego"
//	"github.com/aws/aws-sdk-go/aws"
//	"github.com/aws/aws-sdk-go/aws/credentials"
//	"github.com/aws/aws-sdk-go/aws/session"
//	"github.com/aws/aws-sdk-go/service/ec2"
//)
//
//type AWS struct {
//	Client    *ec2.EC2
//	AccessKey string
//	SecretKey string
//	Region    string
//}
//
//func (cloud *AWS) getExistingVpcs() []*ec2.Vpc {
//	if cloud.Client == nil {
//		err := cloud.init()
//		if err != nil {
//			return nil
//		}
//	}
//
//	input := &ec2.DescribeVpcsInput{}
//
//	result, err := cloud.Client.DescribeVpcs(input)
//	if err != nil {
//		beego.Warn(err.Error())
//		return nil
//	}
//
//	if result != nil && result.Vpcs != nil && len(result.Vpcs) > 0 {
//		return result.Vpcs
//	}
//
//	return nil
//}
//
//func (cloud *AWS) getExistingSubnets(vpcId string) []*ec2.Subnet {
//	if cloud.Client == nil {
//		err := cloud.init()
//		if err != nil {
//			return nil
//		}
//	}
//
//	input := &ec2.DescribeSubnetsInput{
//		Filters: []*ec2.Filter{
//			{
//				Name: aws.String("vpc-id"),
//				Values: []*string{
//					aws.String(vpcId),
//				},
//			},
//		},
//	}
//
//	result, err := cloud.Client.DescribeSubnets(input)
//	if err != nil {
//		beego.Warn(err.Error())
//		return nil
//	}
//
//	if result != nil && result.Subnets != nil && len(result.Subnets) > 0 {
//		return result.Subnets
//	}
//
//	return nil
//}
//
//func (cloud *AWS) getExistingSecurityGroups(vpcId string) []*ec2.SecurityGroup {
//	if cloud.Client == nil {
//		err := cloud.init()
//		if err != nil {
//			return nil
//		}
//	}
//
//	input := &ec2.DescribeSecurityGroupsInput{
//		Filters: []*ec2.Filter{
//			{
//				Name: aws.String("vpc-id"),
//				Values: []*string{
//					aws.String(vpcId),
//				},
//			},
//		},
//	}
//
//	result, err := cloud.Client.DescribeSecurityGroups(input)
//	if err != nil {
//		beego.Warn(err.Error())
//		return nil
//	}
//
//	if result != nil && result.SecurityGroups != nil && len(result.SecurityGroups) > 0 {
//		return result.SecurityGroups
//	}
//
//	return nil
//}
//
//func (cloud *AWS) init() error {
//	if cloud.Client != nil {
//		return nil
//	}
//
//	if cloud.AccessKey == "" || cloud.SecretKey == "" || cloud.Region == "" {
//		text := "invalid cloud credentials"
//		beego.Error(text)
//		return errors.New(text)
//	}
//
//	region := cloud.Region
//	creds := credentials.NewStaticCredentials(cloud.AccessKey, cloud.SecretKey, "")
//
//	cloud.Client = ec2.New(session.New(&aws.Config{Region: &region, Credentials: creds}))
//
//	return nil
//}
