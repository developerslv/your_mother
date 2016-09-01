package bot

import (
	log "github.com/Sirupsen/logrus"
	"github.com/ably/ably-go/ably"
	"github.com/ably/ably-go/ably/proto"
	"github.com/thoj/go-ircevent"
	"net/rpc"
	"strings"
	"time"
)

type IRCSettings struct {
	Nick    string
	User    string
	Server  string
	Channel string
	Debug   bool

	RPCNetwork string
	RPCAddress string

	SubscriberKey     string
	SubscriberChannel string
}

type IRCClient struct {
	con       *irc.Connection
	settings  *IRCSettings
	responder chan *ircResponse

	subChannelName string
	subClientKey   string
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

	client.subClientKey = settings.SubscriberKey
	client.subChannelName = settings.SubscriberChannel

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

	go c.newSubLoop()

	c.con.Loop()

	return nil
}

func (c *IRCClient) newSubLoop() {
	for {
		client, err := ably.NewRealtimeClient(ably.NewClientOptions(c.settings.SubscriberKey))

		if err != nil {
			log.WithError(err).Error("Failed to create ably client")
		}

		log.Debug("Entering subscriber loop")

		channel := client.Channels.Get(c.subChannelName)

		sub, err := channel.Subscribe()

		if err != nil {
			log.WithError(err).Error("Failed to subscribe to %s", c.subChannelName)
			continue
		}

		for msg := range sub.MessageChannel() {
			log.Debugf("Got message %s", msg.Data)

			go func(msg *proto.Message) {
				c.responder <- &ircResponse{channel: c.settings.Channel, message: msg.Data}
			}(msg)
		}
	}
}

func (c *IRCClient) messageCallback(e *irc.Event) {
	msg := strings.TrimSpace(e.Message())

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
