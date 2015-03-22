package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/nfnt/resize"
	"github.com/runningwild/mipper/mipper"
)

var inputFilename = flag.String("input", "", "Input image.")
var outputDir = flag.String("output", "output", "Output dir.")
var maxTileSize = flag.Int("max-size", 1000, "Max length of a tile's side.")

func subImage(im *image.RGBA, bounds image.Rectangle) *image.RGBA {
	return &image.RGBA{
		Pix:    im.Pix[bounds.Min.Y*im.Stride+bounds.Min.X*4:],
		Stride: im.Stride,
		Rect:   bounds.Sub(bounds.Min),
	}
}

func chop(im *image.RGBA, srcSize int, scale float64, outputDir string) (int, error) {
	var wg sync.WaitGroup
	errs := make(chan error)
	for x := 0; x < im.Bounds().Dx(); x += srcSize {
		for y := 0; y < im.Bounds().Dy(); y += srcSize {
			wg.Add(1)
			go func(x, y int) {
				defer wg.Done()
				width := srcSize
				if width+x > im.Bounds().Dx() {
					width = im.Bounds().Dx() - x
				}
				height := srcSize
				if height+y > im.Bounds().Dy() {
					height = im.Bounds().Dy() - y
				}
				sub := subImage(im, image.Rect(x, y, x+width, y+height))
				tile := resize.Resize(uint(float64(width)*scale), uint(float64(height)*scale), sub, resize.Lanczos3)
				outName := fmt.Sprintf("%d_%d.png", x/srcSize, y/srcSize)
				f, err := os.Create(filepath.Join(outputDir, outName))
				if err != nil {
					errs <- fmt.Errorf("Unable to create file %q: %v", outName, err)
					return
				}
				err = png.Encode(f, tile)
				f.Close()
				if err != nil {
					errs <- fmt.Errorf("Unable to write tile %q: %v", outName, err)
					return
				}
				errs <- nil
			}(x, y)
		}
	}
	go func() {
		wg.Wait()
		close(errs)
	}()
	var firstErr error
	tilesCreated := 0
	for err := range errs {
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
		} else {
			tilesCreated++
		}
	}
	return tilesCreated, firstErr
}

func main() {
	flag.Parse()
	os.MkdirAll(*outputDir, 0777)
	f, err := os.Open(*inputFilename)
	if err != nil {
		log.Fatalf("Unable to open %q: %v", *inputFilename, err)
	}
	im, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		log.Fatalf("Unable to decode %q: %v", *inputFilename, err)
	}
	canvas := image.NewRGBA(im.Bounds())
	draw.Draw(canvas, canvas.Bounds(), im, image.Point{}, draw.Over)
	level := 0
	scale := 1.0
	size := *maxTileSize
	for {
		fmt.Printf("Chopping %d\n", level)
		dir := filepath.Join(*outputDir, fmt.Sprintf("%d", level))
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			log.Fatalf("Unable to mkdirall: %v", err)
		}
		tiles, err := chop(canvas, size, scale, dir)
		if err != nil {
			log.Fatalf("Chop failed: %v", err)
		}
		fmt.Printf("Made %d tiles\n", tiles)
		level++
		scale /= 2
		size *= 2
		if tiles == 1 {
			break
		}
	}

	cfg := mipper.Config{
		Levels:   level,
		TileSize: *maxTileSize,
		Dx:       canvas.Bounds().Dx(),
		Dy:       canvas.Bounds().Dy(),
	}
	data, err := json.Marshal(&cfg)
	if err != nil {
		log.Fatalf("Unable to encode config file: %v", err)
	}
	err = ioutil.WriteFile(filepath.Join(*outputDir, "config.json"), data, 0664)
	if err != nil {
		log.Fatalf("Unable to write config file: %v", err)
	}
}
