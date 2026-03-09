package main

import (
	"path/filepath"
	"testing"

	"github.com/awesome-gocui/gocui"
)

func setupCPUTurnUI(t *testing.T) (*UI, *gocui.Gui) {
	t.Helper()
	g := newTestGUI(t)
	u := NewUI()
	dir := t.TempDir()
	u.configDir = dir
	u.settingsPath = filepath.Join(dir, "settings.json")
	u.savePath = filepath.Join(dir, "game.json")
	u.settings = DefaultSettings()
	u.gui = g
	return u, g
}

func TestDoCPUTurnNormalPath(t *testing.T) {
	u, g := setupCPUTurnUI(t)
	u.game = NewGame(12)
	u.game.CPUHand = []Card{AllCards[0]}
	u.game.PlayerHand = []Card{AllCards[4]}
	u.game.Field = []Card{AllCards[2]}
	u.game.Deck = []Card{AllCards[12]}
	u.game.IsPlayerTurn = false
	u.difficulty = DifficultyNormal
	cpuTurnRunning.Store(false)

	u.doCPUTurn(g)

	if cpuTurnRunning.Load() {
		t.Error("cpuTurnRunning should be false after doCPUTurn")
	}
}

func TestExecuteCPUTurnNormal(t *testing.T) {
	u, g := setupCPUTurnUI(t)
	u.game = NewGame(12)
	u.game.CPUHand = []Card{AllCards[0]}
	u.game.PlayerHand = []Card{AllCards[4]}
	u.game.Field = []Card{AllCards[2]}
	u.game.Deck = []Card{AllCards[12], AllCards[20], AllCards[24], AllCards[30]}
	u.game.IsPlayerTurn = false
	u.difficulty = DifficultyNormal

	if err := u.executeCPUTurn(g); err != nil {
		t.Fatalf("executeCPUTurn error: %v", err)
	}
	if !u.game.IsPlayerTurn {
		t.Error("IsPlayerTurn should be true after CPU turn")
	}
	if u.phase != PhasePlayerSelectHand {
		t.Errorf("phase = %d, want PhasePlayerSelectHand", u.phase)
	}
}

func TestExecuteCPUTurnYaku(t *testing.T) {
	u, g := setupCPUTurnUI(t)
	u.game = NewGame(12)
	u.game.CPUCaptured = cardsFromIDList(0, 8)
	u.game.CPUHand = []Card{AllCards[30]}
	u.game.PlayerHand = []Card{AllCards[4]}
	u.game.Field = []Card{AllCards[28]}
	u.game.Deck = []Card{AllCards[12], AllCards[20]}
	u.game.CPUPrevYaku = nil
	u.game.IsPlayerTurn = false
	u.difficulty = DifficultyEasy

	if err := u.executeCPUTurn(g); err != nil {
		t.Fatalf("executeCPUTurn error: %v", err)
	}
	if u.game.CPUScore == 0 {
		t.Error("CPUScore should be > 0 after yaku win")
	}
}

func TestExecuteCPUTurnKoiKoi(t *testing.T) {
	u, g := setupCPUTurnUI(t)
	u.game = NewGame(12)
	u.game.CPUCaptured = cardsFromIDList(0, 8)
	u.game.CPUHand = []Card{AllCards[30], AllCards[14], AllCards[18]}
	u.game.PlayerHand = []Card{AllCards[4], AllCards[16], AllCards[22]}
	u.game.Field = []Card{AllCards[28]}
	u.game.Deck = []Card{AllCards[12], AllCards[20]}
	u.game.CPUPrevYaku = nil
	u.game.IsPlayerTurn = false
	u.difficulty = DifficultyHard

	if err := u.executeCPUTurn(g); err != nil {
		t.Fatalf("executeCPUTurn error: %v", err)
	}
	if !u.game.CPUKoiKoi {
		t.Error("CPUKoiKoi should be true")
	}
	if u.phase != PhaseCPUKoiKoi {
		t.Errorf("phase = %d, want PhaseCPUKoiKoi(%d)", u.phase, PhaseCPUKoiKoi)
	}
	if len(u.cpuKoiKoiYaku) == 0 {
		t.Error("cpuKoiKoiYaku should not be empty")
	}

	// OKボタン押下でプレイヤーターンへ
	if err := u.onCPUKoiKoiOK(g); err != nil {
		t.Fatalf("onCPUKoiKoiOK error: %v", err)
	}
	if u.phase != PhasePlayerSelectHand {
		t.Errorf("phase after OK = %d, want PhasePlayerSelectHand(%d)", u.phase, PhasePlayerSelectHand)
	}
	if u.cpuKoiKoiYaku != nil {
		t.Error("cpuKoiKoiYaku should be nil after OK")
	}
}

func TestExecuteCPUTurnRoundOver(t *testing.T) {
	u, g := setupCPUTurnUI(t)
	u.game = NewGame(12)
	u.game.CPUHand = []Card{AllCards[12]}
	u.game.PlayerHand = nil
	u.game.Field = []Card{AllCards[0]}
	u.game.Deck = []Card{AllCards[20]}
	u.game.IsPlayerTurn = false
	u.difficulty = DifficultyNormal

	if err := u.executeCPUTurn(g); err != nil {
		t.Fatalf("executeCPUTurn error: %v", err)
	}
}

func TestExecuteCPUTurnNoCapture(t *testing.T) {
	u, g := setupCPUTurnUI(t)
	u.game = NewGame(12)
	u.game.CPUHand = []Card{AllCards[0]}
	u.game.PlayerHand = []Card{AllCards[4]}
	u.game.Field = []Card{AllCards[12]}
	u.game.Deck = []Card{AllCards[16]}
	u.game.IsPlayerTurn = false
	u.difficulty = DifficultyNormal

	if err := u.executeCPUTurn(g); err != nil {
		t.Fatalf("executeCPUTurn error: %v", err)
	}
}

func TestExecuteCPUTurnEmptyHand(t *testing.T) {
	u, g := setupCPUTurnUI(t)
	u.game = NewGame(12)
	u.game.CPUHand = nil
	u.game.PlayerHand = []Card{AllCards[4]}
	u.game.Field = []Card{AllCards[0]}
	u.game.Deck = []Card{AllCards[12]}
	u.game.IsPlayerTurn = false

	if err := u.executeCPUTurn(g); err != nil {
		t.Fatalf("executeCPUTurn error: %v", err)
	}
	if !u.game.IsPlayerTurn {
		t.Error("IsPlayerTurn should be true after empty hand guard")
	}
	if u.phase != PhasePlayerSelectHand {
		t.Errorf("phase = %d, want PhasePlayerSelectHand", u.phase)
	}
}
