package main

import (
	"hometown-bot/bot"
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

	bot.BotToken = botToken
	bot.GuildID = guildId
	bot.Run()
}
