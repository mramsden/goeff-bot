package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func main() {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if len(token) == 0 {
		log.Fatal("missing DISCORD_BOT_TOKEN environment variable")
	}

	notifyChannel := os.Getenv("DISCORD_NOTIFY_CHANNEL")
	if len(notifyChannel) == 0 {
		log.Fatal("missing DISCORD_NOTIFY_CHANNEL environment variable")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}

	dg.AddHandler(voiceChannelStateUpdate(notifyChannel))
	dg.Identify.Intents = discordgo.IntentsGuildVoiceStates

	err = dg.Open()
	if err != nil {
		log.Fatal("error opening connection", err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

var connectedMembers []string
var connectedMembersLock sync.RWMutex

func userJoinedServer(s *discordgo.Session, channelID, notifyChannel string, member *discordgo.Member) {
	connectedMembersLock.RLock()
	defer connectedMembersLock.RUnlock()
	if slices.Contains(connectedMembers, member.User.ID) {
		return
	}

	connectedMembersLock.Lock()
	connectedMembers = append(connectedMembers, member.User.ID)
	connectedMembersLock.Unlock()

	channel, err := s.Channel(channelID)
	if err != nil {
		if restErr, ok := err.(*discordgo.RESTError); ok && restErr.Response.StatusCode == 403 {
			return
		}

		log.Println("could not resolve channel:", err)
		return
	}

	memberName := member.DisplayName()
	if len(memberName) == 0 || len(channel.Name) == 0 {
		return
	}

	_, err = s.ChannelMessageSendEmbed(notifyChannel, &discordgo.MessageEmbed{
		Description: fmt.Sprintf("%s just joined %s", memberName, channel.Name),
	})
	if err != nil {
		log.Println("failed sending notification to channel:", err)
	}
}

func userLeftServer(member *discordgo.Member) {
	connectedMembersLock.Lock()
	defer connectedMembersLock.Unlock()

	newConnectedMembers := []string{}
	for _, connectedMember := range connectedMembers {
		if connectedMember != member.User.ID {
			newConnectedMembers = append(newConnectedMembers, member.User.ID)
		}
	}
	connectedMembers = newConnectedMembers
}

func voiceChannelStateUpdate(notifyChannel string) func(*discordgo.Session, *discordgo.VoiceStateUpdate) {
	return func(s *discordgo.Session, state *discordgo.VoiceStateUpdate) {
		if state.ChannelID == "" {
			userLeftServer(state.Member)
		} else {
			userJoinedServer(s, state.ChannelID, notifyChannel, state.Member)
		}
	}
}
