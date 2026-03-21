package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"lumina-blog/config"
	"lumina-blog/models"

	"github.com/gin-gonic/gin"
)

// SEOAnalyzeRequest represents the request for SEO analysis via query params
type SEOAnalyzeRequest struct {
	Title   string `form:"title" binding:"required"`
	Content string `form:"content" binding:"required"`
}

// SEOAnalysisResult represents the SEO analysis result
type SEOAnalysisResult struct {
	TitleScore         int      `json:"titleScore"`
	DescriptionScore   int      `json:"descriptionScore"`
	KeywordDensity     float64  `json:"keywordDensity"`
	ReadabilityScore   int      `json:"readabilityScore"`
	Suggestions        []string `json:"suggestions"`
	SchemaSuggestion   string   `json:"schemaSuggestion"`
}

// SEOAPIResponse represents the API response
type SEOAPIResponse struct {
	Success bool             `json:"success"`
	Data    *SEOAnalysisResult `json:"data,omitempty"`
	Error   string           `json:"error,omitempty"`
}

// extractKeyword extracts the main keyword from title (first significant word)
func extractKeyword(title string) string {
	// Remove common stop words and get first meaningful word
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "is": true, "are": true, "was": true,
		"were": true, "be": true, "been": true, "being": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "must": true,
		"how": true, "what": true, "why": true, "when": true, "where": true, "who": true,
		"this": true, "that": true, "these": true, "those": true, "it": true,
	}

	words := strings.Fields(strings.ToLower(title))
	for _, word := range words {
		cleaned := strings.TrimFunc(word, func(r rune) bool {
			return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
		})
		if len(cleaned) > 2 && !stopWords[cleaned] {
			return cleaned
		}
	}
	// If no keyword found, use the first word
	if len(words) > 0 {
		return strings.TrimFunc(words[0], func(r rune) bool {
			return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
		})
	}
	return ""
}

// calculateTitleScore calculates score based on title quality
func calculateTitleScore(title string, keyword string) int {
	score := 0
	
	// Length check (optimal: 30-60 chars)
	length := len(title)
	if length >= 30 && length <= 60 {
		score += 40
	} else if length >= 20 && length <= 70 {
		score += 20
	} else if length > 0 {
		score += 10
	}
	
	// Keyword presence in title
	if keyword != "" && strings.Contains(strings.ToLower(title), keyword) {
		score += 40
	}
	
	// Has power words or numbers
	powerWords := []string{"best", "top", "guide", "how", "tips", "review", "tutorial"}
	hasPowerWord := false
	for _, pw := range powerWords {
		if strings.Contains(strings.ToLower(title), pw) {
			hasPowerWord = true
			break
		}
	}
	if hasPowerWord {
		score += 20
	}
	
	// Min score 0, max 100
	if score > 100 {
		score = 100
	}
	return score
}

// calculateDescriptionScore calculates score for meta description
func calculateDescriptionScore(content string, keyword string) int {
	score := 0
	
	// Get first paragraph as description proxy
	firstPara := strings.Split(content, "\n")[0]
	if len(firstPara) > 200 {
		firstPara = firstPara[:200]
	}
	
	descLength := len(firstPara)
	
	// Length check (optimal: 120-160 chars for meta description, using first para as proxy)
	if descLength >= 120 && descLength <= 160 {
		score += 50
	} else if descLength >= 80 && descLength <= 200 {
		score += 30
	} else if descLength > 0 {
		score += 10
	}
	
	// Keyword presence
	if keyword != "" && strings.Contains(strings.ToLower(firstPara), keyword) {
		score += 50
	}
	
	if score > 100 {
		score = 100
	}
	return score
}

// calculateKeywordDensity calculates keyword density in content
func calculateKeywordDensity(content string, keyword string) float64 {
	if keyword == "" || len(content) == 0 {
		return 0
	}
	
	contentLower := strings.ToLower(content)
	keywordLower := strings.ToLower(keyword)
	
	// Count keyword occurrences
	count := strings.Count(contentLower, keywordLower)
	
	// Count total words
	words := strings.Fields(contentLower)
	totalWords := len(words)
	
	if totalWords == 0 {
		return 0
	}
	
	density := (float64(count) / float64(totalWords)) * 100
	return density
}

// calculateReadabilityScore calculates readability score (simplified Flesch-like)
func calculateReadabilityScore(content string) int {
	words := strings.Fields(content)
	if len(words) == 0 {
		return 0
	}
	
	// Count sentences (approximate by punctuation)
	sentences := 0
	for _, r := range content {
		if r == '.' || r == '!' || r == '?' {
			sentences++
		}
	}
	if sentences == 0 {
		sentences = 1
	}
	
	// Average words per sentence
	avgWordsPerSentence := float64(len(words)) / float64(sentences)
	
	// Calculate average word length
	totalCharCount := 0
	for _, word := range words {
		totalCharCount += len(word)
	}
	avgWordLength := float64(totalCharCount) / float64(len(words))
	
	// Simplified readability score
	// Lower avg words per sentence and shorter words = more readable
	score := 100
	
	// Penalize long sentences
	if avgWordsPerSentence > 20 {
		score -= 30
	} else if avgWordsPerSentence > 15 {
		score -= 15
	} else if avgWordsPerSentence > 10 {
		score -= 5
	}
	
	// Penalize long words
	if avgWordLength > 6 {
		score -= 25
	} else if avgWordLength > 5 {
		score -= 10
	}
	
	if score < 0 {
		score = 0
	}
	return score
}

// generateSuggestions generates optimization suggestions
func generateSuggestions(title string, content string, keyword string, density float64, titleScore int, descScore int, readScore int) []string {
	var suggestions []string
	
	// Title suggestions
	if titleScore < 50 {
		if len(title) < 30 {
			suggestions = append(suggestions, "Title is too short. Aim for 30-60 characters.")
		} else if len(title) > 60 {
			suggestions = append(suggestions, "Title is too long. Keep it under 60 characters.")
		}
		if keyword != "" && !strings.Contains(strings.ToLower(title), keyword) {
			suggestions = append(suggestions, "Include your target keyword in the title.")
		}
	}
	
	// Description suggestions
	if descScore < 50 {
		suggestions = append(suggestions, "The first paragraph should be 120-160 characters. Include your keyword naturally.")
	}
	
	// Keyword density suggestions
	if density == 0 && keyword != "" {
		suggestions = append(suggestions, "Keyword not found in content. Include your target keyword naturally throughout the text.")
	} else if density < 0.5 {
		suggestions = append(suggestions, "Keyword density is low ("+fmt.Sprintf("%.1f", density)+"%). Consider using the keyword more naturally.")
	} else if density > 3 {
		suggestions = append(suggestions, "Keyword density is too high ("+fmt.Sprintf("%.1f", density)+"%). Reduce keyword usage to avoid over-optimization.")
	}
	
	// Readability suggestions
	if readScore < 50 {
		suggestions = append(suggestions, "Content may be difficult to read. Try shorter sentences and simpler words.")
	}
	
	// Content length suggestions
	wordCount := len(strings.Fields(content))
	if wordCount < 300 {
		suggestions = append(suggestions, "Content is short ("+fmt.Sprintf("%d", wordCount)+" words). Consider adding more valuable information (300+ words recommended).")
	}
	
	// Add positive suggestions if score is good
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Great job! Your content is well-optimized for SEO.")
	}
	
	return suggestions
}

// generateSchemaSuggestion generates schema.org structured data suggestion
func generateSchemaSuggestion(title string, content string) string {
	// Basic Article schema suggestion
	keyword := extractKeyword(title)
	
	schema := `{
  "@context": "https://schema.org",
  "@type": "Article",
  "headline": "` + escapeJSON(title) + `",
  "description": "` + escapeJSON(strings.Split(content, "\n")[0]) + `",
  "author": {
    "@type": "Person",
    "name": "Author Name"
  },
  "datePublished": "` + getCurrentDate() + `",
  "dateModified": "` + getCurrentDate() + `",
  "articleSection": "Blog",
  "keywords": "` + keyword + `"
}`
	
	return schema
}

// escapeJSON escapes special characters for JSON
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// getCurrentDate returns current date in ISO format
func getCurrentDate() string {
	return time.Now().Format("2006-01-02")
}

// SEOAnalyze handles the GET /api/seo/analyze endpoint
func SEOAnalyze(c *gin.Context) {
	var req SEOAnalyzeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, SEOAPIResponse{
			Success: false,
			Error:   "title and content query parameters are required",
		})
		return
	}
	
	// Extract keyword from title
	keyword := extractKeyword(req.Title)
	
	// Calculate scores
	titleScore := calculateTitleScore(req.Title, keyword)
	descScore := calculateDescriptionScore(req.Content, keyword)
	density := calculateKeywordDensity(req.Content, keyword)
	readScore := calculateReadabilityScore(req.Content)
	
	// Generate suggestions
	suggestions := generateSuggestions(req.Title, req.Content, keyword, density, titleScore, descScore, readScore)
	
	// Generate schema suggestion
	schemaSuggestion := generateSchemaSuggestion(req.Title, req.Content)
	
	// Build result
	result := &SEOAnalysisResult{
		TitleScore:       titleScore,
		DescriptionScore: descScore,
		KeywordDensity:   density,
		ReadabilityScore: readScore,
		Suggestions:      suggestions,
		SchemaSuggestion: schemaSuggestion,
	}
	
	c.JSON(http.StatusOK, SEOAPIResponse{
		Success: true,
		Data:    result,
	})
}

// GetBaiduTongjiID returns the Baidu Tongji (统计) ID for frontend use
func GetBaiduTongjiID(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"baidu_tongji_id": config.SEOConfig.BaiduTongjiID,
		"enabled":         config.SEOConfig.BaiduTongjiID != "",
	})
}

// GetSiteConfig returns site configuration for SEO
func GetSiteConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"site_name":        config.SEOConfig.SiteName,
		"site_url":         config.SEOConfig.SiteURL,
		"description":      config.SEOConfig.Description,
		"baidu_tongji_id":  config.SEOConfig.BaiduTongjiID,
		"google_analytics": config.SEOConfig.GoogleAnalytics,
	})
}

// GenerateSitemapXML generates XML sitemap for search engines
func GenerateSitemapXML(c *gin.Context) {
	var posts []models.Post
	config.DB.Where("status = ?", "published").Find(&posts)
	
	var sitemap strings.Builder
	sitemap.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sitemap.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	
	// Add static pages
	staticPages := []struct{
		loc       string
		priority  string
	}{
		{config.SEOConfig.SiteURL + "/", "1.0"},
		{config.SEOConfig.SiteURL + "/posts", "0.9"},
		{config.SEOConfig.SiteURL + "/categories", "0.7"},
		{config.SEOConfig.SiteURL + "/tags", "0.7"},
		{config.SEOConfig.SiteURL + "/about", "0.6"},
	}
	
	for _, page := range staticPages {
		sitemap.WriteString(`  <url>` + "\n")
		sitemap.WriteString(`    <loc>` + page.loc + `</loc>` + "\n")
		sitemap.WriteString(`    <changefreq>weekly</changefreq>` + "\n")
		sitemap.WriteString(`    <priority>` + page.priority + `</priority>` + "\n")
		sitemap.WriteString(`  </url>` + "\n")
	}
	
	// Add blog posts
	for _, post := range posts {
		postURL := config.SEOConfig.SiteURL + "/post/" + post.Slug
		changefreq := "weekly"
		if post.ViewCount > 1000 {
			changefreq = "daily"
		}
		
		sitemap.WriteString(`  <url>` + "\n")
		sitemap.WriteString(`    <loc>` + postURL + `</loc>` + "\n")
		sitemap.WriteString(`    <lastmod>` + post.UpdatedAt.Format("2006-01-02") + `</lastmod>` + "\n")
		sitemap.WriteString(`    <changefreq>` + changefreq + `</changefreq>` + "\n")
		sitemap.WriteString(`    <priority>0.8</priority>` + "\n")
		sitemap.WriteString(`  </url>` + "\n")
	}
	
	sitemap.WriteString(`</urlset>`)
	
	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, sitemap.String())
}

// GenerateRobotsTxt generates robots.txt content
func GenerateRobotsTxt(c *gin.Context) {
	var robots strings.Builder
	robots.WriteString("User-agent: *\n")
	robots.WriteString("Allow: /\n")
	robots.WriteString("\n")
	
	// Disallow admin and API paths
	robots.WriteString("Disallow: /api/admin/\n")
	robots.WriteString("Disallow: /api/auth/\n")
	robots.WriteString("Disallow: /api/backups/\n")
	robots.WriteString("Disallow: /api/comments\n")
	robots.WriteString("\n")
	
	// Sitemap location
	robots.WriteString("Sitemap: " + config.SEOConfig.SiteURL + "/sitemap.xml\n")
	
	// Baidu-specific
	robots.WriteString("\n")
	robots.WriteString("# Baidu Bot\n")
	robots.WriteString("User-agent: Baiduspider\n")
	robots.WriteString("Allow: /\n")
	
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, robots.String())
}
