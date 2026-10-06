package registry

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestCatalogFetcherSources(t *testing.T) {
	for _, tc := range []struct {
		name     string
		data     []byte
		validate func([]byte) error
	}{
		{"general", embeddedModelsJSON, validateCatalogBytes},
		{"codex", embeddedCodexClientModelsJSON, ValidateCodexClientModelsJSON},
		{"devin", embeddedDevinModelsJSON, validateDevinCatalogBytes},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var officialHits atomic.Int32
			official := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { officialHits.Add(1); _, _ = w.Write(tc.data) }))
			defer official.Close()
			custom := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(tc.data) }))
			defer custom.Close()
			broken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("invalid")) }))
			defer broken.Close()
			path := filepath.Join(t.TempDir(), "catalog.json")
			if errWrite := os.WriteFile(path, tc.data, 0600); errWrite != nil {
				t.Fatal(errWrite)
			}
			fetch := catalogFetcher(tc.data, []string{official.URL}, tc.validate)
			for _, source := range []string{embeddedCatalogSource, custom.URL, path} {
				data, errFetch := fetch(context.Background(), source)
				if errFetch != nil || string(data) != string(tc.data) {
					t.Fatalf("source %s: %v", source, errFetch)
				}
			}
			for _, source := range []string{broken.URL, path + ".missing"} {
				if _, errFetch := fetch(context.Background(), source); errFetch == nil {
					t.Fatal("accepted invalid custom source")
				}
			}
			if officialHits.Load() != 0 {
				t.Fatal("custom sources fell back to official URL")
			}
			if _, errFetch := fetch(context.Background(), ""); errFetch != nil {
				t.Fatal(errFetch)
			}
			if officialHits.Load() != 1 {
				t.Fatal("default source was not fetched")
			}
			if errWrite := os.WriteFile(path, []byte("invalid"), 0600); errWrite != nil {
				t.Fatal(errWrite)
			}
			if _, errFetch := fetch(context.Background(), path); errFetch == nil {
				t.Fatal("file was not reread")
			}
		})
	}
}

func TestCatalogSourceSwitchRejectsStalePublication(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	published := make(chan string, 3)
	u := &catalogUpdater{
		generation: 1,
		fetch: func(ctx context.Context, source string) ([]byte, error) {
			if source == "old" {
				close(started)
				<-release
			}
			return []byte(source), nil
		},
		publish: func(data []byte) ([]string, error) { published <- string(data); return nil, nil },
	}
	go func() { u.refresh(ctx, "old", 1); close(done) }()
	<-started
	u.configure(ctx, "new")
	select {
	case value := <-published:
		if value != "new" {
			t.Fatalf("published %s", value)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("new source blocked behind old fetch")
	}
	close(release)
	<-done
	select {
	case value := <-published:
		t.Fatalf("stale publication: %s", value)
	default:
	}
	u.configure(ctx, "new")
	u.mu.Lock()
	generation := u.generation
	u.mu.Unlock()
	if generation != 2 {
		t.Fatal("unchanged source restarted")
	}
}

func TestCatalogRefreshKeepsLastValidData(t *testing.T) {
	last := "last valid"
	u := &catalogUpdater{
		fetch:   func(context.Context, string) ([]byte, error) { return nil, errors.New("unavailable") },
		publish: func(data []byte) ([]string, error) { last = string(data); return nil, nil },
	}
	u.refresh(context.Background(), "", 0)
	if last != "last valid" {
		t.Fatal("failed fetch changed catalog")
	}
	u.fetch = func(context.Context, string) ([]byte, error) { return []byte("invalid"), nil }
	u.publish = func([]byte) ([]string, error) { return nil, errors.New("invalid") }
	u.refresh(context.Background(), "", 0)
	if last != "last valid" {
		t.Fatal("invalid catalog changed data")
	}
}

// TestPublishCatalogBytesKeepsForkSections guards the invariant that a remote
// refresh of the shared upstream catalog cannot drop the provider sections this
// fork publishes: MiniMax Code managed accounts, CodeBuddy, DeepSeek Web, TRAE,
// Cline, Qoder, and Meta. Those models are built into the embedded catalog and
// absent from the upstream payload, so a wholesale replace would erase them and
// silently unlist every one of those providers.
func TestPublishCatalogBytesKeepsForkSections(t *testing.T) {
	install := func(t *testing.T, data *staticModelsJSON) *staticModelsJSON {
		t.Helper()
		modelsCatalogStore.mu.Lock()
		previous := modelsCatalogStore.data
		modelsCatalogStore.data = data
		modelsCatalogStore.mu.Unlock()
		t.Cleanup(func() {
			modelsCatalogStore.mu.Lock()
			modelsCatalogStore.data = previous
			modelsCatalogStore.mu.Unlock()
		})
		return previous
	}

	forkSections := func(data *staticModelsJSON) map[string]int {
		return map[string]int{
			"minimax":      len(data.Minimax),
			"codebuddy-cn": len(data.CodeBuddyCN),
			"codebuddy-ai": len(data.CodeBuddyAI),
			"deepseek-web": len(data.DeepSeekWeb),
			"trae":         len(data.Trae),
			"cline":        len(data.Cline),
			"qoder-cn":     len(data.QoderCN),
			"qoder-ai":     len(data.QoderAI),
			"meta":         len(data.Meta),
		}
	}

	current := &staticModelsJSON{
		Claude:      []*ModelInfo{{ID: "claude-opus-5-5"}},
		Minimax:     []*ModelInfo{{ID: "minimax-m2.7"}},
		CodeBuddyCN: []*ModelInfo{{ID: "hy3"}},
		CodeBuddyAI: []*ModelInfo{{ID: "claude-sonnet-5"}},
		DeepSeekWeb: []*ModelInfo{{ID: "deepseek-v4-pro"}},
		Trae:        []*ModelInfo{{ID: "trae-solo"}},
		Cline:       []*ModelInfo{{ID: "cline-1"}},
		QoderCN:     []*ModelInfo{{ID: "qmodel"}},
		QoderAI:     []*ModelInfo{{ID: "qmodel-ai"}},
		Meta:        []*ModelInfo{{ID: "muse-spark-1.3"}},
	}
	install(t, current)

	// An upstream catalog that carries official providers only.
	upstream, errMarshal := json.Marshal(&staticModelsJSON{Claude: []*ModelInfo{{ID: "claude-opus-5-5", DisplayName: "Opus 5.5"}}})
	if errMarshal != nil {
		t.Fatal(errMarshal)
	}
	if _, errPublish := publishCatalogBytes(upstream); errPublish != nil {
		t.Fatalf("publishCatalogBytes() error = %v", errPublish)
	}

	published := getModels()
	for provider, want := range forkSections(current) {
		if want == 0 {
			t.Fatalf("test setup bug: %s has no local models to preserve", provider)
		}
	}
	got := forkSections(published)
	for provider, want := range forkSections(current) {
		if got[provider] != want {
			t.Fatalf("refresh dropped fork section %s: got %d models, want %d", provider, got[provider], want)
		}
	}
	if len(published.Claude) != 1 || published.Claude[0].DisplayName != "Opus 5.5" {
		t.Fatalf("incoming catalog did not win for a shared section: %+v", published.Claude)
	}
}
