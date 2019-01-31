package networks

import (
	"strings"
	"antelope/models/utils"
	"io/ioutil"
	"github.com/astaxie/beego"
	"encoding/json"
	"antelope/models"
	"time"
	"gopkg.in/mgo.v2/bson"
)

var (
	networkHost    = beego.AppConfig.String("network_url")
)
type AWSNetwork struct {
	EnvironmentId    string        `json:"environment_id" bson:"environment_id"`
	Name             string        `json:"name" bson:"name"`
	Type             models.Type   `json:"type" bson:"type"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"`
	NetworkStatus    string        `json:"status" bson:"status"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	Definition       []*AWSDefinition `json:"definition" bson:"definition"`
}

type AWSDefinition struct {
	ID             bson.ObjectId    `json:"_id" bson:"_id,omitempty"`
	Vpc            Vpc              `json:"vpc" bson:"vpc"`
	Subnets        []*Subnet        `json:"subnets" bson:"subnets"`
	SecurityGroups []*SecurityGroup `json:"security_groups" bson:"security_groups"`
}

type Vpc struct {
	ID    bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	VpcId string        `json:"vpc_id" bson:"vpc_id"`
	Name  string        `json:"name" bson:"name"`
	CIDR  string        `json:"cidr" bson:"cidr"`
}

type Subnet struct {
	ID       bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	SubnetId string        `json:"subnet_id" bson:"subnet_id"`
	Name     string        `json:"name" bson:"name"`
	CIDR  	 string        `json:"cidr" bson:"cidr"`
}

type SecurityGroup struct {
	ID              bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	SecurityGroupId string        `json:"security_group_id" bson:"security_group_id"`
	Name            string        `json:"name" bson:"name"`
	Description     string        `json:"description" bson:"description"`
}

type AzureNetwork struct {
	EnvironmentId    string        `json:"environment_id" bson:"environment_id"`
	Name             string        `json:"name" bson:"name"`
	Type             models.Type   `json:"type" bson:"type"`
	Cloud            models.Cloud  `json:"cloud" bson:"cloud"`
	NetworkStatus    string        `json:"status" bson:"status"`
	CreationDate     time.Time     `json:"-" bson:"creation_date"`
	ModificationDate time.Time     `json:"-" bson:"modification_date"`
	Definition       []*AzureDefinition `json:"definition" bson:"definition"`
}

type AzureDefinition struct {
	ID             bson.ObjectId    `json:"_id" bson:"_id,omitempty"`
	Vnet            VNet              `json:"vnet" bson:"vnet"`
	Subnets        []*Subnet        `json:"subnets" bson:"subnets"`
	SecurityGroups []*SecurityGroup `json:"security_groups" bson:"security_groups"`
}

type VNet struct {
	ID    bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	VnetId string        `json:"vnet_id" bson:"vnet_id"`
	Name  string        `json:"name" bson:"name"`
	CIDR  string        `json:"cidr" bson:"cidr"`
}

func GetNetworkStatus(envId string ,cloudType string) ( interface{}, error){

	networkUrl := strings.Replace(networkHost,"{cloud_provider}",cloudType,-1)
	client := utils.InitReq()

	req , err :=utils.CreateGetRequest(envId, networkUrl)
	if err != nil {
		beego.Error("%s", err)
		return AWSNetwork{}, err
	}

	response, err := client.SendRequest(req)
	if err != nil {
		beego.Error("%s", err)
		return AWSNetwork{}, err
	}
	defer response.Body.Close()
	if cloudType == "aws" {
		var network AWSNetwork

		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			beego.Error("%s", err)
			return AWSNetwork{}, err
		}

		err = json.Unmarshal(contents, &network)
		if err != nil {
			beego.Error("%s", err)
			return AWSNetwork{}, err
		}
		return network, nil
	}else if cloudType == "azure"{
		var network AzureNetwork
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			beego.Error("%s", err)
			return AzureNetwork{}, err
		}

		err = json.Unmarshal(contents, &network)
		if err != nil {
			beego.Error("%s", err)
			return AzureNetwork{}, err
		}
		return network, nil
	}
	return nil,err
}