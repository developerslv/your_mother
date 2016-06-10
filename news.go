package main

import (
	// "fmt"
	"github.com/zabawaba99/firego"
	"strconv"
	"sync"
)

type HackerNews struct{}

type HackerNewsStory struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Score int    `json:"score"`
}

func (s *HackerNewsStory) String() string {
	return s.URL + " - " + s.Title + "(" + strconv.Itoa(s.Score) + ")"
}

func NewHackersNews() *HackerNews {
	return &HackerNews{}
}

func (h *HackerNews) GetTop(cnt int) ([]*HackerNewsStory, error) {
	topStoriesFB := firego.New("https://hacker-news.firebaseio.com/v0/beststories", nil)

	var v []uint64

	if e := topStoriesFB.Value(&v); e != nil {
		return nil, e
	}

	return h.GetStories(v[:cnt]), nil
}

func (h *HackerNews) GetStories(ids []uint64) []*HackerNewsStory {
	var wg sync.WaitGroup

	c := make(chan *HackerNewsStory, len(ids))

	wg.Add(len(ids))

	result := make([]*HackerNewsStory, 0, len(ids))

	go func() {
		for {
			story, ok := <-c

			if !ok {
				break
			}

			if story != nil {
				result = append(result, story)
			}

			wg.Done()
		}
	}()

	for _, id := range ids {
		go func(id uint64) {
			story, _ := h.GetStory(id)

			c <- story
		}(id)
	}

	wg.Wait()
	close(c)

	return result
}

func (h *HackerNews) GetStory(id uint64) (*HackerNewsStory, error) {
	storyFB := firego.New("https://hacker-news.firebaseio.com/v0/item/"+strconv.FormatUint(id, 10), nil)

	var story HackerNewsStory

	if e := storyFB.Value(&story); e != nil {
		return nil, e
	}

	return &story, nil
}
