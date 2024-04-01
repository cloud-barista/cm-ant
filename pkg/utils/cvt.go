package utils

import (
	"reflect"
)

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
