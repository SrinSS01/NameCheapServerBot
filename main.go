package main

import (
	"NS/commands"
	config2 "NS/config"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	config    = config2.Config{}
	Logger    *zap.Logger
	discord   *discordgo.Session
	_commands = []*discordgo.ApplicationCommand{
		commands.CreateDomain.Command,
		commands.NS.Command,
		commands.Add.Command,
		commands.CreateEmail.Command,
		commands.SSL.Command,
	}
	commandHandlers = map[string]func(*discordgo.Session, *discordgo.InteractionCreate){
		commands.CreateDomain.Command.Name: commands.CreateDomain.Execute,
		commands.NS.Command.Name:           commands.NS.Execute,
		commands.Add.Command.Name:          commands.Add.Execute,
		commands.CreateEmail.Command.Name:  commands.CreateEmail.Execute,
		commands.SSL.Command.Name:          commands.SSL.Execute,
	}
)

func init() {
	file, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Print("Enter bot token: ")
		if _, err := fmt.Scanln(&config.Token); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter api user: ")
		if _, err := fmt.Scanln(&config.ApiUser); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter api key: ")
		if _, err := fmt.Scanln(&config.ApiKey); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter user name: ")
		if _, err := fmt.Scanln(&config.UserName); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter client ip: ")
		if _, err := fmt.Scanln(&config.ClientIP); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		fmt.Print("Enter default password: ")
		if _, err := fmt.Scanln(&config.DefaultPassword); err != nil {
			log.Fatal("Error during Scanln(): ", err)
		}
		configJson()
		return
	}
	if err := json.Unmarshal(file, &config); err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
}

func configJson() {
	marshal, err := json.Marshal(&config)
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
	Logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
}

// init discord
func init() {
	var err error
	discord, err = discordgo.New("Bot " + config.Token)
	if err != nil {
		Logger.Fatal("Error creating Discord session", zap.Error(err))
		return
	}
	discord.Identify.Intents = discordgo.IntentMessageContent
}

// init discord handlers
func init() {
	discord.AddHandler(onReady)
	discord.AddHandler(slashCommandInteraction)
}

func onReady(session *discordgo.Session, _ *discordgo.Ready) {
	Logger.Info(session.State.User.Username + " is ready")
}

func slashCommandInteraction(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Type != discordgo.InteractionApplicationCommand {
		return
	}
	commandHandlers[interaction.ApplicationCommandData().Name](session, interaction)
}

func main() {
	commands.CreateDomain.Config = &config
	commands.NS.Config = &config
	commands.Add.Config = &config
	commands.CreateEmail.Config = &config
	if err := discord.Open(); err != nil {
		Logger.Fatal("Error opening connection", zap.Error(err))
		return
	}
	for _, command := range _commands {
		_, err := discord.ApplicationCommandCreate(discord.State.User.ID, "", command)
		if err != nil {
			Logger.Fatal("Error creating slash command", zap.Error(err))
			return
		}
	}
	Logger.Info("Bot is running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	if err := discord.Close(); err != nil {
		Logger.Fatal("Error closing connection", zap.Error(err))
		return
	}
	Logger.Info("Bot is shutting down")
}
