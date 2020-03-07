package ibm

import (
	"antelope/models"
	"antelope/models/utils"
	"antelope/models/vault"
	"encoding/json"
	"io/ioutil"
)

func GetIBM(credentials vault.IBMCredentials) (IBM, error) {
	return IBM{
		APIKey: credentials.IAMKey,
	}, nil
}

type IBM struct {
	APIKey       string
	IAMToken     string
	Region       string
	RefreshToken string
}
type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Expiration   int    `json:"expiration"`
	Scope        string `json:"scope"`
}
type KubeClusterInput struct {
	PublicEndpoint bool            `json:"disablePublicServiceEndpoint"`
	KubeVersion    string          `json:"kubeVersion"`
	Name           string          `json:"name"`
	WorkerPool     WorkerPoolInput `json:"workerPool"`
}
type WorkerPoolInput struct {
	Cluster     string `json:"cluster"`
	MachineType string `json:"flavor"`
	WorkerName  string `json:"name"`
	VPCId       string `json:"vpcID"`
	Count       int    `json:"workerCount"`
	Zone        []Zone `json:"zones"`
}

type Zone struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
}
type KubeClusterResponse struct {
	ID string `json:"id"`
}
type WorkerPoolResponse struct {
	ID string `json:"workerPoolID"`
}

func (cloud *IBM) init(region string, ctx utils.Context) error {

	client := utils.InitReq()

	m := make(map[string]string)
	m["Content-Type"] = "application/x-www-form-urlencoded"
	m["Accept"] = "application/json"

	payloadSlice := []string{"grant_type=urn:ibm:params:oauth:grant-type:apikey&apikey=", cloud.APIKey}
	bytes_, err := json.Marshal(payloadSlice)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	// Create a new request given a method, URL, and optional body.
	req, err := utils.CreatePostRequest(bytes_, models.IBM_IAM_Endpoint)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	// Adding headers to the request
	utils.SetHeaders(req, m)

	// Requesting server
	res, err := client.SendRequest(req)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer res.Body.Close()

	// Reading response
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	// body is []byte format
	// parse the JSON-encoded body and stores the result in the struct object for the res
	var token Token
	json.Unmarshal([]byte(body), &token)

	// saving the token
	cloud.IAMToken = token.TokenType + " " + token.AccessToken
	cloud.Region = region
	cloud.RefreshToken = token.RefreshToken
	return nil
}

/*func (cloud *IBM) create(cluster Cluster_Def, ctx utils.Context, companyId string, token string) (Cluster_Def, error) {


var doNetwork types.IBM
	url := getNetworkHost("do", cluster.ProjectId)
	network, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil || network == nil {
		return cluster, errors.New("error in fetching network")
	}
	err = json.Unmarshal(network.([]byte), &doNetwork)

	if err != nil {
		beego.Error(err.Error())
		return cluster, err
	}



	for index, pool := range cluster.NodePools {

		beego.Info("IBMOperations creating nodes")

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
}*/
func (cloud *IBM) createCluster(rg string, cluster Cluster_Def, ctx utils.Context) (string, error) {

	input := KubeClusterInput{
		PublicEndpoint: cluster.PublicEndpoint,
		KubeVersion:    cluster.KubeVersion,
		Name:           cluster.Name,
	}

	workerpool := WorkerPoolInput{
		MachineType: cluster.NodePools[0].MachineType,
		WorkerName:  cluster.NodePools[0].Name,
		VPCId:       cluster.VPCId,
		Count:       cluster.NodePools[0].NodeCount,
	}

	var zones []Zone
	zone := Zone{
		Name:     cluster.NodePools[0].Zone[0].Name,
		Provider: cluster.NodePools[0].Zone[0].Provider,
	}
	zones = append(zones, zone)

	workerpool.Zone = zones

	input.WorkerPool = workerpool

	bytes, err := json.Marshal(input)

	req, _ := utils.CreatePostRequest(bytes, models.IBM_Kubernetes_Endpoint)

	m := make(map[string]string)

	m["Content-Type"] = "application/json"
	m["Accept"] = "application/json"
	m["Authorization"] = cloud.IAMToken
	m["X-Auth-Refresh-Token"] = cloud.RefreshToken
	m["X-Auth-Resource-Group"] = rg
	utils.SetHeaders(req, m)

	client := utils.InitReq()
	res, err := client.SendRequest(req)

	defer res.Body.Close()

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}

	if res.StatusCode != 201 {
		ctx.SendLogs("error in worker pool creation", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}

	// body is []byte format
	// parse the JSON-encoded body and stores the result in the struct object for the res
	var kubeResponse KubeClusterResponse
	err = json.Unmarshal([]byte(body), &kubeResponse)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}
	return kubeResponse.ID, nil
}
func (cloud *IBM) createWorkerPool(rg, cluster string, pool NodePool, ctx utils.Context) (string, error) {

	workerpool := WorkerPoolInput{
		MachineType: pool.MachineType,
		WorkerName:  pool.Name,
		Count:       pool.NodeCount,
	}

	var zones []Zone
	zone := Zone{
		Name:     pool.Zone[0].Name,
		Provider: pool.Zone[0].Provider,
	}
	zones = append(zones, zone)

	workerpool.Zone = zones

	bytes, err := json.Marshal(workerpool)

	req, _ := utils.CreatePostRequest(bytes, models.IBM_WorkerPool_Endpoint)

	m := make(map[string]string)

	m["Content-Type"] = "application/json"
	m["Accept"] = "application/json"
	m["Authorization"] = cloud.IAMToken
	m["X-Auth-Refresh-Token"] = cloud.RefreshToken
	m["X-Auth-Resource-Group"] = rg //rg id
	utils.SetHeaders(req, m)

	client := utils.InitReq()
	res, err := client.SendRequest(req)

	defer res.Body.Close()

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}

	if res.StatusCode != 201 {
		ctx.SendLogs("error in worker pool creation", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}

	// body is []byte format
	// parse the JSON-encoded body and stores the result in the struct object for the res
	var response WorkerPoolResponse
	err = json.Unmarshal([]byte(body), &response)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}
	return response.ID, nil
}
