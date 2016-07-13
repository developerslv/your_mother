package suppliers

import (
	"container/ring"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/zabawaba99/firego"
	"strconv"
	"sync"
	"time"
)

type HackerNews struct {
	previousTop *ring.Ring
	topLock     *sync.Mutex
	sentId      uint64
	lastTop     *HackerNewsStory
}

type HackerNewsStory struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Score int    `json:"score"`
	Type  string `json:"type"`
	Id    uint64 `json:"id"`
}

func (s *HackerNewsStory) String() string {
	return s.URL + " - " + s.Title + "(" + strconv.Itoa(s.Score) + ")"
}

func NewHackersNews() *HackerNews {
	return &HackerNews{topLock: &sync.Mutex{}, previousTop: ring.New(50)}
}

func (h *HackerNews) GetTop(cnt int) ([]*HackerNewsStory, error) {
	topStoriesFB := firego.New("https://hacker-news.firebaseio.com/v0/topstories", nil)

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
	for retryCount := 0; retryCount <= 3; retryCount += 1 {
		storyFB := firego.New("https://hacker-news.firebaseio.com/v0/item/"+strconv.FormatUint(id, 10), nil)

		log.Debugf("Will fetch story %d", id)

		var story HackerNewsStory
		if e := storyFB.Value(&story); e != nil {
			return nil, e
		}

		if story.Id == id {
			return &story, nil
		}

		log.WithFields(log.Fields{
			"requestedId": id,
			"obtainedId":  story.Id,
		}).Debugf("Failed to obtain requested story")

		time.Sleep(time.Second * 10) //there is delay until story appears in item feed
	}

	return nil, errors.New("Failed to obtain story that matches requested id")
}

//Ugly, but works, as right now there isn't a stream down which to send
//new stories it requires sending qurying rpc server for new stories
func (h *HackerNews) GetLastTop() *HackerNewsStory {
	h.topLock.Lock()
	defer h.topLock.Unlock()

	if h.lastTop == nil || h.sentId == h.lastTop.Id {
		return nil
	}

	h.sentId = h.lastTop.Id
	return h.lastTop
}

func (h *HackerNews) BackgroundNewLoop() {
	go h.doNewLoop()
}

func (h *HackerNews) doNewLoop() {
	for {
		subscribeFB := firego.New("https://hacker-news.firebaseio.com/v0/topstories", nil)
		notifications := make(chan firego.Event)
		subscribeFB.Watch(notifications)
		h.readNewTopNews(notifications)
	}
}

func (h *HackerNews) readNewTopNews(notifications chan firego.Event) {
	for notification := range notifications {
		if notification.Type == firego.EventTypeError {
			err, ok := notification.Data.(error)

			if ok {
				log.WithError(err).Error("Failed to do wath on an item because of error")
			} else {
				log.WithField("data", notification.Data).Error("Failed to watch item with unknown error")
			}

			return
		}

		ids, ok := notification.Data.([]interface{})
		if !ok {
			log.WithField("data", notification.Data).Error("Failed to unparse received data")
			continue
		}

		id, ok := ids[0].(float64)

		if !ok {
			log.WithField("id", ids[0]).Error("Failed to unparse received id")
			continue
		}

		storyId := uint64(id)

		if h.idWasSeen(storyId) {
			log.WithField("id", storyId).Debug("Got already seen story")
			continue
		}

		log.WithField("id", storyId).Debug("Story wasnt seen")
		h.addAsSeen(storyId)

		go func(id uint64) {
			story, err := h.GetStory(id)

			if err != nil {
				log.WithError(err).Error("Failed to fetch story because of error")
				return
			}

			h.setStory(story)
		}(storyId)
	}
}

func (h *HackerNews) setStory(story *HackerNewsStory) {
	h.topLock.Lock()
	defer h.topLock.Unlock()
	h.lastTop = story
}

func (h *HackerNews) idWasSeen(id uint64) bool {
	h.topLock.Lock()
	defer h.topLock.Unlock()

	current := h.previousTop

	for good := true; good; good = current != h.previousTop {
		idVal, ok := current.Value.(uint64)
		if ok && id == idVal {
			return true
		}
		current = current.Next()
	}

	return false
}

func (h *HackerNews) addAsSeen(id uint64) {
	h.topLock.Lock()
	h.topLock.Unlock()

	h.previousTop = h.previousTop.Next()
	h.previousTop.Value = id
}
