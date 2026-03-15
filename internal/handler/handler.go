package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lumina/blog/internal/model"
	"github.com/lumina/blog/internal/repository"
	"github.com/lumina/blog/internal/service"
)

type Handler struct {
	repo        *repository.Repository
	backupSvc   *service.BackupService
}

func New(repo *repository.Repository, backupSvc *service.BackupService) *Handler {
	return &Handler{
		repo:      repo,
		backupSvc: backupSvc,
	}
}

// Article handlers

func (h *Handler) GetArticle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}

	article, err := h.repo.GetArticleByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	c.JSON(http.StatusOK, article)
}

func (h *Handler) GetArticleBySlug(c *gin.Context) {
	slug := c.Param("slug")
	article, err := h.repo.GetArticleBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	c.JSON(http.StatusOK, article)
}

func (h *Handler) CreateArticle(c *gin.Context) {
	var article model.Article
	if err := c.ShouldBindJSON(&article); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article.CreatedAt = article.UpdatedAt

	id, err := h.repo.CreateArticle(&article)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	article.ID = id
	c.JSON(http.StatusCreated, article)
}

func (h *Handler) UpdateArticle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}

	// Get current article for backup
	article, err := h.repo.GetArticleByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	// Create automatic backup before update
	if _, err := h.backupSvc.AutoBackupBeforeUpdate(article); err != nil {
		// Log but don't fail the update
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create backup: " + err.Error()})
		return
	}

	// Parse and update
	var updateData model.Article
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article.Title = updateData.Title
	article.Content = updateData.Content
	article.Slug = updateData.Slug
	article.Status = updateData.Status
	article.Author = updateData.Author

	if err := h.repo.UpdateArticle(article); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, article)
}

func (h *Handler) DeleteArticle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}

	if err := h.repo.DeleteArticle(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "article deleted"})
}

// Backup handlers

func (h *Handler) ListBackups(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.backupSvc.GetBackupList(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetBackup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid backup id"})
		return
	}

	result, err := h.backupSvc.GetBackupDetail(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "backup not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) RestoreBackup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid backup id"})
		return
	}

	var req model.RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "operator is required"})
		return
	}

	article, preBackup, err := h.backupSvc.RestoreFromBackup(id, req.Operator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "restore successful",
		"article":          article,
		"pre_restore_backup_id": preBackup.ID,
	})
}

func (h *Handler) DeleteBackup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid backup id"})
		return
	}

	if err := h.backupSvc.DeleteBackup(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "backup deleted"})
}

// Health check
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
