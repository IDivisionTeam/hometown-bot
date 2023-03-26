package lobby

import (
	"hometown-bot/utils/color"
	"log"
	"strconv"

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
	commandHandlers                 = getCommandHandlers()             // Command interaction
)

// FIXME: TEMP CACHE, replace with real DB
var (
	tempChannels  []string
	tempLobbies   []string = []string{"1088547840734810216"} // General channel
	tempLobbyName string   = "Trio"
	tempUserLimit int      = 2
)

func HandleSlashCommands(discord *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if handler, ok := commandHandlers[interaction.ApplicationCommandData().Name]; ok {
		handler(discord, interaction)
	}
}

// FIXME: split into small functions
func HandleVoiceUpdates(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	if vsu == nil {
		return
	}

	previousState := vsu.BeforeUpdate
	if previousState != nil && previousState.ChannelID == vsu.ChannelID {
		return
	}

	for _, l := range tempLobbies {
		if l == vsu.ChannelID {
			log.Println("Searching for lobby...")
			channel, err := s.Channel(l)
			if err != nil {
				log.Printf("Unable to find a lobby! \n%v\n", err)
				continue
			}

			userName := vsu.Member.User.Username
			nickname := vsu.Member.Nick
			name := ""
			 if len(nickname) == 0 {
				name = userName
			}  else {
				name = nickname
			}

			data := discordgo.GuildChannelCreateData{
				// TODO: replace temp [`name`, `capacity`] with real data form DB
				Name:      tempLobbyName + " " + name,
				Type:      discordgo.ChannelTypeGuildVoice,
				ParentID:  channel.ParentID,
				UserLimit: tempUserLimit,
			}

			log.Println("Found! Creating a new channel...")
			newChannel, err := s.GuildChannelCreateComplex(GuildID, data)
			if err != nil {
				log.Printf("Unable to create a new channel! \n%v\n", err)
				continue
			}

			// TODO: save to channel DB
			tempChannels = append(tempChannels, newChannel.ID)

			log.Println("Created! Moving a user to the new channel...")
			merr := s.GuildMemberMove(GuildID, vsu.Member.User.ID, &newChannel.ID)
			if merr != nil {
				log.Printf("Unable to move a user to the new channel! \n%v\n", merr)
				continue
			}
		}
	}

	if previousState != nil {
		for _, tc := range tempChannels {
			if tc == previousState.ChannelID {
				channel, err := s.Channel(tc)
				if err != nil {
					log.Printf("Unable to find a channel! \n%v\n", err)
					continue
				}

				if len(channel.Members) == 0 {
					log.Println("Channel " + channel.Name + " is empty, deleting...")

					st, err := s.ChannelDelete(tc)
					if err != nil {
						log.Printf("Unable to to delete the channel! \n%v\n", err)
						continue
					}

					// TODO: remove from channel DB
					for i, v := range tempChannels {
						if v == st.ID {
							tempChannels = append(tempChannels[:i], tempChannels[i+1:]...)
							break
						}
					}
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

func getCommandHandlers() map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
				commandResponse = handleCommandRegister(s, i)
			case commandCapacity:
				commandResponse = handleCommandCapacity(s, i)
			case commandName:
				commandResponse = handleCommandName(s, i)
			case commandList:
				commandResponse = handleCommandList(i)
			case commandRemove:
				commandResponse = handleCommandRemove(s, i)
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

func handleCommandRegister(s *discordgo.Session, i *discordgo.InteractionCreate) CommandResponse {
	// TODO: write channel in Lobby DB
	options := i.ApplicationCommandData().Options
	channel := options[0].Options[0].ChannelValue(s)

	return CommandResponse{
		title:       "✅ OK",
		description: "Lobby " + channel.Name + " successfully registered.",
		colorType:   color.Success,
	}
}

func handleCommandCapacity(s *discordgo.Session, i *discordgo.InteractionCreate) CommandResponse {
	// TODO: write channel in Channel DB
	options := i.ApplicationCommandData().Options
	channel := options[0].Options[0].ChannelValue(s)
	capacity := options[0].Options[1].IntValue()

	return CommandResponse{
		title:       "✅ OK",
		description: "Capacity " + strconv.FormatInt(capacity, 10) + " successfully set for " + channel.Name + ".",
		colorType:   color.Success,
	}
}

func handleCommandName(s *discordgo.Session, i *discordgo.InteractionCreate) CommandResponse {
	// TODO: write channel in Channel DB
	options := i.ApplicationCommandData().Options
	channel := options[0].Options[0].ChannelValue(s)
	name := options[0].Options[1].StringValue()

	return CommandResponse{
		title:       "✅ OK",
		description: "Name" + name + " successfully set for " + channel.Name + ".",
		colorType:   color.Success,
	}
}

func handleCommandList(i *discordgo.InteractionCreate) CommandResponse {
	// TODO: read Lobby DB put in description

	return CommandResponse{
		title:       "✅ OK",
		description: "Active Lobbies:\n",
		colorType:   color.Success,
	}
}

func handleCommandRemove(s *discordgo.Session, i *discordgo.InteractionCreate) CommandResponse {
	// TODO: remove channel from Lobby DB
	options := i.ApplicationCommandData().Options
	channel := options[0].Options[0].ChannelValue(s)

	return CommandResponse{
		title:       "✅ OK",
		description: "Lobby" + channel.Name + " successfully unregistered.",
		colorType:   color.Success,
	}
}
