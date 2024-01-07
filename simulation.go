package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"rd/utils"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
)

const width = 150
const height = 150

// create and initialize a new config as current setup
var setup utils.Config = utils.NewConfig(width, height, 1, 0.5, 0.055, 0.062, 1)

func clamp(n float64, min, max float64) float64 {
	// restrict a value between two bounds
	if n < min {
		return min
	} else if n > max {
		return max
	}
	return n
}
func reactionDiffusion(i, j, w, h int) color.Color {
	// update the pixels colors according to the reaction diffusion state matrices
	if i < width && j < height {
		amount := setup.A.At(i, j) - setup.B.At(i, j)
		col := uint8(clamp(amount, 0, 1) * 255)
		return color.RGBA{
			col,
			col,
			col,
			0xff}
	} else {
		return color.Black
	}
}

func randomColor(i, j, w, h int) color.Color {
	// update the pixel colors with random values (used for tests)
	return color.RGBA{
		uint8(rand.Intn(255)),
		uint8(rand.Intn(255)),
		uint8(rand.Intn(255)),
		0xff}
}

func animate(w fyne.Window, raster *canvas.Raster) {
	// update the canvas at a regulat time tick
	for range time.Tick(time.Millisecond * 10) {
		setup.Update()
		w.Canvas().Refresh(raster)
	}
}

func main() {
	// define the window and its properties
	rdApp := app.New()
	w := rdApp.NewWindow("Reaction Diffusion")
	w.SetFixedSize(true) // starts as floating window
	// raster is the pixel matrix and its update function
	raster := canvas.NewRasterWithPixels(reactionDiffusion)
	w.SetContent(raster)
	// define window size
	widthMargin := float32(math.Round(width*0.23) - 7)
	heightMargin := float32(math.Round(height*0.23) - 7)
	fmt.Println(widthMargin, heightMargin)
	//raster.Resize(fyne.NewSize(width, height))
	w.Resize(fyne.NewSize(width-widthMargin, height-heightMargin))
	// launch animation
	go animate(w, raster)
	w.ShowAndRun()
}
