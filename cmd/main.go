package main

import (
	"flag"
	"fmt"
	"os"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"

	"github.com/llan0/multichat/internal/logger"
	"github.com/llan0/multichat/internal/ui"
	"github.com/llan0/multichat/internal/version"
)

const defaultChannel = "xqc"

func main() {
	showVersion := flag.Bool("version", false, "print version information")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Info())
		return
	}

	log, err := logger.New()
	if err != nil {
		os.Exit(1)
	}
	defer log.Sync()

	channel := defaultChannel
	if flag.NArg() > 0 {
		channel = flag.Arg(0)
	}

	ui.Run(log, channel)
}
