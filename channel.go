package main

import (
	"fmt"
	"github.com/gossamer-irc/lib"
)

type MemberForEachFn func(conn *IrcConnection, member *lib.Client, membership *lib.Membership)

func (ircd *Ircd) ForEachLocalMember(channel *lib.Channel, fn MemberForEachFn) {
	for member, membership := range channel.LocalMember {
		conn, found := ircd.connByClient[member]
		if !found {
			// TODO: better error handling for missing client connections
			continue
		}
		fn(conn, member, membership)
	}
}

func (ircd *Ircd) SendTopic(conn *IrcConnection, client *lib.Client, channel *lib.Channel) {
	chanName := fmt.Sprintf("#%s:%s", channel.Subnet.Name, channel.Name)
	conn.Send(IrcTopicNumericMessage{
		To:      client.Nick,
		Channel: chanName,
		Topic:   channel.Topic,
	})
	conn.Send(IrcTopicOriginMessage{
		To:      client.Nick,
		Channel: chanName,
		Author:  channel.TopicBy,
		Ts:      uint64(channel.TopicTs.Unix()),
	})
}

func (ircd *Ircd) SendNames(conn *IrcConnection, client *lib.Client, channel *lib.Channel) {
	chanName := fmt.Sprintf("#%s:%s", channel.Subnet.Name, channel.Name)
	name := make([]IrcChannelNameEntry, 0, 15)
	tlen := 0
	for member, membership := range channel.Member {
		prefix := ""
		if membership.IsOwner {
			prefix = "~"
		} else if membership.IsAdmin {
			prefix = "&"
		} else if membership.IsOp {
			prefix = "@"
		} else if membership.IsHalfop {
			prefix = "%"
		} else if membership.IsVoice {
			prefix = "+"
		}
		entry := IrcChannelNameEntry{
			Nick:   ircd.ClientAsSeenBy(member, client).Nick,
			Prefix: prefix,
		}
		name = append(name, entry)
		tlen += len(entry.Nick)
		if tlen > 300 {
			tlen = 0
			conn.Send(IrcChannelNamesReply{
				To:      client.Nick,
				Channel: chanName,
				Name:    name,
			})
			name = name[:0]
		}
	}
	if len(name) > 0 {
		conn.Send(IrcChannelNamesReply{
			To:      client.Nick,
			Channel: chanName,
			Name:    name,
		})
	}
}
