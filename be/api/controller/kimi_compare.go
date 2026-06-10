package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"go-service-starter/config"
	"go-service-starter/core/libx"

	"github.com/gin-gonic/gin"
)

const photographyCompareRubricID = "photo_compare_v1"
const photographyCompareMaxTokens = 3000

var photographyCompareSystemPrompt = strings.TrimSpace(`
你是专业摄影选片顾问。用户会上传多张候选照片（附图按顺序编号，从第 1 张起）。
请用同一量表对每张照片从五个维度打分（0–100 整数），横向对比后给出排序与最佳推荐。

【五维评分】
- composition（构图）：主体、留白、平衡、引导线
- color（色彩）：色调、和谐、饱和度、情绪
- exposure（曝光）：层次、高光暗部、过曝欠曝
- sharpness（清晰度）：对焦、细节、噪点与伪影
- creativity（创意性）：视角、瞬间、表现力与辨识度

【输出】仅输出一个 JSON 对象，勿 Markdown、勿 JSON 外文字：
{
  "best_index": <int，0 起，综合最佳>,
  "best_reason": "<string，为何选这张，2–4 句>",
  "summary": "<string，整体选片结论 1–2 句>",
  "photos": [
    {
      "index": <int，与附图顺序一致，从 0 起>,
      "dimension_scores": {
        "composition": <int 0–100>,
        "color": <int 0–100>,
        "exposure": <int 0–100>,
        "sharpness": <int 0–100>,
        "creativity": <int 0–100>
      },
      "pros": "<string，优点>",
      "cons": "<string，不足>"
    }
  ]
}

【硬性要求】
- photos 数组长度必须等于附图张数；index 与附图顺序一一对应。
- dimension_scores 五项均为整数；pros/cons 各一句，与分数同向。
- best_index 必须是 photos 中 overall 最高者（服务端会复算排序，你须自洽）。
- 默认简体中文。
`)

type compareDimensionScores struct {
	Composition int `json:"composition"`
	Color       int `json:"color"`
	Exposure    int `json:"exposure"`
	Sharpness   int `json:"sharpness"`
	Creativity  int `json:"creativity"`
}

type comparePhotoItem struct {
	Index           int                  `json:"index"`
	DimensionScores compareDimensionScores `json:"dimension_scores"`
	Pros            string               `json:"pros"`
	Cons            string               `json:"cons"`
}

type compareModelJSON struct {
	BestIndex  int                `json:"best_index"`
	BestReason string             `json:"best_reason"`
	Summary    string             `json:"summary"`
	Photos     []comparePhotoItem `json:"photos"`
}

func weightedCompareOverall(s compareDimensionScores) int {
	v := 0.25*float64(s.Composition) + 0.20*float64(s.Color) +
		0.25*float64(s.Exposure) + 0.20*float64(s.Sharpness) + 0.10*float64(s.Creativity)
	return clampScoreInt(int(v + 0.5))
}

func parseCompareModelJSON(jsonStr string, imageCount int) (compareModelJSON, error) {
	var p compareModelJSON
	jsonStr = strings.TrimSpace(jsonStr)
	if err := json.Unmarshal([]byte(jsonStr), &p); err != nil {
		return p, err
	}
	if len(p.Photos) != imageCount {
		return p, fmt.Errorf("photos 数量 %d 与上传图片数 %d 不一致", len(p.Photos), imageCount)
	}
	if strings.TrimSpace(p.BestReason) == "" {
		return p, fmt.Errorf("缺少 best_reason")
	}
	for _, ph := range p.Photos {
		if ph.Index < 0 || ph.Index >= imageCount {
			return p, fmt.Errorf("photos[%d].index 超出范围", ph.Index)
		}
		for _, pair := range []struct {
			name string
			v    int
		}{
			{"composition", ph.DimensionScores.Composition},
			{"color", ph.DimensionScores.Color},
			{"exposure", ph.DimensionScores.Exposure},
			{"sharpness", ph.DimensionScores.Sharpness},
			{"creativity", ph.DimensionScores.Creativity},
		} {
			if pair.v < 0 || pair.v > 100 {
				return p, fmt.Errorf("photos[%d].%s 须在 0–100", ph.Index, pair.name)
			}
		}
	}
	return p, nil
}

func (k *KimiController) photoCompareMethodHint(c *gin.Context) {
	libx.Err(c, http.StatusMethodNotAllowed,
		"请使用 POST（需先登录；Bearer token）；multipart 字段 images 或 image（至少 2 张，最多 8 张）。路径：POST /api/kimi/photography/compare-images",
		nil)
}

// PhotographyCompareImages 多图质量对比与最佳推荐（产品 2.3）。
func (k *KimiController) PhotographyCompareImages(c *gin.Context) {
	cfg := config.GetConfig()
	key := sanitizeKimiAPIKey(cfg.Kimi.APIKey)
	if key == "" {
		libx.Err(c, http.StatusInternalServerError, "未配置 kimi.api_key", nil)
		return
	}

	images, httpStatus, errMsg := readMultipleImageBinaries(c)
	if httpStatus != 0 {
		libx.Err(c, httpStatus, errMsg, nil)
		return
	}

	model := kimiModelK26
	base := strings.TrimSpace(cfg.Kimi.BaseURL)
	if base == "" {
		base = "https://api.moonshot.ai/v1"
	}

	urls := make([]string, len(images))
	metaList := make([]map[string]any, len(images))
	for i, img := range images {
		urls[i] = img.DataURL
		metaList[i] = img.Meta
	}

	userLine := fmt.Sprintf("共 %d 张候选照片，请按系统要求横向对比、打分、排序并推荐最佳一张。", len(images))
	userContent := buildKimiUserContent(userLine, urls, nil)
	msgs := []map[string]any{
		{"role": "system", "content": photographyCompareSystemPrompt},
		{"role": "user", "content": userContent},
	}

	extras := kimiK26ChatExtras(photographyCompareMaxTokens)
	extras["response_format"] = map[string]string{"type": "json_object"}

	status, respBody, err := k.postKimiChat(c.Request.Context(), model, base, key, msgs, extras)
	if err != nil {
		if k.respondKimiGateBusy(c, err) {
			return
		}
		log.Printf("kimi compare upstream error: %v", err)
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
	payload, err := parseCompareModelJSON(extractJSONFromModelText(text), len(images))
	if err != nil {
		libx.Err(c, http.StatusBadGateway, "模型未返回合法 JSON，请重试", err)
		return
	}

	type rankedPhoto struct {
		Index           int                  `json:"index"`
		OverallScore    int                  `json:"overall_score"`
		DimensionScores compareDimensionScores `json:"dimension_scores"`
		Pros            string               `json:"pros"`
		Cons            string               `json:"cons"`
		Image           map[string]any       `json:"image,omitempty"`
	}

	ranked := make([]rankedPhoto, 0, len(payload.Photos))
	for _, ph := range payload.Photos {
		ranked = append(ranked, rankedPhoto{
			Index:           ph.Index,
			OverallScore:    weightedCompareOverall(ph.DimensionScores),
			DimensionScores: ph.DimensionScores,
			Pros:            strings.TrimSpace(ph.Pros),
			Cons:            strings.TrimSpace(ph.Cons),
			Image:           metaList[ph.Index],
		})
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].OverallScore > ranked[j].OverallScore
	})

	ranking := make([]int, len(ranked))
	for i, r := range ranked {
		ranking[i] = r.Index
	}

	bestIdx := ranked[0].Index

	libx.Ok(c, "ok", gin.H{
		"rubric_id":      photographyCompareRubricID,
		"model":          model,
		"image_count":    len(images),
		"best_index":     bestIdx,
		"best_reason":    strings.TrimSpace(payload.BestReason),
		"summary":        strings.TrimSpace(payload.Summary),
		"ranking":        ranking,
		"overall_weights": gin.H{
			"composition": 0.25,
			"color":       0.20,
			"exposure":    0.25,
			"sharpness":   0.20,
			"creativity":  0.10,
		},
		"photos": ranked,
	})
}
