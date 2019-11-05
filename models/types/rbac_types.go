package types

import "antelope/models"

type Response struct {
	CompanyId string `json:"companyId"`
	UserId    string `json:"username"`
}
type UserRole struct {
	Roles []Role `json:"roles"`
}

type Role struct {
	Name models.Role `json:"name"`
}
