package key_utils

import (
	"antelope/models"
	"antelope/models/utils"
	"encoding/json"
	"github.com/astaxie/beego"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"
)

type AWSKey struct {
	KeyName     string         `json:"key_name" bson:"key_name" valid:"required"`
	KeyType     models.KeyType `json:"key_type" bson:"key_type" valid:"required, in(new|cp|aws|user)"`
	KeyMaterial string         `json:"private_key" bson:"private_key"`
	Cloud       models.Cloud   `json:"cloud" bson:"cloud"`
}

type AZUREKey struct {
	CredentialType models.CredentialsType `json:"credential_type"  bson:"credential_type" description:"Type of credentials used for SSH.Valid values are 'key' and 'password'.Default value is 'key'. [optional]"`
	KeyType        models.KeyType         `json:"key_type" bson:"key_type" validate:"required,eq=new|eq=cp|eq=aws|eq=user" description:"Type of key. By default value is 'new' [required]"`
	KeyName        string                 `json:"key_name" bson:"key_name" validate:"required" description:"Name of the key.[required]"`
	Username       string                 `json:"username" bson:"username"  description:"Name of the user ['optional']"`
	AdminPassword  string                 `json:"admin_password" bson:"admin_password" description:"It is the password ['required' if credentials_type is 'password']"`
	PrivateKey     string                 `json:"private_key" bson:"private_key,omitempty" description:"Private key for SSH ['required'] if credentials_type is 'key']"`
	PublicKey      string                 `json:"public_key" bson:"public_key,omitempty" description:"Public key for ssh [readonly]"`
	ID             int                    `json:"id" bson:"id,omitempty"`
	FingerPrint    string                 `json:"finger_print" bson:"finger_print,omitempty" description:"Finger prints for the key [optional]"`
	Cloud          models.Cloud           `json:"cloud" bson:"cloud" description:"Name of the cloud [optional]"`
}

type KeyPairResponse struct {
	KeyName    string `json:"key_name"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

func AWSKeyCoverstion(keyInfo []byte, ctx utils.Context) (AWSKey, error) {
	//b, e := json.Marshal(keyInfo)
	var k AWSKey
	//	if e != nil {
	//		ctx.SendLogs(e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	//		return AWSKey{}, e
	//	}
	e := json.Unmarshal(keyInfo, &k)
	if e != nil {
		ctx.SendLogs(e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return AWSKey{}, e
	}
	return k, nil
}
func AzureKeyConversion(keyInfo []byte, ctx utils.Context) (AZUREKey, error) {
	var k AZUREKey
	e := json.Unmarshal(keyInfo, &k)
	if e != nil {
		ctx.SendLogs(e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return AZUREKey{}, e
	}
	return k, nil
}

func GenerateKeyPair(keyName, username string, ctx utils.Context) (KeyPairResponse, error) {

	res := KeyPairResponse{}
	keyName = strings.TrimSpace(keyName)
	username = strings.TrimSpace(username)
	t := time.Now().Local()
	tstamp := t.Format("20060102150405")
	keyName = keyName + "_" + tstamp

	cmd := "ssh-keygen"
	args := []string{"-t", "rsa", "-b", "4096", "-C", username, "-f", keyName}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KeyPairResponse{}, err
	}

	ctx.SendLogs("Successfully generated sshkeys", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	arr, err1 := ioutil.ReadFile(keyName)
	str := string(arr)
	if err1 != nil {
		ctx.SendLogs(err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KeyPairResponse{}, err1
	}

	res.PrivateKey = str
	res.KeyName = keyName

	arr, err1 = ioutil.ReadFile(keyName + ".pub")
	str = string(arr)
	if err1 != nil {
		ctx.SendLogs(err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return KeyPairResponse{}, err1
	}
	res.PublicKey = str
	return res, nil
}

func GenerateKey(cloud models.Cloud, keyName, userName, token, teams string, ctx utils.Context) (AZUREKey, error) {

	var keyInfo AZUREKey
	//_, err := vault.GetSSHKey(string(cloud), keyName, token, ctx, "")
	//if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
	//	ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
	//	beego.Error(err.Error())
	//	return keyInfo, err
	//}
	//if err == nil {
	//	return keyInfo, errors.New("Key already exist")
	//}
	if userName == "" {
		userName = "cloudplex"
	}

	res, err := GenerateKeyPair(keyName, userName, ctx)
	if err != nil {
		beego.Error("vm creation failed with error: " + err.Error())
		return keyInfo, err
	}

	keyInfo.Cloud = cloud
	keyInfo.KeyName = keyName
	keyInfo.Username = userName
	keyInfo.PrivateKey = res.PrivateKey
	keyInfo.PublicKey = strings.TrimSuffix(res.PublicKey, "\n")

	ctx.SendLogs("SSHKey Created. ", models.LOGGING_LEVEL_INFO, models.Audit_Trails)
	beego.Info("SSHKey Created. ", keyInfo.PrivateKey)

	return keyInfo, nil
}
