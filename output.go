package main

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"
)

func (a asepriteImporter) writeFile(filename string, buf *bytes.Buffer) error {
	return os.WriteFile(filepath.Join(a.outputDir, filename), buf.Bytes(), os.ModePerm)
}

func (a asepriteImporter) render(filename string, tmpl *template.Template, data any) error {
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		return err
	}
	return a.writeFile(filename, buf)
}
