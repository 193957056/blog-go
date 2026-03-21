package models

import (
	"time"

	"gorm.io/gorm"
)

// Comment represents a blog comment
type Comment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	PostID    uint           `gorm:"index;not null" json:"post_id"`
	ParentID  *uint          `gorm:"index" json:"parent_id"` // For nested replies
	UserID    *uint          `gorm:"index" json:"user_id"`   // NULL for anonymous
	Author    string         `gorm:"size:50" json:"author"`  // Name for anonymous users
	Email     string         `gorm:"size:100" json:"email"`   // Email for anonymous users
	Content   string         `gorm:"type:text;not null" json:"content"`
	Status    string         `gorm:"size:20;default:'pending'" json:"status"` // pending, approved, spam, deleted
	IP        string         `gorm:"size:50" json:"ip"`                       // Commenter IP
	UserAgent string         `gorm:"size:500" json:"user_agent"`              // Browser info

	// Relations
	Post      Post       `gorm:"foreignKey:PostID" json:"post,omitempty"`
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Parent    *Comment  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Replies   []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// Like represents a post like
type Like struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	PostID    uint      `gorm:"index;not null" json:"post_id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	IP        string    `gorm:"size:50" json:"ip"` // For anonymous likes

	// Relations
	Post Post `gorm:"foreignKey:PostID" json:"post,omitempty"`
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Favorite represents a post favorite
type Favorite struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	PostID    uint      `gorm:"index;not null" json:"post_id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`

	// Relations
	Post Post `gorm:"foreignKey:PostID" json:"post,omitempty"`
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Link represents a blog friend link
type Link struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	URL       string         `gorm:"size:500;not null" json:"url"`
	Logo      string         `gorm:"size:500" json:"logo"`
	Desc      string         `gorm:"size:500" json:"desc"`
	Email     string         `gorm:"size:100" json:"email"`
	Order     int            `gorm:"default:0" json:"order"`
	Status    string         `gorm:"size:20;default:'pending'" json:"status"` // pending, approved, rejected
	ContactQQ string         `gorm:"size:20" json:"contact_qq"`
}

// CommentRequest represents a comment creation request
type CommentRequest struct {
	PostID   uint   `json:"post_id" binding:"required"`
	ParentID *uint  `json:"parent_id"`
	Author   string `json:"author"` // For anonymous
	Email    string `json:"email"`   // For anonymous
	Content  string `json:"content" binding:"required,min=1,max=2000"`
}

// LinkRequest represents a link creation request
type LinkRequest struct {
	Name      string `json:"name" binding:"required,min=1,max=100"`
	URL       string `json:"url" binding:"required,url"`
	Logo      string `json:"logo"`
	Desc      string `json:"desc"`
	Email     string `json:"email"`
	ContactQQ string `json:"contact_qq"`
}

// StatsResponse represents blog statistics
type StatsResponse struct {
	TotalPosts     int64 `json:"total_posts"`
	TotalComments  int64 `json:"total_comments"`
	TotalViews     int64 `json:"total_views"`
	TotalLikes     int64 `json:"total_likes"`
	TotalFavorites int64 `json:"total_favorites"`
	PublishedPosts int64 `json:"published_posts"`
	DraftPosts     int64 `json:"draft_posts"`
	PendingComments int64 `json:"pending_comments"`
}
