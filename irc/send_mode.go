package irc

func (d *Dispatcher) sendChannelMode(client *Client, channel *Channel) {
  var mode string
  for flag, set := range channel.Mode {
    if set {
      mode += string(flag)
    }
  }
  client.Relay.Inbox <- ReplyChannelModeIs.
    WithPrefix(d.Config.Name).
    WithParams(client.Nick, channel.Name, mode)
}