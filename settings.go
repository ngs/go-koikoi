package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Difficulty CPU難易度
type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyNormal Difficulty = "normal"
	DifficultyHard   Difficulty = "hard"
)

var difficultyLabels = map[Difficulty]string{
	DifficultyEasy:   "かんたん",
	DifficultyNormal: "ふつう",
	DifficultyHard:   "つよい",
}

func (d Difficulty) Label() string {
	if l, ok := difficultyLabels[d]; ok {
		return l
	}
	return "ふつう"
}

// Settings ゲーム設定
type Settings struct {
	Rounds     int        `json:"rounds"`
	Difficulty Difficulty `json:"difficulty"`
}

func DefaultSettings() Settings {
	return Settings{
		Rounds:     12,
		Difficulty: DifficultyNormal,
	}
}

func DefaultBaseDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".koikoi")
}

func LoadSettings(path string) (Settings, error) {
	path = filepath.Clean(path)
	s := DefaultSettings()
	data, err := os.ReadFile(path)
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return DefaultSettings(), err
	}
	if s.Rounds < 1 {
		s.Rounds = 12
	}
	if s.Difficulty == "" {
		s.Difficulty = DifficultyNormal
	}
	return s, nil
}

func SaveSettings(path string, s Settings) error {
	return saveJSON(path, s)
}
