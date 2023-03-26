package bot

import (
	"hometown-bot/commands/lobby"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

var (
	GuildID        string        // Discord Server ID
	BotToken       string        // Discord BOT API token
	RemoveCommands bool   = true // Should remove slash commands when bot offline. Default - true.
)

func init() {
	// log.Println("Preloading storage...")
	// storage.Load()
}

func Run() {
	discord, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatal(err)
	}

	lobby.GuildID = GuildID

	log.Println("Bot created! Attaching handlers...")
	discord.AddHandler(lobby.HandleSlashCommands)
	discord.AddHandler(lobby.HandleVoiceUpdates)

	discord.Open()

	log.Println("Websocket open! Creating commands...")
	registeredCommands := createCommands(discord)

	defer discord.Close()

	log.Println("Bot running...")
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	<-channel

	log.Println("Bot stopped! Removing slash commands...")
	removeSlashCommands(discord, registeredCommands)
}

func createCommands(discord *discordgo.Session) []*discordgo.ApplicationCommand {
	registeredCommands := make([]*discordgo.ApplicationCommand, len(lobby.Commands))
	for i, v := range lobby.Commands {
		cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	return registeredCommands
}

func removeSlashCommands(discord *discordgo.Session, registeredCommands []*discordgo.ApplicationCommand) {
	if RemoveCommands {
		for _, v := range registeredCommands {
			err := discord.ApplicationCommandDelete(discord.State.User.ID, GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v\n", v.Name, err)
			}
		}
	}
}
