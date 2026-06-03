package controller

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const photographyAnalyzeMaxTokens = 1024

const photographyAnalyzeRubricID = "photography_analyze_v1"

func wantsPhotographyAnalyzeStream(c *gin.Context) bool {
	s := strings.TrimSpace(strings.ToLower(c.Query("stream")))
	if s == "false" || s == "0" || s == "no" {
		return false
	}
	return true
}

func writeSSE(c *gin.Context, event string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, b); err != nil {
		return err
	}
	if f, ok := c.Writer.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}

// postKimiChatStream 请求上游 stream=true，将增量内容以 SSE event: chunk 转发给客户端。
func (k *KimiController) postKimiChatStream(
	ctx context.Context,
	model, base, apiKey string,
	messages []map[string]any,
	payloadExtras map[string]any,
	c *gin.Context,
) error {
	payload := map[string]any{
		"model":    model,
		"messages": messages,
		"stream":   true,
	}
	for ek, ev := range payloadExtras {
		payload[ek] = ev
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	upCtx, upCancel := kimiUpstreamContext(ctx)
	defer upCancel()

	apiURL := strings.TrimRight(strings.TrimSpace(base), "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(upCtx, http.MethodPost, apiURL, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upstream %d: %s", resp.StatusCode, string(body))
	}

	var full strings.Builder
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		if data == "[DONE]" {
			break
		}
		delta, streamErr := extractKimiStreamDelta(data)
		if streamErr != nil {
			return streamErr
		}
		if delta == "" {
			continue
		}
		full.WriteString(delta)
		if err := writeSSE(c, "chunk", gin.H{"delta": delta}); err != nil {
			return err
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return writeSSE(c, "done", gin.H{
		"text":  full.String(),
		"model": model,
	})
}

func extractKimiStreamDelta(data string) (string, error) {
	var chunk struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
			FinishReason *string `json:"finish_reason"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return "", nil
	}
	if chunk.Error != nil && chunk.Error.Message != "" {
		return "", fmt.Errorf("%s", chunk.Error.Message)
	}
	if len(chunk.Choices) == 0 {
		return "", nil
	}
	return chunk.Choices[0].Delta.Content, nil
}

func writePhotographyStreamHeaders(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
}

// writePhotographyStreamMeta 与拍摄建议接口一致的 SSE meta（format=markdown + 图片压缩信息）。
func writePhotographyStreamMeta(c *gin.Context, rubricID, model string, imgMeta map[string]any, extra map[string]any) error {
	meta := gin.H{
		"rubric_id": rubricID,
		"model":     model,
		"streaming": true,
		"format":    "markdown",
		"image":     imgMeta,
	}
	for k, v := range extra {
		if v != nil && v != "" {
			meta[k] = v
		}
	}
	return writeSSE(c, "meta", meta)
}

// photographyAnalyzeSystemPrompt 摄影分析（流式/非流式共用，控制篇幅以降低超时）。
var photographyAnalyzeSystemPrompt = strings.TrimSpace(`
你是摄影指导。针对用户上传的已拍照片，用 Markdown 输出简洁中文分析：
## 概览（1–2句）
## 构图与曝光
## 色彩与技术
## 改进建议（3～4条可执行要点）
## 总结（1句）
仅依据画面可见信息；先肯定亮点再指出问题；不要冗长复述。
`)

// 兼容旧名
var photographyAnalyzeStreamSystemPrompt = photographyAnalyzeSystemPrompt

func (k *KimiController) photographyAnalyzeStream(
	c *gin.Context,
	model, base, key string,
	userLine string,
	dataURL string,
	compressMeta map[string]any,
	userPrompt string,
) {
	msgs := []map[string]any{
		{"role": "system", "content": photographyAnalyzeStreamSystemPrompt},
		{"role": "user", "content": buildKimiUserContent(userLine, []string{dataURL}, nil)},
	}
	extras := kimiK26ChatExtras(photographyAnalyzeMaxTokens)

	writePhotographyStreamHeaders(c)
	metaExtra := map[string]any{}
	if strings.TrimSpace(userPrompt) != "" {
		metaExtra["prompt"] = strings.TrimSpace(userPrompt)
	}
	if err := writePhotographyStreamMeta(c, photographyAnalyzeRubricID, model, compressMeta, metaExtra); err != nil {
		return
	}

	if err := k.postKimiChatStream(c.Request.Context(), model, base, key, msgs, extras, c); err != nil {
		_ = writeSSE(c, "error", gin.H{"message": publicKimiStreamErr(err)})
	}
}

func publicKimiStreamErr(err error) string {
	if err == nil {
		return "未知错误"
	}
	if pub := publicKimiNetworkErr(err); pub != nil {
		return pub.Error()
	}
	msg := err.Error()
	if len(msg) > 500 {
		return msg[:500] + "…"
	}
	return msg
}
