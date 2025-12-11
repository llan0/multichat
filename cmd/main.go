package main

import (
	"os"

	"github.com/llan0/multichat/internal/logger"
	"github.com/llan0/multichat/internal/ui"
)

const defaultChannel = "xqc"

func main() {
	log, err := logger.New()
	if err != nil {
		os.Exit(1)
	}
	defer log.Sync()

	channel := defaultChannel
	if len(os.Args) > 1 {
		channel = os.Args[1]
	}

	ui.Run(log, channel)
}
