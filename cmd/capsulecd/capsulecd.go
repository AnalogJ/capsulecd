package main

import (
	"fmt"
	"os"
	"time"

	"capsulecd/pkg"
	"capsulecd/pkg/config"
	"capsulecd/pkg/errors"
	"capsulecd/pkg/utils"
	"capsulecd/pkg/version"
	"gopkg.in/urfave/cli.v2"
	"path/filepath"
)

func main() {
	app := &cli.App{
		Name:     "capsulecd",
		Usage:    "Continuous Delivery scripts for automating package releases",
		Version:  version.VERSION,
		Compiled: time.Now(),
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Jason Kulatunga",
				Email: "jason@thesparktree.com",
			},
		},
		Before: func(c *cli.Context) error {

			capsuleUrl := "https://www.capsulecd.com"
			versionInfo := fmt.Sprintf("%s-%s", "linux", version.VERSION)

			subtitle := capsuleUrl + utils.LeftPad2Len(versionInfo," ", 53-len(capsuleUrl))

			fmt.Fprintf(c.App.Writer, fmt.Sprintf(utils.StripIndent(
				`
			  ___   __   ____  ____  _  _  __    ____  ___  ____
			 / __) / _\ (  _ \/ ___)/ )( \(  )  (  __)/ __)(    \
			( (__ /    \ ) __/\___ \) \/ (/ (_/\ ) _)( (__  ) D (
			 \___)\_/\_/(__)  (____/\____/\____/(____)\___)(____/
			%s

			`), subtitle ))
			return nil
		},

		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start a new CapsuleCD package pipeline",
				Action: func(c *cli.Context) error {

					config, _ := config.Create()
					config.Set("scm", c.String("scm"))
					config.Set("package_type", c.String("package_type"))
					//config.Set("dry_run", c.String("dry_run"))

					//load configuration file.
					if c.String("config_file") != "" {
						absConfigPath, err := filepath.Abs(c.String("config_file"))
						if err != nil {
							return err
						}
						err = config.ReadConfig(absConfigPath) //ignore failures to read config file.
						if err != nil {
							return errors.EngineUnspecifiedError("Could not load repository configuration file. Check syntax.")
						}
					}

					//fmt.Println("runner:", config.GetString("runner"))
					fmt.Println("package type:", config.GetString("package_type"))
					fmt.Println("scm:", config.GetString("scm"))
					fmt.Println("repository:", config.GetString("scm_repo_full_name"))
					//fmt.Println("dry run:", config.GetString("dry_run"))

					pipeline := pkg.Pipeline{}
					pipeline.Start(config)

					return nil
				},

				Flags: []cli.Flag{
					//TODO: currently not applicable
					//&cli.StringFlag{
					//	Name:  "runner",
					//	Value: "default", // can be :none, :circleci or :shippable (check the readme for why other hosted providers arn't supported.)
					//	Usage: "The cloud CI runner that is running this PR. (Used to determine the Environmental Variables to parse)",
					//},

					&cli.StringFlag{
						Name:  "scm",
						Value: "default",
						Usage: "The scm for the code, used to determine which git endpoint to clone from, and create releases on",
					},

					&cli.StringFlag{
						Name:    "package_type",
						Aliases: []string{"package-type"},
						Value:   "default",
						Usage:   "The type of package being built.",
					},

					//TODO: currently not applicable.
					//&cli.BoolFlag{
					//	Name:    "dry_run",
					//	Aliases: []string{"dry-run"},
					//	Value:   false,
					//	Usage:   "Specifies that no changes should be pushed to source and no package will be released",
					//},

					&cli.StringFlag{
						Name:    "config_file",
						Aliases: []string{"config-file"},
						Usage:   "Specifies the location of the config file",
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
