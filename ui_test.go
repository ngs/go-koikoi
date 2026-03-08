package main

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/awesome-gocui/gocui"
)

func TestCellWidth(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"abc", 3},
		{"", 0},
		{"松", 2},
		{"松に鶴", 6},
		{"[松:光]", 7},
		{"hello世界", 9},
	}
	for _, tt := range tests {
		if got := cellWidth(tt.input); got != tt.want {
			t.Errorf("cellWidth(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestIsWide(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'a', false},
		{'1', false},
		{' ', false},
		{'松', true},
		{'鶴', true},
		{'光', true},
		{'ア', true},      // katakana fullwidth character
		{'A', false},     // ASCII
		{0x1100, true},   // Hangul Jamo
		{0x115f, true},   // Hangul Jamo end
		{0x2329, true},   // Left-pointing angle bracket
		{0x232a, true},   // Right-pointing angle bracket
		{0x2e80, true},   // CJK Radicals
		{0x303f, false},  // Ideographic half fill space (exception)
		{0xac00, true},   // Hangul Syllables
		{0xd7a3, true},   // Hangul Syllables end
		{0xf900, true},   // CJK Compatibility Ideographs
		{0xfaff, true},   // CJK Compatibility end
		{0xfe10, true},   // Vertical Forms
		{0xfe19, true},   // Vertical Forms end
		{0xfe30, true},   // CJK Compatibility Forms
		{0xfe6f, true},   // CJK Compatibility Forms end
		{0xff01, true},   // Fullwidth Forms
		{0xff60, true},   // Fullwidth Forms end
		{0xffe0, true},   // Fullwidth Signs
		{0xffe6, true},   // Fullwidth Signs end
		{0x20000, true},  // CJK Unified Ext B
		{0x2fffd, true},  // CJK Unified Ext B end
		{0x30000, true},  // CJK Unified Ext G
		{0x3fffd, true},  // CJK Unified Ext G end
		{0x10ff, false},  // Below range
		{0x3fffe, false}, // Above range
	}
	for _, tt := range tests {
		if got := isWide(tt.r); got != tt.want {
			t.Errorf("isWide(0x%x) = %v, want %v", tt.r, got, tt.want)
		}
	}
}

func TestCenterPad(t *testing.T) {
	tests := []struct {
		text  string
		width int
		want  string
	}{
		{"hello", 10, "  hello"},
		{"hi", 2, "hi"},
		{"hi", 1, "hi"}, // text wider than width
	}
	for _, tt := range tests {
		got := centerPad(tt.text, tt.width)
		if got != tt.want {
			t.Errorf("centerPad(%q, %d) = %q, want %q", tt.text, tt.width, got, tt.want)
		}
	}
}

func TestCenterPadLabel(t *testing.T) {
	// Not highlighted
	got := centerPadLabel("test", 10, false)
	if strings.Contains(got, ansiReverse) {
		t.Error("centerPadLabel(false) should not contain ANSI reverse")
	}

	// Highlighted
	got = centerPadLabel("test", 10, true)
	if !strings.Contains(got, ansiReverse) {
		t.Error("centerPadLabel(true) should contain ANSI reverse")
	}
	if !strings.Contains(got, ansiReset) {
		t.Error("centerPadLabel(true) should contain ANSI reset")
	}

	// Text wider than width, highlighted
	got = centerPadLabel("very long text", 5, true)
	if !strings.Contains(got, ansiReverse) {
		t.Error("centerPadLabel with wide text should still highlight")
	}

	// Text wider than width, not highlighted
	got = centerPadLabel("very long text", 5, false)
	if got != "very long text" {
		t.Errorf("centerPadLabel wide text not highlighted = %q, want %q", got, "very long text")
	}
}

func TestCardLabel(t *testing.T) {
	tests := []struct {
		card Card
		want string
	}{
		{AllCards[0], "[松:光]"},
		{AllCards[1], "[松:短]"},
		{AllCards[2], "[松:カ]"},
		{AllCards[4], "[梅:種]"},
	}
	for _, tt := range tests {
		if got := cardLabel(tt.card); got != tt.want {
			t.Errorf("cardLabel(%d) = %q, want %q", tt.card.ID, got, tt.want)
		}
	}
}

func TestCapturedNames(t *testing.T) {
	cards := []Card{AllCards[0], AllCards[8]}
	got := capturedNames(cards)
	if got != "松に鶴, 桜に幕" {
		t.Errorf("capturedNames = %q, want %q", got, "松に鶴, 桜に幕")
	}

	got = capturedNames(nil)
	if got != "" {
		t.Errorf("capturedNames(nil) = %q, want empty", got)
	}
}

func TestCountCapturedGroups(t *testing.T) {
	// 空
	n := countCapturedGroups(nil)
	if n != 1 {
		t.Errorf("countCapturedGroups(nil) = %d, want 1", n)
	}

	// 光のみ
	cards := cardsFromIDList(0, 8, 28)
	n = countCapturedGroups(cards)
	if n != 2 {
		t.Errorf("countCapturedGroups(光3枚) = %d, want 2", n)
	}

	// 複数種類
	cards = cardsFromIDList(0, 4, 1, 2)
	n = countCapturedGroups(cards)
	if n != 4 {
		t.Errorf("countCapturedGroups(全種類) = %d, want 4", n)
	}
}

func TestNewUI(t *testing.T) {
	u := NewUI()
	if u == nil {
		t.Fatal("NewUI returned nil")
	}
	if u.logLines != nil {
		t.Error("logLines should be nil")
	}
}

func TestAddLog(t *testing.T) {
	u := NewUI()
	u.addLog("test message")
	if len(u.logLines) != 1 {
		t.Errorf("logLines has %d entries, want 1", len(u.logLines))
	}
	if !strings.Contains(u.logLines[0], "test message") {
		t.Errorf("logLines[0] = %q, should contain 'test message'", u.logLines[0])
	}
	if !strings.Contains(u.logLines[0], "[") {
		t.Error("log line should contain timestamp")
	}
}

func TestAddLogTruncation(t *testing.T) {
	u := NewUI()
	for i := 0; i < 1100; i++ {
		u.addLog("msg")
	}
	if len(u.logLines) != 1000 {
		t.Errorf("logLines should be truncated to 1000, got %d", len(u.logLines))
	}
}

func TestPhaseConstants(t *testing.T) {
	if PhasePlayerSelectHand != 0 {
		t.Errorf("PhasePlayerSelectHand = %d, want 0", PhasePlayerSelectHand)
	}
	if PhaseGameEnd != 7 {
		t.Errorf("PhaseGameEnd = %d, want 7", PhaseGameEnd)
	}
}

func TestRoundName(t *testing.T) {
	u := NewUI()
	u.game = NewGame(12)

	tests := []struct {
		round int
		want  string
	}{
		{1, "睦月"},
		{2, "如月"},
		{12, "師走"},
		{13, "睦月"},
	}
	for _, tt := range tests {
		u.game.Round = tt.round
		if got := u.roundName(); got != tt.want {
			t.Errorf("roundName() for round %d = %q, want %q", tt.round, got, tt.want)
		}
	}
}

func TestCapturedGroupsDefinition(t *testing.T) {
	if len(capturedGroups) != 4 {
		t.Errorf("capturedGroups has %d entries, want 4", len(capturedGroups))
	}
}

func TestANSIConstants(t *testing.T) {
	if ansiReset != "\033[0m" {
		t.Error("ansiReset incorrect")
	}
	if ansiReverse != "\033[7m" {
		t.Error("ansiReverse incorrect")
	}
	if ansiYellow != "\033[33m" {
		t.Error("ansiYellow incorrect")
	}
	if ansiDim != "\033[2m" {
		t.Error("ansiDim incorrect")
	}
}

func TestRoundsOptions(t *testing.T) {
	expected := []int{3, 6, 12}
	if len(roundsOptions) != len(expected) {
		t.Fatalf("roundsOptions has %d entries, want %d", len(roundsOptions), len(expected))
	}
	for i, v := range expected {
		if roundsOptions[i] != v {
			t.Errorf("roundsOptions[%d] = %d, want %d", i, roundsOptions[i], v)
		}
	}
}

func TestDiffOptions(t *testing.T) {
	expected := []Difficulty{DifficultyEasy, DifficultyNormal, DifficultyHard}
	if len(diffOptions) != len(expected) {
		t.Fatalf("diffOptions has %d entries, want %d", len(diffOptions), len(expected))
	}
	for i, v := range expected {
		if diffOptions[i] != v {
			t.Errorf("diffOptions[%d] = %q, want %q", i, diffOptions[i], v)
		}
	}
}

// --- テスト用ヘルパー ---

func newTestUI(t *testing.T) *UI {
	t.Helper()
	dir := t.TempDir()
	u := NewUI()
	u.game = NewGame(12)
	u.game.StartRound()
	u.configDir = dir
	u.settingsPath = filepath.Join(dir, "settings.json")
	u.savePath = filepath.Join(dir, "game.json")
	u.settings = DefaultSettings()
	u.difficulty = DifficultyNormal
	u.phase = PhasePlayerSelectHand
	return u
}

func newTestGUI(t *testing.T) *gocui.Gui {
	t.Helper()
	g, err := gocui.NewGui(gocui.OutputSimulator, true)
	if err != nil {
		t.Fatalf("NewGui(OutputSimulator) error: %v", err)
	}
	t.Cleanup(func() { g.Close() })
	return g
}

func newTestUIWithGUI(t *testing.T) (*UI, *gocui.Gui) {
	t.Helper()
	u := newTestUI(t)
	g := newTestGUI(t)
	u.gui = g
	g.SetManagerFunc(u.layout)
	return u, g
}

// --- handleQuit テスト ---

func TestHandleQuit(t *testing.T) {
	u := newTestUI(t)

	// 通常状態 → 終了確認表示
	u.handleQuit(nil, nil)
	if !u.showQuitConf {
		t.Error("handleQuit should show quit confirmation")
	}
	if u.quitCursor != 1 {
		t.Errorf("quitCursor = %d, want 1 (いいえ)", u.quitCursor)
	}
}

func TestHandleQuitFromQuitConf(t *testing.T) {
	u := newTestUI(t)
	u.showQuitConf = true
	u.handleQuit(nil, nil)
	if u.showQuitConf {
		t.Error("handleQuit in quit confirmation should close it")
	}
}

func TestHandleQuitFromOptions(t *testing.T) {
	u := newTestUI(t)
	u.showOptions = true
	u.handleQuit(nil, nil)
	if u.showOptions {
		t.Error("handleQuit should close options")
	}
}

func TestHandleQuitFromHelp(t *testing.T) {
	u := newTestUI(t)
	u.showHelp = true
	u.handleQuit(nil, nil)
	if u.showHelp {
		t.Error("handleQuit should close help")
	}
}

func TestHandleQuitFromLog(t *testing.T) {
	u := newTestUI(t)
	u.showLog = true
	u.handleQuit(nil, nil)
	if u.showLog {
		t.Error("handleQuit should close log")
	}
}

func TestHandleQuitFromOptConf(t *testing.T) {
	u := newTestUI(t)
	u.showOptConf = true
	u.handleQuit(nil, nil)
	if u.showOptConf {
		t.Error("handleQuit should close optconf")
	}
}

// --- handleEsc テスト ---

func TestHandleEsc(t *testing.T) {
	u := newTestUI(t)

	// 何も表示されていない場合
	err := u.handleEsc(nil, nil)
	if err != nil {
		t.Errorf("handleEsc returned error: %v", err)
	}

	// OptConf
	u.showOptConf = true
	u.handleEsc(nil, nil)
	if u.showOptConf {
		t.Error("handleEsc should close optconf")
	}

	// QuitConf
	u.showQuitConf = true
	u.handleEsc(nil, nil)
	if u.showQuitConf {
		t.Error("handleEsc should close quitconf")
	}

	// Options
	u.showOptions = true
	u.handleEsc(nil, nil)
	if u.showOptions {
		t.Error("handleEsc should close options")
	}

	// Help
	u.showHelp = true
	u.handleEsc(nil, nil)
	if u.showHelp {
		t.Error("handleEsc should close help")
	}

	// Log
	u.showLog = true
	u.handleEsc(nil, nil)
	if u.showLog {
		t.Error("handleEsc should close log")
	}
}

// --- handleYes / handleNo テスト ---

func TestHandleYes(t *testing.T) {
	u := newTestUI(t)

	// 終了確認なし → 何もしない
	err := u.handleYes(nil, nil)
	if err != nil {
		t.Error("handleYes without quitconf should return nil")
	}

	// 終了確認あり → ErrQuit
	u.showQuitConf = true
	err = u.handleYes(nil, nil)
	if !errors.Is(err, gocui.ErrQuit) {
		t.Error("handleYes with quitconf should return ErrQuit")
	}
}

func TestHandleNo(t *testing.T) {
	u := newTestUI(t)

	// 終了確認なし → 何もしない
	err := u.handleNo(nil, nil)
	if err != nil {
		t.Error("handleNo without quitconf should return nil")
	}

	// 終了確認あり → 閉じる
	u.showQuitConf = true
	u.handleNo(nil, nil)
	if u.showQuitConf {
		t.Error("handleNo should close quitconf")
	}
}

// --- handleToggle テスト ---

func TestHandleToggleLog(t *testing.T) {
	u := newTestUI(t)

	u.handleToggleLog(nil, nil)
	if !u.showLog {
		t.Error("toggle log should open")
	}
	u.handleToggleLog(nil, nil)
	if u.showLog {
		t.Error("toggle log should close")
	}

	// QuitConf 表示中は無効
	u.showQuitConf = true
	u.handleToggleLog(nil, nil)
	if u.showLog {
		t.Error("toggle log should be blocked by quitconf")
	}
	u.showQuitConf = false

	// Options 表示中は無効
	u.showOptions = true
	u.handleToggleLog(nil, nil)
	if u.showLog {
		t.Error("toggle log should be blocked by options")
	}
}

func TestHandleToggleHelp(t *testing.T) {
	u := newTestUI(t)

	u.handleToggleHelp(nil, nil)
	if !u.showHelp {
		t.Error("toggle help should open")
	}
	u.handleToggleHelp(nil, nil)
	if u.showHelp {
		t.Error("toggle help should close")
	}

	// QuitConf 表示中は無効
	u.showQuitConf = true
	u.handleToggleHelp(nil, nil)
	if u.showHelp {
		t.Error("toggle help should be blocked by quitconf")
	}
	u.showQuitConf = false

	u.showOptions = true
	u.handleToggleHelp(nil, nil)
	if u.showHelp {
		t.Error("toggle help should be blocked by options")
	}
}

func TestHandleToggleOptions(t *testing.T) {
	u := newTestUI(t)
	u.settings.Rounds = 6
	u.settings.Difficulty = DifficultyHard

	// QuitConf 表示中は無効
	u.showQuitConf = true
	u.handleToggleOptions(nil, nil)
	if u.showOptions {
		t.Error("toggle options should be blocked by quitconf")
	}
	u.showQuitConf = false

	// 開く
	u.handleToggleOptions(nil, nil)
	if !u.showOptions {
		t.Error("toggle options should open")
	}
	if u.optRounds != 6 {
		t.Errorf("optRounds = %d, want 6", u.optRounds)
	}
	if u.optDifficulty != DifficultyHard {
		t.Errorf("optDifficulty = %q, want hard", u.optDifficulty)
	}
	if u.optCursor != 0 {
		t.Errorf("optCursor = %d, want 0", u.optCursor)
	}

	// 開いている状態でもう一度 → 閉じる
	u.handleToggleOptions(nil, nil)
	if u.showOptions {
		t.Error("toggle options should close")
	}
}

func TestHandleToggleOptionsClosesOthers(t *testing.T) {
	u := newTestUI(t)
	u.showHelp = true
	u.showLog = true

	u.handleToggleOptions(nil, nil)
	if u.showHelp {
		t.Error("options should close help")
	}
	if u.showLog {
		t.Error("options should close log")
	}
}

// --- handleUp / handleDown テスト ---

func TestHandleUp(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhasePlayerSelectHand
	u.cursor = 3

	u.handleUp(nil, nil)
	if u.cursor != 2 {
		t.Errorf("cursor = %d, want 2", u.cursor)
	}

	u.cursor = 0
	u.handleUp(nil, nil)
	if u.cursor != 0 {
		t.Error("cursor should not go below 0")
	}
}

func TestHandleUpKoiKoi(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhaseKoiKoi
	u.koikoiCursor = 1

	u.handleUp(nil, nil)
	if u.koikoiCursor != 0 {
		t.Errorf("koikoiCursor = %d, want 0", u.koikoiCursor)
	}

	u.handleUp(nil, nil)
	if u.koikoiCursor != 0 {
		t.Error("koikoiCursor should not go below 0")
	}
}

func TestHandleUpOptions(t *testing.T) {
	u := newTestUI(t)
	u.showOptions = true
	u.optCursor = 2

	u.handleUp(nil, nil)
	if u.optCursor != 1 {
		t.Errorf("optCursor = %d, want 1", u.optCursor)
	}
	u.handleUp(nil, nil)
	if u.optCursor != 0 {
		t.Errorf("optCursor = %d, want 0", u.optCursor)
	}
	u.handleUp(nil, nil)
	if u.optCursor != 0 {
		t.Error("optCursor should not go below 0")
	}
}

func TestHandleUpBlockedByPopups(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhasePlayerSelectHand
	u.cursor = 3

	u.showOptConf = true
	u.handleUp(nil, nil)
	if u.cursor != 3 {
		t.Error("up should be blocked by optconf")
	}
	u.showOptConf = false

	u.showQuitConf = true
	u.handleUp(nil, nil)
	if u.cursor != 3 {
		t.Error("up should be blocked by quitconf")
	}
	u.showQuitConf = false

	u.showLog = true
	u.handleUp(nil, nil)
	if u.cursor != 3 {
		t.Error("up should be blocked by log")
	}
	u.showLog = false

	u.showHelp = true
	u.handleUp(nil, nil)
	if u.cursor != 3 {
		t.Error("up should be blocked by help")
	}
}

func TestHandleDown(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhasePlayerSelectHand
	u.cursor = 0

	u.handleDown(nil, nil)
	if u.cursor != 1 {
		t.Errorf("cursor = %d, want 1", u.cursor)
	}

	// 手札の最後まで行けばそれ以上進まない
	u.cursor = len(u.game.PlayerHand) - 1
	u.handleDown(nil, nil)
	if u.cursor != len(u.game.PlayerHand)-1 {
		t.Error("cursor should not exceed hand length")
	}
}

func TestHandleDownKoiKoi(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhaseKoiKoi
	u.koikoiCursor = 0

	u.handleDown(nil, nil)
	if u.koikoiCursor != 1 {
		t.Errorf("koikoiCursor = %d, want 1", u.koikoiCursor)
	}
	u.handleDown(nil, nil)
	if u.koikoiCursor != 1 {
		t.Error("koikoiCursor should not exceed 1")
	}
}

func TestHandleDownOptions(t *testing.T) {
	u := newTestUI(t)
	u.showOptions = true
	u.optCursor = 0

	u.handleDown(nil, nil)
	if u.optCursor != 1 {
		t.Errorf("optCursor = %d, want 1", u.optCursor)
	}
	u.handleDown(nil, nil)
	if u.optCursor != 2 {
		t.Errorf("optCursor = %d, want 2", u.optCursor)
	}
	u.handleDown(nil, nil)
	if u.optCursor != 2 {
		t.Error("optCursor should not exceed 2")
	}
}

func TestHandleDownBlockedByPopups(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhasePlayerSelectHand
	u.cursor = 0

	u.showOptConf = true
	u.handleDown(nil, nil)
	if u.cursor != 0 {
		t.Error("down should be blocked by optconf")
	}
	u.showOptConf = false

	u.showQuitConf = true
	u.handleDown(nil, nil)
	if u.cursor != 0 {
		t.Error("down should be blocked by quitconf")
	}
	u.showQuitConf = false

	u.showLog = true
	u.handleDown(nil, nil)
	if u.cursor != 0 {
		t.Error("down should be blocked by log")
	}
	u.showLog = false

	u.showHelp = true
	u.handleDown(nil, nil)
	if u.cursor != 0 {
		t.Error("down should be blocked by help")
	}
}

// --- handleLeft / handleRight テスト ---

func TestHandleLeftField(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhasePlayerSelectField
	u.matchingCards = []Card{AllCards[0], AllCards[1]}
	u.fieldCursor = 1

	u.handleLeft(nil, nil)
	if u.fieldCursor != 0 {
		t.Errorf("fieldCursor = %d, want 0", u.fieldCursor)
	}
	u.handleLeft(nil, nil)
	if u.fieldCursor != 0 {
		t.Error("fieldCursor should not go below 0")
	}
}

func TestHandleLeftFieldDraw(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhasePlayerSelectFieldDraw
	u.matchingCards = []Card{AllCards[0], AllCards[1]}
	u.fieldCursor = 1

	u.handleLeft(nil, nil)
	if u.fieldCursor != 0 {
		t.Errorf("fieldCursor = %d, want 0", u.fieldCursor)
	}
}

func TestHandleLeftQuitConf(t *testing.T) {
	u := newTestUI(t)
	u.showQuitConf = true
	u.quitCursor = 1

	u.handleLeft(nil, nil)
	if u.quitCursor != 0 {
		t.Errorf("quitCursor = %d, want 0", u.quitCursor)
	}
	u.handleLeft(nil, nil)
	if u.quitCursor != 0 {
		t.Error("quitCursor should not go below 0")
	}
}

func TestHandleLeftOptConf(t *testing.T) {
	u := newTestUI(t)
	u.showOptConf = true
	u.optConfCursor = 1

	u.handleLeft(nil, nil)
	if u.optConfCursor != 0 {
		t.Errorf("optConfCursor = %d, want 0", u.optConfCursor)
	}
	u.handleLeft(nil, nil)
	if u.optConfCursor != 0 {
		t.Error("optConfCursor should not go below 0")
	}
}

func TestHandleLeftOptions(t *testing.T) {
	u := newTestUI(t)
	u.showOptions = true

	// ラウンド数
	u.optCursor = 0
	u.optRounds = 6
	u.handleLeft(nil, nil)
	if u.optRounds != 3 {
		t.Errorf("optRounds = %d, want 3", u.optRounds)
	}
	u.handleLeft(nil, nil)
	if u.optRounds != 3 {
		t.Error("optRounds should not go below minimum")
	}

	// 難易度
	u.optCursor = 1
	u.optDifficulty = DifficultyNormal
	u.handleLeft(nil, nil)
	if u.optDifficulty != DifficultyEasy {
		t.Errorf("optDifficulty = %q, want easy", u.optDifficulty)
	}
	u.handleLeft(nil, nil)
	if u.optDifficulty != DifficultyEasy {
		t.Error("optDifficulty should not go below easy")
	}

	// ボタン行
	u.optCursor = 2
	u.optBtnCursor = 1
	u.handleLeft(nil, nil)
	if u.optBtnCursor != 0 {
		t.Errorf("optBtnCursor = %d, want 0", u.optBtnCursor)
	}
	u.handleLeft(nil, nil)
	if u.optBtnCursor != 0 {
		t.Error("optBtnCursor should not go below 0")
	}
}

func TestHandleLeftBlockedByPopups(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhasePlayerSelectField
	u.matchingCards = []Card{AllCards[0], AllCards[1]}
	u.fieldCursor = 1

	u.showLog = true
	u.handleLeft(nil, nil)
	if u.fieldCursor != 1 {
		t.Error("left should be blocked by log")
	}
	u.showLog = false

	u.showHelp = true
	u.handleLeft(nil, nil)
	if u.fieldCursor != 1 {
		t.Error("left should be blocked by help")
	}
}

func TestHandleRightField(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhasePlayerSelectField
	u.matchingCards = []Card{AllCards[0], AllCards[1]}
	u.fieldCursor = 0

	u.handleRight(nil, nil)
	if u.fieldCursor != 1 {
		t.Errorf("fieldCursor = %d, want 1", u.fieldCursor)
	}
	u.handleRight(nil, nil)
	if u.fieldCursor != 1 {
		t.Error("fieldCursor should not exceed matches length")
	}
}

func TestHandleRightFieldDraw(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhasePlayerSelectFieldDraw
	u.matchingCards = []Card{AllCards[0], AllCards[1]}
	u.fieldCursor = 0

	u.handleRight(nil, nil)
	if u.fieldCursor != 1 {
		t.Errorf("fieldCursor = %d, want 1", u.fieldCursor)
	}
}

func TestHandleRightQuitConf(t *testing.T) {
	u := newTestUI(t)
	u.showQuitConf = true
	u.quitCursor = 0

	u.handleRight(nil, nil)
	if u.quitCursor != 1 {
		t.Errorf("quitCursor = %d, want 1", u.quitCursor)
	}
	u.handleRight(nil, nil)
	if u.quitCursor != 1 {
		t.Error("quitCursor should not exceed 1")
	}
}

func TestHandleRightOptConf(t *testing.T) {
	u := newTestUI(t)
	u.showOptConf = true
	u.optConfCursor = 0

	u.handleRight(nil, nil)
	if u.optConfCursor != 1 {
		t.Errorf("optConfCursor = %d, want 1", u.optConfCursor)
	}
	u.handleRight(nil, nil)
	if u.optConfCursor != 1 {
		t.Error("optConfCursor should not exceed 1")
	}
}

func TestHandleRightOptions(t *testing.T) {
	u := newTestUI(t)
	u.showOptions = true

	// ラウンド数
	u.optCursor = 0
	u.optRounds = 6
	u.handleRight(nil, nil)
	if u.optRounds != 12 {
		t.Errorf("optRounds = %d, want 12", u.optRounds)
	}
	u.handleRight(nil, nil)
	if u.optRounds != 12 {
		t.Error("optRounds should not exceed maximum")
	}

	// 難易度
	u.optCursor = 1
	u.optDifficulty = DifficultyNormal
	u.handleRight(nil, nil)
	if u.optDifficulty != DifficultyHard {
		t.Errorf("optDifficulty = %q, want hard", u.optDifficulty)
	}
	u.handleRight(nil, nil)
	if u.optDifficulty != DifficultyHard {
		t.Error("optDifficulty should not exceed hard")
	}

	// ボタン行
	u.optCursor = 2
	u.optBtnCursor = 0
	u.handleRight(nil, nil)
	if u.optBtnCursor != 1 {
		t.Errorf("optBtnCursor = %d, want 1", u.optBtnCursor)
	}
	u.handleRight(nil, nil)
	if u.optBtnCursor != 1 {
		t.Error("optBtnCursor should not exceed 1")
	}
}

func TestHandleRightBlockedByPopups(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhasePlayerSelectField
	u.matchingCards = []Card{AllCards[0], AllCards[1]}
	u.fieldCursor = 0

	u.showLog = true
	u.handleRight(nil, nil)
	if u.fieldCursor != 0 {
		t.Error("right should be blocked by log")
	}
	u.showLog = false

	u.showHelp = true
	u.handleRight(nil, nil)
	if u.fieldCursor != 0 {
		t.Error("right should be blocked by help")
	}
}

// --- handleEnter テスト ---

func TestHandleEnterQuitConfYes(t *testing.T) {
	u := newTestUI(t)
	u.showQuitConf = true
	u.quitCursor = 0 // はい

	err := u.handleEnter(nil, nil)
	if !errors.Is(err, gocui.ErrQuit) {
		t.Error("handleEnter quit confirm 'yes' should return ErrQuit")
	}
}

func TestHandleEnterQuitConfNo(t *testing.T) {
	u := newTestUI(t)
	u.showQuitConf = true
	u.quitCursor = 1 // いいえ

	u.handleEnter(nil, nil)
	if u.showQuitConf {
		t.Error("handleEnter quit confirm 'no' should close dialog")
	}
}

func TestHandleEnterOptConfYes(t *testing.T) {
	u := newTestUI(t)
	u.showOptConf = true
	u.optConfCursor = 0 // はい
	u.optRounds = 6
	u.optDifficulty = DifficultyHard

	err := u.handleEnter(nil, nil)
	if err != nil {
		t.Errorf("handleEnter optconf yes returned error: %v", err)
	}
	if u.showOptConf {
		t.Error("optconf should be closed")
	}
	if u.settings.Rounds != 6 {
		t.Errorf("settings.Rounds = %d, want 6", u.settings.Rounds)
	}
}

func TestHandleEnterOptConfNo(t *testing.T) {
	u := newTestUI(t)
	u.showOptConf = true
	u.optConfCursor = 1 // いいえ

	u.handleEnter(nil, nil)
	if u.showOptConf {
		t.Error("optconf should be closed")
	}
}

func TestHandleEnterOptionsCancel(t *testing.T) {
	u := newTestUI(t)
	u.showOptions = true
	u.optCursor = 2
	u.optBtnCursor = 0 // キャンセル

	u.handleEnter(nil, nil)
	if u.showOptions {
		t.Error("options should be closed on cancel")
	}
}

func TestHandleEnterOptionsApply(t *testing.T) {
	u := newTestUI(t)
	u.showOptions = true
	u.optCursor = 2
	u.optBtnCursor = 1 // 適用

	u.handleEnter(nil, nil)
	if !u.showOptConf {
		t.Error("apply should show confirmation dialog")
	}
	if u.optConfCursor != 1 {
		t.Errorf("optConfCursor = %d, want 1 (default to no)", u.optConfCursor)
	}
}

func TestHandleEnterOptionsNonButton(t *testing.T) {
	u := newTestUI(t)
	u.showOptions = true
	u.optCursor = 0 // round row, not a button

	err := u.handleEnter(nil, nil)
	if err != nil {
		t.Errorf("handleEnter on non-button option returned error: %v", err)
	}
}

func TestHandleEnterLog(t *testing.T) {
	u := newTestUI(t)
	u.showLog = true

	u.handleEnter(nil, nil)
	if u.showLog {
		t.Error("enter should close log")
	}
}

func TestHandleEnterHelp(t *testing.T) {
	u := newTestUI(t)
	u.showHelp = true

	u.handleEnter(nil, nil)
	if u.showHelp {
		t.Error("enter should close help")
	}
}

func TestHandleEnterPhasePlayerSelectHand(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhasePlayerSelectHand
	u.cursor = 0
	u.game.PlayerHand = []Card{AllCards[12]} // 藤
	u.game.Field = []Card{AllCards[0]}       // 松(マッチなし)
	u.game.Deck = []Card{AllCards[20]}

	err := u.handleEnter(nil, nil)
	if err != nil {
		t.Errorf("handleEnter PlayerSelectHand returned error: %v", err)
	}
}

func TestHandleEnterPhasePlayerSelectField(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhasePlayerSelectField
	u.playedCard = AllCards[2]
	u.game.PlayerHand = []Card{AllCards[2]}
	u.matchingCards = []Card{AllCards[0], AllCards[1]}
	u.game.Field = []Card{AllCards[0], AllCards[1]}
	u.fieldCursor = 0
	u.game.Deck = []Card{AllCards[20]}

	err := u.handleEnter(nil, nil)
	if err != nil {
		t.Errorf("handleEnter PlayerSelectField returned error: %v", err)
	}
}

func TestHandleEnterPhasePlayerDrawResult(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhasePlayerDrawResult
	u.game.PlayerCaptured = nil

	err := u.handleEnter(nil, nil)
	if err != nil {
		t.Errorf("handleEnter PlayerDrawResult returned error: %v", err)
	}
}

func TestHandleEnterPhasePlayerSelectFieldDraw(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhasePlayerSelectFieldDraw
	u.drawnCard = AllCards[2]
	u.matchingCards = []Card{AllCards[0], AllCards[1]}
	u.game.Field = []Card{AllCards[0], AllCards[1]}
	u.fieldCursor = 0

	err := u.handleEnter(nil, nil)
	if err != nil {
		t.Errorf("handleEnter PlayerSelectFieldDraw returned error: %v", err)
	}
}

func TestHandleEnterPhaseKoiKoi(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhaseKoiKoi
	u.koikoiCursor = 1 // 勝負
	u.game.PlayerCaptured = cardsFromIDList(0, 8, 28)

	err := u.handleEnter(nil, nil)
	if err != nil {
		t.Errorf("handleEnter KoiKoi returned error: %v", err)
	}
}

func TestHandleEnterPhaseRoundEnd(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhaseRoundEnd

	err := u.handleEnter(nil, nil)
	if err != nil {
		t.Errorf("handleEnter RoundEnd returned error: %v", err)
	}
	// 次のラウンドが開始される
	if u.phase != PhasePlayerSelectHand && u.phase != PhaseCPUTurn {
		t.Errorf("phase = %d, want PhasePlayerSelectHand or PhaseCPUTurn", u.phase)
	}
}

func TestHandleEnterDefault(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhaseCPUTurn // no handler

	err := u.handleEnter(nil, nil)
	if err != nil {
		t.Errorf("handleEnter on default phase returned error: %v", err)
	}
}

func TestHandleEnterGameEnd(t *testing.T) {
	u := newTestUI(t)
	u.phase = PhaseGameEnd

	err := u.handleEnter(nil, nil)
	if err != nil {
		t.Errorf("handleEnter GameEnd returned error: %v", err)
	}
}

// --- applyOptions テスト ---

func TestApplyOptions(t *testing.T) {
	u := newTestUI(t)
	u.optRounds = 3
	u.optDifficulty = DifficultyHard

	err := u.applyOptions()
	if err != nil {
		t.Fatalf("applyOptions error: %v", err)
	}

	if u.settings.Rounds != 3 {
		t.Errorf("settings.Rounds = %d, want 3", u.settings.Rounds)
	}
	if u.settings.Difficulty != DifficultyHard {
		t.Errorf("settings.Difficulty = %q, want hard", u.settings.Difficulty)
	}
	if u.difficulty != DifficultyHard {
		t.Errorf("difficulty = %q, want hard", u.difficulty)
	}
	if u.game.MaxRounds != 3 {
		t.Errorf("game.MaxRounds = %d, want 3", u.game.MaxRounds)
	}
	if u.showOptions {
		t.Error("options should be closed")
	}
	if u.showOptConf {
		t.Error("optconf should be closed")
	}
	if u.cursor != 0 {
		t.Error("cursor should be reset")
	}
}

func TestApplyOptionsPlayerFirst(t *testing.T) {
	u := newTestUI(t)
	u.optRounds = 3
	u.optDifficulty = DifficultyEasy

	u.applyOptions()
	// NewGame always sets NextParentIsPlayer = true
	if u.phase != PhasePlayerSelectHand {
		t.Errorf("phase = %d, want PhasePlayerSelectHand", u.phase)
	}
}

// --- autoSave テスト ---

func TestAutoSave(t *testing.T) {
	u := newTestUI(t)
	u.autoSave()

	// ファイルが生成されたか確認
	_, err := LoadGame(u.savePath)
	if err != nil {
		t.Errorf("autoSave failed to create save file: %v", err)
	}
}

// --- フェーズ遷移テスト ---

func TestEndPlayerTurnRoundOver(t *testing.T) {
	u := newTestUI(t)
	u.game.PlayerHand = nil
	u.game.CPUHand = nil

	u.endPlayerTurn(nil)
	// IsRoundOver → finishRound
	if u.game.IsPlayerTurn {
		t.Error("IsPlayerTurn should be false")
	}
}

func TestEndPlayerTurnContinue(t *testing.T) {
	u := newTestUI(t)
	// CPUにまだ手札がある
	u.game.CPUHand = []Card{AllCards[0]}
	u.game.PlayerHand = []Card{AllCards[1]}

	u.endPlayerTurn(nil)
	if u.phase != PhaseCPUTurn {
		t.Errorf("phase = %d, want PhaseCPUTurn", u.phase)
	}
}

func TestFinishRoundGameEnd(t *testing.T) {
	u := newTestUI(t)
	u.game.Round = 12
	u.game.MaxRounds = 12

	u.finishRound(nil)
	if u.phase != PhaseGameEnd {
		t.Errorf("phase = %d, want PhaseGameEnd", u.phase)
	}
	if u.game.Round != 13 {
		t.Errorf("Round = %d, want 13", u.game.Round)
	}
}

func TestFinishRoundContinue(t *testing.T) {
	u := newTestUI(t)
	u.game.Round = 1
	u.game.MaxRounds = 12

	u.finishRound(nil)
	if u.phase != PhaseRoundEnd {
		t.Errorf("phase = %d, want PhaseRoundEnd", u.phase)
	}
	if u.game.Round != 2 {
		t.Errorf("Round = %d, want 2", u.game.Round)
	}
}

func TestOnNextRoundPlayerFirst(t *testing.T) {
	u := newTestUI(t)
	u.game.NextParentIsPlayer = true

	u.onNextRound(nil)
	if u.phase != PhasePlayerSelectHand {
		t.Errorf("phase = %d, want PhasePlayerSelectHand", u.phase)
	}
	if u.cursor != 0 {
		t.Error("cursor should be reset")
	}
}

func TestOnNextRoundCPUFirst(t *testing.T) {
	u := newTestUI(t)
	u.game.NextParentIsPlayer = false

	u.onNextRound(nil)
	if u.phase != PhaseCPUTurn {
		t.Errorf("phase = %d, want PhaseCPUTurn", u.phase)
	}
}

// --- onSelectHand テスト ---

func TestOnSelectHandEmptyHand(t *testing.T) {
	u := newTestUI(t)
	u.game.PlayerHand = nil

	err := u.onSelectHand(nil)
	if err != nil {
		t.Errorf("onSelectHand with empty hand returned error: %v", err)
	}
}

func TestOnSelectHandCursorOverflow(t *testing.T) {
	u := newTestUI(t)
	u.cursor = 100 // 手札数より大きい

	err := u.onSelectHand(nil)
	if err != nil {
		t.Errorf("onSelectHand with cursor overflow returned error: %v", err)
	}
}

func TestOnSelectHandTwoMatch(t *testing.T) {
	u := newTestUI(t)
	u.game.PlayerHand = []Card{AllCards[2]}         // 松のカス1
	u.game.Field = []Card{AllCards[0], AllCards[1]} // 松に鶴, 松に赤短
	u.cursor = 0

	u.onSelectHand(nil)
	if u.phase != PhasePlayerSelectField {
		t.Errorf("phase = %d, want PhasePlayerSelectField", u.phase)
	}
	if len(u.matchingCards) != 2 {
		t.Errorf("matchingCards = %d, want 2", len(u.matchingCards))
	}
}

func TestOnSelectHandNoMatch(t *testing.T) {
	u := newTestUI(t)
	u.game.PlayerHand = []Card{AllCards[12]} // 藤
	u.game.Field = []Card{AllCards[0]}       // 松
	u.game.Deck = []Card{AllCards[20]}       // 山札用

	u.onSelectHand(nil)
	// マッチなし → 場に置いて山札を引く
	if u.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (reset)", u.cursor)
	}
}

func TestOnSelectHandOneMatch(t *testing.T) {
	u := newTestUI(t)
	u.game.PlayerHand = []Card{AllCards[2]} // 松のカス1
	u.game.Field = []Card{AllCards[0]}      // 松に鶴(1枚マッチ)
	u.game.Deck = []Card{AllCards[20]}      // 山札

	u.onSelectHand(nil)
	if u.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (reset)", u.cursor)
	}
}

// --- onSelectFieldForHand テスト ---

func TestOnSelectFieldForHand(t *testing.T) {
	u := newTestUI(t)
	u.playedCard = AllCards[2]
	u.game.PlayerHand = []Card{AllCards[2]}
	u.matchingCards = []Card{AllCards[0], AllCards[1]}
	u.game.Field = []Card{AllCards[0], AllCards[1]}
	u.fieldCursor = 1
	u.game.Deck = []Card{AllCards[20]}

	u.onSelectFieldForHand(nil)
	if u.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (reset)", u.cursor)
	}
}

// --- doPlayerDraw テスト ---

func TestDoPlayerDrawEmptyDeck(t *testing.T) {
	u := newTestUI(t)
	u.game.Deck = nil
	u.game.PlayerCaptured = nil

	u.doPlayerDraw(nil)
	// デッキが空 → checkPlayerYaku に進む
}

func TestDoPlayerDrawTwoMatch(t *testing.T) {
	u := newTestUI(t)
	u.game.Field = []Card{AllCards[0], AllCards[1]} // 松2枚
	u.game.Deck = []Card{AllCards[2]}               // 松のカス(2枚マッチ)

	u.doPlayerDraw(nil)
	if u.phase != PhasePlayerSelectFieldDraw {
		t.Errorf("phase = %d, want PhasePlayerSelectFieldDraw", u.phase)
	}
}

func TestDoPlayerDrawNoMatch(t *testing.T) {
	u := newTestUI(t)
	u.game.Field = []Card{AllCards[0]} // 松
	u.game.Deck = []Card{AllCards[12]} // 藤(マッチなし)

	u.doPlayerDraw(nil)
	// マッチなし → 場に置く → checkPlayerYaku
}

func TestDoPlayerDrawOneMatch(t *testing.T) {
	u := newTestUI(t)
	u.game.Field = []Card{AllCards[0]} // 松に鶴
	u.game.Deck = []Card{AllCards[2]}  // 松のカス1(1枚マッチ)

	u.doPlayerDraw(nil)
}

// --- onSelectFieldForDraw テスト ---

func TestOnSelectFieldForDraw(t *testing.T) {
	u := newTestUI(t)
	u.drawnCard = AllCards[2]
	u.matchingCards = []Card{AllCards[0], AllCards[1]}
	u.game.Field = []Card{AllCards[0], AllCards[1]}
	u.fieldCursor = 0

	u.onSelectFieldForDraw(nil)
}

// --- onDrawResult テスト ---

func TestOnDrawResult(t *testing.T) {
	u := newTestUI(t)
	u.game.PlayerCaptured = nil

	err := u.onDrawResult(nil)
	if err != nil {
		t.Errorf("onDrawResult returned error: %v", err)
	}
}

// --- checkPlayerYaku テスト ---

func TestCheckPlayerYakuNewYaku(t *testing.T) {
	u := newTestUI(t)
	u.game.PlayerCaptured = cardsFromIDList(0, 8, 28) // 三光
	u.game.PlayerPrevYaku = nil
	u.game.PlayerHand = []Card{AllCards[4]} // 手札あり

	u.checkPlayerYaku(nil)
	if u.phase != PhaseKoiKoi {
		t.Errorf("phase = %d, want PhaseKoiKoi", u.phase)
	}
	if len(u.newYaku) == 0 {
		t.Error("newYaku should not be empty")
	}
}

func TestCheckPlayerYakuNewYakuNoHand(t *testing.T) {
	u := newTestUI(t)
	u.game.PlayerCaptured = cardsFromIDList(0, 8, 28) // 三光
	u.game.PlayerPrevYaku = nil
	u.game.PlayerHand = nil // 手札なし → 自動勝負

	u.checkPlayerYaku(nil)
	// 手札なし → 自動勝負 → finishRound
	if u.game.PlayerScore == 0 {
		t.Error("PlayerScore should be > 0 after auto-win")
	}
}

func TestCheckPlayerYakuNoNewYaku(t *testing.T) {
	u := newTestUI(t)
	u.game.PlayerCaptured = cardsFromIDList(2, 3) // 役なし
	u.game.PlayerHand = nil
	u.game.CPUHand = nil

	u.checkPlayerYaku(nil)
	// 新しい役なし → endPlayerTurn → IsRoundOver → finishRound
}

// --- onKoiKoiDecision テスト ---

func TestOnKoiKoiDecisionKoiKoi(t *testing.T) {
	u := newTestUI(t)
	u.koikoiCursor = 0                   // こいこい
	u.game.CPUHand = []Card{AllCards[0]} // ラウンド終了させない
	u.game.PlayerHand = []Card{AllCards[1]}

	u.onKoiKoiDecision(nil)
	if !u.game.PlayerKoiKoi {
		t.Error("PlayerKoiKoi should be true")
	}
	if u.phase != PhaseCPUTurn {
		t.Errorf("phase = %d, want PhaseCPUTurn", u.phase)
	}
}

func TestOnKoiKoiDecisionShoubu(t *testing.T) {
	u := newTestUI(t)
	u.koikoiCursor = 1                                // 勝負
	u.game.PlayerCaptured = cardsFromIDList(0, 8, 28) // 三光

	u.onKoiKoiDecision(nil)
	if u.game.PlayerScore == 0 {
		t.Error("PlayerScore should be > 0 after shoubu")
	}
}

// --- quit テスト ---

func TestQuit(t *testing.T) {
	err := quit(nil, nil)
	if !errors.Is(err, gocui.ErrQuit) {
		t.Error("quit should return ErrQuit")
	}
}

// --- gocui シミュレータベースのテスト ---

func TestLayoutAndDraw(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.phase = PhasePlayerSelectHand

	// layout 関数をシミュレータ上で実行
	err := u.layout(g)
	if err != nil {
		t.Fatalf("layout error: %v", err)
	}

	// 各ビューが生成されているか確認
	views := []string{"header", "cpu", "field", "hand", "cpucap", "deck", "mycap", "status"}
	for _, name := range views {
		if _, err := g.View(name); err != nil {
			t.Errorf("view %q not found: %v", name, err)
		}
	}
}

func TestLayoutWithLogPopup(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showLog = true
	u.addLog("テスト行")

	err := u.layout(g)
	if err != nil {
		t.Fatalf("layout error: %v", err)
	}

	if _, err := g.View("log"); err != nil {
		t.Error("log view should exist when showLog is true")
	}
}

func TestLayoutLogClosed(t *testing.T) {
	u, g := newTestUIWithGUI(t)

	// まず開いてビューを作成
	u.showLog = true
	u.layout(g)
	// 次に閉じる
	u.showLog = false
	u.layout(g)

	if _, err := g.View("log"); err == nil {
		t.Error("log view should be deleted when showLog is false")
	}
}

func TestLayoutWithHelpPopup(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showHelp = true

	err := u.layout(g)
	if err != nil {
		t.Fatalf("layout error: %v", err)
	}

	if _, err := g.View("help"); err != nil {
		t.Error("help view should exist when showHelp is true")
	}
}

func TestLayoutHelpClosed(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showHelp = true
	u.layout(g)
	u.showHelp = false
	u.layout(g)

	if _, err := g.View("help"); err == nil {
		t.Error("help view should be deleted when showHelp is false")
	}
}

func TestLayoutWithOptionsPopup(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showOptions = true
	u.optRounds = 12
	u.optDifficulty = DifficultyNormal

	err := u.layout(g)
	if err != nil {
		t.Fatalf("layout error: %v", err)
	}

	if _, err := g.View("options"); err != nil {
		t.Error("options view should exist when showOptions is true")
	}
}

func TestLayoutOptionsClosed(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showOptions = true
	u.optRounds = 12
	u.optDifficulty = DifficultyNormal
	u.layout(g)
	u.showOptions = false
	u.layout(g)

	if _, err := g.View("options"); err == nil {
		t.Error("options view should be deleted when showOptions is false")
	}
}

func TestLayoutWithQuitConfPopup(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showQuitConf = true

	err := u.layout(g)
	if err != nil {
		t.Fatalf("layout error: %v", err)
	}

	if _, err := g.View("quitconf"); err != nil {
		t.Error("quitconf view should exist when showQuitConf is true")
	}
}

func TestLayoutQuitConfClosed(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showQuitConf = true
	u.layout(g)
	u.showQuitConf = false
	u.layout(g)

	if _, err := g.View("quitconf"); err == nil {
		t.Error("quitconf view should be deleted")
	}
}

func TestLayoutWithOptConfPopup(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showOptConf = true

	err := u.layout(g)
	if err != nil {
		t.Fatalf("layout error: %v", err)
	}

	if _, err := g.View("optconf"); err != nil {
		t.Error("optconf view should exist when showOptConf is true")
	}
}

func TestLayoutOptConfClosed(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showOptConf = true
	u.layout(g)
	u.showOptConf = false
	u.layout(g)

	if _, err := g.View("optconf"); err == nil {
		t.Error("optconf view should be deleted")
	}
}

// --- 各 draw 関数のテスト ---

func TestDrawHeader(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.layout(g)
	u.drawHeader(g)
}

func TestDrawCPU(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.layout(g)
	u.drawCPU(g)
}

func TestDrawFieldAllPhases(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.layout(g)

	// PhasePlayerSelectHand
	u.phase = PhasePlayerSelectHand
	u.cursor = 0
	u.drawField(g)

	// PhasePlayerSelectField
	u.phase = PhasePlayerSelectField
	u.matchingCards = []Card{AllCards[0], AllCards[1]}
	u.fieldCursor = 0
	u.drawField(g)

	// PhasePlayerSelectFieldDraw
	u.phase = PhasePlayerSelectFieldDraw
	u.drawField(g)

	// Default phase
	u.phase = PhaseCPUTurn
	u.drawField(g)

	// Empty field
	u.game.Field = nil
	u.phase = PhasePlayerSelectHand
	u.drawField(g)

	u.phase = PhaseCPUTurn
	u.drawField(g)
}

func TestDrawHandAllPhases(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.layout(g)

	// PhasePlayerSelectHand
	u.phase = PhasePlayerSelectHand
	u.cursor = 0
	u.drawHand(g, -1, 14, 50)

	// PhasePlayerSelectField
	u.phase = PhasePlayerSelectField
	u.drawHand(g, -1, 14, 50)

	// PhasePlayerSelectFieldDraw
	u.phase = PhasePlayerSelectFieldDraw
	u.drawHand(g, -1, 14, 50)

	// PhaseKoiKoi
	u.phase = PhaseKoiKoi
	u.newYaku = []Yaku{{"三光", 5}}
	u.koikoiCursor = 0
	u.game.PlayerCaptured = cardsFromIDList(0, 8, 28)
	u.drawHand(g, -1, 14, 50)

	// PhaseKoiKoi with doubling
	u.game.CPUKoiKoi = true
	u.drawHand(g, -1, 14, 50)
	u.game.CPUKoiKoi = false

	// PhaseRoundEnd
	u.phase = PhaseRoundEnd
	u.drawHand(g, -1, 14, 50)

	// PhaseGameEnd - player wins
	u.phase = PhaseGameEnd
	u.game.PlayerScore = 10
	u.game.CPUScore = 5
	u.drawHand(g, -1, 14, 50)

	// PhaseGameEnd - CPU wins
	u.game.PlayerScore = 5
	u.game.CPUScore = 10
	u.drawHand(g, -1, 14, 50)

	// PhaseGameEnd - draw
	u.game.PlayerScore = 5
	u.game.CPUScore = 5
	u.drawHand(g, -1, 14, 50)

	// Default phase
	u.phase = PhaseCPUTurn
	u.drawHand(g, -1, 14, 50)
}

func TestDrawCpuCaptured(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.layout(g)
	u.drawCpuCaptured(g)

	u.game.CPUCaptured = cardsFromIDList(0, 4, 1, 2)
	u.drawCpuCaptured(g)
}

func TestDrawDeck(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.layout(g)
	u.drawDeck(g)
}

func TestDrawMyCaptured(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.layout(g)
	u.drawMyCaptured(g)

	u.game.PlayerCaptured = cardsFromIDList(0, 8, 28)
	u.drawMyCaptured(g)
}

func TestDrawLog(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showLog = true
	u.layout(g)

	// 通常
	u.addLog("test1")
	u.drawLog(g)

	// 50行以上
	for i := 0; i < 60; i++ {
		u.addLog("line")
	}
	u.drawLog(g)
}

func TestDrawHelp(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showHelp = true
	u.layout(g)
	u.drawHelp(g)
}

func TestDrawOptions(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showOptions = true
	u.optRounds = 12
	u.optDifficulty = DifficultyNormal
	u.optCursor = 0
	u.layout(g)

	u.drawOptions(g)

	// ボタン行の各カーソル位置
	u.optCursor = 2
	u.optBtnCursor = 0
	u.drawOptions(g)

	u.optBtnCursor = 1
	u.drawOptions(g)
}

func TestDrawQuitConf(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showQuitConf = true
	u.layout(g)

	u.quitCursor = 0
	u.drawQuitConf(g)

	u.quitCursor = 1
	u.drawQuitConf(g)
}

func TestDrawOptConf(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showOptConf = true
	u.layout(g)

	u.optConfCursor = 0
	u.drawOptConf(g)

	u.optConfCursor = 1
	u.drawOptConf(g)
}

func TestDrawStatusAllPhases(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.layout(g)

	phases := []Phase{
		PhasePlayerSelectHand,
		PhasePlayerSelectField,
		PhasePlayerSelectFieldDraw,
		PhaseKoiKoi,
		PhaseCPUTurn,
		PhaseRoundEnd,
		PhaseGameEnd,
		PhasePlayerDrawResult,
	}
	for _, p := range phases {
		u.phase = p
		u.drawStatus(g)
	}
}

func TestSetKeybindings(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	err := u.setKeybindings(g)
	if err != nil {
		t.Fatalf("setKeybindings error: %v", err)
	}
}

func TestSetKeybindingsError(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	// Ctrl+C をブラックリストに登録してエラーを発生させる
	g.BlacklistKeybinding(gocui.KeyCtrlC)
	err := u.setKeybindings(g)
	if err == nil {
		t.Error("setKeybindings should return error for blacklisted key")
	}
}

func TestWriteCapturedDetail(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.layout(g)

	v, _ := g.View("mycap")
	if v == nil {
		t.Fatal("mycap view not found")
	}

	// 空
	writeCapturedDetail(v, nil)

	// 役あり
	v.Clear()
	cards := cardsFromIDList(0, 8, 28, 4, 1, 2)
	writeCapturedDetail(v, cards)

	// 役なし、カードあり
	v.Clear()
	cards = cardsFromIDList(2, 3)
	writeCapturedDetail(v, cards)
}

func TestSetOverlayTitle(t *testing.T) {
	g := newTestGUI(t)
	// ベースのビューを作る
	g.SetView("test", 0, 1, 40, 10, 0)
	setOverlayTitle(g, "test", 0, 1, 40, " タイトル ")

	// タイトルビューが生成されている
	if _, err := g.View("test_t"); err != nil {
		t.Error("overlay title view should be created")
	}
}

func TestSetOverlayTitleNarrow(t *testing.T) {
	g := newTestGUI(t)
	g.SetView("narrow", 0, 1, 5, 10, 0)
	// 狭い幅でもエラーにならない
	setOverlayTitle(g, "narrow", 0, 1, 5, " とても長いタイトル ")
}

// --- draw 関数の nil view ガードテスト ---

func TestDrawFunctionsNilView(t *testing.T) {
	g := newTestGUI(t)
	u := newTestUI(t)
	// layout を呼ばないので view が存在しない → nil view ガードが発動

	u.drawHeader(g)
	u.drawCPU(g)
	u.drawField(g)
	u.drawHand(g, -1, 0, 40)
	u.drawCpuCaptured(g)
	u.drawDeck(g)
	u.drawMyCaptured(g)
	u.drawLog(g)
	u.drawHelp(g)
	u.drawOptions(g)
	u.drawQuitConf(g)
	u.drawOptConf(g)
	u.drawStatus(g)
}

// --- layout 再呼び出しテスト (既存ビュー更新パス) ---

func TestLayoutSecondCall(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	// 1回目: ビュー生成 (ErrUnknownView)
	u.layout(g)
	// 2回目: 既存ビュー更新 (err == nil)
	err := u.layout(g)
	if err != nil {
		t.Fatalf("layout second call error: %v", err)
	}
}

func TestLayoutSecondCallWithPopups(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showLog = true
	u.showHelp = true
	u.showQuitConf = true
	u.showOptConf = true
	u.showOptions = true
	u.optRounds = 12
	u.optDifficulty = DifficultyNormal

	u.layout(g)
	// 2回目
	u.layout(g)
}

// --- doCPUTurn テスト ---

func TestDoCPUTurnAlreadyRunning(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	cpuTurnRunning.Store(true)
	u.doCPUTurn(g)
	// should return immediately due to double-execution guard
	cpuTurnRunning.Store(false)
}

// --- drawField の残りブランチ ---

func TestDrawFieldSelectHandEmptyField(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.layout(g)

	u.phase = PhasePlayerSelectHand
	u.cursor = 0
	u.game.Field = nil
	u.drawField(g)
}

func TestDrawFieldSelectHandWithCursor(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.layout(g)

	u.phase = PhasePlayerSelectHand
	// cursor がハンド外の場合
	u.game.PlayerHand = nil
	u.game.Field = []Card{AllCards[0]}
	u.drawField(g)
}

func TestDrawFieldSelectFieldCursorHighlight(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.layout(g)

	// 2枚マッチで cursor=1 (2番目をハイライト)
	u.phase = PhasePlayerSelectField
	u.matchingCards = []Card{AllCards[0], AllCards[1]}
	u.fieldCursor = 1
	u.game.Field = []Card{AllCards[0], AllCards[1], AllCards[4]} // 非マッチ札も含む
	u.drawField(g)
}

// --- drawOptions の btnPad 負数テスト ---

func TestDrawOptionsNarrow(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showOptions = true
	u.optRounds = 12
	u.optDifficulty = DifficultyNormal
	u.optCursor = 1 // 難易度行
	u.layout(g)
	u.drawOptions(g)
}

// --- drawQuitConf / drawOptConf の全カーソル位置 ---

func TestDrawQuitConfBothCursors(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showQuitConf = true
	u.layout(g)

	// cursor=0 (はい)
	u.quitCursor = 0
	u.drawQuitConf(g)

	// set cursor to "No" button so only one button is highlighted
	u.quitCursor = 1
	u.drawQuitConf(g)
}

func TestDrawOptConfBothCursors(t *testing.T) {
	u, g := newTestUIWithGUI(t)
	u.showOptConf = true
	u.layout(g)

	u.optConfCursor = 0
	u.drawOptConf(g)

	u.optConfCursor = 1
	u.drawOptConf(g)
}

// --- CPUChooseHandCard Easy のランダムブランチ強化 ---

func TestCPUChooseHandCardEasyNoMatchRandom(t *testing.T) {
	g := NewGame(12)
	g.CPUHand = []Card{AllCards[12]} // 藤
	g.Field = []Card{AllCards[0]}    // 松(マッチなし)

	// Easy のランダムパスでも動作する
	for i := 0; i < 30; i++ {
		hand, field := CPUChooseHandCard(g, DifficultyEasy)
		if hand.ID != 12 {
			t.Errorf("iteration %d: expected hand ID 12, got %d", i, hand.ID)
		}
		if field != nil {
			t.Errorf("iteration %d: expected nil field card for no-match", i)
		}
	}
}
