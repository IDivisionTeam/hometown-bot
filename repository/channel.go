package repository

import (
    "database/sql"
    "fmt"
    "hometown-bot/model"
)

type ChannelRepository struct {
    db *sql.DB
}

func NewChannel(db *sql.DB) *ChannelRepository {
    return &ChannelRepository{db: db}
}

const SelectChannelById = `
SELECT id, parent_id
FROM channels
WHERE id = ?
`

func (cr *ChannelRepository) GetChannel(id string) (model.Channel, error) {
    var channel model.Channel

    err := cr.db.QueryRow(SelectChannelById, &channel.Id, &channel.ParentID).Scan()
    if err != nil {
        return model.Channel{}, fmt.Errorf("unable to get channel for id %s: %w", id, err)
    }

    return channel, nil
}

const SelectChannels = `
SELECT id, parent_id
FROM channels
`

func (cr *ChannelRepository) GetChannels() ([]model.Channel, error) {
    rows, err := cr.db.Query(SelectChannels)
    if err != nil {
        return []model.Channel{}, fmt.Errorf("unable to get channels: %w", err)
    }

    defer rows.Close()

    var channels []model.Channel
    for rows.Next() {
        var channel model.Channel

        err = rows.Scan(&channel.Id, &channel.ParentID)
        if err != nil {
            return nil, fmt.Errorf("unable to get channels: %w", err)
        }

        channels = append(channels, channel)
    }

    return channels, nil
}

const ReplaceChannel = `
REPLACE INTO channels (id, parent_id)
VALUES(?, ?)
`

func (cr *ChannelRepository) SetChannel(channel *model.Channel) error {
    _, err := cr.db.Exec(ReplaceChannel, channel.Id, channel.ParentID)
    if err != nil {
        return err
    }

    return nil
}

const DeleteChannel = `
DELETE FROM channels
WHERE id = ?
`

func (cr *ChannelRepository) DeleteChannel(id string) error {
    _, err := cr.db.Exec(DeleteChannel, id)
    if err != nil {
        return fmt.Errorf("unable to delete channel for id %s: %w", id, err)
    }

    return nil
}
