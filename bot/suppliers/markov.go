package suppliers

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func fieldSplit(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsNumber(r)
}

type Markov struct {
	words []string
}

func NewMarkov(folder string) (*Markov, error) {
	lines, err := readText(folder)

	if err != nil {
		return nil, err
	}

	m := &Markov{words: make([]string, 0, len(lines))}
	m.splitInWords(lines)

	return m, nil
}

func (m *Markov) AppendWords(sentence string) {
	for _, w := range strings.FieldsFunc(sentence, fieldSplit) {
		if len(w) > 1 {
			m.words = append(m.words, strings.ToLower(w))
		}
	}
}

func (m *Markov) Generate(sentence string, prefixLength int, length int) string {
	if length > 10 {
		length = 10
	}

	if prefixLength > 5 {
		prefixLength = 5
	}

	chain := NewMarkovChain(m.words, prefixLength)
	return chain.Generate(sentence, length)
}

func readText(folder string) ([]string, error) {
	dir, err := os.Open(folder)

	if err != nil {
		return nil, err
	}

	files, err := dir.Readdir(-1)

	lines := make([]string, 0)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fd, err := os.Open(filepath.Join(folder, file.Name()))

		if err != nil {
			return nil, err
		}

		defer fd.Close()

		content, err := ioutil.ReadAll(fd)

		if err != nil {
			return nil, err
		}

		lines = append(lines, strings.Split(string(content), "\n")...)
	}

	return lines, nil
}

func (m *Markov) splitInWords(lines []string) {
	tmpLines := make([]string, 0, len(lines))

	for _, line := range lines {
		if len(line) > 10 { //starts with timestamp
			tmpLines = append(tmpLines, line)
		}
	}

	for i, line := range tmpLines {
		tmpLines[i] = line[11:] //timestamp prefix
	}

	lines = make([]string, 0, len(tmpLines))

	for _, line := range tmpLines {
		if len(line) > 3 && line[:3] == "***" { //joins and quits are ignored
			continue
		}

		lines = append(lines, line)
	}

	for i, line := range lines {
		if line[0] != '<' { //user message starts with nick
			continue
		}

		for j, c := range line {
			if c == '>' && j+2 < len(line) { //closing tag after nick then there is space and actual message text
				lines[i] = line[j+2:]
			}
		}
	}

	for _, line := range lines {
		m.AppendWords(line)
	}
}
