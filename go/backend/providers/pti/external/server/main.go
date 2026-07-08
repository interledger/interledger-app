package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/interledger/interledger-app/go/backend/providers/pti/external/mock"
	"github.com/interledger/interledger-app/go/log"
	"go.uber.org/zap"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	pti := mock.NewPTI()

	var wg sync.WaitGroup
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: pti.Routes(),
	}

	wg.Add(1)
	go func(sigCh chan os.Signal, wg *sync.WaitGroup) {
		defer wg.Done()
		<-sigCh

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		log.Info(fmt.Sprintf("got signal attempting graceful HTTP shutdown: %s", server.Addr))
		_ = server.Shutdown(shutdownCtx)
		err := pti.Save()
		if err != nil {
			log.Error("Failed to save mock PTI state to file.")
		}
	}(ch, &wg)

	go func() {
		log.Info("PTI mock server started", zap.String("address", server.Addr))
		err := server.ListenAndServe()
		// http.ErrServerClosed is returned immediately after Shutdown is called.
		//Don't panic and let the HTTP shutdown inside the 30-second timeout.
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("failed to start HTTP server", zap.Error(err))
		}
	}()

	wg.Wait()
}
