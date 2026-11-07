package loomy

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Login flows.
//
// Two flows reach the same session:
//
//  1. Phone SMS code: POST /login/phone/sendMsgCode then
//     POST /login/phone/checkCode with the returned msgid.
//  2. WeChat QR: fetch the authorization page to extract the embedded QR uuid,
//     long-poll the WeChat confirmation, then exchange the confirmed `code` for
//     an `rcode` (POST /login/thirdAccount/bind/auth). An account that already
//     bound a phone number completes with /bind/skip; otherwise the operator
//     binds a phone through /bind/sendMsg + /bind/checkCode.
//
// The WeChat callback page is not used. It answers 404 upstream and the
// authorization code arrives through the long-poll instead, which is why this
// implementation can run without a browser window.

// accountEnvelope is the `{base, param}` request wrapper the account service
// expects. The trace id is regenerated per call.
func accountEnvelope(param map[string]any) map[string]any {
	return map[string]any{
		"base": map[string]any{
			"appid":   AppID,
			"modelid": "Web",
			"version": "1.0.0",
			"devid":   "web",
			// Hardcoded by the client even on other platforms; do not "fix".
			"ua":      ClientUA,
			"traceid": NewNonce(),
		},
		"param": param,
	}
}

// postAccount issues one signed account request and unwraps the envelope.
func (c *Client) postAccount(ctx context.Context, path string, param map[string]any) (json.RawMessage, error) {
	serialized, errMarshal := json.Marshal(accountEnvelope(param))
	if errMarshal != nil {
		return nil, errMarshal
	}
	body := string(serialized)
	headers := AuthHeaders(AccessKeyID, AccessKeySecret, http.MethodPost, path, body, time.Now().UTC().Format(http.TimeFormat), NewNonce())

	req, errRequest := http.NewRequestWithContext(ctx, http.MethodPost, c.AccountBaseURL()+path, bytes.NewBufferString(body))
	if errRequest != nil {
		return nil, errRequest
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, errDo := c.httpClient.Do(req)
	if errDo != nil {
		return nil, &Error{Message: fmt.Sprintf("account request failed (%s): %v", path, errDo)}
	}
	defer func() { _ = resp.Body.Close() }()
	raw, errRead := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if errRead != nil {
		return nil, &Error{StatusCode: resp.StatusCode, Message: errRead.Error()}
	}
	env, errEnvelope := ParseEnvelope(raw)
	if errEnvelope != nil {
		if upstream, ok := errEnvelope.(*Error); ok && upstream.StatusCode == 0 {
			upstream.StatusCode = resp.StatusCode
		}
		return nil, errEnvelope
	}
	return env.Data, nil
}

// SendSMSCode asks the account service to deliver a verification code and
// returns the msgid that must accompany the subsequent login.
func (c *Client) SendSMSCode(ctx context.Context, phone string) (string, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return "", &Error{StatusCode: http.StatusBadRequest, Message: "phone number is required"}
	}
	data, err := c.postAccount(ctx, "/login/phone/sendMsgCode", map[string]any{
		"ccode": "86", "phone": phone, "expire": SMSCodeTTLSeconds,
	})
	if err != nil {
		return "", err
	}
	var payload struct {
		MsgID string `json:"msgid"`
	}
	if errUnmarshal := json.Unmarshal(data, &payload); errUnmarshal != nil || strings.TrimSpace(payload.MsgID) == "" {
		return "", &Error{Message: "SMS response has no msgid"}
	}
	return payload.MsgID, nil
}

// LoginWithSMSCode exchanges phone + code + msgid for a session.
func (c *Client) LoginWithSMSCode(ctx context.Context, phone, code, msgID string) (*TokenData, error) {
	phone = strings.TrimSpace(phone)
	code = strings.TrimSpace(code)
	msgID = strings.TrimSpace(msgID)
	if phone == "" || code == "" || msgID == "" {
		return nil, &Error{StatusCode: http.StatusBadRequest, Message: "phone, code and msgid are required"}
	}
	data, err := c.postAccount(ctx, "/login/phone/checkCode", map[string]any{
		"ccode": "86", "phone": phone, "mcode": code, "msgid": msgID, "expire": SessionTTLSeconds,
	})
	if err != nil {
		return nil, err
	}
	token, errToken := tokenFromData(data)
	if errToken != nil {
		return nil, errToken
	}
	token.Phone = phone
	return token, nil
}

// tokenFromData decodes the session/userid pair shared by every login response.
func tokenFromData(data json.RawMessage) (*TokenData, error) {
	var payload struct {
		Session  string `json:"session"`
		UserID   string `json:"userid"`
		Phone    string `json:"phone"`
		Nickname string `json:"nickname"`
	}
	if errUnmarshal := json.Unmarshal(data, &payload); errUnmarshal != nil {
		return nil, &Error{Message: "login response is not a JSON object"}
	}
	session := strings.TrimSpace(payload.Session)
	userID := strings.TrimSpace(payload.UserID)
	if session == "" {
		return nil, &Error{Message: "login response has no session"}
	}
	if userID == "" {
		return nil, &Error{Message: "login response has no userid"}
	}
	return &TokenData{
		Session:  session,
		UserID:   userID,
		Phone:    strings.TrimSpace(payload.Phone),
		Nickname: strings.TrimSpace(payload.Nickname),
	}, nil
}

// WeChatAuthURL is the allow-listed QR authorization page.
func WeChatAuthURL(state string) string {
	return "https://open.weixin.qq.com/connect/qrconnect" +
		"?appid=" + EscapeRFC3986(WeChatAppID) +
		"&redirect_uri=" + EscapeRFC3986(WeChatRedirectURI) +
		"&response_type=code&scope=snsapi_login" +
		"&state=" + EscapeRFC3986(state) +
		"#wechat_redirect"
}

// WeChatQRImageURL is the QR image address for an extracted uuid.
func WeChatQRImageURL(uuid string) string {
	return "https://open.weixin.qq.com/connect/qrcode/" + EscapeRFC3986(uuid)
}

var (
	wechatUUIDPattern    = regexp.MustCompile(`/connect/qrcode/([A-Za-z0-9_\-=+/]+)`)
	wechatPollUUID       = regexp.MustCompile(`l/qrconnect\?uuid=([A-Za-z0-9_\-=+/]+)`)
	wechatErrcodePattern = regexp.MustCompile(`wx_errcode\s*=\s*(\d+)`)
	wechatCodePattern    = regexp.MustCompile(`wx_code\s*=\s*'([^']*)'`)
	wechatUUIDChars      = regexp.MustCompile(`^[A-Za-z0-9_\-=+/]{6,64}$`)
)

const wechatUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// ExtractWeChatUUID pulls the embedded QR uuid out of the authorization page.
// Two patterns are tried because the page exposes it both as the QR image src
// and inside the embedded long-poll snippet.
func ExtractWeChatUUID(html string) string {
	for _, pattern := range []*regexp.Regexp{wechatUUIDPattern, wechatPollUUID} {
		match := pattern.FindStringSubmatch(html)
		if len(match) != 2 {
			continue
		}
		if wechatUUIDChars.MatchString(match[1]) {
			return match[1]
		}
	}
	return ""
}

// FetchWeChatUUID loads the authorization page and extracts the QR uuid.
func (c *Client) FetchWeChatUUID(ctx context.Context, state string) (string, error) {
	req, errRequest := http.NewRequestWithContext(ctx, http.MethodGet, WeChatAuthURL(state), nil)
	if errRequest != nil {
		return "", errRequest
	}
	req.Header.Set("User-Agent", wechatUserAgent)
	req.Header.Set("Referer", "https://open.weixin.qq.com/")
	resp, errDo := c.httpClient.Do(req)
	if errDo != nil {
		return "", &Error{Message: fmt.Sprintf("wechat authorize page failed: %v", errDo)}
	}
	defer func() { _ = resp.Body.Close() }()
	raw, errRead := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if errRead != nil {
		return "", &Error{StatusCode: resp.StatusCode, Message: errRead.Error()}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &Error{StatusCode: resp.StatusCode, Message: "wechat authorize page rejected the request"}
	}
	uuid := ExtractWeChatUUID(string(raw))
	if uuid == "" {
		return "", &Error{Message: "wechat authorize page carries no QR uuid"}
	}
	return uuid, nil
}

// FetchWeChatQRImage downloads the QR image and returns it as a data URL.
func (c *Client) FetchWeChatQRImage(ctx context.Context, uuid string) (string, error) {
	req, errRequest := http.NewRequestWithContext(ctx, http.MethodGet, WeChatQRImageURL(uuid), nil)
	if errRequest != nil {
		return "", errRequest
	}
	req.Header.Set("User-Agent", wechatUserAgent)
	req.Header.Set("Referer", "https://open.weixin.qq.com/")
	resp, errDo := c.httpClient.Do(req)
	if errDo != nil {
		return "", &Error{Message: fmt.Sprintf("wechat QR download failed: %v", errDo)}
	}
	defer func() { _ = resp.Body.Close() }()
	raw, errRead := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if errRead != nil {
		return "", &Error{StatusCode: resp.StatusCode, Message: errRead.Error()}
	}
	mime := detectImageMime(raw)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || mime == "" || len(raw) < 200 {
		return "", &Error{StatusCode: resp.StatusCode, Message: "wechat QR response is not an image"}
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(raw), nil
}

// detectImageMime recognises the formats WeChat has been observed to return.
func detectImageMime(raw []byte) string {
	switch {
	case len(raw) >= 3 && raw[0] == 0xff && raw[1] == 0xd8 && raw[2] == 0xff:
		return "image/jpeg"
	case len(raw) >= 4 && raw[0] == 0x89 && raw[1] == 0x50 && raw[2] == 0x4e && raw[3] == 0x47:
		return "image/png"
	case len(raw) >= 3 && raw[0] == 0x47 && raw[1] == 0x49 && raw[2] == 0x46:
		return "image/gif"
	}
	return ""
}

// WeChatPollStatus is the state machine of the QR long-poll.
//
// The mapping follows the authorization page's own script: 408 waiting, 404
// scanned-awaiting-confirmation, 405 confirmed (the code is in that frame),
// 403 cancelled, 402 expired. Reversing 404 and 405 is the documented failure
// mode this implementation exists to avoid.
type WeChatPollStatus string

const (
	WeChatWaiting   WeChatPollStatus = "waiting"
	WeChatScanned   WeChatPollStatus = "scanned"
	WeChatConfirmed WeChatPollStatus = "confirmed"
	WeChatCancelled WeChatPollStatus = "cancelled"
	WeChatExpired   WeChatPollStatus = "expired"
	WeChatPollError WeChatPollStatus = "error"
)

// WeChatPollResult is one long-poll outcome.
type WeChatPollResult struct {
	Status  WeChatPollStatus
	Code    string
	Errcode string
}

// PollWeChat performs one long-poll round. Transport errors are reported as
// WeChatPollError rather than returned, because a transient failure must not
// abort the flow.
func (c *Client) PollWeChat(ctx context.Context, uuid, lastErrcode string) WeChatPollResult {
	query := "uuid=" + EscapeRFC3986(uuid)
	if lastErrcode != "" {
		query += "&last=" + EscapeRFC3986(lastErrcode)
	}
	query += "&_=" + fmt.Sprint(time.Now().UnixMilli())
	req, errRequest := http.NewRequestWithContext(ctx, http.MethodGet,
		c.WeChatOriginBase()+"/connect/l/qrconnect?"+query, nil)
	if errRequest != nil {
		return WeChatPollResult{Status: WeChatPollError, Errcode: errRequest.Error()}
	}
	req.Header.Set("User-Agent", wechatUserAgent)
	req.Header.Set("Referer", WeChatAuthURL(""))
	client := *c.httpClient
	client.Timeout = WeChatPollTimeout
	resp, errDo := client.Do(req)
	if errDo != nil {
		return WeChatPollResult{Status: WeChatPollError, Errcode: errDo.Error()}
	}
	defer func() { _ = resp.Body.Close() }()
	raw, errRead := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if errRead != nil {
		return WeChatPollResult{Status: WeChatPollError, Errcode: errRead.Error()}
	}
	body := string(raw)
	errcode := submatch(wechatErrcodePattern, body)
	code := submatch(wechatCodePattern, body)
	switch errcode {
	case "405":
		// Confirmed. A 405 without a code is an unexpected shape: keep polling
		// rather than exchanging an empty code.
		if code == "" {
			return WeChatPollResult{Status: WeChatScanned, Errcode: errcode}
		}
		return WeChatPollResult{Status: WeChatConfirmed, Code: code, Errcode: errcode}
	case "404":
		return WeChatPollResult{Status: WeChatScanned, Errcode: errcode}
	case "403":
		return WeChatPollResult{Status: WeChatCancelled, Errcode: errcode}
	case "402":
		return WeChatPollResult{Status: WeChatExpired, Errcode: errcode}
	default:
		return WeChatPollResult{Status: WeChatWaiting, Errcode: errcode}
	}
}

func submatch(pattern *regexp.Regexp, value string) string {
	match := pattern.FindStringSubmatch(value)
	if len(match) != 2 {
		return ""
	}
	return match[1]
}

// WeChatBindAuth is the first-step result of a WeChat authorization.
type WeChatBindAuth struct {
	// Bind is 1 when the iFlytek account already bound a phone number.
	Bind int
	// RCode is the context identifier the remaining steps use.
	RCode    string
	Nickname string
	HeadPic  string
}

// BindWeChat exchanges a WeChat authorization code for a bind context.
func (c *Client) BindWeChat(ctx context.Context, code string) (*WeChatBindAuth, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, &Error{StatusCode: http.StatusBadRequest, Message: "wechat code is required"}
	}
	data, err := c.postAccount(ctx, "/login/thirdAccount/bind/auth", map[string]any{
		"tcode": map[string]any{"code": code},
		"type":  "wx",
	})
	if err != nil {
		return nil, err
	}
	var payload struct {
		Bind     int    `json:"bind"`
		RCode    string `json:"rcode"`
		Nickname string `json:"nickname"`
		HeadPic  string `json:"headpic"`
	}
	if errUnmarshal := json.Unmarshal(data, &payload); errUnmarshal != nil {
		return nil, &Error{Message: "wechat bind response is not a JSON object"}
	}
	rcode := strings.TrimSpace(payload.RCode)
	if rcode == "" {
		return nil, &Error{Message: "wechat bind response has no rcode"}
	}
	bind := 0
	if payload.Bind == 1 {
		bind = 1
	}
	return &WeChatBindAuth{Bind: bind, RCode: rcode, Nickname: payload.Nickname, HeadPic: payload.HeadPic}, nil
}

// WeChatBindSkip completes login for an account that already bound a phone.
func (c *Client) WeChatBindSkip(ctx context.Context, rcode string) (*TokenData, error) {
	data, err := c.postAccount(ctx, "/login/thirdAccount/bind/skip", map[string]any{
		"rcode": strings.TrimSpace(rcode), "expire": SessionTTLSeconds,
	})
	if err != nil {
		return nil, err
	}
	return tokenFromData(data)
}

// WeChatBindSendMsg sends the phone-binding verification code.
func (c *Client) WeChatBindSendMsg(ctx context.Context, rcode, phone string) (string, error) {
	data, err := c.postAccount(ctx, "/login/thirdAccount/bind/sendMsg", map[string]any{
		"rcode": strings.TrimSpace(rcode), "phone": strings.TrimSpace(phone),
		"ccode": "86", "expire": SMSCodeTTLSeconds,
	})
	if err != nil {
		return "", err
	}
	var payload struct {
		MsgID string `json:"msgid"`
	}
	if errUnmarshal := json.Unmarshal(data, &payload); errUnmarshal != nil || strings.TrimSpace(payload.MsgID) == "" {
		return "", &Error{Message: "phone bind response has no msgid"}
	}
	return payload.MsgID, nil
}

// WeChatBindCheckCode verifies the binding code and completes login.
func (c *Client) WeChatBindCheckCode(ctx context.Context, rcode, code, msgID string) (*TokenData, error) {
	data, err := c.postAccount(ctx, "/login/thirdAccount/bind/checkCode", map[string]any{
		"rcode": strings.TrimSpace(rcode), "mcode": strings.TrimSpace(code),
		"msgid": strings.TrimSpace(msgID), "expire": SessionTTLSeconds,
	})
	if err != nil {
		return nil, err
	}
	return tokenFromData(data)
}

// FetchModels reads the business model catalogue, keeping chat entries only.
//
// The filter is `type == "chat"`. Input modalities are not a filter: several
// chat models accept images, which does not make them image-generation models.
func (c *Client) FetchModels(ctx context.Context, session string) ([]Model, error) {
	data, err := c.fetchModelList(ctx, session)
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	if errUnmarshal := json.Unmarshal(data, &rows); errUnmarshal != nil {
		var wrapper struct {
			Data []map[string]any `json:"data"`
		}
		if errWrapper := json.Unmarshal(data, &wrapper); errWrapper != nil {
			return nil, &Error{Message: "model list is not an array"}
		}
		rows = wrapper.Data
	}
	models := make([]Model, 0, len(rows))
	for _, row := range rows {
		model, ok := modelFromRow(row)
		if !ok {
			continue
		}
		models = append(models, model)
	}
	return models, nil
}

// fetchModelList accepts both the business envelope used by older gateways and
// the OpenAI-style {object:"list",data:[...]} response used by the current one.
func (c *Client) fetchModelList(ctx context.Context, session string) (json.RawMessage, error) {
	if strings.TrimSpace(session) == "" {
		return nil, &Error{StatusCode: http.StatusUnauthorized, Message: "loomy session is empty"}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.APIBaseURL()+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("token", session)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &Error{Message: err.Error()}
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, &Error{StatusCode: resp.StatusCode, Message: err.Error()}
	}
	var wrapper struct {
		Code *string         `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, &Error{StatusCode: resp.StatusCode, Message: "model list is not a JSON object"}
	}
	if wrapper.Code != nil {
		env, err := ParseEnvelope(raw)
		if err != nil {
			return nil, err
		}
		return env.Data, nil
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, &Error{StatusCode: resp.StatusCode, Message: "model list request failed"}
	}
	if len(wrapper.Data) == 0 || string(wrapper.Data) == "null" {
		return nil, &Error{StatusCode: resp.StatusCode, Message: "model list has no data"}
	}
	return wrapper.Data, nil
}

// modelFromRow maps one catalogue row, reporting false for non-chat entries.
func modelFromRow(row map[string]any) (Model, bool) {
	id, _ := row["id"].(string)
	id = strings.TrimSpace(id)
	if id == "" {
		return Model{}, false
	}
	if kind, _ := row["type"].(string); kind != "chat" {
		return Model{}, false
	}
	model := Model{ID: id}
	if name, _ := row["name"].(string); strings.TrimSpace(name) != "" {
		model.DisplayName = strings.TrimSpace(name)
	} else {
		model.DisplayName = id
	}
	if contextLength, ok := numberValue(row["context_length"]); ok && contextLength > 0 {
		model.ContextLimit = int(contextLength)
	}
	if raw, ok := row["reasoning_efforts"].([]any); ok {
		for _, entry := range raw {
			if level, okLevel := entry.(string); okLevel && strings.TrimSpace(level) != "" {
				model.ReasoningLevels = append(model.ReasoningLevels, strings.TrimSpace(level))
			}
		}
	}
	if capabilities, ok := row["capabilities"].(map[string]any); ok {
		if reasoning, okReasoning := capabilities["reasoning"].(bool); okReasoning {
			model.SupportsReason = reasoning
		}
		if raw, okModality := capabilities["input_modalities"].([]any); okModality {
			for _, entry := range raw {
				if modality, okString := entry.(string); okString && strings.EqualFold(modality, "image") {
					model.SupportsImage = true
				}
			}
		}
	}
	return model, true
}

// FetchCreditDetail reads both points pools. It returns nil when the response
// shape is unusable, which callers must render as unknown rather than zero.
func (c *Client) FetchCreditDetail(ctx context.Context, session string) (*CreditDetail, error) {
	data, err := c.BusinessGET(ctx, session, "/points/records?pageNo=1&pageSize=1&recordType=all")
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if errUnmarshal := json.Unmarshal(data, &payload); errUnmarshal != nil {
		return nil, &Error{Message: "points response is not a JSON object"}
	}
	permanent, okPermanent := numberValue(payload["balance"])
	if !okPermanent {
		return nil, &Error{Message: "points response carries no balance"}
	}
	daily, okDaily := numberValue(payload["dailyBalance"])
	if !okDaily {
		daily = 0
	}
	total, okTotal := numberValue(payload["availableBalance"])
	if !okTotal {
		total = permanent + daily
	}
	detail := &CreditDetail{Permanent: permanent, Daily: daily, Total: total}
	if quota, okQuota := numberValue(payload["dailyQuota"]); okQuota {
		detail.DailyQuota = quota
		detail.DailyQuotaKnown = true
	}
	if consumed, okConsumed := numberValue(payload["dailyConsumed"]); okConsumed {
		detail.DailyConsumed = consumed
	}
	if cycle, okCycle := payload["dailyCycleDate"].(string); okCycle {
		detail.DailyCycleDate = cycle
	}
	return detail, nil
}

// ClaimDailyQuota triggers the daily allowance grant. The endpoint is a write,
// so it is never called from a read path.
func (c *Client) ClaimDailyQuota(ctx context.Context, session string) (json.RawMessage, error) {
	return c.businessRequest(ctx, http.MethodPost, session, "/points/first-login", map[string]any{})
}

// numberValue reads a JSON number (or numeric string) without inventing zero.
func numberValue(value any) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), true
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	case json.Number:
		parsed, errParse := typed.Int64()
		if errParse != nil {
			return 0, false
		}
		return parsed, true
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return 0, false
		}
		var parsed json.Number = json.Number(trimmed)
		value, errParse := parsed.Int64()
		if errParse != nil {
			return 0, false
		}
		return value, true
	default:
		return 0, false
	}
}
