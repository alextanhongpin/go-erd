package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Template struct {
	Title     string
	Relations []Relation
	Entities  []Entity
}

type Attribute struct {
	Name string
}

func NewAttribute(name string) Attribute {
	return Attribute{strings.TrimSpace(name)}
}

func (a Attribute) primary() bool {
	return strings.HasPrefix(a.Name, "*") || a.foreignAndPrimary()
}

func (a Attribute) foreignAndPrimary() bool {
	return strings.HasPrefix(a.Name, "*+") || strings.HasPrefix(a.Name, "+*")
}

func (a Attribute) foreign() bool {
	return strings.HasPrefix(a.Name, "+") || a.foreignAndPrimary()
}

func (a Attribute) String() string {
	name := a.Name
	if a.primary() {
		name = fmt.Sprintf("<U>%s</U>", name)
	}
	if a.foreign() {
		name = fmt.Sprintf("<I>%s</I>", name)
	}
	return name
}

type Option struct {
	Color string `json:"color"`
}

type Entity struct {
	Comments   []string
	Name       string
	Attributes []Attribute
	Option     Option
}

type Relation struct {
	From, To, FromCardinal, ToCardinal string
}

var relationsMapper = map[string]string{
	"1": "1",
	"?": "0..1",
	"+": "1..N",
	"*": "0..N",
}

func (r Relation) String() string {
	return fmt.Sprintf("%q -> %q[taillabel=%q, headlabel=%q]",
		r.From,
		r.To,
		relationsMapper[r.ToCardinal],
		relationsMapper[r.FromCardinal],
	)
}

func NewRelationFromSlice(r []string) Relation {
	if len(r) != 4 {
		return Relation{}
	}
	return Relation{
		From:         strings.TrimSpace(r[0]),
		To:           strings.TrimSpace(r[3]),
		FromCardinal: strings.TrimSpace(r[1]),
		ToCardinal:   strings.TrimSpace(r[2]),
	}
}

func main() {
	var in, out string
	flag.StringVar(&in, "in", "in.txt", "the input file to read")
	flag.StringVar(&out, "out", "out.png", "the output file to write")
	flag.Parse()
	f, err := os.Open(in)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	scanner := bufio.NewScanner(f)

	var partitions []string
	var tmp []string

	addToPartition := func() {
		partitions = append(partitions, strings.Join(tmp, "\n"))
		tmp = tmp[:0]
	}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 && len(tmp) > 0 {
			addToPartition()
		}
		tmp = append(tmp, line)
	}
	if len(tmp) > 0 {
		addToPartition()
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "shouldn't see an error scanning a string")
	}

	var entities []Entity
	var relations []Relation
	var title string
	for _, part := range partitions {
		entity := matchEntity(part)
		if t := matchTitle(part); len(t) > 0 {
			title = t
		}
		comments := matchComments(part)
		relation := matchRelations(part)

		if len(relation) > 0 {
			for _, r := range relation {
				relations = append(relations, NewRelationFromSlice(r))
			}
		}
		if len(entity) > 0 {
			rawOpts := entity[1]
			opt := Option{Color: "#eeeeee"}
			_ = json.Unmarshal([]byte(rawOpts), &opt)

			rows := parseEntity(entity[0])
			name, rawAttributes := rows[0], rows[1:]
			var attributes []Attribute
			for _, raw := range rawAttributes {
				attributes = append(attributes, NewAttribute(raw))
			}

			e := Entity{
				Name:       name,
				Attributes: attributes,
				Comments:   comments,
				Option:     opt,
			}

			entities = append(entities, e)
		}
	}
	var data Template
	data.Title = title
	data.Entities = entities
	data.Relations = relations

	outFile := strings.Split(out, ".")
	ext := outFile[1]
	path, _ := exec.LookPath("dot")
	cmd := exec.Command(path, fmt.Sprintf("-T%s", ext))
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer stdin.Close()
		Render(data, stdin)
	}()
	writeCmd, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	mode := int(0644)
	ioutil.WriteFile(out, writeCmd, os.FileMode(mode))
}

func matchEntity(s string) []string {
	re := regexp.MustCompile(`(?m)(\[.+\]([{}:"#\n\s\w]+)?[\s\w\[\]!"#$%&'()*+, -./:;<=>?@[\]^_{|}~]+$)`)
	matches := re.FindAllStringSubmatch(s, -1)
	if len(matches) > 0 {
		var result []string
		for _, r := range matches {
			result = append(result, r[1:]...)
		}
		return result
	}
	return nil
}

func matchTitle(s string) string {
	s = strings.TrimSpace(s)
	result := regexp.MustCompile(`(?m)^Title:([\w\s]+)$`).FindStringSubmatch(s)
	if len(result) > 0 {
		return strings.TrimSpace(result[1])
	}
	return ""
}

func matchComments(s string) []string {
	re := regexp.MustCompile(`(?m)^#([\w!"#$%&'()*+, -./:;<=>?@[\]^_{|}~]+)$`)
	matches := re.FindAllStringSubmatch(s, -1)
	if len(matches) > 0 {
		var result []string
		for _, r := range matches {
			result = append(result, r[1:]...)
		}
		return result
	}
	return nil
}

func matchRelations(s string) [][]string {
	re := regexp.MustCompile(`(?m)^([\w ]+)([?1+*])--([?1+*])([\w ]+)$`)
	matches := re.FindAllStringSubmatch(s, -1)
	if len(matches) > 0 {
		var result [][]string
		for _, r := range matches {
			result = append(result, r[1:])
		}
		return result
	}
	return nil
}

func parseEntity(s string) []string {
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	head, body := lines[0], lines[1:]
	name := regexp.MustCompile(`\[(.+?)\]`).FindStringSubmatch(head)[1]

	var result []string
	result = append(result, strings.TrimSpace(name))
	for _, row := range body {
		result = append(result, strings.TrimSpace(row))
	}
	return result
}

var tpl = `
digraph G {
    // Title.
    pencolor=black
    fontsize=16
    labelloc=t
    label = "{{- .Title -}}"
    rankdir=LR;
    graph [pad="0.5", nodesep="1", ranksep="2"];
    //
    // Box for entities.
    //
    node [shape=none, margin=0]
    //
    // Relationship Edges.
    //
    edge[arrowhead=none, arrowtail=none, dirType=both, style=dashed,color="#888888"];
    //
    // Entities.
    //
    {{- range $entity := .Entities}}
    "{{$entity.Name}}" [label={{noescape "<"}}
	<table border="0" cellborder="1" cellspacing="0" cellpadding="4">
	    <tr><td  bgColor="{{$entity.Option.Color}}" align="left">{{noescape $entity.Name}}</td></tr>
	    {{- range $attr := $entity.Attributes}}
	    <tr><td align="left">{{noescape $attr.String}}</td></tr>
	    {{- end}}
	</table>
    >]
    {{- end}}
    //
    // Relationships.
    //
    {{- range $rel := .Relations}}
    {{noescape $rel.String}}
    {{- end}}
}`

func Render(data Template, writer io.Writer) {
	var noescape = func(value string) template.HTML {
		return template.HTML(value)
	}
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"noescape": noescape,
	}).Parse(tpl))
	tmpl.Execute(writer, data)
}
