package slash

import (
	"errors"
	"fmt"
	"log"

	"andy/slash/cloc"

	"github.com/bwmarrin/discordgo"
)

func AddHandlers(dgs *discordgo.Session) {
	dgs.AddHandler(func(s *discordgo.Session, _ *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) (string, error){
		"cloc": cloc.Handle,
	}

	dgs.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			content, err := h(s, i)
			_, errFollowup := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: content,
			})
			if errFollowup != nil {
				err = errors.Join(err, fmt.Errorf("followup create: %w", errFollowup))
			}

			if err != nil {
				log.Println("Error:", err)
			}
		}
	})
}

func RegisterCommands(guildID string, dgs *discordgo.Session) []*discordgo.ApplicationCommand {
	commands := []*discordgo.ApplicationCommand{
		cloc.Cmd,
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))

	log.Println("Adding commands...")

	for i, v := range commands {
		cmd, err := dgs.ApplicationCommandCreate(dgs.State.User.ID, guildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}

		registeredCommands[i] = cmd
	}

	return registeredCommands
}
