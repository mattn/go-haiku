package haiku

import (
	"testing"

	"github.com/ikawaha/kagome-dict/uni"
)

func TestNumberReading(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"0", "ゼロ"},
		{"1", "イチ"},
		{"10", "ジュウ"},
		{"12", "ジュウニ"},
		{"100", "ヒャク"},
		{"1000", "セン"},
		{"2024", "ニセンニジュウヨン"},
		{"10000", "イチマン"},
		{"10001", "イチマンイチ"},
		{"100000000", "イチオク"},
		{"120000000", "イチオクニセンマン"},
		{"１２", "ジュウニ"},
		{"007", "ゼロゼロナナ"},
	}
	for _, tt := range tests {
		got, ok := numberReading(tt.in)
		if !ok || got != tt.want {
			t.Errorf("numberReading(%q) = %q, %v; want %q", tt.in, got, ok, tt.want)
		}
	}
	if _, ok := numberReading("1a"); ok {
		t.Error("numberReading must reject non-digits")
	}
}

func TestAcronymReading(t *testing.T) {
	if got := acronymReading("NHK"); got != "エヌエイチケー" {
		t.Errorf("got %q", got)
	}
	if got := acronymReading("ＴＶ"); got != "ティーブイ" {
		t.Errorf("got %q", got)
	}
}

func TestMatchNumbers(t *testing.T) {
	opt := &Opt{Dict: uni.Dict()}
	tests := []struct {
		text string
		want bool
	}{
		// 4時 is read ヨジ, not ヨンジ
		{"朝4時に目覚ましが鳴る眠いなあ", true},
		{"おれも朝4時から待つパパになる", false},
		// digits split by the dictionary are read as one number
		{"みんなA100だのH100だのと", true},
		{"1,000円が9,000円になりました", true},
		// emoji and brackets without reading are ignored
		{"古池や蛙飛び込む水の音🐸", true},
		{"マロンちゃん（先代犬）もたびだった", true},
	}
	for _, tt := range tests {
		if got := MatchWithOpt(tt.text, []int{5, 7, 5}, opt); got != tt.want {
			t.Errorf("MatchWithOpt(%q) = %v; want %v", tt.text, got, tt.want)
		}
	}
}

func TestHiraganaToKatakana(t *testing.T) {
	if got := hiraganaToKatakana("ぱわー"); got != "パワー" {
		t.Errorf("got %q", got)
	}
}
