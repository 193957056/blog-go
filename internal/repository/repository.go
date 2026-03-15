package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/lumina/blog/internal/model"
)

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Article operations
func (r *Repository) GetArticleByID(id int64) (*model.Article, error) {
	var article model.Article
	err := r.db.Get(&article, "SELECT * FROM articles WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *Repository) GetArticleBySlug(slug string) (*model.Article, error) {
	var article model.Article
	err := r.db.Get(&article, "SELECT * FROM articles WHERE slug = ?", slug)
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *Repository) CreateArticle(article *model.Article) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO articles (title, content, slug, status, author, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		article.Title, article.Content, article.Slug, article.Status, article.Author, article.CreatedAt, article.UpdatedAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) UpdateArticle(article *model.Article) error {
	_, err := r.db.Exec(`
		UPDATE articles 
		SET title = ?, content = ?, slug = ?, status = ?, author = ?, updated_at = ?
		WHERE id = ?`,
		article.Title, article.Content, article.Slug, article.Status, article.Author, article.UpdatedAt, article.ID)
	return err
}

func (r *Repository) DeleteArticle(id int64) error {
	_, err := r.db.Exec("DELETE FROM articles WHERE id = ?", id)
	return err
}

// Backup operations
func (r *Repository) CreateBackup(backup *model.Backup) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO backups (article_id, title, content, slug, version, file_path, reason, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		backup.ArticleID, backup.Title, backup.Content, backup.Slug, backup.Version, backup.FilePath, backup.Reason, backup.CreatedAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) GetBackupByID(id int64) (*model.Backup, error) {
	var backup model.Backup
	err := r.db.Get(&backup, "SELECT * FROM backups WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &backup, nil
}

func (r *Repository) GetBackupsByArticleID(articleID int64, limit int) ([]model.Backup, error) {
	var backups []model.Backup
	err := r.db.Select(&backups, `
		SELECT * FROM backups 
		WHERE article_id = ? 
		ORDER BY version DESC 
		LIMIT ?`, articleID, limit)
	return backups, err
}

func (r *Repository) GetAllBackups(page, pageSize int) ([]model.Backup, error) {
	var backups []model.Backup
	offset := (page - 1) * pageSize
	err := r.db.Select(&backups, `
		SELECT * FROM backups 
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?`, pageSize, offset)
	return backups, err
}

func (r *Repository) CountBackups() (int, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM backups")
	return count, err
}

func (r *Repository) GetLatestBackup(articleID int64) (*model.Backup, error) {
	var backup model.Backup
	err := r.db.Get(&backup, `
		SELECT * FROM backups 
		WHERE article_id = ? 
		ORDER BY version DESC 
		LIMIT 1`, articleID)
	if err != nil {
		return nil, err
	}
	return &backup, nil
}

func (r *Repository) GetNextVersion(articleID int64) (int, error) {
	var version int
	err := r.db.Get(&version, `
		SELECT COALESCE(MAX(version), 0) + 1 
		FROM backups 
		WHERE article_id = ?`, articleID)
	return version, err
}

func (r *Repository) DeleteBackup(id int64) error {
	_, err := r.db.Exec("DELETE FROM backups WHERE id = ?", id)
	return err
}

func (r *Repository) GetBackupsCountByArticle(articleID int64) (int, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM backups WHERE article_id = ?", articleID)
	return count, err
}

func (r *Repository) DeleteOldBackups(articleID int64, keepCount int) error {
	_, err := r.db.Exec(`
		DELETE FROM backups 
		WHERE article_id = ? 
		AND id NOT IN (
			SELECT id FROM backups 
			WHERE article_id = ? 
			ORDER BY version DESC 
			LIMIT ?
		)`, articleID, articleID, keepCount)
	return err
}

// Backup log operations
func (r *Repository) CreateBackupLog(log *model.BackupLog) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO backup_logs (article_id, from_version, to_version, restore_backup_id, operator, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		log.ArticleID, log.FromVersion, log.ToVersion, log.RestoreBackupID, log.Operator, log.CreatedAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) GetBackupLogsByArticleID(articleID int64) ([]model.BackupLog, error) {
	var logs []model.BackupLog
	err := r.db.Select(&logs, `
		SELECT * FROM backup_logs 
		WHERE article_id = ? 
		ORDER BY created_at DESC`, articleID)
	return logs, err
}
