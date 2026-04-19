package main

import (
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// DrawText draws a string onto our custom colorBuffer using a fixed 7x13 bitmap font.
// It iterates through the alpha mask of each character and sets the pixels individually.
func DrawText(x, y int, text string, color uint32) {
	face := basicfont.Face7x13
	startX := x

	// The font's glyphs are drawn relative to a baseline. We offset Y by the Ascent
	// so that the (x,y) coordinates you pass in represent the absolute top-left corner.
	baselineY := y + face.Ascent

	for _, r := range text {
		if r == '\n' {
			baselineY += face.Height
			startX = x
			continue
		}

		dr, mask, maskp, advance, ok := face.Glyph(fixed.Point26_6{}, r)
		if !ok {
			continue
		}

		// Iterate over the rectangular bounds of the glyph's mask
		for my := 0; my < dr.Dy(); my++ {
			for mx := 0; mx < dr.Dx(); mx++ {
				// Get the alpha value of the pixel
				_, _, _, a := mask.At(maskp.X+mx, maskp.Y+my).RGBA()

				// If the alpha is greater than 0, it means there is a pixel here for the character
				if a > 0 {
					SetPixel(startX+dr.Min.X+mx, baselineY+dr.Min.Y+my, color)
				}
			}
		}
		startX += advance.Floor()
	}
}
