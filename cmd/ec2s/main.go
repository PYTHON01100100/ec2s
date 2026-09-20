// Command ec2s is a terminal UI for managing AWS EC2 instances across
// multiple accounts and regions at once.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/PYTHON01100100/ec2s/internal/app"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	configPath := flag.String("config", "", "path to ec2s config file (default: search order documented in README)")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("ec2s " + version)
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := app.Run(ctx, *configPath); err != nil {
		fmt.Fprintln(os.Stderr, "ec2s:", err)
		os.Exit(1)
	}
}
