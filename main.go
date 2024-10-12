package main

import (
	"context"
	"errors"
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

	dg.AddHandler(voiceChannelStateUpdate)
	dg.AddHandler(connected)
	dg.Identify.Intents = discordgo.IntentsGuildVoiceStates

	err = dg.Open()
	if err != nil {
		log.Fatal("error opening connection", err)
	}

	presence.Start(context.Background(), makePresenceHandler(dg, notifyChannel))

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	if err := dg.Close(); err != nil {
		log.Printf("failed closing discord session %s", err)
	}
}

// Discord event handlers

func connected(_ *discordgo.Session, _ *discordgo.Ready) {
	log.Println("connected to discord")
}

func voiceChannelStateUpdate(_ *discordgo.Session, state *discordgo.VoiceStateUpdate) {
	if state.ChannelID == "" {
		presence.MemberLeft(state.GuildID, state.UserID)
	} else {
		_ = presence.MemberJoined(state.GuildID, state.ChannelID, state.UserID)
	}
}

// Presence handlers

func makePresenceHandler(dg *discordgo.Session, notifyChannel string) func(presence.Member) {
	return func(m presence.Member) {
		channel, err := dg.Channel(m.ChannelID)
		if err != nil {
			var restErr *discordgo.RESTError
			if errors.As(err, &restErr) && restErr.Response.StatusCode == 403 {
				return
			}

			log.Println("could not resolve channel:", err)
		}

		member, err := dg.GuildMember(m.GuildID, m.MemberID)
		if err != nil {
			log.Println("could not resolve guild member:", err)
			return
		}

		// do not announce bot users joining a channel
		if member.User.Bot {
			return
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
	}
}
