package main

import (
	"container/ring"
	"github.com/zabawaba99/firego"
	"log"
	"strconv"
	"sync"
)

type HackerNews struct {
	previousTop *ring.Ring
	topLock     *sync.Mutex
}

type HackerNewsStory struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Score int    `json:"score"`
	Type  string `json:"type"`
}

func (s *HackerNewsStory) String() string {
	return s.URL + " - " + s.Title + "(" + strconv.Itoa(s.Score) + ")"
}

func NewHackersNews() *HackerNews {
	return &HackerNews{topLock: &sync.Mutex{}, previousTop: ring.New(50)}
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

func (h *HackerNews) SubscribeToNew() (chan *HackerNewsStory, error) {
	subscribeFB := firego.New("https://hacker-news.firebaseio.com/v0/beststories", nil)
	notifications := make(chan firego.Event)
	subscribeFB.Watch(notifications)

	newStories := make(chan *HackerNewsStory)

	go func() {
		for notification := range notifications {
			ids, ok := notification.Data.([]interface{})
			if !ok {
				log.Printf("Failed to unparse received data %v", notification.Data)
				continue
			}

			id, ok := ids[0].(float64)

			if !ok {
				log.Printf("Failed to unparse received id %v", ids[0])
				continue
			}

			storyId := uint64(id)

			if h.idWasSeen(storyId) {
				log.Printf("%d was seen", storyId)
				continue
			}

			log.Printf("%d already seen", storyId)
			h.addAsSeen(storyId)

			go func(id uint64) {
				story, err := h.GetStory(id)

				if err != nil {
					log.Printf("Failed to fetch story because of error %s", err)
					return
				}

				newStories <- story
			}(storyId)
		}
	}()

	return newStories, nil
}

func (h *HackerNews) idWasSeen(id uint64) bool {
	h.topLock.Lock()
	defer h.topLock.Unlock()

	exists := false

	h.previousTop.Do(func(v interface{}) {
		idVal, ok := v.(uint64)
		if ok && id == idVal {
			exists = true
		}
	})

	return exists
}

func (h *HackerNews) addAsSeen(id uint64) {
	h.topLock.Lock()
	h.topLock.Unlock()

	h.previousTop = h.previousTop.Next()
	h.previousTop.Value = id
}
