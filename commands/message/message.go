package message

import (
    "fmt"
    "github.com/bwmarrin/discordgo"
    "hometown-bot/log"
    "hometown-bot/model"
)

const (
    message       string = "message" // Command group
    commandAll    string = "all"     // Subcommand message all
    optionChannel string = "channel" // Option for commandAll
    optionMessage string = "message" // Option for commandAll
)

var (
    dmPermission             bool  = false                            // Does not allow using Bot in DMs
    defaultMemberPermissions int64 = discordgo.PermissionManageServer // Caller permission to use commands
    Commands                       = getMessageCommandGroup()         // Command group
)

type Command struct {
    commandHandlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) // Command interaction
}

func New() *Command {
    return &Command{
        commandHandlers: createCommandHandlers(),
    }
}

func getMessageCommandGroup() []*discordgo.ApplicationCommand {
    return []*discordgo.ApplicationCommand{
        {
            Name:                     message,
            Description:              "Message' commands group.",
            DefaultMemberPermissions: &defaultMemberPermissions,
            DMPermission:             &dmPermission,
            Options: []*discordgo.ApplicationCommandOption{
                getRegisterCommand(),
            },
        },
    }
}

func getRegisterCommand() *discordgo.ApplicationCommandOption {
    return &discordgo.ApplicationCommandOption{
        Name:        commandAll,
        Type:        discordgo.ApplicationCommandOptionSubCommand,
        Description: "Message to a channel.",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionChannel,
                Name:        optionChannel,
                Description: "A channel to be messaged.",
                ChannelTypes: []discordgo.ChannelType{
                    discordgo.ChannelTypeGuildText,
                },
                Required: true,
            },
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        optionMessage,
                Description: "A message to be sent.",
                Required:    true,
            },
        },
    }
}

func (mc *Command) HandleSlashCommands(discord *discordgo.Session, interaction *discordgo.InteractionCreate) {
    if handler, ok := mc.commandHandlers[interaction.ApplicationCommandData().Name]; ok {
        handler(discord, interaction)
    }
}

func createCommandHandlers() map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    return map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
        message: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
            log.Info().Printf("trigger %s command interaction", message)

            slashCommand := i.ApplicationCommandData().Options[0].Name
            channel, commandResponse := &discordgo.Channel{}, "Oops, something went wrong.\nHol' up, you aren't supposed to see this message."

            if slashCommand == commandAll {
                channel, commandResponse = handleMessageAll(s, i)
            }

            log.Info().Printf("message: sending interaction response for %s", slashCommand)
            if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Embeds: []*discordgo.MessageEmbed{
                        model.CommandSuccess(
                            fmt.Sprintf(
                                "Message \"%s\" successfully sent to channel \"%s\".",
                                commandResponse,
                                channel.Name,
                            ),
                        ).ToEmbededMessage(),
                    },
                    Flags: discordgo.MessageFlagsEphemeral,
                },
            }); err != nil {
                log.Error().Printf("message: interaction response to %s[%s]: %v", channel.Name, channel.ID, err)
            }

            log.Info().Printf("message: sending message for %s", slashCommand)
            if _, err := s.ChannelMessageSend(
                channel.ID,
                commandResponse,
            ); err != nil {
                log.Error().Printf(
                    "message: send message[%s] to %s[%s]: %v",
                    commandResponse,
                    channel.Name,
                    channel.ID,
                    err,
                )
            }
        },
    }
}

func handleMessageAll(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.Channel, string) {
    options := i.ApplicationCommandData().Options
    channel := options[0].Options[0].ChannelValue(s)
    msg := options[0].Options[1].StringValue()

    return channel, msg
}
