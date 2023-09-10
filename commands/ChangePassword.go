package commands

import (
	"NS/config"
	"NS/ns"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type ChangePasswordCommand struct {
	Command *discordgo.ApplicationCommand
	Config  *config.Config
}

var ChangePassword = ChangePasswordCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "change-password",
		Description: "Change password",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "email",
				Description: "Email address to change",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "password",
				Description: "New password for the email account",
				Required:    true,
			},
		},
	},
}

func (c *ChangePasswordCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var cPanelResponse ns.CPanelResponse

	commandData := interaction.ApplicationCommandData()
	email := commandData.Options[0].StringValue()
	password := commandData.Options[1].StringValue()
	_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Processing your request...",
		},
	})
	result := changeEmailPassword(email, password, c.Config.BasicAuth.Username, c.Config.BasicAuth.Password)
	if !isJSON(result) {
		// If the result is not a JSON string, send it directly
		_, _ = session.ChannelMessageSendReply(interaction.ChannelID, result, interaction.Message.MessageReference)
		return
	}
	err := json.Unmarshal([]byte(result), &cPanelResponse)
	if err != nil {
		// Handle JSON parsing error
		_, _ = session.ChannelMessageSendReply(interaction.ChannelID, "**❗ Warning! Trouble parsing response!**\\n"+result, interaction.Message.MessageReference)
		return
	}

	progressMessage := "**🔧 Attempting to change password for email:** " + email + "...\n"

	msg, _ := session.ChannelMessageSendReply(interaction.ChannelID, progressMessage, interaction.Message.MessageReference)
	// Check the reason field
	if data, ok := cPanelResponse.CPanelResult.Data.([]interface{}); ok {
		if len(data) > 0 {
			if obj, ok := data[0].(map[string]interface{}); ok {
				reason := obj["reason"].(string)
				if reason == "OK" {
					progressMessage += "**✅ Successfully changed!**"
				} else if strings.HasPrefix(reason, "You do not have an email account named") {
					progressMessage += "**⚠️ Email account does not exist!**"
				} else {
					progressMessage += "**❌ Deletion failed!**\\n" + reason
				}
			}
		} else {
			progressMessage += "**❌ Deletion failed!**\\n" + result
		}
	}

	// Respond to the Discord interaction with the appropriate message
	_, _ = session.ChannelMessageSendReply(msg.ChannelID, progressMessage, msg.Reference())
}

func changeEmailPassword(email, newPassword, cpanelUsername, cpanelPassword string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "Invalid email address format"
	}
	//cpanelUsername := "swapped2"
	//cpanelPassword := "Mmady5113x"
	user := parts[0]
	domain := parts[1]

	changePasswordURL := fmt.Sprintf("https://wch-llc.com:2083/execute/Email/passwd_pop?email=%s@%s&password=%s&domain=%s", user, domain, newPassword, domain)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", changePasswordURL, nil)
	if err != nil {
		return fmt.Sprintf("❌ Error creating request: %v", err.Error())
	}

	authString := fmt.Sprintf("%s:%s", cpanelUsername, cpanelPassword)
	authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(authString)))
	req.Header.Add("Authorization", authHeader)

	fmt.Println("Sending request to cPanel...")
	response, err := client.Do(req)
	fmt.Println("Received response from cPanel.")
	if err != nil {
		return fmt.Sprintf("❌ Error sending request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		return fmt.Sprintf("❌ Request failed with status code %d", response.StatusCode)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Sprintf("❌ Error reading response: %v", err)
	}

	// Modified line
	return handleUAPIResponse(string(responseBody))
}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func handleUAPIResponse(responseBody string) string {
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(responseBody), &response); err != nil {
		return "🚫 Error processing the server response."
	}

	if status, ok := response["status"].(float64); ok {
		if status == 1 {
			return "🟢 Password changed successfully!"
		} else if errors, ok := response["errors"].([]interface{}); ok && len(errors) > 0 {
			return fmt.Sprintf("🚫 %s", errors[0])
		} else {
			return "🚫 Failed to change the password. Please try again."
		}
	}
	return "🚫 Unexpected server response."
}
