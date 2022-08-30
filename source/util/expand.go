package util

import (
	"os"
	"regexp"
	"strings"
	"unicode"
)

// The name of a variable can contain only letters (a to z or A to Z), numbers ( 0 to 9) or
// the underscore character ( _), and can't begin with number.
const envVariable = `\${([a-zA-Z_]{1}[\w]+)((?:\,\,|\^\^|\,|\^)?)[\|]{2}(.*?)}`

// reg exp
var variableReg *regexp.Regexp

func init() {
	variableReg = regexp.MustCompile(envVariable)
}

// if string like ${NAME||archaius}
// will query environment variable for ${NAME}
// if environment variable is "" return default string `archaius`
// support multi variable, e.g.:
//    value string => addr:${IP||127.0.0.1}:${PORT||8080}
//    if environment variable =>  IP=0.0.0.0 PORT=443 , result => addr:0.0.0.0:443
//    if no exist environment variable                , result => addr:127.0.0.1:8080
// support toupper / tolower like bash,
//    1. ^^: whole string to upper
//    2.  ^: capital to upper
//    3. ,,: whole string to lower
//    4.  ,: capital to lower
//    e.g.:
//       os.Setenv("env", "TesT")
//       ExpandValueEnv("${env^^||local}")  return "TEST"
//       ExpandValueEnv("${env,,||local}")  return "test"
//       ExpandValueEnv("${env^||local}")   return "TesT"
//       ExpandValueEnv("${env,||local}")   return "tesT"
func ExpandValueEnv(value string) (realValue string) {
	value = strings.TrimSpace(value)
	submatch := variableReg.FindAllStringSubmatch(value, -1)
	if len(submatch) == 0 {
		return value
	}

	realValue = value
	for _, sub := range submatch {
		if len(sub) != 4 { //rule matching behaves in an unexpected way
			continue
		}
		item := os.Getenv(sub[1])
		if item == "" {
			item = sub[3]
		} else {
			if sub[2] == "^^" {
				item = strings.ToUpper(item)
			} else if sub[2] == ",," {
				item = strings.ToLower(item)
			} else if sub[2] == "^" {
				item = makeFirstUpperCase(item)
			} else if sub[2] == "," {
				item = makeFirstLowerCase(item)
			}
		}
		realValue = strings.ReplaceAll(realValue, sub[0], item)
	}

	return
}

func makeFirstLowerCase(s string) string {
	if len(s)==0 {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func makeFirstUpperCase(s string) string {
	if len(s)==0 {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

