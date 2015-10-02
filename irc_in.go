package main

import (
	"fmt"
	"strconv"
	"strings"
)

type IrcClientMessage interface {
	isIrcClientMessage() bool
}

type GenericIrcClientMessage struct {
	Command string
	Args    []string
}

func ParseIrc(line string) (msg *GenericIrcClientMessage, valid bool) {
	split := strings.Split(line, " ")
	if len(split) == 0 {
		return
	}
	command := strings.ToUpper(split[0])
	args := split[1:]

	remainingStart := len(args)
	for idx, arg := range args {
		if arg[0:1] == ":" {
			remainingStart = idx
			args[idx] = args[idx][1:]
			break
		}
	}

	if remainingStart < len(args) {
		remainder := args[remainingStart:]
		args = args[0 : remainingStart+1]
		args[remainingStart] = strings.Join(remainder, " ")
	}

	msg = &GenericIrcClientMessage{
		Command: command,
		Args:    args,
	}
	valid = true
	return
}

func (msg GenericIrcClientMessage) isIrcClientMessage() bool {
	return true
}

func (msg GenericIrcClientMessage) String() string {
	return fmt.Sprintf("irc(%s, [%s]", msg.Command, strings.Join(msg.Args, ", "))
}

type InvalidIrcClientMessage struct {
	Command string
	MinArgs uint
	Error   string
}

func (msg InvalidIrcClientMessage) isIrcClientMessage() bool {
	return true
}

func (msg InvalidIrcClientMessage) String() string {
	return fmt.Sprintf("invalid(%s, %d)", msg.Command, msg.MinArgs)
}

type NickIrcClientMessage struct {
	Nick string
}

func (msg NickIrcClientMessage) isIrcClientMessage() bool {
	return true
}

func (msg NickIrcClientMessage) String() string {
	return fmt.Sprintf("nick(%s)", msg.Nick)
}

type UserIrcClientMessage struct {
	Ident, Gecos string
}

func (msg UserIrcClientMessage) isIrcClientMessage() bool {
	return true
}

func (msg UserIrcClientMessage) String() string {
	return fmt.Sprintf("user(%s, %s)", msg.Ident, msg.Gecos)
}

type PMIrcClientMessage struct {
	To      string
	Message string
}

func (msg PMIrcClientMessage) isIrcClientMessage() bool {
	return true
}

func (msg PMIrcClientMessage) String() string {
	return fmt.Sprintf("pm(%s, %s", msg.To, msg.Message)
}

type ConnectIrcClientMessage struct {
	Target string
	Host   string
	Port   uint16
}

func (msg ConnectIrcClientMessage) isIrcClientMessage() bool {
	return true
}

func (msg ConnectIrcClientMessage) String() string {
	return fmt.Sprintf("connect(%s, %d)", msg.Host, msg.Port)
}

func InterpretIrc(msg *GenericIrcClientMessage) IrcClientMessage {
	switch msg.Command {
	case "NICK":
		if len(msg.Args) < 1 {
			return &InvalidIrcClientMessage{
				Command: "NICK",
				MinArgs: 1,
			}
		}
		return &NickIrcClientMessage{
			Nick: msg.Args[0],
		}
	case "USER":
		if len(msg.Args) < 4 {
			return &InvalidIrcClientMessage{
				Command: "USER",
				MinArgs: 4,
			}
		}
		return &UserIrcClientMessage{
			Ident: msg.Args[0],
			Gecos: msg.Args[3],
		}
	case "PRIVMSG":
		if len(msg.Args) < 2 {
			return &InvalidIrcClientMessage{
				Command: "USER",
				MinArgs: 2,
			}
		}
		return &PMIrcClientMessage{
			To:      msg.Args[0],
			Message: msg.Args[1],
		}
	case "CONNECT":
		if len(msg.Args) < 3 {
			return &InvalidIrcClientMessage{
				Command: "CONNECT",
				MinArgs: 3,
			}
		}
		port64, err := strconv.ParseUint(msg.Args[2], 10, 16)
		if err != nil {
			return &InvalidIrcClientMessage{
				Command: "CONNECT",
				MinArgs: 3,
				Error:   fmt.Sprintf("Bad port: %s", msg.Args[2]),
			}
		}
		port := uint16(port64)
		return &ConnectIrcClientMessage{
			Target: msg.Args[0],
			Host:   msg.Args[1],
			Port:   port,
		}
	default:
		return msg
	}
}
