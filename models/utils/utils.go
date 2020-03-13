package utils

import (
	"antelope/models"
	"bytes"
	"github.com/astaxie/beego"
	"net/http"
	"time"
)

type HTTPClient struct {
	client *http.Client
}

type Key struct {
	CredentialType models.CredentialsType `json:"credential_type"  bson:"credential_type"`
	NewKey         models.KeyType         `json:"key_type"  bson:"key_type"`
	KeyName        string                 `json:"key_name" bson:"key_name"`
	Username       string                 `json:"username" bson:"username,omitempty"`
	AdminPassword  string                 `json:"admin_password" bson:"admin_password,omitempty"`
	PrivateKey     string                 `json:"private_key" bson:"private_key,omitempty"`
	PublicKey      string                 `json:"public_key" bson:"public_key,omitempty"`
	Cloud          models.Cloud           `json:"cloud" bson:"cloud"`
}

type KeyPairResponse struct {
	KeyName    string `json:"key_name"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

func CreatePutRequest(request_data []byte, url string) (*http.Request, error) {

	//beego.Info("requesting ", url)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(request_data))

	if err != nil {
		beego.Error("%s", err)
		return nil, err
	}

	//req.Header.Set("Content-Type", "application/json")
	return req, nil
}
func CreatePostRequest(request_data []byte, url string) (*http.Request, error) {

	//beego.Info("requesting ", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(request_data))

	if err != nil {
		beego.Error("%s", err)
		return nil, err
	}

	//req.Header.Set("Content-Type", "application/json")
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

func CreateDeleteRequest(url string) (*http.Request, error) {

	//beego.Info("requesting ", url)

	req, err := http.NewRequest("DELETE", url, nil)

	if err != nil {
		beego.Error("%s", err)
		return nil, err
	}

	//	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
func SetHeaders(req *http.Request, headers map[string]string) {
	for key, value := range headers {
		req.Header.Set(key, value)
	}
}
func (httpReq *HTTPClient) SendRequest(req *http.Request) (*http.Response, error) {
	start := time.Now()

	response, err := httpReq.client.Do(req)

	if err != nil {
		beego.Error("%s", err)

		elapsed := time.Since(start)
		beego.Warn("SEGMENT: "+req.Host+" took:", elapsed)

		return nil, err
	}

	elapsed := time.Since(start)
	beego.Warn("SEGMENT: "+req.Host+" took:", elapsed)

	return response, err
}

func InitReq() HTTPClient {

	var client HTTPClient
	client.client = &http.Client{}
	return client

}
