package main

import (
    "database/sql"
    "hometown-bot/bot"
    "hometown-bot/log"
    "hometown-bot/repository"
    "hometown-bot/storage"
    "os"
)

func main() {
    log.Info().Println("env: reading keys")
    botToken, ok := os.LookupEnv("BOT_TOKEN")
    if !ok {
        log.Error().Println("env: key BOT_TOKEN is empty or not set")
        os.Exit(1)
    }

    log.Info().Println("storage: initializing")
    db, err := storage.Load()
    if err != nil {
        log.Error().Printf("storage: %v", err)
        os.Exit(1)
    }

    defer func(db *sql.DB) {
        log.Info().Println("storage: closing")
        err := db.Close()
        if err != nil {
            log.Error().Printf("storage: %v", err)
            os.Exit(1)
        }
    }(db)

    log.Info().Println("repository: initializing")
    channelRepository := repository.NewChannel(db)
    channelMembersRepository := repository.NewChannelMembers(db)
    lobbyRepository := repository.NewLobby(db)

    log.Info().Println("bot: initializing")
    bot.Token = botToken
    b := bot.Create(*channelRepository, *channelMembersRepository, *lobbyRepository)

    if err := b.Run(); err != nil {
        log.Error().Printf("bot: %v", err)
        os.Exit(1)
    }
}
