//go:build js

package main

import (
	"os"
	"syscall/js"
)

// js/wasm ではファイルシステムが使えないため、path をキーに
// ブラウザの localStorage へ保存する。

func localStorage() js.Value {
	return js.Global().Get("localStorage")
}

func storageKey(path string) string {
	return "koikoi:" + path
}

func readFileData(path string) ([]byte, error) {
	ls := localStorage()
	if ls.IsUndefined() || ls.IsNull() {
		return nil, os.ErrNotExist
	}
	v := ls.Call("getItem", storageKey(path))
	if v.IsNull() {
		return nil, os.ErrNotExist
	}
	return []byte(v.String()), nil
}

func writeFileData(path string, data []byte) error {
	ls := localStorage()
	if ls.IsUndefined() || ls.IsNull() {
		return os.ErrInvalid
	}
	ls.Call("setItem", storageKey(path), string(data))
	return nil
}

func removeFileData(path string) {
	ls := localStorage()
	if ls.IsUndefined() || ls.IsNull() {
		return
	}
	ls.Call("removeItem", storageKey(path))
}
