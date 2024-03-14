package utils

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func CreateUniqIdBaseOnUnixTime() string {
	currentTime := time.Now()
	uniqId := fmt.Sprintf("%d-%s", currentTime.UnixMilli(), uuid.New().String())
	return uniqId
}

func SyncSysCall(cmdStr string) error {
	cmd := exec.Command("bash", "-c", cmdStr)

	out, err := cmd.Output()
	if err != nil {
		log.Println("error while execute system call,", err)
		return err
	}

	log.Println(string(out))
	return nil
}

func AsyncSysCall(cmdStr string) error {
	cmd := exec.Command("bash", "-c", cmdStr)

	err := cmd.Start()
	if err != nil {
		log.Println("error while execute system call", err)
		return err
	}
	return nil
}

func StructToMap(obj interface{}) map[string]interface{} {
	objValue := reflect.ValueOf(obj)
	objType := objValue.Type()

	data := make(map[string]interface{})

	for i := 0; i < objValue.NumField(); i++ {
		field := objType.Field(i)
		fieldName := FirstRuneToLower(field.Name)
		data[fieldName] = objValue.Field(i).Interface()
	}

	return data
}

func FirstRuneToLower(s string) string {
	if len(s) == 0 {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

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

func InterfaceToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	// Add more cases for other types as needed
	default:
		return fmt.Sprintf("%v", v)
	}
}

func CreateFolder(filename string) error {
	err := os.Mkdir(filename, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
