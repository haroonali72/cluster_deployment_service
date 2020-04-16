package aks

import (
	"antelope/models/types"
	"strings"
)

func ApiError(err error, code int) (cError types.CustomCPError) {

//	cError.StatusCode = code
	cError.Description = err.Error()
	cError.Message = ValidationError(err.Error())

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
