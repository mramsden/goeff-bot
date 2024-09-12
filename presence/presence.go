package presence

import (
	"context"
	"errors"
)

type connectedMember struct {
	id        string
	channelID string
}

type Member struct {
	GuildID   string
	MemberID  string
	ChannelID string
}

var guildMembers map[string][]connectedMember
var updateStream chan Member

func init() {
	guildMembers = make(map[string][]connectedMember)
	updateStream = make(chan Member, 10)
}

func Start(ctx context.Context, memberJoined func(Member)) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-updateStream:
				if len(update.ChannelID) > 0 {
					if addMember(update) {
						memberJoined(update)
					}
					break
				}
				_ = removeMember(update)
			}
		}
	}()
}

func addMember(update Member) bool {
	members := guildMembers[update.GuildID]

	for i, connected := range members {
		if connected.id == update.MemberID {
			members[i] = connectedMember{
				id:        update.MemberID,
				channelID: update.ChannelID,
			}
			guildMembers[update.GuildID] = members
			return false
		}
	}

	guildMembers[update.GuildID] = append(members, connectedMember{
		id:        update.MemberID,
		channelID: update.ChannelID,
	})

	return true
}

func removeMember(update Member) bool {
	members := guildMembers[update.GuildID]
	for i, member := range members {
		if member.id == update.MemberID {
			guildMembers[update.GuildID] = append(members[:i], members[i+1:]...)
			return true
		}
	}

	return false
}

func MemberJoined(guildID, channelID, memberID string) error {
	if len(channelID) == 0 {
		return errors.New("presence: channel should be a non-empty numerical id string")
	}

	updateStream <- Member{
		GuildID:   guildID,
		ChannelID: channelID,
		MemberID:  memberID,
	}

	return nil
}

func MemberLeft(guildID, memberID string) {
	updateStream <- Member{
		GuildID:  guildID,
		MemberID: memberID,
	}
}
