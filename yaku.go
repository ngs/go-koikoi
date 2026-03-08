package main

// Yaku 役
type Yaku struct {
	Name   string
	Points int
}

// CheckYaku 獲得札から成立している役を判定する（任天堂ルール準拠）
func CheckYaku(captured []Card) []Yaku {
	var yakus []Yaku

	hikari := filterByType(captured, Hikari)
	tane := filterByType(captured, Tane)
	tanzaku := filterByType(captured, Tanzaku)
	kasu := filterByType(captured, Kasu)

	// 光系（排他: 五光 > 四光 > 雨四光 > 三光）
	hikariIDs := cardIDs(hikari)
	hikariNoYanagi := 0
	for _, id := range hikariIDs {
		if id != 40 { // 柳に小野道風(40)以外
			hikariNoYanagi++
		}
	}

	hasGokou := containsAll(hikariIDs, 0, 8, 28, 40, 44)
	hasSikou := !hasGokou && containsAll(hikariIDs, 0, 8, 28, 44) // 柳を除く光4枚
	hasAmeSikou := !hasGokou && !hasSikou && contains(hikariIDs, 40) && len(hikari) >= 4
	hasSankou := !hasGokou && !hasSikou && !hasAmeSikou && hikariNoYanagi >= 3

	switch {
	case hasGokou:
		yakus = append(yakus, Yaku{"五光", 10})
	case hasSikou:
		yakus = append(yakus, Yaku{"四光", 8})
	case hasAmeSikou:
		yakus = append(yakus, Yaku{"雨四光", 7})
	case hasSankou:
		yakus = append(yakus, Yaku{"三光", 5})
	}

	// 猪鹿蝶: 5文 + 種札が増えるごとに+1文
	taneIDs := cardIDs(tane)
	hasInoshikacho := containsAll(taneIDs, 24, 36, 20) // 萩に猪, 紅葉に鹿, 牡丹に蝶
	if hasInoshikacho {
		extra := len(tane) - 3
		yakus = append(yakus, Yaku{"猪鹿蝶", 5 + extra})
	}

	// 花見で一杯: 桜に幕(8) + 菊に盃(32) → 5文
	allIDs := cardIDs(captured)
	if containsAll(allIDs, 8, 32) {
		yakus = append(yakus, Yaku{"花見で一杯", 5})
	}

	// 月見で一杯: 芒に月(28) + 菊に盃(32) → 5文
	if containsAll(allIDs, 28, 32) {
		yakus = append(yakus, Yaku{"月見で一杯", 5})
	}

	// 短冊系（排他: 赤短・青短の重複 > 赤短/青短 > タン）
	tanzakuIDs := cardIDs(tanzaku)
	hasAkatan := containsAll(tanzakuIDs, 1, 5, 9)   // 松・梅・桜の赤短
	hasAotan := containsAll(tanzakuIDs, 21, 33, 37) // 牡丹・菊・紅葉の青短

	switch {
	case hasAkatan && hasAotan:
		// 赤短・青短の重複: 10文 + 短冊が増えるごとに+1文
		extra := len(tanzaku) - 6
		if extra < 0 {
			extra = 0
		}
		yakus = append(yakus, Yaku{"赤短・青短の重複", 10 + extra})
	case hasAkatan:
		extra := len(tanzaku) - 3
		if extra < 0 {
			extra = 0
		}
		yakus = append(yakus, Yaku{"赤短", 5 + extra})
	case hasAotan:
		extra := len(tanzaku) - 3
		if extra < 0 {
			extra = 0
		}
		yakus = append(yakus, Yaku{"青短", 5 + extra})
	}

	// タネ: 種札5枚以上で1文 + 1枚ごとに+1文（猪鹿蝶ができたら無効）
	if !hasInoshikacho && len(tane) >= 5 {
		yakus = append(yakus, Yaku{"タネ", 1 + (len(tane) - 5)})
	}

	// タン: 短冊札5枚以上で1文 + 1枚ごとに+1文（赤短か青短ができたら無効）
	if !hasAkatan && !hasAotan && len(tanzaku) >= 5 {
		yakus = append(yakus, Yaku{"タン", 1 + (len(tanzaku) - 5)})
	}

	// カス: カス札10枚以上で1文 + 1枚ごとに+1文
	if len(kasu) >= 10 {
		yakus = append(yakus, Yaku{"カス", 1 + (len(kasu) - 10)})
	}

	return yakus
}

// TotalPoints 役の合計点を返す
func TotalPoints(yakus []Yaku) int {
	total := 0
	for _, y := range yakus {
		total += y.Points
	}
	return total
}

func filterByType(cards []Card, t CardType) []Card {
	var result []Card
	for _, c := range cards {
		if c.Type == t {
			result = append(result, c)
		}
	}
	return result
}

func cardIDs(cards []Card) []int {
	ids := make([]int, len(cards))
	for i, c := range cards {
		ids[i] = c.ID
	}
	return ids
}

func containsAll(ids []int, targets ...int) bool {
	for _, t := range targets {
		if !contains(ids, t) {
			return false
		}
	}
	return true
}

func contains(ids []int, target int) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}
