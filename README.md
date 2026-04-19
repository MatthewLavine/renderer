# Go Software Rasterizer

A high-performance, dependency-light 3D software rasterizer built from scratch in Go. This engine implements the entire graphics pipeline on the CPU, manually calculating every pixel before pushing the final frame to an SDL2 window.

## Features

### 🛠️ The 3D Pipeline
*   **4x4 Matrix Math**: Fully integrated matrix transformation pipeline (Scale, Rotate, Translate).
*   **Coordinate Spaces**: Vertex transformations from Local Space → World Space → View Space → NDC → Screen Space.
*   **Projection**: 4x4 Perspective Projection Matrix with Field of View (FOV) and Aspect Ratio support.
*   **Clipping**: Near-plane culling to prevent division-by-zero crashes when geometry passes behind the camera.

### 🎨 Rendering Engine
*   **Barycentric Rasterizer**: A modern per-pixel rasterizer that calculates weights for perfect attribute interpolation.
*   **Z-Buffer (Depth Buffer)**: Pixel-perfect depth sorting that handles intersecting geometry flawlessly.
*   **Back-face Culling**: Optimized rendering that discards triangles facing away from the camera.
*   **Flat Shading**: Directional lighting system that calculates surface intensity based on face normals.
*   **Multiple Render Modes**: Real-time hot-swapping between Wireframe, Solid, and Shaded modes.

### 🚀 Systems & Architecture
*   **OBJ Loader**: Custom parser for Wavefront `.obj` files.
*   **Entity System**: Support for multiple 3D objects with independent transforms sharing mesh data.
*   **Display Features**: 1080p internal resolution, resizable window, and 16:9 aspect ratio locking with automatic letterboxing.
*   **HUD Overlay**: Real-time performance statistics (FPS, Update time, Vertex/Face counts).

## Controls

| Key | Action |
|-----|--------|
| **W / S** | Move Forward / Backward |
| **A / D** | Strafe Left / Right |
| **Q / E** | Fly Up / Down |
| **Arrows** | Look Around (Pitch / Yaw) |
| **1 - 5** | Change Render Mode (Wireframe -> Shaded) |
| **R** | Reset Camera Position |
| **ESC** | Exit Application |

## Getting Started

### Prerequisites
*   **Go** (1.18+)
*   **SDL2 Development Headers**
    *   Ubuntu/Debian: `sudo apt install libsdl2-dev`
    *   Fedora: `sudo dnf install SDL2-devel`
    *   macOS: `brew install sdl2`

### Running the Engine
1. Clone the repository.
2. Build the application:
   ```bash
   go build -o renderer_app
   ```
3. Run the executable:
   ```bash
   ./renderer_app
   ```

## Project Structure
*   `main.go`: Application entry point, main loop, and input handling.
*   `display.go`: Core rasterization logic, Z-Buffer, and SDL abstraction.
*   `matrix.go`: 4x4 Matrix math and transformations.
*   `vector.go`: Vec2/Vec3/Vec4 structures and vector math.
*   `camera.go`: Camera system and View matrix logic.
*   `mesh.go`: Mesh and Entity structures.
*   `obj_parser.go`: OBJ file loading logic.
*   `text.go`: Custom bitmap font renderer for the HUD.
