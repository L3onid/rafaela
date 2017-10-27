package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"strings"
	"../rafaela/commands"
	"../rafaela/modules"
)

var (
	prefix = "/"
	rcmds map[string]Cmds = make(map[string]Cmds)
)

type Cmds struct {
	Exec func(s *discordgo.Session, m *discordgo.MessageCreate) (error)
	perm int
	args bool
}

func main() {
	dg, err := discordgo.New("token")
	if err != nil {
		fmt.Println("Error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection,", err)
		return
	}
	fmt.Println("The connection was successful. Press CTRL+C to exit.")

	//website.Site()

	rcmds["ping"] = Cmds{comandos.Ping, 0, false}
	rcmds["play"] = Cmds{comandos.Play, 0, true}
	rcmds["playlist"] = Cmds{comandos.Playlist, 0, true}

	if err != nil {
		panic(err)
	}

	<-make(chan struct{})
	return
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot == true{
		return
	}

	//cleverbot
	replace := strings.Replace(m.Content, "!", "", -1)
	if strings.HasPrefix(replace,"<@"+s.State.User.ID+">") && strings.Replace(replace, "<@"+s.State.User.ID+">", "", -1) != ""{
			s.ChannelTyping(m.ChannelID)
			s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+">, " + modulos.Cleverbot(m.Content, m.Author.ID))
	}

	if strings.HasPrefix(m.Content, prefix) {
		args := strings.Fields(m.Content[len(prefix):])
		var invoked string
		if len(args) != 0{
			invoked = args[0]
		}

		if _, ok := rcmds[invoked]; ok {
			s.ChannelTyping(m.ChannelID)
			if rcmds[invoked].perm != 0{
				user, _ := s.UserChannelPermissions(m.Author.ID, m.ChannelID)
				if user&rcmds[invoked].perm == 0 {
					msgerror("You do not have permission to run this command.", invoked, s, m)
					return
				}
			}
			err := rcmds[invoked].Exec(s, m)
			if err != nil {
				msgerror(err.Error(), invoked, s, m)
				return
			}
		}
	}
}

func msgerror(e string, invoked string, s *discordgo.Session, m *discordgo.MessageCreate){
	s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{Footer: &discordgo.MessageEmbedFooter{Text: "In response to " + m.Author.Username + "#" + m.Author.Discriminator,}, Fields: []*discordgo.MessageEmbedField{{Name: ":red_circle: Error!", Value:  e, Inline: false,},},})
}
