package commands

import (
	"NS/config"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type SSLCommand struct {
	Command *discordgo.ApplicationCommand
}

var SSL = SSLCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "ssl",
		Description: "check SSL certificates",
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

func (c *SSLCommand) ExecuteDash(session *discordgo.Session, messageCreate *discordgo.MessageCreate, domain string) {
	matched, err := regexp.MatchString("^\\w+(?:\\.\\w+)+$", domain)
	if err != nil || !matched {
		_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, "Wrong domain format", messageCreate.Reference())
		return
	}
	if IsSSLInstalled(domain) {
		// SSL is installed
		_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, fmt.Sprintf("SSL is installed on %s", domain), messageCreate.Reference())
	} else {
		// SSL is not installed
		_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, fmt.Sprintf("SSL is not installed on %s", domain), messageCreate.Reference())
	}
}

func (c *SSLCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	options := interaction.ApplicationCommandData().Options
	domain := options[0].StringValue()

	// Check if SSL is installed
	if IsSSLInstalled(domain) {
		// SSL is installed
		err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("SSL is installed on %s", domain),
			},
		})
		if err != nil {
			fmt.Println("Failed to respond to interaction:", err)
		}
	} else {
		// SSL is not installed
		err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("SSL is not installed on %s", domain),
			},
		})
		if err != nil {
			fmt.Println("Failed to respond to interaction:", err)
		}
	}
}

func IsSSLInstalled(domain string) bool {
	// Construct the URL for WHM's SSL installation check (adjust as per WHM's API documentation)
	url := config.WHMHost + "/json-api/listaccts?api.version=1&search=" + domain + "&searchtype=domain"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Failed to create request:", err)
		return false
	}

	req.Header.Add("Authorization", "whm root:"+config.WHMAPIKey)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to execute request:", err)
		return false
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Failed to close response body:", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read response:", err)
		return false
	}

	// This is a basic check. Depending on WHM's response, you might need to adjust this condition.
	return strings.Contains(string(body), "SSL Installed Status or Marker") // Replace with the actual marker or status indicating SSL is installed
}
