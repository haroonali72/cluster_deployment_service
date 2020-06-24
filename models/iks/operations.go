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
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Region            string                 `json:"region"`
	ResourceGroupName string                 `json:"resourceGroupName"`
	State             string                 `json:"state"`
	KubernetesVersion string                 `json:"masterKubeVersion"`
	WorkerCount       int                    `json:"workerCount"`
	WorkerPools       []KubeWorkerPoolStatus `json:"nodePools"`
	EtcdPort          string                 `json:"etcdPort"`
	Crn               string                 `json:"crn"`
	Status            models.Type            `json:"status"`
}
type KubeClusterStatus1 struct {
	ID                string                  `json:"id,omitempty"`
	Name              string                  `json:"name,omitempty"`
	Region            string                  `json:"region,omitempty"`
	Status            models.Type             `json:"status,omitempty"`
	ResourceGroup     string                  `json:"resource_group,omitempty"`
	State             string                  `json:"state,omitempty"`
	KubernetesVersion string                  `json:"kubernetes_version,omitempty"`
	PoolCount         int                     `json:"nodepool_count,omitempty"`
	WorkerPools       []KubeWorkerPoolStatus1 `json:"node_pools"`
}

type KubeWorkerPoolStatus struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"poolName"`
	Flavour     string                  `json:"flavor"`
	Autoscaling bool                    `json:"autoscaleEnabled"`
	Count       int                     `json:"workerCount"`
	Nodes       []KubeWorkerNodesStatus `json:"nodes"`
}
type KubeWorkerPoolStatus1 struct {
	ID          string                   `json:"id,omitempty"`
	Name        string                   `json:"name,omitempty"`
	Flavour     string                   `json:"machine_type,omitempty"`
	Autoscaling Autoscaling              `json:"auto_scaling,omitempty"`
	Nodes       []KubeWorkerNodesStatus1 `json:"nodes"`
	Count       int                      `json:"node_count,omitempty"`
	SubnetId    string                   `json:"subnet_id,omitempty"`
}

type Autoscaling struct {
	AutoScale bool  `json:"autoscale,omitempty"  bson:"autoscaling,omitempty" description:"Autoscaling configuration, possible value 'true' or 'false' [required]"`
	MinNodes  int64 `json:"min_scaling_group_size,omitempty"  bson:"min_scaling_group_size,omitempty" description:"Min VM count ['required' if autoscaling is enabled]"`
	MaxNodes  int64 `json:"max_scaling_group_size,omitempty"  bson:"max_scaling_group_size,omitempty" description:"Max VM count, must be greater than min count ['required' if autoscaling is enabled]"`
}
type KubeWorkerNodesStatus1 struct {
	PoolId    string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	State     string `json:"state,omitempty"`
	PrivateIp string `json:"private_ip,omitempty"`
	PublicIp  string `json:"public_ip,omitempty"`
}
type KubeWorkerNodesStatus struct {
	ID                string              `json:"id"`
	Flavour           string              `json:"flavor"`
	Network           NetworkInfo         `json:"networkInformation"`
	Lifecycle         LifeCycle           `json:"lifecycle"`
	Location          string              `json:"location"`
	PoolId            string              `json:"poolID"`
	NetworkInterfaces []networkInterfaces `json:"networkInterfaces"`
}
type networkInterfaces struct {
	SubnetId  string `json:"subnetID,omitempty"`
	IpAddress string `json:"ipAddress,omitempty"`
	Cidr      string `json:"cidr,omitempty"`
}

type LifeCycle struct {
	State string `json:"actualState"`
}
type NetworkInfo struct {
	PrivateIp string `json:"privateIP"`
	PublicIp  string `json:"publicIP"`
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
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "Error while getting IBM Auth Token", 500)
		return cpErr
	}

	if res.StatusCode != 200 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			ctx.SendLogs("Error while getting IBM Auth Token: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			cpErr := ApiError(err, "Error while getting IBM Auth Token", 512)
			return cpErr
		}
		beego.Info(string(body))
		ctx.SendLogs("Error while getting IBM Auth Token: "+string(body), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(errors.New(string(body)), "Error while getting IBM Auth Token", 512)
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

		utils.SendLog(companyId, cpErr.Error, "error", cluster.ProjectId)
		utils.SendLog(companyId, cpErr.Description, "error", cluster.ProjectId)

		return cluster, cpErr

	}
	cluster.ClusterId = clusterId

	for {
		response, cpErr := cloud.fetchClusterStatus(&cluster, ctx, companyId)
		if cpErr != (types.CustomCPError{}) {

			utils.SendLog(companyId, cpErr.Error, "error", cluster.ProjectId)
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

			utils.SendLog(companyId, cpErr.Error, "error", cluster.ProjectId)
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
	req.Close = true
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
		cpErr := ApiError(errors.New(string(body)), "error occured during iks cluster creation", 512)
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
	req.Close = true
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
			cpErr := ApiError(err, "error occurred while adding workpool: "+pool.Name, 512)
			return cpErr
		}

		if res.StatusCode == 409 {

			var ibmResponse IBMResponse
			err = json.Unmarshal(body, &ibmResponse)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				cpErr := ApiError(err, "error occurred while adding workpool: "+pool.Name, 512)
				return cpErr
			}
			if !strings.Contains(ibmResponse.Description, "already exits") {
				ctx.SendLogs(errors.New(string(body)).Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				cpErr := ApiError(errors.New(string(body)), "error occurred while adding workpool: "+pool.Name, 512)
				return cpErr
			}
		} else {
			ctx.SendLogs(errors.New(string(body)).Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			cpErr := ApiError(errors.New(string(body)), "error occurred while adding workpool: "+pool.Name, 512)
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
	req.Close = true
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
			cpErr := ApiError(err, "error occurred while adding zone to workpool: "+poolName, 512)
			return cpErr
		}

		if res.StatusCode == 409 {

			var ibmResponse IBMResponse
			err = json.Unmarshal(body, &ibmResponse)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				cpErr := ApiError(err, "error occurred while adding zone to workpool: "+poolName, 512)
				return cpErr
			}
			if !strings.Contains(ibmResponse.Description, "The zone is already part of the worker pool") {
				ctx.SendLogs(errors.New(string(body)).Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				cpErr := ApiError(errors.New(string(body)), "error occurred while adding zone to workpool: "+poolName, 512)
				return cpErr
			}
		} else {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			cpErr := ApiError(errors.New(string(body)), "error occurred while adding zone to workpool: "+poolName, 512)
			return cpErr
		}
	}
	return types.CustomCPError{}
}
func (cloud *IBM) GetAllVersions(ctx utils.Context) (Versions, types.CustomCPError) {
	url := "https://" + cloud.Region + ".containers.cloud.ibm.com/global/v2/getVersions" + models.IBM_Version

	req, _ := utils.CreateGetRequest(url)
	req.Close = true
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
		cpErr := ApiError(errors.New(string(body)), "error occurred while getting kubernetes versions", 512)
		return Versions{}, cpErr
	}
	// Reading response

	var kube Versions
	err = json.Unmarshal(body, &kube)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(errors.New("Profilenot valid.Check profile credentails again."), "Error in validating profile", 500)
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
	req.Close = true
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
		cpErr := ApiError(errors.New(string(body)), "error occurred while terminating cluster "+cluster.Name, 512)
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

	req, err := utils.CreateGetRequest(models.IBM_Kube_GetCluster_Endpoint + "?cluster=" + cluster.ClusterId)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while getting status of cluster", 500)
		return KubeClusterStatus{}, cpErr

	}
	req.Close = true
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
		cpErr := ApiError(errors.New(res.Status), "error in fetching cluster", 512)
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

func (cloud *IBM) fetchStatus(cluster *Cluster_Def, ctx utils.Context, companyId string) (KubeClusterStatus, types.CustomCPError) {

	kubeCluster, cperr := cloud.fetchClusterStatus(cluster, ctx, companyId)
	if cperr != (types.CustomCPError{}) {
		return KubeClusterStatus{}, cperr
	}
	req, _ := utils.CreateGetRequest(models.IBM_Kube_GetWorker_Endpoint + "?cluster=" + cluster.ClusterId)
	req.Close = true
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
		return KubeClusterStatus{}, cpErr
	}

	defer res.Body.Close()

	beego.Info(res.Status)

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while fetching cluster", 500)
		return KubeClusterStatus{}, cpErr
	}

	beego.Info(string(body))

	if res.StatusCode != 200 {
		ctx.SendLogs(errors.New(string(body)).Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(errors.New(string(body)), "error occurred while fetching cluster", 512)
		return KubeClusterStatus{}, cpErr
	}

	err = json.Unmarshal([]byte(body), &kubeCluster.WorkerPools)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while fetching cluster", 500)
		return KubeClusterStatus{}, cpErr
	}
	for index, poolId := range kubeCluster.WorkerPools {
		{
			nodes, err_ := cloud.fetchNodes(cluster, poolId.ID, ctx, companyId)
			if err_ != (types.CustomCPError{}) {
				return KubeClusterStatus{}, err_
			}
			kubeCluster.WorkerPools[index].Nodes = nodes
		}
	}
	return kubeCluster, types.CustomCPError{}
}

func (cloud *IBM) GetAllInstances(ctx utils.Context) (AllInstancesResponse, types.CustomCPError) {
	url := models.IBM_All_Instances_Endpoint + cloud.Region + "&provider=vpc-classic"

	req, _ := utils.CreateGetRequest(url)
	req.Close = true
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
		cpErr := ApiError(errors.New(string(body)), "Error occurred while getting machine types.", 512)
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
func (cloud *IBM) fetchNodes(cluster *Cluster_Def, poolId string, ctx utils.Context, companyId string) ([]KubeWorkerNodesStatus, types.CustomCPError) {

	req, _ := utils.CreateGetRequest(models.IBM_Kube_GetNodes_Endpoint + "?cluster=" + cluster.ClusterId + "&pool=" + poolId)
	req.Close = true
	m := make(map[string]string)

	m["Content-Type"] = "application/json"
	m["Accept"] = "application/json"
	m["Authorization"] = cloud.IAMToken
	//m["X-Region"] = cloud.Region
	m["X-Auth-Resource-Group"] = cluster.ResourceGroup
	utils.SetHeaders(req, m)

	client := utils.InitReq()

	res, err := client.SendRequest(req)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while fetching cluster", 500)
		return []KubeWorkerNodesStatus{}, cpErr
	}

	defer res.Body.Close()

	beego.Info(res.Status)

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while fetching cluster", 500)
		return []KubeWorkerNodesStatus{}, cpErr
	}

	beego.Info(string(body))

	if res.StatusCode != 200 {
		ctx.SendLogs(errors.New(string(body)).Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(errors.New(string(body)), "error occurred while fetching cluster", 512)
		return []KubeWorkerNodesStatus{}, cpErr
	}

	var response []KubeWorkerNodesStatus
	err = json.Unmarshal([]byte(body), &response)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cpErr := ApiError(err, "error occurred while fetching cluster", 500)
		return []KubeWorkerNodesStatus{}, cpErr
	}
	return response, types.CustomCPError{}
}
