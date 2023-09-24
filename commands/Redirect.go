package commands

import (
	"NS/config"
	"NS/ns"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type RedirectCommand struct {
	Config  *config.Config
	Command *discordgo.ApplicationCommand
}

var Redirect = RedirectCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "redirect",
		Description: "adds a redirect for the domain in the cpanel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "domain",
				Description: "The domain to redirect",
				Required:    true,
			},
		},
	},
}

func RequestAddRedirect(username, password, domain string) (*ns.RedirectCommandResponse, error) {
	redirectURL := "https://" + domain
	apiUrl := fmt.Sprintf("https://%s:%s@199.188.203.195:2083/execute/Mime/add_redirect?domain=%s&redirect=%s&redirect_wildcard=0&redirect_www=0&src=/&type=permanent", username, password, domain, redirectURL)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(apiUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to add redirect for domain: %s. Error: %s", domain, err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err.Error())
		}
	}(resp.Body)
	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %s", err.Error())
	}
	var response ns.RedirectCommandResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %s", err.Error())
	}
	return &response, nil
}
func (c *RedirectCommand) ExecuteDash(session *discordgo.Session, messageCreate *discordgo.MessageCreate, domain string) {
	matched, _ := regexp.MatchString("^\\w+(?:\\.\\w+)+$", domain)
	if !matched {
		_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, "Please provide a valid domain name", messageCreate.Reference())
		return
	}
	username := c.Config.BasicAuth.Username
	password := c.Config.BasicAuth.Password
	response, err := RequestAddRedirect(username, password, domain)
	if err != nil {
		_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, err.Error(), messageCreate.Reference())
		return
	}

	// Handle the response based on the structure of the CommandResponse
	if response.Status == 1 { // Assuming a status of 1 indicates success
		builder := strings.Builder{}
		builder.WriteString("Successfully added the redirect.")
		if len(response.Messages) > 0 {
			builder.WriteByte(' ')
			builder.WriteString(strings.Join(response.Messages, " "))
		}
		_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, builder.String(), messageCreate.Reference())
	} else {
		errorMessage := strings.Builder{}
		errorMessage.WriteString("Failed to add the redirect.")
		if len(response.Errors) > 0 {
			errorMessage.WriteByte(' ')
			errorMessage.WriteString(strings.Join(response.Errors, " "))
		}
		_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, errorMessage.String(), messageCreate.Reference())
	}
}

func (c *RedirectCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var err error
	options := interaction.ApplicationCommandData().Options
	domain := options[0].StringValue()
	username := c.Config.BasicAuth.Username
	password := c.Config.BasicAuth.Password

	if len(strings.TrimSpace(domain)) == 0 {
		_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please enter a valid domain name",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	/*// Construct the URL for adding the redirect with embedded credentials using cPanel UAPI v2.0 format
	redirectURL := "https://" + domain
	apiUrl := fmt.Sprintf("https://%s:%s@199.188.203.195:2083/execute/Mime/add_redirect?domain=%s&redirect=%s&redirect_wildcard=0&redirect_www=0&src=/&type=permanent", username, password, domain, redirectURL)

	// Create a new HTTP client that skips SSL certificate verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Send the GET request
	resp, err := client.Get(apiUrl)
	if err != nil {
		content := fmt.Sprintf("Failed to add redirect for domain: %s. Error: %s", domain, err.Error())
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err.Error())
		}
	}(resp.Body)

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		content := fmt.Sprintf("Failed to read response body: %s", err.Error())
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	var response ns.RedirectCommandResponse
	fmt.Println("Raw Response:", string(body))
	err = json.Unmarshal(body, &response)
	if err != nil {
		content := fmt.Sprintf("Failed to parse response: %s", err.Error())
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}*/
	response, err := RequestAddRedirect(username, password, domain)
	if err != nil {
		content := err.Error()
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	// Handle the response based on the structure of the CommandResponse
	if response.Status == 1 { // Assuming a status of 1 indicates success
		successMessage := "Successfully added the redirect."
		if len(response.Messages) > 0 {
			successMessage += " " + strings.Join(response.Messages, " ")
		}
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &successMessage,
		})
	} else {
		errorMessage := "Failed to add the redirect."
		if len(response.Errors) > 0 {
			errorMessage += " " + strings.Join(response.Errors, " ")
		}
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &errorMessage,
		})
	}
}
