package utils

import (
	"net/http"
	"github.com/astaxie/beego"
	"bytes"
)

type HTTPClient struct {
	client *http.Client
}

func CreatePostRequest (request_data []byte, url string) (*http.Request, error){

	beego.Info("requesting ", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(request_data))

	if err != nil {
		beego.Error("%s", err)
		return nil , err
	}

	req.Header.Set("Content-Type", "application/json")
	return req , nil
}

func CreateGetRequest ( url string, key string) (*http.Request, error){

	beego.Info("requesting ", url)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		beego.Error("%s", err)
		return nil , err
	}

	q := req.URL.Query()
	q.Add("name",key)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	return req , nil
}

func (httpReq *HTTPClient) SendRequest (req * http.Request)(*http.Response, error){

	response, err := httpReq.client.Do(req)
	if err != nil {
		beego.Error("%s", err)
		return nil, err
	}
	return response, err
}

func InitReq () HTTPClient{

	var logger HTTPClient
	logger.client = &http.Client{}
	return logger
}