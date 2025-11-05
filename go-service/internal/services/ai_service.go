package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/code-zt/vidnotes/config"
	"github.com/code-zt/vidnotes/internal/models"
)

type AIService interface {
	SendMessage(ctx context.Context, session *models.AISession, userMessage string) (string, error)
	CreateSessionContext(videoSummary string) string
	ImproveSummary(ctx context.Context, currentSummary string, issues []string) (string, error)
	FixSummaryErrors(ctx context.Context, currentSummary string) (string, error)
	CreateSummaryFromDialogue(ctx context.Context, messages []models.AIMessage) (string, error)
}

type OpenRouterService struct {
	config *config.OpenRouterConfig
	client *http.Client
}

func NewOpenRouterService(cfg *config.OpenRouterConfig) AIService {
	return &OpenRouterService{
		config: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

type OpenRouterRequest struct {
	Model    string              `json:"model"`
	Messages []OpenRouterMessage `json:"messages"`
	Stream   bool                `json:"stream"`
}

type OpenRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterResponse struct {
	Choices []OpenRouterChoice `json:"choices"`
	Error   *OpenRouterError   `json:"error,omitempty"`
}

type OpenRouterChoice struct {
	Message OpenRouterMessage `json:"message"`
}

type OpenRouterError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (s *OpenRouterService) CreateSessionContext(videoSummary string) string {
	if videoSummary == "" {
		return "Video analysis assistant. Respond in user's language."
	}

	if len(videoSummary) > 8000 {
		videoSummary = videoSummary[:8000] + "\n\n... (text truncated)"
	}

	return fmt.Sprintf("Video context:\n%s\n\nRespond in user's language.", videoSummary)
}

func (s *OpenRouterService) SendMessage(ctx context.Context, session *models.AISession, userMessage string) (string, error) {
	if s.config.APIKey == "" {
		return "", models.ErrAIServiceUnavailable
	}

	systemMessage := s.CreateSessionContext(session.Summary)

	messages := []OpenRouterMessage{
		{Role: "system", Content: systemMessage},
	}

	startIdx := 0
	if len(session.Messages) > 10 {
		startIdx = len(session.Messages) - 10
	}

	for i := startIdx; i < len(session.Messages); i++ {
		messages = append(messages, OpenRouterMessage{
			Role:    session.Messages[i].Role,
			Content: session.Messages[i].Content,
		})
	}

	messages = append(messages, OpenRouterMessage{
		Role:    "user",
		Content: userMessage,
	})

	requestBody := OpenRouterRequest{
		Model:    s.config.Model,
		Messages: messages,
		Stream:   false,
	}

	return s.sendOpenRouterRequest(ctx, requestBody)
}

func (s *OpenRouterService) ImproveSummary(ctx context.Context, currentSummary string, issues []string) (string, error) {
	if s.config.APIKey == "" {
		return "", models.ErrAIServiceUnavailable
	}

	if len(currentSummary) > 6000 {
		currentSummary = currentSummary[:6000] + "\n\n... (text truncated)"
	}

	issuesText := ""
	if len(issues) > 0 {
		issuesText = "Focus on: " + strings.Join(issues, ", ")
	}

	prompt := fmt.Sprintf(`Improve this video summary. Keep original language.

%s

%s

Return only improved summary.`, currentSummary, issuesText)

	messages := []OpenRouterMessage{
		{
			Role:    "system",
			Content: "Improve text quality. Keep original language. Return only improved text.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	requestBody := OpenRouterRequest{
		Model:    s.config.Model,
		Messages: messages,
		Stream:   false,
	}

	return s.sendOpenRouterRequest(ctx, requestBody)
}

func (s *OpenRouterService) FixSummaryErrors(ctx context.Context, currentSummary string) (string, error) {
	if s.config.APIKey == "" {
		return "", models.ErrAIServiceUnavailable
	}

	if len(currentSummary) > 6000 {
		currentSummary = currentSummary[:6000] + "\n\n... (text truncated)"
	}

	prompt := fmt.Sprintf(`Clean errors in this summary. Keep original language.

%s

Remove noise, fix errors. Return only corrected text.`, currentSummary)

	messages := []OpenRouterMessage{
		{
			Role:    "system",
			Content: "Fix text errors. Keep original language. Return only corrected text.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	requestBody := OpenRouterRequest{
		Model:    s.config.Model,
		Messages: messages,
		Stream:   false,
	}

	return s.sendOpenRouterRequest(ctx, requestBody)
}

func (s *OpenRouterService) CreateSummaryFromDialogue(ctx context.Context, messages []models.AIMessage) (string, error) {
	if s.config.APIKey == "" {
		return "", models.ErrAIServiceUnavailable
	}

	var dialogue strings.Builder
	for _, msg := range messages {
		if msg.Role == "user" || msg.Role == "assistant" {
			content := msg.Content
			if len(content) > 1000 {
				content = content[:1000] + "..."
			}
			dialogue.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, content))
		}
	}

	dialogueText := dialogue.String()
	if len(dialogueText) > 6000 {
		dialogueText = dialogueText[:6000] + "\n\n... (dialogue truncated)"
	}

	prompt := fmt.Sprintf(`Create video summary from dialogue. Use dialogue language.

%s

Extract key facts and themes. Return only summary.`, dialogueText)

	messagesReq := []OpenRouterMessage{
		{
			Role:    "system",
			Content: "Create concise summaries. Use same language as input.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	requestBody := OpenRouterRequest{
		Model:    s.config.Model,
		Messages: messagesReq,
		Stream:   false,
	}

	return s.sendOpenRouterRequest(ctx, requestBody)
}

func (s *OpenRouterService) sendOpenRouterRequest(ctx context.Context, requestBody OpenRouterRequest) (string, error) {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	req.Header.Set("HTTP-Referer", "https://vidnotes.app")
	req.Header.Set("X-Title", "VidNotes AI")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp OpenRouterResponse
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error != nil {
			return "", fmt.Errorf("openrouter error: %s", errorResp.Error.Message)
		}
		return "", fmt.Errorf("openrouter API error: %s", string(body))
	}

	var openRouterResp OpenRouterResponse
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(openRouterResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return openRouterResp.Choices[0].Message.Content, nil
}
