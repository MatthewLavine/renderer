package main

import (
	"math"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	WindowWidth  = 1280
	WindowHeight = 720
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

// DrawFilledTriangle draws a solid, filled triangle using the flat-top/flat-bottom scanline algorithm
func DrawFilledTriangle(x0, y0, x1, y1, x2, y2 int, color uint32) {
	// 1. Sort the vertices by y-coordinate ascending (y0 <= y1 <= y2)
	if y0 > y1 {
		y0, y1 = y1, y0
		x0, x1 = x1, x0
	}
	if y1 > y2 {
		y1, y2 = y2, y1
		x1, x2 = x2, x1
	}
	if y0 > y1 {
		y0, y1 = y1, y0
		x0, x1 = x1, x0
	}

	if y1 == y2 {
		// Draw flat-bottom triangle
		fillFlatBottomTriangle(x0, y0, x1, y1, x2, y2, color)
	} else if y0 == y1 {
		// Draw flat-top triangle
		fillFlatTopTriangle(x0, y0, x1, y1, x2, y2, color)
	} else {
		// Split the triangle into a flat-bottom and a flat-top triangle
		// Calculate the new vertex (Mx, My)
		My := y1
		Mx := ((x2-x0)*(y1-y0))/(y2-y0) + x0

		fillFlatBottomTriangle(x0, y0, x1, y1, Mx, My, color)
		fillFlatTopTriangle(x1, y1, Mx, My, x2, y2, color)
	}
}

func fillFlatBottomTriangle(x0, y0, x1, y1, x2, y2 int, color uint32) {
	// Calculate the inverse slopes (how much x changes for each 1 unit of y)
	invSlope1 := float64(x1-x0) / float64(y1-y0)
	invSlope2 := float64(x2-x0) / float64(y2-y0)

	xStart := float64(x0)
	xEnd := float64(x0)

	// Loop scanlines from top to bottom
	for y := y0; y <= y2; y++ {
		drawHorizontalLine(int(xStart), int(xEnd), y, color)
		xStart += invSlope1
		xEnd += invSlope2
	}
}

func fillFlatTopTriangle(x0, y0, x1, y1, x2, y2 int, color uint32) {
	invSlope1 := float64(x2-x0) / float64(y2-y0)
	invSlope2 := float64(x2-x1) / float64(y2-y1)

	xStart := float64(x0)
	xEnd := float64(x1)

	for y := y0; y <= y2; y++ {
		drawHorizontalLine(int(xStart), int(xEnd), y, color)
		xStart += invSlope1
		xEnd += invSlope2
	}
}

func drawHorizontalLine(xStart, xEnd, y int, color uint32) {
	if xStart > xEnd {
		xStart, xEnd = xEnd, xStart
	}
	for x := xStart; x <= xEnd; x++ {
		SetPixel(x, y, color)
	}
}
