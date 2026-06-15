package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"go-service-starter/config"
	"go-service-starter/core/libx"

	"github.com/gin-gonic/gin"
)

const photographyAnalyzeRubricV2 = "photography_analyze_v2"
const photographyAnalyzeJSONMaxTokensDefault = 3000

func photographyAnalyzeMaxTokens() int {
	n := config.GetConfig().Kimi.AnalyzeMaxTokens
	if n <= 0 {
		n = config.GetConfig().Kimi.ScoreMaxTokens
	}
	if n <= 0 {
		n = photographyAnalyzeJSONMaxTokensDefault
	}
	if n < 1200 {
		n = 1200
	}
	return n
}

// photographyAnalyzeDimensionScores 与产品 2.1 一致：构图 / 色彩 / 曝光 / 拍摄内容
type photographyAnalyzeDimensionScores struct {
	Composition int `json:"composition"`
	Color       int `json:"color"`
	Exposure    int `json:"exposure"`
	Content     int `json:"content"`
}

type photographyAnalyzeDimensionNotes struct {
	Composition string `json:"composition"`
	Color       string `json:"color"`
	Exposure    string `json:"exposure"`
	Content     string `json:"content"`
}

type photographyAnalyzeModelJSON struct {
	DimensionScores     photographyAnalyzeDimensionScores `json:"dimension_scores"`
	DimensionNotes      photographyAnalyzeDimensionNotes  `json:"dimension_notes"`
	OverallAnalysis     string                            `json:"overall_analysis"`
	ImprovementTips     []string                          `json:"improvement_tips"`
	FocusedDimension    string                            `json:"focused_dimension"`
	FocusedDeepAnalysis string                            `json:"focused_deep_analysis"`
}

var analyzeDimensionLabels = map[string]string{
	"composition": "构图",
	"color":       "色彩",
	"exposure":    "曝光",
	"content":     "拍摄内容",
}

func normalizeAnalyzeFocusDimension(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	aliases := map[string]string{
		"composition": "composition",
		"color":       "color",
		"exposure":    "exposure",
		"content":     "content",
		"构图":          "composition",
		"色彩":          "color",
		"曝光":          "exposure",
		"内容识别":        "content",
		"内容":          "content",
		"拍摄内容":        "content",
	}
	if v, ok := aliases[raw]; ok {
		return v, true
	}
	if v, ok := aliases[strings.ToLower(raw)]; ok {
		return v, true
	}
	return "", false
}

// ensureMultipartParsed 供 form-data 字段（focus_dimension 等）在读图前可用。
func ensureMultipartParsed(c *gin.Context) {
	ct := c.ContentType()
	if !strings.Contains(ct, "multipart/form-data") {
		return
	}
	_ = c.Request.ParseMultipartForm(maxImageUploadBytes)
}

func readAnalyzeFocusDimension(c *gin.Context) (focus string, hasFocus bool) {
	ensureMultipartParsed(c)
	for _, src := range []string{
		strings.TrimSpace(c.Query("focus_dimension")),
		strings.TrimSpace(c.PostForm("focus_dimension")),
		strings.TrimSpace(c.Query("dimension")),
		strings.TrimSpace(c.PostForm("dimension")),
	} {
		if src == "" {
			continue
		}
		if v, ok := normalizeAnalyzeFocusDimension(src); ok {
			return v, true
		}
	}
	return "", false
}

func photographyAnalyzeJSONSystemPrompt(hasFocus bool, focusKey string) string {
	focusBlock := ""
	if hasFocus {
		label := analyzeDimensionLabels[focusKey]
		focusBlock = fmt.Sprintf(`
【深入维度】用户选定「%s」（%s），须额外输出 focused_deep_analysis：
- 篇幅 200–500 字，明显长于 dimension_notes.%s（全局仅一句）；
- 含画面证据、分项诊断、优先级改进动作；
- focused_dimension 填 "%s"。`, label, focusKey, focusKey, focusKey)
	} else {
		focusBlock = `
【深入维度】用户未选定深入维度：focused_dimension 填 ""，focused_deep_analysis 填 ""。`
	}

	return strings.TrimSpace(`
你是专业摄影分析师。针对用户上传的已拍照片，从四个维度评分并给出整体与分项评价。
仅依据画面可见信息；看不清处写明「信息不足」并相应降分。

【四维定义】每项 0–100 整数；dimension_notes 各一句，与分数同向。
- composition（构图）：主体位置、留白、平衡、引导线、层次与裁切；
- color（色彩）：色调、饱和度、搭配、白平衡与情绪；
- exposure（曝光）：明暗层次、高光/暗部、过曝或欠曝；
- content（拍摄内容）：主体是否明确、场景类型、关键元素、干扰物、主题表达力；

【输出】仅输出一个 JSON 对象，勿 Markdown、勿 JSON 外文字：
{
  "dimension_scores": {
    "composition": <int 0–100>,
    "color": <int 0–100>,
    "exposure": <int 0–100>,
    "content": <int 0–100>
  },
  "dimension_notes": {
    "composition": "<string，一句>",
    "color": "<string，一句>",
    "exposure": "<string，一句>",
    "content": "<string，一句>"
  },
  "overall_analysis": "<string，200–400 字，综合四维的整体评价与改进方向>",
  "improvement_tips": ["<string，3–5 条可执行建议>"],
  "focused_dimension": "<string，composition|color|exposure|content 或空串>",
  "focused_deep_analysis": "<string，深入分析正文；无深入维度时为空串>"
}
` + focusBlock + `
【硬性要求】
- improvement_tips 3–5 条；语气友善、具体；默认简体中文。`)
}

func photographyAnalyzeStreamSystemPrompt(hasFocus bool, focusKey string) string {
	base := strings.TrimSpace(`
你是专业摄影分析师。针对附图输出 Markdown（简体中文）：

## 四维评分
| 维度 | 分数 | 简评 |
| 构图 | 0–100 | 一句 |
| 色彩 | 0–100 | 一句 |
| 曝光 | 0–100 | 一句 |
| 拍摄内容 | 0–100 | 一句 |

## 整体分析
200–400 字，综合四维评价与改进方向。

## 改进建议
3–5 条要点列表。`)

	if hasFocus {
		label := analyzeDimensionLabels[focusKey]
		base += fmt.Sprintf(`

## 深入分析：%s
针对「%s」维度单独展开 200–500 字：画面证据、分项诊断、可执行改进（须比整体分析更细更深）。`, label, label)
	}
	return base
}

func parsePhotographyAnalyzeModelJSON(jsonStr string, expectFocus string) (photographyAnalyzeModelJSON, error) {
	var p photographyAnalyzeModelJSON
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return p, fmt.Errorf("模型返回空 JSON")
	}
	if err := json.Unmarshal([]byte(jsonStr), &p); err != nil {
		return p, err
	}
	for _, pair := range []struct {
		name string
		score int
		note  string
	}{
		{"composition", p.DimensionScores.Composition, p.DimensionNotes.Composition},
		{"color", p.DimensionScores.Color, p.DimensionNotes.Color},
		{"exposure", p.DimensionScores.Exposure, p.DimensionNotes.Exposure},
		{"content", p.DimensionScores.Content, p.DimensionNotes.Content},
	} {
		if pair.score < 0 || pair.score > 100 {
			return p, fmt.Errorf("dimension_scores.%s 须在 0–100", pair.name)
		}
		if strings.TrimSpace(pair.note) == "" {
			return p, fmt.Errorf("缺少 dimension_notes.%s", pair.name)
		}
	}
	if strings.TrimSpace(p.OverallAnalysis) == "" {
		return p, fmt.Errorf("缺少 overall_analysis")
	}
	if len(p.ImprovementTips) < 3 {
		return p, fmt.Errorf("improvement_tips 不足 3 条")
	}
	if expectFocus != "" {
		if strings.TrimSpace(p.FocusedDeepAnalysis) == "" {
			return p, fmt.Errorf("已请求深入维度 %s，但 focused_deep_analysis 为空", expectFocus)
		}
	}
	return p, nil
}

func parsePhotographyAnalyzeFromKimiBody(respBody []byte, expectFocus string) (photographyAnalyzeModelJSON, error) {
	text, err := parseKimiChatResponse(respBody)
	if err != nil {
		return photographyAnalyzeModelJSON{}, err
	}
	raw := extractJSONFromModelText(text)
	if raw == "" {
		preview := text
		if len(preview) > 160 {
			preview = preview[:160] + "…"
		}
		return photographyAnalyzeModelJSON{}, fmt.Errorf("模型返回空 JSON（原文预览: %q）", preview)
	}
	return parsePhotographyAnalyzeModelJSON(raw, expectFocus)
}

func (k *KimiController) requestPhotographyAnalyzeJSON(
	ctx context.Context, model, base, key string, msgs []map[string]any, expectFocus string,
) (photographyAnalyzeModelJSON, error) {
	maxTok := photographyAnalyzeMaxTokens()
	extras := kimiK26ChatExtras(maxTok)
	extras["response_format"] = map[string]string{"type": "json_object"}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			log.Printf("photography/analyze: JSON 解析失败，重试第 %d 次", attempt+1)
		}
		status, respBody, err := k.postKimiChat(ctx, model, base, key, msgs, extras)
		if err != nil {
			return photographyAnalyzeModelJSON{}, err
		}
		if status != http.StatusOK {
			if status == http.StatusUnauthorized {
				return photographyAnalyzeModelJSON{}, fmt.Errorf("kimi_auth_failed")
			}
			return photographyAnalyzeModelJSON{}, fmt.Errorf("kimi_status_%d: %s", status, string(respBody))
		}
		payload, parseErr := parsePhotographyAnalyzeFromKimiBody(respBody, expectFocus)
		if parseErr == nil {
			return payload, nil
		}
		lastErr = parseErr
		log.Printf("photography/analyze: parse_err=%v max_tokens=%d", parseErr, maxTok)
	}
	return photographyAnalyzeModelJSON{}, lastErr
}

func weightedAnalyzeOverallScore(s photographyAnalyzeDimensionScores) int {
	v := 0.30*float64(s.Composition) + 0.25*float64(s.Color) +
		0.25*float64(s.Exposure) + 0.20*float64(s.Content)
	return clampScoreInt(int(v + 0.5))
}

func (k *KimiController) respondPhotographyAnalyzeJSON(
	c *gin.Context, payload photographyAnalyzeModelJSON, model string,
	imgMeta map[string]any, userPrompt string, expectFocus string,
) {
	overall := weightedAnalyzeOverallScore(payload.DimensionScores)
	data := gin.H{
		"rubric_id":         photographyAnalyzeRubricV2,
		"model":             model,
		"format":            "json",
		"image":             imgMeta,
		"overall_score":     overall,
		"overall_weights": gin.H{
			"composition": 0.30,
			"color":       0.25,
			"exposure":    0.25,
			"content":     0.20,
		},
		"dimension_scores":  payload.DimensionScores,
		"dimension_notes":   payload.DimensionNotes,
		"dimension_labels":  analyzeDimensionLabels,
		"overall_analysis":  strings.TrimSpace(payload.OverallAnalysis),
		"improvement_tips":  payload.ImprovementTips,
		"focused_dimension": strings.TrimSpace(payload.FocusedDimension),
		"focused_deep_analysis": strings.TrimSpace(payload.FocusedDeepAnalysis),
	}
	if userPrompt != "" {
		data["prompt"] = userPrompt
	}
	if expectFocus != "" {
		data["focus_dimension"] = expectFocus
		data["focus_dimension_label"] = analyzeDimensionLabels[expectFocus]
	}
	libx.Ok(c, "ok", data)
}

// respondPhotographyScoreFlat score-image 扁平四维响应（与 analyze-image 同量表，字段兼容 Apifox）。
func (k *KimiController) respondPhotographyScoreFlat(
	c *gin.Context, payload photographyAnalyzeModelJSON, model string,
	imgMeta map[string]any, userPrompt string, expectFocus string,
) {
	s := payload.DimensionScores
	overall := weightedAnalyzeOverallScore(s)
	notes := map[string]string{
		"composition": strings.TrimSpace(payload.DimensionNotes.Composition),
		"color":       strings.TrimSpace(payload.DimensionNotes.Color),
		"exposure":    strings.TrimSpace(payload.DimensionNotes.Exposure),
		"content":     strings.TrimSpace(payload.DimensionNotes.Content),
	}
	lowDiff := s.Color >= 46 && s.Color <= 54 && s.Composition >= 46 && s.Composition <= 54 &&
		s.Exposure >= 46 && s.Exposure <= 54 && s.Content >= 46 && s.Content <= 54
	disp := gin.H{"low_differentiation": lowDiff}
	if lowDiff {
		disp["hint"] = "四项子分均落在 46–54，易为模型「安全中段」；请对照 dimension_notes 人工复核。"
	}

	data := gin.H{
		"rubric_id":         photographyAnalyzeRubricV2,
		"model":             model,
		"format":            "flat",
		"image":             imgMeta,
		"composition_score": s.Composition,
		"color_score":       s.Color,
		"exposure_score":    s.Exposure,
		"content_score":     s.Content,
		"dimension_scores": gin.H{
			"composition": s.Composition,
			"color":       s.Color,
			"exposure":    s.Exposure,
			"content":     s.Content,
		},
		"overall_score":     overall,
		"overall_weights": gin.H{
			"composition": 0.30,
			"color":       0.25,
			"exposure":    0.25,
			"content":     0.20,
		},
		"dimension_notes":  notes,
		"dimension_labels": analyzeDimensionLabels,
		"text_analysis":    strings.TrimSpace(payload.OverallAnalysis),
		"improvement_tips": payload.ImprovementTips,
		"score_dispersion": disp,
	}
	if userPrompt != "" {
		data["prompt"] = userPrompt
	}
	if expectFocus != "" {
		data["focus_dimension"] = expectFocus
		data["focus_dimension_label"] = analyzeDimensionLabels[expectFocus]
		data["focused_dimension"] = expectFocus
		data["focused_deep_analysis"] = strings.TrimSpace(payload.FocusedDeepAnalysis)
	}
	libx.Ok(c, "ok", data)
}

func buildAnalyzeUserLine(base string, extra string, focusKey string, hasFocus bool) string {
	line := base
	if extra != "" {
		line += "\n\n【用户补充关注点】" + extra
	}
	if hasFocus {
		line += fmt.Sprintf("\n\n【深入分析请求】请对「%s」维度输出 focused_deep_analysis，须比 overall_analysis 更详细更深入。",
			analyzeDimensionLabels[focusKey])
	} else {
		line += "\n\n【分析模式】仅输出四维评分、整体分析与改进建议；focused_dimension 与 focused_deep_analysis 留空。"
	}
	return line
}
