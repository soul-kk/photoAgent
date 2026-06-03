package imageprep

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"image/jpeg"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// 上传 Kimi 前压缩：控制长边与 JPEG 质量，降低上游视觉 token 与延迟。
const (
	MaxLongEdge = 1024
	JPEGQuality = 80
)

// Meta 压缩结果元信息（可写入 SSE meta 事件）。
type Meta struct {
	OriginalBytes   int  `json:"original_bytes"`
	CompressedBytes int  `json:"compressed_bytes"`
	Resized         bool `json:"resized"`
}

// CompressForUpload 将常见图片格式解码后缩放到长边不超过 MaxLongEdge，再编码为 JPEG。
func CompressForUpload(raw []byte, mimeHint string) (out []byte, meta Meta, err error) {
	meta.OriginalBytes = len(raw)
	if len(raw) == 0 {
		return nil, meta, fmt.Errorf("图像数据为空")
	}
	im, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, meta, fmt.Errorf("解码图片失败: %w", err)
	}
	b := im.Bounds()
	w, h := b.Dx(), b.Dy()
	nw, nh := w, h
	if w >= h && w > MaxLongEdge {
		nw = MaxLongEdge
		nh = max(1, h*MaxLongEdge/w)
		meta.Resized = true
	} else if h > w && h > MaxLongEdge {
		nh = MaxLongEdge
		nw = max(1, w*MaxLongEdge/h)
		meta.Resized = true
	}
	var rgba *image.RGBA
	if nw == w && nh == h {
		rgba = copyToRGBA(im)
	} else {
		dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
		draw.CatmullRom.Scale(dst, dst.Bounds(), im, b, draw.Over, nil)
		rgba = dst
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, rgba, &jpeg.Options{Quality: JPEGQuality}); err != nil {
		return nil, meta, fmt.Errorf("JPEG 编码失败: %w", err)
	}
	out = buf.Bytes()
	meta.CompressedBytes = len(out)
	return out, meta, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func copyToRGBA(src image.Image) *image.RGBA {
	b := src.Bounds()
	out := image.NewRGBA(b)
	draw.Draw(out, b, src, b.Min, draw.Src)
	return out
}
