package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/jlhg/mox"
)

func main() {
	app := &cli.App{
		Usage: "mox CLI application",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "Load configuration from `FILE`",
				Required: true,
			},
		},
		Before: func(ctx *cli.Context) (err error) {
			var config *mox.Config
			config, err = mox.NewConfig(ctx.String("config"))
			if err != nil {
				err = fmt.Errorf("failed to parse config: %w", err)
				return
			}

			err = config.Init()
			if err != nil {
				err = fmt.Errorf("failed to initialize config: %w", err)
				return
			}

			ctx.Context = context.WithValue(ctx.Context, "config", config)

			return
		},
		Commands: []*cli.Command{
			{
				Name:    "download",
				Aliases: []string{"dl"},
				Usage:   "Download comics",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:     "id",
						Aliases:  []string{"i"},
						Usage:    "Comics ID",
						Required: true,
					},
				},
				Action: func(ctx *cli.Context) (err error) {
					var c *mox.MoxClient

					c, err = mox.NewMoxClient(ctx.Context)
					if err != nil {
						return
					}

					err = c.Login()
					if err != nil {
						return
					}

					err = c.DownloadComics(ctx.Int("id"))
					if err != nil {
						return
					}

					err = c.Logout()
					if err != nil {
						return
					}

					return
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
