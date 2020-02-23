package dnstun

import (
	"strings"
)

const (
	enUS = "abcdefghijklmnopqrstuvwxyz0123456789-,;.!?:'\"/\\|_@#$%^&*~`+-=<>()[]{}"
)

type Tokenizer struct {
	dict     map[rune]int
	maxChars int
}

func NewTokenizer(alphabet string, maxChars int) Tokenizer {
	dict := make(map[rune]int, len(alphabet))
	for i, char := range alphabet {
		dict[char] = i + 1
	}
	return Tokenizer{dict, maxChars}
}

func (t Tokenizer) TextToSeq(text string) []int64 {
	runes := ([]rune)(strings.ToLower(text))

	numChars := len(runes)
	if numChars > t.maxChars {
		numChars = t.maxChars
	}

	seq := make([]int64, t.maxChars)
	for i := numChars - 1; i >= 0; i-- {
		no := t.dict[runes[i]]
		seq[numChars-i-1] = int64(no)
	}
	return seq
}
