package do

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/key_utils"
	"antelope/models/types"
	userData2 "antelope/models/userData"
	"antelope/models/utils"
	"antelope/models/vault"
	"context"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"strconv"
	"strings"
	"time"
)

func getNetworkHost(cloudType, projectId string) string {

	host := beego.AppConfig.String("network_url") + models.WeaselGetEndpoint

	if strings.Contains(host, "{cloud}") {
		host = strings.Replace(host, "{cloud}", cloudType, -1)
	}

	if strings.Contains(host, "{projectId}") {
		host = strings.Replace(host, "{projectId}", projectId, -1)
	}

	return host
}

type DO struct {
	AccessKey string
	Region    string
	Client    *godo.Client
	Resources map[string][]string
}
type TokenSource struct {
	AccessToken string
}

func (cloud *DO) init(ctx utils.Context) types.CustomCPError {
	if cloud.Client != nil {
		return types.CustomCPError{}
	}

	if cloud.AccessKey == "" {
		text := "invalid cloud credentials"
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(errors.New(text), "Error while getting Do Credentials Token", 500)
		return cpErr
	}

	tokenSource := &TokenSource{
		AccessToken: cloud.AccessKey,
	}

	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	cloud.Client = godo.NewClient(oauthClient)
	cloud.Resources = make(map[string][]string)
	return types.CustomCPError{}
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}
func getWoodpecker() string {
	return beego.AppConfig.String("woodpecker_url") + models.WoodpeckerEnpoint
}

func (cloud *DO) createCluster(cluster Cluster_Def, ctx utils.Context, companyId string, token string) (Cluster_Def, types.CustomCPError) {
	var cpErr types.CustomCPError
	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != (types.CustomCPError{}) {
			return cluster, err
		}
	}
	var doNetwork types.DONetwork
	url := getNetworkHost("do", cluster.ProjectId)

	network, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil || network == nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr = ApiError(err, "Error while fetching network", 500)
		return cluster, cpErr
	}
	err = json.Unmarshal(network.([]byte), &doNetwork)

	if err != nil {
		cpErr := ApiError(err, "Error while fetching network", 500)
		return cluster, cpErr
	}

	utils.SendLog(companyId, "Creating DO Project With ID : "+cluster.ProjectId, "info", cluster.ProjectId)
	cpErr, cluster.DOProjectId = cloud.createProject(cluster.ProjectId, ctx)
	if cpErr != (types.CustomCPError{}) {
		return cluster, cpErr
	}
	cloud.Resources["project"] = append(cloud.Resources["project"], cluster.DOProjectId)
	utils.SendLog(companyId, "Project Created Successfully : "+cluster.ProjectId, "info", cluster.ProjectId)

	for index, pool := range cluster.NodePools {
		key, cpErr := cloud.getKey(*pool, cluster.ProjectId, ctx, companyId, token)
		if cpErr != (types.CustomCPError{}) {
			return cluster, cpErr
		}

		utils.SendLog(companyId, "Creating Node Pools : "+cluster.Name, "info", cluster.ProjectId)
		droplets, cpErr := cloud.createInstances(*pool, doNetwork, key, ctx, token, cluster.ProjectId)
		if cpErr != (types.CustomCPError{}) {
			utils.SendLog(companyId, "Error in instances creation: "+cpErr.Description, "info", cluster.ProjectId)
			return cluster, cpErr
		}
		utils.SendLog(companyId, "Node Pool Created Successfully : "+cluster.Name, "info", cluster.ProjectId)
		var nodes []*Node
		if droplets != nil && len(droplets) > 0 {
			var dropletsIds []int
			for in, droplet := range droplets {

				dropletsIds = append(dropletsIds, droplet.ID)

				cloud.Resources["droplets"] = append(cloud.Resources["droplets"], strconv.Itoa(droplet.ID))

				publicIp, _ := droplet.PublicIPv4()

				privateIp, _ := droplet.PrivateIPv4()

				var volID string
				if pool.IsExternal {

					utils.SendLog(companyId, "Creating Volume : "+pool.Name+strconv.Itoa(in), "info", cluster.ProjectId)
					volume, cpErr := cloud.createVolume(pool.Name+strconv.Itoa(in), pool.ExternalVolume, ctx)
					if cpErr != (types.CustomCPError{}) {
						return cluster, cpErr
					}
					cloud.Resources["volumes"] = append(cloud.Resources["volumes"], volume.ID)
					volID = volume.ID
					cpErr = cloud.attachVolume(volume.ID, droplets[in].ID, ctx)
					if cpErr != (types.CustomCPError{}) {
						return cluster, cpErr
					}
					utils.SendLog(companyId, "Volume Created Successfully : "+pool.Name+strconv.Itoa(in), "info", cluster.ProjectId)

				}
				nodes = append(nodes, &Node{CloudId: droplet.ID, NodeState: droplet.Status, Name: droplet.Name, PublicIP: publicIp, PrivateIP: privateIp, UserName: "root", VolumeId: volID})
			}

			cpErr := cloud.assignResources(dropletsIds, cluster.DOProjectId, ctx)
			if cpErr != (types.CustomCPError{}) {
				return cluster, cpErr
			}

			sgId := cloud.getSgId(doNetwork, *pool.PoolSecurityGroups[0])
			cpErr = cloud.assignSG(sgId, dropletsIds, ctx)
			if cpErr != (types.CustomCPError{}) {
				return cluster, cpErr
			}
		}
		cluster.NodePools[index].Nodes = nodes
	}

	return cluster, types.CustomCPError{}
}
func (cloud *DO) getKey(pool NodePool, projectId string, ctx utils.Context, companyId string, token string) (existingKey key_utils.AZUREKey, err types.CustomCPError) {

	bytes, err_ := vault.GetSSHKey(string(models.DO), pool.KeyInfo.KeyName, token, ctx, "")
	if err_ != nil {
		ctx.SendLogs("droplet creation failed with error: "+err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		err := ApiError(err_, "Error while fetching ssh key", 500)
		return key_utils.AZUREKey{}, err
	}
	existingKey, err_ = key_utils.AzureKeyConversion(bytes, ctx)
	if err_ != nil {
		ctx.SendLogs(err_.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		err := ApiError(err_, "Error while fetching ssh key", 500)
		return key_utils.AZUREKey{}, err
	}

	if existingKey.ID != 0 || existingKey.FingerPrint != "" {

		return existingKey, types.CustomCPError{}
	}
	err = ApiError(errors.New("Key not found"), "Error while fetching ssh key", 500)
	return key_utils.AZUREKey{}, err
}
func (cloud *DO) createInstances(pool NodePool, network types.DONetwork, key key_utils.AZUREKey, ctx utils.Context, token, projectId string) ([]godo.Droplet, types.CustomCPError) {

	vpcId := cloud.getVPCId(network, pool.VPC)
	var nodeNames []string
	var i int64
	i = 0
	for i < pool.NodeCount {
		n := pool.Name + "-" + strconv.FormatInt(i, 10)
		nodeNames = append(nodeNames, n)
		i = i + 1
	}

	imageInput := godo.DropletCreateImage{
		ID:   pool.Image.ImageId,
		Slug: pool.Image.Slug,
	}

	sshKeyInput := godo.DropletCreateSSHKey{
		ID:          key.ID,
		Fingerprint: key.FingerPrint,
	}

	var keys []godo.DropletCreateSSHKey
	keys = append(keys, sshKeyInput)
	pool.PrivateNetworking = true

	var tags []string
	tags = append(tags, projectId)
	input := &godo.DropletMultiCreateRequest{
		Names:             nodeNames,
		Region:            cloud.Region,
		Size:              pool.MachineType,
		Image:             imageInput,
		SSHKeys:           keys,
		PrivateNetworking: pool.PrivateNetworking,
		Tags:              tags,
		VPCUUID:           vpcId,
	}

	var fileName []string
	userData, err := userData2.GetUserData(token, getWoodpecker()+"/"+projectId, fileName, pool.PoolRole, ctx)

	if err != nil {
		ctx.SendLogs("Error in creating node pool : "+pool.Name+"\n"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Droplets Creation Failed", 500)
		return nil, cpErr
	}
	if input.UserData != "no user data found" {

		input.UserData = userData
	}

	droplets, _, err := cloud.Client.Droplets.CreateMultiple(context.Background(), input)
	if err != nil {
		ctx.SendLogs("Error in creating node pool : "+pool.Name+"\n"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Droplets Creation Failed", 512)
		return nil, cpErr
	}
	return droplets, types.CustomCPError{}
}
func (cloud *DO) getDroplets(dropletId int, ctx utils.Context) (godo.Droplet, types.CustomCPError) {

	droplet, _, err := cloud.Client.Droplets.Get(context.Background(), dropletId)
	if err != nil {
		ctx.SendLogs("Error in getting droplets"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Error in getting droplets", 512)
		return godo.Droplet{}, cpErr
	}
	return *droplet, types.CustomCPError{}
}
func (cloud *DO) createProject(projectId string, ctx utils.Context) (types.CustomCPError, string) {
	projectInput := &godo.CreateProjectRequest{
		Name:        projectId,
		Purpose:     "Operational / Developer tooling",
		Description: "deploying customer solution on DO",
		Environment: "Development",
	}
	project, _, err := cloud.Client.Projects.Create(context.Background(), projectInput)
	if err != nil {
		ctx.SendLogs("Error in creating project on DO : "+projectId+"\n"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Error while creating digital ocean  project", 512)
		return cpErr, ""
	}
	return types.CustomCPError{}, project.ID
}
func (cloud *DO) deleteProject(projectId string, ctx utils.Context) types.CustomCPError {

	_, err := cloud.Client.Projects.Delete(context.Background(), projectId)
	if err != nil {
		ctx.SendLogs("Error in creating project on DO : "+projectId+"\n"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Project Deletion Failed", 512)
		return cpErr
	}
	return types.CustomCPError{}
}
func (cloud *DO) assignResources(droptlets []int, doProjectId string, ctx utils.Context) types.CustomCPError {

	var resources []interface{}
	for _, id := range droptlets {
		resources = append(resources, "do:droplet:"+strconv.Itoa(id))
	}

	_, _, err := cloud.Client.Projects.AssignResources(context.Background(), doProjectId, resources...)
	if err != nil {
		ctx.SendLogs("Error in resource assignment : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Error in resource assignment", 512)

		return cpErr
	}
	return types.CustomCPError{}
}
func (cloud *DO) importKey(name, publicKey string, ctx utils.Context) (types.CustomCPError, godo.Key) {

	input := &godo.KeyCreateRequest{
		Name:      name,
		PublicKey: publicKey,
	}
	key, _, err := cloud.Client.Keys.Create(context.Background(), input)
	if err != nil {
		ctx.SendLogs("Error in importing key on DO : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Error in importing key on DO", 512)
		return cpErr, godo.Key{}
	}
	return types.CustomCPError{}, *key
}
func (cloud *DO) deleteKey(id int, ctx utils.Context) types.CustomCPError {

	_, err := cloud.Client.Keys.DeleteByID(context.Background(), id)
	if err != nil {
		ctx.SendLogs("Error in key deletion on DO : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Key deletion failed on DO", 512)
		return cpErr
	}
	return types.CustomCPError{}
}
func (cloud *DO) getCores(ctx utils.Context) ([]godo.Region, types.CustomCPError) {
	input := &godo.ListOptions{}
	regions, _, err := cloud.Client.Regions.List(context.Background(), input)
	if err != nil {
		ctx.SendLogs("Error in  getting info from DO : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Error in getting region informaton from DO", 512)
		return []godo.Region{}, cpErr
	}
	return regions, types.CustomCPError{}
}
func (cloud *DO) createVolume(poolName string, vol Volume, ctx utils.Context) (godo.Volume, types.CustomCPError) {

	input := &godo.VolumeCreateRequest{
		SizeGigaBytes:   vol.VolumeSize,
		Region:          cloud.Region,
		Name:            poolName,
		FilesystemType:  "ext4",
		FilesystemLabel: "example",
	}
	volume, _, err := cloud.Client.Storage.CreateVolume(context.Background(), input)
	if err != nil {
		ctx.SendLogs("Error in creating volume : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Error in creating volume", 512)
		return godo.Volume{}, cpErr
	}
	return *volume, types.CustomCPError{}
}
func (cloud *DO) deleteVolume(volumeName string, ctx utils.Context, dropletId int) types.CustomCPError {

	if dropletId != -1 {
		for true {
			time.Sleep(time.Second * 5)
			_, err := cloud.getDroplets(dropletId, ctx)
			if err != (types.CustomCPError{}) && strings.Contains(err.Description, strings.ToLower("not be found")) {
				break
			}

		}
	}
	time.Sleep(time.Second * 25)
	_, err := cloud.Client.Storage.DeleteVolume(context.Background(), volumeName)
	if err != nil {
		ctx.SendLogs("Error in deleting volume: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Volume Deletion Failed", 512)
		return cpErr
	}
	return types.CustomCPError{}
}
func (cloud *DO) attachVolume(volumeId string, dropletID int, ctx utils.Context) types.CustomCPError {

	for true {
		time.Sleep(time.Second * 5)
		droplet, err := cloud.getDroplets(dropletID, ctx)
		if err != (types.CustomCPError{}) {
			ctx.SendLogs("Error in volume attachment: "+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
		if droplet.Status == "active" {
			break
		}
	}
	_, _, err := cloud.Client.StorageActions.Attach(context.Background(), volumeId, dropletID)
	if err != nil {
		cpErr := ApiError(err, "Error in volume attachment", 512)
		ctx.SendLogs("Error in volume attachment: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cpErr
	}
	return types.CustomCPError{}
}
func (cloud *DO) fetchStatus(cluster *Cluster_Def, ctx utils.Context, companyId string, token string) types.CustomCPError {
	var cpErr types.CustomCPError
	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != (types.CustomCPError{}) {
			ctx.SendLogs("Failed to get latest status"+err.Description, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}
	for in, _ := range cluster.NodePools {

		var keyInfo key_utils.AZUREKey

		if cluster.NodePools[in].KeyInfo.CredentialType == models.SSHKey {
			bytes, err := vault.GetSSHKey(string(models.DO), cluster.NodePools[in].KeyInfo.KeyName, token, ctx, "")
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				if strings.Contains(err.Error(), "not found") {
					cpErr = ApiError(err, err.Error(), 404)
				} else if strings.Contains(err.Error(), "not authorized") {
					cpErr = ApiError(err, err.Error(), 401)
				}
				return cpErr
			}
			keyInfo, err = key_utils.AzureKeyConversion(bytes, ctx)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				cpErr := ApiError(err, "Error in getting key", 500)
				return cpErr
			}

		}
		cluster.NodePools[in].KeyInfo = keyInfo

		for index, node := range cluster.NodePools[in].Nodes {

			droplet, err := cloud.getDroplets(node.CloudId, ctx)

			if err != (types.CustomCPError{}) {
				return err
			}

			cluster.NodePools[in].Nodes[index].NodeState = droplet.Status
			privateIp, _ := droplet.PrivateIPv4()
			publicIp, _ := droplet.PublicIPv4()
			cluster.NodePools[in].Nodes[index].PublicIP = publicIp
			cluster.NodePools[in].Nodes[index].PrivateIP = privateIp

		}
	}
	return types.CustomCPError{}
}
func (cloud *DO) terminateCluster(cluster *Cluster_Def, ctx utils.Context, companyId string) types.CustomCPError {
	var cpErr types.CustomCPError
	if cloud.Client == nil {
		cpErr = cloud.init(ctx)
		if cpErr != (types.CustomCPError{}) {
			return cpErr
		}
	}

	utils.SendLog(companyId, "Terminating Node Pools : "+cluster.Name, "info", cluster.ProjectId)
	for in, pool := range cluster.NodePools {

		for _, node := range cluster.NodePools[in].Nodes {

			utils.SendLog(companyId, "Terminating Droplet : "+node.Name, "info", cluster.ProjectId)
			cpErr = cloud.deleteDroplet(node.CloudId, ctx)

			if cpErr != (types.CustomCPError{}) {
				return cpErr
			}
			utils.SendLog(companyId, "Droplet "+node.Name+" Terminated Successfully ", "info", cluster.ProjectId)

			if pool.IsExternal {
				utils.SendLog(companyId, "Terminating Volume With ID : "+node.VolumeId, "info", cluster.ProjectId)
				cpErr = cloud.deleteVolume(node.VolumeId, ctx, node.CloudId)
				if cpErr != (types.CustomCPError{}) {
					return cpErr
				}
				utils.SendLog(companyId, "Volume "+node.VolumeId+"Terminated Successfully ", "info", cluster.ProjectId)
			}

		}
	}
	utils.SendLog(companyId, "Node Pools Terminated Successfully : "+cluster.Name, "info", cluster.ProjectId)

	utils.SendLog(companyId, "Deleting DO Project : "+cluster.Name, "info", cluster.ProjectId)
	cpErr = cloud.deleteProject(cluster.DOProjectId, ctx)
	if cpErr != (types.CustomCPError{}) {
		return cpErr
	}
	utils.SendLog(companyId, "DO Project Deleted Successfully : "+cluster.Name, "info", cluster.ProjectId)
	return types.CustomCPError{}
}
func (cloud *DO) deleteDroplet(dropletId int, ctx utils.Context) types.CustomCPError {

	_, err := cloud.Client.Droplets.Delete(context.Background(), dropletId)
	if err != nil {
		cpErr := ApiError(err, "Error in getting droplets", 512)
		ctx.SendLogs("Error in getting droplets"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cpErr
	}
	return types.CustomCPError{}
}
func (cloud *DO) CleanUp(ctx utils.Context) types.CustomCPError {

	if cloud.Resources["droplets"] != nil {
		for _, dropletId := range cloud.Resources["droplets"] {
			id, err := strconv.Atoi(dropletId)
			if err != nil {
				cpErr := ApiError(err, "Error in getting droplet info", 500)
				return cpErr
			}
			cpErr := cloud.deleteDroplet(id, ctx)
			if cpErr != (types.CustomCPError{}) {
				return cpErr
			}
		}
	}
	if cloud.Resources["volumes"] != nil {

		volumes := cloud.Resources["volumes"]
		for _, volume := range volumes {
			cpErr := cloud.deleteVolume(volume, ctx, -1)
			if cpErr != (types.CustomCPError{}) {
				return cpErr
			}
		}

	}
	if cloud.Resources["project"] != nil {
		cpErr := cloud.deleteProject(cloud.Resources["project"][0], ctx)
		if cpErr != (types.CustomCPError{}) {
			return cpErr
		}
	}
	return types.CustomCPError{}
}
func (cloud *DO) assignSG(firewallId string, dropletId []int, ctx utils.Context) types.CustomCPError {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != (types.CustomCPError{}) {
			return err
		}
	}
	_, err := cloud.Client.Firewalls.AddDroplets(context.Background(), firewallId, dropletId...)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Error in assigning firewall", 512)
		return cpErr
	}

	return types.CustomCPError{}
}
func (cloud *DO) getSgId(doNetwork types.DONetwork, sgName string) string {
	for _, network := range doNetwork.Definition {
		for _, sg := range network.SecurityGroups {
			if sg.Name == sgName {
				return sg.SecurityGroupId
			}
		}
	}
	return ""
}
func (cloud *DO) getVPCId(doNetwork types.DONetwork, vpcName string) string {
	for _, network := range doNetwork.Definition {
		for _, vpc := range network.VPCs {
			if vpc.Name == vpcName {
				return vpc.VPCId
			}
		}
	}
	return ""
}
