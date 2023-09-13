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

var (
	_config   = config.Config{}
	discord   *discordgo.Session
	_commands = []*discordgo.ApplicationCommand{
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
	commandHandlers = map[string]func(*discordgo.Session, *discordgo.InteractionCreate){
		commands.CreateDomain.Command.Name:   commands.CreateDomain.Execute,
		commands.NS.Command.Name:             commands.NS.Execute,
		commands.Add.Command.Name:            commands.Add.Execute,
		commands.CreateEmail.Command.Name:    commands.CreateEmail.Execute,
		commands.SSL.Command.Name:            commands.SSL.Execute,
		commands.UploadFile.Command.Name:     commands.UploadFile.Execute,
		commands.Automate.Command.Name:       commands.Automate.Execute,
		commands.CheckDomain.Command.Name:    commands.CheckDomain.Execute,
		commands.Redirect.Command.Name:       commands.Redirect.Execute,
		commands.ChangePassword.Command.Name: commands.ChangePassword.Execute,
		commands.DeleteEmail.Command.Name:    commands.DeleteEmail.Execute,
		commands.DeleteFile.Command.Name:     commands.DeleteFile.Execute,
		commands.Monitor.Command.Name:        commands.Monitor.Execute,
	}
	dashCommandHandlers = map[string]func(*discordgo.Session, *discordgo.MessageCreate, string){
		commands.Add.Command.Name:            commands.Add.ExecuteDash,
		commands.ChangePassword.Command.Name: commands.ChangePassword.ExecuteDash,
	}
)

func init() {
	file, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Print("Enter bot token: ")
		if _, err := fmt.Scanln(&_config.Token); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter api user: ")
		if _, err := fmt.Scanln(&_config.ApiUser); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter api key: ")
		if _, err := fmt.Scanln(&_config.ApiKey); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter user name: ")
		if _, err := fmt.Scanln(&_config.UserName); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter client ip: ")
		if _, err := fmt.Scanln(&_config.ClientIP); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter default password: ")
		if _, err := fmt.Scanln(&_config.DefaultPassword); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter basic auth username: ")
		if _, err := fmt.Scanln(&_config.BasicAuth.Username); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter basic auth password: ")
		if _, err := fmt.Scanln(&_config.BasicAuth.Password); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter default name servers: ")
		if _, err := fmt.Scanln(&_config.DefaultNameServers); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		configJson()
		return
	}
	if err := json.Unmarshal(file, &_config); err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
}

func configJson() {
	marshal, err := json.Marshal(&_config)
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
	discord, err = discordgo.New("Bot " + _config.Token)
	if err != nil {
		logger.Logger.Fatal("Error creating Discord session", zap.Error(err))
		return
	}
	discord.Identify.Intents = discordgo.IntentMessageContent
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
	commandHandlers[interaction.ApplicationCommandData().Name](session, interaction)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	regex, err := regexp.Compile("-(?P<name>[\\w-]+)(?:\\s+(?P<args>.*))?")
	if err != nil {
		return
	}
	matches := regex.FindStringSubmatch(m.Content)
	if len(matches) == 0 {
		return
	}
	name := matches[regex.SubexpIndex("name")]
	args := matches[regex.SubexpIndex("args")]
	f := dashCommandHandlers[name]
	if f != nil {
		return
	}
	f(s, m, args)
}

func main() {
	commands.CreateDomain.Config = &_config
	commands.NS.Config = &_config
	commands.CreateEmail.Config = &_config
	commands.Automate.Config = &_config
	commands.CheckDomain.Config = &_config
	commands.Redirect.Config = &_config
	commands.DeleteEmail.Config = &_config
	commands.ChangePassword.Config = &_config
	commands.DeleteFile.Config = &_config
	commands.Monitor.Config = &_config
	util.Config = &_config
	if err := discord.Open(); err != nil {
		logger.Logger.Fatal("Error opening connection", zap.Error(err))
		return
	}
	for _, command := range _commands {
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
