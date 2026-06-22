package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"go-service-starter/config"
	"go-service-starter/core/kimigate"
	"go-service-starter/core/libx"
	"go-service-starter/domain"

	"github.com/gin-gonic/gin"
)

const photographyShootAdviceRubricID = "shoot_advice_v1"

const photographyShootAdviceMaxTokens = 2500

// photographyShootAdviceSystemPrompt AI 拍摄建议：场景图四维简评 + 前期拍摄建议（产品 2.2）
var photographyShootAdviceSystemPrompt = strings.TrimSpace(`
你是专业摄影前期指导与场景评估师。用户会上传一张「当前所处场景」的参考照片，并说明打算拍摄的主体。
请仅依据画面可见信息分析，不臆造场景中不存在的元素；看不清处须在对应 dimension_notes 写明「信息不足」。

【任务】
1. 结合用户主体描述，给出可执行的前期拍摄建议（机位、焦段、要点）。


【输出】仅输出一个 JSON 对象，勿 Markdown、勿 JSON 外文字。键必须完全一致：
{
  "scene_summary": "<string，1–2 句概括场景与主体关系>",
  "dimension_notes": {
    "composition": "<string，一句，关于场景的评价>",
    "color": "<string，一句，关于色彩的评价>",
    "exposure": "<string，一句，关于曝光的评价>",
    "content": "<string，一句，关于内容的评价>"
  },
  "subject_plan": "<string，结合用户主体，说明最佳表现方式>",
  "camera_position": {
    "description": "<string，推荐机位与站位>",
    "angle": "<string，如正面/侧面/俯拍/仰拍等>",
    "distance": "<string，建议距离>"
  },
  "focal_length": {
    "range": "<string，如 35–50mm>",
    "category": "<string，广角/标准/中长焦/长焦>",
    "reason": "<string>"
  },
  "shooting_tips": ["<string，可执行要点>", "..."],
  "alternatives": [{ "style": "<string>", "description": "<string>" }],
  "summary": "<string，1 句话：优先改善哪一维或最先执行的一点>"
}

【硬性要求】
- shooting_tips 3–5 条；alternatives 0–2 条，无则 []。
- 语气友善、具体；默认简体中文。
`)

type photographyShootAdviceDimensionNotes struct {
	Composition string `json:"composition"`
	Color       string `json:"color"`
	Exposure    string `json:"exposure"`
	Content     string `json:"content"`
}

type photographyShootAdviceModelJSON struct {
	SceneSummary    string                             `json:"scene_summary"`
	DimensionNotes  photographyShootAdviceDimensionNotes `json:"dimension_notes"`
	SubjectPlan      string                              `json:"subject_plan"`
	CameraPosition struct {
		Description string `json:"description"`
		Angle       string `json:"angle"`
		Distance    string `json:"distance"`
	} `json:"camera_position"`
	FocalLength struct {
		Range    string `json:"range"`
		Category string `json:"category"`
		Reason   string `json:"reason"`
	} `json:"focal_length"`
	ShootingTips  []string `json:"shooting_tips"`
	Alternatives  []struct {
		Style       string `json:"style"`
		Description string `json:"description"`
	} `json:"alternatives"`
	Summary string `json:"summary"`
}

func readShootAdviceSubject(c *gin.Context) (subject string, httpStatus int, errMsg string) {
	subject = strings.TrimSpace(c.Query("subject"))
	if subject == "" {
		subject = strings.TrimSpace(c.PostForm("subject"))
	}
	if subject == "" {
		return "", http.StatusBadRequest, "请提供拍摄主体描述（multipart 字段 subject 或查询参数 subject）"
	}
	if len(subject) > 500 {
		return "", http.StatusBadRequest, "subject 描述过长（最多 500 字）"
	}
	return subject, 0, ""
}

func parsePhotographyShootAdviceModelJSON(jsonStr string) (photographyShootAdviceModelJSON, error) {
	var p photographyShootAdviceModelJSON
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return p, fmt.Errorf("模型输出为空")
	}
	if err := json.Unmarshal([]byte(jsonStr), &p); err != nil {
		return p, err
	}
	if strings.TrimSpace(p.SceneSummary) == "" {
		return p, fmt.Errorf("缺少 scene_summary")
	}
	for _, pair := range []struct {
		name string
		note string
	}{
		{"composition", p.DimensionNotes.Composition},
		{"color", p.DimensionNotes.Color},
		{"exposure", p.DimensionNotes.Exposure},
		{"content", p.DimensionNotes.Content},
	} {
		if strings.TrimSpace(pair.note) == "" {
			return p, fmt.Errorf("缺少 dimension_notes.%s", pair.name)
		}
	}
	if len(p.ShootingTips) < 3 {
		return p, fmt.Errorf("shooting_tips 不足 3 条")
	}
	return p, nil
}

func (k *KimiController) photoShootAdviceMethodHint(c *gin.Context) {
	libx.Err(c, http.StatusMethodNotAllowed,
		"请使用 POST（需先登录；Bearer token）；multipart：image/file + subject（拍摄主体描述）。可选 ?stream=true 为 SSE 流式 Markdown。路径：POST /api/kimi/photography/shoot-advice",
		nil)
}

// PhotographyShootAdvice AI 拍摄建议：场景图 + 主体描述 → 机位/焦段/要点等 JSON（产品 2.2）。
func (k *KimiController) PhotographyShootAdvice(c *gin.Context) {
	cfg := config.GetConfig()
	key := sanitizeKimiAPIKey(cfg.Kimi.APIKey)
	if key == "" {
		libx.Err(c, http.StatusInternalServerError, "未配置 kimi.api_key（或环境变量 KIMI_API_KEY / MOONSHOT_API_KEY）", nil)
		return
	}

	subject, httpStatus, errMsg := readShootAdviceSubject(c)
	if httpStatus != 0 {
		libx.Err(c, httpStatus, errMsg, nil)
		return
	}

	model := kimiModelK26
	base := strings.TrimSpace(cfg.Kimi.BaseURL)
	if base == "" {
		base = "https://api.moonshot.ai/v1"
	}

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

	userLine := fmt.Sprintf("用户计划拍摄的物体/主体：%s\n请根据附图场景，按系统要求输出拍摄前期建议 JSON。", subject)

	if wantsPhotographyAnalyzeStream(c) {
		k.photographyShootAdviceStream(c, model, base, key, userLine, dataURL, imgMeta, subject)
		return
	}

	userContent := buildKimiUserContent(userLine, []string{dataURL}, nil)
	msgs := []map[string]any{
		{"role": "system", "content": photographyShootAdviceSystemPrompt},
		{"role": "user", "content": userContent},
	}
	extras := kimiK26ChatExtras(photographyShootAdviceMaxTokens)
	extras["response_format"] = map[string]string{"type": "json_object"}

	status, respBody, err := k.postKimiChat(c.Request.Context(), model, base, key, msgs, extras)
	if err != nil {
		if k.respondKimiGateBusy(c, err) {
			return
		}
		log.Printf("kimi shoot-advice upstream error: %v", err)
		libx.Err(c, http.StatusBadGateway, "调用 Kimi 失败", publicKimiNetworkErr(err))
		return
	}
	k.finishKimiShootAdviceFromUpstream(c, status, respBody, model, subject, imgMeta)
}

func (k *KimiController) finishKimiShootAdviceFromUpstream(
	c *gin.Context, status int, respBody []byte, model, subject string, imgMeta map[string]any,
) {
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
	payload, err := parsePhotographyShootAdviceModelJSON(jsonStr)
	if err != nil {
		libx.Err(c, http.StatusBadGateway, "模型未返回合法 JSON，请重试", err)
		return
	}
	libx.Ok(c, "ok", gin.H{
		"rubric_id": photographyShootAdviceRubricID,
		"model":     model,
		"subject":   subject,
		"image":     imgMeta,
		"advice":    payload,
	})

	// 保存历史记录
	if k.historyRepo != nil {
		payloadMap := make(map[string]any)
		payloadBytes, _ := json.Marshal(payload)
		json.Unmarshal(payloadBytes, &payloadMap)
		go k.saveAnalysisHistory(libx.Uid(c), domain.AnalysisTypeShootAdvice, subject, "", payloadMap)
	}
}

// photographyShootAdviceStreamSystemPrompt 流式：场景简评 + 拍摄建议。
var photographyShootAdviceStreamSystemPrompt = strings.TrimSpace(`
你是专业摄影前期指导。用户上传当前场景参考图并说明拍摄主体。
请用 Markdown 输出（简体中文）：
## 场景概览
## 场景解读
### 构图
### 色彩
### 曝光
### 内容识别
## 主体表现建议
## 推荐机位与焦段
## 拍摄要点（3–5 条）
## 总结
仅依据画面可见信息；看不清处写明「信息不足」；语气友善、可执行。不要输出机位示意图标注列表。
`)

func (k *KimiController) photographyShootAdviceStream(
	c *gin.Context,
	model, base, key string,
	userLine, dataURL string,
	imgMeta map[string]any,
	subject string,
) {
	msgs := []map[string]any{
		{"role": "system", "content": photographyShootAdviceStreamSystemPrompt},
		{"role": "user", "content": buildKimiUserContent(userLine, []string{dataURL}, nil)},
	}
	extras := kimiK26ChatExtras(photographyShootAdviceMaxTokens)

	writePhotographyStreamHeaders(c)
	_ = writePhotographyStreamMeta(c, photographyShootAdviceRubricID, model, imgMeta, map[string]any{
		"subject": subject,
	})
	if err := k.postKimiChatStream(c.Request.Context(), model, base, key, msgs, extras, c); err != nil {
		if errors.Is(err, kimigate.ErrTooManyConcurrent) {
			c.Header("Retry-After", "30")
			_ = writeSSE(c, "error", gin.H{"message": "AI 服务繁忙，请稍后重试", "code": 503})
			return
		}
		_ = writeSSE(c, "error", gin.H{"message": publicKimiStreamErr(err)})
	}
}
