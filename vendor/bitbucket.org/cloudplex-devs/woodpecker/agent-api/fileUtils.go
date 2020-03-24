package agent_api

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func readFile(path string) (string, error) {

	text, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return string(text), nil
}

func createFile(path string) {
	// detect if file exists
	var _, err = os.Stat(path)

	// if inner directory does not exist
	if err != nil {
		err := os.MkdirAll(path[:strings.LastIndex(path, "/")], os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {

			return
		}
		defer file.Close()
	}
}

func writeFile(data, path string) error {

	var file, err = os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// write some text line-by-line to file
	_, err = file.WriteString(data)
	if err != nil {
		return err
	}

	// save changes
	err = file.Sync()
	if err != nil {
		return err
	}

	//utils.Info.Println("done creating file", path)

	return nil
}

func deleteFile(path string) error {

	err := os.Remove(path)
	return err
}
