package main

import (
	"fmt"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	running  bool
)

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

	// Define 3 vertices for a 2D triangle
	p0 := Vec2{X: WindowWidth / 2, Y: 100}
	p1 := Vec2{X: 100, Y: 500}
	p2 := Vec2{X: 700, Y: 500}

	// Draw the wireframe triangle
	DrawTriangle(
		int(p0.X), int(p0.Y),
		int(p1.X), int(p1.Y),
		int(p2.X), int(p2.Y),
		0xFF00FF00, // Green
	)
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
		processInput()
		update()
		render()
		// Sleep briefly to maintain roughly 60 FPS
		sdl.Delay(16) 
	}
}
