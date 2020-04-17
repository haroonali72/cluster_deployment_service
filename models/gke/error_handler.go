package gke

import (
	"antelope/models/types"
	"strings"
)
func ApiErrors (err error,message string) (cError types.CustomCPError){
	errr :=strings.Fields(err.Error())
	cError.StatusCode = 502
	cError.Description = err.Error()
	if (errr[2]=="304"){
		cError.Message =NotModified(err.Error())
	}else if errr[2]=="400"{
		cError.Message =BadRequest(err.Error())
	} else if errr[2]=="401" {
		cError.Message = Unauthorized(err.Error())
	}else if errr[2]=="402"{
		cError.Message =QuotaReached(err.Error())
	}else if errr[2]=="403"{
		cError.Message =Forbidden(err.Error())
	}else if errr[2]=="404" {
		cError.Message = NotFound(err.Error())
	}	else if errr[2]=="409"{
		cError.Message =Conflict(err.Error())
	}else if errr[2]=="410" {
		cError.Message = Gone(err.Error())
	}else if errr[2]=="429" {
		cError.Message = ResourceExhausted(err.Error())
	}else if errr[2]=="500" {
		cError.Message = InternalServerError(err.Error())
	}else if errr[2]=="503" {
		cError.Message = ServiceUnavailable(err.Error())
	}else {
		return cError
	}
	if cError.Message==""{
		cError.Message=message
	}
	return cError
}
func NotModified(err string)string{
	return err
}
func BadRequest(err string)string{
	if strings.Contains(err,"node_count") && strings.Contains(err,"greater than zero") {
		return "The cluster cannot be created without nodepools.Add a nodepool in the cluster"
	} else if strings.Contains(err,"out of range")  {

	}else if strings.Contains(err,"Location") && strings.Contains(err,"does not exist")  {
		if strings.Contains(err,"a") || strings.Contains(err,"b") || strings.Contains(err,"c") ||strings.Contains(err,"d") || strings.Contains(err,"e"){
			return 	 "The zone is invalid.Select another zone from"
		}
		return "The region is invalid.Select a valid region from"
	}
	return ""
}
func Unauthorized(err string) string{
	return "Invalid Profile.Use valid google cretentials profile"
}
func QuotaReached(err string)string{
	return ""
}
func Forbidden(err string)string{
	if strings.Contains(err,"Permission denied") && strings.Contains(err,"locations"){
		return "You do not have permission to create resource in this region.Select another region."
	}
	return "You does not have sufficient permission For this resource. The OAuth token does not have the right scopes. Enable the premission on console or use another profile."

}
func NotFound(err string)string{
	if strings.Contains(err,"/cluster/"){
		return "The cluster is not in running state on console."
	}
	return ""
}
func Conflict(err string)string{
	return ""
}
func Gone(err string)string{
	return""
}
func InternalServerError(err string)string{
	return ""
}
func ServiceUnavailable(err string)string{
	return "The server is down.Retry after some time."

}

func  ResourceExhausted(err string)string {
	return "The resource quota  limit is reached.Either delete some resources or increase the resource limit for your account."
}

func ValidationError(err string) string{
	return ""
}

