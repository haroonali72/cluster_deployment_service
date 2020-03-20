package doks

import (
	"antelope/models"
	"antelope/models/db"
	rbacAuthentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"antelope/models/vault"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

type KubernetesClusterConfig struct{
	KubeconfigYAML []byte
}
type KubernetesCluster struct {
	ID            		string   							`json:"id" bson:"id"`
	ProjectId			string								`json:"project_id" bson:"project_id" valid:"required"`
	CompanyId			string								`json:"company_id" bson:"company_id" valid:"required"`
	Cloud            	models.Cloud  						`json:"cloud" bson:"cloud" valid:"required"`
	CreationDate     	time.Time     						`json:"-" bson:"creation_date"`
	ModificationDate 	time.Time     						`json:"-" bson:"modification_date"`
	CloudplexStatus  	string        						`json:"status" bson:"status"`
	Name          		string   							`json:"name,omitempty" bson:"name" valid:"required"`
	Region    			string  							`json:"region,omitempty" bson:"region"`
	Version		  		string   							`json:"version,omitempty" bson:"version"`
	ClusterSubnet 		string   							`json:"cluster_subnet,omitempty" bson:"cluster_subnet"`
	ServiceSubnet 		string   							`json:"service_subnet,omitempty" bson:"service_subnet"`
	IPv4          		string   							`json:"ipv4,omitempty" bson:"ivp4"`
	Endpoint      		string   							`json:"endpoint,omitempty" bson:"endpoint"`
	Tags          		[]string 							`json:"tags,omitempty" bson:"tags"`
	VPCUUID       		string   							`json:"vpc_uuid" bson:"vpc_uuid"`
	NodePools 			[]*KubernetesNodePool 				`json:"node_pools,omitempty" bson:"node_pools"`
	MaintenancePolicy 	*KubernetesMaintenancePolicy 		`json:"maintenance_policy,omitempty" bson:"maintenance_policy"`
	AutoUpgrade       	bool                         		`json:"auto_upgrade,omitempty" bson:"auto_upgrade"`
	Status   	 		*KubernetesClusterStatus 			`json:"kube_status,omitempty" bson:"kube_status"`
}
type KubernetesNodePool struct {
	ID        	string            		`json:"id,omitempty"  bson:"id"`
	Name     	string            		`json:"name,omitempty"  bson:"name"`
	Size      	string            		`json:"size,omitempty"  bson:"size"`
	Count     	int               		`json:"count,omitempty"  bson:"count"`
	Tags      	[]string          		`json:"tags,omitempty"  bson:"tags"`
	Labels    	map[string]string 		`json:"labels,omitempty"  bson:"labels"`
	AutoScale 	bool             		`json:"auto_scale,omitempty"  bson:"auto_scale"`
	MinNodes  	int               		`json:"min_nodes,omitempty"  bson:"min_nodes"`
	MaxNodes  	int               		`json:"max_nodes,omitempty"  bson:"max_nodes"`
	Nodes 		[]*KubernetesNode 		`json:"nodes,omitempty"  bson:"nodes"`
}
type KubernetesNode struct {
	ID        	string                	`json:"id,omitempty" bson:"id"`
	Name      	string                	`json:"name,omitempty" bson:"name"`
	Status    	*KubernetesNodeStatus 	`json:"status,omitempty" bson:"status"`
	DropletID 	string                	`json:"droplet_id,omitempty" bson:"droplet_id"`
	CreatedAt 	time.Time 				`json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt 	time.Time 				`json:"updated_at,omitempty" bson:"updated_at"`
}
type KubernetesNodeSize struct {
	Name string `json:"name" bson:"name"`
	Slug string `json:"slug" bson:"slug"`
}
type KubernetesRegion struct {
	Name string `json:"name" bson:"name"`
	Slug string `json:"slug" bson:"slug"`
}
type KubernetesMaintenancePolicy struct {
	StartTime 	string                      `json:"start_time" bson:"start_time"`
	Duration  	string                      `json:"duration" bson:"duration"`
	Day       	string 						`json:"day" bson:"day"`
}
type KubernetesClusterStatus struct {
	State   	string 						 `json:"state,omitempty" bson:"state"`
	Message 	string                       `json:"message,omitempty" bson:"message"`
}
type KubernetesNodeStatus struct {
	State   	string 						`json:"state,omitempty" bson:"state"`
	Message 	string 						`json:"message,omitempty" bson:"message"`
}



func GetKubernetesCluster(projectId string, companyId string, ctx utils.Context) (cluster KubernetesCluster, err error) {
	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs( "DOKSGetClusterModel:  Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err1
	}
	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoDOKSClusterCollection)
	err = c.Find(bson.M{"project_id": projectId, "company_id": companyId}).One(&cluster)
	if err != nil {
		ctx.SendLogs("DOKSGetClusterModel:  Get - Got error while fetching from database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging )
		return cluster, err
	}

	return cluster, nil
}
func GetAllKubernetesCluster(data rbacAuthentication.List, ctx utils.Context) (clusters []KubernetesCluster, err error) {
	var copyData []string
	for _, d := range data.Data {
		copyData = append(copyData, d)
	}

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("DOKSGetAllClusterModel:  GetAll - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging )
		return clusters, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoDOKSClusterCollection)
	err = c.Find(bson.M{"project_id": bson.M{"$in": copyData}}).All(&clusters)
	if err != nil {
		ctx.SendLogs("DOKSGetAllClusterModel:  GetAll - Got error while fetching from database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return clusters, err
	}

	return clusters, nil
}
func AddKubernetesCluster(cluster KubernetesCluster, ctx utils.Context) error {
	_, err := GetKubernetesCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err == nil {
		text := fmt.Sprintf("DOKSAddClusterModel:  Add - Cluster for project '%s' already exists in the database.", cluster.ProjectId)
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs("DOKSAddClusterModel:  Add - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging )
		return err
	}
	defer session.Close()

	if cluster.CreationDate.IsZero() {
		cluster.CreationDate = time.Now()
		cluster.ModificationDate = time.Now()
		if cluster.CloudplexStatus == "" {
			cluster.CloudplexStatus = "new"
		}
		cluster.Cloud = models.DOKS
	}

	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoDOKSClusterCollection, cluster)
	if err != nil {
		ctx.SendLogs("DOKSAddClusterModel:  Add - Got error while inserting cluster to the database:  "+err.Error(), models.LOGGING_LEVEL_ERROR,models.Backend_Logging)
		return err
	}

	return nil
}
func UpdateKubernetesCluster(cluster KubernetesCluster, ctx utils.Context) error {
	oldCluster, err := GetKubernetesCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err != nil {
		text := "DOKSUpdateClusterModel:  Update - Cluster '" + cluster.Name + "' does not exist in the database: " + err.Error()
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New(text)
	}

	if oldCluster.CloudplexStatus == string(models.Deploying) {
		ctx.SendLogs("DOKSUpdateClusterModel:  Update - Cluster is in deploying state.", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("cluster is in deploying state")
	}
	if oldCluster.CloudplexStatus == string(models.Terminating) {
		ctx.SendLogs("DOKSUpdateClusterModel:  Update - Cluster is in terminating state.", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("cluster is in terminating state")
	}
	if strings.ToLower(oldCluster.CloudplexStatus) == strings.ToLower(string(models.ClusterCreated)) {
		ctx.SendLogs("DOKSUpdateClusterModel:  Update - Cluster is in running state.", models.LOGGING_LEVEL_ERROR, models.Backend_Logging,)
		return errors.New("cluster is in running state")
	}
cluster.ProjectId="string"
	err = DeleteKubernetesCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err != nil {
		ctx.SendLogs("DOKSUpdateClusterModel:  Update - Got error deleting cluster "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	cluster.CreationDate = oldCluster.CreationDate
	cluster.ModificationDate = time.Now()

	err = AddKubernetesCluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs("DOKSUpdateClusterModel:  Update - Got error creating cluster "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}
func DeleteKubernetesCluster(projectId, companyId string, ctx utils.Context) error {
	session, err := db.GetMongoSession(ctx)
	if err != nil {
		ctx.SendLogs("DOKSDeleteClusterModel:  Delete - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoDOKSClusterCollection)
	err = c.Remove(bson.M{"project_id": projectId, "company_id": companyId})
	if err != nil {
		ctx.SendLogs("DOKSDeleteClusterModel:  Delete - Got error while deleting from the database: "+err.Error(), models.LOGGING_LEVEL_ERROR,models.Backend_Logging)
		return err
	}

	return nil
}
func PrintError(confError error, name, projectId string, companyId string) {
	if confError != nil {
		beego.Error(confError.Error())
		_, _ = utils.SendLog(companyId, "Cluster creation failed : "+name, "error", projectId)
		_, _ = utils.SendLog(companyId, confError.Error(), "error", projectId)
	}
}
func DeployKubernetesCluster(cluster KubernetesCluster, credentials vault.DOCredentials, companyId string, token string, ctx utils.Context, ) (confError error) {

	publisher := utils.Notifier{}
	confError = publisher.Init_notifier()

	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
		ctx.SendLogs(confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return confError
	}

	doksOps, err := GetDOKS(credentials)
	if err != nil {
		ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	err = doksOps.init(ctx)
	if err != nil {
		ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.Status.State = "Cluster creation failed"
		confError = UpdateKubernetesCluster(cluster, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}

	_, _ = utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	_,confError = doksOps.createCluster(cluster,  ctx,companyId,token)
	if confError != nil {
		ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		PrintError(confError, cluster.Name, cluster.ProjectId, companyId)

		cluster.Status.State = "Cluster creation failed"
		confError = UpdateKubernetesCluster(cluster, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return nil
	}

	cluster.Status.State = "Cluster Created"

	confError = UpdateKubernetesCluster(cluster, ctx)
	if confError != nil {
		PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
		ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return confError
	}

	_, _ = utils.SendLog(companyId, "Cluster created successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	return nil
}
func FetchStatus(credentials vault.DOCredentials, token, projectId, companyId string, ctx utils.Context) (KubernetesCluster, error) {
	cluster, err := GetKubernetesCluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel:  Fetch -  Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err
	}

	gkeOps, err := GetDOKS(credentials)
	if err != nil {
		ctx.SendLogs("GKEClusterModel:  Fetch -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err
	}

	err = gkeOps.init(ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel:  Fetch -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err
	}

	//_,err = gkeOps.fetchStatus(&cluster, ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel:  Fetch - Failed to get latest status "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err
	}

	return cluster, nil
}
func TerminateCluster(credentials vault.DOCredentials, projectId, companyId string, ctx utils.Context) error {
	publisher := utils.Notifier{}
	pubErr := publisher.Init_notifier()
	if pubErr != nil {
		ctx.SendLogs("GKEClusterModel:  Terminate -"+pubErr.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return pubErr
	}

	cluster, err := GetKubernetesCluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Terminate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	if cluster.Status.State == "" || cluster.Status.State == "new" {
		text := "GKEClusterModel : Terminate - Cannot terminate a new cluster"
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return errors.New(text)
	}

	gkeOps, err := GetDOKS(credentials)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	cluster.Status.State = string(models.Terminating)
	_, _ = utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	err = gkeOps.init(ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Terminate -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.Status.State = "Cluster Termination Failed"
		err = UpdateKubernetesCluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("GKEClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			_, _ = utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			_, _ = utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
			return err
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}

	_,err = gkeOps.deleteCluster(cluster, ctx,projectId,companyId)
	if err != nil {
		_, _ = utils.SendLog(companyId, "Cluster termination failed: "+cluster.Name, "error", cluster.ProjectId)

		cluster.Status.State = "Cluster Termination Failed"
		err = UpdateKubernetesCluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("GKEClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			_, _ = utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			_, _ = utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
			publisher.Notify(cluster.ProjectId, "Status Available", ctx)
			return err
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return nil
	}

	cluster.Status.State = "Cluster Terminated"

	err = UpdateKubernetesCluster(cluster, ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Terminate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		_, _ = utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
		_, _ = utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}
	_, _ = utils.SendLog(companyId, "Cluster terminated successfully "+cluster.Name, "info", cluster.ProjectId)
	publisher.Notify(cluster.ProjectId, "Status Available", ctx)
	return nil
}
func GetServerConfig(credentials vault.DOCredentials, ctx utils.Context) ( KubernetesClusterConfig, error) {
	gkeOps, err := GetDOKS(credentials)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : GetServerConfig - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubernetesClusterConfig{}, err
	}

	err = gkeOps.init(ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : GetServerConfig -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KubernetesClusterConfig{}, err
	}

	return KubernetesClusterConfig{},nil
}
/*
func TestDOKS(credentials vault.DOCredentials, companyId string, token string, ctx utils.Context, projetcID string, clusterName string) (confError error) {

	data2, err := woodpecker.GetCertificate(projetcID, token, ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Apply Agent -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	filePath := "/tmp/" + companyId + "/" + projetcID + "/"
	cmd := "mkdir -p " + filePath + " && echo '" + data2 + "'>" + filePath + "agent.yaml && echo '" + credentials.RawData + "'>" + filePath + "gcp-auth.json"
	output, err := remoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return
	}

	if credentials.Zone != "" {
		cmd = "sudo docker run --rm --name " + companyId + projetcID + " -e gcpProject=" + credentials.AccountData.ProjectId + " -e cluster=" + clusterName + " -e zone=" + credentials.Region + "-" + credentials.Zone + " -e serviceAccount=" + filePath + "gcp-auth.json" + " -e yamlFile=" + filePath + "agent.yaml -v " + filePath + ":" + filePath + " " + models.GCPAuthContianrName
	} else {
		cmd = "sudo docker run --rm --name " + companyId + projetcID + " -e gcpProject=" + credentials.AccountData.ProjectId + " -e cluster=" + clusterName + " -e region=" + credentials.Region + " -e serviceAccount=" + filePath + "gcp-auth.json" + " -e yamlFile=" + filePath + "agent.yaml -v " + filePath + ":" + filePath + " " + models.GCPAuthContianrName
	}

	output, err = remoteRun("ubuntu", beego.AppConfig.String("jump_host_ip"), beego.AppConfig.String("jump_host_ssh_key"), cmd)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Apply Agent -"+err.Error()+output, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return
	}
	return nil
}
func remoteRun(user string, addr string, privateKey string, cmd string) (string, error) {
	clientPem, err := ioutil.ReadFile(privateKey)
	if err != nil {
		return "", err
	}

	key, err := ssh.ParsePrivateKey(clientPem)
	if err != nil {
		return "", err
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	client, err := ssh.Dial("tcp", net.JoinHostPort(addr, "22"), config)
	if err != nil {
		return "", err
	}
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	var b bytes.Buffer
	session.Stdout = &b
	err = session.Run(cmd)
	return b.String(), err
}

 */
func GetDOKS(credentials vault.DOCredentials) (DOKS, error) {
	return DOKS{
		AccessKey : credentials.AccessKey,
		Region   : credentials.Region,
	}, nil
}
