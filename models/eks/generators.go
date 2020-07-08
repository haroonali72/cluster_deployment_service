package eks

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/google/uuid"
)

func GenerateClusterUpdateLoggingRequest(name string, logging Logging) *eks.UpdateClusterConfigInput {
	id, _ := uuid.NewRandom()
	input := &eks.UpdateClusterConfigInput{
		ClientRequestToken: aws.String(id.String()),
		Name:               aws.String(name),
		Logging:            generateLoggingFromRequest(logging),
	}

	return input

}
func GenerateClusterUpdateNetworkRequest(name string, vpcConfig VpcConfigRequest) *eks.UpdateClusterConfigInput {
	id, _ := uuid.NewRandom()
	input := &eks.UpdateClusterConfigInput{
		ClientRequestToken: aws.String(id.String()),
		Name:               aws.String(name),
		ResourcesVpcConfig: generateResourcesVpcConfigFromRequest(vpcConfig),
	}

	return input

}
func GeneratNodeConfigUpdateRequest(clusterName, poolName string, scalingConfig NodePoolScalingConfig) *eks.UpdateNodegroupConfigInput {
	id, _ := uuid.NewRandom()
	input := &eks.UpdateNodegroupConfigInput{
		ClientRequestToken: aws.String(id.String()),
		ClusterName:        aws.String(clusterName),
		NodegroupName:      aws.String(poolName),
		ScalingConfig:      generateScalingConfigFromRequest(&scalingConfig),
	}

	return input

}
func GenerateUpdateClusterVersionRequest(clusterName, version string) *eks.UpdateClusterVersionInput {
	id, _ := uuid.NewRandom()
	input := &eks.UpdateClusterVersionInput{
		ClientRequestToken: aws.String(id.String()),
		Version:            aws.String(version),
	}
	return input

}

func GenerateClusterCreateRequest(c EKSCluster) *eks.CreateClusterInput {
	id, _ := uuid.NewRandom()
	input := &eks.CreateClusterInput{
		ClientRequestToken: aws.String(id.String()),
		Name:               aws.String(c.Name),
		RoleArn:            c.RoleArn,
		Tags:               c.Tags,
		Version:            c.Version,
		Logging:            generateLoggingFromRequest(c.Logging),
		ResourcesVpcConfig: generateResourcesVpcConfigFromRequest(c.ResourcesVpcConfig),
	}

	if c.EncryptionConfig != nil && c.EncryptionConfig.EnableEncryption {
		input.EncryptionConfig = generateEncryptionConfigFromRequest([]*EncryptionConfig{c.EncryptionConfig})
	}

	return input

}

func generateEncryptionConfigFromRequest(v []*EncryptionConfig) []*eks.EncryptionConfig {
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

func generateLoggingFromRequest(v Logging) *eks.Logging {
	return &eks.Logging{
		ClusterLogging: []*eks.LogSetup{
			{
				Enabled: aws.Bool(v.EnableApi),
				Types:   []*string{aws.String("api")},
			},
			{
				Enabled: aws.Bool(v.EnableAudit),
				Types:   []*string{aws.String("audit")},
			},
			{
				Enabled: aws.Bool(v.EnableAuthenticator),
				Types:   []*string{aws.String("authenticator")},
			},
			{
				Enabled: aws.Bool(v.EnableControllerManager),
				Types:   []*string{aws.String("controllerManager")},
			},
			{
				Enabled: aws.Bool(v.EnableScheduler),
				Types:   []*string{aws.String("scheduler")},
			},
		},
	}
}

func generateResourcesVpcConfigFromRequest(v VpcConfigRequest) *eks.VpcConfigRequest {
	return &eks.VpcConfigRequest{
		EndpointPrivateAccess: v.EndpointPrivateAccess,
		EndpointPublicAccess:  v.EndpointPublicAccess,
		PublicAccessCidrs:     v.PublicAccessCidrs,
		SecurityGroupIds:      v.SecurityGroupIds,
		SubnetIds:             v.SubnetIds,
	}
}

func GenerateNodePoolCreateRequest(n NodePool, clusterName string) *eks.CreateNodegroupInput {
	id, _ := uuid.NewRandom()
	input := &eks.CreateNodegroupInput{
		AmiType:            n.AmiType,
		ClientRequestToken: aws.String(id.String()),
		ClusterName:        aws.String(clusterName),
		DiskSize:           n.DiskSize,
		InstanceTypes:      []*string{n.InstanceType},
		Labels:             n.Labels,
		NodeRole:           n.NodeRole,
		NodegroupName:      aws.String(n.NodePoolName),
		ScalingConfig:      generateScalingConfigFromRequest(n.ScalingConfig),
		Subnets:            n.Subnets,
		Tags:               n.Tags,
	}

	if n.RemoteAccess != nil && n.RemoteAccess.EnableRemoteAccess {
		input.RemoteAccess = generateRemoteAccessFromRequest(n.RemoteAccess)
	}

	return input
}

func generateRemoteAccessFromRequest(v *RemoteAccessConfig) *eks.RemoteAccessConfig {
	if v == nil {
		return nil
	}

	return &eks.RemoteAccessConfig{
		Ec2SshKey:            v.Ec2SshKey,
		SourceSecurityGroups: v.SourceSecurityGroups,
	}
}

func generateScalingConfigFromRequest(v *NodePoolScalingConfig) *eks.NodegroupScalingConfig {
	if v == nil {
		return nil
	}

	return &eks.NodegroupScalingConfig{
		DesiredSize: v.DesiredSize,
		MaxSize:     v.MaxSize,
		MinSize:     v.MinSize,
	}
}
