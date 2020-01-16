package userData

import (
	"antelope/models/api_handler"
	"antelope/models/types"
	"antelope/models/utils"
	b64 "encoding/base64"
	"encoding/json"
	"github.com/astaxie/beego"
	"gopkg.in/yaml.v2"
)

func GetUserData(token, url string, ctx utils.Context) (string, error) {

	var data types.Data

	rawData, err := api_handler.GetAPIStatus(token, url, ctx)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(rawData.([]byte), &data)
	if err != nil {
		return "", err
	}

	var userData types.UserData

	config, err := json.Marshal(data.Config)
	if err != nil {
		return "", err
	}

	encodedData := b64.StdEncoding.EncodeToString(config)

	var writeFile types.WriteFile
	writeFile.Contents = encodedData
	writeFile.Encoding = "b64"
	writeFile.Path = "/usr/local/bin/userDataContents"
	writeFile.Owner = "root:root"
	writeFile.Permission = "0644"

	var arrayOfFiles []types.WriteFile
	arrayOfFiles = append(arrayOfFiles, writeFile)

	userData.WriteFile = arrayOfFiles

	var commands [][]string
	commands = append(commands, []string{"cd", "/usr/local/bin"})
	commands = append(commands, []string{"wget", data.Agent})
	commands = append(commands, []string{"chmod", "+x", "agent"})
	commands = append(commands, []string{"nohup", "./agent", "&"})

	userData.RunCmd = commands

	out, err := yaml.Marshal(userData)
	if err != nil {
		return "", err
	}
	master_data := "#cloud-config\n" + string(out)
	beego.Info(master_data)
	return master_data, nil

}
