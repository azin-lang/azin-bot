package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"

	"azin/commands/miscellaneous"
	"azin/commands/tickets"
)

// commands sync faster, global syncing is too damn long
const GuildID = "1518351193473024062"

func main() {
	token := os.Getenv("TOKEN")

	s, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("failed to create session: ", err)
	}
	s.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if tickets.Handle(s, i) {
			return
		}
		miscellaneous.Handle(s, i)
	})
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		tickets.TrackMessage(m)
		HandleMessage(s, m)
	})

	if err := s.Open(); err != nil {
		log.Fatal("failed to open connection: ", err)
	}
	defer s.Close()

	registerCommands(s)

	log.Println("Azin is running, yay.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}

func registerCommands(s *discordgo.Session) {
	var all []*discordgo.ApplicationCommand
	all = append(all, tickets.Commands...)
	all = append(all, miscellaneous.Commands...)

	for _, cmd := range all {
		if _, err := s.ApplicationCommandCreate(s.State.User.ID, GuildID, cmd); err != nil {
			log.Printf("failed to register %s: %v", cmd.Name, err)
		}
	}
}
