package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/olekukonko/tablewriter"
)

// Bot parameters
var (
	GuildID = flag.String(
		"guild",
		"",
		"Test guild ID. If not passed - bot registers commands globally",
	)
	BotToken       = flag.String("token", "", "Bot access token")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

var s *discordgo.Session

func init() { flag.Parse() }

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	if err != nil {
		log.Fatalf("Could not create new SSH agent: %v", err)
	}
}

// CheckRepoExists We use git ls-remote to see if the user has provided with a valid git repo url.
func CheckRepoExists(repo string) bool {
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		URLs: []string{repo},
	})

	_, err := rem.List(&git.ListOptions{})
	return err == nil
}

func RepoClonePath(dataPath string, u *transport.Endpoint) string {
	return path.Join(dataPath, u.Host, u.Path)
}

func Handle_cloc(s *discordgo.Session, i *discordgo.InteractionCreate) {
	repo := i.ApplicationCommandData().Options[0].StringValue()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		// Data: &discordgo.InteractionResponseData{
		// 	Flags: discordgo.MessageFlagsEphemeral,
		// },
	})

	// TODO: Send progress report to Discord.
	u, err := transport.NewEndpoint(repo)
	if err != nil {
		log.Println("Error parsing repo url:", err)
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Invalid repo URL",
		})
		return
	}
	switch u.Protocol {
	case "http", "https":
	default:
		// TODO: maybe try  sshAuth?
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Files ar enot allowd",
		})
		return
	}

	if !CheckRepoExists(repo) {
		log.Println("Repo does not exist:", repo)
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Invalid repo URL",
		})
		return
	}

	cloneDest := RepoClonePath("_data", u)
	defer func() {
		CleanAfterClone(cloneDest)
	}()

	log.Println("Cloning")
	ctxClone, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(time.Minute),
	)
	defer cancel()
	if _, err = git.PlainCloneContext(ctxClone, cloneDest, false, &git.CloneOptions{
		URL:   repo,
		Depth: 1,
		// Progress: nil, // TODO: Parse output for progress report.
	}); err != nil {
		log.Println("Error cloning repo:", err)
	}

	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(time.Minute),
	)
	defer cancel()

	cmd := exec.CommandContext(ctx, "scc", "-f", "json", cloneDest)
	p, err := cmd.StdoutPipe()
	defer p.Close()

	log.Println("SCC subprocess spawned")
	if err != nil {
		log.Println("Error creating pipe to scc subprocess:", err)
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong, my bad...",
		})
		return
	}

	if err = cmd.Start(); err != nil {
		log.Println("Error starting scc process:", err)
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong, my bad...",
		})
		return
	}

	log.Println("SCC subprocess started")
	var (
		decoder = json.NewDecoder(p)
		result  []SccOutput
	)

	if err = decoder.Decode(&result); err != nil {
		log.Println("Error decoding scc output:", err)
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong, my bad...",
		})

		defer cmd.Cancel()
		return
	}

	if err = cmd.Wait(); err != nil {
		log.Println("Error running scc subprocess:", err)
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Something went wrong, my bad...",
		})

		defer cmd.Cancel()
		return
	}

	var (
		builder = &strings.Builder{}
		tb      = tablewriter.NewWriter(builder)
		t       = SccOutput{}
		rows    = make([][]string, len(result))
	)

	builder.WriteString("```md\n")
	for i, r := range result {
		rows[i] = []string{
			r.Name,
			strconv.Itoa(r.Count),
			strconv.Itoa(r.Blank),
			strconv.Itoa(r.Comment),
			strconv.Itoa(r.Code),
			strconv.Itoa(r.Lines),
		}
		t.Count += r.Count
		t.Blank += r.Blank
		t.Comment += r.Comment
		t.Code += r.Code
		t.Lines += r.Lines
	}

	tb.SetHeader([]string{"Language", "Files", "Blanks", "Comments", "Code", "Lines"})
	tb.SetFooter([]string{
		"Total",
		strconv.Itoa(t.Count),
		strconv.Itoa(t.Blank),
		strconv.Itoa(t.Comment),
		strconv.Itoa(t.Code),
		strconv.Itoa(t.Lines),
	})

	tb.AppendBulk(rows)

	tb.Render()
	builder.WriteString("\n```")
	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: builder.String(),
	})
}

func CleanAfterClone(clonedAt string) {
	// TODO: Make an LRU cache for repos.
	switch err := os.RemoveAll(clonedAt); {
	case err == os.ErrNotExist:
		log.Println("Nothing to clean at:", clonedAt)
	case err != nil:
		log.Printf("Error while cleaning:%v :%v", clonedAt, err)
	default:
		log.Println("Cleaned:", clonedAt)
	}
}

type SccOutput struct {
	Name    string
	Lines   int
	Code    int
	Comment int
	Blank   int
	Count   int

	// Bytes              int
	// CodeBytes          int
	// Files              []string
	// Complexity         int
	// WeightedComplexity int
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
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
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"cloc": Handle_cloc,
	}
)

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, _ *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if *RemoveCommands {
		log.Println("Removing commands...")
		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
