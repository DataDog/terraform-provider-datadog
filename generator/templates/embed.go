package templates

import "embed"

//go:embed datasource/*.tmpl
var DatasourceTemplates embed.FS
