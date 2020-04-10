package main

import (
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
	raw := string(content)
	var data renderer.Data
	relationsMapper := map[string]string{
		"1": "1",    // exactly 1
		"*": "0..N", // 0 or more
		"+": "1..N", // 1 or more
		"?": "0..1", // 0 or 1
	}
	{
		re := regexp.MustCompile(`(?m)(\[.+\]([{}:"\n\s\w]+)?[\s\w\n*+]+$)`)
		result := re.FindAllStringSubmatch(raw, -1)
		for _, entities := range result {
			entity := parseEntity(entities[0])
			header, columns := entity[0], entity[1:]
			if len(data.Entities) == 0 {
				data.Entities = make([]renderer.Entity, 0)
			}
			data.Entities = append(data.Entities, renderer.Entity{Title: header, Columns: columns})
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
				fromEntity,
				toEntity,
				relationsMapper[toRelation],
				relationsMapper[fromRelation],
			})

		}
	}
	renderer.Render(data)

}
func parseHeader(header string) string {
	// TODO: Map options for the header.
	re := regexp.MustCompile(`\[(.+)\](\s+\{.+\})?`)
	result := re.FindStringSubmatch(header)
	if len(result) > 2 {
		return strings.TrimSpace(result[1])
	}
	return strings.TrimSpace(result[1])
}

type Column struct {
	Name string
}

func (c Column) String() string {
	return c.Name
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

func parseEntity(raw string) []string {
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
	header = parseHeader(header)
	columns := parseBody(body)
	var result []string
	result = append(result, header)
	for _, col := range columns {
		res := col.Name
		// NOTE: Italic and underline only works for .svg, not .png.
		if col.IsPrimaryKey() {
			res = fmt.Sprintf(`<U>%s</U>`, res)
		}
		if col.IsForeignKey() {
			res = fmt.Sprintf(`<I>%s</I>`, res)
		}
		result = append(result, res)
	}
	return result
}
