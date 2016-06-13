package bot

import (
	log "github.com/Sirupsen/logrus"
	"github.com/dainis/your_mother/bot/suppliers"
	"strconv"
	"strings"
)

var ircCommands = [...]string{
	"!weather {city} - Weather temperature",
	"!weather_cities - List available cities",
	"!hn_top {n}     - Top n hn links",
	"!help           - Help",
	"!irc_version    - Display rpc git hash",
	"!rpc_version    - Display irc git hash",
	"!repo           - Display repo",
}

type BotRPCServer struct {
	weather *suppliers.Weather
	news    *suppliers.HackerNews
	markov  *suppliers.Markov
}

func NewRPCServer(weather *suppliers.Weather, news *suppliers.HackerNews, markov *suppliers.Markov) *BotRPCServer {
	return &BotRPCServer{weather: weather, news: news, markov: markov}
}

func (srv *BotRPCServer) Execute(e RPCCommand, resp *CommandResponse) error {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered from %v", r)
		}
	}()

	msg := strings.TrimSpace(e.Message)
	channel := e.Arguments[0]

	log.Debugf("Got request for %s", msg)

	if channel[0] != '#' { //private message
		channel = e.Nick
	}

	resp.Channel = channel

	if msg == "!weather" {
		w, err := srv.weather.ForCity("Rīga")

		if err != nil {
			log.Printf("Weather fetch failed %s", err)
			resp.AppendLine(":(")
		} else {
			resp.AppendLine(w)
		}
	}

	if msg == "!weather_cities" {
		cities, err := srv.weather.ListCities()
		resp.Channel = e.Nick

		if err != nil {
			log.Printf("Weather fetch city list %s", err)
			resp.AppendLine(":(")
		} else {
			for _, city := range cities {
				resp.AppendLine(city)
			}
		}
	}

	if strings.HasPrefix(msg, "!weather ") {
		city := msg[len("!weather "):]

		w, err := srv.weather.ForCity(city)

		if err != nil {
			log.WithError(err).Error("Weather fetch weather for city %s failed", city)
			resp.AppendLine(":(")
		} else {
			resp.AppendLine(w)
		}
	}

	if strings.HasPrefix(msg, "!hn_top") {
		count, err := strconv.Atoi(strings.TrimSpace(msg[len("!hn_top"):]))

		if err != nil {
			count = 5
		}

		if count > 5 {
			count = 5
		}

		stories, err := srv.news.GetTop(count)

		if err != nil {
			log.Printf("Failed to fetch hn best stories %d %s", count, err)
			resp.AppendLine(":(")
		} else {
			for i, s := range stories {
				resp.AppendLine(strconv.Itoa(i+1) + " : " + s.String())
			}
		}
	}

	if msg == "!new_top" {
		story := srv.news.GetLastTop()
		if story != nil {
			resp.AppendLine("Trending @ HN : " + story.String())
		}
	}

	if msg == "!help" {
		for _, cmd := range ircCommands {
			resp.Channel = e.Nick
			resp.AppendLine(cmd)
		}
	}

	if msg == "!rpc_version" {
		resp.AppendLine("Git hash : " + GitHash)
	}

	if msg == "!repo" {
		resp.AppendLine("Repo : " + GitRepo)
	}

	if strings.HasPrefix(msg, "!markov") {
		usageString := "usage !markov <prefix length> <max words in generated resp> <input sentence>"
		parts := strings.Fields(msg)

		if len(parts) < 4 {
			resp.AppendLine(usageString)
			return nil
		}

		sentence := strings.Join(parts[3:], " ")

		log.Debugf("Got markov request with prefix length %s, max words %s, sentence %s", parts[1], parts[2], sentence)

		prefixLength, err := strconv.Atoi(parts[1])
		if err != nil {
			resp.AppendLine(usageString)
			return nil
		}

		words, err := strconv.Atoi(parts[2])
		if err != nil {
			resp.AppendLine(usageString)
			return nil
		}

		generated := srv.markov.Generate(sentence, prefixLength, words)

		if len(generated) == 0 {
			resp.AppendLine(":(")
		} else {
			resp.AppendLine(generated)
		}
	}

	return nil
}
