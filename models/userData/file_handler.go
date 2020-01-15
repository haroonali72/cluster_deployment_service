package userData

import (
	"antelope/models/api_handler"
	"antelope/models/types"
	"antelope/models/utils"
	b64 "encoding/base64"
	"encoding/json"
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
	encodedData := b64.StdEncoding.EncodeToString([]byte(config))

	var writeFile types.WriteFile
	writeFile.Contents = encodedData
	writeFile.Encoding = "b64"
	writeFile.Path = "/etc/"
	writeFile.Owner = "root:root"
	writeFile.Permission = "0644"

	var arrayOfFiles []types.WriteFile
	arrayOfFiles = append(arrayOfFiles, writeFile)

	userData.WriteFile = arrayOfFiles

	var cmd1 []string
	cmd1 = append(cmd1, "wget")
	cmd1 = append(cmd1, data.Agent)
	cmd1 = append(cmd1, "-O")
	cmd1 = append(cmd1, "/etc/")

	var commands [][]string
	commands = append(commands, cmd1)

	userData.RunCmd = commands

	out, err := yaml.Marshal(userData)
	if err != nil {
		return "", err
	}
	master_data := "#cloud-config\n" + string(out)

	return master_data, nil

}
