package mipper

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

type Config struct {
	// Levels and TileSize don't both need to be specified, since either can be calculated from the
	// other, but it's convenient.

	// Number of levels of mipping.  The highest level will contain exactly one tile.
	Levels int

	// Maximum length of a side of a tile.
	TileSize int

	// Dx and Dy are measured in pixels from the original image.  This is used to calculate the dpi
	// at maximum resolution.
	Dx, Dy int
}

type Tile interface {
	Level() int
	Bounds() (x, y, dx, dy float64)
	Filename() string
}

type tile struct {
	level        int
	x, y, dx, dy float64
	filename     string
}

func (t tile) Level() int {
	return t.level
}
func (t tile) Bounds() (x, y, dx, dy float64) {
	return t.x, t.y, t.dx, t.dy
}
func (t tile) Filename() string {
	return t.filename
}

type MipManager struct {
	config       Config
	loaded       map[string]tile
	root         tile
	maxDpr       int
	maxDpi       int
	maxDx, maxDy float64
}

func Make(configData []byte) (*MipManager, error) {
	var config Config
	err := json.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}
	mm := &MipManager{
		config: config,
		loaded: make(map[string]tile),
	}
	if mm.config.Dx > mm.config.Dy {
		mm.maxDpr = mm.config.Dx
		mm.maxDpi = mm.config.Dx * mm.config.Dx
		mm.maxDx = 1
		mm.maxDy = float64(mm.config.Dy) / float64(mm.config.Dx)
	} else {
		mm.maxDpr = mm.config.Dy
		mm.maxDpi = mm.config.Dy * mm.config.Dy
		mm.maxDx = float64(mm.config.Dx) / float64(mm.config.Dy)
		mm.maxDy = 1
	}
	mm.root = mm.makeTile(mm.config.Levels-1, 0, 0)
	return mm, nil
}

func (mm *MipManager) makeTile(level, x, y int) tile {
	tileSize := mm.config.TileSize * (1 << uint(level))
	var t tile
	t.level = level
	t.x = float64(x*tileSize) / float64(mm.maxDpr)
	t.y = float64(y*tileSize) / float64(mm.maxDpr)
	t.dx = float64(tileSize) / float64(mm.maxDpr)
	if t.x+t.dx > mm.maxDx {
		t.dx = mm.maxDx - t.x
	}
	t.dy = float64(tileSize) / float64(mm.maxDpr)
	if t.y+t.dy > mm.maxDy {
		t.dy = mm.maxDy - t.y
	}
	t.filename = filepath.Join(fmt.Sprintf("%d", level), fmt.Sprintf("%d_%d.png", x, y))
	return t
}

func (mm *MipManager) tilesAt(level int, x, y, dx, dy float64) map[string]tile {
	tileSize := mm.config.TileSize * (1 << uint(level))
	x0 := int(x * float64(mm.maxDpr) / float64(tileSize))
	x1 := int(((x+dx)*float64(mm.maxDpr) + float64(tileSize-1)) / float64(tileSize))
	if x0 < 0 {
		x0 = 0
	}
	if x1*tileSize > mm.config.Dx {
		x1 = mm.config.Dx / tileSize
	}
	y0 := int(y * float64(mm.maxDpr) / float64(tileSize))
	y1 := int(((y+dy)*float64(mm.maxDpr) + float64(tileSize-1)) / float64(tileSize))
	if y0 < 0 {
		y0 = 0
	}
	if y1*tileSize > mm.config.Dy {
		y1 = mm.config.Dy / tileSize
	}
	tiles := make(map[string]tile)
	for x := x0; x < x1; x++ {
		for y := y0; y < y1; y++ {
			t := mm.makeTile(level, x, y)
			tiles[t.filename] = t
		}
	}
	return tiles
}

func (mm *MipManager) Adjust(x, y, dx, dy float64, px int) (load, unload []Tile) {
	dpr := float64(px) / dx
	dpi := dpr * dpr
	levelDpi := float64(mm.maxDpi)
	level := 0
	for dpi < levelDpi/4 && level < mm.config.Levels-1 {
		levelDpi /= 4
		level++
	}
	if level > mm.config.Levels {
		level = mm.config.Levels - 1
	}
	cur := mm.tilesAt(level, x, y, dx, dy)

	cur[mm.root.filename] = mm.root
	for filename, t := range cur {
		if _, ok := mm.loaded[filename]; !ok {
			load = append(load, t)
		}
	}
	for filename, t := range mm.loaded {
		if _, ok := cur[filename]; !ok {
			unload = append(unload, t)
		}
	}
	mm.loaded = cur
	return
}

func (mm *MipManager) ListAllTiles() []Tile {
	var tiles []Tile
	for level := mm.config.Levels - 1; level >= 0; level-- {
		tileSize := mm.config.TileSize * (1 << uint(level))
		for x := 0; x*tileSize < mm.config.Dx; x++ {
			for y := 0; y*tileSize < mm.config.Dy; y++ {
				tiles = append(tiles, mm.makeTile(level, x, y))
			}
		}
	}
	return tiles
}

func (mm *MipManager) Config() Config {
	return mm.config
}
