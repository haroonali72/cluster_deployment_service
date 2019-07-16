package networks

import (
	"antelope/models"
	"antelope/models/logging"
	"antelope/models/utils"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"strings"
	"time"
)

type GCPNetwork struct {
	Definition []*AWSDefinition `json:"definition" bson:"definition"`
}

type AWSNetwork struct {
	EnvironmentId    string           `json:"project_id" bson:"project_id"`
	Name             string           `json:"name" bson:"name"`
	Type             models.Type      `json:"type" bson:"type"`
	Cloud            models.Cloud     `json:"cloud" bson:"cloud"`
	NetworkStatus    string           `json:"status" bson:"status"`
	CreationDate     time.Time        `json:"-" bson:"creation_date"`
	ModificationDate time.Time        `json:"-" bson:"modification_date"`
	Definition       []*AWSDefinition `json:"definition" bson:"definition"`
}

type AWSDefinition struct {
	Vpc            Vpc              `json:"vpc" bson:"vpc"`
	Subnets        []*Subnet        `json:"subnets" bson:"subnets"`
	SecurityGroups []*SecurityGroup `json:"security_groups" bson:"security_groups"`
}

type Vpc struct {
	VpcId string `json:"vpc_id" bson:"vpc_id"`
	Name  string `json:"name" bson:"name"`
	CIDR  string `json:"cidr" bson:"cidr"`
}

type Subnet struct {
	SubnetId string `json:"subnet_id" bson:"subnet_id"`
	Name     string `json:"name" bson:"name"`
	CIDR     string `json:"cidr" bson:"cidr"`
}

type SecurityGroup struct {
	SecurityGroupId string `json:"security_group_id" bson:"security_group_id"`
	Name            string `json:"name" bson:"name"`
	Description     string `json:"description" bson:"description"`
}

type AzureNetwork struct {
	EnvironmentId    string             `json:"environment_id" bson:"environment_id"`
	Name             string             `json:"name" bson:"name"`
	Type             models.Type        `json:"type" bson:"type"`
	Cloud            models.Cloud       `json:"cloud" bson:"cloud"`
	NetworkStatus    string             `json:"status" bson:"status"`
	CreationDate     time.Time          `json:"-" bson:"creation_date"`
	ModificationDate time.Time          `json:"-" bson:"modification_date"`
	Definition       []*AzureDefinition `json:"definition" bson:"definition"`
}

type AzureDefinition struct {
	Vnet           VNet             `json:"vnet" bson:"vnet"`
	Subnets        []*Subnet        `json:"subnets" bson:"subnets"`
	SecurityGroups []*SecurityGroup `json:"security_groups" bson:"security_groups"`
}

type VNet struct {
	ID     bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	VnetId string        `json:"vnet_id" bson:"vnet_id"`
	Name   string        `json:"name" bson:"name"`
	CIDR   string        `json:"cidr" bson:"cidr"`
}

func GetAPIStatus(host, projectId string, cloudType string, ctx logging.Context) (interface{}, error) {

	if strings.Contains(host, "{cloud_provider}") {
		host = strings.Replace(host, "{cloud_provider}", cloudType, -1)
	}

	client := utils.InitReq()

	url := host + "/" + projectId
	req, err := utils.CreateGetRequest(url)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return nil, err
	}

	response, err := client.SendRequest(req)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return nil, err
	}
	defer response.Body.Close()
	//	var network AzureNetwork
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return nil, err
	}

	return contents, nil

}
