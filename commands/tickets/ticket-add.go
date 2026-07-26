package tickets

import "github.com/bwmarrin/discordgo"

var AddCommand = &discordgo.ApplicationCommand{
	Name:        "ticket-add",
	Description: "[Mod Only] Give a user access to this ticket",
	Options: []*discordgo.ApplicationCommandOption{
		{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "User to add", Required: true},
	},
}

func handleAdd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isStaff(i.Member.Roles) {
		respondEphemeral(s, i, "Only staff can add users to tickets.")
		return
	}
	if _, ok := getTicket(i.ChannelID); !ok {
		respondEphemeral(s, i, "This isn't a ticket channel.")
		return
	}

	target := i.ApplicationCommandData().Options[0].UserValue(s)

	ch, err := s.Channel(i.ChannelID)
	if err != nil {
		respondEphemeral(s, i, "Failed to look up channel: "+err.Error())
		return
	}
	for _, ow := range ch.PermissionOverwrites {
		if ow.ID == target.ID && ow.Type == discordgo.PermissionOverwriteTypeMember {
			respondEphemeral(s, i, "User already has access to this ticket")
			return
		}
	}

	err = s.ChannelPermissionSet(i.ChannelID, target.ID, discordgo.PermissionOverwriteTypeMember,
		discordgo.PermissionViewChannel|discordgo.PermissionSendMessages, 0)
	if err != nil {
		respondEphemeral(s, i, "Failed to add user: "+err.Error())
		return
	}

	addParticipant(i.ChannelID, target.ID)
	respondEphemeral(s, i, target.Mention()+" has been added to the ticket.")
}
