package config

import (
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestV8MigrationPreservesForkExtensions(t *testing.T) {
	data := []byte(`metrics-enabled: true
system-prompt-override:
  enabled: true
  prompt: "Keep the configured prompt."
provider-models:
  minimax:
    disabled: [MiniMax-M2.7]
remote-management:
  anti-bot:
    enabled: true
`)
	families := []string{"codebuddy-cn", "codebuddy-ai", "deepseek-web", "xiaohuanxiong", "codearts", "trae", "cline", "qoder-cn", "qoder-ai"}
	for _, provider := range families {
		data = append(data, []byte(fmt.Sprintf("%s-api-key:\n  - api-key: test-key\n    base-url: https://example.com\n    weight: 2\n", provider))...)
	}
	before, err := ParseConfigBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	migrated, changed, err := NormalizeConfigLayout(data, true)
	if err != nil || !changed {
		t.Fatalf("migration: changed=%v error=%v", changed, err)
	}
	if err = ValidateV8Config(migrated); err != nil {
		t.Fatalf("validate migrated extensions: %v", err)
	}
	after, err := ParseConfigBytes(migrated)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("v8 migration changed fork extension values")
	}
	var doc yaml.Node
	if err = yaml.Unmarshal(migrated, &doc); err != nil {
		t.Fatal(err)
	}
	root := doc.Content[0]
	paths := []string{"observability.metrics-enabled", "requests.system-prompt-override.enabled", "routing.provider-models.minimax", "management.anti-bot.enabled"}
	for _, provider := range families {
		paths = append(paths, "api-keys."+provider)
	}
	for _, path := range paths {
		if yamlPath(root, path) == nil {
			t.Errorf("missing migrated extension %s", path)
		}
	}
}
