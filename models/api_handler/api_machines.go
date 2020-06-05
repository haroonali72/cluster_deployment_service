package api_handler

import (
	"antelope/models/utils"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"strings"
)

func GetAwsMachines() ([]string, error) {

	client := utils.InitReq()
	host := "https://raw.githubusercontent.com/awsdocs/amazon-ec2-user-guide/master/doc_source/general-purpose-instances.md"
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
	first_index := strings.Index(s, "| Instance type |")
	last_index := strings.LastIndex(s, "For more information about the hardware specifications")
	regionsInfo := s[first_index : last_index+1]
	regionsInfo = strings.TrimSpace(regionsInfo)
	regionsInfo = strings.ReplaceAll(regionsInfo, "<code>", "")
	regionsInfo = strings.ReplaceAll(regionsInfo, "</code> ", "")
	information := strings.Split(regionsInfo, "\n")
	var mach []string
	for _, info := range information {
		if info == "" {
			break
		}
		machineInfo := strings.Split(info, "| ")

		if machineInfo[1] == "" || machineInfo[1] == "--- " || machineInfo[1] == "|" || machineInfo[1] == "Instance type " || machineInfo[1] == "Default vCPUs" || machineInfo[1] == "Memory(GiB)|" {
			continue
		}
		mach = append(mach, machineInfo[1])

		/*for _, machine := range machineInfo {
			machine := strings.TrimSpace(machine)
			if machine == "" || machine == "Instance type" || machine == "Default vCPUs" || machine == "Memory(GiB)|" {
				continue
			}
			if machine[len(machine)-1] == '|' || machine[len(machine)-1] == '>' || machine[len(machine)-1] == '-' {
				break
			}
			mach = append(mach, machine)
		} */

	}
	return mach, nil
}
