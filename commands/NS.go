package commands

import (
	"NS/config"
	"NS/ns"
	"encoding/xml"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/go-resty/resty/v2"
	"strings"
)

type NSCommand struct {
	Command *discordgo.ApplicationCommand
	Config  *config.Config
}

var NS = NSCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "ns",
		Description: "change the name server",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "domain",
				Description: "The domain to change the name server",
				Required:    true,
			},
		},
	},
}

func (c *NSCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var err error
	apiUser := c.Config.ApiUser
	apiKey := c.Config.ApiKey
	userName := c.Config.UserName
	clientIP := c.Config.ClientIP
	options := interaction.ApplicationCommandData().Options
	domain := options[0].StringValue()
	nameservers := "ns1.swapped1.lat,ns2.swapped1.lat"
	parts := strings.SplitN(domain, ".", 2)

	if len(domain) == 0 || len(parts) != 2 {
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

	sld, tld := parts[0], parts[1]
	_url := fmt.Sprintf("https://api.namecheap.com/xml.response?ApiUser=%s&ApiKey=%s&UserName=%s&Command=namecheap.domains.dns.setCustom&ClientIp=%s&SLD=%s&TLD=%s&NameServers=%s", apiUser, apiKey, userName, clientIP, sld, tld, nameservers)
	resp, err := resty.New().R().Get(_url)
	if err != nil {
		content := fmt.Sprintf("Error making the request: %s", err.Error())
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	var apiResponse ns.ApiResponse
	err = xml.Unmarshal(resp.Body(), &apiResponse)
	if err != nil {
		content := fmt.Sprintf("Error parsing the response: %s", err.Error())
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	if apiResponse.Status == "ERROR" {
		content := fmt.Sprintf("Error in API response: %s", apiResponse.Status)
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	content := fmt.Sprintf("Name server changed successfully")
	_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}
