package userData

import (
	"antelope/models"
	"antelope/models/api_handler"
	"antelope/models/types"
	"antelope/models/utils"
	b64 "encoding/base64"
	"encoding/json"
	"github.com/astaxie/beego"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func GetUserData(token, url string, scriptNames []string, poolRole models.PoolRole, ctx utils.Context) (string, error) {
	var enableUserData bool
	var userData types.UserData
	var arrayOfFiles []types.WriteFile
	var data types.Data
	if poolRole == models.Master {

		rawData, err := api_handler.GetAPIStatus(token, url, ctx)
		if err != nil {
			return "", err
		}
		err = json.Unmarshal(rawData.([]byte), &data)
		if err != nil {
			return "", err
		}

		config, err := json.Marshal(data.Config)
		if err != nil {
			return "", err
		}

		encodedData := b64.StdEncoding.EncodeToString(config)

		var writeFile types.WriteFile
		writeFile.Contents = encodedData
		writeFile.Encoding = "b64"
		writeFile.Path = "/usr/local/bin/client-conf.json"
		writeFile.Owner = "root:root"
		writeFile.Permission = "0644"
		arrayOfFiles = append(arrayOfFiles, writeFile)
		enableUserData = true
	}
	for _, name := range scriptNames {

		if name != "" {

			fileContents, err := ioutil.ReadFile("/app/scripts/" + name)
			if err != nil {
				return "", err
			}

			encodedScript := b64.StdEncoding.EncodeToString([]byte(fileContents))

			var writeScript types.WriteFile
			writeScript.Contents = encodedScript
			writeScript.Encoding = "b64"
			writeScript.Path = "/usr/local/bin/" + name
			writeScript.Owner = "root:root"
			writeScript.Permission = "700"
			arrayOfFiles = append(arrayOfFiles, writeScript)
			enableUserData = true
		}
	}
	if !enableUserData {
		return "no user data found", nil
	}

	userData.WriteFile = arrayOfFiles

	var commands [][]string
	if poolRole == "master" {
		commands = append(commands, []string{"cd", "/usr/local/bin"})
		commands = append(commands, []string{"wget", data.Agent})
		commands = append(commands, []string{"chmod", "+x", "agent"})
		commands = append(commands, []string{"nohup", "./agent", "&>", "agent.out", "&"})
	}
	for _, names := range scriptNames {
		commands = append(commands, []string{"nohup", "./" + names, "&>", "volume.out", "&"})
	}
	userData.RunCmd = commands

	out, err := yaml.Marshal(userData)
	if err != nil {
		return "", err
	}
	master_data := "#cloud-config\n" + string(out)
	beego.Info(master_data)
	return master_data, nil

}
