package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
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

	org, err := conn.FindOrg()
	handleErr(err)

	log.WithFields(log.Fields{
		"org id":       org.ID,
		"member count": len(org.MemberIDs),
	}).Debug("Organization ID located.")

	myID, err := conn.FindMyUserID()
	handleErr(err)

	log.WithField("user id", myID).Debug("My user ID located.")

	archiveBoardName := newBoardName()
	archiveBoardID, err := conn.CreateBoard(archiveBoardName)
	handleErr(err)

	log.WithFields(log.Fields{
		"board id":   archiveBoardID,
		"board name": archiveBoardName,
	}).Info("Created archive board.")

	for _, memberID := range org.MemberIDs {
		if memberID != myID {
			log.WithField("member ID", memberID).Debug("Granting access")
			err = conn.AddMember(archiveBoardID, memberID)
			handleErr(err)
		}
	}

	log.WithField("member count", len(org.MemberIDs)).Info("Granted access to this organization.")

	autoListIDs, err := conn.GetListIDs(archiveBoardID)
	handleErr(err)

	for _, listID := range autoListIDs {
		log.WithField("list id", listID).Debug("Deleting list")
		err = conn.DeleteList(listID)
		handleErr(err)
	}

	log.Info("Deleted pre-existing lists.")

	err = conn.MoveList(doneList.ID, archiveBoardID, 1)
	handleErr(err)

	log.Info("Moved Done list to the archive board.")
	err = conn.AddList("Done", currentSprintID, strconv.Itoa(doneList.Index))
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

	daysUntilFriday := (time.Friday - n.Weekday() - 7) % 7
	friday := n.AddDate(0, 0, int(daysUntilFriday))

	return fmt.Sprintf("DevEx Sprint %s", friday.Format("2006-01-02"))
}
