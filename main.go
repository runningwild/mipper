package main

import (
	"fmt"
	"github.com/runningwild/mipper/mipper"
)

type params struct {
	x, y, dx, dy int
}

func main() {
	m := mipper.Make(mipper.Config{TileSize: 500, Dx: 12000, Dy: 5000, Levels: 6})
	ps := []params{
		{0, 0, 12000, 5000},
		{0, 0, 2001, 2001},
		{0, 0, 2000, 2000},
		{0, 0, 1500, 1500},
		{0, 0, 1000, 1000},
		{0, 0, 750, 750},
		{0, 0, 500, 500},
		{5000, 1200, 7500, 4000},
	}
	for _, p := range ps {
		load, unload := m.Adjust(p.x, p.y, p.dx, p.dy)
		fmt.Printf("View: %v\n", p)
		fmt.Printf("Load:\n")
		for _, t := range load {
			x, y, dx, dy := t.Bounds()
			fmt.Printf("%q %d %d %d %d\n", t.Filename(), x, y, dx, dy)
		}
		fmt.Printf("Unload:\n")
		for _, t := range unload {
			x, y, dx, dy := t.Bounds()
			fmt.Printf("%q %d %d %d %d\n", t.Filename(), x, y, dx, dy)
		}
	}
}
