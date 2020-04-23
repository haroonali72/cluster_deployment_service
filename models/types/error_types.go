package types

import "antelope/models"

type CustomCPError struct {
	StatusCode int `json:"-"  bson:"status_code"`
	Error     string `json:"error,omitempty"  bson:"error"`
	Description string `json:"description,omitempty"  bson:"description"`
}
type ClusterError struct {
	Cloud     models.Cloud  `json:"cloud"  bson:"cloud"`
	ProjectId string        `json:"project_id"  bson:"project_id"`
	CompanyId string        `json:"company_id" bson:"company_id"`
	Err       CustomCPError `json:"error" bson:"error"`
}
