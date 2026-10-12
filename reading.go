package haiku

import (
	"strings"

	"github.com/ikawaha/kagome-dict/dict"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

type token struct {
	Surface  string
	Features []string
	number   bool
	joined   bool
	yomi     string
}

// tokenize splits text into tokens without the ones to ignore. Symbols such
// as emoji are ignored when the dictionary has no reading for them and
// consecutive digits are joined into one number, because dictionaries differ
// in how they read them.
func tokenize(t *tokenizer.Tokenizer, d *dict.Dict, text string) []token {
	text = reIgnoreText.ReplaceAllString(text, " ")
	for {
		s := reDigitComma.ReplaceAllString(text, "$1$2")
		if s == text {
			break
		}
		text = s
	}

	var tokens []token
	joinable := false
	for _, tok := range t.Tokenize(text) {
		c := tok.Features()
		if isIgnore(d, c) {
			joinable = false
			continue
		}
		number := reDigits.MatchString(tok.Surface)
		if number && joinable {
			tokens[len(tokens)-1].Surface += tok.Surface
			tokens[len(tokens)-1].joined = true
			continue
		}
		next := token{Surface: tok.Surface, Features: c, number: number}
		if reSymbols.MatchString(tok.Surface) && !reWord.MatchString(reading(d, next)) {
			joinable = false
			continue
		}
		tokens = append(tokens, next)
		joinable = number
	}
	readCounters(d, tokens)
	return tokens
}

// reading returns the reading of the token in katakana.
func reading(d *dict.Dict, tok token) string {
	if tok.yomi != "" {
		return tok.yomi
	}
	if y, ok := ownNumberReading(d, tok); ok {
		return y
	}
	y := dictReading(d, tok)
	if reWord.MatchString(y) {
		return y
	}
	if reAcronym.MatchString(tok.Surface) {
		return acronymReading(tok.Surface)
	}
	if reHiragana.MatchString(tok.Surface) {
		return hiraganaToKatakana(tok.Surface)
	}
	return y
}

// hiraganaToKatakana reads hiragana which the dictionary does not know, such
// as colloquial spellings.
func hiraganaToKatakana(s string) string {
	return strings.Map(func(r rune) rune {
		if 'ぁ' <= r && r <= 'ゖ' {
			return r + 'ァ' - 'ぁ'
		}
		return r
	}, s)
}

// dictReading returns the reading of the token given by the dictionary.
func dictReading(d *dict.Dict, tok token) string {
	if reKana.MatchString(tok.Surface) {
		return tok.Surface
	}
	c := tok.Features
	if len(c) == 3 {
		return c[2]
	}
	idx := dictIdx(d, dict.PronunciationIndex)
	if idx >= 0 && idx < len(c) {
		return c[idx]
	}
	return tok.Surface
}

// ownNumberReading reads a number which the dictionary could not read as a
// whole: digits split into several tokens, or a number without reading.
func ownNumberReading(d *dict.Dict, tok token) (string, bool) {
	if !tok.number || (!tok.joined && reWord.MatchString(dictReading(d, tok))) {
		return "", false
	}
	return numberReading(tok.Surface)
}

// counterReadings has the readings of numbers followed by a counter, which
// differ in the count of morae from the plain reading of the number. The
// value is the reading of the number and of the counter.
var counterReadings = map[string]map[int][2]string{
	"人": {1: {"ヒト", "リ"}, 2: {"フタ", "リ"}},
	"日": {
		3: {"ミッ", "カ"}, 4: {"ヨッ", "カ"}, 6: {"ムイ", "カ"}, 7: {"ナノ", "カ"},
		8: {"ヨウ", "カ"}, 10: {"トオ", "カ"}, 14: {"ジュウヨッ", "カ"},
		20: {"ハツ", "カ"}, 24: {"ニジュウヨッ", "カ"},
	},
	"つ": {
		1: {"ヒト", "ツ"}, 2: {"フタ", "ツ"}, 3: {"ミッ", "ツ"}, 4: {"ヨッ", "ツ"},
		5: {"イツ", "ツ"}, 6: {"ムッ", "ツ"}, 7: {"ナナ", "ツ"}, 8: {"ヤッ", "ツ"},
		9: {"ココノ", "ツ"},
	},
	"月": {4: {"シ", ""}, 9: {"ク", ""}},
}

// readCounters fixes the readings of numbers written in digits when they are
// followed by a counter such as 時 or 人.
func readCounters(d *dict.Dict, tokens []token) {
	for i := 0; i+1 < len(tokens); i++ {
		if !tokens[i].number {
			continue
		}
		y := reading(d, tokens[i])
		n, ok := numberValue(tokens[i].Surface)
		if !ok {
			continue
		}
		counter := tokens[i+1].Surface
		if counter == "" {
			continue
		}
		if r, ok := counterReadings[counter][n]; ok {
			tokens[i].yomi = r[0]
			if r[1] != "" {
				tokens[i+1].yomi = r[1]
			}
			continue
		}
		head := []rune(counter)[0]
		switch {
		case strings.HasSuffix(y, "ヨン") && strings.ContainsRune("時年円人", head):
			tokens[i].yomi = strings.TrimSuffix(y, "ン")
		case strings.HasSuffix(y, "キュウ") && head == '時':
			tokens[i].yomi = strings.TrimSuffix(y, "キュウ") + "ク"
		}
	}
}

// numberValue returns the value of a number written in digits.
func numberValue(s string) (int, bool) {
	n := 0
	for _, r := range s {
		switch {
		case '0' <= r && r <= '9':
			n = n*10 + int(r-'0')
		case '０' <= r && r <= '９':
			n = n*10 + int(r-'０')
		default:
			return 0, false
		}
		if n > 1000000 {
			return 0, false
		}
	}
	return n, true
}

var (
	digitReadings = []string{"ゼロ", "イチ", "ニ", "サン", "ヨン", "ゴ", "ロク", "ナナ", "ハチ", "キュウ"}
	unitReadings  = []string{"", "ジュウ", "ヒャク", "セン"}
	largeReadings = []string{"", "マン", "オク", "チョウ", "ケイ"}

	letterReadings = map[rune]string{
		'A': "エー", 'B': "ビー", 'C': "シー", 'D': "ディー", 'E': "イー",
		'F': "エフ", 'G': "ジー", 'H': "エイチ", 'I': "アイ", 'J': "ジェー",
		'K': "ケー", 'L': "エル", 'M': "エム", 'N': "エヌ", 'O': "オー",
		'P': "ピー", 'Q': "キュー", 'R': "アール", 'S': "エス", 'T': "ティー",
		'U': "ユー", 'V': "ブイ", 'W': "ダブリュー", 'X': "エックス", 'Y': "ワイ",
		'Z': "ゼット",
	}
)

// numberReading returns the reading of a number written in digits. Sound
// changes inside a number such as ロッピャク are not applied since they do not
// change the count of morae. Counters are handled by readCounters.
func numberReading(s string) (string, bool) {
	digits := make([]int, 0, len(s))
	for _, r := range s {
		switch {
		case '0' <= r && r <= '9':
			digits = append(digits, int(r-'0'))
		case '０' <= r && r <= '９':
			digits = append(digits, int(r-'０'))
		default:
			return "", false
		}
	}
	if len(digits) == 0 {
		return "", false
	}
	// Read digit by digit when it does not look like a quantity
	if (len(digits) > 1 && digits[0] == 0) || len(digits) > 4*len(largeReadings) {
		var b strings.Builder
		for _, n := range digits {
			b.WriteString(digitReadings[n])
		}
		return b.String(), true
	}
	if len(digits) == 1 && digits[0] == 0 {
		return digitReadings[0], true
	}

	var b strings.Builder
	for i, n := range digits {
		place := len(digits) - 1 - i
		unit := place % 4
		if n != 0 {
			// 10, 100 and 1000 are read without イチ
			if n != 1 || unit == 0 {
				b.WriteString(digitReadings[n])
			}
			b.WriteString(unitReadings[unit])
		}
		if unit == 0 && place > 0 {
			from := i - 3
			if from < 0 {
				from = 0
			}
			for _, g := range digits[from : i+1] {
				if g != 0 {
					b.WriteString(largeReadings[place/4])
					break
				}
			}
		}
	}
	return b.String(), true
}

// acronymReading spells out an acronym such as NHK.
func acronymReading(s string) string {
	var b strings.Builder
	for _, r := range s {
		if 'Ａ' <= r && r <= 'Ｚ' {
			r = r - 'Ａ' + 'A'
		}
		b.WriteString(letterReadings[r])
	}
	return b.String()
}
