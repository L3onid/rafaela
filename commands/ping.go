package commands

import(
	"github.com/bwmarrin/discordgo"
	"github.com/sparrc/go-ping"
	"time"
	"fmt"
	"errors"
)

func Ping(s *discordgo.Session, m *discordgo.MessageCreate) (error){
	pinger, err := ping.NewPinger("104.16.60.37")
	if err != nil {
		print(err.Error())
		return errors.New("There was an error while trying to connect to the server.")
	}
	pinger.Count = 1
	pinger.Run()
	ms := int64(pinger.Statistics().AvgRtt / time.Millisecond)
	s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Footer: &discordgo.MessageEmbedFooter{
			Text: "In response to " + m.Author.Username + "#" + m.Author.Discriminator,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Pong!",
				Value:  "Response time: " + fmt.Sprint(ms) + " milliseconds.",
				Inline: true,
			},
		},
	})
	return nil
}
