package cloc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"andy/helpers"

	"github.com/bwmarrin/discordgo"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/olekukonko/tablewriter"
)

// CheckRepoExists We use git ls-remote to see if the user has provided with a valid git repo url.
func CheckRepoExists(repo string) error {
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		URLs: []string{repo},
	})

	_, err := rem.List(&git.ListOptions{})
	if err != nil {
		return fmt.Errorf("ls-remote: %w", err)
	}

	return nil
}

func RepoClonePath(dataPath string, u *transport.Endpoint) string {
	return path.Join(dataPath, u.Host, u.Path)
}

func runScc(cloneDest string) (results []SccOutput, err error) {
	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(time.Minute),
	)
	defer cancel()

	cmd := exec.CommandContext(ctx, "scc", "-f", "json", cloneDest)

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create pipe to scc: %w", err)
	}

	defer pipe.Close()

	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("start scc: %w", err)
	}

	decoder := json.NewDecoder(pipe)

	if err = decoder.Decode(&results); err != nil {
		defer func() {
			if errScc := cmd.Cancel(); err != nil {
				err = errors.Join(err, fmt.Errorf("cancel scc after start: %w", errScc))
			}
		}()

		return nil, fmt.Errorf("decode scc output: %w", err)
	}

	if err = cmd.Wait(); err != nil {
		defer func() {
			if errScc := cmd.Cancel(); err != nil {
				err = errors.Join(errScc, fmt.Errorf("scc wait: %w", errScc))
			}
		}()

		return nil, fmt.Errorf("wait for scc: %w", err)
	}

	return results, nil
}

func generateSCCTable(result []SccOutput, repo string) string {
	var (
		builder = &strings.Builder{}
		tbl     = tablewriter.NewWriter(builder)
		rows    = make([][]string, len(result))
		counts  SccOutput
	)

	builder.WriteString("```md\n")

	for i, res := range result {
		rows[i] = []string{
			res.Name,
			strconv.Itoa(res.Count),
			strconv.Itoa(res.Blank),
			strconv.Itoa(res.Comment),
			strconv.Itoa(res.Code),
			strconv.Itoa(res.Lines),
		}
		counts.Count += res.Count
		counts.Blank += res.Blank
		counts.Comment += res.Comment
		counts.Code += res.Code
		counts.Lines += res.Lines
	}

	tbl.SetHeader([]string{"Language", "Files", "Blanks", "Comments", "Code", "Lines"})
	tbl.SetFooter([]string{
		"Total",
		strconv.Itoa(counts.Count),
		strconv.Itoa(counts.Blank),
		strconv.Itoa(counts.Comment),
		strconv.Itoa(counts.Code),
		strconv.Itoa(counts.Lines),
	})

	tbl.SetCaption(true, fmt.Sprintf("Source line count from: %s", repo))
	tbl.AppendBulk(rows)

	tbl.Render()
	builder.WriteString("\n```")

	return builder.String()
}

func CleanAfterClone(clonedAt string) {
	switch err := os.RemoveAll(clonedAt); {
	case errors.Is(err, os.ErrNotExist):
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

var ErrNotHTTP = errors.New("illegal protocol")

func Handle(s *discordgo.Session, i *discordgo.InteractionCreate) (string, error) {
	repo := i.ApplicationCommandData().Options[0].StringValue()

	if err := helpers.Defer(s, i); err != nil {
		return "", err
	}

	parsedURL, err := transport.NewEndpoint(repo)
	if err != nil {
		return "Invalid repository URL", fmt.Errorf("parse endpoint: %w", err)
	}

	switch parsedURL.Protocol {
	case "http", "https":
	default:
		return "Files are not allowed", ErrNotHTTP
	}

	if err = CheckRepoExists(repo); err != nil {
		return "Repository does not exist", fmt.Errorf("check repository: %w", err)
	}

	cloneDest := RepoClonePath("_data", parsedURL)
	defer func() {
		CleanAfterClone(cloneDest)
	}()

	log.Println("Cloning into:", cloneDest)

	ctxClone, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(time.Minute),
	)
	defer cancel()

	if _, err = git.PlainCloneContext(ctxClone, cloneDest, false, &git.CloneOptions{
		URL:      repo,
		Depth:    1,
		Progress: os.Stderr,
		Tags:     git.TagFollowing,
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return "Took too long to clone the repo", fmt.Errorf("clone deadline: %w", err)
		}

		return "Something went wrong, my bad...", fmt.Errorf("clone: %w", err)
	}

	result, err := runScc(cloneDest)
	if err != nil {
		return "Something went wrong, my bad...", err
	}

	generatedTable := generateSCCTable(result, repo)

	return generatedTable, err
}
