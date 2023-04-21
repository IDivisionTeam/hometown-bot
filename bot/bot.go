package bot

import (
	"fmt"
	"hometown-bot/commands/lobby"
	"hometown-bot/commands/reset"
	"hometown-bot/repository"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

var (
	BotToken       string        // Discord BOT API token
	RemoveCommands bool   = true // Should remove slash commands when bot offline. Default - true.
)

type Bot struct {
	channelRepository repository.ChannelRepository
	lobbyRepository   repository.LobbyRepository
}

func Create(channelRepository repository.ChannelRepository, lobbyRepository repository.LobbyRepository) *Bot {
	return &Bot{
		channelRepository: channelRepository,
		lobbyRepository:   lobbyRepository,
	}
}

func (bot *Bot) Run() error {
	discord, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		return fmt.Errorf("unable to create a new bot session: %w", err)
	}

	lobbyCommands := lobby.New(bot.channelRepository, bot.lobbyRepository)
	resetCommands := reset.New(bot.channelRepository, bot.lobbyRepository)

	log.Println("Bot created! Attaching handlers...")
	discord.AddHandler(lobbyCommands.HandleSlashCommands)
	discord.AddHandler(resetCommands.HandleSlashCommands)
	discord.AddHandler(lobbyCommands.HandleVoiceUpdates)

	err = discord.Open()
	if err != nil {
		return fmt.Errorf("unable to create socket: %w", err)
	}

	log.Println("Websocket open! Creating commands...")
	registeredCommands, err := createCommands(discord)
	if err != nil {
		return fmt.Errorf("unable to register bot commands: %w", err)
	}

	defer discord.Close()

	log.Println("Bot running...")
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	<-channel

	log.Println("Bot stopped! Removing slash commands...")
	err = removeSlashCommands(discord, registeredCommands)
	if err != nil {
		return fmt.Errorf("unable to remove bot commands: %w", err)
	}

	return nil
}

func createCommands(discord *discordgo.Session) ([]*discordgo.ApplicationCommand, error) {
	joinedCommands := append(lobby.Commands, reset.Commands...)

	registeredCommands := make([]*discordgo.ApplicationCommand, len(joinedCommands))
	for i, v := range joinedCommands {
		cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, discord.State.Application.GuildID, v)
		if err != nil {
			return nil, fmt.Errorf("cannot create '%s' command: %w", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	return registeredCommands, nil
}

func removeSlashCommands(discord *discordgo.Session, registeredCommands []*discordgo.ApplicationCommand) error {
	if RemoveCommands {
		for _, v := range registeredCommands {
			err := discord.ApplicationCommandDelete(discord.State.User.ID, discord.State.Application.GuildID, v.ID)
			if err != nil {
				return fmt.Errorf("cannot delete '%s' command: %w", v.Name, err)
			}
		}
	}
	return nil
}
