package bot

import (
    "fmt"
    "hometown-bot/commands/lobby"
    "hometown-bot/commands/message"
    "hometown-bot/commands/reset"
    "hometown-bot/log"
    "hometown-bot/repository"
    "os"
    "os/signal"

    "github.com/bwmarrin/discordgo"
)

// Token - discord bot API
var Token string

type Bot struct {
    channelRepository        repository.ChannelRepository
    channelMembersRepository repository.ChannelMembersRepository
    lobbyRepository          repository.LobbyRepository
}

func Create(
    channelRepository repository.ChannelRepository,
    channelMembersRepository repository.ChannelMembersRepository,
    lobbyRepository repository.LobbyRepository,
) *Bot {
    return &Bot{
        channelRepository:        channelRepository,
        channelMembersRepository: channelMembersRepository,
        lobbyRepository:          lobbyRepository,
    }
}

func (bot *Bot) Run() error {
    log.Debug().Println("bot: start new session")
    discord, err := discordgo.New("Bot " + Token)
    if err != nil {
        return fmt.Errorf("unable to create a new bot session: %w", err)
    }

    log.Debug().Println("bot: load commands")
    lobbyCommands := lobby.New(bot.channelRepository, bot.channelMembersRepository, bot.lobbyRepository)
    resetCommands := reset.New(bot.channelRepository, bot.lobbyRepository)
    messageCommands := message.New()

    log.Debug().Println("bot: attach handlers for commands")
    discord.AddHandler(lobbyCommands.HandleSlashCommands)
    discord.AddHandler(resetCommands.HandleSlashCommands)
    discord.AddHandler(lobbyCommands.HandleVoiceUpdates)
    discord.AddHandler(messageCommands.HandleSlashCommands)

    log.Debug().Println("bot: establish socket connection")
    if err := discord.Open(); err != nil {
        return fmt.Errorf("unable to create socket: %w", err)
    }

    log.Debug().Println("bot: create commands for discord")
    registeredCommands, err := createCommands(discord)
    if err != nil {
        return fmt.Errorf("unable to register bot commands: %w", err)
    }

    defer func(discord *discordgo.Session) {
        if err := discord.Close(); err != nil {
            log.Error().Printf("bot: unable to close bot socket: %v", err)
        }
    }(discord)

    log.Info().Println("bot: running..")
    channel := make(chan os.Signal, 1)
    signal.Notify(channel, os.Interrupt)
    <-channel

    log.Info().Println("bot: stopping.. removing slash commands")
    if err := removeSlashCommands(discord, registeredCommands); err != nil {
        return fmt.Errorf("unable to remove bot commands: %w", err)
    }

    return nil
}

func createCommands(discord *discordgo.Session) ([]*discordgo.ApplicationCommand, error) {
    joinedCommands := append(lobby.Commands, reset.Commands...)
    joinedCommands = append(joinedCommands, message.Commands...)

    registeredCommands := make([]*discordgo.ApplicationCommand, len(joinedCommands))
    for i, v := range joinedCommands {
        log.Debug().Printf("bot: add command: %s", v.Name)
        cmd, err := discord.ApplicationCommandCreate(
            discord.State.User.ID,
            discord.State.Application.GuildID,
            v,
        )
        if err != nil {
            return nil, fmt.Errorf("cannot create '%s' command: %w", v.Name, err)
        }
        registeredCommands[i] = cmd
    }

    return registeredCommands, nil
}

func removeSlashCommands(discord *discordgo.Session, registeredCommands []*discordgo.ApplicationCommand) error {
    for _, v := range registeredCommands {
        log.Debug().Printf("bot: remove command: %s[%s]", v.Name, v.ID)
        err := discord.ApplicationCommandDelete(
            discord.State.User.ID,
            discord.State.Application.GuildID,
            v.ID,
        )
        if err != nil {
            return fmt.Errorf("cannot delete '%s' command: %w", v.Name, err)
        }
    }

    return nil
}
