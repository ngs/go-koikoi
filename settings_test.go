package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDifficultyLabel(t *testing.T) {
	tests := []struct {
		diff Difficulty
		want string
	}{
		{DifficultyEasy, "かんたん"},
		{DifficultyNormal, "ふつう"},
		{DifficultyHard, "つよい"},
		{Difficulty("unknown"), "ふつう"},
	}
	for _, tt := range tests {
		if got := tt.diff.Label(); got != tt.want {
			t.Errorf("Difficulty(%q).Label() = %q, want %q", tt.diff, got, tt.want)
		}
	}
}

func TestDefaultSettings(t *testing.T) {
	s := DefaultSettings()
	if s.Rounds != 12 {
		t.Errorf("DefaultSettings().Rounds = %d, want 12", s.Rounds)
	}
	if s.Difficulty != DifficultyNormal {
		t.Errorf("DefaultSettings().Difficulty = %q, want %q", s.Difficulty, DifficultyNormal)
	}
}

func TestDefaultBaseDir(t *testing.T) {
	dir := DefaultBaseDir()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".koikoi")
	if dir != expected {
		t.Errorf("DefaultBaseDir() = %q, want %q", dir, expected)
	}
}

func TestSaveAndLoadSettings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	s := Settings{Rounds: 6, Difficulty: DifficultyHard}
	if err := SaveSettings(path, s); err != nil {
		t.Fatalf("SaveSettings error: %v", err)
	}

	loaded, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("LoadSettings error: %v", err)
	}
	if loaded.Rounds != 6 {
		t.Errorf("loaded.Rounds = %d, want 6", loaded.Rounds)
	}
	if loaded.Difficulty != DifficultyHard {
		t.Errorf("loaded.Difficulty = %q, want %q", loaded.Difficulty, DifficultyHard)
	}
}

func TestLoadSettingsFileNotFound(t *testing.T) {
	s, err := LoadSettings("/nonexistent/settings.json")
	if err == nil {
		t.Error("LoadSettings should return error for missing file")
	}
	// デフォルト設定が返される
	if s.Rounds != 12 {
		t.Errorf("Default Rounds = %d, want 12", s.Rounds)
	}
}

func TestLoadSettingsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte("invalid json"), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := LoadSettings(path)
	if err == nil {
		t.Error("LoadSettings should return error for invalid JSON")
	}
	// デフォルト設定が返される
	if s.Rounds != 12 {
		t.Errorf("Default Rounds = %d, want 12", s.Rounds)
	}
}

func TestLoadSettingsInvalidRounds(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(`{"rounds":0,"difficulty":"easy"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("LoadSettings error: %v", err)
	}
	if s.Rounds != 12 {
		t.Errorf("Rounds should default to 12 when <1, got %d", s.Rounds)
	}
}

func TestLoadSettingsEmptyDifficulty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(`{"rounds":6,"difficulty":""}`), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("LoadSettings error: %v", err)
	}
	if s.Difficulty != DifficultyNormal {
		t.Errorf("Difficulty should default to normal when empty, got %q", s.Difficulty)
	}
}

func TestSaveSettingsCreatesDir(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub", "dir")
	path := filepath.Join(subdir, "settings.json")

	s := Settings{Rounds: 3, Difficulty: DifficultyEasy}
	if err := SaveSettings(path, s); err != nil {
		t.Fatalf("SaveSettings error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("settings file was not created")
	}
}

func TestDifficultyConstants(t *testing.T) {
	if DifficultyEasy != "easy" {
		t.Errorf("DifficultyEasy = %q", DifficultyEasy)
	}
	if DifficultyNormal != "normal" {
		t.Errorf("DifficultyNormal = %q", DifficultyNormal)
	}
	if DifficultyHard != "hard" {
		t.Errorf("DifficultyHard = %q", DifficultyHard)
	}
}
