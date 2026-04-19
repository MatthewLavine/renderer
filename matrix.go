package main

import "math"

// Mat4 represents a 4x4 transformation matrix
type Mat4 [4][4]float64

// Mat4Identity returns the identity matrix
func Mat4Identity() Mat4 {
	return Mat4{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
}

// Mat4Scale returns a scaling matrix
func Mat4Scale(sx, sy, sz float64) Mat4 {
	return Mat4{
		{sx, 0, 0, 0},
		{0, sy, 0, 0},
		{0, 0, sz, 0},
		{0, 0, 0, 1},
	}
}

// Mat4Translate returns a translation matrix
func Mat4Translate(tx, ty, tz float64) Mat4 {
	return Mat4{
		{1, 0, 0, tx},
		{0, 1, 0, ty},
		{0, 0, 1, tz},
		{0, 0, 0, 1},
	}
}

// Mat4RotateX returns a matrix for rotation around the X-axis
func Mat4RotateX(angle float64) Mat4 {
	c := math.Cos(angle)
	s := math.Sin(angle)
	return Mat4{
		{1, 0, 0, 0},
		{0, c, -s, 0},
		{0, s, c, 0},
		{0, 0, 0, 1},
	}
}

// Mat4RotateY returns a matrix for rotation around the Y-axis
func Mat4RotateY(angle float64) Mat4 {
	c := math.Cos(angle)
	s := math.Sin(angle)
	return Mat4{
		{c, 0, s, 0},
		{0, 1, 0, 0},
		{-s, 0, c, 0},
		{0, 0, 0, 1},
	}
}

// Mat4RotateZ returns a matrix for rotation around the Z-axis
func Mat4RotateZ(angle float64) Mat4 {
	c := math.Cos(angle)
	s := math.Sin(angle)
	return Mat4{
		{c, -s, 0, 0},
		{s, c, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
}

// Mat4MulMat4 multiplies two 4x4 matrices and returns the result (a * b)
func Mat4MulMat4(a, b Mat4) Mat4 {
	var result Mat4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			result[i][j] = a[i][0]*b[0][j] + a[i][1]*b[1][j] + a[i][2]*b[2][j] + a[i][3]*b[3][j]
		}
	}
	return result
}

// Mat4MulVec3 multiplies a 4x4 matrix by a 3D vector (assuming a 4th component W=1)
func Mat4MulVec3(m Mat4, v Vec3) Vec3 {
	var result Vec3
	result.X = m[0][0]*v.X + m[0][1]*v.Y + m[0][2]*v.Z + m[0][3]
	result.Y = m[1][0]*v.X + m[1][1]*v.Y + m[1][2]*v.Z + m[1][3]
	result.Z = m[2][0]*v.X + m[2][1]*v.Y + m[2][2]*v.Z + m[2][3]
	return result
}
