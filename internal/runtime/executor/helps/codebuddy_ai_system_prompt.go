package helps

import (
	_ "embed"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
)

// CodeBuddy AI (international, https://www.codebuddy.ai) requires the first
// chat message to be a system prompt; the gateway rejects requests whose first
// message is not role "system" with HTTP 400 code 11128. The official CLI
// renders the product.json "cli-agent-prompt" template and unshifts it as the
// first message before every request.
//
// The embedded template below is that prompt captured from the official CLI
// with its volatile sections replaced by placeholders:
//
//   - {{CODEBUDDY_AI_ENV_BLOCK}}    the <env> ... </env> runtime environment block
//   - {{CODEBUDDY_AI_MODEL_NAME}}   the model display name
//   - {{CODEBUDDY_AI_MODEL_ID}}     the gateway model id
//
//go:embed codebuddy_ai_system_prompt.txt
var codeBuddyAISystemPromptTemplate string

const (
	codeBuddyAIEnvPlaceholder       = "{{CODEBUDDY_AI_ENV_BLOCK}}"
	codeBuddyAIModelNamePlaceholder = "{{CODEBUDDY_AI_MODEL_NAME}}"
	codeBuddyAIModelIDPlaceholder   = "{{CODEBUDDY_AI_MODEL_ID}}"
	codeBuddyAIDocsDirPlaceholder   = "{{CODEBUDDY_AI_DOCS_DIR}}"

	// CodeBuddyAIEnvBlockPrefix marks the start of the environment block in the
	// rendered prompt. It is exported for tests and callers that need to locate
	// the block after rendering.
	CodeBuddyAIEnvBlockPrefix = "<env>"

	// codeBuddyAIDefaultModelID is used when the request model is unknown.
	codeBuddyAIDefaultModelID = "default-model"

	// codeBuddyAIDefaultDocsDir is the official documentation location used when
	// no local CLI docs directory is configured.
	codeBuddyAIDefaultDocsDir = "https://cnb.cool/codebuddy/codebuddy-code/-/git/raw/main/docs"
)

// CodeBuddyAIEnvInfo carries optional environment details for the rendered
// system prompt. Fields left empty are derived from the host at render time.
type CodeBuddyAIEnvInfo struct {
	// WorkingDir overrides the working directory reported in the prompt.
	WorkingDir string
	// IsGitRepo reports whether WorkingDir is a git repository.
	IsGitRepo bool
	// GitRepoSet marks IsGitRepo as authoritative; when false the line is omitted.
	GitRepoSet bool
	// Platform overrides the OS platform token (e.g. "linux", "darwin", "windows").
	Platform string
	// OSVersion overrides the OS version string.
	OSVersion string
	// Shell overrides the default shell name.
	Shell string
	// Date overrides today's date.
	Date string
}

// CodeBuddyAISystemPromptOptions controls rendering of the CodeBuddy AI system
// prompt.
type CodeBuddyAISystemPromptOptions struct {
	// ModelName is the model display name used in the background info block.
	ModelName string
	// ModelID is the gateway model id used in the background info block.
	ModelID string
	// DocsDir is the documentation directory advertised to the model. When empty
	// the official online documentation location is used.
	DocsDir string
	// Env optionally overrides the environment block fields.
	Env CodeBuddyAIEnvInfo
}

// RenderCodeBuddyAISystemPrompt renders the embedded CodeBuddy AI agent system
// prompt with the given model and environment details. The result always
// satisfies the gateway's requirement that the first chat message be a system
// prompt.
func RenderCodeBuddyAISystemPrompt(opts CodeBuddyAISystemPromptOptions) string {
	modelID := strings.TrimSpace(opts.ModelID)
	if modelID == "" {
		modelID = codeBuddyAIDefaultModelID
	}
	modelName := strings.TrimSpace(opts.ModelName)
	if modelName == "" {
		modelName = codeBuddyAIDisplayName(modelID)
	}

	prompt := strings.ReplaceAll(codeBuddyAISystemPromptTemplate, codeBuddyAIModelIDPlaceholder, modelID)
	prompt = strings.ReplaceAll(prompt, codeBuddyAIModelNamePlaceholder, modelName)
	prompt = strings.ReplaceAll(prompt, codeBuddyAIDocsDirPlaceholder, codeBuddyAIDocsDir(opts.DocsDir))
	prompt = strings.ReplaceAll(prompt, codeBuddyAIEnvPlaceholder, renderCodeBuddyAIEnvBlock(opts.Env))
	return prompt
}

// codeBuddyAIDocsDir resolves the documentation directory advertised in the
// prompt, falling back to the official online documentation.
func codeBuddyAIDocsDir(docsDir string) string {
	if dir := strings.TrimSpace(docsDir); dir != "" {
		return strings.TrimRight(dir, "/")
	}
	return codeBuddyAIDefaultDocsDir
}

// codeBuddyAIDisplayName returns the catalog display name for a CodeBuddy AI
// model, falling back to the raw id when the model is not in the catalog.
func codeBuddyAIDisplayName(modelID string) string {
	modelID = strings.TrimSpace(modelID)
	for _, model := range registry.GetCodeBuddyAIModels() {
		if model == nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(model.ID), modelID) {
			if name := strings.TrimSpace(model.DisplayName); name != "" {
				return name
			}
			break
		}
	}
	return modelID
}

// renderCodeBuddyAIEnvBlock builds the <env> block describing the runtime.
func renderCodeBuddyAIEnvBlock(env CodeBuddyAIEnvInfo) string {
	workingDir := strings.TrimSpace(env.WorkingDir)
	if workingDir == "" {
		workingDir = codeBuddyAIWorkingDir()
	}
	platform := strings.TrimSpace(env.Platform)
	if platform == "" {
		platform = runtime.GOOS
	}
	shell := strings.TrimSpace(env.Shell)
	if shell == "" {
		shell = defaultShellName()
	}
	osVersion := strings.TrimSpace(env.OSVersion)
	if osVersion == "" {
		osVersion = runtime.GOOS
	}
	date := strings.TrimSpace(env.Date)
	if date == "" {
		date = time.Now().Format("Monday, Jan 2, 2006")
	}

	var b strings.Builder
	b.WriteString(CodeBuddyAIEnvBlockPrefix)
	b.WriteString("\n")
	fmt.Fprintf(&b, "Working directory: %s\n", workingDir)
	if env.GitRepoSet {
		if env.IsGitRepo {
			b.WriteString("Is directory a git repo: Yes\n")
		} else {
			b.WriteString("Is directory a git repo: No\n")
		}
	}
	fmt.Fprintf(&b, "Platform: %s\n\n", platform)
	fmt.Fprintf(&b, "OS Version: %s\n", osVersion)
	fmt.Fprintf(&b, "Default shell: %s\n", shell)
	fmt.Fprintf(&b, "Today's date: %s", date)
	b.WriteString("</env>")
	return b.String()
}

var (
	codeBuddyAIWorkingDirOnce sync.Once
	codeBuddyAIWorkingDirPath string
)

func codeBuddyAIWorkingDir() string {
	codeBuddyAIWorkingDirOnce.Do(func() {
		if dir, err := os.Getwd(); err == nil {
			codeBuddyAIWorkingDirPath = dir
		}
	})
	if codeBuddyAIWorkingDirPath == "" {
		return "."
	}
	return codeBuddyAIWorkingDirPath
}

// defaultShellName reports the shell used in the prompt environment block.
func defaultShellName() string {
	if runtime.GOOS == "windows" {
		return "powershell"
	}
	return "bash"
}
