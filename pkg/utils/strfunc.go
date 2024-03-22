package utils

import (
	"fmt"
	"github.com/google/uuid"
	"strconv"
	"time"
	"unicode"
)

func CreateUniqIdBaseOnUnixTime() string {
	currentTime := time.Now()
	uniqId := fmt.Sprintf("%d-%s", currentTime.UnixMilli(), uuid.New().String())
	return uniqId
}

func FirstRuneToLower(s string) string {
	if len(s) == 0 {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
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
