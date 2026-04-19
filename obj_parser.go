package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// LoadOBJ reads an OBJ file from disk and populates the currentMesh
func LoadOBJ(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Initialize the global mesh
	currentMesh = Mesh{
		Rotation:    Vec3{X: 0, Y: 0, Z: 0},
		Scale:       Vec3{X: 1, Y: 1, Z: 1},
		Translation: Vec3{X: 0, Y: -1.5, Z: 6}, // Push the mesh 6 units into the screen, and down 1.5 units to center it
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "v":
			// Parse Vertex
			if len(fields) >= 4 {
				x, _ := strconv.ParseFloat(fields[1], 64)
				y, _ := strconv.ParseFloat(fields[2], 64)
				z, _ := strconv.ParseFloat(fields[3], 64)
				currentMesh.Vertices = append(currentMesh.Vertices, Vec3{X: x, Y: y, Z: z})
			}
		case "f":
			// Parse Face
			// OBJ format can be "f v1/vt1/vn1 v2/vt2/vn2 v3/vt3/vn3"
			// We only care about the first part (the vertex index)
			if len(fields) >= 4 {
				a := parseFaceVertex(fields[1])
				b := parseFaceVertex(fields[2])
				c := parseFaceVertex(fields[3])

				// OBJ indices are 1-based, our arrays are 0-based
				currentMesh.Faces = append(currentMesh.Faces, Face{
					A:     a - 1,
					B:     b - 1,
					C:     c - 1,
					Color: 0xFF00FF00, // Default green color for wireframes
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	fmt.Printf("\nLoaded %s: %d vertices, %d faces\n", filename, len(currentMesh.Vertices), len(currentMesh.Faces))
	return nil
}

// parseFaceVertex extracts the integer vertex index from an OBJ face token like "1/1/1" or just "1"
func parseFaceVertex(token string) int {
	parts := strings.Split(token, "/")
	val, _ := strconv.Atoi(parts[0])
	return val
}
