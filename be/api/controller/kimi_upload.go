package controller

import (
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	minMultiImageCount = 2
	maxMultiImageCount = 8
)

type uploadedImage struct {
	Index   int
	Raw     []byte
	Mime    string
	DataURL string
	Meta    map[string]any
}

// readMultipleImageBinaries 从 multipart 读取多张图片，字段名 images 或 image（可重复）。
func readMultipleImageBinaries(c *gin.Context) (images []uploadedImage, httpStatus int, errMsg string) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImageUploadBytes*maxMultiImageCount)
	if err := c.Request.ParseMultipartForm(maxImageUploadBytes * maxMultiImageCount); err != nil {
		return nil, http.StatusBadRequest, "multipart 解析失败: " + err.Error()
	}
	form := c.Request.MultipartForm
	if form == nil {
		return nil, http.StatusBadRequest, "请使用 multipart/form-data 上传图片"
	}

	var headers []*multipart.FileHeader
	if fh, ok := form.File["images"]; ok && len(fh) > 0 {
		headers = fh
	} else if fh, ok := form.File["image"]; ok && len(fh) > 0 {
		headers = fh
	} else if fh, ok := form.File["files"]; ok && len(fh) > 0 {
		headers = fh
	}
	if len(headers) < minMultiImageCount {
		return nil, http.StatusBadRequest,
			"请至少上传 2 张图片（multipart 字段 images 或 image，可多选）"
	}
	if len(headers) > maxMultiImageCount {
		return nil, http.StatusBadRequest,
			"最多支持 8 张图片对比"
	}

	out := make([]uploadedImage, 0, len(headers))
	for i, fh := range headers {
		if fh.Size > maxImageUploadBytes {
			return nil, http.StatusRequestEntityTooLarge, "单张图片超过大小上限"
		}
		src, err := fh.Open()
		if err != nil {
			return nil, http.StatusBadRequest, "无法打开上传文件"
		}
		b, err := io.ReadAll(io.LimitReader(src, maxImageUploadBytes+1))
		src.Close()
		if err != nil {
			return nil, http.StatusBadRequest, "读取上传文件失败"
		}
		if len(b) == 0 {
			return nil, http.StatusBadRequest, "存在空图片文件"
		}
		if len(b) > maxImageUploadBytes {
			return nil, http.StatusRequestEntityTooLarge, "单张图片超过大小上限"
		}
		mime := fh.Header.Get("Content-Type")
		dataURL, meta, err := imageBinaryToDataURL(b, mime)
		if err != nil {
			return nil, http.StatusBadRequest, "第 " + itoa(i+1) + " 张图片处理失败: " + err.Error()
		}
		out = append(out, uploadedImage{
			Index:   i,
			Raw:     nil,
			Mime:    mime,
			DataURL: dataURL,
			Meta:    meta,
		})
	}
	return out, 0, ""
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func readToneStyleDescription(c *gin.Context) (desc string, httpStatus int, errMsg string) {
	desc = strings.TrimSpace(c.PostForm("style_description"))
	if desc == "" {
		desc = strings.TrimSpace(c.PostForm("style"))
	}
	if desc == "" {
		desc = strings.TrimSpace(c.Query("style_description"))
	}
	if desc == "" {
		return "", http.StatusBadRequest, "请提供影调风格描述（字段 style_description，如「王家卫电影感」）"
	}
	if len(desc) > 500 {
		return "", http.StatusBadRequest, "风格描述过长（最多 500 字）"
	}
	return desc, 0, ""
}
