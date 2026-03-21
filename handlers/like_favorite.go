package handlers

import (
	"net/http"
	"strconv"

	"lumina-blog/config"
	"lumina-blog/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LikeFavoriteHandlers handles like and favorite endpoints

// LikePost likes a post
func LikePost(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Check if post exists
	var post models.Post
	if err := config.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Check if already liked
	var existingLike models.Like
	query := config.DB.Where("post_id = ? AND user_id = ?", postID, userID)
	
	// For anonymous likes via IP
	ip := c.ClientIP()
	if err := query.First(&existingLike).Error; err == nil {
		// Already liked, remove like (toggle)
		config.DB.Delete(&existingLike)
		
		// Decrement like count in post
		if post.ViewCount > 0 {
			config.DB.Model(&post).Update("view_count", post.ViewCount-1)
		}
		
		c.JSON(http.StatusOK, gin.H{"liked": false, "message": "Like removed"})
		return
	}

	// Create new like
	like := models.Like{
		PostID: uint(postID),
		UserID: uint(userID.(uint)),
		IP:     ip,
	}

	if err := config.DB.Create(&like).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like post"})
		return
	}

	// Note: We don't increment view_count for likes to avoid confusion
	// view_count is for actual page views

	c.JSON(http.StatusCreated, gin.H{"liked": true, "message": "Post liked successfully"})
}

// GetPostLikes returns like count and user's like status
func GetPostLikes(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	// Count likes
	var likeCount int64
	config.DB.Model(&models.Like{}).Where("post_id = ?", postID).Count(&likeCount)

	// Check if current user liked
	var userLiked bool
	if userID, exists := c.Get("user_id"); exists {
		var like models.Like
		if err := config.DB.Where("post_id = ? AND user_id = ?", postID, userID).First(&like).Error; err == nil {
			userLiked = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"likes":    likeCount,
		"userLiked": userLiked,
	})
}

// FavoritePost favorites a post
func FavoritePost(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Check if post exists
	var post models.Post
	if err := config.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Check if already favorited
	var existingFavorite models.Favorite
	if err := config.DB.Where("post_id = ? AND user_id = ?", postID, userID).First(&existingFavorite).Error; err == nil {
		// Already favorited, remove favorite (toggle)
		config.DB.Delete(&existingFavorite)
		c.JSON(http.StatusOK, gin.H{"favorited": false, "message": "Favorite removed"})
		return
	}

	// Create new favorite
	favorite := models.Favorite{
		PostID: uint(postID),
		UserID: uint(userID.(uint)),
	}

	if err := config.DB.Create(&favorite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to favorite post"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"favorited": true, "message": "Post favorited successfully"})
}

// GetPostFavorites returns favorite count and user's favorite status
func GetPostFavorites(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	// Count favorites
	var favoriteCount int64
	config.DB.Model(&models.Favorite{}).Where("post_id = ?", postID).Count(&favoriteCount)

	// Check if current user favorited
	var userFavorited bool
	if userID, exists := c.Get("user_id"); exists {
		var favorite models.Favorite
		if err := config.DB.Where("post_id = ? AND user_id = ?", postID, userID).First(&favorite).Error; err == nil {
			userFavorited = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"favorites":    favoriteCount,
		"userFavorited": userFavorited,
	})
}

// GetUserLikes returns all posts liked by current user
func GetUserLikes(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var likes []models.Like
	if err := config.DB.
		Preload("Post", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, title, slug, cover, status")
		}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&likes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch likes"})
		return
	}

	// Extract posts from likes
	posts := make([]models.Post, 0, len(likes))
	for _, like := range likes {
		if like.Post.ID != 0 {
			posts = append(posts, like.Post)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"likes": posts,
		"total": len(posts),
	})
}

// GetUserFavorites returns all posts favorited by current user
func GetUserFavorites(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var favorites []models.Favorite
	if err := config.DB.
		Preload("Post", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, title, slug, cover, status")
		}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&favorites).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch favorites"})
		return
	}

	// Extract posts from favorites
	posts := make([]models.Post, 0, len(favorites))
	for _, fav := range favorites {
		if fav.Post.ID != 0 {
			posts = append(posts, fav.Post)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"favorites": posts,
		"total":    len(posts),
	})
}
