package commands

import (
	"NS/config"
	"NS/ns"
	"NS/util"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

type CreateEmailCommand struct {
	Command *discordgo.ApplicationCommand
	Config  *config.Config
}

var CreateEmail = CreateEmailCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "createemail",
		Description: "create an email",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "domain",
				Description: "The domain to create the email on",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "email",
				Description: "The email to create",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "password",
				Description: "The password for the email",
			},
		},
	},
}

func (c *CreateEmailCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var err error
	options := interaction.ApplicationCommandData().Options
	domain := options[0].StringValue()
	email := options[1].StringValue()
	password := c.Config.DefaultPassword
	localPart := strings.Split(domain, "@")[0]
	if len(options) == 3 {
		password = options[2].StringValue()
	}
	if len(email) == 0 {
		err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please enter a valid email and password",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			return
		}
		return
	}
	// discord defer reply
	err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return
	}

	emailURL := fmt.Sprintf("https://199.188.203.195:2083/json-api/cpanel?cpanel_jsonapi_func=addpop&cpanel_jsonapi_module=Email&cpanel_jsonapi_version=2&domain=%s&email=%s&password=%s", domain, localPart, password)
	response, err := util.MakeRequest("GET", emailURL, "", nil)
	if err != nil {
		return
	}

	var emailCreateResponse ns.Response
	err = json.Unmarshal(response, &emailCreateResponse)
	if err != nil {
		return
	}

	data := emailCreateResponse.Cpanelresult.Data
	var content string
	if data.Result == "0" {
		content = fmt.Sprintf("Failed to create email account: %s", data.Reason)
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
	} else {
		content = fmt.Sprintf("Successfully created email account: %s", localPart+"@"+domain)
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
	}
}
