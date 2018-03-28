package utils

import (
	"bitbucket.org/enlab/mopds/modules/prettyjson"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
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

func PrintJson(object interface{}, color bool) {
	if color {
		jsonBytes, err := prettyjson.Marshal(object)
		if err == nil {
			fmt.Println(string(jsonBytes))
		} else {
			log.Fatalln("Invalid object")
		}
	} else {
		jsonBytes, err := json.MarshalIndent(object, "", "  ")
		if err == nil {
			fmt.Println(string(jsonBytes))
		} else {
			log.Fatalln("Invalid object")
		}
	}
}

func GetLangCode(s string) int {
	langcode := 9
	LangCodes := map[int]string{
		1: "АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЬЫЪЭЮЯабвгдеёжзийклмнопрстуфхцчшщьыъэюя",
		2: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
		3: "0123456789",
	}

	for i := range LangCodes {
		for pos, c := range s {
			if pos == 0 {
				if strings.ContainsRune(LangCodes[i], c) {
					langcode = i
				}
			}
		}
	}
	return langcode
}

type Genres []struct {
	Genre      string `json:"genre"`
	Section    string `json:"section"`
	Subsection string `json:"subsection"`
}

type Genre struct {
	Genre      string `json:"genre"`
	Section    string `json:"section"`
	Subsection string `json:"subsection"`
}

func LoadGenres(file string) (Genres, error) {
	var genres Genres
	genresFile, err := os.Open(file)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer genresFile.Close()
	jsonParser := json.NewDecoder(genresFile)
	jsonParser.Decode(&genres)

	return genres, nil
}

func (genres Genres) GetDetail(genre_code string) Genre {
	var genre Genre

	for i := 0; i < len(genres); i++ {
		if genre_code == genres[i].Genre {
			genre = genres[i]
		}
	}

	return genre
}

func GetSizeCatalog(catalog string) int64 {
	fi, e := os.Stat(catalog)
	if e != nil {
		fmt.Println(e)
	}
	// get the size
	return fi.Size()
}

func GetSizeFile(s string) int {
	fi, e := os.Stat(s)
	if e != nil {
		log.Fatal(e)
	}
	// get the size
	size := int(fi.Size())

	return size
}

func TrimSlice(in []string) []string {
	for len(in) > 0 && in[len(in)-1] == "" {
		in = in[:len(in)-1]
	}
	return in
}
