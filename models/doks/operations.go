package doks

import (
	"antelope/models"
	"antelope/models/utils"
	"errors"
	"github.com/astaxie/beego"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"strings"
)

func getNetworkHost(cloudType, projectId string) string {

	host := beego.AppConfig.String("network_url") + models.WeaselGetEndpoint

	if strings.Contains(host, "{cloud}") {
		host = strings.Replace(host, "{cloud}", cloudType, -1)
	}

	if strings.Contains(host, "{projectId}") {
		host = strings.Replace(host, "{projectId}", projectId, -1)
	}

	return host
}

type DOKS struct {
	AccessKey string
	Region    string
	Client    *godo.Client
	Resources map[string][]string
}
type TokenSource struct {
	AccessToken string
}

func (cloud *DOKS) init(ctx utils.Context) error {
	if cloud.Client != nil {
		return nil
	}

	if cloud.AccessKey == "" {
		text := "invalid cloud credentials"
		ctx.SendLogs(text, models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(text)
		return errors.New(text)
	}

	tokenSource := &TokenSource{
		AccessToken: cloud.AccessKey,
	}

	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	cloud.Client = godo.NewClient(oauthClient)
	cloud.Resources = make(map[string][]string)
	return nil
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}
