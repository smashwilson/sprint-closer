package main

// Profile is the JSON-serialized configuration.
type Profile struct {
	Key          string `json:"key"`
	Token        string `json:"token"`
	Organization string `json:"organization"`
}
