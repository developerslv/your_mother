package main

import (
	"github.com/thoj/go-ircevent"
	"log"
	"strconv"
	"strings"
	"time"
)

var ircCommands = [...]string{
	"!weather {city} - Weather temperature",
	"!weather_cities - List available cities",
	"!hn_top {n}     - Best n hn links",
	"!help           - Help",
}

type IRCSettings struct {
	Nick    string
	User    string
	Server  string
	Channel string
}

type IRCClient struct {
	con       *irc.Connection
	settings  *IRCSettings
	responder chan *ircResponse
	weather   *Weather
	news      *HackerNews
}

type ircResponse struct {
	channel string
	message string
}

func NewIRCClient(settings IRCSettings, weather *Weather, news *HackerNews) *IRCClient {
	client := &IRCClient{}
	client.con = irc.IRC(settings.Nick, settings.User)
	client.settings = &settings
	client.responder = make(chan *ircResponse)
	client.weather = weather
	client.news = news
	client.con.Debug = true
	return client
}

func (c *IRCClient) Start() error {
	if err := c.con.Connect(c.settings.Server); err != nil {
		return err
	}

	go func() {
		for {
			msg, ok := <-c.responder

			if !ok {
				break
			}

			c.con.Privmsg(msg.channel, msg.message)

			time.Sleep(time.Second * 1)
		}
	}()

	c.con.AddCallback("PRIVMSG", c.messageCallback)

	c.con.Join(c.settings.Channel)

	c.con.Loop()

	return nil
}

func (c *IRCClient) messageCallback(e *irc.Event) {
	msg := strings.TrimSpace(e.Message())
	channel := e.Arguments[0]

	if channel[0] != '#' {
		channel = e.Nick
	}

	if msg == "!weather" {
		go func() {
			w, err := c.weather.ForCity("Rīga")
			if err != nil {
				log.Printf("Weather fetch failed %s", err)
				c.responder <- &ircResponse{channel: channel, message: ":("}
				return
			}

			c.responder <- &ircResponse{channel: channel, message: w}
		}()

		return
	}

	if msg == "!weather_cities" {
		go func() {
			cities, err := c.weather.ListCities()
			if err != nil {
				log.Printf("Weather fetch city list %s", err)
				c.responder <- &ircResponse{channel: channel, message: ":("}
				return
			}

			for _, city := range cities {
				c.responder <- &ircResponse{channel: e.Nick, message: city}
			}
		}()

		return
	}

	if strings.HasPrefix(msg, "!weather ") {
		go func() {
			city := msg[len("!weather "):]

			w, err := c.weather.ForCity(city)

			if err != nil {
				log.Printf("Weather fetch weather for city %s failed %s", city, err)
				c.responder <- &ircResponse{channel: channel, message: ":("}
				return
			}

			c.responder <- &ircResponse{channel: channel, message: w}
		}()

		return
	}

	if strings.HasPrefix(msg, "!hn_top") {
		go func() {
			count, err := strconv.Atoi(strings.TrimSpace(msg[len("!hn_top"):]))

			if err != nil {
				count = 5
			}

			if count > 5 {
				count = 5
			}

			stories, err := c.news.GetTop(count)

			if err != nil {
				log.Printf("Failed to fetch hn best stories %d %s", count, err)
				c.responder <- &ircResponse{channel: channel, message: ":("}
				return
			}

			for i, s := range stories {
				c.responder <- &ircResponse{channel: channel, message: strconv.Itoa(i+1) + " : " + s.String()}
			}
		}()
	}

	if msg == "!help" {
		go func() {
			for _, cmd := range ircCommands {
				c.responder <- &ircResponse{channel: e.Nick, message: cmd}
			}
		}()
	}
}

func (c *IRCClient) Stop() {
	close(c.responder)
	c.con.Disconnect()
}
