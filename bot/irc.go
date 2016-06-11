package bot

import (
	log "github.com/Sirupsen/logrus"
	"github.com/thoj/go-ircevent"
	"net/rpc"
	"strings"
	"time"
)

type IRCSettings struct {
	Nick       string
	User       string
	Server     string
	Channel    string
	Debug      bool
	RPCNetwork string
	RPCAddress string
}

type IRCClient struct {
	con       *irc.Connection
	settings  *IRCSettings
	responder chan *ircResponse
}

type ircResponse struct {
	channel string
	message string
}

func NewIRCClient(settings *IRCSettings) *IRCClient {
	client := &IRCClient{}
	client.con = irc.IRC(settings.Nick, settings.User)
	client.settings = settings
	client.responder = make(chan *ircResponse)
	client.con.Debug = settings.Debug

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

	go c.newTopLoop()

	c.con.Loop()

	return nil
}

func (c *IRCClient) newTopLoop() {
	for {
		c.makeRequest(&RPCCommand{
			Message:   "!new_top",
			Arguments: []string{c.settings.Channel},
			Nick:      c.settings.Nick,
		})
		time.Sleep(time.Minute * 5)
	}
}

func (c *IRCClient) messageCallback(e *irc.Event) {
	msg := strings.TrimSpace(e.Message())

	if msg[0] != '!' {
		return
	}

	if msg == "!irc_version" {
		go func() {
			c.responder <- &ircResponse{channel: e.Arguments[0], message: "Git hash " + GitHash}
		}()

		return
	}

	go c.makeRequest(&RPCCommand{
		Message:   e.Message(),
		Arguments: e.Arguments,
		Code:      e.Code,
		Host:      e.Host,
		Nick:      e.Nick,
		Raw:       e.Raw,
		Source:    e.Source,
		User:      e.User,
	})
}

func (c *IRCClient) makeRequest(command *RPCCommand) {
	log.WithFields(log.Fields{
		"network": c.settings.RPCNetwork,
		"address": c.settings.RPCAddress,
	}).Debug("Will make request to RPC server")

	client, err := rpc.DialHTTP(c.settings.RPCNetwork, c.settings.RPCAddress)

	if err != nil {
		log.WithError(err).Error("Failed to connect to RPC server")
		return
	}

	var result CommandResponse

	log.WithFields(log.Fields{
		"command": command,
	}).Debugf("Will make rpc request")

	err = client.Call("BotRPCServer.Execute", command, &result)

	if err != nil {
		log.WithError(err).Error("Failed to execute RPC call")
		return
	}

	for _, line := range result.Lines {
		c.responder <- &ircResponse{channel: result.Channel, message: line}
	}
}
