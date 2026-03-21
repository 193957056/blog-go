package handlers

import (
	"net/http"
	"strconv"

	"lumina-blog/config"
	"lumina-blog/models"

	"github.com/gin-gonic/gin"
)

// LinkHandlers handles link-related endpoints

// GetLinks returns all approved links (public)
func GetLinks(c *gin.Context) {
	var links []models.Link
	if err := config.DB.
		Where("status = ?", "approved").
		Order("order ASC, created_at DESC").
		Find(&links).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch links"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"links": links,
		"total": len(links),
	})
}

// GetAllLinks returns all links (admin)
func GetAllLinks(c *gin.Context) {
	var links []models.Link
	if err := config.DB.
		Order("order ASC, created_at DESC").
		Find(&links).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch links"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"links": links,
		"total": len(links),
	})
}

// CreateLink creates a new link (admin)
func CreateLink(c *gin.Context) {
	var req models.LinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link := models.Link{
		Name:      req.Name,
		URL:       req.URL,
		Logo:      req.Logo,
		Desc:      req.Desc,
		Email:     req.Email,
		ContactQQ: req.ContactQQ,
		Status:    "approved", // Admin created links are auto-approved
	}

	if err := config.DB.Create(&link).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create link"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Link created successfully",
		"link":    link,
	})
}

// UpdateLink updates a link (admin)
func UpdateLink(c *gin.Context) {
	linkID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid link ID"})
		return
	}

	var link models.Link
	if err := config.DB.First(&link, linkID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	var req struct {
		Name      string `json:"name"`
		URL       string `json:"url"`
		Logo      string `json:"logo"`
		Desc      string `json:"desc"`
		Email     string `json:"email"`
		Order     int    `json:"order"`
		Status    string `json:"status"`
		ContactQQ string `json:"contact_qq"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.URL != "" {
		updates["url"] = req.URL
	}
	if req.Logo != "" {
		updates["logo"] = req.Logo
	}
	if req.Desc != "" {
		updates["desc"] = req.Desc
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Order > 0 || req.Order < 0 {
		updates["order"] = req.Order
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.ContactQQ != "" {
		updates["contact_qq"] = req.ContactQQ
	}

	if err := config.DB.Model(&link).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update link"})
		return
	}

	config.DB.First(&link, linkID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Link updated successfully",
		"link":    link,
	})
}

// DeleteLink deletes a link (admin)
func DeleteLink(c *gin.Context) {
	linkID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid link ID"})
		return
	}

	var link models.Link
	if err := config.DB.First(&link, linkID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	if err := config.DB.Delete(&link).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Link deleted successfully"})
}

// ApplyLink applies for a new link (public)
func ApplyLink(c *gin.Context) {
	var req models.LinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link := models.Link{
		Name:      req.Name,
		URL:       req.URL,
		Logo:      req.Logo,
		Desc:      req.Desc,
		Email:     req.Email,
		ContactQQ: req.ContactQQ,
		Status:    "pending", // Requires approval
	}

	if err := config.DB.Create(&link).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit link application"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Link application submitted successfully, pending approval",
		"link":    link,
	})
}
