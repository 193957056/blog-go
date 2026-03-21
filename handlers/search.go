package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"lumina-blog/config"
	"lumina-blog/models"

	"github.com/gin-gonic/gin"
)

// SearchHandles handles search-related endpoints

// SearchResult represents a search result with highlighted content
type SearchResult struct {
	models.Post
	HighlightTitle   string `json:"highlight_title"`
	HighlightContent string `json:"highlight_content"`
}

// Search searches posts by keyword
func Search(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search keyword is required"})
		return
	}

	keyword = strings.TrimSpace(keyword)
	if len(keyword) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search keyword must be at least 2 characters"})
		return
	}

	limit := 20 // Default limit

	// Search in published posts only (case-insensitive)
	var posts []models.Post
	keywordLower := strings.ToLower(keyword)
	
	// Use raw SQL to avoid GORM issues
	if err := config.DB.Raw(`
		SELECT * FROM posts 
		WHERE status = 'published' 
		AND (LOWER(title) LIKE ? OR LOWER(content) LIKE ?)
		ORDER BY view_count DESC, created_at DESC
		LIMIT ?
	`, "%"+keywordLower+"%", "%"+keywordLower+"%", limit).
		Preload("Category").
		Preload("Tags").
		Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	// Process results with highlighting
	results := make([]SearchResult, len(posts))
	for i, post := range posts {
		results[i] = SearchResult{
			Post:             post,
			HighlightTitle:   highlightText(post.Title, keyword),
			HighlightContent: highlightContent(post.Content, keyword),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"results":    results,
		"total":      len(results),
		"keyword":    keyword,
		"message":    fmt.Sprintf("Found %d results for '%s'", len(results), keyword),
	})
}

// highlightText adds <mark> tags around matched keywords in title
func highlightText(text, keyword string) string {
	// Create case-insensitive regex
	regex, err := regexp.Compile("(?i)" + regexp.QuoteMeta(keyword))
	if err != nil {
		return text
	}
	return regex.ReplaceAllString(text, "<mark>$0</mark>")
}

// highlightContent extracts relevant excerpt with highlighting
func highlightContent(content string, keyword string) string {
	// Remove markdown headers and get plain text
	plainText := removeMarkdown(content)
	
	// Truncate to a reasonable excerpt length
	maxLength := 300
	if len(plainText) <= maxLength {
		return highlightText(plainText, keyword)
	}

	// Find the keyword position
	lowerContent := strings.ToLower(plainText)
	lowerKeyword := strings.ToLower(keyword)
	pos := strings.Index(lowerContent, lowerKeyword)

	var start, end int
	if pos == -1 {
		// Keyword not found, just return first chunk
		start = 0
		end = maxLength
	} else {
		// Try to center the keyword in the excerpt
		halfLength := maxLength / 2
		start = pos - halfLength
		end = pos + len(keyword) + halfLength

		if start < 0 {
			start = 0
			end = maxLength
		}
		if end > len(plainText) {
			end = len(plainText)
			start = end - maxLength
			if start < 0 {
				start = 0
			}
		}
	}

	excerpt := plainText[start:end]
	
	// Add ellipsis
	if start > 0 {
		excerpt = "..." + excerpt
	}
	if end < len(plainText) {
		excerpt = excerpt + "..."
	}

	return highlightText(excerpt, keyword)
}

// removeMarkdown removes markdown formatting for plain text search
func removeMarkdown(text string) string {
	// Remove headers (# ## ###)
	text = regexp.MustCompile(`(?m)^#{1,6}\s+`).ReplaceAllString(text, "")
	
	// Remove bold/italic (**text** or *text*)
	text = regexp.MustCompile(`\*\*(.+?)\*\*`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`\*(.+?)\*`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`__(.+?)__`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`_(.+?)_`).ReplaceAllString(text, "$1")
	
	// Remove code blocks (```code```)
	text = regexp.MustCompile("```[\\s\\S]*?```").ReplaceAllString(text, "")
	
	// Remove inline code (`code`)
	text = regexp.MustCompile("`(.+?)`").ReplaceAllString(text, "$1")
	
	// Remove links [text](url)
	text = regexp.MustCompile(`\[(.+?)\]\(.+?\)`).ReplaceAllString(text, "$1")
	
	// Remove images ![alt](url)
	text = regexp.MustCompile(`!\[.*?\]\(.+?\)`).ReplaceAllString(text, "")
	
	// Remove horizontal rules
	text = regexp.MustCompile(`(?m)^[-*_]{3,}$`).ReplaceAllString(text, "")
	
	// Remove extra whitespace and newlines
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	
	return strings.TrimSpace(text)
}
