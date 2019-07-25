package utils

import (
	"bytes"
	"github.com/astaxie/beego"
	"net/http"
)

type HTTPClient struct {
	client *http.Client
}

func CreatePostRequest(request_data []byte, url string) (*http.Request, error) {

	//beego.Info("requesting ", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(request_data))

	if err != nil {
		beego.Error("%s", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func CreateGetRequest(url string) (*http.Request, error) {

	//beego.Info("requesting", url)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		beego.Error("%s", err)
		return nil, err
	}

	req.Proto = "HTTP/1.0"
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (httpReq *HTTPClient) SendRequest(req *http.Request) (*http.Response, error) {

	response, err := httpReq.client.Do(req)
	if err != nil {
		beego.Error("%s", err)
		return nil, err
	}
	return response, err
}

func InitReq() HTTPClient {

	var client HTTPClient
	client.client = &http.Client{}
	return client

}
