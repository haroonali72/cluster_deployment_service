package d_duck

import (
	"errors"
	"io/ioutil"
	"net/http"
)

type Client struct {
	Host string
	Port string
}

const (
	TextXml         string = "text/xml"
	ApplicationJson string = "application/json"
)

func (c *Client) GetRawCatalog() ([]byte, error) {
	basePath := "http://" + c.Host + ":" + c.Port
	fullPath := basePath + "/1.0/kb/catalog/xml"

	return Get(fullPath, TextXml)
}

func (c *Client) GetRawProduct(subscriptionId string) ([]byte, error) {
	basePath := "http://" + c.Host + ":" + c.Port
	fullPath := basePath + "/1.0/kb/catalog/product?subscriptionId=" + subscriptionId

	return Get(fullPath, ApplicationJson)
}

var Get = func(url string, accept string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth("admin", "password")
	req.Header.Set("accept", accept)
	req.Header.Set("X-Killbill-ApiKey", "cloudplex")
	req.Header.Set("X-Killbill-ApiSecret", "cloudplex")

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
