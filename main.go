package main

import (
	"fmt"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	running  bool
)

// Project takes a 3D point and squashes it onto a 2D plane (Perspective)
func Project(point Vec3) Vec2 {
	fovFactor := 640.0 // Controls how strong the perspective distortion is
	return Vec2{
		X: (fovFactor * point.X) / point.Z,
		Y: (fovFactor * point.Y) / point.Z,
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

	// Load our 3D mesh from disk
	err := LoadOBJ("assets/cube.obj")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading OBJ file: %v\n", err)
	}
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

	// Spin the mesh globally
	currentMesh.Rotation.X += 0.01
	currentMesh.Rotation.Y += 0.01
	currentMesh.Rotation.Z += 0.01

	// Loop over all faces in the mesh
	for _, face := range currentMesh.Faces {
		// Get the 3 vertices for this face
		vertices := [3]Vec3{
			currentMesh.Vertices[face.A],
			currentMesh.Vertices[face.B],
			currentMesh.Vertices[face.C],
		}

		var projectedPoints [3]Vec2

		// Transform and project each vertex
		for i, vertex := range vertices {
			// 1. Scale
			transformed := vertex.Mult(currentMesh.Scale.X) // Assuming uniform scale

			// 2. Rotate
			transformed = transformed.RotateX(currentMesh.Rotation.X)
			transformed = transformed.RotateY(currentMesh.Rotation.Y)
			transformed = transformed.RotateZ(currentMesh.Rotation.Z)

			// 3. Translate (Push it away from camera)
			transformed = transformed.Add(currentMesh.Translation)

			// 4. Project from 3D to 2D
			projected := Project(transformed)

			// 5. Center the projection on the screen
			projected.X += float64(WindowWidth) / 2.0
			projected.Y += float64(WindowHeight) / 2.0

			projectedPoints[i] = projected
		}

		// Draw the wireframe triangle for this face
		DrawTriangle(
			int(projectedPoints[0].X), int(projectedPoints[0].Y),
			int(projectedPoints[1].X), int(projectedPoints[1].Y),
			int(projectedPoints[2].X), int(projectedPoints[2].Y),
			face.Color,
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
