//go:build js

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall/js"
)

// js/wasm ではファイルシステムが使えないため、ブラウザの localStorage へ保存する。
// localStorage はストレージ無効化環境（Cookie ブロック時の Firefox、Safari の
// プライベートブラウジング等）でプロパティアクセスや setItem 自体が例外を投げる
// ことがある。syscall/js は JS 例外を panic に変換するため、すべて recover で
// error に変換し、保存失敗でゲームが落ちないようにする。

// storageKey はパス構成に依存しない固定キーへ正規化する
// (例: ".koikoi/settings.json" -> "koikoi:settings.json")。
func storageKey(path string) string {
	return "koikoi:" + filepath.Base(path)
}

func localStorageValue() (v js.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("localStorage: %v", r)
		}
	}()
	v = js.Global().Get("localStorage")
	if v.IsUndefined() || v.IsNull() {
		return js.Undefined(), os.ErrNotExist
	}
	return v, nil
}

func safeCall(v js.Value, method string, args ...any) (res js.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("localStorage.%s: %v", method, r)
		}
	}()
	return v.Call(method, args...), nil
}

// warnJS は console.warn への best-effort 出力（失敗しても無視）。
func warnJS(msg string) {
	defer func() { _ = recover() }()
	js.Global().Get("console").Call("warn", msg)
}

func readFileData(path string) ([]byte, error) {
	ls, err := localStorageValue()
	if err != nil {
		return nil, err
	}
	v, err := safeCall(ls, "getItem", storageKey(path))
	if err != nil {
		return nil, err
	}
	if v.IsNull() {
		return nil, os.ErrNotExist
	}
	return []byte(v.String()), nil
}

func writeFileData(path string, data []byte) error {
	ls, err := localStorageValue()
	if err != nil {
		warnJS("koikoi: 保存できません: " + err.Error())
		return err
	}
	if _, err := safeCall(ls, "setItem", storageKey(path), string(data)); err != nil {
		warnJS("koikoi: 保存に失敗しました: " + err.Error())
		return err
	}
	return nil
}

func removeFileData(path string) {
	ls, err := localStorageValue()
	if err != nil {
		return
	}
	_, _ = safeCall(ls, "removeItem", storageKey(path))
}
