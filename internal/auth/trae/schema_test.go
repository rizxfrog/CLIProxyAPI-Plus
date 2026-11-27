package trae

import (
	"strings"
	"testing"
)

// TestDescribeModelListReportsShapeWithoutValues pins the diagnostic contract:
// the report names fields and observed types so an operator can confirm the
// inferred schema, but must never echo a value. A live catalogue row carries
// account credentials and encrypted prompt blobs.
func TestDescribeModelListReportsShapeWithoutValues(t *testing.T) {
	raw := []byte(`{"code":0,"message":"","config_info_list":[
		{"config_name":"glm-5.2","config_switch":true,"context_window_tokens":{"dev":128000,"max":200000},"display_config":{"display_name":"GLM-5.2","multimodal":true},"ak":"SECRET-AK","extra_config":"SECRET-PROMPT"},
		{"config_name":"kimi-k3","config_switch":false}
	]}`)

	report := DescribeModelList(raw)
	rendered := report.String()

	for _, want := range []string{
		"envelope keys: code, config_info_list, message",
		"config_info_list: 2 entries",
		"config_name(string)",
		"context_window_tokens(object{dev,max})",
		"parsed: 1, skipped: 1",
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("report missing %q\n%s", want, rendered)
		}
	}

	for _, secret := range []string{"SECRET-AK", "SECRET-PROMPT", "GLM-5.2", "128000", "200000"} {
		if strings.Contains(rendered, secret) {
			t.Errorf("report leaked value %q\n%s", secret, rendered)
		}
	}
	// Sensitive keys stay visible as redacted markers so the operator still sees
	// the field exists.
	for _, key := range []string{"ak(<redacted>)", "extra_config(<redacted>)"} {
		if !strings.Contains(rendered, key) {
			t.Errorf("report missing redacted marker %q\n%s", key, rendered)
		}
	}
}

// TestDescribeModelListFlagsMissingRoutingKey pins the anomaly path: schema drift
// must surface as an explicit warning instead of silent metadata loss.
func TestDescribeModelListFlagsMissingRoutingKey(t *testing.T) {
	report := DescribeModelList([]byte(`{"code":0,"config_info_list":[{"name":"glm-5.2"}]}`))
	rendered := report.String()
	if !strings.Contains(rendered, "no row carries config_name") {
		t.Fatalf("missing routing-key anomaly\n%s", rendered)
	}
	if !strings.Contains(rendered, "no context window field found") {
		t.Fatalf("missing context-window anomaly\n%s", rendered)
	}
}

func TestDescribeModelListHandlesEmptyAndMalformedBodies(t *testing.T) {
	if got := DescribeModelList([]byte(`{"code":0,"config_info_list":[]}`)).String(); !strings.Contains(got, "config_info_list is empty") {
		t.Fatalf("empty body not flagged: %s", got)
	}
	if got := DescribeModelList([]byte(`<html>404</html>`)).String(); !strings.Contains(got, "not a JSON object") {
		t.Fatalf("malformed body not flagged: %s", got)
	}
	if got := DescribeModelList([]byte(`{"code":1001}`)).String(); !strings.Contains(got, "config_info_list missing") {
		t.Fatalf("missing envelope key not flagged: %s", got)
	}
}
