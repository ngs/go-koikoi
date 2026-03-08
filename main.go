package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/awesome-gocui/gocui"
)

func main() {
	var configDir string
	flag.StringVar(&configDir, "config", "", "設定ディレクトリのパス (デフォルト: ~/.koikoi)")
	flag.Parse()

	if configDir == "" {
		configDir = DefaultBaseDir()
	}

	settingsPath := filepath.Join(configDir, "settings.json")
	savePath := filepath.Join(configDir, "game.json")

	settings, _ := LoadSettings(settingsPath)

	ui := NewUI()
	ui.configDir = configDir
	ui.settingsPath = settingsPath
	ui.savePath = savePath
	ui.settings = settings
	ui.difficulty = settings.Difficulty

	// セーブデータがあれば復元
	if sd, err := LoadGame(savePath); err == nil {
		ui.game = SaveDataToGame(&sd)
		ui.logLines = sd.LogLines
		ui.difficulty = sd.Difficulty
		ui.phase = PhasePlayerSelectHand
		if !ui.game.IsPlayerTurn {
			ui.phase = PhaseCPUTurn
		}
		ui.addLog("--- セーブデータを復元しました ---")
	} else {
		ui.game = NewGame(settings.Rounds)
		ui.game.StartRound()
		ui.addLog(fmt.Sprintf("--- %s 開始 ---", ui.roundName()))
		ui.phase = PhasePlayerSelectHand
		if !ui.game.IsPlayerTurn {
			ui.phase = PhaseCPUTurn
		}
	}

	if err := ui.Run(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
}
