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
	guildId, ok := os.LookupEnv("GUILD_ID")
	if !ok {
		log.Fatal("Must set Discord token as env variable: GUILD_ID")
	}
	
	log.Println("Loading storage...")
	db, err := storage.Load()
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	channelRepository := repository.NewChannelRepository(db)
	lobbyRepository := repository.NewLobbyRepository(db)

	bot.BotToken = botToken
	bot.GuildID = guildId
	b:= bot.Create(*channelRepository, *lobbyRepository)

	err = b.Run()
	if err != nil {
		log.Fatal(err)
	}
}
