package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"lumina-blog/cache"
	"lumina-blog/config"
	"lumina-blog/models"
	"lumina-blog/services"

	"github.com/gin-gonic/gin"
)

// stripHTML removes HTML tags from content
func stripHTML(html string) string {
	// Replace <br> and </p> with spaces
	re := regexp.MustCompile(`<[^>]+>`)
	text := re.ReplaceAllString(html, "")
	// Clean up whitespace
	text = strings.Join(strings.Fields(text), " ")
	return text
}

// CalculateReadTime estimates reading time in minutes
func CalculateReadTime(content string) int {
	wordsPerMinute := 200
	words := len(strings.Fields(content))
	minutes := words / wordsPerMinute
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}

// GetPosts returns all posts with pagination
func GetPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.DefaultQuery("status", "published")
	category := c.Query("category")
	lang := c.DefaultQuery("lang", "") // Language filter: zh-CN, en, ja, etc.

	// Build cache key
	cacheKey := fmt.Sprintf("posts:page=%d:limit=%d:status=%s:category=%s:lang=%s", 
		page, limit, status, category, lang)
	
	// Try to get from cache
	if cached, found := cache.ArticleCache.Get(cacheKey); found {
		c.JSON(http.StatusOK, cached)
		return
	}

	offset := (page - 1) * limit

	var posts []models.Post
	var total int64

	// Build query
	db := config.DB.Model(&models.Post{})

	// Always filter by published unless explicitly requesting all
	if status == "all" {
		// Show all posts
	} else {
		db = db.Where("status = ?", "published")
	}

	// Filter by language if specified
	if lang != "" {
		db = db.Where("lang = ?", lang)
	}

	if category != "" {
		db = db.Joins("JOIN categories ON categories.id = posts.category_id").Where("categories.slug = ?", category)
	}

	// Count total
	db.Count(&total)

	// Get posts with preloads
	config.DB.Preload("Category").Preload("Tags").
		Order("created_at desc").
		Offset(offset).Limit(limit).
		Find(&posts)

	response := gin.H{
		"posts": posts,
		"total": total,
		"page":  page,
		"limit": limit,
	}
	
	// Cache the response
	cache.ArticleCache.Set(cacheKey, response)
	
	c.JSON(http.StatusOK, response)
}

// GetPost returns a single post by slug
func GetPost(c *gin.Context) {
	slug := c.Param("slug")
	lang := c.Query("lang") // Optional language filter
	
	// Build cache key
	cacheKey := fmt.Sprintf("post:slug=%s:lang=%s", slug, lang)
	
	// Try to get from cache
	if cached, found := cache.ArticleCache.Get(cacheKey); found {
		c.JSON(http.StatusOK, cached)
		return
	}
	
	var post models.Post

	// Build query with optional language filter
	db := config.DB.Preload("Category").Preload("Tags").Where("slug = ?", slug)
	
	if lang != "" {
		db = db.Where("lang = ?", lang)
	}

	if err := db.First(&post).Error; err != nil {
		// If not found with language filter, try without it to find any version
		if lang != "" {
			if err := config.DB.Preload("Category").Preload("Tags").Where("slug = ?", slug).First(&post).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
				return
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
	}

	// Increment view count
	config.DB.Model(&post).Update("view_count", post.ViewCount+1)

	// Add SEO fields
	post.CanonicalURL = config.SEOConfig.SiteURL + "/post/" + post.Slug
	
	// Use excerpt as meta description, or generate from content
	if post.Excerpt != "" {
		post.MetaDescription = post.Excerpt
	} else {
		// Generate from first 160 chars of content (strip markdown/html)
		desc := stripHTML(post.Content)
		if len(desc) > 160 {
			desc = desc[:160]
			if lastSpace := strings.LastIndex(desc, " "); lastSpace > 0 {
				desc = desc[:lastSpace] + "..."
			}
		}
		post.MetaDescription = desc
	}

	// Cache the response
	cache.ArticleCache.Set(cacheKey, post)
	
	c.JSON(http.StatusOK, post)
}

// CreatePost creates a new post
func CreatePost(c *gin.Context) {
	var post models.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate slug from title
	post.Slug = strings.ToLower(strings.ReplaceAll(post.Title, " ", "-"))
	post.ReadTime = CalculateReadTime(post.Content)

	if post.Status == "" {
		post.Status = "draft"
	}

	// Handle tags
	if tags, ok := c.GetPostForm("tags"); ok {
		var tagList []models.Tag
		config.DB.Where("slug IN ?", strings.Split(tags, ",")).Find(&tagList)
		post.Tags = tagList
	}

	config.DB.Create(&post)
	c.JSON(http.StatusCreated, post)
}

// UpdatePost updates an existing post
func UpdatePost(c *gin.Context) {
	id := c.Param("id")
	var post models.Post

	if err := config.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update read time if content changed
	if content, ok := updateData["content"].(string); ok {
		updateData["read_time"] = CalculateReadTime(content)
	}

	config.DB.Model(&post).Updates(updateData)
	c.JSON(http.StatusOK, post)
}

// DeletePost deletes a post
func DeletePost(c *gin.Context) {
	id := c.Param("id")
	var post models.Post

	if err := config.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	config.DB.Delete(&post)
	c.JSON(http.StatusOK, gin.H{"message": "Post deleted"})
}

// GetCategories returns all categories
func GetCategories(c *gin.Context) {
	var categories []models.Category
	config.DB.Find(&categories)
	c.JSON(http.StatusOK, categories)
}

// CreateCategory creates a new category
func CreateCategory(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if category.Slug == "" {
		category.Slug = strings.ToLower(strings.ReplaceAll(category.Name, " ", "-"))
	}

	config.DB.Create(&category)
	c.JSON(http.StatusCreated, category)
}

// DeleteCategory deletes a category
func DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	if err := config.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	config.DB.Delete(&category)
	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}

// GetTags returns all tags
func GetTags(c *gin.Context) {
	var tags []models.Tag
	config.DB.Find(&tags)
	c.JSON(http.StatusOK, tags)
}

// GetStats returns blog statistics
func GetStats(c *gin.Context) {
	var postCount int64
	var categoryCount int64
	var tagCount int64
	var totalViews int64

	config.DB.Model(&models.Post{}).Count(&postCount)
	config.DB.Model(&models.Category{}).Count(&categoryCount)
	config.DB.Model(&models.Tag{}).Count(&tagCount)
	config.DB.Model(&models.Post{}).Select("COALESCE(SUM(view_count), 0)").Row().Scan(&totalViews)

	// Get recent posts
	var recentPosts []models.Post
	config.DB.Preload("Category").Order("created_at desc").Limit(5).Find(&recentPosts)

	c.JSON(http.StatusOK, gin.H{
		"post_count":    postCount,
		"category_count": categoryCount,
		"tag_count":     tagCount,
		"total_views":   totalViews,
		"recent_posts": recentPosts,
	})
}

// TranslatePostRequest represents the request for translating a post
type TranslatePostRequest struct {
	TargetLang string `json:"target_lang" binding:"required"`
}

// TranslatePost translates a post to the target language
func TranslatePost(c *gin.Context) {
	id := c.Param("id")
	var post models.Post

	// Get the original post
	if err := config.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var req TranslatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target language is required"})
		return
	}

	// Check if translated version already exists
	var existingPost models.Post
	err := config.DB.Where("slug = ? AND lang = ?", post.Slug, req.TargetLang).First(&existingPost).Error
	if err == nil {
		// Translated version exists, return it
		config.DB.Preload("Category").Preload("Tags").First(&existingPost)
		c.JSON(http.StatusOK, gin.H{
			"message":     "Translated version already exists",
			"post":        existingPost,
			"is_new":      false,
		})
		return
	}

	// Initialize AI service if not already done
	if aiService == nil {
		aiService = services.NewOpenAIService()
	}

	// Translate title
	translatedTitle, err := aiService.Translate(post.Title, req.TargetLang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to translate title: " + err.Error()})
		return
	}

	// Translate content
	translatedContent, err := aiService.Translate(post.Content, req.TargetLang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to translate content: " + err.Error()})
		return
	}

	// Translate excerpt
	translatedExcerpt := ""
	if post.Excerpt != "" {
		translatedExcerpt, err = aiService.Translate(post.Excerpt, req.TargetLang)
		if err != nil {
			// Use empty excerpt if translation fails
			translatedExcerpt = ""
		}
	}

	// Create new translated post
	newPost := models.Post{
		Title:      translatedTitle,
		Slug:       post.Slug + "-" + req.TargetLang,
		Content:    translatedContent,
		Excerpt:    translatedExcerpt,
		Cover:      post.Cover,
		Status:     "draft", // Translated posts default to draft
		ViewCount:  0,
		ReadTime:   CalculateReadTime(translatedContent),
		Lang:       req.TargetLang,
		CategoryID: post.CategoryID,
	}

	// Copy tags
	newPost.Tags = post.Tags

	// Save the translated post
	if err := config.DB.Create(&newPost).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create translated post: " + err.Error()})
		return
	}

	// Load relations
	config.DB.Preload("Category").Preload("Tags").First(&newPost)

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Post translated successfully",
		"post":     newPost,
		"is_new":   true,
		"language": req.TargetLang,
	})
}

// GetPostByID returns a single post by ID (for admin operations)
func GetPostByID(c *gin.Context) {
	id := c.Param("id")
	var post models.Post

	if err := config.DB.Preload("Category").Preload("Tags").First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, post)
}
