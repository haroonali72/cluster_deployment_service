package doks

import (
	"antelope/models"
	"antelope/models/db"
	"antelope/models/utils"
	"antelope/models/vault"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
	rbacAuthentication "antelope/models/rbac_authentication"
)

type KubernetesClusterConfig struct{
	KubeconfigYAML []byte
}
type KubernetesCluster struct {
	ID            		string   							`json:"id,omitempty"`
	ProjectId			string								`json:"id,omitempty"`
	CompanyId			string								`json:"id,omitempty"`
	Cloud            	models.Cloud  						`json:"cloud" bson:"cloud"`
	CreationDate     	time.Time     						`json:"-" bson:"creation_date"`
	ModificationDate 	time.Time     						`json:"-" bson:"modification_date"`
	CloudplexStatus  	string        						`json:"status" bson:"status"`
	Name          		string   							`json:"name,omitempty"`
	RegionSlug    		string  							`json:"region,omitempty"`
	VersionSlug  		string   							`json:"version,omitempty"`
	ClusterSubnet 		string   							`json:"cluster_subnet,omitempty"`
	ServiceSubnet 		string   							`json:"service_subnet,omitempty"`
	IPv4          		string   							`json:"ipv4,omitempty"`
	Endpoint      		string   							`json:"endpoint,omitempty"`
	Tags          		[]string 							`json:"tags,omitempty"`
	VPCUUID       		string   							`json:"vpc_uuid,omitempty"`
	NodePools 			[]*KubernetesNodePool 				`json:"node_pools,omitempty"`
	MaintenancePolicy 	*KubernetesMaintenancePolicy 		`json:"maintenance_policy,omitempty"`
	AutoUpgrade       	bool                         		`json:"auto_upgrade,omitempty"`
	Status   	 		*KubernetesClusterStatus 			`json:"status,omitempty"`
	CreatedAt 			time.Time                			`json:"created_at,omitempty"`
	UpdatedAt 			time.Time                			`json:"updated_at,omitempty"`
}

type KubernetesNodePool struct {
	ID        	string            		`json:"id,omitempty"`
	Name     	string            		`json:"name,omitempty"`
	Size      	string            		`json:"size,omitempty"`
	Count     	int               		`json:"count,omitempty"`
	Tags      	[]string          		`json:"tags,omitempty"`
	Labels    	map[string]string 		`json:"labels,omitempty"`
	AutoScale 	bool             		`json:"auto_scale,omitempty"`
	MinNodes  	int               		`json:"min_nodes,omitempty"`
	MaxNodes  	int               		`json:"max_nodes,omitempty"`

	Nodes 		[]*KubernetesNode 		`json:"nodes,omitempty"`
}
type KubernetesNode struct {
	ID        string                `json:"id,omitempty"`
	Name      string                `json:"name,omitempty"`
	Status    *KubernetesNodeStatus `json:"status,omitempty"`
	DropletID string                `json:"droplet_id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
type KubernetesMaintenancePolicy struct {
	StartTime string                         `json:"start_time"`
	Duration  string                         `json:"duration"`
	Day       KubernetesMaintenancePolicyDay `json:"day"`
}
type KubernetesClusterStatus struct {
	State   KubernetesClusterStatusState `json:"state,omitempty"`
	Message string                       `json:"message,omitempty"`
}
type KubernetesNodeStatus struct {
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
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

	err = doksOps.init()
	if err != nil {
		ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.Status = "Cluster creation failed"
		confError = UpdateDOKSCluster(cluster, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}

	_, _ = utils.SendLog(companyId, "Creating Cluster : "+cluster.Name, "info", cluster.ProjectId)
	confError = doksOps.CreateCluster(cluster, token, ctx)

	if confError != nil {
		ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		PrintError(confError, cluster.Name, cluster.ProjectId, companyId)

		cluster.Status = "Cluster creation failed"
		confError = UpdateDOKSCluster(cluster, ctx)
		if confError != nil {
			PrintError(confError, cluster.Name, cluster.ProjectId, companyId)
			ctx.SendLogs("DOKSDeployClusterModel:  Deploy - "+confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		}

		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return nil
	}

	cluster.Status = "Cluster Created"

	confError = UpdateDOKSCluster(cluster, ctx)
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
	cluster, err := GetGKECluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel:  Fetch -  Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err
	}

	gkeOps, err := GetGKE(credentials)
	if err != nil {
		ctx.SendLogs("GKEClusterModel:  Fetch -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err
	}

	err = gkeOps.init()
	if err != nil {
		ctx.SendLogs("GKEClusterModel:  Fetch -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return cluster, err
	}

	err = gkeOps.fetchClusterStatus(&cluster, ctx)
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

	cluster, err := GetGKECluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Terminate - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	if cluster.Status == "" || cluster.Status == "new" {
		text := "GKEClusterModel : Terminate - Cannot terminate a new cluster"
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return errors.New(text)
	}

	gkeOps, err := GetGKE(credentials)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Terminate - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	cluster.Status = string(models.Terminating)
	_, _ = utils.SendLog(companyId, "Terminating cluster: "+cluster.Name, "info", cluster.ProjectId)

	err = gkeOps.init()
	if err != nil {
		ctx.SendLogs("GKEClusterModel : Terminate -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		cluster.Status = "Cluster Termination Failed"
		err = UpdateGKECluster(cluster, ctx)
		if err != nil {
			ctx.SendLogs("GKEClusterModel : Terminate - Got error while connecting to the database:"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			_, _ = utils.SendLog(companyId, "Error in cluster updation in mongo: "+cluster.Name, "error", cluster.ProjectId)
			_, _ = utils.SendLog(companyId, err.Error(), "error", cluster.ProjectId)
			return err
		}
		publisher.Notify(cluster.ProjectId, "Status Available", ctx)
		return err
	}

	err = gkeOps.deleteCluster(cluster, ctx)
	if err != nil {
		_, _ = utils.SendLog(companyId, "Cluster termination failed: "+cluster.Name, "error", cluster.ProjectId)

		cluster.Status = "Cluster Termination Failed"
		err = UpdateGKECluster(cluster, ctx)
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

	cluster.Status = "Cluster Terminated"

	err = UpdateGKECluster(cluster, ctx)
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
func GetServerConfig(credentials vault.DOCredentials, ctx utils.Context) (*doks.ServerConfig, error) {
	gkeOps, err := GetGKE(credentials)
	if err != nil {
		ctx.SendLogs("GKEClusterModel : GetServerConfig - "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	err = gkeOps.init()
	if err != nil {
		ctx.SendLogs("GKEClusterModel : GetServerConfig -"+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	return gkeOps.getGKEVersions(ctx)
}
