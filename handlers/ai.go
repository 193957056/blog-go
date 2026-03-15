package handlers

import (
	"net/http"

	"lumina-blog/services"

	"github.com/gin-gonic/gin"
)

// AIPolishRequest represents the request for text polishing
type AIPolishRequest struct {
	Text string `json:"text" binding:"required"`
}

// AISummaryRequest represents the request for summary
type AISummaryRequest struct {
	Content string `json:"content" binding:"required"`
}

// AISEORequest represents the request for SEO analysis
type AISEORequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// AITranslateRequest represents the request for translation
type AITranslateRequest struct {
	Text      string `json:"text" binding:"required"`
	TargetLang string `json:"target_lang" binding:"required"`
}

// AIResponse represents a unified API response
type AIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Global AI service instance
var aiService services.AIService

// InitAIService initializes the AI service
func InitAIService() {
	aiService = services.NewOpenAIService()
}

// AIPolish handles the text polishing request
func AIPolish(c *gin.Context) {
	var req AIPolishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AIResponse{
			Success: false,
			Error:   "Text is required",
		})
		return
	}

	result, err := aiService.Polish(req.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AIResponse{
		Success: true,
		Data: map[string]string{
			"polished": result,
		},
	})
}

// AISummary handles the summary generation request
func AISummary(c *gin.Context) {
	var req AISummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AIResponse{
			Success: false,
			Error:   "Content is required",
		})
		return
	}

	result, err := aiService.Summary(req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AIResponse{
		Success: true,
		Data: map[string]string{
			"summary": result,
		},
	})
}

// AISEOSuggestion handles the SEO analysis request
func AISEOSuggestion(c *gin.Context) {
	var req AISEORequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AIResponse{
			Success: false,
			Error:   "Title and content are required",
		})
		return
	}

	result, err := aiService.SEOAnalysis(req.Title, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AIResponse{
		Success: true,
		Data:    result,
	})
}

// AITranslate handles the translation request
func AITranslate(c *gin.Context) {
	var req AITranslateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AIResponse{
			Success: false,
			Error:   "Text and target language are required",
		})
		return
	}

	result, err := aiService.Translate(req.Text, req.TargetLang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AIResponse{
		Success: true,
		Data: map[string]string{
			"translated": result,
			"language":  req.TargetLang,
		},
	})
}
