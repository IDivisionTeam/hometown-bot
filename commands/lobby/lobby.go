package lobby

import (
    "database/sql"
    "errors"
    "fmt"
    "hometown-bot/commands"
    "hometown-bot/log"
    "hometown-bot/model"
    "hometown-bot/repository"
    "strconv"
    "strings"

    "github.com/bwmarrin/discordgo"
)

const (
    lobby           string = "lobby"    // Command group
    commandRegister string = "register" // Subcommand lobby register
    commandCapacity string = "capacity" // Subcommand channel capacity
    commandName     string = "name"     // Subcommand channel name
    commandList     string = "list"     // Subcommand lobby list
    commandRemove   string = "remove"   // Subcommand lobby remove
    optionChannel   string = "channel"  // Option for commandRegister
    optionLobby     string = "lobby"    // Option for commandCapacity, commandName, commandRemove
    optionCapacity  string = "capacity" // Option for commandCapacity
    optionName      string = "name"     // Option for commandName
)

var (
    dmPermission             bool  = false                            // Does not allow using Bot in DMs
    defaultMemberPermissions int64 = discordgo.PermissionManageServer // Caller permission to use commands
    Commands                       = getLobbyCommandGroup()           // Command group
)

type Command struct {
    channelRepository        repository.ChannelRepository
    channelMembersRepository repository.ChannelMembersRepository
    lobbyRepository          repository.LobbyRepository
    commandHandlers          map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) // Command interaction
}

func New(
    channelRepository repository.ChannelRepository,
    channelMembersRepository repository.ChannelMembersRepository,
    lobbyRepository repository.LobbyRepository,
) *Command {
    commands := Command{
        channelRepository:        channelRepository,
        channelMembersRepository: channelMembersRepository,
        lobbyRepository:          lobbyRepository,
    }

    commands.commandHandlers = commands.createCommandHandlers()
    return &commands
}

func (lc *Command) HandleSlashCommands(discord *discordgo.Session, interaction *discordgo.InteractionCreate) {
    if handler, ok := lc.commandHandlers[interaction.ApplicationCommandData().Name]; ok {
        handler(discord, interaction)
    }
}

// FIXME: split into small functions
func (lc *Command) HandleVoiceUpdates(s *discordgo.Session, event *discordgo.VoiceStateUpdate) {
    channels, err := lc.channelRepository.GetChannels()
    if err != nil {
        log.Error().Printf("voice updates: get channels: %v", err)
    }

    isSomeoneLeftVoiceChannel := event.BeforeUpdate != nil && event.BeforeUpdate.ChannelID != ""
    if isSomeoneLeftVoiceChannel {
        for _, channel := range channels {
            channelId := event.BeforeUpdate.ChannelID

            if channel.Id == channelId {
                userId := event.VoiceState.Member.User.ID
                log.Info().Printf("voice updates: remove user %s from the channel %s upon leaving", userId, channelId)

                if err := lc.channelMembersRepository.DeleteChannelMember(event.GuildID, userId, channel.Id); err != nil {
                    log.Error().Printf("voice updates: delete member count: %v", err)
                }
            }
        }
    }

    isSomeoneJoinVoiceChannel := event.VoiceState != nil && event.VoiceState.ChannelID != ""
    if isSomeoneJoinVoiceChannel {
        for _, channel := range channels {
            channelId := event.VoiceState.ChannelID

            if channel.Id == channelId {
                userId := event.VoiceState.Member.User.ID
                log.Info().Printf("voice updates: add user %s to the channel %s upon join", userId, channelId)

                if err := lc.channelMembersRepository.SetChannelMember(event.GuildID, userId, channelId); err != nil {
                    log.Error().Printf("voice updates: insert member count: %v", err)
                }
            }
        }
    }

    channels, err = lc.channelRepository.GetChannels()
    if err != nil {
        log.Error().Printf("voice updates: get channels: %v", err)
    }

    log.Info().Println("voice updates: verify self-destructing channels if they have enough people to exist")
    for _, channel := range channels {
        channelMembersCount, err := lc.channelMembersRepository.GetChannelMembersCount(event.GuildID, channel.Id)

        switch {
        case errors.Is(err, sql.ErrNoRows):
            log.Error().Printf("voice updates: get channel members count: %v", err)
            continue
        case err != nil:
            log.Error().Printf("voice updates: get channel members count: %v", err)
            continue
        }

        shouldDeleteSelfDestructingChannel := channelMembersCount == 0
        if shouldDeleteSelfDestructingChannel {
            log.Info().Printf("voice updates: channel %s is empty, deleting..", channel.Id)

            if _, err := s.ChannelDelete(channel.Id); err != nil {
                log.Error().Printf("voice updates: API: unable to delete channel %s: %v", channel.Id, err)
                continue
            }

            if err := lc.channelRepository.DeleteChannel(channel.Id); err != nil {
                log.Error().Printf("voice updates: db: unable to delete channel %s: %v", channel.Id, err)
                continue
            }

            if err := lc.channelMembersRepository.DeleteChannelMembers(event.GuildID, channel.Id); err != nil {
                log.Error().Printf("voice updates: db: unable to delete channel members %s: %v", channel.Id, err)
                continue
            }
        }
    }

    if event == nil {
        log.Warn().Println("voice updates: event is empty, skip")
        return
    }

    previousState := event.BeforeUpdate
    isChannelIdentical := previousState != nil && previousState.ChannelID == event.ChannelID
    if isChannelIdentical {
        log.Warn().Println("voice updates: received updates for the same channel, skip")
        return
    }

    // TODO: verify that logic works as expected for all users
    lobbies, err := lc.lobbyRepository.GetLobbies(event.GuildID)
    if err != nil {
        log.Error().Printf("voice updates: get lobbies: %v", err)
    }

    for _, l := range lobbies {
        hasLobby := l.Id == event.ChannelID

        if hasLobby {
            userName := event.Member.User.Username
            nickname := event.Member.Nick
            name := ""

            if len(nickname) == 0 {
                name = userName
            } else {
                name = nickname
            }

            if l.Template.Valid && l.Template.String != "" {
                name = fmt.Sprintf("%s %s", l.Template.String, name)
            } else {
                name = fmt.Sprintf("Кімната %s", name)
            }

            userLimit := 0
            if l.Capacity.Valid {
                userLimit = int(l.Capacity.Int32)
            }

            data := discordgo.GuildChannelCreateData{
                Name:      name,
                Type:      discordgo.ChannelTypeGuildVoice,
                ParentID:  l.CategoryID,
                UserLimit: userLimit,
            }

            log.Info().Printf("voice updates: creating self-destructing channel %s", name)

            newChannel, err := s.GuildChannelCreateComplex(event.GuildID, data)
            if err != nil {
                log.Error().Printf("voice updates: unable to create self-destructing channel: %v", err)
                continue
            }

            channel := model.Channel{
                Id:       newChannel.ID,
                ParentID: l.Id,
            }

            if err := lc.channelRepository.SetChannel(&channel); err != nil {
                log.Error().Printf("voice updates: unable to save self-destructing channel: %v", err)
                return
            }

            log.Info().Printf(
                "voice updates: move channel creator %s[%s] to the channel %s",
                event.Member.User.Username,
                event.Member.User.ID,
                name,
            )

            if err := s.GuildMemberMove(event.GuildID, event.Member.User.ID, &newChannel.ID); err != nil {
                log.Error().Printf(
                    "voice updates: unable to move channel creator %s[%s] to the channel %s: %v",
                    event.Member.User.Username,
                    event.Member.User.ID,
                    name,
                    err,
                )
                continue
            }
        }
    }
}

/* ------ COMMANDS ------ */

func getLobbyCommandGroup() []*discordgo.ApplicationCommand {
    return []*discordgo.ApplicationCommand{
        {
            Name:                     lobby,
            Description:              "Lobbies' commands group.",
            DefaultMemberPermissions: &defaultMemberPermissions,
            DMPermission:             &dmPermission,
            Options: []*discordgo.ApplicationCommandOption{
                getRegisterCommand(),
                getCapacityCommand(),
                getNameCommand(),
                getListCommand(),
                getRemoveCommand(),
            },
        },
    }
}

func getRegisterCommand() *discordgo.ApplicationCommandOption {
    return &discordgo.ApplicationCommandOption{
        Name:        commandRegister,
        Type:        discordgo.ApplicationCommandOptionSubCommand,
        Description: "Register a new lobby.",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionChannel,
                Name:        optionChannel,
                Description: "A channel to be registered.",
                ChannelTypes: []discordgo.ChannelType{
                    discordgo.ChannelTypeGuildVoice,
                },
                Required: true,
            },
        },
    }
}

func getCapacityCommand() *discordgo.ApplicationCommandOption {
    return &discordgo.ApplicationCommandOption{
        Name:        commandCapacity,
        Type:        discordgo.ApplicationCommandOptionSubCommand,
        Description: "Select new lobbies' capacity.",
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
            {
                Type:        discordgo.ApplicationCommandOptionInteger,
                Name:        optionCapacity,
                Description: "A new lobbies' capacity.",
                Required:    true,
            },
        },
    }
}

func getNameCommand() *discordgo.ApplicationCommandOption {
    return &discordgo.ApplicationCommandOption{
        Name:        commandName,
        Type:        discordgo.ApplicationCommandOptionSubCommand,
        Description: "Select new channels' name when created.",
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
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        optionName,
                Description: "A new channels' name when created.",
                Required:    true,
            },
        },
    }
}

func getListCommand() *discordgo.ApplicationCommandOption {
    return &discordgo.ApplicationCommandOption{
        Name:        commandList,
        Type:        discordgo.ApplicationCommandOptionSubCommand,
        Description: "Show registered lobbies.",
    }
}

func getRemoveCommand() *discordgo.ApplicationCommandOption {
    return &discordgo.ApplicationCommandOption{
        Name:        commandRemove,
        Type:        discordgo.ApplicationCommandOptionSubCommand,
        Description: "Remove an existing lobby.",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionChannel,
                Name:        optionLobby,
                Description: "A lobby to be removed.",
                ChannelTypes: []discordgo.ChannelType{
                    discordgo.ChannelTypeGuildVoice,
                },
                Required: true,
            },
        },
    }
}

/* ------ INTERACTIONS ------ */

func (lc *Command) createCommandHandlers() map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    return map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
        lobby: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
            log.Info().Printf("trigger %s command interaction", lobby)

            slashCommand := i.ApplicationCommandData().Options[0].Name
            commandResponse := model.CommandError(
                fmt.Sprintf("Oops, something went wrong.\nHol' up, you aren't supposed to see this message."),
            )

            switch slashCommand {
            case commandRegister:
                commandResponse = lc.handleCommandRegister(s, i)
            case commandCapacity:
                commandResponse = lc.handleCommandCapacity(s, i)
            case commandName:
                commandResponse = lc.handleCommandName(s, i)
            case commandList:
                commandResponse = lc.handleCommandList(s, i)
            case commandRemove:
                commandResponse = lc.handleCommandRemove(s, i)
            }

            log.Info().Printf("lobby: sending interaction response for %s", slashCommand)
            if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Embeds: []*discordgo.MessageEmbed{
                        commandResponse.ToEmbededMessage(),
                    },
                    Flags: discordgo.MessageFlagsEphemeral,
                },
            }); err != nil {
                log.Error().Printf("lobby: interaction response: %v", err)
            }
        },
    }
}

func (lc *Command) handleCommandRegister(s *discordgo.Session, i *discordgo.InteractionCreate) model.CommandResponse {
    options := i.ApplicationCommandData().Options
    channel := options[0].Options[0].ChannelValue(s)

    lobby := model.Lobby{
        Id:         channel.ID,
        CategoryID: channel.ParentID,
        GuildID:    i.GuildID,
    }

    affectedRows, err := lc.lobbyRepository.SetLobby(&lobby)
    if err != nil {
        log.Error().Printf("lobby: register command: unable to upsert %s[%s]: %v", channel.Name, lobby.Id, err)
        return model.CommandError(
            fmt.Sprintf("Lobby \"%s\" cannot be registered.", channel.Name),
        )
    }

    if affectedRows == 0 {
        log.Warn().Printf("lobby: register command: %s[%s] already registered as lobby", channel.Name, lobby.Id)
        return model.CommandWarning(
            fmt.Sprintf("\"%s\" is already registered as a lobby!", channel.Name),
        )
    }

    log.Info().Printf("lobby: register command: save lobby %s[%s]", channel.Name, lobby.Id)
    return model.CommandSuccess(
        fmt.Sprintf("Lobby \"%s\" successfully registered.", channel.Name),
    )
}

func (lc *Command) handleCommandCapacity(s *discordgo.Session, i *discordgo.InteractionCreate) model.CommandResponse {
    options := i.ApplicationCommandData().Options
    channel := options[0].Options[0].ChannelValue(s)
    capacity := options[0].Options[1].IntValue()

    if capacity <= 0 {
        log.Warn().Printf("lobby: capacity command: capacity = %d for %s[%s], it cannot be negative or zero", capacity, channel.Name, channel.ID)
        return model.CommandWarning("User limit cannot be negative or zero!")
    }

    if response, err := commands.HasLobby(lc.lobbyRepository, channel, i.GuildID); err != nil {
        log.Warn().Printf("lobby: capacity command: %v", err)
        return response
    }

    lobby := model.Lobby{
        Id:         channel.ID,
        CategoryID: channel.ParentID,
        Capacity: sql.NullInt32{
            Valid: true,
            Int32: int32(capacity),
        },
    }

    if err := lc.lobbyRepository.UpsertLobby(&lobby); err != nil {
        log.Error().Printf("lobby: capacity command: unable to update lobby %s[%s]: %v", channel.Name, channel.ID, err)
        return model.CommandError(
            fmt.Sprintf("Unable to update capacity for \"%s\"", channel.Name),
        )
    }

    log.Info().Printf("lobby: capacity command: save capacity %d for %s[%s]", capacity, channel.Name, lobby.Id)
    return model.CommandSuccess(
        fmt.Sprintf("Capacity %d successfully set for \"%s\".", capacity, channel.Name),
    )
}

func (lc *Command) handleCommandName(s *discordgo.Session, i *discordgo.InteractionCreate) model.CommandResponse {
    options := i.ApplicationCommandData().Options
    channel := options[0].Options[0].ChannelValue(s)
    name := options[0].Options[1].StringValue()

    if response, err := commands.HasLobby(lc.lobbyRepository, channel, i.GuildID); err != nil {
        log.Warn().Printf("lobby: name command: %v", err)
        return response
    }

    lobby := model.Lobby{
        Id:         channel.ID,
        CategoryID: channel.ParentID,
        Template: sql.NullString{
            Valid:  true,
            String: name,
        },
    }

    if err := lc.lobbyRepository.UpsertLobby(&lobby); err != nil {
        log.Error().Printf(
            "lobby: name command: unable to update lobby name = %s %s[%s]: %v",
            name,
            channel.Name,
            lobby.Id,
            err,
        )

        return model.CommandError(
            fmt.Sprintf("Unable to update name = \"%s\" for \"%s\"", name, channel.Name),
        )
    }

    log.Info().Printf("lobby: name command: save name %s for %s[%s]", name, channel.Name, lobby.Id)
    return model.CommandSuccess(
        fmt.Sprintf("Name \"%s\" successfully set for \"%s\".", name, channel.Name),
    )
}

func (lc *Command) handleCommandList(s *discordgo.Session, i *discordgo.InteractionCreate) model.CommandResponse {
    lobbies, err := lc.lobbyRepository.GetLobbies(i.GuildID)
    if err != nil {
        log.Error().Printf("lobby: list command: unable to get lobby list for Guild[%s]: %v", i.GuildID, err)
        return model.CommandError(
            fmt.Sprintf("There are no registered lobbies for this Discord Server!"),
        )
    }

    var registeredChannels []string
    for i, lobby := range lobbies {
        channel, err := s.Channel(lobby.Id)
        if err != nil {
            log.Error().Printf("lobby: list command: unable to get channels: %v", err)

            return model.CommandError(
                fmt.Sprintf("Unable to get Discord channels!"),
            )
        }

        if channel.ID == lobby.Id {
            var template string
            if lobby.Template.Valid && len(lobby.Template.String) > 0 {
                template = lobby.Template.String
            } else {
                template = "Кімната %username%"
            }

            var capacity string
            if lobby.Capacity.Valid {
                capacity = strconv.FormatInt(int64(lobby.Capacity.Int32), 10)
            } else if lobby.Capacity.Int32 == 0 {
                capacity = "unlimited"
            } else {
                capacity = "unlimited"
            }

            lobbyIndex := i + 1
            finalString := fmt.Sprintf(
                "%d. Name: %s, Channel template: %s, Capacity: %s",
                lobbyIndex,
                channel.Name,
                template,
                capacity,
            )

            log.Debug().Printf("lobby: list command: channels\n%s", finalString)
            registeredChannels = append(registeredChannels, finalString)
        }
    }

    if len(registeredChannels) == 0 {
        log.Warn().Println("lobby: list command: there are no registered channels")
        return model.CommandWarning("There are no active lobbies.")
    }

    activeLobbies := strings.Join(registeredChannels, "\n")
    log.Info().Printf("lobby: list command: active lobbies found:\n%s", activeLobbies)
    return model.CommandSuccess(
        fmt.Sprintf("Active Lobbies:\n%s", activeLobbies),
    )
}

func (lc *Command) handleCommandRemove(s *discordgo.Session, i *discordgo.InteractionCreate) model.CommandResponse {
    options := i.ApplicationCommandData().Options
    channel := options[0].Options[0].ChannelValue(s)

    affectedRows, err := lc.lobbyRepository.DeleteLobby(channel.ID, i.GuildID)
    if err != nil {
        log.Error().Printf("lobby: remove command: unable to delete lobby %s[%s]: %v", channel.Name, channel.ID, err)

        return model.CommandError(
            fmt.Sprintf("Unable to delete \"%s\" lobby.", channel.Name),
        )
    }

    if affectedRows == 0 {
        log.Warn().Printf("lobby: remove command: db: %s[%s] is not a lobby", channel.Name, channel.ID)
        return model.CommandWarning(
            fmt.Sprintf("\"%s\" is not a lobby!", channel.Name),
        )
    }

    log.Info().Printf("lobby: remove command: lobby %s successfully deleted", channel.Name)
    return model.CommandSuccess(
        fmt.Sprintf("Lobby \"%s\" successfully deleted", channel.Name),
    )
}
