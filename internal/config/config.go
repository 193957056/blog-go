package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Backup   BackupConfig   `yaml:"backup"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type DatabaseConfig struct {
	Type     string `yaml:"type"`
	Path     string `yaml:"path"`
	MaxConns int    `yaml:"maxConns"`
}

type BackupConfig struct {
	Directory       string `yaml:"directory"`
	MaxBackups      int    `yaml:"maxBackups"`
	RetainDays      int    `yaml:"retainDays"`
	EnableAutoBackup bool   `yaml:"enableAutoBackup"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Backup.Directory == "" {
		cfg.Backup.Directory = "./backups"
	}
	if cfg.Backup.MaxBackups == 0 {
		cfg.Backup.MaxBackups = 10
	}

	return &cfg, nil
}
