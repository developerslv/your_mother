package suppliers

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGenerateString(t *testing.T) {
	m, err := NewMarkov(os.Getenv("TEST_LOGS_PATH"))
	assert.Nil(t, err)

	word := m.Generate("xn lox", 2, 5)

	assert.NotEmpty(t, word)
}
