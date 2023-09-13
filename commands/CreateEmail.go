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
