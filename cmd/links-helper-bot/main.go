package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/Daniel1212649/LinksHelperBot/clients/telegram"
	"github.com/Daniel1212649/LinksHelperBot/config"
	eventconsumer "github.com/Daniel1212649/LinksHelperBot/consumer/event-consumer"
	tgevents "github.com/Daniel1212649/LinksHelperBot/events/telegram"
	"github.com/Daniel1212649/LinksHelperBot/storage/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("can't connect to postgres: %v", err)
	}
	defer db.Close()

	tgClient := telegram.New(cfg.TelegramAPIHost, cfg.TelegramBotToken, cfg.HTTPTimeout)
	eventsProcessor := tgevents.New(tgClient, db)
	consumer := eventconsumer.New(eventsProcessor, eventsProcessor, cfg.PollBatchSize, cfg.PollInterval)
	tgevents.StartReminderScheduler(ctx, tgClient, db, time.Minute)

	log.Printf("service started env=%s", cfg.AppEnv)

	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("service stopped: %v", err)
	}

	log.Print("service stopped gracefully")
}
