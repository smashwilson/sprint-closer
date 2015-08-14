package main

import (
	"encoding/json"
	"os"

	log "github.com/smashwilson/sprint-closer/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

// Profile is the JSON-serialized configuration.
type Profile struct {
	Key          string `json:"key"`
	Token        string `json:"token"`
	Organization string `json:"organization"`
}

// LoadProfile attempts to locate and load a profile from disk.
func LoadProfile(path string) (*Profile, error) {
	inf, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.WithField("profile path", path).Error("You have no profile yet!")
		}

		return nil, err
	}

	p := &Profile{}
	err = json.NewDecoder(inf).Decode(p)
	if err != nil {
		return p, err
	}

	return p, nil
}
