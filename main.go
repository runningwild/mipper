package main

import (
	"fmt"
	"github.com/runningwild/mipper/mipper"
)

type params struct {
	x, y, dx, dy float64
}

func main() {
	m, err := mipper.Make([]byte(`{"TileSize": 500, "Dx": 12000, "Dy": 5000, "Levels": 6}`))
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	fmt.Printf("All Tiles:\n")
	for _, t := range m.ListAllTiles() {
		x, y, dx, dy := t.Bounds()
		fmt.Printf("%q: %2.3v %2.3v %2.3v %2.3v\n", t.Filename(), x, y, dx, dy)
	}
	const dx = 12000
	const dy = 5000
	const maxDpr = dx
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
		load, unload := m.Adjust(float64(p.x)/maxDpr, float64(p.y)/maxDpr, float64(p.dx)/maxDpr, float64(p.dy)/maxDpr, 100)
		fmt.Printf("View: %v\n", p)
		fmt.Printf("Load:\n")
		for _, t := range load {
			x, y, dx, dy := t.Bounds()
			fmt.Printf("%q %f %f %f %f\n", t.Filename(), x, y, dx, dy)
		}
		fmt.Printf("Unload:\n")
		for _, t := range unload {
			x, y, dx, dy := t.Bounds()
			fmt.Printf("%q %f %f %f %f\n", t.Filename(), x, y, dx, dy)
		}
	}
}
