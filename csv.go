package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type csvImporter struct {
	outputDir string
}

func (i csvImporter) Import(files []string) error {
	for _, file := range files {
		if err := i.importOne(file); err != nil {
			return err
		}
	}
	return nil
}

func (i csvImporter) importOne(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	r := csv.NewReader(f)
	headers, err := r.Read()
	if err != nil {
		return err
	}
	var code []string
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		var fields []string
		for i, field := range row {
			if i >= len(headers) {
				break
			}
			fields = append(fields, fmt.Sprintf("%s = %q", headers[i], field))
		}
		code = append(code, strings.Join(fields, ","))
	}
	_, name := filepath.Split(file)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	return i.render(name+".lua", csvTemplate, code)
}

func (i csvImporter) render(filename string, tmpl *template.Template, code []string) error {
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, code); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(i.outputDir, filename), buf.Bytes(), os.ModePerm)
}

var csvTemplate = template.Must(template.New("").Parse(`
local M = {}
{{- range . }}
table.insert(M, { {{ . }} })
{{- end }}
return M
`))
