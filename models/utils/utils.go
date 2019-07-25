package utils

import (
	"antelope/models"
	"bytes"
	"encoding/json"
	"github.com/astaxie/beego"
	"io/ioutil"
	"net/http"
	"os/exec"
	"time"
)

type HTTPClient struct {
	client *http.Client
}

type Key struct {
	CredentialType models.CredentialsType `json:"credential_type"  bson:"credential_type"`
	NewKey         models.KeyType         `json:"key_type"  bson:"key_type"`
	KeyName        string                 `json:"key_name" bson:"key_name"`
	AdminPassword  string                 `json:"admin_password" bson:"admin_password,omitempty"`
	PrivateKey     string                 `json:"private_key" bson:"private_key,omitempty"`
	PublicKey      string                 `json:"public_key" bson:"public_key,omitempty"`
	Cloud          models.Cloud           `json:"cloud" bson:"cloud"`
}

type KeyPairResponse struct {
	KeyName   string `json:"key_name"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

func CreatePostRequest(request_data []byte, url string) (*http.Request, error) {

	beego.Info("requesting ", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(request_data))

	if err != nil {
		beego.Error("%s", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func CreateGetRequest(url string) (*http.Request, error) {

	beego.Info("requesting", url)

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

func KeyConversion(keyInfo interface{}) (Key, error) {
	b, e := json.Marshal(keyInfo)
	var k Key
	if e != nil {
		beego.Error(e)
		return Key{}, e
	}
	e = json.Unmarshal(b, &k)
	if e != nil {
		beego.Error(e)
		return Key{}, e
	}
	return k, nil
}

func GenerateKeyPair(keyName string) (KeyPairResponse, error) {
	res := KeyPairResponse{}

	t := time.Now().Local()
	tstamp := t.Format("20060102150405")
	keyName = keyName + "_" + tstamp

	cmd := "ssh-keygen"
	args := []string{"-t", "rsa", "-b", "4096", "-C", "ssh_key@cloudplex.com", "-f", keyName}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		beego.Error(err)
		return KeyPairResponse{}, err
	}
	beego.Info("Successfully generated sshkeys")

	arr, err1 := ioutil.ReadFile(keyName)
	str := string(arr)
	if err1 != nil {
		beego.Error(err1)
		return KeyPairResponse{}, err1
	}

	res.PrivateKey = str
	res.KeyName = keyName

	arr, err1 = ioutil.ReadFile(keyName + ".pub")
	str = string(arr)
	if err1 != nil {
		beego.Error(err1)
		return KeyPairResponse{}, err1
	}
	res.PublicKey = str
	return res, nil
}

func InitReq() HTTPClient {

	var client HTTPClient
	client.client = &http.Client{}
	return client

}
