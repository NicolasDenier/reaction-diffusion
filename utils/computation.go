package utils

import (
	"math/rand"
	"time"

	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/mat"
)

// for random generation
var seed = rand.NewSource(time.Now().UnixNano())
var r = rand.New(seed)

type Config struct {
	// matrices
	A, B, Kernel *mat.Dense
	// parameters
	DA, DB, F, K, Dt float64
}

type compute interface {
	InitState()
	Update()
}

func fill(length int, value float64) []float64 {
	// create a new slice filled with the provided value
	grid := make([]float64, length)
	for i := range grid {
		grid[i] = value
	}
	return grid
}

func randInt(min, max int) int {
	return r.Intn(max-min) + min
}

func Clamp(n float64, min, max float64) float64 {
	// restrict a value between two bounds
	if n < min {
		return min
	} else if n > max {
		return max
	}
	return n
}

func (c Config) InitState() {
	// define the initial state of B
	// for now, random rectangles
	h, w := c.B.Dims()
	// random number of rectagles
	for k := 0; k < randInt(1, 10); k++ {
		// random widths
		w1 := randInt(5, 50)
		w2 := randInt(5, 50)
		// center of rectangle position
		x := randInt(w1, w-w1)
		y := randInt(w2, h-w2)
		// fill the rectangle to 1
		for i := x - w1; i < x+w1; i++ {
			for j := y - w2; j < y+w2; j++ {
				c.B.Set(i, j, 1)
			}
		}
	}
}

func NewEmptyConfig(h, w int) Config {
	// create a new config without the numerical variables
	ones := fill(h*w, 1)
	zeros := fill(h*w, 0)
	setup := Config{
		A:      mat.NewDense(h, w, ones),
		B:      mat.NewDense(h, w, zeros),
		Kernel: mat.NewDense(3, 3, []float64{0.05, 0.2, 0.05, 0.2, -1, 0.2, 0.05, 0.2, 0.05}),
	}
	setup.InitState()
	return setup
}

func NewConfig(h, w int, DA, DB, f, k, dt float64) Config {
	// create a new config with all variables initialized
	setup := NewEmptyConfig(h, w)
	setup.DA = DA
	setup.DB = DB
	setup.F = f
	setup.K = k
	setup.Dt = dt
	return setup
}

func padMatrix(m *mat.Dense, padding int) *mat.Dense {
	// add zero-padding around a matrix
	h, w := m.Dims()
	nh := h + 2*padding
	nw := w + 2*padding
	// full of zeros
	padded := mat.NewDense(nh, nw, fill(nh*nw, 0))
	// copy matrix at the center
	for i := 0; i < h; i++ {
		for j := 0; j < h; j++ {
			padded.Set(i+padding, j+padding, m.At(i, j))
		}
	}
	return padded
}

// TODO redefine with matrix multiplication (should be more efficient)
func convolve(m, kernel *mat.Dense) *mat.Dense {
	// perform a convolution between a matrix and a kernel matrix
	// here the kernel is always the same, 3*3 and the padding is always 1
	p := 1
	h, w := m.Dims()
	// n will receive the new computed values and ref is a copy of n before it is altered
	n := padMatrix(m, p)
	ref := mat.DenseCopyOf(n)
	n.Apply(func(i, j int, _ float64) float64 {
		// first flip the kernel, no need here as it is symmetrical
		// do not compute for padded values
		if i < p || i >= h+p || j < p || j >= w+p {
			return 0
		} else {
			// take a submatrix of n, same size as kernel
			subm := mat.DenseCopyOf(ref.Slice(i-1, i+2, j-1, j+2))
			// multiply element wise with kernel
			subm.MulElem(subm, kernel)
			// sum all elements
			sum := floats.Sum(subm.RawMatrix().Data)
			return sum
		}
	},
		n)
	// remove padding
	return mat.DenseCopyOf(n.Slice(p, h+p, p, w+p))
}

// func normalize(m *mat.Dense) {
// 	// normalize a matrix
// 	max := mat.Max(m)
// 	min := mat.Min(m)
// 	m.Apply(func(_, _ int, v float64) float64 {
// 		return (v - min) / (max - min)
// 	}, m)
// }

// func PrintMat(m *mat.Dense) {
// 	fa := mat.Formatted(m, mat.Prefix("    "), mat.Squeeze())
// 	fmt.Printf("m = %v", fa)
// 	fmt.Println()
// }

func (c *Config) Update() {
	// compute the new states of A and B, step by step
	// diffusion
	DAL := convolve(c.A, c.Kernel)
	DAL.Scale(c.DA, DAL)
	DBL := convolve(c.B, c.Kernel)
	DBL.Scale(c.DB, DBL)
	// reaction
	ABB := mat.DenseCopyOf(c.A)
	ABB.MulElem(ABB, c.B)
	ABB.MulElem(ABB, c.B)
	// feed
	feed := mat.DenseCopyOf(c.A)
	feed.Apply(func(_, _ int, v float64) float64 { return c.F * (1 - v) }, feed)
	// kill
	kill := mat.DenseCopyOf(c.B)
	kill.Scale(-(c.K + c.F), kill)
	// compute next A
	changeA := mat.DenseCopyOf(ABB)
	changeA.Scale(-1, changeA)
	changeA.Add(changeA, DAL)
	changeA.Add(changeA, feed)
	changeA.Scale(c.Dt, changeA)
	NextA := mat.DenseCopyOf(c.A)
	NextA.Add(NextA, changeA)
	// compute next B
	changeB := mat.DenseCopyOf(ABB)
	changeB.Add(changeB, DBL)
	changeB.Add(changeB, kill)
	changeB.Scale(c.Dt, changeB)
	NextB := mat.DenseCopyOf(c.B)
	NextB.Add(NextB, changeB)
	// update
	c.A = mat.DenseCopyOf(NextA)
	c.B = mat.DenseCopyOf(NextB)
}
