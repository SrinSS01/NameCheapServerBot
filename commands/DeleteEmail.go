package commands

import (
	"NS/config"
	"NS/ns"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type DeleteEmailCommand struct {
	Command *discordgo.ApplicationCommand
	Config  *config.Config
}

var DeleteEmail = DeleteEmailCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "delete-email",
		Description: "Delete an email account from cPanel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "email",
				Description: "Email address to delete",
				Required:    true,
			},
		},
	},
}

func (d *DeleteEmailCommand) ExecuteDash(session *discordgo.Session, messageCreate *discordgo.MessageCreate, email string) {
	email = strings.TrimSpace(email)
	matched, _ := regexp.MatchString("^(?:[a-z0-9!#$%&'*+/=?^_{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])$", email)
	if !matched {
		_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, "Please provide a valid email address", messageCreate.Reference())
		return
	}
	username := d.Config.BasicAuth.Username
	password := d.Config.BasicAuth.Password
	msg, _ := session.ChannelMessageSendReply(messageCreate.ChannelID, "**🔍 Attempting to delete email:** "+email+"...", messageCreate.Reference())
	result, err := RequestDeleteEmail(email, username, password)
	if err != nil {
		// Handle JSON parsing error
		_, _ = session.ChannelMessageSendReply(msg.ChannelID, err.Error(), msg.Reference())
		return
	}
	if data, ok := result.CPanelResult.Data.([]interface{}); ok {
		// Check the reason field
		if len(data) > 0 {
			if obj, ok := data[0].(map[string]interface{}); ok {
				reason := obj["reason"].(string)
				if reason == "OK" {
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, "**✅ Successfully deleted!**", msg.Reference())
				} else if strings.HasPrefix(reason, "You do not have an email account named") {
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, "**⚠️ Email account does not exist!**", msg.Reference())
				} else {
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, "**❌ Deletion failed!**\n"+reason, msg.Reference())
				}
			}
			return
		}
		marshal, _ := json.Marshal(result)
		_, _ = session.ChannelMessageSendReply(msg.ChannelID, "**❌ Deletion failed!**\n"+string(marshal), msg.Reference())
	}
}

func (d *DeleteEmailCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	commandData := interaction.ApplicationCommandData()
	email := commandData.Options[0].StringValue()
	username := d.Config.BasicAuth.Username
	password := d.Config.BasicAuth.Password
	content := "**🔍 Attempting to delete email:** " + email + "..."
	msg, _ := session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	result, err := RequestDeleteEmail(email, username, password)
	if err != nil {
		// Handle JSON parsing error
		_, _ = session.ChannelMessageSendReply(msg.ChannelID, err.Error(), msg.Reference())
		return
	}

	if data, ok := result.CPanelResult.Data.([]interface{}); ok {
		// Check the reason field
		if len(data) > 0 {
			if obj, ok := data[0].(map[string]interface{}); ok {
				reason := obj["reason"].(string)
				if reason == "OK" {
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, "**✅ Successfully deleted!**", msg.Reference())
				} else if strings.HasPrefix(reason, "You do not have an email account named") {
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, "**⚠️ Email account does not exist!**", msg.Reference())
				} else {
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, "**❌ Deletion failed!**\n"+reason, msg.Reference())
				}
			}
			return
		}
		marshal, _ := json.Marshal(result)
		_, _ = session.ChannelMessageSendReply(msg.ChannelID, "**❌ Deletion failed!**\n"+string(marshal), msg.Reference())
	}
}

func RequestDeleteEmail(email string, cpanelUsername string, cpanelPassword string) (*ns.CPanelResponse, error) {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid email address format")
	}
	domain := parts[1]
	localPart := parts[0]

	//cpanelUsername :=  "swapped2"
	//cpanelPassword :=  "Mmady5113x"

	deleteURL := fmt.Sprintf("https://wch-llc.com:2083/json-api/cpanel?cpanel_jsonapi_func=delpop&cpanel_jsonapi_module=Email&cpanel_jsonapi_version=2&domain=%s&email=%s", domain, localPart)

	client := &http.Client{}
	req, err := http.NewRequest("GET", deleteURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	authString := fmt.Sprintf("%s:%s", cpanelUsername, cpanelPassword)
	authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(authString)))
	req.Header.Add("Authorization", authHeader)

	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code %d", response.StatusCode)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var cPanelResponse ns.CPanelResponse
	err = json.Unmarshal(responseBody, &cPanelResponse)
	if err != nil {
		return nil, fmt.Errorf("**❌ Error parsing response!**\n```\n%s\n```", err.Error())
	}
	return &cPanelResponse, nil
}
