package haiku

import (
	"regexp"
	"strings"

	"github.com/ikawaha/kagome-dict/dict"
	"github.com/ikawaha/kagome-dict/uni"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

var (
	reWord       = regexp.MustCompile(`^[ァ-ヾ]+$`)
	reIgnoreText = regexp.MustCompile(`[\[\]「」『』、。]`)
	reIgnoreChar = regexp.MustCompile(`[ァィゥェォャュョ]`)
	reKana       = regexp.MustCompile(`[ァ-タダ-ヶ]`)
)

type Opt struct {
	Udic *dict.Dict
}

func isEnd(c []string) bool {
	return c[1] != "非自立" && !strings.HasPrefix(c[5], "連用") && c[5] != "未然形"
}

func isSpace(c []string) bool {
	return c[0] == "空白"
}

// isWord return true when the kind of the word is possible to be leading of
// sentence.
func isWord(c []string) bool {
	for _, f := range []string{"名詞", "形容詞", "形容動詞", "副詞", "連体詞", "接続詞", "感動詞", "接頭詞", "フィラー"} {
		if f == c[0] && c[1] != "非自立" && c[1] != "接尾" {
			return true
		}
	}
	if c[0] == "代名詞" {
		return true
	}
	if c[0] == "記号" && c[1] == "一般" {
		return true
	}
	if c[0] == "助詞" && c[1] != "副助詞" {
		return true
	}
	if c[0] == "動詞" && c[1] != "接尾" {
		return true
	}
	if c[0] == "カスタム人名" || c[0] == "カスタム名詞" {
		return true
	}
	return false
}

// countChars return count of characters with ignoring japanese small letters.
func countChars(s string) int {
	return len([]rune(reIgnoreChar.ReplaceAllString(s, "")))
}

// Match return true when text matches with rule(s).
func Match(text string, rule []int) bool {
	return MatchWithOpt(text, rule, nil)
}

// MatchWithOpt return true when text matches with rule(s).
func MatchWithOpt(text string, rule []int, opt *Opt) bool {
	if len(rule) == 0 {
		return false
	}
	d := opt.Udic
	if d == nil {
		d = uni.Dict()
	}
	t, err := tokenizer.New(d, tokenizer.OmitBosEos())
	if err != nil {
		return false
	}
	text = reIgnoreText.ReplaceAllString(text, " ")
	tokens := t.Tokenize(text)
	pos := 0
	r := make([]int, len(rule))
	copy(r, rule)

	var tmp []tokenizer.Token
	for _, token := range tokens {
		c := token.Features()
		if len(c) > 0 && c[0] != "空白" {
			tmp = append(tmp, token)
		}
	}
	tokens = tmp

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		c := tok.Features()
		if len(c) < 7 {
			return false
		}
		var y string
		if len(c) < 10 {
			y = c[6]
		} else {
			y = c[9]
		}
		if y == "*" {
			y = tok.Surface
		}
		if !reWord.MatchString(y) {
			if y == "、" {
				continue
			}
			return false
		}
		if pos >= len(rule) || (r[pos] == rule[pos] && !isWord(c)) {
			return false
		}
		n := countChars(y)
		r[pos] -= n
		if r[pos] == 0 {
			pos++
			if pos == len(r) && i == len(tokens)-1 {
				return true
			}
		}
	}
	return false
}

func FindWithOpt(text string, rule []int, opt *Opt) ([]string, error) {
	if len(rule) == 0 {
		return nil, nil
	}
	d := opt.Udic
	if d == nil {
		d = uni.Dict()
	}
	t, err := tokenizer.New(d, tokenizer.OmitBosEos())
	if err != nil {
		return nil, err
	}
	text = reIgnoreText.ReplaceAllString(text, " ")
	tokens := t.Tokenize(text)
	pos := 0
	r := make([]int, len(rule))
	copy(r, rule)
	sentence := ""
	start := 0
	ambigous := 0

	var tmp []tokenizer.Token
	for _, token := range tokens {
		c := token.Features()
		if len(c) > 0 && c[0] != "空白" {
			tmp = append(tmp, token)
		}
	}
	tokens = tmp

	for i := 0; i < len(tokens); i++ {
		if reKana.MatchString(tokens[i].Surface) {
			surface := tokens[i].Surface
			var j int
			for j = i + 1; j < len(tokens); j++ {
				if reKana.MatchString(tokens[j].Surface) {
					surface += tokens[j].Surface
				} else {
					break
				}
			}
			tokens[i].Surface = surface
			for k := 0; k < (j - i); k++ {
				if i+1+k < len(tokens) && j+k < len(tokens) {
					tokens[i+1+k] = tokens[j+k]
				}
			}
			tokens = tokens[:len(tokens)-(j-i)+1]
			i = j
		}
	}

	ret := []string{}
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		c := tok.Features()
		if len(c) < 7 {
			continue
		}
		var y string
		if len(c) < 10 {
			y = c[6]
		} else {
			y = c[9]
		}
		if y == "*" {
			y = tok.Surface
		}
		if !reWord.MatchString(y) {
			if y == "、" {
				continue
			}
			pos = 0
			ambigous = 0
			sentence = ""
			copy(r, rule)
			continue
		}
		if pos >= len(rule) || (r[pos] == rule[pos] && !isWord(c)) {
			pos = 0
			ambigous = 0
			sentence = ""
			copy(r, rule)
			continue
		}
		ambigous += strings.Count(y, "ッ") + strings.Count(y, "ー")
		n := countChars(y)
		r[pos] -= n
		sentence += tok.Surface
		if r[pos] >= 0 && (r[pos] == 0 || r[pos]+ambigous == 0) {
			pos++
			if pos == len(r) || pos == len(r)+1 {
				if isEnd(c) {
					ret = append(ret, sentence)
					start = i + 1
				}
				pos = 0
				ambigous = 0
				sentence = ""
				copy(r, rule)
				continue
			}
			sentence += " "
		} else if r[pos] < 0 {
			i = start + 1
			start++
			pos = 0
			ambigous = 0
			sentence = ""
			copy(r, rule)
		}
	}
	return ret, nil
}

// Find returns sentences that text matches with rule(s).
func Find(text string, rule []int) []string {
	res, _ := FindWithOpt(text, rule, nil)
	return res
}
