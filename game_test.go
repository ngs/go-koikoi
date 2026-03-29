package main

import "testing"

func TestNewGame(t *testing.T) {
	g := NewGame(12)
	if g.MaxRounds != 12 {
		t.Errorf("MaxRounds = %d, want 12", g.MaxRounds)
	}
	if g.Round != 1 {
		t.Errorf("Round = %d, want 1", g.Round)
	}
	// 親決めで引いた札が記録されているか
	if g.PlayerDrawnCard == nil {
		t.Error("PlayerDrawnCard should not be nil after NewGame")
	}
	if g.CPUDrawnCard == nil {
		t.Error("CPUDrawnCard should not be nil after NewGame")
	}
	// 親決めの結果が正しいか（月が若い方が親）
	if g.PlayerDrawnCard.Month < g.CPUDrawnCard.Month && !g.NextParentIsPlayer {
		t.Error("Player drew earlier month but is not parent")
	}
	if g.CPUDrawnCard.Month < g.PlayerDrawnCard.Month && g.NextParentIsPlayer {
		t.Error("CPU drew earlier month but player is parent")
	}
}

func TestDetermineParentBothPaths(t *testing.T) {
	// 複数回実行して両方のパスをカバー
	playerWins := 0
	cpuWins := 0
	for i := 0; i < 100; i++ {
		g := NewGame(12)
		if g.NextParentIsPlayer {
			playerWins++
		} else {
			cpuWins++
		}
	}
	// 統計的にどちらも0でないはず
	if playerWins == 0 {
		t.Error("Player never won parent determination in 100 tries")
	}
	if cpuWins == 0 {
		t.Error("CPU never won parent determination in 100 tries")
	}
}

func TestStartRound(t *testing.T) {
	g := NewGame(12)
	g.StartRound()

	if len(g.PlayerHand) != 8 {
		t.Errorf("PlayerHand has %d cards, want 8", len(g.PlayerHand))
	}
	if len(g.CPUHand) != 8 {
		t.Errorf("CPUHand has %d cards, want 8", len(g.CPUHand))
	}
	if len(g.Field) != 8 {
		t.Errorf("Field has %d cards, want 8", len(g.Field))
	}
	if len(g.Deck) != 24 { // 48 - 8 - 8 - 8
		t.Errorf("Deck has %d cards, want 24", len(g.Deck))
	}
	if g.PlayerCaptured != nil {
		t.Error("PlayerCaptured should be nil")
	}
	if g.CPUCaptured != nil {
		t.Error("CPUCaptured should be nil")
	}
	if g.PlayerPrevYaku != nil {
		t.Error("PlayerPrevYaku should be nil")
	}
	if g.CPUPrevYaku != nil {
		t.Error("CPUPrevYaku should be nil")
	}
	if g.PlayerKoiKoi {
		t.Error("PlayerKoiKoi should be false")
	}
	if g.CPUKoiKoi {
		t.Error("CPUKoiKoi should be false")
	}
}

func TestStartRoundParent(t *testing.T) {
	g := NewGame(12)
	g.NextParentIsPlayer = true
	g.StartRound()
	if !g.IsPlayerTurn {
		t.Error("IsPlayerTurn should be true when NextParentIsPlayer is true")
	}

	g.NextParentIsPlayer = false
	g.StartRound()
	if g.IsPlayerTurn {
		t.Error("IsPlayerTurn should be false when NextParentIsPlayer is false")
	}
}

func TestStartRoundAllCardsUnique(t *testing.T) {
	g := NewGame(12)
	g.StartRound()

	seen := make(map[int]bool)
	all := append(append(append(append([]Card{}, g.PlayerHand...), g.CPUHand...), g.Field...), g.Deck...)
	if len(all) != 48 {
		t.Fatalf("Total cards = %d, want 48", len(all))
	}
	for _, c := range all {
		if seen[c.ID] {
			t.Errorf("Card ID %d appears more than once", c.ID)
		}
		seen[c.ID] = true
	}
}

func TestDrawFromDeck(t *testing.T) {
	g := NewGame(12)
	g.Deck = []Card{AllCards[0], AllCards[1]}

	c, ok := g.DrawFromDeck()
	if !ok {
		t.Error("DrawFromDeck should return true")
	}
	if c.ID != 0 {
		t.Errorf("DrawFromDeck returned card ID %d, want 0", c.ID)
	}
	if len(g.Deck) != 1 {
		t.Errorf("Deck has %d cards, want 1", len(g.Deck))
	}

	c, ok = g.DrawFromDeck()
	if !ok {
		t.Error("DrawFromDeck should return true")
	}
	if c.ID != 1 {
		t.Errorf("DrawFromDeck returned card ID %d, want 1", c.ID)
	}

	_, ok = g.DrawFromDeck()
	if ok {
		t.Error("DrawFromDeck should return false when deck is empty")
	}
}

func TestMatchingFieldCards(t *testing.T) {
	g := NewGame(12)
	g.Field = []Card{AllCards[0], AllCards[1], AllCards[4], AllCards[8]}

	// 松(January)のカードでマッチ
	matches := g.MatchingFieldCards(AllCards[2]) // 松のカス1
	if len(matches) != 2 {                       // 松に鶴(0), 松に赤短(1)
		t.Errorf("MatchingFieldCards got %d matches, want 2", len(matches))
	}

	// 桜(March)のカードでマッチ
	matches = g.MatchingFieldCards(AllCards[9]) // 桜に赤短
	if len(matches) != 1 {                      // 桜に幕(8)
		t.Errorf("MatchingFieldCards got %d matches, want 1", len(matches))
	}

	// マッチなし
	matches = g.MatchingFieldCards(AllCards[12]) // 藤に不如帰
	if len(matches) != 0 {
		t.Errorf("MatchingFieldCards got %d matches, want 0", len(matches))
	}
}

func TestPlayCardNoMatch(t *testing.T) {
	g := NewGame(12)
	g.PlayerHand = []Card{AllCards[12]} // 藤に不如帰
	g.Field = []Card{AllCards[0]}       // 松に鶴

	captured := g.PlayCard(AllCards[12], nil, true)
	if len(captured) != 0 {
		t.Errorf("PlayCard with no match should return 0 captured, got %d", len(captured))
	}
	if len(g.Field) != 2 {
		t.Errorf("Field should have 2 cards, got %d", len(g.Field))
	}
	if len(g.PlayerHand) != 0 {
		t.Errorf("PlayerHand should have 0 cards, got %d", len(g.PlayerHand))
	}
}

func TestPlayCardOneMatch(t *testing.T) {
	g := NewGame(12)
	g.PlayerHand = []Card{AllCards[2]} // 松のカス1
	g.Field = []Card{AllCards[0]}      // 松に鶴

	captured := g.PlayCard(AllCards[2], nil, true)
	if len(captured) != 2 {
		t.Errorf("PlayCard with 1 match should return 2 captured, got %d", len(captured))
	}
	if len(g.Field) != 0 {
		t.Errorf("Field should have 0 cards, got %d", len(g.Field))
	}
	if len(g.PlayerCaptured) != 2 {
		t.Errorf("PlayerCaptured should have 2 cards, got %d", len(g.PlayerCaptured))
	}
}

func TestPlayCardTwoMatchWithChoice(t *testing.T) {
	g := NewGame(12)
	g.PlayerHand = []Card{AllCards[2]}         // 松のカス1
	g.Field = []Card{AllCards[0], AllCards[1]} // 松に鶴, 松に赤短

	choice := AllCards[1] // 松に赤短を選択
	captured := g.PlayCard(AllCards[2], &choice, true)
	if len(captured) != 2 {
		t.Errorf("PlayCard with 2 match+choice should return 2 captured, got %d", len(captured))
	}
	// 選択した松に赤短が取れているか
	if captured[1].ID != 1 {
		t.Errorf("captured[1].ID = %d, want 1", captured[1].ID)
	}
	if len(g.Field) != 1 {
		t.Errorf("Field should have 1 card, got %d", len(g.Field))
	}
}

func TestPlayCardTwoMatchNoChoice(t *testing.T) {
	g := NewGame(12)
	g.PlayerHand = []Card{AllCards[2]}         // 松のカス1
	g.Field = []Card{AllCards[0], AllCards[1]} // 松に鶴, 松に赤短

	// fieldChoice == nil の場合、最初のマッチを取る
	captured := g.PlayCard(AllCards[2], nil, true)
	if len(captured) != 2 {
		t.Errorf("PlayCard with 2 match no choice should return 2 captured, got %d", len(captured))
	}
	if captured[1].ID != 0 {
		t.Errorf("captured[1].ID = %d, want 0 (first match)", captured[1].ID)
	}
}

func TestPlayCardThreeMatch(t *testing.T) {
	g := NewGame(12)
	g.PlayerHand = []Card{AllCards[3]}                      // 松のカス2
	g.Field = []Card{AllCards[0], AllCards[1], AllCards[2]} // 松に鶴, 松に赤短, 松のカス1

	captured := g.PlayCard(AllCards[3], nil, true)
	if len(captured) != 4 { // 手札1枚 + 場3枚
		t.Errorf("PlayCard with 3 match should return 4 captured, got %d", len(captured))
	}
	if len(g.Field) != 0 {
		t.Errorf("Field should have 0 cards, got %d", len(g.Field))
	}
}

func TestPlayCardCPU(t *testing.T) {
	g := NewGame(12)
	g.CPUHand = []Card{AllCards[2]}
	g.Field = []Card{AllCards[0]}

	captured := g.PlayCard(AllCards[2], nil, false)
	if len(captured) != 2 {
		t.Errorf("PlayCard for CPU should return 2 captured, got %d", len(captured))
	}
	if len(g.CPUCaptured) != 2 {
		t.Errorf("CPUCaptured should have 2 cards, got %d", len(g.CPUCaptured))
	}
	if len(g.CPUHand) != 0 {
		t.Errorf("CPUHand should have 0 cards, got %d", len(g.CPUHand))
	}
}

func TestPlayDrawnCardNoMatch(t *testing.T) {
	g := NewGame(12)
	g.Field = []Card{AllCards[0]}

	captured := g.PlayDrawnCard(AllCards[12], nil, true) // 藤に不如帰
	if len(captured) != 0 {
		t.Errorf("PlayDrawnCard with no match should return 0, got %d", len(captured))
	}
	if len(g.Field) != 2 {
		t.Errorf("Field should have 2 cards, got %d", len(g.Field))
	}
}

func TestPlayDrawnCardOneMatch(t *testing.T) {
	g := NewGame(12)
	g.Field = []Card{AllCards[0]}

	captured := g.PlayDrawnCard(AllCards[2], nil, true) // 松のカス1
	if len(captured) != 2 {
		t.Errorf("PlayDrawnCard with 1 match should return 2, got %d", len(captured))
	}
}

func TestPlayDrawnCardTwoMatchWithChoice(t *testing.T) {
	g := NewGame(12)
	g.Field = []Card{AllCards[0], AllCards[1]}

	choice := AllCards[1]
	captured := g.PlayDrawnCard(AllCards[2], &choice, true)
	if len(captured) != 2 {
		t.Errorf("PlayDrawnCard with 2 match+choice should return 2, got %d", len(captured))
	}
	if captured[1].ID != 1 {
		t.Errorf("captured[1].ID = %d, want 1", captured[1].ID)
	}
}

func TestPlayDrawnCardTwoMatchNoChoice(t *testing.T) {
	g := NewGame(12)
	g.Field = []Card{AllCards[0], AllCards[1]}

	captured := g.PlayDrawnCard(AllCards[2], nil, true)
	if len(captured) != 2 {
		t.Errorf("PlayDrawnCard with 2 match no choice should return 2, got %d", len(captured))
	}
	if captured[1].ID != 0 {
		t.Errorf("captured[1].ID = %d, want 0", captured[1].ID)
	}
}

func TestPlayDrawnCardThreeMatch(t *testing.T) {
	g := NewGame(12)
	g.Field = []Card{AllCards[0], AllCards[1], AllCards[2]}

	captured := g.PlayDrawnCard(AllCards[3], nil, true)
	if len(captured) != 4 {
		t.Errorf("PlayDrawnCard with 3 match should return 4, got %d", len(captured))
	}
	if len(g.Field) != 0 {
		t.Errorf("Field should have 0 cards, got %d", len(g.Field))
	}
}

func TestPlayDrawnCardCPU(t *testing.T) {
	g := NewGame(12)
	g.Field = []Card{AllCards[0]}

	captured := g.PlayDrawnCard(AllCards[2], nil, false)
	if len(captured) != 2 {
		t.Errorf("PlayDrawnCard for CPU should return 2, got %d", len(captured))
	}
	if len(g.CPUCaptured) != 2 {
		t.Errorf("CPUCaptured should have 2, got %d", len(g.CPUCaptured))
	}
}

func TestCheckNewYakuPlayer(t *testing.T) {
	g := NewGame(12)
	g.PlayerCaptured = cardsFromIDList(0, 8, 28) // 三光
	g.PlayerPrevYaku = nil

	newYaku := g.CheckNewYaku(true)
	if len(newYaku) == 0 {
		t.Error("CheckNewYaku should return new yakus")
	}
	found := false
	for _, y := range newYaku {
		if y.Name == "三光" {
			found = true
		}
	}
	if !found {
		t.Error("三光 should be in new yakus")
	}
}

func TestCheckNewYakuCPU(t *testing.T) {
	g := NewGame(12)
	g.CPUCaptured = cardsFromIDList(1, 5, 9) // 赤短
	g.CPUPrevYaku = nil

	newYaku := g.CheckNewYaku(false)
	found := false
	for _, y := range newYaku {
		if y.Name == "赤短" {
			found = true
		}
	}
	if !found {
		t.Error("赤短 should be in new yakus")
	}
}

func TestCheckNewYakuNoNew(t *testing.T) {
	g := NewGame(12)
	g.PlayerCaptured = cardsFromIDList(0, 8, 28) // 三光
	g.PlayerPrevYaku = []Yaku{{"三光", 5}}

	newYaku := g.CheckNewYaku(true)
	if len(newYaku) != 0 {
		t.Errorf("CheckNewYaku should return 0 new yakus, got %d", len(newYaku))
	}
}

func TestCheckNewYakuUpgraded(t *testing.T) {
	g := NewGame(12)
	g.PlayerCaptured = cardsFromIDList(0, 8, 28, 44) // 四光
	g.PlayerPrevYaku = []Yaku{{"三光", 5}}

	newYaku := g.CheckNewYaku(true)
	found := false
	for _, y := range newYaku {
		if y.Name == yakuSikou {
			found = true
		}
	}
	if !found {
		t.Error("四光 should be detected as new yaku (upgrade from 三光)")
	}
}

func TestUpdatePrevYaku(t *testing.T) {
	g := NewGame(12)
	g.PlayerCaptured = cardsFromIDList(0, 8, 28)
	g.UpdatePrevYaku(true)
	if len(g.PlayerPrevYaku) == 0 {
		t.Error("PlayerPrevYaku should be updated")
	}

	g.CPUCaptured = cardsFromIDList(1, 5, 9)
	g.UpdatePrevYaku(false)
	if len(g.CPUPrevYaku) == 0 {
		t.Error("CPUPrevYaku should be updated")
	}
}

func TestCalcFinalScoreBasic(t *testing.T) {
	g := NewGame(12)
	g.PlayerCaptured = cardsFromIDList(0, 8, 28) // 三光 5文
	score := g.CalcFinalScore(true)
	if score != 5 {
		t.Errorf("CalcFinalScore = %d, want 5", score)
	}
}

func TestCalcFinalScoreDoubling(t *testing.T) {
	g := NewGame(12)
	g.PlayerCaptured = cardsFromIDList(0, 8, 28, 44) // 四光 8文 → 16文(7以上2倍)
	score := g.CalcFinalScore(true)
	if score != 16 {
		t.Errorf("CalcFinalScore = %d, want 16", score)
	}
}

func TestCalcFinalScoreKoiKoiPenalty(t *testing.T) {
	g := NewGame(12)
	g.PlayerCaptured = cardsFromIDList(0, 8, 28) // 三光 5文
	g.CPUKoiKoi = true                           // 相手がこいこいしていた
	score := g.CalcFinalScore(true)
	if score != 10 { // 5 * 2 = 10
		t.Errorf("CalcFinalScore with opponent koikoi = %d, want 10", score)
	}
}

func TestCalcFinalScoreDoublingAndKoiKoi(t *testing.T) {
	g := NewGame(12)
	g.PlayerCaptured = cardsFromIDList(0, 8, 28, 44) // 四光 8文 → 16(7以上) → 32(こいこい)
	g.CPUKoiKoi = true
	score := g.CalcFinalScore(true)
	if score != 32 {
		t.Errorf("CalcFinalScore = %d, want 32", score)
	}
}

func TestCalcFinalScoreCPU(t *testing.T) {
	g := NewGame(12)
	g.CPUCaptured = cardsFromIDList(0, 8, 28) // 三光 5文
	g.PlayerKoiKoi = true
	score := g.CalcFinalScore(false)
	if score != 10 { // 5 * 2 = 10
		t.Errorf("CalcFinalScore for CPU = %d, want 10", score)
	}
}

func TestIsRoundOver(t *testing.T) {
	g := NewGame(12)
	g.PlayerHand = []Card{AllCards[0]}
	g.CPUHand = []Card{AllCards[1]}
	if g.IsRoundOver() {
		t.Error("IsRoundOver should be false when hands are not empty")
	}

	g.PlayerHand = nil
	g.CPUHand = nil
	if !g.IsRoundOver() {
		t.Error("IsRoundOver should be true when hands are empty")
	}

	g.PlayerHand = nil
	g.CPUHand = []Card{AllCards[0]}
	if g.IsRoundOver() {
		t.Error("IsRoundOver should be false when CPU hand is not empty")
	}
}

func TestRemoveCard(t *testing.T) {
	cards := []Card{AllCards[0], AllCards[1], AllCards[2]}
	result := removeCard(cards, AllCards[1])
	if len(result) != 2 {
		t.Errorf("removeCard result has %d cards, want 2", len(result))
	}
	for _, c := range result {
		if c.ID == 1 {
			t.Error("removed card should not be in result")
		}
	}
}

func TestRemoveCardNotFound(t *testing.T) {
	cards := []Card{AllCards[0], AllCards[1]}
	result := removeCard(cards, AllCards[99%48])
	if len(result) != 2 {
		t.Errorf("removeCard with missing card should not change slice, got %d", len(result))
	}
}

func TestCryptoIntn(t *testing.T) {
	for i := 0; i < 100; i++ {
		v := cryptoIntn(10)
		if v < 0 || v >= 10 {
			t.Errorf("cryptoIntn(10) = %d, want 0-9", v)
		}
	}
}
