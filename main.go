package main

import (
	"fmt"
	"math"
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

	scene        []*Entity
	renderMethod = RenderShaded
)

type RenderMethod int

const (
	RenderWireframe RenderMethod = iota
	RenderSolid
	RenderSolidWireframe
	RenderShaded
	RenderShadedWireframe
)

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

// Project function removed, we now use Mat4Projection from matrix.go 

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
	teapotMesh, err := LoadOBJ("assets/teapot.obj")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading OBJ file: %v\n", err)
	}

	// Initialize the camera slightly backed off so we can see the scene
	globalCamera = Camera{
		Position: Vec3{X: 0, Y: 0, Z: -5},
		Yaw:      0,
		Pitch:    0,
	}

	// Create 9 teapots in a 3x3 grid
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			entity := &Entity{
				Mesh:        teapotMesh,
				Rotation:    Vec3{X: 0, Y: 0, Z: 0},
				Scale:       Vec3{X: 1, Y: 1, Z: 1},
				Translation: Vec3{X: float64(x) * 5.0, Y: (float64(y) * 5.0) - 1.5, Z: 15},
			}
			scene = append(scene, entity)
		}
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

	keys := sdl.GetKeyboardState()
	
	moveSpeed := 0.2
	turnSpeed := 0.02

	// Calculate forward and right vectors based on camera's Yaw (Y rotation)
	forward := Vec3{X: math.Sin(globalCamera.Yaw), Y: 0, Z: math.Cos(globalCamera.Yaw)}
	right := Vec3{X: math.Cos(globalCamera.Yaw), Y: 0, Z: -math.Sin(globalCamera.Yaw)}

	// WASD Movement
	if keys[sdl.SCANCODE_W] == 1 {
		globalCamera.Position = globalCamera.Position.Add(forward.Mult(moveSpeed))
	}
	if keys[sdl.SCANCODE_S] == 1 {
		globalCamera.Position = globalCamera.Position.Sub(forward.Mult(moveSpeed))
	}
	if keys[sdl.SCANCODE_A] == 1 {
		globalCamera.Position = globalCamera.Position.Sub(right.Mult(moveSpeed))
	}
	if keys[sdl.SCANCODE_D] == 1 {
		globalCamera.Position = globalCamera.Position.Add(right.Mult(moveSpeed))
	}
	
	// Vertical Movement (Q/E)
	if keys[sdl.SCANCODE_Q] == 1 {
		globalCamera.Position.Y += moveSpeed
	}
	if keys[sdl.SCANCODE_E] == 1 {
		globalCamera.Position.Y -= moveSpeed
	}

	// Camera Rotation (Arrow Keys)
	if keys[sdl.SCANCODE_UP] == 1 {
		globalCamera.Pitch += turnSpeed
	}
	if keys[sdl.SCANCODE_DOWN] == 1 {
		globalCamera.Pitch -= turnSpeed
	}
	if keys[sdl.SCANCODE_LEFT] == 1 {
		globalCamera.Yaw -= turnSpeed
	}
	if keys[sdl.SCANCODE_RIGHT] == 1 {
		globalCamera.Yaw += turnSpeed
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

	var trianglesToRender []TriangleToRender

	// 1. Calculate View Matrix based on the Camera's position and rotation
	viewMatrix := Mat4View(globalCamera.Pitch, globalCamera.Yaw, globalCamera.Position)

	// 2. Calculate Projection Matrix
	aspectRatio := float64(WindowWidth) / float64(WindowHeight)
	projMatrix := Mat4Projection(60.0, aspectRatio, 0.1, 100.0)

	// Loop over all entities in the scene
	for _, entity := range scene {
		// Spin the entity around its Y axis
		entity.Rotation.Y += 0.01

		// 1. Pre-calculate the World Matrix for this entity
		scaleMatrix := Mat4Scale(entity.Scale.X, entity.Scale.Y, entity.Scale.Z)
		rotationXMatrix := Mat4RotateX(entity.Rotation.X)
		rotationYMatrix := Mat4RotateY(entity.Rotation.Y)
		rotationZMatrix := Mat4RotateZ(entity.Rotation.Z)
		translationMatrix := Mat4Translate(entity.Translation.X, entity.Translation.Y, entity.Translation.Z)

		// Combine them: World = Translate * RotZ * RotY * RotX * Scale
		worldMatrix := Mat4Identity()
		worldMatrix = Mat4MulMat4(scaleMatrix, worldMatrix)
		worldMatrix = Mat4MulMat4(rotationXMatrix, worldMatrix)
		worldMatrix = Mat4MulMat4(rotationYMatrix, worldMatrix)
		worldMatrix = Mat4MulMat4(rotationZMatrix, worldMatrix)
		worldMatrix = Mat4MulMat4(translationMatrix, worldMatrix)

		// Loop over all faces in the entity's mesh
		for fIndex, face := range entity.Mesh.Faces {
			// Get the 3 vertices for this face
			vertices := [3]Vec3{
				entity.Mesh.Vertices[face.A],
				entity.Mesh.Vertices[face.B],
				entity.Mesh.Vertices[face.C],
			}

			var transformedVertices [3]Vec3

			// Transform each vertex
			for i, vertex := range vertices {
				// a. Transform into World Space
				transformed := Mat4MulVec3(worldMatrix, vertex)
				// b. Transform into View (Camera) Space
				viewed := Mat4MulVec3(viewMatrix, transformed)
				transformedVertices[i] = viewed
			}

			// --- Back-Face Culling ---
			// 1. Calculate the face normal (in View Space)
			edge1 := transformedVertices[1].Sub(transformedVertices[0])
			edge2 := transformedVertices[2].Sub(transformedVertices[0])
			normal := edge1.Cross(edge2).Normalize()

			// 2. Calculate the camera ray
			// Since we are in View Space, the camera is ALWAYS exactly at (0,0,0)!
			cameraRay := transformedVertices[0].Sub(Vec3{X: 0, Y: 0, Z: 0})

			// 3. Calculate dot product
			dot := normal.Dot(cameraRay)

			// If dot > 0, the face is pointing away from the camera
			// We bypass culling if we are in pure Wireframe mode so we can see through the model
			if renderMethod != RenderWireframe {
				if dot > 0 {
					continue
				}
			}

			// --- Near-Plane Culling ---
			// If any vertex is behind the camera (Z < 0.1), discard the entire face to prevent
			// Divide by Zero crashes and visual artifacts. (A perfect engine would clip the triangle instead)
			if transformedVertices[0].Z < 0.1 || transformedVertices[1].Z < 0.1 || transformedVertices[2].Z < 0.1 {
				continue
			}

			var projectedPoints [3]Vec2

			// Project each transformed vertex
			for i, transformed := range transformedVertices {
				// 4. Project from 3D to 2D using Projection Matrix
				// This converts View Space into Normalized Device Coordinates [-1, 1]
				projected3D := Mat4MulVec4Project(projMatrix, transformed)

				// 5. Scale Normalized Device Coordinates to Screen Space
				// NDC has +Y pointing UP. Screen Space has +Y pointing DOWN.
				projected2D := Vec2{
					X: (projected3D.X + 1.0) * 0.5 * float64(WindowWidth),
					Y: (1.0 - projected3D.Y) * 0.5 * float64(WindowHeight),
				}

				projectedPoints[i] = projected2D
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

		totalVertices := 0
		totalFaces := 0
		for _, e := range scene {
			if e.Mesh != nil {
				totalVertices += len(e.Mesh.Vertices)
				totalFaces += len(e.Mesh.Faces)
			}
		}

		stats := fmt.Sprintf(
			"Update Time: %.2f ms\nFPS: %.1f\nVertices: %d\nFaces: %d\nMode: %s (Press 1-5)",
			updateDuration.Seconds()*1000.0, lastFPS, totalVertices, totalFaces, modeStr,
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
