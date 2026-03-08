package main

import "testing"

const (
	yakuSankou   = "三光"
	yakuSikou    = "四光"
	yakuAmeSikou = "雨四光"
	yakuAkatan   = "赤短"
	yakuAotan    = "青短"
	yakuTane     = "タネ"
	yakuTan      = "タン"
)

func cardsFromIDList(ids ...int) []Card {
	var cards []Card
	for _, id := range ids {
		if id >= 0 && id < len(AllCards) {
			cards = append(cards, AllCards[id])
		}
	}
	return cards
}

func TestCheckYakuGokou(t *testing.T) {
	// 五光: 松に鶴(0), 桜に幕(8), 芒に月(28), 柳に小野道風(40), 桐に鳳凰(44)
	captured := cardsFromIDList(0, 8, 28, 40, 44)
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == "五光" && y.Points == 10 {
			found = true
		}
		// 五光がある場合、四光・雨四光・三光は出ない
		if y.Name == yakuSikou || y.Name == yakuAmeSikou || y.Name == yakuSankou {
			t.Errorf("五光と%sが同時に成立", y.Name)
		}
	}
	if !found {
		t.Error("五光が検出されない")
	}
}

func TestCheckYakuSikou(t *testing.T) {
	// 四光: 松に鶴(0), 桜に幕(8), 芒に月(28), 桐に鳳凰(44) (柳を除く光4枚)
	captured := cardsFromIDList(0, 8, 28, 44)
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == yakuSikou && y.Points == 8 {
			found = true
		}
	}
	if !found {
		t.Error("四光が検出されない")
	}
}

func TestCheckYakuAmeSikou(t *testing.T) {
	// 雨四光: 柳に小野道風(40)を含む光4枚
	captured := cardsFromIDList(0, 8, 28, 40)
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == yakuAmeSikou && y.Points == 7 {
			found = true
		}
	}
	if !found {
		t.Error("雨四光が検出されない")
	}
}

func TestCheckYakuSankou(t *testing.T) {
	// 三光: 柳以外の光3枚
	captured := cardsFromIDList(0, 8, 28)
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == yakuSankou && y.Points == 5 {
			found = true
		}
	}
	if !found {
		t.Error("三光が検出されない")
	}
}

func TestCheckYakuSankouWithYanagi(t *testing.T) {
	// 柳を含む光3枚では三光にならない
	captured := cardsFromIDList(0, 8, 40)
	yakus := CheckYaku(captured)
	for _, y := range yakus {
		if y.Name == yakuSankou {
			t.Error("柳を含む光3枚で三光が成立してしまった")
		}
	}
}

func TestCheckYakuInoshikacho(t *testing.T) {
	// 猪鹿蝶: 萩に猪(24), 紅葉に鹿(36), 牡丹に蝶(20)
	captured := cardsFromIDList(24, 36, 20)
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == "猪鹿蝶" && y.Points == 5 {
			found = true
		}
	}
	if !found {
		t.Error("猪鹿蝶が検出されない")
	}
}

func TestCheckYakuInoshikachoExtra(t *testing.T) {
	// 猪鹿蝶 + 種札1枚追加 = 6文
	captured := cardsFromIDList(24, 36, 20, 4) // 梅に鶯(4)追加
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == "猪鹿蝶" && y.Points == 6 {
			found = true
		}
	}
	if !found {
		t.Error("猪鹿蝶+種札で6文にならない")
	}
}

func TestCheckYakuHanamiDeIppai(t *testing.T) {
	// 花見で一杯: 桜に幕(8) + 菊に盃(32)
	captured := cardsFromIDList(8, 32)
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == "花見で一杯" && y.Points == 5 {
			found = true
		}
	}
	if !found {
		t.Error("花見で一杯が検出されない")
	}
}

func TestCheckYakuTsukimiDeIppai(t *testing.T) {
	// 月見で一杯: 芒に月(28) + 菊に盃(32)
	captured := cardsFromIDList(28, 32)
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == "月見で一杯" && y.Points == 5 {
			found = true
		}
	}
	if !found {
		t.Error("月見で一杯が検出されない")
	}
}

func TestCheckYakuAkatan(t *testing.T) {
	// 赤短: 松に赤短(1), 梅に赤短(5), 桜に赤短(9)
	captured := cardsFromIDList(1, 5, 9)
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == yakuAkatan && y.Points == 5 {
			found = true
		}
	}
	if !found {
		t.Error("赤短が検出されない")
	}
}

func TestCheckYakuAkatanExtra(t *testing.T) {
	// 赤短 + 短冊1枚追加 = 6文
	captured := cardsFromIDList(1, 5, 9, 13) // 藤に短冊(13)追加
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == yakuAkatan && y.Points == 6 {
			found = true
		}
	}
	if !found {
		t.Error("赤短+短冊で6文にならない")
	}
}

func TestCheckYakuAotan(t *testing.T) {
	// 青短: 牡丹に短冊(21), 菊に短冊(33), 紅葉に短冊(37)
	captured := cardsFromIDList(21, 33, 37)
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == yakuAotan && y.Points == 5 {
			found = true
		}
	}
	if !found {
		t.Error("青短が検出されない")
	}
}

func TestCheckYakuAotanExtra(t *testing.T) {
	// 青短 + 短冊1枚追加 = 6文
	captured := cardsFromIDList(21, 33, 37, 13)
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == yakuAotan && y.Points == 6 {
			found = true
		}
	}
	if !found {
		t.Error("青短+短冊で6文にならない")
	}
}

func TestCheckYakuAkatanAotanOverlap(t *testing.T) {
	// 赤短・青短の重複: 10文
	captured := cardsFromIDList(1, 5, 9, 21, 33, 37)
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == "赤短・青短の重複" && y.Points == 10 {
			found = true
		}
		if y.Name == yakuAkatan || y.Name == yakuAotan {
			t.Errorf("重複がある場合に%sが単独で成立してしまった", y.Name)
		}
	}
	if !found {
		t.Error("赤短・青短の重複が検出されない")
	}
}

func TestCheckYakuAkatanAotanOverlapExtra(t *testing.T) {
	// 赤短・青短の重複 + 短冊1枚 = 11文
	captured := cardsFromIDList(1, 5, 9, 21, 33, 37, 13)
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == "赤短・青短の重複" && y.Points == 11 {
			found = true
		}
	}
	if !found {
		t.Error("赤短・青短の重複+短冊で11文にならない")
	}
}

func TestCheckYakuTane(t *testing.T) {
	// タネ: 種札5枚以上で1文 (猪鹿蝶の3枚を含まないようにする)
	// 梅に鶯(4), 藤に不如帰(12), 菖蒲に八橋(16), 芒に雁(29), 菊に盃(32)
	captured := cardsFromIDList(4, 12, 16, 29, 32)
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == yakuTane && y.Points == 1 {
			found = true
		}
	}
	if !found {
		t.Error("タネが検出されない")
	}
}

func TestCheckYakuTaneExtra(t *testing.T) {
	// タネ: 種札6枚 = 2文
	captured := cardsFromIDList(4, 12, 16, 29, 32, 41) // 柳に燕(41)追加
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == yakuTane && y.Points == 2 {
			found = true
		}
	}
	if !found {
		t.Error("タネ6枚で2文にならない")
	}
}

func TestCheckYakuTaneDisabledByInoshikacho(t *testing.T) {
	// 猪鹿蝶がある場合、タネは無効
	// 猪(24), 鹿(36), 蝶(20), 梅に鶯(4), 藤に不如帰(12)
	captured := cardsFromIDList(24, 36, 20, 4, 12)
	yakus := CheckYaku(captured)
	for _, y := range yakus {
		if y.Name == yakuTane {
			t.Error("猪鹿蝶がある場合にタネが成立してしまった")
		}
	}
}

func TestCheckYakuTan(t *testing.T) {
	// タン: 短冊札5枚以上で1文 (赤短・青短の札を含まないようにする)
	// 藤に短冊(13), 菖蒲に短冊(17), 萩に短冊(25), 柳に短冊(42), 牡丹に短冊(21) → 青短成立するので除外
	// 藤に短冊(13), 菖蒲に短冊(17), 萩に短冊(25), 柳に短冊(42), 紅葉に短冊(37) → 青短ではない(37は青短)
	// 赤短も青短も成立しない5枚: 藤(13), 菖蒲(17), 萩(25), 柳(42) + もう1枚赤短・青短にならない = 牡丹(21)は青短の1枚
	// 正しくは: 13, 17, 25, 42 と、赤短でも青短でもない短冊 -> これは4枚しかない...
	// 短冊は全12枚: 1,5,9(赤短), 21,33,37(青短), 13,17,25,42 (普通)
	// 普通の短冊4枚+赤短1枚だと赤短は成立しない: 1, 13, 17, 25, 42
	captured := cardsFromIDList(13, 17, 25, 42, 1) // 赤短1枚だけでは赤短成立しない
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == yakuTan && y.Points == 1 {
			found = true
		}
	}
	if !found {
		t.Error("タンが検出されない")
	}
}

func TestCheckYakuTanDisabledByAkatan(t *testing.T) {
	// 赤短がある場合、タンは無効
	captured := cardsFromIDList(1, 5, 9, 13, 17)
	yakus := CheckYaku(captured)
	for _, y := range yakus {
		if y.Name == yakuTan {
			t.Error("赤短がある場合にタンが成立してしまった")
		}
	}
}

func TestCheckYakuTanDisabledByAotan(t *testing.T) {
	// 青短がある場合、タンは無効
	captured := cardsFromIDList(21, 33, 37, 13, 17)
	yakus := CheckYaku(captured)
	for _, y := range yakus {
		if y.Name == yakuTan {
			t.Error("青短がある場合にタンが成立してしまった")
		}
	}
}

func TestCheckYakuKasu(t *testing.T) {
	// カス: カス札10枚以上で1文
	kasuIDs := []int{2, 3, 6, 7, 10, 11, 14, 15, 18, 19}
	captured := cardsFromIDList(kasuIDs...)
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == "カス" && y.Points == 1 {
			found = true
		}
	}
	if !found {
		t.Error("カスが検出されない")
	}
}

func TestCheckYakuKasuExtra(t *testing.T) {
	// カス: カス札11枚 = 2文
	kasuIDs := []int{2, 3, 6, 7, 10, 11, 14, 15, 18, 19, 22}
	captured := cardsFromIDList(kasuIDs...)
	yakus := CheckYaku(captured)
	found := false
	for _, y := range yakus {
		if y.Name == "カス" && y.Points == 2 {
			found = true
		}
	}
	if !found {
		t.Error("カス11枚で2文にならない")
	}
}

func TestCheckYakuNoYaku(t *testing.T) {
	// 役なし
	captured := cardsFromIDList(2, 3, 6)
	yakus := CheckYaku(captured)
	if len(yakus) != 0 {
		t.Errorf("役なしのはずが%d個の役が検出された: %v", len(yakus), yakus)
	}
}

func TestCheckYakuEmpty(t *testing.T) {
	yakus := CheckYaku(nil)
	if len(yakus) != 0 {
		t.Errorf("空の獲得札で%d個の役が検出された", len(yakus))
	}
}

func TestTotalPoints(t *testing.T) {
	tests := []struct {
		yakus []Yaku
		want  int
	}{
		{nil, 0},
		{[]Yaku{}, 0},
		{[]Yaku{{"五光", 10}}, 10},
		{[]Yaku{{yakuAkatan, 5}, {"猪鹿蝶", 5}}, 10},
		{[]Yaku{{yakuSankou, 5}, {"カス", 1}, {yakuTane, 2}}, 8},
	}
	for _, tt := range tests {
		if got := TotalPoints(tt.yakus); got != tt.want {
			t.Errorf("TotalPoints(%v) = %d, want %d", tt.yakus, got, tt.want)
		}
	}
}

func TestFilterByType(t *testing.T) {
	cards := cardsFromIDList(0, 1, 2, 4, 8)
	hikari := filterByType(cards, Hikari)
	if len(hikari) != 2 { // 0=松に鶴(光), 8=桜に幕(光)
		t.Errorf("filterByType(Hikari) got %d cards, want 2", len(hikari))
	}
	tanzaku := filterByType(cards, Tanzaku)
	if len(tanzaku) != 1 { // 1=松に赤短
		t.Errorf("filterByType(Tanzaku) got %d cards, want 1", len(tanzaku))
	}
	kasu := filterByType(cards, Kasu)
	if len(kasu) != 1 { // 2=松のカス1
		t.Errorf("filterByType(Kasu) got %d cards, want 1", len(kasu))
	}
	tane := filterByType(cards, Tane)
	if len(tane) != 1 { // 4=梅に鶯
		t.Errorf("filterByType(Tane) got %d cards, want 1", len(tane))
	}
}

func TestCardIDs(t *testing.T) {
	cards := cardsFromIDList(0, 8, 28)
	ids := cardIDs(cards)
	if len(ids) != 3 {
		t.Fatalf("cardIDs() got %d IDs, want 3", len(ids))
	}
	expected := []int{0, 8, 28}
	for i, id := range ids {
		if id != expected[i] {
			t.Errorf("cardIDs()[%d] = %d, want %d", i, id, expected[i])
		}
	}
}

func TestContainsAll(t *testing.T) {
	ids := []int{0, 8, 28, 40, 44}
	if !containsAll(ids, 0, 8, 28) {
		t.Error("containsAll should return true")
	}
	if !containsAll(ids, 0, 8, 28, 40, 44) {
		t.Error("containsAll should return true for all IDs")
	}
	if containsAll(ids, 0, 8, 99) {
		t.Error("containsAll should return false when ID not present")
	}
	if !containsAll(ids) {
		t.Error("containsAll with no targets should return true")
	}
}

func TestExtraPoints(t *testing.T) {
	tests := []struct {
		count, base, want int
	}{
		{6, 6, 0},
		{7, 6, 1},
		{3, 3, 0},
		{5, 3, 2},
		{2, 3, 0}, // count < base
		{0, 6, 0},
	}
	for _, tt := range tests {
		if got := extraPoints(tt.count, tt.base); got != tt.want {
			t.Errorf("extraPoints(%d, %d) = %d, want %d", tt.count, tt.base, got, tt.want)
		}
	}
}

func TestContains(t *testing.T) {
	ids := []int{0, 8, 28}
	if !contains(ids, 0) {
		t.Error("contains(0) should return true")
	}
	if !contains(ids, 28) {
		t.Error("contains(28) should return true")
	}
	if contains(ids, 99) {
		t.Error("contains(99) should return false")
	}
	if contains(nil, 0) {
		t.Error("contains(nil, 0) should return false")
	}
}

// 光の排他性テスト: 雨四光は柳を含む光4枚だが、四光(柳除く光4枚)が優先
func TestCheckYakuHikariExclusivity(t *testing.T) {
	// card IDs 0, 8, 28, 44 form sikou (four hikari excluding yanagi)
	captured := cardsFromIDList(0, 8, 28, 44)
	yakus := CheckYaku(captured)
	for _, y := range yakus {
		if y.Name == yakuAmeSikou || y.Name == yakuSankou {
			t.Errorf("四光がある場合に%sが成立してしまった", y.Name)
		}
	}
}

// 複合役テスト
func TestCheckYakuMultiple(t *testing.T) {
	// 三光(0,8,28) + 花見で一杯(8,32) + 月見で一杯(28,32)
	captured := cardsFromIDList(0, 8, 28, 32)
	yakus := CheckYaku(captured)
	names := make(map[string]bool)
	for _, y := range yakus {
		names[y.Name] = true
	}
	if !names[yakuSankou] {
		t.Error("三光が検出されない")
	}
	if !names["花見で一杯"] {
		t.Error("花見で一杯が検出されない")
	}
	if !names["月見で一杯"] {
		t.Error("月見で一杯が検出されない")
	}
}
