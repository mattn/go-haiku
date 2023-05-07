package haiku

import (
	"bufio"
	"log"
	"os"
	"strings"
	"testing"
)

func testMatch(t *testing.T, filename string, rules []int, judge bool) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "#") {
			continue
		}
		if MatchWithOpt(text, rules, &Opt{Debug: true}) != judge {
			t.Fatalf("%q for %q must be %v", text, filename, rules)
		}
	}
}

func TestHaiku(t *testing.T) {
	testMatch(t, "testdata/haiku.good", []int{5, 7, 5}, true)
	testMatch(t, "testdata/haiku.bad", []int{5, 7, 5}, false)
	testMatch(t, "testdata/tanka.good", []int{5, 7, 5, 7, 7}, true)
	testMatch(t, "testdata/tanka.bad", []int{5, 7, 5, 7, 7}, false)
}
