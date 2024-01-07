package main

import (
	"fmt"
	"image"
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
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

/*
Reaction-Diffusion system
Gray-Scott model
https://karlsims.com/rd.html

press 's' to save image
*/

const width = 300
const height = 300

// define system parameters
var DA utils.Parameter // 1
var DB utils.Parameter // 0.5
var f utils.Parameter  // 0.055
var k utils.Parameter  // 0.062

// create and initialize a new config as current setup
var setup utils.Config = utils.NewEmptyConfig(width, height)

func initializeParameters(DA_value, DB_value, f_value, k_value, dt_value float64) {
	// assign each parameter to a setup variable and set the initial values
	DA.Initialize(DA_value, &setup.DA)
	DB.Initialize(DB_value, &setup.DB)
	f.Initialize(f_value, &setup.F)
	k.Initialize(k_value, &setup.K)
	setup.Dt = dt_value // no slider for this one
}

func reactionDiffusion(i, j, w, h int) color.Color {
	// update the pixels colors according to the reaction diffusion state matrices
	if i < width && j < height {
		amount := setup.A.At(i, j) - setup.B.At(i, j)
		col := uint8(utils.Clamp(amount, 0, 1) * 255)
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

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func cropImage(img image.Image) image.Image {
	// crop an image to keep only the left half (the raster)
	cropSize := image.Rect(0, 0, width, height)
	new := image.NewRGBA(cropSize)
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			r, g, b, a := img.At(i, j).RGBA()
			new.SetRGBA(i, j, color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)})
		}
	}
	return new
}

func saveImage(w fyne.Window) error {
	// capture the current rendered image
	img := w.Canvas().Capture()
	img = cropImage(img)
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
	initializeParameters(1, 0.5, 0.055, 0.062, 1)
	// define the window and its properties
	rdApp := app.New()
	w := rdApp.NewWindow("Reaction Diffusion")
	w.SetFixedSize(true) // starts as floating window
	w.SetPadded(false)
	// raster is the pixel matrix and its update function
	raster := canvas.NewRasterWithPixels(reactionDiffusion)
	controls := container.New(layout.NewVBoxLayout(),
		DA.GetSliderBox(0, 1, "DA"),
		DB.GetSliderBox(0, 1, "DB"),
		f.GetSliderBox(0.002, 0.12, "f"),
		k.GetSliderBox(0.01, 0.07, "k"))
	grid := container.New(layout.NewGridLayout(2), raster, controls)
	w.SetContent(grid)
	// define window size
	widthMargin := float32(math.Round(width*0.23) + 1)
	heightMargin := float32(math.Round(height*0.23) + 1)
	//w.Resize(fyne.NewSize(width-widthMargin, height-heightMargin))
	w.Resize(fyne.NewSize(2*(width-widthMargin), height-heightMargin))
	raster.Resize(fyne.NewSize(width, height))
	// launch animation
	go animate(raster)

	// listen for key press
	w.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		switch k.Name {
		// screenshot
		case "S":
			fmt.Println("Image saved")
			saveImage(w)
		// close
		case "C":
			w.Close()
		}
	})
	w.ShowAndRun()
}
