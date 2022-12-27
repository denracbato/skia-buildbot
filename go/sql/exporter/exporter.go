package exporter

import (
	"fmt"
	"reflect"
	"strings"
)

// GenerateSQL takes in a "table type", that is a table whose fields are slices. Each field
// will be interpreted as a table. The sql struct tags will be used to generate the SQL schema.
// A package name is taken in to be included in the returned string. If a malformed type is passed
// in, this function will panic.
func GenerateSQL(inputType interface{}, pkg string) string {
	header := fmt.Sprintf("package %s\n\n// Generated by //go/sql/exporter/\n// DO NOT EDIT\n\nconst Schema = `", pkg)

	body := strings.Builder{}
	t := reflect.TypeOf(inputType)
	for i := 0; i < t.NumField(); i++ {
		table := t.Field(i) // Fields of the outer type are expected to be tables.
		if table.Type.Kind() != reflect.Slice {
			panic(`Expected table should be a slice: ` + table.Name)
		}
		body.WriteString("CREATE TABLE IF NOT EXISTS ")
		body.WriteString(table.Name)
		body.WriteString(" (")
		row := table.Type.Elem()
		wasFirst := true
		for j := 0; j < row.NumField(); j++ {
			col := row.Field(j)
			sqlText, ok := col.Tag.Lookup("sql")
			if !ok {
				panic(`Field missing "sql" tag:` + table.Name + "." + row.Name())
			}
			if !wasFirst {
				body.WriteString(",")
			}
			wasFirst = false
			body.WriteString("\n  ")
			body.WriteString(strings.TrimSpace(sqlText))
		}
		body.WriteString("\n);\n")
	}
	const footer = "`\n"
	return header + body.String() + footer
}
