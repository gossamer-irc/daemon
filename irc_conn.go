package main

import (
	"bufio"
	"github.com/gossamer-irc/lib"
	"io"
	"log"
)

type IrcConnectionEvent struct {
	Connection *IrcConnection
	Message    IrcClientMessage
	Err        error
}

type IrcConnection struct {
	ircd     *Ircd
	client   *lib.Client
	sendQ    *lib.SendQ
	reader   *bufio.Reader
	recv     chan<- IrcConnectionEvent
	trans    chan IrcConnectionEvent
	exit     chan struct{}
	onSuffix bool
}

func NewIrcConnection(ircd *Ircd, reader io.Reader, writer io.WriteCloser, recv chan<- IrcConnectionEvent) *IrcConnection {
	irc := &IrcConnection{
		ircd:   ircd,
		reader: bufio.NewReaderSize(reader, 512),
		// TODO: configurable buffer size
		sendQ: lib.NewSendQ(writer, 2048),
		trans: make(chan IrcConnectionEvent),
		recv:  recv,
		exit:  make(chan struct{}),
	}
	go irc.controlLoop()
	go irc.readLoop()
	return irc
}

func (irc *IrcConnection) SetClient(client *lib.Client) {
	irc.client = client
}

func (irc *IrcConnection) Send(msg IrcMessage) {
	irc.sendQ.Write([]byte(msg.ToIrc(irc.ircd)))
	irc.sendQ.Write([]byte("\r\n"))
}

func (irc *IrcConnection) controlLoop() {
	for {
		select {
		case event := <-irc.trans:
			log.Printf("Forwarding event.")
			irc.recv <- event
			break
		case <-irc.exit:
			for _ = range irc.trans {
			}
			return
		case sqErr := <-irc.sendQ.ErrChan():
			if sqErr != nil {
				irc.recv <- IrcConnectionEvent{
					Connection: irc,
					Err:        sqErr,
				}
			}
		}
	}
}

func (irc *IrcConnection) readLoop() {
	for {
		select {
		case <-irc.exit:
			close(irc.trans)
			irc.trans = nil
			return
		default:
			// Attempt a read.
			data, isPrefix, err := irc.reader.ReadLine()
			log.Printf("Read line: %s", string(data))
			if err != nil {
				irc.trans <- IrcConnectionEvent{
					Connection: irc,
					Err:        err,
				}
				return
			}
			if irc.onSuffix {
				// We are skipping these bytes beause they're trailing from a line we already processed.
				if !isPrefix {
					// This is the last trailing segment.
					irc.onSuffix = false
				}
			} else {
				if isPrefix {
					irc.onSuffix = true
				}
				// Process the new line.
				line := string(data)

				// Parse the line into a GenericIrcClientMessage
				generic, valid := ParseIrc(line)
				if !valid {
					log.Printf("Invalid.")
					continue
				}
				irc.trans <- IrcConnectionEvent{
					Connection: irc,
					Message:    InterpretIrc(generic),
				}
			}
		}
	}
}
