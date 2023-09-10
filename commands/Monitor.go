package commands

import (
	"NS/config"
	"github.com/bwmarrin/discordgo"
)

type MonitorCommand struct {
	Command *discordgo.ApplicationCommand
	Config  *config.Config
}

var Monitor = MonitorCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "monitor",
		Description: "tracking information for the messages in the account's message queue",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "email",
				Description: "Email to monitor",
				Required:    true,
			},
		},
	},
}

func (m *MonitorCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	//commandData := interaction.ApplicationCommandData()
	//email := commandData.Options[0].StringValue()
	//cPanelUserName := m.Config.CPanelUsername
	//cPanelPassword := m.Config.CPanelPassword
	/*
	   	// You can edit this code!
	      // Click here and start typing.
	      package main

	      import (
	      	"fmt"
	      	"os"
	      	"os/signal"
	      	"syscall"
	      	"time"
	      )

	      func main() {
	      	task(time.Now())
	      	tick := time.NewTicker(time.Second * 5)
	      	scheduler(tick)
	      	sigs := make(chan os.Signal, 1)
	      	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	      	<-sigs
	      	tick.Stop()
	      }

	      func scheduler(tick *time.Ticker) {
	      	for t := range tick.C {
	      		task(t)
	      	}
	      }

	      func task(t time.Time) {
	      	fmt.Println("hello! printed at ", t)
	      }
	*/
}
