package helpers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func Defer(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		err = fmt.Errorf("deferred interaction response: %w", err)
	}

	return err
}

type FollowupResponse struct {
	Error   error
	Content string
}
