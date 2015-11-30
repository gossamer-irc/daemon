package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
)

type LinkListener struct {
	Ircd     *Ircd
	Listener net.Listener
}

type LinkEvent struct {
	Listener *LinkListener
	Conn     net.Conn
	Server   string
}

func (ircd *Ircd) NewLinkListener(host string, port uint16) *LinkListener {
	listener, err := tls.Listen("tcp", fmt.Sprintf("%s:%d", host, port), &tls.Config{
		Certificates: []tls.Certificate{ircd.tlsCert},
		RootCAs:      ircd.tlsCaPool,
		ClientCAs:    ircd.tlsCaPool,
		ServerName:   ircd.node.Me.Name,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	})
	if err != nil {
		log.Fatalf("TLS Listen error: %s", err)
	}
	ll := &LinkListener{
		Ircd:     ircd,
		Listener: listener,
	}
	ircd.wg.Add(1)
	go ll.Run()
	return ll
}

func (ll *LinkListener) Run() {
	defer ll.Ircd.wg.Done()
	for {
		rawConn, err := ll.Listener.Accept()
		if err != nil {
			log.Printf("Accept() error: %s", err)
			return
		}

		tlsConn := rawConn.(*tls.Conn)

		// TODO: Run this in a separate goroutine to avoid blocking the listener.
		tlsConn.Handshake()
		state := tlsConn.ConnectionState()
		if !state.HandshakeComplete || len(state.PeerCertificates) < 1 {
			tlsConn.Close()
			log.Printf("Aborted connection due to incomplete handshake")
			continue
		}

		log.Printf("Connection from: %s", state.PeerCertificates[0].Subject.CommonName)
		ll.Ircd.linkEvent <- LinkEvent{
			Listener: ll,
			Conn:     rawConn,
			Server:   state.PeerCertificates[0].Subject.CommonName,
		}
	}
}
