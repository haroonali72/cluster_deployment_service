package woodpecker

import (
	"antelope/models"
	"antelope/models/utils"
	"errors"
	"github.com/astaxie/beego"
	"io/ioutil"
	"regexp"
	"strings"
)

func GetCertificate(projectID, token string, ctx utils.Context) (string, error) {

	client := utils.InitReq()
	basePath := ""
	if basePath = getWoodPeckerHost(); basePath == "" {
		beego.Error("can not get woodpecker_url in env")
		return "", errors.New("can not get woodpecker_url in env")
	}
	basePath = basePath + models.WoodPeckerCertificate
	basePath = strings.Replace(basePath, "{profileId}", projectID, 1)
	req, err := utils.CreateGetRequest(basePath)
	if err != nil {
		beego.Info(err.Error())
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}
	req.Header.Set("token", token)
	response, err := client.SendRequest(req)
	if err != nil {
		beego.Info(err.Error())
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return "", err
	}
	contents, err := ioutil.ReadAll(response.Body)
	re := regexp.MustCompile("(?m)[\r\n]+^.*keel.sh/match-ta.*$")
	res := re.ReplaceAllString(string(contents), "")
	re = regexp.MustCompile("(?m)[\r\n]+^.*keel.sh/policy.*$")
	res = re.ReplaceAllString(res, "")
	return res, err
}

func getWoodPeckerHost() string {
	return beego.AppConfig.String("woodpecker_url")
}
