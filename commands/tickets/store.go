package tickets

import (
	"sync"
	"time"
)

const (
	SupportCat   = "1530905504358994021"
	FoundersRole = "1518351827924553980"
	ModRole      = "1523467087538950294"
)

type Ticket struct {
	CreatorID    string
	CreatedAt    time.Time
	Participants map[string]bool
}

var (
	mu      sync.Mutex
	tickets = map[string]*Ticket{}
)

func getTicket(channelID string) (*Ticket, bool) {
	mu.Lock()
	defer mu.Unlock()
	t, ok := tickets[channelID]
	return t, ok
}

func addTicket(channelID, creatorID string) {
	mu.Lock()
	defer mu.Unlock()
	tickets[channelID] = &Ticket{
		CreatorID:    creatorID,
		CreatedAt:    time.Now(),
		Participants: map[string]bool{creatorID: true},
	}
}

func addParticipant(channelID, userID string) {
	mu.Lock()
	defer mu.Unlock()
	if t, ok := tickets[channelID]; ok {
		t.Participants[userID] = true
	}
}

func removeTicket(channelID string) {
	mu.Lock()
	defer mu.Unlock()
	delete(tickets, channelID)
}

func isStaff(roleIDs []string) bool {
	for _, r := range roleIDs {
		if r == FoundersRole || r == ModRole {
			return true
		}
	}
	return false
}
