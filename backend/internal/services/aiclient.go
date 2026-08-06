package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const defaultOpenRouterURL = "https://openrouter.ai/api/v1"
const defaultModel = "deepseek/deepseek-v4-flash"

type AIClient struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

func NewAIClient(baseURL, apiKey, model string) *AIClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = defaultOpenRouterURL
	}
	model = strings.TrimSpace(model)
	if model == "" {
		model = defaultModel
	}
	return &AIClient{
		baseURL: baseURL,
		apiKey:  strings.TrimSpace(apiKey),
		model:   model,
		client:  &http.Client{Timeout: 45 * time.Second},
	}
}

type inferResponse struct {
	Text      string `json:"text"`
	Strengths string `json:"strengths"`
	Risks     string `json:"risks"`
	Potential string `json:"potential"`
}

type openRouterRequest struct {
	Model    string              `json:"model"`
	Messages []openRouterMessage `json:"messages"`
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (a *AIClient) Infer(ctx context.Context, task, features string) (inferResponse, error) {
	if a.apiKey == "" {
		return inferResponse{}, fmt.Errorf("missing OPENROUTER_API_KEY")
	}

	body, _ := json.Marshal(openRouterRequest{
		Model: a.model,
		Messages: []openRouterMessage{
			{Role: "system", Content: systemPrompt(task)},
			{Role: "user", Content: features},
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return inferResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/poisonaifih/roasty")
	req.Header.Set("X-Title", "Roasty")

	res, err := a.client.Do(req)
	if err != nil {
		return inferResponse{}, err
	}
	defer res.Body.Close()

	var parsed openRouterResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return inferResponse{}, err
	}
	if res.StatusCode >= 300 {
		msg := fmt.Sprintf("openrouter status %d", res.StatusCode)
		if parsed.Error != nil && parsed.Error.Message != "" {
			msg = parsed.Error.Message
		}
		return inferResponse{}, fmt.Errorf("%s", msg)
	}
	if len(parsed.Choices) == 0 {
		return inferResponse{}, fmt.Errorf("openrouter empty response")
	}

	text := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if strings.EqualFold(task, "SCOUT") {
		return parseScout(text), nil
	}
	return inferResponse{Text: text}, nil
}

func systemPrompt(task string) string {
	const english = " Always reply in English only. Never use Indonesian or any other language."
	switch strings.ToUpper(strings.TrimSpace(task)) {
	case "SCOUT":
		return "You are a coffee roastery sourcing advisor. Given bean features, reply in exactly this format on one line: Strengths: <short> | Risks: <short> | Potential: <short>. Use plain English, no markdown." + english
	case "RESTOCK":
		return "You are a coffee inventory advisor. Given stock signals, reply with one short restock recommendation in plain English." + english
	case "CRM":
		return "You are a coffee shop CRM advisor. Given order and payment signals, reply with one short follow-up message in plain English." + english
	case "SHOP":
		return "You are a coffee sourcing advisor. Given origin/variety and live shop search hits, reply with one short note comparing online vs offline buying options in plain English." + english
	default:
		return "You are a coffee roastery operations advisor. Reply briefly in plain English." + english
	}
}

func parseScout(text string) inferResponse {
	out := inferResponse{Text: text}
	if !strings.Contains(text, "|") {
		out.Strengths = text
		return out
	}
	parts := strings.Split(text, "|")
	if len(parts) >= 1 {
		out.Strengths = trimLabel(parts[0], "strengths")
	}
	if len(parts) >= 2 {
		out.Risks = trimLabel(parts[1], "risks")
	}
	if len(parts) >= 3 {
		out.Potential = trimLabel(parts[2], "potential")
	}
	return out
}

func trimLabel(s, label string) string {
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)
	prefix := label + ":"
	if strings.HasPrefix(lower, prefix) {
		return strings.TrimSpace(s[len(prefix):])
	}
	return s
}
