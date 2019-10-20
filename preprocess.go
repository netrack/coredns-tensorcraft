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

func (t Tokenizer) TextToSeq(text string) []int {
	runes := ([]rune)(strings.ToLower(text))

	numChars := len(runes)
	if numChars > t.maxChars {
		numChars = t.maxChars
	}

	println(numChars)

	seq := make([]int, t.maxChars)
	for i := numChars - 1; i >= 0; i-- {
		no := t.dict[runes[i]]
		seq[numChars-i-1] = no
	}
	return seq
}
