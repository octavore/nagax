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
	"fmt"

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
	defaultChannel string
	env            service.Environment
}

type slackClient interface {
	PostMessage(channel string, params ...slack.MsgOption) (string, string, error)
}

func (m *Module) Init(c *service.Config) {
	c.Setup = func() error {
		err := m.Config.ReadConfig(&m.config)
		if err != nil {
			return errors.Wrap(err)
		}
		m.defaultChannel = m.config.SlackConfig.Channel
		if m.defaultChannel == "" {
			m.defaultChannel = "#activity"
		}
		m.env = c.Env()
		return nil
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

// Post a message to the default channel
func (m *Module) Post(txt string, params ...slack.MsgOption) {
	m.PostC(m.defaultChannel, txt, params...)
}

// PostC posts a message to the given channel
func (m *Module) PostC(channel, txt string, params ...slack.MsgOption) {
	if !m.env.IsProduction() {
		txt = fmt.Sprintf("(%s) %s", m.env.String(), txt)
	}
	if m.LogMessages {
		m.Logger.Infof("[%s] %s", m.defaultChannel, txt)
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
