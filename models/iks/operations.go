package iks

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	"io/ioutil"
	"strings"
	"time"
)

func GetIBM(credentials vault.IBMCredentials) (IBM, error) {
	return IBM{
		APIKey: credentials.IAMKey,
	}, nil
}

type IBMResponse struct {
	IncidentId  string `json:"incidentID"`
	Code        string `json:"code"`
	Description string `json:"description"`
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
	PublicEndpoint bool                   `json:"disablePublicServiceEndpoint"`
	KubeVersion    string                 `json:"kubeVersion"`
	Name           string                 `json:"name"`
	Provider       string                 `json:"provider"`
	WorkerPool     ClusterWorkerPoolInput `json:"workerPool"`
}
type ClusterWorkerPoolInput struct {
	DiskEncryption bool          `json:"diskEncryption"`
	MachineType    string        `json:"flavor"`
	WorkerName     string        `json:"name"`
	VPCId          string        `json:"vpcID"`
	Count          int           `json:"workerCount"`
	Zones          []ClusterZone `json:"zones"`
}
type ClusterZone struct {
	Id     string `json:"id"`
	Subnet string `json:"subnetID"`
}
type WorkerPoolInput struct {
	Cluster     string `json:"cluster"`
	MachineType string `json:"flavor"`
	WorkerName  string `json:"name"`
	VPCId       string `json:"vpcID"`
	Count       int    `json:"workerCount"`
}
type ZoneInput struct {
	Cluster    string `json:"cluster"`
	Id         string `json:"id"`
	Subnet     string `json:"subnetID"`
	WorkerPool string `json:"workerpool"`
}
type KubeClusterResponse struct {
	ID string `json:"id"`
}
type WorkerPoolResponse struct {
	ID string `json:"workerPoolID"`
}
type KubeClusterStatus struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Region            string `json:"region"`
	ResourceGroupName string `json:"resourceGroupName"`
	State             string `json:"state"`
	WorkerCount       int    `json:"workerCount"`
}
type AllInstancesResponse struct {
	Profile []InstanceProfile `json:"profiles"`
}
type InstanceProfile struct {
	Family string `json:"family"`
	Name   string `json:"name"`
}
type Versions struct {
	Major string `json:"major"`
	Minor string `json:"minor"`
	Patch string `json:"patch"`
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
func (cloud *IBM) create(cluster Cluster_Def, ctx utils.Context, companyId string, token string) (Cluster_Def, error) {

	var ibmNetwork types.IBMNetwork
	url := getNetworkHost("ibm", cluster.ProjectId)
	network, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil || network == nil {
		return cluster, errors.New("error in fetching network")
	}
	err = json.Unmarshal(network.([]byte), &ibmNetwork)

	if err != nil {
		beego.Error(err.Error())
		return cluster, err
	}

	vpcID := cloud.GetVPC(cluster.VPCId, ibmNetwork)
	if vpcID == "" {
		return cluster, errors.New("error in fetching network")
	}

	clusterId, err := cloud.createCluster(vpcID, cluster, ibmNetwork, ctx)
	if err != nil {
		beego.Error(err.Error())
		return cluster, err
	}
	cluster.ClusterId = clusterId

	for {
		response, err := cloud.fetchStatus(&cluster, ctx, companyId)
		if err != nil {
			beego.Error(err.Error())
			return cluster, err
		}
		if response.State == "normal" {
			break
		} else {
			time.Sleep(60 * time.Second)
		}
	}

	for index, pool := range cluster.NodePools {
		if index == 0 {
			continue
		}
		beego.Info("IBMOperations creating worker pools")

		utils.SendLog(companyId, "Creating Worker Pools : "+cluster.Name, "info", cluster.ProjectId)

		err := cloud.createWorkerPool(cluster.ResourceGroup, clusterId, vpcID, pool, ibmNetwork, ctx)
		if err != nil {
			utils.SendLog(companyId, "Error in instances creation: "+err.Error(), "info", cluster.ProjectId)
			return cluster, err
		}
		utils.SendLog(companyId, "Node Pool Created Successfully : "+cluster.Name, "info", cluster.ProjectId)
	}

	return cluster, nil
}
func (cloud *IBM) createCluster(vpcId string, cluster Cluster_Def, network types.IBMNetwork, ctx utils.Context) (string, error) {

	input := KubeClusterInput{
		PublicEndpoint: cluster.PublicEndpoint,
		KubeVersion:    cluster.KubeVersion,
		Name:           cluster.Name,
		Provider:       "vpc-classic",
	}

	workerpool := ClusterWorkerPoolInput{
		DiskEncryption: false,
		MachineType:    cluster.NodePools[0].MachineType,
		WorkerName:     cluster.NodePools[0].Name,
		VPCId:          vpcId,
		Count:          cluster.NodePools[0].NodeCount,
	}

	subentId := cloud.GetSubnets(cluster.NodePools[0], network)
	if subentId == "" {
		return "", errors.New("error in gettinh subnet id")
	}
	zone := ClusterZone{
		Id:     cloud.Region,
		Subnet: cluster.NodePools[0].SubnetID,
	}
	var zones []ClusterZone
	zones = append(zones, zone)

	workerpool.Zones = zones
	input.WorkerPool = workerpool

	bytes, err := json.Marshal(input)

	req, _ := utils.CreatePostRequest(bytes, models.IBM_Kube_Cluster_Endpoint)

	m := make(map[string]string)

	m["Content-Type"] = "application/json"
	m["Accept"] = "application/json"
	m["Authorization"] = cloud.IAMToken
	m["X-Auth-Refresh-Token"] = cloud.RefreshToken
	m["X-Auth-Resource-Group"] = cluster.ResourceGroup
	utils.SetHeaders(req, m)

	client := utils.InitReq()
	res, err := client.SendRequest(req)

	defer res.Body.Close()

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}

	if res.StatusCode != 201 {
		ctx.SendLogs("error in cluster creation", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
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
func (cloud *IBM) createWorkerPool(rg, clusterID, vpcID string, pool *NodePool, network types.IBMNetwork, ctx utils.Context) error {

	workerpool := WorkerPoolInput{
		Cluster:     clusterID,
		MachineType: pool.MachineType,
		WorkerName:  pool.Name,
		VPCId:       vpcID,
		Count:       pool.NodeCount,
	}

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
		return err
	}

	if res.StatusCode != 201 {
		if res.StatusCode == 409 {
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return err
			}
			var ibmResponse IBMResponse
			err = json.Unmarshal(body, &ibmResponse)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return err
			}
			if !strings.Contains(ibmResponse.Description, "already exits") {
				ctx.SendLogs("error in worker pool creation", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return err
			}
		} else {
			ctx.SendLogs("error in worker pool creation", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	err = cloud.AddZonesToPools(rg, pool.Name, pool.SubnetID, clusterID, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	return nil
}
func (cloud *IBM) AddZonesToPools(rg, poolName, subnetID, clusterID string, ctx utils.Context) error {

	zoneInput := ZoneInput{
		Cluster:    clusterID,
		Id:         cloud.Region,
		Subnet:     subnetID,
		WorkerPool: poolName,
	}

	bytes, err := json.Marshal(zoneInput)

	req, _ := utils.CreatePostRequest(bytes, models.IBM_Zone)

	m := make(map[string]string)

	m["Content-Type"] = "application/json"
	m["Accept"] = "application/json"
	m["Authorization"] = cloud.IAMToken
	m["X-Auth-Resource-Group"] = rg //rg id
	utils.SetHeaders(req, m)

	client := utils.InitReq()
	res, err := client.SendRequest(req)

	defer res.Body.Close()

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	if res.StatusCode != 201 {
		if res.StatusCode == 409 {
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return err
			}
			var ibmResponse IBMResponse
			err = json.Unmarshal(body, &ibmResponse)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return err
			}
			if !strings.Contains(ibmResponse.Description, "The zone is already part of the worker pool") {
				ctx.SendLogs("error in worker pool creation", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return err
			}
		} else {
			ctx.SendLogs("error in worker pool creation", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}
	return nil
}
func (cloud *IBM) GetAllVersions(ctx utils.Context) (Versions, error) {
	url := "https://" + cloud.Region + models.IBM_ALL_Kube_Version_Endpoint + models.IBM_Version

	req, _ := utils.CreateGetRequest(url)

	utils.SetHeaders(req, nil)

	client := utils.InitReq()
	res, err := client.SendRequest(req)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Versions{}, err
	}
	defer res.Body.Close()

	// Reading response
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Versions{}, err
	}

	// body is []byte format
	// parse the JSON-encoded body and stores the result in the struct object for the res
	var versions Versions
	err = json.Unmarshal(body, &versions)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Versions{}, err
	}

	return versions, nil
}
func (cloud *IBM) GetSubnets(pool *NodePool, network types.IBMNetwork) string {
	for _, definition := range network.Definition {
		for _, subnet := range definition.Subnets {
			if subnet.Name == pool.SubnetID {
				return subnet.SubnetId
			}
		}
	}
	return ""
}
func (cloud *IBM) GetVPC(vpcID string, network types.IBMNetwork) string {
	for _, definition := range network.Definition {
		if vpcID == definition.Vpc.Name {
			return definition.Vpc.VpcId

		}
	}
	return ""
}
func (cloud *IBM) terminateCluster(cluster *Cluster_Def, ctx utils.Context) error {

	req, _ := utils.CreateDeleteRequest(models.IBM_Kube_Delete_Cluster_Endpoint + cluster.ClusterId + "?yes")

	m := make(map[string]string)

	m["Content-Type"] = "application/json"
	m["Accept"] = "application/json"
	m["Authorization"] = cloud.IAMToken
	m["X-Auth-Refresh-Token"] = cloud.RefreshToken
	m["X-Auth-Resource-Group"] = cluster.ResourceGroup
	utils.SetHeaders(req, m)

	client := utils.InitReq()
	res, err := client.SendRequest(req)

	defer res.Body.Close()

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	if res.StatusCode != 204 {
		ctx.SendLogs("error in cluster creation", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	for {
		response, err := cloud.fetchStatus(cluster, ctx, "")
		if err == nil && response.State == "deleting" {
			beego.Error(err.Error())
			return err
		}
		break
	}
	return nil
}
func (cloud *IBM) fetchStatus(cluster *Cluster_Def, ctx utils.Context, companyId string) (KubeClusterStatus, error) {

	req, _ := utils.CreateGetRequest(models.IBM_Kube_GetCluster_Endpoint + "?cluster=" + cluster.ClusterId)

	m := make(map[string]string)

	m["Content-Type"] = "application/json"
	m["Accept"] = "application/json"
	m["Authorization"] = cloud.IAMToken
	m["X-Auth-Refresh-Token"] = cloud.RefreshToken
	m["X-Auth-Resource-Group"] = cluster.ResourceGroup
	utils.SetHeaders(req, m)

	client := utils.InitReq()
	res, err := client.SendRequest(req)

	defer res.Body.Close()

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubeClusterStatus{}, err
	}

	if res.StatusCode != 200 {
		ctx.SendLogs("error in fetching cluster ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubeClusterStatus{}, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubeClusterStatus{}, err
	}

	// body is []byte format
	// parse the JSON-encoded body and stores the result in the struct object for the res
	var response KubeClusterStatus
	err = json.Unmarshal([]byte(body), &response)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubeClusterStatus{}, err
	}
	return response, nil
}
func (cloud *IBM) GetAllInstances(ctx utils.Context) (AllInstancesResponse, error) {
	url := "https://" + cloud.Region + models.IBM_All_Instances_Endpoint + models.IBM_Version

	req, _ := utils.CreateGetRequest(url)

	m := make(map[string]string)

	m["Content-Type"] = "application/json"
	m["Accept"] = "application/json"
	m["Authorization"] = cloud.IAMToken

	utils.SetHeaders(req, m)

	client := utils.InitReq()
	res, err := client.SendRequest(req)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return AllInstancesResponse{}, err
	}
	defer res.Body.Close()

	// Reading response
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return AllInstancesResponse{}, err
	}

	// body is []byte format
	// parse the JSON-encoded body and stores the result in the struct object for the res
	var InstanceList AllInstancesResponse
	err = json.Unmarshal(body, &InstanceList)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return AllInstancesResponse{}, err
	}

	return InstanceList, nil
}
