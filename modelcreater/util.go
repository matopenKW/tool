package modelcreater

import (
	"fmt"
	"strings"
)

func ConvertSnakeToCamel(s string, isPascal bool) string {
	if len(s) == 0 {
		return ""
	}
	st := strings.Split(s, "_")

	ret := ""
	for _, s := range st {
		ret += fmt.Sprintf("%s%s", strings.ToUpper(s[0:1]), s[1:])
	}

	if !isPascal {
		return fmt.Sprintf("%s%s", strings.ToLower(ret[0:1]), ret[1:])
	}
	return ret
}

func GetColumnType(typeStr string) string {
	switch typeStr {
	case "text":
		return "string"
	case "datetime", "timestamp":
		return "time.Time"
	case "double":
		return "float64"
	default:
		if strings.Index(typeStr, "varchar") > -1 {
			return "string"
		} else if strings.Index(typeStr, "int") > -1 {
			return "int"
		} else if strings.Index(typeStr, "tinyint") > -1 {
			return "bool"
		}
		return ""
	}
}
