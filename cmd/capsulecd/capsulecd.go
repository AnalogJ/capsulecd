package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/urfave/cli.v2"
)
func main() {
	app := &cli.App{
		Name: "capsulecd",
		Usage: "Continuous Delivery scripts for automating package releases",
		Version: "v19.99.0",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Jason Kulatunga",
				Email: "jason@thesparktree.com",
			},
		},

		Commands: []*cli.Command{
			{
				Name:    "start",
				Usage:   "Start a new CapsuleCD package pipeline",
				Action:  func(c *cli.Context) error {


					fmt.Println("runner:", c.String("runner"))
					fmt.Println("source:", c.String("source"))
					fmt.Println("package type:", c.String("package_type"))
					fmt.Println("dry run:", c.Bool("dry_run"))
					fmt.Println("config file:", c.String("config_file"))
					return nil
				},

				Flags: []cli.Flag {
					&cli.StringFlag{
						Name: "runner",
						Value: "default", // can be :none, :circleci or :shippable (check the readme for why other hosted providers arn't supported.)
						Usage: "The cloud CI runner that is running this PR. (Used to determine the Environmental Variables to parse)",
					},

					&cli.StringFlag{
						Name: "source",
						Value: "default",
						Usage: "The source for the code, used to determine which git endpoint to clone from, and create releases on",
					},

					&cli.StringFlag{
						Name: "package_type",
						Aliases: []string{"package-type"},
						Value: "default",
						Usage: "The type of package being built.",
					},

					&cli.BoolFlag{
						Name: "dry_run",
						Aliases: []string{"dry-run"},
						Value: false,
						Usage: "Specifies that no changes should be pushed to source and no package will be released",
					},

					&cli.StringFlag{
						Name: "config_file",
						Aliases: []string{"config-file"},
						Usage: "Specifies the location of the config file",
					},
				},
			},
		},
	}

	app.Run(os.Args)
}