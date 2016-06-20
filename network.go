package main

import (
	"fmt"
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
	ircd.ForEachLocalMember(channel, func(conn *IrcConnection, member *lib.Client, _ *lib.Membership) {
		conn.Send(&IrcJoinMessage{
			From: ircd.ClientAsSeenBy(client, member),
			To:   fmt.Sprintf("#%s:%s", channel.Subnet.Name, channel.Name),
		})
		if member == client {
			// Send additional join info.
			ircd.SendTopic(conn, client, channel)
			ircd.SendNames(conn, client, channel)
		}
	})
}

func (ircd *Ircd) OnChannelPart(channel *lib.Channel, client *lib.Client, reason string) {
	panic("unimplemented")
}

func (ircd *Ircd) OnChannelMessage(from *lib.Client, to *lib.Channel, message string) {
	ircd.ForEachLocalMember(to, func(conn *IrcConnection, member *lib.Client, _ *lib.Membership) {
		conn.Send(&IrcChannelMessage{
			From:    ircd.ClientAsSeenBy(from, member),
			To:      fmt.Sprintf("#%s:%s", to.Subnet.Name, to.Name),
			Message: message,
		})
	})
}

func (ircd *Ircd) OnChannelModeChange(channel *lib.Channel, by *lib.Client, delta lib.ChannelModeDelta, memberDelta []lib.MemberModeDelta) {
	ircd.ForEachLocalMember(channel, func(conn *IrcConnection, member *lib.Client, _ *lib.Membership) {
		from := ircd.node.Me.Name
		if by != nil {
			from = ircd.ClientAsSeenBy(by, member).String()
		}
		conn.Send(&IrcChannelModeMessage{
			From: from,
			To:   fmt.Sprintf("#%s:%s", channel.Subnet.Name, channel.Name),
			Mode: lib.StringifyChannelModes(delta, memberDelta, func(target *lib.Client) string {
				return ircd.ClientAsSeenBy(target, member).Nick
			}),
		})
	})
}
