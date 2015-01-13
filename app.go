// Copyright 2014 Claudemiro Alves Feitosa Neto. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"sync"

	log "github.com/golang/glog"
)

// An App
type App struct {
	sync.Mutex

	Name                string
	AppID               string
	Key                 string
	Secret              string
	OnlySSL             bool
	ApplicationDisabled bool
	UserEvents          bool
	WebHooks            bool
	URLWebHook          string

	Channels    map[string]*Channel    `json:"-"`
	Subscribers map[string]*Subscriber `json:"-"`
}

// Alloc memory for Subscribers and Channels
func (a *App) Init() {
	a.Subscribers = make(map[string]*Subscriber)
	a.Channels = make(map[string]*Channel)
}

// Only Presence channels
func (a *App) PresenceChannels() []*Channel {
	var channels []*Channel

	for _, c := range a.Channels {
		if c.IsPresence() {
			channels = append(channels, c)
		}
	}

	return channels
}

// Only Private channels
func (a *App) PrivateChannels() []*Channel {
	var channels []*Channel

	for _, c := range a.Channels {
		if c.IsPrivate() {
			channels = append(channels, c)
		}
	}

	return channels
}

// Only Public channels
func (a *App) PublicChannels() []*Channel {
	var channels []*Channel

	for _, c := range a.Channels {
		if c.IsPublic() {
			channels = append(channels, c)
		}
	}

	return channels
}

// Disconnect Socket
func (a *App) Disconnect(socketID string) {
	log.Infof("Disconnecting socket %+v", socketID)

	s, err := a.FindSubscriber(socketID)

	if err != nil {
		log.Infof("Socket not found, %+v", err)
		return
	}

	// Unsubscribe from channels
	for _, c := range a.Channels {
		if c.IsSubscribed(s) {
			c.Unsubscribe(a, s)
		}
	}

	// Remove from app
	a.Lock()
	_, exists := a.Subscribers[s.SocketID]
	if !exists {
		return
	}

	delete(a.Subscribers, s.SocketID)

	a.Unlock()
}

// Connect a new Subscriber
func (a *App) Connect(s *Subscriber) {
	log.Infof("Adding a new Subscriber %s to app %s", s.SocketID, a.Name)
	a.Lock()
	a.Subscribers[s.SocketID] = s
	a.Unlock()
}

// Find a Subscriber on this app
func (a *App) FindSubscriber(socketID string) (*Subscriber, error) {

	s, exists := a.Subscribers[socketID]

	if exists {
		return s, nil
	}

	return nil, errors.New("Subscriber not found")
}

// Add a new Channel to this APP
func (a *App) AddChannel(c *Channel) {
	log.Infof("Adding a new channel %s to app %s", c.ChannelID, a.Name)

	a.Lock()
	a.Channels[c.ChannelID] = c
	a.Unlock()
}

// Returns a Channel from this app
// If not found then the channel is created and added to this app
func (a *App) FindOrCreateChannelByChannelID(n string) *Channel {
	c, err := a.FindChannelByChannelID(n)

	if err != nil {
		c = NewChannel(n)
		a.AddChannel(c)
	}

	return c
}

// Find the channel by channel ID
func (a *App) FindChannelByChannelID(n string) (*Channel, error) {

	c, exists := a.Channels[n]

	if exists {
		return c, nil
	}

	return nil, errors.New("Channel does not exists")
}

func (a *App) Publish(c *Channel, event RawEvent, ignore string) error {
	return c.Publish(a, event, ignore)
}

func (a *App) Unsubscribe(c *Channel, s *Subscriber) error {
	return c.Unsubscribe(a, s)
}

func (a *App) Subscribe(c *Channel, s *Subscriber, data string) {
	c.Subscribe(a, s, data)
}
