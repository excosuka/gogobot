package event_consumer

import (
	"errors"
	"gogobot/events"
	"gogobot/events/telegram"
	"gogobot/events/telegram/types"
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
1. Потеря событий: ретраи, возвращение в хранZилище, фоллбек, подтверждение
2. обработка всей пачки: останавливаться после первой ошибки
3. Параллельная обработка
*/

func (c *EventConsumer) handleEvents(eventsFor []events.Event) error {
	for _, event := range eventsFor {
		switch event.Type {
		case events.Message:
			payload := event.Payload.(types.MessagePayload)
			meta := event.Meta.(telegram.Meta)
			log.Printf("got message from %s: %s", meta.Username, payload.Text)
		case events.Callback:
			payload := event.Payload.(types.CallbackPayload)
			meta := event.Meta.(telegram.Meta)
			log.Printf("got callback from %s: %s", meta.Username, payload.Data)
		default:
			log.Printf("got unknown event type")
		}

		if err := c.processor.Process(event); err != nil {
			var ue telegram.UserError
			if errors.As(err, &ue) {
				if p, ok := c.processor.(*telegram.Processor); ok {
					_ = ue.Send(p, event)
				}
				continue
			}
			log.Printf("[ERR] processor error: %s", err)

		}
	}

	return nil
}
