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
SELECT id, template, capacity, category_id
FROM lobbies
WHERE id = ?
`

func (cr *LobbyRepository) GetLobby(id string) (model.Lobby, error) {
	var lobby model.Lobby

	err := cr.db.QueryRow(SelectLobbyById, &lobby.Id, &lobby.Template, &lobby.Capacity, &lobby.CategoryID).Scan()
	if err != nil {
		return model.Lobby{}, fmt.Errorf("unable to get lobby for id %s: %w", id, err)
	}

	return lobby, nil
}

const SelectLobbies = `
SELECT id, template, capacity, category_id
FROM lobbies
`

func (cr *LobbyRepository) GetLobbies() ([]model.Lobby, error) {
	rows, err := cr.db.Query(SelectLobbies)
	if err != nil {
		return []model.Lobby{}, fmt.Errorf("unable to get lobbies: %w", err)
	}

	defer rows.Close()

	var lobbies []model.Lobby
	for rows.Next() {
		var lobby model.Lobby

		err = rows.Scan(&lobby.Id, &lobby.Template, &lobby.Capacity, &lobby.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("unable to get lobbies: %w", err)
		}

		lobbies = append(lobbies, lobby)
	}

	return lobbies, nil
}

const ReplaceLobby = `
INSERT INTO lobbies (id, template, capacity, category_id)
VALUES(?, ?, ?, ?)
ON CONFLICT(id) 
WHERE id = ?
DO NOTHING
`

func (cr *LobbyRepository) SetLobby(lobby *model.Lobby) (int64, error) {
	result, err := cr.db.Exec(ReplaceLobby, lobby.Id, lobby.Template, lobby.Capacity, lobby.CategoryID, lobby.Id)
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
	_, err := cr.db.Exec(UpsertLobby, lobby.Id, lobby.Template, lobby.Capacity)
	if err != nil {
		return err
	}

	return nil
}

const DeleteLobby = `
DELETE FROM lobbies
WHERE id = ?
`

func (cr *LobbyRepository) DeleteLobby(id string) (int64, error) {
	result, err := cr.db.Exec(DeleteLobby, id)
	if err != nil {
		return 0, fmt.Errorf("unable to delete lobby for id %s: %w", id, err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affectedRows, nil
}
