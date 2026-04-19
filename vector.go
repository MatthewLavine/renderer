package main

import "math"

// Vec2 represents a 2D vector
type Vec2 struct {
	X, Y float64
}

// Vec3 represents a 3D vector
type Vec3 struct {
	X, Y, Z float64
}

// --- Vec2 Methods ---

func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{X: v.X + other.X, Y: v.Y + other.Y}
}

func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{X: v.X - other.X, Y: v.Y - other.Y}
}

func (v Vec2) Mult(scalar float64) Vec2 {
	return Vec2{X: v.X * scalar, Y: v.Y * scalar}
}

func (v Vec2) Div(scalar float64) Vec2 {
	return Vec2{X: v.X / scalar, Y: v.Y / scalar}
}

func (v Vec2) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vec2) Dot(other Vec2) float64 {
	return (v.X * other.X) + (v.Y * other.Y)
}

func (v Vec2) Normalize() Vec2 {
	length := v.Length()
	if length == 0 {
		return Vec2{0, 0}
	}
	return v.Div(length)
}

// --- Vec3 Methods ---

func (v Vec3) Add(other Vec3) Vec3 {
	return Vec3{X: v.X + other.X, Y: v.Y + other.Y, Z: v.Z + other.Z}
}

func (v Vec3) Sub(other Vec3) Vec3 {
	return Vec3{X: v.X - other.X, Y: v.Y - other.Y, Z: v.Z - other.Z}
}

func (v Vec3) Mult(scalar float64) Vec3 {
	return Vec3{X: v.X * scalar, Y: v.Y * scalar, Z: v.Z * scalar}
}

func (v Vec3) Div(scalar float64) Vec3 {
	return Vec3{X: v.X / scalar, Y: v.Y / scalar, Z: v.Z / scalar}
}

func (v Vec3) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// Dot product returns a scalar value indicating how much two vectors point in the same direction
func (v Vec3) Dot(other Vec3) float64 {
	return (v.X * other.X) + (v.Y * other.Y) + (v.Z * other.Z)
}

// Cross product returns a new vector perpendicular to both input vectors
func (v Vec3) Cross(other Vec3) Vec3 {
	return Vec3{
		X: v.Y*other.Z - v.Z*other.Y,
		Y: v.Z*other.X - v.X*other.Z,
		Z: v.X*other.Y - v.Y*other.X,
	}
}

func (v Vec3) Normalize() Vec3 {
	length := v.Length()
	if length == 0 {
		return Vec3{0, 0, 0}
	}
	return v.Div(length)
}

func (v Vec3) RotateX(angle float64) Vec3 {
	return Vec3{
		X: v.X,
		Y: v.Y*math.Cos(angle) - v.Z*math.Sin(angle),
		Z: v.Y*math.Sin(angle) + v.Z*math.Cos(angle),
	}
}

func (v Vec3) RotateY(angle float64) Vec3 {
	return Vec3{
		X: v.X*math.Cos(angle) + v.Z*math.Sin(angle),
		Y: v.Y,
		Z: -v.X*math.Sin(angle) + v.Z*math.Cos(angle),
	}
}

func (v Vec3) RotateZ(angle float64) Vec3 {
	return Vec3{
		X: v.X*math.Cos(angle) - v.Y*math.Sin(angle),
		Y: v.X*math.Sin(angle) + v.Y*math.Cos(angle),
		Z: v.Z,
	}
}
