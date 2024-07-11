package reset

import (
    "database/sql"
    "fmt"
    "hometown-bot/commands"
    "hometown-bot/log"
    "hometown-bot/model"
    "hometown-bot/repository"

    "github.com/bwmarrin/discordgo"
)

const (
    reset           string = "reset"    // Command root
    commandGroup    string = "lobby"    // Command group
    commandCapacity string = "capacity" // Subcommand channel capacity
    commandName     string = "name"     // Subcommand channel name
    optionLobby     string = "lobby"    // Option for [commandCapacity], [commandName]
)

var (
    dmPermission             bool  = false                            // Does not allow using Bot in DMs
    defaultMemberPermissions int64 = discordgo.PermissionManageServer // Caller permission to use commands
    Commands                       = getCommands()
)

type Command struct {
    channelRepository repository.ChannelRepository
    lobbyRepository   repository.LobbyRepository
    commandHandlers   map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) // Command interaction
}

func New(channelRepository repository.ChannelRepository, lobbyRepository repository.LobbyRepository) *Command {
    commands := Command{
        channelRepository: channelRepository,
        lobbyRepository:   lobbyRepository,
    }

    commands.commandHandlers = commands.createCommandHandlers()
    return &commands
}

func (rc *Command) HandleSlashCommands(discord *discordgo.Session, interaction *discordgo.InteractionCreate) {
    if handler, ok := rc.commandHandlers[interaction.ApplicationCommandData().Name]; ok {
        handler(discord, interaction)
    }
}

/* ------ COMMANDS ------ */

func getCommands() []*discordgo.ApplicationCommand {
    return []*discordgo.ApplicationCommand{
        {
            Name:                     reset,
            Description:              "Reset bot settings.",
            DefaultMemberPermissions: &defaultMemberPermissions,
            DMPermission:             &dmPermission,
            Options: []*discordgo.ApplicationCommandOption{
                getLobbyCommandGroup(),
            },
        },
    }
}

func getLobbyCommandGroup() *discordgo.ApplicationCommandOption {
    return &discordgo.ApplicationCommandOption{
        Name:        commandGroup,
        Description: "Lobby settings",
        Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
        Options: []*discordgo.ApplicationCommandOption{
            getCapacityCommand(),
            getNameCommand(),
        },
    }
}

func getCapacityCommand() *discordgo.ApplicationCommandOption {
    return &discordgo.ApplicationCommandOption{
        Name:        commandCapacity,
        Type:        discordgo.ApplicationCommandOptionSubCommand,
        Description: "Set new room capacity to default.",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionChannel,
                Name:        optionLobby,
                Description: "A lobby to be configured.",
                ChannelTypes: []discordgo.ChannelType{
                    discordgo.ChannelTypeGuildVoice,
                },
                Required: true,
            },
        },
    }
}

func getNameCommand() *discordgo.ApplicationCommandOption {
    return &discordgo.ApplicationCommandOption{
        Name:        commandName,
        Type:        discordgo.ApplicationCommandOptionSubCommand,
        Description: "Set new room name to default \"Кімната %nickname%\".",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionChannel,
                Name:        optionLobby,
                Description: "A lobby to be configured.",
                ChannelTypes: []discordgo.ChannelType{
                    discordgo.ChannelTypeGuildVoice,
                },
                Required: true,
            },
        },
    }
}

/* ------ INTERACTIONS ------ */

func (rc *Command) createCommandHandlers() map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    return map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
        reset: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
            log.Info().Printf("trigger %s:%s command interaction", commandGroup, reset)

            slashCommand := i.ApplicationCommandData().Options[0].Name
            subcommandCommand := i.ApplicationCommandData().Options[0].Options[0].Name
            commandResponse := model.CommandError(
                fmt.Sprintf("Oops, something went wrong.Hol' up, you aren't supposed to see this message."),
            )

            switch subcommandCommand {
            case commandCapacity:
                commandResponse = rc.handleCommandCapacity(s, i)
            case commandName:
                commandResponse = rc.handleCommandName(s, i)
            }

            log.Info().Printf("reset: sending interaction response for %s:%s", slashCommand, subcommandCommand)
            if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Embeds: []*discordgo.MessageEmbed{
                        commandResponse.ToEmbededMessage(),
                    },
                    Flags: discordgo.MessageFlagsEphemeral,
                },
            }); err != nil {
                log.Error().Printf("reset: interaction response: %v", err)
            }
        },
    }
}

func (rc *Command) handleCommandCapacity(s *discordgo.Session, i *discordgo.InteractionCreate) model.CommandResponse {
    options := i.ApplicationCommandData().Options
    channel := options[0].Options[0].Options[0].ChannelValue(s)

    if response, err := commands.HasLobby(rc.lobbyRepository, channel, i.GuildID); err != nil {
        log.Warn().Printf("reset: capacity command: %v", err)
        return response
    }

    lobby := model.Lobby{
        Id:         channel.ID,
        CategoryID: channel.ParentID,
        Capacity: sql.NullInt32{
            Valid: true,
            Int32: 0,
        },
    }

    if err := rc.lobbyRepository.UpsertLobby(&lobby); err != nil {
        log.Error().Printf(
            "reset: capacity command: unable to reset capacity for lobby %s[%s]: %v",
            channel.Name,
            channel.ID,
            err,
        )

        return model.CommandError(
            fmt.Sprintf("Unable to reset capacity for \"%s\"", channel.Name),
        )
    }

    log.Info().Printf("reset: capacity command: capacity reset for %s[%s]", channel.Name, lobby.Id)
    return model.CommandSuccess(
        fmt.Sprintf("Capacity successfully reset for \"%s\".", channel.Name),
    )
}

func (rc *Command) handleCommandName(s *discordgo.Session, i *discordgo.InteractionCreate) model.CommandResponse {
    options := i.ApplicationCommandData().Options
    channel := options[0].Options[0].Options[0].ChannelValue(s)

    if response, err := commands.HasLobby(rc.lobbyRepository, channel, i.GuildID); err != nil {
        log.Warn().Printf("reset: capacity command: %v", err)
        return response
    }

    lobby := model.Lobby{
        Id:         channel.ID,
        CategoryID: channel.ParentID,
        Template: sql.NullString{
            Valid:  true,
            String: "",
        },
    }

    if err := rc.lobbyRepository.UpsertLobby(&lobby); err != nil {
        log.Error().Printf(
            "reset: name command: unable to reset name for lobby %s[%s]: %v",
            channel.Name,
            channel.ID,
            err,
        )

        return model.CommandError(
            fmt.Sprintf("Unable to reset name for \"%s\"", channel.Name),
        )
    }

    log.Info().Printf("reset: name command: name reset for %s[%s]", channel.Name, lobby.Id)
    return model.CommandSuccess(
        fmt.Sprintf("Name successfully reset for \"%s\".", channel.Name),
    )
}
