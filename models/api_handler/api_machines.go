package api_handler

import (
	"antelope/models/utils"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"strings"
)

func GetAwsMachines() ([]string, error) {

	client := utils.InitReq()
	host := "https://raw.githubusercontent.com/awsdocs/amazon-ec2-user-guide/master/doc_source/instance-types.md"
	req, err := utils.CreateGetRequest(host)
	if err != nil {
		return []string{}, err
	}
	response, err := client.SendRequest(req)
	if err != nil {
		return []string{}, err
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []string{}, err
	}

	md := blackfriday.MarkdownBasic(contents)

	s := string(md)
	first_index := strings.Index(s, "| General purpose |")
	last_index := strings.LastIndex(s, "| General purpose |")
	regionsInfo := s[first_index : last_index+1]
	regionsInfo = strings.TrimSpace(regionsInfo)
	regionsInfo = strings.ReplaceAll(regionsInfo, "<code>", "")
	regionsInfo = strings.ReplaceAll(regionsInfo, "</code> ", "")
	information := strings.Split(regionsInfo, "\n")
	var mach []string
	for _, info := range information {

		machineInfo := strings.Split(info, "| ")

		for _, machine := range machineInfo {
			machine := strings.TrimSpace(machine)
			if machine == "" || machine == "General purpose" || machine == "Compute optimized" || machine == "Memory optimized" || machine == "Storage optimized" || machine == "Accelerated computing" {
				continue
			}
			if machine[len(machine)-1] == '|' || machine[len(machine)-1] == '>' || machine[len(machine)-1] == '-' {
				break
			}
			mach = append(mach, machine)
		}

	}
	return mach, nil
}
