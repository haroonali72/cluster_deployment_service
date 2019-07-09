package utils

import (
	"encoding/base64"
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/astaxie/beego"
	"strings"
)

type GcpCredentials struct {
	Raw           string `json:"raw"`
	Type          string `json:"type" valid:"required"`
	ProjectId     string `json:"project_id" valid:"required"`
	PrivateKeyId  string `json:"private_key_id" valid:"required"`
	PrivateKey    string `json:"private_key" valid:"required"`
	ClientEmail   string `json:"client_email" valid:"required"`
	ClientId      string `json:"client_id" valid:"required"`
	AuthUri       string `json:"auth_uri" valid:"required"`
	TokenUri      string `json:"token_uri" valid:"required"`
	AuthProvider  string `json:"auth_provider_x509_cert_url" valid:"required"`
	ClientCertUrl string `json:"client_x509_cert_url" valid:"required"`
}

func IsValdidGcpCredentials(credentials string) (bool, GcpCredentials) {
	gcpCredentials := GcpCredentials{}

	decoded, err := base64.StdEncoding.DecodeString(credentials)
	if err != nil {
		beego.Error(err.Error())
		return false, gcpCredentials
	}

	decodedCredentials := string(decoded)
	if decodedCredentials == "" ||
		strings.Contains(strings.ToLower(decodedCredentials), "bearer") ||
		strings.Contains(strings.ToLower(decodedCredentials), "aws") ||
		strings.Contains(strings.ToLower(decodedCredentials), "azure") {
		return false, gcpCredentials
	}

	err = json.Unmarshal(decoded, &gcpCredentials)
	if err != nil {
		beego.Error(err.Error())
		return false, gcpCredentials
	}

	_, err = govalidator.ValidateStruct(gcpCredentials)
	if err != nil {
		beego.Error(err.Error())
		return false, gcpCredentials
	}

	gcpCredentials.Raw = decodedCredentials
	return true, gcpCredentials
}

