package model


{{ .ModelID }}

// {{ .ModelName }} is relation between user and account.
type {{ .ModelName }} struct {
	{{range $index, $columnName := .ColumnNames}}{{ $columnName }}
{{end}}
}

// {{ .ModelName }}s is a set of a {{ .ModelName }}.
type {{ .ModelName }}s []*{{ .ModelName }}
