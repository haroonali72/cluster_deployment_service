package d_duck

import (
	"errors"
	"io/ioutil"
	"net/http"
)

type Client struct {
	Host      string
	Port      string
	Username  string
	Password  string
	ApiKey    string
	ApiSecret string
}

const (
	TextXml         string = "text/xml"
	ApplicationJson string = "application/json"
)

func (c *Client) GetRawCatalog() ([]byte, error) {
	basePath := "http://" + c.Host + ":" + c.Port
	fullPath := basePath + "/1.0/kb/catalog/xml"

	return Get(c, fullPath, TextXml)
}

func (c *Client) GetRawProduct(subscriptionId string) ([]byte, error) {
	basePath := "http://" + c.Host + ":" + c.Port
	fullPath := basePath + "/1.0/kb/catalog/product?subscriptionId=" + subscriptionId

	return Get(c, fullPath, ApplicationJson)
}

func (c *Client) GetSubscriptionData(accountId string) ([]byte, error) {
	basePath := "http://" + c.Host + ":" + c.Port
	fullPath := basePath + "/1.0/kb/accounts/" + accountId + "/bundles"

	return Get(c, fullPath, ApplicationJson)
}

var Get = func(c *Client, url string, accept string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("accept", accept)
	req.Header.Set("X-Killbill-ApiKey", c.ApiKey)
	req.Header.Set("X-Killbill-ApiSecret", c.ApiSecret)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New(response.Status)
	}

	return ioutil.ReadAll(response.Body)
}
