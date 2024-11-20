package utils

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
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

func CreateFolderIfNotExist(filePath string) error {
	if exist := ExistCheck(filePath); !exist {
		err := CreateFolder(filePath)

		if err != nil {
			return err
		}
	}

	return nil
}
func CreateFolder(filePath string) error {
	err := os.Mkdir(filePath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func ReadCSV(filename string) (*[][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Error().Msgf("can not opent files: %s; %s", filename, err)
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	parsedCsv, err := reader.ReadAll()
	if err != nil {
		log.Printf("Failed to read CSV file from path %s; %v", filename, err)
		return nil, err
	}

	return &parsedCsv, nil
}

func ExistCheck(path string) bool {
	fileInfo, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			log.Error().Msgf("file / folder does not exist: %s", path)

		} else {
			log.Error().Msg(err.Error())
		}

		return false
	}

	if fileInfo.Size() == 0 {
		return false
	}

	return true
}

func ReadToString(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Error().Msgf("file doesn't exist on correct path; %v", err)
		return "", err
	}

	return string(data), nil
}
