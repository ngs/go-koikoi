# tcell webfiles へのパッチ

`tcell.js` / `termstyle.css` / `beep.wav` は tcell 同梱の webfiles
(`$(go env GOMODCACHE)/github.com/gdamore/tcell/v2@<version>/webfiles/`)
のコピー。`tcell.js` には以下のパッチを当てている。tcell を更新して
webfiles を取り直す場合は再適用が必要
(`diff -u <webfiles>/tcell.js web/tcell.js` で差分確認)。

現在の取得元: `github.com/gdamore/tcell/v2@v2.13.11-0.20260802140533-eb51949b0a0e`
(v2 ブランチ。後述の上流化済みパッチを含む最初のタグ付きリリースはまだ無い)

## 現行パッチ

1. **ワイド文字への `.wide` クラス付与** (`drawCell` 内)
   - 上流の `drawCell` が受け取る幅引数 `w` が 2 以上のとき、span に
     `.wide` を付ける。**index.html の CSS
     `#terminal .wide { display: inline-block; width: 2ch; }`** で全角
     グリフを正確に半角2文字分にする（フォント依存の幅ずれ対策）。
     このパッチは index.html の CSS とセットで初めて成立する。

2. **`tcellTermSize()` の追加**
   - 現在のセル数 [cols, rows] を明示 API としてページへ公開する。
     index.html の `terminalCells()` がトップレベル `var` の window 漏れに
     依存しないようにするため。

3. **beep の autoplay reject 対策** (`beep()`)
   - 初回ユーザー操作前の `play()` は autoplay policy で reject される
     ため `.catch(() => {})` を付与。

## 上流化済み（再適用不要）

- **ワイド文字が覆うセルの清掃** — かつては JS 側で手書きの Unicode 範囲
  テーブル `isWide()` を持ち、幅2の文字を描いたら x+1 を空にしていた。
  [gdamore/tcell#1150](https://github.com/gdamore/tcell/pull/1150) が
  `wscreen.go` から `drawCell` に幅を渡すようにし、清掃も上流の
  `for (var i = x + 1; i < x + (w || 1) ...)` に入ったため削除。Go 側
  (uniseg) の幅判定がそのまま使われるので、旧 `isWide()` の既知の限界
  （絵文字の一部・異体字セレクタ・Ambiguous width）も解消。
- **クリック座標のセル換算を都度計算 + target 非依存化** (`eventCell()`) —
  [gdamore/tcell#1151](https://github.com/gdamore/tcell/pull/1151) として
  上流にマージ済み。

いずれも 2026-08-02 に tcell の `v2` ブランチへマージ。タグ付きリリース
（v2.13.11 以降の見込み）が出たら go.mod を pseudo-version からそのタグへ
差し替える。

## `index.html` 側の対になる契約

- `terminalCells()` → fork gocui (`ngs/gocui`) の `gui_js.go` が起動時に
  呼び、ウィンドウいっぱいのセル数でターミナルを開く
- `tcellSetSize(cols, rows)` → fork gocui が登録し、resize リスナーから
  呼ぶと実行中のターミナルがリサイズされる
