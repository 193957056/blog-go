package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lumina/blog/internal/config"
	"github.com/lumina/blog/internal/model"
	"github.com/lumina/blog/internal/repository"
)

type BackupService struct {
	repo   *repository.Repository
	config *config.BackupConfig
}

func NewBackupService(repo *repository.Repository, cfg *config.BackupConfig) *BackupService {
	return &BackupService{
		repo:   repo,
		config: cfg,
	}
}

// CreateBackup creates a backup for an article
func (s *BackupService) CreateBackup(article *model.Article, reason string) (*model.Backup, error) {
	// Get next version number
	version, err := s.repo.GetNextVersion(article.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next version: %w", err)
	}

	// Ensure backup directory exists
	if err := os.MkdirAll(s.config.Directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create backup file
	filename := fmt.Sprintf("article_%d_v%d_%s.json", article.ID, version, time.Now().Format("20060102_150405"))
	filepath := filepath.Join(s.config.Directory, filename)

	backupData := map[string]interface{}{
		"article_id": article.ID,
		"title":      article.Title,
		"content":    article.Content,
		"slug":       article.Slug,
		"status":     article.Status,
		"author":     article.Author,
		"created_at": article.CreatedAt,
		"updated_at": article.UpdatedAt,
		"version":    version,
	}

	data, err := json.MarshalIndent(backupData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal backup data: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write backup file: %w", err)
	}

	// Create backup record in database
	backup := &model.Backup{
		ArticleID: article.ID,
		Title:     article.Title,
		Content:   article.Content,
		Slug:      article.Slug,
		Version:   version,
		FilePath:  filepath,
		Reason:    reason,
		CreatedAt: time.Now(),
	}

	id, err := s.repo.CreateBackup(backup)
	if err != nil {
		// Try to delete the backup file if DB insert fails
		os.Remove(filepath)
		return nil, fmt.Errorf("failed to create backup record: %w", err)
	}
	backup.ID = id

	// Cleanup old backups
	if err := s.cleanupOldBackups(article.ID); err != nil {
		// Log error but don't fail the backup
		fmt.Printf("Warning: failed to cleanup old backups: %v\n", err)
	}

	return backup, nil
}

// RestoreFromBackup restores an article from a backup
func (s *BackupService) RestoreFromBackup(backupID int64, operator string) (*model.Article, *model.Backup, error) {
	// Get the backup to restore from
	backup, err := s.repo.GetBackupByID(backupID)
	if err != nil {
		return nil, nil, fmt.Errorf("backup not found: %w", err)
	}

	// Get current article state
	article, err := s.repo.GetArticleByID(backup.ArticleID)
	if err != nil {
		return nil, nil, fmt.Errorf("article not found: %w", err)
	}

	// Create a backup of current state before restore (preserve history)
	currentBackup, err := s.CreateBackup(article, "restore")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create backup before restore: %w", err)
	}

	// Log the restore operation
	log := &model.BackupLog{
		ArticleID:        article.ID,
		FromVersion:      int(article.UpdatedAt.Unix()),
		ToVersion:        backup.Version,
		RestoreBackupID:  backupID,
		Operator:         operator,
		CreatedAt:        time.Now(),
	}
	if _, err := s.repo.CreateBackupLog(log); err != nil {
		fmt.Printf("Warning: failed to create backup log: %v\n", err)
	}

	// Update article with backup data
	article.Title = backup.Title
	article.Content = backup.Content
	article.Slug = backup.Slug
	article.UpdatedAt = time.Now()

	if err := s.repo.UpdateArticle(article); err != nil {
		return nil, nil, fmt.Errorf("failed to restore article: %w", err)
	}

	return article, currentBackup, nil
}

// GetBackupList returns list of all backups with pagination
func (s *BackupService) GetBackupList(page, pageSize int) (*model.BackupListResponse, error) {
	backups, err := s.repo.GetAllBackups(page, pageSize)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.CountBackups()
	if err != nil {
		return nil, err
	}

	// Convert to summaries
	summaries := make([]model.BackupSummary, len(backups))
	for i, b := range backups {
		summaries[i] = model.BackupSummary{
			ID:        b.ID,
			ArticleID: b.ArticleID,
			Title:     b.Title,
			Version:   b.Version,
			Reason:    b.Reason,
			CreatedAt: b.CreatedAt,
		}
	}

	return &model.BackupListResponse{
		Backups:   summaries,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// GetBackupDetail returns detailed backup info
func (s *BackupService) GetBackupDetail(backupID int64) (*model.BackupDetailResponse, error) {
	backup, err := s.repo.GetBackupByID(backupID)
	if err != nil {
		return nil, err
	}

	// Try to read the backup file for full content
	if backup.Content == "" && backup.FilePath != "" {
		if data, err := os.ReadFile(backup.FilePath); err == nil {
			var backupData map[string]interface{}
			if json.Unmarshal(data, &backupData) == nil {
				if content, ok := backupData["content"].(string); ok {
					backup.Content = content
				}
			}
		}
	}

	return &model.BackupDetailResponse{
		Backup: *backup,
	}, nil
}

// DeleteBackup deletes a backup
func (s *BackupService) DeleteBackup(backupID int64) error {
	backup, err := s.repo.GetBackupByID(backupID)
	if err != nil {
		return err
	}

	// Delete the backup file
	if backup.FilePath != "" {
		os.Remove(backup.FilePath)
	}

	// Delete the database record
	return s.repo.DeleteBackup(backupID)
}

// cleanupOldBackups removes old backups, keeping only the most recent N
func (s *BackupService) cleanupOldBackups(articleID int64) error {
	count, err := s.repo.GetBackupsCountByArticle(articleID)
	if err != nil {
		return err
	}

	if count > s.config.MaxBackups {
		// Get backups to delete (old ones beyond maxBackups)
		backups, err := s.repo.GetBackupsByArticleID(articleID, count)
		if err != nil {
			return err
		}

		// Delete oldest backups beyond the limit
		for i := s.config.MaxBackups; i < len(backups); i++ {
			if err := s.DeleteBackup(backups[i].ID); err != nil {
				fmt.Printf("Warning: failed to delete old backup %d: %v\n", backups[i].ID, err)
			}
		}
	}

	return nil
}

// AutoBackupBeforeUpdate creates an automatic backup before article update
func (s *BackupService) AutoBackupBeforeUpdate(article *model.Article) (*model.Backup, error) {
	if !s.config.EnableAutoBackup {
		return nil, nil
	}
	return s.CreateBackup(article, "update")
}
