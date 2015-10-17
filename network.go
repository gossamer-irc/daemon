package main

import (
	"github.com/gossamer-irc/lib"
)

func (ircd *Ircd) OnServerLink(server *lib.Server, hub *lib.Server) {
}

func (ircd *Ircd) OnPrivateMessage(from *lib.Client, to *lib.Client, message string) {
	conn, found := ircd.connByClient[to]
	if !found {
		return
	}
	conn.Send(&IrcPrivateMessage{
		From:    ircd.ClientAsSeenBy(from, to),
		To:      to.Nick,
		Message: message,
	})
}

func (ircd *Ircd) OnChannelJoin(channel *lib.Channel, client *lib.Client, membership *lib.Membership) {
}

func (ircd *Ircd) OnChannelMessage(from *lib.Client, to *lib.Channel, message string) {
}

func (ircd *Ircd) OnChannelModeChange(channel *lib.Channel, by *lib.Client, delta []lib.MemberModeDelta) {
}
