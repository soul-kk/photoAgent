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
	"math"
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
		"请使用 POST（需先登录；Bearer token）；multipart：image/file；可选 prompt、focus_dimension（构图|色彩|曝光|内容识别 或 composition|color|exposure|content）。默认 stream=false：四维 JSON 评分+整体分析；指定 focus_dimension 时额外返回 focused_deep_analysis。stream=true 为 SSE Markdown。路径：POST /api/kimi/photography/analyze-image",
		nil)
}

func (k *KimiController) photoScoreMethodHint(c *gin.Context) {
	libx.Err(c, http.StatusMethodNotAllowed,
		"请使用 POST（需先登录；Header：Authorization: Bearer <token>）；multipart 字段 image 或 file。路径：POST /api/kimi/photography/score-image 或 POST /api/kimi/score-photo；返回各维度 0–100 分与文字分析 JSON",
		nil)
}

// 单张图片二进制上限（multipart 整请求或原始 body 都会受此上限约束）
const maxImageUploadBytes = 20 << 20

// kimiModelK26 多模态接口固定使用该模型（不使用配置文件里的默认模型占位）
const kimiModelK26 = "kimi-k2.6"

const photographyUserBase = "请根据上述要求，对附图从摄影角度进行分析并给出建议。"

// photographyScoreSystemPrompt 固定量表 + 仅 JSON 输出，便于多图可比与复现（该模型 temperature 由接口固定为 1）。
var photographyScoreSystemPrompt = strings.TrimSpace(`
你是摄影赛事初审评委：同一量表、跨所有图片可比；禁止讨好用户、禁止「安全牌」中间分。
所有分数为 0–100 的整数；只依据画面可见证据；看不清则在对应 dimension_notes 写明「信息不足」并该维度倾向给 35–45（不要默认 50）。

【分数—评语一致性（必须满足）】
dimension_notes 对该维度的褒贬，必须与该维度分数同向：评语以批评为主、且未写实质亮点时，该维度分通常不得高于 48；若写明「严重/显著/尽失/严重模糊/伪影严重」等，该维度须落在 15–40。
若某维度评语同时含明显优缺点，分数应落在 40–65 并体现主次，不得四项一律塞 48–52。
禁止在 exposure/content 指出严重画面缺陷时，仍将对应分数拉到 45 以上，除非明确解释是可辨认的创意表达。

【锚点刻度——每维打分前先选档再微调个位数，避免扎堆整十】
每维优先用 17、23、31、38、44、52、58、63、71、78、86 等非克隆数，除非该维确实恰为典型「典型中段」才可使用 50 附近。

1) composition_score：主体、线条、留白、平衡、裁切与视线引导。
   - 15–25：严重失衡、致命裁切、主体淹没。
   - 35–45：完整但不讲究；缺少引导与节奏。
   - 55–65：稳妥可用，有一点设计感或一笔败笔。
   - 72–82：结构清晰，引导明确，有章法。
2) color_score：白平衡、色调与题材关系；和谐度与饱和度控制。
   - 15–25：明显偏色/脏灰/过饱和干扰主题。
   - 35–45：尚可但单调、层次弱或轻微偏色。
   - 55–65：整体舒服、略有平庸或轻微遗憾。
   - 72–82：色彩推动情绪与层次，克制有力。
   - 88+：极少给出；需画面与题材高度匹配且具辨识度。
3) exposure_score：明暗层次、高光/暗部细节、是否过曝或欠曝、主体亮度是否合理。
   - 15–25：明显过曝、死黑或主体亮度严重不可辨。
   - 35–45：曝光可看但层次不足，主体或背景有明显明暗问题。
   - 55–65：曝光基本稳妥，有少量高光/暗部遗憾。
   - 72–82：层次清楚，明暗关系服务主体。
4) content_score：主体是否清晰、画面信息是否可读、关键元素与干扰物、内容表达是否明确。
   - 15–25：主体难辨，关键内容被遮挡或干扰严重。
   - 35–45：能识别内容但表达松散，干扰物或信息不足明显。
   - 55–65：主体和场景关系基本清楚，但记忆点一般。
   - 72–82：内容明确，有可感知的主题关系或瞬间表达。

【输出】仅输出一个 JSON 对象，勿 Markdown、勿 JSON 外文字。
键必须完全一致：
{
  "composition_score": <int>,
  "color_score": <int>,
  "exposure_score": <int>,
  "content_score": <int>,
  "text_analysis": "<string，中文：总体评价 + 可执行改进建议，分段可用 \\n>",
  "dimension_notes": {
    "composition": "<一句，须与 composition_score 同向>",
    "color": "<一句，须与 color_score 同向>",
    "exposure": "<一句，须与 exposure_score 同向>",
    "content": "<一句，须与 content_score 同向>"
  }
}
不要输出 overall_score，由服务端按固定权重计算。
`)

const photographyScoreUserBase = "请仅依据附图，按系统量表输出四维评分 JSON；dimension_notes 每项一句话；text_analysis 200–500 字为宜。"

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
	extras := kimiK26ChatExtras(photographyAnalyzeJSONMaxTokens)
	extras["response_format"] = map[string]string{"type": "json_object"}

	tKimi := time.Now()
	status, respBody, err := k.postKimiChat(c.Request.Context(), model, base, key, msgs, extras)
	log.Printf("photography/analyze: kimi round-trip %s focus=%v err=%v", time.Since(tKimi), hasFocus, err)
	if err != nil {
		if k.respondKimiGateBusy(c, err) {
			return
		}
		log.Printf("kimi upstream error: %v", err)
		libx.Err(c, http.StatusBadGateway, "调用 Kimi 失败", publicKimiNetworkErr(err))
		return
	}
	expectFocus := ""
	if hasFocus {
		expectFocus = focusKey
	}
	k.finishPhotographyAnalyzeJSONFromUpstream(c, status, respBody, model, imgMeta, extra, expectFocus)
	log.Printf("photography/analyze: done total=%s", time.Since(t0))
}

const photographyScoreRubricID = "photography_v2"

type photographyScoreModelJSON struct {
	ColorScore       float64           `json:"color_score"`
	CompositionScore float64           `json:"composition_score"`
	TechniqueScore   *float64          `json:"technique_score"`
	ExposureScore    *float64          `json:"exposure_score"`
	ContentScore     *float64          `json:"content_score"`
	TextAnalysis     string            `json:"text_analysis"`
	DimensionNotes   map[string]string `json:"dimension_notes"`
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

func roundClampScore(f float64) int {
	return clampScoreInt(int(math.Round(f)))
}

func roundClampScorePtr(f *float64) (int, bool) {
	if f == nil {
		return 0, false
	}
	return roundClampScore(*f), true
}

// weightedOverallPhotographyScore 固定权重，保证同一张图的四项子分转为总分时口径一致（与模型可能给出的总分解耦）。
func weightedOverallPhotographyScore(composition, color, exposure, content int) int {
	v := 0.30*float64(composition) + 0.25*float64(color) + 0.25*float64(exposure) + 0.20*float64(content)
	return clampScoreInt(int(math.Round(v)))
}

func extractJSONFromModelText(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
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

func parsePhotographyScoreModelJSON(jsonStr string) (photographyScoreModelJSON, error) {
	var p photographyScoreModelJSON
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return p, fmt.Errorf("模型输出为空")
	}
	if err := json.Unmarshal([]byte(jsonStr), &p); err != nil {
		return p, err
	}
	return p, nil
}

func (k *KimiController) finishKimiScoreFromUpstream(c *gin.Context, status int, respBody []byte, model string) {
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
	text, err := parseKimiChatResponse(respBody)
	if err != nil {
		libx.Err(c, http.StatusBadGateway, "解析 Kimi 响应失败", err)
		return
	}
	jsonStr := extractJSONFromModelText(text)
	payload, err := parsePhotographyScoreModelJSON(jsonStr)
	if err != nil {
		libx.Err(c, http.StatusBadGateway, "模型未返回合法 JSON，请重试或检查提示词冲突", err)
		return
	}
	cScore := roundClampScore(payload.ColorScore)
	coScore := roundClampScore(payload.CompositionScore)
	eScore, hasExposure := roundClampScorePtr(payload.ExposureScore)
	if !hasExposure {
		if fallback, ok := roundClampScorePtr(payload.TechniqueScore); ok {
			eScore = fallback
		}
	}
	contentScore, hasContent := roundClampScorePtr(payload.ContentScore)
	if !hasContent {
		contentScore = weightedOverallPhotographyScore(coScore, cScore, eScore, eScore)
	}

	overall := weightedOverallPhotographyScore(coScore, cScore, eScore, contentScore)
	notes := payload.DimensionNotes
	if notes == nil {
		notes = map[string]string{}
	}
	if notes["exposure"] == "" && notes["technique"] != "" {
		notes["exposure"] = notes["technique"]
	}
	if notes["content"] == "" {
		notes["content"] = "内容识别维度未返回独立评语，请结合总体分析复核。"
	}
	lowDiff := cScore >= 46 && cScore <= 54 && coScore >= 46 && coScore <= 54 && eScore >= 46 && eScore <= 54 && contentScore >= 46 && contentScore <= 54
	disp := gin.H{"low_differentiation": lowDiff}
	if lowDiff {
		disp["hint"] = "四项子分均落在 46–54，易为模型「安全中段」；请对照 dimension_notes 人工复核。"
	}
	libx.Ok(c, "ok", gin.H{
		"rubric_id":         photographyScoreRubricID,
		"model":             model,
		"composition_score": coScore,
		"color_score":       cScore,
		"exposure_score":    eScore,
		"content_score":     contentScore,
		"technique_score":   eScore,
		"overall_score":     overall,
		"overall_weights":   gin.H{"composition": 0.30, "color": 0.25, "exposure": 0.25, "content": 0.20},
		"text_analysis":     strings.TrimSpace(payload.TextAnalysis),
		"dimension_notes":   notes,
		"score_dispersion":  disp,
	})
}

// PhotographyScoreImage 多模态摄影评分：构图 / 色彩 / 曝光 / 内容识别四项 0–100 整数 + 文字分析，JSON 结构由服务端校验并统一计算 overall。
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

	extra := strings.TrimSpace(c.Query("prompt"))
	if extra == "" {
		extra = strings.TrimSpace(c.PostForm("prompt"))
	}
	userLine := photographyScoreUserBase
	if extra != "" {
		userLine = photographyScoreUserBase + "\n\n【用户补充】 " + extra
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

	userContent := buildKimiUserContent(userLine, []string{dataURL}, nil)
	msgs := []map[string]any{
		{"role": "system", "content": photographyScoreSystemPrompt},
		{"role": "user", "content": userContent},
	}

	scoreMaxTokens := config.GetConfig().Kimi.ScoreMaxTokens
	if scoreMaxTokens <= 0 {
		scoreMaxTokens = 800
	}
	extras := kimiK26ChatExtras(scoreMaxTokens)
	extras["response_format"] = map[string]string{
		"type": "json_object",
	}

	status, respBody, err := k.postKimiChat(c.Request.Context(), model, base, key, msgs, extras)
	if err != nil {
		if k.respondKimiGateBusy(c, err) {
			return
		}
		log.Printf("kimi upstream error: %v", err)
		libx.Err(c, http.StatusBadGateway, "调用 Kimi 失败", publicKimiNetworkErr(err))
		return
	}
	k.finishKimiScoreFromUpstream(c, status, respBody, model)
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
