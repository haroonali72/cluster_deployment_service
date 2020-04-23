package types

import "antelope/models"

type CustomCPError struct {
	StatusCode int `json:"status_code,omitempty"  bson:"code"`
	Error     string `json:"error,omitempty"  bson:"message"`
	Description string `json:"description,omitempty"  bson:"description"`
}
type ClusterError struct {
	Cloud     models.Cloud  `json:"cloud"  bson:"cloud"`
	ProjectId string        `json:"project_id"  bson:"project_id"`
	CompanyId string        `json:"company_id" bson:"company_id"`
	Err       CustomCPError `json:"error" bson:"error"`
}
