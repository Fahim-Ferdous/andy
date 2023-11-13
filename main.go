/*
Copyright © 2023 Fahim Ferdous

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"andy/slash"

	"github.com/bwmarrin/discordgo"
)

func main() {
	var (
		guildID = flag.String(
			"guild",
			"",
			"Test guild ID. If not passed, bot registered commands globally.",
		)
		botToken       = flag.String("token", "", "Bot access token.")
		removeCommands = flag.Bool("removeCommands", true, "Remove all commands on exit.")
	)

	flag.Parse()

	var (
		dgs *discordgo.Session
		err error
	)

	dgs, err = discordgo.New("Bot " + *botToken)
	if err != nil {
		log.Fatalf("Invalid bot token: %v", err)
	}

	slash.AddHandlers(dgs)

	if err = dgs.Open(); err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	defer dgs.Close()

	registeredCommands := slash.RegisterCommands(*guildID, dgs)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if *removeCommands {
		log.Println("Removing commands...")

		for _, v := range registeredCommands {
			err := dgs.ApplicationCommandDelete(dgs.State.User.ID, *guildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}

	log.Println("Gracefully shutting down.")
}
