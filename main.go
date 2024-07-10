package main

import (
    "database/sql"
    "hometown-bot/bot"
    "hometown-bot/repository"
    "hometown-bot/storage"
    "log"
    "os"
)

func main() {
    botToken, ok := os.LookupEnv("BOT_TOKEN")
    if !ok {
        log.Fatalln("Must set Discord token as env variable: BOT_TOKEN")
    }

    log.Println("Loading storage...")
    db, err := storage.Load()
    if err != nil {
        log.Fatalln(err)
    }

    defer func(db *sql.DB) {
        err := db.Close()
        if err != nil {
            log.Fatalln(err)
        }
    }(db)

    channelRepository := repository.NewChannel(db)
    channelMembersRepository := repository.NewChannelMembers(db)
    lobbyRepository := repository.NewLobby(db)

    bot.Token = botToken
    b := bot.Create(*channelRepository, *channelMembersRepository, *lobbyRepository)

    err = b.Run()
    if err != nil {
        log.Fatalln(err)
    }
}
