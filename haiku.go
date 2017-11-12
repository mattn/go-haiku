package haiku

import (
	"regexp"
	"strings"

	"github.com/ikawaha/kagome/tokenizer"
)

var (
	reWord       = regexp.MustCompile(`^[ァ-ヾ]+$`)
	reIgnoreText = regexp.MustCompile(`[\[\]「」『』]`)
	reIgnoreChar = regexp.MustCompile(`[ァィゥェォャュョ]`)
)

func isEnd(c []string) bool {
	return c[1] != "非自立" && !strings.HasPrefix(c[5], "連用")
}

// isWord return true when the kind of the word is possible to be leading of
// sentence.
func isWord(c []string) bool {
	for _, f := range []string{"名詞", "形容詞", "形容動詞", "副詞", "連体詞", "接続詞", "感動詞", "接頭詞", "フィラー"} {
		if f == c[0] && c[1] != "非自立" && c[1] != "接尾" {
			return true
		}
	}
	if c[0] == "動詞" && c[1] != "接尾" {
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
	t := tokenizer.New()
	text = reIgnoreText.ReplaceAllString(text, " ")
	tokens := t.Tokenize(text)
	pos := 0
	r := make([]int, len(rule))
	copy(r, rule)

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		c := tok.Features()
		if len(c) == 0 {
			continue
		}
		y := c[len(c)-1]
		if !reWord.MatchString(y) {
			if y == "、" {
				continue
			}
			return false
		}
		if r[pos] == rule[pos] && !isWord(c) {
			return false
		}
		n := countChars(y)
		r[pos] -= n
		if r[pos] == 0 {
			pos++
			if pos == len(r) && i == len(tokens)-2 {
				return true
			}
		}
	}
	return false
}

// Find returns sentences that text matches with rule(s).
func Find(text string, rule []int) []string {
	if len(rule) == 0 {
		return nil
	}
	t := tokenizer.New()
	text = reIgnoreText.ReplaceAllString(text, " ")
	tokens := t.Tokenize(text)
	pos := 0
	r := make([]int, len(rule))
	copy(r, rule)
	sentence := ""
	start := 0
	ambigous := 0

	ret := []string{}
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		c := tok.Features()
		if len(c) == 0 {
			continue
		}
		y := c[len(c)-1]
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
		if r[pos] == rule[pos] && !isWord(c) {
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
		if r[pos] == 0 || r[pos]+ambigous == 0 {
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
	return ret
}
