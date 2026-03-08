package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCardIDsFromSlice(t *testing.T) {
	cards := []Card{AllCards[0], AllCards[8], AllCards[28]}
	ids := cardIDsFromSlice(cards)
	expected := []int{0, 8, 28}
	if len(ids) != len(expected) {
		t.Fatalf("cardIDsFromSlice got %d IDs, want %d", len(ids), len(expected))
	}
	for i, id := range ids {
		if id != expected[i] {
			t.Errorf("ids[%d] = %d, want %d", i, id, expected[i])
		}
	}
}

func TestCardIDsFromSliceEmpty(t *testing.T) {
	ids := cardIDsFromSlice(nil)
	if len(ids) != 0 {
		t.Errorf("cardIDsFromSlice(nil) got %d IDs, want 0", len(ids))
	}
}

func TestCardsFromIDs(t *testing.T) {
	ids := []int{0, 8, 28}
	cards := cardsFromIDs(ids)
	if len(cards) != 3 {
		t.Fatalf("cardsFromIDs got %d cards, want 3", len(cards))
	}
	for i, c := range cards {
		if c.ID != ids[i] {
			t.Errorf("cards[%d].ID = %d, want %d", i, c.ID, ids[i])
		}
	}
}

func TestCardsFromIDsInvalid(t *testing.T) {
	ids := []int{0, -1, 48, 99}
	cards := cardsFromIDs(ids)
	if len(cards) != 1 { // only ID 0 is valid
		t.Errorf("cardsFromIDs with invalid IDs got %d cards, want 1", len(cards))
	}
}

func TestCardsFromIDsEmpty(t *testing.T) {
	cards := cardsFromIDs(nil)
	if len(cards) != 0 {
		t.Errorf("cardsFromIDs(nil) got %d cards, want 0", len(cards))
	}
}

func TestGameToSaveDataAndBack(t *testing.T) {
	g := NewGame(6)
	g.StartRound()
	g.PlayerScore = 10
	g.CPUScore = 5
	g.PlayerKoiKoi = true
	g.CPUKoiKoi = false
	g.NextParentIsPlayer = false

	logLines := []string{"line1", "line2"}
	sd := GameToSaveData(g, DifficultyHard, logLines)

	if sd.Round != g.Round {
		t.Errorf("sd.Round = %d, want %d", sd.Round, g.Round)
	}
	if sd.MaxRounds != g.MaxRounds {
		t.Errorf("sd.MaxRounds = %d, want %d", sd.MaxRounds, g.MaxRounds)
	}
	if sd.PlayerScore != 10 {
		t.Errorf("sd.PlayerScore = %d, want 10", sd.PlayerScore)
	}
	if sd.CPUScore != 5 {
		t.Errorf("sd.CPUScore = %d, want 5", sd.CPUScore)
	}
	if !sd.PlayerKoiKoi {
		t.Error("sd.PlayerKoiKoi should be true")
	}
	if sd.CPUKoiKoi {
		t.Error("sd.CPUKoiKoi should be false")
	}
	if sd.NextParentIsPlayer {
		t.Error("sd.NextParentIsPlayer should be false")
	}
	if sd.Difficulty != DifficultyHard {
		t.Errorf("sd.Difficulty = %q, want hard", sd.Difficulty)
	}
	if len(sd.LogLines) != 2 {
		t.Errorf("sd.LogLines has %d entries, want 2", len(sd.LogLines))
	}
	if len(sd.DeckIDs) != len(g.Deck) {
		t.Errorf("sd.DeckIDs has %d, want %d", len(sd.DeckIDs), len(g.Deck))
	}
	if len(sd.FieldIDs) != len(g.Field) {
		t.Errorf("sd.FieldIDs has %d, want %d", len(sd.FieldIDs), len(g.Field))
	}
	if len(sd.PlayerHandIDs) != len(g.PlayerHand) {
		t.Errorf("sd.PlayerHandIDs has %d, want %d", len(sd.PlayerHandIDs), len(g.PlayerHand))
	}
	if len(sd.CPUHandIDs) != len(g.CPUHand) {
		t.Errorf("sd.CPUHandIDs has %d, want %d", len(sd.CPUHandIDs), len(g.CPUHand))
	}

	// Restore
	g2 := SaveDataToGame(&sd)
	if g2.Round != g.Round {
		t.Errorf("restored Round = %d, want %d", g2.Round, g.Round)
	}
	if g2.MaxRounds != g.MaxRounds {
		t.Errorf("restored MaxRounds = %d, want %d", g2.MaxRounds, g.MaxRounds)
	}
	if g2.PlayerScore != 10 {
		t.Errorf("restored PlayerScore = %d, want 10", g2.PlayerScore)
	}
	if g2.CPUScore != 5 {
		t.Errorf("restored CPUScore = %d, want 5", g2.CPUScore)
	}
	if len(g2.Deck) != len(g.Deck) {
		t.Errorf("restored Deck has %d, want %d", len(g2.Deck), len(g.Deck))
	}
	if len(g2.PlayerHand) != len(g.PlayerHand) {
		t.Errorf("restored PlayerHand has %d, want %d", len(g2.PlayerHand), len(g.PlayerHand))
	}
	if !g2.PlayerKoiKoi {
		t.Error("restored PlayerKoiKoi should be true")
	}
	if g2.NextParentIsPlayer {
		t.Error("restored NextParentIsPlayer should be false")
	}
}

func TestSaveAndLoadGame(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "game.json")

	sd := SaveData{
		Round:              3,
		MaxRounds:          12,
		PlayerScore:        15,
		CPUScore:           8,
		DeckIDs:            []int{0, 1, 2},
		FieldIDs:           []int{3, 4},
		PlayerHandIDs:      []int{5, 6, 7},
		CPUHandIDs:         []int{8, 9, 10},
		PlayerCapIDs:       []int{11, 12},
		CPUCapIDs:          []int{13, 14},
		IsPlayerTurn:       true,
		PlayerKoiKoi:       true,
		CPUKoiKoi:          false,
		NextParentIsPlayer: true,
		Difficulty:         DifficultyEasy,
		LogLines:           []string{"log1", "log2"},
	}

	if err := SaveGame(path, &sd); err != nil {
		t.Fatalf("SaveGame error: %v", err)
	}

	loaded, err := LoadGame(path)
	if err != nil {
		t.Fatalf("LoadGame error: %v", err)
	}

	if loaded.Round != 3 {
		t.Errorf("loaded.Round = %d, want 3", loaded.Round)
	}
	if loaded.PlayerScore != 15 {
		t.Errorf("loaded.PlayerScore = %d, want 15", loaded.PlayerScore)
	}
	if !loaded.IsPlayerTurn {
		t.Error("loaded.IsPlayerTurn should be true")
	}
	if loaded.Difficulty != DifficultyEasy {
		t.Errorf("loaded.Difficulty = %q, want easy", loaded.Difficulty)
	}
	if len(loaded.LogLines) != 2 {
		t.Errorf("loaded.LogLines has %d, want 2", len(loaded.LogLines))
	}
	if len(loaded.DeckIDs) != 3 {
		t.Errorf("loaded.DeckIDs has %d, want 3", len(loaded.DeckIDs))
	}
}

func TestLoadGameFileNotFound(t *testing.T) {
	_, err := LoadGame("/nonexistent/game.json")
	if err == nil {
		t.Error("LoadGame should return error for missing file")
	}
}

func TestLoadGameInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "game.json")
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := LoadGame(path)
	if err == nil {
		t.Error("LoadGame should return error for invalid JSON")
	}
}

func TestDeleteSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "game.json")
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	DeleteSave(path)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("DeleteSave should remove the file")
	}
}

func TestDeleteSaveNonExistent(t *testing.T) {
	// Should not panic when deleting a non-existent file
	DeleteSave("/nonexistent/game.json")
	t.Log("DeleteSave on non-existent path did not panic")
}

func TestSaveGameCreatesDir(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub", "dir")
	path := filepath.Join(subdir, "game.json")

	sd := SaveData{Round: 1, MaxRounds: 12}
	if err := SaveGame(path, &sd); err != nil {
		t.Fatalf("SaveGame error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("game file was not created")
	}
}
