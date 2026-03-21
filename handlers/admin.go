package handlers

import (
	"net/http"
	"strconv"

	"lumina-blog/config"
	"lumina-blog/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminHandlers handles admin panel API endpoints

// AdminGetStats returns blog statistics (admin)
func AdminGetStats(c *gin.Context) {
	var stats models.StatsResponse

	// Count posts
	config.DB.Model(&models.Post{}).Count(&stats.TotalPosts)
	config.DB.Model(&models.Post{}).Where("status = ?", "published").Count(&stats.PublishedPosts)
	config.DB.Model(&models.Post{}).Where("status = ?", "draft").Count(&stats.DraftPosts)

	// Count comments
	config.DB.Model(&models.Comment{}).Count(&stats.TotalComments)
	config.DB.Model(&models.Comment{}).Where("status = ?", "pending").Count(&stats.PendingComments)

	// Count total views
	var posts []models.Post
	config.DB.Find(&posts)
	var totalViews int64
	for _, post := range posts {
		totalViews += int64(post.ViewCount)
	}
	stats.TotalViews = totalViews

	// Count likes
	config.DB.Model(&models.Like{}).Count(&stats.TotalLikes)

	// Count favorites
	config.DB.Model(&models.Favorite{}).Count(&stats.TotalFavorites)

	c.JSON(http.StatusOK, stats)
}

// AdminPostList returns all posts (admin view)
func AdminPostList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	offset := (page - 1) * limit

	var posts []models.Post
	var total int64

	query := config.DB.Model(&models.Post{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	query.Preload("Category").Preload("Tags").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&posts)

	c.JSON(http.StatusOK, gin.H{
		"posts":     posts,
		"total":     total,
		"page":      page,
		"page_size": limit,
	})
}

// AdminDeletePost deletes a post (admin)
func AdminDeletePost(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var post models.Post
	if err := config.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Delete related data
	config.DB.Where("post_id = ?", postID).Delete(&models.Comment{})
	config.DB.Where("post_id = ?", postID).Delete(&models.Like{})
	config.DB.Where("post_id = ?", postID).Delete(&models.Favorite{})

	// Delete post
	if err := config.DB.Delete(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

// AdminCommentList returns all comments (admin view)
func AdminCommentList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	postID := c.Query("post_id")

	offset := (page - 1) * limit

	var comments []models.Comment
	var total int64

	query := config.DB.Model(&models.Comment{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if postID != "" {
		query = query.Where("post_id = ?", postID)
	}

	query.Count(&total)

	query.Preload("Post", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, title, slug")
	}).Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, nickname, avatar")
	}).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&comments)

	c.JSON(http.StatusOK, gin.H{
		"comments":  comments,
		"total":    total,
		"page":     page,
		"page_size": limit,
	})
}

// AdminApproveComment approves a comment
func AdminApproveComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	var comment models.Comment
	if err := config.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	if err := config.DB.Model(&comment).Update("status", "approved").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment approved"})
}

// AdminRejectComment rejects a comment
func AdminRejectComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	var comment models.Comment
	if err := config.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	if err := config.DB.Model(&comment).Update("status", "rejected").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment rejected"})
}

// AdminDeleteComment deletes a comment
func AdminDeleteComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	var comment models.Comment
	if err := config.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		return
	}

	// Also delete replies
	config.DB.Where("parent_id = ?", commentID).Delete(&models.Comment{})

	if err := config.DB.Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
}

// AdminUpdatePost updates a post (admin)
func AdminUpdatePost(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var post models.Post
	if err := config.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Recalculate read time if content changed
	if _, hasContent := updates["content"]; hasContent {
		if content, ok := updates["content"].(string); ok {
			updates["read_time"] = CalculateReadTime(content)
		}
	}

	if err := config.DB.Model(&post).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}

	config.DB.Preload("Category").Preload("Tags").First(&post, postID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Post updated successfully",
		"post":    post,
	})
}
