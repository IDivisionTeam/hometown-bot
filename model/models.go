package model

import "database/sql"

type Lobby struct {
	Id         string
	CategoryID string
	Template   sql.NullString
	Capacity   sql.NullInt32
}

type Channel struct {
	Id       string
	ParentID string
}
