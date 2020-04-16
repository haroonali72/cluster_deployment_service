package iks

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func GetIBM(credentials vault.IBMCredentials) IBM {
	return IBM{
		APIKey: credentials.IAMKey,
	}
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
	ID string `json:"clusterID"`
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
type KubeWorkerPoolStatus struct {
	ID     string `json:"id"`
	Name   string `json:"poolName"`
	Region string `json:"flavour"`
	State  string `json:"state"`
}
type AllInstancesResponse struct {
	Profile []InstanceProfile
}
type InstanceProfile struct {
	Id   string
	Name string
}

type Versions struct {
	Kubernetes []Kubernetes `json:"kubernetes"`
}
type Kubernetes struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

func (cloud *IBM) init(region string, ctx utils.Context) types.CustomCPError {
	payloadSlice := "grant_type=urn:ibm:params:oauth:grant-type:apikey&apikey=" + cloud.APIKey
	res, err := http.Post(models.IBM_IAM_Endpoint, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(payloadSlice)))
	if err != nil {
		ctx.SendLogs("Error while getting IBM Auth Token", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Error while getting IBM Auth Token", 500)
		return cpErr
	}

	if res.StatusCode != 200 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			ctx.SendLogs("Error while getting IBM Auth Token", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			cpErr := ApiError(err, "Error while getting IBM Auth Token", 502)
			return cpErr
		}
		beego.Info(string(body))
		ctx.SendLogs("Error while getting IBM Auth Token", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Error while getting IBM Auth Token", 502)
		return cpErr
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		ctx.SendLogs("Error while reading IBM auth token response:  "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "EError while reading IBM auth token response", 500)
		return cpErr
	}

	// body is []byte format
	// parse the JSON-encoded body and stores the result in the struct object for the res
	var token Token
	err = json.Unmarshal(body, &token)
	if err != nil {
		ctx.SendLogs("Error while getting IBM Auth Token: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Error while getting IBM Auth Token", 500)
		return cpErr
	}

	// saving the token
	cloud.IAMToken = token.TokenType + " " + token.AccessToken
	cloud.Region = region
	cloud.RefreshToken = token.RefreshToken
	return types.CustomCPError{}
}
func (cloud *IBM) create(cluster Cluster_Def, ctx utils.Context, companyId string, token string) (Cluster_Def, types.CustomCPError) {
	/*
	   Getting Network
	*/
	var ibmNetwork types.IBMNetwork
	url := getNetworkHost("ibm", cluster.ProjectId)
	network, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(companyId, "unable to fetch network against this application.\n"+err.Error(), "error", cluster.ProjectId)
		cpErr := ApiError(err, "unable to fetch network against this application", 500)
		return cluster, cpErr
	}
	if network == nil {
		ctx.SendLogs("network not found of this application", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(companyId, "unable to fetch network against this application.\n", "error", cluster.ProjectId)
		cpErr := ApiError(errors.New("network not found of this application"), "network not found of this application", 500)
		return cluster, cpErr
	}
	err = json.Unmarshal(network.([]byte), &ibmNetwork)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(companyId, "unable to fetch network against this application.\n"+err.Error(), "error", cluster.ProjectId)
		cpErr := ApiError(err, "unable to fetch network against this application", 500)
		return cluster, cpErr
	}

	vpcID := cloud.GetVPC(cluster.VPCId, ibmNetwork)
	if vpcID == "" {
		ctx.SendLogs("vpc not found", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(companyId, "vpc not found", "error", cluster.ProjectId)
		cpErr := ApiError(errors.New("vpc not found"), "error while creating iks cluster", 500)
		return cluster, cpErr
	}

	utils.SendLog(companyId, "Creating Worker Pool : "+cluster.NodePools[0].Name, "info", cluster.ProjectId)

	clusterId, cpErr := cloud.createCluster(vpcID, cluster, ibmNetwork, ctx)

	if cpErr != (types.CustomCPError{}) {

		utils.SendLog(companyId, cpErr.Message, "error", cluster.ProjectId)
		utils.SendLog(companyId, cpErr.Description, "error", cluster.ProjectId)

		return cluster, cpErr

	}
	cluster.ClusterId = clusterId

	for {
		response, cpErr := cloud.fetchClusterStatus(&cluster, ctx, companyId)
		if cpErr != (types.CustomCPError{}) {

			utils.SendLog(companyId, cpErr.Message, "error", cluster.ProjectId)
			utils.SendLog(companyId, cpErr.Description, "error", cluster.ProjectId)

		}
		beego.Info(response.State)
		if response.State == "normal" {
			break
		} else {
			time.Sleep(60 * time.Second)
		}
	}
	utils.SendLog(companyId, "Worker Pool Created Successfully : "+cluster.NodePools[0].Name, "info", cluster.ProjectId)

	for index, pool := range cluster.NodePools {
		if index == 0 {
			continue
		}
		beego.Info("IBMOperations creating worker pools")

		utils.SendLog(companyId, "Creating Worker Pools : "+cluster.Name, "info", cluster.ProjectId)

		err := cloud.createWorkerPool(cluster.ResourceGroup, clusterId, vpcID, pool, ibmNetwork, ctx)
		if err != (types.CustomCPError{}) {

			utils.SendLog(companyId, cpErr.Message, "error", cluster.ProjectId)
			utils.SendLog(companyId, cpErr.Description, "error", cluster.ProjectId)
		}
		utils.SendLog(companyId, "Node Pool Created Successfully : "+cluster.Name, "info", cluster.ProjectId)
	}

	return cluster, types.CustomCPError{}
}
func (cloud *IBM) createCluster(vpcId string, cluster Cluster_Def, network types.IBMNetwork, ctx utils.Context) (string, types.CustomCPError) {

	input := KubeClusterInput{
		PublicEndpoint: cluster.PublicEndpoint,
		KubeVersion:    cluster.KubeVersion,
		Name:           cluster.Name,
		Provider:       "vpc-classic",
	}

	workerpool := ClusterWorkerPoolInput{
		DiskEncryption: true,
		MachineType:    cluster.NodePools[0].MachineType,
		WorkerName:     cluster.NodePools[0].Name,
		VPCId:          vpcId,
		Count:          cluster.NodePools[0].NodeCount,
	}

	subentId := cloud.GetSubnets(cluster.NodePools[0], network)
	if subentId == "" {
		ctx.SendLogs(errors.New("subnet not found").Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(errors.New("subnet not found"), "subnet not found", 500)
		return "", cpErr
	}
	zone := ClusterZone{
		Id:     cluster.NodePools[0].AvailabilityZone,
		Subnet: subentId,
	}
	var zones []ClusterZone
	zones = append(zones, zone)

	workerpool.Zones = zones
	input.WorkerPool = workerpool

	bytes, err := json.Marshal(input)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error while creating iks cluster request", 500)
		return "", cpErr
	}
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
		cpErr := ApiError(err, "error while sending iks cluster creation request", 500)
		return "", cpErr
	}

	body, err := ioutil.ReadAll(res.Body)
	beego.Info(string(body))
	beego.Info(res.Status)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occured during iks cluster creation", 500)
		return "", cpErr
	}

	if res.StatusCode != 201 {
		ctx.SendLogs(errors.New(string(body)).Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(errors.New(string(body)), "error occured during iks cluster creation", 502)
		return "", cpErr
	}

	var kubeResponse KubeClusterResponse
	err = json.Unmarshal([]byte(body), &kubeResponse)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occured while parsing cluster creation response of ibm", 500)
		return "", cpErr
	}
	return kubeResponse.ID, types.CustomCPError{}
}
func (cloud *IBM) createWorkerPool(rg, clusterID, vpcID string, pool *NodePool, network types.IBMNetwork, ctx utils.Context) types.CustomCPError {
	subnetId := cloud.GetSubnets(pool, network)
	if subnetId == "" {
		ctx.SendLogs(errors.New("subnet not found").Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(errors.New("subnet not found"), "error occurred while adding workepool: "+pool.Name, 500)
		return cpErr
	}
	workerpool := WorkerPoolInput{
		Cluster:     clusterID,
		MachineType: pool.MachineType,
		WorkerName:  pool.Name,
		VPCId:       vpcID,
		Count:       pool.NodeCount,
	}

	bytes, err := json.Marshal(workerpool)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while creating workpool addition request", 500)
		return cpErr
	}

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
		cpErr := ApiError(err, "error occurred while sending workpool  "+pool.Name+" creation request", 500)
		return cpErr
	}

	if res.StatusCode != 201 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			cpErr := ApiError(err, "error occurred while adding workpool: "+pool.Name, 502)
			return cpErr
		}

		if res.StatusCode == 409 {

			var ibmResponse IBMResponse
			err = json.Unmarshal(body, &ibmResponse)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				cpErr := ApiError(err, "error occurred while adding workpool: "+pool.Name, 502)
				return cpErr
			}
			if !strings.Contains(ibmResponse.Description, "already exits") {
				ctx.SendLogs(errors.New(string(body)).Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				cpErr := ApiError(errors.New(string(body)), "error occurred while adding workpool: "+pool.Name, 502)
				return cpErr
			}
		} else {
			ctx.SendLogs(errors.New(string(body)).Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			cpErr := ApiError(errors.New(string(body)), "error occurred while adding workpool: "+pool.Name, 502)
			return cpErr
		}
	}

	cpErr := cloud.AddZonesToPools(rg, pool.Name, subnetId, pool.AvailabilityZone, clusterID, ctx)
	if cpErr != (types.CustomCPError{}) {
		return cpErr
	}
	return types.CustomCPError{}
}
func (cloud *IBM) AddZonesToPools(rg, poolName, subnetID, zone, clusterID string, ctx utils.Context) types.CustomCPError {

	zoneInput := ZoneInput{
		Cluster:    clusterID,
		Id:         zone,
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
		cpErr := ApiError(err, "error occurred while adding zone to workpool: "+poolName, 500)
		return cpErr
	}

	if res.StatusCode != 201 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			cpErr := ApiError(err, "error occurred while adding zone to workpool: "+poolName, 502)
			return cpErr
		}

		if res.StatusCode == 409 {

			var ibmResponse IBMResponse
			err = json.Unmarshal(body, &ibmResponse)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				cpErr := ApiError(err, "error occurred while adding zone to workpool: "+poolName, 502)
				return cpErr
			}
			if !strings.Contains(ibmResponse.Description, "The zone is already part of the worker pool") {
				ctx.SendLogs(errors.New(string(body)).Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				cpErr := ApiError(errors.New(string(body)), "error occurred while adding zone to workpool: "+poolName, 502)
				return cpErr
			}
		} else {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			cpErr := ApiError(errors.New(string(body)), "error occurred while adding zone to workpool: "+poolName, 502)
			return cpErr
		}
	}
	return types.CustomCPError{}
}
func (cloud *IBM) GetAllVersions(ctx utils.Context) (Versions, types.CustomCPError) {
	url := "https://" + cloud.Region + ".containers.cloud.ibm.com/global/v2/getVersions" + models.IBM_Version

	req, _ := utils.CreateGetRequest(url)

	utils.SetHeaders(req, nil)

	client := utils.InitReq()
	res, err := client.SendRequest(req)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while getting kubernetes versions", 500)
		return Versions{}, cpErr
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while getting kubernetes versions", 500)
		return Versions{}, cpErr
	}
	if res.StatusCode != 200 {
		ctx.SendLogs(errors.New(string(body)).Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(errors.New(string(body)), "error occurred while getting kubernetes versions", 502)
		return Versions{}, cpErr
	}
	// Reading response

	var kube Versions
	err = json.Unmarshal(body, &kube)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while parsing kubernetes versions", 500)
		return Versions{}, cpErr
	}

	return kube, types.CustomCPError{}
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
		if strings.ToLower(vpcID) == definition.Vpc.Name {
			return definition.Vpc.VpcId

		}
	}
	return ""
}
func (cloud *IBM) terminateCluster(cluster *Cluster_Def, ctx utils.Context) types.CustomCPError {
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
		cpErr := ApiError(err, "error occurred while sending cluster creation request"+cluster.Name+" termination request", 500)
		return cpErr
	}
	body, _ := ioutil.ReadAll(res.Body)
	beego.Info(string(body))

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while sending cluster creation request"+cluster.Name+" termination request", 500)
		return cpErr
	}
	if res.StatusCode != 204 {
		ctx.SendLogs(string(body), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(errors.New(string(body)), "error occurred while terminating cluster "+cluster.Name, 502)
		return cpErr
	}
	for {
		response, err := cloud.fetchClusterStatus(cluster, ctx, "")
		if err != (types.CustomCPError{}) {
			break
		}
		if err == (types.CustomCPError{}) && response.State == "deleting" {
			break
		}
	}
	return types.CustomCPError{}
}
func (cloud *IBM) fetchClusterStatus(cluster *Cluster_Def, ctx utils.Context, companyId string) (KubeClusterStatus, types.CustomCPError) {

	req, _ := utils.CreateGetRequest(models.IBM_Kube_GetCluster_Endpoint + "?cluster=" + cluster.ClusterId)

	m := make(map[string]string)

	m["Content-Type"] = "application/json"
	m["Accept"] = "application/json"
	m["Authorization"] = cloud.IAMToken
	m["X-Auth-Resource-Group"] = cluster.ResourceGroup
	utils.SetHeaders(req, m)

	client := utils.InitReq()
	res, err := client.SendRequest(req)

	defer res.Body.Close()

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while getting status of cluster", 500)
		return KubeClusterStatus{}, cpErr
	}

	beego.Info(res.Status)
	body, err := ioutil.ReadAll(res.Body)
	beego.Info(string(body))

	if res.StatusCode != 200 {
		ctx.SendLogs(errors.New(string(body)).Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(errors.New(res.Status), "error in fetching cluster", 502)
		return KubeClusterStatus{}, cpErr
	}

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error in fetching cluster", 500)
		return KubeClusterStatus{}, cpErr
	}

	var response KubeClusterStatus
	err = json.Unmarshal([]byte(body), &response)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while parsing ibm cluster status", 500)
		return KubeClusterStatus{}, cpErr
	}
	return response, types.CustomCPError{}
}
func (cloud *IBM) fetchStatus(cluster *Cluster_Def, ctx utils.Context, companyId string) ([]KubeWorkerPoolStatus, types.CustomCPError) {

	req, _ := utils.CreateGetRequest(models.IBM_Kube_GetWorker_Endpoint + "?cluster=" + cluster.ClusterId)

	m := make(map[string]string)

	m["Content-Type"] = "application/json"
	m["Accept"] = "application/json"
	m["Authorization"] = cloud.IAMToken
	m["X-Region"] = cloud.Region
	m["X-Auth-Resource-Group"] = cluster.ResourceGroup
	utils.SetHeaders(req, m)

	client := utils.InitReq()

	res, err := client.SendRequest(req)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while fetching cluster", 500)
		return []KubeWorkerPoolStatus{}, cpErr
	}

	defer res.Body.Close()

	beego.Info(res.Status)

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while fetching cluster", 500)
		return []KubeWorkerPoolStatus{}, cpErr
	}

	beego.Info(string(body))

	if res.StatusCode != 200 {
		ctx.SendLogs(errors.New(string(body)).Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(errors.New(string(body)), "error occurred while fetching cluster", 502)
		return []KubeWorkerPoolStatus{}, cpErr
	}

	var response []KubeWorkerPoolStatus
	err = json.Unmarshal([]byte(body), &response)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while fetching cluster", 500)
		return []KubeWorkerPoolStatus{}, cpErr
	}
	return response, types.CustomCPError{}
}
func (cloud *IBM) GetAllInstances(ctx utils.Context) (AllInstancesResponse, types.CustomCPError) {
	url := models.IBM_All_Instances_Endpoint + cloud.Region + "&provider=vpc-classic"

	req, _ := utils.CreateGetRequest(url)

	client := utils.InitReq()
	res, err := client.SendRequest(req)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while getting machine types.", 500)
		return AllInstancesResponse{}, cpErr
	}
	defer res.Body.Close()

	// Reading response
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while getting machine types.", 500)
		return AllInstancesResponse{}, cpErr
	}
	if res.StatusCode != 200 {
		ctx.SendLogs(string(body), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(errors.New(string(body)), "error occurred while getting machine types.", 502)
		return AllInstancesResponse{}, cpErr
	}

	var InstanceList AllInstancesResponse
	err = json.Unmarshal(body, &InstanceList.Profile)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while parsing supported machine types.", 500)
		return AllInstancesResponse{}, cpErr
	}

	return InstanceList, types.CustomCPError{}
}
