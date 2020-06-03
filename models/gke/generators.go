package gke

import (
	"antelope/models"
	gke "google.golang.org/api/container/v1"
)

func GenerateClusterFromResponse(v gke.Cluster) GKECluster {
	return GKECluster{
		Cloud:                          models.GKE,
		ClusterIpv4Cidr:                v.ClusterIpv4Cidr,
		CreateTime:                     v.CreateTime,
		CurrentMasterVersion:           v.CurrentMasterVersion,
		CurrentNodeCount:               v.CurrentNodeCount,
		Description:                    v.Description,
		EnableKubernetesAlpha:          v.EnableKubernetesAlpha,
		EnableTpu:                      v.EnableTpu,
		Endpoint:                       v.Endpoint,
		ExpireTime:                     v.ExpireTime,
		InitialClusterVersion:          v.InitialClusterVersion,
		LabelFingerprint:               v.LabelFingerprint,
		Location:                       v.Location,
		Locations:                      v.Locations,
		LoggingService:                 v.LoggingService,
		MonitoringService:              v.MonitoringService,
		Name:                           v.Name,
		Network:                        v.Network,
		NodeIpv4CidrSize:               v.NodeIpv4CidrSize,
		ResourceLabels:                 v.ResourceLabels,
		SelfLink:                       v.SelfLink,
		ServicesIpv4Cidr:               v.ServicesIpv4Cidr,
		Status:                         models.Type(v.Status),
		StatusMessage:                  v.StatusMessage,
		Subnetwork:                     v.Subnetwork,
		TpuIpv4CidrBlock:               v.TpuIpv4CidrBlock,
		Zone:                           v.Zone,
		AddonsConfig:                   GenerateAddonsConfigFromResponse(v.AddonsConfig),
		Conditions:                     GenerateConditionsFromResponse(v.Conditions),
		DefaultMaxPodsConstraint:       GenerateMaxPodsConstraintFromResponse(v.DefaultMaxPodsConstraint),
		IpAllocationPolicy:             GenerateIpAllocationPolicyFromResponse(v.IpAllocationPolicy),
		LegacyAbac:                     GenerateLegacyAbacFromResponse(v.LegacyAbac),
		MaintenancePolicy:              GenerateMaintenancePolicyFromResponse(v.MaintenancePolicy),
		MasterAuth:                     GenerateMasterAuthFromResponse(v.MasterAuth),
		MasterAuthorizedNetworksConfig: GenerateMasterAuthorizedNetworksConfigFromResponse(v.MasterAuthorizedNetworksConfig),
		NetworkConfig:                  GenerateNetworkConfigFromResponse(v.NetworkConfig),
		NetworkPolicy:                  GenerateNetworkPolicyFromResponse(v.NetworkPolicy),
		PrivateClusterConfig:           GeneratePrivateClusterConfigFromResponse(v.PrivateClusterConfig),
		ResourceUsageExportConfig:      GenerateResourceUsageExportConfigFromResponse(v.ResourceUsageExportConfig),
		NodePools:                      GenerateNodePoolFromResponse(v.NodePools),

	}
}

func GenerateClusterCreateRequest(project, region, zone string, c GKECluster) *gke.CreateClusterRequest {
	return &gke.CreateClusterRequest{
		Cluster: &gke.Cluster{
			ClusterIpv4Cidr:                c.ClusterIpv4Cidr,
			Description:                    c.Description,
			EnableKubernetesAlpha:          c.EnableKubernetesAlpha,
			EnableTpu:                      c.EnableTpu,
			InitialClusterVersion:          c.InitialClusterVersion,
			LabelFingerprint:               c.LabelFingerprint,
			Locations:                      c.Locations,
			LoggingService:                 c.LoggingService,
			MonitoringService:              c.MonitoringService,
			Name:                           c.Name,
			Network:                        c.Network,
			ResourceLabels:                 c.ResourceLabels,
			Subnetwork:                     c.Subnetwork,
			AddonsConfig:                   GenerateAddonsConfigFromRequest(c.AddonsConfig),
			DefaultMaxPodsConstraint:       GenerateMaxPodsConstraintFromRequest(c.DefaultMaxPodsConstraint),
			IpAllocationPolicy:             GenerateIpAllocationPolicyFromRequest(c.IpAllocationPolicy),
			LegacyAbac:                     GenerateLegacyAbacFromRequest(c.LegacyAbac),
			MaintenancePolicy:              GenerateMaintenancePolicyFromRequest(c.MaintenancePolicy),
			MasterAuthorizedNetworksConfig: GenerateMasterAuthorizedNetworksConfigFromRequest(c.MasterAuthorizedNetworksConfig),
			MasterAuth:                     GenerateMasterAuthFromRequest(c.MasterAuth),
			NetworkConfig:                  GenerateNetworkConfigFromRequest(c.NetworkConfig),
			NetworkPolicy:                  GenerateNetworkPolicyFromRequest(c.NetworkPolicy),
			NodePools:                      GenerateNodePoolFromRequest(c.NodePools),
			PrivateClusterConfig:           GeneratePrivateClusterConfigFromRequest(c.PrivateClusterConfig),
			ResourceUsageExportConfig:      GenerateResourceUsageExportConfigFromRequest(c.ResourceUsageExportConfig),
		},
		Parent: "projects/" + project + "/locations/" + region + "-" + zone,
	}
}

func GenerateAddonsConfigFromResponse(v *gke.AddonsConfig) *AddonsConfig {
	if v == nil {
		return nil
	}

	addonsConfig := AddonsConfig{}
	if v.HorizontalPodAutoscaling != nil {
		addonsConfig.HorizontalPodAutoscaling = &HorizontalPodAutoscaling{
			Disabled: v.HorizontalPodAutoscaling.Disabled,
		}
	}
	if v.HttpLoadBalancing != nil {
		addonsConfig.HttpLoadBalancing = &HttpLoadBalancing{
			Disabled: v.HttpLoadBalancing.Disabled,
		}
	}
	if v.KubernetesDashboard != nil {
		addonsConfig.KubernetesDashboard = &KubernetesDashboard{
			Disabled: v.KubernetesDashboard.Disabled,
		}
	}
	if v.NetworkPolicyConfig != nil {
		addonsConfig.NetworkPolicyConfig = &NetworkPolicyConfig{
			Disabled: v.NetworkPolicyConfig.Disabled,
		}
	}

	return &addonsConfig
}

func GenerateConditionsFromResponse(v []*gke.StatusCondition) []*StatusCondition {
	statusConditions := []*StatusCondition{}

	if len(v) == 0 {
		return statusConditions
	}

	for _, i := range v {
		statusConditions = append(statusConditions, &StatusCondition{Code: i.Code, Message: i.Message})
	}

	return statusConditions
}

func GenerateMaxPodsConstraintFromResponse(v *gke.MaxPodsConstraint) *MaxPodsConstraint {
	if v == nil {
		return nil
	}

	return &MaxPodsConstraint{MaxPodsPerNode: v.MaxPodsPerNode}
}

func GenerateIpAllocationPolicyFromResponse(v *gke.IPAllocationPolicy) *IPAllocationPolicy {
	if v == nil {
		return nil
	}

	return &IPAllocationPolicy{
		ClusterIpv4Cidr:            v.ClusterIpv4Cidr,
		ClusterIpv4CidrBlock:       v.ClusterIpv4CidrBlock,
		ClusterSecondaryRangeName:  v.ClusterSecondaryRangeName,
		CreateSubnetwork:           v.CreateSubnetwork,
		NodeIpv4Cidr:               v.NodeIpv4Cidr,
		NodeIpv4CidrBlock:          v.NodeIpv4CidrBlock,
		ServicesIpv4Cidr:           v.ServicesIpv4Cidr,
		ServicesIpv4CidrBlock:      v.ServicesIpv4CidrBlock,
		ServicesSecondaryRangeName: v.ServicesSecondaryRangeName,
		SubnetworkName:             v.SubnetworkName,
		TpuIpv4CidrBlock:           v.TpuIpv4CidrBlock,
		UseIpAliases:               v.UseIpAliases,
	}
}

func GenerateLegacyAbacFromResponse(v *gke.LegacyAbac) *LegacyAbac {
	if v == nil {
		return nil
	}

	return &LegacyAbac{Enabled: v.Enabled}
}

func GenerateMaintenancePolicyFromResponse(v *gke.MaintenancePolicy) *MaintenancePolicy {
	if v == nil {
		return nil
	}

	maintenancePolicy := MaintenancePolicy{}
	if v.Window != nil {
		maintenanceWindow := MaintenanceWindow{}
		if v.Window.DailyMaintenanceWindow != nil {
			dailyMaintenanceWindow := DailyMaintenanceWindow{
				Duration:  v.Window.DailyMaintenanceWindow.Duration,
				StartTime: v.Window.DailyMaintenanceWindow.StartTime,
			}
			maintenanceWindow.DailyMaintenanceWindow = &dailyMaintenanceWindow
		}
		maintenancePolicy.Window = &maintenanceWindow
	}

	return &maintenancePolicy
}

func GenerateMasterAuthFromResponse(v *gke.MasterAuth) *MasterAuth {
	if v == nil {
		return nil
	}

	masterAuth := MasterAuth{
		ClientCertificate:    v.ClientCertificate,
		ClientKey:            v.ClientKey,
		ClusterCaCertificate: v.ClusterCaCertificate,
		Password:             v.Password,
		Username:             v.Username,
	}
	if v.ClientCertificateConfig != nil {
		masterAuth.ClientCertificateConfig = &ClientCertificateConfig{
			IssueClientCertificate: v.ClientCertificateConfig.IssueClientCertificate,
		}
	}

	return &masterAuth
}

func GenerateMasterAuthorizedNetworksConfigFromResponse(v *gke.MasterAuthorizedNetworksConfig) *MasterAuthorizedNetworksConfig {
	if v == nil {
		return nil
	}

	masterAuth := MasterAuthorizedNetworksConfig{
		CidrBlocks: []*CidrBlock{},
		Enabled:    v.Enabled,
	}

	for _, i := range v.CidrBlocks {
		if i != nil {
			masterAuth.CidrBlocks = append(masterAuth.CidrBlocks, &CidrBlock{
				CidrBlock:   i.CidrBlock,
				DisplayName: i.DisplayName,
			})
		}
	}

	return &masterAuth
}

func GenerateNetworkConfigFromResponse(v *gke.NetworkConfig) *NetworkConfig {
	if v == nil {
		return nil
	}

	return &NetworkConfig{
		Network:    v.Network,
		Subnetwork: v.Subnetwork,
	}
}

func GenerateNetworkPolicyFromResponse(v *gke.NetworkPolicy) *NetworkPolicy {
	if v == nil {
		return nil
	}

	return &NetworkPolicy{
		Enabled:  v.Enabled,
		Provider: v.Provider,
	}
}

func GeneratePrivateClusterConfigFromResponse(v *gke.PrivateClusterConfig) *PrivateClusterConfig {
	if v == nil {
		return nil
	}

	return &PrivateClusterConfig{
		EnablePrivateEndpoint: v.EnablePrivateEndpoint,
		EnablePrivateNodes:    v.EnablePrivateNodes,
		MasterIpv4CidrBlock:   v.MasterIpv4CidrBlock,
		PrivateEndpoint:       v.PrivateEndpoint,
		PublicEndpoint:        v.PublicEndpoint,
	}
}

func GenerateResourceUsageExportConfigFromResponse(v *gke.ResourceUsageExportConfig) *ResourceUsageExportConfig {
	if v == nil {
		return nil
	}

	resourceUsage := ResourceUsageExportConfig{
		EnableNetworkEgressMetering: v.EnableNetworkEgressMetering,
	}
	if v.BigqueryDestination != nil {
		resourceUsage.BigqueryDestination = &BigQueryDestination{
			DatasetId: v.BigqueryDestination.DatasetId,
		}
	}
	if v.ConsumptionMeteringConfig != nil {
		resourceUsage.ConsumptionMeteringConfig = &ConsumptionMeteringConfig{
			Enabled: v.ConsumptionMeteringConfig.Enabled,
		}
	}

	return &resourceUsage
}

func GenerateNodePoolFromResponse(pools []*gke.NodePool) []*NodePool {
	nodePools := []*NodePool{}

	if len(pools) == 0 {
		return nodePools
	}

	for _, v := range pools {
		nodePool := NodePool{
			InitialNodeCount:  v.InitialNodeCount,
			InstanceGroupUrls: v.InstanceGroupUrls,
			Name:              v.Name,
			PodIpv4CidrSize:   v.PodIpv4CidrSize,
			SelfLink:          v.SelfLink,
			Status:            v.Status,
			StatusMessage:     v.StatusMessage,
			Version:           v.Version,
		}
		if v.Autoscaling != nil {
			nodePool.Autoscaling = &NodePoolAutoscaling{
				Enabled:      v.Autoscaling.Enabled,
				MaxNodeCount: v.Autoscaling.MaxNodeCount,
				MinNodeCount: v.Autoscaling.MinNodeCount,
			}
		}
		if v.Conditions != nil {
			nodePool.Conditions = GenerateConditionsFromResponse(v.Conditions)
		}
		if v.Config != nil {
			nodePool.Config = &NodeConfig{
				Accelerators:   []*AcceleratorConfig{},
				DiskSizeGb:     v.Config.DiskSizeGb,
				DiskType:       v.Config.DiskType,
				ImageType:      v.Config.ImageType,
				Labels:         v.Config.Labels,
				LocalSsdCount:  v.Config.LocalSsdCount,
				MachineType:    v.Config.MachineType,
				Metadata:       v.Config.Metadata,
				MinCpuPlatform: v.Config.MinCpuPlatform,
				OauthScopes:    v.Config.OauthScopes,
				Preemptible:    v.Config.Preemptible,
				ServiceAccount: v.Config.ServiceAccount,
				Tags:           v.Config.Tags,
				Taints:         []*NodeTaint{},
			}
			for _, i := range v.Config.Accelerators {
				nodePool.Config.Accelerators = append(nodePool.Config.Accelerators, &AcceleratorConfig{
					AcceleratorCount: i.AcceleratorCount,
					AcceleratorType:  i.AcceleratorType,
				})
			}
			for _, i := range v.Config.Taints {
				nodePool.Config.Taints = append(nodePool.Config.Taints, &NodeTaint{
					Effect: i.Effect,
					Key:    i.Key,
					Value:  i.Value,
				})
			}
		}
		if v.MaxPodsConstraint != nil {
			nodePool.MaxPodsConstraint = GenerateMaxPodsConstraintFromResponse(v.MaxPodsConstraint)
		}
		if v.Management != nil {
			nodePool.Management = &NodeManagement{
				AutoRepair:  v.Management.AutoRepair,
				AutoUpgrade: v.Management.AutoUpgrade,
			}
		}

		nodePools = append(nodePools, &nodePool)
	}

	return nodePools
}

func GenerateAddonsConfigFromRequest(v *AddonsConfig) *gke.AddonsConfig {
	if v == nil {
		return nil
	}

	addonsConfig := gke.AddonsConfig{}
	if v.HorizontalPodAutoscaling != nil {
		addonsConfig.HorizontalPodAutoscaling = &gke.HorizontalPodAutoscaling{
			Disabled: v.HorizontalPodAutoscaling.Disabled,
		}
	}
	if v.HttpLoadBalancing != nil {
		addonsConfig.HttpLoadBalancing = &gke.HttpLoadBalancing{
			Disabled: v.HttpLoadBalancing.Disabled,
		}
	}
	if v.KubernetesDashboard != nil {
		addonsConfig.KubernetesDashboard = &gke.KubernetesDashboard{
			Disabled: v.KubernetesDashboard.Disabled,
		}
	}
	if v.NetworkPolicyConfig != nil {
		addonsConfig.NetworkPolicyConfig = &gke.NetworkPolicyConfig{
			Disabled: v.NetworkPolicyConfig.Disabled,
		}
	}

	return &addonsConfig
}

func GenerateConditionsFromRequest(v []*StatusCondition) []*gke.StatusCondition {
	statusConditions := []*gke.StatusCondition{}

	if len(v) == 0 {
		return statusConditions
	}

	for _, i := range v {
		statusConditions = append(statusConditions, &gke.StatusCondition{Code: i.Code, Message: i.Message})
	}

	return statusConditions
}

func GenerateMaxPodsConstraintFromRequest(v *MaxPodsConstraint) *gke.MaxPodsConstraint {
	if v == nil {
		return nil
	}

	return &gke.MaxPodsConstraint{MaxPodsPerNode: v.MaxPodsPerNode}
}

func GenerateIpAllocationPolicyFromRequest(v *IPAllocationPolicy) *gke.IPAllocationPolicy {
	if v == nil {
		return nil
	}

	return &gke.IPAllocationPolicy{
		ClusterIpv4Cidr:            v.ClusterIpv4Cidr,
		ClusterIpv4CidrBlock:       v.ClusterIpv4CidrBlock,
		ClusterSecondaryRangeName:  v.ClusterSecondaryRangeName,
		CreateSubnetwork:           v.CreateSubnetwork,
		NodeIpv4Cidr:               v.NodeIpv4Cidr,
		NodeIpv4CidrBlock:          v.NodeIpv4CidrBlock,
		ServicesIpv4Cidr:           v.ServicesIpv4Cidr,
		ServicesIpv4CidrBlock:      v.ServicesIpv4CidrBlock,
		ServicesSecondaryRangeName: v.ServicesSecondaryRangeName,
		SubnetworkName:             v.SubnetworkName,
		TpuIpv4CidrBlock:           v.TpuIpv4CidrBlock,
		UseIpAliases:               v.UseIpAliases,
	}
}

func GenerateLegacyAbacFromRequest(v *LegacyAbac) *gke.LegacyAbac {
	if v == nil {
		return nil
	}

	return &gke.LegacyAbac{Enabled: v.Enabled}
}

func GenerateMaintenancePolicyFromRequest(v *MaintenancePolicy) *gke.MaintenancePolicy {
	if v == nil {
		return nil
	}

	maintenancePolicy := gke.MaintenancePolicy{}
	if v.Window != nil {
		maintenanceWindow := gke.MaintenanceWindow{}
		if v.Window.DailyMaintenanceWindow != nil {
			dailyMaintenanceWindow := gke.DailyMaintenanceWindow{
				Duration:  v.Window.DailyMaintenanceWindow.Duration,
				StartTime: v.Window.DailyMaintenanceWindow.StartTime,
			}
			maintenanceWindow.DailyMaintenanceWindow = &dailyMaintenanceWindow
		}
		maintenancePolicy.Window = &maintenanceWindow
	}

	return &maintenancePolicy
}

func GenerateMasterAuthFromRequest(v *MasterAuth) *gke.MasterAuth {
	if v == nil {
		return nil
	}

	masterAuth := gke.MasterAuth{
		ClientCertificate:    v.ClientCertificate,
		ClientKey:            v.ClientKey,
		ClusterCaCertificate: v.ClusterCaCertificate,
		Password:             v.Password,
		Username:             v.Username,
	}
	if v.ClientCertificateConfig != nil {
		masterAuth.ClientCertificateConfig = &gke.ClientCertificateConfig{
			IssueClientCertificate: v.ClientCertificateConfig.IssueClientCertificate,
		}
	}

	return &masterAuth
}

func GenerateMasterAuthorizedNetworksConfigFromRequest(v *MasterAuthorizedNetworksConfig) *gke.MasterAuthorizedNetworksConfig {
	if v == nil {
		return nil
	}

	masterAuth := gke.MasterAuthorizedNetworksConfig{
		CidrBlocks: []*gke.CidrBlock{},
		Enabled:    v.Enabled,
	}

	for _, i := range v.CidrBlocks {
		if i != nil {
			masterAuth.CidrBlocks = append(masterAuth.CidrBlocks, &gke.CidrBlock{
				CidrBlock:   i.CidrBlock,
				DisplayName: i.DisplayName,
			})
		}
	}

	return &masterAuth
}

func GenerateNetworkConfigFromRequest(v *NetworkConfig) *gke.NetworkConfig {
	if v == nil {
		return nil
	}

	return &gke.NetworkConfig{
		Network:    v.Network,
		Subnetwork: v.Subnetwork,
	}
}

func GenerateNetworkPolicyFromRequest(v *NetworkPolicy) *gke.NetworkPolicy {
	if v == nil {
		return nil
	}

	return &gke.NetworkPolicy{
		Enabled:  v.Enabled,
		Provider: v.Provider,
	}
}

func GeneratePrivateClusterConfigFromRequest(v *PrivateClusterConfig) *gke.PrivateClusterConfig {
	if v == nil {
		return nil
	}

	return &gke.PrivateClusterConfig{
		EnablePrivateEndpoint: v.EnablePrivateEndpoint,
		EnablePrivateNodes:    v.EnablePrivateNodes,
		MasterIpv4CidrBlock:   v.MasterIpv4CidrBlock,
		PrivateEndpoint:       v.PrivateEndpoint,
		PublicEndpoint:        v.PublicEndpoint,
	}
}

func GenerateResourceUsageExportConfigFromRequest(v *ResourceUsageExportConfig) *gke.ResourceUsageExportConfig {
	if v == nil {
		return nil
	}

	resourceUsage := gke.ResourceUsageExportConfig{
		EnableNetworkEgressMetering: v.EnableNetworkEgressMetering,
	}
	if v.BigqueryDestination != nil {
		resourceUsage.BigqueryDestination = &gke.BigQueryDestination{
			DatasetId: v.BigqueryDestination.DatasetId,
		}
	}
	if v.ConsumptionMeteringConfig != nil {
		resourceUsage.ConsumptionMeteringConfig = &gke.ConsumptionMeteringConfig{
			Enabled: v.ConsumptionMeteringConfig.Enabled,
		}
	}

	return &resourceUsage
}

func GenerateNodePoolFromRequest(pools []*NodePool) []*gke.NodePool {
	nodePools := []*gke.NodePool{}

	if len(pools) == 0 {
		return nodePools
	}

	for _, v := range pools {
		nodePool := gke.NodePool{
			InitialNodeCount:  v.InitialNodeCount,
			InstanceGroupUrls: v.InstanceGroupUrls,
			Name:              v.Name,
			PodIpv4CidrSize:   v.PodIpv4CidrSize,
			SelfLink:          v.SelfLink,
			Status:            v.Status,
			StatusMessage:     v.StatusMessage,
			Version:           v.Version,
		}
		if v.Autoscaling != nil {
			nodePool.Autoscaling = &gke.NodePoolAutoscaling{
				Enabled:      v.Autoscaling.Enabled,
				MaxNodeCount: v.Autoscaling.MaxNodeCount,
				MinNodeCount: v.Autoscaling.MinNodeCount,
			}
		}
		if v.Conditions != nil {
			nodePool.Conditions = GenerateConditionsFromRequest(v.Conditions)
		}
		if v.Config != nil {
			nodeConfig := &gke.NodeConfig{
				Accelerators:   []*gke.AcceleratorConfig{},
				DiskSizeGb:     v.Config.DiskSizeGb,
				DiskType:       v.Config.DiskType,
				ImageType:      v.Config.ImageType,
				Labels:         v.Config.Labels,
				LocalSsdCount:  v.Config.LocalSsdCount,
				MachineType:    v.Config.MachineType,
				Metadata:       v.Config.Metadata,
				MinCpuPlatform: v.Config.MinCpuPlatform,
				OauthScopes:    v.Config.OauthScopes,
				Preemptible:    v.Config.Preemptible,
				ServiceAccount: v.Config.ServiceAccount,
				Tags:           v.Config.Tags,
				Taints:         []*gke.NodeTaint{},
			}
			for _, i := range v.Config.Accelerators {
				nodeConfig.Accelerators = append(nodeConfig.Accelerators, &gke.AcceleratorConfig{
					AcceleratorCount: i.AcceleratorCount,
					AcceleratorType:  i.AcceleratorType,
				})
			}
			for _, i := range v.Config.Taints {
				nodeConfig.Taints = append(nodeConfig.Taints, &gke.NodeTaint{
					Effect: i.Effect,
					Key:    i.Key,
					Value:  i.Value,
				})
			}
			nodePool.Config = nodeConfig
		}
		if v.MaxPodsConstraint != nil {
			nodePool.MaxPodsConstraint = GenerateMaxPodsConstraintFromRequest(v.MaxPodsConstraint)
		}
		if v.Management != nil {
			nodePool.Management = &gke.NodeManagement{
				AutoRepair:  v.Management.AutoRepair,
				AutoUpgrade: v.Management.AutoUpgrade,
			}
		}

		nodePools = append(nodePools, &nodePool)
	}

	return nodePools
}
