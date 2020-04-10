package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	"github.com/alextanhongpin/go-db-parser/renderer"
)

func main() {
	content, err := ioutil.ReadFile("in.txt")
	if err != nil {
		log.Fatal(err)
	}
	raw := strings.TrimSpace(string(content))
	var data renderer.Data
	relationsMapper := map[string]string{
		"1": "1",    // exactly 1
		"*": "0..N", // 0 or more
		"+": "1..N", // 1 or more
		"?": "0..1", // 0 or 1
	}
	{
		re := regexp.MustCompile(`(?m)^Title:([\w ]+)$`)
		result := re.FindStringSubmatch(raw)
		if len(result) > 1 {
			data.Title = strings.TrimSpace(result[1])
		}
	}
	{
		re := regexp.MustCompile(`(?m)(\[.+\]([{}:"#\n\s\w]+)?[\s\w\n*+()]+$)`)
		result := re.FindAllStringSubmatch(raw, -1)
		for _, entities := range result {
			entity := parseEntity(entities[0])
			var columns []string
			for _, col := range entity.Columns {
				columns = append(columns, col.String())
			}
			data.Entities = append(data.Entities, renderer.Entity{Title: entity.Header.Name, Columns: columns, Option: entity.Header.Option})
		}
	}
	{
		re := regexp.MustCompile(`(?m)^([\w\s]+)([?1+*])--([?1+*])([\w\s]+)$`)
		result := re.FindAllStringSubmatch(raw, -1)
		for _, relations := range result {
			fromEntity, toRelation := strings.TrimSpace(relations[1]), relations[2]
			fromRelation, toEntity := relations[3], strings.TrimSpace(relations[4])
			if len(data.Relations) == 0 {
				data.Relations = make([]renderer.Relation, 0)
			}
			data.Relations = append(data.Relations, renderer.Relation{
				From:         fromEntity,
				To:           toEntity,
				FromCardinal: relationsMapper[toRelation],
				ToCardinal:   relationsMapper[fromRelation],
			})
		}
	}
	renderer.Render(data)
}

type Header struct {
	Name   string
	Option renderer.Option
}

func parseHeader(header string) Header {
	// TODO: Map options for the header.
	re := regexp.MustCompile(`(?m)^\[(.+)\](\s+\{.+\})?`)
	result := re.FindStringSubmatch(header)
	hdr := strings.TrimSpace(result[1])
	defaultOption := renderer.Option{Color: "white"}
	if len(result) > 2 && len(result[2]) > 0 {
		var opt renderer.Option
		if err := json.Unmarshal([]byte(result[2]), &opt); err != nil {
			return Header{
				Name:   hdr,
				Option: defaultOption,
			}
		}
		return Header{
			Name:   hdr,
			Option: opt,
		}
	}
	return Header{Name: hdr, Option: defaultOption}
}

type Column struct {
	Name string
}

func (c Column) String() string {
	name := c.Name
	if c.IsPrimaryKey() {
		name = fmt.Sprintf(`<U>%s</U>`, name)
	}
	if c.IsForeignKey() {
		name = fmt.Sprintf(`<I>%s</I>`, name)
	}
	return name
}

func (c Column) multipleAttributes() bool {
	return strings.HasPrefix(c.Name, "*+") || strings.HasPrefix(c.Name, "+*")
}

func (c Column) IsForeignKey() bool {
	return strings.HasPrefix(c.Name, "+") || c.multipleAttributes()
}

func (c Column) IsPrimaryKey() bool {
	return strings.HasPrefix(c.Name, "*") || c.multipleAttributes()
}

func NewColumn(col string) Column {
	return Column{Name: strings.TrimSpace(col)}
}

func parseBody(body []string) []Column {
	var result []Column
	for _, col := range body {
		result = append(result, NewColumn(col))
	}
	return result
}

type Entity struct {
	Header  Header
	Columns []Column
}

func parseEntity(raw string) Entity {
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
	header, body := parsed[0], parsed[1:]
	headers := parseHeader(header)
	columns := parseBody(body)

	return Entity{Header: headers, Columns: columns}
}
