package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/pranavraja/asefile"
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
		uiNodes    []element
		datas      []string
	)
	for _, file := range filenames {
		var aseFile asefile.AsepriteFile
		if err := aseFile.DecodeFile(file); err != nil {
			return fmt.Errorf("failed to decode %s: %s", file, err)
		}
		if aseFile.Header.ColorDepth != 32 {
			return fmt.Errorf("unsupported color depth %d. please convert to RGBA", aseFile.Header.ColorDepth)
		}
		// e.g. assets/ui/start.aseprite becomes dir=assets/ui and name=start
		dir, name := filepath.Split(file)
		name = strings.TrimSuffix(name, filepath.Ext(name))
		if strings.HasSuffix(dir, "ui/") {
			ui, err := a.importUI(name, aseFile)
			if err != nil {
				return err
			}
			uiNodes = append(uiNodes, ui...)
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
	// Also all UI nodes
	if err := a.render("ui.atlas", atlasTemplate, uiNodes); err != nil {
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
		tiles    []element
		datas    []string
	)
	for _, frame := range file.Frames {
		tilesets := make(map[uint32]asefile.AsepriteTilesetChunk2023)
		for _, tileset := range frame.Tilesets {
			tilesets[tileset.TilesetID] = tileset
			if tileset.Name == "" {
				continue
			}
			if err := a.writeTilesetPNG(fmt.Sprintf("img/%s_tiles_%s.png", filename, tileset.Name), tileset); err != nil {
				return nil, err
			}
		}
		for _, cel := range frame.Cels {
			const (
				Image   = 2
				Tilemap = 3
			)
			switch cel.CelType {
			case Image:
				layer := frame.Layers[cel.LayerIndex].LayerName
				var objectName string
				if strings.HasSuffix(layer, ".object") {
					objectName = strings.TrimSuffix(layer, ".object")
				}
				centerX := cel.X + int16(cel.WidthInPix)/2
				centerY := cel.Y + int16(cel.HeightInPix)/2
				// y coordinates are reversed in defold
				y := int16(file.Header.HeightInPixels) - centerY
				if objectName != "" {
					objects = append(objects, element{
						X:    centerX,
						Y:    y,
						Name: objectName,
					})
					continue
				}
				if err := a.writePNG(fmt.Sprintf("img/%s_%s.png", filename, layer), cel); err != nil {
					return nil, err
				}
				objects = append(objects, element{
					Group: filename,
					Name:  layer,
					X:     centerX,
					Y:     y,
					W:     int(cel.WidthInPix),
					H:     int(cel.HeightInPix),
				})
			case Tilemap:
				layer := frame.Layers[cel.LayerIndex].LayerName
				tileset, ok := tilesets[frame.Layers[cel.LayerIndex].TilesetIndex]
				if !ok {
					continue
				}
				var objectName string
				if strings.HasSuffix(layer, ".object") {
					objectName = strings.TrimSuffix(layer, ".object")
				}
				r := bytes.NewReader(cel.Tiles)
				for y := 0; y < int(cel.HeightInTiles); y++ {
					for x := 0; x < int(cel.WidthInTiles); x++ {
						var tileIndex uint32
						if err := binary.Read(r, binary.LittleEndian, &tileIndex); err != nil {
							return nil, fmt.Errorf("failed to read tile data: %s", err)
						}
						if tileIndex > 0 {
							x := cel.X + int16(x)*int16(tileset.TileWidth) + int16(tileset.TileWidth)/2
							// y coordinates are reversed in defold
							y := int16(file.Header.HeightInPixels) - cel.Y - int16(y)*int16(tileset.TileHeight)
							if objectName != "" {
								objects = append(objects, element{
									X:     x,
									Y:     y,
									Name:  objectName,
									Index: len(objects) + 1,
								})
							} else {
								tiles = append(tiles, element{
									X:     x / int16(tileset.TileWidth),
									Y:     y/int16(tileset.TileHeight) - 1,
									Index: int(tileIndex),
								})
							}
						}
					}
				}
			default:
				log.Printf("unsupported cel type %d", cel.CelType)
			}
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
				dataOffset++
				datas = append(datas, slice.Name)
			}
		}
	}
	var level struct {
		Filename string
		Objects  []element
		Triggers []element
		Tiles    []element
	}
	level.Filename = filename
	level.Objects = objects
	level.Triggers = triggers
	level.Tiles = tiles
	if err := a.render(filename+".atlas", atlasTemplate, level.Objects); err != nil {
		return nil, err
	}
	if len(tiles) > 0 {
		if err := a.render(filename+".tilemap", tilemapTemplate, level); err != nil {
			return nil, err
		}
	}
	return datas, a.render(filename+".collection", collectionTemplate, level)
}

func (a asepriteImporter) importUI(filename string, file asefile.AsepriteFile) ([]element, error) {
	var gui struct {
		Textures []string
		Elements []element
	}
	gui.Textures = append(gui.Textures, "ui")
	needsAllTextures := false
	for _, frame := range file.Frames {
		for _, cel := range frame.Cels {
			layer := frame.Layers[cel.LayerIndex].LayerName
			if err := a.writePNG(fmt.Sprintf("img/%s_%s.png", filename, layer), cel); err != nil {
				return nil, err
			}
			gui.Elements = append(gui.Elements, element{
				Group: filename,
				Name:  layer,
				X:     cel.X,
				// y coordinates are reversed in defold
				Y: int16(file.Header.HeightInPixels) - cel.Y,
				W: int(cel.WidthInPix),
				H: int(cel.HeightInPix),
			})
		}
		if len(frame.Slices) > 0 {
			needsAllTextures = true

			for _, slice := range frame.Slices {
				for _, key := range slice.SliceKeysData {
					gui.Elements = append(gui.Elements, element{
						Name: slice.Name,
						X:    int16(key.SliceXOriginCoords),
						// y coordinates are reversed in defold
						Y: int16(file.Header.HeightInPixels) - int16(key.SliceYOriginCoords),
						W: int(key.SliceWidth),
						H: int(key.SliceHeight),
					})
				}
			}
		}
	}
	if needsAllTextures {
		gui.Textures = append(gui.Textures, "all")
	}
	if err := a.render(filename+".gui", guiTemplate, gui); err != nil {
		return nil, err
	}
	return gui.Elements, nil
}

func (a asepriteImporter) writeTilesetPNG(filename string, tileset asefile.AsepriteTilesetChunk2023) error {
	out, err := zlib.NewReader(bytes.NewReader(tileset.CompressedTilesetImg))
	if err != nil {
		return err
	}
	data, err := io.ReadAll(out)
	if err != nil {
		return err
	}
	w, h := int(tileset.TileWidth), int(tileset.NumTiles)*int(tileset.TileHeight)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	offset := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x, offset = x+1, offset+4 {
			col := color.RGBA{data[offset], data[offset+1], data[offset+2], data[offset+3]}
			img.SetRGBA(x, y, col)
		}
	}
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return err
	}
	return a.writeFile(filename, buf)
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
