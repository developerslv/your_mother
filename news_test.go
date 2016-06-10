package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTop(t *testing.T) {
	h := NewHackersNews()

	s, e := h.GetTop(5)

	assert.Nil(t, e)
	assert.Equal(t, 5, len(s))
}

func TestGetStory(t *testing.T) {
	h := NewHackersNews()

	story, e := h.GetStory(11878149)

	assert.Nil(t, e)
	assert.NotNil(t, story)

	assert.Equal(t, "Why Rust for Low-Level Linux Programming?", story.Title)
	assert.NotZero(t, story.Score)
	assert.Equal(t, "http://groveronline.com/2016/06/why-rust-for-low-level-linux-programming/", story.URL)
}

func TestGetStories(t *testing.T) {
	h := NewHackersNews()

	s := h.GetStories([]uint64{11878149, 11862476, 11871587})

	assert.Equal(t, 3, len(s))

	for _, v := range s {
		assert.NotNil(t, v)
	}
}
