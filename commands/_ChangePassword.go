package commands

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("❗ Error creating Discord session:", err)
		return
	}

	// Register slash commands
	registerCommands(dg)

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			handleInteraction(s, i)
		}
	})

	// Open the connection to Discord
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session:", err)
		return
	}

	fmt.Println("✅ Bot is now running. Press Ctrl+C to exit.")
	// Keep the program running
	select {}
}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Parse the cPanel response
	var cPanelResponse struct {
		CPanelResult struct {
			Data []struct {
				Reason string `json:"reason"`
			} `json:"data"`
		} `json:"cpanelresult"`
	}
	if i.Type == discordgo.InteractionApplicationCommand {
		commandData := i.ApplicationCommandData()
		if commandData.Name == "changepassword" {
			emailOption := commandData.Options[0]
			passwordOption := commandData.Options[1]
			var email string
			if emailOption.Type == discordgo.ApplicationCommandOptionString {
				email = emailOption.StringValue()

				// Define password before using it
				password := passwordOption.StringValue()

				// Process the password change asynchronously
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Content: "Processing your request..."}})
				go func() {
					// Attempt to change the email password
					result := changeEmailPassword(email, password)

					if !isJSON(result) {
						// If the result is not a JSON string, send it directly
						s.ChannelMessageSend(i.ChannelID, result)
						return
					}

					err := json.Unmarshal([]byte(result), &cPanelResponse)
					if err != nil {
						// Handle JSON parsing error
						s.ChannelMessageSend(i.ChannelID, "**❗ Warning! Error parsing response!**\\n"+result)
						return
					}

					progressMessage := "**🔧 Attempting to change password for email:** " + email + "...\n"

					s.ChannelMessageSend(i.ChannelID, progressMessage)
					// Check the reason field
					if len(cPanelResponse.CPanelResult.Data) > 0 {
						reason := cPanelResponse.CPanelResult.Data[0].Reason
						if reason == "OK" {
							progressMessage += "**✅ Successfully changed!**"
						} else if strings.HasPrefix(reason, "You do not have an email account named") {
							progressMessage += "**⚠️ Email account does not exist!**"
						} else {
							progressMessage += "**❌ Deletion failed!**\\n" + reason
						}
					} else {
						progressMessage += "**❌ Deletion failed!**\\n" + result
					}

					// Respond to the Discord interaction with the appropriate message
					finalResponse := &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: progressMessage,
						},
					}
					s.InteractionRespond(i.Interaction, finalResponse)
				}()
			}
		}
	}
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

func changeEmailPassword(email, newPassword string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "Invalid email address format"
	}
	cpanelUsername := "swapped2"
	cpanelPassword := "Mmady5113x"
	user := parts[0]
	domain := parts[1]

	changePasswordURL := fmt.Sprintf("https://wch-llc.com:2083/execute/Email/passwd_pop?email=%s@%s&password=%s&domain=%s", user, domain, newPassword, domain)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", changePasswordURL, nil)
	if err != nil {
		return fmt.Sprintf("❌ Error creating request: %v", err)
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
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Sprintf("❌ Request failed with status code %d", response.StatusCode)
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Sprintf("❌ Error reading response: %v", err)
	}

	// Modified line
	return handleUAPIResponse(string(responseBody))
}

func registerCommands(s *discordgo.Session) {
	user, err := s.User("@me")
	if err != nil {
		fmt.Println("Error fetching bot user:", err)
		return
	}

	var commands = []*discordgo.ApplicationCommand{
		{
			Name:        "changepassword",
			Description: "Change password of an email account from cPanel",
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

	for _, command := range commands {
		_, err := s.ApplicationCommandCreate(user.ID, "", command)
		if err != nil {
			fmt.Println("Cannot create command:", err)
			return
		}
	}
}
