package key_utils

import (
	"antelope/models/logging"
	"antelope/models/utils"
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"time"
)

func KeyConversion(keyInfo interface{}, ctx logging.Context) (utils.Key, error) {
	b, e := json.Marshal(keyInfo)
	var k utils.Key
	if e != nil {
		ctx.SendSDLog(e.Error(), "error")
		return utils.Key{}, e
	}
	e = json.Unmarshal(b, &k)
	if e != nil {
		ctx.SendSDLog(e.Error(), "error")
		return utils.Key{}, e
	}
	return k, nil
}

func GenerateKeyPair(keyName string, ctx logging.Context) (utils.KeyPairResponse, error) {

	res := utils.KeyPairResponse{}

	t := time.Now().Local()
	tstamp := t.Format("20060102150405")
	keyName = keyName + "_" + tstamp

	cmd := "ssh-keygen"
	args := []string{"-t", "rsa", "-b", "4096", "-C", "azure@example.com", "-f", keyName}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		ctx.SendSDLog(err.Error(), "error")
		return utils.KeyPairResponse{}, err
	}
	ctx.SendSDLog("Successfully generated sshkeys", "info")

	arr, err1 := ioutil.ReadFile(keyName)
	str := string(arr)
	if err1 != nil {
		ctx.SendSDLog(err1.Error(), "error")
		return utils.KeyPairResponse{}, err1
	}

	res.PrivateKey = str
	res.KeyName = keyName

	arr, err1 = ioutil.ReadFile(keyName + ".pub")
	str = string(arr)
	if err1 != nil {
		ctx.SendSDLog(err1.Error(), "error")
		return utils.KeyPairResponse{}, err1
	}
	res.PublicKey = str
	return res, nil
}
