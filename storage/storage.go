package storage

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var lobbyTable = `
CREATE TABLE IF NOT EXISTS lobbies(
	id TEXT PRIMARY KEY,
	template TEXT,			/* mutable, default NULL */
	capacity INTEGER,		/* mutable, default NULL */
);`

var channelTable = `
CREATE TABLE IF NOT EXISTS channels(
	id TEXT PRIMARY KEY,
	parent TEXT NOT NULL,	/* immutable */
);`

func Load() error {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return err
	}

	defer db.Close()

	_, err = db.Exec(lobbyTable)
	if err != nil {
		return err
	}

	_, err = db.Exec(channelTable)
	if err != nil {
		return err
	}

	return nil
}
