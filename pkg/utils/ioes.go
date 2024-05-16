package utils

import (
	"encoding/csv"
	"fmt"
	"io"
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
	// 파일 열기
	file, err := os.Open(filename)
	if err != nil {
		log.Println("파일을 열 수 없습니다:", err)
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	var parsedCsv [][]string
	for {
		line, err := reader.Read()

		if err != nil {
			if err == io.EOF {
				return &parsedCsv, nil
			}

			log.Println(err)
			break
		}

		parsedCsv = append(parsedCsv, line)
	}

	return &parsedCsv, nil
}

func ExistCheck(path string) bool {
	fileInfo, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			log.Println("file / folder does not exist on", path)

		} else {
			log.Println(err)
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
		log.Println("file doesn't exist on correct path")
		return "", err
	}

	return string(data), nil
}
