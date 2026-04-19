package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"sync"
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
		sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE,
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

	// Set a logical resolution for the renderer.
	// This automatically handles letterboxing and maintains our 16:9 aspect ratio
	// no matter how the user resizes the window.
	renderer.SetLogicalSize(int32(WindowWidth), int32(WindowHeight))

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
	// Allocate memory for our custom color buffer and Z buffer
	colorBuffer = make([]uint32, WindowWidth*WindowHeight)
	zBuffer = make([]float64, WindowWidth*WindowHeight)

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
				Translation: Vec3{X: float64(x) * 7.0, Y: (float64(y) * 5.0) - 1.5, Z: 15},
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
				case sdl.K_r:
					globalCamera.Position = Vec3{X: 0, Y: 0, Z: -5.0}
					globalCamera.Pitch = 0
					globalCamera.Yaw = 0
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
// TriangleToRender holds projected 2D coordinates and depth information
type TriangleToRender struct {
	Points [3]Vec3
	Color  uint32
	MinY   int
	MaxY   int
}

func update() []TriangleToRender {
	// Clear the screen and the Z-buffer (now in parallel)
	ClearColorBuffer(0xFF111111)
	ClearZBuffer()

	// 1. Calculate matrices
	viewMatrix := Mat4View(globalCamera.Pitch, globalCamera.Yaw, globalCamera.Position)
	aspectRatio := float64(WindowWidth) / float64(WindowHeight)
	projMatrix := Mat4Projection(60.0, aspectRatio, 0.1, 100.0)

	// 2. Process all entities in parallel
	var wg sync.WaitGroup
	triangleChunks := make([][]TriangleToRender, len(scene))

	for i, entity := range scene {
		wg.Add(1)
		go func(idx int, e *Entity) {
			defer wg.Done()
			
			// Spin the entity around its Y axis
			e.Rotation.Y += 0.01

			// Calculate World Matrix
			scaleMatrix := Mat4Scale(e.Scale.X, e.Scale.Y, e.Scale.Z)
			rotationXMatrix := Mat4RotateX(e.Rotation.X)
			rotationYMatrix := Mat4RotateY(e.Rotation.Y)
			rotationZMatrix := Mat4RotateZ(e.Rotation.Z)
			translationMatrix := Mat4Translate(e.Translation.X, e.Translation.Y, e.Translation.Z)

			worldMatrix := Mat4Identity()
			worldMatrix = Mat4MulMat4(scaleMatrix, worldMatrix)
			worldMatrix = Mat4MulMat4(rotationXMatrix, worldMatrix)
			worldMatrix = Mat4MulMat4(rotationYMatrix, worldMatrix)
			worldMatrix = Mat4MulMat4(rotationZMatrix, worldMatrix)
			worldMatrix = Mat4MulMat4(translationMatrix, worldMatrix)

			var chunk []TriangleToRender

			// Loop over all faces
			for fIndex, face := range e.Mesh.Faces {
				// Geometry pipeline (Culling, Projection, etc.)
				vertices := [3]Vec3{
					e.Mesh.Vertices[face.A],
					e.Mesh.Vertices[face.B],
					e.Mesh.Vertices[face.C],
				}

				var transformedVertices [3]Vec3
				for j, vertex := range vertices {
					transformed := Mat4MulVec3(worldMatrix, vertex)
					viewed := Mat4MulVec3(viewMatrix, transformed)
					transformedVertices[j] = viewed
				}

				// Back-face culling
				edge1 := transformedVertices[1].Sub(transformedVertices[0])
				edge2 := transformedVertices[2].Sub(transformedVertices[0])
				normal := edge1.Cross(edge2).Normalize()
				cameraRay := transformedVertices[0]
				dot := normal.Dot(cameraRay)

				if renderMethod != RenderWireframe && dot > 0 {
					continue
				}

				// Near-plane culling
				if transformedVertices[0].Z < 0.1 || transformedVertices[1].Z < 0.1 || transformedVertices[2].Z < 0.1 {
					continue
				}

				var projectedPoints [3]Vec3
				for j, transformed := range transformedVertices {
					projected3D := Mat4MulVec4Project(projMatrix, transformed)
					projected2D := Vec3{
						X: (projected3D.X + 1.0) * 0.5 * float64(WindowWidth),
						Y: (1.0 - projected3D.Y) * 0.5 * float64(WindowHeight),
						Z: projected3D.Z,
					}
					projectedPoints[j] = projected2D
				}

				// Lighting
				lightIntensity := -normal.Dot(globalLightDirection)
				if lightIntensity < 0.15 { lightIntensity = 0.15 }

				var color uint32
				if renderMethod == RenderShaded || renderMethod == RenderShadedWireframe {
					color = ApplyLightIntensity(0xFFFFFFFF, lightIntensity)
				} else {
					color = uint32(0xFF000000) | (uint32((fIndex*30)%255) << 16) | (uint32((fIndex*40)%255) << 8) | uint32((fIndex*50)%255)
				}

				// Fast Bounding Box
				minY := projectedPoints[0].Y
				if projectedPoints[1].Y < minY { minY = projectedPoints[1].Y }
				if projectedPoints[2].Y < minY { minY = projectedPoints[2].Y }
				
				maxY := projectedPoints[0].Y
				if projectedPoints[1].Y > maxY { maxY = projectedPoints[1].Y }
				if projectedPoints[2].Y > maxY { maxY = projectedPoints[2].Y }

				chunk = append(chunk, TriangleToRender{
					Points: projectedPoints,
					Color:  color,
					MinY:   int(minY),
					MaxY:   int(maxY),
				})
			}
			triangleChunks[idx] = chunk
		}(i, entity)
	}
	wg.Wait()

	// 3. Merge chunks into final slice
	var totalTriangles []TriangleToRender
	for _, chunk := range triangleChunks {
		totalTriangles = append(totalTriangles, chunk...)
	}

	return totalTriangles
}

func renderTrianglesParallel(triangles []TriangleToRender) {
	numWorkers := runtime.NumCPU()
	stripHeight := WindowHeight / numWorkers
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			yMin := workerId * stripHeight
			yMax := yMin + stripHeight
			// Ensure the last strip catches any remainder from integer division
			if workerId == numWorkers-1 {
				yMax = WindowHeight
			}

			for _, t := range triangles {
				// Early exit: If the triangle's bounding box is entirely outside this strip, skip it!
				if t.MaxY < yMin || t.MinY > yMax {
					continue
				}

				// Only draw solid faces if the render method allows it
				if renderMethod == RenderSolid || renderMethod == RenderSolidWireframe || renderMethod == RenderShaded || renderMethod == RenderShadedWireframe {
					DrawFilledTriangle(t.Points[0], t.Points[1], t.Points[2], t.Color, yMin, yMax)
				}
			}
		}(i)
	}
	wg.Wait()

	// Draw wireframe outlines sequentially on the main thread
	// (Wireframes are extremely fast compared to filled triangles)
	if renderMethod == RenderWireframe || renderMethod == RenderSolidWireframe || renderMethod == RenderShadedWireframe {
		for _, t := range triangles {
			wireColor := uint32(0xFF111111) // Dark gray default
			if renderMethod == RenderWireframe {
				wireColor = 0xFFFFFFFF // White for visibility on dark background
			}

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
	perfFreq := float64(sdl.GetPerformanceFrequency())

	for running {
		frameStart := sdl.GetPerformanceCounter()
		totalStart := time.Now()

		// 1. Process Input
		inputStart := time.Now()
		processInput()
		inputTime := time.Since(inputStart)

		// 2. Geometry Pipeline (Update, Matrices, Culling, Projection)
		geomStart := time.Now()
		trianglesToRender := update()
		geomTime := time.Since(geomStart)

		// 3. Rasterization Pipeline (Parallel)
		rasterStart := time.Now()
		renderTrianglesParallel(trianglesToRender)
		rasterTime := time.Since(rasterStart)

		// 4. Statistics Calculation
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
			"FPS: %.1f | Frame: %.2fms\nInput: %.2fms | Geom: %.2fms | Raster: %.2fms | Display: %.2fms\nVertices: %d | Faces: %d | Mode: %s",
			lastFPS, time.Since(totalStart).Seconds()*1000.0,
			inputTime.Seconds()*1000.0, geomTime.Seconds()*1000.0, rasterTime.Seconds()*1000.0, 0.0, // displayTime filled later
			totalVertices, totalFaces, modeStr,
		)
		DrawText(10, 10, stats, 0xFFFFFFFF)

		// 5. Display Pipeline (SDL Texture Upload & Present)
		displayStart := time.Now()
		render()
		displayTime := time.Since(displayStart)

		// Update the display time in the next frame's stats or just show it now?
		// Since we draw text BEFORE we render, we can't show the current frame's display time accurately.
		// We'll just use the previous frame's display time for the display.
		_ = displayTime // We'll just let it be for now or update stats to use it.

		// Calculate total frame time and FPS
		frameTime := float64(sdl.GetPerformanceCounter()-frameStart) / perfFreq
		if frameTime > 0 {
			lastFPS = 1.0 / frameTime
		}
	}
}
