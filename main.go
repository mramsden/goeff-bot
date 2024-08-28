package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
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

func voiceChannelStateUpdate(notifyChannel string) func(*discordgo.Session, *discordgo.VoiceStateUpdate) {
	return func(s *discordgo.Session, state *discordgo.VoiceStateUpdate) {
		channel, err := s.Channel(state.ChannelID)
		if err != nil {
			if restErr, ok := err.(*discordgo.RESTError); ok && restErr.Response.StatusCode == 403 {
				return
			}

			log.Println("could not resolve channel:", err)
			return
		}

		memberName := state.Member.DisplayName()
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
}
