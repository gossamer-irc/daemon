package main

import (
	"github.com/gossamer-irc/lib"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type Ircd struct {
	node    *lib.Node
	time    time.Time
	newConn chan *Connection

	connEvent chan IrcConnectionEvent
	linkEvent chan LinkEvent

	clientByConn map[*IrcConnection]*lib.Client
	connByClient map[*lib.Client]*IrcConnection
	pending      map[*IrcConnection]*PendingClient

	tlsCert   tls.Certificate
	tlsCaPool *x509.CertPool
}

func NewIrcd(network, server, serverDesc, subnet string) (ircd *Ircd) {
	config := lib.Config{
		NetName:           network,
		ServerName:        server,
		ServerDesc:        serverDesc,
		DefaultSubnetName: subnet,
	}
	ircd = &Ircd{
		time:         time.Now(),
		newConn:      make(chan *Connection),
		connEvent:    make(chan IrcConnectionEvent),
		linkEvent:    make(chan LinkEvent),
		clientByConn: make(map[*IrcConnection]*lib.Client),
		connByClient: make(map[*lib.Client]*IrcConnection),
		pending:      make(map[*IrcConnection]*PendingClient),
	}
	ircd.node = lib.NewNode(config, ircd)
	return
}

func (ircd *Ircd) LoadTls(caFile, certFile, keyFile string) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load TLS certificate: %s", err)
	}
	ircd.tlsCert = cert

	caBytes, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("Failed to load TLS CA certificate: %s", err)
	}
	ircd.tlsCaPool = x509.NewCertPool()
	log.Printf("Ca bytes: %d", len(caBytes))
	ok := ircd.tlsCaPool.AppendCertsFromPEM(caBytes)
	if !ok {
		log.Fatalf("Failed to load TLS CA certificate: invalid")
	}
}

func (ircd *Ircd) AcceptPendingClient(pc *PendingClient) {
	delete(ircd.pending, pc.Conn)
	client := &lib.Client{
		Nick:   pc.Nick,
		Ident:  pc.Ident,
		Host:   pc.Host,
		Gecos:  pc.Gecos,
		Subnet: pc.Subnet,
	}
	err := ircd.node.AttachClient(client)
	if err != nil {
		log.Printf("Error during attach: %s", err)
		return
	}
	ircd.clientByConn[pc.Conn] = client
	ircd.connByClient[client] = pc.Conn

	// Send the welcome.
	pc.Conn.Send(&IrcWelcomeBanner{client.Nick, client.Ident, client.Host})
	pc.Conn.Send(&IrcWelcomeHost{client.Nick})
	pc.Conn.Send(&IrcWelcomeCreated{client.Nick})
	pc.Conn.Send(&IrcWelcomeSupportedModes{client.Nick})
}

func (ircd *Ircd) Run() {
	for {
		select {
		case conn := <-ircd.newConn:
			if conn.Err != nil {
				log.Printf("Error: %s", conn.Err)
				continue
			}
			irc := NewIrcConnection(ircd, conn.NetConn, conn.NetConn, ircd.connEvent)
			pc := NewPendingClient(ircd, irc, ircd.node.DefaultSubnet, conn.NetConn.RemoteAddr().String())
			ircd.pending[irc] = pc
		case event := <-ircd.connEvent:
			if event.Err != nil {
				log.Printf("Error: %s", event.Err)
				continue
			}
			pc, found := ircd.pending[event.Connection]
			if found {
				pc.Handle(event.Message)
				continue
			}
			client, found := ircd.clientByConn[event.Connection]
			if found {
				ircd.Handle(event.Connection, client, event.Message)
			}
		case event := <-ircd.linkEvent:
			log.Printf("Connection from %s", event.Server)
			ircd.node.BeginLink(event.Conn, event.Conn, nil)
		}
	}
}

func (ircd *Ircd) FindClientByRef(context *lib.Client, ref string) (client *lib.Client, found bool) {
	parts := strings.SplitN(ref, ":", 2)
	search := context.Subnet
	nick := parts[0]
	if len(parts) > 1 {
		search, found = ircd.node.Subnet[strings.ToLower(parts[0])]
		if !found {
			return
		}
		nick = parts[1]
	}

	client, found = search.Client[strings.ToLower(nick)]
	return
}

func (_ *Ircd) ClientAsSeenBy(client, context *lib.Client) IrcNIH {
	if client.Subnet == context.Subnet {
		return IrcNIH{client.Nick, client.Ident, client.Host}
	}
	return IrcNIH{fmt.Sprintf("%s:%s", client.Subnet.Name, client.Nick), client.Ident, client.Host}
}

func (ircd *Ircd) Handle(irc *IrcConnection, client *lib.Client, rawEvent IrcClientMessage) {
	switch event := rawEvent.(type) {
	case *PMIrcClientMessage:
		// Lookup the recepient.
		to, found := ircd.FindClientByRef(client, event.To)
		if !found {
			// TODO: Send error saying they're not found.
			return
		}
		ircd.node.PrivateMessage(client, to, event.Message)
	case *ConnectIrcClientMessage:
		ircd.InitiateConnection(event.Target, event.Host, event.Port)
	}
}

func (ircd *Ircd) InitiateConnection(target, host string, port uint16) {
	// TODO: Move off the main ircd goroutine.
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", host, port), &tls.Config{
		Certificates: []tls.Certificate{ircd.tlsCert},
		RootCAs:      ircd.tlsCaPool,
		ServerName:   target,
	})
	if err != nil {
		log.Printf("Error linking to %s:%d: %s", host, port, err)
		return
	}
	ircd.node.BeginLink(conn, conn, nil)
}
