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
var zBuffer []float64

// ClearColorBuffer fills the entire screen with a single solid color
func ClearColorBuffer(color uint32) {
	for i := range colorBuffer {
		colorBuffer[i] = color
	}
}

// ClearZBuffer resets the depth buffer to the far plane (1.0)
func ClearZBuffer() {
	for i := range zBuffer {
		zBuffer[i] = 1.0
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

// DrawFilledTriangle draws a solid, filled triangle using Barycentric coordinates.
// It features perfect screen clamping and Z-Buffering.
func DrawFilledTriangle(v0, v1, v2 Vec3, color uint32) {
	// 1. Find the bounding box of the triangle on the screen
	minX := int(math.Floor(math.Min(v0.X, math.Min(v1.X, v2.X))))
	maxX := int(math.Ceil(math.Max(v0.X, math.Max(v1.X, v2.X))))
	minY := int(math.Floor(math.Min(v0.Y, math.Min(v1.Y, v2.Y))))
	maxY := int(math.Ceil(math.Max(v0.Y, math.Max(v1.Y, v2.Y))))

	// 2. Screen Clamping!
	// This entirely prevents the massive framerate drops by guaranteeing we NEVER
	// loop over pixels that are outside the physical screen boundaries.
	if minX < 0 { minX = 0 }
	if maxX >= WindowWidth { maxX = WindowWidth - 1 }
	if minY < 0 { minY = 0 }
	if maxY >= WindowHeight { maxY = WindowHeight - 1 }

	// Calculate the total area of the triangle for barycentric division
	area := edgeCrossProduct(v0, v1, v2)
	
	// If area is 0, it's a degenerate line/point, don't draw
	if area == 0 {
		return
	}

	// 3. Loop over only the pixels inside the clamped bounding box
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			p := Vec3{X: float64(x), Y: float64(y), Z: 0}
			
			// Calculate barycentric weights
			w0 := edgeCrossProduct(v1, v2, p)
			w1 := edgeCrossProduct(v2, v0, p)
			w2 := edgeCrossProduct(v0, v1, p)
			
			// Check if point is inside the triangle
			// Depending on winding order, all weights will be either positive or negative
			isInside := (w0 >= 0 && w1 >= 0 && w2 >= 0) || (w0 <= 0 && w1 <= 0 && w2 <= 0)
			
			if isInside {
				// Normalize weights
				w0 /= area
				w1 /= area
				w2 /= area
				
				// Interpolate Z depth
				z := w0*v0.Z + w1*v1.Z + w2*v2.Z
				
				idx := y*WindowWidth + x
				
				// Z-Buffer Depth Test! Only draw if this pixel is CLOSER than the current one
				if z < zBuffer[idx] {
					zBuffer[idx] = z
					colorBuffer[idx] = color
				}
			}
		}
	}
}

// edgeCrossProduct returns the 2D cross product to determine which side of an edge a point lies
func edgeCrossProduct(a, b, c Vec3) float64 {
	return (b.X - a.X) * (c.Y - a.Y) - (b.Y - a.Y) * (c.X - a.X)
}
