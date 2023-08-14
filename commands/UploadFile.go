package commands

import (
	"NS/ns"
	"NS/util"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

type UploadFileCommand struct {
	Command *discordgo.ApplicationCommand
}

var UploadFile = UploadFileCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "uploadfile",
		Description: "upload a file",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "domain",
				Description: "The domain to upload the file to",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "file",
				Description: "The file to upload",
				Required:    true,
			},
		},
	},
}

func (c *UploadFileCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var err error
	applicationCommandData := interaction.ApplicationCommandData()
	options := applicationCommandData.Options
	domain := options[0].StringValue()
	fileId := options[1].Value.(string)
	if len(domain) == 0 {
		err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please enter a valid domain name",
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

	attachment := applicationCommandData.Resolved.Attachments[fileId]
	fileName := attachment.Filename
	fileURL := attachment.URL
	res, err := http.DefaultClient.Get(fileURL)
	if err != nil {
		editDeferredReply(session, interaction, fmt.Sprintf("Error during http.Get(): %s", err.Error()))
		return
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file-1", fileName)
	if err != nil {
		editDeferredReply(session, interaction, fmt.Sprintf("Error during writer.CreateFormFile(): %s", err.Error()))
		return
	}
	_, err = io.Copy(part, res.Body)

	escapedDomain := url.QueryEscape(domain)
	uploadURL := fmt.Sprintf("https://199.188.203.195:2083/execute/Fileman/upload_files?dir=/home/swapped2/%s", escapedDomain)
	response, err := util.MakeRequest("POST", uploadURL, writer.FormDataContentType(), body)
	if err != nil {
		editDeferredReply(session, interaction, fmt.Sprintf("Error during util.MakeRequest(): %s", err.Error()))
		return
	}
	var fileUploadResponse ns.FileUploadResponse
	err = json.Unmarshal(response, &fileUploadResponse)
	if err != nil {
		content := fmt.Sprintf("Error during json.Unmarshal(): %s", err.Error())
		editDeferredReply(session, interaction, content)
		return
	}
	if len(fileUploadResponse.Errors) > 0 {
		var contentBuilder strings.Builder
		for i, errors := range fileUploadResponse.Errors {
			contentBuilder.WriteString(fmt.Sprintf("Error `%d`: ```\n%s```\n", i, errors))
		}
		editDeferredReply(session, interaction, contentBuilder.String())
		return
	}
	if data, ok := fileUploadResponse.Data.(map[string]interface{}); ok {
		if uploads, ok := data["uploads"].([]interface{}); ok {
			for _, upload := range uploads {
				if upload, ok := upload.(map[string]interface{}); ok {
					if upload["status"].(float64) == 0 {
						editDeferredReply(session, interaction, fmt.Sprintf("Failed to upload file: %s", upload["reason"].(string)))
					} else {
						editDeferredReply(session, interaction, "Successfully uploaded file.")
					}
				} else {
					editDeferredReply(session, interaction, "Unable to cast `upload` to `map[string]interface{}`")
				}
			}
		} else {
			editDeferredReply(session, interaction, "Unable to cast `uploads` to `[]interface{}`")
		}
	} else {
		editDeferredReply(session, interaction, "Unable to cast `data` to `map[string]interface{}`")
	}
}

func editDeferredReply(session *discordgo.Session, interaction *discordgo.InteractionCreate, content string) {
	_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}
