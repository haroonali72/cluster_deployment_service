package doks

import (
	"antelope/models/utils"
	"antelope/models/vault"
	"github.com/digitalocean/godo"
	"strings"
)

type CustomError struct{
//	Status				string 			 `json:"status,omitempty"  bson:"status"`
	StatusCode			string			 `json:"code,omitempty"  bson:"code"`
	Type				string 			 `json:"type,omitempty"  bson:"type"`
	Message				string 			 `json:"message,omitempty"  bson:"message"`
	Description			string  		 `json:"description,omitempty"  bson:"description"`
}

func ApiError (err error, credentials vault.DOCredentials,ctx utils.Context,companyId string) (cError CustomError){

	errr :=strings.Fields(err.Error())
	cError.StatusCode = errr[2]
	cError.Type=errr[3]
	cError.Description = err.Error()
	if (errr[2]=="422"){
		cError.Message =ValidationError(err.Error(),credentials,ctx ,companyId )
	}

	return cError

}

func getKubernetesVersion(credentials vault.DOCredentials,ctx utils.Context,companyId string) string{
	config,_:=GetServerConfig(credentials ,ctx , companyId )
	var versions string
	for _,version:= range config.Versions {
		versions = versions + *godo.String(version.KubernetesVersion)+ ": " +*godo.String(version.Slug) + " , "
	}

	return versions
}

func getMachineSizes(credentials vault.DOCredentials,ctx utils.Context,companyId string) string{
	config,_:=GetServerConfig(credentials ,ctx , companyId )
	var sizes string
	for _,size:= range config.Sizes {
		sizes = sizes + *godo.String(size.Name)+ " : " +*godo.String(size.Slug) + " , "
	}
	return sizes
}

func getRegions(credentials vault.DOCredentials,ctx utils.Context,companyId string )string{

	config,_:=GetServerConfig(credentials ,ctx , companyId )
	var regions string
	for _,re:= range config.Regions {
		regions = regions + *godo.String(re.Name)+ "(" +*godo.String(re.Slug) + ") , "
	}
	return regions
}


func ValidationError(description string ,credentials vault.DOCredentials,ctx utils.Context,companyId string) string{

 	if strings.Contains(description,"cluster_spec.missing"){
		if strings.Contains(description,"region"){
			regions := getRegions(credentials ,ctx ,companyId )
			return "Request have some missing value : Region . Select region from : "+regions

		} else if strings.Contains(description,"name"){
			return "Missing Value : Name .Give a valid name"
		} else if strings.Contains(description,"size"){
			sizes := getMachineSizes(credentials ,ctx ,companyId )
			return "Missing Value : Machine size of node pool.  Select machine size from : " +sizes
		} else if strings.Contains(description,"min_count"){
			return "Missing Value : Minimum Node Count for auto scaling.Select a numerical value of minimun number of nodes for autoscaling."
		}else if strings.Contains(description,"max_count"){
			return "Missing Value : Max Count for auto scaling.Select a minimun numerical value of maximun number of nodes for autoscaling.The value should be more than minimun count and less than 25"
		}  else if strings.Contains(description,"version"){
			versions := getKubernetesVersion(credentials ,ctx ,companyId )
			return "Missing Value : Kubernetes Version.Select a Kubernetes Version from : "+versions
		}else {
			return description
		}

	}else if  strings.Contains(description,"cluster_spec.invalid"){
		if strings.Contains(description,"region"){
			regions := getRegions(credentials ,ctx ,companyId )
			return "Invalid Value : Region. Select region from : " +regions
		} else if strings.Contains(description,"name"){
			return "Invalid Value : Name .Give a valid name"
		} else if strings.Contains(description,"size"){
			sizes := getMachineSizes(credentials ,ctx ,companyId )
			return "Invalid Value : Machine size of node pool.  Select machine size from : " +sizes
		} else if strings.Contains(description,"min_count"){
			return "Invalid Value : Minimum Node Count for auto scaling.Select a numerical value of minimun number of nodes for autoscaling."
		}else if strings.Contains(description,"max_count"){
			return "Invalid Value : Max Count for auto scaling.Select a minimun numerical value of maximun number of nodes for autoscaling.The value should be more than minimun count and less than 25"
		}  else if strings.Contains(description,"version"){
			versions := getKubernetesVersion(credentials ,ctx ,companyId )
			return "Invalid Value : Kubernetes Version.Select a Kubernetes Version from : " +versions
		}else {
			return description
		}

	}
	return description
 }