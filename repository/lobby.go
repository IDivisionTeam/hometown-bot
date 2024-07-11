package repository

import (
    "database/sql"
    "fmt"
    "hometown-bot/log"
    "hometown-bot/model"
)

type LobbyRepository struct {
    db *sql.DB
}

func NewLobby(db *sql.DB) *LobbyRepository {
    return &LobbyRepository{db: db}
}

const SelectLobbyById = `
SELECT id, template, capacity, category_id, guild_id
FROM lobbies
WHERE (id = ? AND guild_id = ?)
`

func (cr *LobbyRepository) GetLobby(id string, guildId string) (model.Lobby, error) {
    log.Debug().Printf("repo: get lobby[%s] for guild[%s]", id, guildId)

    var lobby model.Lobby
    if err := cr.db.QueryRow(
        SelectLobbyById,
        id,
        guildId,
    ).Scan(
        &lobby.Id,
        &lobby.Template,
        &lobby.Capacity,
        &lobby.CategoryID,
        &lobby.GuildID,
    ); err != nil {
        return model.Lobby{}, fmt.Errorf(
            "repo: unable to get lobby[%s] for guild[%s]: %w",
            id,
            guildId,
            err,
        )
    }

    return lobby, nil
}

const SelectLobbies = `
SELECT id, template, capacity, category_id, guild_id
FROM lobbies
WHERE guild_id = ?
`

func (cr *LobbyRepository) GetLobbies(guildId string) ([]model.Lobby, error) {
    log.Debug().Printf("repo: get lobbies for guild[%s]", guildId)

    rows, err := cr.db.Query(SelectLobbies, guildId)
    if err != nil {
        return []model.Lobby{}, fmt.Errorf(
            "repo: unable to lobbies for guild[%s]: %w",
            guildId,
            err,
        )
    }

    var lobbies []model.Lobby
    for rows.Next() {
        var lobby model.Lobby

        if err := rows.Scan(
            &lobby.Id,
            &lobby.Template,
            &lobby.Capacity,
            &lobby.CategoryID,
            &lobby.GuildID,
        ); err != nil {
            return nil, fmt.Errorf(
                "repo: unable to lobbies for guild[%s]: %w",
                guildId,
                err,
            )
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
    log.Debug().Printf("repo: set lobby[%s]", lobby.Id)

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
        return 0, fmt.Errorf("repo: unable to set lobby[%s]: %w", lobby.Id, err)
    }

    affectedRows, err := result.RowsAffected()
    if err != nil {
        return 0, fmt.Errorf("repo: unable to set lobby[%s]: %w", lobby.Id, err)
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
    log.Debug().Printf("repo: upsert lobby[%s]", lobby.Id)

    if _, err := cr.db.Exec(
        UpsertLobby,
        lobby.Id,
        lobby.Template,
        lobby.Capacity,
    ); err != nil {
        return fmt.Errorf("repo: unable to upsert lobby[%s]: %w", lobby.Id, err)
    }

    return nil
}

const DeleteLobby = `
DELETE FROM lobbies
WHERE (id = ? AND guild_id = ?)
`

func (cr *LobbyRepository) DeleteLobby(id string, guildId string) (int64, error) {
    log.Debug().Printf("repo: delete lobby[%s] for guild[%s]", id, guildId)

    result, err := cr.db.Exec(DeleteLobby, id, guildId)
    if err != nil {
        return 0, fmt.Errorf("repo: unable to delete lobby[%s] for guild[%s]: %w", id, guildId, err)
    }

    affectedRows, err := result.RowsAffected()
    if err != nil {
        return 0, fmt.Errorf("repo: unable to delete lobby[%s] for guild[%s]: %w", id, guildId, err)
    }

    return affectedRows, nil
}
