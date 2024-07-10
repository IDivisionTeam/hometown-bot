package repository

import (
    "database/sql"
    "fmt"
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
    var output int
    row := cmr.db.QueryRow(CountChannelMembers, guildId, channelId)

    err := row.Scan(&output)
    if err != nil {
        return 0, fmt.Errorf("get channel members for %s: %w", channelId, err)
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
    _, err := cmr.db.Exec(InsertChannelMembers, guildId, userId, channelId, guildId)
    if err != nil {
        return fmt.Errorf("unable to insert user %s for channel %s: %w", userId, channelId, err)
    }

    return nil
}

const DeleteChannelMember = `
DELETE FROM channel_members
WHERE (guild_id = ? AND user_id = ?)
`

func (cmr *ChannelMembersRepository) DeleteChannelMember(guildId string, userId string) error {
    _, err := cmr.db.Exec(DeleteChannelMember, guildId, userId)
    if err != nil {
        return fmt.Errorf("unable to delete user %s for temp channel: %w", userId, err)
    }
    return nil
}

const DeleteChannelMembers = `
DELETE FROM channel_members
WHERE (guild_id = ? AND channel_id = ?)
`

func (cmr *ChannelMembersRepository) DeleteChannelMembers(guildId string, channelId string) error {
    _, err := cmr.db.Exec(DeleteChannelMember, guildId, channelId)
    if err != nil {
        return fmt.Errorf("unable to delete temp channel %s: %w", channelId, err)
    }
    return nil
}
