package models

import (
	"time"

	"lumina-blog/config"
	"gorm.io/gorm"
)

type Post struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Title       string         `gorm:"size:255;not null" json:"title"`
	Slug        string         `gorm:"size:255;uniqueIndex;not null" json:"slug"`
	Content     string         `gorm:"type:text" json:"content"`
	Excerpt     string         `gorm:"size:500" json:"excerpt"`
	Cover       string         `gorm:"size:500" json:"cover"`
	Status      string         `gorm:"size:20;default:'draft'" json:"status"` // draft, published
	ViewCount   int            `gorm:"default:0" json:"view_count"`
	ReadTime    int            `json:"read_time"` // in minutes
	CategoryID  uint           `json:"category_id"`
	Category    Category       `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Tags        []Tag          `gorm:"many2many:post_tags;" json:"tags,omitempty"`
}

type Category struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Slug      string         `gorm:"size:50;uniqueIndex;not null" json:"slug"`
	Color     string         `gorm:"size:20" json:"color"`
	Posts     []Post         `json:"posts,omitempty"`
}

type Tag struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Slug      string         `gorm:"size:50;uniqueIndex;not null" json:"slug"`
	Color     string         `gorm:"size:20" json:"color"`
	Posts     []Post         `gorm:"many2many:post_tags;" json:"posts,omitempty"`
}

// SeedData creates default categories and tags
func SeedData() {
	// Create categories
	categories := []Category{
		{Name: "技术", Slug: "tech", Color: "#7C3AED"},
		{Name: "设计", Slug: "design", Color: "#06B6D4"},
		{Name: "生活", Slug: "life", Color: "#10B981"},
		{Name: "随笔", Slug: "essay", Color: "#F59E0B"},
	}

	for _, cat := range categories {
		var exists Category
		if err := config.DB.Where("slug = ?", cat.Slug).First(&exists).Error; err != nil {
			config.DB.Create(&cat)
		}
	}

	// Create tags
	tags := []Tag{
		{Name: "Vue", Slug: "vue", Color: "#42B883"},
		{Name: "Go", Slug: "go", Color: "#00ADD8"},
		{Name: "前端", Slug: "frontend", Color: "#E34F26"},
		{Name: "后端", Slug: "backend", Color: "#1572B6"},
		{Name: "UI/UX", Slug: "ui-ux", Color: "#FF61F6"},
	}

	for _, tag := range tags {
		var exists Tag
		if err := config.DB.Where("slug = ?", tag.Slug).First(&exists).Error; err != nil {
			config.DB.Create(&tag)
		}
	}
}
