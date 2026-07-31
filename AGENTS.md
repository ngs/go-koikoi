# koikoi

Go 製の TUI 花札こいこいゲーム（任天堂ルール準拠）。
TUI フレームワークに `github.com/awesome-gocui/gocui` を使用。

- リポジトリ: https://github.com/ngs/go-koikoi
- モジュール: `go.ngs.io/koikoi`

## インストール

### Go

```bash
go install go.ngs.io/koikoi@latest
```

### Homebrew

```bash
brew tap ngs/tap
brew install ngs/tap/koikoi
```

## プロジェクト構成

| ファイル | 概要 |
|----------|------|
| `main.go` | エントリポイント。CLI フラグ解析、設定・セーブデータ読込、UI 起動 |
| `card.go` | Card 構造体、全48枚の定義、月・札種の定数 |
| `yaku.go` | 役判定ロジック（任天堂ルール準拠、排他制御含む） |
| `game.go` | Game 構造体、ラウンド管理、マッチング処理、得点計算 |
| `cpu.go` | CPU AI（3段階の難易度: かんたん/ふつう/つよい） |
| `ui.go` | gocui ベースの TUI。画面描画、キーバインド、ゲームフェーズ管理 |
| `settings.go` | 設定の永続化（ラウンド数、難易度） |
| `save.go` | ゲーム進捗のセーブ・ロード |

## ルール

花札こいこいのルール詳細は [docs/rules.md](docs/rules.md) を参照。

## ビルドと実行

```bash
go build -o koikoi .
./koikoi
# 設定ディレクトリを指定する場合
./koikoi -config /path/to/config
```

## 開発上の注意

- TUI は gocui (tcell ベース) を使用。CJK 全角文字はターミナル上で2カラム幅を占める
- gocui の `SetView` の overlap パラメータ (0) でポップアップの角が T 字にマージされるのを防止
- ポップアップの `FrameRunes` は11文字指定で、T 字結合位置にコーナー文字を割り当て
- 札の表示は `[月:種]` 形式（例: `[松:光]`）
- ゲーム状態は JSON でシリアライズ。札は ID (0-47) で保存

## ブラウザ (WASM) 版

- <https://koikoi.ngs.io> で公開。GitHub Pages + `.github/workflows/pages.yml`（master push で自動デプロイ）
- gocui は js/wasm 対応のため fork [ngs/gocui](https://github.com/ngs/gocui) を使用（upstream PR: awesome-gocui/gocui#138。マージされたら本家に戻す）
- `storage.go` / `storage_js.go`: 永続化の抽象。js ビルドでは localStorage に保存
- `web/`: 静的アセット。`tcell.js` は tcell 同梱の webfiles にパッチ済み。パッチは `web/tcell.js` 単体で完結せず **index.html の `.wide` CSS とセット**。一覧と再適用手順は [web/PATCHES.md](web/PATCHES.md)
- `web/index.html` の `terminalCells()` / `tcellSetSize()` は fork gocui の `gui_js.go` と対になる契約（ウィンドウいっぱいのセル数で起動・リサイズ追従）
- `make wasm` で `dist/` 生成、`make serve-wasm` でローカル確認（ポート使用中は自動で空きポート）
