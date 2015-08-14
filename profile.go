package main

import (
	"encoding/json"
	"errors"
	"os"

	log "github.com/smashwilson/sprint-closer/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

// Profile is the JSON-serialized configuration.
type Profile struct {
	Key          string `json:"key"`
	Token        string `json:"token"`
	Organization string `json:"organization"`
}

const noProfileMessage = `Create a file at ~/.trello.json with the following contents:

{
  "key": "",
  "token": "",
  "organization": ""
}

To find a key, log in to Trello and visit:

https://trello.com/1/appKey/generate

To generate a token, use the key and visit:

https://trello.com/1/authorize?key=${KEY}&name=Closer&expiration=never&scope=read,write&response_type=token

To find the organization name, visit the organization on the web UI and look at the last bit of
the URL:

"https://trello.com/automationtesting2" => org name is "automationtesting2"
`

// LoadProfile attempts to locate and load a profile from disk.
func LoadProfile(path string) (*Profile, error) {
	inf, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.WithField("profile path", path).Errorf("You have no profile yet!\n%s", noProfileMessage)
		}

		return nil, err
	}
	defer inf.Close()

	p := &Profile{}
	err = json.NewDecoder(inf).Decode(p)
	if err != nil {
		return p, err
	}

	if p.Key == "" {
		log.Error("To get a Trello API key, log in to the web UI and visit: https://trello.com/1/appKey/generate")
		return p, errors.New("Trello API key missing")
	}

	if p.Token == "" {
		log.Errorf("To get a Trello token, visit https://trello.com/1/authorize?key=%s&name=Closer"+
			"&expiration=never&scope=read,write&response_type=token", p.Key)
		return p, errors.New("Trello token missing")
	}

	if p.Organization == "" {
		return p, errors.New("Trello organization missing")
	}

	return p, nil
}
