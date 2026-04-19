package main

// Camera represents the player's viewpoint in the 3D world
type Camera struct {
	Position Vec3
	Yaw      float64 // Rotation around the Y axis (looking left/right)
	Pitch    float64 // Rotation around the X axis (looking up/down)
}

// Global camera instance
var globalCamera Camera
