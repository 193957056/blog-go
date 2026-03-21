package handlers

import (
	"net/http"
	"strconv"

	"lumina-blog/config"
	"lumina-blog/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)
// CommentHandlers handles comment-related endpoints

// GetCommentsByPost returns all approved comments for a post
func GetCommentsByPost(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var comments []models.Comment
	if err := config.DB.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nickname, avatar")
		}).
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User", func(db *gorm.DB) *gorm.DB {
				return db.Select("id, nickname, avatar")
			}).Where("status = ?", "approved").Order("created_at ASC")
		}).
		Where("post_id = ? AND status = ? AND parent_id IS NULL", postID, "approved").
		Order("created_at DESC").
		Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
		"total":    len(comments),
	})
}

// CreateComment creates a new comment
func CreateComment(c *gin.Context) {
	var req models.CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from context if authenticated
	var userID *uint
	var author, email string

	if userIDVal, exists := c.Get("user_id"); exists {
		id := uint(userIDVal.(uint))
		userID = &id
		
		// Get user info
		var user models.User
		if err := config.DB.First(&user, id).Error; err == nil {
			author = user.Nickname
			email = user.Email
		}
	} else {
		// Anonymous comment - require author and email
		if req.Author == "" || req.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Author and email are required for anonymous comments"})
			return
		}
		author = req.Author
		email = req.Email
	}

	// Check if post exists
	var post models.Post
	if err := config.DB.First(&post, req.PostID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// If replying to another comment, verify parent exists
	if req.ParentID != nil {
		var parentComment models.Comment
		if err := config.DB.First(&parentComment, *req.ParentID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Parent comment not found"})
			return
		}
		// Only allow one level of nesting
		if parentComment.ParentID != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot reply to a nested comment"})
			return
		}
	}

	// Get IP and User-Agent
	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	comment := models.Comment{
		PostID:    req.PostID,
		ParentID:  req.ParentID,
		UserID:    userID,
		Author:    author,
		Email:     email,
		Content:   req.Content,
		Status:    "pending", // Require approval by default
		IP:        ip,
		UserAgent: userAgent,
	}

	if err := config.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	// Preload user for response
	if comment.UserID != nil {
		config.DB.Preload("User").First(&comment, comment.ID)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Comment created successfully, pending approval",
		"comment": comment,
	})
}

// DeleteComment deletes a comment (admin or owner)
func DeleteComment(c *gin.Context) {
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

	// Check if user is admin or comment owner
	if userID, exists := c.Get("user_id"); exists {
		var user models.User
		if err := config.DB.First(&user, userID).Error; err == nil {
			if user.Role != "admin" && (comment.UserID == nil || *comment.UserID != user.ID) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
				return
			}
		}
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		return
	}

	// Soft delete
	if err := config.DB.Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}

// ApproveComment approves a pending comment
func ApproveComment(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{"message": "Comment approved successfully"})
}

// GetAllComments returns all comments (admin)
func GetAllComments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	offset := (page - 1) * limit

	var comments []models.Comment
	var total int64

	query := config.DB.Model(&models.Comment{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	query.Preload("Post", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, title")
	}).Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, nickname, avatar")
	}).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&comments)

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}
