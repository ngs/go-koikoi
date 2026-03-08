package main

import (
	"math/rand"
)

// Game ゲーム状態
type Game struct {
	Deck           []Card
	Field          []Card
	PlayerHand     []Card
	CPUHand        []Card
	PlayerCaptured []Card
	CPUCaptured    []Card
	PlayerScore    int
	CPUScore       int
	Round          int
	MaxRounds      int
	IsPlayerTurn   bool
	// 前回チェック時の役（こいこい判定用）
	PlayerPrevYaku []Yaku
	CPUPrevYaku    []Yaku
	// こいこい宣言フラグ（ラウンド内）
	PlayerKoiKoi bool
	CPUKoiKoi    bool
	// 次ラウンドの親（true=プレイヤー）
	NextParentIsPlayer bool
}

// NewGame 新しいゲームを開始
func NewGame(rounds int) *Game {
	g := &Game{
		MaxRounds:          rounds,
		Round:              1,
		NextParentIsPlayer: true, // 初回はプレイヤーが親
	}
	return g
}

// StartRound ラウンド開始：シャッフルして配る
func (g *Game) StartRound() {
	// デッキ作成・シャッフル
	g.Deck = make([]Card, len(AllCards))
	copy(g.Deck, AllCards)
	rand.Shuffle(len(g.Deck), func(i, j int) {
		g.Deck[i], g.Deck[j] = g.Deck[j], g.Deck[i]
	})

	g.PlayerHand = nil
	g.CPUHand = nil
	g.Field = nil
	g.PlayerCaptured = nil
	g.CPUCaptured = nil
	g.PlayerPrevYaku = nil
	g.CPUPrevYaku = nil
	g.PlayerKoiKoi = false
	g.CPUKoiKoi = false

	// 配札: 手札8枚ずつ、場札8枚
	for i := 0; i < 8; i++ {
		g.PlayerHand = append(g.PlayerHand, g.draw())
		g.CPUHand = append(g.CPUHand, g.draw())
	}
	for i := 0; i < 8; i++ {
		g.Field = append(g.Field, g.draw())
	}

	// 前ラウンドの勝者が親（先攻）
	g.IsPlayerTurn = g.NextParentIsPlayer
}

func (g *Game) draw() Card {
	c := g.Deck[0]
	g.Deck = g.Deck[1:]
	return c
}

// DrawFromDeck 山札から1枚引く
func (g *Game) DrawFromDeck() (Card, bool) {
	if len(g.Deck) == 0 {
		return Card{}, false
	}
	return g.draw(), true
}

// MatchingFieldCards 場に出ている同月の札を返す
func (g *Game) MatchingFieldCards(card Card) []Card {
	var matches []Card
	for _, f := range g.Field {
		if f.Month == card.Month {
			matches = append(matches, f)
		}
	}
	return matches
}

// PlayCard 手札から場に出してマッチング処理を行う
// fieldChoice: 場札の中から選んだ札（マッチが複数ある場合）
// 戻り値: 獲得した札
func (g *Game) PlayCard(handCard Card, fieldChoice *Card, isPlayer bool) []Card {
	// 手札から除去
	if isPlayer {
		g.PlayerHand = removeCard(g.PlayerHand, handCard)
	} else {
		g.CPUHand = removeCard(g.CPUHand, handCard)
	}

	matches := g.MatchingFieldCards(handCard)

	var captured []Card
	if len(matches) == 0 {
		// マッチなし → 場に置く
		g.Field = append(g.Field, handCard)
	} else if len(matches) == 1 {
		// 1枚マッチ → 獲得
		captured = append(captured, handCard, matches[0])
		g.Field = removeCard(g.Field, matches[0])
	} else if len(matches) == 2 && fieldChoice != nil {
		// 2枚マッチ → 選択した1枚を獲得
		captured = append(captured, handCard, *fieldChoice)
		g.Field = removeCard(g.Field, *fieldChoice)
	} else if len(matches) == 3 {
		// 3枚マッチ → 全て獲得
		captured = append(captured, handCard)
		for _, m := range matches {
			captured = append(captured, m)
			g.Field = removeCard(g.Field, m)
		}
	} else if len(matches) == 2 {
		// 2枚マッチでfieldChoice未指定 → 最初の1枚
		captured = append(captured, handCard, matches[0])
		g.Field = removeCard(g.Field, matches[0])
	}

	if isPlayer {
		g.PlayerCaptured = append(g.PlayerCaptured, captured...)
	} else {
		g.CPUCaptured = append(g.CPUCaptured, captured...)
	}

	return captured
}

// PlayDrawnCard 山札から引いた札を場に合わせる
func (g *Game) PlayDrawnCard(drawnCard Card, fieldChoice *Card, isPlayer bool) []Card {
	matches := g.MatchingFieldCards(drawnCard)

	var captured []Card
	if len(matches) == 0 {
		g.Field = append(g.Field, drawnCard)
	} else if len(matches) == 1 {
		captured = append(captured, drawnCard, matches[0])
		g.Field = removeCard(g.Field, matches[0])
	} else if len(matches) == 2 && fieldChoice != nil {
		captured = append(captured, drawnCard, *fieldChoice)
		g.Field = removeCard(g.Field, *fieldChoice)
	} else if len(matches) == 3 {
		captured = append(captured, drawnCard)
		for _, m := range matches {
			captured = append(captured, m)
			g.Field = removeCard(g.Field, m)
		}
	} else if len(matches) == 2 {
		captured = append(captured, drawnCard, matches[0])
		g.Field = removeCard(g.Field, matches[0])
	}

	if isPlayer {
		g.PlayerCaptured = append(g.PlayerCaptured, captured...)
	} else {
		g.CPUCaptured = append(g.CPUCaptured, captured...)
	}

	return captured
}

// CheckNewYaku 新しく成立した役があるかチェック
func (g *Game) CheckNewYaku(isPlayer bool) []Yaku {
	var captured []Card
	var prevYaku []Yaku
	if isPlayer {
		captured = g.PlayerCaptured
		prevYaku = g.PlayerPrevYaku
	} else {
		captured = g.CPUCaptured
		prevYaku = g.CPUPrevYaku
	}

	currentYaku := CheckYaku(captured)

	// 前回との差分を取る
	var newYaku []Yaku
	for _, cy := range currentYaku {
		found := false
		for _, py := range prevYaku {
			if cy.Name == py.Name && cy.Points == py.Points {
				found = true
				break
			}
		}
		if !found {
			newYaku = append(newYaku, cy)
		}
	}

	return newYaku
}

// UpdatePrevYaku 前回の役を更新（こいこい後）
func (g *Game) UpdatePrevYaku(isPlayer bool) {
	if isPlayer {
		g.PlayerPrevYaku = CheckYaku(g.PlayerCaptured)
	} else {
		g.CPUPrevYaku = CheckYaku(g.CPUCaptured)
	}
}

// CalcFinalScore 最終得点を計算する（7文以上2倍、こいこいペナルティ）
func (g *Game) CalcFinalScore(isPlayer bool) int {
	var captured []Card
	var opponentKoiKoi bool
	if isPlayer {
		captured = g.PlayerCaptured
		opponentKoiKoi = g.CPUKoiKoi
	} else {
		captured = g.CPUCaptured
		opponentKoiKoi = g.PlayerKoiKoi
	}

	yakus := CheckYaku(captured)
	total := TotalPoints(yakus)

	// 7文以上で2倍
	if total >= 7 {
		total *= 2
	}

	// 相手がこいこいしていた場合、さらに2倍
	if opponentKoiKoi {
		total *= 2
	}

	return total
}

// IsRoundOver 手札がなくなったか
func (g *Game) IsRoundOver() bool {
	return len(g.PlayerHand) == 0 && len(g.CPUHand) == 0
}

func removeCard(cards []Card, target Card) []Card {
	for i, c := range cards {
		if c.ID == target.ID {
			return append(cards[:i], cards[i+1:]...)
		}
	}
	return cards
}
