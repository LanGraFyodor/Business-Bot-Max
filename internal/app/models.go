package app

import (
	"net/http"
	"sync"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}
type Channel struct {
	TeamID      string `json:"teamId"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}
type Comment struct {
	ID      string `json:"id"`
	Author  User   `json:"author"`
	Text    string `json:"text"`
	Created string `json:"created"`
	System  bool   `json:"system"`
}
type Version struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	MIME    string `json:"mime"`
	Size    int64  `json:"size"`
	Created string `json:"created"`
	Author  string `json:"author"`
}
type Document struct {
	ImportMethod      string    `json:"importMethod,omitempty"`
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	ChannelID         string    `json:"channelId"`
	Owner             User      `json:"owner"`
	Status            string    `json:"status"`
	Due               string    `json:"due"`
	Description       string    `json:"description"`
	Created           string    `json:"created"`
	Updated           string    `json:"updated"`
	Versions          []Version `json:"versions"`
	Comments          []Comment `json:"comments"`
	Revision          int       `json:"revision"`
	IntegrationID     string    `json:"integrationId,omitempty"`
	SourceSystem      string    `json:"sourceSystem,omitempty"`
	EscalatedFrom     string    `json:"escalatedFrom,omitempty"`
	EscalatedRevision int       `json:"escalatedRevision,omitempty"`
	ExternalID        string    `json:"externalId,omitempty"`
}
type IntegrationKey struct {
	ID       string `json:"id"`
	TeamID   string `json:"teamId"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Hash     string `json:"hash"`
	Created  string `json:"created"`
}
type Account struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Registered string `json:"registered"`
}
type State struct {
	DemoSeedVersion int              `json:"demoSeedVersion,omitempty"`
	PollMarker      string           `json:"pollMarker,omitempty"`
	Accounts        []Account        `json:"accounts,omitempty"`
	Teams           []Team           `json:"teams"`
	Messages        []ChannelMessage `json:"messages"`
	Channels        []Channel        `json:"channels"`
	Documents       []Document       `json:"documents"`
	Outbox          []Notification   `json:"outbox,omitempty"`
	Integrations    []IntegrationKey `json:"integrations,omitempty"`
}
type App struct {
	mu        sync.Mutex
	state     State
	dir       string
	demo      bool
	token     string
	members   map[string]bool
	managers  map[string]bool
	client    *http.Client
	downloads map[string]downloadTicket
}
