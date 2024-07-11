package commands

import (
    "fmt"
    "github.com/bwmarrin/discordgo"
    "hometown-bot/model"
    "hometown-bot/repository"
)

func HasLobby(
    repository repository.LobbyRepository,
    channel *discordgo.Channel,
    guildId string,
) (model.CommandResponse, error) {
    if _, err := repository.GetLobby(channel.ID, guildId); err != nil {
        return model.CommandWarning(
                fmt.Sprintf("\"%s\" is not a lobby!", channel.Name),
            ),
            fmt.Errorf("db: %s[%s] is not a lobby: %w", channel.Name, channel.ID, err)
    }

    return model.CommandResponse{}, nil
}
