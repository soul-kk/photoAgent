package controller

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"go-service-starter/config"
	"go-service-starter/core/imageprep"
	"go-service-starter/core/kimigate"
	"go-service-starter/core/libx"

	"github.com/gin-gonic/gin"
)

type KimiController struct {
	gate *kimigate.Gate
}

func NewKimiController() *KimiController {
	cfg := config.GetConfig()
	gate := kimigate.New(kimigate.Options{
		MaxConcurrent: cfg.Kimi.MaxConcurrent,
		TimeoutSec:    cfg.Kimi.TimeoutSec,
		QueueWaitSec:  cfg.Kimi.QueueWaitSec,
	})
	maxC, timeoutSec, queueSec := cfg.Kimi.MaxConcurrent, cfg.Kimi.TimeoutSec, cfg.Kimi.QueueWaitSec
	if maxC <= 0 {
		maxC = 5
	}
	if timeoutSec <= 0 {
		timeoutSec = 300
	}
	if queueSec <= 0 {
		queueSec = 30
	}
	log.Printf("kimigate: max_concurrent=%d timeout_sec=%d queue_wait_sec=%d", maxC, timeoutSec, queueSec)
	return &KimiController{gate: gate}
}

func (k *KimiController) respondKimiGateBusy(c *gin.Context, err error) bool {
	if !errors.Is(err, kimigate.ErrTooManyConcurrent) {
		return false
	}
	c.Header("Retry-After", "30")
	libx.Err(c, http.StatusServiceUnavailable, "AI 服务繁忙，请稍后重试", err)
	return true
}

// kimiK26ChatExtras kimi-k2.6 要求 temperature=1；非流式长生成需足够 max_tokens。
func kimiK26ChatExtras(maxTokens int) map[string]any {
	extras := map[string]any{"temperature": 1}
	if maxTokens > 0 {
		extras["max_tokens"] = maxTokens
	}
	return extras
}

func (k *KimiController) kimiUpstreamContext(parent context.Context) (context.Context, context.CancelFunc) {
	return k.gate.UpstreamContext(parent)
}

func (k *KimiController) RegisterPublic(r *gin.RouterGroup) {
	r.POST("/kimi/generate", k.Generate)
	r.POST("/kimi/recognize-image", k.RecognizeImageBinary)
	r.POST("/kimi/recognize/image", k.RecognizeImageBinary)
}

func (k *KimiController) RegisterProtected(r *gin.RouterGroup) {
	r.POST("/kimi/photography/analyze-image", k.PhotographyAnalyzeImage)
	r.POST("/kimi/analyze-photo", k.PhotographyAnalyzeImage)
	r.GET("/kimi/photography/analyze-image", k.photoAnalyzeMethodHint)
	r.GET("/kimi/analyze-photo", k.photoAnalyzeMethodHint)
	r.POST("/kimi/photography/score-image", k.PhotographyScoreImage)
	r.POST("/kimi/score-photo", k.PhotographyScoreImage)
	r.GET("/kimi/photography/score-image", k.photoScoreMethodHint)
	r.GET("/kimi/score-photo", k.photoScoreMethodHint)
	r.POST("/kimi/photography/shoot-advice", k.PhotographyShootAdvice)
	r.POST("/kimi/shoot-advice", k.PhotographyShootAdvice)
	r.GET("/kimi/photography/shoot-advice", k.photoShootAdviceMethodHint)
	r.GET("/kimi/shoot-advice", k.photoShootAdviceMethodHint)
	r.POST("/kimi/photography/compare-images", k.PhotographyCompareImages)
	r.POST("/kimi/compare-photos", k.PhotographyCompareImages)
	r.GET("/kimi/photography/compare-images", k.photoCompareMethodHint)
	r.GET("/kimi/compare-photos", k.photoCompareMethodHint)
	r.POST("/kimi/photography/tone-style", k.PhotographyToneStyle)
	r.POST("/kimi/tone-style", k.PhotographyToneStyle)
	r.GET("/kimi/photography/tone-style", k.photoToneStyleMethodHint)
	r.GET("/kimi/tone-style", k.photoToneStyleMethodHint)
}

func (k *KimiController) photoAnalyzeMethodHint(c *gin.Context) {
	libx.Err(c, http.StatusMethodNotAllowed,
		"请使用 POST（需先登录；Bearer token）；multipart：image/file；可选 prompt、focus_dimension（构图|色彩|曝光|拍摄内容 或 composition|color|exposure|content）。默认 stream=false：四维 JSON 评分+整体分析；指定 focus_dimension 时额外返回 focused_deep_analysis。stream=true 为 SSE Markdown。路径：POST /api/kimi/photography/analyze-image",
		nil)
}

func (k *KimiController) photoScoreMethodHint(c *gin.Context) {
	libx.Err(c, http.StatusMethodNotAllowed,
		"请使用 POST（需先登录；Bearer token）；multipart：image/file；可选 prompt、focus_dimension（构图|色彩|曝光|拍摄内容）。路径：POST /api/kimi/photography/score-image；返回 exposure/color/composition/content 四维 0–100 分 + text_analysis + improvement_tips；指定 focus_dimension 时额外返回 focused_deep_analysis",
		nil)
}

// 单张图片二进制上限（multipart 整请求或原始 body 都会受此上限约束）
const maxImageUploadBytes = 20 << 20

// kimiModelK26 多模态接口固定使用该模型（不使用配置文件里的默认模型占位）
const kimiModelK26 = "kimi-k2.6"

const photographyUserBase = "请根据上述要求，对附图从摄影角度进行分析并给出建议。"

const photographyScoreUserBase = "请仅依据附图输出四维评分、整体分析与改进建议。"

type kimiGenerateBody struct {
	Text   string   `json:"text"`
	Prompt string   `json:"prompt"`
	Model  string   `json:"model"`
	Images []string `json:"images"` // 可选：每条为完整 data URL、http(s) URL、ms:// 引用，或裸 base64（按 JPEG data URL 拼接）
	Videos []string `json:"videos"` // 可选：同上，视频多为 data:video/...;base64,...
}

func (b *kimiGenerateBody) userText() string {
	t := strings.TrimSpace(b.Text)
	if t != "" {
		return t
	}
	return strings.TrimSpace(b.Prompt)
}

func publicKimiNetworkErr(err error) error {
	if err == nil {
		return nil
	}
	var op *net.OpError
	if errors.As(err, &op) && op.Timeout() {
		return fmt.Errorf("连接 Kimi（Moonshot）接口超时，请检查网络或 HTTPS_PROXY")
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return fmt.Errorf("连接 Kimi（Moonshot）接口超时，请检查网络或 HTTPS_PROXY")
	}
	low := strings.ToLower(err.Error())
	if strings.Contains(low, "timeout") || strings.Contains(low, "i/o timeout") ||
		errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("连接 Kimi（Moonshot）接口超时（非流式 stream=false 常需 1–3 分钟；请将 Apifox 超时设为 300s+，或改用 stream=true）")
	}
	if strings.Contains(low, "connection refused") {
		return fmt.Errorf("连接被拒绝，请检查本机或 HTTPS_PROXY 代理是否可用")
	}
	return fmt.Errorf("无法连接 Kimi 服务，请检查网络、防火墙或代理（勿在响应中暴露上游详情）")
}

func sanitizeKimiAPIKey(key string) string {
	key = strings.TrimSpace(key)
	key = strings.TrimPrefix(strings.TrimPrefix(key, "Bearer "), "bearer ")
	key = strings.Trim(key, `"'`)
	return strings.TrimSpace(key)
}

// normalizeMediaURL 将单条输入转为 Kimi image_url / video_url 可用的 url 字段。
func normalizeMediaURL(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "data:image/") ||
		strings.HasPrefix(s, "data:video/") ||
		strings.HasPrefix(s, "http://") ||
		strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "ms://") {
		return s
	}
	return "data:image/jpeg;base64," + s
}

func normalizeVideoURL(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "data:video/") ||
		strings.HasPrefix(s, "http://") ||
		strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "ms://") {
		return s
	}
	return "data:video/mp4;base64," + s
}

func buildKimiUserContent(text string, images, videos []string) any {
	hasMedia := len(images) > 0 || len(videos) > 0
	if !hasMedia {
		return text
	}
	parts := make([]map[string]any, 0, len(images)+len(videos)+1)
	for _, img := range images {
		u := normalizeMediaURL(img)
		if u == "" {
			continue
		}
		parts = append(parts, map[string]any{
			"type":      "image_url",
			"image_url": map[string]string{"url": u},
		})
	}
	for _, v := range videos {
		u := normalizeVideoURL(v)
		if u == "" {
			continue
		}
		parts = append(parts, map[string]any{
			"type":      "video_url",
			"video_url": map[string]string{"url": u},
		})
	}
	t := strings.TrimSpace(text)
	if t == "" {
		t = "请结合以上内容进行描述或回答。"
	}
	parts = append(parts, map[string]any{"type": "text", "text": t})
	return parts
}

func (k *KimiController) postKimiChat(ctx context.Context, model, base, apiKey string, messages []map[string]any, payloadExtras map[string]any) (statusCode int, respBody []byte, reqErr error) {
	if err := k.gate.Acquire(ctx); err != nil {
		return 0, nil, err
	}
	defer k.gate.Release()

	payload := map[string]any{
		"model":    model,
		"messages": messages,
	}
	for ek, ev := range payloadExtras {
		payload[ek] = ev
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, err
	}
	upCtx, upCancel := k.kimiUpstreamContext(ctx)
	defer upCancel()

	apiURL := strings.TrimRight(strings.TrimSpace(base), "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(upCtx, http.MethodPost, apiURL, bytes.NewReader(raw))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := k.gate.HTTPClient().Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, body, nil
}

// imageBinaryToDataURL 将图片压缩为 JPEG 后编码为 data URL，供 Kimi 多模态调用；压缩失败时回退为原图 data URL。
func imageBinaryToDataURL(raw []byte, mimeHint string) (string, map[string]any, error) {
	if len(raw) == 0 {
		return "", nil, fmt.Errorf("图像数据为空")
	}
	compressed, meta, err := imageprep.CompressForUpload(raw, mimeHint)
	if err == nil {
		b64 := base64.StdEncoding.EncodeToString(compressed)
		metaMap := map[string]any{
			"compressed":       true,
			"original_bytes":   meta.OriginalBytes,
			"compressed_bytes": meta.CompressedBytes,
			"resized":          meta.Resized,
			"mime":             "image/jpeg",
		}
		return "data:image/jpeg;base64," + b64, metaMap, nil
	}
	url, rawErr := imageBinaryToDataURLRaw(raw, mimeHint)
	metaMap := map[string]any{
		"compressed":     false,
		"original_bytes": len(raw),
		"compress_error": err.Error(),
	}
	if rawErr != nil {
		return "", metaMap, rawErr
	}
	return url, metaMap, nil
}

func imageBinaryToDataURLRaw(raw []byte, mimeHint string) (string, error) {
	mimeHint = strings.TrimSpace(strings.ToLower(mimeHint))
	if mimeHint != "" && mimeHint != "application/octet-stream" && strings.HasPrefix(mimeHint, "image/") {
		b64 := base64.StdEncoding.EncodeToString(raw)
		return fmt.Sprintf("data:%s;base64,%s", mimeHint, b64), nil
	}
	sample := raw
	if len(sample) > 512 {
		sample = raw[:512]
	}
	mt := http.DetectContentType(sample)
	switch mt {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		b64 := base64.StdEncoding.EncodeToString(raw)
		return fmt.Sprintf("data:%s;base64,%s", mt, b64), nil
	default:
		return "", fmt.Errorf("无法识别的图片格式 (%s)，请上传 jpeg/png/gif/webp", mt)
	}
}

// readSingleImageBinary 读取一张上传图；httpStatus 非 0 时表示失败，errMsg 为客户端提示。
func readSingleImageBinary(c *gin.Context) (raw []byte, mimeHint string, httpStatus int, errMsg string) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImageUploadBytes)
	ct := c.ContentType()
	if strings.Contains(ct, "multipart/form-data") {
		if err := c.Request.ParseMultipartForm(maxImageUploadBytes); err != nil {
			return nil, "", http.StatusBadRequest, "multipart 解析失败: " + err.Error()
		}
		fh, err := c.FormFile("image")
		if err != nil {
			fh, err = c.FormFile("file")
		}
		if err != nil {
			return nil, "", http.StatusBadRequest, "请使用 multipart 上传字段 image 或 file"
		}
		if fh.Size > maxImageUploadBytes {
			return nil, "", http.StatusRequestEntityTooLarge, "图片超过大小上限"
		}
		src, err := fh.Open()
		if err != nil {
			return nil, "", http.StatusBadRequest, "无法打开上传文件"
		}
		b, err := io.ReadAll(io.LimitReader(src, maxImageUploadBytes+1))
		src.Close()
		if err != nil {
			return nil, "", http.StatusBadRequest, "读取上传文件失败"
		}
		if len(b) > maxImageUploadBytes {
			return nil, "", http.StatusRequestEntityTooLarge, "图片超过大小上限"
		}
		return b, fh.Header.Get("Content-Type"), 0, ""
	}
	if strings.HasPrefix(ct, "image/") || ct == "application/octet-stream" {
		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return nil, "", http.StatusBadRequest, "读取请求正文失败"
		}
		if len(b) == 0 {
			return nil, "", http.StatusBadRequest, "请求正文为空"
		}
		return b, ct, 0, ""
	}
	return nil, "", http.StatusUnsupportedMediaType,
		"请使用 multipart/form-data 上传字段 image 或 file，或使用 Content-Type 为 image/* 或 application/octet-stream 传递原始图片二进制"
}

func (k *KimiController) finishKimiFromUpstream(c *gin.Context, status int, respBody []byte, model string) {
	if status != http.StatusOK {
		if status == http.StatusUnauthorized {
			libx.Err(c, http.StatusUnauthorized,
				"Kimi 鉴权失败：密钥与 base_url 需同属一国别开放平台——在中国大陆申请的密钥请设 kimi.base_url 为 https://api.moonshot.cn/v1；国际/ kimi.ai 侧密钥一般用 https://api.moonshot.ai/v1。并核对密钥未过期、整串粘贴且无多余字符",
				nil)
			return
		}
		libx.Err(c, status, "Kimi 返回错误", fmt.Errorf("%s", string(respBody)))
		return
	}
	out, err := parseKimiChatResponse(respBody)
	if err != nil {
		libx.Err(c, http.StatusBadGateway, "解析 Kimi 响应失败", err)
		return
	}
	libx.Ok(c, "ok", gin.H{
		"text":  out,
		"model": model,
	})
}

// RecognizeImageBinary 接收一张图片的二进制（无需登录）：
// - multipart/form-data：字段名 image 或 file；可选表单字段 prompt、可选查询参数 model
// - 原始正文：Content-Type 为 image/jpeg | image/png | image/webp | image/gif 或 application/octet-stream（按嗅探识别）；可选查询参数 prompt、model
func (k *KimiController) RecognizeImageBinary(c *gin.Context) {
	cfg := config.GetConfig()
	key := sanitizeKimiAPIKey(cfg.Kimi.APIKey)
	if key == "" {
		libx.Err(c, http.StatusInternalServerError, "未配置 kimi.api_key（或环境变量 KIMI_API_KEY / MOONSHOT_API_KEY）", nil)
		return
	}

	model := strings.TrimSpace(c.Query("model"))
	if model == "" {
		model = strings.TrimSpace(cfg.Kimi.Model)
	}
	if model == "" {
		model = kimiModelK26
	}

	base := strings.TrimSpace(cfg.Kimi.BaseURL)
	if base == "" {
		base = "https://api.moonshot.ai/v1"
	}

	prompt := strings.TrimSpace(c.Query("prompt"))
	if prompt == "" {
		prompt = strings.TrimSpace(c.PostForm("prompt"))
	}
	if prompt == "" {
		prompt = "请识别并简要描述这张图片的主要内容。"
	}

	raw, mimeHint, httpStatus, errMsg := readSingleImageBinary(c)
	if httpStatus != 0 {
		libx.Err(c, httpStatus, errMsg, nil)
		return
	}

	dataURL, _, err := imageBinaryToDataURL(raw, mimeHint)
	raw = nil
	if err != nil {
		libx.Err(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	userContent := buildKimiUserContent(prompt, []string{dataURL}, nil)
	msgs := []map[string]any{{"role": "user", "content": userContent}}
	status, respBody, err := k.postKimiChat(c.Request.Context(), model, base, key, msgs, nil)
	if err != nil {
		if k.respondKimiGateBusy(c, err) {
			return
		}
		log.Printf("kimi upstream error: %v", err)
		libx.Err(c, http.StatusBadGateway, "调用 Kimi 失败", publicKimiNetworkErr(err))
		return
	}
	k.finishKimiFromUpstream(c, status, respBody, model)
}

// PhotographyAnalyzeImage 使用 kimi-k2.6 多模态，从摄影角度分析与建议（system + user 引导）；需在 JWT 保护路由下调用。
func (k *KimiController) PhotographyAnalyzeImage(c *gin.Context) {
	t0 := time.Now()
	cfg := config.GetConfig()
	key := sanitizeKimiAPIKey(cfg.Kimi.APIKey)
	if key == "" {
		libx.Err(c, http.StatusInternalServerError, "未配置 kimi.api_key（或环境变量 KIMI_API_KEY / MOONSHOT_API_KEY）", nil)
		return
	}

	model := kimiModelK26

	base := strings.TrimSpace(cfg.Kimi.BaseURL)
	if base == "" {
		base = "https://api.moonshot.ai/v1"
	}

	ensureMultipartParsed(c)
	extra := strings.TrimSpace(c.Query("prompt"))
	if extra == "" {
		extra = strings.TrimSpace(c.PostForm("prompt"))
	}
	focusKey, hasFocus := readAnalyzeFocusDimension(c)
	userLine := buildAnalyzeUserLine(photographyUserBase, extra, focusKey, hasFocus)

	raw, mimeHint, httpStatus, errMsg := readSingleImageBinary(c)
	if httpStatus != 0 {
		libx.Err(c, httpStatus, errMsg, nil)
		return
	}
	log.Printf("photography/analyze: read_image %d bytes in %s", len(raw), time.Since(t0))

	tCompress := time.Now()
	dataURL, imgMeta, err := imageBinaryToDataURL(raw, mimeHint)
	raw = nil
	if err != nil {
		libx.Err(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	log.Printf("photography/analyze: compress meta=%v in %s", imgMeta, time.Since(tCompress))

	stream := wantsPhotographyAnalyzeStream(c)
	if stream {
		k.photographyAnalyzeStream(c, model, base, key, userLine, dataURL, imgMeta, extra, focusKey, hasFocus)
		log.Printf("photography/analyze: stream_handoff focus=%v total_prep=%s", hasFocus, time.Since(t0))
		return
	}

	userContent := buildKimiUserContent(userLine, []string{dataURL}, nil)
	msgs := []map[string]any{
		{"role": "system", "content": photographyAnalyzeJSONSystemPrompt(hasFocus, focusKey)},
		{"role": "user", "content": userContent},
	}
	tKimi := time.Now()
	expectFocus := ""
	if hasFocus {
		expectFocus = focusKey
	}
	payload, err := k.requestPhotographyAnalyzeJSON(c.Request.Context(), model, base, key, msgs, expectFocus)
	log.Printf("photography/analyze: kimi round-trip %s focus=%v err=%v", time.Since(tKimi), hasFocus, err)
	if err != nil {
		if err.Error() == "kimi_auth_failed" {
			libx.Err(c, http.StatusUnauthorized,
				"Kimi 鉴权失败：密钥与 base_url 需同属一国别开放平台", nil)
			return
		}
		if k.respondKimiGateBusy(c, err) {
			return
		}
		if strings.HasPrefix(err.Error(), "kimi_status_") {
			libx.Err(c, http.StatusBadGateway, "Kimi 返回错误", err)
			return
		}
		log.Printf("kimi upstream error: %v", err)
		if strings.Contains(err.Error(), "模型返回") || strings.Contains(err.Error(), "缺少") ||
			strings.Contains(err.Error(), "dimension_scores") || strings.Contains(err.Error(), "improvement_tips") {
			libx.Err(c, http.StatusBadGateway, "模型未返回合法 JSON，请重试", err)
			return
		}
		libx.Err(c, http.StatusBadGateway, "调用 Kimi 失败", publicKimiNetworkErr(err))
		return
	}
	k.respondPhotographyAnalyzeJSON(c, payload, model, imgMeta, extra, expectFocus)
	log.Printf("photography/analyze: done total=%s", time.Since(t0))
}

func clampScoreInt(x int) int {
	if x < 0 {
		return 0
	}
	if x > 100 {
		return 100
	}
	return x
}

func extractJSONFromModelText(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if block := extractMarkdownJSONCodeBlock(s); block != "" {
		if obj := extractJSONObject(block); obj != "" {
			return obj
		}
		return block
	}
	if obj := extractJSONObject(s); obj != "" {
		return obj
	}
	return s
}

func extractMarkdownJSONCodeBlock(s string) string {
	if !strings.HasPrefix(s, "```") {
		return ""
	}
	lines := strings.Split(s, "\n")
	var b strings.Builder
	inBlock := false
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "```") {
			if !inBlock {
				inBlock = true
				continue
			}
			break
		}
		if inBlock {
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	return strings.TrimSpace(b.String())
}

// extractJSONObject 从混杂文本中截取首个完整 JSON 对象；截断时返回已扫描前缀供报错。
func extractJSONObject(s string) string {
	start := strings.Index(s, "{")
	if start < 0 {
		return ""
	}
	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(s); i++ {
		ch := s[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return strings.TrimSpace(s[start : i+1])
			}
		}
	}
	return strings.TrimSpace(s[start:])
}

// PhotographyScoreImage 与 analyze-image 同四维量表，返回扁平字段（composition/color/exposure/content_score）。
func (k *KimiController) PhotographyScoreImage(c *gin.Context) {
	cfg := config.GetConfig()
	key := sanitizeKimiAPIKey(cfg.Kimi.APIKey)
	if key == "" {
		libx.Err(c, http.StatusInternalServerError, "未配置 kimi.api_key（或环境变量 KIMI_API_KEY / MOONSHOT_API_KEY）", nil)
		return
	}

	model := kimiModelK26
	base := strings.TrimSpace(cfg.Kimi.BaseURL)
	if base == "" {
		base = "https://api.moonshot.ai/v1"
	}

	ensureMultipartParsed(c)
	extra := strings.TrimSpace(c.Query("prompt"))
	if extra == "" {
		extra = strings.TrimSpace(c.PostForm("prompt"))
	}
	focusKey, hasFocus := readAnalyzeFocusDimension(c)
	userLine := buildAnalyzeUserLine(photographyScoreUserBase, extra, focusKey, hasFocus)

	raw, mimeHint, httpStatus, errMsg := readSingleImageBinary(c)
	if httpStatus != 0 {
		libx.Err(c, httpStatus, errMsg, nil)
		return
	}

	dataURL, imgMeta, err := imageBinaryToDataURL(raw, mimeHint)
	raw = nil
	if err != nil {
		libx.Err(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	userContent := buildKimiUserContent(userLine, []string{dataURL}, nil)
	msgs := []map[string]any{
		{"role": "system", "content": photographyAnalyzeJSONSystemPrompt(hasFocus, focusKey)},
		{"role": "user", "content": userContent},
	}
	expectFocus := ""
	if hasFocus {
		expectFocus = focusKey
	}
	payload, err := k.requestPhotographyAnalyzeJSON(c.Request.Context(), model, base, key, msgs, expectFocus)
	if err != nil {
		if err.Error() == "kimi_auth_failed" {
			libx.Err(c, http.StatusUnauthorized,
				"Kimi 鉴权失败：密钥与 base_url 需同属一国别开放平台", nil)
			return
		}
		if k.respondKimiGateBusy(c, err) {
			return
		}
		if strings.HasPrefix(err.Error(), "kimi_status_") {
			libx.Err(c, http.StatusBadGateway, "Kimi 返回错误", err)
			return
		}
		if strings.Contains(err.Error(), "模型返回") || strings.Contains(err.Error(), "缺少") ||
			strings.Contains(err.Error(), "dimension_scores") || strings.Contains(err.Error(), "improvement_tips") {
			libx.Err(c, http.StatusBadGateway, "模型未返回合法四维 JSON，请重试", err)
			return
		}
		log.Printf("kimi upstream error: %v", err)
		libx.Err(c, http.StatusBadGateway, "调用 Kimi 失败", publicKimiNetworkErr(err))
		return
	}
	k.respondPhotographyScoreFlat(c, payload, model, imgMeta, extra, expectFocus)
}

func (k *KimiController) Generate(c *gin.Context) {
	var body kimiGenerateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		libx.Err(c, http.StatusBadRequest, "参数无效", err)
		return
	}
	text := body.userText()
	hasMedia := len(body.Images) > 0 || len(body.Videos) > 0
	if text == "" && !hasMedia {
		libx.Err(c, http.StatusBadRequest, "请提供 text/prompt，或与 images/videos 一并用于多模态", nil)
		return
	}

	cfg := config.GetConfig()
	key := sanitizeKimiAPIKey(cfg.Kimi.APIKey)
	if key == "" {
		libx.Err(c, http.StatusInternalServerError, "未配置 kimi.api_key（或环境变量 KIMI_API_KEY / MOONSHOT_API_KEY）", nil)
		return
	}

	model := strings.TrimSpace(body.Model)
	if model == "" {
		model = strings.TrimSpace(cfg.Kimi.Model)
	}
	if model == "" {
		model = kimiModelK26
	}

	base := strings.TrimSpace(cfg.Kimi.BaseURL)
	if base == "" {
		base = "https://api.moonshot.ai/v1"
	}
	base = strings.TrimRight(base, "/")

	userContent := buildKimiUserContent(text, body.Images, body.Videos)

	msgs := []map[string]any{{"role": "user", "content": userContent}}

	status, respBody, err := k.postKimiChat(c.Request.Context(), model, base, key, msgs, nil)
	if err != nil {
		if k.respondKimiGateBusy(c, err) {
			return
		}
		log.Printf("kimi upstream error: %v", err)
		libx.Err(c, http.StatusBadGateway, "调用 Kimi 失败", publicKimiNetworkErr(err))
		return
	}

	k.finishKimiFromUpstream(c, status, respBody, model)
}

func parseKimiChatResponse(b []byte) (string, error) {
	var root struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(b, &root); err != nil {
		return "", err
	}
	if root.Error != nil && root.Error.Message != "" {
		return "", fmt.Errorf("%s", root.Error.Message)
	}
	if len(root.Choices) == 0 {
		return "", fmt.Errorf("choices 为空")
	}
	s := strings.TrimSpace(root.Choices[0].Message.Content)
	if s == "" {
		return "", fmt.Errorf("模型未返回文本")
	}
	return s, nil
}
