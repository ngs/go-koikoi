package main

// CPUChooseHandCard CPUが出す手札を選ぶ
func CPUChooseHandCard(g *Game, diff Difficulty) (handCard Card, fieldCard *Card) {
	// Easy: 一定確率でランダムに選ぶ
	if diff == DifficultyEasy && cryptoIntn(3) == 0 {
		idx := cryptoIntn(len(g.CPUHand))
		hand := g.CPUHand[idx]
		matches := g.MatchingFieldCards(hand)
		if len(matches) == 2 {
			chosen := matches[cryptoIntn(2)]
			return hand, &chosen
		}
		return hand, nil
	}

	bestScore := -1
	bestHandIdx := 0
	var bestFieldCard *Card

	for i, hand := range g.CPUHand {
		matches := g.MatchingFieldCards(hand)
		score := evaluatePlay(hand, matches)

		// Hard: 役に近い札をボーナス評価
		if diff == DifficultyHard {
			score += evaluateYakuPotential(g, hand, matches)
		}

		if score > bestScore {
			bestScore = score
			bestHandIdx = i
			bestFieldCard = nil
			if len(matches) == 2 {
				best := chooseBestMatch(matches)
				bestFieldCard = &best
			}
		}
	}

	handCard = g.CPUHand[bestHandIdx]

	// マッチがない場合、一番価値の低い札を捨てる
	if len(g.MatchingFieldCards(handCard)) == 0 {
		lowestIdx := 0
		lowestVal := cardValue(g.CPUHand[0])
		for i, c := range g.CPUHand {
			v := cardValue(c)
			if v < lowestVal {
				lowestVal = v
				lowestIdx = i
			}
		}
		return g.CPUHand[lowestIdx], nil
	}

	return handCard, bestFieldCard
}

// CPUChooseFieldCard 山札から引いた札に対して場札を選ぶ
func CPUChooseFieldCard(matches []Card) *Card {
	if len(matches) == 2 {
		best := chooseBestMatch(matches)
		return &best
	}
	return nil
}

// CPUDecideKoiKoi こいこいするかどうか判断
func CPUDecideKoiKoi(g *Game, yakus []Yaku, diff Difficulty) bool {
	points := TotalPoints(yakus)

	switch diff {
	case DifficultyEasy:
		// かんたん: 常に勝負（こいこいしない）
		return false
	case DifficultyHard:
		// つよい: 手札に余裕があれば積極的にこいこい
		if len(g.CPUHand) <= 1 {
			return false
		}
		if points >= 10 {
			return false
		}
		return true
	default:
		// ふつう: 従来ロジック
		if len(g.CPUHand) <= 2 {
			return false
		}
		if points >= 7 {
			return false
		}
		return true
	}
}

// evaluateYakuPotential Hard モード用: 役への近さを評価
func evaluateYakuPotential(g *Game, hand Card, matches []Card) int {
	if len(matches) == 0 {
		return 0
	}
	bonus := 0
	cpuIDs := cardIDs(g.CPUCaptured)

	// 光札を取れるなら大きなボーナス
	for _, m := range matches {
		if m.Type == Hikari {
			bonus += 15
		}
	}
	if hand.Type == Hikari {
		bonus += 10
	}

	// 猪鹿蝶に近い札
	inoShiKa := []int{24, 36, 20}
	for _, m := range matches {
		for _, id := range inoShiKa {
			if m.ID == id && !contains(cpuIDs, id) {
				bonus += 8
			}
		}
	}

	// 赤短・青短に近い札
	akatan := []int{1, 5, 9}
	aotan := []int{21, 33, 37}
	for _, m := range matches {
		for _, id := range akatan {
			if m.ID == id && !contains(cpuIDs, id) {
				bonus += 6
			}
		}
		for _, id := range aotan {
			if m.ID == id && !contains(cpuIDs, id) {
				bonus += 6
			}
		}
	}

	return bonus
}

func evaluatePlay(hand Card, matches []Card) int {
	if len(matches) == 0 {
		return 0
	}
	score := cardValue(hand)
	for _, m := range matches {
		score += cardValue(m)
	}
	return score
}

func cardValue(c Card) int {
	switch c.Type {
	case Hikari:
		return 20
	case Tane:
		return 10
	case Tanzaku:
		return 5
	default:
		return 1
	}
}

func chooseBestMatch(matches []Card) Card {
	best := matches[0]
	bestVal := cardValue(matches[0])
	for _, m := range matches[1:] {
		v := cardValue(m)
		if v > bestVal {
			bestVal = v
			best = m
		}
	}
	return best
}
