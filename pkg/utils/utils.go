package utils

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
)

func SysCall(cmdStr string) (string, error) {
	cmd := exec.Command("bash", "-c", cmdStr)

	cmdOut, err := cmd.Output()

	if err != nil {
		return "", err
	}

	return string(cmdOut), nil

}

func StructToMap(obj interface{}) map[string]interface{} {
	objValue := reflect.ValueOf(obj)
	objType := objValue.Type()

	data := make(map[string]interface{})

	for i := 0; i < objValue.NumField(); i++ {
		field := objType.Field(i)
		data[field.Name] = objValue.Field(i).Interface()
	}

	return data
}

func WritePropertiesFile(filePath string, properties map[string]interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for key, value := range properties {
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
