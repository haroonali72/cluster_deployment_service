package api_handler

import (
	"antelope/models/utils"
	"fmt"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"strings"
)


func GetAwsRegions() (map[string]string,error){
	region := make(map[string]string)
	client := utils.InitReq()
	host :="https://raw.githubusercontent.com/awsdocs/amazon-ec2-user-guide/master/doc_source/using-regions-availability-zones.md"
	req, err := utils.CreateGetRequest(host)
	if err != nil {
		return  region,err
	}
	response, err := client.SendRequest(req)
	if err != nil {
		return  region,err
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return  region,err
	}

	md :=blackfriday.MarkdownBasic(contents)

	s := string(md)
	first_index := strings.Index(s,"|  <code>")
	last_index := strings.LastIndex(s,"|  <code>")
	regionsInfo := s[first_index:last_index+1]
	regionsInfo = strings.TrimSpace(regionsInfo)
	regionsInfo=strings.ReplaceAll(regionsInfo,"<code>","")
	regionsInfo=strings.ReplaceAll(regionsInfo,"</code> ","")
	information := strings.Split(regionsInfo, "\n")

	for _,info := range information{
		if info == "|"{
			break
		}
		regionInfo:= strings.Split(info,"| ")
		loc :=strings.Split(regionInfo[2],"(")
		loca :=strings.Split(loc[1],")")
		region[loca[0]]=regionInfo[1]
	}
	return region,nil
}

func GetGcpRegion() (map[string]string,error){
	region := make(map[string]string)
	client := utils.InitReq()
	host :="https://cloud.google.com/compute/docs/regions-zones.md"
	req, err := utils.CreateGetRequest(host)
	if err != nil {
		return  region,err
	}
	response, err := client.SendRequest(req)
	if err != nil {
		return  region,err
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return  region,err
	}
	fmt.Println(contents)
	//d,err :=html.Parse(byte(contents))
	//md :=blackfriday.MarkdownBasic(contents)

	s := string(contents)
	first_index := strings.Index(s,"<table>")
	last_index := strings.Index(s,"</table>")
	regionsInfo := s[first_index:last_index+1]
	regionsInfo = strings.TrimSpace(regionsInfo)
	regionsInfo=strings.ReplaceAll(regionsInfo,"<code>","")
	regionsInfo=strings.ReplaceAll(regionsInfo,"</code> ","")
	information := strings.Split(regionsInfo, "\n")

	for _,info := range information{
		if info == "|"{
			break
		}
		regionInfo:= strings.Split(info,"| ")
		loc :=strings.Split(regionInfo[2],"(")
		loca :=strings.Split(loc[1],")")
		region[loca[0]]=regionInfo[1]
	}
	return region,nil
}
