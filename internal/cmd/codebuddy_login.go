package cmd

import (
	"context"
	"fmt"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	sdkAuth "github.com/router-for-me/CLIProxyAPI/v8/sdk/auth"
	log "github.com/sirupsen/logrus"
)

// DoCodeBuddyCNLogin starts CodeBuddy CN browser authorization and saves the tokens.
func DoCodeBuddyCNLogin(cfg *config.Config, options *LoginOptions) {
	doCodeBuddyLogin(cfg, options, "codebuddy-cn", "CodeBuddy CN")
}

// DoCodeBuddyAILogin starts international CodeBuddy AI browser authorization and
// saves the tokens.
func DoCodeBuddyAILogin(cfg *config.Config, options *LoginOptions) {
	doCodeBuddyLogin(cfg, options, "codebuddy-ai", "CodeBuddy AI")
}

func doCodeBuddyLogin(cfg *config.Config, options *LoginOptions, provider, label string) {
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
