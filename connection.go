package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	log "github.com/smashwilson/sprint-closer/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

// Connection performs authenticated actions against the Trello API.
type Connection struct {
	profile Profile
}

// List captures information about a Trello List.
type List struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"pos"`
}

func (c Connection) url(parts []string, query map[string]string) string {
	pathParts := []string{"1"}
	pathParts = append(pathParts, parts...)

	queryValues := url.Values{
		"key":   []string{c.profile.Key},
		"token": []string{c.profile.Token},
	}

	for key, value := range query {
		queryValues[key] = []string{value}
	}

	u := url.URL{
		Scheme:   "https",
		Host:     "trello.com",
		Path:     strings.Join(pathParts, "/"),
		RawQuery: queryValues.Encode(),
	}

	return u.String()
}

func (c Connection) extract(resp *http.Response, response interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		rbody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			rbody = []byte(err.Error())
		}
		return fmt.Errorf("Unexpected status code: %d\n%s", resp.StatusCode, rbody)
	}

	if response != nil {
		return json.NewDecoder(resp.Body).Decode(response)
	}

	return nil
}

func (c Connection) get(url string, response interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	return c.extract(resp, response)
}

func (c Connection) post(url string, payload, response interface{}) error {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(b)
	}

	resp, err := http.Post(url, "application/json", body)
	if err != nil {
		return err
	}

	return c.extract(resp, response)
}

func (c Connection) put(url string, payload, response interface{}) error {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	return c.extract(resp, response)
}

// CreateBoard creates a new Trello board.
func (c Connection) CreateBoard(name string, sourceBoardID string) error {
	u := c.url([]string{"boards"}, nil)

	reqBody := map[string]string{
		"name":                 name,
		"idOrganization":       c.profile.Organization,
		"prefs_permissonLevel": "org",
	}

	if sourceBoardID != "" {
		reqBody["idBoardSource"] = sourceBoardID
		reqBody["keepFromSource"] = "memberships"
	}

	return c.post(u, reqBody, nil)
}

// FindBoard discovers the ID of an existing board by name.
func (c Connection) FindBoard(name string) (string, error) {
	u := c.url([]string{"organizations", c.profile.Organization, "boards"}, map[string]string{
		"fields": "name",
	})

	var boardResults []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	err := c.get(u, &boardResults)
	if err != nil {
		return "", err
	}

	for _, board := range boardResults {
		log.WithFields(log.Fields{
			"name": board.Name,
			"id":   board.ID,
		}).Debug("Board")

		if board.Name == name {
			return board.ID, nil
		}
	}

	return "", fmt.Errorf("Unable to find a board with the name [%s].", name)
}

// FindList locates a list on a board by name.
func (c Connection) FindList(name string, boardID string) (*List, error) {
	u := c.url([]string{"boards", boardID, "lists"}, nil)

	var listResults []List

	err := c.get(u, &listResults)
	if err != nil {
		return nil, err
	}

	for _, list := range listResults {
		log.WithFields(log.Fields{
			"name": list.Name,
			"id":   list.ID,
			"pos":  list.Position,
		}).Debug("List")

		if list.Name == name {
			return &list, nil
		}
	}

	return nil, fmt.Errorf("Unable to find a list with the name [%s].", name)
}

// MoveList moves a list to a different board.
func (c Connection) MoveList(listID string, toBoardID string, position int) error {
	u := c.url([]string{"lists", listID, "idBoard"}, nil)

	var params struct {
		Value    string `json:"value"`
		Position string `json:"pos"`
	}

	params.Value = toBoardID
	if position != 0 {
		params.Position = strconv.Itoa(position)
	} else {
		params.Position = "top"
	}

	return c.put(u, &params, nil)
}

// AddList creates a new list on the specified board at the given position.
func (c Connection) AddList(name string, boardID string, position int) error {
	u := c.url([]string{"lists"}, nil)

	var params struct {
		Name     string `json:"name"`
		BoardID  string `json:"idBoard"`
		Position string `json:"position"`
	}

	params.Name = name
	params.BoardID = boardID
	if position != 0 {
		params.Position = strconv.Itoa(position)
	} else {
		params.Position = "top"
	}

	return c.post(u, &params, nil)
}
