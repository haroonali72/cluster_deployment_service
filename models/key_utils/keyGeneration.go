package key_utils

import (
	"antelope/models"
	"antelope/models/utils"
	"antelope/models/vault"
	"encoding/json"
	"github.com/astaxie/beego"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"
)

func KeyConversion(keyInfo interface{}, ctx utils.Context) (utils.Key, error) {
	b, e := json.Marshal(keyInfo)
	var k utils.Key
	if e != nil {
		ctx.SendLogs(e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)

		return utils.Key{}, e
	}
	e = json.Unmarshal(b, &k)
	if e != nil {
		ctx.SendLogs(e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return utils.Key{}, e
	}
	return k, nil
}

func GenerateKeyPair(keyName, username string, ctx utils.Context) (utils.KeyPairResponse, error) {

	res := utils.KeyPairResponse{}

	t := time.Now().Local()
	tstamp := t.Format("20060102150405")
	keyName = keyName + "_" + tstamp

	cmd := "ssh-keygen"
	args := []string{"-t", "rsa", "-b", "4096", "-C", username, "-f", keyName}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return utils.KeyPairResponse{}, err
	}

	ctx.SendLogs("Successfully generated sshkeys", models.LOGGING_LEVEL_INFO, models.Backend_Logging)
	arr, err1 := ioutil.ReadFile(keyName)
	str := string(arr)
	if err1 != nil {
		ctx.SendLogs(err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return utils.KeyPairResponse{}, err1
	}

	res.PrivateKey = str
	res.KeyName = keyName

	arr, err1 = ioutil.ReadFile(keyName + ".pub")
	str = string(arr)
	if err1 != nil {
		ctx.SendLogs(err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		return utils.KeyPairResponse{}, err1
	}
	res.PublicKey = str
	return res, nil
}

func GenerateKey(cloud models.Cloud, keyName, userName, token, teams string, ctx utils.Context) (string, error) {
	var keyInfo utils.Key
	_, err := vault.GetAzureSSHKey(string(cloud), keyName, token, ctx)
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error(err.Error())
		return "", err
	}

	res, err := GenerateKeyPair(keyName, userName, ctx)
	if err != nil {
		beego.Error("vm creation failed with error: " + err.Error())
		return "", err
	}

	keyInfo.Cloud = cloud
	keyInfo.KeyName = keyName
	keyInfo.Username = userName
	keyInfo.PrivateKey = res.PrivateKey
	keyInfo.PublicKey = strings.TrimSuffix(res.PublicKey, "\n")

	beego.Info("Private Key in fetch ", keyInfo.PrivateKey)

	_, err = vault.PostAzureSSHKey(cloud, keyInfo, ctx, token, teams)
	if err != nil {
		beego.Error("vm creation failed with error: " + err.Error())
		return "", err
	}

	return keyInfo.PrivateKey, nil
}
func FetchKey(cloud models.Cloud, keyName, userName, token string, ctx utils.Context) (utils.Key, error) {

	var err error
	var key interface{}
	var empty utils.Key

	key, err = vault.GetAzureSSHKey(string(cloud), keyName, token, ctx)

	if err != nil && strings.Contains(strings.ToLower(err.Error()), "not found") {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("vm creation failed with error: " + err.Error())
		return empty, err
	}

	existingKey, err := KeyConversion(key, ctx)
	if err != nil {
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Logging)
		beego.Error("vm creation failed with error: " + err.Error())
		return empty, err
	}

	return existingKey, nil
}
