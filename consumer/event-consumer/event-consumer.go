package eventconsumer

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Daniel1212649/LinksHelperBot/events"
)

type Consumer struct {
	fetcher      events.Fetcher
	processor    events.Processor
	batchSize    int
	pollInterval time.Duration
}

func New(fetcher events.Fetcher, processor events.Processor, batchSize int, pollInterval time.Duration) Consumer {
	return Consumer{
		fetcher:      fetcher,
		processor:    processor,
		batchSize:    batchSize,
		pollInterval: pollInterval,
	}
}

func (c Consumer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		gotEvents, err := c.fetcher.Fetch(ctx, c.batchSize)
		if err != nil {
			log.Printf("[ERR] consumer: %s", err.Error())
			if err := sleep(ctx, c.pollInterval); err != nil {
				return nil
			}
			continue
		}

		if len(gotEvents) == 0 {
			if err := sleep(ctx, c.pollInterval); err != nil {
				return nil
			}
			continue
		}

		c.handleEvents(ctx, gotEvents)
	}
}

func (c Consumer) handleEvents(ctx context.Context, events []events.Event) {
	for _, event := range events {
		log.Printf("got new event: %s", event.Text)

		if err := c.processor.Process(ctx, event); err != nil {
			log.Printf("can't handle event: %s", err.Error())
			continue
		}
	}
}

func sleep(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return errors.New("context canceled")
	case <-timer.C:
		return nil
	}
}
