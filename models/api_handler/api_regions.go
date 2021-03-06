package api_handler

import (
	"antelope/models"
	"antelope/models/utils"
	"github.com/russross/blackfriday"
	"golang.org/x/net/html"

	"io/ioutil"
	"strings"
)

var AzureZoneNotSupportedRegions = []byte(
	`[
			{"region":"East Asia","location":"eastasia"},
			{"region":"Central US","location": "centralus"},
			{"region":"West US","location": "westus"},
			{"region":"North Central US","location":"northcentralus"},
			{"region":"Japan West","location":"japanwest"},
			{"region":"Brazil South","location":"brazilsouth"},
			{"region":"Australia Southeast","location":"austrlocationoutheast"},
			{"region":"South India","location": "southindia"},
			{"region":"Central India","location":"centralindia"},
			{"region":"West India","location": "westindia"},
			{"region":"West India","location": "westindia"},
			{"region":"Canada Central","location":"canadacentral"},
			{"region":"Canada East","location": "canadaeast"},
			{"region":"UK West","location":"ukwest"},
			{"region":"Korea Central","location":"koreacentral"},
			{"region":"Korea South","location": "koreasouth"},
			{"region":"Australia Central","location":"australiacentral"},
			{"region":"UAE North","location":"uaenorth"},
			{"region":"South Africa North","location":"southafricanorth"},
			{"region":"Switzerland North","location":"switzerlandnorth"},
			{"region":"Germany West Central","location":"germanywestcentral"},
			{"region":"Norway East","location": "norwayeast"}

]`)

var AzureZone = []byte(
	`[
	 "1",
	 "2",
	"3"
]`)

func GetAwsRegions() (reg []models.Region, err error) {
	region := new(models.Region)
	//region := make(map[string]string)
	client := utils.InitReq()
	host := "https://raw.githubusercontent.com/awsdocs/amazon-ec2-user-guide/master/doc_source/using-regions-availability-zones.md"
	req, err := utils.CreateGetRequest(host)
	if err != nil {
		return []models.Region{}, err
	}
	response, err := client.SendRequest(req)
	if err != nil {
		return []models.Region{}, err
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []models.Region{}, err
	}

	md := blackfriday.MarkdownBasic(contents)

	s := string(md)
	first_index := strings.Index(s, "<p>| Code")
	last_index := strings.LastIndex(s, "|</p>")
	regionsInfo := s[first_index : last_index+1]
	regionsInfo = strings.TrimSpace(regionsInfo)
	regionsInfo = strings.ReplaceAll(regionsInfo, "<code>", "")
	regionsInfo = strings.ReplaceAll(regionsInfo, "</code> ", "")
	information := strings.Split(regionsInfo, "\n")

	for _, info := range information {
		if info == "|" {
			break
		}
		regionInfo := strings.Split(info, "| ")
		if strings.Contains(regionInfo[2], "(") {
			loc := strings.Split(regionInfo[2], "(")
			loca := strings.Split(loc[1], ")")
			//region[loca[0]]=regionInfo[1]

			region.Name = loca[0]
			region.Location = strings.TrimSpace(regionInfo[1])
			reg = append(reg, *region)
		}
	}
	return reg, nil
}

func GetGcpRegion() (reg []models.Region, err error) {

	//	var region models.GcpRegion
	var regions []string
	var region models.Region
	client := utils.InitReq()
	host := "https://cloud.google.com/compute/docs/regions-zones.md"
	req, err := utils.CreateGetRequest(host)
	if err != nil {
		return reg, err
	}
	response, err := client.SendRequest(req)
	if err != nil {
		return reg, err
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return reg, err
	}

	s := string(contents)
	first_index := strings.Index(s, "<table>")
	last_index := strings.Index(s, "</table>")
	regionsInfo := s[first_index : last_index+1]
	regionsInfo = strings.TrimSpace(regionsInfo)
	domDocTest := html.NewTokenizer(strings.NewReader(regionsInfo))
	previousStartTokenTest := domDocTest.Token()
loopDomTest:
	for {
		tt := domDocTest.Next()
		switch {
		case tt == html.ErrorToken:
			break loopDomTest // End of the document,  done
		case tt == html.StartTagToken:
			previousStartTokenTest = domDocTest.Token()
		case tt == html.TextToken:
			if previousStartTokenTest.Data == "script" {
				continue
			}
			TxtContent := strings.TrimSpace(html.UnescapeString(string(domDocTest.Text())))
			if TxtContent == "<" {
				break
			}
			if len(TxtContent) > 0 && TxtContent != "Region" && TxtContent != "Zones" && TxtContent != "Location" && TxtContent != "Machine types"  && TxtContent != "CPUs" && TxtContent != "Resources"  && TxtContent != "GPUs" && !strings.Contains(TxtContent,"E2") && !strings.Contains(TxtContent,"Broadwell") && !strings.Contains(TxtContent,"Skylake"){
				regions = append(regions, TxtContent)
			}
		}
	}
	exist :=false
	for i := 0; i <= len(regions); i = i +2 {

		if i+1 >= len(regions)  {
			break
		}
		temp := regions[i+1]

		for _,re := range reg {
			if  temp== re.Name  || (strings.Contains(temp, "-a") || strings.Contains(temp, "-b") || strings.Contains(temp, "-c") ){
				exist =true
			}
		}
		if exist ==false {
			check := regions[i]
			region.Location = check[:len(check)-2]
			region.Name = regions[i+1]
			reg = append(reg, region)

		}else{
			exist =false
		}
	}
	return reg, nil
}
