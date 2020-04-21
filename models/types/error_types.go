package types

import "antelope/models"

type CustomCPError struct {
	StatusCode int `json:"code,omitempty"  bson:"code"`
	//Type        string `json:"type,omitempty"  bson:"type"`
	Message     string `json:"message,omitempty"  bson:"message"`
	Description string `json:"description,omitempty"  bson:"description"`
}
type ClusterError struct {
	Cloud     models.Cloud  `json:"cloud"  bson:"cloud"`
	ProjectId string        `json:"project_id"  bson:"project_id"`
	CompanyId string        `json:"company_id" bson:"company_id"`
	Err       CustomCPError `json:"error" bson:"error"`
}
