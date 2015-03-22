package mipper

import (
	"fmt"
	"path/filepath"
)

type Config struct {
	Levels   int
	TileSize int
	Dx, Dy   int
}

type Tile interface {
	Level() int
	Bounds() (x, y, dx, dy int)
	Filename() string
}

type tile struct {
	level        int
	x, y, dx, dy int
	filename     string
}

func (t tile) Level() int {
	return t.level
}
func (t tile) Bounds() (x, y, dx, dy int) {
	return t.x, t.y, t.dx, t.dy
}
func (t tile) Filename() string {
	return t.filename
}

type MipManager struct {
	config Config
	loaded map[string]tile
	root   tile
}

func Make(config Config) *MipManager {
	mm := &MipManager{
		config: config,
		loaded: make(map[string]tile),
	}
	mm.root = mm.makeTile(mm.config.Levels-1, 0, 0)
	return mm
}

func (mm *MipManager) makeTile(level, x, y int) tile {
	tileSize := mm.config.TileSize * (1 << uint(level))
	var t tile
	t.level = level
	t.x = x * tileSize
	t.y = y * tileSize
	t.dx = tileSize
	if t.x+t.dx > mm.config.Dx {
		t.dx = mm.config.Dx - t.x
	}
	t.dy = tileSize
	if t.y+t.dy > mm.config.Dy {
		t.dy = mm.config.Dy - t.y
	}
	t.filename = filepath.Join(fmt.Sprintf("%d", level), fmt.Sprintf("%d_%d.png", x, y))
	return t
}

func (mm *MipManager) tilesAt(x, y, dx, dy int) map[string]tile {
	large := dx
	if dy > large {
		large = dy
	}
	minSize := large
	level := 0
	tileSize := mm.config.TileSize
	for level < mm.config.Levels-1 && tileSize < minSize {
		level++
		tileSize *= 2
	}
	x0 := x / tileSize
	x1 := (x + dx + tileSize - 1) / tileSize
	if x0 < 0 {
		x0 = 0
	}
	if x1*tileSize > mm.config.Dx {
		x1 = mm.config.Dx / tileSize
	}
	y0 := y / tileSize
	y1 := (y + dy + tileSize - 1) / tileSize
	if y0 < 0 {
		y0 = 0
	}
	if y1*tileSize > mm.config.Dy {
		y1 = mm.config.Dy / tileSize
	}
	tiles := make(map[string]tile)
	for x := x0; x <= x1; x++ {
		for y := y0; y <= y1; y++ {
			t := mm.makeTile(level, x, y)
			tiles[t.filename] = t
		}
	}
	return tiles
}

func (mm *MipManager) Adjust(x, y, dx, dy int) (load, unload []Tile) {
	cur := mm.tilesAt(x, y, dx, dy)
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
