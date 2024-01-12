package main

import (
	"flag"
	"log"
)

func main() {
	var output string
	flag.StringVar(&output, "output", "testdata", "Folder to output to")
	flag.Parse()
	importer := asepriteImporter{output}
	if err := importer.Import(flag.Args()); err != nil {
		log.Fatal(err)
	}
}
