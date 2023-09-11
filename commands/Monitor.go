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

	"github.com/bwmarrin/discordgo"
)

type MonitorCommand struct {
	Command *discordgo.ApplicationCommand
	Config  *config.Config
}

var Monitor = MonitorCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "monitor",
		Description: "tracking information for the messages in the account's message queue",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "email",
				Description: "Email to monitor",
				Required:    true,
			},
		},
	},
}

func (m *MonitorCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	commandData := interaction.ApplicationCommandData()
	email := commandData.Options[0].StringValue()
	cPanelUserName := m.Config.BasicAuth.Username
	cPanelPassword := m.Config.BasicAuth.Password
	apiUrl := "https://wch-llc.com:2083/json-api/cpanel?cpanel_jsonapi_user=user&cpanel_jsonapi_apiversion=2&cpanel_jsonapi_module=EmailTrack&cpanel_jsonapi_func=search&success=1&defer=0&recepient=" + url.QueryEscape(email)
	content := "Starting to monitor"
	msg, _ := session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("❌ Error creating request: %v", err.Error()), msg.Reference())
		return
	}
	authString := fmt.Sprintf("%s:%s", cPanelUserName, cPanelPassword)
	authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(authString)))
	req.Header.Add("Authorization", authHeader)
	response, err := client.Do(req)
	if err != nil {
		_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("❌ Error creating request: %v", err.Error()), msg.Reference())
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(response.Body)
	if response.StatusCode != http.StatusOK {
		_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("❌ Request failed with status code %d", response.StatusCode), msg.Reference())
		return
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("❌ Error reading response: %v", err), msg.Reference())
		return
	}
	var emailTrackResponse ns.CPanelResponse
	err = json.Unmarshal(responseBody, &emailTrackResponse)
	if err != nil {
		_, _ = session.ChannelMessageSendReply(msg.ChannelID, "❌ Error parsing response", msg.Reference())
		return
	}
	if len(emailTrackResponse.CPanelResult.Error) != 0 {
		_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("⚠️%s", emailTrackResponse.CPanelResult.Error), msg.Reference())
		return
	}

	/*
	   	// You can edit this code!
	      // Click here and start typing.
	      package main

	      import (
	      	"fmt"
	      	"os"
	      	"os/signal"
	      	"syscall"
	      	"time"
	      )

	      func main() {
	      	task(time.Now())
	      	tick := time.NewTicker(time.Second * 5)
	      	scheduler(tick)
	      	sigs := make(chan os.Signal, 1)
	      	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	      	<-sigs
	      	tick.Stop()
	      }

	      func scheduler(tick *time.Ticker) {
	      	for t := range tick.C {
	      		task(t)
	      	}
	      }

	      func task(t time.Time) {
	      	fmt.Println("hello! printed at ", t)
	      }
	*/
}
