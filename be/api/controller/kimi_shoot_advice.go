package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"go-service-starter/config"
	"go-service-starter/core/libx"

	"github.com/gin-gonic/gin"
)

const photographyShootAdviceRubricID = "shoot_advice_v1"

const photographyShootAdviceMaxTokens = 2500

// photographyShootAdviceSystemPrompt AI 拍摄建议（产品 2.2）
var photographyShootAdviceSystemPrompt = strings.TrimSpace(`
你是专业摄影前期指导。用户会上传一张「当前所处场景」的参考照片，并说明打算拍摄的主体。
请仅依据画面可见信息分析，不臆造场景中不存在的元素；看不清处须在对应字段注明「信息不足」。

【任务】给出可执行的前期拍摄建议，帮助用户在按下快门前确定机位、焦段与用光策略。

【输出】仅输出一个 JSON 对象，勿 Markdown、勿 JSON 外文字。键必须完全一致：
{
  "scene_summary": "<string，1–2 句概括场景类型与第一印象>",
  "scene_analysis": {
    "light": "<string，光线方向、强弱、反差、色温倾向>",
    "space": "<string，空间结构、纵深、视线引导>",
    "background": "<string，背景元素与干扰/可利用点>",
    "atmosphere": "<string，环境氛围与情绪>"
  },
  "subject_plan": "<string，结合用户描述的主体，说明最佳表现方式与在画面中的位置建议>",
  "camera_position": {
    "description": "<string，推荐机位与站位，含高度与左右关系>",
    "angle": "<string，如正面/侧面/俯拍/仰拍/斜侧等>",
    "distance": "<string，建议与主体的大致距离，如 2–3 米>",
    "annotations": [
      {
        "id": 1,
        "area": "<string，画面区域，如左上/中央/前景>",
        "label": "<string，标注说明，如建议站位、相机朝向>",
        "hint": "<string，箭头或框线语义，供示意图渲染>"
      }
    ]
  },
  "focal_length": {
    "range": "<string，如 35–50mm（全画幅等效）>",
    "category": "<string，广角/标准/中长焦/长焦 之一>",
    "reason": "<string，结合场景与主体说明理由>"
  },
  "shooting_tips": ["<string，可执行要点>", "..."],
  "alternatives": [
    {
      "style": "<string，风格或创意名称>",
      "description": "<string，不同机位/焦段/氛围的备选拍法>"
    }
  ],
  "summary": "<string，1 句话总结最优先执行的一点>"
}

【硬性要求】
- shooting_tips 必须 3–5 条，涵盖构图、对焦/景深、光线利用等。
- annotations 至少 2 条，用于机位示意图标注（文字描述即可）。
- alternatives 0–2 条；若无合适备选可输出空数组 []。
- 语气友善、具体，面向初学者；参数用「可尝试」表述。
- 默认简体中文。
`)

type photographyShootAdviceModelJSON struct {
	SceneSummary   string `json:"scene_summary"`
	SceneAnalysis  struct {
		Light      string `json:"light"`
		Space      string `json:"space"`
		Background string `json:"background"`
		Atmosphere string `json:"atmosphere"`
	} `json:"scene_analysis"`
	SubjectPlan     string `json:"subject_plan"`
	CameraPosition  struct {
		Description string `json:"description"`
		Angle       string `json:"angle"`
		Distance    string `json:"distance"`
		Annotations []struct {
			ID    int    `json:"id"`
			Area  string `json:"area"`
			Label string `json:"label"`
			Hint  string `json:"hint"`
		} `json:"annotations"`
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
}

// photographyShootAdviceStream 流式输出 Markdown 分节建议（便于边生成边展示）。
var photographyShootAdviceStreamSystemPrompt = strings.TrimSpace(`
你是专业摄影前期指导。用户上传当前场景参考图并说明拍摄主体。
请用 Markdown 分节输出拍摄建议（简体中文），包含：
## 场景解读（光线、空间、背景、氛围）
## 主体表现建议
## 推荐机位（角度、距离、站位描述）
## 机位示意图标注（列表：区域 + 标注 + 箭头/框线含义，至少 2 条）
## 焦段建议（范围、类型、理由）
## 拍摄要点（3–5 条列表）
## 备选方案（0–2 种不同风格，可选）
## 总结（一句话）
仅依据画面可见信息，不臆造；语气友善、可执行。
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
		_ = writeSSE(c, "error", gin.H{"message": publicKimiStreamErr(err)})
	}
}
