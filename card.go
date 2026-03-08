package main

// Month 月（花の種類）
type Month int

const (
	January   Month = iota // 松
	February               // 梅
	March                  // 桜
	April                  // 藤
	May                    // 菖蒲
	June                   // 牡丹
	July                   // 萩
	August                 // 芒
	September              // 菊
	October                // 紅葉
	November               // 柳
	December               // 桐
)

// CardType 札の種類
type CardType int

const (
	Kasu   CardType = iota // カス
	Tane                   // タネ
	Tanzaku                // 短冊
	Hikari                 // 光
)

// Card 花札1枚
type Card struct {
	ID    int
	Month Month
	Type  CardType
	Name  string
}

var monthNames = [12]string{
	"松", "梅", "桜", "藤", "菖蒲", "牡丹",
	"萩", "芒", "菊", "紅葉", "柳", "桐",
}

var oldMonthNames = [12]string{
	"睦月", "如月", "弥生", "卯月", "皐月", "水無月",
	"文月", "葉月", "長月", "神無月", "霜月", "師走",
}

func (m Month) String() string {
	return monthNames[m]
}

func (m Month) OldName() string {
	return oldMonthNames[m]
}

var typeSymbols = map[CardType]string{
	Hikari:  "光",
	Tanzaku: "短",
	Tane:    "種",
	Kasu:    "カ",
}

// AllCards 花札48枚の定義
var AllCards = []Card{
	// 1月 松
	{0, January, Hikari, "松に鶴"},
	{1, January, Tanzaku, "松に赤短"},
	{2, January, Kasu, "松のカス１"},
	{3, January, Kasu, "松のカス２"},
	// 2月 梅
	{4, February, Tane, "梅に鶯"},
	{5, February, Tanzaku, "梅に赤短"},
	{6, February, Kasu, "梅のカス１"},
	{7, February, Kasu, "梅のカス２"},
	// 3月 桜
	{8, March, Hikari, "桜に幕"},
	{9, March, Tanzaku, "桜に赤短"},
	{10, March, Kasu, "桜のカス１"},
	{11, March, Kasu, "桜のカス２"},
	// 4月 藤
	{12, April, Tane, "藤に不如帰"},
	{13, April, Tanzaku, "藤に短冊"},
	{14, April, Kasu, "藤のカス１"},
	{15, April, Kasu, "藤のカス２"},
	// 5月 菖蒲
	{16, May, Tane, "菖蒲に八橋"},
	{17, May, Tanzaku, "菖蒲に短冊"},
	{18, May, Kasu, "菖蒲のカス１"},
	{19, May, Kasu, "菖蒲のカス２"},
	// 6月 牡丹
	{20, June, Tane, "牡丹に蝶"},
	{21, June, Tanzaku, "牡丹に短冊"},
	{22, June, Kasu, "牡丹のカス１"},
	{23, June, Kasu, "牡丹のカス２"},
	// 7月 萩
	{24, July, Tane, "萩に猪"},
	{25, July, Tanzaku, "萩に短冊"},
	{26, July, Kasu, "萩のカス１"},
	{27, July, Kasu, "萩のカス２"},
	// 8月 芒
	{28, August, Hikari, "芒に月"},
	{29, August, Tane, "芒に雁"},
	{30, August, Kasu, "芒のカス１"},
	{31, August, Kasu, "芒のカス２"},
	// 9月 菊
	{32, September, Tane, "菊に盃"},
	{33, September, Tanzaku, "菊に短冊"},
	{34, September, Kasu, "菊のカス１"},
	{35, September, Kasu, "菊のカス２"},
	// 10月 紅葉
	{36, October, Tane, "紅葉に鹿"},
	{37, October, Tanzaku, "紅葉に短冊"},
	{38, October, Kasu, "紅葉のカス１"},
	{39, October, Kasu, "紅葉のカス２"},
	// 11月 柳
	{40, November, Hikari, "柳に小野道風"},
	{41, November, Tane, "柳に燕"},
	{42, November, Tanzaku, "柳に短冊"},
	{43, November, Kasu, "柳のカス"},
	// 12月 桐
	{44, December, Hikari, "桐に鳳凰"},
	{45, December, Kasu, "桐のカス１"},
	{46, December, Kasu, "桐のカス２"},
	{47, December, Kasu, "桐のカス３"},
}

// DisplayCard 札の表示用文字列
func (c Card) Display() string {
	sym := typeSymbols[c.Type]
	return "[" + c.Month.String() + ":" + sym + "]"
}

// DisplayFull 札の詳細表示
func (c Card) DisplayFull() string {
	return c.Name
}
