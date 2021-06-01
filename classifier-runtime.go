package main

import (
	"fmt"
	"log"
	"os"

	"github.com/shynuu/classifier-runtime/runtime"
	"github.com/urfave/cli/v2"
)

func main() {

	var flush bool = false
	var acm bool = false
	var config string = ""

	app := &cli.App{
		Name:  "classifier-runtime",
		Usage: "The runtime for the slice classifier",
		Authors: []*cli.Author{
			{Name: "Youssouf Drif"},
		},
		Copyright: "Copyright (c) 2021 IRT Saint Exupéry",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Usage:       "Load configuration from `FILE`",
				Destination: &config,
				Required:    true,
				DefaultText: "not set",
			},
			&cli.BoolFlag{
				Name:        "flush",
				Usage:       "Flush IPTABLES and remove existing chains",
				Destination: &flush,
			},
			&cli.BoolFlag{
				Name:        "acm",
				Usage:       "Activate the ACM simulation",
				Destination: &acm,
				DefaultText: "not activated",
			},
		},
		Action: func(c *cli.Context) error {
			err := runtime.InitRuntime(flush, config)
			if err != nil {
				fmt.Println("Init error, exiting...")
				os.Exit(1)
			}

			runtime.Run()

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
