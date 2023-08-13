package commands

import (
	"NS/config"
	"NS/ns"
	"NS/util"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/url"
)

type AddCommand struct {
	Command *discordgo.ApplicationCommand
	Config  *config.Config
}

var Add = AddCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "add",
		Description: "adds the domain as addon in the cpanel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "domain",
				Description: "The domain to add",
				Required:    true,
			},
		},
	},
}

func (c *AddCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var err error
	options := interaction.ApplicationCommandData().Options
	domain := options[0].StringValue()
	if len(domain) == 0 {
		err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please enter a valid domain name",
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
	escapedDomain := url.QueryEscape(domain)
	addonURL := fmt.Sprintf("https://199.188.203.195:2083/json-api/cpanel?cpanel_jsonapi_func=addaddondomain&cpanel_jsonapi_module=AddonDomain&cpanel_jsonapi_version=2&newdomain=%s&subdomain=%s&dir=/home/swapped2/%s", escapedDomain, escapedDomain, escapedDomain)
	response, err := util.MakeRequest("GET", addonURL, "", nil)
	if err != nil {
		content := "Error adding domain: " + err.Error()
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
	}

	var addDomainResponse ns.Response
	err = json.Unmarshal(response, &addDomainResponse)
	if err != nil {
		content := fmt.Sprintf("Error unmarshalling JSON ```json\n%s```: %s", response, err.Error())
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}
	data := addDomainResponse.Cpanelresult.Data
	if data.Result == "0" {
		content := "Failed to add domain: " + data.Reason
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}
	content := "Successfully added domain: " + domain
	_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}
