package types

import (
	"antelope/models"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type GCPNetwork struct {
	Name       string           `json:"name" bson:"name"`
	Definition []*AWSDefinition `json:"definition" bson:"definition"`
	IsPrivate  bool             `json:"is_private" bson:"is_private"`
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
	IsPrivate        bool             `json:"is_private" bson:"is_private"`
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
	Link  string `json:"link" bson:"link"`
}

type Subnet struct {
	SubnetId string `json:"subnet_id" bson:"subnet_id"`
	Name     string `json:"name" bson:"name"`
	CIDR     string `json:"cidr" bson:"cidr"`
	Link     string `json:"link" bson:"link"`
	Zone     string `json:"availability_zone" bson:"availability_zone"`
}

type SecurityGroup struct {
	SecurityGroupId string `json:"security_group_id" bson:"security_group_id"`
	Name            string `json:"name" bson:"name"`
	Description     string `json:"description" bson:"description"`
	Link            string `json:"link" bson:"link"`
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
	IsPrivate        bool               `json:"is_private" bson:"is_private"`
}

type AzureDefinition struct {
	Vnet           VNet             `json:"vpc" bson:"vpc"`
	Subnets        []*Subnet        `json:"subnets" bson:"subnets"`
	SecurityGroups []*SecurityGroup `json:"security_groups" bson:"security_groups"`
	ResourceGroup  string           `json:"resource_group" bson:"resource_group"`
}

type VNet struct {
	ID     bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	VnetId string        `json:"vpc_id" bson:"vpc_id"`
	Name   string        `json:"name" bson:"name"`
	CIDR   string        `json:"cidr" bson:"cidr"`
}
type DONetwork struct {
	ID               bson.ObjectId   `json:"-" bson:"_id,omitempty"`
	ProjectId        string          `json:"project_id" bson:"project_id" valid:"required"`
	Name             string          `json:"name" bson:"name" valid:"required"`
	Type             models.Type     `json:"type" bson:"type" valid:"required,in(New|Existing|new|existing)"`
	Cloud            models.Cloud    `json:"-" bson:"cloud" valid:"in(AWS|Azure|aws|azure)"`
	NetworkStatus    string          `json:"status" bson:"status"`
	CreationDate     time.Time       `json:"-" bson:"creation_date"`
	ModificationDate time.Time       `json:"-" bson:"modification_date"`
	Definition       []*DODefinition `json:"definition" bson:"definition" valid:"required"`
	CompanyId        string          `json:"company_id" bson:"company_id"`
}

type DODefinition struct {
	ID             bson.ObjectId    `json:"-" bson:"_id,omitempty"`
	SecurityGroups []*SecurityGroup `json:"security_groups" bson:"security_groups" valid:"optional"`
	VPCs           []*DOVPC         `json:"vpc" bson:"vpc" valid:"optional"`
}
type DOSecurityGroup struct {
	ID              bson.ObjectId `json:"-" bson:"_id,omitempty"`
	SecurityGroupId string        `json:"security_group_id" bson:"security_group_id"`
	Name            string        `json:"name" bson:"name" valid:"required"`
	Outbound        []DOBound     `json:"outbound" bson:"outbound" valid:"optional"`
	Inbound         []DOBound     `json:"inbound" bson:"inbound" valid:"optional"`
}
type DOVPC struct {
	ID    bson.ObjectId `json:"-" bson:"_id,omitempty"`
	VPCId string        `json:"vpc_id" bson:"vpc_id"`
	Name  string        `json:"name" bson:"name" valid:"required"`
}

type DOBound struct {
	addresses  []string `json:"addresses" bson:"addresses"`
	PortRange  string   `json:"port_range" bson:"port_range"`
	IpProtocol string   `json:"ip_protocol" bson:"ip_protocol" valid:"in(tcp|udp|icmp|all|UDP|TCP|ICMP|ALL|58|-1)"`
}
type IBMNetwork struct {
	ID               bson.ObjectId    `json:"-" bson:"_id,omitempty"`
	ProjectId        string           `json:"project_id" bson:"project_id" valid:"required"`
	Name             string           `json:"name" bson:"name" valid:"required"`
	Type             models.Type      `json:"type" bson:"type" valid:"required,in(New|Existing|new|existing)"`
	Cloud            models.Cloud     `json:"-" bson:"cloud" valid:"in(ibm|Azure|ibm|azure)"`
	CreationDate     time.Time        `json:"-" bson:"creation_date"`
	ModificationDate time.Time        `json:"-" bson:"modification_date"`
	Definition       []*IBMDefinition `json:"definition" bson:"definition" valid:"required"`
	CompanyId        string           `json:"company_id" bson:"company_id"`
}

type IBMDefinition struct {
	ID      bson.ObjectId `json:"-" bson:"_id,omitempty"`
	Vpc     IBMVpc        `json:"vpc" bson:"vpc" valid:"required"`
	Subnets []*IBMSubnet  `json:"subnets" bson:"subnets" valid:"optional"`
}
type IBMVpc struct {
	ID               bson.ObjectId `json:"-" bson:"_id,omitempty"`
	VpcId            string        `json:"vpc_id" bson:"vpc_id"`
	Name             string        `json:"name" bson:"name" valid:"required"`
	AvailabilityZone string        `json:"availability_zone" bson:"availability_zone"`
	CIDR             string        `json:"cidr" bson:"cidr" valid:"required,cidr"`
}
type IBMSubnet struct {
	ID                      bson.ObjectId `json:"-" bson:"_id,omitempty"`
	SubnetId                string        `json:"subnet_id" bson:"subnet_id"`
	Name                    string        `json:"name" bson:"name" valid:"required"`
	CIDR                    string        `json:"cidr" bson:"cidr" valid:"required,cidr"`
	RouteTableAssociationId string        `json:"route_table_association_id" bson:"route_table_association_id,omitempty"`
	AvailabilityZone        string        `json:"availability_zone" bson:"availability_zone"`
	KubeTag                 bool          `json:"tag_assigned" bson:"tag_assigned"`
}
