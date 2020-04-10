package aks

import (
	"strings"
)

type CustomError struct {
	//	Status				string 			 `json:"status,omitempty"  bson:"status"`
	StatusCode  string `json:"code,omitempty"  bson:"code"`
	Type        string `json:"type,omitempty"  bson:"type"`
	Message     string `json:"message,omitempty"  bson:"message"`
	Description string `json:"description,omitempty"  bson:"description"`
}

func ApiError(err error) (cError CustomError) {

	errr := strings.Fields(err.Error())
	cError.StatusCode = errr[2]
	cError.Type = errr[3]
	cError.Description = err.Error()
	if errr[2] == "422" {
		cError.Message = ValidationError(err.Error())
	}

	return cError

}

func ValidationError(description string) string {
	if strings.Contains(description, "APIServerAuthorizedIPRanges") {
		return "API server authorized IP ranges must be valid IP address or CIDR"
	} else if strings.Contains(description, "ServicePrincipalProfile were invalid") {
		return "Service Principal Profile credentials are not valid"
	} else if strings.Contains(description, "agentPoolProfile.name is invalid") {
		return "Agent pool name is invalid. Please see https://aka.ms/aks-naming-rules for more details."
	} else if strings.Contains(description, "ServicePrincipalNotFound") {
		return "Service Principal not found"
	} else if strings.Contains(description, "exceeding approved Total Regional Cores quota") {
		return "Regional Cores Quota is getting exceeded"
	} else {
		return description
	}

}
