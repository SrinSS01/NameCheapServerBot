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
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "nameservers",
				Description: "The name servers to set (comma separated values)",
			},
		},
	},
}

func RequestNameServerChange(apiUser, apiKey, userName, clientIP, sld, tld, nameservers string) (string, error) {
	apiUrl := fmt.Sprintf("https://api.namecheap.com/xml.response?ApiUser=%s&ApiKey=%s&UserName=%s&Command=namecheap.domains.dns.setCustom&ClientIp=%s&SLD=%s&TLD=%s&NameServers=%s", apiUser, apiKey, userName, clientIP, sld, tld, nameservers)
	resp, err := resty.New().R().Get(apiUrl)
	if err != nil {
		return "", fmt.Errorf("error making the request: %s", err.Error())
	}
	var apiResponse ns.ApiResponse
	err = xml.Unmarshal(resp.Body(), &apiResponse)
	if err != nil {
		return "", fmt.Errorf("error parsing the response: %s", err.Error())
	}

	if apiResponse.Status == "ERROR" {
		return "", fmt.Errorf("error in API response: %s", apiResponse.Status)
	}
	return "Name server changed successfully", nil
}

var nsRegex, _ = regexp.Compile("^(?P<domain>\\w+(?:\\.\\w+)+) +(?P<ns>(\\w+(?:\\.\\w+)+)(?:(?:, *| +)(\\w+(?:\\.\\w+)+))*)$")
var nsR = regexp.MustCompile("\\w+(?:\\.\\w+)+")

func (c *NSCommand) ExecuteDash(session *discordgo.Session, messageCreate *discordgo.MessageCreate, args string) {
	apiUser := c.Config.ApiUser
	apiKey := c.Config.ApiKey
	userName := c.Config.UserName
	clientIP := c.Config.ClientIP
	matches := nsRegex.FindStringSubmatch(args)
	if len(matches) == 0 {
		_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, "Please provide valid arguments", messageCreate.Reference())
		return
	}
	domain := matches[nsRegex.SubexpIndex("domain")]
	nameServers := matches[nsRegex.SubexpIndex("ns")]
	nsString := strings.Join(nsR.FindAllString(nameServers, -1), ",")
	parts := strings.Split(domain, ".")
	sld, tld := parts[0], parts[1]
	res, err := RequestNameServerChange(apiUser, apiKey, userName, clientIP, sld, tld, nsString)
	if err != nil {
		_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, err.Error(), messageCreate.Reference())
		return
	}
	_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, res, messageCreate.Reference())
}
func (c *NSCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var err error
	apiUser := c.Config.ApiUser
	apiKey := c.Config.ApiKey
	userName := c.Config.UserName
	clientIP := c.Config.ClientIP
	options := interaction.ApplicationCommandData().Options
	domain := options[0].StringValue()
	nameservers := c.Config.DefaultNameServers
	parts := strings.SplitN(domain, ".", 2)

	if len(options) > 1 {
		nameservers = options[1].StringValue()
	}

	if len(strings.TrimSpace(domain)) == 0 || len(parts) != 2 {
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
	if len(strings.TrimSpace(nameservers)) == 0 {
		err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please enter a valid name servers",
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
	res, err := RequestNameServerChange(apiUser, apiKey, userName, clientIP, sld, tld, nameservers)
	if err != nil {
		content := err.Error()
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}
	/*_url := fmt.Sprintf("https://api.namecheap.com/xml.response?ApiUser=%s&ApiKey=%s&UserName=%s&Command=namecheap.domains.dns.setCustom&ClientIp=%s&SLD=%s&TLD=%s&NameServers=%s", apiUser, apiKey, userName, clientIP, sld, tld, nameservers)
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
	}*/
	_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &res,
	})
}
