package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
	"strconv"
)

const (
	RE_NOT_SEPARATORS    = `[^[\s\.,:\*\+;\?\\\-—_\(\)\[\]{}<>'"#«»№\/!]+`
	RE_UNSUPPORTED_CHARS = `[\\\/\*\+\?]`
)

var (
	re_split = regexp.MustCompile(RE_NOT_SEPARATORS)
	re_unsup = regexp.MustCompile(RE_UNSUPPORTED_CHARS)
)


func ReplaceUnsupported(str string) string {
	return re_unsup.ReplaceAllString(str, "_")
}

func UpperInitial(str string) string {
	if len(str) > 0 {
		process := strings.ToLower(str)
		r, size := utf8.DecodeRuneInString(process)
		return string(unicode.ToUpper(r)) + process[size:]
	}
	return ""
}

func UpperInitialAll(src string) string {
	return re_split.ReplaceAllStringFunc(src, func(str string) string {
		return UpperInitial(str)
	})
}

func SplitBySeparators(src string) []string {
	return re_split.FindAllString(src, -1)
}

func ParseBool(str string) (value bool, err error) {
	switch str {
	case "1", "t", "T", "true", "TRUE", "True":
		return true, nil
	case "0", "f", "F", "false", "FALSE", "False":
		return false, nil
	}
	return false, fmt.Errorf("ParseBool %s", str)
}

func ParseInt(value string) (result int) {
	if result, err := strconv.Atoi(value); err == nil {
		return result
	}
	return result
}

func PrintJson(object interface{}) {
	jsonBytes, err := json.MarshalIndent(object, "", "  ")
	if err == nil {
		fmt.Println(string(jsonBytes))
	} else {
		log.Fatalln("Invalid object")
	}
}


