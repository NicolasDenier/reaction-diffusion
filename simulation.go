package main

import (
	"fmt"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"rd/utils"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
)

/*
Reaction-Diffusion system
Gray-Scott model
https://karlsims.com/rd.html

press 's' to save image
*/

const width = 300
const height = 300

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

func saveImage(w fyne.Window) error {
	// capture the current rendered image
	img := w.Canvas().Capture()
	// create the file
	t := time.Now()
	date := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	path := fmt.Sprintf("images/%s.png", date)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	// encode the image to PNG format
	err = png.Encode(file, img)
	if err != nil {
		return err
	}
	return nil

}

func animate(raster *canvas.Raster) {
	// update the canvas at a regulat time tick
	for range time.Tick(time.Millisecond * 10) {
		setup.Update()
		raster.Refresh()
	}
}

func main() {
	// define the window and its properties
	rdApp := app.New()
	w := rdApp.NewWindow("Reaction Diffusion")
	w.SetFixedSize(true) // starts as floating window
	w.SetPadded(false)
	// raster is the pixel matrix and its update function
	raster := canvas.NewRasterWithPixels(reactionDiffusion)
	w.SetContent(raster)
	// define window size
	widthMargin := float32(math.Round(width*0.23) + 1)
	heightMargin := float32(math.Round(height*0.23) + 1)
	//raster.Resize(fyne.NewSize(width, height))
	w.Resize(fyne.NewSize(width-widthMargin, height-heightMargin))
	// launch animation
	go animate(raster)
	// listen for button press
	w.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		if k.Name == "S" {
			fmt.Println("Image saved")
			saveImage(w)
		}
	})
	w.ShowAndRun()
}
