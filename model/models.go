package model

import (
	"database/sql"
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
