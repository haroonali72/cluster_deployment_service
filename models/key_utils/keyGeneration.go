package key_utils

import (
	"antelope/models"
	"antelope/models/utils"
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"time"
)

func KeyConversion(keyInfo interface{}, ctx utils.Context) (utils.Key, error) {
	b, e := json.Marshal(keyInfo)
	var k utils.Key
	if e != nil {
		ctx.SendLogs(e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Log)

		return utils.Key{}, e
	}
	e = json.Unmarshal(b, &k)
	if e != nil {
		ctx.SendLogs(e.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Log)
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
		ctx.SendLogs(err.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Log)
		return utils.KeyPairResponse{}, err
	}

	ctx.SendLogs("Successfully generated sshkeys", models.LOGGING_LEVEL_INFO, models.Backend_Log)
	arr, err1 := ioutil.ReadFile(keyName)
	str := string(arr)
	if err1 != nil {
		ctx.SendLogs(err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Log)
		return utils.KeyPairResponse{}, err1
	}

	res.PrivateKey = str
	res.KeyName = keyName

	arr, err1 = ioutil.ReadFile(keyName + ".pub")
	str = string(arr)
	if err1 != nil {
		ctx.SendLogs(err1.Error(), models.LOGGING_LEVEL_ERROR, models.Backend_Log)
		return utils.KeyPairResponse{}, err1
	}
	res.PublicKey = str
	return res, nil
}
