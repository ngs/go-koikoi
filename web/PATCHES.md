# tcell webfiles へのパッチ

`tcell.js` / `termstyle.css` / `beep.wav` は tcell 同梱の webfiles
(`$(go env GOMODCACHE)/github.com/gdamore/tcell/v2@<version>/webfiles/`)
のコピー。`tcell.js` には以下のパッチを当てている。tcell を更新して
webfiles を取り直す場合は再適用が必要
(`diff -u <webfiles>/tcell.js web/tcell.js` で差分確認)。

1. **ワイド文字の隣接セル清掃と `.wide` クラス付与** (`drawCell` 内 + `isWide()`)
   - Go 側は幅2の文字を描くと x+1 のセルをスキップするが、JS 側の DOM は
     x+1 に古い文字が残る。ワイド文字描画時に x+1 を空にする。
   - あわせて span に `.wide` クラスを付け、**index.html の CSS
     `#terminal .wide { display: inline-block; width: 2ch; }`** で全角
     グリフを正確に半角2文字分にする（フォント依存の幅ずれ対策）。
     このパッチは index.html の CSS とセットで初めて成立する。
   - 既知の限界: `isWide()` は手書きの Unicode 範囲テーブルで、Go 側
     (uniseg) の幅判定と完全一致しない（絵文字の一部・異体字セレクタ・
     Ambiguous width）。こいこいの表示（仮名/漢字/罫線）では実害なし。
     根本対応は tcell 側 `drawCell` に width 引数を渡す上流 PR。

2. **クリック座標のセル換算を都度計算 + target 非依存化** (`eventCell()`)
   - 上流はロード時（ターミナルが空で幅0）に1回だけセル寸法を計算する
     ため、クリックが常に右下セル扱いになる。イベント毎に計算し直す。
   - `offsetX/offsetY` は event.target（色付きセルでは内側 span）基準で
     ずれるため、`clientX/clientY - getBoundingClientRect()` を使う。

3. **`tcellTermSize()` の追加**
   - 現在のセル数 [cols, rows] を明示 API としてページへ公開する。
     index.html の `terminalCells()` がトップレベル `var` の window 漏れに
     依存しないようにするため。

4. **beep の autoplay reject 対策** (`beep()`)
   - 初回ユーザー操作前の `play()` は autoplay policy で reject される
     ため `.catch(() => {})` を付与。

`index.html` 側の対になる契約:

- `terminalCells()` → fork gocui (`ngs/gocui`) の `gui_js.go` が起動時に
  呼び、ウィンドウいっぱいのセル数でターミナルを開く
- `tcellSetSize(cols, rows)` → fork gocui が登録し、resize リスナーから
  呼ぶと実行中のターミナルがリサイズされる
