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
		yakus = append(yakus, Yaku{"赤短・青短の重複", 10 + extraPoints(len(tanzaku), 6)})
	case hasAkatan:
		yakus = append(yakus, Yaku{"赤短", 5 + extraPoints(len(tanzaku), 3)})
	case hasAotan:
		yakus = append(yakus, Yaku{"青短", 5 + extraPoints(len(tanzaku), 3)})
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

// YakuReach リーチ中の役（あと1枚で成立）
type YakuReach struct {
	Name    string
	Missing []Card // 不足している札 (nil = 同種札あと1枚)
}

// CheckReach 獲得札からリーチ中の役を判定する
func CheckReach(captured []Card) []YakuReach {
	var reaches []YakuReach

	hikari := filterByType(captured, Hikari)
	tane := filterByType(captured, Tane)
	tanzaku := filterByType(captured, Tanzaku)
	kasu := filterByType(captured, Kasu)

	hikariIDs := cardIDs(hikari)
	taneIDs := cardIDs(tane)
	tanzakuIDs := cardIDs(tanzaku)
	allIDs := cardIDs(captured)

	existingYaku := CheckYaku(captured)
	hasYaku := func(name string) bool {
		for _, y := range existingYaku {
			if y.Name == name {
				return true
			}
		}
		return false
	}

	missingCards := func(required []int, have []int) []Card {
		var cards []Card
		for _, r := range required {
			if !contains(have, r) {
				cards = append(cards, AllCards[r])
			}
		}
		return cards
	}

	// --- 光系 ---
	allHikari := []int{0, 8, 28, 40, 44}
	noYanagiHikari := []int{0, 8, 28, 44}

	hikariNoYanagi := 0
	for _, id := range hikariIDs {
		if id != 40 {
			hikariNoYanagi++
		}
	}
	hasYanagi := contains(hikariIDs, 40)

	// 五光リーチ (光札4枚)
	if !hasYaku("五光") && len(hikari) == 4 {
		if m := missingCards(allHikari, hikariIDs); len(m) == 1 {
			reaches = append(reaches, YakuReach{"五光", m})
		}
	}

	// 四光リーチ (柳以外の光3枚、柳なし)
	if !hasYaku("四光") && !hasYaku("五光") && hikariNoYanagi == 3 && !hasYanagi {
		if m := missingCards(noYanagiHikari, hikariIDs); len(m) == 1 {
			reaches = append(reaches, YakuReach{"四光", m})
		}
	}

	// 雨四光リーチ
	if !hasYaku("雨四光") && !hasYaku("四光") && !hasYaku("五光") {
		if hasYanagi && hikariNoYanagi == 2 {
			// 柳あり + 柳以外2枚 → 柳以外の光あと1枚
			m := missingCards(noYanagiHikari, hikariIDs)
			reaches = append(reaches, YakuReach{"雨四光", m})
		} else if !hasYanagi && hikariNoYanagi == 3 {
			// 三光成立中 → 柳を取れば雨四光
			reaches = append(reaches, YakuReach{"雨四光", []Card{AllCards[40]}})
		}
	}

	// 三光リーチ (柳以外の光2枚、柳なし)
	if !hasYaku("三光") && !hasYaku("雨四光") && !hasYaku("四光") && !hasYaku("五光") && hikariNoYanagi == 2 && !hasYanagi {
		m := missingCards(noYanagiHikari, hikariIDs)
		reaches = append(reaches, YakuReach{"三光", m})
	}

	// --- 猪鹿蝶 ---
	inoshikacho := []int{20, 24, 36}
	if !hasYaku("猪鹿蝶") {
		if m := missingCards(inoshikacho, taneIDs); len(m) == 1 {
			reaches = append(reaches, YakuReach{"猪鹿蝶", m})
		}
	}

	// --- 花見で一杯 ---
	hanami := []int{8, 32}
	if !hasYaku("花見で一杯") {
		if m := missingCards(hanami, allIDs); len(m) == 1 {
			reaches = append(reaches, YakuReach{"花見で一杯", m})
		}
	}

	// --- 月見で一杯 ---
	tsukimi := []int{28, 32}
	if !hasYaku("月見で一杯") {
		if m := missingCards(tsukimi, allIDs); len(m) == 1 {
			reaches = append(reaches, YakuReach{"月見で一杯", m})
		}
	}

	// --- 短冊系 ---
	akatan := []int{1, 5, 9}
	aotan := []int{21, 33, 37}
	akatanDone := hasYaku("赤短") || hasYaku("赤短・青短の重複")
	aotanDone := hasYaku("青短") || hasYaku("赤短・青短の重複")
	bothDone := hasYaku("赤短・青短の重複")

	if !bothDone {
		akatanMissing := missingCards(akatan, tanzakuIDs)
		aotanMissing := missingCards(aotan, tanzakuIDs)
		akatanReach := len(akatanMissing) == 1
		aotanReach := len(aotanMissing) == 1

		switch {
		case akatanDone && aotanReach:
			reaches = append(reaches, YakuReach{"赤短・青短の重複", aotanMissing})
		case aotanDone && akatanReach:
			reaches = append(reaches, YakuReach{"赤短・青短の重複", akatanMissing})
		default:
			if !akatanDone && akatanReach {
				reaches = append(reaches, YakuReach{"赤短", akatanMissing})
			}
			if !aotanDone && aotanReach {
				reaches = append(reaches, YakuReach{"青短", aotanMissing})
			}
		}
	}

	// --- タネ (5枚、猪鹿蝶未成立時のみ) ---
	if !hasYaku("タネ") && !hasYaku("猪鹿蝶") && len(tane) == 4 {
		reaches = append(reaches, YakuReach{"タネ", nil})
	}

	// --- タン (5枚、赤短/青短未成立時のみ) ---
	if !hasYaku("タン") && !akatanDone && !aotanDone && len(tanzaku) == 4 {
		reaches = append(reaches, YakuReach{"タン", nil})
	}

	// --- カス (10枚) ---
	if !hasYaku("カス") && len(kasu) == 9 {
		reaches = append(reaches, YakuReach{"カス", nil})
	}

	return reaches
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

func extraPoints(count, base int) int {
	return max(count-base, 0)
}
