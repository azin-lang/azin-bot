package miscellaneous

import "github.com/bwmarrin/discordgo"

var Commands = []*discordgo.ApplicationCommand{StatsCommand}

func Handle(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if i.Type != discordgo.InteractionApplicationCommand {
		return false
	}
	if i.ApplicationCommandData().Name != "stats" {
		return false
	}
	handleStats(s, i)
	return true
}
