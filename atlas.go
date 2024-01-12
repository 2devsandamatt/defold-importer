package main

import "text/template"

var atlasTemplate = template.Must(template.New("").Parse(`
{{- range . }}
images {
image: "/import/img/{{ .Group }}_{{ .Name }}.png"
sprite_trim_mode: SPRITE_TRIM_MODE_OFF
}
{{ end -}}
margin: 2
extrude_borders: 0
inner_padding: 0
`))

var animationsTemplate = template.Must(template.New("").Parse(`
{{- range .}}
animations {
  id: "{{ .ID }}"
  {{- range .Frames }}
  images {
    image: "/import/img/{{ . }}"
    sprite_trim_mode: SPRITE_TRIM_MODE_OFF
  }
  {{- end }}
  playback: {{ .PlaybackMode }}
  fps: {{ .FPS }}
  flip_horizontal: 0
  flip_vertical: 0
}
{{ end -}}
margin: 2
extrude_borders: 0
inner_padding: 0
`))
