package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/pastel-sketchbook/reins/internal/cli"
)

// version is injected via ldflags at build time.
var version = "dev"

func main() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	slog.SetDefault(slog.New(handler))

	ctx := context.Background()
	cli.SetVersion(version)
	os.Exit(cli.Run(ctx, os.Args))
}
