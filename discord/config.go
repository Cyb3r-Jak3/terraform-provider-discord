package discord

import (
	"github.com/bwmarrin/discordgo"
)

type Config struct {
	Token    string
	ClientID string
	Secret   string
}

type Context struct {
	Session *discordgo.Session
	Config  *Config
}

func (c *Config) Client(version string) (*Context, error) {
	session, err := discordgo.New(c.Token)
	session.UserAgent = "discord-terraform/" + version
	if err != nil {
		return nil, err
	}

	return &Context{Config: c, Session: session}, nil
}
