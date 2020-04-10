package parser

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/alextanhongpin/go-db-parser/renderer"
)

var headerRegexp *regexp.Regexp
var defaultOption = &renderer.Option{
	Color: "#eeeeee",
}

func init() {
	headerRegexp = regexp.MustCompile(`(?m)^\[(.+)\](\s+\{.+\})?`)
}

// Header represents the Entity header.
type Header struct {
	Name   string
	Option renderer.Option
}

// ParseHeader takes in a header of the format:
// `[table_name] {"color": "white"}`
// and returns a struct.
func ParseHeader(header string) Header {
	result := headerRegexp.FindStringSubmatch(header)
	opt := *defaultOption

	if len(result) > 2 && len(result[2]) > 0 {
		_ = json.Unmarshal([]byte(result[2]), &opt)
	}
	return Header{
		Name:   strings.TrimSpace(result[1]),
		Option: opt,
	}
}
