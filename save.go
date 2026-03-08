package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SaveData ゲーム進捗保存データ
type SaveData struct {
	Round              int        `json:"round"`
	MaxRounds          int        `json:"max_rounds"`
	PlayerScore        int        `json:"player_score"`
	CPUScore           int        `json:"cpu_score"`
	DeckIDs            []int      `json:"deck_ids"`
	FieldIDs           []int      `json:"field_ids"`
	PlayerHandIDs      []int      `json:"player_hand_ids"`
	CPUHandIDs         []int      `json:"cpu_hand_ids"`
	PlayerCapIDs       []int      `json:"player_captured_ids"`
	CPUCapIDs          []int      `json:"cpu_captured_ids"`
	IsPlayerTurn       bool       `json:"is_player_turn"`
	PlayerKoiKoi       bool       `json:"player_koikoi"`
	CPUKoiKoi          bool       `json:"cpu_koikoi"`
	NextParentIsPlayer bool       `json:"next_parent_is_player"`
	Difficulty         Difficulty `json:"difficulty"`
	LogLines           []string   `json:"log_lines"`
}

func cardIDsFromSlice(cards []Card) []int {
	ids := make([]int, len(cards))
	for i, c := range cards {
		ids[i] = c.ID
	}
	return ids
}

func cardsFromIDs(ids []int) []Card {
	cards := make([]Card, 0, len(ids))
	for _, id := range ids {
		if id >= 0 && id < len(AllCards) {
			cards = append(cards, AllCards[id])
		}
	}
	return cards
}

func GameToSaveData(g *Game, difficulty Difficulty, logLines []string) SaveData {
	return SaveData{
		Round:              g.Round,
		MaxRounds:          g.MaxRounds,
		PlayerScore:        g.PlayerScore,
		CPUScore:           g.CPUScore,
		DeckIDs:            cardIDsFromSlice(g.Deck),
		FieldIDs:           cardIDsFromSlice(g.Field),
		PlayerHandIDs:      cardIDsFromSlice(g.PlayerHand),
		CPUHandIDs:         cardIDsFromSlice(g.CPUHand),
		PlayerCapIDs:       cardIDsFromSlice(g.PlayerCaptured),
		CPUCapIDs:          cardIDsFromSlice(g.CPUCaptured),
		IsPlayerTurn:       g.IsPlayerTurn,
		PlayerKoiKoi:       g.PlayerKoiKoi,
		CPUKoiKoi:          g.CPUKoiKoi,
		NextParentIsPlayer: g.NextParentIsPlayer,
		Difficulty:         difficulty,
		LogLines:           logLines,
	}
}

func SaveDataToGame(sd *SaveData) *Game {
	g := &Game{
		Round:              sd.Round,
		MaxRounds:          sd.MaxRounds,
		PlayerScore:        sd.PlayerScore,
		CPUScore:           sd.CPUScore,
		Deck:               cardsFromIDs(sd.DeckIDs),
		Field:              cardsFromIDs(sd.FieldIDs),
		PlayerHand:         cardsFromIDs(sd.PlayerHandIDs),
		CPUHand:            cardsFromIDs(sd.CPUHandIDs),
		PlayerCaptured:     cardsFromIDs(sd.PlayerCapIDs),
		CPUCaptured:        cardsFromIDs(sd.CPUCapIDs),
		IsPlayerTurn:       sd.IsPlayerTurn,
		PlayerKoiKoi:       sd.PlayerKoiKoi,
		CPUKoiKoi:          sd.CPUKoiKoi,
		NextParentIsPlayer: sd.NextParentIsPlayer,
	}
	return g
}

func saveJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func SaveGame(path string, sd *SaveData) error {
	return saveJSON(path, sd)
}

func LoadGame(path string) (SaveData, error) {
	path = filepath.Clean(path)
	var sd SaveData
	data, err := os.ReadFile(path)
	if err != nil {
		return sd, err
	}
	if err := json.Unmarshal(data, &sd); err != nil {
		return sd, err
	}
	return sd, nil
}

func DeleteSave(path string) {
	_ = os.Remove(path)
}
