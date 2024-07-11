package repository

import (
    "database/sql"
    "fmt"
    "hometown-bot/log"
)

type ChannelMembersRepository struct {
    db *sql.DB
}

func NewChannelMembers(db *sql.DB) *ChannelMembersRepository {
    return &ChannelMembersRepository{db: db}
}

const CountChannelMembers = `
SELECT COUNT(*) FROM channel_members
WHERE (guild_id = ? AND channel_id = ?)
`

func (cmr *ChannelMembersRepository) GetChannelMembersCount(guildId string, channelId string) (int, error) {
    log.Debug().Printf("repo: get channel[%s] member count for guild[%s]", channelId, guildId)

    var output int
    row := cmr.db.QueryRow(CountChannelMembers, guildId, channelId)

    if err := row.Scan(&output); err != nil {
        return 0, fmt.Errorf(
            "repo: unable to get channel[%s] member count for guild[%s]: %w",
            channelId,
            guildId,
            err,
        )
    }

    return output, nil
}

const InsertChannelMembers = `
INSERT INTO channel_members (guild_id, user_id, channel_id)
VALUES(?, ?, ?)
ON CONFLICT(user_id) 
DO UPDATE
SET
	channel_id = EXCLUDED.channel_id
WHERE guild_id = ?
`

func (cmr *ChannelMembersRepository) SetChannelMember(guildId string, userId string, channelId string) error {
    log.Debug().Printf("repo: set channel[%s] member[%s] for guild[%s]", channelId, userId, guildId)

    if _, err := cmr.db.Exec(InsertChannelMembers, guildId, userId, channelId, guildId); err != nil {
        return fmt.Errorf(
            "repo: unable to set channel[%s] member[%s] for guild[%s]: %w",
            channelId,
            userId,
            guildId,
            err,
        )
    }

    return nil
}

const DeleteChannelMember = `
DELETE FROM channel_members
WHERE (guild_id = ? AND user_id = ? AND channel_id = ?)
`

func (cmr *ChannelMembersRepository) DeleteChannelMember(guildId string, userId string, channelId string) error {
    log.Debug().Printf("repo: delete channel[%s] member[%s] for guild[%s]", channelId, userId, guildId)

    if _, err := cmr.db.Exec(DeleteChannelMember, guildId, userId, channelId); err != nil {
        return fmt.Errorf(
            "repo: unable to delete channel[%s] member[%s] for guild[%s]: %w",
            channelId,
            userId,
            guildId,
            err,
        )
    }

    return nil
}

const DeleteChannelMembers = `
DELETE FROM channel_members
WHERE (guild_id = ? AND channel_id = ?)
`

func (cmr *ChannelMembersRepository) DeleteChannelMembers(guildId string, channelId string) error {
    log.Debug().Printf("repo: delete channel[%s] members for guild[%s]", channelId, guildId)

    if _, err := cmr.db.Exec(DeleteChannelMembers, guildId, channelId); err != nil {
        return fmt.Errorf(
            "repo: unable to delete channel[%s] members for guild[%s]: %w",
            channelId,
            guildId,
            err,
        )
    }

    return nil
}
