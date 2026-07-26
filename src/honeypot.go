package main

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	HoneypotChannel = "1531069878281175060"

	imageSpamThreshold    = 4
	imageSpamWindow       = 5 * time.Second
	imageSpamMuteDuration = 2 * time.Minute
)

func HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.ChannelID == HoneypotChannel {
		handleHoneypot(s, m)
		return
	}
	if m.Author.Bot {
		return
	}
	checkImageSpam(s, m)
}

func handleHoneypot(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		return
	}
	if err := s.GuildBanCreateWithReason(m.GuildID, m.Author.ID, "tripped honeypot channel", 0); err != nil {
		log.Printf("failed to ban %s: %v", m.Author.ID, err)
	}
}

var (
	imageMu         sync.Mutex
	imageTimestamps = map[string][]time.Time{}
)

func checkImageSpam(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !hasImage(m.Message) {
		return
	}

	now := time.Now()
	cutoff := now.Add(-imageSpamWindow)

	imageMu.Lock()
	recent := append(imageTimestamps[m.Author.ID], now)
	recent = dropBefore(recent, cutoff)
	count := len(recent)
	if count >= imageSpamThreshold {
		delete(imageTimestamps, m.Author.ID) // reset so it doesn't re-trigger on every message after
	} else {
		imageTimestamps[m.Author.ID] = recent
	}
	imageMu.Unlock()

	if count < imageSpamThreshold {
		return
	}

	until := now.Add(imageSpamMuteDuration)
	if err := s.GuildMemberTimeout(m.GuildID, m.Author.ID, &until); err != nil {
		log.Printf("failed to timeout %s: %v", m.Author.ID, err)
	}

	dm, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		return // DMs closed, nothing more we can do
	}
	s.ChannelMessageSend(dm.ID, "You were muted because you've sent 4 or more images within 5 seconds\n. This is a spam prevention feature, you're not in trouble.")
}

func hasImage(msg *discordgo.Message) bool {
	for _, a := range msg.Attachments {
		if strings.HasPrefix(a.ContentType, "image/") {
			return true
		}
	}
	return false
}

func dropBefore(times []time.Time, cutoff time.Time) []time.Time {
	kept := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	return kept
}
