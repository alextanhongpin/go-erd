package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

type TemplateData struct {
	Entities  []Entity
	Relations []Relation
}
type Entity struct {
	Title   string
	Columns []string
}

type Relation struct {
	From, To, FromCardinal, ToCardinal string
}

var tpl = `
digraph G {
    // Title.
    pencolor=black
    fontsize=16
    labelloc=t
    label = "title"

    rankdir=LR;
    graph [pad="0.5", nodesep="1", ranksep="2"];

    // Box for entities
    node [shape=none, margin=0]

    // One-to-many relation (from one, to many)
    // edge [arrowhead=crow, arrowtail=none, dirType=both]
    edge[arrowhead=none, arrowtail=none, dirType=both, style=dashed,color="#888888"];

    //
    // Entities
    //
    {{- range $entity := .Entities}}
    "{{$entity.Title}}" [label={{noescape "<"}}
	<table border="0" cellborder="1" cellspacing="0" cellpadding="4">
	    <tr><td align="left">{{noescape $entity.Title}}</td></tr>
	    {{- range $col := $entity.Columns}}
	    <tr><td align="left">{{noescape $col}}</td></tr>
	    {{- end}}
	</table>
    >]
    {{- end}}

    //
    // Relationships
    //
    {{- range $rel := .Relations}}
    "{{$rel.From}}"->"{{$rel.To}}"[taillabel="{{$rel.FromCardinal}}", headlabel="{{$rel.ToCardinal}}"];
    {{- end}}
}`

func previewTemplate(data TemplateData) {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"noescape": func(value string) template.HTML {
			return template.HTML(value)
		}}).Parse(tpl))
	f, err := os.Create("out.dot")
	if err != nil {
		log.Println("create file: ", err)
		return
	}
	// err := tmpl.Execute(os.Stdout, data)
	err = tmpl.Execute(f, data)
	if err != nil {
		log.Fatalf("execution failed: %s", err)
	}
}

func main() {
	content, err := ioutil.ReadFile("in.txt")
	if err != nil {
		log.Fatal(err)
	}
	raw := string(content)
	var data TemplateData
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
				data.Entities = make([]Entity, 0)
			}
			data.Entities = append(data.Entities, Entity{Title: header, Columns: columns})
		}
	}
	{
		re := regexp.MustCompile(`(?m)^([\w\s]+)([?1+*])--([?1+*])([\w\s]+)$`)
		result := re.FindAllStringSubmatch(raw, -1)
		for _, relations := range result {
			fromEntity, toRelation := strings.TrimSpace(relations[1]), relations[2]
			fromRelation, toEntity := relations[3], strings.TrimSpace(relations[4])

			if len(data.Relations) == 0 {
				data.Relations = make([]Relation, 0)
			}
			data.Relations = append(data.Relations, Relation{
				fromEntity,
				toEntity,
				relationsMapper[toRelation],
				relationsMapper[fromRelation],
			})

		}
	}
	previewTemplate(data)
	// parser(raw)

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
		if col.IsPrimaryKey() {
			res = fmt.Sprintf(`<u>%s</u>`, res)
		}
		if col.IsForeignKey() {
			res = fmt.Sprintf(`<i>%s</i>`, res)
		}
		result = append(result, res)
	}
	return result
}
