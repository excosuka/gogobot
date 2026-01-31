package main

import (
	"flag"
	tgClient "gogobot/clients/telegram"
	event_consumer "gogobot/consumer/event-consumer"
	"gogobot/events/session"
	"gogobot/events/telegram"
	"gogobot/events/telegram/types"
	"gogobot/events/telegram/types/stateStorage"
	"gogobot/parserService"
	"gogobot/storage/files"
	"gogobot/storage/pageService"
	"gogobot/storage/searchService"
	"time"

	"log"
)

const sessionTTL = 200 * time.Second

const (
	tgBotHost   = "api.telegram.org"
	storagePath = "storage"
	batchSize   = 100
)

func main() {
	storage := files.NewStorage(storagePath)

	client := tgClient.New(tgBotHost, mustToken())
	sessionMgr := session.NewTTLManager(
		sessionTTL,
		func(s *types.UserSession) {

			if s.UserState != types.StateIdle {
				_ = client.SendMessage(
					s.ChatId,
					"⌛ Dialog was reset cause of inactive",
				)
			}
			stateStorage.ResetSession(s)
		},
	)

	parserService := parserService.New()

	eventsProcessor := telegram.New(
		client,
		pageService.New(storage),
		stateStorage.New(),
		searchService.New(storage),
		sessionMgr,
		parserService,
	)

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
