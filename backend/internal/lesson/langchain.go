package lesson

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Carlos-hub/planejai/backend/internal/store"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
)

// lcGen is a Generator backed by a LangChainGo llms.Model. It is brand
// agnostic: the concrete provider is selected at construction time via
// NewGeneratorForProvider.
type lcGen struct {
	llm llms.Model
}

// defaultModels maps each supported provider to its fixed model. Adjust here
// to change which model a provider uses.
var defaultModels = map[string]string{
	"anthropic": "claude-opus-4-8",
	"openai":    "gpt-4o",
	"googleai":  "gemini-2.0-flash",
	"deepseek":  "deepseek-chat",
	"llama":     "Llama-3.3-70B-Instruct",
}

// providerBaseURLs holds custom OpenAI-compatible base URLs for providers that
// speak the OpenAI API but live elsewhere.
var providerBaseURLs = map[string]string{
	"deepseek": "https://api.deepseek.com",
	"llama":    "https://api.llama.com/compat/v1",
}

// NewGeneratorForProvider builds a Generator for the given provider using the
// supplied API key. The model is fixed per provider (defaultModels). Only
// internal/lesson imports provider-specific packages.
func NewGeneratorForProvider(ctx context.Context, provider, apiKey string) (Generator, error) {
	model, ok := defaultModels[provider]
	if !ok {
		return nil, fmt.Errorf("provider não suportado: %s", provider)
	}
	switch provider {
	case "anthropic":
		m, err := anthropic.New(anthropic.WithModel(model), anthropic.WithToken(apiKey))
		if err != nil {
			return nil, err
		}
		return &lcGen{llm: m}, nil
	case "openai", "deepseek", "llama":
		opts := []openai.Option{openai.WithModel(model), openai.WithToken(apiKey)}
		if base, ok := providerBaseURLs[provider]; ok {
			opts = append(opts, openai.WithBaseURL(base))
		}
		m, err := openai.New(opts...)
		if err != nil {
			return nil, err
		}
		return &lcGen{llm: m}, nil
	case "googleai":
		m, err := googleai.New(ctx, googleai.WithAPIKey(apiKey), googleai.WithDefaultModel(model))
		if err != nil {
			return nil, err
		}
		return &lcGen{llm: m}, nil
	default:
		return nil, fmt.Errorf("provider não suportado: %s", provider)
	}
}

// Generate creates a new LessonData from a BNCC skill and duration by
// prompting the underlying LLM.
func (g *lcGen) Generate(ctx context.Context, skill store.BnccSkill, duracaoMin int) (LessonData, error) {
	prompt := systemPrompt + "\n\n" + generateUserPrompt(skill, duracaoMin)
	resp, err := llms.GenerateFromSinglePrompt(ctx, g.llm, prompt)
	if err != nil {
		return LessonData{}, fmt.Errorf("llm generate failed: %w", err)
	}
	return parseLessonResponse(resp)
}

// Enhance improves an existing LessonData draft based on a BNCC skill by
// prompting the underlying LLM.
func (g *lcGen) Enhance(ctx context.Context, draft LessonData, skill store.BnccSkill) (LessonData, error) {
	draftBytes, err := json.Marshal(draft)
	if err != nil {
		return LessonData{}, fmt.Errorf("failed to marshal draft: %w", err)
	}
	prompt := systemPrompt + "\n\n" + enhanceUserPrompt(skill, string(draftBytes))
	resp, err := llms.GenerateFromSinglePrompt(ctx, g.llm, prompt)
	if err != nil {
		return LessonData{}, fmt.Errorf("llm generate failed: %w", err)
	}
	return parseLessonResponse(resp)
}

// parseLessonResponse attempts to parse the LLM response directly as
// LessonData JSON; if that fails, it extracts the first JSON object found in
// the response and retries. If parsing still fails, the parse error is
// returned so the caller can set status=falha.
func parseLessonResponse(resp string) (LessonData, error) {
	ld, err := ParseLessonData([]byte(resp))
	if err == nil {
		return ld, nil
	}

	extracted := extractJSON(resp)
	if extracted == nil {
		return LessonData{}, err
	}

	ld, err2 := ParseLessonData(extracted)
	if err2 != nil {
		return LessonData{}, err2
	}
	return ld, nil
}

// extractJSON trims markdown code fences from s and returns the substring
// from the first '{' to the last '}'. Returns nil if no such substring is
// found.
func extractJSON(s string) []byte {
	trimmed := strings.TrimSpace(s)

	// Strip markdown fences like ```json ... ``` or ``` ... ```
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimPrefix(trimmed, "json")
		trimmed = strings.TrimPrefix(trimmed, "JSON")
		trimmed = strings.TrimSpace(trimmed)
		if idx := strings.LastIndex(trimmed, "```"); idx != -1 {
			trimmed = trimmed[:idx]
		}
		trimmed = strings.TrimSpace(trimmed)
	}

	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start == -1 || end == -1 || end < start {
		return nil
	}
	return []byte(trimmed[start : end+1])
}
