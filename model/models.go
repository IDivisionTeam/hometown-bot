package model

import (
    "database/sql"
    "github.com/bwmarrin/discordgo"
    "hometown-bot/util/discord"
)

type Lobby struct {
    Id         string
    CategoryID string
    GuildID    string
    Template   sql.NullString
    Capacity   sql.NullInt32
}

type Channel struct {
    Id       string
    ParentID string
}

type CommandResponse struct {
    Title       string
    Description string
    ColorType   discord.Color
}

func CommandSuccess(description string) CommandResponse {
    return CommandResponse{
        Title:       "✅ OK",
        Description: description,
        ColorType:   discord.Success,
    }
}

func CommandWarning(description string) CommandResponse {
    return CommandResponse{
        Title:       "🧀 Warning",
        Description: description,
        ColorType:   discord.Warning,
    }
}

func CommandError(description string) CommandResponse {
    return CommandResponse{
        Title:       "🚨 Error",
        Description: description,
        ColorType:   discord.Failure,
    }
}

func (c CommandResponse) ToEmbededMessage() *discordgo.MessageEmbed {
    return &discordgo.MessageEmbed{
        Title:       c.Title,
        Description: c.Description,
        Color:       discord.GetColorFrom(c.ColorType),
    }
}
