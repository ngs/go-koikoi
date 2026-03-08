package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/awesome-gocui/gocui"
)

func main() {
	var configDir string
	flag.StringVar(&configDir, "config", "", "設定ディレクトリのパス (デフォルト: ~/.koikoi)")
	flag.Parse()

	if configDir == "" {
		configDir = DefaultBaseDir()
	}

	ui := NewUI()
	ui.Init(configDir)

	if err := ui.Run(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
}
