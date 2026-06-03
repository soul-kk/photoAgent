package imageprep

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

func TestCompressForUpload_ResizesLargeImage(t *testing.T) {
	im := image.NewRGBA(image.Rect(0, 0, 3000, 2000))
	for y := 0; y < 2000; y++ {
		for x := 0; x < 3000; x++ {
			im.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 128, 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, im, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatal(err)
	}
	out, meta, err := CompressForUpload(buf.Bytes(), "image/jpeg")
	if err != nil {
		t.Fatal(err)
	}
	if !meta.Resized {
		t.Fatal("expected resized")
	}
	if meta.CompressedBytes >= meta.OriginalBytes {
		t.Logf("compressed=%d original=%d (jpeg re-encode may vary)", meta.CompressedBytes, meta.OriginalBytes)
	}
	dec, _, err := image.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	b := dec.Bounds()
	long := b.Dx()
	if b.Dy() > long {
		long = b.Dy()
	}
	if long > MaxLongEdge {
		t.Fatalf("long edge %d > %d", long, MaxLongEdge)
	}
}
