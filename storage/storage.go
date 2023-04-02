package storage

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var lobbyTable = `
CREATE TABLE IF NOT EXISTS lobbies(
	id TEXT PRIMARY KEY,
	category_id TEXT, 		/* immutable */
	guild_id TEXT, 		/* immutable */
	template TEXT,			/* mutable, default NULL */
	capacity INTEGER		/* mutable, default NULL */
);`

var channelTable = `
CREATE TABLE IF NOT EXISTS channels(
	id TEXT PRIMARY KEY,
	parent_id TEXT NOT NULL	/* immutable */
);`

func Load() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./storage.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(lobbyTable)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(channelTable)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	log.Println("Storage loaded!")
	return db, nil
}
