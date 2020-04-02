package eks

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/google/uuid"
)

func GenerateClusterCreateRequest(project, region, zone string, c EKSCluster) *eks.CreateClusterInput {
	id, _ := uuid.NewRandom()
	return &eks.CreateClusterInput{
		ClientRequestToken: aws.String(id.String()),
		Name:               aws.String(c.Name),
		RoleArn:            aws.String(c.RoleArn),
		Tags:               c.Tags,
		Version:            c.Version,
		EncryptionConfig:   GenerateEncryptionConfigFromRequest(c.EncryptionConfig),
		Logging:            GenerateLoggingFromRequest(c.Logging),
		ResourcesVpcConfig: GenerateResourcesVpcConfigFromRequest(c.ResourcesVpcConfig),
	}
}

func GenerateEncryptionConfigFromRequest(v []*EncryptionConfig) []*eks.EncryptionConfig {
	encryptionConfigs := []*eks.EncryptionConfig{}

	for _, i := range v {
		encryptionConfig := eks.EncryptionConfig{Resources: i.Resources}
		if i.Provider != nil {
			encryptionConfig.Provider = &eks.Provider{
				KeyArn: i.Provider.KeyArn,
			}
		}
		encryptionConfigs = append(encryptionConfigs, &encryptionConfig)
	}

	return encryptionConfigs
}

func GenerateLoggingFromRequest(v *Logging) *eks.Logging {
	if v == nil {
		return nil
	}

	clusterLogging := []*eks.LogSetup{}

	for _, i := range v.ClusterLogging {
		clusterLogging = append(clusterLogging, &eks.LogSetup{
			Enabled: i.Enabled,
			Types:   i.Types,
		})
	}

	return &eks.Logging{ClusterLogging: clusterLogging}
}

func GenerateResourcesVpcConfigFromRequest(v VpcConfigRequest) *eks.VpcConfigRequest {
	return &eks.VpcConfigRequest{
		EndpointPrivateAccess: v.EndpointPrivateAccess,
		EndpointPublicAccess:  v.EndpointPublicAccess,
		PublicAccessCidrs:     v.PublicAccessCidrs,
		SecurityGroupIds:      v.SecurityGroupIds,
		SubnetIds:             v.SubnetIds,
	}
}
