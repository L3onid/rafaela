package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"io"
	"fmt"
	"github.com/rylio/ytdl"
	"time"
	"errors"
	"net/http"
	"google.golang.org/api/youtube/v3"
	"google.golang.org/api/googleapi/transport"
	"flag"
	"strconv"
	"strings"
)

func GetArgs(conteudo string) (string){
	args := strings.Fields(conteudo[len("/"):])
	invoked := args[0]
	args = args[1:]
	argstr := conteudo[len("/")+len(invoked):]
	if argstr != "" {
		argstr = argstr[1:]
	}
	return argstr
}

type Info struct {
	vinfo *ytdl.VideoInfo
	msg *discordgo.MessageCreate
}

const developerKey = "my youtube api key"

func Play(s *discordgo.Session, m *discordgo.MessageCreate) (error){
	args := GetArgs(m.Content)
	if args == ""{
		return errors.New("Specify a video.")
	}

	videoInfo, err := ytdl.GetVideoInfo(args)
	if err != nil {
		oi, err := SearchYoutube(args)
		if err != nil {
			return err
		}
		_, _ = s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{Footer: &discordgo.MessageEmbedFooter{Text: "In response to " + m.Author.Username + "#" + m.Author.Discriminator,}, Fields: []*discordgo.MessageEmbedField{{Name:   "Results", Value: "What video do you want? Tell me the number.\n\n" + oi, Inline: true,},},})
		//in progress (when the user answers a number, the song will be added)
		return nil
	}
	r, err := http.Head("https://www.youtube.com/watch?v=" + videoInfo.ID)
	if err != nil || (r.StatusCode != 200){
		oi, err := SearchYoutube(args)
		if err != nil {
			return err
		}
		_, _ = s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{Footer: &discordgo.MessageEmbedFooter{Text: "In response to " + m.Author.Username + "#" + m.Author.Discriminator,}, Fields: []*discordgo.MessageEmbedField{{Name:   "Results", Value: "What video do you want? Tell me the number.\n\n" + oi, Inline: true,},},})
		//in progress (when the user answers a number, the song will be added)
		return nil
	}

	err = AddPlaylist(Info{videoInfo, m}, s)
	return err
}

var (
	PlaylistMap map[string][]Info = make(map[string][]Info)
)


func AddPlaylist(info Info, s *discordgo.Session) (error){
	c, _ := s.Channel(info.msg.ChannelID)
	guild, _ := s.Guild(c.GuildID)
	if len(guild.VoiceStates) == 0{
		return errors.New("It is necessary that there be voice channels on the server so that I can enter.")
	}
	for _, vs := range guild.VoiceStates {
		if vs.UserID == info.msg.Author.ID {
			userP, _ := s.UserChannelPermissions(s.State.User.ID, vs.ChannelID)
			if userP&discordgo.PermissionVoiceConnect == 0 {
				return errors.New("I am not allowed to connect to this channel.")
			}
			if userP&discordgo.PermissionVoiceSpeak == 0 {
				return errors.New("I am not allowed to speak in this channel.")
			}
			if _, ok := PlaylistMap[c.GuildID]; !ok {
				PlaylistMap[c.GuildID] = append(PlaylistMap[c.GuildID], info)
				vc, _ := s.ChannelVoiceJoin(c.GuildID, vs.ChannelID, false, true)
				err := PlayPlaylist(s, vc, guild)
				return err
			}
			PlaylistMap[c.GuildID] = append(PlaylistMap[c.GuildID], info)
			user, _ := s.User(info.msg.Author.ID)
			_, _ = s.ChannelMessageSendEmbed(info.msg.ChannelID, &discordgo.MessageEmbed{Footer: &discordgo.MessageEmbedFooter{Text: "In response to " + info.msg.Author.Username + "#" + info.msg.Author.Discriminator,}, Fields: []*discordgo.MessageEmbedField{{Name:   "Added!", Value:  "**" + info.vinfo.Title + "**, sent by " + user.Mention() + ", was added to the playlist.", Inline: true,},},})
			return nil
		}
	}
	return errors.New("You must be connected to some voice channel so I can enter.")
}

func PlayPlaylist(s *discordgo.Session, vc *discordgo.VoiceConnection, c *discordgo.Guild) (error) {
	for {
		if len(PlaylistMap[c.ID]) != 0 {
			err := PlayMusic(s, vc, PlaylistMap[c.ID][0])
			if err != nil {
				s.ChannelMessageSendEmbed(PlaylistMap[c.ID][0].msg.ChannelID, &discordgo.MessageEmbed{Footer: &discordgo.MessageEmbedFooter{Text: "In response to " + PlaylistMap[c.ID][0].msg.Author.Username + "#" + PlaylistMap[c.ID][0].msg.Author.Discriminator,}, Fields: []*discordgo.MessageEmbedField{{Name: ":red_circle: Error!", Value: err.Error(), Inline: false,},},})
			}
			if len(PlaylistMap[c.ID]) == 1{
				_, _ = s.ChannelMessageSendEmbed(PlaylistMap[c.ID][0].msg.ChannelID, &discordgo.MessageEmbed{Fields: []*discordgo.MessageEmbedField{{Name: "My playlist is over!", Value: "Disconnecting from the voice channel...", Inline: true,},},})
				delete(PlaylistMap, c.ID)
				time.Sleep(time.Second * 1)
				vc.Disconnect()
				return nil
			}
			PlaylistMap[c.ID] = append(PlaylistMap[c.ID][:0], PlaylistMap[c.ID][0+1:]...)
		}
	}
}

func PlayMusic(s *discordgo.Session, vc *discordgo.VoiceConnection, info Info) (error){
	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 64
	options.Application = dca.AudioApplicationLowDelay

	format := info.vinfo.Formats.Best(ytdl.FormatAudioBitrateKey)[0]
	downloadURL, err := info.vinfo.GetDownloadURL(format)
	if err != nil {
		errors.New("There was an error downloading the video.")
	}
	encodingSession, err := dca.EncodeFile(downloadURL.String(), options)
	if err != nil {
		errors.New("There was an error encoding audio data.")
	}
	defer encodingSession.Cleanup()
	done := make(chan error)
	user, _ := s.User(info.msg.Author.ID)
	_, err = s.ChannelMessageSendEmbed(info.msg.ChannelID, &discordgo.MessageEmbed{Footer: &discordgo.MessageEmbedFooter{Text: "In response to " + info.msg.Author.Username + "#" + info.msg.Author.Discriminator,}, Fields: []*discordgo.MessageEmbedField{{Name:   "Broadcasting...", Value:  "Now playing " + "[" + info.vinfo.Title + "](https://www.youtube.com/watch?v=" + info.vinfo.ID + ")" + " for " + info.vinfo.Duration.String() + ", sent by " + user.Mention() + ".", Inline: true,},}, Thumbnail: &discordgo.MessageEmbedThumbnail{URL: info.vinfo.GetThumbnailURL(ytdl.ThumbnailQualityHigh).String(),},})
	dca.NewStream(encodingSession, vc, done)
	err = <-done
	if err != nil && err != io.EOF {
		fmt.Println("Error", err)
		errors.New("Some error occurred while starting the broadcast.")
	}
	return nil
}

func SearchYoutube(s string) (string, error){
	var (
		query      = flag.String("query", s, "Search term")
		maxResults = flag.Int64("max-results", 5, "Max YouTube results")
	)
	client := &http.Client{
		Transport: &transport.APIKey{Key: developerKey},
	}
	service, err := youtube.New(client)
	if err != nil {
		return "", errors.New("Error trying to connect to YouTube.")
	}
	call := service.Search.List("id,snippet").
		Q(*query).
		MaxResults(*maxResults)
	response, err := call.Do()
	if len(response.Items) == 0{
		return "", errors.New("I did not find anything related to what you said.")
	}
	var videos string
	var i int = 1
	for _, item := range response.Items {
		videos += "[" +strconv.Itoa(i) + "] [" + item.Snippet.Title + "]" + "(https://www.youtube.com/watch?v="+item.Id.VideoId+")\n"
		if i == len(response.Items){
			break
		}
		i++
	}
	return videos, nil
}
