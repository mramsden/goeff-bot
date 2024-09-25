package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/mramsden/goeff-bot/presence"
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
	dg.AddHandler(connected)
	dg.Identify.Intents = discordgo.IntentsGuildVoiceStates

	err = dg.Open()
	if err != nil {
		log.Fatal("error opening connection", err)
	}

	presence.Start(context.Background(), func(m presence.Member) {
		channel, err := dg.Channel(m.ChannelID)
		if err != nil {
			if restErr, ok := err.(*discordgo.RESTError); ok && restErr.Response.StatusCode == 403 {
				return
			}

			log.Println("could not resolve channel:", err)
		}

		member, err := dg.GuildMember(m.GuildID, m.MemberID)
		if err != nil {
			log.Println("could not resolve guild member:", err)
		}

		_, err = dg.ChannelMessageSendEmbed(notifyChannel, &discordgo.MessageEmbed{
			Description: fmt.Sprintf("%s just joined %s", member.DisplayName(), channel.Name),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL:    member.AvatarURL("1024"),
				Width:  512,
				Height: 512,
			},
		})
		if err != nil {
			log.Println("failed sending notification to channel:", err)
		}
	})

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func connected(_ *discordgo.Session, _ *discordgo.Ready) {
	log.Println("connected to discord")
}

func voiceChannelStateUpdate(notifyChannel string) func(*discordgo.Session, *discordgo.VoiceStateUpdate) {
	return func(s *discordgo.Session, state *discordgo.VoiceStateUpdate) {
		if state.ChannelID == "" {
			presence.MemberLeft(state.GuildID, state.UserID)
		} else {
			presence.MemberJoined(state.GuildID, state.ChannelID, state.UserID)
		}
	}
}
