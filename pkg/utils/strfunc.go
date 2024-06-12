package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
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

func GetFirstPart(input, delim string) string {
	parts := strings.Split(input, delim)
	return parts[0]
}
