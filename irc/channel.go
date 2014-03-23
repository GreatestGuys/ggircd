package irc

import (
  "log"
)

const (
  ChannelModeOp        = "o"
  ChannelModePrivate   = "p"
  ChannelModeSecret    = "s"
  ChannelModeInvite    = "i"
  ChannelModeTopicOp   = "t"
  ChannelModeNoOutside = "n"
  ChannelModeModerated = "m"
  ChannelModeUserLimit = "l"
  ChannelModeBanMask   = "b"
  ChannelModeVoice     = "v"
  ChannelModeKey       = "k"
)

type ModeFlag string

type Mode map[ModeFlag]bool

type Channel struct {
  Name string

  Mode Mode

  Topic string

  Limit int
  Key   string

  BanNick string
  BanUser string
  BanHost string

  Clients map[int64]bool
  Ops     map[int64]bool
  Voice   map[int64]bool
}

// A set of all of the valid modes recognized by the server.
var ValidChannelModes = map[ModeFlag]bool{
  ChannelModeOp:        true,
  ChannelModePrivate:   true,
  ChannelModeSecret:    true,
  ChannelModeInvite:    true,
  ChannelModeTopicOp:   true,
  ChannelModeNoOutside: true,
  ChannelModeModerated: true,
  ChannelModeUserLimit: true,
  ChannelModeBanMask:   true,
  ChannelModeVoice:     true,
  ChannelModeKey:       true,
}

// GetChannel will return the Channel for a given name and will create one if
// the channel is not already present.
func (d *Dispatcher) GetChannel(name string) *Channel {
  name = Lowercase(name)

  if name[0] != '&' && name[0] != '#' {
    log.Printf("malformed channel name....")
    return nil
  }

  if channel, found := d.channels[name]; found {
    return channel
  }

  mode := make(Mode)
  for _, r := range d.Config.DefaultChannelMode {
    f := ModeFlag(r)
    if ValidChannelModes[f] {
      mode[f] = true
    } else {
      log.Printf("Unknown default channel flag: %s", f)
    }
  }

  d.channelToClient[name] = make(map[int64]bool)

  channel := &Channel{
    Name:    name,
    Mode:    mode,
    Clients: make(map[int64]bool),
    Ops:     make(map[int64]bool),
    Voice:   make(map[int64]bool),
  }
  d.channels[name] = channel
  return channel
}

func (c *Channel) IsBanned(client *Client) bool {
  Todo("Handle banned users")
  return false
}

// CanPrivMsg returns a boolean indicating whether or not the given client has
// permission to message the channel.
func (c *Channel) CanPrivMsg(client *Client) bool {
  if c.Mode[ChannelModeNoOutside] && !c.Clients[client.ID] {
    return false
  }

  if c.Mode[ChannelModeModerated] && !c.Voice[client.ID] && !c.Ops[client.ID] {
    return false
  }

  return true
}

// AddToChannel adds a Client to the given Channel. This method sends the
// appropriate messages. Access control is not provided.
func (d *Dispatcher) AddToChannel(channel *Channel, client *Client) {
  channel.Clients[client.ID] = true
  client.Channels[channel.Name] = true
  d.channelToClient[channel.Name][client.ID] = true

  // Make the client an op if they are the first in the channel.
  if len(channel.Clients) == 1 {
    channel.Ops[client.ID] = true
  }

  joinMsg := Message{
    Prefix:  client.Prefix(),
    Command: CmdJoin,
    Params:  []string{channel.Name},
  }
  for cid := range channel.Clients {
    c := d.clients[cid]
    c.Relay.Inbox <- joinMsg
  }

  d.sendTopic(client, channel)
  d.sendNames(client, channel)
}

// KillChannel unregisters the channel from the dispatcher and deletes all
// associated map entries. This method assumes that there are not currently any
// clients in the channel.
func (d *Dispatcher) KillChannel(channel *Channel) {
  delete(d.channels, channel.Name)
  delete(d.channelToClient, channel.Name)
}

// RemoveFromChannel removes a client from a given channel and sends out the
// appropriate messages.
func (d *Dispatcher) RemoveFromChannel(channel *Channel, client *Client, reason string) {
  partMsg := Message{
    Prefix:  client.Prefix(),
    Command: CmdPart,
    Params:  []string{channel.Name, reason},
  }
  for cid := range channel.Clients {
    d.clients[cid].Relay.Inbox <- partMsg
  }

  // The parting user should see their part message.
  delete(channel.Clients, client.ID)
  delete(client.Channels, channel.Name)
  delete(d.channelToClient[channel.Name], client.ID)

  // Kill the channel if there is no one left in it.
  if len(channel.Clients) == 0 {
    d.KillChannel(channel)
  }
}

func (d *Dispatcher) SendToChannel(channel *Channel, msg Message) {
  for cid := range channel.Clients {
    d.clients[cid].Relay.Inbox <- msg
  }
}