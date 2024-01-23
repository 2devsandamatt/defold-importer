package main

import "text/template"

var tilemapTemplate = template.Must(template.New("").Parse(`
tile_set: "/game/levels.tilesource"
layers {
  id: "background"
  z: 0.0
  is_visible: 1
  {{- range . }}
  cell {
    x: {{ .X }}
    y: {{ .Y }}
    tile: {{ .Index }}
    h_flip: 0
    v_flip: 0
    rotate90: 0
  }
  {{- end }}
}
material: "/builtins/materials/tile_map.material"
blend_mode: BLEND_MODE_ALPHA
`))
