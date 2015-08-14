package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	log "github.com/smashwilson/sprint-closer/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/smashwilson/sprint-closer/Godeps/_workspace/src/github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "sprint-closer"
	app.Usage = "Close the current DevEx sprint"
	app.Version = "0.0.0"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log, l",
			Value: "info",
			Usage: "Logging level",
		},
		cli.StringFlag{
			Name:  "profile, p",
			Value: path.Join(os.Getenv("HOME"), ".trello.json"),
			Usage: "Path to a JSON profile.",
		},
	}

	app.Action = run

	app.Run(os.Args)
}

func run(c *cli.Context) {
	levelName := strings.ToLower(c.String("log"))
	level, err := log.ParseLevel(levelName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	log.SetLevel(level)

	p, err := LoadProfile(c.String("profile"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	fmt.Printf("Profile: %#v\n", p)
}
