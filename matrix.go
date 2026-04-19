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

// Mat4View returns a View Matrix which represents a camera.
// It applies the inverse of the camera's translation and rotation.
func Mat4View(pitch, yaw float64, position Vec3) Mat4 {
	// 1. Inverse Rotation: Rotate the world in the opposite direction the camera is looking
	rotX := Mat4RotateX(-pitch)
	rotY := Mat4RotateY(-yaw)
	
	// 2. Inverse Translation: Move the world in the opposite direction the camera is moving
	trans := Mat4Translate(-position.X, -position.Y, -position.Z)

	// Combine: View = RotX * RotY * Trans
	// Note: We apply Translation first to the world, then rotate it around the new origin (the camera's eyes)
	view := Mat4Identity()
	view = Mat4MulMat4(trans, view)
	view = Mat4MulMat4(rotY, view)
	view = Mat4MulMat4(rotX, view)

	return view
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

// Mat4Projection creates a perspective projection matrix.
// It normalizes coordinates into a [-1, 1] space based on FOV, Aspect Ratio, and clipping planes.
func Mat4Projection(fovDegrees, aspect, near, far float64) Mat4 {
	fovRadians := fovDegrees * (math.Pi / 180.0)
	f := 1.0 / math.Tan(fovRadians/2.0)

	return Mat4{
		{f / aspect, 0, 0, 0},
		{0, f, 0, 0},
		{0, 0, far / (far - near), -(far * near) / (far - near)},
		{0, 0, 1, 0},
	}
}

// Mat4MulVec4Project multiplies a vector by a projection matrix and performs perspective divide (divide by W).
func Mat4MulVec4Project(m Mat4, v Vec3) Vec3 {
	var result Vec3
	result.X = m[0][0]*v.X + m[0][1]*v.Y + m[0][2]*v.Z + m[0][3]
	result.Y = m[1][0]*v.X + m[1][1]*v.Y + m[1][2]*v.Z + m[1][3]
	result.Z = m[2][0]*v.X + m[2][1]*v.Y + m[2][2]*v.Z + m[2][3]
	
	// The 4th dimension W stores the original Z depth
	w := m[3][0]*v.X + m[3][1]*v.Y + m[3][2]*v.Z + m[3][3]

	// Perspective divide
	if w != 0 {
		result.X /= w
		result.Y /= w
		result.Z /= w
	}
	return result
}
