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
)

type DeleteFileCommand struct {
	Command *discordgo.ApplicationCommand
	Config  *config.Config
}

var DeleteFile = DeleteEmailCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "delete-file",
		Description: "Delete a file from cPanel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "file-name",
				Description: "file to delete",
				Required:    true,
			},
		},
	},
}

func (d DeleteFileCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	commandData := interaction.ApplicationCommandData()
	fileName := commandData.Options[0].StringValue()
	cPanelUserName := d.Config.BasicAuth.Username
	cPanelPassword := d.Config.BasicAuth.Password
	_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	var cPanelResponse ns.CPanelResponse
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
	}
	content := "✅ Successfully deleted"
	_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}
