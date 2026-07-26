package tickets

import "github.com/bwmarrin/discordgo"

var RemoveCommand = &discordgo.ApplicationCommand{
	Name:        "ticket-remove",
	Description: "[Mod Only] Revoke a user's access to this ticket",
	Options: []*discordgo.ApplicationCommandOption{
		{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "User to remove", Required: true},
	},
}

func handleRemove(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isStaff(i.Member.Roles) {
		respondEphemeral(s, i, "Only staff can remove users from tickets.")
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
	found := false
	for _, ow := range ch.PermissionOverwrites {
		if ow.ID == target.ID && ow.Type == discordgo.PermissionOverwriteTypeMember {
			found = true
			break
		}
	}
	if !found {
		respondEphemeral(s, i, "User already cannot access this ticket")
		return
	}

	if err := s.ChannelPermissionDelete(i.ChannelID, target.ID); err != nil {
		respondEphemeral(s, i, "Failed to remove user: "+err.Error())
		return
	}

	respondEphemeral(s, i, target.Mention()+" has been removed from the ticket.")
}
