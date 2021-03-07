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

	uppercase := func(s string) bool {
		for _, upcase := range []string{
			"id",
			"url",
			"json",
		} {
			if s == upcase {
				return true
			}
		}
		return false
	}

	ret := ""
	for _, s := range st {
		if uppercase(s) {
			ret += strings.ToUpper(s)
		} else {
			ret += s
		}
	}

	if !isPascal {
		return fmt.Sprintf("%s%s", strings.ToLower(ret[0:1]), ret[1:])
	}
	return ret
}
