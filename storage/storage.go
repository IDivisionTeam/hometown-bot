package storage

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

/*
+------------------------------+
|go        | sqlite3           |
|----------|-------------------|
|nil       | null              |
|int       | integer           |
|int64     | integer           |
|float64   | float             |
|bool      | integer           |
|[]byte    | blob              |
|string    | text              |
|time.Time | timestamp/datetime|
+------------------------------+
*/

var lobbyTable = `
CREATE TABLE IF NOT EXISTS lobbies(
	id TEXT PRIMARY KEY,
	template TEXT,			/* mutable, default NULL */
	capacity INTEGER,		/* mutable, default NULL */
)`

var channelTable = `
CREATE TABLE IF NOT EXISTS channels(
	id TEXT PRIMARY KEY,
	parent TEXT NOT NULL,	/* immutable */
)`

func Load() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	_, err = db.Exec(lobbyTable)
	if err != nil {
		log.Panicf("Unable to create lobby table!\n%v\n", err)
	}

	_, err = db.Exec(channelTable)
	if err != nil {
		log.Panicf("Unable to create channel table!\n%v\n", err)
	}
}
