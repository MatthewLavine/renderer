package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	running  bool
)

type RenderMethod int

const (
	RenderWireframe RenderMethod = iota
	RenderSolid
	RenderSolidWireframe
	RenderShaded
	RenderShadedWireframe
)

var renderMethod = RenderShaded
var globalLightDirection = Vec3{X: 0, Y: 0, Z: 1} // Light shining directly into the screen

// ApplyLightIntensity applies a given percentage of light (0.0 to 1.0) to a base color
func ApplyLightIntensity(baseColor uint32, intensity float64) uint32 {
	a := (baseColor >> 24) & 0xFF
	r := float64((baseColor >> 16) & 0xFF)
	g := float64((baseColor >> 8) & 0xFF)
	b := float64(baseColor & 0xFF)

	// Apply intensity
	r *= intensity
	g *= intensity
	b *= intensity

	return (a << 24) | (uint32(r) << 16) | (uint32(g) << 8) | uint32(b)
}

// Project takes a 3D point and squashes it onto a 2D plane (Perspective)
func Project(point Vec3) Vec2 {
	fovFactor := 640.0 // Controls how strong the perspective distortion is
	return Vec2{
		X: (fovFactor * point.X) / point.Z,
		Y: -(fovFactor * point.Y) / point.Z, // Negate Y so +Y goes UP the screen
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
	err := LoadOBJ("assets/teapot.obj")
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
			if e.Type == sdl.KEYDOWN {
				switch e.Keysym.Sym {
				case sdl.K_ESCAPE:
					running = false
				case sdl.K_1:
					renderMethod = RenderWireframe
				case sdl.K_2:
					renderMethod = RenderSolid
				case sdl.K_3:
					renderMethod = RenderSolidWireframe
				case sdl.K_4:
					renderMethod = RenderShaded
				case sdl.K_5:
					renderMethod = RenderShadedWireframe
				}
			}
		}
	}
}

// TriangleToRender holds projected 2D coordinates and depth information for Painter's algorithm
type TriangleToRender struct {
	Points   [3]Vec2
	Color    uint32
	AvgDepth float64
}

func update() {
	// Clear the screen to a dark gray/black
	ClearColorBuffer(0xFF111111)

	// Spin the mesh globally around the Y axis
	currentMesh.Rotation.Y += 0.01

	var trianglesToRender []TriangleToRender
	cameraPosition := Vec3{X: 0, Y: 0, Z: 0}

	// Loop over all faces in the mesh
	for fIndex, face := range currentMesh.Faces {
		// Get the 3 vertices for this face
		vertices := [3]Vec3{
			currentMesh.Vertices[face.A],
			currentMesh.Vertices[face.B],
			currentMesh.Vertices[face.C],
		}

		var transformedVertices [3]Vec3

		// Transform each vertex
		for i, vertex := range vertices {
			// 1. Scale
			transformed := vertex.Mult(currentMesh.Scale.X)

			// 2. Rotate
			transformed = transformed.RotateX(currentMesh.Rotation.X)
			transformed = transformed.RotateY(currentMesh.Rotation.Y)
			transformed = transformed.RotateZ(currentMesh.Rotation.Z)

			// 3. Translate (Push it away from camera)
			transformed = transformed.Add(currentMesh.Translation)

			transformedVertices[i] = transformed
		}

		// --- Back-Face Culling ---
		// 1. Calculate the face normal
		edge1 := transformedVertices[1].Sub(transformedVertices[0])
		edge2 := transformedVertices[2].Sub(transformedVertices[0])
		normal := edge1.Cross(edge2).Normalize()

		// 2. Calculate the camera ray (vector from camera to the face)
		cameraRay := transformedVertices[0].Sub(cameraPosition)

		// 3. Calculate dot product
		dot := normal.Dot(cameraRay)

		// If dot > 0, the face is pointing away from the camera
		// We bypass culling if we are in pure Wireframe mode so we can see through the model
		if renderMethod != RenderWireframe {
			if dot > 0 {
				continue
			}
		}

		var projectedPoints [3]Vec2

		// Project each transformed vertex
		for i, transformed := range transformedVertices {
			// 4. Project from 3D to 2D
			projected := Project(transformed)

			// 5. Center the projection on the screen
			projected.X += float64(WindowWidth) / 2.0
			projected.Y += float64(WindowHeight) / 2.0

			projectedPoints[i] = projected
		}

		// Calculate average depth for Painter's Algorithm
		avgDepth := (transformedVertices[0].Z + transformedVertices[1].Z + transformedVertices[2].Z) / 3.0

		// Calculate light intensity based on how directly the face points at the light
		// We negate the dot product because light comes towards +Z and normal points towards -Z
		lightIntensity := -normal.Dot(globalLightDirection)
		if lightIntensity < 0.15 {
			lightIntensity = 0.15 // Ambient light baseline so shadows aren't pitch black
		}

		var color uint32
		if renderMethod == RenderShaded || renderMethod == RenderShadedWireframe {
			baseColor := uint32(0xFFFFFFFF) // White base color
			color = ApplyLightIntensity(baseColor, lightIntensity)
		} else {
			// Generate a distinct color based on the face index (Flat random colors)
			color = uint32(0xFF000000) | (uint32((fIndex*30)%255) << 16) | (uint32((fIndex*40)%255) << 8) | uint32((fIndex*50)%255)
		}

		trianglesToRender = append(trianglesToRender, TriangleToRender{
			Points:   projectedPoints,
			Color:    color,
			AvgDepth: avgDepth,
		})
	}

	// --- Painter's Algorithm ---
	// Sort the triangles by average depth (furthest away is drawn first)
	sort.Slice(trianglesToRender, func(i, j int) bool {
		return trianglesToRender[i].AvgDepth > trianglesToRender[j].AvgDepth
	})

	// Render all sorted triangles
	for _, t := range trianglesToRender {
		if renderMethod == RenderSolid || renderMethod == RenderSolidWireframe || renderMethod == RenderShaded || renderMethod == RenderShadedWireframe {
			// Draw the solid, filled triangle
			DrawFilledTriangle(
				int(t.Points[0].X), int(t.Points[0].Y),
				int(t.Points[1].X), int(t.Points[1].Y),
				int(t.Points[2].X), int(t.Points[2].Y),
				t.Color,
			)
		}

		if renderMethod == RenderWireframe || renderMethod == RenderSolidWireframe || renderMethod == RenderShadedWireframe {
			wireColor := uint32(0xFF111111) // Dark gray default
			if renderMethod == RenderWireframe {
				wireColor = 0xFFFFFFFF // White for visibility on dark background
			}

			// Draw the wireframe outline
			DrawTriangle(
				int(t.Points[0].X), int(t.Points[0].Y),
				int(t.Points[1].X), int(t.Points[1].Y),
				int(t.Points[2].X), int(t.Points[2].Y),
				wireColor,
			)
		}
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

	var lastFPS float64
	targetFrameTime := 1.0 / 60.0 // Exactly 60 FPS (approx 0.0166667 seconds)
	perfFreq := float64(sdl.GetPerformanceFrequency())

	for running {
		frameStart := sdl.GetPerformanceCounter()

		processInput()

		// Track how long the update logic takes
		updateStart := time.Now()
		update()
		updateDuration := time.Since(updateStart)

		// Draw Statistics Overlay
		var modeStr string
		switch renderMethod {
		case RenderWireframe:
			modeStr = "Wireframe"
		case RenderSolid:
			modeStr = "Solid"
		case RenderSolidWireframe:
			modeStr = "Solid + Wireframe"
		case RenderShaded:
			modeStr = "Shaded"
		case RenderShadedWireframe:
			modeStr = "Shaded + Wireframe"
		}
		stats := fmt.Sprintf(
			"Update Time: %.2f ms\nFPS: %.1f\nVertices: %d\nFaces: %d\nMode: %s (Press 1-5)",
			updateDuration.Seconds()*1000.0, lastFPS, len(currentMesh.Vertices), len(currentMesh.Faces), modeStr,
		)
		DrawText(10, 10, stats, 0xFFFFFFFF) // Draw white text

		render()

		// Precise Frame Pacing
		elapsed := float64(sdl.GetPerformanceCounter()-frameStart) / perfFreq
		if elapsed < targetFrameTime {
			timeToWait := targetFrameTime - elapsed
			// Ask the OS to sleep for the bulk of the time to save CPU, but leave ~2ms buffer
			if timeToWait > 0.002 {
				sdl.Delay(uint32((timeToWait - 0.002) * 1000.0))
			}
			// Busy-wait the remaining fraction of a millisecond for perfect precision
			for float64(sdl.GetPerformanceCounter()-frameStart)/perfFreq < targetFrameTime {
				// Spin
			}
		}

		// Calculate total frame time and FPS
		frameTime := float64(sdl.GetPerformanceCounter()-frameStart) / perfFreq
		if frameTime > 0 {
			lastFPS = 1.0 / frameTime
		}
	}
}
