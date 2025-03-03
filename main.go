package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/hostwithquantum/hugo-metamonster/internal/cmd"
	"github.com/hostwithquantum/hugo-metamonster/internal/metamonster"
	"github.com/urfave/cli/v3"
)

var (
	// ignore these sections
	ignoreSections = []string{}

	// log level
	logLevel = slog.LevelInfo
)

func main() {
	cmd := &cli.Command{
		Description: "hugo-metamonster — a cli tool to add the optimized title, description and keywords to your Hugo content",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "csv",
				Required: true,
				Usage:    "path to the metamonster export (csv)",
				Action: func(_ context.Context, _ *cli.Command, s string) error {
					return cmd.Exists(s)
				},
			},
			&cli.StringFlag{
				Name:     "site-path",
				Aliases:  []string{"site", "s"},
				Required: true,
				Usage:    "path to your hugo site's root",
				Action: func(_ context.Context, _ *cli.Command, s string) error {
					if _, err := os.Open(s); err != nil {
						return fmt.Errorf("directory %q does not exists or cannot be read", s)
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:  "hugo-path",
				Value: "hugo",
				Usage: "path to hugo (if not in $PATH)",
				Action: func(_ context.Context, _ *cli.Command, s string) error {
					if s == "hugo" {
						if _, err := exec.LookPath(s); err != nil {
							return fmt.Errorf("%q is not available in path", s)
						}
						return nil
					}
					return cmd.Exists(s)
				},
			},
			&cli.BoolFlag{
				Name:  "debug",
				Value: false,
				Usage: "enable debug logging",
			},
			&cli.StringSliceFlag{
				Name:        "ignore",
				Value:       []string{"content/company"},
				Usage:       "comma separated list of paths to ignore (e.g. content/company,content/faq)",
				Destination: &ignoreSections,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Bool("debug") {
				logLevel = slog.LevelDebug
			}

			// init global logger
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: logLevel,
			})))

			pages, err := metamonster.ListContent(ctx, cmd.String("site-path"), cmd.String("hugo-path"))
			if err != nil {
				return err
			}

			if pages == nil {
				return fmt.Errorf("unable to find any pages (in %q), but no error occurred", cmd.String("site-path"))
			}

			// extract metamonster
			report, err := metamonster.Report(ctx, cmd.String("csv"))
			if err != nil {
				return err
			}

			for url, f := range pages {
				var ignoreThis bool = false

				for _, ignore := range ignoreSections {
					if strings.Contains(f, ignore) {
						ignoreThis = true
						break
					}
				}

				if ignoreThis {
					slog.WarnContext(ctx, "ignoring", slog.String("url", url))
					continue
				}

				// check if URL is in metamonster csv
				if _, ok := report[url]; !ok {
					slog.WarnContext(ctx, "skipping", slog.String("url", url))
					continue
				}

				// update file
				if err := metamonster.Update(ctx, f, report[url]); err != nil {
					return err
				}
			}
			return nil
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("failed to run", slog.Any("err", err))
		os.Exit(1)
	}
	os.Exit(0)
}
