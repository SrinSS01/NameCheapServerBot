package commands

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	Token = "MTE0MDY3OTgzNTA0MDA4ODI2NQ.GEX-it.OtYKALR-1yH4aDdp555Mx5xJEQm9h0aFOlrYl0" // Replace with your Discord bot token
)

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
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

	fmt.Println("Bot is now running. Press Ctrl+C to exit.")
	// Keep the program running
	select {}
}

func handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		commandData := i.ApplicationCommandData()
		if commandData.Name == "deleteemail" {
			emailOption := commandData.Options[0]
			if emailOption.Type == discordgo.ApplicationCommandOptionString {
				email := emailOption.StringValue()

				// Get the response from the deleteEmail function
				result := deleteEmail(email)
				progressMessage := "**🔍 Attempting to delete email:** " + email + "...\n"

				// Parse the cPanel response
				var cPanelResponse struct {
					CPanelResult struct {
						Data []struct {
							Reason string `json:"reason"`
						} `json:"data"`
					} `json:"cpanelresult"`
				}
				err := json.Unmarshal([]byte(result), &cPanelResponse)
				if err != nil {
					// Handle JSON parsing error
					progressMessage += "**❌ Error parsing response!**\\n" + result
				} else {
					// Check the reason field
					if len(cPanelResponse.CPanelResult.Data) > 0 {
						reason := cPanelResponse.CPanelResult.Data[0].Reason
						if reason == "OK" {
							progressMessage += "**✅ Successfully deleted!**"
						} else if strings.HasPrefix(reason, "You do not have an email account named") {
							progressMessage += "**⚠️ Email account does not exist!**"
						} else {
							progressMessage += "**❌ Deletion failed!**\\n" + reason
						}
					} else {
						progressMessage += "**❌ Deletion failed!**\\n" + result
					}
				}

				// Respond to the Discord interaction with the appropriate message
				finalResponse := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: progressMessage,
					},
				}
				s.InteractionRespond(i.Interaction, finalResponse)
			}
		}
	}
}

func deleteEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "Invalid email address format"
	}
	domain := parts[1]
	localPart := parts[0]

	cpanelUsername := "swapped2"
	cpanelPassword := "Mmady5113x"

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
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Sprintf("Request failed with status code %d", response.StatusCode)
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Sprintf("Error reading response: %v", err)
	}

	// Modified line
	return string(responseBody)
}

func registerCommands(s *discordgo.Session) {
	user, err := s.User("@me")
	if err != nil {
		fmt.Println("Error fetching bot user:", err)
		return
	}

	var commands = []*discordgo.ApplicationCommand{
		{
			Name:        "deleteemail",
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

	for _, command := range commands {
		_, err := s.ApplicationCommandCreate(user.ID, "", command)
		if err != nil {
			fmt.Println("Cannot create command:", err)
			return
		}
	}
}
