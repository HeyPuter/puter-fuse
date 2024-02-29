package lang

import (
	"os"
	"strings"
)

func PathSplit(path string) []string {
	components := strings.Split(path, string(os.PathSeparator))
	// remove empty strings
	result := []string{}
	for _, c := range components {
		if c == "" {
			continue
		}
		result = append(result, c)
	}
	return result
}
