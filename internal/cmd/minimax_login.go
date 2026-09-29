package cmd

import (
	"context"
	"fmt"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	sdkAuth "github.com/router-for-me/CLIProxyAPI/v8/sdk/auth"
	log "github.com/sirupsen/logrus"
)

// DoMinimaxLogin starts the MiniMax Code international OAuth device
// authorization flow and saves the resulting tokens.
func DoMinimaxLogin(cfg *config.Config, options *LoginOptions) {
	doMinimaxLogin(cfg, options, "minimax", "MiniMax Code (International)")
}

// DoMinimaxCNLogin starts the MiniMax Code mainland China OAuth device
// authorization flow and saves the resulting tokens.
func DoMinimaxCNLogin(cfg *config.Config, options *LoginOptions) {
	doMinimaxLogin(cfg, options, "minimax-cn", "MiniMax Code (China)")
}

func doMinimaxLogin(cfg *config.Config, options *LoginOptions, provider, label string) {
	if options == nil {
		options = &LoginOptions{}
	}
	record, savedPath, err := newAuthManager().Login(context.Background(), provider, cfg, &sdkAuth.LoginOptions{
		NoBrowser: options.NoBrowser,
		Metadata:  map[string]string{},
		Prompt:    options.Prompt,
	})
	if err != nil {
		log.Errorf("%s authentication failed: %v", label, err)
		return
	}
	if savedPath != "" {
		fmt.Printf("Authentication saved to %s\n", savedPath)
	}
	if record != nil && record.Label != "" {
		fmt.Printf("Authenticated as %s\n", record.Label)
	}
	fmt.Printf("%s authentication successful!\n", label)
}
