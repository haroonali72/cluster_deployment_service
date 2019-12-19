package utils

import (
	"github.com/astaxie/beego"
	"os"
)

func GetUserData(fileName string) (error, string) {
	return FileRead(fileName)
}
func FileRead(fileName string) (error, string) {

	f, err := os.Open(fileName)
	if err != nil {
		return err, ""
	}
	defer f.Close()
	var contents []byte
	n2, err := f.Read(contents)
	if err != nil {
		return err, ""
	}
	beego.Info("read %d bytes\n", n2)
	return nil, string(contents)
}
