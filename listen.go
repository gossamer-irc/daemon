package main

import (
	"crypto/tls"
	"fmt"
	"net"
)

type Listener struct {
	Ircd     *Ircd
	Host     string
	Port     uint16
	Tls      bool
	Listener net.Listener
	connChan chan<- *Connection
}

type Connection struct {
	NetConn net.Conn
	Login   string
	Err     error
}

func (ircd *Ircd) NewListener(host string, port uint16, tls bool) *Listener {
	listener := &Listener{
		Ircd:     ircd,
		Host:     host,
		Port:     port,
		Tls:      tls,
		connChan: ircd.newConn,
	}
	go listener.run()
	return listener
}

func (l *Listener) run() {
	netListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", l.Host, l.Port))
	if err != nil {
		l.connChan <- &Connection{
			Err: err,
		}
		return
	}
	if l.Tls {
		netListener = tls.NewListener(netListener, &tls.Config{
			Certificates: []tls.Certificate{l.Ircd.tlsCert},
			RootCAs:      l.Ircd.tlsCaPool,
			ClientCAs:    l.Ircd.tlsCaPool,
			ServerName:   l.Ircd.node.Me.Name,
			ClientAuth:   tls.VerifyClientCertIfGiven,
		})
	}
	l.Listener = netListener
	for {
		conn, err := l.Listener.Accept()
		if err != nil {
			l.connChan <- &Connection{
				Err: err,
			}
			return
		}
		l.connChan <- &Connection{
			NetConn: conn,
		}
	}
}
