//go:build !js

package main

import (
	"os"
	"path/filepath"
)

// readFileData は path のファイル内容を読み込む。
func readFileData(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// writeFileData は path にデータを書き込む（親ディレクトリも作成）。
func writeFileData(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// removeFileData は path のファイルを削除する。
func removeFileData(path string) {
	_ = os.Remove(path)
}
