package main

import (
	"fmt"
	"os"

	log "github.com/smashwilson/sprint-closer/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/smashwilson/sprint-closer/Godeps/_workspace/src/github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "sprint-closer"
	app.Usage = "Close the current DevEx sprint"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log, l",
			Value: "INFO",
			Usage: "Logging level",
		},
	}

	app.Action = run

	app.Run(os.Args)
}

func run(c *cli.Context) {
	levelName := c.String("log")
	level, err := log.ParseLevel(levelName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	log.SetLevel(level)
}
