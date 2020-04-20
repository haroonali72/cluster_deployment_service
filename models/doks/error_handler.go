package doks

import (
	"antelope/models/types"
	"antelope/models/utils"
	"antelope/models/vault"
	"github.com/digitalocean/godo"
	"strings"
)

func ApiError (err error,message string, credentials vault.DOCredentials,ctx utils.Context) (cError types.CustomCPError){

	errr :=strings.Fields(err.Error())
	cError.StatusCode = 502
	cError.Description = err.Error()
	if (errr[2]=="422"){
		cError.Message =ValidationError(err.Error(),credentials,ctx )
	}else if errr[2]=="404"{
		cError.Message =NotFoundError(err.Error(),errr[0],ctx)
	} else if strings.Contains(err.Error(),"Invalid cloud credentials"){
		cError.StatusCode=402
		cError.Message="Invalid cloud credentials"
		cError.Description="The Access Token is not valid."
	}
	if cError.Message==""{
		cError.Message= message
	}
	return cError

}

func getKubernetesVersion(credentials vault.DOCredentials,ctx utils.Context) string{
	config,_:=GetServerConfig(credentials ,ctx )
	var versions string
	for _,version:= range config.Versions {
		versions = versions + *godo.String(version.KubernetesVersion)+ ": " +*godo.String(version.Slug) + " , "
	}
	return versions
}

func getMachineSizes(credentials vault.DOCredentials,ctx utils.Context) string{
	config,_:=GetServerConfig(credentials ,ctx )
	var sizes string
	for _,size:= range config.Sizes {
		sizes = sizes + *godo.String(size.Name)+ " : " +*godo.String(size.Slug) + " , "
	}
	return sizes
}

func getRegions(credentials vault.DOCredentials,ctx utils.Context )string{

	config,_:=GetServerConfig(credentials ,ctx )
	var regions string
	for _,re:= range config.Regions {
		regions = regions + *godo.String(re.Name)+ "(" +*godo.String(re.Slug) + ") , "
	}
	return regions
}


func ValidationError(description string ,credentials vault.DOCredentials,ctx utils.Context) string{

 	if strings.Contains(description,"cluster_spec.missing"){
		if strings.Contains(description,"region"){
			regions := getRegions(credentials ,ctx  )
			return "Request have some missing value : Region . Select region from : "+regions

		} else if strings.Contains(description,"name"){
			return "Missing Value : Name .Give a valid name"
		} else if strings.Contains(description,"size"){
			sizes := getMachineSizes(credentials ,ctx  )
			return "Missing Value : Machine size of node pool.  Select machine size from : " +sizes
		} else if strings.Contains(description,"min_count"){
			return "Missing Value : Minimum Node Count for auto scaling.Select a numerical value of minimun number of nodes for autoscaling."
		}else if strings.Contains(description,"max_count"){
			return "Missing Value : Max Count for auto scaling.Select a minimun numerical value of maximun number of nodes for autoscaling.The value should be more than minimun count and less than 25"
		}  else if strings.Contains(description,"version"){
			versions := getKubernetesVersion(credentials ,ctx )
			return "Missing Value : Kubernetes Version.Select a Kubernetes Version from : "+versions
		}else {
			return description
		}

	}else if  strings.Contains(description,"cluster_spec.invalid") {
		if strings.Contains(description, "region") {
			regions := getRegions(credentials, ctx)
			return "Invalid Value : Region. Select region from : " + regions
		} else if strings.Contains(description, "name") {
			return "Invalid Value : Name .Give a valid name"
		} else if strings.Contains(description, "size") {
			sizes := getMachineSizes(credentials, ctx)
			return "Invalid Value : Machine size of node pool.  Select machine size from : " + sizes
		} else if strings.Contains(description, "min_count") {
			return "Invalid Value : Minimum Node Count for auto scaling.Select a numerical value of minimun number of nodes for autoscaling."
		} else if strings.Contains(description, "max_count") {
			return "Invalid Value : Max Count for auto scaling.Select a minimun numerical value of maximun number of nodes for autoscaling.The value should be more than minimun count and less than 25"
		} else if strings.Contains(description, "version") {
			versions := getKubernetesVersion(credentials, ctx)
			return "Invalid Value : Kubernetes Version.Select a Kubernetes Version from : " + versions
		}else {
			return description
		}
	}else if  strings.Contains(description,"no nodes"){
		return "Node Count is empty.Give a numeric value for node_count "
	}else if  strings.Contains(description,"autoscale") && strings.Contains(description,"exceed limits") {
		return "The max_count value is more than the droplet limit of your account. Either decrease the max_node count value or increase droplet limit of your account."
	}else if  strings.Contains(description,"additional nodes") {
		return "The node count  is more than the droplet limit of your account. Either decrease the node count or increase droplet limit of your account."
	}else if  strings.Contains(description,"This size is currently restricted") {
		return "The selected machine size is restricted for your account.Contact Digital Ocean to unlock this machine size."
	}

	return description
 }

 func NotFoundError(description ,etype string , ctx utils.Context) string{
	 if strings.Contains(description,"kubeconfig"){
		 return "KubeConfig File cannot be fetched.Give a valid cluster name to fetch the kubeconfig file. "
	 }else if strings.Contains(etype,"GET"){
		 return "The status of the cluster cannot be fetched. Cluster should be in running state for it's status to be fetched"
	 }else if strings.Contains(etype,"DELETE"){
		 return "The Cluster cannot be deleted. Only cluster that are in running state can be deleted"
	 }else{
		 return description
	 }
 }