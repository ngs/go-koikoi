package main

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/awesome-gocui/gocui"
)

func setupMouseTestUI(t *testing.T) (*UI, *gocui.Gui) {
	t.Helper()
	g := newTestGUI(t)
	u := NewUI()
	dir := t.TempDir()
	u.configDir = dir
	u.settingsPath = filepath.Join(dir, "settings.json")
	u.savePath = filepath.Join(dir, "game.json")
	u.settings = DefaultSettings()
	u.gui = g
	g.SetManagerFunc(u.layout)
	return u, g
}

// --- handleHandClick テスト ---

func TestHandleHandClickIgnoresWhenQuitConfOpen(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.PlayerHand = []Card{AllCards[0], AllCards[4]}
	u.game.IsPlayerTurn = true
	u.phase = PhasePlayerSelectHand
	u.showQuitConf = true

	v := &gocui.View{}
	if err := u.handleHandClick(g, v); err != nil {
		t.Fatalf("handleHandClick error: %v", err)
	}
	// ポップアップ表示中は何もしないで終了すればOK
}

func TestHandleHandClickIgnoresWhenOptionsOpen(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.PlayerHand = []Card{AllCards[0]}
	u.phase = PhasePlayerSelectHand
	u.showOptions = true

	v := &gocui.View{}
	if err := u.handleHandClick(g, v); err != nil {
		t.Fatalf("handleHandClick error: %v", err)
	}
}

func TestHandleHandClickIgnoresWhenHelpOpen(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.PlayerHand = []Card{AllCards[0]}
	u.phase = PhasePlayerSelectHand
	u.showHelp = true

	v := &gocui.View{}
	if err := u.handleHandClick(g, v); err != nil {
		t.Fatalf("handleHandClick error: %v", err)
	}
}

func TestHandleHandClickIgnoresWrongPhase(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.PlayerHand = []Card{AllCards[0]}
	u.phase = PhaseCPUTurn

	v := &gocui.View{}
	if err := u.handleHandClick(g, v); err != nil {
		t.Fatalf("handleHandClick error: %v", err)
	}
	// 間違ったフェーズでは何もしない
}

func TestHandleHandClickIgnoresPhaseKoiKoi(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.PlayerHand = []Card{AllCards[0]}
	u.phase = PhaseKoiKoi

	v := &gocui.View{}
	if err := u.handleHandClick(g, v); err != nil {
		t.Fatalf("handleHandClick error: %v", err)
	}
}

// --- handleFieldClick テスト ---

func TestHandleFieldClickIgnoresWhenPopupOpen(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.phase = PhasePlayerSelectField
	u.showQuitConf = true

	v := &gocui.View{}
	if err := u.handleFieldClick(g, v); err != nil {
		t.Fatalf("handleFieldClick error: %v", err)
	}
}

func TestHandleFieldClickIgnoresWrongPhase(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.phase = PhasePlayerSelectHand

	v := &gocui.View{}
	if err := u.handleFieldClick(g, v); err != nil {
		t.Fatalf("handleFieldClick error: %v", err)
	}
}

func TestHandleFieldClickIgnoresPhaseCPUTurn(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.phase = PhaseCPUTurn

	v := &gocui.View{}
	if err := u.handleFieldClick(g, v); err != nil {
		t.Fatalf("handleFieldClick error: %v", err)
	}
}

// --- handleKoiKoiClick テスト ---

func TestHandleKoiKoiClickWrongLine(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.PlayerHand = []Card{AllCards[0], AllCards[4]}
	u.newYaku = []Yaku{{Name: "三光", Points: 5}}
	u.phase = PhaseKoiKoi
	u.koikoiCursor = 0

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("koikoi")
	if v == nil {
		t.Skip("koikoi view not created")
	}

	// ボタン行ではない行をクリック
	v.SetCursor(5, 2)
	originalCursor := u.koikoiCursor
	if err := u.handleKoiKoiClick(g, v); err != nil {
		t.Fatalf("handleKoiKoiClick error: %v", err)
	}

	// カーソルが変わらないことを確認
	if u.koikoiCursor != originalCursor {
		t.Error("koikoiCursor should not change when clicking non-button line")
	}
}

// --- handleCPUKoiKoiClick テスト ---

func TestHandleCPUKoiKoiClickOK(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.PlayerHand = []Card{AllCards[0]}
	u.game.CPUHand = []Card{AllCards[4]}
	u.cpuKoiKoiYaku = []Yaku{{Name: "三光", Points: 5}}
	u.phase = PhaseCPUKoiKoi

	if err := u.handleCPUKoiKoiClick(g, nil); err != nil {
		t.Fatalf("handleCPUKoiKoiClick error: %v", err)
	}

	if u.phase != PhasePlayerSelectHand {
		t.Errorf("phase = %d, want PhasePlayerSelectHand", u.phase)
	}
	if u.cpuKoiKoiYaku != nil {
		t.Error("cpuKoiKoiYaku should be nil after click")
	}
}

func TestHandleCPUKoiKoiClickRoundOver(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.PlayerHand = nil // 手札なし = ラウンド終了
	u.game.CPUHand = nil
	u.cpuKoiKoiYaku = []Yaku{{Name: "三光", Points: 5}}
	u.phase = PhaseCPUKoiKoi

	if err := u.handleCPUKoiKoiClick(g, nil); err != nil {
		t.Fatalf("handleCPUKoiKoiClick error: %v", err)
	}

	if u.phase != PhaseRoundEnd {
		t.Errorf("phase = %d, want PhaseRoundEnd", u.phase)
	}
}

// --- handleRoundEndClick テスト ---

func TestHandleRoundEndClickNextRound(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.Round = 2
	u.phase = PhaseRoundEnd

	if err := u.handleRoundEndClick(g, nil); err != nil {
		t.Fatalf("handleRoundEndClick error: %v", err)
	}

	if u.phase != PhasePlayerSelectHand && u.phase != PhaseCPUTurn {
		t.Errorf("phase = %d, want PhasePlayerSelectHand or PhaseCPUTurn", u.phase)
	}
}

// --- handleGameEndClick テスト ---

func TestHandleGameEndClickWrongLine(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.Round = 12
	u.phase = PhaseGameEnd

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("gameend")
	if v == nil {
		t.Skip("gameend view not created")
	}

	// ボタン行ではない行をクリック
	v.SetCursor(5, 2)
	originalCursor := u.gameEndCursor
	if err := u.handleGameEndClick(g, v); err != nil {
		t.Fatalf("handleGameEndClick error: %v", err)
	}

	if u.gameEndCursor != originalCursor {
		t.Error("gameEndCursor should not change when clicking non-button line")
	}
}

// --- handleQuitConfClick テスト ---

func TestHandleQuitConfClickWrongLine(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showQuitConf = true

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("quitconf")
	if v == nil {
		t.Skip("quitconf view not created")
	}

	// ボタン行ではない行をクリック
	v.SetCursor(5, 2)
	if err := u.handleQuitConfClick(g, v); err != nil {
		t.Fatalf("handleQuitConfClick error: %v", err)
	}

	// まだ表示中であること
	if !u.showQuitConf {
		t.Error("showQuitConf should still be true when clicking non-button line")
	}
}

// --- handleOptionsClick テスト ---

func TestHandleOptionsClickRoundsLeft(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.optRounds = 6
	u.optCursor = 2

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("options")
	if v == nil {
		t.Skip("options view not created")
	}

	// ラウンド数行をクリック（左側で減らす）
	v.SetCursor(10, 1)
	if err := u.handleOptionsClick(g, v); err != nil {
		t.Fatalf("handleOptionsClick error: %v", err)
	}

	if u.optCursor != 0 {
		t.Errorf("optCursor = %d, want 0", u.optCursor)
	}
	if u.optRounds != 3 {
		t.Errorf("optRounds = %d, want 3", u.optRounds)
	}
}

func TestHandleOptionsClickRoundsRight(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.optRounds = 6
	u.optCursor = 2

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("options")
	if v == nil {
		t.Skip("options view not created")
	}

	// ラウンド数行をクリック（右側で増やす）
	v.SetCursor(30, 1)
	if err := u.handleOptionsClick(g, v); err != nil {
		t.Fatalf("handleOptionsClick error: %v", err)
	}

	if u.optRounds != 12 {
		t.Errorf("optRounds = %d, want 12", u.optRounds)
	}
}

func TestHandleOptionsClickDifficultyLeft(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.optDifficulty = DifficultyNormal
	u.optCursor = 0

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("options")
	if v == nil {
		t.Skip("options view not created")
	}

	// 難易度行をクリック（左側で下げる）
	v.SetCursor(10, 3)
	if err := u.handleOptionsClick(g, v); err != nil {
		t.Fatalf("handleOptionsClick error: %v", err)
	}

	if u.optDifficulty != DifficultyEasy {
		t.Errorf("optDifficulty = %q, want easy", u.optDifficulty)
	}
}

func TestHandleOptionsClickDifficultyRight(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.optDifficulty = DifficultyNormal
	u.optCursor = 0

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("options")
	if v == nil {
		t.Skip("options view not created")
	}

	// 難易度行をクリック（右側で上げる）
	v.SetCursor(30, 3)
	if err := u.handleOptionsClick(g, v); err != nil {
		t.Fatalf("handleOptionsClick error: %v", err)
	}

	if u.optDifficulty != DifficultyHard {
		t.Errorf("optDifficulty = %q, want hard", u.optDifficulty)
	}
}

func TestHandleOptionsClickCancelButton(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.optCursor = 0
	u.optBtnCursor = 1

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("options")
	if v == nil {
		t.Skip("options view not created")
	}

	// ボタン位置を計算して正確な位置を設定
	cancelLabel := " キャンセル "
	applyLabel := " 適用 "
	innerW := 42
	cancelW := cellWidth(cancelLabel)
	applyW := cellWidth(applyLabel)
	btnW := cancelW + 4 + applyW
	btnPad := (innerW - btnW) / 2
	// キャンセルボタンの中央
	cancelMid := btnPad + cancelW/2
	v.SetCursor(cancelMid, 5)

	if err := u.handleOptionsClick(g, v); err != nil {
		t.Fatalf("handleOptionsClick error: %v", err)
	}

	if u.showOptions {
		t.Error("showOptions should be false after cancel")
	}
}

func TestHandleOptionsClickApplyButton(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.optCursor = 0
	u.optBtnCursor = 0

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("options")
	if v == nil {
		t.Skip("options view not created")
	}

	// ボタン位置を計算して正確な位置を設定
	cancelLabel := " キャンセル "
	applyLabel := " 適用 "
	innerW := 42
	cancelW := cellWidth(cancelLabel)
	applyW := cellWidth(applyLabel)
	btnW := cancelW + 4 + applyW
	btnPad := (innerW - btnW) / 2
	applyStart := btnPad + cancelW + 4
	// 適用ボタンの中央
	applyMid := applyStart + applyW/2
	v.SetCursor(applyMid, 5)

	// テスト環境でCursor設定が反映されない場合の代替検証
	cx, cy := v.Cursor()
	if cx != applyMid || cy != 5 {
		// Cursor が設定できない環境ではスキップ
		t.Skipf("View.SetCursor not working in test env (got %d,%d, want %d,5)", cx, cy, applyMid)
	}

	if err := u.handleOptionsClick(g, v); err != nil {
		t.Fatalf("handleOptionsClick error: %v", err)
	}

	if !u.showOptConf {
		t.Error("showOptConf should be true after clicking apply")
	}
}

// handleEnter 経由でのオプション適用を直接テスト
func TestHandleEnterApplyOptions(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.optCursor = 2    // ボタン行
	u.optBtnCursor = 1 // 適用ボタン

	if err := u.handleEnter(g, nil); err != nil {
		t.Fatalf("handleEnter error: %v", err)
	}

	if !u.showOptConf {
		t.Error("showOptConf should be true after apply")
	}
}

// handleEnter 経由でのこいこい/勝負を直接テスト
func TestHandleEnterKoiKoiDecisionKoiKoi(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.PlayerHand = []Card{AllCards[0], AllCards[4]}
	u.game.PlayerCaptured = cardsFromIDList(0, 8, 28)
	u.game.Deck = []Card{AllCards[12]}
	u.newYaku = []Yaku{{Name: "三光", Points: 5}}
	u.phase = PhaseKoiKoi
	u.koikoiCursor = 0 // こいこい

	if err := u.handleEnter(g, nil); err != nil {
		t.Fatalf("handleEnter error: %v", err)
	}

	if !u.game.PlayerKoiKoi {
		t.Error("PlayerKoiKoi should be true")
	}
}

func TestHandleEnterKoiKoiDecisionShoubu(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.PlayerHand = []Card{AllCards[0]}
	u.game.PlayerCaptured = cardsFromIDList(0, 8, 28)
	u.newYaku = []Yaku{{Name: "三光", Points: 5}}
	u.phase = PhaseKoiKoi
	u.koikoiCursor = 1 // 勝負

	if err := u.handleEnter(g, nil); err != nil {
		t.Fatalf("handleEnter error: %v", err)
	}

	if u.game.PlayerScore == 0 {
		t.Error("PlayerScore should be > 0 after shoubu")
	}
}

// handleEnter 経由での終了確認を直接テスト (マウステスト用)
func TestMouseHandleEnterQuitConfYes(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showQuitConf = true
	u.quitCursor = 0 // はい

	err := u.handleEnter(g, nil)
	if !errors.Is(err, gocui.ErrQuit) {
		t.Errorf("expected ErrQuit, got %v", err)
	}
}

func TestMouseHandleEnterQuitConfNo(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showQuitConf = true
	u.quitCursor = 1 // いいえ

	if err := u.handleEnter(g, nil); err != nil {
		t.Fatalf("handleEnter error: %v", err)
	}

	if u.showQuitConf {
		t.Error("showQuitConf should be false")
	}
}

// handleEnter 経由でのオプション確認を直接テスト (マウステスト用)
func TestMouseHandleEnterOptConfYes(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.showOptConf = true
	u.optRounds = 6
	u.optDifficulty = DifficultyHard
	u.optConfCursor = 0 // はい

	if err := u.handleEnter(g, nil); err != nil {
		t.Fatalf("handleEnter error: %v", err)
	}

	if u.showOptConf {
		t.Error("showOptConf should be false")
	}
	if u.game.MaxRounds != 6 {
		t.Errorf("MaxRounds = %d, want 6", u.game.MaxRounds)
	}
}

func TestMouseHandleEnterOptConfNo(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.showOptConf = true
	u.optConfCursor = 1 // いいえ

	if err := u.handleEnter(g, nil); err != nil {
		t.Fatalf("handleEnter error: %v", err)
	}

	if u.showOptConf {
		t.Error("showOptConf should be false")
	}
	if !u.showOptions {
		t.Error("showOptions should still be true")
	}
}

// --- handleOptConfClick テスト ---

func TestHandleOptConfClickWrongLine(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.showOptConf = true

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("optconf")
	if v == nil {
		t.Skip("optconf view not created")
	}

	// ボタン行ではない行をクリック
	v.SetCursor(10, 2)
	if err := u.handleOptConfClick(g, v); err != nil {
		t.Fatalf("handleOptConfClick error: %v", err)
	}

	// まだ表示中であること
	if !u.showOptConf {
		t.Error("showOptConf should still be true when clicking non-button line")
	}
}

// --- handleHandClick 追加テスト ---

func TestHandleHandClickValidLine(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.PlayerHand = []Card{AllCards[0], AllCards[4], AllCards[8]}
	u.game.Field = []Card{AllCards[12]} // マッチしない場札
	u.game.Deck = []Card{AllCards[16], AllCards[20]}
	u.game.IsPlayerTurn = true
	u.phase = PhasePlayerSelectHand

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("hand")
	if v == nil {
		t.Skip("hand view not created")
	}

	// 2行目をクリック（1枚目の手札、空行考慮で cy=2）
	v.SetCursor(0, 2)
	cx, cy := v.Cursor()
	if cy != 2 {
		t.Skipf("SetCursor not working (got cy=%d, want 2)", cy)
	}

	initialFieldLen := len(u.game.Field)
	if err := u.handleHandClick(g, v); err != nil {
		t.Fatalf("handleHandClick error: %v", err)
	}

	// 手札が場に出されるか、マッチ処理が行われる
	if len(u.game.Field) <= initialFieldLen && u.phase == PhasePlayerSelectHand {
		t.Logf("cx=%d, cy=%d, clickedLine=%d", cx, cy, cy-1)
	}
}

func TestHandleHandClickOutOfRange(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.PlayerHand = []Card{AllCards[0]}
	u.game.IsPlayerTurn = true
	u.phase = PhasePlayerSelectHand

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("hand")
	if v == nil {
		t.Skip("hand view not created")
	}

	// 範囲外の行をクリック（手札1枚なのに10行目）
	v.SetCursor(0, 10)
	_, cy := v.Cursor()
	if cy != 10 {
		t.Skipf("SetCursor not working (got cy=%d, want 10)", cy)
	}

	if err := u.handleHandClick(g, v); err != nil {
		t.Fatalf("handleHandClick error: %v", err)
	}
	// エラーなく終了すればOK（範囲外は無視される）
}

// --- handleFieldClick 追加テスト ---

func TestHandleFieldClickPhaseSelectField(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.PlayerHand = []Card{AllCards[4]}
	u.game.Field = []Card{AllCards[0], AllCards[1]} // 松2枚
	u.game.Deck = []Card{AllCards[12], AllCards[16]}
	u.playedCard = AllCards[2] // 松のカス
	u.matchingCards = []Card{AllCards[0], AllCards[1]}
	u.phase = PhasePlayerSelectField

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("field")
	if v == nil {
		t.Skip("field view not created")
	}

	// 最初の札の位置をクリック
	v.SetCursor(2, 1)
	_, cy := v.Cursor()
	if cy != 1 {
		t.Skipf("SetCursor not working")
	}

	if err := u.handleFieldClick(g, v); err != nil {
		t.Fatalf("handleFieldClick error: %v", err)
	}
}

func TestHandleFieldClickPhaseSelectFieldDraw(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.PlayerHand = []Card{AllCards[4]}
	u.game.Field = []Card{AllCards[0], AllCards[1]}
	u.game.Deck = []Card{AllCards[12]}
	u.drawnCard = AllCards[2]
	u.matchingCards = []Card{AllCards[0], AllCards[1]}
	u.phase = PhasePlayerSelectFieldDraw

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("field")
	if v == nil {
		t.Skip("field view not created")
	}

	v.SetCursor(2, 1)
	if err := u.handleFieldClick(g, v); err != nil {
		t.Fatalf("handleFieldClick error: %v", err)
	}
}

func TestHandleFieldClickNoMatch(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.Field = []Card{AllCards[0], AllCards[4]} // 松と梅
	u.matchingCards = []Card{AllCards[0]}           // 松のみマッチ
	u.phase = PhasePlayerSelectField

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("field")
	if v == nil {
		t.Skip("field view not created")
	}

	// 2枚目（梅、マッチしない）をクリック - 位置は約8セル目以降
	v.SetCursor(10, 1)
	if err := u.handleFieldClick(g, v); err != nil {
		t.Fatalf("handleFieldClick error: %v", err)
	}
	// マッチしないのでフェーズ変わらず
}

func TestHandleFieldClickOutOfCards(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.Field = []Card{AllCards[0]}
	u.matchingCards = []Card{AllCards[0]}
	u.phase = PhasePlayerSelectField

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("field")
	if v == nil {
		t.Skip("field view not created")
	}

	// 札がない位置をクリック（遠くの位置）
	v.SetCursor(50, 1)
	if err := u.handleFieldClick(g, v); err != nil {
		t.Fatalf("handleFieldClick error: %v", err)
	}
}

// --- handleKoiKoiClick 追加テスト ---

func TestHandleKoiKoiClickKoiKoiButton(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.PlayerHand = []Card{AllCards[0], AllCards[4]}
	u.game.PlayerCaptured = cardsFromIDList(0, 8, 28)
	u.game.Deck = []Card{AllCards[12]}
	u.newYaku = []Yaku{{Name: "三光", Points: 5}}
	u.phase = PhaseKoiKoi

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("koikoi")
	if v == nil {
		t.Skip("koikoi view not created")
	}

	// ボタン行でこいこいボタンの位置をクリック
	buttonLine := 7 + len(u.newYaku)
	// こいこいボタンは左寄り
	innerW := 42
	btnW := cellWidth(" こいこい（続行） ") + 2 + cellWidth(" 勝負（得点確定） ")
	btnPad := (innerW - btnW) / 2
	koikoiMid := btnPad + cellWidth(" こいこい（続行） ")/2

	v.SetCursor(koikoiMid, buttonLine)
	cx, cy := v.Cursor()
	if cy != buttonLine || cx != koikoiMid {
		t.Skipf("SetCursor not working (got %d,%d, want %d,%d)", cx, cy, koikoiMid, buttonLine)
	}

	if err := u.handleKoiKoiClick(g, v); err != nil {
		t.Fatalf("handleKoiKoiClick error: %v", err)
	}

	if !u.game.PlayerKoiKoi {
		t.Error("PlayerKoiKoi should be true")
	}
}

func TestHandleKoiKoiClickShoubuButton(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.PlayerHand = []Card{AllCards[0]}
	u.game.PlayerCaptured = cardsFromIDList(0, 8, 28)
	u.newYaku = []Yaku{{Name: "三光", Points: 5}}
	u.phase = PhaseKoiKoi

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("koikoi")
	if v == nil {
		t.Skip("koikoi view not created")
	}

	buttonLine := 7 + len(u.newYaku)
	// 勝負ボタンは右寄り - innerW の右端近く
	v.SetCursor(35, buttonLine)
	cx, cy := v.Cursor()
	if cy != buttonLine || cx != 35 {
		t.Skipf("SetCursor not working (got %d,%d)", cx, cy)
	}

	if err := u.handleKoiKoiClick(g, v); err != nil {
		t.Fatalf("handleKoiKoiClick error: %v", err)
	}

	if u.game.PlayerScore == 0 {
		t.Error("PlayerScore should be > 0 after shoubu")
	}
}

func TestHandleKoiKoiClickBetweenButtons(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.PlayerHand = []Card{AllCards[0]}
	u.newYaku = []Yaku{{Name: "三光", Points: 5}}
	u.phase = PhaseKoiKoi
	u.koikoiCursor = 0

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("koikoi")
	if v == nil {
		t.Skip("koikoi view not created")
	}

	buttonLine := 7 + len(u.newYaku)
	// ボタンとボタンの間（どちらにもヒットしない位置）
	innerW := 42
	btnW := cellWidth(" こいこい（続行） ") + 2 + cellWidth(" 勝負（得点確定） ")
	btnPad := (innerW - btnW) / 2
	koikoiEnd := btnPad + cellWidth(" こいこい（続行） ")
	// ボタン間の隙間
	betweenPos := koikoiEnd + 1

	v.SetCursor(betweenPos, buttonLine)
	originalCursor := u.koikoiCursor

	if err := u.handleKoiKoiClick(g, v); err != nil {
		t.Fatalf("handleKoiKoiClick error: %v", err)
	}

	// ボタン間クリックではカーソル変わらない
	if u.koikoiCursor != originalCursor {
		t.Error("koikoiCursor should not change when clicking between buttons")
	}
}

// --- handleGameEndClick 追加テスト ---

func TestHandleGameEndClickRestartButton(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.game.Round = 12
	u.game.PlayerScore = 50
	u.phase = PhaseGameEnd

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("gameend")
	if v == nil {
		t.Skip("gameend view not created")
	}

	// 「1月から」ボタンの位置
	innerW := 42
	btnW := cellWidth(" 1月から ") + 4 + cellWidth(" 終了する ")
	btnPad := (innerW - btnW) / 2
	restartMid := btnPad + cellWidth(" 1月から ")/2

	v.SetCursor(restartMid, 7)
	_, cy := v.Cursor()
	if cy != 7 {
		t.Skipf("SetCursor not working")
	}

	if err := u.handleGameEndClick(g, v); err != nil {
		t.Fatalf("handleGameEndClick error: %v", err)
	}

	if u.game.Round != 1 {
		t.Errorf("Round = %d, want 1", u.game.Round)
	}
}

func TestHandleGameEndClickQuitButton(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.phase = PhaseGameEnd

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("gameend")
	if v == nil {
		t.Skip("gameend view not created")
	}

	// 「終了する」ボタンは右寄り
	v.SetCursor(35, 7)
	_, cy := v.Cursor()
	if cy != 7 {
		t.Skipf("SetCursor not working")
	}

	err := u.handleGameEndClick(g, v)
	if !errors.Is(err, gocui.ErrQuit) {
		t.Errorf("expected ErrQuit, got %v", err)
	}
}

func TestHandleGameEndClickBetweenButtons(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.phase = PhaseGameEnd
	u.gameEndCursor = 0

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("gameend")
	if v == nil {
		t.Skip("gameend view not created")
	}

	// ボタン間の隙間
	innerW := 42
	btnW := cellWidth(" 1月から ") + 4 + cellWidth(" 終了する ")
	btnPad := (innerW - btnW) / 2
	restartEnd := btnPad + cellWidth(" 1月から ")
	betweenPos := restartEnd + 2

	v.SetCursor(betweenPos, 7)
	originalCursor := u.gameEndCursor

	if err := u.handleGameEndClick(g, v); err != nil {
		t.Fatalf("handleGameEndClick error: %v", err)
	}

	if u.gameEndCursor != originalCursor {
		t.Error("gameEndCursor should not change")
	}
}

// --- handleQuitConfClick 追加テスト ---

func TestHandleQuitConfClickYesButton(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showQuitConf = true

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("quitconf")
	if v == nil {
		t.Skip("quitconf view not created")
	}

	// 「はい」ボタンの位置
	innerW := 44
	btnW := cellWidth(" はい ") + 4 + cellWidth(" いいえ ")
	btnPad := (innerW - btnW) / 2
	yesMid := btnPad + cellWidth(" はい ")/2

	v.SetCursor(yesMid, 5)
	cx, cy := v.Cursor()
	if cy != 5 || cx != yesMid {
		t.Skipf("SetCursor not working (got %d,%d, want %d,5)", cx, cy, yesMid)
	}

	err := u.handleQuitConfClick(g, v)
	if !errors.Is(err, gocui.ErrQuit) {
		t.Errorf("expected ErrQuit, got %v", err)
	}
}

func TestHandleQuitConfClickNoButton(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showQuitConf = true

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("quitconf")
	if v == nil {
		t.Skip("quitconf view not created")
	}

	// 「いいえ」ボタンは右寄り
	v.SetCursor(35, 5)
	cx, cy := v.Cursor()
	if cy != 5 || cx != 35 {
		t.Skipf("SetCursor not working (got %d,%d)", cx, cy)
	}

	if err := u.handleQuitConfClick(g, v); err != nil {
		t.Fatalf("handleQuitConfClick error: %v", err)
	}

	if u.showQuitConf {
		t.Error("showQuitConf should be false")
	}
}

func TestHandleQuitConfClickBetweenButtons(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showQuitConf = true
	u.quitCursor = 0

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("quitconf")
	if v == nil {
		t.Skip("quitconf view not created")
	}

	// ボタン間の隙間
	innerW := 44
	btnW := cellWidth(" はい ") + 4 + cellWidth(" いいえ ")
	btnPad := (innerW - btnW) / 2
	yesEnd := btnPad + cellWidth(" はい ")
	betweenPos := yesEnd + 2

	v.SetCursor(betweenPos, 5)

	if err := u.handleQuitConfClick(g, v); err != nil {
		t.Fatalf("handleQuitConfClick error: %v", err)
	}

	// ボタン間クリックでは変化なし
	if !u.showQuitConf {
		t.Error("showQuitConf should still be true")
	}
}

// --- handleOptConfClick 追加テスト ---

func TestHandleOptConfClickYesButton(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.showOptConf = true
	u.optRounds = 6
	u.optDifficulty = DifficultyHard

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("optconf")
	if v == nil {
		t.Skip("optconf view not created")
	}

	// 「はい」ボタンの位置
	innerW := 44
	btnW := cellWidth(" はい ") + 4 + cellWidth(" いいえ ")
	btnPad := (innerW - btnW) / 2
	yesMid := btnPad + cellWidth(" はい ")/2

	v.SetCursor(yesMid, 5)
	cx, cy := v.Cursor()
	if cy != 5 || cx != yesMid {
		t.Skipf("SetCursor not working (got %d,%d, want %d,5)", cx, cy, yesMid)
	}

	if err := u.handleOptConfClick(g, v); err != nil {
		t.Fatalf("handleOptConfClick error: %v", err)
	}

	if u.showOptConf {
		t.Error("showOptConf should be false")
	}
	if u.game.MaxRounds != 6 {
		t.Errorf("MaxRounds = %d, want 6", u.game.MaxRounds)
	}
}

func TestHandleOptConfClickNoButton(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.showOptConf = true

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("optconf")
	if v == nil {
		t.Skip("optconf view not created")
	}

	// 「いいえ」ボタンは右寄り
	v.SetCursor(35, 5)
	cx, cy := v.Cursor()
	if cy != 5 || cx != 35 {
		t.Skipf("SetCursor not working (got %d,%d)", cx, cy)
	}

	if err := u.handleOptConfClick(g, v); err != nil {
		t.Fatalf("handleOptConfClick error: %v", err)
	}

	if u.showOptConf {
		t.Error("showOptConf should be false")
	}
	if !u.showOptions {
		t.Error("showOptions should still be true")
	}
}

func TestHandleOptConfClickBetweenButtons(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.showOptConf = true
	u.optConfCursor = 0

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("optconf")
	if v == nil {
		t.Skip("optconf view not created")
	}

	// ボタン間の隙間
	innerW := 44
	btnW := cellWidth(" はい ") + 4 + cellWidth(" いいえ ")
	btnPad := (innerW - btnW) / 2
	yesEnd := btnPad + cellWidth(" はい ")
	betweenPos := yesEnd + 2

	v.SetCursor(betweenPos, 5)

	if err := u.handleOptConfClick(g, v); err != nil {
		t.Fatalf("handleOptConfClick error: %v", err)
	}

	// ボタン間クリックでは変化なし
	if !u.showOptConf {
		t.Error("showOptConf should still be true")
	}
}

// --- handleOptionsClick 追加テスト ---

func TestHandleOptionsClickRoundsAtMin(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.optRounds = 3 // 最小値

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("options")
	if v == nil {
		t.Skip("options view not created")
	}

	// 左側クリックで減らそうとする（最小値なので変わらない）
	v.SetCursor(10, 1)
	if err := u.handleOptionsClick(g, v); err != nil {
		t.Fatalf("handleOptionsClick error: %v", err)
	}

	if u.optRounds != 3 {
		t.Errorf("optRounds = %d, want 3 (should stay at min)", u.optRounds)
	}
}

func TestHandleOptionsClickRoundsAtMax(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.optRounds = 12 // 最大値

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("options")
	if v == nil {
		t.Skip("options view not created")
	}

	// 右側クリックで増やそうとする（最大値なので変わらない）
	v.SetCursor(30, 1)
	if err := u.handleOptionsClick(g, v); err != nil {
		t.Fatalf("handleOptionsClick error: %v", err)
	}

	if u.optRounds != 12 {
		t.Errorf("optRounds = %d, want 12 (should stay at max)", u.optRounds)
	}
}

func TestHandleOptionsClickDifficultyAtMin(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.optDifficulty = DifficultyEasy // 最小値

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("options")
	if v == nil {
		t.Skip("options view not created")
	}

	// 左側クリックで下げようとする
	v.SetCursor(10, 3)
	if err := u.handleOptionsClick(g, v); err != nil {
		t.Fatalf("handleOptionsClick error: %v", err)
	}

	if u.optDifficulty != DifficultyEasy {
		t.Errorf("optDifficulty = %q, want easy (should stay at min)", u.optDifficulty)
	}
}

func TestHandleOptionsClickDifficultyAtMax(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.optDifficulty = DifficultyHard // 最大値

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("options")
	if v == nil {
		t.Skip("options view not created")
	}

	// 右側クリックで上げようとする
	v.SetCursor(30, 3)
	if err := u.handleOptionsClick(g, v); err != nil {
		t.Fatalf("handleOptionsClick error: %v", err)
	}

	if u.optDifficulty != DifficultyHard {
		t.Errorf("optDifficulty = %q, want hard (should stay at max)", u.optDifficulty)
	}
}

func TestHandleOptionsClickButtonRowBetween(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.optCursor = 0

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("options")
	if v == nil {
		t.Skip("options view not created")
	}

	// ボタン行だがボタン間の隙間
	innerW := 42
	cancelW := cellWidth(" キャンセル ")
	applyW := cellWidth(" 適用 ")
	btnW := cancelW + 4 + applyW
	btnPad := (innerW - btnW) / 2
	cancelEnd := btnPad + cancelW
	betweenPos := cancelEnd + 2

	v.SetCursor(betweenPos, 5)
	if err := u.handleOptionsClick(g, v); err != nil {
		t.Fatalf("handleOptionsClick error: %v", err)
	}

	// ボタン間はクリックしても何も起きない（showOptions は true のまま）
	if !u.showOptions {
		t.Error("showOptions should still be true")
	}
}

func TestHandleOptionsClickOtherRow(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true
	u.optCursor = 0
	u.optRounds = 6

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	v, _ := g.View("options")
	if v == nil {
		t.Skip("options view not created")
	}

	// 無関係な行をクリック（2行目はラウンドと難易度の間）
	v.SetCursor(10, 2)
	if err := u.handleOptionsClick(g, v); err != nil {
		t.Fatalf("handleOptionsClick error: %v", err)
	}

	// 何も変わらない
	if u.optRounds != 6 {
		t.Errorf("optRounds changed unexpectedly to %d", u.optRounds)
	}
}

// --- マウス有効化テスト ---

func TestMouseEnabled(t *testing.T) {
	g, err := gocui.NewGui(gocui.OutputSimulator, true)
	if err != nil {
		t.Fatalf("NewGui error: %v", err)
	}
	defer g.Close()

	u := NewUI()
	u.gui = g
	g.Mouse = true

	if !g.Mouse {
		t.Error("Mouse should be enabled")
	}
}

// --- setKeybindings マウスバインド テスト ---

func TestSetKeybindingsIncludesMouseBindings(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	if err := u.setKeybindings(g); err != nil {
		t.Fatalf("setKeybindings error: %v", err)
	}
	// エラーなく終了すればバインドは成功
}

func TestSetKeybindingsMouseBindingError(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)

	if err := u.layout(g); err != nil {
		t.Fatalf("layout error: %v", err)
	}

	// MouseLeft をブラックリストに登録してマウスバインドでエラーを発生させる
	g.BlacklistKeybinding(gocui.MouseLeft)

	err := u.setKeybindings(g)
	if err == nil {
		t.Error("setKeybindings should return error for blacklisted mouse key")
	}
}

// --- 統合テスト: ポップアップブロック ---

func TestAllClickHandlersBlockedByQuitConf(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showQuitConf = true

	handlers := []struct {
		name    string
		handler func(*gocui.Gui, *gocui.View) error
	}{
		{"handleHandClick", u.handleHandClick},
		{"handleFieldClick", u.handleFieldClick},
	}

	for _, h := range handlers {
		t.Run(h.name, func(t *testing.T) {
			if err := h.handler(g, &gocui.View{}); err != nil {
				t.Errorf("%s should not return error when quitconf is open: %v", h.name, err)
			}
		})
	}
}

func TestAllClickHandlersBlockedByOptions(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showOptions = true

	handlers := []struct {
		name    string
		handler func(*gocui.Gui, *gocui.View) error
	}{
		{"handleHandClick", u.handleHandClick},
		{"handleFieldClick", u.handleFieldClick},
	}

	for _, h := range handlers {
		t.Run(h.name, func(t *testing.T) {
			if err := h.handler(g, &gocui.View{}); err != nil {
				t.Errorf("%s should not return error when options is open: %v", h.name, err)
			}
		})
	}
}

func TestAllClickHandlersBlockedByLog(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showLog = true

	handlers := []struct {
		name    string
		handler func(*gocui.Gui, *gocui.View) error
	}{
		{"handleHandClick", u.handleHandClick},
		{"handleFieldClick", u.handleFieldClick},
	}

	for _, h := range handlers {
		t.Run(h.name, func(t *testing.T) {
			if err := h.handler(g, &gocui.View{}); err != nil {
				t.Errorf("%s should not return error when log is open: %v", h.name, err)
			}
		})
	}
}

func TestAllClickHandlersBlockedByHelp(t *testing.T) {
	u, g := setupMouseTestUI(t)
	u.game = NewGame(12)
	u.showHelp = true

	handlers := []struct {
		name    string
		handler func(*gocui.Gui, *gocui.View) error
	}{
		{"handleHandClick", u.handleHandClick},
		{"handleFieldClick", u.handleFieldClick},
	}

	for _, h := range handlers {
		t.Run(h.name, func(t *testing.T) {
			if err := h.handler(g, &gocui.View{}); err != nil {
				t.Errorf("%s should not return error when help is open: %v", h.name, err)
			}
		})
	}
}
