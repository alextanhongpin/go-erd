package parser

import (
	"fmt"
	"strings"
)

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

func ParseColumns(columns []string) []Column {
	var result []Column
	for _, col := range columns {
		result = append(result, NewColumn(col))
	}
	return result
}
