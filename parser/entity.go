package parser

import "strings"

type Entity struct {
	Header  Header
	Columns []Column
}

func ParseEntity(raw string) Entity {
	raw = strings.TrimSpace(raw)
	lines := strings.Split(raw, "\n")
	var parsed []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		// Skip comments.
		if strings.HasPrefix(line, "#") {
			continue
		}
		parsed = append(parsed, line)
	}
	head, body := parsed[0], parsed[1:]
	header := ParseHeader(head)
	columns := ParseColumns(body)

	return Entity{Header: header, Columns: columns}
}
