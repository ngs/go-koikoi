package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/awesome-gocui/gocui"
)

const (
	ansiReset   = "\033[0m"
	ansiReverse = "\033[7m"
	ansiYellow  = "\033[33m"
	ansiDim     = "\033[2m"
)

type Phase int

const (
	PhaseParentDetermination Phase = iota
	PhasePlayerSelectHand
	PhasePlayerSelectField
	PhasePlayerDrawResult
	PhasePlayerSelectFieldDraw
	PhaseKoiKoi
	PhaseCPUKoiKoi
	PhaseCPUTurn
	PhaseRoundEnd
	PhaseGameEnd
)

const (
	msgPlayerWin = "あなたの勝ちです！おめでとうございます！"
	msgCPUWin    = "CPUの勝ちです。次は頑張りましょう！"
	msgDraw      = "引き分けです！"

	labelCancel = " キャンセル "
	labelApply  = " 適用 "
)

type UI struct {
	gui           *gocui.Gui
	game          *Game
	phase         Phase
	cursor        int
	fieldCursor   int
	matchingCards []Card
	playedCard    Card
	drawnCard     Card
	logLines      []string
	koikoiCursor  int
	gameEndCursor int // ゲーム終了ダイアログ (0=1月から, 1=終了する)
	newYaku       []Yaku
	cpuKoiKoiYaku []Yaku     // CPUこいこい時の役（ダイアログ表示用）
	roundResult   string     // ラウンド結果メッセージ
	gameResult    string     // ゲーム終了結果メッセージ
	showLog       bool       // 行動履歴ポップアップ表示
	showHelp      bool       // ヘルプポップアップ表示
	showOptions   bool       // オプションポップアップ表示
	showQuitConf  bool       // 終了確認ポップアップ表示
	quitCursor    int        // 終了確認カーソル (0=はい, 1=いいえ)
	optCursor     int        // オプション画面のカーソル位置 (0=ラウンド, 1=難易度, 2=ボタン行)
	optBtnCursor  int        // オプションボタンカーソル (0=キャンセル, 1=適用)
	optRounds     int        // オプション: ラウンド数
	optDifficulty Difficulty // オプション: 難易度
	showOptConf   bool       // オプション適用確認ポップアップ
	optConfCursor int        // 適用確認カーソル (0=はい, 1=いいえ)

	// 親決めアニメーション
	parentDetStep    int   // 0=プレイヤー札回転中, 1=CPU札表示中, 2=結果表示
	parentDetDisplay *Card // 回転中に表示する札

	// 設定・保存パス
	configDir    string
	settingsPath string
	savePath     string
	settings     Settings
	difficulty   Difficulty
}

func NewUI() *UI {
	return &UI{
		logLines: nil,
	}
}

// Init は設定・セーブデータを読み込み、ゲームの初期状態をセットアップする。
func (u *UI) Init(configDir string) {
	u.configDir = configDir
	u.settingsPath = filepath.Join(configDir, "settings.json")
	u.savePath = filepath.Join(configDir, "game.json")
	u.settings, _ = LoadSettings(u.settingsPath)
	u.difficulty = u.settings.Difficulty

	if sd, err := LoadGame(u.savePath); err == nil {
		u.restoreSave(&sd)
	} else {
		u.newGame(u.settings.Rounds)
	}
}

func (u *UI) restoreSave(sd *SaveData) {
	u.game = SaveDataToGame(sd)
	u.logLines = sd.LogLines
	u.difficulty = sd.Difficulty
	if u.game.Round > u.game.MaxRounds {
		// 旧バージョンのセーブデータ（ゲーム終了状態）は破棄して新規開始
		DeleteSave(u.savePath)
		u.newGame(u.settings.Rounds)
		return
	}
	u.addLog("--- セーブデータを復元しました ---")
	u.setInitialPhase()
}

func (u *UI) newGame(rounds int) {
	u.game = NewGame(rounds)
	u.phase = PhaseParentDetermination
	u.parentDetStep = 0
	u.parentDetDisplay = nil
}

func (u *UI) logParentDetermination() {
	if u.game.PlayerDrawnCard == nil || u.game.CPUDrawnCard == nil {
		return
	}
	playerCard := u.game.PlayerDrawnCard
	cpuCard := u.game.CPUDrawnCard
	u.addLog("--- 親決め ---")
	u.addLog(fmt.Sprintf("  あなた: %s  CPU: %s", playerCard.Name, cpuCard.Name))
	if u.game.NextParentIsPlayer {
		u.addLog("  → あなたが親（先攻）")
	} else {
		u.addLog("  → CPUが親（先攻）")
	}
}

func (u *UI) setInitialPhase() {
	u.phase = PhasePlayerSelectHand
	if !u.game.IsPlayerTurn {
		u.phase = PhaseCPUTurn
	}
}

func (u *UI) addLog(msg string) {
	ts := time.Now().Format("2006-01-02 15:04:05")
	u.logLines = append(u.logLines, fmt.Sprintf("[%s] %s", ts, msg))
	if len(u.logLines) > 1000 {
		u.logLines = u.logLines[len(u.logLines)-1000:]
	}
}

func (u *UI) autoSave() {
	sd := GameToSaveData(u.game, u.difficulty, u.logLines)
	_ = SaveGame(u.savePath, &sd)
}

func (u *UI) Run() error {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		return err
	}
	defer g.Close()

	u.gui = g
	g.SetManagerFunc(u.layout)
	g.Cursor = false
	g.Mouse = true
	g.ASCII = false

	if err := u.setKeybindings(g); err != nil {
		return err
	}

	return g.MainLoop()
}

func (u *UI) roundName() string {
	m := Month((u.game.Round - 1) % 12)
	return m.OldName()
}

// ---- タイトルオーバーレイ ----

func setOverlayTitle(g *gocui.Gui, name string, x0, y0, x1 int, title string) {
	viewName := name + "_t"
	w := cellWidth(title)
	tx1 := min(x0+2+w, x1-1)
	v, err := g.SetView(viewName, x0+1, y0-1, tx1, y0+1, 1)
	if err != nil && !errors.Is(err, gocui.ErrUnknownView) {
		return
	}
	v.Frame = false
	v.Clear()
	fmt.Fprint(v, title)
}

func cellWidth(s string) int {
	w := 0
	for _, r := range s {
		if isWide(r) {
			w += 2
		} else {
			w++
		}
	}
	return w
}

func isWide(r rune) bool {
	return r >= 0x1100 &&
		(r <= 0x115f || r == 0x2329 || r == 0x232a ||
			(r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) ||
			(r >= 0xac00 && r <= 0xd7a3) ||
			(r >= 0xf900 && r <= 0xfaff) ||
			(r >= 0xfe10 && r <= 0xfe19) ||
			(r >= 0xfe30 && r <= 0xfe6f) ||
			(r >= 0xff01 && r <= 0xff60) ||
			(r >= 0xffe0 && r <= 0xffe6) ||
			(r >= 0x20000 && r <= 0x2fffd) ||
			(r >= 0x30000 && r <= 0x3fffd))
}

// ---- レイアウト ----

func (u *UI) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	headerH := 6 // ヘッダー(内部: 空行+ボックス3行+空行)
	statusH := 2 // ステータス(内部1行)

	// 左カラム(60%), 右カラム(40%)
	leftW := maxX * 60 / 100
	bodyTop := headerH
	bodyBot := maxY - statusH

	// 左カラム3分割: CPU(5行), 場札(5行), 手札(残り)
	cpuLeftH := 5
	fieldLeftH := 5

	type R struct{ x0, y0, x1, y1 int }

	rHeader := R{-1, -1, maxX, headerH}
	rCPU := R{-1, bodyTop - 1, leftW, bodyTop - 1 + cpuLeftH}
	rField := R{-1, rCPU.y1 - 1, leftW, rCPU.y1 - 1 + fieldLeftH}
	rHand := R{-1, rField.y1 - 1, leftW, bodyBot}

	// 右カラムは獲得札の内容に応じて動的にサイズ変更
	cpuCapLines := countCapturedGroups(u.game.CPUCaptured) + 2 // 前後空白行
	cpuCapH := max(cpuCapLines+2, 4)                           // フレーム分
	deckH := 5

	rCpuCap := R{leftW - 1, bodyTop - 1, maxX, bodyTop - 1 + cpuCapH}
	rDeck := R{leftW - 1, rCpuCap.y1 - 1, maxX, rCpuCap.y1 - 1 + deckH}
	rMyCap := R{leftW - 1, rDeck.y1 - 1, maxX, bodyBot}

	rStatus := R{-1, bodyBot - 1, maxX, maxY}

	// --- ヘッダー ---
	if v, err := g.SetView("header", rHeader.x0, rHeader.y0, rHeader.x1, rHeader.y1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
	}
	u.drawHeader(g)

	// --- CPU (左上) ---
	if v, err := g.SetView("cpu", rCPU.x0, rCPU.y0, rCPU.x1, rCPU.y1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
	}
	setOverlayTitle(g, "cpu", rCPU.x0, rCPU.y0, rCPU.x1, " CPU ")
	u.drawCPU(g)

	// --- 場札 (左中) ---
	if v, err := g.SetView("field", rField.x0, rField.y0, rField.x1, rField.y1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
	}
	setOverlayTitle(g, "field", rField.x0, rField.y0, rField.x1, " 場札 ")
	u.drawField(g)

	// --- 手札 (左下) ---
	if v, err := g.SetView("hand", rHand.x0, rHand.y0, rHand.x1, rHand.y1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
	}
	u.drawHand(g, rHand.x0, rHand.y0, rHand.x1)

	// --- CPU獲得札 (右上) ---
	if v, err := g.SetView("cpucap", rCpuCap.x0, rCpuCap.y0, rCpuCap.x1, rCpuCap.y1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
		v.Wrap = true
	}
	setOverlayTitle(g, "cpucap", rCpuCap.x0, rCpuCap.y0, rCpuCap.x1, " CPU獲得札 ")
	u.drawCpuCaptured(g)

	// --- 山札 (右中) ---
	if v, err := g.SetView("deck", rDeck.x0, rDeck.y0, rDeck.x1, rDeck.y1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
	}
	setOverlayTitle(g, "deck", rDeck.x0, rDeck.y0, rDeck.x1, " 山札 ")
	u.drawDeck(g)

	// --- 自分の獲得札 (右下) ---
	if v, err := g.SetView("mycap", rMyCap.x0, rMyCap.y0, rMyCap.x1, rMyCap.y1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
		v.Wrap = true
	}
	setOverlayTitle(g, "mycap", rMyCap.x0, rMyCap.y0, rMyCap.x1, " 獲得札 ")
	u.drawMyCaptured(g)

	// --- ステータスバー ---
	if v, err := g.SetView("status", rStatus.x0, rStatus.y0, rStatus.x1, rStatus.y1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
	}
	u.drawStatus(g)

	// --- 行動履歴ポップアップ ---
	if u.showLog {
		pw := maxX * 60 / 100
		ph := maxY * 70 / 100
		px0 := (maxX - pw) / 2
		py0 := (maxY - ph) / 2
		if v, err := g.SetView("log", px0, py0, px0+pw, py0+ph, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Frame = true
			v.FrameRunes = []rune{'━', '┃', '┏', '┓', '┗', '┛', '┏', '┓', '┏', '┗', '┏'}
			v.Wrap = true
			v.Autoscroll = true
		}
		setOverlayTitle(g, "log", px0, py0, px0+pw, " 行動履歴 (lキーで閉じる) ")
		u.drawLog(g)
	} else {
		_ = g.DeleteView("log")
		_ = g.DeleteView("log_t")
	}

	// --- ヘルプポップアップ ---
	if u.showHelp {
		pw := 62
		ph := 30
		px0 := (maxX - pw) / 2
		py0 := (maxY - ph) / 2
		if v, err := g.SetView("help", px0, py0, px0+pw, py0+ph, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Frame = true
			v.FrameRunes = []rune{'━', '┃', '┏', '┓', '┗', '┛', '┏', '┓', '┏', '┗', '┏'}
		}
		setOverlayTitle(g, "help", px0, py0, px0+pw, " ヘルプ (?キーで閉じる) ")
		u.drawHelp(g)
	} else {
		_ = g.DeleteView("help")
		_ = g.DeleteView("help_t")
	}

	// --- オプションポップアップ ---
	if u.showOptions {
		pw := 44
		ph := 11
		px0 := (maxX - pw) / 2
		py0 := (maxY - ph) / 2
		if v, err := g.SetView("options", px0, py0, px0+pw, py0+ph, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Frame = true
			v.FrameRunes = []rune{'━', '┃', '┏', '┓', '┗', '┛', '┏', '┓', '┏', '┗', '┏'}
		}
		setOverlayTitle(g, "options", px0, py0, px0+pw, " オプション (Escで閉じる) ")
		u.drawOptions(g)
	} else {
		_ = g.DeleteView("options")
		_ = g.DeleteView("options_t")
	}

	// --- 終了確認ポップアップ ---
	if u.showQuitConf {
		pw := 46
		ph := 10
		px0 := (maxX - pw) / 2
		py0 := (maxY - ph) / 2
		if v, err := g.SetView("quitconf", px0, py0, px0+pw, py0+ph, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Frame = true
			v.FrameRunes = []rune{'━', '┃', '┏', '┓', '┗', '┛', '┏', '┓', '┏', '┗', '┏'}
		}
		setOverlayTitle(g, "quitconf", px0, py0, px0+pw, " 終了確認 ")
		u.drawQuitConf(g)
	} else {
		_ = g.DeleteView("quitconf")
		_ = g.DeleteView("quitconf_t")
	}

	// --- オプション適用確認ポップアップ ---
	if u.showOptConf {
		pw := 46
		ph := 10
		px0 := (maxX - pw) / 2
		py0 := (maxY - ph) / 2
		if v, err := g.SetView("optconf", px0, py0, px0+pw, py0+ph, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Frame = true
			v.FrameRunes = []rune{'━', '┃', '┏', '┓', '┗', '┛', '┏', '┓', '┏', '┗', '┏'}
		}
		setOverlayTitle(g, "optconf", px0, py0, px0+pw, " 確認 ")
		u.drawOptConf(g)
	} else {
		_ = g.DeleteView("optconf")
		_ = g.DeleteView("optconf_t")
	}

	// --- こいこいポップアップ ---
	if u.phase == PhaseKoiKoi {
		pw := 44
		yakuCount := len(u.newYaku)
		ph := 10 + yakuCount // header + yakus + total + buttons + footer
		px0 := (maxX - pw) / 2
		py0 := (maxY - ph) / 2
		if v, err := g.SetView("koikoi", px0, py0, px0+pw, py0+ph, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Frame = true
			v.FrameRunes = []rune{'━', '┃', '┏', '┓', '┗', '┛', '┏', '┓', '┏', '┗', '┏'}
		}
		setOverlayTitle(g, "koikoi", px0, py0, px0+pw, " こいこい？ ")
		u.drawKoiKoi(g)
	} else {
		_ = g.DeleteView("koikoi")
		_ = g.DeleteView("koikoi_t")
	}

	// --- 親決めポップアップ ---
	if u.phase == PhaseParentDetermination {
		pw := 44
		ph := 11
		px0 := (maxX - pw) / 2
		py0 := (maxY - ph) / 2
		if v, err := g.SetView("parentdet", px0, py0, px0+pw, py0+ph, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Frame = true
			v.FrameRunes = []rune{'━', '┃', '┏', '┓', '┗', '┛', '┏', '┓', '┏', '┗', '┏'}
		}
		setOverlayTitle(g, "parentdet", px0, py0, px0+pw, " 親決め ")
		u.drawParentDetermination(g)
		// アニメーション開始
		if u.parentDetStep == 0 {
			u.startParentDetAnimation(g)
		}
	} else {
		_ = g.DeleteView("parentdet")
		_ = g.DeleteView("parentdet_t")
	}

	// --- CPUこいこいポップアップ ---
	if u.phase == PhaseCPUKoiKoi {
		pw := 44
		yakuCount := len(u.cpuKoiKoiYaku)
		ph := 10 + yakuCount
		px0 := (maxX - pw) / 2
		py0 := (maxY - ph) / 2
		if v, err := g.SetView("cpukoikoi", px0, py0, px0+pw, py0+ph, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Frame = true
			v.FrameRunes = []rune{'━', '┃', '┏', '┓', '┗', '┛', '┏', '┓', '┏', '┗', '┏'}
		}
		setOverlayTitle(g, "cpukoikoi", px0, py0, px0+pw, " CPUがこいこい！ ")
		u.drawCPUKoiKoi(g)
	} else {
		_ = g.DeleteView("cpukoikoi")
		_ = g.DeleteView("cpukoikoi_t")
	}

	// --- ラウンド終了ポップアップ ---
	if u.phase == PhaseRoundEnd {
		pw := 44
		ph := 11
		px0 := (maxX - pw) / 2
		py0 := (maxY - ph) / 2
		if v, err := g.SetView("roundend", px0, py0, px0+pw, py0+ph, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Frame = true
			v.FrameRunes = []rune{'━', '┃', '┏', '┓', '┗', '┛', '┏', '┓', '┏', '┗', '┏'}
		}
		setOverlayTitle(g, "roundend", px0, py0, px0+pw, " ラウンド終了 ")
		u.drawRoundEnd(g)
	} else {
		_ = g.DeleteView("roundend")
		_ = g.DeleteView("roundend_t")
	}

	// --- ゲーム終了ポップアップ ---
	if u.phase == PhaseGameEnd {
		pw := 44
		ph := 13
		px0 := (maxX - pw) / 2
		py0 := (maxY - ph) / 2
		if v, err := g.SetView("gameend", px0, py0, px0+pw, py0+ph, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Frame = true
			v.FrameRunes = []rune{'━', '┃', '┏', '┓', '┗', '┛', '┏', '┓', '┏', '┗', '┏'}
		}
		setOverlayTitle(g, "gameend", px0, py0, px0+pw, " ゲーム終了 ")
		u.drawGameEnd(g)
	} else {
		_ = g.DeleteView("gameend")
		_ = g.DeleteView("gameend_t")
	}

	if u.phase == PhaseCPUTurn {
		go u.doCPUTurn(g)
	}

	return nil
}

// ---- 描画 ----

func (u *UI) drawHeader(g *gocui.Gui) {
	v, _ := g.View("header")
	if v == nil {
		return
	}
	v.Clear()

	m := Month((u.game.Round - 1) % 12)
	leftText := fmt.Sprintf(" %s %d/%d ", m.OldName(), u.game.Round, u.game.MaxRounds)
	rightText := fmt.Sprintf(" あなた %d文 / CPU %d文 ", u.game.PlayerScore, u.game.CPUScore)

	lw := cellWidth(leftText)
	rw := cellWidth(rightText)

	fmt.Fprintln(v)
	fmt.Fprintf(v, "  ┏%s┳%s┓\n", strings.Repeat("━", lw), strings.Repeat("━", rw))
	fmt.Fprintf(v, "  ┃%s┃%s┃\n", leftText, rightText)
	fmt.Fprintf(v, "  ┗%s┻%s┛\n", strings.Repeat("━", lw), strings.Repeat("━", rw))
}

func (u *UI) drawCPU(g *gocui.Gui) {
	v, _ := g.View("cpu")
	if v == nil {
		return
	}
	v.Clear()
	backs := ""
	for range u.game.CPUHand {
		backs += "[??] "
	}
	fmt.Fprintln(v)
	fmt.Fprintf(v, " 手札(%d枚): %s\n", len(u.game.CPUHand), backs)
	fmt.Fprintln(v)
}

func (u *UI) drawField(g *gocui.Gui) {
	v, _ := g.View("field")
	if v == nil {
		return
	}
	v.Clear()

	switch u.phase {
	case PhasePlayerSelectField, PhasePlayerSelectFieldDraw:
		var cards []string
		for _, c := range u.game.Field {
			isMatch := false
			mIdx := 0
			for mi, m := range u.matchingCards {
				if m.ID == c.ID {
					isMatch = true
					mIdx = mi
					break
				}
			}
			label := cardLabel(c)
			switch {
			case isMatch && mIdx == u.fieldCursor:
				cards = append(cards, ansiReverse+label+ansiReset)
			case isMatch:
				cards = append(cards, ansiYellow+label+ansiReset)
			default:
				cards = append(cards, ansiDim+label+ansiReset)
			}
		}
		fmt.Fprintln(v)
		fmt.Fprintln(v, " "+strings.Join(cards, " "))

	case PhasePlayerSelectHand:
		var matchIDs map[int]bool
		if u.cursor < len(u.game.PlayerHand) {
			hovered := u.game.PlayerHand[u.cursor]
			matches := u.game.MatchingFieldCards(hovered)
			matchIDs = make(map[int]bool)
			for _, m := range matches {
				matchIDs[m.ID] = true
			}
		}
		var cards []string
		for _, c := range u.game.Field {
			label := cardLabel(c)
			if matchIDs != nil && matchIDs[c.ID] {
				cards = append(cards, ansiYellow+label+ansiReset)
			} else {
				cards = append(cards, label)
			}
		}
		if len(cards) == 0 {
			fmt.Fprintln(v, " (なし)")
		} else {
			fmt.Fprintln(v)
			fmt.Fprintln(v, " "+strings.Join(cards, " "))
		}

	default:
		var parts []string
		for _, c := range u.game.Field {
			parts = append(parts, cardLabel(c))
		}
		if len(parts) == 0 {
			fmt.Fprintln(v, " (なし)")
		} else {
			fmt.Fprintln(v)
			fmt.Fprintln(v, " "+strings.Join(parts, " "))
		}
	}
}

func (u *UI) drawHand(g *gocui.Gui, x0, y0, x1 int) {
	v, _ := g.View("hand")
	if v == nil {
		return
	}
	v.Clear()

	switch u.phase {
	case PhasePlayerSelectHand:
		setOverlayTitle(g, "hand", x0, y0, x1, " 手札 ")
		fmt.Fprintln(v)
		for i, c := range u.game.PlayerHand {
			matches := u.game.MatchingFieldCards(c)
			matchInfo := ""
			if len(matches) > 0 {
				matchInfo = fmt.Sprintf(" (%d)", len(matches))
			}
			line := fmt.Sprintf(" %s %s%s", cardLabel(c), c.Name, matchInfo)
			if i == u.cursor {
				fmt.Fprintf(v, "%s%s%s\n", ansiReverse, line, ansiReset)
			} else {
				fmt.Fprintf(v, "%s\n", line)
			}
		}

	case PhasePlayerSelectField, PhasePlayerSelectFieldDraw:
		setOverlayTitle(g, "hand", x0, y0, x1, " 手札 ")
		fmt.Fprintln(v)
		for _, c := range u.game.PlayerHand {
			fmt.Fprintf(v, " %s %s\n", cardLabel(c), c.Name)
		}
		fmt.Fprintln(v)
		fmt.Fprintln(v, "  Left/Right で場札を選択、Enterで決定")

	case PhaseKoiKoi:
		setOverlayTitle(g, "hand", x0, y0, x1, " 手札 ")
		fmt.Fprintln(v)
		for _, c := range u.game.PlayerHand {
			fmt.Fprintf(v, " %s %s\n", cardLabel(c), c.Name)
		}

	case PhaseRoundEnd:
		setOverlayTitle(g, "hand", x0, y0, x1, " 手札 ")

	case PhaseGameEnd:
		setOverlayTitle(g, "hand", x0, y0, x1, " 手札 ")

	default:
		setOverlayTitle(g, "hand", x0, y0, x1, " 手札 ")
		fmt.Fprintln(v)
		for _, c := range u.game.PlayerHand {
			fmt.Fprintf(v, " %s %s\n", cardLabel(c), c.Name)
		}
	}
}

func (u *UI) drawCpuCaptured(g *gocui.Gui) {
	v, _ := g.View("cpucap")
	if v == nil {
		return
	}
	v.Clear()
	fmt.Fprintln(v)
	writeCapturedDetail(v, u.game.CPUCaptured)
	fmt.Fprintln(v)
}

func (u *UI) drawDeck(g *gocui.Gui) {
	v, _ := g.View("deck")
	if v == nil {
		return
	}
	v.Clear()
	fmt.Fprintln(v)
	fmt.Fprintf(v, " 残り%d枚\n", len(u.game.Deck))
	fmt.Fprintln(v)
}

func (u *UI) drawMyCaptured(g *gocui.Gui) {
	v, _ := g.View("mycap")
	if v == nil {
		return
	}
	v.Clear()
	fmt.Fprintln(v)
	writeCapturedDetail(v, u.game.PlayerCaptured)

	// リーチ表示
	reaches := CheckReach(u.game.PlayerCaptured, u.game.CPUCaptured)
	if len(reaches) > 0 {
		fmt.Fprintln(v)
		fmt.Fprintf(v, " %s── リーチ ──%s\n", ansiDim, ansiReset)
		for _, r := range reaches {
			if len(r.Missing) > 0 {
				var names []string
				for _, c := range r.Missing {
					names = append(names, c.Name)
				}
				fmt.Fprintf(v, " %s%s: %s%s\n", ansiDim, r.Name, strings.Join(names, " / "), ansiReset)
			} else {
				fmt.Fprintf(v, " %s%s: あと1枚%s\n", ansiDim, r.Name, ansiReset)
			}
		}
	}

	fmt.Fprintln(v)
}

func (u *UI) drawLog(g *gocui.Gui) {
	v, _ := g.View("log")
	if v == nil {
		return
	}
	v.Clear()
	start := 0
	if len(u.logLines) > 50 {
		start = len(u.logLines) - 50
	}
	for _, l := range u.logLines[start:] {
		fmt.Fprintln(v, l)
	}
}

func (u *UI) drawHelp(g *gocui.Gui) {
	v, _ := g.View("help")
	if v == nil {
		return
	}
	v.Clear()
	fmt.Fprintln(v)
	fmt.Fprintln(v, "  === 操作 ================================================")
	fmt.Fprintln(v)
	fmt.Fprintln(v, "  ↑/↓ 手札選択  ←/→ 場札選択  Enter 決定")
	fmt.Fprintln(v, "  マウスクリック: 手札/場札/ボタンを直接選択")
	fmt.Fprintln(v, "  l 行動履歴  ? ヘルプ  q 終了  Ctrl+C 強制終了")
	fmt.Fprintln(v)
	fmt.Fprintln(v, "  === 出来役一覧 ==========================================")
	fmt.Fprintln(v)
	fmt.Fprintln(v, "  五光　　　　　 10文  光札5枚")
	fmt.Fprintln(v, "  四光　　　　　  8文  柳以外の光札4枚")
	fmt.Fprintln(v, "  雨四光　　　　  7文  柳を含む光札4枚")
	fmt.Fprintln(v, "  三光　　　　　  5文  柳以外の光札3枚")
	fmt.Fprintln(v, "  花見で一杯　　  5文  桜に幕 + 菊に盃")
	fmt.Fprintln(v, "  月見で一杯　　  5文  芒に月 + 菊に盃")
	fmt.Fprintln(v, "  猪鹿蝶　　　　  5文  猪・鹿・蝶 (+1文/種札)")
	fmt.Fprintln(v, "  赤短　　　　　  5文  松・梅・桜の短冊 (+1文/短冊)")
	fmt.Fprintln(v, "  青短　　　　　  5文  牡丹・菊・紅葉の短冊 (+1文/短冊)")
	fmt.Fprintln(v, "  赤短・青短重複 10文  赤短+青短 (+1文/短冊)")
	fmt.Fprintln(v, "  タネ　　　　　  1文  種札5枚～ (+1文/枚)")
	fmt.Fprintln(v, "  タン　　　　　  1文  短冊札5枚～ (+1文/枚)")
	fmt.Fprintln(v, "  カス　　　　　  1文  カス札10枚～ (+1文/枚)")
	fmt.Fprintln(v)
	fmt.Fprintln(v, "  === 特殊ルール ==========================================")
	fmt.Fprintln(v)
	fmt.Fprintln(v, "  •7文以上で得点2倍")
	fmt.Fprintln(v, "  •こいこい後に相手が上がると相手の得点2倍")
	fmt.Fprintln(v)
}

func centerPadLabel(text string, width int, highlighted bool) string {
	w := cellWidth(text)
	if w >= width {
		if highlighted {
			return ansiReverse + text + ansiReset
		}
		return text
	}
	left := (width - w) / 2
	right := width - w - left
	padded := strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
	if highlighted {
		return ansiReverse + padded + ansiReset
	}
	return padded
}

func (u *UI) drawOptions(g *gocui.Gui) {
	v, _ := g.View("options")
	if v == nil {
		return
	}
	v.Clear()

	innerW := 42 // pw(44) - frame(2)

	fmt.Fprintln(v)

	// ラウンド数
	roundsLabel := fmt.Sprintf("ラウンド数:  ◀ %d ▶", u.optRounds)
	fmt.Fprintln(v, centerPadLabel(roundsLabel, innerW, u.optCursor == 0))
	fmt.Fprintln(v)

	// 難易度
	diffLabel := fmt.Sprintf("CPU難易度:   ◀ %s ▶", u.optDifficulty.Label())
	fmt.Fprintln(v, centerPadLabel(diffLabel, innerW, u.optCursor == 1))
	fmt.Fprintln(v)

	// キャンセル / 適用ボタン
	cancelStr := labelCancel
	applyStr := labelApply
	if u.optCursor == 2 && u.optBtnCursor == 0 {
		cancelStr = ansiReverse + cancelStr + ansiReset
	}
	if u.optCursor == 2 && u.optBtnCursor == 1 {
		applyStr = ansiReverse + applyStr + ansiReset
	}
	btnLine := cancelStr + "    " + applyStr
	btnW := cellWidth(labelCancel) + 4 + cellWidth(labelApply)
	btnPad := max((innerW-btnW)/2, 0)
	fmt.Fprintln(v, strings.Repeat(" ", btnPad)+btnLine)
	fmt.Fprintln(v)
	fmt.Fprintln(v)
	fmt.Fprintln(v, centerPad("↑/↓:項目  ←/→:変更  Enter:決定", innerW))
}

func centerPad(text string, width int) string {
	w := cellWidth(text)
	if w >= width {
		return text
	}
	left := (width - w) / 2
	return strings.Repeat(" ", left) + text
}

func (u *UI) drawQuitConf(g *gocui.Gui) {
	v, _ := g.View("quitconf")
	if v == nil {
		return
	}
	v.Clear()

	innerW := 44 // pw(46) - frame(2)

	fmt.Fprintln(v)
	fmt.Fprintln(v, centerPad("終了しますか？", innerW))
	fmt.Fprintln(v, centerPad("(進捗は保存されます)", innerW))
	fmt.Fprintln(v)

	yesLabel := " はい "
	noLabel := " いいえ "
	if u.quitCursor == 0 {
		yesLabel = ansiReverse + yesLabel + ansiReset
	}
	if u.quitCursor == 1 {
		noLabel = ansiReverse + noLabel + ansiReset
	}
	btnLine := yesLabel + "    " + noLabel
	// ボタン部分のセルの幅（ANSIエスケープを除く）
	btnW := cellWidth(" はい ") + 4 + cellWidth(" いいえ ")
	btnPad := max((innerW-btnW)/2, 0)
	fmt.Fprintln(v, strings.Repeat(" ", btnPad)+btnLine)
	fmt.Fprintln(v)
	fmt.Fprintln(v)
	fmt.Fprintln(v, centerPad("y:はい  n/q/Esc:いいえ  ←/→:選択", innerW))
}

func (u *UI) drawOptConf(g *gocui.Gui) {
	v, _ := g.View("optconf")
	if v == nil {
		return
	}
	v.Clear()

	innerW := 44

	fmt.Fprintln(v)
	fmt.Fprintln(v, centerPad("設定を適用してゲームをリセットします", innerW))
	fmt.Fprintln(v, centerPad("よろしいですか？", innerW))
	fmt.Fprintln(v)

	yesLabel := " はい "
	noLabel := " いいえ "
	if u.optConfCursor == 0 {
		yesLabel = ansiReverse + yesLabel + ansiReset
	}
	if u.optConfCursor == 1 {
		noLabel = ansiReverse + noLabel + ansiReset
	}
	btnLine := yesLabel + "    " + noLabel
	btnW := cellWidth(" はい ") + 4 + cellWidth(" いいえ ")
	btnPad := max((innerW-btnW)/2, 0)
	fmt.Fprintln(v, strings.Repeat(" ", btnPad)+btnLine)
	fmt.Fprintln(v)
	fmt.Fprintln(v)
	fmt.Fprintln(v, centerPad("←/→:選択  Enter:決定  Esc:戻る", innerW))
}

func (u *UI) drawKoiKoi(g *gocui.Gui) {
	v, _ := g.View("koikoi")
	if v == nil {
		return
	}
	v.Clear()

	innerW := 42 // pw(44) - frame(2)

	fmt.Fprintln(v)
	fmt.Fprintln(v, centerPad("*** 役が成立しました！ ***", innerW))
	fmt.Fprintln(v)
	for _, y := range u.newYaku {
		fmt.Fprintln(v, centerPad(fmt.Sprintf("%s (%d文)", y.Name, y.Points), innerW))
	}
	fmt.Fprintln(v)
	basePoints := TotalPoints(CheckYaku(u.game.PlayerCaptured))
	finalScore := u.game.CalcFinalScore(true)
	if finalScore != basePoints {
		fmt.Fprintln(v, centerPad(fmt.Sprintf("合計: %d文 → %d文", basePoints, finalScore), innerW))
	} else {
		fmt.Fprintln(v, centerPad(fmt.Sprintf("合計: %d文", basePoints), innerW))
	}
	fmt.Fprintln(v)

	koiLabel := " こいこい（続行） "
	shoubuLabel := " 勝負（得点確定） "
	if u.koikoiCursor == 0 {
		koiLabel = ansiReverse + koiLabel + ansiReset
	}
	if u.koikoiCursor == 1 {
		shoubuLabel = ansiReverse + shoubuLabel + ansiReset
	}
	btnLine := koiLabel + "  " + shoubuLabel
	btnW := cellWidth(" こいこい（続行） ") + 2 + cellWidth(" 勝負（得点確定） ")
	btnPad := max((innerW-btnW)/2, 0)
	fmt.Fprintln(v, strings.Repeat(" ", btnPad)+btnLine)
	fmt.Fprintln(v)
	fmt.Fprintln(v, centerPad("←/→:選択  Enter:決定", innerW))
}

func (u *UI) drawParentDetermination(g *gocui.Gui) {
	v, _ := g.View("parentdet")
	if v == nil {
		return
	}
	v.Clear()

	innerW := 42

	fmt.Fprintln(v)
	fmt.Fprintln(v, centerPad("山札から札を引いて親を決めます", innerW))
	fmt.Fprintln(v)

	switch u.parentDetStep {
	case 0:
		// プレイヤーの札が回転中
		displayCard := "？？？"
		if u.parentDetDisplay != nil {
			displayCard = u.parentDetDisplay.Name
		}
		fmt.Fprintln(v, centerPad(fmt.Sprintf("あなた: %s", displayCard), innerW))
		fmt.Fprintln(v, centerPad("CPU:    ---", innerW))
		fmt.Fprintln(v)
		fmt.Fprintln(v, centerPad("", innerW))
		fmt.Fprintln(v)
		stopLabel := ansiReverse + " 止める " + ansiReset
		stopW := cellWidth(" 止める ")
		stopPad := max((innerW-stopW)/2, 0)
		fmt.Fprintln(v, strings.Repeat(" ", stopPad)+stopLabel)
		fmt.Fprintln(v)
		fmt.Fprintln(v, centerPad("Enter:止める", innerW))
	case 1:
		// CPUの札を表示中（ディレイ後に結果へ）
		fmt.Fprintln(v, centerPad(fmt.Sprintf("あなた: %s", u.game.PlayerDrawnCard.Name), innerW))
		displayCard := "？？？"
		if u.parentDetDisplay != nil {
			displayCard = u.parentDetDisplay.Name
		}
		fmt.Fprintln(v, centerPad(fmt.Sprintf("CPU:    %s", displayCard), innerW))
		fmt.Fprintln(v)
		fmt.Fprintln(v, centerPad("", innerW))
		fmt.Fprintln(v)
		fmt.Fprintln(v, centerPad("", innerW))
		fmt.Fprintln(v)
		fmt.Fprintln(v, centerPad("", innerW))
	default:
		// 結果表示
		fmt.Fprintln(v, centerPad(fmt.Sprintf("あなた: %s", u.game.PlayerDrawnCard.Name), innerW))
		fmt.Fprintln(v, centerPad(fmt.Sprintf("CPU:    %s", u.game.CPUDrawnCard.Name), innerW))
		fmt.Fprintln(v)
		if u.game.NextParentIsPlayer {
			fmt.Fprintln(v, centerPad("→ あなたが親（先攻）です", innerW))
		} else {
			fmt.Fprintln(v, centerPad("→ CPUが親（先攻）です", innerW))
		}
		fmt.Fprintln(v)
		okLabel := ansiReverse + "  OK  " + ansiReset
		okW := cellWidth("  OK  ")
		okPad := max((innerW-okW)/2, 0)
		fmt.Fprintln(v, strings.Repeat(" ", okPad)+okLabel)
		fmt.Fprintln(v)
		fmt.Fprintln(v, centerPad("Enter:続ける", innerW))
	}
}

func (u *UI) drawCPUKoiKoi(g *gocui.Gui) {
	v, _ := g.View("cpukoikoi")
	if v == nil {
		return
	}
	v.Clear()

	innerW := 42

	fmt.Fprintln(v)
	fmt.Fprintln(v, centerPad("*** CPUがこいこいしました ***", innerW))
	fmt.Fprintln(v)
	for _, y := range u.cpuKoiKoiYaku {
		fmt.Fprintln(v, centerPad(fmt.Sprintf("%s (%d文)", y.Name, y.Points), innerW))
	}
	fmt.Fprintln(v)
	basePoints := TotalPoints(CheckYaku(u.game.CPUCaptured))
	fmt.Fprintln(v, centerPad(fmt.Sprintf("合計: %d文", basePoints), innerW))
	fmt.Fprintln(v)

	okLabel := ansiReverse + "  OK  " + ansiReset
	okW := cellWidth("  OK  ")
	okPad := max((innerW-okW)/2, 0)
	fmt.Fprintln(v, strings.Repeat(" ", okPad)+okLabel)
	fmt.Fprintln(v)
	fmt.Fprintln(v, centerPad("Enter:閉じる", innerW))
}

func (u *UI) drawRoundEnd(g *gocui.Gui) {
	v, _ := g.View("roundend")
	if v == nil {
		return
	}
	v.Clear()

	innerW := 42 // pw(44) - frame(2)

	// ラウンド番号は Round-1 (finishRound で既にインクリメント済み)
	roundNum := u.game.Round - 1
	m := Month((roundNum - 1) % 12)

	fmt.Fprintln(v)
	fmt.Fprintln(v, centerPad(fmt.Sprintf("第%d局 (%s) 終了", roundNum, m.OldName()), innerW))
	fmt.Fprintln(v)
	fmt.Fprintln(v, centerPad(u.roundResult, innerW))
	fmt.Fprintln(v)
	fmt.Fprintln(v, centerPad(fmt.Sprintf("あなた: %d文  CPU: %d文", u.game.PlayerScore, u.game.CPUScore), innerW))
	fmt.Fprintln(v)

	nextM := Month((u.game.Round - 1) % 12)
	btnLabel := fmt.Sprintf(" 次の月へ（%s） ", nextM.OldName())
	btnW := cellWidth(btnLabel)
	btnPad := max((innerW-btnW)/2, 0)
	fmt.Fprintln(v, strings.Repeat(" ", btnPad)+ansiReverse+btnLabel+ansiReset)
}

func (u *UI) drawGameEnd(g *gocui.Gui) {
	v, _ := g.View("gameend")
	if v == nil {
		return
	}
	v.Clear()

	innerW := 42 // pw(44) - frame(2)

	fmt.Fprintln(v)
	fmt.Fprintln(v, centerPad(fmt.Sprintf("あなた: %d文  CPU: %d文", u.game.PlayerScore, u.game.CPUScore), innerW))
	fmt.Fprintln(v)
	fmt.Fprintln(v, centerPad(u.gameResult, innerW))
	fmt.Fprintln(v)

	// ラウンド結果
	fmt.Fprintln(v, centerPad(u.roundResult, innerW))
	fmt.Fprintln(v)

	restartLabel := " 1月から "
	quitLabel := " 終了する "
	if u.gameEndCursor == 0 {
		restartLabel = ansiReverse + restartLabel + ansiReset
	}
	if u.gameEndCursor == 1 {
		quitLabel = ansiReverse + quitLabel + ansiReset
	}
	btnLine := restartLabel + "    " + quitLabel
	btnW := cellWidth(" 1月から ") + 4 + cellWidth(" 終了する ")
	btnPad := max((innerW-btnW)/2, 0)
	fmt.Fprintln(v, strings.Repeat(" ", btnPad)+btnLine)
	fmt.Fprintln(v)
	fmt.Fprintln(v, centerPad("←/→:選択  Enter:決定", innerW))
}

func (u *UI) applyOptions() error {
	// 設定を保存
	u.settings.Rounds = u.optRounds
	u.settings.Difficulty = u.optDifficulty
	u.difficulty = u.optDifficulty
	_ = SaveSettings(u.settingsPath, u.settings)

	// セーブデータ削除（リスタート）
	DeleteSave(u.savePath)

	// ゲームリセット
	u.game = NewGame(u.settings.Rounds)
	u.logLines = nil
	u.addLog(fmt.Sprintf("--- 設定変更: %dラウンド / %s ---", u.settings.Rounds, u.difficulty.Label()))
	u.cursor = 0
	u.showOptions = false
	u.showOptConf = false
	u.phase = PhaseParentDetermination
	u.parentDetStep = 0
	u.parentDetDisplay = nil
	return nil
}

func (u *UI) drawStatus(g *gocui.Gui) {
	v, _ := g.View("status")
	if v == nil {
		return
	}
	v.Clear()
	switch u.phase {
	case PhaseParentDetermination:
		switch u.parentDetStep {
		case 0:
			fmt.Fprint(v, " Enter/Click: 止める")
		case 1:
			fmt.Fprint(v, " CPUの札を引いています...")
		default:
			fmt.Fprint(v, " Enter/Click: 続ける")
		}
	case PhasePlayerSelectHand:
		fmt.Fprint(v, " Up/Down/Click: 選択 | Enter: 決定 | o: オプション | ?: ヘルプ | q: 終了")
	case PhasePlayerSelectField, PhasePlayerSelectFieldDraw:
		fmt.Fprint(v, " Left/Right/Click: 場札選択 | Enter: 決定")
	case PhaseKoiKoi:
		fmt.Fprint(v, " Left/Right/Click: 選択 | Enter: 決定")
	case PhaseCPUKoiKoi:
		fmt.Fprint(v, " Enter/Click: OK")
	case PhaseCPUTurn:
		fmt.Fprint(v, " CPUが考え中...")
	case PhaseRoundEnd:
		fmt.Fprint(v, " Enter/Click: 次の月へ")
	case PhaseGameEnd:
		fmt.Fprint(v, " Left/Right/Click: 選択 | Enter: 決定")
	default:
		fmt.Fprint(v, " Enter: 続行")
	}
}

// ---- ヘルパー ----

var capturedGroups = []struct {
	label string
	typ   CardType
}{
	{"光", Hikari},
	{"種", Tane},
	{"短", Tanzaku},
	{"カス", Kasu},
}

func writeCapturedDetail(v *gocui.View, cards []Card) {
	if len(cards) == 0 {
		fmt.Fprintln(v, " (なし)")
		return
	}
	yakus := CheckYaku(cards)
	if len(yakus) > 0 {
		parts := ""
		for _, y := range yakus {
			parts += fmt.Sprintf(" %s[%s %d文]%s", ansiYellow, y.Name, y.Points, ansiReset)
		}
		fmt.Fprintln(v, parts)
	}
	for _, gr := range capturedGroups {
		filtered := filterByType(cards, gr.typ)
		if len(filtered) == 0 {
			continue
		}
		var names []string
		for _, c := range filtered {
			names = append(names, c.Name)
		}
		fmt.Fprintf(v, " %s(%d): %s\n", gr.label, len(filtered), strings.Join(names, " / "))
	}
}

func countCapturedGroups(cards []Card) int {
	n := 0
	for _, gr := range capturedGroups {
		if len(filterByType(cards, gr.typ)) > 0 {
			n++
		}
	}
	yakus := CheckYaku(cards)
	if len(yakus) > 0 {
		n++
	}
	if n == 0 {
		n = 1
	}
	return n
}

func cardLabel(c Card) string {
	return fmt.Sprintf("[%s:%s]", c.Month.String(), typeSymbols[c.Type])
}

// ---- キーバインド ----

func (u *UI) setKeybindings(g *gocui.Gui) error {
	bindings := []struct {
		key     interface{}
		handler func(*gocui.Gui, *gocui.View) error
	}{
		{gocui.KeyCtrlC, quit},
		{'q', u.handleQuit},
		{'l', u.handleToggleLog},
		{'?', u.handleToggleHelp},
		{'o', u.handleToggleOptions},
		{gocui.KeyArrowUp, u.handleUp},
		{gocui.KeyArrowDown, u.handleDown},
		{gocui.KeyArrowLeft, u.handleLeft},
		{gocui.KeyArrowRight, u.handleRight},
		{gocui.KeyEnter, u.handleEnter},
		{gocui.KeyEsc, u.handleEsc},
		{'y', u.handleYes},
		{'n', u.handleNo},
	}
	for _, b := range bindings {
		if err := g.SetKeybinding("", b.key, gocui.ModNone, b.handler); err != nil {
			return err
		}
	}

	// マウスクリックバインド
	mouseBindings := []struct {
		viewName string
		handler  func(*gocui.Gui, *gocui.View) error
	}{
		{"hand", u.handleHandClick},
		{"field", u.handleFieldClick},
		{"koikoi", u.handleKoiKoiClick},
		{"cpukoikoi", u.handleCPUKoiKoiClick},
		{"parentdet", u.handleParentDetClick},
		{"roundend", u.handleRoundEndClick},
		{"gameend", u.handleGameEndClick},
		{"quitconf", u.handleQuitConfClick},
		{"options", u.handleOptionsClick},
		{"optconf", u.handleOptConfClick},
	}
	for _, b := range mouseBindings {
		if err := g.SetKeybinding(b.viewName, gocui.MouseLeft, gocui.ModNone, b.handler); err != nil {
			return err
		}
	}
	return nil
}

func quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func (u *UI) handleQuit(_ *gocui.Gui, _ *gocui.View) error {
	if u.showOptConf {
		u.showOptConf = false
		return nil
	}
	if u.showOptions {
		u.showOptions = false
		return nil
	}
	if u.showHelp {
		u.showHelp = false
		return nil
	}
	if u.showLog {
		u.showLog = false
		return nil
	}
	if u.showQuitConf {
		// q は「いいえ」として扱う
		u.showQuitConf = false
		return nil
	}
	// 終了確認を表示
	u.showQuitConf = true
	u.quitCursor = 1 // デフォルトは「いいえ」
	return nil
}

func (u *UI) handleEsc(_ *gocui.Gui, _ *gocui.View) error {
	if u.showOptConf {
		u.showOptConf = false
		return nil
	}
	if u.showQuitConf {
		u.showQuitConf = false
		return nil
	}
	if u.showOptions {
		u.showOptions = false
		return nil
	}
	if u.showHelp {
		u.showHelp = false
		return nil
	}
	if u.showLog {
		u.showLog = false
		return nil
	}
	return nil
}

func (u *UI) handleYes(_ *gocui.Gui, _ *gocui.View) error {
	if u.showQuitConf {
		u.autoSave()
		return gocui.ErrQuit
	}
	return nil
}

func (u *UI) handleNo(_ *gocui.Gui, _ *gocui.View) error {
	if u.showQuitConf {
		u.showQuitConf = false
		return nil
	}
	return nil
}

func (u *UI) handleToggleLog(_ *gocui.Gui, _ *gocui.View) error {
	if u.showQuitConf || u.showOptions {
		return nil
	}
	u.showLog = !u.showLog
	return nil
}

func (u *UI) handleToggleHelp(_ *gocui.Gui, _ *gocui.View) error {
	if u.showQuitConf || u.showOptions {
		return nil
	}
	u.showHelp = !u.showHelp
	return nil
}

func (u *UI) handleToggleOptions(_ *gocui.Gui, _ *gocui.View) error {
	if u.showQuitConf {
		return nil
	}
	if u.showOptions {
		u.showOptions = false
		return nil
	}
	u.showOptions = true
	u.showHelp = false
	u.showLog = false
	u.optCursor = 0
	u.optRounds = u.settings.Rounds
	u.optDifficulty = u.settings.Difficulty
	return nil
}

var roundsOptions = []int{3, 6, 12}
var diffOptions = []Difficulty{DifficultyEasy, DifficultyNormal, DifficultyHard}

func (u *UI) handleUp(_ *gocui.Gui, _ *gocui.View) error {
	if u.showOptConf || u.showQuitConf {
		return nil
	}
	if u.showOptions {
		if u.optCursor > 0 {
			u.optCursor--
		}
		return nil
	}
	if u.showLog || u.showHelp {
		return nil
	}
	if u.phase == PhasePlayerSelectHand && u.cursor > 0 {
		u.cursor--
	}
	return nil
}

func (u *UI) handleDown(_ *gocui.Gui, _ *gocui.View) error {
	if u.showOptConf || u.showQuitConf {
		return nil
	}
	if u.showOptions {
		if u.optCursor < 2 {
			u.optCursor++
		}
		return nil
	}
	if u.showLog || u.showHelp {
		return nil
	}
	if u.phase == PhasePlayerSelectHand && u.cursor < len(u.game.PlayerHand)-1 {
		u.cursor++
	}
	return nil
}

func (u *UI) handleLeft(_ *gocui.Gui, _ *gocui.View) error {
	if u.showOptConf {
		if u.optConfCursor > 0 {
			u.optConfCursor--
		}
		return nil
	}
	if u.showOptions {
		switch u.optCursor {
		case 0: // ラウンド数
			for i, r := range roundsOptions {
				if r == u.optRounds && i > 0 {
					u.optRounds = roundsOptions[i-1]
					break
				}
			}
		case 1: // 難易度
			for i, d := range diffOptions {
				if d == u.optDifficulty && i > 0 {
					u.optDifficulty = diffOptions[i-1]
					break
				}
			}
		case 2: // ボタン行
			if u.optBtnCursor > 0 {
				u.optBtnCursor--
			}
		}
		return nil
	}
	if u.showQuitConf {
		if u.quitCursor > 0 {
			u.quitCursor--
		}
		return nil
	}
	if u.showLog || u.showHelp {
		return nil
	}
	switch u.phase {
	case PhasePlayerSelectField, PhasePlayerSelectFieldDraw:
		if u.fieldCursor > 0 {
			u.fieldCursor--
		}
	case PhaseKoiKoi:
		if u.koikoiCursor > 0 {
			u.koikoiCursor--
		}
	case PhaseGameEnd:
		if u.gameEndCursor > 0 {
			u.gameEndCursor--
		}
	}
	return nil
}

func (u *UI) handleRight(_ *gocui.Gui, _ *gocui.View) error {
	if u.showOptConf {
		if u.optConfCursor < 1 {
			u.optConfCursor++
		}
		return nil
	}
	if u.showOptions {
		switch u.optCursor {
		case 0: // ラウンド数
			for i, r := range roundsOptions {
				if r == u.optRounds && i < len(roundsOptions)-1 {
					u.optRounds = roundsOptions[i+1]
					break
				}
			}
		case 1: // 難易度
			for i, d := range diffOptions {
				if d == u.optDifficulty && i < len(diffOptions)-1 {
					u.optDifficulty = diffOptions[i+1]
					break
				}
			}
		case 2: // ボタン行
			if u.optBtnCursor < 1 {
				u.optBtnCursor++
			}
		}
		return nil
	}
	if u.showQuitConf {
		if u.quitCursor < 1 {
			u.quitCursor++
		}
		return nil
	}
	if u.showLog || u.showHelp {
		return nil
	}
	switch u.phase {
	case PhasePlayerSelectField, PhasePlayerSelectFieldDraw:
		if u.fieldCursor < len(u.matchingCards)-1 {
			u.fieldCursor++
		}
	case PhaseKoiKoi:
		if u.koikoiCursor < 1 {
			u.koikoiCursor++
		}
	case PhaseGameEnd:
		if u.gameEndCursor < 1 {
			u.gameEndCursor++
		}
	}
	return nil
}

func (u *UI) handleEnter(g *gocui.Gui, _ *gocui.View) error {
	if u.showQuitConf {
		if u.quitCursor == 0 {
			// はい → 保存して終了
			u.autoSave()
			return gocui.ErrQuit
		}
		u.showQuitConf = false
		return nil
	}
	if u.showOptConf {
		if u.optConfCursor == 0 {
			// はい → 適用してリスタート
			return u.applyOptions()
		}
		// いいえ → 確認ダイアログだけ閉じる
		u.showOptConf = false
		return nil
	}
	if u.showOptions {
		if u.optCursor == 2 {
			if u.optBtnCursor == 0 {
				// キャンセル → 保存せず閉じる
				u.showOptions = false
			} else {
				// 適用 → 確認ダイアログ表示
				u.showOptConf = true
				u.optConfCursor = 1 // デフォルトは「いいえ」
			}
		}
		return nil
	}
	if u.showLog {
		u.showLog = false
		return nil
	}
	if u.showHelp {
		u.showHelp = false
		return nil
	}
	switch u.phase {
	case PhaseParentDetermination:
		return u.onParentDeterminationOK(g)
	case PhasePlayerSelectHand:
		return u.onSelectHand(g)
	case PhasePlayerSelectField:
		return u.onSelectFieldForHand(g)
	case PhasePlayerDrawResult:
		return u.onDrawResult(g)
	case PhasePlayerSelectFieldDraw:
		return u.onSelectFieldForDraw(g)
	case PhaseKoiKoi:
		return u.onKoiKoiDecision(g)
	case PhaseCPUKoiKoi:
		return u.onCPUKoiKoiOK(g)
	case PhaseRoundEnd:
		return u.onNextRound(g)
	case PhaseGameEnd:
		return u.onGameEndDecision(g)
	}
	return nil
}

// ---- フェーズ遷移 ----

func (u *UI) onParentDeterminationOK(g *gocui.Gui) error {
	switch u.parentDetStep {
	case 0:
		// プレイヤーの札確定 → CPUの札表示へ
		u.parentDetDisplay = u.game.PlayerDrawnCard
		u.parentDetStep = 1
		u.showCPUCardWithDelay(g)
		return nil
	case 1:
		// CPUの札表示中は何もしない
		return nil
	default:
		// 結果表示後 → ゲーム開始
		u.logParentDetermination()
		u.game.StartRound()
		u.addLog(fmt.Sprintf("--- %s 開始 ---", u.roundName()))
		u.cursor = 0
		if u.game.IsPlayerTurn {
			u.phase = PhasePlayerSelectHand
		} else {
			u.phase = PhaseCPUTurn
		}
		return nil
	}
}

func (u *UI) onSelectHand(g *gocui.Gui) error {
	if len(u.game.PlayerHand) == 0 {
		return nil
	}
	if u.cursor >= len(u.game.PlayerHand) {
		u.cursor = len(u.game.PlayerHand) - 1
	}
	handCard := u.game.PlayerHand[u.cursor]
	u.playedCard = handCard
	matches := u.game.MatchingFieldCards(handCard)

	if len(matches) == 2 {
		u.matchingCards = matches
		u.fieldCursor = 0
		u.phase = PhasePlayerSelectField
		u.addLog(fmt.Sprintf("%sを出す - 取る札を選んでください", handCard.Name))
	} else {
		captured := u.game.PlayCard(handCard, nil, true)
		if len(captured) > 0 {
			u.addLog(fmt.Sprintf("%s -> 獲得: %s", handCard.Name, capturedNames(captured)))
		} else {
			u.addLog(fmt.Sprintf("%s を場に出しました", handCard.Name))
		}
		u.cursor = 0
		u.doPlayerDraw(g)
	}
	return nil
}

func (u *UI) onSelectFieldForHand(g *gocui.Gui) error {
	chosen := u.matchingCards[u.fieldCursor]
	captured := u.game.PlayCard(u.playedCard, &chosen, true)
	u.addLog(fmt.Sprintf("獲得: %s", capturedNames(captured)))
	u.cursor = 0
	u.doPlayerDraw(g)
	return nil
}

func (u *UI) doPlayerDraw(g *gocui.Gui) {
	drawn, ok := u.game.DrawFromDeck()
	if !ok {
		u.checkPlayerYaku(g)
		return
	}
	u.drawnCard = drawn
	matches := u.game.MatchingFieldCards(drawn)

	if len(matches) == 2 {
		u.matchingCards = matches
		u.fieldCursor = 0
		u.phase = PhasePlayerSelectFieldDraw
		u.addLog(fmt.Sprintf("山札: %s - 取る札を選んでください", drawn.Name))
	} else {
		captured := u.game.PlayDrawnCard(drawn, nil, true)
		if len(captured) > 0 {
			u.addLog(fmt.Sprintf("山札: %s -> 獲得: %s", drawn.Name, capturedNames(captured)))
		} else {
			u.addLog(fmt.Sprintf("山札: %s -> 場に置きました", drawn.Name))
		}
		u.checkPlayerYaku(g)
	}
}

func (u *UI) onDrawResult(g *gocui.Gui) error {
	u.checkPlayerYaku(g)
	return nil
}

func (u *UI) onSelectFieldForDraw(g *gocui.Gui) error {
	chosen := u.matchingCards[u.fieldCursor]
	captured := u.game.PlayDrawnCard(u.drawnCard, &chosen, true)
	u.addLog(fmt.Sprintf("獲得: %s", capturedNames(captured)))
	u.checkPlayerYaku(g)
	return nil
}

func (u *UI) checkPlayerYaku(g *gocui.Gui) {
	newYaku := u.game.CheckNewYaku(true)
	if len(newYaku) > 0 {
		for _, y := range newYaku {
			u.addLog(fmt.Sprintf("* %s (%d文) 成立！", y.Name, y.Points))
		}
		// 手札が0枚ならこいこいできないので自動的に勝負
		if len(u.game.PlayerHand) == 0 {
			finalScore := u.game.CalcFinalScore(true)
			u.game.PlayerScore += finalScore
			u.addLog(fmt.Sprintf("手札なし → 自動的に勝負！ %d文獲得！", finalScore))
			u.game.NextParentIsPlayer = true
			u.finishRound(g, fmt.Sprintf("あなたの勝ち！ %d文獲得", finalScore))
			return
		}
		u.newYaku = newYaku
		u.koikoiCursor = 0
		u.phase = PhaseKoiKoi
	} else {
		u.endPlayerTurn(g)
	}
}

func (u *UI) onKoiKoiDecision(g *gocui.Gui) error {
	if u.koikoiCursor == 0 {
		u.addLog("こいこい！")
		u.game.PlayerKoiKoi = true
		u.game.UpdatePrevYaku(true)
		u.endPlayerTurn(g)
	} else {
		finalScore := u.game.CalcFinalScore(true)
		u.game.PlayerScore += finalScore
		u.addLog(fmt.Sprintf("勝負！ %d文獲得！", finalScore))
		u.game.NextParentIsPlayer = true
		u.finishRound(g, fmt.Sprintf("あなたの勝ち！ %d文獲得", finalScore))
	}
	return nil
}

func (u *UI) endPlayerTurn(g *gocui.Gui) {
	u.game.IsPlayerTurn = false
	if u.game.IsRoundOver() {
		u.addLog("手札が尽きました。引き分けです。")
		u.finishRound(g, "引き分け")
		return
	}
	u.phase = PhaseCPUTurn
	u.autoSave()
}

func (u *UI) finishRound(_ *gocui.Gui, result string) {
	u.roundResult = result
	u.game.Round++
	if u.game.Round > u.game.MaxRounds {
		u.game.Round = u.game.MaxRounds // ヘッダー表示用に戻す
		u.phase = PhaseGameEnd
		u.gameEndCursor = 0 // デフォルト: 1月から
		switch {
		case u.game.PlayerScore > u.game.CPUScore:
			u.gameResult = msgPlayerWin
		case u.game.PlayerScore < u.game.CPUScore:
			u.gameResult = msgCPUWin
		default:
			u.gameResult = msgDraw
		}
		u.addLog("--- ゲーム終了 ---")
		// ゲーム終了時はセーブデータ削除
		DeleteSave(u.savePath)
	} else {
		u.phase = PhaseRoundEnd
		m := Month((u.game.Round - 1) % 12)
		u.addLog(fmt.Sprintf("--- 次は%s ---", m.OldName()))
		u.autoSave()
	}
}

func (u *UI) onNextRound(_ *gocui.Gui) error {
	u.game.StartRound()
	u.cursor = 0
	u.addLog(fmt.Sprintf("--- %s 開始 ---", u.roundName()))
	if u.game.IsPlayerTurn {
		u.phase = PhasePlayerSelectHand
	} else {
		u.phase = PhaseCPUTurn
	}
	return nil
}

func (u *UI) onGameEndDecision(_ *gocui.Gui) error {
	if u.gameEndCursor == 1 {
		// 終了する
		return gocui.ErrQuit
	}
	// 1月から再スタート
	u.game = NewGame(u.game.MaxRounds)
	u.addLog("--- 1月から再スタート ---")
	u.cursor = 0
	u.phase = PhaseParentDetermination
	u.parentDetStep = 0
	u.parentDetDisplay = nil
	return nil
}

// ---- 親決めアニメーション ----

var parentDetAnimRunning atomic.Bool

func (u *UI) startParentDetAnimation(g *gocui.Gui) {
	if g == nil {
		// テスト時は同期的に1回だけ実行
		u.executeParentDetSpin(nil, 0)
		return
	}
	if !parentDetAnimRunning.CompareAndSwap(false, true) {
		return
	}
	go func() {
		defer parentDetAnimRunning.Store(false)
		u.executeParentDetSpin(g, 80*time.Millisecond)
	}()
}

// executeParentDetSpin 親決めアニメーションのスピンロジック（テスト可能）
func (u *UI) executeParentDetSpin(g *gocui.Gui, spinDelay time.Duration) {
	cardIndex := 0
	for u.parentDetStep == 0 && u.phase == PhaseParentDetermination {
		card := AllCards[cardIndex%len(AllCards)]
		u.parentDetDisplay = &card
		if g != nil && u.gui != nil {
			u.gui.Update(func(_ *gocui.Gui) error { return nil })
		}
		if spinDelay > 0 {
			time.Sleep(spinDelay)
		} else {
			// テスト時は1回だけ実行して終了
			break
		}
		cardIndex++
	}
}

func (u *UI) showCPUCardWithDelay(g *gocui.Gui) {
	if g == nil {
		// テスト時はアニメーションをスキップして即座に結果表示
		u.executeCPUCardReveal(nil, 0, 0)
		return
	}
	go func() {
		u.executeCPUCardReveal(g, 100*time.Millisecond, 500*time.Millisecond)
	}()
}

// executeCPUCardReveal CPUの札を公開するロジック（テスト可能）
func (u *UI) executeCPUCardReveal(g *gocui.Gui, spinDelay, finalDelay time.Duration) {
	// CPUの札をルーレット風に数回表示
	for range 10 {
		card := AllCards[cryptoIntn(len(AllCards))]
		u.parentDetDisplay = &card
		if g != nil && u.gui != nil {
			u.gui.Update(func(_ *gocui.Gui) error { return nil })
		}
		if spinDelay > 0 {
			time.Sleep(spinDelay)
		}
	}
	// 最終的なCPUの札を表示
	u.parentDetDisplay = u.game.CPUDrawnCard
	if g != nil && u.gui != nil {
		u.gui.Update(func(_ *gocui.Gui) error { return nil })
	}
	if finalDelay > 0 {
		time.Sleep(finalDelay)
	}
	// 結果表示へ
	u.parentDetStep = 2
	if g != nil && u.gui != nil {
		u.gui.Update(func(_ *gocui.Gui) error { return nil })
	}
}

// ---- CPUターン ----

var cpuTurnRunning atomic.Bool

func (u *UI) doCPUTurn(g *gocui.Gui) {
	if !cpuTurnRunning.CompareAndSwap(false, true) {
		return
	}
	defer cpuTurnRunning.Store(false)

	time.Sleep(800 * time.Millisecond)

	g.Update(func(g *gocui.Gui) error {
		return u.executeCPUTurn(g)
	})
}

func (u *UI) onCPUKoiKoiOK(_ *gocui.Gui) error {
	u.cpuKoiKoiYaku = nil
	u.game.IsPlayerTurn = true
	if u.game.IsRoundOver() {
		u.addLog("手札が尽きました。引き分けです。")
		u.finishRound(nil, "引き分け")
		return nil
	}
	u.cursor = 0
	u.phase = PhasePlayerSelectHand
	return nil
}

func (u *UI) executeCPUTurn(g *gocui.Gui) error {
	if len(u.game.CPUHand) == 0 {
		u.game.IsPlayerTurn = true
		u.phase = PhasePlayerSelectHand
		return nil
	}
	handCard, fieldChoice := CPUChooseHandCard(u.game, u.difficulty)
	captured := u.game.PlayCard(handCard, fieldChoice, false)
	if len(captured) > 0 {
		u.addLog(fmt.Sprintf("CPU: %s -> 獲得: %s", handCard.Name, capturedNames(captured)))
	} else {
		u.addLog(fmt.Sprintf("CPU: %sを場に出した", handCard.Name))
	}

	drawn, ok := u.game.DrawFromDeck()
	if ok {
		drawnMatches := u.game.MatchingFieldCards(drawn)
		drawnFieldChoice := CPUChooseFieldCard(drawnMatches)
		drawnCaptured := u.game.PlayDrawnCard(drawn, drawnFieldChoice, false)
		if len(drawnCaptured) > 0 {
			u.addLog(fmt.Sprintf("CPU山札: %s -> 獲得: %s", drawn.Name, capturedNames(drawnCaptured)))
		} else {
			u.addLog(fmt.Sprintf("CPU山札: %s -> 場へ", drawn.Name))
		}
	}

	newYaku := u.game.CheckNewYaku(false)
	if len(newYaku) > 0 {
		for _, y := range newYaku {
			u.addLog(fmt.Sprintf("CPU * %s (%d文)", y.Name, y.Points))
		}
		if len(u.game.CPUHand) > 0 && CPUDecideKoiKoi(u.game, CheckYaku(u.game.CPUCaptured), u.difficulty) {
			u.addLog("CPU: こいこい！")
			u.game.CPUKoiKoi = true
			u.game.UpdatePrevYaku(false)
			u.cpuKoiKoiYaku = newYaku
			u.phase = PhaseCPUKoiKoi
			return nil
		} else {
			finalScore := u.game.CalcFinalScore(false)
			u.game.CPUScore += finalScore
			u.addLog(fmt.Sprintf("CPU: 勝負！ %d文獲得", finalScore))
			u.game.NextParentIsPlayer = false
			u.finishRound(g, fmt.Sprintf("CPUの勝ち！ %d文獲得", finalScore))
			return nil
		}
	}

	u.game.IsPlayerTurn = true
	if u.game.IsRoundOver() {
		u.addLog("手札が尽きました。引き分けです。")
		u.finishRound(g, "引き分け")
		return nil
	}
	u.cursor = 0
	u.phase = PhasePlayerSelectHand
	return nil
}

func capturedNames(cards []Card) string {
	names := make([]string, 0, len(cards))
	for _, c := range cards {
		names = append(names, c.Name)
	}
	return strings.Join(names, ", ")
}

// ---- マウスクリックハンドラ ----

func (u *UI) handleHandClick(g *gocui.Gui, v *gocui.View) error {
	if u.showQuitConf || u.showOptions || u.showOptConf || u.showLog || u.showHelp {
		return nil
	}
	if u.phase != PhasePlayerSelectHand {
		return nil
	}
	_, cy := v.Cursor()
	_, oy := v.Origin()
	// 行番号はビュー内の相対位置（空行を考慮: 1行目は空行）
	clickedLine := cy + oy - 1 // 空行分を引く
	if clickedLine >= 0 && clickedLine < len(u.game.PlayerHand) {
		u.cursor = clickedLine
		return u.onSelectHand(g)
	}
	return nil
}

func (u *UI) handleFieldClick(g *gocui.Gui, v *gocui.View) error {
	if u.showQuitConf || u.showOptions || u.showOptConf || u.showLog || u.showHelp {
		return nil
	}
	if u.phase != PhasePlayerSelectField && u.phase != PhasePlayerSelectFieldDraw {
		return nil
	}
	cx, _ := v.Cursor()
	ox, _ := v.Origin()
	clickX := cx + ox - 1 // 先頭スペース分を引く

	// 各場札の位置を計算して、クリック位置からどの札かを特定
	pos := 0
	clickedFieldIdx := -1
	for i, c := range u.game.Field {
		label := cardLabel(c)
		labelW := cellWidth(label)
		if clickX >= pos && clickX < pos+labelW {
			clickedFieldIdx = i
			break
		}
		pos += labelW + 1 // 札の幅 + スペース
	}

	if clickedFieldIdx < 0 {
		return nil
	}

	// クリックした場札が matchingCards に含まれているか確認
	clickedCard := u.game.Field[clickedFieldIdx]
	for mi, m := range u.matchingCards {
		if m.ID == clickedCard.ID {
			u.fieldCursor = mi
			if u.phase == PhasePlayerSelectField {
				return u.onSelectFieldForHand(g)
			}
			return u.onSelectFieldForDraw(g)
		}
	}
	return nil
}

func (u *UI) handleKoiKoiClick(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	_, oy := v.Origin()
	clickedLine := cy + oy

	// ボタン行をクリックしたか確認
	yakuCount := len(u.newYaku)
	buttonLine := 7 + yakuCount // header + yakus + total + spacing

	if clickedLine == buttonLine {
		cx, _ := v.Cursor()
		ox, _ := v.Origin()
		clickX := cx + ox
		innerW := 42
		btnW := cellWidth(" こいこい（続行） ") + 2 + cellWidth(" 勝負（得点確定） ")
		btnPad := max((innerW-btnW)/2, 0)

		koikoiEnd := btnPad + cellWidth(" こいこい（続行） ")
		if clickX >= btnPad && clickX < koikoiEnd {
			u.koikoiCursor = 0
			return u.onKoiKoiDecision(g)
		}
		shoubuStart := koikoiEnd + 2
		if clickX >= shoubuStart {
			u.koikoiCursor = 1
			return u.onKoiKoiDecision(g)
		}
	}
	return nil
}

func (u *UI) handleParentDetClick(g *gocui.Gui, _ *gocui.View) error {
	return u.onParentDeterminationOK(g)
}

func (u *UI) handleCPUKoiKoiClick(g *gocui.Gui, _ *gocui.View) error {
	return u.onCPUKoiKoiOK(g)
}

func (u *UI) handleRoundEndClick(g *gocui.Gui, _ *gocui.View) error {
	return u.onNextRound(g)
}

func (u *UI) handleGameEndClick(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	_, oy := v.Origin()
	clickedLine := cy + oy

	// ボタン行は7行目
	if clickedLine == 7 {
		cx, _ := v.Cursor()
		ox, _ := v.Origin()
		clickX := cx + ox
		innerW := 42
		btnW := cellWidth(" 1月から ") + 4 + cellWidth(" 終了する ")
		btnPad := max((innerW-btnW)/2, 0)

		restartEnd := btnPad + cellWidth(" 1月から ")
		if clickX >= btnPad && clickX < restartEnd {
			u.gameEndCursor = 0
			return u.onGameEndDecision(g)
		}
		quitStart := restartEnd + 4
		if clickX >= quitStart {
			u.gameEndCursor = 1
			return u.onGameEndDecision(g)
		}
	}
	return nil
}

func (u *UI) handleQuitConfClick(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	_, oy := v.Origin()
	clickedLine := cy + oy

	// ボタン行は5行目
	if clickedLine == 5 {
		cx, _ := v.Cursor()
		ox, _ := v.Origin()
		clickX := cx + ox
		innerW := 44
		btnW := cellWidth(" はい ") + 4 + cellWidth(" いいえ ")
		btnPad := max((innerW-btnW)/2, 0)

		yesEnd := btnPad + cellWidth(" はい ")
		if clickX >= btnPad && clickX < yesEnd {
			u.quitCursor = 0
			return u.handleEnter(g, nil)
		}
		noStart := yesEnd + 4
		if clickX >= noStart {
			u.quitCursor = 1
			return u.handleEnter(g, nil)
		}
	}
	return nil
}

func (u *UI) handleOptionsClick(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	_, oy := v.Origin()
	clickedLine := cy + oy

	switch clickedLine {
	case 1: // ラウンド数行
		u.optCursor = 0
		cx, _ := v.Cursor()
		ox, _ := v.Origin()
		clickX := cx + ox
		// ◀ または ▶ をクリックで値変更
		if clickX < 21 {
			// 左側 → 減らす
			for i, r := range roundsOptions {
				if r == u.optRounds && i > 0 {
					u.optRounds = roundsOptions[i-1]
					break
				}
			}
		} else {
			// 右側 → 増やす
			for i, r := range roundsOptions {
				if r == u.optRounds && i < len(roundsOptions)-1 {
					u.optRounds = roundsOptions[i+1]
					break
				}
			}
		}
	case 3: // 難易度行
		u.optCursor = 1
		cx, _ := v.Cursor()
		ox, _ := v.Origin()
		clickX := cx + ox
		if clickX < 21 {
			for i, d := range diffOptions {
				if d == u.optDifficulty && i > 0 {
					u.optDifficulty = diffOptions[i-1]
					break
				}
			}
		} else {
			for i, d := range diffOptions {
				if d == u.optDifficulty && i < len(diffOptions)-1 {
					u.optDifficulty = diffOptions[i+1]
					break
				}
			}
		}
	case 5: // ボタン行
		u.optCursor = 2
		cx, _ := v.Cursor()
		ox, _ := v.Origin()
		clickX := cx + ox
		innerW := 42
		cancelW := cellWidth(labelCancel)
		applyW := cellWidth(labelApply)
		btnW := cancelW + 4 + applyW
		btnPad := max((innerW-btnW)/2, 0)

		cancelEnd := btnPad + cancelW
		if clickX >= btnPad && clickX < cancelEnd {
			u.optBtnCursor = 0
			return u.handleEnter(g, nil)
		}
		applyStart := cancelEnd + 4
		if clickX >= applyStart {
			u.optBtnCursor = 1
			return u.handleEnter(g, nil)
		}
	}
	return nil
}

func (u *UI) handleOptConfClick(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	_, oy := v.Origin()
	clickedLine := cy + oy

	// ボタン行は5行目
	if clickedLine == 5 {
		cx, _ := v.Cursor()
		ox, _ := v.Origin()
		clickX := cx + ox
		innerW := 44
		btnW := cellWidth(" はい ") + 4 + cellWidth(" いいえ ")
		btnPad := max((innerW-btnW)/2, 0)

		yesEnd := btnPad + cellWidth(" はい ")
		if clickX >= btnPad && clickX < yesEnd {
			u.optConfCursor = 0
			return u.handleEnter(g, nil)
		}
		noStart := yesEnd + 4
		if clickX >= noStart {
			u.optConfCursor = 1
			return u.handleEnter(g, nil)
		}
	}
	return nil
}
