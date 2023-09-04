package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Netflix/go-env"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func main() {
	var environment Environment
	_, err := env.UnmarshalFromEnviron(&environment)
	if err != nil {
		log.Fatal("Failed to parse environment variables.", zap.Error(err))
	}

	_ = NewBackends(&environment)

	log.Info("Bot is running")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
