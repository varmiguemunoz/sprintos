package tray

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

func iconBytes() []byte {
	const size = 22
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	const cx, cy = float64(size) / 2.0, float64(size) / 2.0
	const r = float64(size)/2.0 - 1.5

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) + 0.5 - cx
			dy := float64(y) + 0.5 - cy
			if dx*dx+dy*dy <= r*r {
				img.Set(x, y, color.RGBA{R: 124, G: 58, B: 237, A: 255})
			}
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
