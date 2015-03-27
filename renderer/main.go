package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/nfnt/resize"
	"github.com/runningwild/mipper/mipper"
)

var configPath = flag.String("config", "./config.json", "Config file to load")
var x = flag.Float64("x", 0, "x")
var y = flag.Float64("y", 0, "y")
var dx = flag.Float64("dx", 1, "dx")
var dy = flag.Float64("dy", 1, "dy")
var px = flag.Int("px", 100, "px")
var out = flag.String("out", "out.png", "output filename")

func main() {
	flag.Parse()
	data, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("%v", err)
	}
	m, err := mipper.Make(data)
	if err != nil {
		log.Fatalf("%v", err)
	}
	load, _ := m.Adjust(*x, *y, *dx, *dy, *px)
	for _, t := range load {
		fmt.Printf("Load %v\n", t.Filename())
	}
	dpr := float64(*px) / float64(*dx)
	py := int(float64(*dy)*dpr + 0.5)
	canvas := image.NewRGBA(image.Rect(0, 0, *px, py))
	lowest := 10000000
	for i := range load {
		if load[i].Level() < lowest {
			lowest = load[i].Level()
		}
	}
	for _, t := range load {
		if t.Level() != lowest {
			continue
		}
		f, err := os.Open(filepath.Join(filepath.Dir(*configPath), t.Filename()))
		if err != nil {
			log.Fatalf("%v", err)
		}
		im, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			log.Fatalf("%v", err)
		}
		tx, ty, tdx, tdy := t.Bounds()
		fmt.Printf("Resize to %v %v\n", uint(tdx*dpr+0.5), uint(tdy*dpr+0.5))
		r := resize.Resize(uint(tdx*dpr+0.5), uint(tdy*dpr+0.5), im, resize.Lanczos2)
		px := int((tx-*x)*dpr + 0.5)
		py := int((ty-*y)*dpr + 0.5)
		fmt.Printf("Drawing to %v %v\n", px, py)
		p := image.Point{px, py}
		draw.Draw(canvas, r.Bounds().Add(p), r, image.Point{}, draw.Over)
	}
	f, err := os.Create(*out)
	if err != nil {
		log.Fatalf("%v", err)
	}
	err = png.Encode(f, canvas)
	f.Close()
	if err != nil {
		log.Fatalf("%v", err)
	}
}
