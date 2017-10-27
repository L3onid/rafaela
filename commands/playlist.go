package commands

import (
	"github.com/bwmarrin/discordgo"
	"errors"
	"strconv"
	"time"
	"fmt"
)

func Playlist(s *discordgo.Session, m *discordgo.MessageCreate) (error){
	c, _ := s.Channel(m.ChannelID)
	if _, ok := PlaylistMap[c.GuildID]; !ok {
		return errors.New("There is no playlist in progress.")
	}
	var (
		list		string
		duration 	time.Duration
		n		int = 1
		fields		[]*discordgo.MessageEmbedField
	)
	if len(PlaylistMap[c.GuildID]) > 1 {
		for i := 1; i < len(PlaylistMap[c.GuildID]); i++ {
			list += "[" + strconv.Itoa(i) + "] " + "[" + PlaylistMap[c.GuildID][i].vinfo.Title + "](https://www.youtube.com/watch?v=" + PlaylistMap[c.GuildID][i].vinfo.ID + ")" + " (" + PlaylistMap[c.GuildID][i].vinfo.Duration.String() + ")\n"
			if i == 10{
				break
			}
		}
		for _, vs := range PlaylistMap[c.GuildID] {
			duration += vs.vinfo.Duration
		}
		n = len(PlaylistMap[c.GuildID])
		fields = []*discordgo.MessageEmbedField{{Name: "Playlist", Value: list, Inline: false,},{Name: "Current audio", Value: "["+PlaylistMap[c.GuildID][0].vinfo.Title+"](https://www.youtube.com/watch?v=" + PlaylistMap[c.GuildID][0].vinfo.ID + ")",  Inline: false,},{Name: "Duration", Value: duration.String(), Inline: true,},{Name: "Size", Value: strconv.Itoa(n), Inline: true,}}
	}else {
		duration = PlaylistMap[c.GuildID][0].vinfo.Duration
		fields = []*discordgo.MessageEmbedField{{Name: "Current audio", Value: "[" + PlaylistMap[c.GuildID][0].vinfo.Title + "](https://www.youtube.com/watch?v=" + PlaylistMap[c.GuildID][0].vinfo.ID + ")", Inline: false,}, {Name: "Duration", Value: duration.String(), Inline: true,}, {Name: "Size", Value: strconv.Itoa(n), Inline: true,}}
	}
	msg, err := s.ChannelMessageSendEmbed(c.ID, &discordgo.MessageEmbed{Footer: &discordgo.MessageEmbedFooter{Text: "In response to " + m.Author.Username + "#" + m.Author.Discriminator,}, Fields: fields,})
	if err != nil{
		fmt.Println(err.Error())
	}
	if len(PlaylistMap[c.GuildID]) > 11 {
		s.MessageReactionAdd(msg.ChannelID, msg.ID, "◀")
		s.MessageReactionAdd(msg.ChannelID, msg.ID, "▶")
	}
	return nil
}
