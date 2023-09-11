package commands

import (
	"NS/config"
	"NS/ns"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/bwmarrin/discordgo"
)

type MonitorCommand struct {
	Command *discordgo.ApplicationCommand
	Config  *config.Config
}

type MonitorTaskScheduler struct {
	MessageCount int
	Task         func(msgOld *discordgo.Message, session *discordgo.Session)
	StopTask     chan bool
}

var monitor = map[string]*MonitorTaskScheduler{}

var Monitor = MonitorCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "monitor",
		Description: "tracking information for the messages in the account's message queue",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "start",
				Description: "start monitoring",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "email",
						Description: "Email to start monitor",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "stop",
				Description: "Stop monitoring",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "email",
						Description: "Email to stop monitor",
						Required:    true,
					},
				},
			},
		},
	},
}

func RequestEmailTrack(client *http.Client, apiUrl, cPanelUserName, cPanelPassword string) (*ns.CPanelResponse, error) {
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("❌ Error creating request: %v", err.Error())
	}
	authString := fmt.Sprintf("%s:%s", cPanelUserName, cPanelPassword)
	authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(authString)))
	req.Header.Add("Authorization", authHeader)
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("❌ Error creating request: %v", err.Error())
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(response.Body)
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("❌ Request failed with status code %d", response.StatusCode)
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("❌ Error reading response: %v", err.Error())
	}
	var emailTrackResponse ns.CPanelResponse
	err = json.Unmarshal(responseBody, &emailTrackResponse)
	if err != nil {
		return nil, fmt.Errorf("❌ Error parsing response: %s", err.Error())
	}
	if len(emailTrackResponse.CPanelResult.Error) != 0 {
		return nil, fmt.Errorf("⚠️%s", emailTrackResponse.CPanelResult.Error)
	}
	return &emailTrackResponse, nil
}

func (m *MonitorCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	commandData := interaction.ApplicationCommandData()
	subCommand := commandData.Options[0]
	email := subCommand.Options[0].StringValue()
	switch subCommand.Name {
	case "stop":
		monitorTaskScheduler := monitor[email]
		if monitorTaskScheduler == nil {
			content := "email not monitored yet"
			_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
			return
		}
		content := "Stopped monitoring " + email
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		monitorTaskScheduler.StopTask <- true
		break
	case "start":
		cPanelUserName := m.Config.BasicAuth.Username
		cPanelPassword := m.Config.BasicAuth.Password
		apiUrl := "https://wch-llc.com:2083/json-api/cpanel?cpanel_jsonapi_user=user&cpanel_jsonapi_apiversion=2&cpanel_jsonapi_module=EmailTrack&cpanel_jsonapi_func=search&success=1&defer=0&recipient=" + url.QueryEscape(email)
		content := "Starting to monitor"
		msg, _ := session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		client := &http.Client{}
		result, err := RequestEmailTrack(client, apiUrl, cPanelUserName, cPanelPassword)
		if err != nil {
			_, _ = session.ChannelMessageSendReply(msg.ChannelID, err.Error(), msg.Reference())
			return
		}
		stop := make(chan bool)

		ticker := time.NewTicker(time.Second * 5)

		task := func(msgOld *discordgo.Message, session *discordgo.Session) {
			for {
				select {
				case <-stop:
					ticker.Stop()
					return
				case <-ticker.C:
					req, err := RequestEmailTrack(client, apiUrl, cPanelUserName, cPanelPassword)
					if err != nil {
						return
					}
					mts := monitor[email]
					if datas, ok := req.CPanelResult.Data.([]interface{}); ok {
						newMsgCount := len(datas)
						if mts.MessageCount < newMsgCount {
							mts.MessageCount = newMsgCount
							data := datas[0]
							if dataObj, ok := data.(map[string]interface{}); ok {
								sender := dataObj["email"].(string)
								recipient := dataObj["recipient"].(string)
								senderip := dataObj["senderip"].(string)
								senderhost := dataObj["senderhost"].(string)
								actionunixtime := dataObj["actionunixtime"].(float64)
								msgOld, _ = session.ChannelMessageSendEmbedReply(msgOld.ChannelID, &discordgo.MessageEmbed{
									Title: "Email Recieved!",
									Color: 0x00FF00,
									Fields: []*discordgo.MessageEmbedField{
										{
											Name:  "To",
											Value: fmt.Sprintf("`%s`", recipient),
										},
										{
											Name:  "From",
											Value: fmt.Sprintf("`%s`", sender),
										},
										{
											Name:  "Sender IP",
											Value: fmt.Sprintf("`%s`", senderip),
										},
										{
											Name:  "Sender Host",
											Value: fmt.Sprintf("`%s`", senderhost),
										},
										{
											Name:  "Action Time",
											Value: fmt.Sprintf("<t:%d>", int64(actionunixtime)),
										},
									},
								}, msgOld.Reference())
							}
						}
					}
				}
			}
		}
		if datas, ok := result.CPanelResult.Data.([]interface{}); ok {
			monitor[email] = &MonitorTaskScheduler{
				MessageCount: len(datas),
				Task:         task,
				StopTask:     stop,
			}
		}
		task(msg, session)
	}
}
