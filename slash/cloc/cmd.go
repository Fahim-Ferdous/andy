package cloc

import (
	"github.com/bwmarrin/discordgo"
)

var Cmd = &discordgo.ApplicationCommand{
	Name:                     "cloc",
	Description:              "Count lines of source for a repository.",
	DescriptionLocalizations: &map[discordgo.Locale]string{},
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "repo",
			Description: "The repository you want to count",
			Required:    true,
		},
	},
}
