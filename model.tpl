{{define "model.tpl"}}
package {{ .PackageName }}

import (
    {{ range $k, $v := .ImportList }}"{{ $k }}"
    {{ end }}
)

/**
{{ .CreateTableDDL }}
*/
type {{ .ModelName }} struct {
{{ range .RowsList }}    {{ . }}{{ end }}
}

func (model {{ .ModelName }}) DBName() string {
    return "{{ .DbName }}"
}

func (model {{ .ModelName }}) TableName() string {
    return "{{ .TableName }}"
}

{{end}}