/*
	package slack provides an easy way to read slack credentials
	from a config file and post (internal) messages to slack.

	```
	{
		"slack_internal": {
			"channel": "#activity",
			"api_token": "..."
		}
	}
	```
*/
package slack

import (
	"github.com/nlopes/slack"
	"github.com/octavore/naga/service"

	"github.com/octavore/nagax/config"
	"github.com/octavore/nagax/logger"
	"github.com/octavore/nagax/util/errors"
)

type Module struct {
	Config *config.Module
	Logger *logger.Module

	LogMessages bool
	client      slackClient
	config      struct {
		SlackConfig struct {
			Channel  string `json:"channel"`
			APIToken string `json:"api_token"`
		} `json:"slack_internal"`
	}
}

type slackClient interface {
	PostMessage(channel string, params ...slack.MsgOption) (string, string, error)
}

func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		return m.Config.ReadConfig(&m.config)
	}

	c.Start = func() {
		if m.config.SlackConfig.APIToken != "" {
			m.client = slack.New(m.config.SlackConfig.APIToken)
		} else {
			m.client = &dummyClient{logger: m.Logger}
			m.LogMessages = true
		}
	}
}

type MsgOption = slack.MsgOption

// Post a message to the default channel
func (m *Module) Post(txt string, params ...MsgOption) {
	m.PostC(m.config.SlackConfig.Channel, txt, params...)
}

// PostC posts a message to the given channel
func (m *Module) PostC(channel, txt string, params ...MsgOption) {
	if m.LogMessages {
		m.Logger.Infof("[%s] %s", m.config.SlackConfig.Channel, txt)
	}

	_, _, err := m.client.PostMessage(channel,
		slack.MsgOptionText(txt, false),
		slack.MsgOptionCompose(params...))
	if err != nil {
		m.Logger.Error(errors.Wrap(err))
	}
}

type dummyClient struct {
	logger *logger.Module
}

func (d *dummyClient) PostMessage(channel string, params ...slack.MsgOption) (string, string, error) {
	return "", "", nil
}
