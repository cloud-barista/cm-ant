package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func WritePropertiesFile(filePath string, properties map[string]interface{}, emptyOmit bool) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for key, value := range properties {
		strValue := InterfaceToString(value)
		if len(strings.TrimSpace(strValue)) == 0 && emptyOmit {
			continue
		}
		_, err := fmt.Fprintf(file, "%s=%v\n", key, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateFolder(filename string) error {
	err := os.Mkdir(filename, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func ExistCheck(path string) bool {
	_, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			log.Println("file / folder does not exist on", path)

		} else {
			log.Println(err)
		}

		return false
	}

	return true
}
