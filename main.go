package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

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
	handleErr(err)
	log.SetLevel(level)

	p, err := LoadProfile(c.String("profile"))
	handleErr(err)

	conn := Connection{profile: *p}

	currentSprintID, err := conn.FindBoard("Current Sprint")
	handleErr(err)

	log.WithField("board id", currentSprintID).Debug("Current sprint board located.")

	doneList, err := conn.FindList("Done", currentSprintID)
	handleErr(err)

	log.WithField("list id", doneList.ID).Debug("Done list located.")

	archiveBoardName := newBoardName()
	err = conn.CreateBoard(archiveBoardName, currentSprintID)
	handleErr(err)

	archiveBoardID, err := conn.FindBoard(archiveBoardName)
	handleErr(err)

	log.WithFields(log.Fields{
		"board id":   archiveBoardID,
		"board name": archiveBoardName,
	}).Info("Created archive board.")

	err = conn.MoveList(doneList.ID, archiveBoardID, 1)
	handleErr(err)

	log.Info("Moved Done list to the archive board.")

	err = conn.AddList("Done", currentSprintID, doneList.Position)
	handleErr(err)

	log.Info("Created Done list on the Current Sprint board.")
}

func handleErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func newBoardName() string {
	n := time.Now()

	daysUntilFriday := time.Friday - n.Weekday()
	friday := n.AddDate(0, 0, int(daysUntilFriday))

	return fmt.Sprintf("DevEx Sprint %s", friday.Format("2006-01-02"))
}
