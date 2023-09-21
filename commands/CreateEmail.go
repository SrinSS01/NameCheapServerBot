package commands

import (
	"NS/config"
	"NS/ns"
	"NS/util"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"regexp"
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

var regex, _ = regexp.Compile("^(?P<domain>\\w+(?:\\.\\w+)+) +(?P<localPart>[a-z0-9!#$%&'*+/=?^_{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")(?: +(?P<password>\\S+))?$")

func RequestEmailCreate(domain, localPart, password string) (*ns.Response, error) {
	apiUrl := fmt.Sprintf("https://199.188.203.195:2083/json-api/cpanel?cpanel_jsonapi_func=addpop&cpanel_jsonapi_module=Email&cpanel_jsonapi_version=2&domain=%s&email=%s&password=%s", domain, localPart, password)
	response, err := util.MakeRequest("GET", apiUrl, "", nil)
	if err != nil {
		return nil, err
	}
	var emailCreateResponse ns.Response
	err = json.Unmarshal(response, &emailCreateResponse)
	if err != nil {
		return nil, err
	}
	return &emailCreateResponse, nil
}

func (c *CreateEmailCommand) ExecuteDash(session *discordgo.Session, messageCreate *discordgo.MessageCreate, args string) {
	matches := regex.FindStringSubmatch(args)
	if len(matches) == 0 {
		_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, "Please provide valid arguments", messageCreate.Reference())
		return
	}
	domain := matches[regex.SubexpIndex("domain")]
	localPart := matches[regex.SubexpIndex("localPart")]
	password := matches[regex.SubexpIndex("password")]
	if password == "" {
		password = c.Config.DefaultPassword
	}
	emailCreateResponse, err := RequestEmailCreate(domain, localPart, password)
	if err != nil {
		_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, err.Error(), messageCreate.Reference())
		return
	}

	if cpanelresult, ok := emailCreateResponse.Cpanelresult.(map[string]interface{}); ok {
		if datas, _ok := cpanelresult["data"].([]interface{}); _ok {
			for _, data := range datas {
				if data.(map[string]interface{})["result"].(float64) == 0 {
					_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, fmt.Sprintf("Failed to create email account: %s", data.(map[string]interface{})["reason"].(string)), messageCreate.Reference())
				} else {
					_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, fmt.Sprintf("Successfully created email account: %s", localPart+"@"+domain), messageCreate.Reference())
				}
			}
		} else if data, __ok := cpanelresult["data"].(map[string]interface{}); __ok {
			if data["result"].(float64) == 0 {
				_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, fmt.Sprintf("Failed to create email account: %s", data["reason"].(string)), messageCreate.Reference())
			} else {
				_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, fmt.Sprintf("Successfully created email account: %s", localPart+"@"+domain), messageCreate.Reference())
			}
		} else {
			_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, fmt.Sprintf("Error casting data to type `map[string]interface{}` or `[]interface{}`, ```json\n%s\n```", cpanelresult["data"]), messageCreate.Reference())
		}
	} else {
		marshal, _ := json.Marshal(emailCreateResponse)
		_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, "Error casting cpanelresult to type map[string]interface{}, ```json\n"+string(marshal)+"\n```", messageCreate.Reference())
	}
}

func (c *CreateEmailCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var err error
	options := interaction.ApplicationCommandData().Options
	domain := options[0].StringValue()
	email := options[1].StringValue()
	password := c.Config.DefaultPassword
	localPart := email
	if len(options) == 3 {
		password = options[2].StringValue()
	}
	if len(strings.TrimSpace(email)) == 0 || len(strings.TrimSpace(domain)) == 0 || len(strings.TrimSpace(password)) == 0 {
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

	emailCreateResponse, err := RequestEmailCreate(domain, localPart, password)
	if err != nil {
		content := err.Error()
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	var content string
	if cpanelresult, ok := emailCreateResponse.Cpanelresult.(map[string]interface{}); ok {
		if datas, _ok := cpanelresult["data"].([]interface{}); _ok {
			for _, data := range datas {
				if data.(map[string]interface{})["result"].(float64) == 0 {
					content = fmt.Sprintf("Failed to create email account: %s", data.(map[string]interface{})["reason"].(string))
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
		} else if data, __ok := cpanelresult["data"].(map[string]interface{}); __ok {
			if data["result"].(float64) == 0 {
				content = fmt.Sprintf("Failed to create email account: %s", data["reason"].(string))
				_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
					Content: &content,
				})
			} else {
				content = fmt.Sprintf("Successfully created email account: %s", localPart+"@"+domain)
				_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
					Content: &content,
				})
			}
		} else {
			content = fmt.Sprintf("Error casting data to type `map[string]interface{}` or `[]interface{}`, ```json\n%s\n```", cpanelresult["data"])
			_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
		}
	} else {
		marshal, _ := json.Marshal(emailCreateResponse)
		content = "Error casting cpanelresult to type map[string]interface{}, ```json\n" + string(marshal) + "\n```"
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
	}
}
