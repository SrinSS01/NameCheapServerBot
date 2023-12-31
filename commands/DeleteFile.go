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
	"net/url"
	"regexp"
	"strings"
)

type DeleteFileCommand struct {
	Command *discordgo.ApplicationCommand
	Config  *config.Config
}

var DeleteFile = DeleteFileCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "delete-file",
		Description: "Delete a file from cPanel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "domain",
				Description: "domain to delete the file from",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "file-name",
				Description: "file to delete",
				Required:    true,
			},
		},
	},
}

var argsRegex, _ = regexp.Compile("^(?P<domain>[a-z0-9!#$%&'*+/=?^_{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\") +(?P<fileName>[\\w\\-. ]+)$")

func RequestDeleteFile(fileName, cPanelUserName, cPanelPassword string) (string, error) {
	var cPanelResponse ns.CPanelResponse
	apiUrl := fmt.Sprintf("https://wch-llc.com:2083/json-api/cpanel?cpanel_jsonapi_user=user&cpanel_jsonapi_apiversion=2&cpanel_jsonapi_module=Fileman&cpanel_jsonapi_func=fileop&op=trash&sourcefiles=%s&doubledecode=1", fileName)
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return "", fmt.Errorf("❌ Error creating request: %v", err.Error())
	}
	authString := fmt.Sprintf("%s:%s", cPanelUserName, cPanelPassword)
	authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(authString)))
	req.Header.Add("Authorization", authHeader)
	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("❌ Error creating request: %v", err.Error())
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(response.Body)
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("❌ Request failed with status code %d", response.StatusCode)
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("❌ Error reading response: %s", err.Error())
	}
	err = json.Unmarshal(responseBody, &cPanelResponse)
	if err != nil {
		return "", fmt.Errorf("❌ Error parsing response")
	}
	if len(cPanelResponse.CPanelResult.Error) != 0 {
		return "", fmt.Errorf("⚠️%s", cPanelResponse.CPanelResult.Error)
	}
	return "✅ Successfully deleted", nil
}
func (d *DeleteFileCommand) ExecuteDash(session *discordgo.Session, messageCreate *discordgo.MessageCreate, args string) {
	matches := argsRegex.FindStringSubmatch(strings.TrimSpace(args))
	if len(matches) == 0 {
		_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, "Please provide valid arguments", messageCreate.Reference())
		return
	}
	domain := matches[argsRegex.SubexpIndex("domain")]
	fileName := fmt.Sprintf("/home/swapped2/%s/%s", url.QueryEscape(domain), matches[argsRegex.SubexpIndex("fileName")])
	cPanelUserName := d.Config.BasicAuth.Username
	cPanelPassword := d.Config.BasicAuth.Password
	res, err := RequestDeleteFile(fileName, cPanelUserName, cPanelPassword)
	if err != nil {
		_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, err.Error(), messageCreate.Reference())
		return
	}
	_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, res, messageCreate.Reference())
}

func (d *DeleteFileCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	commandData := interaction.ApplicationCommandData()
	domain := commandData.Options[0].StringValue()
	fileName := fmt.Sprintf("/home/swapped2/%s/%s", url.QueryEscape(domain), commandData.Options[1].StringValue())
	cPanelUserName := d.Config.BasicAuth.Username
	cPanelPassword := d.Config.BasicAuth.Password
	_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	/*var cPanelResponse ns.CPanelResponse
	apiUrl := fmt.Sprintf("https://wch-llc.com:2083/json-api/cpanel?cpanel_jsonapi_user=user&cpanel_jsonapi_apiversion=2&cpanel_jsonapi_module=Fileman&cpanel_jsonapi_func=fileop&op=trash&sourcefiles=%s&doubledecode=1", fileName)
	client := &http.Client{}

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		content := fmt.Sprintf("❌ Error creating request: %v", err.Error())
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}
	authString := fmt.Sprintf("%s:%s", cPanelUserName, cPanelPassword)
	authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(authString)))
	req.Header.Add("Authorization", authHeader)
	response, err := client.Do(req)
	if err != nil {
		content := fmt.Sprintf("❌ Error creating request: %v", err.Error())
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		content := fmt.Sprintf("❌ Request failed with status code %d", response.StatusCode)
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		content := fmt.Sprintf("❌ Error reading response: %v", err)
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}
	err = json.Unmarshal(responseBody, &cPanelResponse)
	if err != nil {
		content := "❌ Error parsing response"
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}
	if len(cPanelResponse.CPanelResult.Error) != 0 {
		content := fmt.Sprintf("⚠️%s", cPanelResponse.CPanelResult.Error)
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}*/
	res, err := RequestDeleteFile(fileName, cPanelUserName, cPanelPassword)
	if err != nil {
		content := err.Error()
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}
	_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &res,
	})
}
