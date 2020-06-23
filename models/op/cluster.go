package op

import (
	"antelope/models"
	"antelope/models/db"
	"antelope/models/key_utils"
	rbac_athentication "antelope/models/rbac_authentication"
	"antelope/models/utils"
	"antelope/models/vault"
	"errors"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Cluster_Def struct {
	ID               bson.ObjectId `json:"-" bson:"_id,omitempty"`
	ProjectId        string        `json:"project_id" bson:"project_id" validate:"required" description:"ID of project [required]"`
	Kube_Credentials interface{}   `json:"-" bson:"kube_credentials"`
	Name             string        `json:"name" bson:"name" validate:"required" description:"Name of cluster [required]"`
	Status           models.Type   `json:"status" bson:"status" validate:"eq=New|eq=new|eq=Cluster Creation Failed" description:"Cluster status can be New, Cluster Created, Cluster Terminated. By default value will be 'New' [readonly]"`
	Cloud            models.Cloud  `json:"-" bson:"cloud"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	NodePools        []*NodePool   `json:"node_pools" bson:"node_pools" validate:"required,dive"`
	CompanyId        string        `json:"company_id" bson:"company_id" description:"ID of company which you are belong to [optional]"`
	TokenName        string        `json:"-" bson:"token_name"`
}

type NodePool struct {
	ID        bson.ObjectId      `json:"-" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name" validate:"required" description:"Name of node pool [required]"`
	NodeCount int64              `json:"node_count" bson:"node_count" validate:"required,gte=0" description:"Count of node pool [required]"`
	Nodes     []*Node            `json:"nodes" bson:"nodes" validate:"required,dive"`
	KeyInfo   key_utils.AZUREKey `json:"key_info" bson:"key_info" validate:"required,dive"`
	PoolRole  models.PoolRole    `json:"pool_role" bson:"pool_role" validate:"required" description:"Pool role can be master or slave [required]"`
}

type Node struct {
	Name      string `json:"name" bson:"name,omitempty" validate:"required" description:"Name of node [required]"`
	PrivateIP string `json:"private_ip" bson:"private_ip,omitempty" description:"Private IP of node [readonly]"`
	PublicIP  string `json:"public_ip" bson:"public_ip,omitempty" description:"Public IP of node [readonly]"`
	UserName  string `json:"user_name" bson:"user_name,omitempty" validate:"required" description:"User name which will be used for ssh into machine [required]"`
}
type Cluster struct {
	Name      string      `json:"name,omitempty" bson:"name,omitempty" description:"Cluster name"`
	ProjectId string      `json:"project_id" bson:"project_id"  description:"ID of project"`
	Status    models.Type `json:"status,omitempty" bson:"status,omitempty" description:"Status of cluster"`
}

func GetCluster(projectId, companyId string, ctx utils.Context) (cluster Cluster_Def, err error) {

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Cluster model: Get - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoOPClusterCollection)
	err = c.Find(bson.M{"project_id": projectId, "company_id": companyId}).One(&cluster)
	if err != nil {
		ctx.SendLogs("Cluster model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return Cluster_Def{}, err
	}
	return cluster, nil
}
func GetAllCluster(ctx utils.Context, input rbac_athentication.List) (opClusters []Cluster, err error) {
	var clusters []Cluster_Def
	var copyData []string
	for _, d := range input.Data {
		copyData = append(copyData, d)
	}

	session, err1 := db.GetMongoSession(ctx)
	if err1 != nil {
		ctx.SendLogs("Cluster model: GetAll - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return nil, err1
	}

	defer session.Close()
	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoOPClusterCollection)
	err = c.Find(bson.M{"project_id": bson.M{"$in": copyData}, "company_id": ctx.Data.Company}).All(&clusters)
	if err != nil {
		ctx.SendLogs("Cluster model: GetAll - Got error while connecting to the database: "+err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return nil, err
	}

	return opClusters, nil
}

func checkMasterPools(cluster Cluster_Def) error {
	noOfMasters := 0
	for _, pools := range cluster.NodePools {
		if pools.PoolRole == models.Master {
			noOfMasters += 1
			if noOfMasters == 2 {
				return errors.New("Cluster can't have more than 1 master")
			}
		}
	}
	return nil
}
func CreateCluster(cluster Cluster_Def, ctx utils.Context, token string, teams string) error {
	_, err := GetCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err == nil { //cluster found
		ctx.SendLogs("Cluster model: Create - Cluster  already exists in the database: "+cluster.Name, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("Cluster model: Create - Cluster  already exists in the database: " + cluster.Name)
	}
	/*
		inserting key in vault
	**/
	for index, pool := range cluster.NodePools {
		pool.KeyInfo.KeyName = pool.Name
		_, err := vault.PostSSHKey(pool.KeyInfo, pool.Name, models.OP, ctx, token, teams, "")
		if err != nil {
			ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
		for i, nodes := range pool.Nodes {
			cluster.NodePools[index].Nodes[i].PrivateIP = nodes.PublicIP
		}
	}

	mc := db.GetMongoConf()
	err = db.InsertInMongo(mc.MongoOPClusterCollection, cluster)
	if err != nil {
		ctx.SendLogs("Cluster model: Create - Got error inserting cluster to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}
func UpdateCluster(cluster Cluster_Def, update bool, ctx utils.Context, teams, token string) error {
	oldCluster, err := GetCluster(cluster.ProjectId, cluster.CompanyId, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Update - Cluster   does not exist in the database: "+cluster.Name+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	if oldCluster.Status == (models.Deploying) && update {
		ctx.SendLogs("cluster is in deploying state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("cluster is in deploying state")
	}
	if oldCluster.Status == (models.Terminating) && update {
		ctx.SendLogs("cluster is in terminating state", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("cluster is in terminating state")
	}

	if oldCluster.Status == "Cluster Created" && update {
		ctx.SendLogs("Cluster is in runnning state ", models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return errors.New("Cluster is in runnning state")

	}
	err = DeleteCluster(cluster.ProjectId, cluster.CompanyId, ctx, token)
	if err != nil {
		ctx.SendLogs("Cluster model: Update - Got error deleting cluster: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	cluster.CreationDate = oldCluster.CreationDate
	cluster.ModificationDate = time.Now()

	err = CreateCluster(cluster, ctx, token, teams)
	if err != nil {
		ctx.SendLogs("Cluster model: Update - Got error deleting cluster: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}
func DeleteCluster(projectId, companyId string, ctx utils.Context, token string) error {
	oldCluster, err := GetCluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Update - Cluster   does not exist in the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	session, err := db.GetMongoSession(ctx)
	if err != nil {

		ctx.SendLogs("Cluster model: Delete - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	defer session.Close()
	for _, pool := range oldCluster.NodePools {
		err := vault.DeleteSSHkey(string(models.OP), pool.Name, token, ctx, "")
		if err != nil {
			ctx.SendLogs("Cluster model: Delete - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
			return err
		}
	}

	mc := db.GetMongoConf()
	c := session.DB(mc.MongoDb).C(mc.MongoOPClusterCollection)
	err = c.Remove(bson.M{"project_id": projectId, "company_id": companyId})
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}

	return nil
}
func PrintError(confError error, name, projectId string, ctx utils.Context, companyId string) {
	if confError != nil {
		ctx.SendLogs(confError.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		utils.SendLog(companyId, "Cluster creation failed : "+name, "error", projectId)
		utils.SendLog(companyId, confError.Error(), "error", projectId)

	}
}
func CheckCluster(projectId, companyId string, ctx utils.Context) error {
	cluster, err := GetCluster(projectId, companyId, ctx)
	if err != nil {
		ctx.SendLogs("Cluster model: Get - Got error while connecting to the database: "+err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	agent, err := GetGrpcAgentConnection(ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	err = agent.InitializeAgentClient(projectId, companyId)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return err
	}
	for _, pool := range cluster.NodePools {
		for _, node := range pool.Nodes {
			err = agent.ExecCommand(*node, pool.KeyInfo.PrivateKey, ctx)
			if err != nil {
				ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
				return err
			}
		}
	}

	return nil
}
