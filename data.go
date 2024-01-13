package main

import "text/template"

var dataTemplate = template.Must(template.New("").Funcs(template.FuncMap{
	"inc": func(i int) int { return i + 1 },
}).Parse(`
local data = {}
{{- range $i, $v := . }}
data[{{ inc $i }}] = {{ printf "%q" $v }}
{{- end }}
data.attached = function (id)
	local script = msg.url(nil, id, "data")
	local index = go.get(script, "data")
	return data[index]
end
return data
`))
