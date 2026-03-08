package main

import "testing"

func TestMonthString(t *testing.T) {
	expected := []string{"松", "梅", "桜", "藤", "菖蒲", "牡丹", "萩", "芒", "菊", "紅葉", "柳", "桐"}
	months := []Month{January, February, March, April, May, June, July, August, September, October, November, December}
	for i, m := range months {
		if got := m.String(); got != expected[i] {
			t.Errorf("Month(%d).String() = %q, want %q", m, got, expected[i])
		}
	}
}

func TestMonthOldName(t *testing.T) {
	expected := []string{"睦月", "如月", "弥生", "卯月", "皐月", "水無月", "文月", "葉月", "長月", "神無月", "霜月", "師走"}
	months := []Month{January, February, March, April, May, June, July, August, September, October, November, December}
	for i, m := range months {
		if got := m.OldName(); got != expected[i] {
			t.Errorf("Month(%d).OldName() = %q, want %q", m, got, expected[i])
		}
	}
}

func TestAllCardsCount(t *testing.T) {
	if len(AllCards) != 48 {
		t.Fatalf("AllCards has %d cards, want 48", len(AllCards))
	}
}

func TestAllCardsIDs(t *testing.T) {
	for i, c := range AllCards {
		if c.ID != i {
			t.Errorf("AllCards[%d].ID = %d, want %d", i, c.ID, i)
		}
	}
}

func TestAllCardsMonthDistribution(t *testing.T) {
	monthCount := make(map[Month]int)
	for _, c := range AllCards {
		monthCount[c.Month]++
	}
	for m := January; m <= December; m++ {
		if monthCount[m] != 4 {
			t.Errorf("Month %s has %d cards, want 4", m.String(), monthCount[m])
		}
	}
}

func TestCardDisplay(t *testing.T) {
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
		if got := tt.card.Display(); got != tt.want {
			t.Errorf("Card(%d).Display() = %q, want %q", tt.card.ID, got, tt.want)
		}
	}
}

func TestCardDisplayFull(t *testing.T) {
	tests := []struct {
		card Card
		want string
	}{
		{AllCards[0], "松に鶴"},
		{AllCards[8], "桜に幕"},
		{AllCards[40], "柳に小野道風"},
	}
	for _, tt := range tests {
		if got := tt.card.DisplayFull(); got != tt.want {
			t.Errorf("Card(%d).DisplayFull() = %q, want %q", tt.card.ID, got, tt.want)
		}
	}
}

func TestCardTypeConstants(t *testing.T) {
	if Kasu != 0 || Tane != 1 || Tanzaku != 2 || Hikari != 3 {
		t.Errorf("CardType constants: Kasu=%d Tane=%d Tanzaku=%d Hikari=%d", Kasu, Tane, Tanzaku, Hikari)
	}
}

func TestMonthConstants(t *testing.T) {
	if January != 0 || December != 11 {
		t.Errorf("Month constants: January=%d December=%d", January, December)
	}
}

func TestTypeSymbols(t *testing.T) {
	tests := []struct {
		ct   CardType
		want string
	}{
		{Hikari, "光"},
		{Tanzaku, "短"},
		{Tane, "種"},
		{Kasu, "カ"},
	}
	for _, tt := range tests {
		if got := typeSymbols[tt.ct]; got != tt.want {
			t.Errorf("typeSymbols[%d] = %q, want %q", tt.ct, got, tt.want)
		}
	}
}
