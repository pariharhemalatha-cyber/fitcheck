package ai

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"image/jpeg"
	"os"

	"golang.org/x/image/draw"
)

const maxVisionBytes = 1 << 20 // 1 MB — resize above this
const maxVisionDim   = 1024

// prepareImageForVision reads and optionally downscales an image for API upload.
func prepareImageForVision(imagePath string) ([]byte, string, error) {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, "", err
	}

	mime := mimeFromExt(imagePath)
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data, mime, nil
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	needsResize := len(data) > maxVisionBytes || w > maxVisionDim || h > maxVisionDim
	if !needsResize {
		return data, mime, nil
	}

	scale := float64(maxVisionDim) / float64(max(w, h))
	nw := int(float64(w) * scale)
	nh := int(float64(h) * scale)
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 85}); err != nil {
		return nil, "", fmt.Errorf("encode resized image: %w", err)
	}
	return buf.Bytes(), "image/jpeg", nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
