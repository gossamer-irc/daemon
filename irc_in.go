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

type ChannelIrcClientMessage struct {
	To      string
	Message string
}

func (msg ChannelIrcClientMessage) isIrcClientMessage() bool {
	return true
}

func (msg ChannelIrcClientMessage) String() string {
	return fmt.Sprintf("chanmsg(%s, %s", msg.To, msg.Message)
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

type JoinIrcClientMessage struct {
	Targets []string
	Keys    []string
}

func (msg JoinIrcClientMessage) isIrcClientMessage() bool {
	return true
}

func (msg JoinIrcClientMessage) String() string {
	return fmt.Sprintf("join([%s], [%s])", strings.Join(msg.Targets, ", "), strings.Join(msg.Keys, ", "))
}

type ChannelModeChangeIrcClientMessage struct {
	Target string
	Mode   string
	Arg    []string
}

func (msg ChannelModeChangeIrcClientMessage) isIrcClientMessage() bool {
	return true
}

func (msg ChannelModeChangeIrcClientMessage) String() string {
	return fmt.Sprintf("chmode(%s, %s, [%s])", msg.Target, msg.Mode, strings.Join(msg.Arg, ", "))
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
		if strings.HasPrefix(msg.Args[0], "#") {
			return &ChannelIrcClientMessage{
				To:      msg.Args[0],
				Message: msg.Args[1],
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
	case "JOIN":
		if len(msg.Args) < 1 {
			return &InvalidIrcClientMessage{
				Command: "JOIN",
				MinArgs: 1,
			}
		}
		var keys []string = nil
		if len(msg.Args) > 1 {
			keys = strings.Split(msg.Args[0], ",")
		}
		return &JoinIrcClientMessage{
			Targets: strings.Split(msg.Args[0], ","),
			Keys:    keys,
		}
	case "MODE":
		if len(msg.Args) < 1 {
			return &InvalidIrcClientMessage{
				Command: "MODE",
				MinArgs: 1,
			}
		}
		tname := msg.Args[0]
		if strings.HasPrefix(tname, "#") {
			// Channel mode.
			if len(msg.Args) == 1 {
				// Request for the current channel mode.
				return msg
			} else {
				return &ChannelModeChangeIrcClientMessage{
					Target: tname,
					Mode:   msg.Args[1],
					Arg:    msg.Args[2:],
				}
			}
		} else {
			return msg
		}
	default:
		return msg
	}
}
