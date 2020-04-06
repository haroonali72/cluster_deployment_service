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
	Message				string 			 `json:"message,omitempty"  bson:"message"`
	Description			string  		 `json:"description,omitempty"  bson:"description"`
}

func ApiError (err error, credentials vault.DOCredentials,ctx utils.Context,companyId string) (cError CustomError){

	errr :=strings.Fields(err.Error())
	cError.StatusCode = errr[0]
	cError.Message = errr[1]
	if (errr[0]=="422"){
		cError.Description=ValidationError(errr[3],credentials,ctx ,companyId )
	}

	return cError

}
func getKubernetesVersion(credentials vault.DOCredentials,ctx utils.Context,companyId string) []*godo.KubernetesVersion{
	kubeversion,_:=GetServerConfig(credentials ,ctx , companyId )
	return kubeversion.Versions
}
func getMachineSizes(credentials vault.DOCredentials,ctx utils.Context,companyId string) []*godo.KubernetesNodeSize{
	machines,_:=GetServerConfig(credentials ,ctx , companyId )
	return machines.Sizes
}
func getRegions(credentials vault.DOCredentials,ctx utils.Context,companyId string )[]*godo.KubernetesRegion{
	regions,_:=GetServerConfig(credentials ,ctx , companyId )
	return regions.Regions
}


func ValidationError(description string ,credentials vault.DOCredentials,ctx utils.Context,companyId string) string{

 	if strings.Contains(description,"cluster_spec.missing"){
		if strings.Contains(description,"region"){
			regions := getRegions(credentials ,ctx ,companyId )
			return "Missing Value : Region. Select region from : "+regions

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