package model

type {{ .ModelName }} struct {
    {{range $index, $column := .Columns}}{{ GetField $column }}
    {{end}}
}


