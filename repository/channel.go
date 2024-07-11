package repository

import (
    "database/sql"
    "fmt"
    "hometown-bot/log"
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

    log.Debug().Printf("repo: get channel %s", id)
    if err := cr.db.QueryRow(SelectChannelById, &channel.Id, &channel.ParentID).Scan(); err != nil {
        return model.Channel{}, fmt.Errorf("repo: unable to get channel[%s]: %w", id, err)
    }

    return channel, nil
}

const SelectChannels = `
SELECT id, parent_id
FROM channels
`

func (cr *ChannelRepository) GetChannels() ([]model.Channel, error) {
    log.Debug().Println("repo: get channels")

    rows, err := cr.db.Query(SelectChannels)
    if err != nil {
        return []model.Channel{}, fmt.Errorf("repo: unable to get channels: %w", err)
    }

    var channels []model.Channel
    for rows.Next() {
        var channel model.Channel

        if err := rows.Scan(&channel.Id, &channel.ParentID); err != nil {
            return nil, fmt.Errorf("repo: unable to get channels: %w", err)
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
    log.Debug().Printf("repo: set channel[%s]", channel.Id)

    if _, err := cr.db.Exec(ReplaceChannel, channel.Id, channel.ParentID); err != nil {
        return fmt.Errorf("repo: unable to set channel[%s]: %w", channel.Id, err)
    }

    return nil
}

const DeleteChannel = `
DELETE FROM channels
WHERE id = ?
`

func (cr *ChannelRepository) DeleteChannel(id string) error {
    log.Debug().Printf("repo: delete channel[%s]", id)

    if _, err := cr.db.Exec(DeleteChannel, id); err != nil {
        return fmt.Errorf("repo: unable to delete channel[%s]: %w", id, err)
    }

    return nil
}
