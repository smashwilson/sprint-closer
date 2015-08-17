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

// Org contains a little information about a Trello organization.
type Org struct {
	ID        string
	MemberIDs []string
}

// List captures information about a Trello List.
type List struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Position float64 `json:"pos"`
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

func (c Connection) delete(url string) error {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	_, err = http.DefaultClient.Do(req)
	return err
}

// CreateBoard creates a new Trello board.
func (c Connection) CreateBoard(name string) (string, error) {
	u := c.url([]string{"boards"}, nil)

	reqBody := map[string]string{
		"name":                 name,
		"idOrganization":       c.profile.Organization,
		"prefs_permissonLevel": "org",
	}

	var resp struct {
		ID string `json:"id"`
	}

	err := c.post(u, reqBody, &resp)
	return resp.ID, err
}

// FindMyUserID returns the user ID associated with the token we're using.
func (c Connection) FindMyUserID() (string, error) {
	u := c.url([]string{"members", "me"}, nil)

	var respBody struct {
		ID string `json:"id"`
	}

	err := c.get(u, &respBody)
	return respBody.ID, err
}

// FindOrg looks up the ID and members of the configured organization.
func (c Connection) FindOrg() (*Org, error) {
	u := c.url([]string{"organizations", c.profile.Organization}, map[string]string{
		"members": "all",
	})

	var respBody struct {
		ID      string `json:"id"`
		Members []struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		} `json:"members"`
	}

	err := c.get(u, &respBody)
	if err != nil {
		return nil, err
	}

	memberIDs := make([]string, 0, len(respBody.Members))
	for _, member := range respBody.Members {
		memberIDs = append(memberIDs, member.ID)
	}

	return &Org{
		ID:        respBody.ID,
		MemberIDs: memberIDs,
	}, err
}

// AddMember adds all members of the named organization to a board.
func (c Connection) AddMember(boardID string, memberID string) error {
	u := c.url([]string{"boards", boardID, "members", memberID}, map[string]string{
		"type": "normal",
	})

	return c.put(u, nil, nil)
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

// GetListIDs returns an array of IDs of the lists on an existing board.
func (c Connection) GetListIDs(boardID string) ([]string, error) {
	u := c.url([]string{"boards", boardID, "lists"}, nil)

	var respBody []struct {
		ID string `json:"id"`
	}

	err := c.get(u, &respBody)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(respBody))
	for _, each := range respBody {
		ids = append(ids, each.ID)
	}

	return ids, nil
}

// DeleteList deletes a list, or as close to it as Trello lets you.
func (c Connection) DeleteList(listID string) error {
	u := c.url([]string{"lists", listID, "closed"}, map[string]string{
		"value": "true",
	})

	return c.put(u, nil, nil)
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
			"name":     list.Name,
			"id":       list.ID,
			"position": list.Position,
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
func (c Connection) AddList(name, boardID string, position float64) error {
	u := c.url([]string{"boards", boardID, "lists"}, map[string]string{
		"name": name,
		"pos":  strconv.FormatFloat(position, 'f', 1, 64),
	})

	return c.post(u, nil, nil)
}
