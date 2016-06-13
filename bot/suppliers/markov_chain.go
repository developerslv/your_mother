package suppliers

import (
	"math/rand"
	"strings"
	"time"
)

type MarkovChain struct {
	chain        map[string][]string
	prefixLength int
}

func NewMarkovChain(words []string, prefixLength int) *MarkovChain {
	chain := &MarkovChain{chain: make(map[string][]string), prefixLength: prefixLength}
	chain.buildChain(words)
	return chain
}

func (m *MarkovChain) buildChain(words []string) {
	for i := range words {
		if len(words) <= i+m.prefixLength {
			continue
		}

		prefix := strings.Join(words[i:i+m.prefixLength], " ")
		m.chain[prefix] = append(m.chain[prefix], words[i+m.prefixLength])
	}
}

func (m *MarkovChain) Generate(sentence string, length int) string {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	prefixWords := strings.FieldsFunc(sentence, fieldSplit)

	if len(prefixWords) < m.prefixLength {
		return ""
	}

	for i := range prefixWords {
		prefixWords[i] = strings.ToLower(prefixWords[i])
	}

	generated := make([]string, 0, length)

	for i := 0; i < length; i++ {
		prefix := strings.Join(prefixWords[:m.prefixLength], " ")
		prefixWords = prefixWords[1:]

		if len(m.chain[prefix]) == 0 {
			break
		}

		newWord := m.chain[prefix][r.Intn(len(m.chain[prefix]))]

		generated = append(generated, newWord)
		prefixWords = append(prefixWords, newWord)
	}

	return strings.Join(generated, " ")
}
