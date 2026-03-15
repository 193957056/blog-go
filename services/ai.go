package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// AIRequest represents the request to AI API
type AIRequest struct {
	Model    string   `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool     `json:"stream"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIResponse represents the response from AI API
type AIResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

type AIService interface {
	Polish(text string) (string, error)
	Summary(content string) (string, error)
	SEOAnalysis(title, content string) (map[string]interface{}, error)
	Translate(text, targetLang string) (string, error)
}

// OpenAIService implements AIService using OpenAI-compatible API
type OpenAIService struct {
	BaseURL  string
	APIKey   string
	Model    string
}

// NewOpenAIService creates a new OpenAI-compatible AI service
func NewOpenAIService() *OpenAIService {
	baseURL := os.Getenv("AI_API_URL")
	apiKey := os.Getenv("AI_API_KEY")
	model := os.Getenv("AI_MODEL")

	// Default values if not set
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	return &OpenAIService{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
	}
}

// CallAI makes a call to the AI API
func (s *OpenAIService) CallAI(prompt string) (string, error) {
	if s.APIKey == "" {
		return "", fmt.Errorf("AI_API_KEY is not set")
	}

	messages := []Message{
		{Role: "user", Content: prompt},
	}

	reqBody := AIRequest{
		Model:    s.Model,
		Messages: messages,
		Stream:   false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.BaseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var aiResp AIResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(aiResp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return aiResp.Choices[0].Message.Content, nil
}

// Polish improves the text using AI
func (s *OpenAIService) Polish(text string) (string, error) {
	prompt := fmt.Sprintf(`请润色以下文本，使其更加流畅、专业。直接返回润色后的文本，不要添加任何解释或额外内容：

%s`, text)

	return s.CallAI(prompt)
}

// Summary generates a summary of the content
func (s *OpenAIService) Summary(content string) (string, error) {
	// Truncate content if too long
	maxLen := 8000
	if len(content) > maxLen {
		content = content[:maxLen] + "..."
	}

	prompt := fmt.Sprintf(`请为以下文章生成一个简洁的摘要（100字以内），直接返回摘要，不要添加任何解释：

%s`, content)

	return s.CallAI(prompt)
}

// SEOAnalysis performs SEO analysis on title and content
func (s *OpenAIService) SEOAnalysis(title, content string) (map[string]interface{}, error) {
	// Truncate content if too long
	maxLen := 5000
	if len(content) > maxLen {
		content = content[:maxLen] + "..."
	}

	prompt := fmt.Sprintf(`请对以下文章进行SEO分析，直接返回JSON格式的分析结果，不要添加任何解释：

{
  "score": 分数(0-100),
  "suggestions": ["建议1", "建议2", "建议3"],
  "keywords": ["关键词1", "关键词2", "关键词3"]
}

文章标题: %s
文章内容: %s`, title, content)

	result, err := s.CallAI(prompt)
	if err != nil {
		return nil, err
	}

	// Try to parse as JSON
	result = strings.TrimSpace(result)
	// Remove markdown code blocks if present
	result = strings.TrimPrefix(result, "```json")
	result = strings.TrimPrefix(result, "```")
	result = strings.TrimSuffix(result, "```")

	var seoResult map[string]interface{}
	if err := json.Unmarshal([]byte(result), &seoResult); err != nil {
		// If parsing fails, return basic result
		return map[string]interface{}{
			"score":       70,
			"suggestions": []string{"无法解析SEO建议", "建议检查标题和内容"},
			"keywords":    []string{},
		}, nil
	}

	return seoResult, nil
}

// Translate translates text to target language
func (s *OpenAIService) Translate(text, targetLang string) (string, error) {
	prompt := fmt.Sprintf(`请将以下文本翻译成%s，直接返回翻译后的文本，不要添加任何解释：

%s`, targetLang, text)

	return s.CallAI(prompt)
}
