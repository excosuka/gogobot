package main

import (
	"flag"
	tgClient "gogobot/clients/telegram"
	event_consumer "gogobot/consumer/event-consumer"
	"gogobot/events/telegram"
	"gogobot/events/telegram/types/stateStorage"
	"gogobot/storage/files"

	"log"
)

// 8482487054:AAHWjRdwt16-9KXTnIrrbD2D9OGhAQz73wE

const (
	tgBotHost   = "api.telegram.org"
	storagePath = "storage"
	batchSize   = 100
)

func main() {

	eventsProcessor := telegram.New(
		tgClient.New(tgBotHost, mustToken()),
		files.NewStorage(storagePath),
		stateStorage.New())

	log.Println("Starting telegram bot")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(); err != nil {
		log.Fatal("service is stopped", err)
	}

}

func mustToken() string {
	token := flag.String(
		"bot-token",
		"",
		"token for access to telegram api for bots",
	)

	flag.Parse()

	if *token == "" {
		log.Fatal("token is not specified")
	}

	return *token

}
