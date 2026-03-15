package model

import (
	"time"
)

// Article represents a blog article
type Article struct {
	ID        int64     `db:"id" json:"id"`
	Title     string    `db:"title" json:"title"`
	Content   string    `db:"content" json:"content"`
	Slug      string    `db:"slug" json:"slug"`
	Status    string    `db:"status" json:"status"` // draft, published
	Author    string    `db:"author" json:"author"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Backup represents an article backup
type Backup struct {
	ID         int64     `db:"id" json:"id"`
	ArticleID  int64     `db:"article_id" json:"article_id"`
	Title      string    `db:"title" json:"title"`
	Content    string    `db:"content" json:"content"`
	Slug       string    `db:"slug" json:"slug"`
	Version    int       `db:"version" json:"version"`
	FilePath   string    `db:"file_path" json:"file_path"`
	Reason     string    `db:"reason" json:"reason"` // update, restore, manual
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// BackupLog represents a rollback operation log
type BackupLog struct {
	ID            int64     `db:"id" json:"id"`
	ArticleID     int64     `db:"article_id" json:"article_id"`
	FromVersion   int       `db:"from_version" json:"from_version"`
	ToVersion     int       `db:"to_version" json:"to_version"`
	RestoreBackupID int64   `db:"restore_backup_id" json:"restore_backup_id"`
	Operator      string    `db:"operator" json:"operator"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

// API response types
type BackupListResponse struct {
	Backups   []BackupSummary `json:"backups"`
	Total     int             `json:"total"`
	Page      int             `json:"page"`
	PageSize  int             `json:"page_size"`
}

type BackupSummary struct {
	ID        int64     `json:"id"`
	ArticleID int64     `json:"article_id"`
	Title     string    `json:"title"`
	Version   int       `json:"version"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

type BackupDetailResponse struct {
	Backup     Backup  `json:"backup"`
	Article    Article `json:"article,omitempty"`
}

type RestoreRequest struct {
	Operator string `json:"operator" binding:"required"`
}
