package commands

import (
	"NS/config"
	"github.com/bwmarrin/discordgo"
)

type ChangePasswordCommand struct {
	Command *discordgo.ApplicationCommand
	Config  *config.Config
}

var ChangePassword = ChangePasswordCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "change-password",
		Description: "Change password",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "email",
				Description: "Email",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "password",
				Description: "Password",
				Required:    true,
			},
		},
	},
}
