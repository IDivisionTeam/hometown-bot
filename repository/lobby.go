package repository

import (
	"database/sql"
	"fmt"
	"hometown-bot/model"
)

type LobbyRepository struct {
	db *sql.DB
}

func NewLobbyRepository(db *sql.DB) *LobbyRepository {
	return &LobbyRepository{db: db}
}

const SelectLobbyById = `
SELECT id, template, capacity, category_id, guild_id
FROM lobbies
WHERE (id = ? AND guild_id = ?)
`

func (cr *LobbyRepository) GetLobby(id string, guildId string) (model.Lobby, error) {
	var lobby model.Lobby

	err := cr.db.QueryRow(
		SelectLobbyById,
		&lobby.Id,
		&lobby.Template,
		&lobby.Capacity,
		&lobby.CategoryID,
		&lobby.GuildID,
	).Scan()
	if err != nil {
		return model.Lobby{}, fmt.Errorf("unable to get lobby for id %s: %w", id, err)
	}

	return lobby, nil
}

const SelectLobbies = `
SELECT id, template, capacity, category_id, guild_id
FROM lobbies
WHERE guild_id = ?
`

func (cr *LobbyRepository) GetLobbies(guildId string) ([]model.Lobby, error) {
	rows, err := cr.db.Query(SelectLobbies, guildId)
	if err != nil {
		return []model.Lobby{}, fmt.Errorf("unable to get lobbies: %w", err)
	}

	defer rows.Close()

	var lobbies []model.Lobby
	for rows.Next() {
		var lobby model.Lobby

		err = rows.Scan(
			&lobby.Id,
			&lobby.Template,
			&lobby.Capacity,
			&lobby.CategoryID,
			&lobby.GuildID,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to get lobbies: %w", err)
		}

		lobbies = append(lobbies, lobby)
	}

	return lobbies, nil
}

const ReplaceLobby = `
INSERT INTO lobbies (id, template, capacity, category_id, guild_id)
VALUES(?, ?, ?, ?, ?)
ON CONFLICT(id) 
WHERE (id = ? AND guild_id = ?)
DO NOTHING
`

func (cr *LobbyRepository) SetLobby(lobby *model.Lobby) (int64, error) {
	result, err := cr.db.Exec(
		ReplaceLobby,
		lobby.Id,
		lobby.Template,
		lobby.Capacity,
		lobby.CategoryID,
		lobby.GuildID,
		lobby.Id,
		lobby.GuildID,
	)
	if err != nil {
		return 0, err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affectedRows, nil
}

const UpsertLobby = `
INSERT INTO lobbies (id, template, capacity)
VALUES(?, ?, ?)
ON CONFLICT(id) 
DO UPDATE
SET
	template = coalesce(EXCLUDED.template, template),
	capacity = coalesce(EXCLUDED.capacity, capacity)
`

func (cr *LobbyRepository) UpsertLobby(lobby *model.Lobby) error {
	_, err := cr.db.Exec(
		UpsertLobby,
		lobby.Id,
		lobby.Template,
		lobby.Capacity,
	)
	if err != nil {
		return err
	}

	return nil
}

const DeleteLobby = `
DELETE FROM lobbies
WHERE (id = ? AND guild_id = ?)
`

func (cr *LobbyRepository) DeleteLobby(id string, guildId string) (int64, error) {
	result, err := cr.db.Exec(DeleteLobby, id, guildId)
	if err != nil {
		return 0, fmt.Errorf("unable to delete lobby for id %s: %w", id, err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affectedRows, nil
}
