package main

import (
	"NS/commands"
	"NS/config"
	"NS/logger"
	"NS/util"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"
)

type Execute struct {
	Slash func(*discordgo.Session, *discordgo.InteractionCreate)
	Dash  func(*discordgo.Session, *discordgo.MessageCreate, string)
}

var (
	cnfg    = config.Config{}
	discord *discordgo.Session
	cmds    = []*discordgo.ApplicationCommand{
		commands.CreateDomain.Command,
		commands.NS.Command,
		commands.Add.Command,
		commands.CreateEmail.Command,
		commands.SSL.Command,
		commands.UploadFile.Command,
		commands.Automate.Command,
		commands.CheckDomain.Command,
		commands.Redirect.Command,
		commands.ChangePassword.Command,
		commands.DeleteEmail.Command,
		commands.DeleteFile.Command,
		commands.Monitor.Command,
	}
	commandHandlers = map[string]Execute{
		commands.CreateDomain.Command.Name: {
			Slash: commands.CreateDomain.Execute,
			Dash:  commands.CreateDomain.ExecuteDash,
		},
		commands.NS.Command.Name: {
			Slash: commands.NS.Execute,
			Dash:  commands.NS.ExecuteDash,
		},
		commands.Add.Command.Name: {
			Slash: commands.Add.Execute,
			Dash:  commands.Add.ExecuteDash,
		},
		commands.CreateEmail.Command.Name: {
			Slash: commands.CreateEmail.Execute,
			Dash:  commands.CreateEmail.ExecuteDash,
		},
		commands.SSL.Command.Name: {
			Slash: commands.SSL.Execute,
			Dash:  commands.SSL.ExecuteDash,
		},
		commands.UploadFile.Command.Name: {
			Slash: commands.UploadFile.Execute,
			Dash:  commands.UploadFile.ExecuteDash,
		},
		commands.Automate.Command.Name: {
			Slash: commands.Automate.Execute,
		},
		commands.CheckDomain.Command.Name: {
			Slash: commands.CheckDomain.Execute,
			Dash:  commands.CheckDomain.ExecuteDash,
		},
		commands.Redirect.Command.Name: {
			Slash: commands.Redirect.Execute,
			Dash:  commands.Redirect.ExecuteDash,
		},
		commands.ChangePassword.Command.Name: {
			Slash: commands.ChangePassword.Execute,
			Dash:  commands.ChangePassword.ExecuteDash,
		},
		commands.DeleteEmail.Command.Name: {
			Slash: commands.DeleteEmail.Execute,
			Dash:  commands.DeleteEmail.ExecuteDash,
		},
		commands.DeleteFile.Command.Name: {
			Slash: commands.DeleteFile.Execute,
			Dash:  commands.DeleteFile.ExecuteDash,
		},
		commands.Monitor.Command.Name: {
			Slash: commands.Monitor.Execute,
			Dash:  commands.Monitor.ExecuteDash,
		},
	}
)

func init() {
	file, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Print("Enter bot token: ")
		if _, err := fmt.Scanln(&cnfg.Token); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter api user: ")
		if _, err := fmt.Scanln(&cnfg.ApiUser); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter api key: ")
		if _, err := fmt.Scanln(&cnfg.ApiKey); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter user name: ")
		if _, err := fmt.Scanln(&cnfg.UserName); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter client ip: ")
		if _, err := fmt.Scanln(&cnfg.ClientIP); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter default password: ")
		if _, err := fmt.Scanln(&cnfg.DefaultPassword); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter basic auth username: ")
		if _, err := fmt.Scanln(&cnfg.BasicAuth.Username); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter basic auth password: ")
		if _, err := fmt.Scanln(&cnfg.BasicAuth.Password); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter default name servers: ")
		if _, err := fmt.Scanln(&cnfg.DefaultNameServers); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		configJson()
		return
	}
	if err := json.Unmarshal(file, &cnfg); err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
}

func configJson() {
	marshal, err := json.Marshal(&cnfg)
	if err != nil {
		log.Fatal("Error during Marshal(): ", err)
		return
	}
	if err := os.WriteFile("config.json", marshal, 0644); err != nil {
		log.Fatal("Error during WriteFile(): ", err)
	}
}

// init logger
func init() {
	var err error
	logger.Logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
}

// init discord
func init() {
	var err error
	discord, err = discordgo.New("Bot " + cnfg.Token)
	if err != nil {
		logger.Logger.Fatal("Error creating Discord session", zap.Error(err))
		return
	}
	discord.Identify.Intents |= discordgo.IntentMessageContent
}

// init discord handlers
func init() {
	discord.AddHandler(onReady)
	discord.AddHandler(slashCommandInteraction)
	discord.AddHandler(messageCreate)
}

func onReady(session *discordgo.Session, _ *discordgo.Ready) {
	logger.Logger.Info(session.State.User.Username + " is ready")
}

func slashCommandInteraction(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Type != discordgo.InteractionApplicationCommand {
		return
	}
	commandHandlers[interaction.ApplicationCommandData().Name].Slash(session, interaction)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	regex, err := regexp.Compile("-(?P<name>[\\w-]+)(?:\\s+(?P<args>.*))?")
	if err != nil {
		return
	}
	matches := regex.FindStringSubmatch(m.Content)
	if len(matches) == 0 {
		return
	}
	_, _ = s.ChannelMessageSendReply(m.ChannelID, "Processing...", m.Reference())
	name := matches[regex.SubexpIndex("name")]
	args := matches[regex.SubexpIndex("args")]
	f := commandHandlers[name].Dash
	if f == nil {
		return
	}
	f(s, m, args)
}

func main() {
	commands.CreateDomain.Config = &cnfg
	commands.NS.Config = &cnfg
	commands.CreateEmail.Config = &cnfg
	commands.Automate.Config = &cnfg
	commands.CheckDomain.Config = &cnfg
	commands.Redirect.Config = &cnfg
	commands.DeleteEmail.Config = &cnfg
	commands.ChangePassword.Config = &cnfg
	commands.DeleteFile.Config = &cnfg
	commands.Monitor.Config = &cnfg
	util.Config = &cnfg
	if err := discord.Open(); err != nil {
		logger.Logger.Fatal("Error opening connection", zap.Error(err))
		return
	}
	for _, command := range cmds {
		_, err := discord.ApplicationCommandCreate(discord.State.User.ID, "", command)
		if err != nil {
			logger.Logger.Fatal("Error creating slash command", zap.Error(err))
			return
		}
	}
	logger.Logger.Info("Bot is running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	if err := discord.Close(); err != nil {
		logger.Logger.Fatal("Error closing connection", zap.Error(err))
		return
	}
	logger.Logger.Info("Bot is shutting down")
}
