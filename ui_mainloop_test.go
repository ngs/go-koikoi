//go:build !race

package main

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/awesome-gocui/gocui"
)

// These tests use gocui's OutputSimulator + StartGui which spawns a pollEvent
// goroutine accessing a global tcell.SimulationScreen. The gocui library does
// not synchronize this goroutine with Close(), causing unavoidable data races
// between consecutive tests. They are excluded from -race builds.

func TestDoCPUTurnWithMainLoop(t *testing.T) {
	g, err := gocui.NewGui(gocui.OutputSimulator, true)
	if err != nil {
		t.Fatalf("NewGui error: %v", err)
	}
	defer g.Close()

	dir := t.TempDir()
	u := NewUI()
	u.game = NewGame(12)
	u.game.CPUHand = []Card{AllCards[0]}
	u.game.PlayerHand = []Card{AllCards[4]}
	u.game.Field = []Card{AllCards[2]}
	u.game.Deck = []Card{AllCards[12], AllCards[20], AllCards[24], AllCards[30]}
	u.game.IsPlayerTurn = false
	u.configDir = dir
	u.settingsPath = filepath.Join(dir, "settings.json")
	u.savePath = filepath.Join(dir, "game.json")
	u.settings = DefaultSettings()
	u.difficulty = DifficultyNormal
	u.phase = PhaseCPUTurn
	u.gui = g
	cpuTurnRunning.Store(false)

	g.SetManagerFunc(u.layout)
	if err := u.setKeybindings(g); err != nil {
		t.Fatalf("setKeybindings: %v", err)
	}

	testingScreen := g.GetTestingScreen()
	cleanup := testingScreen.StartGui()
	defer cleanup()

	time.Sleep(1500 * time.Millisecond)
}

func TestDoCPUTurnYakuWithMainLoop(t *testing.T) {
	g, err := gocui.NewGui(gocui.OutputSimulator, true)
	if err != nil {
		t.Fatalf("NewGui error: %v", err)
	}
	defer g.Close()

	dir := t.TempDir()
	u := NewUI()
	u.game = NewGame(12)
	u.game.CPUCaptured = cardsFromIDList(0, 8)
	u.game.CPUHand = []Card{AllCards[30]}
	u.game.PlayerHand = []Card{AllCards[4]}
	u.game.Field = []Card{AllCards[28]}
	u.game.Deck = []Card{AllCards[12], AllCards[20]}
	u.game.CPUPrevYaku = nil
	u.game.IsPlayerTurn = false
	u.configDir = dir
	u.settingsPath = filepath.Join(dir, "settings.json")
	u.savePath = filepath.Join(dir, "game.json")
	u.settings = DefaultSettings()
	u.difficulty = DifficultyEasy
	u.phase = PhaseCPUTurn
	u.gui = g
	cpuTurnRunning.Store(false)

	g.SetManagerFunc(u.layout)
	if err := u.setKeybindings(g); err != nil {
		t.Fatalf("setKeybindings: %v", err)
	}

	testingScreen := g.GetTestingScreen()
	cleanup := testingScreen.StartGui()
	defer cleanup()

	time.Sleep(1500 * time.Millisecond)
}

func TestDoCPUTurnKoiKoiWithMainLoop(t *testing.T) {
	g, err := gocui.NewGui(gocui.OutputSimulator, true)
	if err != nil {
		t.Fatalf("NewGui error: %v", err)
	}
	defer g.Close()

	dir := t.TempDir()
	u := NewUI()
	u.game = NewGame(12)
	u.game.CPUCaptured = cardsFromIDList(0, 8)
	u.game.CPUHand = []Card{AllCards[30], AllCards[14]}
	u.game.PlayerHand = []Card{AllCards[4], AllCards[16]}
	u.game.Field = []Card{AllCards[28]}
	u.game.Deck = []Card{AllCards[12], AllCards[20]}
	u.game.CPUPrevYaku = nil
	u.game.IsPlayerTurn = false
	u.configDir = dir
	u.settingsPath = filepath.Join(dir, "settings.json")
	u.savePath = filepath.Join(dir, "game.json")
	u.settings = DefaultSettings()
	u.difficulty = DifficultyHard
	u.phase = PhaseCPUTurn
	u.gui = g
	cpuTurnRunning.Store(false)

	g.SetManagerFunc(u.layout)
	if err := u.setKeybindings(g); err != nil {
		t.Fatalf("setKeybindings: %v", err)
	}

	testingScreen := g.GetTestingScreen()
	cleanup := testingScreen.StartGui()
	defer cleanup()

	time.Sleep(1500 * time.Millisecond)
}

func TestDoCPUTurnRoundOverWithMainLoop(t *testing.T) {
	g, err := gocui.NewGui(gocui.OutputSimulator, true)
	if err != nil {
		t.Fatalf("NewGui error: %v", err)
	}
	defer g.Close()

	dir := t.TempDir()
	u := NewUI()
	u.game = NewGame(12)
	u.game.CPUHand = []Card{AllCards[12]}
	u.game.PlayerHand = nil
	u.game.Field = []Card{AllCards[0]}
	u.game.Deck = []Card{AllCards[20]}
	u.game.IsPlayerTurn = false
	u.configDir = dir
	u.settingsPath = filepath.Join(dir, "settings.json")
	u.savePath = filepath.Join(dir, "game.json")
	u.settings = DefaultSettings()
	u.difficulty = DifficultyNormal
	u.phase = PhaseCPUTurn
	u.gui = g
	cpuTurnRunning.Store(false)

	g.SetManagerFunc(u.layout)
	if err := u.setKeybindings(g); err != nil {
		t.Fatalf("setKeybindings: %v", err)
	}

	testingScreen := g.GetTestingScreen()
	cleanup := testingScreen.StartGui()
	defer cleanup()

	time.Sleep(1500 * time.Millisecond)
}
