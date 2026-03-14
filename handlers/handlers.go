package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"lumina-blog/config"
	"lumina-blog/models"

	"github.com/gin-gonic/gin"
)

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

	if category != "" {
		db = db.Joins("JOIN categories ON categories.id = posts.category_id").Where("categories.slug = ?", category)
	}

	// Count total
	db.Count(&total)

	// Get posts with preloads
	config.DB.Preload("Category").Preload("Tags").
		Where("status = ?", "published").
		Order("created_at desc").
		Offset(offset).Limit(limit).
		Find(&posts)

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetPost returns a single post by slug
func GetPost(c *gin.Context) {
	slug := c.Param("slug")
	var post models.Post

	if err := config.DB.Preload("Category").Preload("Tags").Where("slug = ?", slug).First(&post).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Increment view count
	config.DB.Model(&post).Update("view_count", post.ViewCount+1)

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
