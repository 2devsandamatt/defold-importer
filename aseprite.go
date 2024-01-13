package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"path/filepath"
	"strings"

	"github.com/Racinettee/asefile"
)

type element struct {
	Index int
	Group string
	Name  string
	X     int16
	Y     int16
	W     int
	H     int
}

type animation struct {
	ID           string
	Frames       []string
	PlaybackMode string
	FPS          uint16
}

type asepriteImporter struct {
	outputDir string
}

func (a asepriteImporter) Import(filenames []string) error {
	var (
		animations []animation
		datas      []string
	)
	for _, file := range filenames {
		var aseFile asefile.AsepriteFile
		if err := aseFile.DecodeFile(file); err != nil {
			return fmt.Errorf("failed to decode %s: %s", file, err)
		}
		// e.g. assets/ui/start.aseprite becomes dir=assets/ui and name=start
		dir, name := filepath.Split(file)
		name = strings.TrimSuffix(name, filepath.Ext(name))
		if strings.HasSuffix(dir, "ui/") {
			if err := a.importUI(name, aseFile); err != nil {
				return err
			}
		} else if strings.HasSuffix(dir, "sprites/") {
			anim, err := a.importSprite(name, aseFile)
			if err != nil {
				return err
			}
			animations = append(animations, anim...)
		} else if strings.HasSuffix(dir, "levels/") {
			data, err := a.importLevel(name, len(datas), aseFile)
			if err != nil {
				return err
			}
			datas = append(datas, data...)
		} else {
			log.Printf("no support for importing %s yet", file)
		}
	}
	if err := a.render("data.lua", dataTemplate, datas); err != nil {
		return err
	}
	if err := a.writeFile("data.script", bytes.NewBufferString(`go.property("data", 1)`)); err != nil {
		return err
	}
	// Combine all game animations into single atlas for performance
	if err := a.render("all.atlas", animationsTemplate, animations); err != nil {
		return err
	}
	return nil
}

func (a asepriteImporter) importSprite(filename string, file asefile.AsepriteFile) ([]animation, error) {
	var anims []animation
	for i, frame := range file.Frames {
		if err := a.writePNG(fmt.Sprintf("img/%s_%d.png", filename, i), frame.Cels...); err != nil {
			return anims, err
		}
		for _, tag := range frame.Tags.Tags {
			anim := animation{
				ID:  fmt.Sprintf("%s_%s", filename, tag.TagName),
				FPS: 12,
			}
			anim.PlaybackMode = "PLAYBACK_LOOP_FORWARD"
			if tag.Repeats == 1 {
				anim.PlaybackMode = "PLAYBACK_ONCE_FORWARD"
			}
			var duration uint16
			for i := tag.FromFrame; i <= tag.ToFrame; i++ {
				frameDuration := file.Frames[i].FrameDurationMilliseconds
				if duration == 0 {
					duration = frameDuration
				} else if duration != frameDuration {
					log.Printf("WARNING: frame duration inconsistency for animation %s: wanted %d, got %d", anim.ID, duration, frameDuration)
				}
				anim.Frames = append(anim.Frames, fmt.Sprintf("%s_%d.png", filename, i))
			}
			if duration == 0 {
				return anims, fmt.Errorf("unexpected zero animation duration for %s", anim.ID)
			}
			anim.FPS = 1000 / duration
			anims = append(anims, anim)
		}
	}
	if err := a.render(filename+".sprite", spriteTemplate, filename); err != nil {
		return anims, err
	}
	return anims, nil
}

func (a asepriteImporter) importLevel(filename string, dataOffset int, file asefile.AsepriteFile) ([]string, error) {
	var (
		objects  []element
		triggers []element
	)
	for _, frame := range file.Frames {
		for _, cel := range frame.Cels {
			layer := frame.Layers[cel.LayerIndex].LayerName
			if err := a.writePNG(fmt.Sprintf("img/%s_%s.png", filename, layer), cel); err != nil {
				return nil, err
			}
			centerX := cel.X + int16(cel.WidthInPix)/2
			centerY := cel.Y + int16(cel.HeightInPix)/2
			objects = append(objects, element{
				Group: filename,
				Name:  layer,
				X:     centerX,
				// y coordinates are reversed in defold
				Y: int16(file.Header.HeightInPixels) - centerY,
				W: int(cel.WidthInPix),
				H: int(cel.HeightInPix),
			})
		}
		for _, slice := range frame.Slices {
			for _, data := range slice.SliceKeysData {
				centerX := int16(data.SliceXOriginCoords) + int16(data.SliceWidth)/2
				centerY := int16(data.SliceYOriginCoords) + int16(data.SliceHeight)/2
				triggers = append(triggers, element{
					Index: dataOffset + 1,
					Group: filename,
					Name:  slice.Name,
					X:     centerX,
					// y coordinates are reversed in defold
					Y: int16(file.Header.HeightInPixels) - centerY,
					// For some reason box2d shapes end up twice the width you specify
					W: int(data.SliceWidth / 2),
					H: int(data.SliceHeight / 2),
				})
			}
		}
	}
	var level struct {
		Filename string
		Objects  []element
		Triggers []element
	}
	level.Filename = filename
	level.Objects = objects
	level.Triggers = triggers
	if err := a.render(filename+".atlas", atlasTemplate, level.Objects); err != nil {
		return nil, err
	}
	var data []string
	for _, trigger := range triggers {
		data = append(data, fmt.Sprintf(`{ name = %q }`, trigger.Name))
	}
	return data, a.render(filename+".collection", collectionTemplate, level)
}

func (a asepriteImporter) importUI(filename string, file asefile.AsepriteFile) error {
	var gui []element
	for _, frame := range file.Frames {
		for _, cel := range frame.Cels {
			layer := frame.Layers[cel.LayerIndex].LayerName
			if err := a.writePNG(fmt.Sprintf("img/%s_%s.png", filename, layer), cel); err != nil {
				return err
			}
			gui = append(gui, element{
				Group: filename,
				Name:  layer,
				X:     cel.X,
				// y coordinates are reversed in defold
				Y: int16(file.Header.HeightInPixels) - cel.Y,
				W: int(cel.WidthInPix),
				H: int(cel.HeightInPix),
			})
		}
	}
	if err := a.render(filename+".atlas", atlasTemplate, gui); err != nil {
		return err
	}
	return a.render(filename+".gui", guiTemplate, gui)
}

// TODO: this function seems like it would support multiple cels and flatten the layers,
// however they currently all would have to be the same x, y, width and height,
// which is kinda useless. Will improve later
func (a asepriteImporter) writePNG(filename string, cels ...asefile.AsepriteCelChunk2005) error {
	w, h := 1, 1
	if len(cels) > 0 {
		first := cels[0]
		w, h = int(first.WidthInPix), int(first.HeightInPix)
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for _, cel := range cels {
		offset := 0
		data := cel.RawCelData
		for y := 0; y < h; y++ {
			for x := 0; x < w; x, offset = x+1, offset+4 {
				col := color.RGBA{data[offset], data[offset+1], data[offset+2], data[offset+3]}
				img.SetRGBA(x, y, col)
			}
		}
	}
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return err
	}
	return a.writeFile(filename, buf)
}
