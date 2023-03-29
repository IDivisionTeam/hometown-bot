package lobby

import (
	"database/sql"
	"fmt"
	"hometown-bot/model"
	"hometown-bot/repository"
	"hometown-bot/utils/color"
	"log"
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
	optionChannel   string = "channel"  // Option for [commandRegister]
	optionLobby     string = "lobby"    // Option for [commandCapacity], [commandName], [commandRemove]
	optionCapacity  string = "capacity" // Option for [commandCapacity]
	optionName      string = "name"     // Option for [commandName]
)

type CommandResponse struct {
	title       string
	description string
	colorType   color.ColorType
}

var (
	GuildID                  string                                    // Discord Server ID
	dmPermission             bool   = false                            // Does not allow using Bot in DMs
	defaultMemberPermissions int64  = discordgo.PermissionManageServer // Caller permission to use commands
	Commands                        = getLobbyCommandGroup()           // Command group
)

type LobbyCommands struct {
	channelRepository repository.ChannelRepository
	lobbyRepository   repository.LobbyRepository
	commandHandlers   map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) // Command interaction
}

func New(channelRepository repository.ChannelRepository, lobbyRepository repository.LobbyRepository) *LobbyCommands {
	commands := LobbyCommands{
		channelRepository: channelRepository,
		lobbyRepository:   lobbyRepository,
	}

	commands.commandHandlers = commands.getCommandHandlers()
	return &commands
}

func (lc *LobbyCommands) HandleSlashCommands(discord *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if handler, ok := lc.commandHandlers[interaction.ApplicationCommandData().Name]; ok {
		handler(discord, interaction)
	}
}

// FIXME: split into small functions
func (lc *LobbyCommands) HandleVoiceUpdates(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	if vsu == nil {
		return
	}

	previousState := vsu.BeforeUpdate
	if previousState != nil && previousState.ChannelID == vsu.ChannelID {
		return
	}

	lobbies, _ := lc.lobbyRepository.GetLobbies()

	for _, l := range lobbies {
		if l.Id == vsu.ChannelID {
			log.Println("Searching for lobby...")

			userName := vsu.Member.User.Username
			nickname := vsu.Member.Nick
			name := ""
			if len(nickname) == 0 {
				name = userName
			} else {
				name = nickname
			}

			if l.Template.Valid {
				name = l.Template.String + " " + name
			} else {
				name = "Кімната " + name
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

			log.Println("Found! Creating a new channel...")
			newChannel, err := s.GuildChannelCreateComplex(GuildID, data)
			if err != nil {
				log.Println("unable to create a new channel! %w", err)
				continue
			}

			channel := model.Channel{
				Id:       newChannel.ID,
				ParentID: l.Id,
			}

			lc.channelRepository.SetChannel(&channel)

			log.Println("Created! Moving a user to the new channel...")
			err = s.GuildMemberMove(GuildID, vsu.Member.User.ID, &newChannel.ID)
			if err != nil {
				log.Println("unable to move a user to the new channel! %w", err)
				continue
			}
		}
	}

	if previousState != nil {
		channels, _ := lc.channelRepository.GetChannels()
		for _, tc := range channels {
			if tc.Id == previousState.ChannelID {
				channel, err := s.Channel(tc.Id)
				if err != nil {
					log.Println("unable to find a channel! %w", err)
					continue
				}

				if len(channel.Members) == 0 {
					log.Println("Channel " + channel.Name + " is empty, deleting...")

					_, err := s.ChannelDelete(tc.Id)
					if err != nil {
						log.Println("unable to to delete the channel! %w", err)
						continue
					}

					lc.channelRepository.DeleteChannel(tc.Id)
				}
			}
		}
	}
}

/* Commands */

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

/* Interactions */

func (lc *LobbyCommands) getCommandHandlers() map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		lobby: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			slashCommand := i.ApplicationCommandData().Options[0].Name
			commandResponse := CommandResponse{
				title: "🚨 Error",
				description: "Oops, something went wrong.\n" +
					"Hol' up, you aren't supposed to see this message.",
				colorType: color.Failure,
			}

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

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:       commandResponse.title,
							Description: commandResponse.description,
							Color:       color.GetColorFrom(commandResponse.colorType),
						},
					},
					Flags: discordgo.MessageFlagsEphemeral,
				},
			})
		},
	}
}

func (lc *LobbyCommands) handleCommandRegister(s *discordgo.Session, i *discordgo.InteractionCreate) CommandResponse {
	options := i.ApplicationCommandData().Options
	channel := options[0].Options[0].ChannelValue(s)

	lobby := model.Lobby{
		Id:         channel.ID,
		CategoryID: channel.ParentID,
	}

	affectedRows, err := lc.lobbyRepository.SetLobby(&lobby)
	if err != nil {
		log.Println("unable to upsert the lobby! %w", err)

		return CommandResponse{
			title:       "🚨 Error",
			description: "Lobby \"" + channel.Name + "\" cannot be registered.",
			colorType:   color.Failure,
		}
	}

	if affectedRows == 0 {
		return CommandResponse{
			title:       "🧀 Warning",
			description: "\"" + channel.Name + "\" is already registered as a lobby!",
			colorType:   color.Warning,
		}
	}

	return CommandResponse{
		title:       "✅ OK",
		description: "Lobby \"" + channel.Name + "\" successfully registered.",
		colorType:   color.Success,
	}
}

func (lc *LobbyCommands) handleCommandCapacity(s *discordgo.Session, i *discordgo.InteractionCreate) CommandResponse {
	options := i.ApplicationCommandData().Options
	channel := options[0].Options[0].ChannelValue(s)
	capacity := options[0].Options[1].IntValue()

	if capacity < 0 {
		return CommandResponse{
			title:       "🧀 Warning",
			description: "User limit cannot be negative!",
			colorType:   color.Warning,
		}
	}

	_, err := lc.lobbyRepository.GetLobby(channel.ID)
	if err != nil {
		return CommandResponse{
			title:       "🧀 Warning",
			description: "\"" + channel.Name + "\" is not a lobby!",
			colorType:   color.Warning,
		}
	}

	lobby := model.Lobby{
		Id:         channel.ID,
		CategoryID: channel.ParentID,
		Capacity: sql.NullInt32{
			Valid: true,
			Int32: int32(capacity),
		},
	}

	err = lc.lobbyRepository.UpsertLobby(&lobby)
	if err != nil {
		log.Println("unable to update lobby! %w", err)

		return CommandResponse{
			title:       "🚨 Error",
			description: "Unable to update lobby!",
			colorType:   color.Failure,
		}
	}

	if capacity == 0 {
		return CommandResponse{
			title:       "✅ OK",
			description: "Capacity successfully reset for \"" + channel.Name + "\".",
			colorType:   color.Success,
		}
	}

	return CommandResponse{
		title:       "✅ OK",
		description: "Capacity " + strconv.FormatInt(capacity, 10) + " successfully set for \"" + channel.Name + "\".",
		colorType:   color.Success,
	}
}

func (lc *LobbyCommands) handleCommandName(s *discordgo.Session, i *discordgo.InteractionCreate) CommandResponse {
	options := i.ApplicationCommandData().Options
	channel := options[0].Options[0].ChannelValue(s)
	name := options[0].Options[1].StringValue()

	_, err := lc.lobbyRepository.GetLobby(channel.ID)
	if err != nil {
		return CommandResponse{
			title:       "🧀 Warning",
			description: "\"" + channel.Name + "\" is not a lobby!",
			colorType:   color.Warning,
		}
	}

	lobby := model.Lobby{
		Id:         channel.ID,
		CategoryID: channel.ParentID,
		Template: sql.NullString{
			Valid:  true,
			String: name,
		},
	}

	err = lc.lobbyRepository.UpsertLobby(&lobby)
	if err != nil {
		log.Println("unable to update lobby! %w", err)

		return CommandResponse{
			title:       "🚨 Error",
			description: "Unable to update lobby!",
			colorType:   color.Failure,
		}
	}

	return CommandResponse{
		title:       "✅ OK",
		description: "Name " + name + " successfully set for " + channel.Name + ".",
		colorType:   color.Success,
	}
}

func (lc *LobbyCommands) handleCommandList(s *discordgo.Session, i *discordgo.InteractionCreate) CommandResponse {
	lobbies, err := lc.lobbyRepository.GetLobbies()
	if err != nil {
		log.Println("unable to get lobbies! %w", err)

		return CommandResponse{
			title:       "🚨 Error",
			description: "Unable to get lobbies!",
			colorType:   color.Failure,
		}
	}

	var registeredChannels []string
	for i, lobby := range lobbies {
		channel, err := s.Channel(lobby.Id)
		if err != nil {
			log.Println("unable to get channels! %w", err)

			return CommandResponse{
				title:       "🚨 Error",
				description: "Unable to get channels!",
				colorType:   color.Failure,
			}
		}
		if channel.ID == lobby.Id {

			var template string
			if lobby.Template.Valid {
				template = lobby.Template.String
			} else {
				template = "Кімната %username%"
			}

			var capacity string
			if lobby.Capacity.Valid {
				capacity = strconv.FormatInt(int64(lobby.Capacity.Int32), 10)
			} else if lobby.Capacity.Int32 == 0 {
				capacity = "default"
			} else {
				capacity = "default"
			}

			finalString := fmt.Sprintf("%d. Name: %s, Channel template: %s, Capacity: %s", i+1, channel.Name, template, capacity)
			registeredChannels = append(registeredChannels, finalString)
		}
	}

	if len(registeredChannels) == 0 {
		return CommandResponse{
			title:       "🧀 Warning",
			description: "There are no active lobbies.",
			colorType:   color.Warning,
		}
	}

	return CommandResponse{
		title:       "✅ OK",
		description: "Active Lobbies:\n" + strings.Join(registeredChannels, "\n"),
		colorType:   color.Success,
	}
}

func (lc *LobbyCommands) handleCommandRemove(s *discordgo.Session, i *discordgo.InteractionCreate) CommandResponse {
	options := i.ApplicationCommandData().Options
	channel := options[0].Options[0].ChannelValue(s)

	affectedRows, err := lc.lobbyRepository.DeleteLobby(channel.ID)
	if err != nil {
		log.Println("unable to delete the lobby! %w", err)

		return CommandResponse{
			title:       "🚨 Error",
			description: "Unable unregister \"" + channel.Name + "\" lobby.",
			colorType:   color.Failure,
		}
	}

	if affectedRows == 0 {
		return CommandResponse{
			title:       "🧀 Warning",
			description: "\"" + channel.Name + "\" is not a lobby!",
			colorType:   color.Warning,
		}
	}

	return CommandResponse{
		title:       "✅ OK",
		description: "Lobby \"" + channel.Name + "\" successfully unregistered.",
		colorType:   color.Success,
	}
}
