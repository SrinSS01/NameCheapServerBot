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

func (m MonitorCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	//commandData := interaction.ApplicationCommandData()
	//email := commandData.Options[0].StringValue()
	//cPanelUserName := m.Config.CPanelUsername
	//cPanelPassword := m.Config.CPanelPassword

}
