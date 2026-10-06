package cliproxy

import (
	"context"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	runtimeexecutor "github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

func TestRegisterQwenWebAccount(t *testing.T) {
	const id = "qwen-web-service-test"
	modelRegistry := registry.GetGlobalRegistry()
	defer modelRegistry.UnregisterClient(id)
	service := &Service{cfg: &config.Config{}, coreManager: coreauth.NewManager(nil, nil, nil)}
	auth := &coreauth.Auth{ID: id, Provider: "qwen-web", Metadata: map[string]any{"access_token": "test-session"}}
	var errRegister error
	auth, errRegister = service.coreManager.Register(context.Background(), auth)
	if errRegister != nil {
		t.Fatal(errRegister)
	}
	service.registerExecutorForAuth(auth, false)
	registered, ok := service.coreManager.Executor("qwen-web")
	if !ok {
		t.Fatal("executor missing")
	}
	if _, ok := registered.(*runtimeexecutor.QwenWebExecutor); !ok {
		t.Fatalf("wrong executor %T", registered)
	}
	service.registerModelsForAuth(context.Background(), auth)
	models := modelRegistry.GetModelsForClient(id)
	if len(models) != 2 {
		t.Fatalf("registered models: %+v", models)
	}
	info := registry.LookupModelInfo("qwen-web-image")
	if info == nil || info.Type != registry.OpenAIImageModelType {
		t.Fatal("image not routable")
	}
}
