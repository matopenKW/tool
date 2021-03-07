package model

{{ .ModelName }}
{{ .TableName }}

{{range $index, $column := .Columns}}{{ Func $column 0 }},{{end}}
