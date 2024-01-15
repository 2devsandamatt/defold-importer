package main

import (
	"flag"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type importer struct {
	root string

	aseprite asepriteImporter
	ink      inkImporter
}

func (i importer) Import() error {
	var (
		asepriteFiles []string
		inkFiles      []string
	)
	if err := filepath.WalkDir(i.root, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		case ".aseprite":
			asepriteFiles = append(asepriteFiles, path)
		case ".ink":
			inkFiles = append(inkFiles, path)
		default:
			log.Printf("WARNING: skipping unsupported asset %s", path)
		}
		return nil
	}); err != nil {
		return err
	}
	os.MkdirAll(filepath.Join(i.aseprite.outputDir, "img"), os.ModePerm)
	if err := i.aseprite.Import(asepriteFiles); err != nil {
		return err
	}
	if err := i.ink.Import(inkFiles); err != nil {
		return err
	}
	return nil
}

func main() {
	var output string
	flag.StringVar(&output, "output", "import", "Folder to output to")
	flag.Parse()
	importer := importer{root: flag.Arg(0)}
	importer.aseprite = asepriteImporter{output}
	importer.ink = inkImporter{output}
	if err := importer.Import(); err != nil {
		log.Fatal(err)
	}
}
