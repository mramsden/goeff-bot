package presence

import (
	"context"
	"testing"
	"time"
)

func TestPresenceReportsForNewMemberID(t *testing.T) {
	guildID := "1"
	channelID := "1"
	memberID := "1"

	ctx := context.Background()
	c := make(chan Member)

	Start(ctx, func(m Member) {
		c <- m
	})

	_ = MemberJoined(guildID, channelID, memberID)
	select {
	case <-c:
		break
	case <-time.After(1 * time.Second):
		t.Fatal("expected a member to be returned from presence monitor")
		return
	}

	_ = MemberJoined(guildID, channelID, memberID)
	select {
	case <-c:
		t.Fatal("did not expect a member to be returned from presence monitor if they have already joined")
		return
	case <-time.After(1 * time.Second):
		break
	}

	MemberLeft(guildID, memberID)
	_ = MemberJoined(guildID, channelID, memberID)
	select {
	case <-c:
		break
	case <-time.After(1 * time.Second):
		t.Fatal("expected a member to be returned from presence monitor")
		return
	}
}

func TestMemberJoinedExpectsNonEmptyChannelID(t *testing.T) {
	if err := MemberJoined("1", "", "1"); err == nil {
		t.Fatal("expected error if channel id is empty")
	}
}
