package main

import "testing"

func TestCPUChooseHandCardEasy(t *testing.T) {
	g := NewGame(12)
	g.CPUHand = []Card{AllCards[0], AllCards[4], AllCards[12]}
	g.Field = []Card{AllCards[2], AllCards[6]} // 松のカス, 梅のカス

	// Easy でも動作することを確認（ランダム要素あるため繰り返し）
	for i := 0; i < 20; i++ {
		hand, _ := CPUChooseHandCard(g, DifficultyEasy)
		found := false
		for _, c := range g.CPUHand {
			if c.ID == hand.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("CPUChooseHandCard returned card not in hand: ID=%d", hand.ID)
		}
	}
}

func TestCPUChooseHandCardNormal(t *testing.T) {
	g := NewGame(12)
	g.CPUHand = []Card{AllCards[0], AllCards[2]} // 松に鶴(光), 松のカス1
	g.Field = []Card{AllCards[1]}                // 松に赤短

	hand, _ := CPUChooseHandCard(g, DifficultyNormal)
	// 光札の方が価値が高いので松に鶴を選ぶはず
	if hand.ID != 0 {
		t.Errorf("CPUChooseHandCard(Normal) chose ID=%d, want 0 (higher value)", hand.ID)
	}
}

func TestCPUChooseHandCardHard(t *testing.T) {
	g := NewGame(12)
	g.CPUHand = []Card{AllCards[2], AllCards[14]} // 松のカス1, 藤のカス1
	g.Field = []Card{AllCards[0], AllCards[12]}   // 松に鶴(光), 藤に不如帰(種)
	g.CPUCaptured = nil

	hand, _ := CPUChooseHandCard(g, DifficultyHard)
	// Hard では光札を取れるマッチにボーナスがつく
	if hand.ID != 2 {
		t.Errorf("CPUChooseHandCard(Hard) chose ID=%d, want 2 (match with hikari)", hand.ID)
	}
}

func TestCPUChooseHandCardNoMatch(t *testing.T) {
	g := NewGame(12)
	g.CPUHand = []Card{AllCards[0], AllCards[2]} // 松に鶴(光), 松のカス1
	g.Field = []Card{AllCards[12]}               // 藤(マッチなし)

	hand, _ := CPUChooseHandCard(g, DifficultyNormal)
	// マッチがない場合、一番価値の低い札を出す
	if hand.ID != 2 {
		t.Errorf("CPUChooseHandCard with no match chose ID=%d, want 2 (lowest value)", hand.ID)
	}
}

func TestCPUChooseHandCardTwoMatch(t *testing.T) {
	g := NewGame(12)
	g.CPUHand = []Card{AllCards[2]}            // 松のカス1
	g.Field = []Card{AllCards[0], AllCards[1]} // 松に鶴, 松に赤短 (2枚マッチ)

	hand, fieldCard := CPUChooseHandCard(g, DifficultyNormal)
	if hand.ID != 2 {
		t.Errorf("hand.ID = %d, want 2", hand.ID)
	}
	if fieldCard == nil {
		t.Fatal("fieldCard should not be nil when 2 matches")
	}
	// chooseBestMatch は価値の高い方を選ぶ: 光(20) > 短冊(5)
	if fieldCard.ID != 0 {
		t.Errorf("fieldCard.ID = %d, want 0 (hikari is best)", fieldCard.ID)
	}
}

func TestCPUChooseFieldCard(t *testing.T) {
	matches := []Card{AllCards[0], AllCards[1]} // 松に鶴(光), 松に赤短
	choice := CPUChooseFieldCard(matches)
	if choice == nil {
		t.Fatal("CPUChooseFieldCard should return a choice for 2 matches")
	}
	if choice.ID != 0 { // 光が価値が高い
		t.Errorf("CPUChooseFieldCard chose ID=%d, want 0", choice.ID)
	}
}

func TestCPUChooseFieldCardNonTwo(t *testing.T) {
	// 1枚マッチ
	matches := []Card{AllCards[0]}
	choice := CPUChooseFieldCard(matches)
	if choice != nil {
		t.Error("CPUChooseFieldCard should return nil for non-2 matches")
	}

	// 0枚
	choice = CPUChooseFieldCard(nil)
	if choice != nil {
		t.Error("CPUChooseFieldCard should return nil for 0 matches")
	}

	// 3枚
	matches = []Card{AllCards[0], AllCards[1], AllCards[2]}
	choice = CPUChooseFieldCard(matches)
	if choice != nil {
		t.Error("CPUChooseFieldCard should return nil for 3 matches")
	}
}

func TestCPUDecideKoiKoiEasy(t *testing.T) {
	g := NewGame(12)
	g.CPUHand = make([]Card, 5)
	yakus := []Yaku{{"三光", 5}}
	if CPUDecideKoiKoi(g, yakus, DifficultyEasy) {
		t.Error("Easy CPU should never koikoi")
	}
}

func TestCPUDecideKoiKoiNormal(t *testing.T) {
	g := NewGame(12)

	// 手札3枚、5文 → こいこいする
	g.CPUHand = make([]Card, 3)
	yakus := []Yaku{{"三光", 5}}
	if !CPUDecideKoiKoi(g, yakus, DifficultyNormal) {
		t.Error("Normal CPU should koikoi with 3 cards and 5 points")
	}

	// 手札2枚 → こいこいしない
	g.CPUHand = make([]Card, 2)
	if CPUDecideKoiKoi(g, yakus, DifficultyNormal) {
		t.Error("Normal CPU should not koikoi with 2 cards")
	}

	// 手札3枚、7文 → こいこいしない
	g.CPUHand = make([]Card, 3)
	yakus = []Yaku{{yakuAmeSikou, 7}}
	if CPUDecideKoiKoi(g, yakus, DifficultyNormal) {
		t.Error("Normal CPU should not koikoi with 7 points")
	}
}

func TestCPUDecideKoiKoiHard(t *testing.T) {
	g := NewGame(12)

	// 手札3枚、5文 → こいこいする
	g.CPUHand = make([]Card, 3)
	yakus := []Yaku{{"三光", 5}}
	if !CPUDecideKoiKoi(g, yakus, DifficultyHard) {
		t.Error("Hard CPU should koikoi with 3 cards and 5 points")
	}

	// 手札1枚 → こいこいしない
	g.CPUHand = make([]Card, 1)
	if CPUDecideKoiKoi(g, yakus, DifficultyHard) {
		t.Error("Hard CPU should not koikoi with 1 card")
	}

	// 手札3枚、10文 → こいこいしない
	g.CPUHand = make([]Card, 3)
	yakus = []Yaku{{"五光", 10}}
	if CPUDecideKoiKoi(g, yakus, DifficultyHard) {
		t.Error("Hard CPU should not koikoi with 10 points")
	}

	// 手札2枚、9文 → こいこいする
	g.CPUHand = make([]Card, 2)
	yakus = []Yaku{{yakuSikou, 8}, {"カス", 1}}
	if !CPUDecideKoiKoi(g, yakus, DifficultyHard) {
		t.Error("Hard CPU should koikoi with 2 cards and 9 points")
	}
}

func TestEvaluatePlay(t *testing.T) {
	// マッチなし
	if evaluatePlay(AllCards[0], nil) != 0 {
		t.Error("evaluatePlay with no matches should return 0")
	}

	// 光 + 光マッチ
	score := evaluatePlay(AllCards[0], []Card{AllCards[8]})
	if score != 40 { // 20 + 20
		t.Errorf("evaluatePlay = %d, want 40", score)
	}

	// カス + 種マッチ
	score = evaluatePlay(AllCards[2], []Card{AllCards[4]})
	if score != 11 { // 1 + 10
		t.Errorf("evaluatePlay = %d, want 11", score)
	}
}

func TestCardValue(t *testing.T) {
	tests := []struct {
		card Card
		want int
	}{
		{AllCards[0], 20}, // 光
		{AllCards[4], 10}, // 種
		{AllCards[1], 5},  // 短冊
		{AllCards[2], 1},  // カス
	}
	for _, tt := range tests {
		if got := cardValue(tt.card); got != tt.want {
			t.Errorf("cardValue(%s) = %d, want %d", tt.card.Name, got, tt.want)
		}
	}
}

func TestChooseBestMatch(t *testing.T) {
	// 光 vs 短冊
	matches := []Card{AllCards[1], AllCards[0]} // 短冊, 光
	best := chooseBestMatch(matches)
	if best.ID != 0 {
		t.Errorf("chooseBestMatch chose ID=%d, want 0 (hikari)", best.ID)
	}

	// 種 vs カス
	matches = []Card{AllCards[2], AllCards[4]} // カス, 種
	best = chooseBestMatch(matches)
	if best.ID != 4 {
		t.Errorf("chooseBestMatch chose ID=%d, want 4 (tane)", best.ID)
	}
}

func TestEvaluateYakuPotentialNoMatches(t *testing.T) {
	g := NewGame(12)
	g.CPUCaptured = nil
	bonus := evaluateYakuPotential(g, AllCards[0], nil)
	if bonus != 0 {
		t.Errorf("evaluateYakuPotential with no matches = %d, want 0", bonus)
	}
}

func TestEvaluateYakuPotentialHikari(t *testing.T) {
	g := NewGame(12)
	g.CPUCaptured = nil
	// 手札が光で、マッチ先にも光がある
	bonus := evaluateYakuPotential(g, AllCards[0], []Card{AllCards[8]}) // 松に鶴(光) + 桜に幕(光)
	// hand.Type==Hikari: +10, match Hikari: +15 = 25
	if bonus != 25 {
		t.Errorf("evaluateYakuPotential = %d, want 25", bonus)
	}
}

func TestEvaluateYakuPotentialInoshikacho(t *testing.T) {
	g := NewGame(12)
	g.CPUCaptured = nil
	// 猪鹿蝶の札がマッチにある
	bonus := evaluateYakuPotential(g, AllCards[26], []Card{AllCards[24]}) // 萩のカス + 萩に猪
	if bonus != 8 {
		t.Errorf("evaluateYakuPotential inoshikacho = %d, want 8", bonus)
	}
}

func TestEvaluateYakuPotentialInoshikachoAlreadyCaptured(t *testing.T) {
	g := NewGame(12)
	g.CPUCaptured = []Card{AllCards[24]} // 猪を既に持っている
	bonus := evaluateYakuPotential(g, AllCards[26], []Card{AllCards[24]})
	// 既に持っているのでボーナスなし
	if bonus != 0 {
		t.Errorf("evaluateYakuPotential already captured = %d, want 0", bonus)
	}
}

func TestEvaluateYakuPotentialAkatan(t *testing.T) {
	g := NewGame(12)
	g.CPUCaptured = nil
	// 赤短の札がマッチにある
	bonus := evaluateYakuPotential(g, AllCards[2], []Card{AllCards[1]}) // 松のカス + 松に赤短
	if bonus != 6 {
		t.Errorf("evaluateYakuPotential akatan = %d, want 6", bonus)
	}
}

func TestEvaluateYakuPotentialAotan(t *testing.T) {
	g := NewGame(12)
	g.CPUCaptured = nil
	// 青短の札がマッチにある
	bonus := evaluateYakuPotential(g, AllCards[22], []Card{AllCards[21]}) // 牡丹のカス + 牡丹に短冊(青短)
	if bonus != 6 {
		t.Errorf("evaluateYakuPotential aotan = %d, want 6", bonus)
	}
}

func TestCPUChooseHandCardEasyRandom(t *testing.T) {
	// Easy モードでランダム分岐を通ることを確認するため複数回実行
	g := NewGame(12)
	g.CPUHand = []Card{AllCards[0], AllCards[4], AllCards[8]}
	g.Field = []Card{AllCards[2], AllCards[6], AllCards[10]} // 各月1枚ずつマッチ

	for i := 0; i < 50; i++ {
		hand, _ := CPUChooseHandCard(g, DifficultyEasy)
		if hand.ID < 0 {
			t.Errorf("iteration %d: got invalid card ID %d", i, hand.ID)
		}
	}
}

func TestCPUChooseHandCardEasyTwoFieldMatch(t *testing.T) {
	// Easy のランダムパスで2枚マッチする場合
	g := NewGame(12)
	g.CPUHand = []Card{AllCards[2]}            // 松のカス1
	g.Field = []Card{AllCards[0], AllCards[1]} // 松に鶴, 松に赤短 (2枚マッチ)

	for i := 0; i < 50; i++ {
		hand, field := CPUChooseHandCard(g, DifficultyEasy)
		if hand.ID != 2 {
			t.Errorf("iteration %d: expected hand ID 2, got %d", i, hand.ID)
		}
		if field != nil {
			// ランダムパスで2枚マッチの場合 fieldCard が返される
			return
		}
	}
}
