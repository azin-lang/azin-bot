package tickets

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var CreateCommand = &discordgo.ApplicationCommand{
	Name:        "ticket-create",
	Description: "Open a support ticket",
}

func handleCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := i.Member.User

	overwrites := []*discordgo.PermissionOverwrite{
		{ID: i.GuildID, Type: discordgo.PermissionOverwriteTypeRole, Deny: discordgo.PermissionViewChannel},
		{ID: user.ID, Type: discordgo.PermissionOverwriteTypeMember, Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages},
		{ID: FoundersRole, Type: discordgo.PermissionOverwriteTypeRole, Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages},
		{ID: ModRole, Type: discordgo.PermissionOverwriteTypeRole, Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages},
	}

	ch, err := s.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
		Name:                 fmt.Sprintf("ticket-%s", strings.ToLower(user.Username)),
		Type:                 discordgo.ChannelTypeGuildText,
		ParentID:             SupportCat,
		PermissionOverwrites: overwrites,
	})
	if err != nil {
		respondEphemeral(s, i, "Failed to create ticket channel: "+err.Error())
		return
	}

	addTicket(ch.ID, user.ID)

	_, err = s.ChannelMessageSendComplex(ch.ID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Description: "A founder/mod will be with you shortly, describe your issue below and then wait.",
			Color:       0x2b2d31,
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{Components: []discordgo.MessageComponent{
				discordgo.Button{Label: "Close", Style: discordgo.DangerButton, CustomID: "ticket_close"},
				discordgo.Button{Label: "Close with Reason", Style: discordgo.SecondaryButton, CustomID: "ticket_close_reason"},
			}},
		},
	})
	if err != nil {
		removeTicket(ch.ID)
		respondEphemeral(s, i, "Channel created but the intro message failed to send: "+err.Error())
		return
	}

	respondEphemeral(s, i, "Ticket created: "+ch.Mention())
}

func handleCloseButtons(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isStaff(i.Member.Roles) {
		respondEphemeral(s, i, "Only staff can close tickets.")
		return
	}

	if i.MessageComponentData().CustomID == "ticket_close" {
		closeTicket(s, i, "")
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "ticket_close_reason_modal",
			Title:    "Close Ticket",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:  "reason",
						Label:     "Reason",
						Style:     discordgo.TextInputParagraph,
						Required:  true,
						MaxLength: 500,
					},
				}},
			},
		},
	})
}

func handleCloseModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isStaff(i.Member.Roles) {
		respondEphemeral(s, i, "Only staff can close tickets.")
		return
	}
	row := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow)
	reason := row.Components[0].(*discordgo.TextInput).Value
	closeTicket(s, i, reason)
}

func closeTicket(s *discordgo.Session, i *discordgo.InteractionCreate, reason string) {
	channelID := i.ChannelID
	t, ok := getTicket(channelID)
	if !ok {
		respondEphemeral(s, i, "This isn't a ticket channel.")
		return
	}

	if reason == "" {
		reason = "No reason provided"
	}

	participants := make([]string, 0, len(t.Participants))
	for id := range t.Participants {
		participants = append(participants, "<@"+id+">")
	}

	embed := &discordgo.MessageEmbed{
		Title: "Ticket Closed",
		Color: 0x2b2d31,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Created", Value: fmt.Sprintf("<t:%d:F>", t.CreatedAt.Unix())},
			{Name: "Closed", Value: fmt.Sprintf("<t:%d:F>", time.Now().Unix())},
			{Name: "Closed By", Value: i.Member.User.Mention()},
			{Name: "Participants", Value: strings.Join(participants, ", ")},
			{Name: "Reason", Value: reason},
		},
	}

	for id := range t.Participants {
		dm, err := s.UserChannelCreate(id)
		if err != nil {
			continue // DMs closed or user left the server, nothing more we can do
		}
		s.ChannelMessageSendEmbed(dm.ID, embed)
	}

	removeTicket(channelID)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "Closing ticket..."},
	})
	s.ChannelDelete(channelID)
}
