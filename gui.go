package main

import "text/template"

var guiTemplate = template.Must(template.New("").Parse(`
script: ""
{{ range .Textures }}
textures {
  name: "{{ . }}"
  texture: "/import/{{ . }}.atlas"
}
{{ end }}
background_color {
  x: 0.0
  y: 0.0
  z: 0.0
  w: 0.0
}
{{ range .Elements }}
nodes {
  position {
    x: {{ .X }}.0
    y: {{ .Y }}.0
    z: 0.0
    w: 1.0
  }
  rotation {
    x: 0.0
    y: 0.0
    z: 0.0
    w: 1.0
  }
  scale {
    x: 1.0
    y: 1.0
    z: 1.0
    w: 1.0
  }
  size {
    x: {{ .W }}.0
    y: {{ .H }}.0
    z: 0.0
    w: 1.0
  }
  color {
    x: 1.0
    y: 1.0
    z: 1.0
    w: 1.0
  }
  type: TYPE_BOX
  blend_mode: BLEND_MODE_ALPHA
  {{ if .Group }}
  texture: "ui/{{ .Group }}_{{ .Name }}"
  {{ else }}
  texture: ""
  {{ end }}
  id: "{{ .Name }}"
  xanchor: XANCHOR_NONE
  yanchor: YANCHOR_NONE
  pivot: PIVOT_NW
  adjust_mode: ADJUST_MODE_FIT
  layer: ""
  inherit_alpha: true
  slice9 {
    x: 0.0
    y: 0.0
    z: 0.0
    w: 0.0
  }
  clipping_mode: CLIPPING_MODE_NONE
  clipping_visible: true
  clipping_inverted: false
  alpha: 1.0
  template_node_child: false
  {{ if .Group }}
  size_mode: SIZE_MODE_AUTO
  {{ else }}
  size_mode: SIZE_MODE_MANUAL
  {{ end }}
  custom_type: 0
  enabled: true
  visible: true
}
{{ end }}
material: "/builtins/materials/gui.material"
adjust_reference: ADJUST_REFERENCE_PARENT
max_nodes: 512
`))
