package main

import (
	"hometown-bot/bot"
	"hometown-bot/repository"
	"hometown-bot/storage"
	"log"
	"os"
)

func main() {
	botToken, ok := os.LookupEnv("BOT_TOKEN")
	if !ok {
		log.Fatal("Must set Discord token as env variable: BOT_TOKEN")
	}

	log.Println("Loading storage...")
	db, err := storage.Load()
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	channelRepository := repository.NewChannelRepository(db)
	channelMembersRepository := repository.NewChannelMembersRepository(db)
	lobbyRepository := repository.NewLobbyRepository(db)

	bot.BotToken = botToken
	b := bot.Create(*channelRepository, *channelMembersRepository, *lobbyRepository)

	err = b.Run()
	if err != nil {
		log.Fatal(err)
	}
}
