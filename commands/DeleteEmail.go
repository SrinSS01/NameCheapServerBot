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

func (d *DeleteEmailCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	commandData := interaction.ApplicationCommandData()
	email := commandData.Options[0].StringValue()
	username := d.Config.BasicAuth.Username
	password := d.Config.BasicAuth.Password
	result := deleteEmail(email, username, password)
	content := "**🔍 Attempting to delete email:** " + email + "..."
	msg, _ := session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	//progressMessage := "**🔍 Attempting to delete email:** " + email + "...\n"

	var cPanelResponse ns.CPanelResponse
	err := json.Unmarshal([]byte(result), &cPanelResponse)
	if err != nil {
		// Handle JSON parsing error
		_, _ = session.ChannelMessageSendReply(msg.ChannelID, "**❌ Error parsing response!**\n```"+result+"```", msg.Reference())
		return
	}

	if data, ok := cPanelResponse.CPanelResult.Data.([]interface{}); ok {
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
		_, _ = session.ChannelMessageSendReply(msg.ChannelID, "**❌ Deletion failed!**\n"+result, msg.Reference())
	}
}

func deleteEmail(email string, cpanelUsername string, cpanelPassword string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "Invalid email address format"
	}
	domain := parts[1]
	localPart := parts[0]

	//cpanelUsername :=  "swapped2"
	//cpanelPassword :=  "Mmady5113x"

	deleteURL := fmt.Sprintf("https://wch-llc.com:2083/json-api/cpanel?cpanel_jsonapi_func=delpop&cpanel_jsonapi_module=Email&cpanel_jsonapi_version=2&domain=%s&email=%s", domain, localPart)

	client := &http.Client{}
	req, err := http.NewRequest("GET", deleteURL, nil)
	if err != nil {
		return fmt.Sprintf("Error creating request: %v", err)
	}

	authString := fmt.Sprintf("%s:%s", cpanelUsername, cpanelPassword)
	authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(authString)))
	req.Header.Add("Authorization", authHeader)

	response, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("Error sending request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		return fmt.Sprintf("Request failed with status code %d", response.StatusCode)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Sprintf("Error reading response: %v", err)
	}

	// Modified line
	return string(responseBody)
}
