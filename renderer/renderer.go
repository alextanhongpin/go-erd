package renderer

import (
	"html/template"
	"os"
)

type Data struct {
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

func Render(data Data) {
	var noescape = func(value string) template.HTML {
		return template.HTML(value)
	}
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"noescape": noescape,
	}).Parse(tpl))
	// f, err := os.Create("out.dot")
	// if err != nil {
	//         log.Println("create file: ", err)
	//         return
	// }
	tmpl.Execute(os.Stdout, data)
	// err = tmpl.Execute(f, data)
	// if err != nil {
	// log.Fatalf("execution failed: %s", err)
	// }
}
