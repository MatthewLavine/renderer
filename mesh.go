package main

// Face represents a single triangle, storing the indices of the 3 vertices
type Face struct {
	A, B, C int
	Color   uint32
}

// Mesh holds the raw vertices, the triangular faces, and spatial properties
type Mesh struct {
	Vertices    []Vec3
	Faces       []Face
	Rotation    Vec3
	Scale       Vec3
	Translation Vec3
}

// Global mesh instance
var currentMesh Mesh
