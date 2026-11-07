package dumate

import "testing"

func TestResolveBaseURL(t *testing.T) {
	for _, tc := range []struct {
		name     string
		metadata map[string]any
		want     string
	}{
		{
			name:     "explicit base url",
			metadata: map[string]any{"base_url": "http://127.0.0.1:52194/api/qianfanproxy/v1"},
			want:     "http://127.0.0.1:52194/api/qianfanproxy/v1",
		},
		{
			name:     "base url without suffix gains it",
			metadata: map[string]any{"base_url": "http://127.0.0.1:52194"},
			want:     "http://127.0.0.1:52194" + InferencePathSuffix,
		},
		{
			name:     "trailing slash is trimmed",
			metadata: map[string]any{"base_url": "http://127.0.0.1:52194/api/qianfanproxy/v1/"},
			want:     "http://127.0.0.1:52194/api/qianfanproxy/v1",
		},
		{
			name:     "port becomes the base url",
			metadata: map[string]any{"port": float64(52194)},
			want:     "http://127.0.0.1:52194/api/qianfanproxy/v1",
		},
		{
			name:     "string port becomes the base url",
			metadata: map[string]any{"port": "52194"},
			want:     "http://127.0.0.1:52194/api/qianfanproxy/v1",
		},
		{
			name:     "base url wins over port",
			metadata: map[string]any{"base_url": "http://host:1234/api/qianfanproxy/v1", "port": float64(1)},
			want:     "http://host:1234/api/qianfanproxy/v1",
		},
		{
			name:     "neither field is empty",
			metadata: map[string]any{"account_id": "abc"},
			want:     "",
		},
		{
			name:     "nil metadata is empty",
			metadata: nil,
			want:     "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveBaseURL(tc.metadata); got != tc.want {
				t.Fatalf("ResolveBaseURL() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestResolveAPIKey(t *testing.T) {
	if got := ResolveAPIKey(map[string]any{"api_key": "custom"}); got != "custom" {
		t.Fatalf("explicit api_key = %q", got)
	}
	// The desktop client always sends a non-empty bearer token, even though the
	// real authorization is carried by the X-Dumate-* headers.
	if got := ResolveAPIKey(map[string]any{}); got != PlaceholderAPIKey {
		t.Fatalf("default api_key = %q, want %q", got, PlaceholderAPIKey)
	}
	if got := ResolveAPIKey(nil); got != PlaceholderAPIKey {
		t.Fatalf("nil metadata api_key = %q, want %q", got, PlaceholderAPIKey)
	}
}

func TestRuntimeHeaders(t *testing.T) {
	full := RuntimeHeaders(map[string]any{
		"account_id": "91fc4f6327264f38a269db24427804e0",
		"device_id":  "1|windows.deadbeef",
		"inapp_key":  "3a7d79475dff98880ebce69d59450ee89c3c225742ed8d3b5bb712a4f3ee9e87",
	})
	want := map[string]string{
		HeaderAccountID: "91fc4f6327264f38a269db24427804e0",
		HeaderDeviceID:  "1|windows.deadbeef",
		HeaderInappKey:  "3a7d79475dff98880ebce69d59450ee89c3c225742ed8d3b5bb712a4f3ee9e87",
	}
	if len(full) != len(want) {
		t.Fatalf("header count = %d, want %d (%v)", len(full), len(want), full)
	}
	for name, value := range want {
		if full[name] != value {
			t.Fatalf("header %s = %q, want %q", name, full[name], value)
		}
	}

	// Empty fields must be omitted rather than sent blank, so the upstream
	// rejects the request with its own error instead of treating it as signed.
	partial := RuntimeHeaders(map[string]any{"account_id": "abc", "inapp_key": "  "})
	if len(partial) != 1 || partial[HeaderAccountID] != "abc" {
		t.Fatalf("partial headers = %v", partial)
	}
	if headers := RuntimeHeaders(map[string]any{}); headers != nil {
		t.Fatalf("empty metadata produced %v, want nil", headers)
	}
}
