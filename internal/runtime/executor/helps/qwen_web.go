package helps

import (
	"bufio"
	"encoding/json"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/auth/qwenweb"
	"github.com/tidwall/gjson"
)

const QwenWebUpstreamModel = "qwen3.7-plus"

// QwenWebRejectUnsupported runs before and after protocol translation so tools
// cannot disappear silently in a translator that does not recognize them.
func QwenWebRejectUnsupported(body []byte) error {
	var walk func(gjson.Result) bool
	walk = func(v gjson.Result) bool {
		if !v.IsObject() && !v.IsArray() {
			return false
		}
		bad := false
		v.ForEach(func(k, val gjson.Result) bool {
			switch k.String() {
			case "tools", "functions", "tool_calls":
				if !val.IsArray() || len(val.Array()) > 0 {
					bad = true
				}
			case "tool_choice", "function_call", "tool_call_id", "functionCall", "functionResponse":
				if val.Exists() && val.Type != gjson.Null {
					bad = true
				}
			case "role":
				if val.String() == "tool" || val.String() == "function" {
					bad = true
				}
			case "type":
				if val.String() == "tool_use" || val.String() == "tool_result" || val.String() == "function_call" || val.String() == "function_call_output" {
					bad = true
				}
			}
			if !bad {
				bad = walk(val)
			}
			return !bad
		})
		return bad
	}
	if walk(gjson.ParseBytes(body)) {
		return &qwenweb.Error{Code: 400, Message: "qwen-web does not support tool calling or tool history"}
	}
	for _, field := range []string{"thinking", "reasoning", "reasoning_effort", "thinkingConfig", "generationConfig.thinkingConfig"} {
		if gjson.GetBytes(body, field).Exists() {
			return &qwenweb.Error{Code: 400, Message: "qwen-web does not support configurable reasoning in this version"}
		}
	}
	return nil
}

// QwenWebPrompt encodes the complete ordered transcript. Each API request uses
// a new upstream chat; no hidden server conversation or shared history exists.
func QwenWebPrompt(body []byte) (string, error) {
	messages := gjson.GetBytes(body, "messages")
	if !messages.IsArray() || len(messages.Array()) == 0 {
		return "", &qwenweb.Error{Code: 400, Message: "messages must be a non-empty array"}
	}
	type turn struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	turns := []turn{}
	for _, message := range messages.Array() {
		role := message.Get("role").String()
		switch role {
		case "user", "assistant", "system", "developer":
		default:
			return "", &qwenweb.Error{Code: 400, Message: "unsupported message role for qwen-web"}
		}
		content := message.Get("content")
		text := ""
		if content.Type == gjson.String {
			text = content.String()
		} else if content.IsArray() {
			for _, part := range content.Array() {
				if part.Get("type").String() != "text" || part.Get("text").Type != gjson.String {
					return "", &qwenweb.Error{Code: 400, Message: "qwen-web chat currently accepts text only"}
				}
				text += part.Get("text").String() + "\n"
			}
		} else {
			return "", &qwenweb.Error{Code: 400, Message: "message content must be text"}
		}
		if message.Get("audio").Exists() {
			return "", &qwenweb.Error{Code: 400, Message: "qwen-web chat currently accepts text only"}
		}
		turns = append(turns, turn{role, text})
	}
	if turns[len(turns)-1].Role != "user" {
		return "", &qwenweb.Error{Code: 400, Message: "qwen-web requires a final user message"}
	}
	if len(turns) == 1 {
		return turns[0].Content, nil
	}
	raw, _ := json.Marshal(turns)
	return "Continue the conversation represented by this ordered JSON transcript. Respect the system/developer instructions and answer the final user message. Return only the assistant response, not the transcript.\n" + string(raw), nil
}

// QwenWebUpstreamImageModel is the image model the web frontend selects through
// the message metadata while the chat model remains the text model.
const QwenWebUpstreamImageModel = "qwen-image-2.0-pro"

// QwenWebPayload builds the payload observed from the real web frontend.
// Text and image turns share /api/v2/chat/completions; the turn kind is carried
// by chat_type, sub_chat_type and the message metadata.
func QwenWebPayload(chatID, prompt string, image bool, ratio string) map[string]any {
	kind := "t2t"
	if image {
		kind = "t2i"
	}
	features := map[string]any{
		"thinking_enabled": false, "output_schema": "phase", "research_mode": "normal",
		"auto_thinking": false, "auto_search": false, "code_interpreter": false,
		"function_calling": false, "plugins_enabled": false,
	}
	meta := map[string]any{"subChatType": kind}
	if image {
		features["thinking_mode"] = "Fast"
		features["auto_search"] = true
		meta["model"] = QwenWebUpstreamImageModel
		meta["size"] = ratio
	} else {
		features["thinking_mode"] = "Disabled"
		features["thinking_format"] = "summary"
	}
	stamp := time.Now().Unix()
	payload := map[string]any{
		// The web frontend sends both spellings; keep them consistent.
		"model": QwenWebUpstreamModel, "chatId": chatID, "chat_id": chatID,
		"stream": true, "version": "2.1", "incremental_output": true,
		"chat_mode": "normal", "parent_id": nil, "parentId": "", "timestamp": stamp,
		"messages": []any{map[string]any{
			"id": nil, "fid": uuid.NewString(), "parentId": nil, "parent_id": nil,
			"childrenIds": []string{uuid.NewString()}, "role": "user", "content": prompt,
			"user_action": "chat", "files": []any{}, "timestamp": stamp,
			"models": []string{QwenWebUpstreamModel}, "model": "",
			"chat_type": kind, "sub_chat_type": kind,
			"feature_config": features, "extra": map[string]any{"meta": meta},
		}},
	}
	if image {
		payload["size"] = ratio
	}
	return payload
}

type QwenWebDelta struct {
	Text, Reasoning string
	Phase           string
	FinishReason    string
	Raw             []byte
}

// ConsumeQwenWebSSE parses the real Qwen Web event stream. Control frames such
// as response.created and keep-alive response.info are ignored. Image URLs are
// delivered as the delta content of the image_gen phase.
// Upstream error text is deliberately never reflected to the caller.
func ConsumeQwenWebSSE(r io.Reader, emit func(QwenWebDelta) error) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(nil, 4<<20)
	done := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			done = true
			break
		}
		if data == "" {
			continue
		}
		if !json.Valid([]byte(data)) {
			return &qwenweb.Error{Code: 502, Message: "invalid Qwen Web stream event"}
		}
		if msg := qwenWebRiskControlMessage(data); msg != "" {
			return &qwenweb.Error{Code: 502, Message: msg}
		}
		obj := gjson.Parse(data)
		if (obj.Get("success").Exists() && !obj.Get("success").Bool()) || obj.Get("error").Exists() {
			return &qwenweb.Error{Code: 502, Message: "Qwen Web generation failed"}
		}
		if obj.Get("data.choices").Exists() {
			obj = obj.Get("data")
		}
		delta := obj.Get("choices.0.delta")
		if !delta.Exists() {
			// response.created / response.info carry no user-visible content.
			continue
		}
		if delta.Get("error").Exists() || delta.Get("status").String() == "failed" {
			return &qwenweb.Error{Code: 502, Message: "Qwen Web generation failed"}
		}
		phase := delta.Get("phase").String()
		text := delta.Get("content").String()
		reasoning := delta.Get("reasoning_content").String()
		if reasoning == "" {
			reasoning = delta.Get("reasoning").String()
		}
		if reasoning == "" && delta.Get("extra.summary_thought.content").Exists() {
			reasoning = strings.Join(gjsonStringSlice(delta.Get("extra.summary_thought.content")), "")
		}
		if phase == "thinking_summary" {
			reasoning = text
			text = ""
		}
		finish := obj.Get("choices.0.finish_reason")
		if finish.String() == "error" || finish.String() == "content_filter" {
			return &qwenweb.Error{Code: 502, Message: "Qwen Web did not complete the answer"}
		}
		if delta.Get("status").String() == "finished" || delta.Get("status").String() == "completed" || (finish.Type == gjson.String && finish.String() != "") {
			done = true
		}
		if err := emit(QwenWebDelta{
			Text:         text,
			Reasoning:    reasoning,
			Phase:        phase,
			Raw:          []byte(data),
			FinishReason: finish.String(),
		}); err != nil {
			return err
		}
	}
	if scanner.Err() != nil {
		return &qwenweb.Error{Code: 502, Message: "Qwen Web stream interrupted"}
	}
	if !done {
		return &qwenweb.Error{Code: 502, Message: "Qwen Web stream ended without completion"}
	}
	return nil
}

var qwenWebRiskControlMarkers = []string{"FAIL_SYS_USER_VALIDATE", "RGV587_ERROR", "_____tmd_____", "action=captcha", "/punish"}

// qwenWebRiskControlMessage detects Ali risk-control responses. These arrive as
// HTTP 200 with a JSON body instead of an event stream, so they must be reported
// as verification failures rather than empty completions.
func qwenWebRiskControlMessage(data string) string {
	for _, marker := range qwenWebRiskControlMarkers {
		if strings.Contains(data, marker) {
			return "Qwen Web requires browser verification for this session or network; complete the check on chat.qwen.ai and import a fresh session"
		}
	}
	return ""
}

func gjsonStringSlice(value gjson.Result) []string {
	out := []string{}
	for _, item := range value.Array() {
		out = append(out, item.String())
	}
	return out
}

// QwenWebImages extracts generated image URLs from one stream event.
// The real web stream returns the URL as the delta content of the image_gen
// phase. Only HTTPS URLs from Qwen's own CDN hosts are accepted so a malicious
// upstream cannot make the proxy emit or later fetch arbitrary URLs.
// The server never downloads these URLs.
func QwenWebImages(raw []byte) []string {
	out := []string{}
	seen := map[string]bool{}
	add := func(s string, explicit bool) {
		u, err := url.Parse(s)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
			return
		}
		if !qwenWebImageHost(u.Host) && !explicit {
			return
		}
		ext := strings.ToLower(path.Ext(u.Path))
		image := ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" || ext == ".gif"
		if (explicit || image) && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	if !json.Valid(raw) {
		return out
	}
	delta := gjson.GetBytes(raw, "choices.0.delta")
	if !delta.Exists() {
		return out
	}
	// Image generation delivers the URL directly in the content field. Even for
	// the image_gen phase the host must still be a Qwen CDN, so a hostile or
	// compromised upstream cannot steer the caller to an arbitrary URL.
	if delta.Get("phase").String() == "image_gen" {
		if content := strings.TrimSpace(delta.Get("content").String()); content != "" {
			add(content, false)
		}
	}
	for _, value := range delta.Get("extra").Array() {
		if value.Type == gjson.String {
			add(value.String(), false)
		}
	}
	return out
}

var qwenWebImageHosts = []string{"cdn.qwenlm.ai", "qwenlm.ai", "wanx.alicdn.com", "img.alicdn.com"}

func qwenWebImageHost(host string) bool {
	host = strings.ToLower(host)
	for _, allowed := range qwenWebImageHosts {
		if host == allowed || strings.HasSuffix(host, "."+allowed) {
			return true
		}
	}
	return false
}

// QwenWebDetailImages extracts image URLs from a chat detail payload. The chat
// history stores assistant content_list entries, and the image_gen entry holds
// the generated URL.
func QwenWebDetailImages(raw []byte) []string {
	if !json.Valid(raw) {
		return nil
	}
	out := []string{}
	seen := map[string]bool{}
	add := func(s string) {
		u, err := url.Parse(strings.TrimSpace(s))
		if err != nil || u.Scheme != "https" || !qwenWebImageHost(u.Host) {
			return
		}
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	var walk func(gjson.Result, int)
	walk = func(v gjson.Result, depth int) {
		if depth > 32 {
			return
		}
		if v.Get("phase").String() == "image_gen" {
			if content := strings.TrimSpace(v.Get("content").String()); content != "" {
				add(content)
			}
		}
		if v.IsObject() || v.IsArray() {
			v.ForEach(func(_ gjson.Result, val gjson.Result) bool {
				walk(val, depth+1)
				return true
			})
		}
	}
	walk(gjson.ParseBytes(raw), 0)
	return out
}
