package main

import (
	"github.com/gossamer-irc/lib"
	"strings"
)

type PendingClient struct {
	Ircd   *Ircd
	Conn   *IrcConnection
	Subnet *lib.Subnet
	Nick   string
	Ident  string
	Gecos  string
	Host   string
}

func NewPendingClient(ircd *Ircd, conn *IrcConnection, subnet *lib.Subnet, host string) *PendingClient {
	return &PendingClient{
		Ircd:   ircd,
		Subnet: subnet,
		Conn:   conn,
		Host:   host,
	}
}

func (pc *PendingClient) Handle(raw IrcClientMessage) {
	switch msg := raw.(type) {
	case *InvalidIrcClientMessage:
		break
	case *NickIrcClientMessage:
		// Check whether this nick is taken.
		lnick := strings.ToLower(msg.Nick)
		_, found := pc.Subnet.Client[lnick]
		if found {
			pc.Conn.Send(&IrcNickInUse{msg.Nick})
			return
		}
		pc.Nick = msg.Nick
		pc.CheckReady()
		break
	case *UserIrcClientMessage:
		pc.Ident = msg.Ident
		pc.Gecos = msg.Gecos
		pc.CheckReady()
		break
	}
}

func (pc *PendingClient) CheckReady() {
	if pc.Nick == "" || pc.Ident == "" || pc.Gecos == "" {
		return
	}
	lnick := strings.ToLower(pc.Nick)
	_, found := pc.Subnet.Client[lnick]
	if found {
		pc.Conn.Send(&IrcNickInUse{pc.Nick})
		pc.Nick = ""
		return
	}
	pc.Ircd.AcceptPendingClient(pc)
}
