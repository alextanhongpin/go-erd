package main

import (
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	"github.com/alextanhongpin/go-db-parser/parser"
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
			entity := parser.ParseEntity(entities[0])
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
