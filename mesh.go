package main

// Face represents a single triangle, storing the indices of the 3 vertices
type Face struct {
	A, B, C int
	Color   uint32
}

// Mesh holds the raw vertices and the triangular faces
type Mesh struct {
	Vertices []Vec3
	Faces    []Face
}

// Entity represents an object in the 3D scene, referencing a Mesh and holding its own transform
type Entity struct {
	Mesh        *Mesh
	Rotation    Vec3
	Scale       Vec3
	Translation Vec3
}
