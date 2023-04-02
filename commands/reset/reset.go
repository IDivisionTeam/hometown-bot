package reset

import (
	"database/sql"
	"hometown-bot/model"
	"hometown-bot/repository"
	"hometown-bot/utils/color"
	"log"

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

type ResetCommands struct {
	channelRepository repository.ChannelRepository
	lobbyRepository   repository.LobbyRepository
	commandHandlers   map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) // Command interaction
}

func New(channelRepository repository.ChannelRepository, lobbyRepository repository.LobbyRepository) *ResetCommands {
	commands := ResetCommands{
		channelRepository: channelRepository,
		lobbyRepository:   lobbyRepository,
	}

	commands.commandHandlers = commands.getCommandHandlers()
	return &commands
}

func (rc *ResetCommands) HandleSlashCommands(discord *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if handler, ok := rc.commandHandlers[interaction.ApplicationCommandData().Name]; ok {
		handler(discord, interaction)
	}
}

/* Commands */

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

/* Interactions */

func (rc *ResetCommands) getCommandHandlers() map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		reset: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// slashCommand := i.ApplicationCommandData().Options[0].Name as for now I do not need to handle groups
			subcommandCommand := i.ApplicationCommandData().Options[0].Options[0].Name
			commandResponse := model.CommandResponse{
				Title: "🚨 Error",
				Description: "Oops, something went wrong.\n" +
					"Hol' up, you aren't supposed to see this message.",
				ColorType: color.Failure,
			}

			switch subcommandCommand {
			case commandCapacity:
				commandResponse = rc.handleCommandCapacity(s, i)
			case commandName:
				commandResponse = rc.handleCommandName(s, i)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:       commandResponse.Title,
							Description: commandResponse.Description,
							Color:       color.GetColorFrom(commandResponse.ColorType),
						},
					},
					Flags: discordgo.MessageFlagsEphemeral,
				},
			})
		},
	}
}

func (rc *ResetCommands) handleCommandCapacity(s *discordgo.Session, i *discordgo.InteractionCreate) model.CommandResponse {
	options := i.ApplicationCommandData().Options
	channel := options[0].Options[0].Options[0].ChannelValue(s)

	_, err := rc.lobbyRepository.GetLobby(channel.ID, i.GuildID)
	if err == nil {
		return model.CommandResponse{
			Title:       "🧀 Warning",
			Description: "\"" + channel.Name + "\" is not a lobby!",
			ColorType:   color.Warning,
		}
	}

	lobby := model.Lobby{
		Id:         channel.ID,
		CategoryID: channel.ParentID,
		Capacity: sql.NullInt32{
			Valid: true,
			Int32: 0,
		},
	}

	err = rc.lobbyRepository.UpsertLobby(&lobby)
	if err != nil {
		log.Println("unable to update lobby! %w", err)

		return model.CommandResponse{
			Title:       "🚨 Error",
			Description: "Unable to update lobby!",
			ColorType:   color.Failure,
		}
	}

	return model.CommandResponse{
		Title:       "✅ OK",
		Description: "Capacity successfully reset for \"" + channel.Name + "\".",
		ColorType:   color.Success,
	}
}

func (rc *ResetCommands) handleCommandName(s *discordgo.Session, i *discordgo.InteractionCreate) model.CommandResponse {
	options := i.ApplicationCommandData().Options
	channel := options[0].Options[0].Options[0].ChannelValue(s)

	_, err := rc.lobbyRepository.GetLobby(channel.ID, i.GuildID)
	if err == nil {
		return model.CommandResponse{
			Title:       "🧀 Warning",
			Description: "\"" + channel.Name + "\" is not a lobby!",
			ColorType:   color.Warning,
		}
	}

	template := sql.NullString{
		Valid:  true,
		String: "",
	}

	lobby := model.Lobby{
		Id:         channel.ID,
		CategoryID: channel.ParentID,
		Template:   template,
	}

	err = rc.lobbyRepository.UpsertLobby(&lobby)
	if err != nil {
		log.Println("unable to update lobby! %w", err)

		return model.CommandResponse{
			Title:       "🚨 Error",
			Description: "Unable to update lobby!",
			ColorType:   color.Failure,
		}
	}

	return model.CommandResponse{
		Title:       "✅ OK",
		Description: "Name successfully reset to default for " + channel.Name + ".",
		ColorType:   color.Success,
	}
}
