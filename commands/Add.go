package commands

import (
	"NS/ns"
	"NS/util"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/url"
	"regexp"
	"strings"
)

type AddCommand struct {
	Command *discordgo.ApplicationCommand
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

var DomainRegex = regexp.MustCompile("^[a-z0-9!#$%&'*+/=?^_{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\"$")

func RequestAddDomain(domain string) (string, error) {
	escapedDomain := url.QueryEscape(domain)
	apiUrl := fmt.Sprintf("https://199.188.203.195:2083/json-api/cpanel?cpanel_jsonapi_func=addaddondomain&cpanel_jsonapi_module=AddonDomain&cpanel_jsonapi_version=2&newdomain=%s&subdomain=%s&dir=/home/swapped2/%s", escapedDomain, escapedDomain, escapedDomain)
	response, err := util.MakeRequest("GET", apiUrl, "", nil)
	if err != nil {
		return "", err
	}

	var addDomainResponse ns.Response
	err = json.Unmarshal(response, &addDomainResponse)
	if err != nil {
		return "", fmt.Errorf("Error unmarshalling JSON ```json\n%s```: %s", response, err.Error())
	}

	if cpanelresult, ok := addDomainResponse.Cpanelresult.(map[string]interface{}); ok {
		if data, ok := cpanelresult["data"].(map[string]interface{}); ok {
			if data["result"].(string) == "0" {
				return "", fmt.Errorf("Failed to add domain: " + data["reason"].(string))
			}
			return "Successfully added domain: " + domain, nil
		} else if datas, ok := cpanelresult["data"].([]interface{}); ok {
			builder := strings.Builder{}
			for _, data := range datas {
				if data.(map[string]interface{})["result"].(float64) == 0 {
					builder.WriteString("Failed to add domain: " + data.(map[string]interface{})["reason"].(string))
					builder.WriteByte('\n')
				}
				builder.WriteString("✅ Successfully added domain: " + domain)
			}
			return builder.String(), nil
		} else {
			return "", fmt.Errorf("error casting data to type []interface{} or map[string]interface{}")
		}
	} else {
		return "", fmt.Errorf("Error casting cpanelresult to type map[string]interface{}, ```json\n" + string(response) + "```")
	}
}

func (a *AddCommand) ExecuteDash(s *discordgo.Session, m *discordgo.MessageCreate, domain string) {
	matched := DomainRegex.MatchString(strings.TrimSpace(domain))
	if !matched {
		_, _ = s.ChannelMessageSendReply(m.ChannelID, "Wrong domain format", m.Reference())
		return
	}
	result, err := RequestAddDomain(domain)
	if err != nil {
		_, _ = s.ChannelMessageSendReply(m.ChannelID, err.Error(), m.Reference())
		return
	}
	_, _ = s.ChannelMessageSendReply(m.ChannelID, result, m.Reference())
}

func (a *AddCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
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
	result, err := RequestAddDomain(domain)
	if err != nil {
		eStr := err.Error()
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &eStr,
		})
		return
	}
	_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &result,
	})
	/*escapedDomain := url.QueryEscape(domain)
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
	var content string
	if cpanelresult, ok := addDomainResponse.Cpanelresult.(map[string]interface{}); ok {
		if data, ok := cpanelresult["data"].(map[string]interface{}); ok {
			if data["result"].(string) == "0" {
				content = "Failed to add domain: " + data["reason"].(string)
				_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
					Content: &content,
				})
				return
			}
			content := "Successfully added domain: " + domain
			_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
		} else if datas, ok := cpanelresult["data"].([]interface{}); ok {
			for _, data := range datas {
				if data.(map[string]interface{})["result"].(float64) == 0 {
					content = "Failed to add domain: " + data.(map[string]interface{})["reason"].(string)
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
		} else {
			content := "Error casting data to type []interface{} or map[string]interface{}"
			_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
			return
		}
	} else {
		content := "Error casting cpanelresult to type map[string]interface{}, ```json\n" + string(response) + "```"
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}*/
}
