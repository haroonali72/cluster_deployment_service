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

func (cloud *DO) init(ctx utils.Context) error {
	if cloud.Client != nil {
		return nil
	}

	if cloud.AccessKey == "" {
		text := "invalid cloud credentials"
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(text)
		return errors.New(text)
	}

	tokenSource := &TokenSource{
		AccessToken: cloud.AccessKey,
	}

	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	cloud.Client = godo.NewClient(oauthClient)
	cloud.Resources = make(map[string][]string)
	return nil
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

func (cloud *DO) createCluster(cluster Cluster_Def, ctx utils.Context, companyId string, token string) (Cluster_Def, error) {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != nil {
			return cluster, err
		}
	}
	var doNetwork types.DONetwork
	url := getNetworkHost("do", cluster.ProjectId)
	network, err := api_handler.GetAPIStatus(token, url, ctx)

	err = json.Unmarshal(network.([]byte), &doNetwork)

	if err != nil {
		beego.Error(err.Error())
		return cluster, err
	}

	utils.SendLog(companyId, "Creating DO Project With ID : "+cluster.ProjectId, "info", cluster.ProjectId)
	err, cluster.DOProjectId = cloud.createProject(cluster.ProjectId, ctx)
	if err != nil {
		return cluster, err
	}
	cloud.Resources["project"] = append(cloud.Resources["project"], cluster.DOProjectId)
	utils.SendLog(companyId, "Project Created Successfully : "+cluster.ProjectId, "info", cluster.ProjectId)

	for index, pool := range cluster.NodePools {
		key, err := cloud.getKey(*pool, cluster.ProjectId, ctx, companyId, token)
		if err != nil {
			return cluster, err
		}
		beego.Info("DOOperations creating nodes")

		utils.SendLog(companyId, "Creating Node Pools : "+cluster.Name, "info", cluster.ProjectId)
		droplets, err := cloud.createInstances(*pool, doNetwork, key, ctx, token, cluster.ProjectId)
		if err != nil {
			utils.SendLog(companyId, "Error in instances creation: "+err.Error(), "info", cluster.ProjectId)
			return cluster, err
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
					volume, err := cloud.createVolume(pool.Name+strconv.Itoa(in), pool.ExternalVolume, ctx)
					if err != nil {
						return cluster, err
					}
					cloud.Resources["volumes"] = append(cloud.Resources["volumes"], volume.ID)
					volID = volume.ID
					err = cloud.attachVolume(volume.ID, droplets[in].ID, ctx)
					if err != nil {
						return cluster, err
					}
					utils.SendLog(companyId, "Volume Created Successfully : "+pool.Name+strconv.Itoa(in), "info", cluster.ProjectId)

				}
				nodes = append(nodes, &Node{CloudId: droplet.ID, NodeState: droplet.Status, Name: droplet.Name, PublicIP: publicIp, PrivateIP: privateIp, UserName: "root", VolumeId: volID})
			}

			err := cloud.assignResources(dropletsIds, cluster.DOProjectId, ctx)
			if err != nil {
				return cluster, err
			}

			sgId, err := cloud.getSgId(doNetwork, *pool.PoolSecurityGroups[0])
			err = cloud.assignSG(sgId, dropletsIds, ctx)
			if err != nil {
				return cluster, err
			}
		}
		cluster.NodePools[index].Nodes = nodes
	}

	return cluster, nil
}
func (cloud *DO) getKey(pool NodePool, projectId string, ctx utils.Context, companyId string, token string) (existingKey key_utils.AZUREKey, err error) {

	//if pool.KeyInfo.CredentialType == models.SSHKey {

	bytes, err := vault.GetSSHKey(string(models.DO), pool.KeyInfo.KeyName, token, ctx, "")
	if err != nil {
		ctx.SendLogs("droplet creation failed with error: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return key_utils.AZUREKey{}, err
	}
	existingKey, err = key_utils.AzureKeyConversion(bytes, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return key_utils.AZUREKey{}, err
	}

	if existingKey.ID != 0 && existingKey.FingerPrint != "" {

		return existingKey, nil
	}
	//}
	return key_utils.AZUREKey{}, errors.New("key not found")
}
func (cloud *DO) createInstances(pool NodePool, network types.DONetwork, key key_utils.AZUREKey, ctx utils.Context, token, projectId string) ([]godo.Droplet, error) {

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
	}

	var fileName []string
	userData, err := userData2.GetUserData(token, getWoodpecker()+"/"+projectId, fileName, pool.PoolRole, ctx)

	if err != nil {
		ctx.SendLogs("Error in creating node pool : "+pool.Name+"\n"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}
	if input.UserData != "no user data found" {

		input.UserData = userData
	}

	droplets, _, err := cloud.Client.Droplets.CreateMultiple(context.Background(), input)
	if err != nil {
		ctx.SendLogs("Error in creating node pool : "+pool.Name+"\n"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return nil, err
	}
	return droplets, nil
}
func (cloud *DO) getDroplets(dropletId int, ctx utils.Context) (godo.Droplet, error) {

	droplet, _, err := cloud.Client.Droplets.Get(context.Background(), dropletId)
	if err != nil {
		ctx.SendLogs("Error in getting droplets"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return godo.Droplet{}, err
	}
	return *droplet, nil
}
func (cloud *DO) createProject(projectId string, ctx utils.Context) (error, string) {
	projectInput := &godo.CreateProjectRequest{
		Name:        projectId,
		Purpose:     "Operational / Developer tooling",
		Description: "deploying customer solution on DO",
		Environment: "Development",
	}
	project, _, err := cloud.Client.Projects.Create(context.Background(), projectInput)
	if err != nil {
		ctx.SendLogs("Error in creating project on DO : "+projectId+"\n"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return err, ""
	}
	return nil, project.ID
}

func (cloud *DO) deleteProject(projectId string, ctx utils.Context) error {

	_, err := cloud.Client.Projects.Delete(context.Background(), projectId)
	if err != nil {
		ctx.SendLogs("Error in creating project on DO : "+projectId+"\n"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return err
	}
	return nil
}
func (cloud *DO) assignResources(droptlets []int, doProjectId string, ctx utils.Context) error {

	var resources []interface{}
	for _, id := range droptlets {
		resources = append(resources, "do:droplet:"+strconv.Itoa(id))
	}

	_, _, err := cloud.Client.Projects.AssignResources(context.Background(), doProjectId, resources...)
	if err != nil {
		ctx.SendLogs("Error in resource assignement : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func (cloud *DO) importKey(name, publicKey string, ctx utils.Context) (error, godo.Key) {

	input := &godo.KeyCreateRequest{
		Name:      name,
		PublicKey: publicKey,
	}
	key, _, err := cloud.Client.Keys.Create(context.Background(), input)
	if err != nil {
		ctx.SendLogs("Error in key generation on DO : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err, godo.Key{}
	}
	return nil, *key
}
func (cloud *DO) deleteKey(id int, ctx utils.Context) error {

	_, err := cloud.Client.Keys.DeleteByID(context.Background(), id)
	if err != nil {
		ctx.SendLogs("Error in key generation on DO : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func (cloud *DO) getCores(ctx utils.Context) ([]godo.Region, error) {
	input := &godo.ListOptions{}
	regions, _, err := cloud.Client.Regions.List(context.Background(), input)
	if err != nil {
		ctx.SendLogs("Error in  getting info from DO : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return []godo.Region{}, err
	}
	return regions, nil
}
func (cloud *DO) createVolume(poolName string, vol Volume, ctx utils.Context) (godo.Volume, error) {

	input := &godo.VolumeCreateRequest{
		SizeGigaBytes:   vol.VolumeSize,
		Region:          cloud.Region,
		Name:            poolName,
		FilesystemType:  "ext4",
		FilesystemLabel: "example",
	}
	volume, _, err := cloud.Client.Storage.CreateVolume(context.Background(), input)
	if err != nil {
		ctx.SendLogs("Error in  getting info from DO : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return godo.Volume{}, err
	}
	return *volume, nil
}
func (cloud *DO) deleteVolume(volumeName string, ctx utils.Context, dropletId int) error {

	if dropletId != -1 {
		for true {
			time.Sleep(time.Second * 5)
			_, err := cloud.getDroplets(dropletId, ctx)
			if err != nil && strings.Contains(err.Error(), strings.ToLower("not be found")) {
				break
			}

		}
	}
	time.Sleep(time.Second * 25)
	_, err := cloud.Client.Storage.DeleteVolume(context.Background(), volumeName)
	if err != nil {
		ctx.SendLogs("Error in  getting info from DO : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func (cloud *DO) attachVolume(volumeId string, dropletID int, ctx utils.Context) error {

	for true {
		time.Sleep(time.Second * 5)
		droplet, err := cloud.getDroplets(dropletID, ctx)
		if err != nil {
			return err
		}
		if droplet.Status == "active" {
			break
		}
	}
	_, _, err := cloud.Client.StorageActions.Attach(context.Background(), volumeId, dropletID)
	if err != nil {
		ctx.SendLogs("Error in  getting info from DO : "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func (cloud *DO) fetchStatus(cluster *Cluster_Def, ctx utils.Context, companyId string, token string) error {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != nil {
			ctx.SendLogs("Failed to get latest status"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}
	for in, _ := range cluster.NodePools {

		var keyInfo key_utils.AZUREKey

		if cluster.NodePools[in].KeyInfo.CredentialType == models.SSHKey {
			bytes, err := vault.GetSSHKey(string(models.DO), cluster.NodePools[in].KeyInfo.KeyName, token, ctx, "")
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return err
			}
			keyInfo, err = key_utils.AzureKeyConversion(bytes, ctx)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return err
			}

		}
		cluster.NodePools[in].KeyInfo = keyInfo

		for index, node := range cluster.NodePools[in].Nodes {

			droplet, err := cloud.getDroplets(node.CloudId, ctx)

			if err != nil {
				return err
			}

			cluster.NodePools[in].Nodes[index].NodeState = droplet.Status
			privateIp, _ := droplet.PrivateIPv4()
			publicIp, _ := droplet.PublicIPv4()
			cluster.NodePools[in].Nodes[index].PublicIP = publicIp
			cluster.NodePools[in].Nodes[index].PrivateIP = privateIp

		}
	}
	return nil
}
func (cloud *DO) terminateCluster(cluster *Cluster_Def, ctx utils.Context, companyId string) error {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != nil {
			ctx.SendLogs("Failed to terminate cluster"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	utils.SendLog(companyId, "Terminating Node Pools : "+cluster.Name, "info", cluster.ProjectId)
	for in, pool := range cluster.NodePools {

		for _, node := range cluster.NodePools[in].Nodes {

			utils.SendLog(companyId, "Terminating Droplet : "+node.Name, "info", cluster.ProjectId)
			err := cloud.deleteDroplet(node.CloudId, ctx)

			if err != nil {
				return err
			}
			utils.SendLog(companyId, "Droplet "+node.Name+" Terminated Successfully ", "info", cluster.ProjectId)

			if pool.IsExternal {
				utils.SendLog(companyId, "Terminating Volume With ID : "+node.VolumeId, "info", cluster.ProjectId)
				err := cloud.deleteVolume(node.VolumeId, ctx, node.CloudId)
				if err != nil {
					return err
				}
				utils.SendLog(companyId, "Volume "+node.VolumeId+"Terminated Successfully ", "info", cluster.ProjectId)
			}

		}
	}
	utils.SendLog(companyId, "Node Pools Terminated Successfully : "+cluster.Name, "info", cluster.ProjectId)

	utils.SendLog(companyId, "Deleting DO Project : "+cluster.Name, "info", cluster.ProjectId)
	err := cloud.deleteProject(cluster.DOProjectId, ctx)
	if err != nil {
		return err
	}
	utils.SendLog(companyId, "DO Project Deleted Successfully : "+cluster.Name, "info", cluster.ProjectId)
	return nil
}
func (cloud *DO) deleteDroplet(dropletId int, ctx utils.Context) error {

	_, err := cloud.Client.Droplets.Delete(context.Background(), dropletId)
	if err != nil {
		ctx.SendLogs("Error in getting droplets"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func (cloud *DO) CleanUp(ctx utils.Context) error {

	if cloud.Resources["droplets"] != nil {
		for _, dropletId := range cloud.Resources["droplets"] {
			id, err := strconv.Atoi(dropletId)
			if err != nil {
				return err
			}
			err = cloud.deleteDroplet(id, ctx)
			if err != nil {
				return err
			}
		}
	}
	if cloud.Resources["volumes"] != nil {

		volumes := cloud.Resources["volumes"]
		for _, volume := range volumes {
			err := cloud.deleteVolume(volume, ctx, -1)
			if err != nil {
				return err
			}
		}

	}
	if cloud.Resources["project"] != nil {
		err := cloud.deleteProject(cloud.Resources["project"][0], ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
func (cloud *DO) assignSG(firewallId string, dropletId []int, ctx utils.Context) error {

	if cloud.Client == nil {
		err := cloud.init(ctx)
		if err != nil {
			return err
		}
	}
	_, err := cloud.Client.Firewalls.AddDroplets(context.Background(), firewallId, dropletId...)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return err
	}

	return nil
}
func (cloud *DO) getSgId(doNetwork types.DONetwork, sgName string) (string, error) {
	for _, network := range doNetwork.Definition {
		for _, sg := range network.SecurityGroups {
			if sg.Name == sgName {
				return sg.SecurityGroupId, nil
			}
		}
	}
	return "", nil
}
