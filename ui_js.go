//go:build js

package main

import (
	"syscall/js"

	"github.com/ngs/gocui"
)

// registerWebResize は koikoiSetSize(cols, rows) をグローバルへ公開し、
// ホストページがブラウザウィンドウに収まるセル数をターミナルへ渡せるようにする。
// ページ側は untrusted な境界なので、引数が数値でなければ黙って無視する。
//
// 公開したあとに resize イベントを1度発火させ、起動時のサイズ決定も
// ページ側の resize リスナー1本に集約する。
func registerWebResize(g *gocui.Gui) {
	js.Global().Set("koikoiSetSize", js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) != 2 || args[0].Type() != js.TypeNumber || args[1].Type() != js.TypeNumber {
			return nil
		}
		w, h := args[0].Int(), args[1].Int()
		if w > 0 && h > 0 {
			g.SetSize(w, h)
		}
		return nil
	}))
	// dispatchEvent / Event が無い環境でも落とさない
	defer func() { _ = recover() }()
	js.Global().Get("window").Call("dispatchEvent", js.Global().Get("Event").New("resize"))
}
