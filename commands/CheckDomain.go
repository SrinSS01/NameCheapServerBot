package commands

import (
	"NS/config"
	"NS/ns"
	"encoding/xml"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/go-resty/resty/v2"
	"regexp"
	"strings"
)

type CheckDomainCommand struct {
	Command *discordgo.ApplicationCommand
	Config  *config.Config
}

var CheckDomain = CheckDomainCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "checkdomain",
		Description: "check if a domain is available",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "domain",
				Description: "The domain to check",
				Required:    true,
			},
		},
	},
}

func RequestDomainCheck(apiUser, apiKey, userName, clientIP, domain string) (*ns.ApiResponse, error) {
	apiUrl := fmt.Sprintf("https://api.namecheap.com/xml.response?ApiUser=%s&ApiKey=%s&UserName=%s&Command=namecheap.domains.check&ClientIp=%s&DomainList=%s", apiUser, apiKey, userName, clientIP, domain)
	resp, err := resty.New().R().Get(apiUrl)
	if err != nil {
		return nil, fmt.Errorf("Error making the request: %s", err.Error())
	}
	var apiResponse ns.ApiResponse
	body := resp.Body()
	err = xml.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Error parsing the response: %s", err.Error())
	}
	if apiResponse.Status != "OK" {
		return nil, fmt.Errorf("Error in API response: \n```xml\n%s\n```", string(body))
	}
	return &apiResponse, nil
}

func (c *CheckDomainCommand) ExecuteDash(s *discordgo.Session, m *discordgo.MessageCreate, domain string) {
	apiUser := c.Config.ApiUser
	apiKey := c.Config.ApiKey
	userName := c.Config.UserName
	clientIP := c.Config.ClientIP
	matched, err := regexp.MatchString("^\\w+(?:\\.\\w+)+$", strings.TrimSpace(domain))
	if err != nil || !matched {
		_, _ = s.ChannelMessageSendReply(m.ChannelID, "Wrong domain format", m.Reference())
		return
	}
	apiResponse, err := RequestDomainCheck(apiUser, apiKey, userName, clientIP, domain)
	if err != nil {
		_, _ = s.ChannelMessageSendReply(m.ChannelID, err.Error(), m.Reference())
		return
	}
	var msg = m.Message
	for _, result := range apiResponse.CommandResponse.DomainCheckData {
		availability := "available"
		if !result.Available {
			availability = "unavailable"
		}
		content := fmt.Sprintf("Domain: %s, Availability: %s", result.Domain, availability)

		msg, _ = s.ChannelMessageSendReply(msg.ChannelID, content, msg.Reference())
	}
}

func (c *CheckDomainCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var err error
	apiUser := c.Config.ApiUser
	apiKey := c.Config.ApiKey
	userName := c.Config.UserName
	clientIP := c.Config.ClientIP
	options := interaction.ApplicationCommandData().Options
	domain := options[0].StringValue()
	if len(strings.TrimSpace(domain)) == 0 {
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
	_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	apiResponse, err := RequestDomainCheck(apiUser, apiKey, userName, clientIP, domain)
	if err != nil {
		content := err.Error()
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	/*_url := fmt.Sprintf("https://api.namecheap.com/xml.response?ApiUser=%s&ApiKey=%s&UserName=%s&Command=namecheap.domains.check&ClientIp=%s&DomainList=%s", apiUser, apiKey, userName, clientIP, domain)
	// discord defer reply
	err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return
	}
	resp, err := resty.New().R().Get(_url)
	if err != nil {
		content := fmt.Sprintf("Error making the request: %s", err.Error())
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	var apiResponse ns.ApiResponse
	body := resp.Body()
	err = xml.Unmarshal(body, &apiResponse)
	if err != nil {
		content := fmt.Sprintf("Error parsing the response: %s", err.Error())
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	if apiResponse.Status != "OK" {
		content := fmt.Sprintf("Error in API response: \n```xml\n%s\n```", string(body))
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}*/
	var msg *discordgo.Message = nil
	for _, result := range apiResponse.CommandResponse.DomainCheckData {
		availability := "available"
		if !result.Available {
			availability = "unavailable"
		}
		content := fmt.Sprintf("Domain: %s, Availability: %s", result.Domain, availability)

		if msg != nil {
			msg, _ = session.ChannelMessageSendReply(msg.ChannelID, content, msg.Reference())
			continue
		}
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
	}
}
