package tickets

import "github.com/bwmarrin/discordgo"

var Commands = []*discordgo.ApplicationCommand{CreateCommand, AddCommand, RemoveCommand}

func Handle(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		switch i.ApplicationCommandData().Name {
		case "ticket-create":
			handleCreate(s, i)
		case "ticket-add":
			handleAdd(s, i)
		case "ticket-remove":
			handleRemove(s, i)
		default:
			return false
		}
	case discordgo.InteractionMessageComponent:
		switch i.MessageComponentData().CustomID {
		case "ticket_close", "ticket_close_reason":
			handleCloseButtons(s, i)
		default:
			return false
		}
	case discordgo.InteractionModalSubmit:
		if i.ModalSubmitData().CustomID != "ticket_close_reason_modal" {
			return false
		}
		handleCloseModal(s, i)
	default:
		return false
	}
	return true
}

func TrackMessage(m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	addParticipant(m.ChannelID, m.Author.ID)
}
