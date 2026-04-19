package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	running  bool

	// Global rotation variable
	cubeRotation float64

	// The 8 corners of a 3D cube
	cubePoints = [8]Vec3{
		{X: -1, Y: -1, Z: -1}, // 0
		{X: -1, Y: 1, Z: -1},  // 1
		{X: 1, Y: 1, Z: -1},   // 2
		{X: 1, Y: -1, Z: -1},  // 3
		{X: -1, Y: -1, Z: 1},  // 4
		{X: -1, Y: 1, Z: 1},   // 5
		{X: 1, Y: 1, Z: 1},    // 6
		{X: 1, Y: -1, Z: 1},   // 7
	}
)

// Project takes a 3D point and squashes it onto a 2D plane (Perspective)
func Project(point Vec3) Vec2 {
	fovFactor := 640.0 // Controls how strong the perspective distortion is
	return Vec2{
		X: (fovFactor * point.X) / point.Z,
		Y: (fovFactor * point.Y) / point.Z,
	}
}

// Temporary functions to rotate our points until we build full Matrix math
func RotateX(v Vec3, angle float64) Vec3 {
	return Vec3{
		X: v.X,
		Y: v.Y*math.Cos(angle) - v.Z*math.Sin(angle),
		Z: v.Y*math.Sin(angle) + v.Z*math.Cos(angle),
	}
}

func RotateY(v Vec3, angle float64) Vec3 {
	return Vec3{
		X: v.X*math.Cos(angle) + v.Z*math.Sin(angle),
		Y: v.Y,
		Z: -v.X*math.Sin(angle) + v.Z*math.Cos(angle),
	}
}

func initializeWindow() bool {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing SDL: %v\n", err)
		return false
	}

	var err error
	window, err = sdl.CreateWindow(
		"Software Rasterizer",
		sdl.WINDOWPOS_CENTERED,
		sdl.WINDOWPOS_CENTERED,
		WindowWidth,
		WindowHeight,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating SDL Window: %v\n", err)
		return false
	}

	renderer, err = sdl.CreateRenderer(window, -1, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating SDL Renderer: %v\n", err)
		return false
	}

	// Create a texture that we will stream our pixel data into every frame
	texture, err = renderer.CreateTexture(
		sdl.PIXELFORMAT_ARGB8888,
		sdl.TEXTUREACCESS_STREAMING,
		WindowWidth,
		WindowHeight,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating SDL Texture: %v\n", err)
		return false
	}

	return true
}

func setup() {
	// Allocate memory for our custom color buffer
	colorBuffer = make([]uint32, WindowWidth*WindowHeight)
}

func processInput() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.QuitEvent:
			running = false
		case *sdl.KeyboardEvent:
			if e.Keysym.Sym == sdl.K_ESCAPE {
				running = false
			}
		}
	}
}

func update() {
	// Clear the screen to a dark gray/black
	ClearColorBuffer(0xFF111111)

	// Spin the cube
	cubeRotation += 0.02

	var projectedPoints [8]Vec2

	for i, point := range cubePoints {
		// 1. Rotate the point in 3D space
		rotated := RotateX(point, cubeRotation)
		rotated = RotateY(rotated, cubeRotation)

		// 2. Push the point away from the camera into the screen
		rotated.Z += 5.0

		// 3. Project from 3D to 2D
		projected := Project(rotated)

		// 4. Center the projection on the screen
		projected.X += float64(WindowWidth) / 2.0
		projected.Y += float64(WindowHeight) / 2.0

		projectedPoints[i] = projected
	}

	// Draw the 12 edges of the wireframe cube

	// Front face
	for i := 0; i < 4; i++ {
		DrawLine(
			int(projectedPoints[i].X), int(projectedPoints[i].Y),
			int(projectedPoints[(i+1)%4].X), int(projectedPoints[(i+1)%4].Y),
			0xFF00FF00, // Green
		)
	}

	// Back face
	for i := 0; i < 4; i++ {
		DrawLine(
			int(projectedPoints[i+4].X), int(projectedPoints[i+4].Y),
			int(projectedPoints[((i+1)%4)+4].X), int(projectedPoints[((i+1)%4)+4].Y),
			0xFF00FF00,
		)
	}

	// Connecting edges
	for i := 0; i < 4; i++ {
		DrawLine(
			int(projectedPoints[i].X), int(projectedPoints[i].Y),
			int(projectedPoints[i+4].X), int(projectedPoints[i+4].Y),
			0xFF00FF00,
		)
	}
}

func render() {
	RenderColorBuffer(renderer, texture)
}

func destroyWindow() {
	texture.Destroy()
	renderer.Destroy()
	window.Destroy()
	sdl.Quit()
}

func main() {
	running = initializeWindow()
	if !running {
		os.Exit(1)
	}
	defer destroyWindow()

	setup()

	for running {
		frameStart := sdl.GetTicks()

		processInput()

		// Track how long the update logic takes
		updateStart := time.Now()
		update()
		updateDuration := time.Since(updateStart)

		render()

		// Sleep briefly to maintain roughly 60 FPS (16.6ms)
		timeToWait := 16 - (sdl.GetTicks() - frameStart)
		if timeToWait > 0 && timeToWait <= 16 {
			sdl.Delay(timeToWait)
		}

		// Calculate total time taken for the entire frame (including delay)
		frameTime := sdl.GetTicks() - frameStart

		// Calculate frames per second
		fps := 0.0
		if frameTime > 0 {
			fps = 1000.0 / float64(frameTime)
		}

		// Use \r to overwrite the same line in the terminal instead of scrolling
		fmt.Printf("\rUpdate: %-5d µs | FPS: %-5.1f    ", updateDuration.Microseconds(), fps)
	}
	fmt.Println() // Print a final newline when the app exits
}
