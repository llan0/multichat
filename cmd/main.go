package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/llan0/multichat/internal/adapters/kick"
	"github.com/llan0/multichat/internal/adapters/twitch"
	"github.com/llan0/multichat/internal/logger"
	"github.com/llan0/multichat/internal/models"
	"github.com/llan0/multichat/internal/service"
	"github.com/llan0/multichat/internal/ui"
	"go.uber.org/zap"
)

func main() {
	enableProfiling := flag.Bool("pprof", false, "Enable pprof profiling server on :6060")
	flag.Parse()

	logger.Log.Info("starting multichat application")
	defer logger.Sync()

	// start pprof server if enabled
	if *enableProfiling {
		go func() {
			// stats endpoint for quick perf metrics
			http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				fmt.Fprintf(w, "Memory: %.2f MB\n", float64(m.Alloc)/1024/1024)
				fmt.Fprintf(w, "Goroutines: %d\n", runtime.NumGoroutine())
				fmt.Fprintf(w, "GC Count: %d\n", m.NumGC)
			})

			logger.Log.Info("pprof server started", zap.String("address", "http://localhost:6060/debug/pprof/"))
			logger.Log.Info("quick stats: http://localhost:6060/stats")
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				logger.Log.Error("pprof server error", zap.Error(err))
			}
		}()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Log.Info("shutdown signal received")
		cancel()
	}()

	// mock producers
	var twitchProducer service.Producer = &twitchProducer{}
	var kickProducer service.Producer = &kickProducer{}

	logger.Log.Info("initializing producers", zap.Int("producer_count", 2))

	// mocks -> service -> UI
	ui.Start(ctx, twitchProducer, kickProducer)

	logger.Log.Info("application shutdown complete")
}

// twitch producer
type twitchProducer struct{}

func (t *twitchProducer) Stream(ctx context.Context) <-chan models.ChatMessage {
	return twitch.Stream(ctx)
}

// kick producer
type kickProducer struct{}

func (k *kickProducer) Stream(ctx context.Context) <-chan models.ChatMessage {
	return kick.Stream(ctx)
}
