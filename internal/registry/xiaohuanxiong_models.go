package registry

import (
	_ "embed"
	"encoding/json"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

// The Xiaohuanxiong catalog is embedded from a dedicated file rather than the
// shared models/models.json.
//
// The shared catalog is replaced wholesale by StartModelsUpdater when it fetches
// the upstream catalog, and that upstream payload does not (yet) carry a
// xiaohuanxiong section. Keeping the models here means a remote refresh can never
// erase this provider's catalog.
//
//go:embed models/xiaohuanxiong_models.json
var embeddedXiaohuanxiongModelsJSON []byte

// xiaohuanxiongModelsFilePayload mirrors the on-disk shape.
type xiaohuanxiongModelsFilePayload struct {
	Xiaohuanxiong []*ModelInfo `json:"xiaohuanxiong,omitempty"`
	Models        []*ModelInfo `json:"models,omitempty"`
}

type xiaohuanxiongModelsStore struct {
	mu     sync.RWMutex
	models []*ModelInfo
}

var xiaohuanxiongCatalogStore = &xiaohuanxiongModelsStore{}

func init() {
	if _, errLoad := loadXiaohuanxiongModelsFromBytes(embeddedXiaohuanxiongModelsJSON, "embed"); errLoad != nil {
		log.Warnf("registry: failed to parse embedded xiaohuanxiong_models.json: %v", errLoad)
	}
}

// loadXiaohuanxiongModelsFromBytes installs a catalog parsed from raw JSON.
func loadXiaohuanxiongModelsFromBytes(data []byte, source string) ([]*ModelInfo, error) {
	var payload xiaohuanxiongModelsFilePayload
	if errUnmarshal := json.Unmarshal(data, &payload); errUnmarshal != nil {
		return nil, errUnmarshal
	}
	models := payload.Xiaohuanxiong
	if len(models) == 0 {
		models = payload.Models
	}
	normalized := normalizeXiaohuanxiongModels(models)
	xiaohuanxiongCatalogStore.mu.Lock()
	xiaohuanxiongCatalogStore.models = normalized
	xiaohuanxiongCatalogStore.mu.Unlock()
	log.Debugf("registry: loaded %d xiaohuanxiong models from %s", len(normalized), source)
	return cloneModelInfos(normalized), nil
}

// normalizeXiaohuanxiongModels fills defaults and drops unusable entries.
func normalizeXiaohuanxiongModels(models []*ModelInfo) []*ModelInfo {
	out := make([]*ModelInfo, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		if model == nil {
			continue
		}
		id := strings.TrimSpace(model.ID)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		model.ID = id
		if strings.TrimSpace(model.Type) == "" {
			model.Type = "xiaohuanxiong"
		}
		if strings.TrimSpace(model.Object) == "" {
			model.Object = "model"
		}
		if strings.TrimSpace(model.OwnedBy) == "" {
			model.OwnedBy = "sensetime"
		}
		out = append(out, model)
	}
	return out
}

// staticXiaohuanxiongModels is the last-resort catalog used when the embedded
// file cannot be parsed. It mirrors what has been verified against the upstream
// gateway so the provider stays usable.
var staticXiaohuanxiongModels = []*ModelInfo{
	{
		ID: "raccoon-chat-ml-5-5", Object: "model", OwnedBy: "sensetime",
		Type: "xiaohuanxiong", DisplayName: "Raccoon Chat ML 5.5",
		ContextLength: 180000, MaxCompletionTokens: 80000,
		Thinking: &ThinkingSupport{ZeroAllowed: true, Levels: []string{"low", "high"}},
	},
	{
		ID: "glm-5-3", Object: "model", OwnedBy: "sensetime",
		Type: "xiaohuanxiong", DisplayName: "GLM-5.3",
		ContextLength: 200000, MaxCompletionTokens: 65536,
		Thinking: &ThinkingSupport{ZeroAllowed: true, Levels: []string{"low", "high"}},
	},
	{
		ID: "sn-deepseek-v4-1-flash", Object: "model", OwnedBy: "sensetime",
		Type: "xiaohuanxiong", DisplayName: "DeepSeek V4.1 Flash",
		ContextLength: 1000000, MaxCompletionTokens: 100000,
		Thinking: &ThinkingSupport{ZeroAllowed: true, Levels: []string{"low", "high"}},
	},
	{
		ID: "sn-sensenova-6-8-flash-lite", Object: "model", OwnedBy: "sensetime",
		Type: "xiaohuanxiong", DisplayName: "SenseNova 6.8 Flash Lite",
		ContextLength: 131072, MaxCompletionTokens: 65536,
	},
}

// GetXiaohuanxiongModels returns the Xiaohuanxiong (SenseTime Raccoon) catalog.
func GetXiaohuanxiongModels() []*ModelInfo {
	xiaohuanxiongCatalogStore.mu.RLock()
	models := xiaohuanxiongCatalogStore.models
	xiaohuanxiongCatalogStore.mu.RUnlock()
	if len(models) > 0 {
		return cloneModelInfos(models)
	}
	return cloneModelInfos(staticXiaohuanxiongModels)
}

// LookupXiaohuanxiongModel resolves one model from the embedded catalog.
func LookupXiaohuanxiongModel(modelID string) *ModelInfo {
	clean := strings.TrimSpace(modelID)
	if clean == "" {
		return nil
	}
	xiaohuanxiongCatalogStore.mu.RLock()
	defer xiaohuanxiongCatalogStore.mu.RUnlock()
	for _, model := range xiaohuanxiongCatalogStore.models {
		if model != nil && model.ID == clean {
			return model
		}
	}
	return nil
}
