package suppliers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestForCity(t *testing.T) {
	w := NewWeather()

	v, e := w.ForCity("Rīga - LU")

	assert.Nil(t, e)
	assert.NotEmpty(t, v)
}

func TestListCities(t *testing.T) {
	w := NewWeather()

	c, e := w.ListCities()

	assert.Nil(t, e)
	assert.NotEqual(t, 0, len(c))
}
