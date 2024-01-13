package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type inkImporter struct {
	outputDir string
}

func (i inkImporter) Import(files []string) error {
	for _, file := range files {
		contents, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		_, name := filepath.Split(file)
		name = strings.TrimSuffix(name, filepath.Ext(name))
		outputFile := filepath.Join(i.outputDir, name+".lua")
		out := fmt.Sprintf(inkTemplate, string(contents))
		if err := os.WriteFile(outputFile, []byte(out), os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

var inkTemplate = `
local narrator = require('narrator.narrator')
local book = narrator.parse_content([[
%s
]])
local story = narrator.init_story(book)
return story
`
