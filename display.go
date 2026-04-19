package main

import (
	"math"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	WindowWidth  = 800
	WindowHeight = 600
)

var colorBuffer []uint32

// ClearColorBuffer fills the entire screen with a single solid color
func ClearColorBuffer(color uint32) {
	for i := range colorBuffer {
		colorBuffer[i] = color
	}
}

// SetPixel sets the color of a specific pixel on the screen
func SetPixel(x, y int, color uint32) {
	if x >= 0 && x < WindowWidth && y >= 0 && y < WindowHeight {
		colorBuffer[(WindowWidth*y)+x] = color
	}
}

// RenderColorBuffer pushes our custom pixel array to the SDL Texture and presents it
func RenderColorBuffer(renderer *sdl.Renderer, texture *sdl.Texture) {
	// Pitch is the number of bytes in a row of pixel data (Width * 4 bytes per pixel)
	pitch := WindowWidth * 4

	texture.Update(nil, unsafe.Pointer(&colorBuffer[0]), pitch)
	renderer.Copy(texture, nil, nil)
	renderer.Present()
}

// DrawLine draws a line between two points using the DDA algorithm
func DrawLine(x0, y0, x1, y1 int, color uint32) {
	deltaX := float64(x1 - x0)
	deltaY := float64(y1 - y0)

	// Find the longest side to determine how many pixels we need to draw
	var sideLength float64
	if math.Abs(deltaX) >= math.Abs(deltaY) {
		sideLength = math.Abs(deltaX)
	} else {
		sideLength = math.Abs(deltaY)
	}

	// Calculate how much we should increment x and y each step
	xInc := deltaX / sideLength
	yInc := deltaY / sideLength

	currentX := float64(x0)
	currentY := float64(y0)

	for i := 0; i <= int(sideLength); i++ {
		SetPixel(int(math.Round(currentX)), int(math.Round(currentY)), color)
		currentX += xInc
		currentY += yInc
	}
}

// DrawTriangle draws a wireframe triangle by connecting three points with lines
func DrawTriangle(x0, y0, x1, y1, x2, y2 int, color uint32) {
	DrawLine(x0, y0, x1, y1, color)
	DrawLine(x1, y1, x2, y2, color)
	DrawLine(x2, y2, x0, y0, color)
}
