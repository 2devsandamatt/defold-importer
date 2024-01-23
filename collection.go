package main

import "text/template"

var collectionTemplate = template.Must(template.New("").Parse(`
name: "{{ .Filename }}"
scale_along_z: 0
embedded_instances {
  id: "world"
{{- range .Objects }}
{{- if .Group }}
  children: "{{ .Name }}"
{{- else }}
  children: "{{ .Name }}{{ .Index }}"
{{- end }}
{{- end }}
{{- range .Triggers }}
  children: "{{ .Name }}"
{{- end }}
  data: ""
  position {
    x: 0.0
    y: 0.0
    z: 0.0
  }
  rotation {
    x: 0.0
    y: 0.0
    z: 0.0
    w: 1.0
  }
  scale3 {
    x: 1.0
    y: 1.0
    z: 1.0
  }
}
{{- range .Objects }}
{{- if .Group }}
embedded_instances {
  id: "{{ .Name }}"
  data: "embedded_components {\n"
  "  id: \"sprite\"\n"
  "  type: \"sprite\"\n"
  "  data: \"tile_set: \\\"/import/{{ $.Filename }}.atlas\\\"\\n"
  "default_animation: \\\"{{ $.Filename }}_{{ .Name }}\\\"\\n"
  "material: \\\"/builtins/materials/sprite.material\\\"\\n"
  "blend_mode: BLEND_MODE_ALPHA\\n"
  "\"\n"
  "  position {\n"
  "    x: 0.0\n"
  "    y: 0.0\n"
  "    z: 0.0\n"
  "  }\n"
  "  rotation {\n"
  "    x: 0.0\n"
  "    y: 0.0\n"
  "    z: 0.0\n"
  "    w: 1.0\n"
  "  }\n"
  "}\n"
  ""
  position {
    x: {{ .X }}.0
    y: {{ .Y }}.0
    z: 0.0
  }
  rotation {
    x: 0.0
    y: 0.0
    z: 0.0
    w: 1.0
  }
  scale3 {
    x: 1.0
    y: 1.0
    z: 1.0
  }
}
{{- else }}
instances {
  id: "{{ .Name }}{{ .Index }}"
  prototype: "/game/objects/{{ .Name }}.go"
  position {
    x: {{ .X }}.0
    y: {{ .Y }}.0
    z: 0.4
  }
  rotation {
    x: 0.0
    y: 0.0
    z: 0.0
    w: 1.0
  }
  scale3 {
    x: 1.0
    y: 1.0
    z: 1.0
  }
}
{{- end }}
{{- end }}
{{- range .Triggers }}
embedded_instances {
  id: "{{ .Name }}"
  data: "components {\n"
  "  id: \"data\"\n"
  "  component: \"/import/data.script\"\n"
  "  position {\n"
  "    x: 0.0\n"
  "    y: 0.0\n"
  "    z: 0.0\n"
  "  }\n"
  "  rotation {\n"
  "    x: 0.0\n"
  "    y: 0.0\n"
  "    z: 0.0\n"
  "    w: 1.0\n"
  "  }\n"
  "  properties {\n"
  "    id: \"data\"\n"
  "    value: \"{{ .Index }}.0\"\n"
  "    type: PROPERTY_TYPE_NUMBER\n"
  "  }\n"
  "  property_decls {\n"
  "  }\n"
  "}\n"
  "embedded_components {\n"
  "  id: \"collisionobject\"\n"
  "  type: \"collisionobject\"\n"
  "  data: \"collision_shape: \\\"\\\"\\n"
  "type: COLLISION_OBJECT_TYPE_TRIGGER\\n"
  "mass: 0.0\\n"
  "friction: 0.1\\n"
  "restitution: 0.5\\n"
  "group: \\\"trigger\\\"\\n"
  "mask: \\\"player\\\"\\n"
  "embedded_collision_shape {\\n"
  "  shapes {\\n"
  "    shape_type: TYPE_BOX\\n"
  "    position {\\n"
  "      x: 0.0\\n"
  "      y: 0.0\\n"
  "      z: 0.0\\n"
  "    }\\n"
  "    rotation {\\n"
  "      x: 0.0\\n"
  "      y: 0.0\\n"
  "      z: 0.0\\n"
  "      w: 1.0\\n"
  "    }\\n"
  "    index: 0\\n"
  "    count: 3\\n"
  "  }\\n"
  "  data: {{ .W }}.0\\n"
  "  data: {{ .H }}.0\\n"
  "  data: 10.0\\n"
  "}\\n"
  "linear_damping: 0.0\\n"
  "angular_damping: 0.0\\n"
  "locked_rotation: false\\n"
  "bullet: false\\n"
  "\"\n"
  "  position {\n"
  "    x: 0.0\n"
  "    y: 0.0\n"
  "    z: 0.0\n"
  "  }\n"
  "  rotation {\n"
  "    x: 0.0\n"
  "    y: 0.0\n"
  "    z: 0.0\n"
  "    w: 1.0\n"
  "  }\n"
  "}\n"
  ""
  position {
    x: {{ .X }}.0
    y: {{ .Y }}.0
    z: 0.0
  }
  rotation {
    x: 0.0
    y: 0.0
    z: 0.0
    w: 1.0
  }
  scale3 {
    x: 1.0
    y: 1.0
    z: 1.0
  }
}
{{ end }}
`))
