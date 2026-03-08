package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_MissingFileCreatesDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LINEAR_CONFIG_DIR", dir)
	t.Setenv("LINEAR_API_KEY", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.APIKey != "" {
		t.Errorf("expected empty APIKey, got %q", cfg.APIKey)
	}
	if cfg.DefaultTeam != "" {
		t.Errorf("expected empty DefaultTeam, got %q", cfg.DefaultTeam)
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LINEAR_CONFIG_DIR", dir)
	t.Setenv("LINEAR_API_KEY", "")

	content := "api_key: lin_api_test123\ndefault_team: ENG\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.APIKey != "lin_api_test123" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "lin_api_test123")
	}
	if cfg.DefaultTeam != "ENG" {
		t.Errorf("DefaultTeam = %q, want %q", cfg.DefaultTeam, "ENG")
	}
}

func TestLoad_EmptyConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LINEAR_CONFIG_DIR", dir)
	t.Setenv("LINEAR_API_KEY", "")

	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.APIKey != "" {
		t.Errorf("expected empty APIKey, got %q", cfg.APIKey)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LINEAR_CONFIG_DIR", dir)
	t.Setenv("LINEAR_API_KEY", "lin_api_from_env")

	content := "api_key: lin_api_from_file\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.APIKey != "lin_api_from_env" {
		t.Errorf("APIKey = %q, want env value %q", cfg.APIKey, "lin_api_from_env")
	}
}

func TestLoad_EnvOverride_NoFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LINEAR_CONFIG_DIR", dir)
	t.Setenv("LINEAR_API_KEY", "lin_api_env_only")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.APIKey != "lin_api_env_only" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "lin_api_env_only")
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LINEAR_CONFIG_DIR", dir)
	t.Setenv("LINEAR_API_KEY", "")

	cfg := &Config{
		APIKey:      "lin_api_saved",
		DefaultTeam: "TEAM",
	}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() after Save() error = %v", err)
	}
	if loaded.APIKey != cfg.APIKey {
		t.Errorf("APIKey = %q, want %q", loaded.APIKey, cfg.APIKey)
	}
	if loaded.DefaultTeam != cfg.DefaultTeam {
		t.Errorf("DefaultTeam = %q, want %q", loaded.DefaultTeam, cfg.DefaultTeam)
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "nested", "subdir")
	t.Setenv("LINEAR_CONFIG_DIR", nested)
	t.Setenv("LINEAR_API_KEY", "")

	cfg := &Config{APIKey: "lin_api_nested"}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(nested, "config.yaml")); err != nil {
		t.Errorf("config file not created: %v", err)
	}
}
