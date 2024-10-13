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
			log.Println("could not resolve channel:", err)
			return
		}

		gm, err := dg.GuildMember(m.GuildID, m.MemberID)
		if err != nil {
			log.Println("could not resolve guild gm:", err)
			return
		}

		// do not announce bot users joining a channel
		if gm.User.Bot {
			return
		}

		message := &discordgo.MessageSend{
			Content: fmt.Sprintf("%s just joined %s", gm.DisplayName(), channel.Name),
			Embed: &discordgo.MessageEmbed{
				Fields: []*discordgo.MessageEmbedField{
					{Name: "Member", Value: fmt.Sprintf("<@%s>", m.MemberID)},
					{Name: "Channel", Value: fmt.Sprintf("<#%s>", m.ChannelID)},
				},
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL:    gm.AvatarURL("1024"),
					Width:  512,
					Height: 512,
				},
			},
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse:       []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeUsers},
				Users:       []string{},
				RepliedUser: false,
			},
		}
		_, err = dg.ChannelMessageSendComplex(notifyChannel, message)
		if err != nil {
			log.Println("failed sending notification to channel:", err)
		}
	}
}
