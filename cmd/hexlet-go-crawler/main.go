package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"code/crawler"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:      "hexlet-go-crawler",
		Usage:     "analyze a website structure",
		ArgsUsage: "<url>",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "depth",
				Value: 10,
				Usage: "crawl depth",
			},
			&cli.IntFlag{
				Name:  "retries",
				Value: 1,
				Usage: "number of retries for failed requests",
			},
			&cli.DurationFlag{
				Name:  "delay",
				Value: 0,
				Usage: "delay between requests (example: 200ms, 1s)",
			},
			&cli.DurationFlag{
				Name:  "timeout",
				Value: 15 * time.Second,
				Usage: "per-request timeout",
			},
			&cli.Float64Flag{
				Name:  "rps",
				Value: 0,
				Usage: "limit requests per second (overrides delay)",
			},
			&cli.StringFlag{
				Name:  "user-agent",
				Usage: "custom user agent",
			},
			&cli.IntFlag{
				Name:  "workers",
				Value: 4,
				Usage: "number of concurrent workers",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return cli.ShowAppHelp(c)
			}

			url := c.Args().Get(0)
			userAgent := c.String("user-agent")

			httpClient := &http.Client{
				Timeout: c.Duration("timeout"),
			}

			opts := crawler.Options{
				URL:         url,
				Depth:       c.Int("depth"),
				Retries:     c.Int("retries"),
				Delay:       c.Duration("delay"),
				Timeout:     c.Duration("timeout"),
				UserAgent:   userAgent,
				Concurrency: c.Int("workers"),
				IndentJSON:  true,
				HTTPClient:  httpClient,
			}

			result, err := crawler.Analyze(context.Background(), opts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return nil
			}

			fmt.Println(string(result))
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}
