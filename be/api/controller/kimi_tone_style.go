package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"go-service-starter/config"
	"go-service-starter/core/libx"

	"github.com/gin-gonic/gin"
)

const photographyToneStyleRubricID = "tone_style_v1"
const photographyToneStyleMaxTokens = 2000

var photographyToneStyleSystemPrompt = strings.TrimSpace(`
你是专业摄影后期调色师。用户用文字描述想要的影调风格（可能附带一张参考照片）。
请理解风格语义，输出可执行的后期参数建议（适用于 Lightroom / 手机修图 / Photoshop Camera Raw 等）。

【输出】仅输出一个 JSON 对象，勿 Markdown、勿 JSON 外文字：
{
  "style_name": "<string，归纳的风格名称>",
  "style_match_summary": "<string，如何理解用户描述，1–2 句>",
  "parameters": {
    "exposure": "<string，如 +0.3 EV>",
    "contrast": "<string，如 -10 或 +15>",
    "highlights": "<string>",
    "shadows": "<string>",
    "whites": "<string>",
    "blacks": "<string>",
    "saturation": "<string>",
    "vibrance": "<string>",
    "temperature": "<string，色温 K 或 暖/冷>",
    "tint": "<string>",
    "hue_adjustments": "<string，可选，如 橙-5 绿+3>",
    "curve": "<string，可选，曲线倾向>",
    "grain": "<string，可选>"
  },
  "adjustment_notes": ["<string，3–5 条操作步骤>"],
  "before_after_description": "<string，描述调整前后画面观感差异，用于文字示意对比>",
  "preview_hints": "<string，如何用现有工具快速预览该风格>"
}

【硬性要求】
- parameters 中 exposure/contrast/saturation/temperature 必填且有具体数值或方向。
- adjustment_notes 3–5 条，按推荐操作顺序。
- before_after_description 须对比「调整前」与「调整后」的明暗、色调、氛围。
- 若附图存在，须结合画面给出更贴合的参数；无附图则按风格常识推断。
- 默认简体中文。
`)

type toneStyleParameters struct {
	Exposure        string `json:"exposure"`
	Contrast        string `json:"contrast"`
	Highlights      string `json:"highlights"`
	Shadows         string `json:"shadows"`
	Whites          string `json:"whites"`
	Blacks          string `json:"blacks"`
	Saturation      string `json:"saturation"`
	Vibrance        string `json:"vibrance"`
	Temperature     string `json:"temperature"`
	Tint            string `json:"tint"`
	HueAdjustments  string `json:"hue_adjustments"`
	Curve           string `json:"curve"`
	Grain           string `json:"grain"`
}

type toneStyleModelJSON struct {
	StyleName              string              `json:"style_name"`
	StyleMatchSummary      string              `json:"style_match_summary"`
	Parameters             toneStyleParameters `json:"parameters"`
	AdjustmentNotes        []string            `json:"adjustment_notes"`
	BeforeAfterDescription string              `json:"before_after_description"`
	PreviewHints           string              `json:"preview_hints"`
}

func parseToneStyleModelJSON(jsonStr string) (toneStyleModelJSON, error) {
	var p toneStyleModelJSON
	jsonStr = strings.TrimSpace(jsonStr)
	if err := json.Unmarshal([]byte(jsonStr), &p); err != nil {
		return p, err
	}
	if strings.TrimSpace(p.StyleName) == "" {
		return p, fmt.Errorf("缺少 style_name")
	}
	if strings.TrimSpace(p.Parameters.Exposure) == "" ||
		strings.TrimSpace(p.Parameters.Contrast) == "" ||
		strings.TrimSpace(p.Parameters.Saturation) == "" ||
		strings.TrimSpace(p.Parameters.Temperature) == "" {
		return p, fmt.Errorf("parameters 缺少必填项")
	}
	if len(p.AdjustmentNotes) < 3 {
		return p, fmt.Errorf("adjustment_notes 不足 3 条")
	}
	if strings.TrimSpace(p.BeforeAfterDescription) == "" {
		return p, fmt.Errorf("缺少 before_after_description")
	}
	return p, nil
}

func (k *KimiController) photoToneStyleMethodHint(c *gin.Context) {
	libx.Err(c, http.StatusMethodNotAllowed,
		"请使用 POST（需先登录）；multipart：style_description（必填）+ 可选 image/file 参考图。路径：POST /api/kimi/photography/tone-style",
		nil)
}

// PhotographyToneStyle 影调风格后期参数建议（产品 2.4）。
func (k *KimiController) PhotographyToneStyle(c *gin.Context) {
	cfg := config.GetConfig()
	key := sanitizeKimiAPIKey(cfg.Kimi.APIKey)
	if key == "" {
		libx.Err(c, http.StatusInternalServerError, "未配置 kimi.api_key", nil)
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImageUploadBytes)
	if err := c.Request.ParseMultipartForm(maxImageUploadBytes); err != nil {
		libx.Err(c, http.StatusBadRequest, "multipart 解析失败: "+err.Error(), nil)
		return
	}

	styleDesc, httpStatus, errMsg := readToneStyleDescription(c)
	if httpStatus != 0 {
		libx.Err(c, httpStatus, errMsg, nil)
		return
	}

	var dataURL string
	var imgMeta map[string]any
	hasImage := false
	if form := c.Request.MultipartForm; form != nil {
		var fhName string
		for _, key := range []string{"image", "file"} {
			if fhs, ok := form.File[key]; ok && len(fhs) > 0 {
				fhName = key
				fh := fhs[0]
				src, err := fh.Open()
				if err != nil {
					libx.Err(c, http.StatusBadRequest, "无法打开参考图", nil)
					return
				}
				b, err := io.ReadAll(io.LimitReader(src, maxImageUploadBytes+1))
				src.Close()
				if err != nil {
					libx.Err(c, http.StatusBadRequest, "读取参考图失败", nil)
					return
				}
				if len(b) > 0 && len(b) <= maxImageUploadBytes {
					url, meta, err := imageBinaryToDataURL(b, fh.Header.Get("Content-Type"))
					if err != nil {
						libx.Err(c, http.StatusBadRequest, err.Error(), nil)
						return
					}
					dataURL = url
					imgMeta = meta
					hasImage = true
				}
				_ = fhName
				break
			}
		}
	}

	model := kimiModelK26
	base := strings.TrimSpace(cfg.Kimi.BaseURL)
	if base == "" {
		base = "https://api.moonshot.ai/v1"
	}

	userLine := "用户想要的影调风格描述：" + styleDesc
	if hasImage {
		userLine += "\n附图为用户提供的参考照片，请结合画面色调给出更贴合的参数。"
	} else {
		userLine += "\n用户未提供参考图，请根据风格描述给出通用参数建议。"
	}

	var userContent any
	if hasImage {
		userContent = buildKimiUserContent(userLine, []string{dataURL}, nil)
	} else {
		userContent = userLine
	}

	msgs := []map[string]any{
		{"role": "system", "content": photographyToneStyleSystemPrompt},
		{"role": "user", "content": userContent},
	}

	extras := kimiK26ChatExtras(photographyToneStyleMaxTokens)
	extras["response_format"] = map[string]string{"type": "json_object"}

	status, respBody, err := k.postKimiChat(c.Request.Context(), model, base, key, msgs, extras)
	if err != nil {
		if k.respondKimiGateBusy(c, err) {
			return
		}
		log.Printf("kimi tone-style upstream error: %v", err)
		libx.Err(c, http.StatusBadGateway, "调用 Kimi 失败", publicKimiNetworkErr(err))
		return
	}

	if status != http.StatusOK {
		if status == http.StatusUnauthorized {
			libx.Err(c, http.StatusUnauthorized, "Kimi 鉴权失败，请检查 api_key 与 base_url", nil)
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
	payload, err := parseToneStyleModelJSON(extractJSONFromModelText(text))
	if err != nil {
		libx.Err(c, http.StatusBadGateway, "模型未返回合法 JSON，请重试", err)
		return
	}

	resp := gin.H{
		"rubric_id": photographyToneStyleRubricID,
		"model":     model,
		"style_description": styleDesc,
		"has_reference_image": hasImage,
		"style":     payload,
	}
	if hasImage {
		resp["image"] = imgMeta
	}
	libx.Ok(c, "ok", resp)
}
