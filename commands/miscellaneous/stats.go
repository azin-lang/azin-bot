package miscellaneous

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

var StatsCommand = &discordgo.ApplicationCommand{
	Name:        "stats",
	Description: "Show bot resource usage and response time",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionBoolean,
			Name:        "ephemeral",
			Description: "Only show the result to you (default: false)",
			Required:    false,
		},
	},
}

func handleStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	start := time.Now()

	ephemeral := false
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == "ephemeral" {
			ephemeral = opt.BoolValue()
		}
	}

	cpuPct, cpuErr := processCPUPercent()
	ramMB, ramErr := processRSSMB()

	embed := &discordgo.MessageEmbed{
		Title: "I'm alive!",
		Color: 0x57f287,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "CPU Usage", Value: fmtOrUnavailable(fmt.Sprintf("%.1f%%", cpuPct), cpuErr), Inline: true},
			{Name: "RAM Usage", Value: fmtOrUnavailable(fmt.Sprintf("%.1f MB", ramMB), ramErr), Inline: true},
			{Name: "Response Time", Value: time.Since(start).Round(time.Millisecond).String(), Inline: true},
		},
	}

	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  flags,
		},
	})
}

func fmtOrUnavailable(val string, err error) string {
	if err != nil {
		return "unavailable"
	}
	return val
}
