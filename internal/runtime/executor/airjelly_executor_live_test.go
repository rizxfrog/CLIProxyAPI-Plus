package executor

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/auth/airjelly"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v8/sdk/translator"
)

func TestAirJellyLiveChat(t *testing.T) {
	path := os.Getenv("AIRJELLY_TEST_SESSION_FILE")
	if path == "" || os.Getenv("AIRJELLY_GATEWAY_HMAC_KEY") == "" {
		t.Skip("live credentials not configured")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var outer map[string]string
	if err = json.Unmarshal(b, &outer); err != nil {
		t.Fatal(err)
	}
	var session airjelly.Session
	for _, v := range outer {
		if err = json.Unmarshal([]byte(v), &session); err != nil {
			t.Fatal(err)
		}
		break
	}
	auth := &cliproxyauth.Auth{Provider: airjelly.Provider, Attributes: map[string]string{"access_token": session.AccessToken}}
	executor := NewAirJellyExecutor(nil)
	request := cliproxyexecutor.Request{Model: airjelly.Model, Payload: []byte(`{"model":"airjelly-lite","messages":[{"role":"user","content":"Reply only OK."}],"max_tokens":200}`)}
	opts := cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatOpenAI}
	response, err := executor.Execute(context.Background(), auth, request, opts)
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err = json.Unmarshal(response.Payload, &parsed); err != nil {
		t.Fatalf("invalid completion: %v", err)
	}
	if len(parsed.Choices) == 0 || parsed.Choices[0].Message.Content == "" {
		t.Fatalf("empty completion: %s", response.Payload)
	}
	t.Logf("received content (%d chars)", len(parsed.Choices[0].Message.Content))
}
