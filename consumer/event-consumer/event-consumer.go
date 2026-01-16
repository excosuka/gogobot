package event_consumer

import (
	"gogobot/events"
	"log"
	"time"
)

type EventConsumer struct {
	fetcher   events.Fetcher
	processor events.Processor
	batchSize int
}

func New(fetcher events.Fetcher, processor events.Processor, batchSize int) EventConsumer {
	return EventConsumer{fetcher: fetcher, processor: processor, batchSize: batchSize}
}

func (c EventConsumer) Start() error {
	for {
		// механизм ретрая можно допилить
		gotEvents, err := c.fetcher.Fetch(c.batchSize)
		// Доработать обработку ошибок
		if err != nil {
			log.Printf("[ERR] consumer:  %s", err.Error())
			continue
		}

		if len(gotEvents) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		if err := c.handleEvents(gotEvents); err != nil {
			log.Print(err)

			continue
		}
	}
}

/*
Проблемы и идеи доработки
1. Потеря событий: ретраи, возвращение в хранилище, фоллбек, подтверждение
2. обработка всей пачки: останавливаться после первой ошибки
3. Параллельная обработка
*/

func (c *EventConsumer) handleEvents(events []events.Event) error {
	for _, event := range events {
		log.Printf("got new event: %s", event.Text)

		if err := c.processor.Process(event); err != nil {
			log.Printf("[ERR] processor had met an error:  %s", err.Error())

			continue
		}

	}
	return nil
}
